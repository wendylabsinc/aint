# shell-chmod-permissive

**Flags:** chmod commands that grant world-writable permissions (`777`, `a+rwx`, or `ugo+rwx`), with or without `-R`.

**Why it matters:** world-writable files and directories allow any user on the system to modify them, potentially enabling privilege escalation, data corruption, or unauthorized access. This is rarely intentional and exposes the system to local attack.

**Fix:** use a restrictive permission mode that grants write access only to the owner and necessary group members. Common safe modes are `755` (rwxr-xr-x for directories) and `644` (rw-r--r-- for files). If group write access is needed, use `775` or `664` instead.

```bash
# Flags this:
chmod 777 /var/www/uploads
chmod -R a+rwx /opt/app/data

# Prefer this:
chmod 755 /var/www/uploads           # rwx for owner, rx for group and others
chmod -R 775 /opt/app/data          # rwx for owner and group, rx for others
chmod 644 config.txt                # rw for owner, r for group and others
```
