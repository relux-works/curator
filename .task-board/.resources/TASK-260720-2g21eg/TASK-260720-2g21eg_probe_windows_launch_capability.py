from __future__ import annotations

import ctypes
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any, cast

from csk.builds import go_v1


TASK_ROOT = Path(__file__).resolve().parent
MANAGER = (TASK_ROOT / "venv" / "Scripts" / "csk.exe").resolve(strict=True)


def inspect_child() -> int:
    kernel32 = go_v1._windows_kernel32()
    get_file_type = kernel32.GetFileType
    get_file_type.argtypes = [ctypes.c_void_p]
    get_file_type.restype = ctypes.c_uint32
    peek = kernel32.PeekNamedPipe
    peek.argtypes = [
        ctypes.c_void_p,
        ctypes.c_void_p,
        ctypes.c_uint32,
        ctypes.POINTER(ctypes.c_uint32),
        ctypes.POINTER(ctypes.c_uint32),
        ctypes.POINTER(ctypes.c_uint32),
    ]
    peek.restype = ctypes.c_int
    records: list[tuple[int, int, int, bytes, int]] = []
    for value in range(4, go_v1._MAX_WINDOWS_HANDLE_SCAN, 4):
        handle = ctypes.c_void_p(value)
        if get_file_type(handle) != 3:
            continue
        prefix = ctypes.create_string_buffer(len(go_v1._WORKER_LAUNCH_MAGIC))
        read = ctypes.c_uint32()
        available = ctypes.c_uint32()
        result = int(
            peek(
                handle,
                prefix,
                len(prefix),
                ctypes.byref(read),
                ctypes.byref(available),
                None,
            )
        )
        records.append(
            (
                value,
                result,
                available.value,
                prefix.raw[: read.value],
                ctypes.get_last_error(),  # type: ignore[attr-defined]
            )
        )
    print(
        f"child_pid={os.getpid()} parent_pid={os.getppid()} pipes={records!r}",
        flush=True,
    )
    try:
        context = go_v1._consume_worker_launch_context()
    except BaseException as exc:
        print(
            f"consume_error={type(exc).__name__}: {exc!s} "
            f"cause={exc.__cause__!r}",
            flush=True,
        )
        return 3
    print(f"context={context.public_dict()!r}", flush=True)
    return 0


def run_parent() -> int:
    identity = go_v1._resolve_manager_identity(MANAGER)
    use_base = sys.argv[1:] == ["--base"]
    selected_interpreter = (
        identity.interpreter.runtime.base_executable.path
        if use_base
        else identity.interpreter.invocation_path
    )
    cache = Path(
        tempfile.mkdtemp(
            prefix="tester-capability-cache-",
            dir=TASK_ROOT,
        )
    )
    environment = go_v1._indispensable_worker_environment(
        go_v1.PLATFORM_WINDOWS,
        cache,
        identity.startup.site_root,
        identity.startup.python_home,
    )
    capability = go_v1._PreparedWorkerLaunch(go_v1.PLATFORM_WINDOWS)
    options: dict[str, object] = {
        "cwd": MANAGER.parent,
        "env": environment,
        "stdin": subprocess.DEVNULL,
        "stdout": subprocess.PIPE,
        "stderr": subprocess.PIPE,
        "close_fds": True,
        "shell": False,
    }
    capability.add_popen_options(options)
    print(
        f"parent_pid={os.getpid()} capability_handle="
        f"{capability._windows_handle} "
        f"selected_interpreter={selected_interpreter!s}",
        flush=True,
    )
    try:
        process = subprocess.Popen(
            (
                os.fspath(selected_interpreter),
                *go_v1.WORKER_LAUNCH_FLAGS,
                os.fspath(Path(__file__).resolve()),
                "--child",
            ),
            **cast(Any, options),
        )
    finally:
        capability.close_parent_copy()
    print(f"popen_pid={process.pid}", flush=True)
    stdout, stderr = process.communicate(timeout=30)
    print(f"returncode={process.returncode}", flush=True)
    print(f"stdout={stdout!r}", flush=True)
    print(f"stderr={stderr!r}", flush=True)
    return process.returncode


if __name__ == "__main__":
    raise SystemExit(
        inspect_child()
        if sys.argv[1:] == ["--child"]
        else run_parent()
    )
