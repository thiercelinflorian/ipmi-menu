from __future__ import annotations

import base64
import json
import os
from pathlib import Path
from typing import Any, Dict

from .settings import DEFAULT_LOCALE

CONFIG_DIR = Path.home() / ".config" / "ipmi-menu"
PREFERENCES_FILE = CONFIG_DIR / "preferences.json"

CURRENT_VERSION = 1


def _ensure_config_dir() -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)


def _encode_password(password: str) -> str:
    return base64.b64encode(password.encode("utf-8")).decode("ascii")


def _decode_password(encoded: str) -> str:
    return base64.b64decode(encoded.encode("ascii")).decode("utf-8")


def _is_base64(value: str) -> bool:
    try:
        decoded = base64.b64decode(value.encode("ascii"))
        return base64.b64encode(decoded).decode("ascii") == value
    except Exception:
        return False


def _migrate(prefs: Dict[str, Any]) -> Dict[str, Any]:
    """Migrate preferences from older formats to the current version."""
    if prefs.get("version", 0) >= CURRENT_VERSION:
        return prefs

    # Migrate plaintext password to base64-encoded
    pw = prefs.get("password")
    if pw is not None and not _is_base64(pw):
        prefs["password"] = _encode_password(pw)

    prefs["version"] = CURRENT_VERSION
    return prefs


def load_preferences() -> Dict[str, Any]:
    if not PREFERENCES_FILE.exists():
        return {}
    try:
        with open(PREFERENCES_FILE, "r", encoding="utf-8") as f:
            prefs = json.load(f)
    except (json.JSONDecodeError, OSError):
        return {}

    if prefs.get("version", 0) < CURRENT_VERSION:
        prefs = _migrate(prefs)
        save_preferences(prefs)

    return prefs


def save_preferences(prefs: Dict[str, Any]) -> None:
    _ensure_config_dir()
    with open(PREFERENCES_FILE, "w", encoding="utf-8") as f:
        json.dump(prefs, f, indent=2)
    os.chmod(PREFERENCES_FILE, 0o600)


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
    encoded = prefs.get("password")
    if encoded is None:
        return None
    try:
        return _decode_password(encoded)
    except Exception:
        return encoded


def set_preferred_password(password: str | None) -> None:
    prefs = load_preferences()
    if password is not None:
        prefs["password"] = _encode_password(password)
    elif "password" in prefs:
        del prefs["password"]
    save_preferences(prefs)
