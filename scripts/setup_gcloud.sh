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

# Create a secure temporary file
TOKEN_FILE=$(mktemp)

# Get the access token and write it to the file
if gcloud auth application-default print-access-token > "$TOKEN_FILE"; then
    # Set the gcloud property
    gcloud config set auth/access_token_file "$TOKEN_FILE"
    echo "Successfully set auth/access_token_file to $TOKEN_FILE"
else
    echo "Failed to get access token" >&2
    rm -f "$TOKEN_FILE"
    exit 1
fi
