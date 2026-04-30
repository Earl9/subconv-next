# Security Policy

SubConv Next is a local or trusted-LAN subscription conversion tool. The Web UI has no built-in login in V1. Do not expose the management page directly to the public Internet unless it is protected by HTTPS, a reverse proxy, authentication, VPN, or equivalent access control.

## Supported Versions

| Version | Supported |
| --- | --- |
| `main` / latest | Yes |
| V1.x | Yes, after V1 tags are published |
| Older snapshots | No |

## Security Boundaries

- SubConv Next converts and serves Mihomo / Clash Meta YAML. It is not a public authentication gateway.
- `/healthz` is unauthenticated by design and returns only basic health, version, data directory, and uptime.
- `/s/{token}/mihomo.yaml` is a bearer-style private subscription link. Anyone holding the URL can fetch the generated YAML.
- If a private subscription link leaks, use `重新生成私密链接` in the Web UI. The old token is invalidated immediately.
- Deleting a published subscription removes the published directory and invalidates the old link.
- Public Internet use must be placed behind HTTPS and access control.

## Sensitive Information

The project attempts to redact sensitive values in APIs and logs, including:

- Upstream subscription URL tokens and query secrets.
- Published subscription tokens.
- Node `password` values.
- Node `uuid` values.
- WireGuard `private-key` values.
- WireGuard `pre-shared-key` values.
- `Authorization` headers.
- `Cookie` headers.

Logs should show only token hints such as `abcd...wxyz`, redacted subscription paths such as `/s/<redacted>/mihomo.yaml`, or host-level URL information.

## Private Subscription Links

Published links use this form:

```text
/s/{token}/mihomo.yaml
```

Treat the full URL like a password. Do not paste it into public issues, screenshots, logs, or chat. If the link is shared by mistake, rotate it immediately. A normal `重新生成配置` keeps the same link; `重新生成私密链接` rotates the token and makes the old link return `404`.

## Local Drafts

Local drafts are stored in the current browser only after the user explicitly clicks `保存为本机草稿` or `更新本机草稿`.

Local drafts may store a `publish_id` and `token_hint`, but must not store the full published token or full `/s/{token}/mihomo.yaml` URL. Do not use local drafts on shared or public computers.

## Reporting a Vulnerability

Please open a GitHub Security Advisory or private issue if available.

Do not post real subscription links, published tokens, upstream URLs with tokens, node passwords, UUIDs, WireGuard keys, cookies, authorization headers, or unredacted logs in a public issue.

When reporting, include:

- A minimal reproduction.
- Affected version or commit.
- Deployment mode.
- Redacted logs and redacted configuration.

See [docs/security.md](docs/security.md) for the implementation security model.
