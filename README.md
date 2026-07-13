# SubConv Next

[English](README.md) | [简体中文](README.zh-CN.md)

SubConv Next is a self-hosted subscription converter for Mihomo / Clash Meta. A single Go binary provides the conversion API and Web UI for aggregating upstream subscriptions, editing nodes, managing routing rules, generating validated Mihomo YAML, and publishing private subscription links.

It supports Docker deployments and a native OpenWrt integration built around procd, UCI, rpcd, ACLs, and LuCI JavaScript views.

## Highlights

- Aggregate multiple subscription sources with stable source IDs, labels, and optional emoji prefixes.
- Parse Base64 subscriptions, Clash/Mihomo YAML, and individual node URIs.
- Edit, disable, delete, restore, bulk rename, and add manual nodes.
- Configure built-in routing presets, custom rules, remote rule providers, and custom policy groups.
- Generate optional country and country auto-test groups.
- Preserve upstream `Subscription-Userinfo` metadata for traffic and expiration reporting.
- Publish unguessable `/s/{token}/mihomo.yaml` subscription URLs.
- Keep browser-local drafts without storing complete publication tokens.
- Validate proxy references, rule providers, policy groups, filtered nodes, and final `MATCH` order before writing output.
- Run on `linux/amd64` and `linux/arm64`, with persistent `/config` and `/data` volumes.
- Manage OpenWrt service status, settings, backups, restores, and logs from LuCI.

Supported protocols include `ss`, `ssr`, `vmess`, `vless`, `trojan`, `hysteria2`, `tuic`, `anytls`, `wireguard`, and `mieru`.

## Screenshots

### Web UI

![SubConv Next Web UI](docs/assets/screenshot-main.png)

<details>
<summary>Mobile Web UI</summary>

![SubConv Next mobile Web UI](docs/assets/screenshot-mobile.png)

</details>

### OpenWrt LuCI

![SubConv Next OpenWrt LuCI overview](docs/assets/screenshot-luci.png)

<details>
<summary>Mobile LuCI</summary>

![SubConv Next mobile LuCI overview](docs/assets/screenshot-luci-mobile.png)

</details>

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
      - ./config:/config
      - ./data:/data
    environment:
      SUBCONV_HOST: 0.0.0.0
      SUBCONV_PORT: 9876
      SUBCONV_DATA_DIR: /data
      SUBCONV_LOG_LEVEL: info
```

Start the service and check its health:

```sh
mkdir -p config data
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

Open <http://127.0.0.1:9876/>.

Use `ghcr.io/earl9/subconv-next:latest` to track releases, or pin a release tag for reproducible deployments.

### OpenWrt / Kwrt

Install both the core service and LuCI package:

```sh
opkg install \
  /tmp/subconv-next_<version>_aarch64_generic.ipk \
  /tmp/luci-app-subconv-next_<version>_all.ipk

ubus call luci.subconv status
curl -fsS http://127.0.0.1:9876/healthz
```

The `aarch64_generic` package has been verified on Kwrt 25.12.2 `rockchip/armv8`. Check `opkg print-architecture` before selecting an asset for another device.

Important paths:

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
/usr/libexec/rpcd/luci.subconv
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/overview.js
```

When UCI option `enabled` is `1`, installation enables and starts the service. The LuCI entry is **Services > SubConv Next**.

## Downloads

Release artifacts are published through [GitHub Releases](https://github.com/Earl9/subconv-next/releases) and are not committed to the repository. Assets include Linux binaries, OpenWrt IPKs, the LuCI package, and `checksums.txt`.

## Security

The standalone Web UI has no built-in account system. Run it on localhost or a trusted LAN, or place it behind HTTPS and an authenticated reverse proxy, VPN, or equivalent access control. LuCI access uses the router's existing authentication and rpcd ACL model.

Published subscription URLs are bearer links. Anyone holding a valid `/s/{token}/mihomo.yaml` URL can retrieve its generated configuration. Rotate the link from the Web UI if it is exposed.

Logs and API responses redact known secrets, including full tokens, upstream URL credentials, passwords, UUIDs, private keys, pre-shared keys, `Authorization`, and `Cookie` values. See [SECURITY.md](SECURITY.md) for the security model and reporting process.

## Documentation

- [Docker deployment](docs/docker.md)
- [Configuration](docs/configuration.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Release checklist](docs/release-checklist.md)
- [OpenWrt build and packaging](docs/openwrt-build.md)
- [Security details](docs/security.md)
- [OpenWrt package notes](docs/03-openwrt-package.md)
- [LuCI application notes](docs/10-luci-app.md)

## Development

Go 1.22 or newer is required.

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

Build a local binary:

```sh
go build -o subconv-next ./cmd/subconv-next
```

## License

SubConv Next is released under the [MIT License](LICENSE).
