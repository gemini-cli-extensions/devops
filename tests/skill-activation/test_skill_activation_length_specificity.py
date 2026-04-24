import csv
import time

import glob
import hashlib
import json
import os
import re
import shutil
import subprocess
import yaml
import matplotlib.pyplot as plt
import numpy as np

# Get directory of the current script
script_dir = os.path.dirname(os.path.abspath(__file__))
# Workspace root
workspace_root = os.path.dirname(os.path.dirname(script_dir))

src_skills_dir = os.path.join(workspace_root, "skills")
dest_skills_dir = os.path.join(script_dir, "skills")

json_path = os.path.join(script_dir, "eval_queries_length_specificity.json")
output_path = os.path.join(script_dir, "skill-activation-length-specificity-eval.yaml")


def get_hash(text):
    return hashlib.md5(text.encode("utf-8")).hexdigest()


def copy_and_modify_skills():
    print(f"Copying from {src_skills_dir} to {dest_skills_dir}")
    try:
        shutil.copytree(src_skills_dir, dest_skills_dir, dirs_exist_ok=True)
    except TypeError:
        if os.path.exists(dest_skills_dir):
            shutil.rmtree(dest_skills_dir)
        shutil.copytree(src_skills_dir, dest_skills_dir)

    for skill_name in os.listdir(dest_skills_dir):
        skill_path = os.path.join(dest_skills_dir, skill_name)
        if not os.path.isdir(skill_path):
            continue

        refs_dir = os.path.join(skill_path, "references")
        if os.path.exists(refs_dir):
            shutil.rmtree(refs_dir)
            print(f"Removed references in {skill_name}")

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
            extracted_name = skill_name

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
                        extracted_name = extracted_name.strip("'\"")
                elif line.startswith(" ") or line.startswith("\t"):
                    if keep:
                        new_header_lines.append(line)
                elif ":" in line:
                    keep = False

            new_content = "---\n"
            new_content += "\n".join(new_header_lines) + "\n"
            new_content += "---\n\n"
            new_content += f"Simply say '{extracted_name}' and **terminate** without doing anything else!\n"

            with open(skill_md_path, "w") as f:
                f.write(new_content)
                print(f"Updated SKILL.md for {skill_name}")


def generate_yaml():
    with open(json_path, "r") as f:
        prompts_data = json.load(f)["google-cicd-deploy"]

    eval_data = {
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
        "tasks": [],
    }

    skill = "google-cicd-deploy"

    for item in prompts_data:
        length = item["length"]
        specificity = item["specificity"]
        prompt = item["prompt"]
        prompt_hash = get_hash(prompt)

        task = {
            "name": f"SAT_{skill}_length-{length}_specificity-{specificity}-{prompt_hash}",
            "instruction": prompt,
            "trialConfig": {
                "setup": f"git clone https://github.com/CoasterJX/tinyjam.git /tmp/skill-activation/{prompt_hash}",
                "cleanup": f"rm -rf /tmp/skill-activation/{prompt_hash}",
            },
            "agentWorkingDir": f"/tmp/skill-activation/{prompt_hash}",
            "graders": [
                {
                    "type": "tool_usage",
                    "expectedTools": [
                        {"name": "activate_skill", "args": {"name": skill}}
                    ],
                }
            ],
        }
        eval_data["tasks"].append(task)

    with open(output_path, "w") as f:
        yaml.dump(eval_data, f, default_flow_style=False, sort_keys=False, indent=2)

    print(f"Generated {output_path} with {len(eval_data['tasks'])} tasks.")


def generate_heatmap(bucket_name, local_dir="local_results", output_img="heatmap.png"):
    os.makedirs(local_dir, exist_ok=True)

    print("Downloading files from GCS...")
    # Using gsutil -m cp with glob to download all json files
    cmd = f"gsutil -m cp gs://{bucket_name}/**.json {local_dir}/"
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Warning or Error downloading files: {result.stderr}")

    files = glob.glob(os.path.join(local_dir, "*.json"))
    print(f"Found {len(files)} JSON files.")

    grid = np.full((10, 6), np.nan)
    pattern = re.compile(r"length-(\d+)_specificity-(\d+)")

    for file_path in files:
        filename = os.path.basename(file_path)
        match = pattern.search(filename)
        if match:
            length = int(match.group(1))
            specificity = int(match.group(2))

            try:
                with open(file_path, "r") as f:
                    data = json.load(f)
                    pass_rate = data.get("pass_rate")

                    if pass_rate is not None:
                        if 1 <= length <= 10 and 1 <= specificity <= 6:
                            grid[length - 1, specificity - 1] = pass_rate
            except Exception as e:
                print(f"Error reading {filename}: {e}")

    fig, ax = plt.subplots(figsize=(8, 10))
    im = ax.imshow(grid, cmap="RdYlGn", vmin=0.0, vmax=1.0)

    ax.set_xticks(np.arange(6))
    ax.set_yticks(np.arange(10))
    ax.set_xticklabels(np.arange(1, 7))
    ax.set_yticklabels(np.arange(1, 11))

    ax.set_xlabel("Specificity")
    ax.set_ylabel("Length")
    ax.set_title("Skillgrade Pass Rate Heatmap")

    cbar = ax.figure.colorbar(im, ax=ax)
    cbar.ax.set_ylabel("Pass Rate", rotation=-90, va="bottom")

    for i in range(10):
        for j in range(6):
            val = grid[i, j]
            if not np.isnan(val):
                text = ax.text(
                    j, i, f"{val:.2f}", ha="center", va="center", color="black"
                )
            else:
                text = ax.text(j, i, "N/A", ha="center", va="center", color="gray")

    fig.tight_layout()
    plt.savefig(output_img)
    print(f"Heatmap successfully saved to {output_img}")

    # Upload to bucket
    print(f"Uploading {output_img} to gs://{bucket_name}")
    cmd = f"gsutil cp {output_img} gs://{bucket_name}/"
    subprocess.run(cmd, shell=True, check=True)

    # Cleanup
    shutil.rmtree(local_dir)
    os.remove(output_img)
    print("Cleaned up local results and heatmap.")


if __name__ == "__main__":
    bucket_name = "skill-activation-length-specificity-debug"
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

    copy_and_modify_skills()
    generate_yaml()

    print("Running skillgrade...")
    try:
        # Start skillgrade as a background process
        process = subprocess.Popen(
            [
                "skillgrade",
                "--config=skill-activation-length-specificity-eval.yaml",
                "--no-redact",
                f"--output=gs://{bucket_name}",
            ],
            cwd=script_dir,
        )

        # Periodically generate heatmap while skillgrade is running
        while process.poll() is None:
            print("Updating heatmap (waiting 5 minutes)...")
            generate_heatmap(bucket_name)
            time.sleep(300)  # 300 seconds = 5 minutes

        # Run one last time after skillgrade finishes to catch the final results
        print("Skillgrade finished. Generating final heatmap...")
        generate_heatmap(bucket_name)

    except Exception as e:
        print(f"Error running skillgrade: {e}")

    # Cleanup
    print(f"Removing {dest_skills_dir}")
    if os.path.exists(dest_skills_dir):
        shutil.rmtree(dest_skills_dir)

    print(f"Removing {output_path}")
    if os.path.exists(output_path):
        os.remove(output_path)

    print("Done.")
