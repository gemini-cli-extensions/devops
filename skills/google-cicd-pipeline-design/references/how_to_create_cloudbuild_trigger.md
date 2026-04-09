# End-to-End Guide: Creating a Cloud Build Trigger

This document outlines the standard, idempotent procedure for creating a Google Cloud Build trigger. Creating the trigger is the **final** action in a sequence of prerequisite checks and resource provisioning steps. The agent must ensure all dependencies are met before attempting to create the trigger itself.

---

## Core Principle: Idempotency

Every step in this process must be **idempotent**. This means the agent must **always check if a resource already exists** before attempting to create it. This prevents errors and ensures the process can be run multiple times safely.

---

## Prerequisite Checklist

The following dependencies must be satisfied in order before creating the trigger.

### 1. Ensure `cloudbuild.yaml` Exists

The trigger needs a build configuration file to execute.

* **Action**: Check for a `cloudbuild.yaml` file at the root of the source repository.
* **If it does not exist**: Generate one by translating the user-approved plan. The steps in the generated YAML must be a direct translation of the components defined in the plan's `stages` object. The specifics of the steps (e.g., using `pytest` vs. `mvn test`) should be informed by discovering the application archetype (e.g., by finding a `pyproject.toml` or `pom.xml`).

### 2. Ensure Artifact Registry Repository Exists

The `cloudbuild.yaml` file will reference an Artifact Registry repository to push container images. This repository must exist before a build can succeed.

* **Action**: Parse the `cloudbuild.yaml` file to identify the Artifact Registry image path (e.g., `us-central1-docker.pkg.dev/my-project/my-app-repo/my-image`).
* **Extract** the repository portion (`us-central1-docker.pkg.dev/my-project/my-app-repo`).
* **Check** if this repository already exists in the target GCP project.
* **If it does not exist**: Create it using the available tools.

### 3. Ensure Developer Connect and Repository Link Exist

Cloud Build triggers connect to source code via Developer Connect. The entire connection and repository link must be in place.

1.  **Check for Connection**: First, check if a Developer Connect connection already exists for the relevant Git provider (e.g., GitHub) in the target project and region.
2.  **Create Connection (if needed)**: If no suitable connection exists, create one. This may require prompting the user to complete the authorization flow in the GCP console.
3.  **Obtain Source URI**: The agent must know the exact URI of the source code repository (e.g., `https://github.com/my-org/my-app`). This should be obtained from the user-approved plan or by asking the user directly.
4.  **Check for Repository Link**: Check if a repository link for that specific URI already exists within the Developer Connect connection.
5.  **Create Repository Link (if needed)**: If the link does not exist, create it. This link is the resource that the Cloud Build trigger will formally point to.

### 4. Ensure Dedicated Service Account Exists

Cloud Build triggers should **always** use a dedicated, user-managed service account instead of the default Compute Engine service account to follow the principle of least privilege. **CRITICAL: NEVER use the default Compute Engine service account (typically formatted as `[PROJECT_NUMBER]-compute@developer.gserviceaccount.com`), even if you observe existing triggers in the project using it.**

1.  **Check for Service Account**: Check if a dedicated service account for Cloud Build (e.g., `cloud-build-runner@<PROJECT_ID>.iam.gserviceaccount.com`) already exists.
2.  **Create Service Account (if needed)**: If no dedicated service account exists, create one using `gcloud iam service-accounts create`.
3.  **Required Roles (Ensured Automatically)**: The `create_build_trigger` MCP tool will **automatically ensure** the necessary roles (e.g., `roles/logging.logWriter`, `roles/artifactregistry.writer`, `roles/developerconnect.tokenAccessor`, `roles/run.developer`, `roles/storage.admin`, `roles/serviceusage.serviceUsageConsumer`, `roles/cloudbuild.builds.editor`, `roles/cloudbuild.workerpools.use`, and specific IAM delegation on the default Compute SA) are granted to the relevant service accounts. You **do not** need to grant these permissions manually.

---

## Final Step: Creating the Trigger

Once all prerequisites are met, the agent can create the trigger itself using the `create_build_trigger` tool. **You MUST provide the email of the dedicated service account identified or created in Step 4 to the `service_account` parameter.**
