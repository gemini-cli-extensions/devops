#!/usr/bin/env bash
#
# Copyright 2026 Google LLC
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

# Initialize checks
CHECK_EXISTS="false"
MSG_EXISTS="Not checked"
CHECK_YAML="false"
MSG_YAML="Not checked"
CHECK_STEPS="false"
MSG_STEPS="Not checked"
CHECK_IMAGE="false"
MSG_IMAGE="Not checked"
CHECK_SA="false"
MSG_SA="Not checked"
CHECK_LOGGING="false"
MSG_LOGGING="Not checked"

# Check for required environment variables
MISSING_VARS=()
[[ -z "$APP_DIR" ]] && MISSING_VARS+=("APP_DIR")

if [[ ${#MISSING_VARS[@]} -ne 0 ]]; then
  MSG_EXISTS="Missing environment variables: ${MISSING_VARS[*]}"
else
  CLOUDBUILD_FILE="$APP_DIR/cloudbuild.yaml"

  # 1. Check if cloudbuild.yaml exists
  if [[ -f "$CLOUDBUILD_FILE" ]]; then
    CHECK_EXISTS="true"
    MSG_EXISTS="cloudbuild.yaml exists"

    # 2. Check if valid YAML and other properties using Python
    # We use python to avoid external dependencies like yq which might not be present.
    # Python is usually available in these environments.

    python3 -c "
import yaml
import sys

def check_cloudbuild(file_path):
    try:
        with open(file_path, 'r') as f:
            data = yaml.safe_load(f)
    except yaml.YAMLError as e:
        print(f\"YAML_INVALID:{e}\")
        sys.exit(0)
    except Exception as e:
        print(f\"ERROR:{e}\")
        sys.exit(0)

    print(\"YAML_VALID\")

    steps = data.get('steps', [])
    step_ids = [s.get('id') for s in steps if s.get('id')]
    
    required_steps = ['build', 'lint', 'test', 'deploy']
    missing_steps = [s for s in required_steps if s not in step_ids]
    
    if not missing_steps:
        print(\"STEPS_OK\")
    else:
        print(f\"STEPS_MISSING:{','.join(missing_steps)}\")

    images = data.get('images', [])
    if images:
        print(\"IMAGE_OK\")
    else:
        print(\"IMAGE_MISSING\")

    sa = data.get('serviceAccount')
    if sa:
        print(\"SA_OK\")
    else:
        print(\"SA_MISSING\")

    options = data.get('options', {})
    logging = options.get('logging')
    
    acceptable_logging = ['CLOUD_LOGGING_ONLY', 'LEGACY', 'STACKDRIVER_ONLY']
    if logging in acceptable_logging:
         print(\"LOGGING_OK\")
    elif not logging:
         # If not specified, it might be ok depending on project settings, 
         # but prompt says 'The logging option specifies cloud logging or both'.
         # We will assume it needs to be explicit for this check to pass.
         print(\"LOGGING_MISSING\")
    else:
         print(f\"LOGGING_INVALID:{logging}\")

check_cloudbuild('$CLOUDBUILD_FILE')
" > $OUTPUT_DIR/cloudbuild_check_output

    # Parse python output
    if grep -q "YAML_VALID" $OUTPUT_DIR/cloudbuild_check_output; then
        CHECK_YAML="true"
        MSG_YAML="cloudbuild.yaml is valid YAML"
    else
        CHECK_YAML="false"
        MSG_YAML=$(grep "YAML_INVALID" $OUTPUT_DIR/cloudbuild_check_output | cut -d: -f2-)
        [[ -z "$MSG_YAML" ]] && MSG_YAML="Invalid YAML format"
    fi

    if grep -q "STEPS_OK" $OUTPUT_DIR/cloudbuild_check_output; then
        CHECK_STEPS="true"
        MSG_STEPS="All required steps (build, lint, test, deploy) found"
    else
        CHECK_STEPS="false"
        MISSING=$(grep "STEPS_MISSING" $OUTPUT_DIR/cloudbuild_check_output | cut -d: -f2-)
        MSG_STEPS="Missing steps: $MISSING"
    fi

    if grep -q "IMAGE_OK" $OUTPUT_DIR/cloudbuild_check_output; then
        CHECK_IMAGE="true"
        MSG_IMAGE="Image defined in images field"
    else
        CHECK_IMAGE="false"
        MSG_IMAGE="No image defined in images field"
    fi

    if grep -q "SA_OK" $OUTPUT_DIR/cloudbuild_check_output; then
        CHECK_SA="true"
        MSG_SA="Specific service account being used"
    else
        CHECK_SA="false"
        MSG_SA="No specific service account defined"
    fi

    if grep -q "LOGGING_OK" $OUTPUT_DIR/cloudbuild_check_output; then
        CHECK_LOGGING="true"
        MSG_LOGGING="Logging option specifies cloud logging or both"
    else
        CHECK_LOGGING="false"
        LOGGING_VAL=$(grep "LOGGING_INVALID" $OUTPUT_DIR/cloudbuild_check_output | cut -d: -f2-)
        if [[ -z "$LOGGING_VAL" ]]; then
             if grep -q "LOGGING_MISSING" $OUTPUT_DIR/cloudbuild_check_output; then
                 MSG_LOGGING="Logging option is missing (not explicit)"
             else
                 MSG_LOGGING="Logging option invalid"
             fi
        else
             MSG_LOGGING="Logging option invalid: $LOGGING_VAL"
        fi
    fi

    rm -f $OUTPUT_DIR/cloudbuild_check_output
  else
    CHECK_EXISTS="false"
    MSG_EXISTS="cloudbuild.yaml does not exist"
  fi
fi

# Calculate score
TOTAL_CHECKS=6
PASSED_CHECKS=0
[[ "$CHECK_EXISTS" == "true" ]] && ((PASSED_CHECKS++))
[[ "$CHECK_YAML" == "true" ]] && ((PASSED_CHECKS++))
[[ "$CHECK_STEPS" == "true" ]] && ((PASSED_CHECKS++))
[[ "$CHECK_IMAGE" == "true" ]] && ((PASSED_CHECKS++))
[[ "$CHECK_SA" == "true" ]] && ((PASSED_CHECKS++))
[[ "$CHECK_LOGGING" == "true" ]] && ((PASSED_CHECKS++))

SCORE=$(awk "BEGIN {printf \"%.2f\", $PASSED_CHECKS / $TOTAL_CHECKS}")

# Construct JSON
cat <<EOF
{
  "score": $SCORE,
  "details": "$PASSED_CHECKS/$TOTAL_CHECKS checks passed",
  "checks": [
    {"name": "cloudbuild-exists", "passed": $CHECK_EXISTS, "message": "$MSG_EXISTS"},
    {"name": "cloudbuild-yaml", "passed": $CHECK_YAML, "message": "$MSG_YAML"},
    {"name": "cloudbuild-steps", "passed": $CHECK_STEPS, "message": "$MSG_STEPS"},
    {"name": "cloudbuild-image", "passed": $CHECK_IMAGE, "message": "$MSG_IMAGE"},
    {"name": "cloudbuild-sa", "passed": $CHECK_SA, "message": "$MSG_SA"},
    {"name": "cloudbuild-logging", "passed": $CHECK_LOGGING, "message": "$MSG_LOGGING"}
  ]
}
EOF
