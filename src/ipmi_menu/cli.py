#!/usr/bin/env python3
from __future__ import annotations

import sys
from typing import Optional

from ipmi_menu.config.messages import get_available_languages, load_messages
from ipmi_menu.config.preferences import get_preferred_language, set_preferred_language
from ipmi_menu.config.settings import (
    DEFAULT_INTERFACE,
    DEFAULT_PASSWORD,
    DEFAULT_PORT,
    DEFAULT_TIMEOUT,
    DEFAULT_USER,
)
from ipmi_menu.core.detect import detect
from ipmi_menu.core.ipmi import (
    has_ipmitool,
    ipmi,
    ipmi_lan_print,
    ipmi_sdr_list,
    looks_like_auth_error,
    power,
    sol_activate,
    bootdev,
)
from ipmi_menu.ui.prompts import confirm_critical, menu, yesno


def die(msg: str, code: int = 2) -> None:
    if msg:
        print(msg, file=sys.stderr)
    raise SystemExit(code)


def require_ipmi_ok(
    msg,
    host: str,
    user: str,
    password: Optional[str],
    interface: str,
    port: int,
    timeout: int,
) -> None:
    rc, out, err = ipmi(host, user, password, interface, port, timeout, ["mc", "info"])
    if rc == 0 and out:
        return

    blob = "\n".join([out or "", err or ""]).strip()
    if looks_like_auth_error(blob):
        die(msg.t("errors.ipmi_auth", details=blob), 1)
    die(msg.t("errors.ipmi_generic", details=blob or msg.t("errors.unknown")), 1)


def main() -> None:
    msg = load_messages(get_preferred_language())

    if not has_ipmitool():
        die(msg.t("errors.ipmitool_missing"))

    host = input(msg.t("prompts.bmc_ip")).strip()
    if not host:
        die(msg.t("errors.bmc_ip_required"))

    user = input(msg.t("prompts.user", default=DEFAULT_USER)).strip() or DEFAULT_USER

    password_in = input(msg.t("prompts.password"))
    if password_in == "":
        password: Optional[str] = DEFAULT_PASSWORD
        pw_mode = "default"
    elif password_in.strip() == "-":
        password = ""
        pw_mode = "empty"
    else:
        password = password_in.rstrip("\n")
        pw_mode = "custom"

    interface = DEFAULT_INTERFACE
    port = DEFAULT_PORT
    timeout = DEFAULT_TIMEOUT

    print(msg.t("info.connect_detect"))
    require_ipmi_ok(msg, host, user, password, interface, port, timeout)

    di = detect(host, user, password, interface, port, timeout)
    print(
        msg.t(
            "info.hw_detected",
            vendor=di.vendor,
            mfg=di.manufacturer or "-",
            product=di.product or "-",
        )
    )
    print(msg.t("info.auth", user=user, port=port, iface=interface, pw_mode=pw_mode))

    while True:
        action = menu(
            msg,
            "menu.action.title",
            [
                ("power", msg.t("menu.action.power")),
                ("sol", msg.t("menu.action.sol")),
                ("boot", msg.t("menu.action.boot")),
                ("info", msg.t("menu.action.info")),
                ("lang", msg.t("menu.action.language")),
                ("quit", msg.t("menu.action.quit")),
            ],
            5,
        )

        if action == "quit":
            raise SystemExit(0)

        if action == "lang":
            languages = get_available_languages()
            lang_options = [
                (code, f"{name} {msg.t('menu.language.current')}" if code == msg.lang else name)
                for code, name in languages
            ]
            lang_options.append(("home", msg.t("menu.home")))
            current_idx = next(
                (i for i, (code, _) in enumerate(languages) if code == msg.lang), 0
            )
            selected_lang = menu(msg, "menu.language.title", lang_options, current_idx)
            if selected_lang != "home" and selected_lang != msg.lang:
                set_preferred_language(selected_lang)
                msg = load_messages(selected_lang)
            continue

        if action == "info":
            print(msg.t("labels.info.sensors"))
            rc_s, out_s, err_s = ipmi_sdr_list(host, user, password, interface, port, timeout)
            if out_s:
                print(out_s)
            if rc_s != 0 and err_s:
                print(err_s, file=sys.stderr)

            print(msg.t("labels.info.misc"))
            for args in (["mc", "info"], ["fru", "print"]):
                rc, out, err = ipmi(host, user, password, interface, port, timeout, list(args))
                if out:
                    print(out)
                if rc != 0 and err:
                    print(err, file=sys.stderr)

            rc, out, err = ipmi_lan_print(host, user, password, interface, port, timeout)
            if out:
                print(out)
            if rc != 0 and err:
                print(err, file=sys.stderr)
            continue

        if action == "sol":
            print(msg.t("labels.sol_exit"))
            rc = sol_activate(host, user, password, interface, port)
            if rc != 0:
                print(msg.t("errors.ipmi_generic", details=msg.t("errors.unknown")), file=sys.stderr)
            continue

        if action == "power":
            mode = menu(
                msg,
                "menu.power.title",
                [
                    ("on", msg.t("menu.power.on")),
                    ("off", msg.t("menu.power.off")),
                    ("cycle", msg.t("menu.power.cycle")),
                    ("reset", msg.t("menu.power.reset")),
                    ("status", msg.t("menu.power.status")),
                    ("soft", msg.t("menu.power.soft")),
                    ("home", msg.t("menu.home")),
                ],
                4,
            )
            if mode == "home":
                continue

            if mode in {"off", "cycle", "reset"}:
                label_key = f"labels.critical.{mode}"
                if not confirm_critical(msg, msg.t(label_key)):
                    print(msg.t("errors.cancelled"))
                    continue

            rc, out, err = power(host, user, password, interface, port, timeout, mode)
            if out:
                print(out)
            if rc != 0 and err:
                print(err, file=sys.stderr)
            continue

        # bootdev
        device = menu(
            msg,
            "menu.bootdev.title",
            [
                ("pxe", msg.t("menu.bootdev.pxe")),
                ("disk", msg.t("menu.bootdev.disk")),
                ("cdrom", msg.t("menu.bootdev.cdrom")),
                ("bios", msg.t("menu.bootdev.bios")),
                ("safe", msg.t("menu.bootdev.safe")),
                ("none", msg.t("menu.bootdev.none")),
                ("home", msg.t("menu.home")),
            ],
            1,
        )
        if device == "home":
            continue

        boot_mode = menu(
            msg,
            "menu.bootmode.title",
            [
                ("uefi", msg.t("menu.bootmode.uefi")),
                ("legacy", msg.t("menu.bootmode.legacy")),
                ("home", msg.t("menu.home")),
            ],
            0,
        )
        if boot_mode == "home":
            continue

        persistent = yesno(msg, msg.t("labels.boot.persistent"), False)
        reboot = yesno(msg, msg.t("labels.boot.reboot_after"), True)

        rc, out, err = bootdev(
            host,
            user,
            password,
            interface,
            port,
            timeout,
            device,
            uefi=(boot_mode == "uefi"),
            persistent=persistent,
        )

        if out:
            print(out)
        if rc != 0:
            if err:
                print(err, file=sys.stderr)
            continue

        if reboot:
            reboot_mode = menu(
                msg,
                "menu.reboot.title",
                [
                    ("cycle", msg.t("menu.reboot.cycle")),
                    ("reset", msg.t("menu.reboot.reset")),
                    ("quit", msg.t("menu.reboot.quit")),
                    ("home", msg.t("menu.home")),
                ],
                0,
            )

            if reboot_mode in {"quit", "home"}:
                continue

            if not confirm_critical(msg, msg.t("labels.critical.reboot")):
                print(msg.t("errors.cancelled"))
                continue

            rc2, out2, err2 = power(host, user, password, interface, port, timeout, reboot_mode)
            if out2:
                print(out2)
            if rc2 != 0 and err2:
                print(err2, file=sys.stderr)
        continue


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        msg = load_messages()
        die(msg.t("errors.interrupted"), 130)
