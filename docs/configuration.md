# Configuration

This document covers the V1 runtime configuration surface for Docker and binary deployments.

## Priority

Configuration priority is:

```text
CLI flags > environment variables > config file defaults
```

Use CLI flags for one-off binary runs. Use environment variables for Docker and process managers.

## CLI

Start the HTTP service:

```sh
subconv-next serve \
  --config /config/config.json \
  --host 0.0.0.0 \
  --port 9876 \
  --data-dir /data \
  --public-base-url https://subconv.example.com \
  --log-level info
```

Supported service flags:

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `config/config.json` | JSON or UCI config file path. |
| `--host` | config value | Listen address override. |
| `--port` | config value | Listen port override. |
| `--data-dir` | config paths | Runtime data root. Must be an absolute path. |
| `--public-base-url` | config value | Public origin used when generating subscription links. |
| `--log-level` | config value | Service and rendered YAML log level. |

Other CLI commands:

```sh
subconv-next version
subconv-next parse
subconv-next generate
```

## Environment Variables

| Variable | Default in Docker | Description |
| --- | --- | --- |
| `SUBCONV_HOST` | `0.0.0.0` | Listen address inside the container or process. |
| `SUBCONV_PORT` | `9876` | Listen port. |
| `SUBCONV_DATA_DIR` | `/data` | Runtime data directory. |
| `SUBCONV_PUBLIC_BASE_URL` | empty | Public origin used in generated subscription links. |
| `SUBCONV_LOG_LEVEL` | `info` | Service and renderer log level. |

Example:

```sh
SUBCONV_HOST=0.0.0.0 \
SUBCONV_PORT=9876 \
SUBCONV_DATA_DIR=/data \
SUBCONV_PUBLIC_BASE_URL=https://subconv.example.com \
SUBCONV_LOG_LEVEL=info \
subconv-next serve --config /config/config.json
```

## Data Directory

`SUBCONV_DATA_DIR` defaults to `/data` in Docker. Docker deployments should mount:

```yaml
volumes:
  - ./data:/data
```

Runtime files include:

```text
/data/state.json
/data/cache/
/data/workspaces/
/data/published/
/data/logs/
```

Published subscription links depend on:

```text
/data/published/{publish_id}/current.yaml
/data/published/{publish_id}/meta.json
```

If `/data` is not mounted, links and workspace state can be lost when the container is removed.

## Public Base URL

`SUBCONV_PUBLIC_BASE_URL` controls the origin returned by publish APIs.

Use it when SubConv Next is behind a reverse proxy:

```sh
SUBCONV_PUBLIC_BASE_URL=https://subconv.example.com
```

This value only changes generated URLs. It does not configure TLS, reverse proxy behavior, authentication, or firewall rules.

## Docker Example

```yaml
services:
  subconv-next:
    image: ghcr.io/OWNER/subconv-next:latest
    container_name: subconv-next
    restart: unless-stopped
    ports:
      - "9876:9876"
    volumes:
      - ./data:/data
    environment:
      SUBCONV_HOST: 0.0.0.0
      SUBCONV_PORT: 9876
      SUBCONV_DATA_DIR: /data
      SUBCONV_PUBLIC_BASE_URL: ""
      SUBCONV_LOG_LEVEL: info
```

See [docker.md](docker.md) for Docker-specific deployment, health check, backup, and multi-arch build steps.

## DNS Defaults

The default rendered DNS block is intentionally scoped for OpenClash compatibility:

```yaml
dns:
  enable: true
  listen: 127.0.0.1:5335
  enhanced-mode: fake-ip
  default-nameserver: [119.29.29.29, 223.5.5.5]
  nameserver-policy:
    '*.linux.do': https://xxx.ddd.oaifree.com/query-dns
    geosite:cn,private,apple:
      - https://doh.pub/dns-query
      - https://dns.alidns.com/dns-query
    linux.do: https://xxx.ddd.oaifree.com/query-dns
  nameserver: ['https://1.1.1.1/dns-query#RULES', 'https://8.8.8.8/dns-query#RULES']
  proxy-server-nameserver: [119.29.29.29, 223.5.5.5]
  direct-nameserver: ['https://doh.pub/dns-query', 'https://dns.alidns.com/dns-query']
  direct-nameserver-follow-policy: true
  fake-ip-range: 198.18.0.0/16
  fake-ip-filter: ['*.lan', '*.local', '*.arpa', time.*.com, ntp.*.com, +.market.xiaomi.com, localhost.ptlogin2.qq.com, '*.msftncsi.com', www.msftconnecttest.com]
```

SubConv Next does not emit `fallback`, `fallback-filter`, or DoT servers by default. `default-nameserver` and `proxy-server-nameserver` use domestic plain DNS for bootstrap and node domains, while the global `nameserver` entries include `#RULES`.
