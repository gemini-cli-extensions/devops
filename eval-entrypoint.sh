#!/bin/bash
set -e

ENV_FILE="/etc/eval-env.sh"
echo "#!/bin/bash" > "$ENV_FILE"

if [ -n "$GOOGLE_APPLICATION_CREDENTIALS_CONTENTS" ]; then
    TMP_FILE=$(mktemp /tmp/gcp-XXXXXX.json)
    echo "$GOOGLE_APPLICATION_CREDENTIALS_CONTENTS" > "$TMP_FILE"
    export GOOGLE_APPLICATION_CREDENTIALS="$TMP_FILE"
    echo "export GOOGLE_APPLICATION_CREDENTIALS=\"$TMP_FILE\"" >> "$ENV_FILE"
fi

TOKEN=$(gcloud auth application-default print-access-token)
export CLOUDSDK_AUTH_ACCESS_TOKEN="$TOKEN"
echo "export CLOUDSDK_AUTH_ACCESS_TOKEN=\"$TOKEN\"" >> "$ENV_FILE"

chmod +x "$ENV_FILE"

exec "$@"

