# OpenWrt Build

This document describes the V1 OpenWrt/Kwrt package for SubConv Next.

The default release artifact is an all-in-one package containing the backend binary, `init.d`, UCI config, `procd` service, persistent data directory, and the LuCI management shell.

## Target

Initial target:

```text
OpenWrt target: rockchip/armv8
CPU architecture: arm64 / aarch64
Package architecture: aarch64_generic
Binary: /usr/bin/subconv-next
Config: /etc/config/subconv-next
Service: /etc/init.d/subconv-next
Data: /etc/subconv-next/data
Process manager: procd
LuCI menu: Services / SubConv Next
```

The Web UI is embedded into the Go binary through `go:embed` in `internal/api/static/assets.go`. The router does not need Node.js, npm, or the Go toolchain at runtime.

## Package Format

Choose the package format by the target device package manager, not only by the OpenWrt release number.

Current Kwrt 25.12.2 `rockchip/armv8` devices observed with:

```text
DISTRIB_ID='Kwrt'
DISTRIB_RELEASE='25.12.2'
DISTRIB_TARGET='rockchip/armv8'
DISTRIB_ARCH='aarch64_generic'
opkg print-architecture: arch aarch64_generic 10
```

must install an IPK:

```sh
opkg install /tmp/subconv-next_*.ipk
```

Official OpenWrt 25.12 SDKs can enable `CONFIG_USE_APK=y` and produce APK packages such as `subconv-next-1.0.0-r1.apk`. That APK is valid only for APK-based firmware and is not the final target for Kwrt/opkg devices.

## Runtime Layout

Installed files:

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/index.js
```

Runtime data under `/etc/subconv-next/data` includes:

```text
state.json
cache/
published/
workspaces/
logs/
```

Back up `/etc/subconv-next/data` if published subscription links must survive router reinstall or flash reset.

## CLI Contract

The binary supports both explicit `serve` and root-level service flags:

```sh
subconv-next serve \
  --config /etc/subconv-next/config.json \
  --host 0.0.0.0 \
  --port 9876 \
  --data-dir /etc/subconv-next/data \
  --public-base-url '' \
  --log-level info
```

The init script reads `/etc/config/subconv-next`, creates the data directory, and starts the service through `procd`. It passes `/etc/subconv-next/config.json` as the application config path. If that file does not exist, the binary starts from built-in defaults and CLI overrides from UCI.

## Package Files

Package source files live in this repository:

```text
openwrt/subconv-next/Makefile
openwrt/subconv-next/files/subconv-next.init
openwrt/subconv-next/files/subconv-next.config
openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/index.js
```

The Makefile builds from a local checkout through `SUBCONV_NEXT_SRC`. It uses the OpenWrt SDK and `golang-package.mk`, but does not fetch source from a remote Git repository.

Go module path:

```text
subconv-next
```

Main package path:

```text
./cmd/subconv-next
```

## Build Requirements

Use the SDK matching the exact firmware release, target, and package manager. For Kwrt/opkg devices, prefer a Kwrt SDK for `25.12.2 rockchip/armv8` that outputs IPK packages for `aarch64_generic`.

For official OpenWrt rockchip/armv8, SDK names look like:

```text
openwrt-sdk-*-rockchip-armv8-*.tar.zst
```

Official OpenWrt 25.12 SDKs may output APK by default because `CONFIG_USE_APK=y`.

Build as a normal user. Do not build the OpenWrt SDK as root.

Install host tools required by your distribution for the OpenWrt SDK, including `zstd`, `tar`, `make`, `gcc`, `g++`, `rsync`, `python3`, and `git`.

## Kwrt/opkg SDK Build Flow

If a Kwrt 25.12.2 `rockchip/armv8` SDK is available, build from the SDK directory:

```sh
./scripts/feeds update -a
./scripts/feeds install -a
mkdir -p package/subconv-next
cp -r /path/to/subconv-next/openwrt/subconv-next/* package/subconv-next/
make defconfig
make package/subconv-next/compile V=s SUBCONV_NEXT_SRC=/path/to/subconv-next
find bin -name "subconv-next*.ipk"
```

If the package directory is symlinked instead of copied, `SUBCONV_NEXT_SRC` should still point to the repository root containing `cmd/subconv-next`.

Expected IPK output path depends on SDK feed layout, for example:

```text
bin/packages/aarch64_generic/base/subconv-next_1.0.0-1_aarch64_generic.ipk
```

## Official OpenWrt 25.12 SDK Notes

The official `openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64` SDK enables `CONFIG_USE_APK=y` by default and outputs APK:

```text
bin/packages/aarch64_generic/base/subconv-next-1.0.0-r1.apk
```

This APK should be documented as the official OpenWrt 25.12 SDK artifact only. Do not use it as the target package for Kwrt/opkg devices.

Attempting to force IPK in that official SDK:

```sh
grep -n "USE_APK" .config
./scripts/config -d USE_APK
make defconfig
grep -n "USE_APK" .config
```

may re-enable `CONFIG_USE_APK=y` during `defconfig`.

Validated result for the official SDK above:

- `make defconfig` re-enabled `CONFIG_USE_APK=y`.
- A compile after attempting to disable APK still invoked `apk mkpkg`.
- The artifact was `bin/packages/aarch64_generic/base/subconv-next-1.0.0-r1.apk`.
- No `subconv-next_*.ipk` was produced by that official SDK.

For Kwrt/opkg devices, use a Kwrt/IPK SDK or the SDK `ipkg-build` packaging script below.

## SDK ipkg-build Packaging

If a Kwrt/IPK SDK is not available, build an arm64 static binary and use the SDK-provided `scripts/ipkg-build` to wrap it as an opkg-compatible IPK:

```sh
SDK_DIR=/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64 \
VERSION=1.0.0 \
RELEASE=3 \
ARCH=aarch64_generic \
./scripts/package-openwrt-ipk-sdk.sh
```

Default output:

```text
dist/subconv-next_1.0.0-3_aarch64_generic.ipk
```

The script defaults to:

- package name: `subconv-next`
- version: `1.0.0-3`
- architecture: `aarch64_generic`
- binary input/output: `dist/openwrt-arm64/subconv-next`
- SDK dir: `/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64`
- version: `1.0.0`
- release: `3`
- architecture: `aarch64_generic`
- binary input/output: `dist/openwrt-arm64/subconv-next`
- IPK output: `dist/subconv-next_1.0.0-3_aarch64_generic.ipk`

It installs:

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/index.js
```

The control metadata is:

```text
Package: subconv-next
Version: 1.0.0-3
Architecture: aarch64_generic
Maintainer: SubConv Next Maintainers
Section: net
Priority: optional
Depends: ca-bundle, luci-base, rpcd, uci
Description: Modern subscription converter for Mihomo / Clash Meta.
```

The main package includes `postinst` and `prerm` scripts:

- `postinst` enables `/etc/init.d/subconv-next` after install, starts it when `subconv-next.main.enabled` is unset or `1`, clears LuCI caches, and restarts `rpcd`/`uhttpd`.
- `prerm` stops and disables the service before package removal.
- Both scripts exit immediately during image/rootfs builds when `IPKG_INSTROOT` is set.
- Runtime data under `/etc/subconv-next/data` is never removed by package scripts.

Override inputs when needed:

```sh
SDK_DIR=/path/to/openwrt-sdk \
VERSION=1.0.0 \
RELEASE=3 \
ARCH=aarch64_generic \
DIST=dist \
./scripts/package-openwrt-ipk-sdk.sh /path/to/arm64/subconv-next
```

Verify the IPK locally:

```sh
tmp="$(mktemp -d)"
cd "$tmp"
tar -xzf /path/to/subconv-next_1.0.0-3_aarch64_generic.ipk
tar -xOzf control.tar.gz ./control
tar -tzf data.tar.gz
```

Do not rename an APK to IPK. Also do not use the older `scripts/package-openwrt-ipk.sh` hand-built ar path as the default Kwrt package path; that ar package was rejected by Kwrt/opkg as malformed.

## Optional Split LuCI App Packaging

The default release artifact is the all-in-one `subconv-next` package above. The split `luci-app-subconv-next` package is kept only as an advanced packaging option. Normal Kwrt users should not install it separately when using `subconv-next_1.0.0-3_aarch64_generic.ipk`.

Package files:

```text
openwrt/luci-app-subconv-next/Makefile
openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json
openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/index.js
```

Build the LuCI IPK with:

```sh
SDK_DIR=/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64 \
VERSION=1.0.0 \
RELEASE=1 \
ARCH=all \
./scripts/package-openwrt-luci-ipk-sdk.sh
```

Default output:

```text
dist/luci-app-subconv-next_1.0.0-1_all.ipk
```

The LuCI package installs:

```text
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/index.js
```

LuCI menu path:

```text
Services / SubConv Next
```

The page supports:

- viewing Running / Stopped status
- editing `enabled`, `host`, `port`, `data_dir`, `public_base_url`, and `log_level`
- start, stop, and restart actions
- opening the SubConv Next Web UI

## Install on Router

For Kwrt/opkg systems:

```sh
scp dist/subconv-next_1.0.0-3_aarch64_generic.ipk root@router:/tmp/
ssh root@router "opkg install /tmp/subconv-next_1.0.0-3_aarch64_generic.ipk"
```

For IPK produced by a Kwrt SDK:

```sh
scp bin/packages/*/*/subconv-next_*.ipk root@router:/tmp/
ssh root@router "opkg install /tmp/subconv-next_*.ipk"
```

For APK-based official OpenWrt builds:

```sh
scp bin/packages/*/*/subconv-next-*.apk root@router:/tmp/
ssh root@router "apk add --allow-untrusted /tmp/subconv-next-*.apk"
```

Verify installed files:

```sh
ssh root@router "ls -l /usr/bin/subconv-next /etc/init.d/subconv-next /etc/config/subconv-next"
ssh root@router "ls -ld /etc/subconv-next/data"
```

## Start and Enable

The `subconv-next` main package enables the service after install and starts it automatically when `subconv-next.main.enabled` is unset or `1`. Manual `enable` and `start` are no longer required for the default install.

Check status and logs:

```sh
/etc/init.d/subconv-next status
logread -f | grep subconv
curl -fsS http://127.0.0.1:9876/healthz
```

Expected health shape:

```json
{"ok":true,"version":"...","data_dir":"/etc/subconv-next/data","uptime_seconds":1}
```

Open the Web UI from a trusted LAN:

```text
http://router-ip:9876
```

Open LuCI:

```text
Services / SubConv Next
```

If LuCI does not show the new menu immediately, reload rpcd and clear browser cache:

```sh
/etc/init.d/rpcd reload
```

## Configure

Show current config:

```sh
uci show subconv-next
```

Example changes:

```sh
uci set subconv-next.main.enabled='1'
uci set subconv-next.main.host='0.0.0.0'
uci set subconv-next.main.port='9876'
uci set subconv-next.main.data_dir='/etc/subconv-next/data'
uci set subconv-next.main.public_base_url='http://router-ip:9876'
uci set subconv-next.main.log_level='info'
uci commit subconv-next
/etc/init.d/subconv-next restart
```

When changing the port through LuCI, use Save & Apply. The LuCI page restarts `subconv-next` after applying UCI changes.

Default UCI config:

```uci
config subconv-next 'main'
        option enabled '1'
        option host '0.0.0.0'
        option port '9876'
        option data_dir '/etc/subconv-next/data'
        option public_base_url ''
        option log_level 'info'
```

## Subscription Verification

After generating a publish link in the Web UI:

```sh
curl -D - -o /tmp/mihomo.yaml "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

When upstream metadata exists, both GET and HEAD should include:

```text
Subscription-Userinfo: upload=...; download=...; total=...; expire=...
Profile-Update-Interval: 24
Content-Disposition: attachment; filename="mihomo.yaml"
```

Restart and confirm the old link still works:

```sh
/etc/init.d/subconv-next restart
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

Reboot verification:

```sh
reboot
# after reconnect
/etc/init.d/subconv-next status
curl -fsS http://127.0.0.1:9876/healthz
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

The service should auto-start and old published links should remain valid as long as `/etc/subconv-next/data` is preserved.

## Uninstall

Stop, disable, and remove the all-in-one package:

```sh
/etc/init.d/subconv-next stop
/etc/init.d/subconv-next disable
opkg remove subconv-next
```

If you installed the optional split LuCI package in an older test build, remove it separately:

```sh
opkg remove luci-app-subconv-next
```

For APK-based systems:

```sh
apk del subconv-next
```

OpenWrt package removal may leave `/etc/config/subconv-next` because it is a conffile. Runtime data under `/etc/subconv-next/data` is intentionally not removed by the package scripts. Remove it manually only if published links, workspaces, logs, and cache are no longer needed:

```sh
rm -rf /etc/subconv-next/data
```

## Device Acceptance Checklist

- `/usr/bin/subconv-next` exists and is executable.
- `/etc/init.d/subconv-next` exists and is executable.
- `/etc/config/subconv-next` exists.
- `/etc/subconv-next/data` exists.
- `/etc/init.d/subconv-next start` starts the service.
- `/etc/init.d/subconv-next enable` enables boot startup.
- Installing `subconv-next_1.0.0-3_aarch64_generic.ipk` auto-enables and auto-starts the service when enabled is `1`.
- Installing the all-in-one package adds `Services / SubConv Next`.
- The LuCI page can change the port and restart the service.
- `curl http://127.0.0.1:9876/healthz` returns OK.
- Browser access to `http://router-ip:9876` works from trusted LAN.
- Generated `/s/{token}/mihomo.yaml` imports into Mihomo / Clash Meta.
- GET and HEAD subscription responses include metadata headers when upstream metadata exists.
- After reboot, service auto-starts and old published links remain valid.

## V1.1 Limits

- LuCI page is a management shell only; the full application remains the built-in SubConv Next Web UI.
- No automatic reverse proxy or TLS setup.
- No proxy core management.
- No kernel or firewall integration.
- SubConv Next only generates and serves subscription YAML.
