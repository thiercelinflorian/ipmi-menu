# ipmi-menu

Pilote des serveurs en IPMI depuis le terminal : un menu interactif, et des
sous-commandes pour les scripts. Un seul binaire, `ipmitool` requis.

```
  ipmi-menu  srv1 · 10.0.0.10                                     ● allumé

  Menu › Démarrage

  ▸ 1 PXE
    2 Disque
    3 CD/DVD · ISO virtuel
    4 Setup BIOS

    5 Mode  UEFI
    6 Persistant  non (une seule fois)
    7 Redémarrer après  oui
```

## Installation

macOS :

```bash
curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh | sh
brew install ipmitool
```

Le script [install.sh](install.sh) installe la dernière version dans
`/usr/local/bin` ou `~/.local/bin`. Pour mettre à jour, relancer la commande.
Sur le GitLab interne, utiliser l'URL Pages du projet à la place de l'URL
GitHub.

Linux : archive `ipmi-menu_linux_<arch>.tar.gz` de la dernière release, et
`apt install ipmitool`. Ou `go install github.com/thiercelinflorian/ipmi-menu@latest`.

Si la v1 (Python) est installée : `pipx uninstall ipmi-menu`.

## Menu

```bash
ipmi-menu            # liste des serveurs
ipmi-menu -s srv1    # directement sur srv1
```

Flèches ou `j`/`k` pour se déplacer, `⏎` ou `1`…`9` pour valider, `esc` pour
revenir, `r` pour relire l'état d'alimentation, `q` pour quitter. Dans la liste
des serveurs, `e` modifie et `d` supprime. Les actions qui coupent ou
redémarrent le serveur demandent une confirmation (`o`).

## Ligne de commande

Pas de confirmation.

```bash
ipmi-menu power -s srv1                     # état
ipmi-menu power on|off|soft|cycle|reset -s srv1
ipmi-menu boot pxe --reboot -s srv1         # --legacy, --persistent
ipmi-menu netboot --reboot -s srv1          # netboot.xyz, --eject pour retirer l'ISO
ipmi-menu sol -s srv1                       # console série, ~. pour quitter
ipmi-menu sensors -s srv1
ipmi-menu hw -s srv1
ipmi-menu lan -s srv1
ipmi-menu raw -s srv1 -- user list 1        # commande ipmitool brute
```

Sans serveur enregistré : `-H ADRESSE -U UTILISATEUR -P MOTDEPASSE`, ou la
variable `IPMI_PASSWORD`. Avec un seul serveur enregistré, `-s` est facultatif.

Complétion : `ipmi-menu completion zsh|bash|fish --help`.

## Serveurs enregistrés

```bash
ipmi-menu host add srv1 10.0.0.10 -U ADMIN   # demande le mot de passe
ipmi-menu host ls
ipmi-menu host rm srv1
```

Fichier `~/.config/ipmi-menu/hosts.yaml`, en 0600, mots de passe en clair.
Le mot de passe est passé à `ipmitool` par variable d'environnement, pas en
argument.

## netboot.xyz

`ipmi-menu netboot` (ou l'entrée du menu Démarrage) monte l'ISO
[netboot.xyz](https://netboot.xyz) dans le CD virtuel du BMC via Redfish, avec
les mêmes identifiants, puis démarre une fois dessus. Le BMC doit savoir monter
une image depuis une URL HTTP et joindre `boot.netboot.xyz`. Sinon, indiquer
un miroir :

```yaml
# ~/.config/ipmi-menu/hosts.yaml
netboot_iso: http://miroir.local/netboot.xyz.iso
```

ou `ipmi-menu netboot --iso URL`.

## Développement

```bash
go test ./...
go build -o ipmi-menu .
PATH=$PWD/internal/ipmi/testdata:$PATH go run .   # menu sans BMC, faux ipmitool
```

Release : pousser un tag `v*`. Sur GitHub, le workflow publie les binaires
macOS et Linux dans la release. Sur GitLab, le job `pages` publie
`install.sh` et les binaires sur GitLab Pages, à chaque tag et à chaque push
sur `main` ([scripts/build-pages.sh](scripts/build-pages.sh)).
