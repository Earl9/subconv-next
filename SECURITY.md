# Security Policy

## Supported Versions

| Version | Supported |
| --- | --- |
| `main` / latest | Yes |
| V1.x | Yes |
| Older snapshots | No |

## Security Boundary

SubConv Next is a local or trusted-LAN subscription conversion tool. The V1 Web UI has no built-in login and should not be exposed directly to the public Internet.

For public access, place it behind HTTPS, authentication, a reverse proxy, VPN, or equivalent access control.

## Sensitive Data

SubConv Next attempts to redact sensitive values in APIs and logs, including upstream subscription URL tokens, published subscription tokens, passwords, UUIDs, WireGuard private keys, pre-shared keys, `Authorization`, and `Cookie` values.

Do not publish real subscription URLs, tokens, node secrets, or unredacted logs in public issues.

## Private Subscription Links

Published subscriptions use bearer-style private URLs:

```text
/s/{token}/mihomo.yaml
```

Anyone holding the full URL can fetch the generated YAML. If a link leaks, use `重新生成私密链接` in the Web UI; the old token is invalidated immediately. Deleting a published subscription also invalidates its old URL.

## Reporting a Vulnerability

Please open a GitHub Security Advisory or private issue if available.

Include the affected version or commit, deployment mode, a minimal reproduction, and redacted logs or configuration. Do not include real tokens, upstream subscription links, node passwords, UUIDs, private keys, cookies, authorization headers, or full published subscription URLs.
