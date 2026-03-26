# IAM Permissions

This document provides insights into the IAM permissions required for Cloud Deploy to operate based on the features enabled for the `DeliveryPipeline`.

## Execution Service Account

**MUST KNOW**: The execution service account used by Cloud Deploy. This is either the default Compute Engine service account or a user-provided service account that was defined in the `Target`.

The execution service account **must always** have the following roles:
* `roles/clouddeploy.jobRunner`
* `roles/iam.serviceAccountUser`

### Runtime Permissions

If deploying to Cloud Run then `roles/run.developer` is **required**.

If deploying to GKE then `roles/container.developer` is **required**.

### Analysis Permissions (Google Cloud Observability)

The following roles are **required**:
* `roles/monitoring.alertViewer`
* `roles/serviceusage.serviceUsageConsumer`

## Automation Service Account

**MUST KNOW**: The service account defined in the `Automation` resources.

The `Automation` service account **requires** the `roles/clouddeploy.operator` role.

## Release Creator

The user or service account that creates a `Release` and `Rollout` **must** have:
* `roles/clouddeploy.releaser`
* `roles/iam.serviceAccountUser`
