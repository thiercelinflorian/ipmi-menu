#!/bin/sh
# Installe la dernière version d'ipmi-menu depuis les releases GitHub.
set -eu

REPO=thiercelinflorian/ipmi-menu

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "Architecture non supportée : $(uname -m)" >&2; exit 1 ;;
esac

dir=${INSTALL_DIR:-/usr/local/bin}
[ -w "$dir" ] || dir="$HOME/.local/bin"
mkdir -p "$dir"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "https://github.com/$REPO/releases/latest/download/ipmi-menu_${os}_${arch}.tar.gz" | tar -xz -C "$tmp" ipmi-menu
install -m 755 "$tmp/ipmi-menu" "$dir/ipmi-menu"
echo "ipmi-menu installé dans $dir"

command -v ipmitool >/dev/null 2>&1 || echo "⚠ ipmitool est requis : brew install ipmitool  ou  sudo apt install ipmitool"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "⚠ Ajoute $dir à ton PATH" ;;
esac
