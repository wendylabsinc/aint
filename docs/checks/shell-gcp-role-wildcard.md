# shell-gcp-role-wildcard

**Flags:** GCP IAM policy bindings that grant `roles/owner` or `roles/editor` — the two broadest, most dangerous predefined roles.

**Why it matters:** owner and editor roles grant almost all permissions in a GCP project. Following the principle of least privilege, users and service accounts should have only the specific roles they need. Broad roles increase the blast radius if credentials are compromised.

**Fix:** replace the broad role with a narrower predefined role (e.g., `roles/compute.instanceAdmin.v1`, `roles/storage.objectViewer`) or a custom IAM role with only the required permissions. Review GCP's role reference to find the appropriate scope.

```bash
# Flags this:
gcloud projects add-iam-policy-binding my-project \
  --member=user:alice@example.com \
  --role=roles/owner

# Prefer this:
gcloud projects add-iam-policy-binding my-project \
  --member=user:alice@example.com \
  --role=roles/compute.instanceAdmin.v1
```
