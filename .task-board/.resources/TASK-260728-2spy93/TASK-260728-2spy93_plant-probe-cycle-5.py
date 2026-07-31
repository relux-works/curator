"""Plant each cycle-5 widening into the shipped common schema and revert it.

For each plant: run the real CLI (``python tools/validate.py``) and record its
exit code and rejecting message, then ask each earlier validation layer in
isolation whether it would have caught the same planted file. The file is
restored from the pre-plant bytes and verified by SHA-256 after every plant.
"""

from __future__ import annotations

import hashlib
import importlib.util
import json
import subprocess
import sys
from pathlib import Path

WORKTREE = Path(__file__).resolve().parent / "curator-spec-worktree"
COMMON = WORKTREE / "schemas" / "v1" / "common.schema.json"
SPEC = importlib.util.spec_from_file_location(
    "plant_validate", WORKTREE / "tools" / "validate.py"
)
assert SPEC is not None and SPEC.loader is not None
validate = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(validate)

LAYERS = (
    "validate_schemas",
    "validate_manifest",
    "validate_vector_semantics",
    "validate_additional_driver_boundary",
)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def plants() -> list[tuple[str, dict, str]]:
    """(label, mutated document, layer expected to reject it on disk).

    The reviewer's third mutant raises ``maximum`` past the CCJ-1 safe integer,
    which ``load_json`` refuses to parse at all, so on disk it never reaches the
    boundary gate. That is an honest pre-existing guard rather than this fix, so
    the plant is kept and attributed to the layer that actually rejects it, and
    a fourth plant removes ``maximum`` instead — the same widening expressed in
    a form the parse guard cannot see.
    """
    document = json.loads(COMMON.read_text(encoding="utf-8"))
    out: list[tuple[str, dict, str]] = []
    for label, target, keyword, value, layer in (
        (
            "portablePath.maxLength 4096 -> 4097",
            "portablePath",
            "maxLength",
            4097,
            "validate_additional_driver_boundary",
        ),
        (
            "sha256.pattern admits uppercase hex",
            "sha256",
            "pattern",
            "^sha256:[0-9a-fA-F]{64}$",
            "validate_additional_driver_boundary",
        ),
        (
            "nonNegativeSafeInteger.maximum + 1",
            "nonNegativeSafeInteger",
            "maximum",
            9007199254740992,
            "validate_schemas",
        ),
    ):
        mutated = json.loads(json.dumps(document))
        mutated["$defs"][target][keyword] = value
        out.append((label, mutated, layer))

    unbounded = json.loads(json.dumps(document))
    del unbounded["$defs"]["nonNegativeSafeInteger"]["maximum"]
    out.append(
        (
            "nonNegativeSafeInteger.maximum removed",
            unbounded,
            "validate_additional_driver_boundary",
        )
    )
    return out


def main() -> int:
    original_bytes = COMMON.read_bytes()
    original_digest = digest(COMMON)
    print(f"pre-plant common.schema.json sha256 = {original_digest}")
    ok = True
    for label, mutated, expected_layer in plants():
        COMMON.write_text(json.dumps(mutated, indent=2) + "\n", encoding="utf-8")
        try:
            completed = subprocess.run(
                [sys.executable, "tools/validate.py"],
                cwd=WORKTREE,
                capture_output=True,
                text=True,
            )
            message = (completed.stderr or completed.stdout).strip().splitlines()
            rejected = completed.returncode != 0
            print(f"[{'ok' if rejected else 'FAIL'}] {label}: CLI exit "
                  f"{completed.returncode} -- {message[-1][:150] if message else ''}")
            ok = ok and rejected

            caught_by = []
            for layer in LAYERS:
                try:
                    getattr(validate, layer)()
                except validate.ValidationFailure as failure:
                    verdict = f"rejected -- {str(failure).splitlines()[0][:110]}"
                    caught_by.append(layer)
                else:
                    verdict = "PASSED (blind to the widening)"
                print(f"        {layer:36s} {verdict}")
            first = caught_by[0] if caught_by else "nothing"
            correct = first == expected_layer
            ok = ok and correct
            print(f"        [{'ok' if correct else 'FAIL'}] first rejecting layer: "
                  f"{first} (expected {expected_layer})")
        finally:
            COMMON.write_bytes(original_bytes)
        restored = digest(COMMON) == original_digest
        print(f"[{'ok' if restored else 'FAIL'}] {label}: reverted byte-identical")
        ok = ok and restored

    completed = subprocess.run(
        [sys.executable, "tools/validate.py"], cwd=WORKTREE, capture_output=True, text=True
    )
    print(f"[{'ok' if completed.returncode == 0 else 'FAIL'}] post-revert "
          f"tools/validate.py exit {completed.returncode}")
    ok = ok and completed.returncode == 0
    print("PLANT PROBE", "PASS" if ok else "FAIL")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
