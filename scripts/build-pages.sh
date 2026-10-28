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
cp docs/demo.gif public/demo.gif

cat > public/index.html <<HTML
<!doctype html>
<html lang="fr">
<meta charset="utf-8">
<title>ipmi-menu $version</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:40em;margin:3em auto;padding:0 1em}pre{background:#f4f4f4;padding:1em;overflow:auto}h2{margin-top:2em}h3{margin-top:1.5em}</style>
<h1>ipmi-menu $version</h1>
<img src="demo.gif" alt="Démo du menu ipmi-menu" style="max-width:100%;border-radius:6px">
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

<h2>Fonctionnement</h2>

<h3>Menu interactif</h3>
<p><code>ipmi-menu</code> affiche la liste des serveurs enregistrés, puis un menu
navigable aux flèches (ou <code>1</code>…<code>9</code>), avec l'état
d'alimentation en en-tête : alimentation, périphérique de démarrage (PXE,
disque, CD, BIOS, en UEFI ou Legacy), console SOL, capteurs, infos matériel,
réseau du BMC. <code>esc</code> pour revenir, <code>r</code> pour relire
l'état, <code>q</code> pour quitter. Seules les actions qui coupent ou
redémarrent le serveur demandent une confirmation.</p>

<h3>Serveurs enregistrés</h3>
<pre>ipmi-menu host add srv1 10.0.0.10 -U ADMIN   # demande le mot de passe
ipmi-menu host ls
ipmi-menu host rm srv1</pre>
<p>Stockés dans <code>~/.config/ipmi-menu/hosts.yaml</code> (droits 0600). Le
mot de passe est transmis à ipmitool par variable d'environnement, jamais en
argument. Avec un seul serveur enregistré, <code>-s</code> est facultatif ;
sans serveur enregistré, <code>-H ADRESSE -U UTILISATEUR -P MOTDEPASSE</code>.</p>

<h3>Ligne de commande</h3>
<p>Chaque action du menu existe en sous-commande, sans confirmation, pour les
scripts.</p>
<pre>ipmi-menu power -s srv1                     # état
ipmi-menu power on|off|soft|cycle|reset -s srv1
ipmi-menu boot pxe --reboot -s srv1         # --legacy, --persistent
ipmi-menu sol -s srv1                       # console série, ~. pour quitter
ipmi-menu sensors -s srv1
ipmi-menu hw -s srv1
ipmi-menu lan -s srv1
ipmi-menu raw -s srv1 -- user list 1        # commande ipmitool brute</pre>

<h3>Mise à jour, désinstallation</h3>
<p>Mise à jour : relancer la commande d'installation. Désinstallation :
<code>rm \$(which ipmi-menu)</code>, et <code>~/.config/ipmi-menu</code> si
besoin.</p>
</html>
HTML
echo "Pages prêt : $base ($version)"
