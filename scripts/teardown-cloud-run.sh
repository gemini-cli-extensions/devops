#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Check for required environment variables
if [ -z "$CLOUD_RUN_SERVICE" ]; then
  echo "Error: CLOUD_RUN_SERVICE environment variable is not set."
  exit 1
fi

if [ -z "$PROJECT_ID" ]; then
  echo "Error: PROJECT_ID environment variable is not set."
  exit 1
fi

if [ -z "$REGION" ]; then
  echo "Error: REGION environment variable is not set."
  exit 1
fi

echo "Deleting Cloud Run service $CLOUD_RUN_SERVICE in project $PROJECT_ID and region $REGION..."

gcloud run services delete "$CLOUD_RUN_SERVICE" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --quiet

echo "Cloud Run service $CLOUD_RUN_SERVICE deletion command completed."
