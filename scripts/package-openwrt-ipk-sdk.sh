#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

SDK_DIR="${SDK_DIR:-/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64}"
VERSION="${VERSION:-1.0.0}"
RELEASE="${RELEASE:-3}"
ARCH="${ARCH:-aarch64_generic}"
PKG_NAME="${PKG_NAME:-subconv-next}"
MAINTAINER="${MAINTAINER:-SubConv Next Maintainers}"
DIST="${DIST:-$ROOT_DIR/dist}"
BIN_PATH="${1:-${SUBCONV_NEXT_BIN:-$ROOT_DIR/dist/openwrt-arm64/subconv-next}}"
USE_PROVIDED_BIN=0
if [ "$#" -gt 0 ] || [ -n "${SUBCONV_NEXT_BIN:-}" ]; then
	USE_PROVIDED_BIN=1
fi

PKG_VERSION="${VERSION}-${RELEASE}"
IPKG_BUILD="$SDK_DIR/scripts/ipkg-build"
IPK_PATH="$DIST/${PKG_NAME}_${PKG_VERSION}_${ARCH}.ipk"
INIT_FILE="$ROOT_DIR/openwrt/subconv-next/files/subconv-next.init"
CONFIG_FILE="$ROOT_DIR/openwrt/subconv-next/files/subconv-next.config"
LUCI_ROOT="$ROOT_DIR/openwrt/luci-app-subconv-next"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

need_cmd grep
need_cmd gzip
need_cmd head
need_cmd ls
need_cmd mkdir
need_cmd od
need_cmd sha256sum
need_cmd stat
need_cmd tar
need_cmd tr

if [ ! -x "$IPKG_BUILD" ]; then
	echo "missing SDK ipkg-build: $IPKG_BUILD" >&2
	exit 1
fi

if [ ! -f "$INIT_FILE" ]; then
	echo "missing init script: $INIT_FILE" >&2
	exit 1
fi

if [ ! -f "$CONFIG_FILE" ]; then
	echo "missing UCI config: $CONFIG_FILE" >&2
	exit 1
fi

for path in \
	"$LUCI_ROOT/root/usr/share/luci/menu.d/luci-app-subconv-next.json" \
	"$LUCI_ROOT/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json" \
	"$LUCI_ROOT/htdocs/luci-static/resources/view/subconv-next/index.js"; do
	if [ ! -f "$path" ]; then
		echo "missing LuCI file: $path" >&2
		exit 1
	fi
done

if [ "$USE_PROVIDED_BIN" -eq 0 ]; then
	need_cmd go
	echo "building linux/arm64 binary: $BIN_PATH"
	mkdir -p "$(dirname -- "$BIN_PATH")"
	(
		cd "$ROOT_DIR"
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
			go build -trimpath -ldflags="-s -w -X main.version=$PKG_VERSION" \
			-o "$BIN_PATH" ./cmd/subconv-next
	)
elif [ ! -x "$BIN_PATH" ]; then
	echo "provided binary is not executable: $BIN_PATH" >&2
	exit 1
fi

if [ ! -x "$BIN_PATH" ]; then
	echo "binary is not executable: $BIN_PATH" >&2
	exit 1
fi

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/subconv-next-ipkg-build.XXXXXX")"
trap 'rm -rf "$WORK_DIR"' EXIT INT TERM

PKGROOT="$WORK_DIR/pkg"
mkdir -p \
	"$PKGROOT/CONTROL" \
	"$PKGROOT/usr/bin" \
	"$PKGROOT/etc/init.d" \
	"$PKGROOT/etc/config" \
	"$PKGROOT/etc/subconv-next/data" \
	"$PKGROOT/usr/share/luci/menu.d" \
	"$PKGROOT/usr/share/rpcd/acl.d" \
	"$PKGROOT/www/luci-static/resources/view/subconv-next"

install -m0755 "$BIN_PATH" "$PKGROOT/usr/bin/subconv-next"
install -m0755 "$INIT_FILE" "$PKGROOT/etc/init.d/subconv-next"
install -m0644 "$CONFIG_FILE" "$PKGROOT/etc/config/subconv-next"
install -m0644 "$LUCI_ROOT/root/usr/share/luci/menu.d/luci-app-subconv-next.json" "$PKGROOT/usr/share/luci/menu.d/luci-app-subconv-next.json"
install -m0644 "$LUCI_ROOT/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json" "$PKGROOT/usr/share/rpcd/acl.d/luci-app-subconv-next.json"
install -m0644 "$LUCI_ROOT/htdocs/luci-static/resources/view/subconv-next/index.js" "$PKGROOT/www/luci-static/resources/view/subconv-next/index.js"
chmod 0755 "$PKGROOT/etc/subconv-next/data"

cat >"$PKGROOT/CONTROL/control" <<EOF_CONTROL
Package: $PKG_NAME
Version: $PKG_VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Section: net
Priority: optional
Depends: ca-bundle, luci-base, rpcd, uci
Description: Modern subscription converter for Mihomo / Clash Meta.
EOF_CONTROL

cat >"$PKGROOT/CONTROL/conffiles" <<'EOF_CONFFILES'
/etc/config/subconv-next
EOF_CONFFILES

cat >"$PKGROOT/CONTROL/postinst" <<'EOF_POSTINST'
#!/bin/sh

[ -n "$IPKG_INSTROOT" ] && exit 0

if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next enable || true

	enabled="$(uci -q get subconv-next.main.enabled)"
	[ -z "$enabled" ] && enabled="1"

	if [ "$enabled" = "1" ]; then
		/etc/init.d/subconv-next start || true
	fi
fi

rm -f /tmp/luci-indexcache
rm -rf /tmp/luci-modulecache

/etc/init.d/rpcd restart >/dev/null 2>&1 || true
/etc/init.d/uhttpd restart >/dev/null 2>&1 || true

exit 0
EOF_POSTINST

cat >"$PKGROOT/CONTROL/prerm" <<'EOF_PRERM'
#!/bin/sh

[ -n "$IPKG_INSTROOT" ] && exit 0

if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next stop || true
	/etc/init.d/subconv-next disable || true
fi

exit 0
EOF_PRERM

chmod 0644 "$PKGROOT/CONTROL/control" "$PKGROOT/CONTROL/conffiles"
chmod 0755 "$PKGROOT/CONTROL/postinst" "$PKGROOT/CONTROL/prerm"

if [ "$(id -u)" = "0" ]; then
	chown -R 0:0 "$PKGROOT"
else
	echo "warning: not running as root; generated package ownership depends on SDK ipkg-build support" >&2
fi

actual_mode() {
	stat -c '%a' "$PKGROOT/$1"
}

check_mode() {
	actual="$(actual_mode "$1")"
	if [ "$actual" != "$2" ]; then
		echo "invalid mode for $1: $actual, want $2" >&2
		exit 1
	fi
}

check_mode usr/bin/subconv-next 755
check_mode etc/init.d/subconv-next 755
check_mode etc/config/subconv-next 644
check_mode etc/subconv-next/data 755
check_mode usr/share/luci/menu.d/luci-app-subconv-next.json 644
check_mode usr/share/rpcd/acl.d/luci-app-subconv-next.json 644
check_mode www/luci-static/resources/view/subconv-next/index.js 644

mkdir -p "$DIST"
rm -f "$IPK_PATH"

if "$IPKG_BUILD" -h 2>&1 | grep -q -- ' -o '; then
	"$IPKG_BUILD" -o 0 -g 0 "$PKGROOT" "$DIST"
else
	echo "notice: $IPKG_BUILD does not support -o/-g; using SDK-compatible invocation" >&2
	"$IPKG_BUILD" "$PKGROOT" "$DIST"
fi

if [ ! -f "$IPK_PATH" ]; then
	echo "expected output not found: $IPK_PATH" >&2
	find "$DIST" -maxdepth 1 -name "${PKG_NAME}_*_${ARCH}.ipk" -print >&2
	exit 1
fi

echo "== package"
ls -lh "$IPK_PATH"
sha256sum "$IPK_PATH"

echo "== ar t $IPK_PATH"
if command -v ar >/dev/null 2>&1 && ar t "$IPK_PATH" 2>/dev/null; then
	:
else
	echo "ar t not applicable; SDK ipkg-build output is not an ar archive"
fi

echo "== outer tar"
tar -tzf "$IPK_PATH"

CHECK_DIR="$WORK_DIR/check"
mkdir -p "$CHECK_DIR"
tar -xzf "$IPK_PATH" -C "$CHECK_DIR"

debian_binary_hex="$(od -An -tx1 "$CHECK_DIR/debian-binary" | tr -d ' \n')"
if [ "$debian_binary_hex" != "322e300a" ]; then
	echo "invalid debian-binary content" >&2
	exit 1
fi

echo "== debian-binary"
cat "$CHECK_DIR/debian-binary"

echo "== control.tar.gz"
tar -tzf "$CHECK_DIR/control.tar.gz"

echo "== control"
tar -xOzf "$CHECK_DIR/control.tar.gz" ./control

if ! tar -tzf "$CHECK_DIR/control.tar.gz" | grep -Fx './postinst' >/dev/null; then
	echo "missing control script: ./postinst" >&2
	exit 1
fi
if ! tar -tzf "$CHECK_DIR/control.tar.gz" | grep -Fx './prerm' >/dev/null; then
	echo "missing control script: ./prerm" >&2
	exit 1
fi

CONTROL_CHECK_DIR="$CHECK_DIR/control"
mkdir -p "$CONTROL_CHECK_DIR"
tar -xzf "$CHECK_DIR/control.tar.gz" -C "$CONTROL_CHECK_DIR"

for script in postinst prerm; do
	actual="$(stat -c '%a' "$CONTROL_CHECK_DIR/$script")"
	if [ "$actual" != "755" ]; then
		echo "invalid control script mode for $script: $actual, want 755" >&2
		exit 1
	fi
done

echo "== data.tar.gz | head -50"
tar -tzf "$CHECK_DIR/data.tar.gz" | head -50

DATA_CHECK_DIR="$CHECK_DIR/data"
mkdir -p "$DATA_CHECK_DIR"
tar -xzf "$CHECK_DIR/data.tar.gz" -C "$DATA_CHECK_DIR"

check_data_mode() {
	actual="$(stat -c '%a' "$DATA_CHECK_DIR/$1")"
	if [ "$actual" != "$2" ]; then
		echo "invalid packaged mode for $1: $actual, want $2" >&2
		exit 1
	fi
}

check_data_mode ./usr/bin/subconv-next 755
check_data_mode ./etc/init.d/subconv-next 755
check_data_mode ./etc/config/subconv-next 644
check_data_mode ./etc/subconv-next/data 755
check_data_mode ./usr/share/luci/menu.d/luci-app-subconv-next.json 644
check_data_mode ./usr/share/rpcd/acl.d/luci-app-subconv-next.json 644
check_data_mode ./www/luci-static/resources/view/subconv-next/index.js 644

echo "$IPK_PATH"
