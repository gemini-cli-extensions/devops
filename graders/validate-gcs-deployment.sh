#!/bin/bash

# Check if GCS_BUCKET is set
if [ -z "$GCS_BUCKET" ]; then
  cat <<EOF
{
  "score": 0.0,
  "details": "GCS_BUCKET environment variable is not set",
  "checks": [
    {"name": "env-var", "passed": false, "message": "GCS_BUCKET environment variable is not set"}
  ]
}
EOF
  exit 1
fi

passed=0
total=3

c1_pass=false
c1_msg="Bucket does not exist"
c2_pass=false
c2_msg="Bucket is empty or could not be listed"
c3_pass=false
c3_msg="Could not verify public access"

# Check 1: Bucket exists
if gcloud storage buckets describe gs://$GCS_BUCKET &> /dev/null; then
  passed=$((passed + 1))
  c1_pass=true
  c1_msg="Bucket $GCS_BUCKET exists"
  
  # Check 2: Bucket contains files
  FILES=$(gcloud storage ls gs://$GCS_BUCKET 2>/dev/null)
  if [ -n "$FILES" ]; then
    passed=$((passed + 1))
    c2_pass=true
    c2_msg="Bucket contains files"
    
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
        passed=$((passed + 1))
        c3_pass=true
        c3_msg="Successfully accessed $URL (HTTP 200)"
      else
        c3_msg="Failed to access $URL (HTTP $HTTP_CODE)"
      fi
    else
      c3_msg="Could not determine a target file to test"
    fi
  else
    c2_msg="Bucket is empty"
  fi
fi

# Calculate score using awk
score=$(awk "BEGIN {printf \"%.2f\", $passed/$total}")

# Output JSON
echo "{\"score\":$score,\"details\":\"$passed/$total checks passed\",\"checks\":[{\"name\":\"bucket-exists\",\"passed\":$c1_pass,\"message\":\"$c1_msg\"},{\"name\":\"contains-files\",\"passed\":$c2_pass,\"message\":\"$c2_msg\"},{\"name\":\"public-access\",\"passed\":$c3_pass,\"message\":\"$c3_msg\"}]}"
