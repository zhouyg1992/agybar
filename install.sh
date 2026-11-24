#!/bin/sh
set -eu

REPO="zhouyg1992/agybar"
BIN_NAME="agybar"
VERSION="${AGYBAR_VERSION:-}"
BIN_DIR=""
SETTINGS_FILE=""
CONFIGURE_SETTINGS=1
INSTALL_PLUGIN=1
TMP_DIR=""
HAVE_TTY=0

usage() {
	cat <<'EOF'
Install agybar from GitHub Releases.

Usage:
  install.sh [options]

Options:
  --version vX.Y.Z   install a specific release (default: latest)
  --dir DIR          install directory (default: ~/.local/bin)
  --settings PATH    agy settings.json path (default: ~/.gemini/antigravity-cli/settings.json)
  --no-settings      install the binary only without modifying settings.json
  --no-plugin        skip registering the agy plugin
  -h, --help         show this help

Environment:
  AGYBAR_VERSION        same as --version
  AGYBAR_BIN_DIR        same as --dir
  AGYBAR_DOWNLOAD_BASE  override the release download base URL
EOF
}

log() {
	printf '%s\n' "$*"
}

step() {
	printf '\n✦ [%s/3] %s\n' "$1" "$2"
}

err() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

cleanup() {
	if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
		rm -rf "$TMP_DIR"
	fi
}
trap cleanup EXIT INT TERM

while [ $# -gt 0 ]; do
	case "$1" in
	--version)
		[ $# -ge 2 ] || err "--version requires a value"
		VERSION="$2"
		shift 2
		;;
	--dir)
		[ $# -ge 2 ] || err "--dir requires a value"
		BIN_DIR="$2"
		shift 2
		;;
	--settings)
		[ $# -ge 2 ] || err "--settings requires a value"
		SETTINGS_FILE="$2"
		shift 2
		;;
	--no-settings)
		CONFIGURE_SETTINGS=0
		shift
		;;
	--no-plugin)
		INSTALL_PLUGIN=0
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		err "unknown option: $1"
		;;
	esac
done

[ -n "${HOME:-}" ] || err "HOME is not set"
BIN_DIR="${BIN_DIR:-${AGYBAR_BIN_DIR:-$HOME/.local/bin}}"
SETTINGS_FILE="${SETTINGS_FILE:-$HOME/.gemini/antigravity-cli/settings.json}"

if { exec 3</dev/tty; } 2>/dev/null; then
	HAVE_TTY=1
fi

need uname
need tar
need mktemp

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
darwin | linux) ;;
*) err "unsupported OS: $OS (use macOS/Linux, or WSL on Windows)" ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
x86_64 | amd64) ARCH="amd64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*) err "unsupported architecture: $ARCH" ;;
esac

latest_version() {
	if command -v curl >/dev/null 2>&1; then
		url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest") || return 1
		case "$url" in
		*/releases/tag/*)
			printf '%s\n' "${url##*/}"
			return 0
			;;
		esac
		return 1
	fi
	if command -v wget >/dev/null 2>&1; then
		tag=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" |
			sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
			head -n 1)
		[ -n "$tag" ] || return 1
		printf '%s\n' "$tag"
		return 0
	fi
	return 1
}

download() {
	url="$1"
	out="$2"
	show_progress="${3:-0}"
	if command -v curl >/dev/null 2>&1; then
		if [ "$show_progress" -eq 1 ] && [ "$HAVE_TTY" -eq 1 ]; then
			curl -fL -# "$url" -o "$out"
		else
			curl -fsSL "$url" -o "$out"
		fi
	elif command -v wget >/dev/null 2>&1; then
		if [ "$show_progress" -eq 1 ] && [ "$HAVE_TTY" -eq 1 ]; then
			wget -O "$out" "$url"
		else
			wget -qO "$out" "$url"
		fi
	else
		err "curl or wget is required"
	fi
}

sha256_file() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		err "sha256sum or shasum is required"
	fi
}

verify_checksum() {
	url="$1"
	file="$2"
	name="$3"
	dir="$4"

	if ! download "$url" "$dir/checksums.txt"; then
		log "warning: checksums.txt unavailable; skipping verification"
		return 0
	fi
	expected=$(awk -v name="$name" '{f=$2; sub(/^\*/, "", f); if (f==name) {print $1; exit}}' "$dir/checksums.txt")
	if [ -z "$expected" ]; then
		log "warning: checksum entry not found; skipping verification"
		return 0
	fi
	actual=$(sha256_file "$file")
	[ "$actual" = "$expected" ] || err "checksum mismatch for $name"
	log "checksum verified"
}

configure_settings() {
	mkdir -p "$(dirname "$SETTINGS_FILE")"
	if command -v python3 >/dev/null 2>&1; then
		python3 - "$SETTINGS_FILE" "$BIN_DIR/$BIN_NAME" <<'PY'
import json, os, sys

path, bin_path = sys.argv[1], sys.argv[2]
data = {}
if os.path.exists(path):
    with open(path, "r", encoding="utf-8") as f:
        raw = f.read().strip()
        if raw:
            data = json.loads(raw)

data["statusLine"] = {
    "type": "command",
    "command": bin_path,
    "enabled": True,
    "padding": 0,
    "stack_with_default": False,
}

with open(path, "w", encoding="utf-8") as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write("\n")
PY
		log "updated $SETTINGS_FILE"
	elif command -v jq >/dev/null 2>&1; then
		tmp_json=$(mktemp)
		if [ -f "$SETTINGS_FILE" ] && [ -s "$SETTINGS_FILE" ]; then
			jq --arg cmd "$BIN_DIR/$BIN_NAME" '.statusLine = {"type":"command","command":$cmd,"enabled":true,"padding":0,"stack_with_default":false}' "$SETTINGS_FILE" >"$tmp_json"
		else
			jq -n --arg cmd "$BIN_DIR/$BIN_NAME" '{statusLine: {"type":"command","command":$cmd,"enabled":true,"padding":0,"stack_with_default":false}}' >"$tmp_json"
		fi
		mv "$tmp_json" "$SETTINGS_FILE"
		log "updated $SETTINGS_FILE"
	else
		log "warning: neither python3 nor jq found; add statusLine manually to $SETTINGS_FILE:"
		cat <<EOF
{
  "statusLine": {
    "type": "command",
    "command": "$BIN_DIR/$BIN_NAME",
    "enabled": true,
    "padding": 0,
    "stack_with_default": false
  }
}
EOF
	fi
}

log ""
log "✦ agybar installer"

step 1 "Download agybar release"
if [ -z "$VERSION" ]; then
	VERSION=$(latest_version) || err "could not determine the latest release; pass --version vX.Y.Z"
fi
case "$VERSION" in
v*) ;;
*) VERSION="v$VERSION" ;;
esac

VERSION_NO_V=${VERSION#v}
ASSET="${BIN_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
BASE_URL="${AGYBAR_DOWNLOAD_BASE:-https://github.com/${REPO}/releases/download/${VERSION}}"

TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t agybar 2>/dev/null || true)
[ -n "$TMP_DIR" ] || err "could not create a temporary directory"

log "  agybar ${VERSION} (${OS}/${ARCH})"
download "$BASE_URL/$ASSET" "$TMP_DIR/$ASSET" 1 || err "download failed: $BASE_URL/$ASSET"
verify_checksum "$BASE_URL/checksums.txt" "$TMP_DIR/$ASSET" "$ASSET" "$TMP_DIR"

step 2 "Install binary"
mkdir -p "$TMP_DIR/pkg"
tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR/pkg" || err "failed to extract $ASSET"

BIN_PATH="$TMP_DIR/pkg/$BIN_NAME"
if [ ! -f "$BIN_PATH" ]; then
	BIN_PATH=$(find "$TMP_DIR/pkg" -type f -name "$BIN_NAME" | head -n 1 || true)
fi
[ -n "${BIN_PATH:-}" ] && [ -f "$BIN_PATH" ] || err "binary not found in archive"

mkdir -p "$BIN_DIR" 2>/dev/null || err "cannot create $BIN_DIR (try --dir with a writable path)"
if command -v install >/dev/null 2>&1; then
	install -m 0755 "$BIN_PATH" "$BIN_DIR/$BIN_NAME" || err "failed to install to $BIN_DIR"
else
	cp "$BIN_PATH" "$BIN_DIR/$BIN_NAME" || err "failed to install to $BIN_DIR"
	chmod 0755 "$BIN_DIR/$BIN_NAME"
fi
log "installed $BIN_DIR/$BIN_NAME"

step 3 "Configure Antigravity CLI"
if [ "$CONFIGURE_SETTINGS" -eq 1 ]; then
	configure_settings
else
	log "skipped settings.json update (--no-settings)"
fi

if [ "$INSTALL_PLUGIN" -eq 1 ] && command -v agy >/dev/null 2>&1; then
	agy plugin install "https://github.com/${REPO}" >/dev/null 2>&1 || true
fi

log ""
log "✦ agybar ${VERSION} installed successfully! Restart agy to see it in action."
case ":${PATH:-}:" in
*":$BIN_DIR:"*) ;;
*) log "add to PATH: export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac
log ""
log "Quick commands:"
log "  $BIN_DIR/$BIN_NAME preview"
log "  $BIN_DIR/$BIN_NAME theme tokyo-night"
log "  $BIN_DIR/$BIN_NAME style badge"
log "docs: https://github.com/${REPO}#readme"
