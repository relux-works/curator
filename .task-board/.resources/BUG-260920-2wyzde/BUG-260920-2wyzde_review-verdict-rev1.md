# BUG-260920-2wyzde — review verdict, revision 1 — ACCEPT

Reviewer run RUN-260920-f1d9f0 (claude-opus-5, independent). Leaf of STORY-260919-37szes.
Reviewed against: 2wyzde-brief.md, 2wyzde-review-rev1-note.md (binding), 37szes-review-brief.md,
curator-spec `protocol/skillfile-sources.md` draft §4 (curator-spec main 802caee) and
`protocol/registry.md` §3/§4 (frozen matching + deny-wins).

## 1. Candidate identity (exact tree, three ways)

- CR-BUG-260920-2wyzde-1 rev1: base `99cb352bad68f24967acf78cea82f0e685b2d0ea`, candidate tree
  `c627a28a3eb8c80bf999909bff08ccc014b7c081`, 6 changed paths.
- Story worktree `.temp/STORY-260919-37szes/worktree`: temp-index recipe
  (`GIT_INDEX_FILE=$(mktemp) git read-tree HEAD && git add -A . && git write-tree`) →
  `c627a28a…` (covers the untracked `internal/install/draftevidence_test.go`). Re-checked at the
  end of the review: still `c627a28a…`, worktree left untouched (an early `git add
  --intent-to-add` probe was undone with `git reset -q`; the index is back at HEAD).
- Patch resource `BUG-260920-2wyzde_change-request_rev1.patch` sha256
  `53859770f1c6804417a6644e8f1c02a02492dc1b56d16e9631a9d75a385d0f85` (matches the CR). Disposable
  clone `/tmp/2wyzde-mutants/base`: `git checkout --detach 99cb352b && git apply --index rev1.patch
  && git commit` → `HEAD^{tree}` = `c627a28a…`. Every mutant and every throwaway probe ran there;
  the clone's tracked tree was `c627a28a…` before and after both drivers.
- Hosted gate (validation log): `scripts/remote-gate.sh` pushed `910e319adad918c71103e57dd0e49bf6a2a9ef48`
  as `gate/STORY-260919-37szes/260920-061004-50352-1`; `git rev-parse 910e319a^{tree}` =
  `c627a28a…`, parent `99cb352b`. `gh run view 35493522363` → headSha `910e319a…`, conclusion
  `success`: Test/Race ubuntu+macos, Test windows, Lint, Naming, Gate self-test ×3, Interop
  conformance gate all success; `Test (rose-air)` skipped (declared: no self-hosted lane configured
  for this Story), `Candidate suite` skipped (matrix expression, pre-existing).

## 2. Contract check (review-note items 1–5)

Production delta: `internal/registry/registry.go` adds `MatchesExact` (name ∧ canonical
source_identity ∧ commit ∧ content_sha256) and `ResolveExact`; `Resolve` is refactored into the
shared `resolveMatched` loop with the frozen predicate `Matches` injected unchanged (deny-wins,
signature check after match, warnings identical). `internal/install/install.go`
`resolveRegistries(..., draft bool)`: the `install.Project` call site passes `draftLock != nil`
(non-nil only when `effectiveManifest.SchemaVersion == 2 && draftSourcesEnabled`), and the loop
resolves through `ResolveExact(usable, node.Name, node.Identity, node.Resolved.Commit, contentHash, fetch)`
on the draft lane, `Resolve` otherwise. `global.go` passes `false` (the global scope has no draft lane).

1. Exact-match layer on the draft lane, network-Git members only: on the draft lane
   `closure.LoadDraftFrozenNodes` sets `node.Identity = pkg.Repository` only for
   `KindNetworkGit` packages (local-snapshot and legacy configured-git members keep `Identity == ""`
   and are skipped by the pre-existing `if node.Identity == "" { continue }`), so `ResolveExact`
   reaches exactly the members draft §4 names. Both sides of the repository compare are canonical
   by construction (`manifest/sources.go` → `buildrepo.ParseSource().Identity`;
   `sourcelock.Package.Validate` → `identity.ValidCanonical`; `registry.ParseRecord` rejects a
   non-canonical `source_identity`), so exact string equality is the right compare and there is no
   canonicalization step left to mutate. Wrong-name / wrong-context / wrong-repository /
   wrong-commit records resolve `unknown` → strict policy refuses at step 14, before staging and
   publication, with the shared typed refusal `"<skill> is not audited by any trusted registry
   (registry_policy is strict)"` (skill name + policy only; asserted free of `http://`, `https://`,
   `ed25519:`, `127.0.0.1`). Verified at `install.Project` for all four single-field shapes
   (§3, §4 below). Accepting the producer's diagnostic decision: the corpus expects one behavior
   class across all evidence conditions and draft §5 defines no separate class.
2. Corpus rows: `attestation-evidence-wrong-name` and `attestation-evidence-wrong-context` are now
   registered through `draftEvidenceConditions` into the 4-layer `driveAttestationEvidence`
   (Layer 0 `ResolveExact` over a static fetch with a frozen-legacy lock; Layer 1 fresh + repair
   `install.Project` over the live signed stub with state-digest preservation; Layer 2 refresh
   non-laundering; Layer 3 compiled-CLI status). The two known-gap drivers are deleted;
   `TestDraftSourcesSemanticCoverage` still binds 94 ids. Ratio line from my own full run
   (worktree, `-run 'TestDraftSourcesSemanticCases$|TestDraftSourcesSemanticCoverage$'`):
   `semantic cases: 92 driven, 1 known-gap, 1 bound, 0 skipped, 94 total`, `ok … 447.2s`, exit 0
   (before this leaf the two rows were known-gap). Remaining known-gap is the unrelated
   `v2-alias-resolution` row; remaining bound is `capture-mutation`.
3. Mutants — see §4: name / context / call-site / lane-swap / predicate-swap all killed at the
   production entry; repository-only and commit-only compares killed at the helper level by
   committed tests and at `install.Project` by my probe (bound B1).
4. Legacy v1 unchanged: `Matches` byte-identical; `Resolve` = `resolveMatched` with the same
   predicate; legacy call sites pass `draft=false`. Independently rerun: registry package full
   (`ok 3.9s`), `TestRegistry*|TestStrictRegistryPolicyFailsUnknown|TestLegacyInstallUntouchedWhenDraftOff`
   (`ok 13.1s`), interop goldens against `curator-spec/conformance/v1`
   (`CURATOR_CONFORMANCE_ROOT=…` → 10/10 PASS incl. `TestGoldenFederationSemantics`, exit 0).
   Production-entry proof: my throwaway probe `TestReviewProbeLegacyLaneORMatchingAtInstallProject`
   (legacy `install.Project`, strict, wrong-name record with matching content / wrong-context record
   with matching identity+commit) → `ok` + marker attestation `audited` on the candidate, and fails
   under the lane-swap mutant M6 (legacy side got exact matching) — so the frozen OR-matching is
   still what the legacy production entry does. `draftaudit.go` / `internal/audit` untouched
   (empty diff): source-audit gate (hwxr26) and attestation/marker renewal paths unchanged.
5. No new or changed JSON shape (nothing to validate closed); no registry endpoint, key or record
   bytes in the refusal (asserted at `install.Project`).

## 3. Independent reruns (zsh, `set -o pipefail`; bash drivers for the clone)

| command (worktree unless noted) | result |
|---|---|
| `gofmt -l internal/registry internal/install internal/crossconformance` | empty, exit 0 |
| `go build ./...` | exit 0 |
| `go vet ./internal/registry/ ./internal/install/ ./internal/crossconformance/` | exit 0 |
| `golangci-lint run ./internal/registry/... ./internal/install/... ./internal/crossconformance/...` | `0 issues.` exit 0 |
| `go test -count=1 -p 1 ./internal/registry/` | `ok 3.863s` exit 0 (TestMatchesExact 4/4, TestResolveExact 4/4 + revocation + identity-less) |
| `go test -count=1 -p 1 ./internal/install/ -run TestDraftEvidenceExactMatch -v` | PASS exact-admits / wrong-name-refuses / wrong-context-refuses, `ok 5.412s` exit 0 |
| `go test -count=1 -p 1 ./internal/crossconformance/ -run 'TestDraftSourcesSemanticCases/attestation-evidence-wrong-(name\|context)$' -v` | both rows `--- PASS` (11.5 s / 9.1 s); parent FAIL only from the by-design `executed(2) != total(94)` guard under `-run`; exit 1 as designed |
| `go test -count=1 -p 1 -timeout 40m ./internal/crossconformance/ -run 'TestDraftSourcesSemanticCases$\|TestDraftSourcesSemanticCoverage$' -v` | `92 driven, 1 known-gap, 1 bound, 0 skipped, 94 total`; `ok 447.246s` exit 0 |
| `go test -count=1 -p 1 ./internal/install/ -run 'TestRegistry\|TestStrictRegistryPolicyFailsUnknown\|TestLegacyInstallUntouchedWhenDraftOff' -v` | 7/7 PASS, `ok 13.116s` exit 0 |
| `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1 go test -count=1 -p 1 ./internal/interop/ -v` | 10/10 PASS, exit 0 |
| clone: `go test ./internal/install/ -run 'TestDraftEvidenceExactMatch\|TestReviewProbeLegacyLaneORMatchingAtInstallProject'` | 5/5 PASS, `ok 21.997s` exit 0 |
| clone: `go test ./internal/install/ -run TestReviewProbeDraftLaneSingleFieldMismatchAtInstallProject` | wrong-repository-only / wrong-commit-only both refused, exit 0 |

Hosted per-platform evidence (`gh run download 35493522363 -n test-evidence-<os>`, `test/go-test.json`):
- ubuntu: `TestDraftEvidenceExactMatch` (3/3) pass; rows wrong-name / wrong-context / wrong-repository /
  wrong-commit / wrong-key pass; ratio `91 driven, 1 known-gap, 1 bound, 1 skipped, 94 total`
  (skipped = `case-alias`, unrelated).
- macos: `TestDraftEvidenceExactMatch` (3/3), `TestMatchesExact`, `TestResolveExact` pass; both rows
  pass; ratio `92 driven, 1 known-gap, 1 bound, 0 skipped, 94 total`.
- windows: `TestDraftEvidenceExactMatch` (3/3), `TestMatchesExact`, `TestResolveExact` pass;
  `TestDraftSourcesSemanticCases` pass (2295 s); the evidence rows report `--- SKIP` with the
  declared reason `test transport wrapper is POSIX-only` (Layer 3, line 420) after 15–23 s of
  Layers 0–2; ratio `42 driven, 0 known-gap, 1 bound, 51 skipped, 94 total` — declared skips only.

## 4. Mutants (disposable clone; `git checkout --` between mutants; per-mutant `go test` recompiles)

| id | mutation | registry pkg | `install.Project` (TestDraftEvidenceExactMatch) | crossconformance rows | verdict |
|---|---|---|---|---|---|
| M1 | `MatchesExact`: drop `record.Name == name` | FAIL wrong_name (helper) + ResolveExact/wrong-name | FAIL wrong-name-refuses (`Status:ok`, attestation recorded) | wrong-name FAIL (Layer 0 `Resolve = audited, want unknown`) | killed |
| M2 | drop `record.ContentSHA256 == contentSHA256` | FAIL wrong_context + ResolveExact/wrong-context | FAIL wrong-context-refuses (`Status:ok`) | wrong-context FAIL (Layer 0) | killed |
| M3 | drop `record.SourceIdentity == sourceIdentity` | FAIL TestMatchesExact/wrong_repository only | committed rows PASS (survive); reviewer probe wrong-repository-only FAIL (`Status:ok`) | PASS (rows mutate commit+content too) | killed by committed helper test; production-entry kill by reviewer probe only → B1 |
| M4 | drop `record.Commit == commit` | FAIL TestMatchesExact/wrong_commit only | committed rows PASS; reviewer probe wrong-commit-only FAIL | PASS | same as M3 → B1 |
| M5 | call site `resolveRegistries(..., draftLock != nil)` → `false` | PASS (untouched) | FAIL both refusals (`Status:ok` + attestation) | both rows FAIL at Layer 1 `fresh install = {… Status:ok …}, want refusal` (line 365 = `install.Project`) | killed at the production entry |
| M6 | `if draft` → `if !draft` (lane swap) | PASS | FAIL both draft refusals; legacy probe FAIL (`skill-a is not audited …`) | both rows FAIL at Layer 1 | killed; legacy side detected by reviewer probe |
| M7 | `ResolveExact` admits via `Matches` (exact helper uncalled) | FAIL ResolveExact/wrong-name, wrong-context | FAIL both refusals | both rows FAIL (Layer 0) | killed |

M1/M2/M7 stop the crossconformance rows at Layer 0 (helper layer), so their Layer 1 sensitivity is
shown by M5/M6 (Layer 0 intact, `install.Project` rewired) — both rows fail exactly there.

## 5. Bounds, residuals, nits (none blocking)

- B1 (committed evidence, not product): the wrong-repository-only and wrong-commit-only shapes are
  killed only by `TestMatchesExact` (helper-direct). The product refuses both at `install.Project`
  (reviewer probe, §3), and the wiring is proven by M5/M7, but no committed row drives those two
  shapes at the production entry (the crossconformance wrong-repository / wrong-commit stubs also
  mutate the content hash, so they exercise the frozen predicate, not the identity/commit
  compares). Recommend two extra rows in `TestDraftEvidenceExactMatch`'s table
  (`source_identity` only; `commit` only) in a follow-up of this Story; AC lists name/context only.
- R1 (spec reading, orchestrator/spec-owner call, no change requested): `ResolveExact` applies the
  exact key to every record, including `revoked` ones. Under strict policy a non-exact revocation
  still refuses (unknown → strict). Under advisory policy the draft lane no longer denies a
  revocation that matches only by content (same bytes revoked at another canonical repository) or
  by identity+commit with another name/hash — the frozen lane still does (registry.md §3
  "content equality intentionally permits one audited tree mirrored from another source" and
  §4 deny-wins). Draft §4 literally says network-Git members "may use existing registry evidence
  only with exact … matching", which this implements; whether deny-wins revocation should keep the
  broad §3 match on the draft lane is a spec question, not a defect against the accepted contract.
- R2 (pre-existing, out of scope): the "context hash" compared is the registry artifact hash the
  frozen client already computes — `hashing.ContentSHA256(node.Snapshot, nil)` over the raw frozen
  package tree — not the lock member's projected context hash (`closure.ContentHashFor`, which
  excludes `agent-skill.json` and non-whitelisted roots, so the two always differ). Draft §4 says
  "existing registry rules establish … context hash" and "no registry redesign", so keeping the
  existing artifact hash is the conformant reading; both stubs echo the query hash and cannot
  distinguish the two. Worth a terminology note to the spec owner.
- N1: Layer 0 of the seven sibling evidence conditions still calls `registry.Resolve`; only the two
  new conditions call `ResolveExact`. Layer 1 (`install.Project`) is the production entry for all
  nine rows, so this is cosmetic.
- N2: `TestDraftEvidenceExactMatch/exact-admits` duplicates the `run` helper body instead of
  calling `run(t, nil)`.
- Windows: crossconformance rows skip at Layer 3 by declaration; the production-entry install test
  ran and passed on windows-latest. rose-air: not configured for this Story (skipped, unverified).
- Host: local runs were bounded (every tool call < 2 min; long suites via background logs); no
  host incidents during this review. Leaked `curator-conformance-bin*` dirs from my own run window
  (11:26–11:34 local, 6 dirs) deleted; nothing else touched outside `/tmp/2wyzde-*`.

## 6. Verdict

ACCEPT revision 1: AC met (draft-lane exact admission at the production entry, both corpus rows
driven-pass with the ratio counting them, name/context narrowing mutants killed at
`install.Project`, legacy v1 lane byte-identical with the switch off, sanitized typed refusal,
gate green on the exact candidate tree). Recorded via
`accept_cr(BUG-260920-2wyzde, revision=1, evidence=BUG-260920-2wyzde_review-verdict-rev1.md)`.
No commit_ack; the developer/implementer producer run integrates.
