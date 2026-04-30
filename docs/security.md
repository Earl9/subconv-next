# Security Model

SubConv Next is a self-hosted subscription converter. It does not provide a public authentication gateway, proxy core, traffic forwarding, port forwarding, or network scanning.

## Deployment Boundary

- Default development configuration binds the service to `127.0.0.1`.
- The provided `docker-compose.yml` may be changed to `0.0.0.0:9876:9876` for trusted LAN access.
- Do not expose the Web UI directly to the public Internet. Put it behind a VPN, reverse proxy with authentication, or another access-control layer.
- `/healthz` is intentionally unauthenticated and returns only basic service health.

## Published Subscription Links

- Fixed `/sub/mihomo.yaml` is disabled and returns `404`.
- Published subscriptions use `/s/{random-token}/mihomo.yaml`.
- `publish_id` and subscription tokens are generated with `crypto/rand`.
- Each `publish_id` stores one `current.yaml`; normal regeneration overwrites that file.
- Rotating the private link changes only the token. The old token immediately returns `404`.
- Deleting a publish removes its published directory and the old link immediately returns `404`.
- Published YAML responses use `Cache-Control: no-store`, `X-Robots-Tag: noindex, nofollow, noarchive`, and `X-Content-Type-Options: nosniff`.

## Local Draft Privacy

Local draft data is only written after the user explicitly chooses `保存为本机草稿` or `更新本机草稿`.

Allowed in `localStorage.SUBCONV_LOCAL_DRAFT`:

- Editing configuration.
- Subscription source name, emoji, URL, and User-Agent.
- `publish_ref.publish_id`.
- `publish_ref.token_hint`.
- `publish_ref.updated_at`.
- Node edit state needed to restore the editing session.

Forbidden in local draft storage:

- Full `/s/{token}/mihomo.yaml` URL.
- Full subscription token.
- Rendered `current.yaml`.
- Runtime logs.
- `access_count` and `last_access_at`.

Restoring a draft creates a new workspace. If the saved `publish_id` still exists, the workspace is rebound to that publish and future `重新生成配置` overwrites the same `current.yaml` without changing the link.

## API Redaction

The following API surfaces must not expose secrets by default:

- `/api/config` redacts access tokens and subscription URL query values.
- `/api/nodes` and node detail responses mask password, uuid, private key, and pre-shared key fields.
- `/api/logs` returns masked log lines.
- `/api/published/{publish_id}` returns the current subscription URL for the requested publish, but does not return a raw token field.

## Logging

Logs are masked before writing to `/data/logs/app.log`.

Masked values include:

- `/s/{token}/mihomo.yaml`
- URL query parameters such as `token`, `key`, `auth`, `password`, and `uuid`
- URI userinfo secrets such as `ss://password@...`
- UUIDs
- `password`, `uuid`, `private-key`, `pre-shared-key`, `authorization`, and `cookie` key/value pairs

Log rotation keeps at most three rotated files, each up to 5 MB.

## YAML Integrity

The renderer receives final nodes only. Disabled, deleted, excluded, invalid, info, and deduplicated nodes must not enter the final YAML.

After rendering, the YAML is parsed back and validated:

- Every `proxies[].name` exists in the final node set.
- `proxy-groups[].proxies` references only existing node names, group names, or built-ins such as `DIRECT`.
- Excluded nodes are not referenced by proxies or groups.
- Rule providers referenced by rules exist.
- Rule targets exist.
- `MATCH` is the last rule.

## Reporting Security Issues

For a public GitHub release, create a private disclosure path before enabling Issues for security reports. Until then, do not post real subscription URLs, tokens, node credentials, or logs containing unredacted secrets in public issues.
