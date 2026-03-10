# End-to-End Guide: Creating a Cloud Build Trigger

This document outlines the standard, idempotent procedure for creating a Google Cloud Build trigger. Creating the trigger is the **final** action in a sequence of prerequisite checks and resource provisioning steps. The agent must ensure all dependencies are met before attempting to create the trigger itself.

---

## ## Core Principle: Idempotency

Every step in this process must be **idempotent**. This means the agent must **always check if a resource already exists** before attempting to create it. This prevents errors and ensures the process can be run multiple times safely.

---

## ## Prerequisite Checklist

The following dependencies must be satisfied in order before creating the trigger.

### ### 1. Ensure `cloudbuild.yaml` Exists

The trigger needs a build configuration file to execute.

* **Action**: Check for a `cloudbuild.yaml` file at the root of the source repository.
* **If it does not exist**: Generate one by translating the user-approved plan. The steps in the generated YAML must be a direct translation of the components defined in the plan's `stages` object. The specifics of the steps (e.g., using `pytest` vs. `mvn test`) should be informed by discovering the application archetype (e.g., by finding a `pyproject.toml` or `pom.xml`).

### ### 2. Ensure Artifact Registry Repository Exists

The `cloudbuild.yaml` file will reference an Artifact Registry repository to push container images. This repository must exist before a build can succeed.

* **Action**: Parse the `cloudbuild.yaml` file to identify the Artifact Registry image path (e.g., `us-central1-docker.pkg.dev/my-project/my-app-repo/my-image`).
* **Extract** the repository portion (`us-central1-docker.pkg.dev/my-project/my-app-repo`).
* **Check** if this repository already exists in the target GCP project.
* **If it does not exist**: Create it using the available tools.

### ### 3. Ensure Developer Connect and Repository Link Exist

Cloud Build triggers connect to source code via Developer Connect. The entire connection and repository link must be in place.

1.  **Check for Connection**: First, check if a Developer Connect connection already exists for the relevant Git provider (e.g., GitHub) in the target project and region.
2.  **Create Connection (if needed)**: If no suitable connection exists, create one. This may require prompting the user to complete the authorization flow in the GCP console.
3.  **Obtain Source URI**: The agent must know the exact URI of the source code repository (e.g., `https://github.com/my-org/my-app`). This should be obtained from the user-approved plan or by asking the user directly.
4.  **Check for Repository Link**: Check if a repository link for that specific URI already exists within the Developer Connect connection.
5.  **Create Repository Link (if needed)**: If the link does not exist, create it. This link is the resource that the Cloud Build trigger will formally point to.

---

## ## Final Step: Creating the Trigger

Once all prerequisites are met, the agent can create the trigger itself using the available tools.