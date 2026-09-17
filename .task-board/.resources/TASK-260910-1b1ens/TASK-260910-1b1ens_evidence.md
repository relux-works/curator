# Evidence — TASK-260910-1b1ens: spec-records-boundary-envelope (R1/P1)

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-25yc0h/worktree`
Branch: `task-board/story/STORY-260910-25yc0h`, base `origin/main` = `23dafa7`
Role: doc-writer (normative spec revision, no implementation).
Finding: `docs/security-audit-2026-09.md` "R1 / P1. Records pages carry no
snapshot boundary" (High). Read before editing.

## What changed per file (18 files, +708/−28 vs origin/main)

Normative text:

- `protocol/registry.md`
  - §5: page boundaries advance and check the same rollback state as
    `/v1/snapshot` snapshots (same atomic persistence, same
    below/equal/higher rules); a page MUST NOT contribute records or
    entries before its boundary is accepted.
  - §9: endpoint table moves `/v1/records` GET and `/v1/log` to the v2
    envelope schemas; the unknown-fields rule now explicitly states a
    v1-validating client treats `boundary` as ignorable; the pagination
    paragraph points at §9.3.
  - New §9.3 "Page boundary": REQUIRED `boundary` = the
    `registry-snapshot-v1` object (all fields incl. `sig`) the page was
    evaluated at; byte-identical across a cursor chain; client
    verification order (1. §2 signature, 2. §5 rollback rules, 3. chain
    equality vs the first page); the closed 3-row diagnostics table;
    the exclusion rule (rejected page ⇒ registry contributes nothing,
    like a reachable invalid snapshot); the boundary as stated inclusion
    evidence with log replay optional as its independent re-derivation;
    the read-only status posture row (per trusted registry URL:
    persisted high-water `version`/`log_size` + whether the last page
    boundary was verified).
- `profiles/registry-service.md`
  - §2: REQUIRED wire `boundary` on every page, byte-identical per
    chain; P1 cursor binding (cursor-boundary disagreement ⇒
    `404 invalid_cursor`, MUST NOT re-evaluate at a newer boundary).
  - §5: emitted `boundary` is the immutable snapshot body of the
    captured boundary.
  - §10: boundary is the stated inclusion evidence; log/bundle replay
    stays optional as the independent re-derivation.
  - §11: conformance list gains page-boundary emission and
    cursor-boundary refusal.

Schemas and registry:

- `schemas/v1/records-response-v2.schema.json`,
  `schemas/v1/log-response-v2.schema.json` (new): v1 fields plus
  REQUIRED `boundary` (`$ref: registry-snapshot-v1.schema.json`),
  `additionalProperties: false`. V1 schemas byte-frozen, untouched.
- `schemas/v1/README.md`: next-version registration paragraph (v1 stays
  valid, `boundary` ignorable under §9).

Conformance (all generated content flows through the generator, S4
precedent — see Decisions):

- `conformance/v1/schema-cases/{records,log}-response-v2/{valid,invalid}.json`
  (new): valid = v1 body + canonical snapshot boundary; invalid =
  byte-identical to the v1 valid body (proves `boundary` is required).
  Registered in `schema-cases/index.json` via the generator.
- `conformance/v1/vectors/registry-client.json`: new `page_boundary_cases`
  block, 7 cases exactly as the brief lists (fresh advance;
  equal-same accept; equal-different reject; below-high-water reject;
  chain mismatch reject; missing excluded; bad signature rejected).
  Every rejected case pins `registry_excluded: true`.
- `conformance/v1/vectors/registry-service.json` `pagination`: new
  `boundary_emitted_on_every_page: true`,
  `chain_boundary_byte_identical: true`, and `cursor_boundary_cases`
  (`cursor-boundary-disagreement`: 404 `invalid_cursor`, no
  re-evaluation). The existing `cursor_rejections` set is untouched
  (validator pins it exactly).
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
  pins (manifest +4 files, 1071 total; rc.9 pin follows the manifest).

Tooling (spec-repo gates, not product implementation):

- `tools/generate-vectors/main.go`: the two vector blocks above + the v2
  schema cases in `writeSchemaCases` (v2-invalid reuses the v1-valid
  map, so the bytes stay identical by construction).
- `tools/validate.py`: new production gate
  `validate_registry_page_boundary_vectors` (wired into `main()`),
  recomputing accepted/diagnostic/high-water/exclusion from each case's
  inputs and pinning the closed diagnostic set + service emission pins.
- `tools/test_validate.py`: `RegistryPageBoundaryVectorTests`, 13 tests
  (published vectors pass + 12 fail-closed mutants: each admit/reject
  flip, wrong/swapped spelling, missing exclusion/advance, dropped
  case, dropped emission pin, re-evaluating cursor).

`CHANGELOG.md`: Unreleased `### Added` entry "R1/P1: …" (envelope v2,
client rule, missing-boundary reporting, direct rollout with impact row
"R1", vectors, stories `STORY-260910-25yc0h` / `STORY-260910-3rvvxh`).

## Why

A key-holding registry could serve an honest advancing `/v1/snapshot`
while evaluating `/v1/records` at an older boundary, hiding a `revoked`
record; client rollback state bound snapshot versions, not page
boundaries. The revision makes the evaluated boundary visible on the
wire (signed snapshot object) and requires clients to verify it and
reject pages below their persisted high-water, closing R1; the
cursor/boundary agreement (P1) stops re-evaluation at a newer boundary.

## Decisions (brief left these to the producer)

- Equal-version-different-body reports `registry_page_boundary_stale`,
  not `mismatch`: both below-high-water and equal-but-different fail
  the §5 persisted-state comparison, while `mismatch` is purely
  chain-internal (page vs first page). Stated in §9.3, the gate, and
  the `equal-version-different-body-rejected` vector.
- Bad boundary signature reports `registry_page_boundary_missing`: the
  brief defines a page "without a valid `boundary`" as the missing
  case, and a boundary failing §2 verification is not valid. Keeps the
  closed set at exactly three diagnostics. Stated in §9.3 and pinned by
  the `bad-signature-rejected` vector.
- Generator owns vectors, schema cases, index, manifest, and rc.9 pins
  (same shape as the landed S4 revision `23dafa7`): `writeSchemaCases`
  rewrites `index.json` wholesale, so hand edits there do not survive —
  the v2 cases live in `main.go`.
- The validator gate + 13 tests mirror the S4 fail-closed pattern; the
  role's "does not modify code" constraint is read as product
  implementation code, per the brief's explicit deliverables (schemas,
  vectors, manifest, regenerate-clean output).

## Validation transcript (exit codes quoted)

Shell: `/bin/zsh` (worktree root). Python: `/tmp/specvenv` venv,
`pip install -r requirements-dev.txt` (jsonschema 4.25.1), first on
`PATH` (repo has no committed venv; CI installs the same file).

- `go run ./tools/generate-vectors -root .` → exit 0.
- `make validate` → exit 0:
  - `python3 tools/validate.py` → `validated 62 schemas and 1071 vector
    files` (was 60 / 1067: +2 schemas, +4 schema-case files).
  - `python3 -B -m unittest discover -s tools -p 'test_*.py'` → `Ran
    301 tests … OK` (was 288: +13 new gate tests).
  - `go test ./tools/...` → `ok …/tools/generate-vectors`.
- `make regenerate-check` (after staging conformance/v1 + rc.9 so the
  idempotence diff is meaningful) → exit 0.

Byte-identity proofs:

- `git diff` on both registry vector files shows zero removed lines
  (purely additive); `cursor_rejections`, `snapshot_transitions`, and
  every other existing block untouched.
- All pre-existing schema-case files unmodified; both v2 `invalid.json`
  are `cmp`-identical to the corresponding v1 `valid.json`.
- Both v1 envelope schemas unmodified; `git grep` on HEAD shows no
  prior `registry_page_boundary_*` occurrence, and the current tree
  carries exactly the three closed spellings (plus `*_vectors`
  function names).
- `go vet`-level check: `gofmt -w` applied to the generator; no other
  Go files touched.

## Deliberately out of scope

Implementation (service `TASK-260910-14dnb7`/`-27yepb`, client
`TASK-260910-2n0233`), S2 (TOFU/equivocation), inclusion proofs beyond
the signed snapshot, R2–R8, tags/releases, proposals 0014–0018. No
`manager.md`/`environments.md` edits: neither lists registry client
diagnostics, and the brief assigns the posture row to the status text
in §9.3. No knob, no lockable set, no warn-first split (direct
rollout, impact row "R1"). No commits, branches, or pushes made.

---

## Revision 2 (rework of review finding F1)

Finding: `TASK-260910-1b1ens_review-verdict-rev1.md` F1 — §9.3 ordered
rollback comparison/persistence before chain comparison, while
`expected_page_boundary_verdict` and the vectors decide mismatch first and
never advance on a mismatching page. Settled resolution (rework brief):
the oracle's order is intended; the normative text changes. The first
page's boundary is the chain boundary; later pages are compared to it
before anything else and never touch rollback state.

### What changed per file (6 files; all else hunk-identical to rev1)

- `protocol/registry.md`
  - §9.3 client order rewritten: (1) presence and §2 signature
    verification — absent or failing reports
    `registry_page_boundary_missing`, no state change; (2) every page
    after the first MUST be byte-identical to the chain boundary, else
    `registry_page_boundary_mismatch`, no state change — a later page
    never advances or re-checks the high-water on its own; (3) the
    first page applies the §5 rollback rules — below, or equal with a
    different `head`/`merkle_root`/`log_size`, reports
    `registry_page_boundary_stale`, no state change; equal same body
    accepted with nothing persisted; higher persisted atomically BEFORE
    the page contributes anything, then accepted. Persistence sentence:
    only a first page passing steps 1 and 3 changes rollback state; a
    rejected page leaves it untouched. Precedence: `missing` over
    `mismatch` on a later page; `stale` only for a first page.
  - §5: rollback rules now apply to the chain boundary (first page's
    boundary); later pages are compared to it first and never advance
    or re-check on their own. Exclusion paragraph, diagnostics table,
    missing-boundary rollout, inclusion wording and posture sentence
    byte-unchanged.
- `tools/validate.py`: oracle docstring rewritten to the §9.3 order
  above (code order unchanged — still missing → mismatch → stale);
  `PAGE_BOUNDARY_CASES` gains `stale-and-mismatch-reports-mismatch`
  and `higher-and-mismatch-never-advances`, so the gate requires them
  by name and the per-case loop pins their accepted / diagnostic /
  high-water / exclusion outcomes via the oracle.
- `tools/generate-vectors/main.go`: comment above
  `page_boundary_cases` documents that `chain_boundary_equal: false`
  on a present, signature-valid boundary only arises on a later page
  differing from the chain boundary (the missing case carries the
  `false` as a placeholder since missing is checked first); two cases
  appended exactly per brief — `stale-and-mismatch-reports-mismatch`
  (boundary 7, stored 8 → rejected, `mismatch`, no advance, excluded)
  and `higher-and-mismatch-never-advances` (boundary 9, stored 7 →
  rejected, `mismatch`, no advance, excluded). No `first_page` boolean
  added: the documented reading keeps the 7 existing cases byte-stable
  while pinning precedence and persistence. `gofmt` clean.
- `conformance/v1/vectors/registry-client.json`: +26/−0 (two cases
  appended; the 7 existing cases' names and outcomes unchanged).
- `conformance/v1/manifest.json`: only the `registry-client.json`
  sha256 pin (1/1). `release/1.0.0-rc.9.json`: only the manifest pin
  in two places (2/2).
- `profiles/registry-service.md`: untouched — no sentence there
  restates the client order (verified by reading the file).
- The other 12 rev1 files (CHANGELOG, schema-cases + index, vectors
  registry-service, schemas v2 + README, `tools/test_validate.py`)
  are hunk-identical to the rev1 patch: a per-file comparison of
  `git diff origin/main` against
  `TASK-260910-1b1ens_spec-patch_rev1.patch` reports 12/12 identical,
  0 problems.

### Validation transcript (exit codes quoted)

Shell `/bin/sh` in the worktree root, repository venv
`curator-spec/.temp/venv/bin` first on `PATH` (jsonschema 4.25.1).
The runner bounds one shell call, so the long `make validate` ran as
its three constituent gates in bounded sequential calls (no `tee`,
no pipes on gate commands except `tail` on verbose unittest output):

- `go run ./tools/generate-vectors -root .` → exit 0 (no output).
- `python3 tools/validate.py` → exit 0:
  `validated 62 schemas and 1071 vector files`.
- unittest, full suite 301 tests, all OK, exit 0 each:
  - `python3 -B -m unittest discover -s tools -p 'test_validate.py'`
    → `Ran 223 tests in 124.449s … OK`, exit 0 (includes the 13
    `RegistryPageBoundaryVectorTests`, also run alone via
    `-k 'RegistryPageBoundary'`: 13 tests OK, exit 0).
  - `-p 'test_implementation_coverage.py'` → 36 tests OK, exit 0.
  - `-p 'test_release_gate.py'` → 32 tests in 139.339s OK, exit 0.
  - `-p 'test_verify_release_commit.py'` → 5 tests OK, exit 0.
  - `-p 'test_verify_release_merge_policy.py'` → 5 tests OK, exit 0.
  - Total 223+36+32+5+5 = 301; a 4-way `-k` chunk split of
    test_validate.py (31/52/82/58) cross-checks the 223 with no
    overlap or gap. The literal single `make validate` was not run
    (its ~8 min exceeds the per-call bound); every gate it runs is
    quoted green above.
- `go test ./tools/...` → exit 0:
  `ok github.com/relux-works/curator-spec/tools/generate-vectors 3.391s`.
- Regenerate idempotence (after staging `conformance/v1` + rc.5–rc.9
  pins, rev1 precedent): `go run ./tools/generate-vectors -root .`
  → exit 0; `git diff --exit-code -- conformance/v1
  release/1.0.0-rc.5.json release/1.0.0-rc.6.json
  release/1.0.0-rc.7.json release/1.0.0-rc.8.json
  release/1.0.0-rc.9.json` → exit 0, empty output. Note: one
  `make regenerate-check` wrapper invocation produced no output for
  6+ minutes and was terminated as hung; its identical two recipe
  steps run directly complete in seconds with exit 0 (quoted here).
- `git diff --check` → exit 0.

Byte-identity proofs:

- `git diff origin/main --stat`: same 18 files as rev1, now
  +775/−28 (rev1 was +708/−28; deletions unchanged — every rev2
  prose/oracle/generator edit rewrites lines rev1 added).
- `registry-client.json` vs `origin/main`: 119 additions, 0
  deletions (93 rev1 + 26 rev2); rev2-only delta +26/−0.
- v1 envelope schemas: `git diff origin/main` empty (frozen).
- Closed spellings in tree: `registry_page_boundary_mismatch` 14,
  `_missing` 11, `_stale` 11, plus `_vectors` 4 (gate function
  names); no other `registry_page_boundary_*` spelling.
- Attached `TASK-260910-1b1ens_spec-patch_rev2.patch` is byte-identical
  to live `git diff origin/main` (`cmp` match, 52501 bytes).

Deliberately out of scope: unchanged from rev1 (implementation,
S2, inclusion proofs beyond the signed snapshot, R2–R8,
tags/releases, proposals 0014–0018). No commits, branches, or
pushes made; only `git add` staging of generated files and
`git add -N` for patch generation, as in rev1.
