from __future__ import annotations

import os
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path.cwd()
sys.path.insert(0, str(ROOT / "src"))
sys.path.insert(0, str(ROOT / "tests"))

from csk.builds import cache_windows  # noqa: E402
from csk.builds.cache import CacheEntryStatus, CacheExpectation  # noqa: E402
import test_build_cache_windows as helpers  # noqa: E402


def artifact_lookup_probe(root: Path) -> str:
    home, store = helpers._new_store(root)
    build_input = helpers._build_input()
    publication, _ = helpers._publication(root, build_input, b"artifact-race")
    store.publish(publication, guard=helpers._HeldGuard())
    late_link = home / "late-artifact-hard-link.exe"
    original = cache_windows._hash_handle

    def add_link_after_hash(
        handle: cache_windows._Handle,
        *,
        expected_size: int,
        label: str,
        error_factory: object,
    ) -> tuple[str, int]:
        result = original(
            handle,
            expected_size=expected_size,
            label=label,
            error_factory=error_factory,  # type: ignore[arg-type]
        )
        if label == "cache artifact" and not late_link.exists():
            os.link(handle.path, late_link)
        return result

    cache_windows._hash_handle = add_link_after_hash  # type: ignore[assignment]
    try:
        inspection = store.inspect(CacheExpectation(input=build_input))
    finally:
        cache_windows._hash_handle = original
    assert late_link.exists()
    return inspection.status.value


def receipt_lookup_probe(root: Path) -> str:
    home, store = helpers._new_store(root)
    build_input = helpers._build_input()
    publication, _ = helpers._publication(root, build_input, b"receipt-race")
    store.publish(publication, guard=helpers._HeldGuard())
    late_link = home / "late-receipt-hard-link"
    original = cache_windows._read_bounded_handle

    def add_link_after_read(
        handle: cache_windows._Handle,
        limit: int,
        label: str,
    ) -> bytes:
        result = original(handle, limit, label)
        if label == "cache receipt" and not late_link.exists():
            os.link(handle.path, late_link)
        return result

    cache_windows._read_bounded_handle = add_link_after_read
    try:
        inspection = store.inspect(CacheExpectation(input=build_input))
    finally:
        cache_windows._read_bounded_handle = original
    assert late_link.exists()
    return inspection.status.value


def publication_source_probe(root: Path) -> str:
    home, store = helpers._new_store(root)
    build_input = helpers._build_input()
    publication, _ = helpers._publication(root, build_input, b"source-race")
    late_link = publication.artifact_source.with_name("late-source-hard-link.exe")
    original = cache_windows._validate_source_unchanged

    def add_link_before_final_validation(source: cache_windows._Handle) -> None:
        if not late_link.exists():
            os.link(source.path, late_link)
        original(source)

    cache_windows._validate_source_unchanged = add_link_before_final_validation
    try:
        result = store.publish(publication, guard=helpers._HeldGuard())
    finally:
        cache_windows._validate_source_unchanged = original
    assert late_link.exists()
    assert home.exists()
    return result.status.value


def manager_home_inheritance_probe(root: Path) -> tuple[str, tuple[int, ...]]:
    home, store = helpers._new_store(root)
    subprocess.run(
        [
            "icacls",
            str(home),
            "/grant",
            "*S-1-1-0:(OI)(CI)(IO)F",
        ],
        check=True,
        capture_output=True,
        text=True,
    )
    with cache_windows._open_raw_handle(
        home,
        desired_access=(
            cache_windows._READ_CONTROL
            | cache_windows._FILE_READ_ATTRIBUTES
        ),
    ) as handle:
        snapshot = cache_windows._security_snapshot(handle)
    everyone_flags = tuple(
        ace.flags for ace in snapshot.aces if ace.sid == "S-1-1-0"
    )
    inspection = store.inspect(CacheExpectation(input=helpers._build_input()))
    return inspection.status.value, everyone_flags


def main() -> int:
    violations: list[str] = []
    with tempfile.TemporaryDirectory(
        prefix="TASK-260720-8nxlgx-review-"
    ) as temporary:
        root = Path(temporary)
        (root / "artifact").mkdir()
        try:
            artifact_status = artifact_lookup_probe(root / "artifact")
            print(f"late artifact hard link: inspect={artifact_status}")
            if artifact_status == CacheEntryStatus.HIT.value:
                violations.append("lookup admitted a multiply linked artifact")
        finally:
            helpers._make_cleanup_mutable(root)

    with tempfile.TemporaryDirectory(
        prefix="TASK-260720-8nxlgx-review-"
    ) as temporary:
        root = Path(temporary)
        (root / "receipt").mkdir()
        try:
            receipt_status = receipt_lookup_probe(root / "receipt")
            print(f"late receipt hard link: inspect={receipt_status}")
            if receipt_status == CacheEntryStatus.HIT.value:
                violations.append("lookup admitted a multiply linked receipt")
        finally:
            helpers._make_cleanup_mutable(root)

    with tempfile.TemporaryDirectory(
        prefix="TASK-260720-8nxlgx-review-"
    ) as temporary:
        root = Path(temporary)
        (root / "source").mkdir()
        try:
            publication_status = publication_source_probe(root / "source")
            print(f"late source hard link: publish={publication_status}")
            if publication_status in {"published", "reused-winner"}:
                violations.append(
                    "publication admitted a multiply linked artifact source"
                )
        finally:
            helpers._make_cleanup_mutable(root)

    with tempfile.TemporaryDirectory(
        prefix="TASK-260720-8nxlgx-review-"
    ) as temporary:
        root = Path(temporary)
        (root / "inheritance").mkdir()
        try:
            inheritance_status, everyone_flags = manager_home_inheritance_probe(
                root / "inheritance"
            )
            print(
                "inherit-only Everyone full-control ACE: "
                f"inspect={inheritance_status}, flags={everyone_flags}"
            )
            if inheritance_status != CacheEntryStatus.UNTRUSTED_PROVENANCE.value:
                violations.append(
                    "manager home accepted an inheritable untrusted mutation ACE"
                )
        finally:
            helpers._make_cleanup_mutable(root)

    for violation in violations:
        print(f"VIOLATION: {violation}")
    return 1 if violations else 0


if __name__ == "__main__":
    raise SystemExit(main())
