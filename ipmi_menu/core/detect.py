from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

from .ipmi import ipmi, normalize_vendor, parse_kv


@dataclass
class DetectInfo:
    vendor: str
    manufacturer: str
    product: str


def detect(host: str, user: str, password: Optional[str], interface: str, port: int, timeout: int) -> DetectInfo:
    manufacturer = ""
    product = ""
    raw = ""

    rc, out, _ = ipmi(host, user, password, interface, port, timeout, ["mc", "info"])
    if rc == 0 and out:
        raw += "\n" + out
        kv = parse_kv(out)
        manufacturer = kv.get("manufacturer", manufacturer)
        product = kv.get("product name", product)

    rc2, out2, _ = ipmi(host, user, password, interface, port, timeout, ["fru", "print"])
    if rc2 == 0 and out2:
        raw += "\n" + out2
        kv2 = parse_kv(out2)
        manufacturer = manufacturer or kv2.get("board mfg", "") or kv2.get("product manufacturer", "")
        product = product or kv2.get("board product", "") or kv2.get("product name", "")

    vendor = normalize_vendor(manufacturer, product, raw)
    return DetectInfo(vendor=vendor, manufacturer=manufacturer, product=product)
