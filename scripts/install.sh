#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-bitnob/bitnob-cli}"
BINARY_NAME="${BINARY_NAME:-bitnob}"
INSTALL_DIR="${INSTALL_DIR:-}"
VERSION="${VERSION:-}"

log() {
  printf '[bitnob-install] %s\n' "$*"
}

fail() {
  printf '[bitnob-install] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    *)
      fail "unsupported OS: $os (supported: linux, darwin)"
      ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      fail "unsupported architecture: $arch (supported: amd64, arm64)"
      ;;
  esac
}

detect_target_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    echo "$INSTALL_DIR"
    return
  fi

  if [ -w "/usr/local/bin" ]; then
    echo "/usr/local/bin"
    return
  fi

  echo "${HOME}/.local/bin"
}

resolve_version() {
  if [ -n "$VERSION" ]; then
    echo "$VERSION"
    return
  fi

  local api_url tag
  api_url="https://api.github.com/repos/${REPO}/releases/latest"
  tag="$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
  [ -n "$tag" ] || fail "unable to resolve latest release tag from ${api_url}"
  echo "$tag"
}

verify_checksum() {
  local asset checksums_file asset_file expected actual
  asset="$1"
  checksums_file="$2"
  asset_file="$3"

  if command -v sha256sum >/dev/null 2>&1; then
    expected="$(grep " ${asset}\$" "$checksums_file" | awk '{print $1}' | head -n1)"
    [ -n "$expected" ] || fail "could not find checksum entry for ${asset}"
    printf '%s  %s\n' "$expected" "$asset_file" | sha256sum -c - >/dev/null
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    expected="$(grep " ${asset}\$" "$checksums_file" | awk '{print $1}' | head -n1)"
    [ -n "$expected" ] || fail "could not find checksum entry for ${asset}"
    actual="$(shasum -a 256 "$asset_file" | awk '{print $1}')"
    [ "$expected" = "$actual" ] || fail "checksum verification failed for ${asset}"
    return
  fi

  fail "no SHA-256 tool found (install sha256sum or shasum)"
}

main() {
  require_cmd curl
  require_cmd tar
  require_cmd mktemp

  local os arch version version_no_v release_base asset archive checksums_url archive_url
  local tmp_dir checksums_file extracted_dir binary_path target_dir target_bin

  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version)"
  version_no_v="${version#v}"

  release_base="https://github.com/${REPO}/releases/download/${version}"
  asset="${BINARY_NAME}_${version_no_v}_${os}_${arch}.tar.gz"
  archive="${asset}"
  archive_url="${release_base}/${archive}"
  checksums_url="${release_base}/checksums.txt"

  tmp_dir="$(mktemp -d)"
  trap "rm -rf '$tmp_dir'" EXIT

  checksums_file="${tmp_dir}/checksums.txt"

  log "installing ${BINARY_NAME} ${version} for ${os}/${arch}"
  log "downloading checksums"
  curl -fsSL -o "$checksums_file" "$checksums_url"

  log "downloading ${archive}"
  curl -fsSL -o "${tmp_dir}/${archive}" "$archive_url"

  log "verifying checksum"
  verify_checksum "$asset" "$checksums_file" "${tmp_dir}/${archive}"

  log "extracting archive"
  tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir"
  extracted_dir="${tmp_dir}/${BINARY_NAME}_${version_no_v}_${os}_${arch}"
  binary_path="${extracted_dir}/${BINARY_NAME}"
  [ -f "$binary_path" ] || fail "binary not found after extraction: ${binary_path}"

  target_dir="$(detect_target_dir)"
  mkdir -p "$target_dir"
  target_bin="${target_dir}/${BINARY_NAME}"

  install -m 0755 "$binary_path" "$target_bin"
  log "installed to ${target_bin}"

  if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
    log "binary directory may not be in PATH: ${target_dir}"
  fi

  "${target_bin}" version || true
}

main "$@"
