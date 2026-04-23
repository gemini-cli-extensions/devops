#!/usr/bin/env python3
#
# Reusable script to validate Terraform configurations for pipeline design evals.

import argparse
import hcl2
import json
import os
import re
import subprocess


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate Terraform configurations for pipeline design evals."
    )
    parser.add_argument(
        "dir", nargs="?", default=".", help="Directory containing files."
    )

    args = parser.parse_args()

    if not os.path.exists(args.dir):
        print(json.dumps({"score": 0.0, "details": f"Directory {args.dir} not found"}))
        return

    error_details = []
    # Run init and validate
    check_validate = False
    try:
        # Run init with -backend=false to avoid GCS access
        init_res = subprocess.run(
            ["terraform", "init", "-backend=false"],
            cwd=args.dir,
            capture_output=True,
            text=True,
        )
        if init_res.returncode == 0:
            val_res = subprocess.run(
                ["terraform", "validate"],
                cwd=args.dir,
                capture_output=True,
                text=True,
            )
            if val_res.returncode == 0:
                check_validate = True
            else:
                error_details.append(f"terraform validate failed: {val_res.stderr.strip()}")
        else:
            error_details.append(f"terraform init failed: {init_res.stderr.strip()}")
    except Exception as e:
        error_details.append(f"Failed to execute terraform: {e}")

    # Resource checks based on file content
    expected_resources = [
        "google_artifact_registry_repository",
        "google_cloudbuild_trigger",
        "google_service_account",
        "google_project_iam_member",
        "google_developer_connect_connection",
        "google_developer_connect_git_repository_link",
    ]

    found_resources = {r: False for r in expected_resources}

    # Scan .tf files
    for root, dirs, files in os.walk(args.dir):
        dirs[:] = [d for d in dirs if d not in [".terraform", ".git"]]
        for file in files:
            if file.endswith(".tf"):
                full_path = os.path.join(root, file)
                try:
                    with open(full_path, "r") as f:
                        dict_content = hcl2.load(f)
                        if "resource" in dict_content:
                            for resource_block in dict_content["resource"]:
                                for r_type in resource_block.keys():
                                    if r_type in expected_resources:
                                        found_resources[r_type] = True
                except Exception as e:
                    error_details.append(f"Failed to parse {full_path}: {e}")

    # Calculate score
    total_checks = len(expected_resources) + 1  # +1 for validate
    passed_checks = 0

    if check_validate:
        passed_checks += 1

    for r in expected_resources:
        if found_resources[r]:
            passed_checks += 1

    score = round(passed_checks / total_checks, 2)

    checks = [{"name": "terraform-validate", "passed": check_validate}]
    for r in expected_resources:
        checks.append(
            {
                "name": f"resource-{r.replace('_', '-')}",
                "passed": found_resources[r],
            }
        )

    details = f"{passed_checks}/{total_checks} checks passed"
    if error_details:
        details += f". Errors: {'; '.join(error_details)}"

    print(
        json.dumps(
            {
                "score": score,
                "details": details,
                "checks": checks,
            }
        )
    )


if __name__ == "__main__":
    run_grader()
