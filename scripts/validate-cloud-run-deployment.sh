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

# Support both PROJECT_NAME and PROJECT_ID
PROJECT_NAME=${PROJECT_NAME:-$PROJECT_ID}

# Initialize checks
CHECK_EXISTS="false"
MSG_EXISTS="Not checked"
CHECK_REACHABLE="false"
MSG_REACHABLE="Not checked"

# Check for required environment variables
MISSING_VARS=()
[ -z "$CLOUD_RUN_SERVICE" ] && MISSING_VARS+=("CLOUD_RUN_SERVICE")
[ -z "$PROJECT_NAME" ] && MISSING_VARS+=("PROJECT_NAME/PROJECT_ID")
[ -z "$REGION" ] && MISSING_VARS+=("REGION")

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
  MSG_EXISTS="Missing environment variables: ${MISSING_VARS[*]}"
else
  # 1. Check if service exists
  if gcloud run services describe "$CLOUD_RUN_SERVICE" --project="$PROJECT_NAME" --region="$REGION" &>/dev/null; then
    CHECK_EXISTS="true"
    MSG_EXISTS="Service exists"


    # 3. Check Reachable
    URL=$(gcloud run services describe "$CLOUD_RUN_SERVICE" \
      --project="$PROJECT_NAME" \
      --region="$REGION" \
      --format='value(status.url)')

    if [ -n "$URL" ]; then
      CURL_OPTS="-s -o /dev/null -w %{http_code}"
      HTTP_STATUS=$(curl $CURL_OPTS "$URL")

      if [ "$HTTP_STATUS" == "000" ]; then
        CHECK_REACHABLE="false"
        MSG_REACHABLE="Could not connect to service URL"
      elif [[ "$HTTP_STATUS" =~ ^5 ]]; then
        CHECK_REACHABLE="false"
        MSG_REACHABLE="Service returned server error (HTTP $HTTP_STATUS)"
      else
        CHECK_REACHABLE="true"
        MSG_REACHABLE="Service is reachable (HTTP $HTTP_STATUS)"
      fi
    else
      CHECK_REACHABLE="false"
      MSG_REACHABLE="Service URL not found"
    fi

  else
    CHECK_EXISTS="false"
    MSG_EXISTS="Service does not exist or is not accessible"
    MSG_REACHABLE="Skipped: Service does not exist"
  fi
fi

# Calculate score
TOTAL_CHECKS=2
PASSED_CHECKS=0
[ "$CHECK_EXISTS" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_REACHABLE" == "true" ] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "service-exists", "passed": $CHECK_EXISTS, "message": "$MSG_EXISTS"},
    {"name": "service-reachable", "passed": $CHECK_REACHABLE, "message": "$MSG_REACHABLE"}
  ]
}
EOF
