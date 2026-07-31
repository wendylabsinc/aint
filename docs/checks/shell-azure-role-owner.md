# shell-azure-role-owner

**Flags:** `az role assignment create ... --role Owner` (with or without
quotes around `Owner`).

**Why it matters:** the Azure `Owner` role grants full control over the
scoped resource, including the ability to manage access for other users —
it's the broadest built-in role available. Granting it where a narrower
role would do increases the blast radius of a compromised identity and
makes it harder to reason about who can change access controls.

**Fix:** grant a narrower built-in role (e.g. `Contributor` if management
access isn't needed, or a service-specific role like
`Storage Blob Data Reader`) scoped to only the resource(s) required.

```bash
# Flags this:
az role assignment create --assignee alice@example.com --role Owner --scope /subscriptions/xxx

# Prefer this:
az role assignment create --assignee alice@example.com \
  --role "Storage Blob Data Reader" \
  --scope /subscriptions/xxx/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/mystorage
```
