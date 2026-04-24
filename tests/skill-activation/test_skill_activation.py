import hashlib
import json
import os
import shutil
import subprocess
import yaml

# Get the directory where the script is located
script_dir = os.path.dirname(os.path.abspath(__file__))
# Workspace root (assuming script is in cicd/tests/skill-activation)
workspace_root = os.path.dirname(os.path.dirname(script_dir))

src_skills_dir = os.path.join(workspace_root, "skills")
dest_skills_dir = os.path.join(script_dir, "skills")

# Check if GCS bucket exists
bucket_name = "skill-activation-test-debug"
print(f"Checking if bucket {bucket_name} exists...")
try:
    subprocess.run(
        ["gcloud", "storage", "buckets", "describe", f"gs://{bucket_name}"],
        check=True,
        capture_output=True,
    )
    print("Bucket exists.")
except subprocess.CalledProcessError:
    raise Exception(
        f"Bucket gs://{bucket_name} does not exist. Please create it first or ensure you have access."
    )

print(f"Copying from {src_skills_dir} to {dest_skills_dir}")


# shutil.copytree requires dest to not exist or dirs_exist_ok=True
try:
    shutil.copytree(src_skills_dir, dest_skills_dir, dirs_exist_ok=True)
except TypeError:
    # Fallback for older python
    if os.path.exists(dest_skills_dir):
        shutil.rmtree(dest_skills_dir)
    shutil.copytree(src_skills_dir, dest_skills_dir)

# Iterate through copied skills
for skill_name in os.listdir(dest_skills_dir):
    skill_path = os.path.join(dest_skills_dir, skill_name)
    if not os.path.isdir(skill_path):
        continue

    # Remove references folder
    refs_dir = os.path.join(skill_path, "references")
    if os.path.exists(refs_dir):
        shutil.rmtree(refs_dir)
        print(f"Removed references in {skill_name}")

    # Process SKILL.md
    skill_md_path = os.path.join(skill_path, "SKILL.md")
    if os.path.exists(skill_md_path):
        with open(skill_md_path, "r") as f:
            content = f.read()

        lines = content.splitlines()
        try:
            first_dash = lines.index("---")
            second_dash = lines.index("---", first_dash + 1)
        except ValueError:
            print(f"Warning: Frontmatter not found in {skill_md_path}")
            continue

        header_lines = lines[first_dash + 1 : second_dash]
        new_header_lines = []
        keep = False
        extracted_name = skill_name  # Fallback

        for line in header_lines:
            if (
                line.startswith("name:")
                or line.startswith("version:")
                or line.startswith("description:")
            ):
                keep = True
                new_header_lines.append(line)
                if line.startswith("name:"):
                    extracted_name = line.split(":", 1)[1].strip()
                    # Remove quotes if any
                    extracted_name = extracted_name.strip("'\"")
            elif line.startswith(" ") or line.startswith("\t"):
                if keep:
                    new_header_lines.append(line)
            elif ":" in line:
                keep = False

        # Construct new content
        new_content = "---\n"
        new_content += "\n".join(new_header_lines) + "\n"
        new_content += "---\n\n"
        new_content += (
            f"Simply say '{extracted_name}' and **terminate** without doing anything else!\n"
        )

        with open(skill_md_path, "w") as f:
            f.write(new_content)
            print(f"Updated SKILL.md for {skill_name}")

# Generate skill-activation-eval.yaml
eval_queries_path = os.path.join(script_dir, "eval_queries.json")
eval_yaml_path = os.path.join(script_dir, "skill-activation-eval.yaml")

if os.path.exists(eval_queries_path):
    print(f"Reading {eval_queries_path}")
    with open(eval_queries_path, "r") as f:
        queries_data = json.load(f)

    tasks = []
    for skill_name, queries in queries_data.items():
        for item in queries:
            query = item["query"]
            should_trigger = item["should_trigger"]

            # Calculate hash
            h = hashlib.md5(query.encode()).hexdigest()[:8]
            task_name = f"{h}-{str(should_trigger).lower()}"

            task = {
                "name": task_name,
                "instruction": query,
                "trialConfig": {
                    "setup": f"git clone https://github.com/sdlc-graph/tinyjam.git /tmp/skill-activation/{h}",
                    "cleanup": f"rm -rf /tmp/skill-activation/{h}",
                },
                "agentWorkingDir": f"/tmp/skill-activation/{h}",
                "graders": [],
            }

            if should_trigger:
                task["graders"].append(
                    {
                        "type": "tool_usage",
                        "expectedTools": [
                            {"name": "activate_skill", "args": {"name": skill_name}}
                        ],
                    }
                )
            else:
                task["graders"].append(
                    {
                        "type": "llm_rubric",
                        "outcome_assertions": [
                            f"Is the exact skill '{skill_name}' not being activated?"
                        ],
                    }
                )

            tasks.append(task)

    output_data = {
        "version": "1",
        "defaults": {
            "agent": "gemini",
            "provider": "local",
            "trials": 5,
            "timeout": 60,
            "threshold": 0.8,
            "grader_model": "gemini-3-flash-preview",
            "docker": {"base": "cicd-evals:latest"},
            "env": {
                "GOOGLE_APPLICATION_CREDENTIALS": "~/.config/gcloud/application_default_credentials.json"
            },
        },
        "tasks": tasks,
    }

    print(f"Writing {eval_yaml_path}")
    with open(eval_yaml_path, "w") as f:
        yaml.dump(output_data, f, sort_keys=False)

    # Bucket check moved to top

    # Run skillgrade
    print("Running skillgrade...")
    try:
        subprocess.run(
            [
                "skillgrade",
                "--config=skill-activation-eval.yaml",
                "--no-redact",
                f"--output=gs://{bucket_name}",
            ],
            check=True,
            cwd=script_dir,
        )
    except subprocess.CalledProcessError as e:
        print(f"Error running skillgrade: {e}")

    # Remove skills folder
    print(f"Removing {dest_skills_dir}")
    if os.path.exists(dest_skills_dir):
        shutil.rmtree(dest_skills_dir)

    # Remove eval yaml file
    print(f"Removing {eval_yaml_path}")
    if os.path.exists(eval_yaml_path):
        os.remove(eval_yaml_path)

else:
    print(f"Warning: {eval_queries_path} not found.")
