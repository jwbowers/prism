# Workspace Lifecycle

This guide covers how Prism manages the lifecycle of your workspaces — from auto-shutdown with time limits, to DNS hostnames, to idle detection that saves costs when you're not working.

---

## Time-to-Live (TTL)

A TTL is a hard time limit on a workspace. Once the TTL expires, the workspace is stopped automatically.

### Setting a TTL at launch

```bash
prism workspace launch python-ml my-project --ttl 8h
prism workspace launch r-research paper-analysis --ttl 24h
prism workspace launch genomics long-run --ttl 48h
```

Duration format: `h` (hours), `m` (minutes), `s` (seconds). Examples: `8h`, `30m`, `1h30m`.

### What happens when the TTL expires

1. The **spored daemon** on the instance checks the TTL continuously.
2. **5 minutes before expiration**: `wall` broadcast to all logged-in users — "Workspace will stop in 5 minutes".
3. **At expiration**: spored stops (or terminates, depending on `--on-complete`) the workspace.
4. **Safety valve**: the Prism daemon also monitors `ExpiresAt` every 10 seconds and stops the workspace if spored fails to act.

### Extending the time limit

In the GUI, click the **Actions** dropdown on the workspace row and select **Extend Time (+4h)**. This adds 4 hours to the current expiry and updates the `prism:ttl` EC2 tag so spored picks up the change.

Via the daemon API:
```bash
curl -X POST http://localhost:8947/api/v1/instances/my-project/extend \
  -H "X-API-Key: $(cat ~/.prism/state.json | jq -r .Config.APIKey)" \
  -d '{"hours": 4}'
```

### Viewing time remaining

The instance table in the GUI shows a **Time Remaining** column with color-coded status:
- 🟢 Green — more than 2 hours
- 🟡 Yellow — less than 2 hours
- 🔴 Red — less than 1 hour or expired

---

## DNS Hostnames (prismcloud.host)

Every Prism workspace automatically registers a DNS hostname at launch.

### Hostname format

```
{name}.{account-base36}.prismcloud.host
```

For example, if your workspace is named `my-project` and your AWS account ID encodes to `abc123`:
```
my-project.abc123.prismcloud.host
```

The account base36 encoding is a compact representation of your AWS account ID — it scopes hostnames so different AWS accounts never collide.

### Custom DNS names

Use `--dns` to specify a different record name:

```bash
prism workspace launch python-ml my-project --dns ml-ws
# → Registers: ml-ws.abc123.prismcloud.host
```

If `--dns` is not specified, the workspace name is sanitized (lowercased, non-alphanumeric characters replaced with hyphens) and used as the DNS name.

### Using the hostname

The hostname appears in the **Hostname** column of the instance table and is used automatically in the SSH connect command:

```bash
ssh ubuntu@ml-ws.abc123.prismcloud.host
```

DNS registration happens automatically via the [spored daemon](#spored-daemon) when the workspace boots. The A record is deleted when the workspace stops.

> **Note**: DNS propagation takes up to 60 seconds after workspace boot.

---

## Idle Detection

Prism automatically detects when a workspace is idle and can hibernate or stop it to save costs. Idle detection runs on the instance via the spored daemon — it uses 7 independent signals that are more accurate than polling CloudWatch.

### How idle detection works

The spored daemon checks the following signals:

| Signal | Default threshold |
|--------|-----------------|
| CPU utilization | < 5% average |
| Network throughput | < 10 KB/min |
| Disk I/O | < 100 KB/min |
| GPU utilization (if present) | < 5% |
| Active terminal sessions (`/dev/pts`) | None active |
| Logged-in users (`who`) | None |
| Recent user activity (`last`) | No activity in threshold window |

All signals must be below threshold simultaneously for the idle timeout to trigger.

### Configuring idle detection

Idle detection is configured via your idle policy. Apply a policy when launching:

```bash
prism workspace launch python-ml my-project --idle-policy
```

Or override the threshold directly:

```bash
prism workspace launch python-ml my-project --idle-timeout 2h
```

This sets `prism:idle-timeout` on the EC2 instance tags, which spored reads at startup.

### Built-in idle policies

| Policy | Idle threshold | Action | Best for |
|--------|---------------|--------|----------|
| `aggressive` | 15 minutes | Hibernate | Dev/test |
| `balanced` | 60 minutes | Hibernate | General use |
| `conservative` | 4 hours | Stop | Long-running jobs |
| `research` | 2 hours | Hibernate | ML/data science |

Manage policies in the GUI via **Settings → Idle Detection**, or via:
```bash
prism admin idle-policy list
prism admin idle-policy apply <workspace> --policy balanced
```

---

## Workspace States

| State | Cost | Description |
|-------|------|-------------|
| `running` | Compute + storage | Instance is running and accessible |
| `stopping` | Compute (brief) | Shutdown in progress |
| `stopped` | Storage only | Instance stopped; data preserved |
| `hibernated` | Storage only | RAM state saved to EBS; resumes in seconds |
| `pending` | Compute | Starting up |
| `terminated` | None | Permanently deleted |

### Cost comparison (example m5.xlarge, us-west-2)

| State | Hourly cost |
|-------|------------|
| Running | ~$0.192/hr |
| Stopped | ~$0.008/hr (EBS only) |
| Hibernated | ~$0.008/hr (EBS only) |
| Terminated | $0.00 |

Hibernation and stopping save ~96% on compute costs while preserving your work.

---

## Spot Interruption Handling

If a workspace is launched with `--spot`, spored monitors the EC2 Instance Metadata Service (IMDS) every 5 seconds for a spot interruption notice. When a 2-minute warning is received:

1. spored broadcasts a `wall` warning to all users
2. Pre-stop hook is executed (if configured via `prism:pre-stop` tag)
3. DNS record is deleted
4. Instance terminates (or stops, per `prism:on-complete` tag)

The 2-minute window gives you time to save work if you're actively using the workspace.
