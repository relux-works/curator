# Review verdict — TASK-260910-1xya7x revision 4 (CR-TASK-260910-1xya7x-4)

Reviewer run: RUN-260920-90c477 (claude-opus-5, reviewer archetype), 2026-09-20.
Verdict: **ACCEPT** — `accept_cr(TASK-260910-1xya7x, revision=4, evidence=this file)`.

Revision 4 is exactly the rework-3 delta (R1–R4 of `TASK-260910-1xya7x_review-verdict-rev3.md` §11):
four test files, no production, spec or frozen file. Every R item is met and re-measured below; the
§10 facts of the rev3 verdict were re-verified where rework 3 touched them (runner, the four re-labelled
rows, transaction e2e wording) and hold. The four product gaps stay filed as follow-ups per the
orchestrator ruling (STORY-260919-37szes) and do not block this leaf.

## 1. Exact candidate identity (verified)

- Story worktree `.temp/STORY-260910-1cnwwp/worktree`: HEAD `71e353e6` (stbg4d rev2 checkpoint on base
  `97a161da`); temp-index `git read-tree HEAD && git add -A . && git write-tree` =
  `01a3b519fab0a0316415799f1d694aa712e76f8f` = the CR-4 candidate tree OID. Worktree not modified.
- Patch `TASK-260910-1xya7x_change-request_rev4.patch` sha256
  `188376d2df6804f05a70db722c74fcd7edaaf0197e245b557287602514c87a04` (matches the CR record).
- Disposable clones `/tmp/1xya7x-rev4/cand` and `/tmp/1xya7x-rev4/probe`: `git checkout --detach 97a161da
  && git apply --index <rev4 patch> && git commit` → `HEAD^{tree}` = `01a3b519…` in both; every driver
  prints `git write-tree` before/after and both clones ended at `01a3b519…`.
- rev3 → rev4 delta (`git diff --stat 2d400147 01a3b519`): `draftsources_semantic_test.go` (+102/−0),
  `draftsources_semantic_evidence_test.go` (16), `draftsources_semantic_v2_test.go` (16),
  `draftsources_transaction_e2e_test.go` (24) — 132 insertions, 26 deletions, nothing else.
- Leaf delta vs the checkpointed sibling (`git diff --name-status 71e353e6 01a3b519`): **153 paths, all
  `A`** (13 865 insertions, 0 deletions/modifications), all under `internal/crossconformance/` and
  `internal/testcli/`. No production file, no `conformance/v1`, no `schemas/v1`, no release/SPEC_PIN
  file touched; frozen v1 goldens byte-identical by construction.
- Hosted gate run 35475528062 (rev4 validation log, all jobs success; rose-air + Candidate suite skipped):
  `gh run view --json headSha` = `7f7d5955bca6263b41929e847650e2c079343012`; `git rev-parse
  7f7d5955^{tree}` = `01a3b519…` — the gate ran the exact revision-4 candidate. Jobs: Test ubuntu 4.1 min,
  macOS 13.3 min, Windows 67.8 min; Race ubuntu/macos, Lint, Naming, Gate self-test ×3, Interop
  conformance gate all success.

## 2. Corpus pin (unchanged since rev3, re-checked)

`DRAFT_SOURCES_PIN` = `802caee548ddc8b19408746d26c7972d39b39cc2` = curator-spec `main` HEAD today. The
vendored corpus bytes are unchanged by rev4 (the delta touches no testdata). AC subset files unchanged:
`ac-semantic-cases.txt` 73 ids (0 `v2-*`; contains `capture-mutation`, `attestation-evidence-wrong-name`,
`-wrong-context`), `ac-schema-cases.txt` 102 (contains all 5 schema-bound instances).

## 3. Independent execution on this host (macOS 15.6, go1.26.0 darwin/amd64, load 8–22)

Driver `/tmp/1xya7x-rev4/run-full.sh` (`#!/bin/bash`, `set -o pipefail`) in the `cand` clone:

| step | result |
|---|---|
| `gofmt -l internal/crossconformance internal/testcli` | exit 0, 0 files |
| `go vet ./internal/crossconformance/ ./internal/testcli/` | exit 0 |
| `go build ./...` | exit 0 |
| `go test -p 1 -count=1 -timeout=1800s -v -run 'Draft' ./internal/crossconformance/` | **exit 0**, `ok … 661.841s` (H2 ran concurrently; load 22) |
| `go test -p 1 -count=1 -v -skip 'Draft' ./internal/crossconformance/` (v1 regressions) | **exit 0**, 21 tests pass, 0 fail, 2 rust host-capability subtest skips (as on every hosted lane) |

Measured from the `-v` log: **243 `=== RUN` / 243 `--- PASS` / 0 FAIL / 0 SKIP**;
`TestDraftSourcesSchemaCases/*` **115/115**, `TestDraftSourcesSnapshotVectors/*` **3/3**,
`TestDraftSourcesSemanticCases/*` **94/94**. Ratio lines: `schema cases: 110 driven, 5 bounds, 115 total`
and (new, R1) `semantic cases: 89 driven, 4 known-gap, 1 bound, 0 skipped, 94 total`. Log markers:
4 `KNOWN-GAP` lines (`draftsources_semantic_evidence_test.go:490` wrong-name, `:506` wrong-context,
`draftsources_semantic_v2_test.go:251` v2-alias-resolution, `:402` v2-user-insteadof-ignored) and 6
`BOUND` lines (5 schema at `draftsources_schema_test.go:83`, 1 semantic `capture-mutation` at
`draftsources_semantic_capture_test.go:38`).

## 4. R1–R4 checked precisely

**R1 (executed-count assertion + ratio line) — met, and the H2 mutation now fails.**
`draftsources_semantic_test.go:29-58`: `resetSemanticOutcomes()` before the loop; each subtest registers
a `t.Cleanup` that classifies the row exactly once (`t.Skipped()` → skipped, else the explicit
bound/known-gap marker, else driven); after the loop `tallySemanticOutcomes()` logs the ratio line and
`t.Fatalf`s unless `driven+knownGap+bound+skipped == len(cases) == wantSemanticCases (94)`.
`recordSemanticOutcome` panics on a conflicting double classification; the registry is mutex-guarded.
Mutant **H2** in the probe clone (`:30` `range cases` → `range cases[1:]`, only that line; `git diff
--stat` 1 file, +1/−1): `go test -p 1 -count=1 -v -run 'TestDraftSourcesSemanticCases$'` → **exit 1**,
93 `--- PASS` subtests, `semantic cases: 88 driven, 4 known-gap, 1 bound, 0 skipped, 94 total`,
`executed(93) != total(94, want 94); run the full matrix without -run subtest filters` — **KILLED**
(rev3: survived with exit 0). File byte-restored (`git checkout --`), probe tree back at `01a3b519…`.
Consequences noted, not findings: (a) a `-run` filter naming a subtest now also fails the parent with
the same message (by design; the subtest verdict stays readable — see M6 below); (b) a `t.Parallel()`
row would drop out of the tally and fail closed; (c) a subtest that FAILS is tallied as driven — the
ratio line is meaningful only on a green run; (d) as in any test suite, an empty row would count as
driven — the per-row production-entry audit of the rev3 verdict §5 remains the control for that.
Hosted confirmation: the ratio line appears in all three lane logs (§5) and sums to 94 on each.

**R2 (known-gap vs bound) — met.** `semanticKnownGap` added (`draftsources_semantic_test.go:95-99`,
`KNOWN-GAP` marker, class `known-gap`); the four rows call it: `draftsources_semantic_evidence_test.go:490`,
`:506`, `draftsources_semantic_v2_test.go:251`, `:402` (drivers renamed `*Bound` → `*Gap`). The
gap-signature assertions are byte-for-byte the same conditions as rev3 (only the message tail changed to
"convert the known gap to a drive"): `resolution.Result != registry.ResultAudited` (`:487`, `:503`),
`len(clones) != 1 || clones[0] != v2AliasEndpoint` (`:247`), `member.Package.Commit.Hex != evilCommit`
(`:399`). `semanticBound` has exactly one remaining caller: `capture-mutation`
(`draftsources_semantic_capture_test.go:38`). Flip probe **M6** (probe clone, production
`internal/registry/registry.go` `Matches` prefixed with `if record.Name != "review" { return false }`,
i.e. exact-name matching "implemented"): `-run 'TestDraftSourcesSemanticCases/attestation-evidence-wrong-(name|commit)$'`
→ exit 1, `--- FAIL: …/attestation-evidence-wrong-name` with `wrong-name Resolve = unknown, want the
locked audited outcome; convert the known gap to a drive`, while `…/wrong-commit` still `--- PASS`
(15.17 s) — the xfail lock is live after the rename. Restored; tree `01a3b519…`.

**R3 (results record) — met.** `TASK-260910-1xya7x_results.md` (rev4): semantic headline `94 pinned =
89 driven-pass + 4 known-gap + 1 bound`, AC subset `70 + 2 + 1`, schema `115 = 110 + 5` / AC `97 + 5`
— all equal to my measured run (§3) and to the hosted ratio lines (§5); a heading "Conformance failures
/ product gaps (driven, fail the draft)" names the three follow-ups (security literal-URL git config;
draft §4 exact evidence matching; transport rev2 alias substitution) and states the xfail treatment;
per-lane coverage quoted verbatim from the rev3 verdict §4 (re-confirmed on the rev4 gate ledgers, §5);
Windows package time quoted from rev3 (2447 s / 2275 s; rev4: 2409 s / 2238 s — same order); the two
inverted-oracle lock positives are noted next to the schema bounds; "Conformance at production entries,
not release qualification". **Arithmetic correction accepted**: the rev3 verdict's F1 tally "87 driven"
was my error (87+4+1 = 92); the driven set was 89 in rev3 already and the 4 rows moved bound → known-gap,
so 94 = 89 + 4 + 1 is the correct record.

**R4 (transaction e2e wording) — met.** `draftsources_transaction_e2e_test.go:3-14` now states the row
proves a transaction PREPARE refusal (mkdir of the `.curator-txn` desired dir: permission denied) leaves
the prior generation byte-identical, and points post-prepare rollback at `write-boundary-retarget`
(Go-API hook after PointPrepared) and `TestDraftFailureAtEveryTargetClassRestoresPriorState`; the
in-body comment (`:57-60`) and the failure message (`:78` "want the prepare refusal") match; results.md
carries the same wording. Optionals (RequireGit reason, `t.Parallel`, probe fold-in) not taken, as
allowed.

## 5. Hosted per-lane evidence, rev4 gate 35475528062 (`test/observed-cases.tsv`, `skips-observed.tsv`, `go-test.json`)

| lane | pkg time | schema | snapshot | semantic (ledger) | ratio line (from the lane log) | other draft skips |
|---|---|---|---|---|---|---|
| macOS | 172 s | 115/115 | 3/3 | 94 pass, 0 skip | `89 driven, 4 known-gap, 1 bound, 0 skipped, 94 total` | none |
| ubuntu | 101 s | 115/115 | 3/3 | 93 pass, 1 skip (`case-alias`, host-capability "case sensitivity") | `88 driven, 4 known-gap, 1 bound, 1 skipped, 94 total` | none |
| windows | **2409 s** (semantic 2238 s) | 115/115 | 0/3 skip (platform-control "Windows does not expose portable executable permission bits…") | 45 pass, 49 skip ("test transport wrapper is POSIX-only") | `42 driven, 2 known-gap, 1 bound, 49 skipped, 94 total` | `CLILocalSkillScriptDependencies/shims_execute_the_frozen_runtime` ("executes POSIX skill commands"), `CLIInstallRestoresPriorState` ("this process can write through a read-only directory…") |
| rose-air | — | unverified (lane skipped) | | | | |

Numbers identical to the rev3 gate (as the producer states; only the ratio line is new). Every
crossconformance skip is classified `allowed-host-capability` / `allowed-platform-control` in
`skips-observed.tsv`; `platform-case-gate.sh` Tier 2 makes an unrecognised skip reason fatal, so with
R1 a narrowing by `t.Skip` is neither silent locally (ratio line, `--- SKIP`) nor admissible on the
hosted lanes. On Windows the two attestation known-gap rows execute (known-gap 2) and the two v2 gap
rows are among the 49 POSIX-wrapper skips; 42 + 2 + 1 + 49 = 94 = 45 pass + 49 skip. Windows budget:
`internal/crossconformance` 2409 s in its own package (< 120 m); the job wall time is bounded by
`internal/install` (3682 s); job 67.8 min.

## 6. Verified and holding from the rev3 verdict §10 (re-checked only where touched)

Exact tree/gate identity (§1); pin and AC subset (§2); 115/3/94 executed with 0 skips on this host and
on hosted macOS (§3, §5); production entries per group unchanged (rev4 changes no driver logic — only
the four marker calls, names and messages); H1 (schema counter) untouched; v1 regressions green (§3);
no production/spec/frozen file touched (§1); Windows within budget (§5); all skip reasons declared
verbatim (§5); no release-qualification claim in the record.

## 7. Bounds and follow-ups (unchanged, for the record)

Real bounds: 5 schema (3 `local-snapshot-v1` no byte reader; marker `valid.json` refused / F2;
`invalid-external-missing-substituted` admitted / F1 pre-existing), 1 semantic (`capture-mutation`),
external-evidence status half by test-side comparison (F4), 2 inverted-oracle lock positives counted
as driven (F6), rose-air unverified, Windows: transport fetch loop and snapshot byte vectors not proven.
Product follow-ups (filed by the orchestrator under STORY-260919-37szes: BUG-260920-3ukdk4,
-2wyzde, -3v7x6j, -2eg8nv, TASK-260920-1sbj7o): the four known-gap rows lock their signatures and will
fail when fixed (M6 shows the flip), which is the intended hand-off to those items.

## Evidence bundle

`TASK-260910-1xya7x_review-rev4-evidence.tar.gz`: `logs/` (full.status, draft.out summary lines,
h2.status/h2.out summary, post.status, v1.out summary, m6.out), driver scripts (run-full.sh, run-h2.sh,
run-post.sh), per-lane gate extracts (observed-cases counts, ratio lines, skip classes, durations).
