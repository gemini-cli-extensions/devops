#!/usr/bin/env python3

import argparse
import json
import os
import re
import yaml


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate Cloud Build configurations."
    )
    parser.add_argument(
        "file",
        nargs="?",
        default="cloudbuild.yaml",
        help="Path to cloudbuild.yaml",
    )
    args = parser.parse_args()

    if not os.path.exists(args.file):
        print(
            json.dumps({"score": 0.0, "details": f"File {args.file} not found"})
        )
        return

    try:
        with open(args.file, "r") as f:
            config = yaml.safe_load(f)
    except Exception as e:
        print(
            json.dumps(
                {"score": 0.0, "details": f"Failed to parse YAML: {str(e)}"}
            )
        )
        return

    check_test = False
    check_build = False
    check_push = False
    check_deploy = False

    steps = config.get("steps", [])
    for step in steps:
        step_id = step.get("id") or ""
        step_name = step.get("name") or ""
        step_args = " ".join(step.get("args") or [])

        # Test check
        if (
            re.search(r"[Tt]est", step_id)
            or re.search(r"npm|python|node|pytest", step_name)
            or re.search(r"[Tt]est", step_args)
        ):
            check_test = True

        # Build check
        if re.search(r"[Bb]uild", step_id) or (
            "docker" in step_name and "build" in step_args
        ):
            check_build = True

        # Push check
        if re.search(r"[Pp]ush", step_id) or (
            "docker" in step_name and "push" in step_args
        ):
            check_push = True

        # Deploy check
        if re.search(r"[Dd]eploy", step_id) or (
            re.search(r"gcloud|cloud-sdk", step_name)
            and re.search(r"deploy|run", step_args)
        ):
            check_deploy = True

    total_checks = 4
    passed_checks = 0
    if check_test:
        passed_checks += 1
    if check_build:
        passed_checks += 1
    if check_push:
        passed_checks += 1
    if check_deploy:
        passed_checks += 1

    score = round(passed_checks / total_checks, 2)

    print(
        json.dumps(
            {
                "score": score,
                "details": f"{passed_checks}/{total_checks} checks passed",
                "checks": [
                    {"name": "test-step", "passed": check_test},
                    {"name": "build-step", "passed": check_build},
                    {"name": "push-step", "passed": check_push},
                    {"name": "deploy-step", "passed": check_deploy},
                ],
            }
        )
    )


if __name__ == "__main__":
    run_grader()
