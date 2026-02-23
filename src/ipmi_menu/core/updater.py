"""Update checker for ipmi-menu."""
from __future__ import annotations

import json
import subprocess
import tempfile
import urllib.request
from shutil import which
from typing import Optional, Tuple

PYPI_URL = "https://pypi.org/pypi/ipmi-menu/json"
INSTALL_SCRIPT_URL = "https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/main/install.sh"


def get_current_version() -> str:
    """Get the currently installed version of ipmi-menu."""
    # Try importlib.metadata first (works when installed via pip/pipx)
    try:
        from importlib.metadata import version
        return version("ipmi-menu")
    except Exception:
        pass

    # Fallback: use __version__ from package
    try:
        from ipmi_menu import __version__
        return __version__
    except Exception:
        pass

    return "0.0.0"


def get_latest_version() -> Optional[str]:
    """Fetch the latest version from PyPI."""
    try:
        req = urllib.request.Request(PYPI_URL, headers={"Accept": "application/json"})
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.load(resp)
            return data.get("info", {}).get("version")
    except Exception:
        return None


def _parse_version(v: str) -> Tuple[int, ...]:
    """Parse version string into tuple of integers for comparison."""
    try:
        parts = v.split(".")
        return tuple(int(p) for p in parts[:3])
    except (ValueError, AttributeError):
        return (0, 0, 0)


def is_update_available() -> Tuple[bool, str, Optional[str]]:
    """
    Check if an update is available.

    Returns:
        Tuple of (update_available, current_version, latest_version)
        latest_version is None if check failed
    """
    current = get_current_version()
    latest = get_latest_version()

    if latest is None:
        return (False, current, None)

    current_tuple = _parse_version(current)
    latest_tuple = _parse_version(latest)

    return (latest_tuple > current_tuple, current, latest)


def run_upgrade() -> int:
    """Execute the upgrade command."""
    # Try pipx first
    if which("pipx"):
        rc = subprocess.call(["pipx", "upgrade", "ipmi-menu"])
        if rc == 0:
            return 0

    # Fallback: download install script to a temp file and run it
    try:
        with tempfile.NamedTemporaryFile(suffix=".sh", delete=False) as tmp:
            req = urllib.request.Request(INSTALL_SCRIPT_URL)
            with urllib.request.urlopen(req, timeout=30) as resp:
                tmp.write(resp.read())
            tmp_path = tmp.name
        return subprocess.call(["bash", tmp_path, "--upgrade"])
    except Exception:
        return 1
