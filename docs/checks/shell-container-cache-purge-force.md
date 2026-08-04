# shell-container-cache-purge-force

**Flags:** a forced, non-interactive container build-cache purge - `docker
builder prune -f`/`-af`, `docker image prune -af`, `docker container prune
-f`, or `container builder delete --force`.

**Why it matters:** these commands are the biggest disk-space reclaim
available on a dev machine running multiple container runtimes, but the
`--force`/`-f` flag skips the interactive confirmation and nukes the build
cache outright - the next build runs cold/slow, and the size of what's being
deleted (can be hundreds of GB) is easy to underestimate.

**Fix:** confirm with the user before running a forced purge, especially
`container builder stop && container builder delete --force`, which recreates
the BuildKit cache from scratch on the next build.

```bash
# Flags this:
docker builder prune -af

# Prefer this: run it yourself after confirming, or without -f/-a to review
# what would be deleted first
docker builder prune
```
