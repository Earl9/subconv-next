# SubConv Next

[English](README.md) | [简体中文](docs/README.zh-CN.md)

SubConv Next is a self-hosted subscription converter for Mihomo / Clash Meta. A single Go binary provides the conversion API and Web UI for aggregating upstream subscriptions, editing nodes, managing routing rules, generating validated Mihomo YAML, and publishing private subscription links.

Docker is the recommended deployment path. The project also includes a native OpenWrt integration built around procd, UCI, rpcd, ACLs, and LuCI JavaScript views.

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

The repository includes a hardened [docker-compose.yml](docker-compose.yml). It initializes `/data` ownership for the unprivileged container user, mounts `/config` read-only, and publishes the web port on loopback by default.

From the repository root, start the service and check its health:

```sh
mkdir -p config data
export SUBCONV_ACCESS_TOKEN="$(openssl rand -hex 32)"
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

Open <http://127.0.0.1:9876/> and sign in with `SUBCONV_ACCESS_TOKEN`. Save the generated token securely and reuse it for updates; do not commit it to the repository.

This requires Docker Engine and the Compose v2 plugin. With Docker Engine only, use the [standalone Docker instructions](docs/docker.md#standalone-docker). For LAN access, set `SUBCONV_HOST_BIND=0.0.0.0` before starting and keep a strong access token. See [Docker deployment](docs/docker.md) for local image builds, updates, persistence, backups, and reverse-proxy settings.

Use `ghcr.io/earl9/subconv-next:latest` to track releases, or pin a release tag for reproducible deployments.

### OpenWrt / Kwrt

Install the all-in-one package, which includes the service and LuCI pages:

```sh
opkg install /tmp/subconv-next_<version>-<release>_aarch64_generic.ipk

ubus call luci.subconv status
curl -fsS http://127.0.0.1:9876/healthz
```

The `aarch64_generic` package has been verified on Kwrt 25.12.2 `rockchip/armv8`. Check `opkg print-architecture` before selecting an asset for another device.

For APK-based official OpenWrt 25.12 firmware, install the matching all-in-one release APK with `apk add --allow-untrusted /tmp/subconv-next-<version>-r<release>.apk`. Do not install an APK on an opkg-based Kwrt device.

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

Release artifacts are published through [GitHub Releases](https://github.com/Earl9/subconv-next/releases) and are not committed to the repository. Assets include Linux binaries, all-in-one OpenWrt IPK and APK packages, and `checksums.txt`.

## Security

The supplied Compose file binds the host port to `127.0.0.1` by default, mounts `/config` read-only, drops Linux capabilities, and uses a read-only container filesystem. For a public converter, enable `SUBCONV_PUBLIC_CONVERTER=true`, keep a strong `SUBCONV_ACCESS_TOKEN` for protected routes, terminate HTTPS at a reverse proxy, keep port `9876` private, and add distributed edge rate limits. Anonymous visitors receive independent random workspaces; approved converter routes are public while routes outside that allowlist still require the management token. Anonymous workspaces cannot disable TLS certificate verification for upstream fetches, and workspace, publication-store, and refresh-concurrency caps are tunable via `SUBCONV_MAX_WORKSPACES`, `SUBCONV_MAX_PUBLICATIONS`, and `SUBCONV_MAX_CONCURRENT_REFRESHES`. `SUBCONV_ALLOW_INSECURE_PUBLIC` disables the entire management boundary and is not suitable for production.

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
