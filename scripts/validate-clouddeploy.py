#!/usr/bin/env python3
#
# Reusable script to validate clouddeploy.yaml using native YAML parsing.

import argparse
import json
import os
import sys
import yaml


def run_grader():
    parser = argparse.ArgumentParser(
        description="Validate clouddeploy.yaml using native YAML parsing."
    )
    parser.add_argument(
        "--runtime",
        choices=["run", "gke"],
        default="run",
        help="Target runtime.",
    )
    parser.add_argument(
        "--check-automations",
        action="store_true",
        help="Check for automations.",
    )
    parser.add_argument(
        "--check-canary",
        action="store_true",
        help="Check for canary strategy.",
    )
    parser.add_argument(
        "dir", nargs="?", default=".", help="Directory containing files."
    )

    args = parser.parse_args()

    cloud_deploy_file = os.path.join(args.dir, "clouddeploy.yaml")

    if not os.path.exists(cloud_deploy_file):
        print(json.dumps({"score": 0.0, "details": "clouddeploy.yaml missing"}))
        return

    try:
        with open(cloud_deploy_file, "r") as stream:
            docs = list(yaml.safe_load_all(stream))

        pipeline = next(
            (d for d in docs if d and d.get("kind") == "DeliveryPipeline"),
            None,
        )
        targets = [d for d in docs if d and d.get("kind") == "Target"]
        automations = [d for d in docs if d and d.get("kind") == "Automation"]

        pipeline_ok = pipeline is not None

        targets_ok = False
        if args.runtime == "run":
            targets_ok = len(targets) > 0 and all(
                t.get("run") is not None for t in targets
            )
        elif args.runtime == "gke":
            targets_ok = len(targets) > 0 and all(
                t.get("gke") is not None for t in targets
            )

        names_ok = any(
            "staging" in t.get("metadata", {}).get("name", "") for t in targets
        ) and any(
            "prod" in t.get("metadata", {}).get("name", "") for t in targets
        )

        automation_ok = True
        if args.check_automations:
            automation_ok = len(automations) > 0

        canary_ok = True
        if args.check_canary:
            canary_ok = False
            if pipeline:
                stages = pipeline.get("serialPipeline", {}).get("stages", [])
                for s in stages:
                    strategy = s.get("strategy", {})
                    if "canary" in strategy:
                        canary = strategy["canary"]
                        percentages = canary.get("canaryDeployment", {}).get("percentages", [])
                        if args.runtime == "gke":
                            kubernetes = canary.get("runtimeConfig", {}).get("kubernetes", {})
                            service_networking = kubernetes.get("serviceNetworking", {})
                            if service_networking.get("service") and service_networking.get("deployment") and len(percentages) > 0:
                                canary_ok = True
                                break

        passed_checks = 0
        total_checks = 3
        if args.check_automations:
            total_checks += 1
        if args.check_canary:
            total_checks += 1

        if pipeline_ok:
            passed_checks += 1
        if targets_ok:
            passed_checks += 1
        if names_ok:
            passed_checks += 1
        if automation_ok and args.check_automations:
            passed_checks += 1
        if canary_ok and args.check_canary:
            passed_checks += 1

        score = round(passed_checks / total_checks, 2)

        checks = [
            {"name": "pipeline-exists", "passed": pipeline_ok},
            {"name": "runtime-correct", "passed": targets_ok},
            {"name": "targets-named-correctly", "passed": names_ok},
        ]

        if args.check_automations:
            checks.append({"name": "automations-exist", "passed": automation_ok})
        if args.check_canary:
            checks.append({"name": "canary-configured", "passed": canary_ok})

        print(
            json.dumps(
                {
                    "score": score,
                    "details": f"{passed_checks}/{total_checks} checks passed",
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
