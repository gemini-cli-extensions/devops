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
