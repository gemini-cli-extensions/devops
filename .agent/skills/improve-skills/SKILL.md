---
name: improve-skills
description: Analyze evaluation results to identify gaps and improve skills with general-purpose updates. Avoids biasing updates solely for evaluation success.
---

# Skill Improvement from Evaluation Analysis

This skill outlines the process for analyzing evaluation results to improve the core skills used by the agent. It emphasizes making robust, general improvements rather than "hacking" the skills to pass specific evaluation checks.

## Workflow

### Step 1: Gather Evaluation Results
1.  Locate the evaluation results. These may be in a local directory or a Google Cloud Storage bucket.
2.  If the results are in GCS, use `gsutil` or `gcloud storage` to list and copy the files locally for analysis.
3.  Read the result JSON files to understand the tasks run, pass rates, and specific failure details in the `grader_results`.

### Step 2: Analyze Against Configuration (`eval.yaml`)
1.  Read the `eval.yaml` file in the repository root to understand the specific requirements and expected behaviors for each task.
2.  Pay attention to the `graders` section, especially `tool_usage` requirements (e.g., expected tools).
3.  Compare the actual tool calls and outputs in the result logs with the expectations in `eval.yaml`.

### Step 3: Identify Skill Gaps
1.  Locate the current skills in the `skills/` directory.
2.  Analyze the relevant skill file (e.g., `SKILL.md`) to see if it covers the failed scenario or if the instructions are unclear.
3.  Identify gaps where the skill file lacks clarity on mandatory steps, tool usage, or best practices that led to the failure.

### Step 4: Formulate General Improvements
1.  **Rule of No Eval Bias**: When updating a skill, do NOT reference "evaluation scores", "evals", or specific evaluation tasks.
2.  Focus on **Scope and Applicability**: Update descriptions to follow a consistent pattern in the YAML front matter:
    - High-level description.
    - Why it's important to activate (e.g., security best practices not handled by tools alone).
    - When to activate (triggers).
3.  **Move Triggers to Description**: Include skill activation triggers within the description field of the YAML front matter to make them immediately visible to the model and encourage activation.
4.  **Soft but Clear Wording**: Avoid overly harsh terms like "CRITICAL" or "MANDATORY" in skills, but clearly state that the skill *should* be activated to ensure best practices are followed.
5.  Focus on **Best Practices**: Add instructions that ensure correct and secure operation (e.g., mandatory security scans, resource checks).
6.  Ensure the improvements are general enough to help in direct user interactions while also addressing the root cause that caused the evaluation to fail.

## Constraints & Rules
*   **Do not hardcode evaluation hacks**: Never update a skill to simply check for a specific file named in an eval or to use a specific tool just because the eval requires it, unless that tool is a best practice for the workflow.
*   **Maintain General Utility**: Skills must remain useful for a human user interacting with the agent.
