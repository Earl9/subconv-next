#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

SDK_DIR="${SDK_DIR:-/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64}"
VERSION="${VERSION:-1.0.0}"
RELEASE="${RELEASE:-1}"
ARCH="${ARCH:-all}"
PKG_NAME="${PKG_NAME:-luci-app-subconv-next}"
MAINTAINER="${MAINTAINER:-SubConv Next Maintainers}"
DIST="${DIST:-$ROOT_DIR/dist}"

PKG_VERSION="${VERSION}-${RELEASE}"
IPKG_BUILD="$SDK_DIR/scripts/ipkg-build"
IPK_PATH="$DIST/${PKG_NAME}_${PKG_VERSION}_${ARCH}.ipk"
LUCI_ROOT="$ROOT_DIR/openwrt/luci-app-subconv-next"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

need_cmd grep
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

for path in \
	"$LUCI_ROOT/root/usr/share/luci/menu.d/luci-app-subconv-next.json" \
	"$LUCI_ROOT/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json" \
	"$LUCI_ROOT/htdocs/luci-static/resources/view/subconv-next/index.js"; do
	if [ ! -f "$path" ]; then
		echo "missing LuCI file: $path" >&2
		exit 1
	fi
done

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/luci-app-subconv-next-ipkg-build.XXXXXX")"
trap 'rm -rf "$WORK_DIR"' EXIT INT TERM

PKGROOT="$WORK_DIR/pkg"
mkdir -p \
	"$PKGROOT/CONTROL" \
	"$PKGROOT/usr/share/luci/menu.d" \
	"$PKGROOT/usr/share/rpcd/acl.d" \
	"$PKGROOT/www/luci-static/resources/view/subconv-next"

install -m0644 "$LUCI_ROOT/root/usr/share/luci/menu.d/luci-app-subconv-next.json" "$PKGROOT/usr/share/luci/menu.d/luci-app-subconv-next.json"
install -m0644 "$LUCI_ROOT/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json" "$PKGROOT/usr/share/rpcd/acl.d/luci-app-subconv-next.json"
install -m0644 "$LUCI_ROOT/htdocs/luci-static/resources/view/subconv-next/index.js" "$PKGROOT/www/luci-static/resources/view/subconv-next/index.js"

cat >"$PKGROOT/CONTROL/control" <<EOF_CONTROL
Package: $PKG_NAME
Version: $PKG_VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Section: luci
Priority: optional
Depends: luci-base, subconv-next
Description: LuCI support for SubConv Next.
EOF_CONTROL

chmod 0644 "$PKGROOT/CONTROL/control"

if [ "$(id -u)" = "0" ]; then
	chown -R 0:0 "$PKGROOT"
else
	echo "warning: not running as root; generated package ownership depends on SDK ipkg-build support" >&2
fi

for path in \
	usr/share/luci/menu.d/luci-app-subconv-next.json \
	usr/share/rpcd/acl.d/luci-app-subconv-next.json \
	www/luci-static/resources/view/subconv-next/index.js; do
	actual="$(stat -c '%a' "$PKGROOT/$path")"
	if [ "$actual" != "644" ]; then
		echo "invalid mode for $path: $actual, want 644" >&2
		exit 1
	fi
done

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

echo "== control"
tar -xOzf "$CHECK_DIR/control.tar.gz" ./control

echo "== data.tar.gz | head -50"
tar -tzf "$CHECK_DIR/data.tar.gz" | head -50

DATA_CHECK_DIR="$CHECK_DIR/data"
mkdir -p "$DATA_CHECK_DIR"
tar -xzf "$CHECK_DIR/data.tar.gz" -C "$DATA_CHECK_DIR"

for path in \
	./usr/share/luci/menu.d/luci-app-subconv-next.json \
	./usr/share/rpcd/acl.d/luci-app-subconv-next.json \
	./www/luci-static/resources/view/subconv-next/index.js; do
	if [ ! -f "$DATA_CHECK_DIR/$path" ]; then
		echo "missing packaged file: $path" >&2
		exit 1
	fi
	actual="$(stat -c '%a' "$DATA_CHECK_DIR/$path")"
	if [ "$actual" != "644" ]; then
		echo "invalid packaged mode for $path: $actual, want 644" >&2
		exit 1
	fi
done

echo "$IPK_PATH"
