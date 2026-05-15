# Google Cloud Run No-Build Deployment Guide

Deploying your application to Google Cloud Run usually involves building a container image (via a local Docker daemon or Cloud Build) and then deploying that image. However, for supported languages, you can deploy your source code **directly** to Cloud Run without a container build step by using a pre-configured Google-managed runtime base image.

---

## MCP Tool Interface

To perform a no-build deployment, call the MCP tool `deploy_cloudrun_service_no_build`. This tool executes the deployment under the hood and returns the resulting Cloud Run service object.

### Interactive Parameter Constraints:
1.  **Mandatory Fields**: Before invoking the tool, you **MUST** ask the user to provide any required arguments (such as `project_id`, `location`, `service_name`) that are not explicitly specified in the conversation before or naturally discoverable in the repository context. Do not assume or guess missing mandatory values.
2.  **Optional Fields**: For optional fields (such as `port`, `allow_public_access`, `command`, `args`, or `env_vars`), prepare reasonable defaults based on the project's language guide below, present these defaults clearly to the user, and ask for explicit confirmation before calling the tool.

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

### Node.js No-Build Deployment Guide

#### 1. Prerequisites & Checklist
*   **Active Files**: The root of your application MUST contain `package.json` and `index.js`.
*   **Node.js Base Image**: The default base image is `nodejs24` (or similar version matching your project requirements).

#### 2. Step-by-Step Workflow
1.  **Install Dependencies Locally**:
    Before deploying, run `npm install` in your local workspace root. This ensures all dependencies listed in `package.json` are installed inside your `./node_modules` directory:
    ```bash
    npm install
    ```
2.  **Deploy via MCP Tool**:
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

### Python No-Build Deployment Guide

#### 1. Prerequisites & Checklist
*   **Identify Entry Point**: Locate the startup file for your application (typically `main.py`).
*   **Requirements Specification**: Ensure there is a `requirements.txt` file in the root directory. If missing, scan the project imports to generate a correct `requirements.txt` (e.g., `pipreqs` or manual entry) and verify that running the app locally with these requirements works successfully.
*   **Python Base Image**: The default base image is `python314` (or matching your Python major/minor version).

#### 2. Step-by-Step Workflow
1.  **Install Dependencies Locally**:
    Since there is no build step on Cloud Run, all external dependencies must be packaged alongside your application source code. To avoid system-level `pip` permission restrictions, always create an isolated virtual environment (e.g. `.venv`), activate it to use its private `pip` to stage dependencies into a dedicated vendor folder (e.g. `./vendor`), and then deactivate:
    ```bash
    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt --target=./vendor
    deactivate
    ```
2.  **Deploy via MCP Tool**:
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

### Go No-Build Deployment Guide

#### 1. Prerequisites & Checklist
*   **Go compiler**: Must be installed locally on your machine.
*   **Linux-targeted Binary**: Since Cloud Run runs containerized Linux environments, you must compile your Go binary for the Linux OS and AMD64 architecture (`linux/amd64`).
*   **OS-only Base Image**: The deploy command uses `osonly24` (a base image containing only the operating system without pre-installed language runtime tools) to execute your pre-compiled binary.

#### 2. Step-by-Step Workflow
1.  **Compile the Binary Locally**:
    Build the Go binary by targeting Linux OS (`linux/amd64`). Replace `main.go` with your project's actual entry point filename if different:
    ```bash
    GOOS="linux" GOARCH=amd64 go build -o main main.go
    ```
2.  **Deploy via MCP Tool**:
    Call the `deploy_cloudrun_service_no_build` MCP tool with the following parameters:
    *   **`project_id`**: The Google Cloud project ID.
    *   **`location`**: The region where your service is deployed.
    *   **`service_name`**: Replace with the desired service name.
    *   **`source`**: `.`
    *   **`base_image`**: `osonly24`
    *   **`command`**: `./main`
    *   **`allow_public_access`**: `true` (if you want it to be a public service)

