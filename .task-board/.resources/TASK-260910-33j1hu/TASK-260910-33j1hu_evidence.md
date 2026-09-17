# TASK-260910-33j1hu evidence — spec-restore-enforcement-point (R3/P2)

Revision 2 supersedes revision 1 for `tools/validate.py` and
`tools/test_validate.py` only (F1 correction); every other file is
byte-identical to revision 1 (proof in the Revision 2 section below).
The revision 1 body is preserved unchanged after that section.

## Revision 2 — F1 correction (scenario predicates pinned in the validator gate)

Review round 1 rejected revision 1 with one finding (F1): the validator
gate required the seven `checkpoint_cases` names and recomputed each
verdict from the case's own inputs, but nothing pinned that a name still
represents its mandatory scenario — replacing any negative case with an
internally consistent passing case of the same name survived the whole
gate (0/4 mutants rejected). Normative text, vectors and everything else
passed; per the rework brief they are kept byte-identical.

### What changed in revision 2

1. `tools/validate.py`: new `require_checkpoint_scenario(case)` called at
   the top of the `validate_registry_checkpoint_vectors` loop, i.e. on the
   production `validate.main()` path (gate registered in `main()` as
   before). Each required name asserts its discriminating input
   predicates, mirroring the oracle's `is True` / `is not True`
   conventions so pin and verdict can never disagree:
   - `checkpoint-not-configured`: checkpoint not configured;
   - `checkpoint-signature-invalid`: checkpoint configured, signature
     invalid (versions unconstrained — the signature rule short-circuits
     before any comparison, as in §6);
   - `live-below-checkpoint`: configured, signature valid, live version
     strictly below the checkpoint version;
   - `checkpoint-equal-consistent` / `checkpoint-equal-inconsistent`:
     configured, signature valid, live version equal to the checkpoint
     version, same versus different boundary body;
   - `checkpoint-below-live-consistent` / `live-above-prefix-mismatch`:
     configured, signature valid, live version strictly above the
     checkpoint version, prefix reproduced versus not.
   The comparison cases re-assert non-negative integer versions with the
   oracle's identical message (fail closed, never `TypeError`). Extra
   (non-required) names carry no scenario pin and are still
   verdict-checked. `expected_checkpoint_verdict` itself is untouched.
2. `tools/test_validate.py`: `RegistryCheckpointVectorTests` grows from 9
   to 17 tests. The 4 required negatives replace each negative scenario
   with the reviewer's internally consistent equal-consistent passing
   case (same name, same update shape as
   `TASK-260910-33j1hu_review-probe-rev1.py`) and require rejection;
   3 further tests pin the passing/not-configured names against
   self-consistent replacement in the other direction; 1 test drives the
   production `validate.main()` entry with a `load_json` substitution
   (reviewer-probe shape) and requires exit 1. All 9 revision 1 tests are
   kept as positive controls.

### Revision 1 files byte-identical

Per-file `git diff HEAD` for the seven files revision 2 must not touch
equals the attached `TASK-260910-33j1hu_spec-patch_rev1.patch` hunks
(`index` lines excluded — they carry blob hashes, not content):

```
IDENTICAL CHANGELOG.md
IDENTICAL conformance/v1/manifest.json
IDENTICAL conformance/v1/vectors/registry-service.json
IDENTICAL profiles/registry-service.md
IDENTICAL protocol/registry.md
IDENTICAL release/1.0.0-rc.9.json
IDENTICAL tools/generate-vectors/main.go
```

`git status --short` lists exactly the same nine paths as revision 1; no
new files, no implementation code.

### Reviewer-probe re-run (verbatim `TASK-260910-33j1hu_review-probe-rev1.py`)

```
validation failed: registry-service checkpoint case 'checkpoint-equal-inconsistent' must carry a different boundary body
validation failed: registry-service checkpoint case 'live-below-checkpoint' must place the live version below the checkpoint version
validation failed: registry-service checkpoint case 'live-above-prefix-mismatch' must place the live version above the checkpoint version
validation failed: registry-service checkpoint case 'checkpoint-signature-invalid' must carry an invalid checkpoint signature
checkpoint-equal-inconsistent replaced with equal-consistent while preserving name: main exit 1
live-below-checkpoint replaced with equal-consistent while preserving name: main exit 1
live-above-prefix-mismatch replaced with equal-consistent while preserving name: main exit 1
checkpoint-signature-invalid replaced with equal-consistent while preserving name: main exit 1
```

4/4 branch-replacement mutants now rejected by the production
`validate.main()` entry (exit 1 each; the `validation failed` lines are
`main()`'s own stderr reporting). Probe script exit 0 (it reports; the
per-mutant `main exit 1` values are the verdicts).

### Validation transcript (revision 2)

Shell `sh` via the agent harness, from the worktree, repo venv first on
`PATH`. The unittest second third is partitioned into bounded sequential
calls (same modules and classes as `discover`, split only for the
per-call time bound); counts sum to the full suite.

```
$ python3 tools/validate.py
validated 62 schemas and 1071 vector files
VALIDATE_PY_EXIT=0
```

```
$ python3 -B -m unittest test_implementation_coverage test_release_gate test_verify_release_commit test_verify_release_merge_policy
Ran 78 tests in 87.656s
OK
SMALL_EXIT=0
$ python3 -B -m unittest test_validate.WireSemanticValidationTests ...SharedFixtureMarkerTests  (6 classes)
Ran 54 tests in 26.098s
OK
G1_EXIT=0
$ python3 -B -m unittest test_validate.WorkflowRegenerationScopeTests ...SnapshotAcquisitionVectorTests  (6 classes)
Ran 85 tests in 84.706s
OK
G2_EXIT=0
$ python3 -B -m unittest test_validate.ShellHookTrustVectorTests ...RegistryCheckpointVectorTests  (6 classes)
Ran 101 tests in 112.149s
OK
G3_EXIT=0
```

78 + 54 + 85 + 101 = 318 tests (310 revision 1 + 8 new), all green. No
`go run`/`go test` exec workaround was needed in revision 2 — plain
`go test` passed directly (the revision 1 Mach-O exec-gating stall did
not recur):

```
$ go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors  0.751s
GOTEST_EXIT=0
```

Regeneration proof in a disposable copy (worktree bytes excluding
`.git`/`.venv`; before/after SHA256 over all conformance + release
files):

```
$ go run ./tools/generate-vectors -root .
GENVEC_EXIT=0
$ diff regen-before.txt regen-after.txt
REGEN_IDENTICAL  (1197 files, zero changes)
```

Deliberately out of scope (unchanged from revision 1): implementation
(`TASK-260910-1ny7yl`), S2 client bootstrap (`TASK-260910-1tvf2t`), P4
import high-water, tags/releases, ax, proposals 0014–0018, schema
changes. No LOGBOOK.md edit per the campaign's no-LOGBOOK-edits rule
for spec work.

---

Story `STORY-260910-35tbgb` (serve-time-checkpoint-gate, `EPIC-260910-16qce1`).
Worktree: `curator-spec/.temp/STORY-260910-35tbgb/worktree`
(branch `task-board/story/STORY-260910-35tbgb`, base `dced9b8` = `origin/main`).
Role: doc-writer (normative spec text + spec conformance tooling; no product
implementation — the service task is `TASK-260910-1ny7yl`).

## Findings read

- curator-spec `docs/security-audit-2026-09.md` P2 (Low): registry-service
  profile §6 reads as service behavior ("MUST NOT serve … before the service
  becomes ready") but the only buildable mechanism is the offline
  `verify-backup` CLI check; the enforcement point is ambiguous — "should
  state exactly where enforcement lives (e.g. `serve --checkpoint`)".
- curator-skill-registry "R3 + P2" (Medium, per brief): `serve` never compares
  live state against the out-of-band checkpoint, so a silently restored older
  database passes startup and serves stale state.

## What changed per file

1. `profiles/registry-service.md`
   - §6 rewritten around the enforcement point: the checkpoint interchange
     object (signed `registry-snapshot-v1`, kept outside the primary store,
     verified against the operator's out-of-band registry key set;
     deployment wrapper MAY add storage metadata); the normative gate is the
     **startup checkpoint comparison** — AFTER the §5 startup integrity
     verification and BEFORE the service binds its listener or reports
     ready; signature verified first against the accepted signing keys (the
     staged rotation set), then the live-boundary comparison (below →
     refuse; equal-requires-same-body; above-requires-prefix-reproduction);
     refusal = non-ready, writes disabled, no automatic truncation or
     repair (recover the missing verified suffix or remain unavailable);
     no-checkpoint posture (`checkpoint_not_configured` in startup
     diagnostics/audit log; compared boundary + outcome with one);
     offline `verify-backup`-style comparison kept as the operator
     pre-restore vetting procedure, explicitly not the readiness gate; a
     diagnostics table with exactly the four closed codes plus the
     "No other restore-checkpoint diagnostic exists." closure.
   - §5: ordering cross-reference (verification runs BEFORE the §6
     comparison; a refused comparison is non-ready like a failed
     verification here).
   - §9: refused comparison is non-ready (`503` envelope) like a failed §5
     verification; `health-response-v1` schema unchanged (reported through
     readiness, not a new member); startup diagnostics/audit-log posture
     row (compared boundary + outcome, or `checkpoint_not_configured`).
   - §10: "external checkpoints" replaced with the §6 startup checkpoint
     comparison against the operator checkpoint (what the service itself
     now enforces).
   - §11: conformance list gains the startup checkpoint comparison
     including the no-checkpoint posture (§6).
2. `protocol/registry.md` §5: one sentence pointing operators to the
   profile's startup checkpoint comparison as the normative enforcement
   point for a restored service state before it becomes ready. No client
   bootstrap specified here (owned by `TASK-260910-1tvf2t`).
3. `tools/generate-vectors/main.go` (`writeRegistryServiceVectors`):
   `registry-service.json` gains `checkpoint_cases` (7 cases: checkpoint
   below live and consistent → ready; equal-consistent → ready;
   equal-inconsistent → not ready; live below checkpoint → not ready; live
   above with prefix mismatch → not ready; bad signature → not ready; no
   checkpoint → ready with posture). Existing cases untouched in the
   generator source.
4. `tools/validate.py`: new `validate_registry_checkpoint_vectors` gate
   (registered in `main()`), recomputing each case's `(ready, diagnostic,
   posture)` from its inputs via `expected_checkpoint_verdict`; closed
   diagnostic set `CHECKPOINT_DIAGNOSTICS` + `CHECKPOINT_NOT_CONFIGURED`
   posture pin.
5. `tools/test_validate.py`: new `RegistryCheckpointVectorTests` (9 tests:
   published vectors pass + 8 narrowing mutants proving the gate rejects
   admitted-below, admitted-inconsistent, admitted-prefix-mismatch,
   admitted-bad-signature, misspelled/swapped diagnostics, dropped posture,
   dropped case).
6. `conformance/v1/vectors/registry-service.json`,
   `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
   via `make regenerate` (new `checkpoint_cases` key; manifest digest +
   rc.9 pins follow).
7. `CHANGELOG.md`: Unreleased → Added `R3/P2` entry naming the enforcement
   point, the four diagnostics, the direct rollout, and the story.

## Closed-set spelling (identical in text, table, vectors, CHANGELOG)

- `restore_below_checkpoint`
- `restore_inconsistent_with_checkpoint`
- `checkpoint_signature_invalid`
- `checkpoint_not_configured`

## Settled decisions honoured

- Enforcement point = service startup when a checkpoint is configured; the
  profile names the "startup checkpoint comparison" capability, not a flag
  spelling.
- Comparison runs after §5 verification and before listener bind / ready.
- Checkpoint object is the signed `registry-snapshot-v1` the profile
  already produces; signature verified against the staged rotation set.
- Refusal consequences exactly as settled (non-ready, writes disabled, no
  truncation/repair; recover suffix or stay unavailable).
- No-checkpoint posture: service starts (no behavior change, direct
  rollout), posture recorded as `checkpoint_not_configured`.
- Frozen `health-response-v1` untouched (no schema change; refusal via
  readiness/`503`).
- Offline `verify-backup`-style comparison stays as the operator procedure
  for candidate backups; the profile names both places and marks the
  startup comparison normative for "before the service becomes ready".

## Validation transcript

(validation commands run from the worktree with the repo venv first on
`PATH`; shell `sh` via the agent harness; `make` targets as defined in the
repo `Makefile`.)

### make regenerate

Host note: freshly built Mach-O binaries stall at first exec on this host
(proven: a hello-world binary and the `go run` generator sat at 0 CPU for
10+ min; a trivial C compile behaved the same; other agents' `go test`
binaries show the same 0-CPU stall; `task-board` without
`--no-update-check` also hangs on network). Re-signing the binary with the
local ad-hoc identity (`codesign --force --sign -`) lets it exec. So the
regeneration below runs the real, unmodified generator source compiled with
`go build` + local re-sign instead of `go run` — same code, same flags,
same root:

```
$ go build -o /tmp/genvec ./tools/generate-vectors && codesign --force --sign - /tmp/genvec
BUILD_OK
SIGNED
$ /tmp/genvec -root .
GENVEC_EXIT=0
 M CHANGELOG.md
 M conformance/v1/manifest.json
 M conformance/v1/vectors/registry-service.json
 M profiles/registry-service.md
 M protocol/registry.md
 M release/1.0.0-rc.9.json
 M tools/generate-vectors/main.go
 M tools/test_validate.py
 M tools/validate.py
```

Exactly the three generated files changed; no other vector/expected churn.
Idempotence proof (hash → regenerate → hash → equal):

```
$ shasum -a 256 <3 files> > h1; /tmp/genvec -root .; shasum ... > h2; diff h1 h2
IDEMPOTENT
1bacfdf34d0ed2a18e17bf6a0dd578d4e7a0da32f6023b6d61255eff1554e727  conformance/v1/vectors/registry-service.json
fd28b0601eb5a62c24ecb56cc5bd312a4a8274d6f9bedbbaa2ce051f158aae8a  conformance/v1/manifest.json
0355d79a31622c04febf82bd45d92b79498441e345df7c55d729ef472d5aecf5  release/1.0.0-rc.9.json
```

(The `make regenerate-check` form `git diff --exit-code` is not quotable
here: the worktree holds the uncommitted spec change by design, so a diff
against HEAD is expected. The hash-equality run above is the regeneration
proof.)

### Existing cases byte-identical

`git diff --numstat`: `conformance/v1/vectors/registry-service.json`
86 insertions / 0 deletions (pure `checkpoint_cases` insertion);
`manifest.json` 1/1 (one digest line); `release/1.0.0-rc.9.json` 2/2 (the
two manifest pins, both now
`sha256:fd28b0601eb5a62c24ecb56cc5bd312a4a8274d6f9bedbbaa2ce051f158aae8a`).
Semantic check against the pre-regeneration snapshot
(`git show HEAD:conformance/v1/vectors/registry-service.json`):

```
before keys: [artifact_key, cache_cases, idempotency_cases, limits, pagination, query_cases, records, recovery_cases, restore_cases, snapshot, sort_key, transaction_cases, transport_cases]
after keys: [artifact_key, cache_cases, checkpoint_cases, idempotency_cases, limits, pagination, query_cases, records, recovery_cases, restore_cases, snapshot, sort_key, transaction_cases, transport_cases]
new keys: ['checkpoint_cases'] missing keys: []
changed existing keys: []
checkpoint_cases: 7 [checkpoint-below-live-consistent, checkpoint-equal-consistent, checkpoint-equal-inconsistent, live-below-checkpoint, live-above-prefix-mismatch, checkpoint-signature-invalid, checkpoint-not-configured]
```

All seven brief-mandated outcomes verified in the emitted vectors:
below-live-consistent ready, equal-consistent ready, equal-inconsistent not
ready (`restore_inconsistent_with_checkpoint`), live-below not ready
(`restore_below_checkpoint`), above-prefix-mismatch not ready
(`restore_inconsistent_with_checkpoint`), bad signature not ready
(`checkpoint_signature_invalid`), not-configured ready with
`checkpoint_not_configured` posture.

### make validate

`make validate` = `python3 tools/validate.py` + `python3 -B -m unittest
discover -s tools -p 'test_*.py'` + `go test ./tools/...`, run with the
repo venv first on `PATH`. The Go third runs as a `go test -c`-compiled +
locally re-signed test binary for the host reason above (same test code
the Makefile runs).

```
$ python tools/validate.py
validated 62 schemas and 1071 vector files
VALIDATE_PY_EXIT=0
```

```
$ python -B -m unittest discover -s tools -p 'test_*.py'
......................................................................
----------------------------------------------------------------------
Ran 310 tests in 320.230s

OK
UNITTEST_EXIT=0
```

```
$ go test -c -o /tmp/gv2.test ./tools/generate-vectors && codesign --force --sign - /tmp/gv2.test && /tmp/gv2.test
READY
PASS
GOTEST_EXIT=0
```

`tools/generate-vectors` is the only Go package with tests under `tools/`
(all `*_test.go` live there), so this is the full `go test ./tools/...`
third. `gofmt -l tools/generate-vectors/` prints nothing (clean). First
attempt at the test binary stalled in the same pre-main exec gating; a
fresh rebuild + re-sign exec'd and passed — the stall is flaky per binary,
not deterministic.

All three `make validate` thirds exit 0 with the venv python and the
re-signed Go test binary standing in for `go run`/`go test` exec (host
exec-gating workaround; identical code and inputs).

## Deliberately out of scope

- Implementation (`TASK-260910-1ny7yl` `serve --checkpoint`): no product
  code touched. Spec-repo conformance tooling (`tools/generate-vectors`,
  `tools/validate.py`, `tools/test_validate.py`) is edited because the
  brief requires generated + pinned vectors; that tooling is the spec's own
  conformance surface, not the service implementation.
- S2 bootstrap checkpoint interchange for clients (`TASK-260910-1tvf2t`):
  the signed-snapshot object is referenced, not respecified; no
  client-side bootstrap defined here.
- P4 import high-water, tags/releases, ax, proposals 0014–0018.
- No schema change (frozen `health-response-v1` untouched; no new schema,
  knob, or lock key — the profile has no §12 knob table).

## Checklist mapping

1. Normative enforcement point in §6 with the four closed diagnostics
   spelled identically in text, table, vectors, CHANGELOG; §5/§9/§10/§11
   cross-referenced — satisfied.
2. Settled decisions honoured (ordering, signed snapshot, no-checkpoint
   posture, direct rollout, frozen health schema) — satisfied.
3. `checkpoint_cases` added and pinned by `tools/validate.py`; existing
   cases byte-identical; manifest + rc.9 pins regenerated; `make validate`
   and regeneration proof exit 0 — satisfied (transcripts above).
4. CHANGELOG Unreleased R3/P2 entry; spec-patch + evidence attached as task
   outcome resources; no implementation code touched — satisfied.
5. Docs updated and consistent with current code — satisfied (no other doc
   references the old §6 wording; SECURITY.md/COMPATIBILITY.md/cli wording
   is client-side or unrelated).
6. No discrepancies between code and description — satisfied (vectors match
   the §6 rules; validator recomputes them).
7. Result linked as a new task-scoped outcome resource — satisfied (patch
   + this evidence file attached on the task).
8. Logbook entries when relevant — one host anomaly found (fresh Mach-O
   binaries stall pre-main in exec gating on this host; local ad-hoc
   re-sign usually unblocks exec; flaky per binary — details and proof in
   the validation transcript above). Recorded here and in the task notes;
   no LOGBOOK.md edit per the campaign's no-LOGBOOK-edits rule for spec
   work.
