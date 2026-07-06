#!/usr/bin/env bash
set -euo pipefail

REPO="rodionlim/portfolio-manager-go"
APP_NAME="portfolio-manager"
INSTALL_DIR="${PORTFOLIO_MANAGER_HOME:-$HOME/portfolio-manager}"

usage() {
  cat <<EOF
Portfolio Manager installer

Usage:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash -s -- --run

Environment:
  PORTFOLIO_MANAGER_HOME  Install directory. Defaults to ~/portfolio-manager.

Options:
  --run                   Start Portfolio Manager after installation.
  -h, --help              Show this help.
EOF
}

RUN_AFTER_INSTALL=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --run)
      RUN_AFTER_INSTALL=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

download() {
  local url="$1"
  local output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fL --progress-bar -o "$output" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q --show-progress -O "$output" "$url"
  else
    echo "Missing required command: curl or wget" >&2
    exit 1
  fi
}

detect_asset() {
  local os
  local arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}:${arch}" in
    Darwin:arm64|Darwin:aarch64)
      echo "portfolio-manager_darwin_arm64"
      ;;
    Linux:x86_64|Linux:amd64)
      echo "portfolio-manager_linux_amd64"
      ;;
    *)
      echo "Unsupported platform: ${os} ${arch}" >&2
      echo "Download a matching asset manually from https://github.com/${REPO}/releases/latest" >&2
      exit 1
      ;;
  esac
}

main() {
  need_cmd uname
  need_cmd mkdir
  need_cmd chmod

  local asset
  local url
  local binary
  asset="$(detect_asset)"
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
  binary="${INSTALL_DIR}/${APP_NAME}"

  echo "Installing Portfolio Manager into ${INSTALL_DIR}"
  mkdir -p "$INSTALL_DIR"
  download "$url" "$binary"
  chmod +x "$binary"

  if command -v xattr >/dev/null 2>&1; then
    xattr -d com.apple.quarantine "$binary" 2>/dev/null || true
  fi

  echo
  echo "Portfolio Manager installed:"
  echo "  ${binary}"
  echo
  echo "Start it with:"
  echo "  cd ${INSTALL_DIR}"
  echo "  ./portfolio-manager"
  echo
  echo "Default URLs:"
  echo "  Backend: http://localhost:8080"
  echo "  MCP:     http://localhost:8081/mcp"

  if [[ "$RUN_AFTER_INSTALL" -eq 1 ]]; then
    echo
    echo "Starting Portfolio Manager..."
    cd "$INSTALL_DIR"
    exec ./portfolio-manager
  fi
}

main
