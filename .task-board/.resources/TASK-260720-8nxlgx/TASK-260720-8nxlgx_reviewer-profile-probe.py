"""Second reviewer probe: DACL inheritance stripping and source strictness.

Two questions the committed focused suite does not answer:

A. A manager home carrying an *inheritable but non-mutating* untrusted ACE is
   tolerated by design. Do the cache roots that ``publish`` creates below it
   actually end up with exactly the protected three-ACE profile, or does the
   inherited untrusted read grant survive into the live namespace?

B. The POSIX backend accepts an owner-readable-but-not-writable publication
   source (mode 0500). What does the Windows backend do with the equivalent
   sealed read/execute source?
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(os.environ.get("CSK_REPO_ROOT", Path.cwd())).resolve()
sys.path.insert(0, str(REPO_ROOT / "src"))
sys.path.insert(0, str(REPO_ROOT / "tests"))

import test_build_cache_windows as harness  # noqa: E402
from csk.builds import cache_windows  # noqa: E402
from csk.builds.cache import (  # noqa: E402
    BuildCacheError,
    CacheEntryStatus,
    CacheExpectation,
    CachePublication,
)

EVERYONE = "S-1-1-0"


def describe_dacl(path: Path) -> tuple[str, tuple[tuple[str, int, int], ...]]:
    with cache_windows._open_raw_handle(
        path,
        desired_access=(
            cache_windows._READ_CONTROL | cache_windows._FILE_READ_ATTRIBUTES
        ),
    ) as handle:
        snapshot = cache_windows._security_snapshot(handle)
    return snapshot.owner_sid, tuple(
        (ace.sid, ace.flags, ace.mask) for ace in snapshot.aces
    )


def probe_inheritable_untrusted_read(root: Path) -> bool:
    work = root / "inheritable-read"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    applied = subprocess.run(
        ["icacls", str(home), "/grant", f"*{EVERYONE}:(OI)(CI)(RX)"],
        check=False,
        capture_output=True,
        text=True,
    )
    if applied.returncode != 0:
        print(f"SKIP inheritable-read: icacls refused: {applied.stderr.strip()}")
        return True

    publication, _ = harness._publication(work, build_input, b"honest bytes")
    try:
        result = store.publish(publication, guard=harness._HeldGuard())
        outcome = result.status.value
    except BuildCacheError as exc:
        outcome = exc.code

    leaks: list[str] = []
    if outcome == "published":
        entry = harness._entry_path(home, build_input)
        for path in (
            home / cache_windows.LIVE_ROOT_NAME,
            home / cache_windows.LIVE_ROOT_NAME / "go-v1",
            home / cache_windows.STAGING_ROOT_NAME,
            home / cache_windows.QUARANTINE_ROOT_NAME,
            entry,
            entry / "bin",
            entry / cache_windows.RECEIPT_FILENAME,
            entry / Path(*build_input.artifact_path.split("/")),
        ):
            if not path.exists():
                continue
            owner, aces = describe_dacl(path)
            untrusted = [ace for ace in aces if ace[0] == EVERYONE]
            if untrusted:
                leaks.append(f"{path.name}: {untrusted}")
    inspection = store.inspect(CacheExpectation(input=build_input))
    ok = outcome == "published" and not leaks and (
        inspection.status is CacheEntryStatus.HIT
    )
    print(
        f"{'HOLD' if ok else 'VIOLATION'}: inheritable untrusted read ACE: "
        f"publish={outcome} inspect={inspection.status.value} "
        f"inherited_everyone_aces={leaks or 'none'}"
    )
    return ok


def probe_sealed_publication_source(root: Path) -> bool:
    work = root / "sealed-source"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    publication, _ = harness._publication(work, build_input, b"read only source")

    # Equivalent of a POSIX mode 0500 private build output: owner may read and
    # execute, nobody may write, read-only attribute set.
    harness._protect(publication.artifact_source, cache_windows._SEALED_ARTIFACT)

    outcome = ""
    detail = ""
    try:
        result = store.publish(publication, guard=harness._HeldGuard())
        outcome = result.status.value
    except BuildCacheError as exc:
        outcome = exc.code
        detail = exc.detail[:140]
    # Fails closed either way; the point is to record which contract applies.
    ok = outcome in {"published", "cache_publication_invalid"}
    print(
        f"{'HOLD' if ok else 'VIOLATION'}: sealed read/execute publication "
        f"source: publish={outcome} detail={detail!r}"
    )
    return ok


def main() -> int:
    if os.name != "nt":
        print("this probe requires a native Windows host", file=sys.stderr)
        return 2
    root = Path(tempfile.mkdtemp(prefix="reviewer-af18fc-profile-"))
    print(f"probe root: {root}", flush=True)
    try:
        results = [
            probe_inheritable_untrusted_read(root),
            probe_sealed_publication_source(root),
        ]
    finally:
        harness._make_cleanup_mutable(root)
        shutil.rmtree(root, ignore_errors=True)
    return 0 if all(results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
