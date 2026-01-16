# ipmi-menu

Menu interactif autour de `ipmitool` (Power, Boot Option, SOL, Infos).

## Prérequis
- Python 3.9+
- `ipmitool` dans le PATH

## Lancer

### Mode simple
```bash
./scripts/ipmi
```

## Notes
- Les textes (prompts/labels/messages) sont centralisés dans `ipmi_menu/config/messages.fr.json`.
- Les defaults (user, password, interface, port, timeout) sont dans `ipmi_menu/config/settings.py`.
