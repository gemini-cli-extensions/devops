# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Check required environment variables
if [ -z "$PROJECT_ID" ]; then
    echo "Error: PROJECT_ID environment variable is required."
    exit 1
fi
if [ -z "$REGION" ]; then
    echo "Error: REGION environment variable is required."
    exit 1
fi
if [ -z "$ARTIFACT_REGISTRY" ]; then
    echo "Error: ARTIFACT_REGISTRY environment variable is required."
    exit 1
fi

# Create .env file
echo "Creating .env file..."
echo "PROJECT_ID=$PROJECT_ID" > .env
echo "REGION=$REGION" >> .env
echo "ARTIFACT_REGISTRY=$ARTIFACT_REGISTRY" >> .env

# Update eval.yaml to remove GOOGLE_APPLICATION_CREDENTIALS (metadata server is available in CI)
echo "Removing GOOGLE_APPLICATION_CREDENTIALS from eval.yaml..."
yq -i 'del(.defaults.env.GOOGLE_APPLICATION_CREDENTIALS)' eval.yaml

# Build MCP server
echo "Building MCP server..."
./build.sh

# Configure Gemini CLI
echo "Configuring Gemini CLI..."
mkdir -p ~/.gemini
cat <<EOF > ~/.gemini/settings.json
{
  "mcpServers": {
    "cicd": {
      "command": "$(pwd)/cicd-mcp-server/cicd-mcp-server",
      "timeout": 300000,
      "trust": true
    }
  }
}
EOF

cat <<EOF > ~/.gemini/trustedFolders.json
{
  "/": "TRUST_PARENT"
}
EOF

# Runs the evals
echo "Running evals with local provider..."

OUTPUT_PATH=$1
if [ -z "$OUTPUT_PATH" ]; then
    echo "Error: Output GCS path is required."
    echo "Usage: $0 <output_gcs_path>"
    exit 1
fi

EVAL_EXIT_CODE=0
skillgrade --ci --provider=local --output="$OUTPUT_PATH" --no-redact || EVAL_EXIT_CODE=$?

echo "Evals results are available to view by running \`skillgrade preview --output=\"$OUTPUT_PATH\"\`"

exit $EVAL_EXIT_CODE
