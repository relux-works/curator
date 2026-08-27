# BUG-260731-2rhy74 — reviewer verdict

**Task:** `BUG-260731-2rhy74: marker-v2-writer-conformance-fixture`  
**Reviewer run:** `RUN-260731-46c299`  
**Date:** 2026-07-31  
**Verdict branch:** `changes_requested`  
**Route:** `to-dev`

## Verdict

Do not accept this revision yet. The curator-spec half correctly separates the
frozen marker-v1 read fixture from a generated marker-v2 writer fixture and its
local release gates pass. The CocoaSkills half, however, does not write those
canonical bytes on Windows, and neither repository has published the reviewed
delta or run PR 16 CI against it. Both are acceptance-criteria failures that can
be resolved through ordinary implementation/publication rework; there is no
human-only architecture or product decision and therefore no Stop-The-Line
`blocked` boundary.

## Blocking finding 1 — Windows text translation changes the canonical marker bytes

`src/csk/install_marker.py:428-430` returns the canonical LF-terminated rendering
as a `str`. Both production call sites pass that string to
`Path.write_text(..., encoding="utf-8")` without a `newline` argument:

- `src/csk/installer.py:2241-2243`
- `src/csk/installer.py:2297-2299`

`Path.write_text` forwards `newline=None` to `TextIOWrapper`. On output,
`newline=None` translates every `\n` to the platform default `os.linesep`; on
Windows that is CRLF. The generated fixture is explicitly LF (`*.json text
eol=lf` in curator-spec), and `serialize_install_marker(...).encode("utf-8")`
also contains LF. A reviewer-side Windows text-layer simulation on the exact
published fixture produced:

```text
serializer_matches_fixture=True
windows_text_write_matches_fixture=False
fixture bytes=787
Windows text-write bytes=825
CRLF sequences=38
```

This is not hypothetical test-only drift. The new
`tests/test_install.py:173-203` reads the actual `.csk-install.json` bytes and
requires them to equal the LF serializer bytes. The CI matrix includes
`windows-latest` for Python 3.11, 3.12, 3.13, and 3.14, so all four Windows
cells are expected to fail that assertion after the unpublished patch lands.
It also means the conformance assertion currently checks the serializer, not
the bytes production reliably commits on every supported platform.

Required rework:

1. Make both marker write sites commit deterministic UTF-8/LF bytes. Prefer a
   byte-returning serialization boundary plus `Path.write_bytes`, or explicitly
   disable text newline translation (`newline="\n"`) at both sites.
2. Keep the existing raw on-disk byte assertion; it is the regression test that
   exposes this defect on Windows.
3. Run the focused test and full CI matrix on real Windows after publication.

## Blocking finding 2 — the coupled suite/consumer revision is not published or tested

The reviewed curator-spec branch
`task/BUG-260731-2rhy74-marker-v2-fixture` remains at base
`671888e1bd2200ec9774f602c4aeeac3a34bbdaa`, with all task changes uncommitted
and `conformance/v1/expected/marker-v2.json` still untracked. A read-only
`git ls-remote --heads origin task/BUG-260731-2rhy74-marker-v2-fixture`
returned no remote branch.

The CocoaSkills PR 16 worktree also has all five task files uncommitted.
`.github/workflows/ci.yml:28-33` still checks out curator-spec
`f5d7673039226ab81de2f4f87e2155ae995c4df3`, which has no
`expected/marker-v2.json`, while `tests/test_build_metadata.py:32-34` already
requires the new manifest digest
`sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`.
No checked-out root can satisfy that mixed state.

GitHub PR 16 is still at head
`c4131bd0964839409d2d132a358a0c1c623edab6`; the task delta is absent from the
PR. Its current CI run `30589736936` is not green: the Ubuntu 3.14 test job
failed on the original schema-2-versus-schema-1 comparison, the other test
matrix jobs were cancelled/reported failed, and only mypy passed.

Required rework:

1. After fixing the deterministic writer bytes, have the commit-owning mover
   commit and publish the curator-spec fixture branch, including the currently
   untracked `marker-v2.json`.
2. Pin CocoaSkills `.github/workflows/ci.yml` to that exact published
   curator-spec commit.
3. Commit/push the CocoaSkills delta to PR 16 and wait for the complete
   Ubuntu/macOS/Windows × Python 3.11–3.14 matrix, mypy, and build job to pass.
4. Attach the published commit identities and green CI run as the next
   task-scoped outcome before another reviewer cycle.

## Independently verified green evidence

- Frozen legacy fixture unchanged:
  `conformance/v1/expected/marker.json` SHA-256
  `80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a`,
  equal to `origin/release/v1.0.0-rc.6`.
- New writer fixture SHA-256:
  `22117126c8932769636bd5ca1f1623f26a3c55d4bb9da3266acf3dc3a3b7fc2c`.
- Regenerated suite manifest SHA-256:
  `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`;
  both marker files are inventoried with their correct hashes.
- `make release-check VERSION=1.0.0-rc.6` passed in the disposable committed
  release-check copy: 42 schemas / 448 vectors, 78 Python tests, Go tests,
  regeneration diff, and rc.6 release gate all green. `gofmt -l tools`,
  `git diff --check`, and post-regeneration git status were clean.
- CocoaSkills focused marker tests: `3 passed`.
- CocoaSkills protocol conformance: `107 passed`.
- CocoaSkills build metadata: `60 passed`.
- CocoaSkills strict mypy: no issues in 67 source files.

The first release-check attempt with system Python failed before validation
because `jsonschema` was not installed. Re-running the identical command with
the task venv passed; this is recorded as an environment diagnostic, not a
product failure.

## Reviewer scope

No curator-spec or CocoaSkills source/test/workflow file was modified by this
reviewer. The only new artifacts are this verdict and its board/logbook record.
