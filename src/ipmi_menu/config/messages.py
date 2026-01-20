from __future__ import annotations

import json
from dataclasses import dataclass
from importlib import resources
from typing import Any, Dict, List, Tuple

from .settings import DEFAULT_LOCALE

# Available languages with their display names
AVAILABLE_LANGUAGES: List[Tuple[str, str]] = [
    ("fr", "Français"),
    ("en", "English"),
]


@dataclass(frozen=True)
class Messages:
    data: Dict[str, str]
    lang: str = DEFAULT_LOCALE

    def t(self, key: str, **fmt: Any) -> str:
        raw = self.data.get(key, key)
        try:
            return raw.format(**fmt)
        except Exception:
            return raw


def get_available_languages() -> List[Tuple[str, str]]:
    """Return list of available languages as (code, display_name) tuples."""
    return AVAILABLE_LANGUAGES


def load_messages(lang: str | None = None) -> Messages:
    lang = lang or DEFAULT_LOCALE
    filename = f"messages.{lang}.json"

    with resources.open_text("ipmi_menu.config", filename, encoding="utf-8") as f:
        data = json.load(f)

    return Messages(data=data, lang=lang)
