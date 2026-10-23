#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

SDK_DIR="${SDK_DIR:-/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64}"
VERSION="${VERSION:-1.0.7}"
RELEASE="${RELEASE:-1}"
ARCH="${ARCH:-aarch64_generic}"
PKG_NAME="${PKG_NAME:-subconv-next}"
DIST="${DIST:-$ROOT_DIR/dist}"
BIN_PATH="${1:-${SUBCONV_NEXT_BIN:-$ROOT_DIR/dist/openwrt-arm64/subconv-next}}"
case "$BIN_PATH" in
	/*) ;;
	*) BIN_PATH="$PWD/$BIN_PATH" ;;
esac
PACKAGE_DIR="$SDK_DIR/package/$PKG_NAME"
BIN_IS_EXPLICIT=0
if [ "$#" -gt 0 ] || [ -n "${SUBCONV_NEXT_BIN:-}" ]; then
	BIN_IS_EXPLICIT=1
fi

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

need_cmd cp
need_cmd dirname
need_cmd find
need_cmd grep
need_cmd ls
need_cmd make
need_cmd mkdir
need_cmd rm
need_cmd sed
need_cmd sha256sum

if [ ! -d "$SDK_DIR" ] || [ ! -f "$SDK_DIR/Makefile" ]; then
	echo "missing OpenWrt SDK: $SDK_DIR" >&2
	exit 1
fi

if [ -e "$PACKAGE_DIR" ]; then
	echo "refusing to overwrite existing SDK package directory: $PACKAGE_DIR" >&2
	exit 1
fi

if [ ! -x "$BIN_PATH" ]; then
	if [ "$BIN_IS_EXPLICIT" -eq 1 ]; then
		echo "arm64 binary is missing or not executable: $BIN_PATH" >&2
		exit 1
	fi
	need_cmd go
	echo "building linux/arm64 binary: $BIN_PATH"
	mkdir -p "$(dirname -- "$BIN_PATH")"
	(
		cd "$ROOT_DIR"
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
			go build -trimpath -ldflags="-s -w -X main.version=$VERSION-$RELEASE" \
			-o "$BIN_PATH" ./cmd/subconv-next
	)
fi

cleanup() {
	rm -rf "$PACKAGE_DIR"
}
trap cleanup EXIT INT TERM

mkdir -p "$SDK_DIR/package"
cp -a "$ROOT_DIR/openwrt/subconv-next" "$PACKAGE_DIR"

(
	cd "$SDK_DIR"
	make "SUBCONV_NEXT_BIN=$BIN_PATH" "SUBCONV_NEXT_SRC=$ROOT_DIR" defconfig
	if grep -q '^# CONFIG_USE_APK is not set$' .config; then
		sed -i 's/^# CONFIG_USE_APK is not set$/CONFIG_USE_APK=y/' .config
	elif ! grep -q '^CONFIG_USE_APK=y$' .config; then
		echo 'CONFIG_USE_APK=y' >> .config
	fi
	make "SUBCONV_NEXT_BIN=$BIN_PATH" "SUBCONV_NEXT_SRC=$ROOT_DIR" defconfig
	if ! grep -q '^CONFIG_USE_APK=y$' .config; then
		echo "OpenWrt SDK did not enable CONFIG_USE_APK" >&2
		exit 1
	fi
	sdk_arch="$(sed -n 's/^CONFIG_TARGET_ARCH_PACKAGES="\([^"]*\)"$/\1/p' .config)"
	if [ "$sdk_arch" != "$ARCH" ]; then
		echo "OpenWrt SDK package architecture is $sdk_arch, expected $ARCH" >&2
		exit 1
	fi
	make "PKG_VERSION=$VERSION" "PKG_RELEASE=$RELEASE" \
		"SUBCONV_NEXT_BIN=$BIN_PATH" \
		"SUBCONV_NEXT_SRC=$ROOT_DIR" \
		package/$PKG_NAME/clean
	make "PKG_VERSION=$VERSION" "PKG_RELEASE=$RELEASE" \
		"SUBCONV_NEXT_BIN=$BIN_PATH" \
		"SUBCONV_NEXT_SRC=$ROOT_DIR" \
		package/$PKG_NAME/compile V=s
)

APK_SOURCE="$(find "$SDK_DIR/bin/packages" -type f \
	-name "$PKG_NAME-$VERSION-r$RELEASE.apk" -print -quit 2>/dev/null || true)"
if [ -z "$APK_SOURCE" ]; then
	echo "expected APK was not produced by the OpenWrt SDK" >&2
	find "$SDK_DIR/bin/packages" -type f -name "$PKG_NAME-*.apk" -print >&2 2>/dev/null || true
	exit 1
fi

mkdir -p "$DIST"
APK_PATH="$DIST/$PKG_NAME-$VERSION-r$RELEASE.apk"
cp -f "$APK_SOURCE" "$APK_PATH"

echo "== package"
ls -lh "$APK_PATH"
sha256sum "$APK_PATH"
echo "$APK_PATH"
