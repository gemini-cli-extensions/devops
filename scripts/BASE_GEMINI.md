# Baseline Instructions for Automated Evaluations

You are running in an automated evaluation environment. The run is headless and non-interactive (invoked with flags like `--yolo` and `-p`),
but we want to evaluate cases where input is provided by the user that you should pull from additional context. Always try to complete the task.

## Context Loading

1.  **Workspace Specifics**: Check the current working directory or workspace root for a `GEMINI.md` file. If present, read it immediately as it contains task-specific instructions and constraints that override or supplement these general guidelines.
2.  **Environment**: Check for environment variables that clearly define properties needed to complete the task.

## Execution Guidelines

1.  **No Tool Guessing Loops**: If a tool call fails or does not produce the expected result, do not enter a loop trying random combinations of tools to try and fix the issue. If you are blocked or cannot proceed based on the instructions, state the blocker clearly and stop.
2.  **Stick to Skills**: Rely strictly on the guidance provided in the skills or task definitions. Do not attempt to invent workflows or use tools in ways not documented.
3.  **Fail Fast**: If the goal cannot be achieved with the available tools and instructions, report the failure immediately.