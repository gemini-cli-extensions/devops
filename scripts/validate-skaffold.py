#!/usr/bin/env python3
#
# Reusable script to validate skaffold.yaml using native YAML parsing.

import argparse
import json
import os
import yaml


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate skaffold.yaml using native YAML parsing."
    )
    parser.add_argument(
        "--profiles", default="staging,prod", help="Expected profiles."
    )
    parser.add_argument(
        "dir", nargs="?", default=".", help="Directory containing files."
    )

    args = parser.parse_args()

    skaffold_file = os.path.join(args.dir, "skaffold.yaml")

    if not os.path.exists(skaffold_file):
        print(json.dumps({"score": 0.0, "details": "skaffold.yaml missing"}))
        return

    try:
        with open(skaffold_file, "r") as stream:
            doc = yaml.safe_load(stream) or {}

        profile_list = args.profiles.split(",")
        
        passed_checks = 0
        total_checks = len(profile_list)

        doc_profiles = doc.get("profiles", [])

        for p in profile_list:
            if any(dp.get("name") == p for dp in doc_profiles):
                passed_checks += 1

        score = round(passed_checks / total_checks, 2)

        checks = [
            {
                "name": f"profile-{p}",
                "passed": any(dp.get("name") == p for dp in doc_profiles),
            }
            for p in profile_list
        ]

        print(
            json.dumps(
                {
                    "score": score,
                    "details": f"{passed_checks}/{total_checks} profiles found",
                    "checks": checks,
                }
            )
        )

    except Exception as e:
        print(
            json.dumps(
                {"score": 0.0, "details": f"Error parsing YAML: {str(e)}"}
            )
        )


if __name__ == "__main__":
    run_grader()
