# shell-aws-iam-wildcard

**Flags:** an IAM policy document (inline JSON in a CLI invocation, e.g.
`aws iam put-role-policy --policy-document '...'`) containing
`"Action": "*"` or `"Resource": "*"`.

**Why it matters:** a wildcard `Action` or `Resource` grants every possible
action, or applies a permission to every possible resource. Combined, they
grant effectively unrestricted access. If the credentials used to assume
this policy are ever compromised, the blast radius is the entire account
rather than a scoped set of actions/resources.

**Fix:** enumerate the specific actions and resource ARNs the policy
actually needs, following the principle of least privilege.

```bash
# Flags this:
aws iam put-role-policy --role-name deploy --policy-name full-access \
  --policy-document '{"Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}'

# Prefer this:
aws iam put-role-policy --role-name deploy --policy-name scoped-s3-read \
  --policy-document '{"Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::my-bucket/*"}]}'
```
