# Evidence — TASK-260910-2ohnjo: spec bounds for MCP env passthrough and declaration surfacing (S4), rev3

Story `STORY-260910-1lf0m5`, wave 1. Role: doc-writer.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`
(branch `task-board/story/STORY-260910-1lf0m5`, base `main` @ `07e2b41`).

Review verdict rev2 (`TASK-260910-2ohnjo_review-verdict-rev2.md`) read first for
this round; everything it marks PASS/CLOSED is unchanged except where this
file says otherwise. Rework brief rev3 is the authority for this round: one
medium correction, nothing else.

Role note: the rework brief rev3 explicitly requires edits to
`tools/validate.py` and `tools/test_validate.py` (spec-validation tooling,
not product implementation). No product code, tags, or proposals touched.
No normative text changed: `protocol/environments.md` §2.3 keeps the JSON
string argument grammar (spaces permitted inside strings).

## Closure of the rev3 correction (medium)

`tools/validate.py` `s4_parse_surfacing_row` (`:4523`) split the row with
`row.split(" ")` and required exactly seven tokens, so a valid argument
containing a space (e.g. `"hello world"`, permitted by
`schemas/v1/agent-mcp-v1.schema.json` string args) produced eight tokens and
the consumer rejected a correctly rendered positive row — CLOSED:

1. `tools/validate.py:4523-4585` — the parser now tokenizes the row by the
   closed column layout: the `marker package version transport` head is split
   with `maxsplit=4`, `command=` is taken up to the next single space, and
   the `args` / `env_names` JSON arrays are decoded with
   `json.JSONDecoder().raw_decode` from the exact `args=` /
   ` env_names=` offsets, so spaces and escapes inside string values are
   preserved. The compact re-serialization check still rejects separator
   whitespace; missing, reordered, or extra columns are still rejected
   (exact `command=` / `args=` / ` env_names=` prefixes in order, end of
   string after the second array, explicit `\n`/`\r` rejection).
2. `tools/validate.py:4444-4451` — `S4_SURFACING_CASES` gains
   `args-with-space` and `args-with-escaped-quote` (6 cases; negatives
   unchanged).
3. `conformance/v1/vectors/environments-env-passthrough.json:262-305` — two
   new positive surfacing cases (insertion only, everything else
   byte-identical): `args-with-space` mirrors the reviewer's repro
   (`args=["-y","figma-developer-mcp","--stdio","hello world"]`, 140 B,
   `sha256:81889e…47bbd`); `args-with-escaped-quote` carries an escaped
   quote (`args=["say \"hi\"","--stdio"]`, 112 B,
   `sha256:8416c6…9770`). The agent-mcp-v1 grammar allows any string
   (maxLength 8192), so the escaped-quote arg is valid input.
4. `tools/test_validate.py:1883-1945` — six new tests: spaced arg parses
   with exact values, escaped quote parses, a delimiter-like
   `env_names=[...]` marker inside a string parses (structural, not naive
   split), extra column rejected, padded JSON rejected, and both new cases
   present and passing the gate. Both round-1 review mutants are kept
   verbatim (`:1737`, `:1747`).
5. `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json` — regenerated
   pins only (`go run ./tools/generate-vectors -root .`): the
   env-passthrough vector sha and the two manifest pins.

Old-vs-new probe (real code, quoted output): the spaced row splits into 8
tokens under the old logic (`old split token count: 8 (old code required
exactly 7 -> rejected)`); the new parser accepts it (`True`) and still
rejects the same row with ` extra=1` (`True`).

## What changed per file, and why (rev3 delta on top of rev2)

1. `tools/validate.py` — structural surfacing-row parser (`:4523-4585`);
   inventory +2 (`:4444-4451`). Nothing else in the file touched.
2. `conformance/v1/vectors/environments-env-passthrough.json` — 25 cases
   (was 23): the two positive cases above (`:262-305`). All other cases,
   bytes, lengths, and hashes unchanged.
3. `tools/test_validate.py` — 6 new tests (`:1883-1945`); 27
   `EnvPassthroughVectorTests` (was 21).
4. `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json` — regenerated
   pins only.
5. Carried from rev2 unchanged: every normative sentence, default, order,
   diagnostic, fixture, schema case, generator line, and the CHANGELOG entry.

Closed-set spellings unchanged; no new diagnostic, knob, or lock key. The
rev2-vs-rev3 patch file lists are identical (same 11 paths; +136 lines).

## Validation transcript (rev3)

Shell: `/bin/sh` via the run tool on macOS, workdir = the story worktree.
Repo venv first on `PATH`
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`).
`set -o pipefail` on every gate call; exit codes are the processes' own.
No step piped through `tee`.

The headless run cannot wait out the literal single-call `make validate`
(~9 min wall in rev2) inside one time-bounded shell call, so its three
components were run as bounded sequential calls covering the identical set;
nothing was accepted from prior evidence — every number below is this
round's own rerun:

1. `python3 tools/validate.py` → exit 0:
   `validated 60 schemas and 1048 vector files` (the S4 consumer runs inside
   this entry point and passes with the two new vectors).
2. `python3 -B -m unittest` over every `tools/test_*.py` class, in splits:
   - small files (`test_implementation_coverage`,
     `test_release_gate`, `test_verify_release_commit`,
     `test_verify_release_merge_policy`) → exit 0: `Ran 78 tests … OK`
     (192.315 s);
   - `test_validate` classes Wire…WorkflowRegeneration (7 classes) →
     exit 0: `Ran 56 tests … OK` (16.012 s);
   - `test_validate` classes Environment…SystemConfigV2 (7 classes,
     incl. all 27 `EnvPassthroughVectorTests` with both review mutants
     and the 6 new parser tests) → exit 0: `Ran 120 tests … OK`
     (128.435 s).
   Total: 254 tests OK (248 rev2 + 6 new), every class of every test file.
3. `go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 8.620s`.
4. Literal `make regenerate-check` → exit 0. Same staging procedure as
   rev2 (the gate diffs the working tree against the git index, so the
   generated paths were staged with `git add conformance/v1
   release/1.0.0-rc.9.json` — stage only, no commit): `go run
   ./tools/generate-vectors -root .` then `git diff --exit-code -- …`,
   `regen_check_exit=0`. Nothing was committed.
5. `git diff --check` → exit 0 (no whitespace errors).
6. Patch self-check against the candidate base `07e2b41` (NOT
   `origin/main`: `main`/`origin/main` moved to `f544a01` with unrelated
   landed changes, so the patch is `git diff 07e2b41` after
   `git add -N .`): live diff and attached
   `TASK-260910-2ohnjo_spec-patch_rev3.patch` share stable patch-id
   `3fcb1de6fe40f2753141b2f8e69fd3ea2c2a3be1`; `git apply --check`
   against archived `07e2b41` in a scratch directory passes (exit 0).
7. `git status --short` lists only the 11 spec/schema/vector/manifest/
   CHANGELOG/tooling paths (tooling explicitly required by the rework
   briefs); no implementation code, no unrelated edits.

## Deliberately out of scope

Everything the producer brief and rework rev3 exclude: implementation
(`TASK-260910-gocke2`), E3 codex seed, S1/S3 hardened defaults, proposals
0014–0018, the closed interpreter contract (still a noted future revision),
`LOGBOOK.md` (untouched per campaign rules). No normative sentence was
narrowed to fit the parser.

## Run anomaly (no impact on the artifact)

At ~03:14 local, new launches of the current board binary
`task-board-main-ac2ad9c0-curatorlike` (the `task-board` wrapper target)
began sticking pre-main in `_dyld_start` (0 CPU, confirmed via `sample`;
both reads and writes affected, another agent's invocation stuck the
same way). Python, git, and go launched normally. The stall is selective
to that binary's dyld launch — the same signature as the rev2 run's
transient Go-binary stall. Board writes for this round (resource attach,
handoff) were therefore executed through the previous build
`task-board-main-6cb09a23-curatorlike`, which launches and operates
normally; the stuck invocations were terminated before retrying so no
duplicate write can land if the stall clears. All validation above ran
before and independently of the stall.

---

# Prior evidence (rev2, retained)

# Evidence — TASK-260910-2ohnjo: spec bounds for MCP env passthrough and declaration surfacing (S4), rev2

Story `STORY-260910-1lf0m5`, wave 1. Role: doc-writer.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`
(branch `task-board/story/STORY-260910-1lf0m5`, base `main` @ `07e2b41`).

Finding read first: `docs/security-audit-2026-09.md` S4 (spec) and the
manager-side S4 in `curator/docs/security-audit-2026-09.md`. Review verdict
rev1 (`TASK-260910-2ohnjo_review-verdict-rev1.md`) read first for this round;
everything it marks PASS is unchanged except where this file says otherwise.

Role note: the rework brief rev2 explicitly requires edits to
`tools/generate-vectors/manager_config.go`, `tools/validate.py`, and
`tools/test_validate.py` (spec-validation tooling, not product
implementation). Those edits are disclosed per file below. No product code,
tags, or proposals touched.

## Closure of review findings 1–4

1. Generator drift (high) — CLOSED. `tools/generate-vectors/manager_config.go:22`
   now emits `"passable_env_names": []any{}` (was `nil`), matching the
   schema (`schemas/v1/manager-config-v2.schema.json`, default `[]`) and
   §12.1. The four schema cases are now produced by
   `managerConfigV2Vectors()` (`manager_config.go:316-328`, appended in the
   same order as the rev1 hand-written cases), not hand-written. `make
   regenerate` was run; the regenerated `manager-config-v2.json` is
   byte-identical to the rev1 hand-edited file (verified with `diff` before
   any other rev2 edit), and `make regenerate-check` exits 0 (see
   transcript). Rev1 finding F1 is therefore closed, not a follow-up.
2. One consistent surfacing order (high) — CLOSED. §9.2
   (`protocol/environments.md:1696-1717`) now lists surfacing as step 3 of
   the update order — "after the audit gate passes and before the lock is
   published or any surface is re-materialized" — with publish moved to step
   4 and the old contradictory "After publishing the new lock" paragraph
   removed. Install (§9.1, `:1649-1652`) and §2.3 (`:480-484`) already state
   the same sequence, so install and update now share one order. The update
   order is covered by the new `surfacing_order_cases` vector family
   (`vectors/environments-env-passthrough.json:275-297`): install and update
   both pin `["audit-gate","surfacing","lock-published","materialization"]`.
   §13 (`:2426-2430`) registers the order coverage.
3. S4 vectors semantically checked (high) — CLOSED.
   `tools/validate.py:4403-4789` adds `validate_environments_env_passthrough_vectors`
   (`:4583`), wired into `main()` at `:4971` directly after
   `validate_environment_vectors`. It recomputes from the declared inputs:
   the effective `passable_env_names` per profile/knob state
   (`s4_effective_passable`, `:4463`), the passed/dropped sets and
   diagnostics (`s4_expected_passthrough`, `:4562`), the allowlist warning
   verdict and `admitted` wording, the surfacing output bytes, byte length,
   and sha256 from the declarations (`s4_surfacing_row`, `:4486`; row parser
   `s4_parse_surfacing_row`, `:4521`, enforcing the closed ordered columns
   and compact JSON), the pinned install/update order, and the schema cases
   against the real `manager-config-v2` grammar plus the `[]` schema
   default. Positive cases must equal the recomputation; negative cases
   (pinned by name) must carry `conforming: false`, a reason, and an
   observation that genuinely contradicts the recomputed rule — a repaired
   observation or flipped flag fails. Case inventories are exact per family.
   `tools/test_validate.py:1705-1911` adds `EnvPassthroughVectorTests`
   (21 tests), including both review round-1 mutants (enforcing absent-knob
   flipped to pass `FIGMA_API_KEY`; surfacing bytes replaced with
   `INVALID\n`, with stale AND refreshed length/hash pins) and
   narrowed-list bypasses under both profiles. Mutant score on the new
   consumer: 2/2 review mutants rejected (was 0/2).
4. Contradictory negative fixture (medium) — CLOSED. `non-empty-allowlist-silent`
   (`vectors/environments-env-passthrough.json:28-38`) is now a positive
   (`diagnostic: null`, `admitted: "only listed declaration packages"`,
   `fails_operation: false`), and the new negative `non-empty-allowlist-warns`
   (`:39-49`) carries the emitting observation
   (`diagnostic: mcp_package_allowlist_empty`, `conforming: false`) matching
   its verdict; both are asserted through the consumer of item 3. The
   `explicit-null-treated-as-empty` negative (`:164-177`) gained its concrete
   contradicting observation (`requested: ["FIGMA_API_KEY"]`, `passed: []`)
   so the consumer can assert it; its verdict and reason are unchanged.

## What changed per file, and why (rev2 delta on top of rev1)

1. `tools/generate-vectors/manager_config.go` — S4 default `nil` → `[]any{}`
   (`:22`); four `schema2-passable-env-*` cases generated (`:313-328`).
2. `protocol/environments.md` — §9.2 update order unified (`:1696-1717`,
   surfacing before lock publish for update as for install); §13 registers
   the order vectors (`:2426-2430`). All other rev1 text unchanged.
3. `conformance/v1/vectors/environments-env-passthrough.json` — 23 cases
   (was 20): allowlist silence positive + emitting negative (`:28-49`),
   null-negative observation (`:172-176`), `surfacing_order_cases` (`:275-297`),
   `fails_operation: false` on the silent positive (`:31`). Surfacing bytes
   unchanged (126 B `sha256:fec4…0bb0`, 159 B `sha256:21779…91af`).
4. `tools/validate.py` — S4 consumer + `main()` wiring (`:4403-4789`, `:4971`).
5. `tools/test_validate.py` — `EnvPassthroughVectorTests`, 21 tests (`:1705-1911`).
6. `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json` — regenerated
   pins only (`go run ./tools/generate-vectors -root .`).
7. Carried from rev1 unchanged: `CHANGELOG.md` (Unreleased "S4 …" entry),
   `profiles/manager.md` consistency lines, `schemas/v1/manager-config-v2.schema.json`
   (default `[]`), `vectors/manager-config-v2.json` content (now
   generator-produced, byte-identical), §2.2/§2.3/§10.3/§12 text.

Closed-set spellings verified identical in prose, tables, schema, vectors,
CHANGELOG, and consumer: `mcp_package_allowlist_empty`,
`mcp_env_passthrough_unlisted`, `mcp_env_passthrough_dropped`,
`passable_env_names`, `s4-warn`, `s4-enforce`. No new knob or lock key.

## Validation transcript (rev2)

Shell: `/bin/sh` via the run tool on macOS, workdir = the story worktree.
Repo venv first on `PATH` (`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`).
No step piped through `tee`; exit codes are the processes' own (`PIPESTATUS[0]`
where a `tail` applied).

1. `python3 tools/validate.py` → exit 0 (13 s):
   `validated 60 schemas and 1048 vector files`
2. `python3 -B -m unittest discover -s tools -p 'test_*.py'` → exit 0:
   `Ran 248 tests in 204.690s — OK` (227 prior + 21 new S4 tests)
3. `go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.285s`
4. Literal `make validate` in one call → exit 0 (9m9s wall; machine loaded):
   `Ran 248 tests in 531.313s — OK`, `go test ./tools/... ok`, `make_validate_exit=0`
5. `make regenerate-check` → exit 0. Note: the Makefile gate diffs the
   working tree against the git index, so with an uncommitted candidate it
   first showed the candidate delta itself (exit 2, expected — the diff was
   exactly the S4 delta, nothing else). After staging the generated paths
   (`git add conformance/v1 release/1.0.0-rc.*.json`; stage only, no commit),
   regeneration reproduces the candidate bytes exactly and the gate exits 0:
   `go run ./tools/generate-vectors -root .` then `git diff --exit-code -- …`,
   `regen_check_exit=0`. Nothing was committed; the worktree remains an
   uncommitted candidate for handoff as required.
6. `git diff --check` → exit 0 (no whitespace errors).
7. Patch self-check: `git diff origin/main | git patch-id --stable` =
   `78c86592a0024537c22842b25b1ef7c918e33538`, identical for the attached
   patch file; `git apply --check` against base `07e2b41` in a scratch clone
   passes.

## Deliberately out of scope

Implementation (`TASK-260910-gocke2`), E3 codex seed, S1/S3 hardened
defaults, proposals 0014–0018, the closed interpreter contract (noted as a
future revision in §10.3), `manager-config.json` (byte-frozen schema-1
family, untouched). `LOGBOOK.md` untouched per campaign rules.

## Notes

- Mid-run anomaly (no impact on the artifact): between ~01:17 and ~01:30 the
  machine refused to start the cached `generate-vectors` binary (new process
  stuck pre-main, 0 CPU; Apple-signed binaries still ran). No repo state was
  changed by the stall; after it cleared, the same binary ran normally
  (verified with `--help`, then the successful regenerations above). Stuck
  processes were killed; none remain (`pgrep` clean).
- The `non-empty-allowlist-warns` negative and the `admitted` wording
  `"only listed declaration packages"` are new closed strings pinned by both
  the vector file and the consumer; the reviewer should confirm the wording.
