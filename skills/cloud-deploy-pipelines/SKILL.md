---
name: cloud-deploy-pipelines
description: >
  Design Cloud Deploy delivery pipelines and manage releases when deploying applications to Cloud Run and Google Kubernetes Engine (GKE). Use when users want to deploy their applications to multiple environments (e.g. dev and prod), leverage deployment strategies (i.e. canary), or  rollback (manually or automatically) when there are issues deploying their application.
version: "0.1.0"
---

# Cloud Deploy Pipelines

## Overview

This skill encompasses the entire lifecycle of Cloud Deploy for a user, from designing and creating delivery pipelines to managing releases and debugging release failures.

## Workflow: Designing a Pipeline

This workflow provides steps for designing a Cloud Deploy `DeliveryPipeline`.

### Constraints & Rules

1.  **NO PLACEHOLDERS**: Never generate YAML with placeholders like `<PROJECT_ID>`. Ask the user for values first.
2.  **Context First**: Always check existing files and conversation context before asking.
3.  **Step-by-Step**: Perform the steps one at a time. The goal is to guide the user through designing a delivery pipeline.

### Step 0: Prerequisites

**Required Context**: Before generating ANY configuration for this workflow, you **MUST** have the following values. Ask the user strictly for any missing information:
- **Project ID**: The Google Cloud project ID for the Cloud Deploy resources.
- **Region**: The region for the Cloud Deploy resources (e.g., `us-central1`).
- **Application Name**: The name of the application that will be deployed. Use the application name to generate the names of the Cloud Deploy resources, such as `DeliveryPipeline` and `Target` resources.
- **Runtime**: Either Cloud Run or Google Kubernetes Engine (GKE).
  - **If Cloud Run**: The Cloud Run project and location.
  - **If GKE**: The GKE cluster name.

### Step 1: Define the target environments

1. Identify the number of deployment environments (e.g., dev, staging, production).
2. Identify if promotions should require user approval.
3. Define each of the deployment environments as Cloud Deploy `Target` resources in a `clouddeploy.yaml` file. 
    - Use `references/configure-targets.md` as a reference when generating the resource YAML.
    - Use the application name provided by the user when naming the Cloud Deploy `Target` resources. For example, if the user wants to deploy an application named "hello-world" to a test and production environment then use "hello-world-test" and "hello-world-prod" as the `Target` IDs.

### Step 2: Define the delivery pipeline

1. Identify whether the user wants to use a canary deployment strategy for any of the target environments.
2. Define the Cloud Deploy `DeliveryPipeline` in the `clouddeploy.yaml` file.
    - Use `references/configure-pipelines.md` as a reference when generating the resource YAML.
    - Use application name as the `DeliveryPipeline` ID.

### Step 3: Define automations

1. Identify whether the user wants to automatically rollback if any failures occur during the rollout.
2. **If the user specified multiple deployment environments in the previous step** 
  - Identify if they want automatic promotions between deployment environments.
3. **If the user specified a canary deployment strategy in the previous step** 
  - Identify if they want to automatically advance the rollout through the phases after a wait period.
4. Define the Cloud Deploy `Automation` resources in the `clouddeploy.yaml` file.
  - Use `references/configure-automations.md` as a reference when generating the resource YAML.

### Step 4: Validate the clouddeploy.yaml file

Ensure that the `clouddeploy.yaml` file is valid. See https://docs.cloud.google.com/deploy/docs/config-files for the schema.

### Step 5: Create the delivery pipeline

Run the following command to create the Cloud Deploy `DeliveryPipeline` and associated resources, using the values collected in Step 0:

```bash
gcloud deploy apply --file=clouddeploy.yaml --region=<REGION> --project=<PROJECT_ID>
```

### Step 6: Create a skaffold.yaml file and runtime manifests

**Required Context**: Before generating a `skaffold.yaml` file, you **MUST** know if the user has manifests for the runtime they are deploying to.

1. **If the user does not have runtime manifests**: Generate some basic ones based on the runtime.
    - **If Cloud Run**: Generate a Cloud Run manifest. Use `references/basic-cloudrun-manifests.md` as a reference.
    - **If GKE**: Generate a Kubernetes `Deployment` and `Service`manifest. Use `references/basic-k8s-manifests.md` as a reference.
2. Create a `skaffold.yaml` file required to create a Cloud Deploy `Release` for the `DeliveryPipeline`.
    - Use `references/configure-skaffold.md` as a reference when generating the `skaffold.yaml` file.

## Release Management

This section covers the various aspects of managing Cloud Deploy `Release` resources.

### Constraints & Rules

In order to manage releases, a `DeliveryPipeline` MUST already be defined and configured in Cloud Deploy. Determine whether a delivery pipeline is defined by checking for a `clouddeploy.yaml` file and checking if the resources exist in Cloud Deploy or ask the user directly. 

**Required Context**:
  - **Project ID**: The Google Cloud project ID of the `DeliveryPipeline`.
  - **Region**: The region of the `DeliveryPipeline` (e.g., `us-central1`).
  - **Delivery Pipeline ID**: The `DeliveryPipeline` ID.

### Create a release

**Use case**: The user wants to deploy a new version of their application.

**Required Context**: Before creating a `Release`, you **MUST** know whether the users runtime manifests are using a placeholder for the container image and the value of the placeholder. This is **CRITICAL** for build artifact substitution in Cloud Deploy. Check the users runtime manifests or ask the user directly. See examples in `references/basic-cloudrun-manifests.md` and `references/basic-k8s-manifests.md`.

Run the following command to create a `Release` for the `DeliveryPipeline`:

```bash
gcloud deploy releases create release-$DATE-$TIME --delivery-pipeline=<DELIVERY_PIPELINE_ID> --region=<REGION> --project=<PROJECT_ID>
```

If the user is leveraging build artifact substitution with a placeholder in the image field of the runtime manifests then use the `--images` flag:

```bash
gcloud deploy releases create release-$DATE-$TIME --delivery-pipeline=<DELIVERY_PIPELINE_ID> --region=<REGION> --project=<PROJECT_ID> --images <IMAGE_PLACEHOLDER>=<IMAGE_URI>
```

**CRITICAL**: If the `skaffold.yaml` is not in the current directory, use the `--source` flag to specify the directory where the `skaffold.yaml` file is located.

Reference documenation for `gcloud deploy releases create`: https://docs.cloud.google.com/sdk/gcloud/reference/deploy/releases/create.

### Promote a release

**Use case**: The user wants to promote the application to the next target environment in the `DeliveryPipeline` progression sequence.

Run the following command to promote a `Release` to the next target in the `DeliveryPipeline` progression sequence:

```bash
gcloud deploy promote --release=<RELEASE_ID> --delivery-pipeline=<DELIVERY_PIPELINE_ID> --region=<REGION> --project=<PROJECT_ID>
```

Reference documentation for `gcloud deploy releases promote`: https://docs.cloud.google.com/sdk/gcloud/reference/deploy/releases/promote.

### Monitor a release

**Use case**: Monitor the status of a release across a `DeliveryPipeline`.

Monitoring a release across a `DeliveryPipeline` consists of checking the status of both the `Release` resource and its child `Rollout` resource(s). Always ensure that the `Release` has completed successfully before checking the status of the `Rollout`.

### Troubleshoot

#### Release failed

Get the release to determine which of the target renders failed and inspect the failure message and failure cause. Additionally the target renders contain a Cloud Build reference where the target render was executed. Retrieve the build logs to determine the root cause of the failure.

#### Rollout failed

Get the rollout to determine which of the jobs failed and inspect the failure message and failure cause. Additionally the `Rollout` contains a Cloud Build reference where the failed job was executed. Retrieve the build logs to determine the root cause of the failure.
