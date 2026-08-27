from __future__ import annotations

import sys


def main() -> int:
    print(f"sys_argv={sys.argv!r}", flush=True)
    print(f"sys_orig_argv={sys.orig_argv!r}", flush=True)
    print(f"sys_executable={sys.executable!r}", flush=True)
    return 0
