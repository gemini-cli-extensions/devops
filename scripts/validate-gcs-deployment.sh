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

GCS_BUCKET="${1:-$GCS_BUCKET}"
GCS_BUCKET="${GCS_BUCKET#gs://}"

if [ -z "$GCS_BUCKET" ]; then
  cat <<EOF
{
  "score": 0.0,
  "details": "GCS bucket name not provided as argument and GCS_BUCKET environment variable is not set",
  "checks": [
    {"name": "env-var", "passed": false, "message": "GCS bucket name not provided as argument and GCS_BUCKET environment variable is not set"}
  ]
}
EOF
  exit 1
fi

CHECK_EXISTS="false"
MSG_EXISTS="Bucket does not exist"
CHECK_CONTAINS="false"
MSG_CONTAINS="Bucket is empty or could not be listed"
CHECK_PUBLIC="false"
MSG_PUBLIC="Could not verify public access"

# Check 1: Bucket exists
if gcloud storage buckets describe gs://$GCS_BUCKET &> /dev/null; then
  CHECK_EXISTS="true"
  MSG_EXISTS="Bucket $GCS_BUCKET exists"
  
  # Check 2: Bucket contains files
  FILES=$(gcloud storage ls gs://$GCS_BUCKET 2>/dev/null)
  if [ -n "$FILES" ]; then
    CHECK_CONTAINS="true"
    MSG_CONTAINS="Bucket contains files"
    
    # Check 3: Public access
    # Find a file to test
    TARGET_FILE=""
    if echo "$FILES" | grep -q "gs://$GCS_BUCKET/index.html"; then
      TARGET_FILE="index.html"
    else
      TARGET_FILE=$(echo "$FILES" | head -n 1 | sed "s|gs://$GCS_BUCKET/||")
    fi
    
    if [ -n "$TARGET_FILE" ]; then
      URL="https://storage.googleapis.com/$GCS_BUCKET/$TARGET_FILE"
      HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$URL")
      
      if [ "$HTTP_CODE" = "200" ]; then
        CHECK_PUBLIC="true"
        MSG_PUBLIC="Successfully accessed $URL (HTTP 200)"
      else
        MSG_PUBLIC="Failed to access $URL (HTTP $HTTP_CODE)"
      fi
    else
      MSG_PUBLIC="Could not determine a target file to test"
    fi
  else
    MSG_CONTAINS="Bucket is empty"
  fi
fi

# Calculate score
TOTAL_CHECKS=3
PASSED_CHECKS=0
[ "$CHECK_EXISTS" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_CONTAINS" == "true" ] && ((PASSED_CHECKS++))
[ "$CHECK_PUBLIC" == "true" ] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "bucket-exists", "passed": $CHECK_EXISTS, "message": "$MSG_EXISTS"},
    {"name": "contains-files", "passed": $CHECK_CONTAINS, "message": "$MSG_CONTAINS"},
    {"name": "public-access", "passed": $CHECK_PUBLIC, "message": "$MSG_PUBLIC"}
  ]
}
EOF
