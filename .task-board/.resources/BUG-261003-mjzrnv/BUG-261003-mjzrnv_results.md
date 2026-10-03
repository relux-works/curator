# BUG-261003-mjzrnv — freeze-guard-tests-break-on-release-tag

Developer evidence for review. Work remains uncommitted in the assigned Story worktree.

## Candidate and scope

- HEAD: `daf15ec8e78c29148063f14905fafc0b8b682786`.
- Modified files: `tools/test_validate.py`, `CHANGELOG.md` (one rc.14 Changed entry).
- Production guard (`tools/validate.py`) unchanged.
- Expected published-record coverage is derived independently from release-tag trees, including records first shipped by another version's tag. The expected inventory does not call the guard's inventory function.
- Every expected record receives byte-identity checks against its tagged snapshot and a whitespace mutation through production `validate.main()`.
- Coverage: **6/6** published records without rc.14's tag, **7/7** with it. rc.13 remains the frozen anchor; rc.6 remains anchored to rc.7.
- The active record's whitespace mutation must pass while unpublished and fail after publication. This exercises the same entry point used by the release validation command.

## Environment

Python 3.14.6 venv: `/tmp/BUG-261003-mjzrnv-venv`.
Created with `python3 -m venv /tmp/BUG-261003-mjzrnv-venv` (exit 0), then installed `requirements-dev.txt` (jsonschema 4.25.1) with the venv Python's `-m pip install -r requirements-dev.txt` (exit 0).

The module-form test command uses `PYTHONPATH=tools` because `validate.py` imports the sibling `assurance` module. Make commands select the venv Python using PATH.

## Gates run directly by this developer

Each accepted gate ran as a standalone process without tee or a pipe chain. No previously attached validation evidence was reused.

| State | Exact command | Real exit | Evidence |
| --- | --- | --- | --- |
| Untagged | `env PYTHONPATH=tools /tmp/BUG-261003-mjzrnv-venv/bin/python -m unittest tools.test_validate.ReleasedRecordImmutabilityTests` | 0 | 9 tests, 12.670s, OK |
| Untagged | `env PYTHONPATH=tools /tmp/BUG-261003-mjzrnv-venv/bin/python -m unittest tools.test_validate` | 0 | 562 tests, 371.417s, OK |
| Untagged | `env PATH=/tmp/BUG-261003-mjzrnv-venv/bin:$PATH make validate` | 0 | 73 schemas / 1294 vector files validated; 672 Python tests, 423.399s, OK; Go generator tests OK, 4.667s |
| Tag setup | `git -c tag.gpgSign=false tag v1.0.0-rc.14 HEAD` | 0 | Temporary local lightweight tag at HEAD |
| Tagged | `env PYTHONPATH=tools /tmp/BUG-261003-mjzrnv-venv/bin/python -m unittest tools.test_validate` | 0 | 562 tests, 365.584s, OK |
| Tagged | `env PATH=/tmp/BUG-261003-mjzrnv-venv/bin:$PATH make validate` | 0 | 73 schemas / 1294 vector files validated; 672 Python tests, 400.239s, OK; Go generator tests OK, 2.120s |
| Tagged, rc.13 mutated | `/tmp/BUG-261003-mjzrnv-venv/bin/python tools/validate.py` | **1** | Expected failure: published rc.13 bytes differ from v1.0.0-rc.13 |
| Tag cleanup | `git tag -d v1.0.0-rc.14` | 0 | Deleted temporary tag; never pushed |
| Untagged, restored | `env PATH=/tmp/BUG-261003-mjzrnv-venv/bin:$PATH make regenerate-check` | 0 | Release-history guard, Go regeneration, generated-artifact diff all green |
| Restored | `git diff --exit-code -- conformance schemas release tools/validate.py LOGBOOK.md` | 0 | All protected paths byte-identical to HEAD |
| Restored | `git diff --check` | 0 | Whitespace clean |
| Restored | `test -z "$(gofmt -l tools)"` | 0 | Repository's Go formatting check clean |
| Restored | `env PYTHONPYCACHEPREFIX=/tmp/BUG-261003-mjzrnv-pycache /tmp/BUG-261003-mjzrnv-venv/bin/python -m py_compile tools/test_validate.py` | 0 | Changed Python test file compiles |

The manual rc.13 negative appended one newline to a saved byte copy, ran the real CLI above, restored the exact saved bytes, and asserted restoration (exit 0). Diagnostic:

```text
validation failed: published release record bytes differ from v1.0.0-rc.13: release/1.0.0-rc.13.json
```

## Narrowed-guard negative

While rc.14 was tagged, ran the following standalone process. It omits only active rc.14 from the guard inventory, leaving all six historical records protected. Both independently derived coverage and production-entry mutation tests failed. **Real exit 1**, expected failure; not a passing gate. Two tests ran in 9.227s and produced two failures:

```text
test_all_published_records_are_covered_and_unchanged ... FAIL
Items in the second set but not the first: 'release/1.0.0-rc.14.json'
test_every_published_record_byte_drift_is_rejected_through_main
  (relative='release/1.0.0-rc.14.json') ... FAIL
AssertionError: 0 != 1
FAILED (failures=2)
```

Reproducer (process-local patch, no production file edits):

```bash
env PYTHONPATH=tools /tmp/BUG-261003-mjzrnv-venv/bin/python -c 'import sys, unittest
from unittest.mock import patch
from tools import test_validate
original = test_validate.validate.published_release_records
active = "release/1.0.0-rc.14.json"
def omit_active():
    return {path: record for path, record in original().items() if path != active}
suite = unittest.TestSuite(test_validate.ReleasedRecordImmutabilityTests(name) for name in ("test_all_published_records_are_covered_and_unchanged", "test_every_published_record_byte_drift_is_rejected_through_main"))
with patch.object(test_validate.validate, "published_release_records", side_effect=omit_active):
    result = unittest.TextTestRunner(verbosity=2).run(suite)
sys.exit(0 if result.wasSuccessful() else 1)'
```

## Preservation and cleanup

A standalone venv Python check asserted the manifest digest against the exact expected value and asserted `git tag --list v1.0.0-rc.14` returned an empty inventory. Exit 0; output:

```text
manifest sha256:6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5
temporary local rc.14 tag absent
```

`git status --short` reports only `CHANGELOG.md` and `tools/test_validate.py` modified. No commits, pushes, branch changes, schema changes, release-record changes, or conformance changes. No production validation gate was weakened.

## Operational anomalies

- Initial `git tag v1.0.0-rc.14 HEAD` failed with exit **128**, `fatal: no tag message?`, because inherited `tag.gpgSign=true` requests a signed tag. Its follow-up `git show-ref --tags v1.0.0-rc.14` exited **1** (tag absent). The test process accidentally started after that failure was deliberately interrupted with SIGINT and exited **130**; it is not counted as tagged evidence. Fixtures were restored and checked, then the explicitly requested lightweight tag was created with a command-local configuration override. Both full tagged gates subsequently ran green.
- The initial focused suite and first module suite briefly overlapped. Untagged `make validate` subsequently reran all 672 Python tests sequentially against the unchanged candidate, so broad handoff evidence does not depend on that overlap.
- One early diagnostic inspection ran `git diff --exit-code` while a suite was actively mutating fixtures and then ran further reads in the same shell. That shell exited 0 but did not preserve the diff subcommand's status; no acceptance evidence is inferred from it. The standalone preservation gate above ran after every suite and mutation had exited and returned 0.
- `LOGBOOK.md` intentionally unchanged per the explicit freeze-fix brief. Findings and decisions are preserved in board notes and this outcome instead.

All required local tests/build/validation commands were run; no required local gate remains unrun. Hosted release publication and integration remain outside the developer handoff scope.
