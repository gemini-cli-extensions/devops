#!/usr/bin/env bash

# Initialize checks
CHECK_VALIDATE="false"
CHECK_PLAN="false"

DIR="${1:-.}"
cd "$DIR" || { echo "{\"score\": 0, \"details\": \"Directory $DIR not found\"}"; exit 0; }

# Create backend override to avoid GCS access
cat <<EOF > backend_override.tf
terraform {
  backend "local" {}
}
EOF

# Run init and validate
if terraform init -backend=false &>/dev/null; then
  if terraform validate &>/dev/null; then
    CHECK_VALIDATE="true"
  fi
fi

# For plan, we need to use the override backend and actual init
rm -rf .terraform
if terraform init &>/dev/null; then
  # Try plan without vars
  if terraform plan -input=false -out=plan.out &>/dev/null; then
    CHECK_PLAN="true"
  fi
fi

# Initialize resource checks
CHECK_REPO="false"
CHECK_TRIGGER="false"
CHECK_SA="false"
CHECK_IAM="false"
CHECK_CONN="false"
CHECK_LINK="false"

if [ "$CHECK_PLAN" == "true" ]; then
  SHOW_OUTPUT=$(terraform show plan.out)
  
  echo "$SHOW_OUTPUT" | grep -q "google_artifact_registry_repository" && CHECK_REPO="true"
  echo "$SHOW_OUTPUT" | grep -q "google_cloudbuild_trigger" && CHECK_TRIGGER="true"
  echo "$SHOW_OUTPUT" | grep -q "google_service_account" && CHECK_SA="true"
  echo "$SHOW_OUTPUT" | grep -q "google_project_iam_member" && CHECK_IAM="true"
  echo "$SHOW_OUTPUT" | grep -q "google_developer_connect_connection" && CHECK_CONN="true"
  echo "$SHOW_OUTPUT" | grep -q "google_developer_connect_git_repository_link" && CHECK_LINK="true"
fi

# Cleanup
rm -f backend_override.tf plan.out
rm -rf .terraform

# Calculate score
TOTAL_CHECKS=7
PASSED_CHECKS=0
[ "$CHECK_VALIDATE" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_REPO" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_TRIGGER" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_SA" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_IAM" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_CONN" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_LINK" == "true" ] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "terraform-validate", "passed": $CHECK_VALIDATE},
    {"name": "resource-artifact-registry", "passed": $CHECK_REPO},
    {"name": "resource-cloudbuild-trigger", "passed": $CHECK_TRIGGER},
    {"name": "resource-service-account", "passed": $CHECK_SA},
    {"name": "resource-iam-member", "passed": $CHECK_IAM},
    {"name": "resource-dev-connect-connection", "passed": $CHECK_CONN},
    {"name": "resource-dev-connect-link", "passed": $CHECK_LINK}
  ]
}
EOF
