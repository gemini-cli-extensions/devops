#!/bin/bash

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
