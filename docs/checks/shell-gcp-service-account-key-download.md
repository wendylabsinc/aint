# shell-gcp-service-account-key-download

**Flags:** `gcloud iam service-accounts keys create ...`.

**Why it matters:** a downloaded service account key is a long-lived
credential file that has to be stored, distributed, and rotated manually —
if it leaks (committed to a repo, left on a laptop, exposed in a CI log),
it's usable until someone notices and revokes it. GCP's workload identity
federation lets workloads authenticate without any long-lived key ever
existing on disk.

**Fix:** prefer workload identity federation (for CI/CD and workloads
running outside GCP) or attached service accounts (for workloads running
on GCP compute) instead of creating and downloading a key file.

```bash
# Flags this:
gcloud iam service-accounts keys create key.json \
  --iam-account=sa@project.iam.gserviceaccount.com

# Prefer this: configure workload identity federation instead
gcloud iam workload-identity-pools create my-pool --location=global
gcloud iam workload-identity-pools providers create-oidc my-provider \
  --workload-identity-pool=my-pool --location=global \
  --issuer-uri=https://token.actions.githubusercontent.com
```
