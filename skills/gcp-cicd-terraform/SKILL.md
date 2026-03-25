---
name: gcp-terraform-cicd
description: "Architect, provision, and troubleshoot production-grade Google Cloud infrastructure using Terraform and OpenTofu. Use to design landing zones (Shared VPCs, Folders), deploy core services (GKE, Cloud Run, Cloud SQL), implement IAM least-privilege, and manage GCS-backed state. Enforces Google’s Cloud Foundation Fabric patterns and rigorous validation protocols to ensure secure, idempotent, and scalable deployments across environments."
version: "0.1.0"
---

# Terraform GCP Skill

This skill provides expert-level guidance for architecting, deploying, and managing GCP infrastructure using Terraform.

## 🎯 When to Use This Skill
- Designing GCP Landing Zones (Projects, Folders, Shared VPCs).
- Provisioning Google Cloud resources (GKE, Cloud Run, Cloud SQL, Spanner).
- Setting up remote state management via **GCS Backend**.
- Implementing IAM least-privilege via Terraform.
- Troubleshooting `terraform plan` discrepancies in GCP.

## 🏗️ Core Architecture: GCS Backend & State
**Always** use a GCS backend for state. Local state is strictly for local prototyping and must never be committed. If local state direcotry is used, add it to .gitignore.

### Standard Backend Configuration
```hcl
terraform {
  backend "gcs" {
    bucket  = "tf-state-${var.project_id}"
    prefix  = "terraform/state/${var.environment}"
  }
}
```

Note: The GCS bucket must have Object Versioning enabled to allow recovery from accidental state corruption or overlapping writes.

### Required Provider Version
Use the Google Cloud Terraform provider version 7.20.0 or higher. This skill utilizes features (e.g., Developer Connect) introduced in Google Provider v7.20.0.

```hcl
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 7.20.0" 
    }
  }
}
```

To retrieve the latest version of the Google provider, use the following command:
```bash
curl -s https://registry.terraform.io/v1/providers/hashicorp/google | jq -r .version
```

## 🛠️ Execution Protocol (Safety First)
The Agent must follow this lifecycle for every infrastructure change to ensure idempotency and prevent production outages:

1. Initialize (`terraform init`):

   - Verify backend connectivity.

   - Ensure provider plugins are downloaded and versions are locked in `.terraform.lock.hcl`.

2. Validate (`terraform validate` & `tflint`):

   - Check for internal HCL consistency.

   - Run `tflint` with the Google plugin to catch cloud-specific deprecated fields or non-optimal configurations.

3. Plan (`terraform plan -out=tfplan`):

   - Generate a speculative execution plan.

   - Mandatory Step: The Agent must summarize the plan for the user, specifically highlighting any Destroy actions.

4. Apply (`terraform apply tfplan`):

   - Strict Constraint: Only execute the application after the user provides explicit manual confirmation.

## 🔍 Live Provider Documentation Lookup
To prevent hallucinations regarding new fields or deprecated attributes, the Agent must verify the grounded truth of the provider schema.

Schema Inspection via CLI
The primary command for retrieving the machine-readable schema of the currently initialized providers is:

```bash
$ terraform providers schema -json
```


## 🛡️ GCP-Specific Standards
1. Resource Naming & "Main" Pattern
To maintain a clean module interface, use the main identifier for singleton resources (the primary resource the module is named after). Use underscores for identifiers and hyphens for names.

    ```hcl
    resource "google_compute_network" "main" {
      name                    = "${var.prefix}-vpc"
      auto_create_subnetworks = false
    }
    ```

2. IAM Management
   - Never use google_project_iam_policy: This resource is authoritative and replaces the entire IAM policy for the project. It is the #1 cause of accidental lockouts.
   - Avoid `google_project_iam_policy`: This resource is authoritative for the entire project and is a common cause of accidental lockouts.

   - Prefer `google_project_iam_member` or `google_project_iam_binding`:
     - `google_project_iam_member` is additive and safely grants a role to a single member.
     - `google_project_iam_binding` is authoritative for a single role. It's useful for managing all members of a role, but be aware it overwrites existing members for that role.
   - Shared VPC: Always distinguish between Host projects (where the network lives) and Service projects (where resources consume the network).
   - Private Google Access: Subnets should always have private_ip_google_access = true.
   - Workload Identity: Prefer GKE Workload Identity over static Service Account JSON keys.

3. Cloud Build Triggers with Developer Connect
   When using Developer Connect git repository links, use `developer_connect_event_config` — NOT `repository_event_config`. The `repository_event_config` block is for Cloud Build v2 repository connections and will not work with Developer Connect resources. An example block to create a Cloud Build trigger with Developer Connect git repository link is as follows:

    ```hcl
    resource "google_cloudbuild_trigger" "main" {
      developer_connect_event_config {
        git_repository_link = google_developer_connect_git_repository_link.main.id
        push {
          branch = var.trigger_branch
        }
      }
    }
    ```

4. Developer Connect Connection
   When configuring `google_developer_connect_connection`, always set `github_app` to `"DEVELOPER_CONNECT"`. Using `"FIREBASE"` is incorrect and will cause triggers to fail.

    ```hcl
    resource "google_developer_connect_connection" "main" {
      location      = var.region
      connection_id = var.connection_id
      project       = var.project_id

      github_config {
        github_app = "DEVELOPER_CONNECT" # CORRECT
        authorizer_credential {
          oauth_token_secret_version = "" # Populated after manual authorization
        }
      }

      depends_on = [google_project_service.main["developerconnect.googleapis.com"]]

      lifecycle {
        ignore_changes = [github_config[0].authorizer_credential]
      }
    }
    ```

## 📂 Directory Structure
Follow this standard to ensure compatibility with Antigravity (AGY) discovery and Google best practices:

```
.
└── terraform/
    ├── main.tf              # Entry point / Resource definitions
    ├── variables.tf         # Typed variables with units and descriptions
    ├── outputs.tf           # Resource ID outputs (no direct input pass-through)
    ├── versions.tf          # Provider version pinning
    ├── network.tf           # (Optional) Grouped networking resources
    ├── examples/            # Example usage for modules
    ├── files/               # Static files (startup scripts, etc.)
    ├── templates/           # .tftpl templates
    ├── scripts/             # Scripts called by Terraform
    └── helpers/             # Scripts NOT called by Terraform
```
## ⚠️ Anti-Patterns (Do NOT do these)
   - ❌ Hardcoded IDs: Never hardcode Project IDs. Use variables or data "google_project" sources.

   - ❌ Service Account Keys: Never generate or store .json keys. Use Workload Identity Federation or the default metadata server.

   - ❌ Hardcoded IDs: Never hardcode Project IDs. Use variables or a `data "google_project"` source.

   - ❌ Broad Scopes: Avoid cloud-platform scopes for GKE nodes; use fine-grained IAM roles instead.

## 🧪 Testing Strategy
   - Static Analysis: Use checkov, trivy, or terrascan to catch insecure GCP configurations (e.g., public GCS buckets).

   - Integration Testing: Use terraform test (v1.6+) to assert that GCP labels and network tags are correctly applied to resources before full deployment.
