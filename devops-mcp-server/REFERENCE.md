# Tool Reference

This document provides detailed information about the tools available in the Google Cloud DevOps MCP Server.

## Artifact Registry

### `artifactregistry.setup_repository`
Sets up a new Artifact Registry repository. Optionally, it can grant Artifact Registry Writer permissions to a service account.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location for the repository.
- `repository_id` (string, required): The ID of the repository.
- `format` (string, required): The format of the repository (e.g., `DOCKER`).
- `service_account_email` (string, optional): The email of the service account to grant Artifact Registry Writer permissions to.

## Cloud Build

### `cloudbuild.list_triggers`
Lists all Cloud Build triggers in a given location.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location for the triggers.

### `cloudbuild.create_trigger`
Creates a new Cloud Build trigger.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location for the trigger.
- `trigger_id` (string, required): The ID of the trigger.
- `repo_link` (string, required): The Developer Connect repository link.
- `service_account` (string, required): The service account to use for the build (e.g., `serviceAccount:123-compute@developer.gserviceaccount.com`).
- `branch` (string, optional): Create builds on push to branch (regex, e.g., `^main$`).
- `tag` (string, optional): Create builds on new tag push (regex, e.g., `^nightly$`).

### `cloudbuild.run_trigger`
Runs a Cloud Build trigger.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location for the trigger.
- `trigger_id` (string, required): The ID of the trigger.
- `branch` (string, optional): The branch to run the trigger at (regex).
- `tag` (string, optional): The tag to run the trigger at (regex).
- `commit_sha` (string, optional): The exact commit SHA to run the trigger at.

## Cloud Run

### `cloudrun.list_services`
Lists the Cloud Run services in a specified GCP project and location.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location.

### `cloudrun.deploy_to_cloud_run_from_image`
Creates a new Cloud Run service or updates an existing one from a container image.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location.
- `service_name` (string, required): The name of the Cloud Run service.
- `revision_name` (string, required): The name of the Cloud Run revision.
- `image_url` (string, required): The URL of the container image to deploy.
- `port` (integer, optional): The port the container listens on.
- `allow_public_access` (boolean, optional): If the service should be public. Default is `false`.

### `cloudrun.deploy_to_cloud_run_from_source`
Creates a new Cloud Run service or updates an existing one from source.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location.
- `service_name` (string, required): The name of the Cloud Run service.
- `source` (string, required): The path to the source code to deploy.
- `port` (integer, optional): The port the container listens on.
- `allow_public_access` (boolean, optional): If the service should be public. Default is `false`.

## Cloud Storage

### `cloudstorage.list_buckets`
Lists Cloud Storage buckets in a specified project.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.

### `cloudstorage.upload_source`
Uploads source to a GCS bucket. If a new bucket is created, it will be public.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `bucket_name` (string, optional): The name of the bucket.
- `destination_dir` (string, required): The name of the destination directory in the bucket.
- `source_path` (string, required): The local path to the source directory.

## Developer Connect

### `devconnect.setup_connection`
Sets up a Developer Connect connection.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location.
- `git_repo_uri` (string, required): The URI of the git repository to connect to.

### `devconnect.add_git_repo_link`
Creates a Developer Connect git repository link when a connection already exists.

**Arguments:**
- `project_id` (string, required): The Google Cloud project ID.
- `location` (string, required): The Google Cloud location.
- `connection_id` (string, required): The ID of the Developer Connect connection.
- `git_repo_uri` (string, required): The URI of the git repository to link.

## OSV

### `osv.scan_secrets`
Scans the specified root directory for secrets using OSV.

**Arguments:**
- `root` (string, required): The absolute path to the root directory to scan.
- `ignore_directories` (array of strings, optional): Absolute directory paths to ignore.

## BM25 (Search)

### `bm25.search_common_cicd_patterns`
Finds common CI/CD patterns in the database.

**Arguments:**
- `query` (string, required): The query to search for.

### `bm25.query_knowledge`
Finds knowledge snippets in the knowledge database.

**Arguments:**
- `query` (string, required): The query to search for.
