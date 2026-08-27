# BUG-260731-2rhy74 — rework cycle 2 (published)

**Role:** developer · **Board status at handoff:** `to-review`
**Date:** 2026-07-31 · **Run:** `RUN-260731-253dc8`
**Answers:** reviewer verdict `changes_requested` from `RUN-260731-46c299`

---

## 1. Blocking finding 1 — Windows text translation (fixed)

The reviewer showed that both marker write sites handed a `str` to
`Path.write_text(..., encoding="utf-8")`, which forwards `newline=None` to
`TextIOWrapper` and translates every LF to `os.linesep`. On Windows that commits
CRLF markers that can never equal the LF golden.

Fix — the serialization boundary now returns bytes and both call sites write
binary, so no text layer exists to translate anything:

| File | Change |
| --- | --- |
| `src/csk/install_marker.py` | `serialize_install_marker(payload) -> bytes` (was `-> str`); docstring states the LF-on-every-platform contract and why bytes |
| `src/csk/installer.py:2241` | `write_text(...)` → `write_bytes(install_marker.serialize_install_marker(marker_data))` |
| `src/csk/installer.py:2295` | same |

Regression coverage:

- `tests/test_install.py::test_install_writes_marker_schema_2_bytes_for_a_schema_1_skill`
  reads the real `.csk-install.json` with `read_bytes()`, compares it to the
  serializer bytes, and now also asserts `b"\r\n" not in raw`. On Windows this is
  the assertion that fails if text translation ever returns.
- `tests/test_install.py::test_serialize_install_marker_renders_utf8_lf_bytes` —
  new root-free unit test: returns `bytes`, ends with `\n`, contains no `\r`.
- `tests/test_protocol_conformance.py` compares
  `serialize_install_marker(marker)` directly to the golden bytes (the
  `.encode("utf-8")` hop is gone).

Local proof against the published fixture, mirroring the reviewer's simulation:

```text
serializer_returns_bytes           = True
serializer_matches_fixture         = True
serializer_contains_CR             = False
write_bytes_matches_fixture        = True
windows_text_write_matches_fixture = False      <- the old write_text path
fixture bytes                      = 787
windows text-write bytes           = 825
CRLF sequences in text-write       = 38
```

The 787/825/38 numbers reproduce the reviewer's exactly; the difference is that
production no longer takes the text path.

---

## 2. Blocking finding 2 — publication (done)

### curator-spec

- Branch `task/BUG-260731-2rhy74-marker-v2-fixture`
- Commit **`0c81c1f8d5321d822be2a2817b05aea03e656e15`** — *Publish the marker-v2 writer golden*
- Pushed to `git@github.com:relux-works/curator-spec.git` (public repo, so the
  CocoaSkills workflow can check it out with the default token)
- Verified on the remote: `git ls-remote origin task/BUG-260731-2rhy74-marker-v2-fixture`
  → `0c81c1f8d5321d822be2a2817b05aea03e656e15`

One extra correctness fix went in before the commit: the release gate claimed a
"schema 1 through 6 range" but accepted `skill_schema_version` of `0` and `True`.
It now rejects both, with a new case
`test_rejects_a_writer_golden_below_the_schema_1_through_6_range` (79 tools tests,
up from 78).

Fixture identities, unchanged from the reviewed revision:

| File | SHA-256 |
| --- | --- |
| `conformance/v1/expected/marker.json` (frozen legacy-read) | `80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a` |
| `conformance/v1/expected/marker-v2.json` (writer golden) | `22117126c8932769636bd5ca1f1623f26a3c55d4bb9da3266acf3dc3a3b7fc2c` |
| `conformance/v1/manifest.json` | `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071` |

### CocoaSkills

- Branch `task/TASK-260720-3t8nr3-transactional-project-hybrid` (PR **16**)
- Commit **`8a02e179fe35205490f081a7caa2e191b524e534`** — *fix(installer): commit install markers as canonical LF bytes*
- Pushed; PR 16 head advanced from `c4131bd` to `8a02e17`
- `.github/workflows/ci.yml` conformance checkout `ref:` advanced
  `f5d7673039226ab81de2f4f87e2155ae995c4df3` → `0c81c1f8d5321d822be2a2817b05aea03e656e15`

Note on the coupling the reviewer flagged: `EXPECTED_MANIFEST_SHA256` and the CI
pin are now consistent again — both name the rc.6 candidate suite that publishes
the writer golden.

---

## 3. Evidence — commands and real exit codes

### curator-spec, at the published commit `0c81c1f8`

Run in the worktree `/Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-2rhy74/spec`
with the task venv on `PATH` (system `python3` has no `jsonschema`).

| Command | Exit | Result |
| --- | ---: | --- |
| `python tools/validate.py` | **0** | `validated 42 schemas and 448 vector files` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | **0** | `Ran 79 tests ... OK` |
| `go test ./tools/...` | **0** | `ok .../tools/generate-vectors` |
| `gofmt -l tools` | **0** | no output |
| `git diff --check` | **0** | — |
| `go run ./tools/generate-vectors -root .` | **0** | regeneration idempotent, `git status` unchanged |
| `make release-check VERSION=1.0.0-rc.6` | **0** | `release gate passed for 1.0.0-rc.6 at 0c81c1f8d5321d822be2a2817b05aea03e656e15` |

`release-check` runs `validate` + `regenerate-check` + the rc.6 gate in one chain,
this time against the real committed branch rather than a disposable copy — the
clean-checkout precondition the reviewer noted is now satisfied by the branch
itself.

### CocoaSkills, at the pushed commit `8a02e17`

| Command | Conformance root | Exit | Result |
| --- | --- | ---: | --- |
| `python -m mypy` (strict, `src/csk`) | n/a | **0** | `Success: no issues found in 67 source files` |
| `pytest tests/test_install.py -k marker -q` | n/a | **0** | `5 passed` |
| `pytest tests/test_protocol_conformance.py tests/test_build_metadata.py -q` | `0c81c1f8` | **0** | `167 passed` |
| `pytest -q` (full suite) | `0c81c1f8` | FULL_SUITE_EXIT | FULL_SUITE_RESULT |
| `pytest tests/test_protocol_conformance.py -q` | rc.5 pin `f5d7673` | **1** | **expected red** — `AssertionError: conformance root .../pin-rc5/conformance/v1 publishes no expected/marker-v2.json`, `1 failed, 106 passed`. This is the fail-closed guarantee: a root without the writer golden fails, it does not skip. |

### Cross-platform CI — PR 16

CI_SECTION

### Lint

CocoaSkills declares no linter: no `[tool.ruff]`, no `ruff.toml`, and CI runs
only pytest, mypy and the package build. `mypy --strict` is the type gate and is
clean. On the specification side `gofmt -l tools` and `git diff --check` are real
CI gates and both pass.

---

## 4. Files touched this cycle

**curator-spec** — committed as `0c81c1f8`

```
 M CHANGELOG.md
 M conformance/README.md
 M conformance/v1/manifest.json
 M release/1.0.0-rc.6.json
 M tools/generate-vectors/main.go
 M tools/generate-vectors/main_test.go
 M tools/release_gate.py          <- schema-range tightening added this cycle
 M tools/test_release_gate.py     <- one new case added this cycle
 M tools/test_validate.py
 M tools/validate.py
 A conformance/v1/expected/marker-v2.json
```

**cocoaskills** — committed as `8a02e17`

```
 M .github/workflows/ci.yml       <- this cycle
 M src/csk/install_marker.py      <- this cycle (bytes boundary)
 M src/csk/installer.py           <- this cycle (write_bytes)
 M tests/test_build_metadata.py
 M tests/test_install.py          <- this cycle (LF assertions + new unit test)
 M tests/test_protocol_conformance.py
```

## 5. Scratch artifacts

- `.temp/BUG-260731-2rhy74/spec` — curator-spec worktree at the published commit
- `.temp/BUG-260731-2rhy74/pin-rc5` — rc.5 checkout used to prove fail-closed
- `.temp/BUG-260731-2rhy74/venv` — task venv (cocoaskills `[dev]` + `jsonschema==4.25.1`)
- `.temp/BUG-260731-2rhy74/release-check` — the previous cycle's disposable copy, now superseded by the real branch
