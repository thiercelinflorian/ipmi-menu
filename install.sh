#!/bin/sh
set -eu

REPO=thiercelinflorian/ipmi-menu
base=${IPMI_MENU_BASE_URL:-https://github.com/$REPO/releases/latest/download}

[ "$(uname -s)" = Darwin ] || { echo "Ce script n'installe ipmi-menu que sur macOS. Binaires Linux : $base" >&2; exit 1; }
case $(uname -m) in
  x86_64) arch=amd64 ;;
  arm64) arch=arm64 ;;
  *) echo "Architecture non supportée : $(uname -m)" >&2; exit 1 ;;
esac

dir=${INSTALL_DIR:-/usr/local/bin}
[ -w "$dir" ] || dir="$HOME/.local/bin"
mkdir -p "$dir"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
url="$base/ipmi-menu_darwin_${arch}.tar.gz"
curl -fsSL "$url" | tar -xz -C "$tmp" ipmi-menu || { echo "Téléchargement impossible : $url" >&2; exit 1; }
install -m 755 "$tmp/ipmi-menu" "$dir/ipmi-menu"
echo "ipmi-menu $("$dir/ipmi-menu" --version | awk '{print $NF}') installé dans $dir"

command -v ipmitool >/dev/null 2>&1 || echo "⚠ ipmitool est requis : brew install ipmitool"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "⚠ Ajoute $dir à ton PATH (export PATH=\"$dir:\$PATH\")" ;;
esac
