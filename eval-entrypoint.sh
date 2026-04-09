#!/bin/bash
set -e

ENV_FILE="/etc/eval-env.sh"
echo "#!/bin/bash" > "$ENV_FILE"

# Try to get ADC token, but don't fail the script if it fails
if TOKEN=$(gcloud auth application-default print-access-token 2>/tmp/adc-error.log); then
    export CLOUDSDK_AUTH_ACCESS_TOKEN="$TOKEN"
    echo "export CLOUDSDK_AUTH_ACCESS_TOKEN=\"$TOKEN\"" >> "$ENV_FILE"
else
    echo "Warning: Failed to obtain ADC token. Credentials may be invalid." >&2
    echo "Error details saved to /etc/eval-adc-error.log" >&2
    cp /tmp/adc-error.log /etc/eval-adc-error.log
    echo "export ADC_AUTH_ERROR=\"true\"" >> "$ENV_FILE"
fi

chmod +x "$ENV_FILE"

exec "$@"

