# ipmi-menu

Pilote des serveurs en IPMI depuis le terminal : menu interactif et commandes
scriptables. Un seul binaire, `ipmitool` requis.

![Démo du menu](docs/demo.gif)

## Installation

macOS : télécharger [install.sh](install.sh) et le lancer (`sh install.sh`).
Il faut aussi `ipmitool` : `brew install ipmitool`.

Linux : archive `ipmi-menu_linux_<arch>.tar.gz` de la dernière release, et
`apt install ipmitool`.

Si la v1 (Python) est installée : `pipx uninstall ipmi-menu`.

## Menu interactif

`ipmi-menu` affiche la liste des serveurs enregistrés, puis un menu navigable
aux flèches (ou `1`…`9`), avec l'état d'alimentation en en-tête : alimentation,
périphérique de démarrage (PXE, disque, CD, BIOS, en UEFI ou Legacy), console
SOL, capteurs, infos matériel, réseau du BMC. `esc` pour revenir, `r` pour
relire l'état, `q` pour quitter. Seules les actions qui coupent ou redémarrent
le serveur demandent une confirmation.

## Serveurs enregistrés

```bash
ipmi-menu host add srv1 10.0.0.10 -U ADMIN   # demande le mot de passe
ipmi-menu host ls
ipmi-menu host rm srv1
```

Stockés dans `~/.config/ipmi-menu/hosts.yaml` (droits 0600). Le mot de passe
est transmis à ipmitool par variable d'environnement, jamais en argument. Avec
un seul serveur enregistré, `-s` est facultatif ; sans serveur enregistré,
`-H ADRESSE -U UTILISATEUR -P MOTDEPASSE`.

## Ligne de commande

Chaque action du menu existe en sous-commande, sans confirmation, pour les
scripts.

```bash
ipmi-menu power -s srv1                     # état
ipmi-menu power on|off|soft|cycle|reset -s srv1
ipmi-menu boot pxe --reboot -s srv1         # --legacy, --persistent
ipmi-menu sol -s srv1                       # console série, ~. pour quitter
ipmi-menu sensors -s srv1
ipmi-menu hw -s srv1
ipmi-menu lan -s srv1
ipmi-menu raw -s srv1 -- user list 1        # commande ipmitool brute
```

## Mise à jour, désinstallation

Mise à jour : relancer `install.sh`. Désinstallation : `rm $(which ipmi-menu)`,
et `~/.config/ipmi-menu` si besoin.

## Développement

```bash
go test ./...
go build -o ipmi-menu .
```

Release : pousser un tag `v*`.
