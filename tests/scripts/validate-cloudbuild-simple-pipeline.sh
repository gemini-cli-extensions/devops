#!/usr/bin/env bash

# Initialize checks
CHECK_TEST="false"
CHECK_BUILD="false"
CHECK_PUSH="false"
CHECK_DEPLOY="false"

FILE="${1:-cloudbuild.yaml}"

if [ ! -f "$FILE" ]; then
  echo "{\"score\": 0, \"details\": \"$FILE not found\"}"
  exit 0
fi

# Use yq to parse cloudbuild.yaml robustly
while IFS="|" read -r id name args; do
  # Check for Test step
  if [[ "$id" =~ [Tt]est ]] || [[ "$name" =~ (npm|python|node|pytest) ]] || [[ "$args" =~ [Tt]est ]]; then
    CHECK_TEST="true"
  fi

  # Check for Build step
  if [[ "$id" =~ [Bb]uild ]] || ([[ "$name" =~ docker ]] && [[ "$args" =~ build ]]); then
    CHECK_BUILD="true"
  fi

  # Check for Push step
  if [[ "$id" =~ [Pp]ush ]] || ([[ "$name" =~ docker ]] && [[ "$args" =~ push ]]); then
    CHECK_PUSH="true"
  fi

  # Check for Deploy step
  if [[ "$id" =~ [Dd]eploy ]] || ([[ "$name" =~ (gcloud|cloud-sdk) ]] && [[ "$args" =~ (deploy|run) ]]); then
    CHECK_DEPLOY="true"
  fi
done < <(yq '.steps[] | (.id // "") + "|" + (.name // "") + "|" + ((.args // []) | join(" "))' "$FILE")

# Calculate score
TOTAL_CHECKS=4
PASSED_CHECKS=0
[ "$CHECK_TEST" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_BUILD" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_PUSH" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_DEPLOY" == "true" ] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "test-step", "passed": $CHECK_TEST},
    {"name": "build-step", "passed": $CHECK_BUILD},
    {"name": "push-step", "passed": $CHECK_PUSH},
    {"name": "deploy-step", "passed": $CHECK_DEPLOY}
  ]
}
EOF
