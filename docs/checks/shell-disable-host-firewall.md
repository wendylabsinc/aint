# shell-disable-host-firewall

**Flags:** `setenforce 0` (disabling SELinux enforcement), `ufw disable`,
or `systemctl stop firewalld`.

**Why it matters:** these commands turn off a host-level defense-in-depth
layer — SELinux mandatory access control or the host firewall. They're
sometimes used as a quick fix to get past a permission or connectivity
error while debugging, but leaving them disabled (or running them in a
provisioning script that lands in production) removes a real barrier
against lateral movement and unauthorized network access.

**Fix:** find and fix the actual SELinux policy or firewall rule that's
blocking the legitimate action, rather than disabling the whole subsystem.

```bash
# Flags this:
setenforce 0

# Prefer this: diagnose and allow the specific denial
audit2allow -a -M mypolicy
semodule -i mypolicy.pp

# Flags this:
ufw disable

# Prefer this: open only the port that's needed
ufw allow 8080/tcp
```
