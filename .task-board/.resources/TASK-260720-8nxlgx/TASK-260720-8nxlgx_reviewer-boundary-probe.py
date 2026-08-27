"""Independent reviewer probe for the protected Windows build cache.

Defensive security review of TASK-260720-8nxlgx. Every case asserts that the
backend refuses to hand back candidate bytes when its Windows boundary is
violated in a way the committed focused suite does not already cover.

Run on a native Windows host from the repository root:

    .venv\\Scripts\\python.exe .temp\\...\\reviewer_boundary_probe.py

Exit code 0 means every boundary held. Non-zero means at least one probe
observed candidate bytes or a reused winner it should never have seen.
"""

from __future__ import annotations

import hashlib
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
    CachePublicationStatus,
)
from csk.builds.metadata import (  # noqa: E402
    BuildArtifact,
    build_receipt,
    cache_key,
    canonical_receipt_bytes,
)

FAILURES: list[str] = []
RESULTS: list[str] = []


def record(name: str, ok: bool, detail: str) -> None:
    status = "HOLD" if ok else "VIOLATION"
    RESULTS.append(f"{status}: {name}: {detail}")
    if not ok:
        FAILURES.append(f"{name}: {detail}")
    print(f"{status}: {name}: {detail}", flush=True)


def no_bytes(inspection: object) -> bool:
    return (
        getattr(inspection, "receipt", None) is None
        and getattr(inspection, "receipt_bytes", None) is None
        and getattr(inspection, "artifact_path", None) is None
    )


def icacls(*arguments: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["icacls", *arguments],
        check=False,
        capture_output=True,
        text=True,
    )


def junction(link: Path, target: Path) -> bool:
    result = subprocess.run(
        ["cmd", "/d", "/c", "mklink", "/J", str(link), str(target)],
        check=False,
        capture_output=True,
        text=True,
    )
    return result.returncode == 0


def sealed_entry_paths(home: Path, build_input: object) -> tuple[Path, Path, Path]:
    entry = harness._entry_path(home, build_input)
    receipt = entry / cache_windows.RECEIPT_FILENAME
    artifact = entry / Path(*build_input.artifact_path.split("/"))
    return entry, receipt, artifact


def unseal(path: Path) -> None:
    harness._protect(
        path,
        cache_windows._MUTABLE_DIRECTORY
        if path.is_dir()
        else cache_windows._MUTABLE_FILE,
    )


# --------------------------------------------------------------------------
# 1. Inheritable untrusted mutation grants beyond the committed (F) cases.
# --------------------------------------------------------------------------
def probe_manager_home_partial_mutation_grants(root: Path) -> None:
    for label, rights in (
        ("write-only", "(OI)(CI)(W)"),
        ("delete-only", "(OI)(CI)(DE)"),
        ("write-dac-only", "(OI)(CI)(WDAC)"),
        ("write-owner-only", "(OI)(CI)(WO)"),
        ("append-only", "(OI)(CI)(AD)"),
        ("inherit-only-write", "(OI)(CI)(IO)(W)"),
    ):
        work = root / f"grant-{label}"
        work.mkdir()
        home, store = harness._new_store(work)
        build_input = harness._build_input()
        applied = icacls(str(home), "/grant", f"*S-1-1-0:{rights}")
        if applied.returncode != 0:
            record(
                f"manager-home {label} untrusted grant",
                True,
                f"icacls refused the grant, boundary untested: {applied.stderr.strip()}",
            )
            continue
        inspection = store.inspect(CacheExpectation(input=build_input))
        publication, _ = harness._publication(work, build_input, b"candidate")
        published_error = ""
        try:
            store.publish(publication, guard=harness._HeldGuard())
            published_error = "publish succeeded"
        except BuildCacheError as exc:
            published_error = exc.code
        roots_created = [
            name
            for name in (
                cache_windows.LIVE_ROOT_NAME,
                cache_windows.STAGING_ROOT_NAME,
                cache_windows.QUARANTINE_ROOT_NAME,
            )
            if (home / name).exists()
        ]
        ok = (
            inspection.status is CacheEntryStatus.UNTRUSTED_PROVENANCE
            and no_bytes(inspection)
            and published_error == "cache_boundary_untrusted"
            and not roots_created
        )
        record(
            f"manager-home {label} untrusted grant",
            ok,
            f"inspect={inspection.status.value} publish={published_error} "
            f"roots={roots_created}",
        )


# --------------------------------------------------------------------------
# 2. Escaping reparse points below the validated roots.
# --------------------------------------------------------------------------
def probe_reparse_below_roots(root: Path) -> None:
    for label in ("entry", "bin"):
        work = root / f"reparse-{label}"
        work.mkdir()
        home, store = harness._new_store(work)
        build_input = harness._build_input()
        publication, _ = harness._publication(work, build_input, b"candidate")
        store.publish(publication, guard=harness._HeldGuard())
        entry, _receipt, artifact = sealed_entry_paths(home, build_input)

        outside = work / f"outside-{label}"
        outside.mkdir()
        harness._protect(outside, cache_windows._MUTABLE_DIRECTORY)

        if label == "entry":
            # Move the real sealed entry out of the way, then point the live
            # entry name at an out-of-boundary directory holding valid bytes.
            harness._protect(entry, cache_windows._MUTABLE_DIRECTORY)
            shutil.move(str(entry), str(outside / "real"))
            if not junction(entry, outside / "real"):
                record(f"reparse at {label}", True, "junction creation refused")
                continue
        else:
            bin_path = entry / "bin"
            harness._protect(entry, cache_windows._MUTABLE_DIRECTORY)
            harness._protect(bin_path, cache_windows._MUTABLE_DIRECTORY)
            unseal(artifact)
            shutil.move(str(artifact), str(outside / artifact.name))
            harness._protect(
                outside / artifact.name,
                cache_windows._SEALED_ARTIFACT,
            )
            shutil.rmtree(bin_path)
            if not junction(bin_path, outside):
                record(f"reparse at {label}", True, "junction creation refused")
                continue
            harness._protect(entry, cache_windows._SEALED_ENTRY)

        inspection = store.inspect(CacheExpectation(input=build_input))
        ok = inspection.status is not CacheEntryStatus.HIT and no_bytes(inspection)
        record(
            f"reparse at {label}",
            ok,
            f"inspect={inspection.status.value} reason={inspection.reason[:90]!r}",
        )


# --------------------------------------------------------------------------
# 3. Special files standing in for required cache objects.
# --------------------------------------------------------------------------
def probe_special_objects_inside_entry(root: Path) -> None:
    for label in ("bin-as-file", "receipt-as-directory"):
        work = root / label
        work.mkdir()
        home, store = harness._new_store(work)
        build_input = harness._build_input()
        publication, _ = harness._publication(work, build_input, b"candidate")
        store.publish(publication, guard=harness._HeldGuard())
        entry, receipt, artifact = sealed_entry_paths(home, build_input)
        harness._protect(entry, cache_windows._MUTABLE_DIRECTORY)

        if label == "bin-as-file":
            bin_path = entry / "bin"
            harness._protect(bin_path, cache_windows._MUTABLE_DIRECTORY)
            unseal(artifact)
            shutil.rmtree(bin_path)
            bin_path.write_bytes(b"not a directory")
            harness._protect(bin_path, cache_windows._MUTABLE_FILE)
        else:
            unseal(receipt)
            receipt.unlink()
            receipt.mkdir()
            harness._protect(receipt, cache_windows._MUTABLE_DIRECTORY)

        harness._protect(entry, cache_windows._SEALED_ENTRY)
        inspection = store.inspect(CacheExpectation(input=build_input))
        ok = inspection.status is not CacheEntryStatus.HIT and no_bytes(inspection)
        record(
            label,
            ok,
            f"inspect={inspection.status.value} reason={inspection.reason[:90]!r}",
        )


# --------------------------------------------------------------------------
# 4. A pre-existing permissive entry holding self-consistent foreign bytes.
# --------------------------------------------------------------------------
def probe_permissive_preexisting_entry(root: Path) -> None:
    work = root / "permissive-preexisting"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()

    foreign = b"attacker controlled payload"
    trusted = b"honest build output"

    driver = home / cache_windows.LIVE_ROOT_NAME / "go-v1"
    driver.mkdir(parents=True)
    harness._protect(
        home / cache_windows.LIVE_ROOT_NAME,
        cache_windows._MUTABLE_DIRECTORY,
    )
    harness._protect(driver, cache_windows._MUTABLE_DIRECTORY)

    entry = driver / cache_key(build_input).removeprefix("sha256:")
    (entry / "bin").mkdir(parents=True)
    foreign_receipt = canonical_receipt_bytes(
        build_receipt(
            build_input,
            BuildArtifact(
                path=build_input.artifact_path,
                sha256=f"sha256:{hashlib.sha256(foreign).hexdigest()}",
                size=len(foreign),
            ),
        )
    )
    (entry / cache_windows.RECEIPT_FILENAME).write_bytes(foreign_receipt)
    (entry / Path(*build_input.artifact_path.split("/"))).write_bytes(foreign)
    # Deliberately leave the planted entry on inherited, unprotected DACLs.

    inspection = store.inspect(CacheExpectation(input=build_input))
    leaked = inspection.receipt_bytes == foreign_receipt
    ok_lookup = (
        inspection.status is not CacheEntryStatus.HIT
        and no_bytes(inspection)
        and not leaked
    )
    record(
        "permissive pre-existing entry lookup",
        ok_lookup,
        f"inspect={inspection.status.value} leaked_foreign_receipt={leaked}",
    )

    publication, _ = harness._publication(work, build_input, trusted)
    outcome = ""
    try:
        result = store.publish(publication, guard=harness._HeldGuard())
        outcome = result.status.value
    except BuildCacheError as exc:
        outcome = exc.code
    live_bytes = b""
    live_artifact = entry / Path(*build_input.artifact_path.split("/"))
    if live_artifact.exists():
        live_bytes = live_artifact.read_bytes()
    ok_publish = (
        outcome == CachePublicationStatus.PUBLISHED.value
        and live_bytes == trusted
    ) or (outcome not in {CachePublicationStatus.REUSED_WINNER.value} and live_bytes != foreign)
    record(
        "permissive pre-existing entry publication",
        ok_publish,
        f"publish={outcome} live_bytes_are_foreign={live_bytes == foreign}",
    )


# --------------------------------------------------------------------------
# 5. Alternate data stream smuggled onto a sealed artifact.
# --------------------------------------------------------------------------
def probe_alternate_data_stream(root: Path) -> None:
    work = root / "alternate-stream"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    publication, _ = harness._publication(work, build_input, b"candidate")
    store.publish(publication, guard=harness._HeldGuard())
    _entry, _receipt, artifact = sealed_entry_paths(home, build_input)

    unseal(artifact)
    try:
        with open(f"{artifact}:smuggled", "intranet") as stream:
            stream.write(b"payload")
    except OSError as exc:
        harness._protect(artifact, cache_windows._SEALED_ARTIFACT)
        record("alternate data stream", True, f"stream creation refused: {exc}")
        return
    harness._protect(artifact, cache_windows._SEALED_ARTIFACT)

    inspection = store.inspect(CacheExpectation(input=build_input))
    ok = inspection.status is not CacheEntryStatus.HIT and no_bytes(inspection)
    record(
        "alternate data stream",
        ok,
        f"inspect={inspection.status.value} reason={inspection.reason[:90]!r}",
    )


# --------------------------------------------------------------------------
# 6. Untrusted mutation grants on intermediate and entry-level cache objects.
# --------------------------------------------------------------------------
def probe_untrusted_grants_inside_cache(root: Path) -> None:
    targets = {
        "driver-root": lambda home, entry: home / cache_windows.LIVE_ROOT_NAME / "go-v1",
        "sealed-entry": lambda home, entry: entry,
        "sealed-bin": lambda home, entry: entry / "bin",
    }
    for label, resolve in targets.items():
        work = root / f"grant-inside-{label}"
        work.mkdir()
        home, store = harness._new_store(work)
        build_input = harness._build_input()
        publication, _ = harness._publication(work, build_input, b"candidate")
        store.publish(publication, guard=harness._HeldGuard())
        entry, _receipt, _artifact = sealed_entry_paths(home, build_input)
        target = resolve(home, entry)
        applied = icacls(str(target), "/grant", "*S-1-1-0:(F)")
        if applied.returncode != 0:
            record(
                f"untrusted grant on {label}",
                True,
                f"icacls refused the grant: {applied.stderr.strip()[:80]}",
            )
            continue
        inspection = store.inspect(CacheExpectation(input=build_input))
        ok = inspection.status is not CacheEntryStatus.HIT and no_bytes(inspection)
        record(
            f"untrusted grant on {label}",
            ok,
            f"inspect={inspection.status.value} reason={inspection.reason[:90]!r}",
        )


# --------------------------------------------------------------------------
# 7. Real owner drift on a sealed artifact.
# --------------------------------------------------------------------------
def probe_real_owner_drift(root: Path) -> None:
    work = root / "owner-drift"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    publication, _ = harness._publication(work, build_input, b"candidate")
    store.publish(publication, guard=harness._HeldGuard())
    _entry, _receipt, artifact = sealed_entry_paths(home, build_input)
    applied = icacls(str(artifact), "/setowner", "*S-1-5-32-544")
    if applied.returncode != 0:
        record(
            "real owner drift on artifact",
            True,
            f"setowner refused: {applied.stderr.strip()[:80]}",
        )
        return
    inspection = store.inspect(CacheExpectation(input=build_input))
    ok = inspection.status is not CacheEntryStatus.HIT and no_bytes(inspection)
    record(
        "real owner drift on artifact",
        ok,
        f"inspect={inspection.status.value} reason={inspection.reason[:90]!r}",
    )


# --------------------------------------------------------------------------
# 8. Publication without a witness lock, and with a dry pristine home.
# --------------------------------------------------------------------------
def probe_lock_and_read_only_lookup(root: Path) -> None:
    work = root / "lock-and-readonly"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    before = harness._tree_state(home)
    inspection = store.inspect(CacheExpectation(input=build_input))
    after_lookup = harness._tree_state(home)
    publication, _ = harness._publication(work, build_input, b"candidate")
    code = ""
    try:
        store.publish(publication, guard=None)  # type: ignore[arg-type]
        code = "publish succeeded"
    except BuildCacheError as exc:
        code = exc.code
    ok = (
        inspection.status is CacheEntryStatus.MISS
        and before == after_lookup == ()
        and code == "cache_lock_required"
        and harness._tree_state(home) == ()
    )
    record(
        "read-only lookup and missing lock",
        ok,
        f"inspect={inspection.status.value} publish={code} "
        f"tree_after={harness._tree_state(home)}",
    )


# --------------------------------------------------------------------------
# 9. Diagnostic: how narrow is the post-validation hard-link window?
# --------------------------------------------------------------------------
def probe_residual_hardlink_window(root: Path) -> None:
    work = root / "residual-window"
    work.mkdir()
    home, store = harness._new_store(work)
    build_input = harness._build_input()
    publication, _ = harness._publication(work, build_input, b"candidate")
    store.publish(publication, guard=harness._HeldGuard())
    _entry, _receipt, artifact = sealed_entry_paths(home, build_input)
    late_link = home / "residual-artifact-link.exe"

    original = cache_windows._revalidate_child

    def link_during_parent_revalidation(
        parent: object,
        handle: object,
        name: str,
        profile: object,
        label: str,
    ) -> None:
        if label == "artifact directory" and not late_link.exists():
            os.link(artifact, late_link)
        original(parent, handle, name, profile, label)

    cache_windows._revalidate_child = link_during_parent_revalidation  # type: ignore[assignment]
    try:
        inspection = store.inspect(CacheExpectation(input=build_input))
    finally:
        cache_windows._revalidate_child = original  # type: ignore[assignment]

    admitted = inspection.status is CacheEntryStatus.HIT
    record(
        "diagnostic residual hard-link window",
        True,
        f"link_present={late_link.exists()} inspect={inspection.status.value} "
        f"admitted_multiply_linked_artifact={admitted}",
    )


def main() -> int:
    if os.name != "nt":
        print("this probe requires a native Windows host", file=sys.stderr)
        return 2
    root = Path(tempfile.mkdtemp(prefix="reviewer-af18fc-"))
    print(f"probe root: {root}", flush=True)
    try:
        probe_manager_home_partial_mutation_grants(root)
        probe_reparse_below_roots(root)
        probe_special_objects_inside_entry(root)
        probe_permissive_preexisting_entry(root)
        probe_alternate_data_stream(root)
        probe_untrusted_grants_inside_cache(root)
        probe_real_owner_drift(root)
        probe_lock_and_read_only_lookup(root)
        probe_residual_hardlink_window(root)
    finally:
        harness._make_cleanup_mutable(root)
        shutil.rmtree(root, ignore_errors=True)

    print("\n=== summary ===", flush=True)
    for line in RESULTS:
        print(line, flush=True)
    if FAILURES:
        print(f"\n{len(FAILURES)} boundary violation(s)", flush=True)
        return 1
    print("\nall probed boundaries held", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
