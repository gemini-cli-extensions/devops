#!/usr/bin/env python3
#
# Reusable script to validate GKE manifests without yq.

import argparse
import json
import os


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate GKE manifests in Python."
    )
    parser.add_argument(
        "dir", nargs="?", default=".", help="Directory containing files."
    )

    args = parser.parse_args()

    deployment_ok = False
    service_ok = False
    
    for root, dirs, files in os.walk(args.dir):
        for file in files:
            if file != "clouddeploy.yaml":
                full_path = os.path.join(root, file)
                try:
                    with open(full_path, "r") as f:
                        content = f.read()
                        if "kind: Deployment" in content:
                            deployment_ok = True
                        if "kind: Service" in content:
                            service_ok = True
                        
                        if deployment_ok and service_ok:
                            break
                except Exception:
                    pass
        if deployment_ok and service_ok:
            break

    passed_checks = 0
    total_checks = 2

    if deployment_ok:
        passed_checks += 1
    if service_ok:
        passed_checks += 1

    score = round(passed_checks / total_checks, 2)

    print(
        json.dumps(
            {
                "score": score,
                "details": f"{passed_checks}/{total_checks} manifests found",
                "checks": [
                    {"name": "deployment-manifest", "passed": deployment_ok},
                    {"name": "service-manifest", "passed": service_ok},
                ],
            }
        )
    )


if __name__ == "__main__":
    run_grader()
