# Go Cloud Run No-Build Deployment Guide

This guide provides step-by-step instructions on how to deploy a Go application directly to Google Cloud Run without a build step by pre-compiling the binary locally.

---

## Prerequisites & Checklist

*   **Go compiler**: Must be installed locally on your machine.
*   **Linux-targeted Binary**: Since Cloud Run runs containerized Linux environments, you must compile your Go binary for the Linux OS and AMD64 architecture (`linux/amd64`).
*   **OS-only Base Image**: The deploy command uses `osonly24` (a base image containing only the operating system without pre-installed language runtime tools) to execute your pre-compiled binary.

---

## Step-by-Step Workflow

### 1. Compile the Binary Locally
Build the Go binary by targeting Linux OS (`linux/amd64`). Replace `main.go` with your project's actual entry point filename if different:

```bash
GOOS="linux" GOARCH=amd64 go build -o main main.go
```

### 2. Run Secret Scanning
Check your workspace for secrets before deployment with the `scan_code_for_secrets` tool to ensure sensitive config files are not inadvertently included.

### 3. Deploy via MCP Tool
Call the `deploy_cloudrun_service_no_build` MCP tool with the following parameters:

*   **`project_id`**: The Google Cloud project ID.
*   **`location`**: The region where your service is deployed.
*   **`service_name`**: Replace with the desired service name.
*   **`source`**: `.`
*   **`base_image`**: `osonly24`
*   **`command`**: `./main`
*   **`allow_public_access`**: `true` (if you want it to be a public service)

---

## Verification
Verify that the deployment was successful by using the service URL output in the `gcloud` command response. Send HTTP requests or view in the browser to check if it responds correctly.
