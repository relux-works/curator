from __future__ import annotations

import ctypes
import tempfile
from pathlib import Path
from types import SimpleNamespace
from typing import Any, Callable

from csk.builds import _windows, source


class _Function:
    def __init__(self, callback: Callable[..., Any]) -> None:
        self._callback = callback
        self.argtypes: object = None
        self.restype: object = None

    def __call__(self, *args: object) -> Any:
        return self._callback(*args)


def _set_stream_name(pointer: object, name: str) -> None:
    data = ctypes.cast(
        pointer,
        ctypes.POINTER(_windows._Win32FindStreamData),
    ).contents
    data.stream_name = name


def probe_root_recheck() -> None:
    original_open_root_fd = source._open_root_fd
    original_reject = source._reject_windows_named_streams
    calls: list[str] = []
    root_stream_present = False

    def reject(_path: Path, relative: str) -> None:
        calls.append(relative)
        if root_stream_present and relative == ".":
            raise source.InvalidSnapshotError("simulated root named stream")

    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory) / "snapshot"
        root.mkdir()
        (root / "file").write_bytes(b"content")
        source._open_root_fd = lambda _path: None
        source._reject_windows_named_streams = reject
        frozen = source.freeze_snapshot(root)
        calls_after_freeze = tuple(calls)
        root_stream_present = True
        calls.clear()
        failed_closed = False
        try:
            frozen.recheck()
        except source.SnapshotMutationError:
            failed_closed = True
        finally:
            frozen.close()
            source._open_root_fd = original_open_root_fd
            source._reject_windows_named_streams = original_reject

    print(f"root_checks_during_freeze={calls_after_freeze!r}")
    print(f"checks_during_recheck={tuple(calls)!r}")
    print(f"root_stream_mutation_failed_closed={failed_closed!r}")


def probe_find_close_failure() -> None:
    last_error = 38
    close_calls: list[int] = []

    def find_first(_path: object, _level: object, data: object, _flags: object) -> int:
        _set_stream_name(data, "::$DATA")
        return 42

    def find_next(_handle: object, _data: object) -> int:
        nonlocal last_error
        last_error = 38
        return 0

    def find_close(handle: int) -> int:
        nonlocal last_error
        close_calls.append(handle)
        last_error = 5
        return 0

    kernel32 = SimpleNamespace(
        FindFirstStreamW=_Function(find_first),
        FindNextStreamW=_Function(find_next),
        FindClose=_Function(find_close),
    )
    had_windll = hasattr(ctypes, "WinDLL")
    original_windll = getattr(ctypes, "WinDLL", None)
    had_get_last_error = hasattr(ctypes, "get_last_error")
    original_get_last_error = getattr(ctypes, "get_last_error", None)
    setattr(ctypes, "WinDLL", lambda _name, *, use_last_error: kernel32)
    setattr(ctypes, "get_last_error", lambda: last_error)
    failed_closed = False
    result: tuple[str, ...] | None = None
    try:
        try:
            result = _windows.named_data_streams(Path("C:/snapshot/file"))
        except OSError:
            failed_closed = True
    finally:
        if had_windll:
            setattr(ctypes, "WinDLL", original_windll)
        else:
            delattr(ctypes, "WinDLL")
        if had_get_last_error:
            setattr(ctypes, "get_last_error", original_get_last_error)
        else:
            delattr(ctypes, "get_last_error")

    print(f"find_close_calls={close_calls!r}")
    print(f"find_close_failure_result={result!r}")
    print(f"find_close_failure_failed_closed={failed_closed!r}")


if __name__ == "__main__":
    probe_root_recheck()
    probe_find_close_failure()
