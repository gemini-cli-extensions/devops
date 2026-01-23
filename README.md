# Gemini CLI DevOps Extension

The DevOps extension supercharges your Gemini CLI experience, providing AI-assisted Continuous Integration and Continuous Delivery ([CI/CD](https://en.wikipedia.org/wiki/CI/CD)) workflows. Effortlessly deploy to Google Cloud services like Cloud Run and Cloud Storage, and generate robust CI/CD pipelines that adhere to testing and security best practices.

> [!IMPORTANT]
> This project is currently experimental. Features, commands, and functionality may change. We highly value and welcome your feedback!

## 📋 Key Features

-   **Intelligent Code Deployment**: Use the `/devops:deploy` command to deploy your codebase. The extension leverages Gemini to analyze your project and recommend the best Google Cloud service: Cloud Run for dynamic applications or Cloud Storage for static websites. Includes pre-deployment scanning for secrets, keys, and passwords to prevent accidental leaks.
-   **AI-Powered CI/CD Pipeline Design**: Generate secure and robust CI/CD pipelines in moments with `/devops:design`. Collaborate with Gemini to tailor the pipeline to your specific needs, including automatic setup of the required Google Cloud infrastructure.
-   **Interactive GCP Management**: The extension provides commands and tools to interact directly with Google Cloud's CI/CD services (Cloud Build, Artifact Registry, Artifact Analysis, Cloud Deploy, Developer Connect) from within Gemini CLI. Run builds, check for vulnerabilities (CVEs), view SBOMs, and pull build logs to investigate failures.
-   **Simplified Complex Release Flows**: Build sophisticated Cloud Deploy release pipelines quickly, guided by simple, interactive questions.
-   **Integrated DevOps MCP Server**: The extension includes a local Model Context Protocol (MCP) server, seamlessly integrating Gemini CLI with Google Cloud CI/CD services.

>[!NOTE]
> As with all Generative AI, Large Language Models (LLMs) can sometimes produce unexpected outputs ("hallucinate"). Please use this extension with care and always verify the generated configurations and commands.

## ⚙️ Installation

To install the DevOps extension, run the following command in your terminal:

```bash
gemini extensions install https://github.com/gemini-cli-extensions/devops
```
*TBD: Add `--ref` or other flags if needed for specific versions or pre-releases, e.g., --ref=nightly --pre-release*

## ✅ Prerequisites

*   [Gemini CLI](https://github.com/google-gemini/gemini-cli): Version **v0.15.0 or newer** must be installed.
*   Gemini CLI Authentication: Ensure you have configured [Authentication Options](https://github.com/google-gemini/gemini-cli/tree/main?tab=readme-ov-file#-authentication-options).
*   `gcloud` CLI: The Google Cloud CLI must be [installed](https://cloud.google.com/sdk/docs/install) and available in your system's PATH.
*   Google Cloud Project: You need a Google Cloud project with the necessary APIs enabled. Depending on your usage, the extension may require:
    *   Cloud Build API
    *   Artifact Registry API
    *   Artifact Analysis API
    *   Developer Connect API
    *   Cloud Run API
    *   Cloud Storage API
*   Application Default Credentials (ADC): Ensure [Application Default Credentials](https://cloud.google.com/docs/authentication/gcloud) are configured in your environment. You can set this up by running:
    ```bash
    gcloud auth login
    gcloud auth application-default login
    ```

> [!WARNING]
> **Security Recommendation:** This extension, through Gemini CLI and your ADC, can modify your Google Cloud resources. We recommend adhering to the principle of least privilege. When running in a production or sensitive environment, consider configuring ADC to use a service account with only the necessary permissions. You can set up ADC to impersonate a service account using a command similar to this:
> `gcloud auth application-default login --impersonate-service-account YOUR-SERVICE-ACCOUNT-EMAIL`
> Learn more about [setting up ADC for local development](https://cloud.google.com/docs/authentication/set-up-adc-local-dev-environment) and [service account impersonation](https://cloud.google.com/docs/authentication/use-service-account-impersonation).
> Always carefully review any generated pipelines or commands before execution.

## ☕ Usage

>[!TODO] *Coming soon! Examples and common workflows will be added here.*
#### `/devops:deploy`
A simple deployment workflow that guides you through deploying to the most suitable Google Cloud service. Recommends Cloud Storage for static content and Cloud Run for dynamic applications. Performs secret scanning before deployment. Analyzes your current workspace and guides you through deploying to the most suitable Google Cloud service. Recommends Cloud Storage for static content and Cloud Run for dynamic applications. Performs secret scanning before deployment.

#### `/devops:design`
Launches an AI-assisted process to design and generate a CI/CD pipeline configuration (cloudbuild.yaml) tailored to your project, including the necessary Google Cloud infrastructure. Gemini will ask clarifying questions about your needs and codebase to create a pipeline for building, testing, and pushing artifacts, primarily using Google Cloud Build and Artifact Registry.

**Design Process:**

1. Requirement Gathering: Gemini inspects your current workspace and asks clarifying questions to understand your application type, build process, testing strategies, and deployment objectives.
2. Pipeline Configuration Generation: Gemini generates a cloudbuild.yaml file defining the pipeline stages (e.g., source checkout, dependency installation, build, test, artifact push to Artifact Registry). The configuration is created using Cloud Build's script mode for clarity.
3. Infrastructure Setup: Based on the requirements, Gemini guides you through setting up the required GCP resources. This may include:
    * Creating or configuring Artifact Registry repositories.
    * Establishing connections to your Git repository (e.g., GitHub) using Developer Connect.
    * Setting up or advising on necessary IAM service accounts and permissions for the Cloud Build service to interact with other GCP services.
4. Validation & Testing: The extension will attempt to validate the generated pipeline configuration, potentially by submitting an initial test build using gcloud builds submit to ensure correctness.
5. Review & Refinement: You can review the generated configuration and infrastructure setup and provide feedback to Gemini for further adjustments.

The goal is to produce a functional, production-ready CI/CD pipeline configuration with all the necessary GCP prerequisites in place.


### 🛠️ Supported MCP Tools

The extension exposes the following tools to Gemini CLI, enabling interaction with Google Cloud services:

#### CI/CD Service Tools
*   `artifactregistry.setup_repository`: Creates a new Artifact Registry repository. Optionally grants Artifact Registry Writer permissions to a specified service account.
*   `cloudbuild.create_trigger`: Creates a new Cloud Build trigger.
*   `cloudbuild.list_triggers`: Lists all Cloud Build triggers in a given project and location.
*   `cloudbuild.run_trigger`: Manually runs an existing Cloud Build trigger.
*   `devconnect.add_git_repo_link`: Creates a Developer Connect Git repository link under an existing connection.
*   `devconnect.setup_connection`: Sets up a new Developer Connect connection (e.g., to GitHub).

#### Deployment Tools
*   `cloudrun.deploy_to_cloud_run_from_image`: Deploys a container image to Cloud Run, creating a new service or updating an existing one.
*   `cloudrun.deploy_to_cloud_run_from_source`: Deploys to Cloud Run directly from source code, typically using Cloud Build and BuildPacks.
*   `cloudrun.list_services`: Lists Cloud Run services in a specified project and location.
*   `cloudstorage.list_buckets`: Lists Cloud Storage buckets in a specified project.
*   `cloudstorage.upload_source`: Uploads files from the local workspace to a GCS bucket. Can create a new public bucket if specified.
*   `osv.scan_secrets`: Scans a specified directory for potential secrets and keys using OSV-Scanner.

#### Knowledge Retrieval Tools
*   `bm25.query_knowledge`: Retrieves relevant snippets from the extension's knowledge base to answer questions.
*   `bm25.search_common_cicd_patterns`: Finds common CI/CD pipeline patterns and best practices.

## 📚 Resources

-   [Gemini CLI Extensions Documentation](https://geminicli.com/extensions/about/): Learn more about how extensions work in Gemini CLI.
-   [GitHub Issues](https://github.com/gemini-cli-extensions/devops/issues): Report bugs, request features, or provide feedback.
- Setup Service Account for ADC Impersonation: [Service Account Impersonation](https://docs.cloud.google.com/docs/authentication/set-up-adc-attached-service-account)

