#!/usr/bin/env python3
#
# Reusable script to validate Cloud Run manifests without yq.

import argparse
import json
import os


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate Cloud Run manifests in Python."
    )
    parser.add_argument(
        "--services",
        default="service-staging,service-prod",
        help="Expected services.",
    )
    parser.add_argument(
        "dir", nargs="?", default=".", help="Directory containing files."
    )

    args = parser.parse_args()

    service_list = args.services.split(",")
    passed_checks = 0
    total_checks = len(service_list)

    found_services = {s: False for s in service_list}
    
    for root, dirs, files in os.walk(args.dir):
        for file in files:
            if file != "clouddeploy.yaml":
                full_path = os.path.join(root, file)
                try:
                    with open(full_path, "r") as f:
                        content = f.read()
                        for s, found in found_services.items():
                            if not found and s in content:
                                found_services[s] = True
                except Exception:
                    pass

    passed_checks = sum(1 for found in found_services.values() if found)
    score = round(passed_checks / total_checks, 2)

    checks = [
        {"name": f"manifest-{s}", "passed": found_services[s]}
        for s in service_list
    ]

    print(
        json.dumps(
            {
                "score": score,
                "details": f"{passed_checks}/{total_checks} services found in manifests",
                "checks": checks,
            }
        )
    )


if __name__ == "__main__":
    run_grader()
