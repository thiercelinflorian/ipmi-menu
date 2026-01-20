from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Dict

from .settings import DEFAULT_LOCALE

CONFIG_DIR = Path.home() / ".config" / "ipmi-menu"
PREFERENCES_FILE = CONFIG_DIR / "preferences.json"


def _ensure_config_dir() -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)


def load_preferences() -> Dict[str, Any]:
    if not PREFERENCES_FILE.exists():
        return {}
    try:
        with open(PREFERENCES_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return {}


def save_preferences(prefs: Dict[str, Any]) -> None:
    _ensure_config_dir()
    with open(PREFERENCES_FILE, "w", encoding="utf-8") as f:
        json.dump(prefs, f, indent=2)


def get_preferred_language() -> str:
    prefs = load_preferences()
    return prefs.get("language", DEFAULT_LOCALE)


def set_preferred_language(lang: str) -> None:
    prefs = load_preferences()
    prefs["language"] = lang
    save_preferences(prefs)


def get_preferred_username() -> str | None:
    prefs = load_preferences()
    return prefs.get("username")


def set_preferred_username(username: str | None) -> None:
    prefs = load_preferences()
    if username:
        prefs["username"] = username
    elif "username" in prefs:
        del prefs["username"]
    save_preferences(prefs)


def get_preferred_password() -> str | None:
    prefs = load_preferences()
    return prefs.get("password")


def set_preferred_password(password: str | None) -> None:
    prefs = load_preferences()
    if password is not None:
        prefs["password"] = password
    elif "password" in prefs:
        del prefs["password"]
    save_preferences(prefs)
