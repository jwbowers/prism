# spored — Instance Lifecycle Daemon

**Added in v0.34.0** — every Prism workspace automatically runs the spored daemon.

## What is spored?

spored is a lightweight Go service that runs on every Prism EC2 instance as a systemd service. It is the on-instance counterpart to the Prism control plane daemon — while `prismd` manages instances remotely via AWS APIs, spored operates locally on the instance itself.

spored is developed in the [spore-host](https://github.com/scttfrdmn/spore-host) project and deployed by Prism with Prism-specific configuration.

## Capabilities

| Capability | Description |
|-----------|-------------|
| **Idle detection** | Monitors 7 local signals (CPU, network, disk I/O, GPU, terminals, users, activity); stops/hibernates when idle |
| **TTL enforcement** | Counts down from `prism:ttl` EC2 tag; warns users 5 min before expiry; stops/terminates at expiry |
| **DNS registration** | Registers `{name}.{account}.prismcloud.host` A record on boot; deletes on shutdown |
| **Spot interruption** | Polls IMDS every 5s; runs pre-stop hook and DNS cleanup within 2-minute warning window |
| **Cost limit** | Stops instance when accumulated cost exceeds `prism:cost-limit` tag value |
| **Pre-stop hooks** | Runs `prism:pre-stop` shell command before any shutdown (configurable timeout) |
| **Status reporting** | `spored status` shows TTL remaining, idle detection state, config |
| **Config reload** | `spored reload` re-reads EC2 tags without restart |

## Automatic Installation

spored is installed automatically during EC2 UserData on every workspace launch. The installation:

1. Detects architecture (x86_64 → amd64, aarch64 → arm64)
2. Detects AWS region from IMDS
3. Downloads binary from `s3://spawn-binaries-{region}/prism/spored-linux-{arch}`
4. Verifies SHA256 checksum (binary deleted if verification fails)
5. Writes systemd service unit to `/etc/systemd/system/spored.service`
6. Enables and starts the service

## Configuration

spored reads configuration from `prism:*` EC2 instance tags at startup. The Prism daemon sets these tags at launch based on `--ttl`, `--dns`, `--idle-timeout`, and `--idle-policy` flags.

| EC2 Tag | Example | Description |
|---------|---------|-------------|
| `prism:dns-name` | `my-workspace` | DNS record name for prismcloud.host |
| `prism:ttl` | `8h` | Time-to-live duration |
| `prism:idle-timeout` | `1h` | Idle detection threshold |
| `prism:hibernate-on-idle` | `true` | Hibernate instead of stop on idle |
| `prism:on-complete` | `stop` | Action when completion file appears |
| `prism:cost-limit` | `10.00` | Maximum spend in USD |
| `prism:price-per-hour` | `0.192` | On-demand hourly rate (for cost tracking) |
| `prism:pre-stop` | `/home/ubuntu/cleanup.sh` | Shell command before shutdown |

Tags can be updated at runtime — run `spored reload` on the instance or use the EC2 console/CLI to update tags and spored will pick up the new values.

## systemd Service

```ini
[Unit]
Description=Prism Instance Daemon (spored)
After=network-online.target cloud-final.service
Wants=network-online.target

[Service]
Type=simple
Environment=SPORED_TAG_PREFIX=prism
Environment=SPORED_DNS_DOMAIN=prismcloud.host
ExecStart=/usr/local/bin/spored
Restart=on-failure
RestartSec=10
TimeoutStopSec=30
NoNewPrivileges=true
LimitNOFILE=8192

[Install]
WantedBy=multi-user.target
```

Key security properties:
- `NoNewPrivileges=true` — prevents privilege escalation
- Runs as `ec2-user` (default) — no root required
- `Restart=on-failure` — auto-recovers from crashes

## DNS Architecture

spored registers workspaces with `prismcloud.host` via a shared Lambda function (the same infrastructure that powers `spore.host`).

```
Instance boot
  └── spored starts
        └── reads prism:dns-name from EC2 tags
              └── fetches IMDS instance identity document + signature
                    └── POSTs signed request to DNS Lambda
                          └── Lambda validates instance identity
                                └── Lambda creates Route53 A record:
                                      {name}.{account-base36}.prismcloud.host → public IP

Instance stop
  └── spored Cleanup()
        └── deletes Route53 A record
```

The account base36 encoding (`strconv.FormatInt(accountID, 36)`) scopes all workspaces under a per-account subdomain, preventing hostname collisions across AWS accounts.

## Monitoring

Check spored status on a running instance via SSH:

```bash
ssh ubuntu@my-workspace.abc123.prismcloud.host spored status
```

Or inspect systemd logs:

```bash
ssh ubuntu@my-workspace.abc123.prismcloud.host sudo journalctl -u spored -f
```

## Relation to Prism Daemon

| Concern | prismd (local control plane) | spored (on-instance) |
|---------|------------------------------|----------------------|
| Idle detection | Policy management + TTL safety valve | Actual 7-signal monitoring |
| TTL enforcement | Sets ExpiresAt at launch; stops if spored fails | Primary enforcer |
| DNS | Constructs hostname for display | Registers/deregisters A records |
| Lifecycle actions | Start/stop/terminate via EC2 API | Stop/hibernate from within instance |
| AWS API calls | Full access | Self-management only (stop/terminate self) |
