# IAM and Security Standards for GCP

Prevent outages and security breaches by following additive IAM patterns and leveraging Workload Identity.

## 🚫 Additive vs. Authoritative IAM
**NEVER** use authoritative resources (`google_project_iam_policy` or `google_project_iam_binding`) for core projects unless you are explicitly managing the *entire* policy and understand the risk of lockouts.

| Resource | Type | Impact | Usage |
| :--- | :--- | :--- | :--- |
| `google_project_iam_member` | **Additive** | Adds a single member/role. Safe. | **Preferred** |
| `google_project_iam_binding` | **Authoritative (Role)** | Overwrites all members for a specific role. | Use with caution. |
| `google_project_iam_policy` | **Authoritative (Project)** | Overwrites the entire project IAM policy. | **BANNED** for general use. |

### Example: Safe Additive IAM
```hcl
resource "google_project_iam_member" "storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.this.email}"
}
```

## 🆔 Workload Identity (GKE)
Bind Kubernetes Service Accounts (KSA) to Google Service Accounts (GSA) without managing JSON keys.

```hcl
# 1. The Google Service Account
resource "google_service_account" "gsa" {
  account_id   = "app-workload-sa"
  project      = var.project_id
}

# 2. The IAM Binding for Workload Identity
resource "google_project_iam_member" "workload_identity_user" {
  project = var.project_id
  role    = "roles/iam.workloadIdentityUser"
  member  = "serviceAccount:${var.project_id}.svc.id.goog[${var.k8s_namespace}/${var.k8s_sa_name}]"
}
```

## 🏛️ Organization Policy Constraints
Common constraints to enforce during `terraform validate` or `plan`:

- `constraints/compute.disableSerialPortAccess`: Prevents unencrypted console access.
- `constraints/compute.vmExternalIpAccess`: Blocks external IPs for VMs.
- `constraints/iam.allowedPolicyMemberDomains`: Restricts IAM to specific Workspace/Cloud Identity domains.

```hcl
resource "google_project_organization_policy" "disable_serial_port" {
  project    = var.project_id
  constraint = "constraints/compute.disableSerialPortAccess"

  boolean_policy {
    enforced = true
  }
}
```
