from __future__ import annotations

import json
from dataclasses import dataclass
from importlib import resources
from typing import Any, Dict

from .settings import DEFAULT_LOCALE


@dataclass(frozen=True)
class Messages:
    data: Dict[str, str]

    def t(self, key: str, **fmt: Any) -> str:
        raw = self.data.get(key, key)
        try:
            return raw.format(**fmt)
        except Exception:
            return raw


def load_messages(lang: str | None = None) -> Messages:
    lang = lang or DEFAULT_LOCALE
    filename = f"messages.{lang}.json"

    with resources.open_text("ipmi_menu.config", filename, encoding="utf-8") as f:
        data = json.load(f)

    return Messages(data=data)
