from __future__ import annotations

from typing import List, Tuple

from ipmi_menu.config.messages import Messages


def menu(msg: Messages, title_key: str, items: List[Tuple[str, str]], default_idx: int = 0) -> str:
    print("\n" + msg.t(title_key))
    for i, (_, label) in enumerate(items, start=1):
        mark = " *" if (i - 1) == default_idx else ""
        print(f"  {i}) {label}{mark}")

    while True:
        v = input(msg.t("confirm.choice.prompt", max=len(items), default=default_idx + 1)).strip()
        if v == "":
            return items[default_idx][0]
        if v.isdigit():
            n = int(v)
            if 1 <= n <= len(items):
                return items[n - 1][0]
        print(msg.t("confirm.choice.invalid"))


def yesno(msg: Messages, label: str, default: bool = True) -> bool:
    yn = "Y/n" if default else "y/N"
    while True:
        v = input(msg.t("confirm.yesno.prompt", label=label, yn=yn)).strip().lower()
        if v == "":
            return default
        if v in {"y", "yes", "o", "oui"}:
            return True
        if v in {"n", "no", "non"}:
            return False
        print(msg.t("confirm.yesno.invalid"))


def confirm_critical(msg: Messages, label: str) -> bool:
    if not yesno(msg, label, False):
        return False
    v = input(msg.t("confirm.final")).strip()
    return v == "OUI"
