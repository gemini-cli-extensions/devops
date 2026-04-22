---
name: author-evals
description: A skill to help author evaluations for skillgrade, including eval.yaml, grader scripts, environment variables, and test project setup.
---

# Authoring Evals for SkillGrade

This skill provides guidelines and best practices for authoring evaluations using the `skillgrade` tool. Evaluations are defined in an `eval.yaml` file and used to test agent capabilities.

## Identifying Target Skills

When authoring an evaluation, first ask the user which skill (or skills) should be evaluated.
The skills are located in the `./skills` directory for the project.
This information is crucial for determining what evaluation cases would be useful.

## Generating Eval Cases

The skill should support generating/suggesting and generating multiple evaluation cases to ensure the skill is well tested.
A single task is usually not enough. Aim for 3-5 distinct tasks that cover different aspects of the skill.

## Eval Configuration (`eval.yaml`)

The `eval.yaml` file defines the tasks, environment, and graders for evaluation.

### Key Sections

- **`defaults`**: Set base configuration for all tasks.
    - `agent`: Model to use (e.g., `gemini`).
    - `provider`: Execution environment (`docker` or `local`).
    - `docker`: Base image and setup commands.
    - `env`: Base environment variables.
    - `environment.mounts`: Bind mounts for Docker.
- **`tasks`**: List of evaluation tasks.
    - `name`: Unique task name.
    - `instruction`: Clear instructions for the agent, including expected output file names.
    - `workspace`: Files to include in the container. Use isolation; do not assume files are present.
    - `trialConfig`: Setup and cleanup hooks for each trial.
    - `graders`: Graders used to evaluate performance.

### Workspace Isolation

By default, **no files are auto-included** in the execution environment. You must explicitly specify all files and directories needed for the task in the `workspace` section.

Example:
```yaml
workspace:
  - src: fixtures/broken-app.js
    dest: app.js
  - content: |
      # Gemini Custom Instructions
      Always use TypeScript for new files.
    dest: /root/.gemini/GEMINI.md
```

## Graders

Authoring effective graders is key to reliable evaluations.

### 1. Deterministic Graders (Recommended)

Deterministic graders run a script or command and evaluate the outcome objectively. They are more reliable and faster than LLM graders.

- **Output Format**: The grader script MUST output a JSON object to `stdout` with at least `score` (0.0 to 1.0) and `details` (string).
- **Bash Example**:
    ```bash
    #!/bin/bash
    # Grade outcome, not steps!
    if test -f output.txt; then
      echo '{"score": 1.0, "details": "Output file exists"}'
    else
      echo '{"score": 0.0, "details": "Output file missing"}'
    fi
    ```
- **Rule**: Use `awk` for arithmetic instead of `bc`, as `bc` is not available in standard `node:20-slim` images.

### 2. LLM Rubric Graders

LLM rubric graders are used for qualitative criteria.

- **Format**: Outcome assertions are recommended for binary checks. Free-text rubrics are also supported.
- **Outcome Assertions Example**:
    ```yaml
    - type: llm_rubric
      outcome_assertions:
        - Did the agent follow the mandatory 3-step workflow?
        - Did it complete in <=5 commands?
      weight: 0.3
    ```

### 3. Tool Usage Graders

Verifies that the agent called specific tools during the trial.

Example:
```yaml
- type: tool_usage
  expectedTools:
    - name: read_file
    - name: write_to_file
      args:
        path: test.txt
```

Final reward is a weighted sum of all grader scores.

## Environment Variables

Proper environment variable configuration ensures trials are isolated and reproducible.

### Dynamic Variables

Use the `{{trial}}` placeholder in environment variables to create unique values per trial (e.g., ports, resource names). SkillGrade also auto-injects:
- `_EVAL_TRIAL`: Current trial ID.
- `_EVAL_UUID`: Unique run identifier.

Example usage in instructions or scripts:
- `PROJECT_ID`: Project to target.
- `REGION`: Deployment region.
- `gs://${_EVAL_UUID}-${_EVAL_TRIAL}`: Unique bucket name.

## IAM Permissions and Test Project Setup

When evaluations interact with cloud resources (like Google Cloud in this project), suggest required permissions and setup.

### IAM Permissions

Suggest minimal permissions required for the test:
- **Storage Admin** for GCS bucket operations.
- **Cloud Run Admin** for service deployments.
- **Artifact Registry Admin** for image uploads.

Advise the user to use limited service accounts rather than Owner roles.

### Setup and Cleanup Scripts

- **Setup scripts**: Clone necessary repositories, install dependencies, or provision base resources.
- **Cleanup scripts**: Crucial for deleting resources created during the evaluation (e.g., GCS buckets, Cloud Run services) to avoid leaks and costs.

Example Cleanup script in task:
```yaml
trialConfig:
  cleanup: scripts/teardown-gcs.sh gs://${_EVAL_UUID}-${_EVAL_TRIAL}
```

## Best Practices

- **Grade outcomes, not steps.** Check that the file was fixed, not that the agent ran a specific command.
- **Instructions must name output files.** The grader must know what to check.
- **Start small.** 3-5 well-designed tasks beat 50 noisy ones.

## Improving Evals based on Results

This process outlined in this section is for analyzing evaluation results to improve the core instructions and setup of the eval itself, rather than the skill file.

### Workflow

1.  **Gather Results**: Locate the evaluation results (local or GCS).
2.  **Analyze Logs**: Look at the agent JSON stream and execution logs.
3.  **Identify Inefficiencies**:
    - Is the agent looping?
    - Is it trying out unnecessary tools not directly related to the skill?
    - Are the instructions ambiguous?
    - **Handle Environment Issues**: If the failure was caused by the test environment (e.g., missing IAM permissions, disabled APIs, missing tools), write a local file (e.g., `ENV_ISSUES.md`) to the workspace root detailing the issue and required action, and do not update the eval to fix it unless the eval setup itself is flawed.
4.  **Formulate Improvements**:
    - Suggest a better set of `GEMINI.md` or agent instructions (either base instructions available to all evals or per-eval instructions).
    - Refine instructions to be more specific about expected outcomes or file names.
    - Adjust the `eval.yaml` configuration (e.g., workspace, timeouts) if needed.
    - Suggest updates to `eval.yaml` to remove requirements that conflict with mandatory tool capabilities, rather than forcing agents to bypass tools.
5.  **Rule of No Tool Bias**: When improving eval instructions based on results, do NOT force the agent to use specific skills or tools. Eval instructions must test the outcome, not the specific steps or tools used. For example, check that a file was fixed, not that a specific linting tool was run.
6.  **Rule of No Implementation Bias (Avoid Overfitting)**: Avoid over-constraining instructions with specific implementation steps (e.g., 'must be defined as a resource in your Terraform') just to pass a rigid deterministic grader. Instructions must focus on high-level patterns and outcomes required by the skill, allowing the agent to make logical decisions based on context (e.g., whether to create new resources or reference existing ones), rather than forcing a specific path that leads to overfitting.
7.  **Differentiate from Skill Improvement**: Remember that `improve-skills` is for improving the *skill itself* (making it more robust). This skill is for improving the *eval setup* to make testing more effective and efficient.

## Relationship with `improve-skills`

The `author-evals` and `improve-skills` skills are related and often used together:
- `author-evals` focuses on authoring and improving **evaluation configurations and instructions**.
- `improve-skills` focuses on improving the **skills themselves** based on eval results.

When an evaluation failure occurs, analyze whether it is due to:
- An **eval flaw** (unclear instructions, wrong setup) → Use `author-evals` to improve the eval.
- A **skill gap** (missing instructions, lack of best practices in the skill) → Use `improve-skills` to improve the skill file.

Be aware that both might be needed to achieve high-quality results.
