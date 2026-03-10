# Troubleshooting GCP-Terraform Errors

A guide to identifying and resolving common errors encountered when managing GCP resources with Terraform.

## 🚫 403 Forbidden / Permission Denied
**Symptoms**: `Error 403: Permission '...' denied on resource '...'`.

**Fixes**:
1. **Service Usage API**: Ensure the `serviceusage.googleapis.com` API is enabled in the project.
2. **Missing Roles**: Verify the principal (Service Account or User) has the required IAM roles. Use `google_project_iam_member` to add them.
3. **Propagation Delay**: IAM changes can take up to 60 seconds to propagate. Retry the operation.

## ⏳ Resource In Use / Deletion Delays
**Symptoms**: `Error: Resource ... is still in use` or `Error deleting Subnetwork: ...`.

**Fixes**:
1. **Dependency Chains**: GCP often has hidden dependencies. For example, a `google_compute_subnetwork` cannot be deleted if a `google_compute_address` or a VM is still using it.
2. **Force Destroy**: For `google_storage_bucket`, use `force_destroy = true` if you want Terraform to delete a bucket containing objects.
3. **Sequential Deletion**: Sometimes `terraform destroy` struggles with complex VPC peering. Try destroying resources in smaller chunks if a timeout occurs.

## 🔒 GCS State Locking
**Symptoms**: `Error: Error acquiring the state lock: ... GCS bucket ... already locked`.

**Fixes**:
1. **Check for Active Processes**: Ensure no other developer or CI job is running.
2. **Manual Unlock**: If a process crashed and left the lock behind, use the force-unlock command with the Lock ID provided in the error message:
   ```bash
   terraform force-unlock <LOCK_ID>
   ```

## 🛠️ Provider Schema Inspection
To see exactly what fields a resource supports and avoid "Attribute not found" errors:

```bash
# Export the full provider schema to JSON
terraform providers schema -json > schema.json

# Use jq to find specific resource details
cat schema.json | jq '.provider_schemas["registry.terraform.io/hashicorp/google"].resource_schemas["google_compute_instance"]'
```
