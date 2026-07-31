# shell-aws-iam-attach-admin

**Flags:** `aws iam attach-user-policy` / `attach-role-policy` /
`attach-group-policy` with `--policy-arn` pointing at the AWS-managed
`AdministratorAccess` policy.

**Why it matters:** `AdministratorAccess` grants full access to every AWS
service and resource in the account, including the ability to modify IAM
itself. Attaching it is almost always broader than what the user, role, or
group actually needs, and it turns any compromise of that principal into a
full account takeover.

**Fix:** attach a narrower AWS-managed policy scoped to the services
actually needed, or write a custom policy with only the required
permissions.

```bash
# Flags this:
aws iam attach-role-policy --role-name deploy \
  --policy-arn arn:aws:iam::aws:policy/AdministratorAccess

# Prefer this:
aws iam attach-role-policy --role-name deploy \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```
