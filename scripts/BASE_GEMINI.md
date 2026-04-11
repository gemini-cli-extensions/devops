# Baseline Instructions for Automated Evaluations

You are running in an automated evaluation environment. The run is headless and
non-interactive (invoked with flags like `--yolo` and `-p`), but we want to
emulate cases where input is provided by the user. For these tests, this
information will not be provided interactively from the user, and you should not
try too hard to discover it. It should be clearly available from either gemini
memory files, or in the environment.

## Strict Prohibitions (Zero Tolerance)

1.  **Limited Parameter Search**: If a required parameter (e.g., Project ID,
    Region, Service Name) is not found in the environment or memory files, you
    may search for and inspect configuration files *within the local
    repository*. However, **DO NOT** search files outside the repository
    directory, and **DO NOT** arbitrarily call cloud tools (like listing
    projects or services) to discover missing information.
2.  **Do Not Attempt Error Recovery**: If a tool fails due to Authentication,
    Permissions, or Billing disabled, **DO NOT** attempt to find keys, change
    regions, search for other projects, or try alternative tools. Report the
    error as the final output and stop.
3.  **No Speculative Tool Use**: Do not try tools just to "see if they work" or
    to gather clues (e.g., listing builds, listing projects) when a previous
    step failed.