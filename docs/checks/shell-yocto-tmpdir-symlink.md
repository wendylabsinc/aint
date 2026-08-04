# shell-yocto-tmpdir-symlink

**Flags:** `ln -s ... build/tmp` (or any symlink command mentioning
`TMPDIR`) - pointing Yocto's build scratch directory at a symlink instead of
a real path/bind mount.

**Why it matters:** Yocto's `pseudo` (fakeroot) records canonical absolute
paths and does not tolerate `TMPDIR` sitting behind a symlink. When
`build/tmp` is a symlink, rootfs-time postinstall scriptlets run under pseudo
fail during `do_rootfs`, and the error surfaces as an unrelated-looking
failure (e.g. `rpcbind` postinst) - `rpcbind` is just the first package in
install order to hit it, not the actual cause. It's deterministic (fails on
retry) and hits every board/branch on the affected host.

**Fix:** bind-mount the scratch onto `build/tmp` (a real path) instead of
symlinking it - this keeps `DEPLOY_DIR` at `build/tmp/deploy` for upload steps
while keeping `pseudo` happy.

```bash
# Flags this:
ln -s /wendy/build-tmp build/tmp

# Prefer this:
mkdir -p build/tmp
mount --bind /wendy/build-tmp build/tmp
```
