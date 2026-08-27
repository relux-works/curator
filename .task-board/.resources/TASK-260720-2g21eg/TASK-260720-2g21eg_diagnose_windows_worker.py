from __future__ import annotations

import tempfile
import time
from pathlib import Path

from csk.builds import go_v1


task_root = Path.cwd() / "TASK-260720-2g21eg-cycle5-61f29f"
manager = (task_root / "venv" / "Scripts" / "csk.exe").resolve(strict=True)
cache = Path(
    tempfile.mkdtemp(
        prefix="tester-worker-cache-",
        dir=task_root,
    )
)

started = time.monotonic()
identity = go_v1._resolve_manager_identity(manager)
print(f"identity_seconds={time.monotonic() - started:.3f}", flush=True)

limits = go_v1.ResourceLimits(timeout_seconds=30)
platform, probes = go_v1.probe_native_controls(limits)
print(
    f"probe_seconds={time.monotonic() - started:.3f} "
    f"platform={platform!r} probes={probes!r}",
    flush=True,
)

domain = go_v1._NativeControlDomain(platform, probes, limits)
process = None
try:
    print(f"argv={go_v1.worker_argv(identity)!r}", flush=True)
    print(
        "environment="
        f"{go_v1._indispensable_worker_environment(platform, cache, identity.startup.site_root, identity.startup.python_home)!r}",
        flush=True,
    )
    process = domain.launch(identity, cache)
    print(
        f"launch_seconds={time.monotonic() - started:.3f} pid={process.pid}",
        flush=True,
    )
    for wait_seconds in (1, 4, 15):
        time.sleep(wait_seconds)
        print(
            f"elapsed={time.monotonic() - started:.3f} "
            f"returncode={process.poll()!r}",
            flush=True,
        )
        if process.poll() is not None:
            break
    if process.poll() is None and process.stdin is not None:
        process.stdin.close()
        process.wait(timeout=15)
    stdout = process.stdout.read() if process.stdout is not None else b""
    stderr = process.stderr.read() if process.stderr is not None else b""
    print(
        f"terminal_seconds={time.monotonic() - started:.3f} "
        f"returncode={process.poll()!r}",
        flush=True,
    )
    print(f"stdout={stdout!r}", flush=True)
    print(f"stderr={stderr!r}", flush=True)
finally:
    try:
        domain.terminate(process)
    finally:
        domain.close()
