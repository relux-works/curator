# Review verdict — TASK-260916-3oh0u8 CR rev2 — ACCEPTED

Reviewer: claude-opus-5-5 (low). Base d41da0fb, candidate tree d13d9519 (worktree verified identical: `git diff --quiet d13d9519` ok). 6 paths, all in scope; no CHANGELOG/LOGBOOK, no stray files.

## Scope check vs rev1 snapshot
Trunk at d41da0fb already carries the E4 core (revisions A/B behind `activeProviderRevision`, A shipped — TestActiveRevisionIsWarningRelease; `provider_directories` knob in internal/config/environments.go, lockable — TestEnvConfigProviderDirectoriesRoundTrip/LockedRefuses; vector test TestUmbrellaProviderResolutionVectors; env status provider posture). Rev2 closes the remaining gaps rather than dropping rev1 scope:
- umbrella.go resolveRevisionB: symlink in a trust root whose canonical target escapes the roots → `subcommand_provider_untrusted`, no fall-through to later roots (environments §11 trust roots; §12 posture).
- umbrella.go resolveRevisionA: outside-roots hint names the canonical target dir.
- main.go cmdStatus: `curator status` now reports provider rows (text + JSON `providers`/`provider_diagnostic`) and `--check` fails on non-current provider rows (§12 posture: refused/failed row non-current).
- PATH override seam (`providerPathOverride`) test-only, nil in production; production PATH via pathEnvironment().
- platform-cases.tsv: 3 rows.
Gap ledger: no conformance-gaps.tsv row is owned by STORY-260916-2otjbn/this task; the provider-directories rows belong to STORY-260922-1cenbr (0017/0018 model) — not in scope.

## Independent validation (zsh, pipefail, CURATOR_CONFORMANCE_ROOT=curator-spec/conformance/v1)
`go build ./... && go vet ./cmd/curator/ && gofmt -l cmd/curator` clean; `go test ./cmd/curator/ -run 'Umbrella|Provider|CuratorStatus|EnvStatus' -count=1` → ok (412 s), EXIT=0; 38 tests PASS incl. vectors, hostile PATH plant warned(A)/refused(B), symlink escape, status posture.

## Mutants (real exit codes)
- M1 drop B symlink-escape refusal → exit 1, TestUmbrellaTrustedRootSymlinkEscape FAIL (killed)
- M2 drop `status --check` provider failure → exit 1, TestCuratorStatusProviderPostureAndCheck FAIL ("missing provider rows made status --check = 0, want 1") (killed)
- M3 revert A hint to non-canonical dir → exit 1, TestUmbrellaTrustedRootSymlinkEscape FAIL (killed)
Tree restored to d13d9519 after mutants (verified).

## Residuals (non-blocking)
- Hosted gate green per orchestrator note; not re-run here.
- Symlink-escape test skips where the host cannot create symlinks (Windows without privilege) — bound.
