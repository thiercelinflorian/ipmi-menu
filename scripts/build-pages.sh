#!/bin/sh
set -eu

base=${CI_PAGES_URL:?CI_PAGES_URL manquant}
version=${CI_COMMIT_TAG:-$(git describe --tags --always)}
version=${version#v}
out=public/latest

rm -rf public
mkdir -p "$out"
for os in darwin linux; do
  for arch in amd64 arm64; do
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w -X main.version=$version" -o ipmi-menu .
    tar -czf "$out/ipmi-menu_${os}_${arch}.tar.gz" ipmi-menu
  done
done
rm -f ipmi-menu
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$out" && sha256sum ./*.tar.gz > checksums.txt)
else
  (cd "$out" && shasum -a 256 ./*.tar.gz > checksums.txt)
fi
echo "$version" > "$out/version"

sed "s#^base=.*#base=\${IPMI_MENU_BASE_URL:-$base/latest}#" install.sh > public/install.sh
chmod +x public/install.sh

cat > public/index.html <<HTML
<!doctype html>
<html lang="fr">
<meta charset="utf-8">
<title>ipmi-menu $version</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:40em;margin:3em auto;padding:0 1em}pre{background:#f4f4f4;padding:1em;overflow:auto}</style>
<h1>ipmi-menu $version</h1>
<p>Installation sur macOS (Apple Silicon ou Intel) :</p>
<pre>curl -fsSL $base/install.sh | sh</pre>
<p>Il faut aussi <code>ipmitool</code> : <code>brew install ipmitool</code>.</p>
<h2>Binaires</h2>
<ul>
<li><a href="latest/ipmi-menu_darwin_arm64.tar.gz">macOS Apple Silicon</a></li>
<li><a href="latest/ipmi-menu_darwin_amd64.tar.gz">macOS Intel</a></li>
<li><a href="latest/ipmi-menu_linux_amd64.tar.gz">Linux amd64</a></li>
<li><a href="latest/ipmi-menu_linux_arm64.tar.gz">Linux arm64</a></li>
<li><a href="latest/checksums.txt">checksums.txt</a></li>
</ul>
</html>
HTML
echo "Pages prêt : $base ($version)"
