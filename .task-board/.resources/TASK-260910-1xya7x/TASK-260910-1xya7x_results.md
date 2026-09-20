# TASK-260910-1xya7x results — draft source conformance + end-to-end (rev4)

Rev4 is test/report-only (rework-3 R1–R4; no production/spec change).
Revs 1–3 evidence stands except where this text corrects the record.

## Tested revision (exact)

- Worktree HEAD `71e353e63489ac6fdb9db67bdb6049033ab0a1bd` (Story
  branch `task-board/story/STORY-260910-1cnwwp`; checkpointed sibling
  stbg4d rev2 on trunk waves 1-3). Candidate: 18 new
  `internal/crossconformance/draftsources_*_test.go` +
  `testdata/draft-sources-v1/` (vendored corpus) +
  `internal/testcli/cli.go`, all UNCOMMITTED. No production/spec file
  touched; frozen v1 and release qualification unchanged.
- Draft corpus pin: curator-spec
  `802caee548ddc8b19408746d26c7972d39b39cc2` (`git archive` of
  `conformance/draft-sources-v1` + `schemas/draft-sources-v1` diffs
  clean vs the vendored tree). AC corpus a4fcaf0 (102 schema / 73
  semantic) is an unchanged subset (delta purely additive: 13
  source-policy-v2 schema cases + schema, 21 `v2-*` semantic cases;
  snapshot-cases byte-identical).
- Toolchain (macOS lane): `go1.26.0 darwin/amd64`, Apple Git-155.

## Case counts (macOS, exit 0)

Schema 115 pinned / 110 driven / 5 bounds; snapshot 3/3/0; semantic
94 pinned = 89 driven-pass + 4 known-gap (driven rows that fail the
draft expectation) + 1 bound (rev4 run: 211 PASS / 0 SKIP / 0 FAIL
/ 4 KNOWN-GAP / 6 BOUND, 301.981s; `semantic cases: 89 driven, 4
known-gap, 1 bound, 0 skipped, 94 total`). AC schema subset: 97 driven
+ 5 bounds; AC semantic subset: 70 + 2 + 1. NOTE on verdict-rev3 §9/F1
"87 driven": 87+4+1 sums to 92, not 94; the measured ratio line reads
89 (70 AC + 19 v2) — the 4 gaps moved bound→known-gap, the driven set
is unchanged from rev3's 89. Drivers: skillfile-v2 ->
`manifest.LoadWithOptions` (DraftSourcesV1); source-policy v1/v2 ->
`config.ParseSourcePolicy`; lock -> `sourcelock.Parse`; marker-v5 ->
`marker.Read`; receipt-v3 -> canonical marshal +
`buildmeta.DecodeReceipt`; audit -> `audit.ParseSourceAudit`.

## Conformance failures / product gaps (driven, fail the draft)

Three follow-ups, OUTSIDE this leaf's scope (surfaced, not caused):
1. Security: the draft literal-URL resolve lane honours user git
   config (hostile `url.<evil>.insteadOf` redirects the clone; lock
   binds the evil commit; resolved lane already isolates).
   Corpus `v2-user-insteadof-ignored`.
2. Draft §4 exact evidence matching (name + canonical repository +
   commit + context) not implemented: `registry.Matches` OR-matching
   accepts wrong-name/wrong-context records. Corpus
   `attestation-evidence-wrong-name`, `-wrong-context`.
3. Transport rev2 alias substitution never reaches the connection
   (executor fetches `attempt.URL`; resolved host is provenance
   only). Corpus `v2-alias-resolution`.
Each row locks its gap signature and fails on a fix (xfail); counted
as known-gap, never as passing.

## Explicit bounds (never passing)

Schema 5: 3x local-snapshot (no byte-reader entry; shape via Capture
vectors); marker valid.json (inconsistent fixture; refusal correct);
marker invalid-external-missing-substituted (reader admits; row locks
it). Note: the two inverted-oracle lock positives
(`skillfile-lock-v1/valid-git.json`, `valid-configured-git.json`:
schema-valid, refused with `source_member_invalid` on member/package
directory disagreement) count as driven; consistent treatment would
list them with the marker `valid.json` bound. Semantic 1:
capture-mutation (in-package proof only). Design: no JSON Schema
engine (production readers driven); no positive authenticated network
fetch; external-CLI status half fail-closed.

## AC clause coverage (production entries)

Pin/counts via CorpusCounts + Pin (DRAFT_SOURCES_PIN + MANIFEST,
no extras). Local skill+script+compiled CLI+deps e2e; broker askpass
(2 prompts, 9 silent refusals + control); 13 transport + 21 v2 rows
via CLI resolve with default-deny fake git; external mirror admitted
/ port+alias refused via install.Project; transaction PREPARE refusal
row (wording corrected rev4; rollback proof is
write-boundary-retarget + the install-package sweep); spaced/unicode
paths, symlink-escape refusal, exec-bit vectors, windows/linux
cross-compile; v1 regressions (21 tests) green on rev3 (reviewer §8;
rev4 touches no covered behavior).

## Platform coverage (actual; no qualification claim)

Rev3 gate run 35469802052 ledgers, per verdict-rev3 §4 (rev4 is
test/report-only; lane behavior unchanged):
- macOS (hosted): pkg 214 s; schema 115/115; snapshot 3/3; semantic
  94/94 pass, 0 skip; no other draft skips.
- ubuntu (hosted): pkg 111 s; schema 115/115; snapshot 3/3; semantic
  93 pass + 1 skip (`case-alias`, host-capability "case
  sensitivity"); no other draft skips.
- Windows (hosted): pkg 2447 s (SemanticCases 2275 s); schema
  115/115; snapshot 0/3 (platform-control: no portable executable
  permission bits); semantic 45 pass + 49 platform-control skips
  ("test transport wrapper is POSIX-only"); 2 e2e skips (frozen-shim
  row, "executes POSIX skill commands"; restore-prior-state row,
  read-only-directory writability).
- rose-air: unverified (lane skipped).
All skip reasons match `.github/ci/skip-classes.tsv` verbatim; no new
class. Windows proves the 115 schema readers, install-level rows and
18 external-evidence rows — not the transport fetch loop or snapshot
vectors. Conformance at production entries, not release qualification.

## Revision 4 — rework-3 fixes (R1–R4, no production change)

R1 `draftsources_semantic_test.go`: the runner classifies every
iterated case (driven/known-gap/bound/skipped; skip via `t.Cleanup` +
`t.Skipped`), logs the ratio line, and asserts `executed ==
len(cases) == wantSemanticCases` (the `skipped` slot extends the
requested shape so the Windows matrix tallies; macOS runs 0 skipped).
R2 New `semanticKnownGap` (`KNOWN-GAP` line) at the 4 driven-failing
rows; gap-signature assertions unchanged; drivers renamed `*Bound` ->
`*Gap`. Only capture-mutation stays a bound. R3 This file. R4
Transaction e2e header/comments now state the row proves a PREPARE
refusal (mkdir of `.curator-txn` desired: permission denied) leaves
the prior generation intact; rollback proof pointed at
write-boundary-retarget + the install-package sweep. Optionals
(RequireGit reason, `t.Parallel`, §7 probe fold-in) not taken:
explicitly non-gating.

## Evidence (macOS, real exit codes)

- `gofmt -l internal/crossconformance internal/testcli`: clean, exit 0.
- `go vet ./internal/crossconformance`: exit 0.
- `go test -p 1 ./internal/crossconformance -run
  'TestDraftSourcesSemanticCases|TestDraftSourcesSchemaCases' -count=1
  -timeout=900s -v`: exit 0, 301.981s; 211 `--- PASS` / 0 FAIL;
  4 KNOWN-GAP + 6 BOUND lines; ratio lines `schema cases: 110 driven,
  5 bounds, 115 total`, `semantic cases: 89 driven, 4 known-gap,
  1 bound, 0 skipped, 94 total`.
- H2 mutant (runner loop `cases` -> `cases[1:]`, line 28 only):
  exit 1, 299.836s, 93 subtests ran, `executed(93) != total(94, want 94)` — KILLED;
  byte-restored (`cmp` clean) + `go vet` exit 0 on restored bytes.
- Windows: verified by the hosted gate only.
