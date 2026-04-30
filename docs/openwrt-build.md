# OpenWrt Build Notes

This is a V1 planning draft. Phase 3 does not implement or modify the OpenWrt package. V1 Docker remains the primary supported deployment target.

The OpenWrt plan is binary plus `init.d` first. LuCI is intentionally out of scope for the initial V1 package.

Target for this release gate:

```text
OpenWrt target: rockchip/armv8
Binary: /usr/bin/subconv-next
Config: /etc/config/subconv-next
Service: /etc/init.d/subconv-next
Data: /etc/subconv-next/data
Manager: procd
```

The Web UI is embedded into the Go binary through `go:embed` in `internal/api/static/assets.go`, so the package does not install separate frontend assets and does not require Node.js on the router.

## Runtime Layout

Default OpenWrt paths:

```text
/usr/bin/subconv-next
/etc/config/subconv-next
/etc/init.d/subconv-next
/etc/subconv-next/data
```

Runtime files under `/etc/subconv-next/data` include:

```text
state.json
cache/
published/
workspaces/
logs/
```

Keep this directory in backups if published subscription links must survive router reinstall or flash reset.

## CLI Contract

The binary supports:

```sh
subconv-next serve \
  --config /etc/config/subconv-next \
  --host 0.0.0.0 \
  --port 9876 \
  --data-dir /etc/subconv-next/data \
  --public-base-url http://router.lan:9876 \
  --log-level info
```

The init script reads `/etc/config/subconv-next` and passes these CLI flags to the binary. CLI flags intentionally override UCI/default config paths so the package can keep all runtime state under `/etc/subconv-next/data`.

## Planned Package Files

The planned package layout is:

```text
openwrt/subconv-next/Makefile
openwrt/subconv-next/files/subconv-next.init
openwrt/subconv-next/files/subconv-next.config
```

Installed files:

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
```

The service should be managed by `procd` and enabled on boot after install. Review config before starting the service on a router.

## Local Binary Build

Build a Linux arm64 binary for rockchip/armv8:

```sh
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o subconv-next ./cmd/subconv-next
file subconv-next
```

Expected:

```text
ELF 64-bit LSB executable, ARM aarch64, statically linked
```

## OpenWrt SDK Build

Use the OpenWrt SDK matching the exact firmware release and target, for example an SDK for `rockchip/armv8`.

Inside the SDK:

```sh
./scripts/feeds update packages
./scripts/feeds install golang
```

Copy or symlink the package directory:

```sh
ln -s /path/to/subconv-next/openwrt/subconv-next package/subconv-next
```

Planned package build command:

```sh
make package/subconv-next/compile V=s SUBCONV_NEXT_SRC=/path/to/subconv-next
```

If the package is copied instead of symlinked, `SUBCONV_NEXT_SRC` is required and must point to the repository root containing `cmd/subconv-next`.

The generated ipk would be under:

```text
bin/packages/*/base/subconv-next_*.ipk
```

## Target Install

Once an ipk is available, copy and install on the router:

```sh
scp bin/packages/*/base/subconv-next_*.ipk root@router:/tmp/
ssh root@router "opkg install /tmp/subconv-next_*.ipk"
```

Review config:

```sh
ssh root@router "cat /etc/config/subconv-next"
```

Common settings:

```sh
uci set subconv-next.service.enabled='1'
uci set subconv-next.service.host='0.0.0.0'
uci set subconv-next.service.port='9876'
uci set subconv-next.service.data_dir='/etc/subconv-next/data'
uci set subconv-next.service.public_base_url='http://router.lan:9876'
uci set subconv-next.service.log_level='info'
uci commit subconv-next
```

Start and enable:

```sh
/etc/init.d/subconv-next enable
/etc/init.d/subconv-next restart
```

Health check:

```sh
curl -fsS http://127.0.0.1:9876/healthz
```

Expected shape:

```json
{"ok":true,"version":"...","uptime_seconds":1}
```

## Manual Install Smoke Test

If testing without ipk packaging:

```sh
scp subconv-next root@router:/usr/bin/subconv-next
scp openwrt/subconv-next/files/subconv-next.init root@router:/etc/init.d/subconv-next
scp openwrt/subconv-next/files/subconv-next.config root@router:/etc/config/subconv-next
ssh root@router "chmod +x /usr/bin/subconv-next /etc/init.d/subconv-next"
ssh root@router "/etc/init.d/subconv-next enable && /etc/init.d/subconv-next start"
ssh root@router "curl -fsS http://127.0.0.1:9876/healthz"
```

## Runtime Verification

On the router:

```sh
/etc/init.d/subconv-next status
logread -e subconv-next
ls -la /etc/subconv-next/data
curl -I http://127.0.0.1:9876/
curl -fsS http://127.0.0.1:9876/healthz
```

After generating a published subscription link, restart and verify the same link still works:

```sh
/etc/init.d/subconv-next restart
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

The link should return `200` as long as `/etc/subconv-next/data` is preserved.

## V1 Limits

- No LuCI page.
- No LuCI implementation work in this phase.
- IPK packaging is planned after Docker release stability is locked.
- No automatic reverse proxy or TLS setup.
- No proxy core management.
- No kernel/network integration; SubConv Next only generates and serves subscription YAML.
