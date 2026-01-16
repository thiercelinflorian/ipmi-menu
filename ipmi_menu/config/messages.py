from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
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
            # Si un format foire, on renvoie la string brute
            return raw


def load_messages(locale: str = DEFAULT_LOCALE) -> Messages:
    base = Path(__file__).resolve().parent
    path = base / f"messages.{locale}.json"
    if not path.exists():
        # fallback fr
        path = base / "messages.fr.json"
    data = json.loads(path.read_text(encoding="utf-8"))
    return Messages(data=data)
