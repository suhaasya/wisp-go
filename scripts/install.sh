#!/usr/bin/env sh
# Install wisp to ~/.local/bin (override with WISP_INSTALL_DIR).
set -eu

REPO="${WISP_INSTALL_REPO:-suhaasya/wisp-go}"
VERSION="${WISP_VERSION:-latest}"
BASE="${WISP_INSTALL_BASE:-https://github.com}/${REPO}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin) goos=darwin ;;
  linux) goos=linux ;;
  *) echo "install.sh: unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) goarch=amd64 ;;
  arm64|aarch64) goarch=arm64 ;;
  *) echo "install.sh: unsupported arch: $arch" >&2; exit 1 ;;
esac

ext=tar.gz
if [ "$goos" = "windows" ]; then
  ext=zip
fi

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "install.sh: missing required command: $1" >&2
    exit 1
  }
}
need curl
need sha256sum
need minisign

if [ "$VERSION" = "latest" ]; then
  tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
else
  tag="$VERSION"
fi
case "$tag" in v*) ;; *) tag="v${tag}" ;; esac
ver=$(echo "$tag" | sed 's/^v//')
archive="wisp_${ver}_${goos}_${goarch}.${ext}"
base_url="${BASE}/releases/download/${tag}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT INT TERM

if [ -n "${WISP_PUBLIC_KEY:-}" ]; then
  pubkey="$WISP_PUBLIC_KEY"
elif [ -f "$(dirname "$0")/../release/minisign.pub" ] 2>/dev/null && [ "$(basename "$0")" != "sh" ]; then
  pubkey="$(cd "$(dirname "$0")/.." && pwd)/release/minisign.pub"
else
  pubkey="$tmpdir/minisign.pub"
  curl -fsSL -o "$pubkey" "https://raw.githubusercontent.com/${REPO}/${tag}/release/minisign.pub"
fi

curl -fsSL -o "$tmpdir/checksums.txt" "${base_url}/checksums.txt"
curl -fsSL -o "$tmpdir/checksums.txt.minisig" "${base_url}/checksums.txt.minisig"
minisign -Vqm "$tmpdir/checksums.txt" -x "$tmpdir/checksums.txt.minisig" -p "$pubkey"

curl -fsSL -o "$tmpdir/$archive" "${base_url}/${archive}"
line=$(grep "  ${archive}\$" "$tmpdir/checksums.txt" || true)
if [ -z "$line" ]; then
  echo "install.sh: no checksum line for $archive" >&2
  exit 1
fi
want=$(echo "$line" | awk '{print $1}')
got=$(sha256sum "$tmpdir/$archive" | awk '{print $1}')
if [ "$want" != "$got" ]; then
  echo "install.sh: checksum mismatch" >&2
  exit 1
fi

case "$ext" in
  tar.gz)
    tar -xzf "$tmpdir/$archive" -C "$tmpdir"
    bin="$tmpdir/wisp"
    ;;
  zip)
    need unzip
    unzip -q "$tmpdir/$archive" -d "$tmpdir"
    bin="$tmpdir/wisp.exe"
    ;;
esac

install_dir="${WISP_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"
install -m 0755 "$bin" "$install_dir/wisp"
echo "Installed wisp ${ver} to ${install_dir}/wisp"
