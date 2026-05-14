# Node.js Cloud Run No-Build Deployment Guide

This guide provides step-by-step instructions on how to deploy a Node.js application directly to Google Cloud Run without a build step.

---

## Prerequisites & Checklist

*   **Active Files**: The root of your application MUST contain `package.json` and `index.js`.
*   **Node.js Base Image**: The default base image is `nodejs24` (or similar version matching your project requirements).

---

## Step-by-Step Workflow

### 1. Install Dependencies Locally
Before deploying, run `npm install` in your local workspace root. This ensures all dependencies listed in `package.json` are installed inside your `./node_modules` directory:

```bash
npm install
```

### 2. Run Secret Scanning
Always verify no secrets are present in your files by running the `scan_code_for_secrets` tool. If secrets are found, ensure they are added to `.gitignore` or `.gcloudignore`.

### 3. Deploy via MCP Tool
Call the `deploy_cloudrun_service_no_build` MCP tool with the following parameters:

*   **`project_id`**: The Google Cloud project ID.
*   **`location`**: The region where your service is deployed.
*   **`service_name`**: Replace with the desired service name.
*   **`source`**: `.`
*   **`base_image`**: `nodejs24`
*   **`command`**: `node`
*   **`args`**: `["index.js"]`
*   **`allow_public_access`**: `true` (if you want it to be a public service)

---

## Verification
Upon successful deployment, the `gcloud` command will output the service URL. Open the URL in your browser or use `curl` to verify your service is running.
