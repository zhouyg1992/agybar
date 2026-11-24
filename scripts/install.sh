#!/usr/bin/env bash
set -euo pipefail

REPO="zhouyg1992/agybar"
INSTALL_DIR="${AGYBAR_INSTALL_DIR:-$HOME/.local/bin}"
BIN_PATH="$INSTALL_DIR/agybar"
SETTINGS_FILE="$HOME/.gemini/antigravity-cli/settings.json"

info()  { printf "\033[1;34m[agybar]\033[0m %s\n" "$*"; }
ok()    { printf "\033[1;32m[agybar]\033[0m %s\n" "$*"; }
warn()  { printf "\033[1;33m[agybar]\033[0m %s\n" "$*"; }
err()   { printf "\033[1;31m[agybar]\033[0m %s\n" "$*" >&2; }

detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  case "$OS" in
    darwin|linux) ;;
    *) err "Unsupported OS: $OS"; exit 1 ;;
  esac
  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) err "Unsupported architecture: $ARCH"; exit 1 ;;
  esac
}

install_from_release() {
  local asset="agybar_${OS}_${ARCH}.tar.gz"
  local url="https://github.com/${REPO}/releases/latest/download/${asset}"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  info "Checking latest prebuilt release (${OS}/${ARCH})..."
  if curl -fsSL "$url" -o "$tmp/$asset" 2>/dev/null; then
    tar -xzf "$tmp/$asset" -C "$tmp"
    mkdir -p "$INSTALL_DIR"
    install -m 755 "$tmp/agybar" "$BIN_PATH"
    ok "Installed prebuilt binary to $BIN_PATH"
    return 0
  fi
  return 1
}

install_from_go() {
  if ! command -v go >/dev/null 2>&1; then
    err "No prebuilt release available yet and 'go' is not installed."
    err "Please install Go (https://go.dev/dl/) or wait for a GitHub Release."
    exit 1
  fi
  info "Building agybar from source with Go..."
  mkdir -p "$INSTALL_DIR"
  GOBIN="$INSTALL_DIR" go install "github.com/${REPO}/cmd/agybar@latest"
  ok "Built and installed binary to $BIN_PATH"
}

configure_settings() {
  mkdir -p "$(dirname "$SETTINGS_FILE")"
  if command -v python3 >/dev/null 2>&1; then
    python3 - "$SETTINGS_FILE" "$BIN_PATH" <<'PY'
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
    ok "Updated $SETTINGS_FILE"
  elif command -v jq >/dev/null 2>&1; then
    local tmp
    tmp="$(mktemp)"
    if [ -f "$SETTINGS_FILE" ] && [ -s "$SETTINGS_FILE" ]; then
      jq --arg cmd "$BIN_PATH" '.statusLine = {"type":"command","command":$cmd,"enabled":true,"padding":0,"stack_with_default":false}' "$SETTINGS_FILE" > "$tmp"
    else
      jq -n --arg cmd "$BIN_PATH" '{statusLine: {"type":"command","command":$cmd,"enabled":true,"padding":0,"stack_with_default":false}}' > "$tmp"
    fi
    mv "$tmp" "$SETTINGS_FILE"
    ok "Updated $SETTINGS_FILE"
  else
    warn "Neither python3 nor jq found; please add statusLine manually to $SETTINGS_FILE:"
    cat <<EOF
{
  "statusLine": {
    "type": "command",
    "command": "$BIN_PATH",
    "enabled": true,
    "padding": 0,
    "stack_with_default": false
  }
}
EOF
  fi
}

install_plugin_if_available() {
  if command -v agy >/dev/null 2>&1; then
    info "Registering agybar plugin with agy..."
    agy plugin install "https://github.com/${REPO}" >/dev/null 2>&1 || true
  fi
}

main() {
  detect_platform
  if ! install_from_release; then
    install_from_go
  fi
  configure_settings
  install_plugin_if_available

  echo
  ok "Installation complete! Restart agy to see agybar at the bottom of your terminal."
  echo "  Preview themes:  $BIN_PATH preview"
  echo "  Switch theme:    $BIN_PATH theme tokyo-night"
  echo "  Switch style:    $BIN_PATH style pill"
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo
    warn "Tip: Add '$INSTALL_DIR' to your PATH to run 'agybar' directly."
  fi
}

main "$@"
