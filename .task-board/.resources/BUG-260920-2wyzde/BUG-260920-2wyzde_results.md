# BUG-260920-2wyzde results — draft §4 exact evidence matching

Ready for review. Shell for all commands below: `bash` with `set -o pipefail`.

## Outcome

Draft-lane registry evidence is now admitted only on exact name + canonical
repository + commit + context hash matching. Both corpus cases flip from
known-gap to driven-pass; narrowing mutants killed; legacy v1 lane unchanged.

## Production change

- `internal/registry/registry.go`: new `MatchesExact` (all four fields must
  match) and `ResolveExact` (deny-wins combination behind the exact
  predicate). `Resolve` keeps §13.3 OR-matching; both entries share the
  unmodified combination loop via `resolveMatched`.
- `internal/install/install.go`: `resolveRegistries` takes `draft`; the
  `install.Project` call site passes `draftLock != nil`, so draft-lane
  network members (the only draft nodes carrying `Identity`) resolve via
  `ResolveExact` with `node.Name`. Gate stays at step 14, before any
  cache/compiler/publication work.
- `internal/install/global.go`: passes `draft=false` (frozen lane).

Diagnostic decision: wrong-name/wrong-context refusals reuse the shared
strict typed refusal `"<skill> is not audited by any trusted registry
(registry_policy is strict)"` — the same class as the eight sibling
evidence rows, and the corpus expects identical behavior across all
evidence conditions. Spec §5 defines no separate evidence class. The
message carries skill name + policy only: no registry URL, endpoint, key,
or record bytes (asserted in the install test). No JSON shapes added or
changed, so no new closed-shape validation was needed; source-audit gate
and attestation renewal paths untouched (no diff in `draftaudit.go` or
renewal code).

## Tests

Committed, in repo layout:

- `internal/registry/registry_test.go`: `TestMatchesExact` (positive +
  four single-field negatives), `TestResolveExact` (exact admits, four
  mismatch shapes unknown with no attestation, exact revocation denies,
  identity-less query unknown; plus frozen-legacy locks that `Resolve`
  still returns audited for wrong-name/wrong-context).
- `internal/install/draftevidence_test.go` (new):
  `TestDraftEvidenceExactMatch` at `install.Project` over a live signed
  httptest registry: exact admits + audited marker attestation; wrong-name
  and wrong-context refuse with the typed refusal and no endpoint/key
  leak (`http://`, `https://`, `ed25519:`, `127.0.0.1` absent).
- `internal/crossconformance/draftsources_semantic_evidence_test.go`:
  new `attestWrongName`/`attestWrongContext` stub conditions folded into
  the `driveAttestationEvidence` 4-layer table (ResolveExact layer 0 with
  legacy-Resolve frozen lock; fresh+repair install refusal with
  state-digest preservation; refresh non-laundering; compiled-CLI status
  nonzero read-only). Gap drivers deleted; matrix still 94 cases.

## Evidence (real exit codes)

- `go test ./internal/registry/ -count=1` → ok, exit 0 (4.8s).
- `go test ./internal/install/ -run
  "TestDraftEvidenceExactMatch|TestRegistry|TestLegacyInstallUntouchedWhenDraftOff|TestDraftInstall"
  -count=1` → ok, exit 0 (31.7s).
- `go test ./internal/crossconformance/ -run "TestDraftSourcesSemanticCases"
  -count=1 -v` → exit 0 (486s):
  - `--- PASS: .../attestation-evidence-wrong-name (19.19s)`
  - `--- PASS: .../attestation-evidence-wrong-context (9.20s)`
  - `semantic cases: 92 driven, 1 known-gap, 1 bound, 0 skipped, 94 total`
  - remaining known-gap is the unrelated `v2-alias-resolution` row;
    remaining bound is the explicit `capture-mutation` bound.
- `go test ./internal/crossconformance/ -run TestDraftSourcesSemanticCoverage`
  → ok, exit 0.
- `go build ./...` → exit 0. `go vet` on registry+install → clean.
  `golangci-lint run` on registry, install, crossconformance → 0 issues.
- `gofmt -l` on the three packages → empty.

## Mutants (all killed, exit 1 with install wrongly `Status:ok`)

- M1 drop name compare in `MatchesExact` → `TestMatchesExact/wrong_name`,
  `TestResolveExact/wrong-name`, and
  `TestDraftEvidenceExactMatch/wrong-name-refuses` fail.
- M2 drop context compare → `TestMatchesExact/wrong_context`,
  `TestResolveExact/wrong-context`, and
  `TestDraftEvidenceExactMatch/wrong-context-refuses` fail.
- M0 draft lane rewired to legacy `Resolve` (original bug shape) → both
  install refusal subtests fail with `Status:ok` + attestation recorded,
  proving the new test detects the reported defect; `exact-admits` stays
  green. All mutants reverted; final diff verified restored.

## Legacy v1 unchanged

`Matches`/`Resolve` behavior preserved (shared loop, predicate-injected);
frozen-lane locks assert `Resolve` still admits both shapes as audited;
full registry package + legacy install tests green; draft flag false on
the v1/global lanes.

## Bounds

- Windows: not run locally; layer 3 keeps its declared POSIX-only skip,
  new install test uses only git + httptest + portable paths — hosted gate
  decides.
- No board/LOGBOOK edits; sibling checkpoint `BUG-260920-3ukdk4` untouched.
  Work left uncommitted in the story worktree for handoff snapshot.
