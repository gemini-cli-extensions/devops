#!/usr/bin/env bash
#
# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# Initialize checks
CHECK_EXISTS="false"
MSG_EXISTS="Not checked"
CHECK_MULTI_STAGE="false"
MSG_MULTI_STAGE="Not checked"

# Check for required environment variables
MISSING_VARS=()
[ -z "$APP_DIR" ] && MISSING_VARS+=("APP_DIR")

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
  MSG_EXISTS="Missing environment variables: ${MISSING_VARS[*]}"
else
  DOCKER_FILE="$APP_DIR/Dockerfile"
  # 1. Check if Dockerfile exists
  if [ -f "$DOCKER_FILE" ]; then
    CHECK_EXISTS="true"
    MSG_EXISTS="Dockerfile exists"
  else
    CHECK_EXISTS="false"
    MSG_EXISTS="Dockerfile does not exist"
  fi

  # 2. Check if Dockerfile is multi-stage
  COUNT=$(grep -ic "^FROM " $DOCKER_FILE)
  if [ "$COUNT" -gt 1 ]; then
    CHECK_MULTI_STAGE="true"
    MSG_MULTI_STAGE="Dockerfile is multi-stage"
  else
    CHECK_MULTI_STAGE="false"
    MSG_MULTI_STAGE="Dockerfile is not multi-stage"
  fi

  # 3. Check if Dockerfile syntax
  if docker build --check -f $DOCKER_FILE $APP_DIR > /dev/null 2>&1; then
    CHECK_SYNTAX="true"
    MSG_SYNTAX="Dockerfile syntax ok"
  else
    CHECK_SYNTAX="false"
    MSG_SYNTAX="Dockerfile syntax error"
  fi
  
fi

# Calculate score
TOTAL_CHECKS=3
PASSED_CHECKS=0
[ "$CHECK_EXISTS" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_MULTI_STAGE" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_SYNTAX" == "true" ] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "dockerfile-exists", "passed": $CHECK_EXISTS, "message": "$MSG_EXISTS"},
    {"name": "dockerfile-multistage", "passed": $CHECK_MULTI_STAGE, "message": "$MSG_MULTI_STAGE"},
    {"name": "dockerfile-syntax", "passed": $CHECK_SYNTAX, "message": "$MSG_SYNTAX"}
  ]
}
EOF
