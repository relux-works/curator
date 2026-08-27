from __future__ import annotations

import runpy
import sys
from pathlib import Path


task_root = Path(__file__).resolve().parent
site_root = task_root / "venv" / "Lib" / "site-packages"
manager = (task_root / "venv" / "Scripts" / "csk.exe").resolve(strict=True)
sys.path.insert(0, str(site_root))

import csk.cli  # noqa: E402


def diagnostic_main(argv: list[str] | None = None) -> int:
    print(f"main_argv={argv!r}", flush=True)
    print(f"sys_argv={sys.argv!r}", flush=True)
    print(f"sys_orig_argv={getattr(sys, 'orig_argv', None)!r}", flush=True)
    print(f"sys_executable={sys.executable!r}", flush=True)
    print(f"manager_exists={manager.exists()!r}", flush=True)
    return 0


csk.cli.main = diagnostic_main
try:
    runpy.run_path(str(manager), run_name="__main__")
except SystemExit as exc:
    print(f"system_exit={exc.code!r}", flush=True)
