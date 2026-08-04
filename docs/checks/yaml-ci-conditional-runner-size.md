# yaml-ci-conditional-runner-size

**Flags:** a `runs-on:` value using a ternary GitHub Actions expression based
on event type, e.g. `runs-on: ${{ github.event_name == 'pull_request' &&
'c7i.8xlarge' || 'c7i.24xlarge' }}`.

**Why it matters:** conditionally downsizing the runner for PRs versus
nightly/main builds was explicitly rejected in favor of a consistent, larger
instance for every event type - inconsistent runner sizing makes build times
unpredictable and re-introduces a config knob that was deliberately removed.

**Fix:** use one instance size unconditionally for all event types on a given
job.

```yaml
# Flags this:
runs-on: ${{ github.event_name == 'pull_request' && 'c7i.8xlarge' || 'c7i.24xlarge' }}

# Prefer this:
runs-on: c7i.24xlarge
```
