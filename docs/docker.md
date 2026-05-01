# Docker Deployment

SubConv Next is Docker-first for V1. The image contains the Go binary and embedded Web UI assets; no Node.js runtime is required.

## Quick Start

```sh
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

Open:

```text
http://127.0.0.1:9876/
```

The image runs `/usr/bin/subconv-next` directly. It does not include a Go toolchain or Node.js runtime in the final stage.

For trusted LAN access, keep the default compose port mapping:

```yaml
ports:
  - "0.0.0.0:9876:9876"
```

For local-only access, change it to:

```yaml
ports:
  - "127.0.0.1:9876:9876"
```

For public internet deployment, put SubConv Next behind a TLS reverse proxy and set `SUBCONV_PUBLIC_BASE_URL` to the public HTTPS origin. Do not publish it without firewall, TLS, and access controls appropriate for your environment.

## Persistence

The runtime data directory is `/data`:

```yaml
volumes:
  - ./config:/config
  - ./data:/data
```

Published subscriptions are stored under:

```text
/data/published/{publish_id}/current.yaml
/data/published/{publish_id}/meta.json
```

Keep `./data` mounted. Without this volume, published links and workspace state are lost on container removal.

To verify persistence, generate a subscription link, restart the container, then request the same link again:

```sh
docker compose restart subconv-next
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

The link should still return `200` as long as `./data` is mounted.

For backup:

```sh
tar -czf subconv-next-data-backup.tgz ./data
```

For restore, stop the container, restore `./data`, then start it again:

```sh
docker compose down
tar -xzf subconv-next-data-backup.tgz
docker compose up -d
```

## Environment Variables

Docker supports these environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `SUBCONV_HOST` | `0.0.0.0` | Listen address inside the container. |
| `SUBCONV_PORT` | `9876` | Listen port and healthcheck port. |
| `SUBCONV_DATA_DIR` | `/data` | Runtime state, cache, logs, and published subscriptions. |
| `SUBCONV_PUBLIC_BASE_URL` | empty | Public origin used in generated subscription links. |
| `SUBCONV_LOG_LEVEL` | `info` | Service and render log level. |

Example:

```sh
SUBCONV_PUBLIC_BASE_URL=https://subconv.example.com docker compose up -d
```

`SUBCONV_PUBLIC_BASE_URL` only changes generated subscription links returned by the API. It does not configure TLS or reverse proxy behavior.

## Runtime Flags

The same values can be passed as explicit flags. Flags take precedence over environment variables:

```sh
subconv-next serve \
  --config /config/config.json \
  --host 0.0.0.0 \
  --port 9876 \
  --data-dir /data \
  --public-base-url https://subconv.example.com \
  --log-level info
```

## Health Check

```sh
curl -fsS http://127.0.0.1:9876/healthz
```

Expected shape:

```json
{"ok":true,"version":"...","data_dir":"/data","uptime_seconds":1}
```

The health response does not include subscription URLs, published tokens, upstream URLs, or node secrets.

## Updating

```sh
docker compose pull
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

If using only local builds, build and run a local image separately:

```sh
docker build -t subconv-next:local .
docker run --rm -p 9876:9876 -v "$PWD/config:/config" -v "$PWD/data:/data" subconv-next:local
```

Keep `./data` mounted during updates.

## Log Redaction Check

Check logs with:

```sh
docker logs subconv-next
```

Logs must not contain complete upstream subscription tokens, full `/s/{token}/mihomo.yaml` links, node passwords, UUIDs, private keys, pre-shared keys, `Authorization`, or `Cookie` values. Published subscriptions should appear with `token_hint` or redacted paths such as `/s/<redacted>/mihomo.yaml`.

## Multi-Arch Build

Prepare buildx once:

```sh
docker buildx create --use --name subconv-next-builder
```

Build amd64 and arm64:

```sh
docker buildx build --platform linux/amd64,linux/arm64 -t subconv-next:local .
```

For registry release:

```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/earl9/subconv-next:v1.0.0 \
  -t ghcr.io/earl9/subconv-next:latest \
  --push \
  .
```

Verify the pushed manifest:

```sh
docker buildx imagetools inspect ghcr.io/earl9/subconv-next:v1.0.0
docker buildx imagetools inspect ghcr.io/earl9/subconv-next:latest
```

The Dockerfile uses `TARGETOS` and `TARGETARCH`, so arm64 builds do not require code changes.

## Subscription Header Verification

After generating a published link, verify final headers with:

```sh
curl -D - -o /tmp/mihomo.yaml "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

When upstream subscriptions provided traffic metadata, both commands should include:

```text
Subscription-Userinfo: upload=...; download=...; total=...; expire=...
Profile-Update-Interval: 24
Content-Disposition: attachment; filename="mihomo.yaml"
```
