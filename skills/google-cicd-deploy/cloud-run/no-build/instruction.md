# Google Cloud Run No-Build Deployment Guide

Deploying your application to Google Cloud Run usually involves building a container image (via a local Docker daemon or Cloud Build) and then deploying that image. However, for supported languages, you can deploy your source code **directly** to Cloud Run without a container build step by using a pre-configured Google-managed runtime base image.

---

## Why Choose No-Build Deployment?

*   **Sub-second Build Times**: Bypasses building container layers entirely, substantially speeding up deployment cycles.
*   **Simpler Developer Experience**: Eliminates the need for a local Docker daemon, dockerfiles, or Cloud Build triggers.
*   **Predictable Environments**: Runs on Google-managed, secured, and optimized runtime base images (such as `nodejs24`, `python314`, and `osonly24`).
*   **Vendor-Locking/Pre-Compilation Support**: Great for compiled binaries (Go) and standard scripting runtimes with bundled dependencies.

---

## Prerequisites

Before you proceed, ensure the following prerequisites are met:
1.  **Supported Language**: This feature currently supports **Node.js**, **Python**, and **Go**.
2.  **Google Cloud SDK**: The `gcloud` CLI must be installed and updated (`gcloud components update`).
3.  **Beta Commands Enabled**: This workflow relies on `gcloud beta`.

---

## MCP Tool Interface

To perform a no-build deployment, call the MCP tool `deploy_cloudrun_service_no_build`. This tool executes the deployment under the hood and returns the resulting Cloud Run service object.

### Tool Arguments:
*   **`project_id`**: The Google Cloud project ID.
*   **`location`**: The location/region where your service is deployed (e.g., `us-central1`).
*   **`service_name`**: The name of your Cloud Run service.
*   **`source`**: The location of your application on the local file system (usually `.`).
*   **`base_image`**: The runtime base image you want to use (e.g., `nodejs24`, `python314`, or `osonly24`).
*   **`command`** (optional): The entry point command that the container starts up with.
*   **`args`** (optional array of strings): The argument(s) to pass to the startup command.
*   **`env_vars`** (optional map): Environment variables (e.g., `{"PYTHONPATH": "./vendor"}`).
*   **`port`** (optional integer): The container port to listen on.
*   **`allow_public_access`** (optional boolean): If the service should be public. Default is `false`.

---

## Language-Specific Detailed Guides

Select the guide corresponding to your project's language:

*   [Node.js No-Build Deployment](nodejs.md)
*   [Python No-Build Deployment](python.md)
*   [Go No-Build Deployment](go.md)
