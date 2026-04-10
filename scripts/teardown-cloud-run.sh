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
