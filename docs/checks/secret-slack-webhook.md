# secret-slack-webhook

**Flags:** a hardcoded Slack incoming webhook URL
(`hooks.slack.com/services/...`).

**Why it matters:** a Slack webhook URL is a bearer credential — anyone who
has it can post messages to that channel as the configured integration, no
further authentication required. Once it's in source control (even in
history after a "removal" commit), it should be treated as compromised and
rotated.

**Fix:** load the webhook URL from an environment variable or secrets
manager instead of hardcoding it, and regenerate the webhook in Slack if
one has already been committed.

```go
// Flags this:
webhookURL := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

// Prefer this:
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```
