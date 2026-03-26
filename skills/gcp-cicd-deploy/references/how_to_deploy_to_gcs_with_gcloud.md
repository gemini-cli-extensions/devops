# How to Deploy to Google Cloud Storage with `gcloud`

This document outlines the standard procedure for deploying static applications (e.g., websites and single-page applications) to a Google Cloud Storage (GCS) bucket using the `gcloud storage` command-line tool.

---

## Step 1: Gather Deployment Parameters

Before executing any commands, the following parameters must be identified:

*   **`PROJECT_ID`**: The Google Cloud project where the bucket will be hosted.
*   **`BUCKET_NAME`**: The name of the GCS bucket (e.g., `gs://my-static-site-bucket`).
*   **`LOCATION`**: The geographic location for the bucket (e.g., `us-central1`).
*   **`SOURCE_PATH`**: The local directory containing the build artifacts (e.g., `./dist`, `./build`).
*   **`DESTINATION_DIR`**: (Optional) The directory within the bucket to upload to.
*   **`MAIN_FILE`**: The entry point of the application (e.g., `index.html`). This is **identified automatically** by inspecting the build artifacts.

---

## Step 2: Prepare the Storage Bucket

### 1. Create the Bucket (if it doesn't exist)
```bash
gcloud storage buckets create gs://<BUCKET_NAME> \
    --project=<PROJECT_ID> \
    --location=<LOCATION>
```

### 2. Enable Public Access
```bash
gcloud storage buckets add-iam-policy-binding gs://<BUCKET_NAME> \
    --member=allUsers \
    --role=roles/storage.objectViewer \
    --project=<PROJECT_ID>
```

---

## Step 3: Deploy Build Artifacts

Use the `gcloud storage cp` command to upload the local build artifacts to the bucket.

### Best Practice: The Recursive Wildcard (`/**`)
Always use the `/**` wildcard to ensure that all files, including **hidden files (dotfiles)**, are correctly matched and uploaded. This is critical for configuration files (e.g., `.well-known/`) and some modern build tool outputs.

```bash
gcloud storage cp -r <SOURCE_PATH>/** gs://<BUCKET_NAME>/<DESTINATION_DIR> \
    --project=<PROJECT_ID>
```

---

## Step 4: Verification

After the upload is complete, the application can be accessed via the standard GCS public URL:

`https://storage.googleapis.com/<BUCKET_NAME>/<MAIN_FILE>`
