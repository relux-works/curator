# TASK-260924-4mzun5 rework 3 results (developer, revision 3)

Revision 2 verdict (`TASK-260924-4mzun5_review-verdict-rev2.md`, changes_requested)
left two blocking findings about declared-tag currentness for legacy package
markers. Both are fixed at the production entries, with permanent
project/global regression tests. No other behavior was changed.

## F4 — explicit tag change falsely refused under strict-tag policy (regression)

Cause: `internal/install/install.go` `detectMovedTagsIn` warned on any commit
difference between the recorded package commit and the live declared-tag
binding. A v1-to-v2 declaration change (both tags unmoved) therefore read as
"moved tag for consumer: v2 old -> new" on project and global entries.

Fix: on a commit difference for a legacy-lane package marker, the gate now
proves same-tag movement from the live tag bindings (core section 10 limits
strict-tag refusal to a moved tag): when the recorded commit is still the
value of a tag, the previous tag-bound installation is intact and the new
binding is an explicit declaration change, which stays silent; only a
recorded commit that lost every tag binding warns. Failed tag enumeration
fails the gate (unreadable evidence is neither a move nor an intact
binding), consistent with generation-read failures. Non-package legacy
markers keep their exact ref comparison; draft-lane package markers are
still skipped (moves observable only at explicit refresh).

- New: `internal/gitops/gitops.go` `TagsPointingAt` — one
  `git for-each-ref` invocation enumerating every tag with direct and
  peeled object names; peeling comparison exact on every git version.
  Parsing verified empirically (lightweight `v1  <sha>`, annotated
  `v2-ann <peeled> <tagobj>`).
- Changed: `internal/install/install.go` `detectMovedTagsIn` (+ `gitops`
  import).

Tests: new permanent `TestReviewChangedDeclaredTagIsNotMovedTag`
(`internal/install/legacy_movedtags_test.go`, project + global subtests,
adapted from the reviewer's probe) — strict reinstall after an explicit
v1-to-v2 change succeeds silently. Control
`TestLegacyLaneMovedTagRefusesStrictSecondInstall` (actual same-tag move
still refuses) is unchanged. Narrowing mutant `movedtag-commit-only`
(drop the tag-binding proof, warn on any commit difference) fails the new
test on both entries by construction.

Stated bound: a move whose old commit remains tagged elsewhere reads as a
declaration change. The closed v5/v6 shape persists no prior ref
(skillfile-sources section 4), so full same-tag proof for that case needs
prior-declaration evidence this gate does not have. Documented in the gate
comment; suggested follow-up: persist the legacy effective lock.

## F6 — status admits a stale declared-ref binding (bypass)

Cause: `cmd/curator/main.go` `scopeStatusDrift` reported legacy package
markers up-to-date on commit equality alone. Changing v1 to an unchanged
v2 at the same commit returned up-to-date and `status --check` exit 0.

Fix: status now compares the recorded package and lock against the staged
effective plan exactly like the install currentness check
(`marker.Current`: package DeepEqual + lock equality). The staged lock
binds the declaration (skillfile-sources section 4: "Status MUST compare
package, lock, attestation and substitution against the effective plan ...
Changed ... declared ref, package or lock makes the installation
non-current"), so a same-commit ref change is non-current. A package with
no staged expectation can never prove current and needs an install
(fail-closed; no commit-only fallback anywhere).

- `internal/install/install.go` + `internal/install/global.go`: status
  plans (`OperationStatus`) now stage the effective lock even though they
  are dry runs (pure computation, writes nothing); `Result` carries
  `LegacyPackages` + `LegacyLockSHA256`.
- `cmd/curator/main.go`: `statusReport` / `scopeStatusDrift` take the
  staged expectation (project and global call sites); package branch
  compares package + lock. `statusDrift` wrapper and 6 existing test call
  sites pass nil/"" (all cover non-package markers; behavior unchanged).

Tests: new permanent `TestReviewCLIStatusDetectsChangedTagAtSameCommit`
(`cmd/curator/audit_directory_test.go`, project, adapted from the
reviewer's probe) and new
`TestReviewGlobalStatusDetectsChangedTagAtSameCommit`
(`cmd/curator/global_status_test.go`, machine-wide entry written for this
revision). Control `TestCLIStatusReportsLegacyPackageMarkerCurrent`
(unchanged install still up-to-date) is unchanged. Narrowing mutant
`status-commit-only` (compare recorded commit vs live resolution instead
of staged package + lock) fails both new tests by construction.

## Files changed (this revision only)

- `internal/gitops/gitops.go` — `TagsPointingAt`
- `internal/install/install.go` — gate proof, `Result` fields, status lock
- `internal/install/global.go` — status lock, `Result` fields
- `internal/install/legacy_movedtags_test.go` — F4 regression
- `cmd/curator/main.go` — staged plan threading, package+lock comparison
- `cmd/curator/audit_directory_test.go` — F6 project regression
- `cmd/curator/global_status_test.go` — F6 global regression
- `cmd/curator/status_test.go`, `cmd/curator/builds_test.go`,
  `cmd/curator/nul_opaque_v1_rework_test.go` — signature updates only
- `.../testdata/draft-sources-v1/` untouched (trunk content preserved)

## Compile-only checks (R223: no local `go test`; hosted gate is arbiter)

- `go build ./...` — exit 0, no package errors.
- `go vet ./...` — exit 0 (vet type-checks the new tests too).
- `gofmt -l` on touched files — clean.
- `git for-each-ref` peeling format verified empirically in a scratch repo.

Not run locally: `go test` (forbidden on the mini by R223/R220).

Hosted full gate on the exact tree (sanctioned `scripts/remote-gate.sh`,
throwaway branch auto-deleted): run 37858487935
(https://github.com/relux-works/curator/actions/runs/37858487935),
snapshot 8ddcf7f3e7525c50d824f22fab5bc97f6a16f70b — conclusion
**success**. All lanes green: Test ubuntu/macos/windows, Race
ubuntu/macos, Lint, Go driver matrix (1.25.5/1.26.0/1.27.1 x
ubuntu/macos/windows), gate self-tests, interop, naming;
candidate-suite and rose-air skipped (as in prior CR gates).
Ubuntu `go-test.json` evidence: the three new tests ran and passed
(`TestReviewChangedDeclaredTagIsNotMovedTag` project+global,
`TestReviewCLIStatusDetectsChangedTagAtSameCommit`,
`TestReviewGlobalStatusDetectsChangedTagAtSameCommit`; 5 pass actions,
0 skips), zero `"Action":"fail"` in the whole stream.

Mutant kills (`movedtag-commit-only`, `status-commit-only`) hold by
construction and are for the reviewer to run on hosted runners per the
review note.
