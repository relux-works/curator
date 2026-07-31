from __future__ import annotations

import importlib.util
import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

from csk.builds import cache_posix
from csk.builds.cache import BuildCacheError, CacheExpectation
from csk.builds.cache_posix import PosixBuildCache


def load_test_helpers() -> object:
    path = Path(
        "/Users/iv/Developer/Wildberries/cocoaskills/"
        ".temp/TASK-260720-2jfnz6/worktree/tests/test_build_cache_posix.py"
    )
    spec = importlib.util.spec_from_file_location("cache_test_helpers", path)
    if spec is None or spec.loader is None:
        raise RuntimeError("could not load focused cache test helpers")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


helpers = load_test_helpers()
held_guard = helpers._HeldGuard()
build_input = helpers._build_input()

with tempfile.TemporaryDirectory(
    prefix="TASK-260720-2jfnz6-race-probe-"
) as temporary:
    root = Path(temporary)
    home, _store = helpers._new_store(root)
    publication, _receipt_hash = helpers._publication(
        root,
        build_input,
        b"identical artifact",
    )
    stores = (PosixBuildCache(home), PosixBuildCache(home))

    first_in_seal_window = threading.Event()
    release_first = threading.Event()
    call_lock = threading.Lock()
    seal_calls = 0
    original_seal = cache_posix._seal_published_entry

    def controlled_seal(driver_fd: int, entry_name: str) -> None:
        global seal_calls
        with call_lock:
            seal_calls += 1
            call_number = seal_calls
        if call_number == 1:
            first_in_seal_window.set()
            if not release_first.wait(timeout=10):
                raise TimeoutError("second publisher did not finish")
        original_seal(driver_fd, entry_name)

    cache_posix._seal_published_entry = controlled_seal
    try:
        with ThreadPoolExecutor(max_workers=2) as executor:
            first = executor.submit(
                stores[0].publish,
                publication,
                guard=held_guard,
            )
            if not first_in_seal_window.wait(timeout=10):
                raise TimeoutError("first publisher did not reach seal window")
            second = executor.submit(
                stores[1].publish,
                publication,
                guard=held_guard,
            )
            try:
                second_outcome = second.result(timeout=10)
            except BaseException as exc:
                second_outcome = exc
            finally:
                release_first.set()
            try:
                first_outcome = first.result(timeout=10)
            except BaseException as exc:
                first_outcome = exc
    finally:
        cache_posix._seal_published_entry = original_seal
        release_first.set()

    quarantine_names = sorted(
        path.name for path in (home / cache_posix.QUARANTINE_ROOT_NAME).iterdir()
    )

    def describe(outcome: object) -> str:
        if isinstance(outcome, BuildCacheError):
            return f"error:{outcome.code}:{outcome.detail}"
        status = getattr(outcome, "status", None)
        value = getattr(status, "value", None)
        return f"result:{value}"

    inspection = stores[0].inspect(CacheExpectation(input=build_input))
    print(f"first_outcome={describe(first_outcome)}")
    print(f"second_outcome={describe(second_outcome)}")
    print(f"final_inspection={inspection.status.value}")
    print(f"seal_calls={seal_calls}")
    print(f"quarantine_count={len(quarantine_names)}")
    print(f"quarantine_names={quarantine_names}")
