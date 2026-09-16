# TASK-260910-14hsti — gate-fix handoff (rev2)

Run: RUN-260916-f6e054 (successor after rev1 remote-gate failure).
Prior evidence: TASK-260910-14hsti_results.md (rev1) and
TASK-260910-14hsti_change-request_rev1-validation.log.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`.
- Uncommitted candidate; the handoff snapshot publishes the authoritative tree.
- Tree delta vs base: `M internal/adapters/adapters.go` + 8 new files
  (`internal/{snapshot,staging,adapters}/boundaries{,_test}.go`,
  `internal/privatedir/staging{,_test}.go`). No other files touched
  (`git status` shows exactly these 9 entries).
- Spec checkout: curator-spec main `871d11b`. Re-read
  protocol/skillfile-sources.md (§2 physical boundaries, §5 diagnostics)
  and the revision-1-only scope; implementation unchanged from rev1
  except the skip-reason fix below. Diagnostics keep spec §5 names
  (`source_output_overlap`, `source_selection_invalid`,
  `source_alias_unknown`, `source_member_missing`, `source_member_invalid`).

## The rev1 failure and this fix

Remote gate run 35074054048: `go test` exit=0 on every lane, but the
platform-case gate failed on ubuntu (Test and Race lanes):

- `FAIL skip with an unrecognised reason on linux: internal/snapshot ::
  TestValidateCaseAliasUsesFilesystemIdentity`
- reason: `filesystem distinguishes case; symlink alias above covers the
  SameFile path`

Fix: reworded both `t.Skip` reasons in
`internal/snapshot/boundaries_test.go` (lines 105, 113) to
`test filesystem is case-sensitive; symlink alias above covers the
SameFile path`, matching the existing host-capability vocabulary in
`.github/ci/skip-classes.tsv` (same phrasing as
`internal/transaction/validation_darwin_test.go` and
`internal/gitops/deadlock_test.go`).

Proof against the gate's own classifier (awk ERE `reason ~ regex` over
the real `.github/ci/skip-classes.tsv`, exit 0):

- new reason -> `host-capability` (policy `allow`)
- old reason -> `UNCLASSIFIED` (reproduces the remote failure)
- `symlinks unavailable: ...` (all other skips in the touched test
  files) -> `host-capability`

No test logic, production code, or ledger file was changed for this fix.

## Direct validation (bash, standalone commands, real exit codes)

All commands run by this producer on the final candidate; no prior-run
evidence accepted for the rows below.

| Command | Exit | Evidence |
|---|---|---:|
| `go test -count=1 ./internal/snapshot/ -run TestValidateCaseAliasUsesFilesystemIdentity -v` | 0 | PASS, no skip: host volume is case-insensitive, `.AGENTS` semantic exercised directly |
| `go test -count=1 ./internal/staging/ ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | All four touched packages green (0.42s / 1.78s / 2.46s / 1.35s) |
| `go vet ./internal/staging/ ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | Clean |
| `test -z "$(gofmt -l ...)"` on the four packages | 0 | `FMT_CLEAN` |
| `git diff --check` | 0 | No whitespace errors |
| `go build -o /tmp/curator-14hsti ./cmd/curator` (binary removed) | 0 | CLI compiles |
| `golangci-lint run ./internal/staging/... ./internal/snapshot/... ./internal/adapters/... ./internal/privatedir/...` | 0 | `0 issues.` |
| awk skip-reason classification vs real `skip-classes.tsv` | 0 | new reason `host-capability`, old reason `UNCLASSIFIED` |

The full landing suite was not run manually; the handoff owns the single
remote-gate execution. Other platforms are unverified. No installs,
daemon restarts, tags, releases, LOGBOOK.md edits, runtime-home changes,
live credential export, or ax calls.

## Measured negative evidence (this run)

Two narrowing mutants, both killed (**2/2**, 0 survivors), bytes restored
(sha256 of both files matches pre-mutation snapshots; `cmp` clean):

| Narrowing mutation | Killer command | Real exit/result |
|---|---|---|
| staging `checkAdmittedOverlap`: reverse `Within(current, input)` refusal disabled (`if reverse && false`), forward check kept | `go test -count=1 ./internal/staging/ -run TestPlanRecheckRefusesAdmittedOverwriteBothDirections` | 1, expected failure: `admitted inside destination err = <nil>` |
| adapters `ValidateDestinations`: reverse `Within(candidate, input)` refusal disabled, forward check kept | `go test -count=1 ./internal/adapters/ -run TestValidateDestinationsRefusesAdmittedOverwrite` | 1, expected failure: `reverse err = <nil>` |

Post-restore confirmation: the four-package suite rerun green, exit 0
(row 2 above ran after restoration).

## Bounds and review needs

- Same stated bounds as rev1: Unicode-normalization aliases of missing
  paths are a blind spot (documented on `staging.Within`); root-input
  "every required context input" is bounded to SKILL.md, the effective
  manifest file, and declared runtime/build roots; revision-2 policy
  shapes are hooks only.
- The linux-only skip path (case-sensitive lane) is verified by
  vocabulary classification, not by executing the skip: this host's
  volume is case-insensitive so the test passes through the direct
  `.AGENTS` assertion here. The remote gate on ubuntu is the
  authoritative check for that path.
- Independent review must verify the exact published CR tree and rerun
  the narrow suites plus the remote gate.
