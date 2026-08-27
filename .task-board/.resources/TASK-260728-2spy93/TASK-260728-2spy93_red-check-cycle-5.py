"""Show which cycle-5 regressions are red against the pre-fix gate.

The new structural pin is stubbed back to its cycle-4 no-op, then the artifact
tests are run. Exit 0 means every regression that must fail without the fix
fails, and the tests that merely document the escape still pass either way.
"""

from __future__ import annotations

import importlib.util
import sys
import unittest
from pathlib import Path

WORKTREE = Path(__file__).resolve().parent / "curator-spec-worktree"
sys.path.insert(0, str(WORKTREE / "tools"))
SPEC = importlib.util.spec_from_file_location(
    "probe_test_validate", WORKTREE / "tools" / "test_validate.py"
)
assert SPEC is not None and SPEC.loader is not None
tests = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(tests)

MUST_BE_RED = (
    "test_artifact_identity_target_cannot_be_widened_by_one_keyword",
    "test_artifact_identity_target_cannot_disappear",
    "test_artifact_identity_target_cannot_drop_a_bound",
    "test_artifact_reference_targets_cannot_be_widened_underneath",
    "test_every_artifact_reference_must_be_pinned",
    "test_artifact_member_pinned_outside_the_shared_definitions",
)
MUST_STAY_GREEN = (
    "test_pinned_targets_are_the_shipped_shared_definitions",
    "test_one_keyword_widening_survives_every_sampled_rejection",
    "test_one_keyword_widening_lets_the_real_validator_accept_it",
)


def run(name: str) -> bool:
    suite = unittest.TestLoader().loadTestsFromName(
        f"AdditionalDriverBoundaryTests.{name}", tests
    )
    result = unittest.TextTestRunner(stream=open("/dev/null", "w"), verbosity=0).run(suite)
    return result.wasSuccessful()


def main() -> int:
    tests.validate.check_artifact_reference_targets = lambda common: None
    ok = True
    for name in MUST_BE_RED:
        passed = run(name)
        ok = ok and not passed
        print(f"[{'ok' if not passed else 'FAIL'}] pre-fix red: {name} -> "
              f"{'PASSED (gap open)' if passed else 'FAILED as required'}")
    for name in MUST_STAY_GREEN:
        passed = run(name)
        ok = ok and passed
        print(f"[{'ok' if passed else 'FAIL'}] pre-fix green: {name} -> "
              f"{'PASSED' if passed else 'FAILED'}")
    print("RED CHECK", "PASS" if ok else "FAIL")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
