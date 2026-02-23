from __future__ import annotations

import logging
import subprocess
from typing import List, Optional, Tuple

logger = logging.getLogger("ipmi_menu")


def _sanitize_cmd(cmd: List[str]) -> List[str]:
    """Replace the password value after -P flag with '****'."""
    sanitized: List[str] = []
    skip_next = False
    for i, arg in enumerate(cmd):
        if skip_next:
            sanitized.append("****")
            skip_next = False
        elif arg == "-P" and i + 1 < len(cmd):
            sanitized.append(arg)
            skip_next = True
        else:
            sanitized.append(arg)
    return sanitized


def run_cmd(cmd: List[str], timeout: Optional[int]) -> Tuple[int, str, str]:
    logger.debug("Running: %s (timeout=%s)", " ".join(_sanitize_cmd(cmd)), timeout)
    try:
        p = subprocess.run(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            timeout=timeout,
        )
        return p.returncode, (p.stdout or "").strip(), (p.stderr or "").strip()
    except subprocess.TimeoutExpired:
        logger.warning("Command timed out after %ss: %s", timeout, " ".join(_sanitize_cmd(cmd)))
        return 124, "", "timeout"
    except FileNotFoundError:
        logger.error("Command not found: %s", cmd[0] if cmd else "<empty>")
        return 127, "", "command not found"
