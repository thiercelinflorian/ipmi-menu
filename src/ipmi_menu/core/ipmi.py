from __future__ import annotations

import re
import subprocess
from shutil import which
from typing import Dict, List, Optional, Tuple

from .utils import run_cmd


def has_ipmitool() -> bool:
    return which("ipmitool") is not None


def parse_kv(text: str) -> Dict[str, str]:
    out: Dict[str, str] = {}
    for ln in text.splitlines():
        if ":" in ln:
            k, v = ln.split(":", 1)
            out[k.strip().lower()] = v.strip()
    return out


def normalize_vendor(manufacturer: str, product: str, raw: str) -> str:
    blob = f"{manufacturer} {product} {raw}".lower()
    rules = [
        (r"supermicro|super micro", "supermicro"),
        (r"dell|idrac|poweredge", "dell"),
        (r"hp|hewlett|ilo", "hp"),
        (r"lenovo|thinksystem|xclarity", "lenovo"),
        (r"asrockrack|asrock", "asrockrack"),
        (r"intel", "intel"),
    ]
    for pat, v in rules:
        if re.search(pat, blob):
            return v
    return (manufacturer or "unknown").strip().lower() or "unknown"


def ipmi_base(host: str, user: str, password: Optional[str], interface: str, port: int) -> List[str]:
    cmd = ["ipmitool", "-I", interface, "-H", host, "-p", str(port), "-U", user]
    if password is not None:
        cmd += ["-P", password]
    return cmd


def ipmi(host: str, user: str, password: Optional[str], interface: str, port: int, timeout: int, args: List[str]) -> Tuple[int, str, str]:
    return run_cmd(ipmi_base(host, user, password, interface, port) + args, timeout)


def looks_like_auth_error(text: str) -> bool:
    t = (text or "").lower()
    needles = [
        "rakp 2 hmac is invalid",
        "unauthorized",
        "invalid user name",
        "invalid password",
        "password invalid",
        "authentication",
        "insufficient privilege",
        "privilege level",
        "access denied",
        "session timeout",
        "unable to establish",
        "no response from",
        "timeout",
    ]
    return any(n in t for n in needles)


def sol_activate(host: str, user: str, password: Optional[str], interface: str, port: int) -> int:
    cmd = ipmi_base(host, user, password, interface, port) + ["sol", "activate"]
    try:
        return subprocess.call(cmd)
    except FileNotFoundError:
        return 127


def power(host: str, user: str, password: Optional[str], interface: str, port: int, timeout: int, mode: str) -> Tuple[int, str, str]:
    mp = {
        "on": ["chassis", "power", "on"],
        "off": ["chassis", "power", "off"],
        "cycle": ["chassis", "power", "cycle"],
        "reset": ["chassis", "power", "reset"],
        "status": ["chassis", "power", "status"],
        "soft": ["chassis", "power", "soft"],
    }
    return ipmi(host, user, password, interface, port, timeout, mp[mode])


def bootdev(
    host: str,
    user: str,
    password: Optional[str],
    interface: str,
    port: int,
    timeout: int,
    device: str,
    *,
    uefi: bool,
    persistent: bool,
) -> Tuple[int, str, str]:
    opts: List[str] = []
    if persistent:
        opts.append("persistent")
    if uefi:
        opts.append("efiboot")

    args = ["chassis", "bootdev", device]
    if opts:
        args += [f"options={','.join(opts)}"]

    rc, out, err = ipmi(host, user, password, interface, port, timeout, args)

    # fallback si options non supportées
    if rc != 0 and opts:
        rc2, out2, err2 = ipmi(host, user, password, interface, port, timeout, args[:-1])
        if rc2 == 0:
            return rc2, out2, err2

    return rc, out, err


def ipmi_lan_print(host: str, user: str, password: Optional[str], interface: str, port: int, timeout: int) -> Tuple[int, str, str]:
    rc, out, err = ipmi(host, user, password, interface, port, timeout, ["lan", "print"])
    if rc == 0:
        return rc, out, err
    # certains BMC attendent un channel
    return ipmi(host, user, password, interface, port, timeout, ["lan", "print", "1"])


def ipmi_sdr_list(host: str, user: str, password: Optional[str], interface: str, port: int, timeout: int) -> Tuple[int, str, str]:
    rc, out, err = ipmi(host, user, password, interface, port, timeout, ["sdr", "list"])
    if rc == 0 and out:
        return rc, out, err
    return ipmi(host, user, password, interface, port, timeout, ["sdr", "list", "all"])
