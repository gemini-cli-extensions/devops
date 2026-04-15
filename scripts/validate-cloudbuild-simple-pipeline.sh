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

# Check for Test step
if grep -q -i "id:.*test" "$FILE"; then
  CHECK_TEST="true"
fi

# Check for Build step
if grep -q -i "id:.*build" "$FILE"; then
  CHECK_BUILD="true"
fi

# Check for Push step
if grep -q -i "id:.*push" "$FILE"; then
  CHECK_PUSH="true"
fi

# Check for Deploy step
if grep -q -i "id:.*deploy" "$FILE"; then
  CHECK_DEPLOY="true"
fi

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
