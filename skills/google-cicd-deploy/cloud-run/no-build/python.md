# Python Cloud Run No-Build Deployment Guide

This guide provides step-by-step instructions on how to deploy a Python application directly to Google Cloud Run without a build step.

---

## Prerequisites & Checklist

*   **Identify Entry Point**: Locate the startup file for your application (typically `main.py`).
*   **Requirements Specification**: Ensure there is a `requirements.txt` file in the root directory. 
    *   *If missing*, scan the project imports to generate a correct `requirements.txt` (e.g., `pipreqs` or manual entry) and verify that running the app locally with these requirements works successfully.
*   **Python Base Image**: The default base image is `python314` (or matching your Python major/minor version).

---

## Step-by-Step Workflow

### 1. Install Dependencies Locally into Vendor Folder
Since there is no build step on Cloud Run, all external dependencies must be packaged alongside your application source code.

Install dependencies into a local `./vendor` directory by running:

```bash
pip3 install -r requirements.txt --target=./vendor
```

### 2. Run Secret Scanning
Ensure there are no secrets present in the vendor or code directories using the `scan_code_for_secrets` tool. Configure `.gcloudignore` or `.gitignore` to prevent uploading sensitive local files.

### 3. Deploy via MCP Tool
Call the `deploy_cloudrun_service_no_build` MCP tool with the following parameters:

*   **`project_id`**: The Google Cloud project ID.
*   **`location`**: The region where your service is deployed.
*   **`service_name`**: Replace with the desired service name.
*   **`source`**: `.`
*   **`base_image`**: `python314`
*   **`command`**: `python`
*   **`args`**: `["main.py"]`
*   **`env_vars`**: `{"PYTHONPATH": "./vendor"}`
*   **`allow_public_access`**: `true` (if you want it to be a public service)

---

## Verification
Upon successful deployment, check the output log for the service URL. Verify the deployment is correct by sending a GET request or visiting the URL in a web browser.
