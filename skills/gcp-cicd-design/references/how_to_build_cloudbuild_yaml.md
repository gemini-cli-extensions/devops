# How to Create a `cloudbuild.yaml` File

This document outlines the standard procedure for automatically generating a `cloudbuild.yaml` configuration file. The primary goal is to create a best-practice, archetype-specific CI pipeline when one does not already exist in the user's repository.

---

## When to Generate a `cloudbuild.yaml`

The core principle is to be non-destructive and idempotent. The generation process should only be triggered under one specific condition:

* **A `cloudbuild.yaml` file does not exist at the root of the source repository.**

If a `cloudbuild.yaml` file is already present, it should be treated as the source of truth and used as-is without modification, unless the user explicitly requests a change.

---

## Step 1: Discovering the Application Archetype

Before a `cloudbuild.yaml` can be generated, the application's archetype must be identified. This is done by inspecting the local filesystem for common project files. This discovery step is crucial for tailoring the build steps (e.g., linting, testing) to the specific language or framework.

The following mapping should be used:

* `pom.xml` → **Java (Maven)**
* `build.gradle` → **Java (Gradle)**
* `package.json` → **Node.js**
* `requirements.txt` or `pyproject.toml` → **Python**
* `go.mod` → **Go**

---

## Step 2: Generating the Default CI Pipeline

If no `cloudbuild.yaml` exists, a new one should be generated with a standard, four-step CI sequence. These steps must be tailored to the discovered application archetype.

1.  **Lint**: Run a static code analysis tool to check for stylistic or programmatic errors. The specific linter should match the application archetype (e.g., `pylint` for Python, `eslint` for Node.js).
2.  **Test**: Execute the project's unit tests. The test runner should match the archetype (e.g., `pytest` for Python, `go test` for Go, `mvn test` for Maven).
3.  **Build Container**: Use the native `gcr.io/cloud-builders/docker` builder to build the container image from the `Dockerfile` in the repository.
4.  **Push Container**: Push the newly built container image to the verified Artifact Registry repository.

---

## Key Best Practices and Configuration

When generating the `cloudbuild.yaml`, several best practices must be included to ensure security and efficiency.

* **Image Tagging**: The container image must be tagged with the `$SHORT_SHA` substitution variable. This ensures a unique, traceable image for every single commit.
* **Use the `images` Attribute**: The final image URI should be explicitly listed under the top-level `images` attribute in the `cloudbuild.yaml`. This allows Cloud Build to push the image concurrently with other build steps, potentially speeding up the build.
* **Enable Provenance**: To enhance supply chain security, build provenance should always be enabled. This is done by adding an `options` block and setting `requestedVerifyOption: VERIFIED`.

---

## Example: Python Application

Here is a complete, best-practice `cloudbuild.yaml` generated for a Python application that uses `pytest` for testing.

```yaml
# Auto-generated cloudbuild.yaml for a Python application

steps:
  # Step 1: Install dependencies
  - name: 'python:3.11'
    entrypoint: 'pip'
    args: ['install', '-r', 'requirements.txt', '--user']

  # Step 2: Run unit tests with pytest
  - name: 'python:3.11'
    entrypoint: 'python'
    args: ['-m', 'pytest']

  # Step 3: Build the container image
  # The image is tagged with the short commit SHA for traceability.
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'build'
      - '-t'
      - '${_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${_REPO_NAME}/${_IMAGE_NAME}:$SHORT_SHA'
      - '.'

  # Step 4: Push the container image to Artifact Registry
  # This step runs in parallel with the final steps of the build.
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'push'
      - '${_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${_REPO_NAME}/${_IMAGE_NAME}:$SHORT_SHA'

# Explicitly list the final image to be pushed for potential build speed improvements.
images:
  - '${_LOCATION}-docker.pkg.dev/${PROJECT_ID}/${_REPO_NAME}/${_IMAGE_NAME}:$SHORT_SHA'

# Enable SLSA Level 3 provenance for enhanced supply chain security.
options:
  requestedVerifyOption: VERIFIED