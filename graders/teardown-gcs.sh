#!/bin/bash

# Check if GCS_BUCKET is set
if [ -z "$GCS_BUCKET" ]; then
  cat <<EOF
{
  "score": 0.0,
  "details": "GCS_BUCKET environment variable is not set"
}
EOF
  exit 1
fi

PROJECT_ARG=""
if [ -n "$PROJECT_ID" ]; then
  PROJECT_ARG="--project=$PROJECT_ID"
fi

# Attempt to delete the bucket and its contents
# We use gcloud storage rm -r to delete objects.
# We ignore errors because the bucket might be empty or not exist.
gcloud storage rm -r gs://$GCS_BUCKET/** $PROJECT_ARG &> /dev/null

# Delete the bucket
gcloud storage buckets delete gs://$GCS_BUCKET $PROJECT_ARG --quiet &> /dev/null

# Check if the bucket still exists
if gcloud storage buckets describe gs://$GCS_BUCKET $PROJECT_ARG &> /dev/null; then
  cat <<EOF
{
  "score": 0.0,
  "details": "Failed to delete bucket $GCS_BUCKET"
}
EOF
  exit 1
else
  cat <<EOF
{
  "score": 1.0,
  "details": "Successfully deleted bucket $GCS_BUCKET"
}
EOF
  exit 0
fi
