# SubConv Next

## Project Summary

SubConv Next is a modern subscription converter for Mihomo / Clash Meta. It provides a single Go binary with an embedded Web UI for aggregating upstream subscriptions, editing nodes, generating Mihomo YAML, and publishing private subscription links.

The V1 release focuses on practical self-hosting: Docker deployment, OpenWrt / Kwrt all-in-one IPK packaging, traffic metadata headers, safer generated YAML, and local-first workspace behavior.

## Features

- Multi-source subscription aggregation with source names and Emoji labels.
- Node names keep the upstream raw name and add only the `Emoji + sourceName` prefix; duplicate final names get `#2`, `#3`, etc.
- Built-in node editing, disabling, deleting, restoring, batch rename, and manual node support.
- Mihomo / Clash Meta YAML generation with rule groups, full-mode rule policy node lists, and no country / region policy groups.
- Aggregated `Subscription-Userinfo` headers so Clash / Mihomo can show traffic usage and expiry time.
- Private random published subscription links under `/s/{token}/mihomo.yaml`.
- Local browser drafts without storing the full published token.
- Final YAML validation for proxy group references, filtered nodes, auto-select contents, and `MATCH` ordering.
- Docker image with `/data` persistence and `linux/amd64` plus `linux/arm64` support.
- OpenWrt / Kwrt all-in-one IPK with backend service, init.d, UCI config, procd, data directory, LuCI menu, and LuCI management page.

Supported protocols include `ss`, `ssr`, `vmess`, `vless`, `trojan`, `hysteria2`, `tuic`, `anytls`, and `wireguard`. `anytls` and `wireguard` are still treated as experimental in V1.

## Screenshots

Screenshots are not committed to the repository yet. Use the Web UI at `http://127.0.0.1:9876/` after starting the service.

## Quick Start

### Docker Compose

Create `docker-compose.yml`:

```yaml
services:
  subconv-next:
    image: ghcr.io/earl9/subconv-next:latest
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
      SUBCONV_LOG_LEVEL: info
```

Start it:

```sh
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

Open:

```text
http://127.0.0.1:9876/
```

Use `ghcr.io/earl9/subconv-next:v1.0.0` for a pinned release, or `ghcr.io/earl9/subconv-next:latest` for the latest published image.

### OpenWrt / Kwrt

OpenWrt / Kwrt support is provided as an all-in-one IPK for `rockchip/armv8` / `aarch64_generic`.

```sh
opkg install /tmp/subconv-next_1.0.0-3_aarch64_generic.ipk
curl -fsS http://127.0.0.1:9876/healthz
```

Installed paths:

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/index.js
```

The package auto-enables and starts the service when UCI `enabled=1`. LuCI entry: `Services / SubConv Next`.

## Downloads

Release files are published as GitHub Release Assets, not committed to the repository:

- GitHub Releases: <https://github.com/Earl9/subconv-next/releases>
- `subconv-next-linux-amd64`
- `subconv-next-linux-arm64`
- `subconv-next_1.0.0-3_aarch64_generic.ipk`
- `checksums.txt`

## Security

SubConv Next has no built-in login in V1. Run it on localhost or a trusted LAN. If exposing it beyond a trusted network, put it behind HTTPS, authentication, a reverse proxy, VPN, or equivalent access control.

Private subscription URLs are bearer links: anyone holding `/s/{token}/mihomo.yaml` can fetch the generated YAML. Rotate a private link from the Web UI if it leaks.

Logs and APIs are designed to redact full tokens, upstream URL secrets, passwords, UUIDs, private keys, pre-shared keys, `Authorization`, and `Cookie` values. See [SECURITY.md](SECURITY.md) for the security boundary and reporting guidance.

## Documentation

- [Docker deployment](docs/docker.md)
- [Configuration](docs/configuration.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Release checklist](docs/release-checklist.md)
- [OpenWrt / Kwrt build and packaging](docs/openwrt-build.md)
- [Security model details](docs/security.md)
- [OpenWrt package notes](docs/03-openwrt-package.md)
- [LuCI app notes](docs/10-luci-app.md)

## Development

Requires Go 1.22+.

```sh
go test ./...
go test -race ./...
go vet ./...
```

Run locally:

```sh
go run ./cmd/subconv-next serve \
  --host 127.0.0.1 \
  --port 9876 \
  --data-dir "$PWD/data" \
  --log-level info
```

Build a binary:

```sh
go build -o subconv-next ./cmd/subconv-next
```

## License

SubConv Next is released under the MIT License. See [LICENSE](LICENSE).
