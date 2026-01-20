# ipmi-menu

Menu interactif autour de `ipmitool` pour gérer vos serveurs via IPMI/BMC.

## Fonctionnalités

- **Power** : Allumer, éteindre, redémarrer (cycle/reset/soft), statut
- **Boot Options** : Configurer le périphérique de boot (PXE, disque, CD-ROM, BIOS), mode UEFI/Legacy, boot persistant ou one-shot
- **SOL (Serial Over LAN)** : Accès console distante via le port série
- **Infos** : Capteurs (températures, ventilateurs), infos matériel (FRU), configuration réseau BMC
- **Multilingue** : Interface en français et anglais
- **Paramètres** : Sauvegarde des credentials par défaut (username/password)

## Prérequis
- Python 3.9+
- `ipmitool` dans le PATH

## Installation rapide (recommandée)

Le script d'installation détecte automatiquement votre OS (macOS/Linux) et installe toutes les dépendances nécessaires.

### Installation
```bash
curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh | bash
```

### Mise à jour
```bash
curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh | bash -s -- --upgrade
```

### Désinstallation
```bash
curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh | bash -s -- --uninstall
```

## Installation manuelle

### macOS
```bash
brew install pipx ipmitool
pipx install ipmi-menu
pipx ensurepath
```

### Linux (Debian/Ubuntu)
```bash
sudo apt-get install -y pipx ipmitool
pipx install ipmi-menu
pipx ensurepath
```

## Utilisation

Redémarrez votre terminal puis lancez :
```bash
ipmi-menu
```