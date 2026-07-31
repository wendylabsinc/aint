# shell-docker-privileged

**Flags:** `docker run` with either `--privileged` or a bind-mount of
`/var/run/docker.sock`.

**Why it matters:** `--privileged` disables most of the container isolation
that Docker normally provides — the container gets access to all host
devices and can effectively act as root on the host. Mounting the Docker
socket has the same practical effect via a different path: a process with
access to `docker.sock` can start a new privileged container and use it to
read/write anywhere on the host filesystem. Both amount to giving the
container host root.

**Fix:** avoid `--privileged`; grant only the specific capabilities the
container needs via `--cap-add`. Avoid mounting the Docker socket into
containers that don't strictly need to control the Docker daemon — if one
does, treat it as equivalent to giving that container root on the host.

```bash
# Flags this:
docker run --privileged -it myimage

# Prefer this (grant only what's needed):
docker run --cap-add=NET_ADMIN -it myimage
```
