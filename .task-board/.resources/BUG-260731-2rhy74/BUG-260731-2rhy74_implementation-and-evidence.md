# BUG-260731-2rhy74 — marker-v2 writer conformance fixture

**Role:** developer · **Board status at handoff:** `to-review`
**Date:** 2026-07-31

---

## 1. Defect, reproduced

`protocol/core.md` §10 requires managers to **write marker schema 2 for every schema 1
through 6 installation mutation**. `tests/test_protocol_conformance.py::test_shared_fixture_context_hash_and_marker`
compared that writer output to `conformance/v1/expected/marker.json`, which is
schema-1 legacy-read evidence. Reproduced against the rc.6 candidate root before any
change:

```
E       AssertionError: assert {'schema_vers...evision', ...} == {'activation'...1234567', ...}
E         Differing items:
E         {'schema_version': 2} != {'schema_version': 1}
E         Left contains 2 more items:
E         {'build_roots': [], 'builds': {}}
tests/test_protocol_conformance.py:76: AssertionError
1 failed, 105 deselected
```

The delta is exactly three members and nothing else: the writer is correct, the golden
was the wrong role.

---

## 2. Change — `curator-spec`

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/spec`
Branch: `task/BUG-260731-2rhy74-marker-v2-fixture` off `origin/release/v1.0.0-rc.6` (`671888e`)
**Uncommitted** — per standing instruction the agent does not commit or stage.

| File | Change |
| --- | --- |
| `tools/generate-vectors/main.go` | `sharedFixtureMarkerV2()` derives the writer golden from the marker-v1 object; `main()` writes `expected/marker-v2.json` |
| `conformance/v1/expected/marker-v2.json` | **new generated fixture** (`sha256:22117126c8932769636bd5ca1f1623f26a3c55d4bb9da3266acf3dc3a3b7fc2c`) |
| `conformance/v1/manifest.json` | regenerated — one added entry; every other entry byte-unchanged |
| `release/1.0.0-rc.6.json` | regenerated pin → `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071` |
| `tools/validate.py` | new `validate_shared_fixture_markers()` registered in `main()` |
| `tools/release_gate.py` | `expected/marker.json` + `expected/marker-v2.json` added to `RC6_REQUIRED_FILES` and `RC6_REQUIRED_MANIFEST_FILES`; new `validate_shared_fixture_marker_release_surface()` |
| `tools/test_validate.py` | `SharedFixtureMarkerTests` — 10 tests |
| `tools/test_release_gate.py` | 8 new `ProtocolRC6ReleaseGateTests` cases |
| `tools/generate-vectors/main_test.go` | `TestSharedFixturePublishesBothLegacyAndWriterMarkers` |
| `conformance/README.md`, `CHANGELOG.md` | document both marker roles |

### The fixture

`expected/marker-v2.json` restates the same golden installation as `expected/marker.json`
with exactly `{schema_version, build_roots, builds}` changed:

```json
  "build_roots": [],
  "builds": {},
  "schema_version": 2,
  "skill_schema_version": 5,
```

No `build_source` — the golden skill activates no compiled command, and marker v2
requires `build_source` **exactly when** `builds` is non-empty.

### Legacy fixture preserved byte-for-byte

`conformance/v1/expected/marker.json` is untouched: `sha256:80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a`,
identical to the published rc.5 bytes. It is now frozen by a hard-coded constant in
**both** `tools/validate.py` and `tools/release_gate.py`, so a future in-place upgrade of
the legacy file fails the gate instead of silently rewriting release history.

### Fail-closed guarantees added

Both `validate.py` and `release_gate.py` reject: a missing writer golden; an edited
legacy marker; a writer golden that is not schema 2; a `skill_schema_version` outside
1–6; invented `build_roots`/`builds`; `build_source` without builds; and any member
drift from the legacy installation. `validate.py` additionally validates each marker
against its own JSON Schema and re-checks sorted-unique set members.

---

## 3. Change — CocoaSkills (PR 16 branch)

Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-3t8nr3/worktree`
Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid` (PR #16) — **uncommitted**.

| File | Change |
| --- | --- |
| `src/csk/install_marker.py` | new `serialize_install_marker()` — the exact `.csk-install.json` rendering |
| `src/csk/installer.py` | both marker write sites now call it instead of duplicating `json.dumps(...)` |
| `tests/test_protocol_conformance.py` | requires `expected/marker-v2.json`, compares writer bytes; new `test_shared_fixture_legacy_marker_v1_stays_readable` |
| `tests/test_install.py` | new `test_install_writes_marker_schema_2_bytes_for_a_schema_1_skill` (runs without a conformance root) |
| `tests/test_build_metadata.py` | `EXPECTED_MANIFEST_SHA256` advanced to the rc.6 candidate digest |

The consumer now asserts three things about the writer golden:

1. `installer._marker_payload(...)` equals the golden object;
2. `install_marker.serialize_install_marker(payload)` equals the golden **bytes**
   (read with `read_bytes()`, no newline translation);
3. the golden bytes read back through `install_marker.read_install_marker` as an
   `InstallMarkerV2` whose `to_json()` is that same payload.

Legacy-read compatibility is asserted separately and positively: `expected/marker.json`
parses as `InstallMarkerV1` (schema 1, skill schema 5) and round-trips.

### Fail-closed, verified

Against the currently pinned rc.5 root (`f5d7673`), which has no writer golden:

```
E  AssertionError: conformance root .../pin-rc5/conformance/v1 publishes no expected/marker-v2.json
1 failed, 106 passed
```

No skip, no silent pass.

---

## 4. Evidence — commands and real exit codes

### curator-spec — full `make release-check VERSION=1.0.0-rc.6` chain

Run inside a disposable committed copy (`.temp/BUG-260731-2rhy74/release-check`,
`git init` + one commit) because `release_gate.validate_checkout` requires a clean
checkout at a candidate commit and the working branch is intentionally uncommitted.
Contents are byte-identical to the worktree.

| Command | Exit | Output |
| --- | ---: | --- |
| `python tools/validate.py` | **0** | `validated 42 schemas and 448 vector files` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | **0** | `Ran 78 tests ... OK` (60 before, 18 new) |
| `go test ./tools/...` | **0** | `ok github.com/relux-works/curator-spec/tools/generate-vectors` |
| `go run ./tools/generate-vectors -root .` | **0** | — |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json` | **0** | regenerate-check clean |
| `python tools/release_gate.py --version 1.0.0-rc.6 --commit HEAD` | **0** | `release gate passed for 1.0.0-rc.6 at 9255e74b` |

The same `regenerate-check` was also proved on the real worktree using a scratch
`GIT_INDEX_FILE` (real index untouched): exit **0**.

Specification CI also runs two formatting gates, both clean on the worktree:
`gofmt -l tools` → empty, exit **0**; `git diff --check` → exit **0**.

### CocoaSkills

| Command | Root | Exit | Result |
| --- | --- | ---: | --- |
| `pytest tests/test_protocol_conformance.py -q` | rc.6 + fixture | **0** | `107 passed` |
| `pytest tests/test_protocol_conformance.py -q` | rc.5 pin `f5d7673` | **1** | `1 failed, 106 passed` — **expected red**, the fail-closed assertion above |
| `pytest tests/test_build_metadata.py -q` | rc.6 + fixture | **0** | `60 passed` |
| `pytest tests/test_install.py -k marker_schema_2_bytes -q` | n/a | **0** | `1 passed` |
| `python -m mypy` (strict, `src/csk`) | n/a | **0** | `Success: no issues found in 67 source files` |
| `pytest -q` (full suite) | rc.6 + fixture | **0** | `1303 passed, 49 skipped in 1103.35s` |

The full suite was also run once against the rc.6 root **before** `EXPECTED_MANIFEST_SHA256`
was advanced: 60 failures, all from `test_build_metadata._root()`'s digest assertion, none
from behaviour. With the digest advanced, `tests/test_build_metadata.py` alone is `60 passed`.

### Lint

CocoaSkills declares no linter: there is no `[tool.ruff]` section, no `ruff.toml`, and
CI runs only `pytest`, `mypy` and the package build. `mypy --strict` over `src/csk` is
the type gate and is clean. For completeness `ruff check` was run on the five changed
files both at `HEAD` and on the working tree: **identical output in both** — exit `1`
with the same 11 pre-existing findings (`4 I001`, `2 PYI034`, `2 PYI036`, `2 RUF012`,
`1 DTZ007`), none of them in code this task added. No new lint finding is introduced.
On the specification side `gofmt`/`git diff --check` are real CI gates and pass.

---

## 5. Remaining step — requires publication (not done here)

The consumer and the suite must move together, and one value cannot be known until the
spec change is published.

`.github/workflows/ci.yml` still pins `relux-works/curator-spec@f5d7673…` (rc.5). That
root has no `expected/marker-v2.json`, so PR 16 CI stays red on
`test_shared_fixture_context_hash_and_marker` until the pin advances.

**Exactly one edit remains, after the curator-spec change lands:**

```yaml
      - name: Checkout Curator Protocol conformance suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: <commit that publishes conformance/v1/expected/marker-v2.json>
          path: protocol-spec
```

`tests/test_build_metadata.py::EXPECTED_MANIFEST_SHA256` is already advanced to
`sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`, which the
generator derives deterministically from the suite content and does not depend on the
commit SHA. This was required, not optional: with the old rc.5 digest **no** conformance
root could satisfy both that constant and the writer-golden assertion — rc.5 has the
digest but not the fixture, rc.6 has the fixture but not the digest.

Why the agent did not do it: committing, pushing and choosing the published pin are
human decisions under this project's standing instructions ("never commit or stage
files automatically"). Both repositories are left as reviewable working trees.

---

## 6. Files touched

**curator-spec** (`.temp/BUG-260731-2rhy74/spec`)

```
 M CHANGELOG.md
 M conformance/README.md
 M conformance/v1/manifest.json
 M release/1.0.0-rc.6.json
 M tools/generate-vectors/main.go
 M tools/generate-vectors/main_test.go
 M tools/release_gate.py
 M tools/test_release_gate.py
 M tools/test_validate.py
 M tools/validate.py
?? conformance/v1/expected/marker-v2.json
```

**cocoaskills** (`.temp/TASK-260720-3t8nr3/worktree`)

```
 M src/csk/install_marker.py
 M src/csk/installer.py
 M tests/test_build_metadata.py
 M tests/test_install.py
 M tests/test_protocol_conformance.py
```

## 7. Scratch artifacts

- `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/spec` — curator-spec worktree
- `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/release-check` — disposable committed copy for the release gate
- `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/pin-rc5` — rc.5 pin checkout used to prove fail-closed
- `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/venv` — task venv (cocoaskills `[dev]` + `jsonschema==4.25.1`)
