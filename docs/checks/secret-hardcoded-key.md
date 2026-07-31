# secret-hardcoded-key

**Flags:** hardcoded API keys or tokens in source code, including AWS access keys, OpenAI API keys, and literal secrets assigned to variables.

**Why it matters:** if credentials are committed to a repository, anyone with access to the repo — or the repository history, even after deletion — can impersonate your account, access protected resources, or run up charges. Secrets should live only in environment variables, secure vaults, or configuration files kept outside version control.

**Fix:** remove the hardcoded secret, regenerate it in your service's control panel, and store it securely as an environment variable or in a secret management tool (e.g., `process.env.API_KEY`, `os.getenv("SECRET")`). If a secret was accidentally committed, revoke it immediately and rewrite Git history or alert your ops team.

```python
# Flags this:
api_key = "sk-abcdefghijklmnop1234"

# Prefer this:
import os
api_key = os.getenv("API_KEY")  # set via environment, not code
```
