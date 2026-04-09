#!/bin/bash

# Check if GCS_BUCKET is set
if [ -z "$GCS_BUCKET" ]; then
  echo "Error: GCS_BUCKET environment variable is not set" >&2
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
  echo "Error: Failed to delete bucket $GCS_BUCKET" >&2
  exit 1
else
  echo "Successfully deleted bucket $GCS_BUCKET"
  exit 0
fi
