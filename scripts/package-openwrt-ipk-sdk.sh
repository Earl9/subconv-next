#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

SDK_DIR="${SDK_DIR:-/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64}"
VERSION="${VERSION:-1.0.7}"
RELEASE="${RELEASE:-33}"
ARCH="${ARCH:-aarch64_generic}"
PKG_NAME="${PKG_NAME:-subconv-next}"
MAINTAINER="${MAINTAINER:-SubConv Next Maintainers}"
DIST="${DIST:-$ROOT_DIR/dist}"
GO_BUILDER_IMAGE="${GO_BUILDER_IMAGE:-golang:1.25.12-alpine}"
USE_DOCKER_GO="${USE_DOCKER_GO:-auto}"
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

if [ "$USE_PROVIDED_BIN" -eq 0 ]; then
	echo "building linux/arm64 binary: $BIN_PATH"
	mkdir -p "$(dirname -- "$BIN_PATH")"
	case "$USE_DOCKER_GO" in
		1|true|yes) use_docker_go=1 ;;
		0|false|no) use_docker_go=0 ;;
		auto) if command -v docker >/dev/null 2>&1; then use_docker_go=1; else use_docker_go=0; fi ;;
		*) echo "USE_DOCKER_GO must be auto, true, or false" >&2; exit 1 ;;
	esac
	if [ "$use_docker_go" -eq 1 ]; then
		need_cmd docker
		output_dir="$(CDPATH= cd -- "$(dirname -- "$BIN_PATH")" && pwd)"
		output_name="$(basename -- "$BIN_PATH")"
		docker run --rm \
			-e CGO_ENABLED=0 \
			-e GOOS=linux \
			-e GOARCH=arm64 \
			-v "$ROOT_DIR:/src" \
			-v "$output_dir:/out" \
			-w /src \
			"$GO_BUILDER_IMAGE" \
			go build -trimpath -ldflags="-s -w -X main.version=$PKG_VERSION" \
			-o "/out/$output_name" ./cmd/subconv-next
	else
		need_cmd go
		(
			cd "$ROOT_DIR"
			CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
				go build -trimpath -ldflags="-s -w -X main.version=$PKG_VERSION" \
				-o "$BIN_PATH" ./cmd/subconv-next
		)
	fi
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
	"$PKGROOT/usr/share/subconv-next"

install -m0755 "$BIN_PATH" "$PKGROOT/usr/bin/subconv-next"
install -m0755 "$INIT_FILE" "$PKGROOT/etc/init.d/subconv-next"
install -m0644 "$CONFIG_FILE" "$PKGROOT/etc/config/subconv-next"
install -m0644 "$CONFIG_FILE" "$PKGROOT/usr/share/subconv-next/subconv-next.config"
chmod 0755 "$PKGROOT/etc/subconv-next/data"

cat >"$PKGROOT/CONTROL/control" <<EOF_CONTROL
Package: $PKG_NAME
Version: $PKG_VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Section: net
Priority: optional
Depends: ca-bundle, uci
Description: Modern subscription converter for Mihomo / Clash Meta.
EOF_CONTROL

cat >"$PKGROOT/CONTROL/preinst" <<'EOF_PREINST'
#!/bin/sh

root="${IPKG_INSTROOT:-}"
config_path="$root/etc/config/subconv-next"
backup_path="$root/tmp/subconv-next.config.keep"
persistent_backup="$root/etc/config/subconv-next-opkg.backup"

if [ -f "$config_path" ]; then
	mkdir -p "$root/tmp"
	cp "$config_path" "$backup_path" || true
	cp "$config_path" "$persistent_backup" || true
	chmod 0600 "$persistent_backup" 2>/dev/null || true
fi

exit 0
EOF_PREINST

cat >"$PKGROOT/CONTROL/postinst" <<'EOF_POSTINST'
#!/bin/sh

root="${IPKG_INSTROOT:-}"
config_path="$root/etc/config/subconv-next"
default_config="$root/usr/share/subconv-next/subconv-next.config"
backup_path="$root/tmp/subconv-next.config.keep"
persistent_backup="$root/etc/config/subconv-next-opkg.backup"

if [ -f "$backup_path" ]; then
	mkdir -p "$root/etc/config"
	cp "$backup_path" "$config_path"
	chmod 0644 "$config_path"
	cp "$config_path" "$persistent_backup" || true
	chmod 0600 "$persistent_backup" 2>/dev/null || true
elif [ -f "$persistent_backup" ]; then
	mkdir -p "$root/etc/config"
	cp "$persistent_backup" "$config_path"
	chmod 0644 "$config_path"
elif [ ! -e "$config_path" ] && [ -f "$default_config" ]; then
	mkdir -p "$root/etc/config"
	cp "$default_config" "$config_path"
	chmod 0644 "$config_path"
fi
[ ! -f "$persistent_backup" ] || chmod 0600 "$persistent_backup" 2>/dev/null || true
rm -f "$backup_path"

[ -n "$IPKG_INSTROOT" ] && exit 0

if [ -x /etc/init.d/subconv-next ]; then
	/etc/init.d/subconv-next enable >/dev/null 2>&1 || true
	enabled="$(uci -q get subconv-next.main.enabled)"
	[ -n "$enabled" ] || enabled=1
	if [ "$enabled" = 1 ]; then
		/etc/init.d/subconv-next start >/dev/null 2>&1 || true
	fi
fi

exit 0
EOF_POSTINST

cat >"$PKGROOT/CONTROL/prerm" <<'EOF_PRERM'
#!/bin/sh

[ -n "$IPKG_INSTROOT" ] && exit 0

if [ -x /etc/init.d/subconv-next ]; then
	if [ -f /etc/config/subconv-next ]; then
		cp /etc/config/subconv-next /etc/config/subconv-next-opkg.backup >/dev/null 2>&1 || true
		chmod 0600 /etc/config/subconv-next-opkg.backup >/dev/null 2>&1 || true
	fi
	/etc/init.d/subconv-next stop >/dev/null 2>&1 || true
	/etc/init.d/subconv-next disable >/dev/null 2>&1 || true
fi

exit 0
EOF_PRERM

chmod 0644 "$PKGROOT/CONTROL/control"
chmod 0755 "$PKGROOT/CONTROL/preinst" "$PKGROOT/CONTROL/postinst" "$PKGROOT/CONTROL/prerm"

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
check_mode usr/share/subconv-next/subconv-next.config 644

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

if tar -tzf "$CHECK_DIR/control.tar.gz" | grep -Fx './conffiles' >/dev/null; then
	echo "unexpected control metadata: ./conffiles" >&2
	exit 1
fi

echo "== control"
tar -xOzf "$CHECK_DIR/control.tar.gz" ./control

for script in preinst postinst prerm; do
	if ! tar -tzf "$CHECK_DIR/control.tar.gz" | grep -Fx "./$script" >/dev/null; then
		echo "missing control script: ./$script" >&2
		exit 1
	fi
done

CONTROL_CHECK_DIR="$CHECK_DIR/control"
mkdir -p "$CONTROL_CHECK_DIR"
tar -xzf "$CHECK_DIR/control.tar.gz" -C "$CONTROL_CHECK_DIR"

for script in preinst postinst prerm; do
	actual="$(stat -c '%a' "$CONTROL_CHECK_DIR/$script")"
	if [ "$actual" != "755" ]; then
		echo "invalid control script mode for $script: $actual, want 755" >&2
		exit 1
	fi
done

echo "== data.tar.gz | head -50"
tar -tzf "$CHECK_DIR/data.tar.gz" | head -50

if ! tar -tzf "$CHECK_DIR/data.tar.gz" | grep -Fx './etc/config/subconv-next' >/dev/null; then
	echo "missing packaged config: ./etc/config/subconv-next" >&2
	exit 1
fi

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
check_data_mode ./usr/share/subconv-next/subconv-next.config 644

echo "$IPK_PATH"
