#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/subconv-next-portable-ipkg.XXXXXX")"
trap 'rm -rf "$WORK_DIR"' EXIT INT TERM

mkdir -p "$WORK_DIR/sdk/scripts"

cat >"$WORK_DIR/sdk/scripts/ipkg-build" <<'EOF_IPKG_BUILD'
#!/bin/sh
set -eu

usage="Usage: $0 <pkg_directory> [<destination_directory>]"

case "${1:-}" in
	-h|--help)
		echo "$usage"
		exit 0
		;;
esac

case "$#" in
	1)
		dest_dir="$PWD"
		;;
	2)
		dest_dir="$2"
		;;
	*)
		echo "$usage" >&2
		exit 1
		;;
esac

pkg_dir="$(realpath "$1")"
control_dir="$pkg_dir/CONTROL"
if [ ! -d "$control_dir" ]; then
	echo "*** Error: Directory $pkg_dir has no CONTROL subdirectory." >&2
	exit 1
fi

field_value() {
	sed -n "s/^$1:[[:space:]]*//p" "$control_dir/control" | head -n 1
}

pkg="$(field_value Package)"
version="$(field_value Version)"
arch="$(field_value Architecture)"
if [ -z "$pkg" ] || [ -z "$version" ] || [ -z "$arch" ]; then
	echo "*** Error: control file must include Package, Version, and Architecture." >&2
	exit 1
fi

if echo "$pkg" | grep '[^a-zA-Z0-9_.+-]' >/dev/null; then
	echo "*** Error: Package name '$pkg' contains illegal characters." >&2
	exit 1
fi

mkdir -p "$dest_dir"
tmp_dir="$dest_dir/IPKG_BUILD.$$"
mkdir "$tmp_dir"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

timestamp="@${SOURCE_DATE_EPOCH:-0}"

echo CONTROL > "$tmp_dir/tarX"
(
	cd "$pkg_dir"
	tar --exclude=./CONTROL --exclude=./CONTROL/\* --format=gnu --numeric-owner --sort=name --mtime="$timestamp" -cpf - . | gzip -n - > "$tmp_dir/data.tar.gz"
)

(
	cd "$control_dir"
	tar --format=gnu --numeric-owner --sort=name --mtime="$timestamp" -cf - . | gzip -n - > "$tmp_dir/control.tar.gz"
)

printf '2.0\n' > "$tmp_dir/debian-binary"

pkg_file="$dest_dir/${pkg}_${version}_${arch}.ipk"
rm -f "$pkg_file"
(
	cd "$tmp_dir"
	tar --format=gnu --numeric-owner --sort=name --mtime="$timestamp" -cf - ./debian-binary ./data.tar.gz ./control.tar.gz | gzip -n - > "$pkg_file"
)

echo "Packaged contents of $pkg_dir into $pkg_file"
EOF_IPKG_BUILD

chmod 0755 "$WORK_DIR/sdk/scripts/ipkg-build"

SDK_DIR="$WORK_DIR/sdk" exec "$ROOT_DIR/scripts/package-openwrt-ipk-sdk.sh" "$@"
