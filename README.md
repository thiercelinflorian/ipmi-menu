# ipmi-menu

Piloter des serveurs en IPMI depuis le terminal : un menu plein écran navigable
aux flèches, et des commandes directes pour les scripts. Un seul binaire, seul
`ipmitool` est requis.

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

  ✓ Prochain démarrage : PXE, serveur redémarré
  ↑↓ choisir · ⏎ valider · esc retour · r rafraîchir l'état · q quitter
```

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh | sh
```

Ou avec Go : `go install github.com/thiercelinflorian/ipmi-menu@latest`.

Il faut aussi `ipmitool` : `brew install ipmitool` ou `sudo apt install ipmitool`.

Mise à jour : relancer la même commande. Désinstallation : supprimer le binaire
(`rm $(which ipmi-menu)`) et éventuellement `~/.config/ipmi-menu`.

> Depuis la v1 (Python) : `pipx uninstall ipmi-menu` avant d'installer.

## Menu interactif

```bash
ipmi-menu            # choisir un serveur dans la liste
ipmi-menu -s srv1    # ouvrir directement sur srv1
```

| Touche | Action |
| --- | --- |
| `↑` `↓` / `j` `k` | se déplacer |
| `⏎` ou `1`…`9` | valider |
| `esc` | retour |
| `r` | rafraîchir l'état d'alimentation |
| `e` / `d` | modifier / supprimer un serveur (liste des serveurs) |
| `q` | quitter |

Seules les actions qui coupent ou redémarrent le serveur demandent une
confirmation, en une touche (`o`).

## Ligne de commande

Aucune confirmation : les commandes sont faites pour être scriptées.

```bash
ipmi-menu power -s srv1                  # état
ipmi-menu power on|off|soft|cycle|reset -s srv1
ipmi-menu boot pxe --reboot -s srv1      # options : --legacy --persistent
ipmi-menu sol -s srv1                    # console série, ~. pour quitter
ipmi-menu sensors | hw | lan -s srv1
ipmi-menu raw -s srv1 -- user list 1     # n'importe quelle commande ipmitool
```

Sans serveur enregistré : `-H ADRESSE -U UTILISATEUR -P MOTDEPASSE` (ou la
variable `IPMI_PASSWORD`). S'il n'y a qu'un seul serveur enregistré, `-s` est
facultatif.

Autocomplétion : `ipmi-menu completion zsh|bash|fish --help`.

## Serveurs enregistrés

```bash
ipmi-menu host add srv1 10.0.0.10 -U ADMIN   # demande le mot de passe
ipmi-menu host ls
ipmi-menu host rm srv1
```

Ils sont stockés dans `~/.config/ipmi-menu/hosts.yaml`, lisible par vous seul
(0600). Les mots de passe y sont en clair. Ils sont transmis à `ipmitool` par
variable d'environnement, jamais sur la ligne de commande (invisibles dans `ps`).

## Développement

```bash
go test ./...
go build -o ipmi-menu .
```

Les tests utilisent un faux `ipmitool` (`internal/ipmi/testdata/ipmitool`), qui
sert aussi à essayer le menu sans BMC :
`PATH=$PWD/internal/ipmi/testdata:$PATH go run .`

Une release (binaires macOS/Linux, amd64/arm64) est publiée automatiquement en
poussant un tag `v*`.
