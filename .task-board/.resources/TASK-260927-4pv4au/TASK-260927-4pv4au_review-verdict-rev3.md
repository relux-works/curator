# TASK-260927-4pv4au — review verdict rev3: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate: CR-TASK-260927-4pv4au-3, base 6bd98d49, tree c7e73fa2 — the worktree tree was recomputed through a temporary index and matches c7e73fa2 exactly. Shell: zsh with pipefail. All exit codes below are real.

## F1 (rev2) — fixed
- cmd/curator/main.go:205-231 `runEnforcedShim` no longer emits anything. The only emission is in `loadConfig` (main.go:326-328), and it goes to stderr.
- internal/install/scriptpolicy_test.go is unchanged versus base, so its original fixture is restored. internal/scriptworker/derive_test.go:1160-1186 keeps the original schema-1 fixture and adds a check that no warning bytes reach stdout or stderr. A new test, `TestProductionEnforcedShimDoesNotLeakPermissiveWarning` (derive_test.go:1210), covers the explicit-permissive and schema-1 cases through the production launcher binary.
- Tests: `go test ./internal/scriptworker -run 'TestProductionBinaryLaunchesWhenHostProvides|TestProductionEnforcedShimDoesNotLeak'` exit 0; `go test ./internal/install -run TestEnforcedInstallAndLaunchAtCLIEntry` exit 0.
- Mutant M1 (the emission re-added to runEnforcedShim on stderr): both tests fail, exit 1.

## Deferred rev2 checks
The corpus is curator-spec 23435129 conformance/v1, extracted with `git archive` into $TMPDIR, with CURATOR_CONFORMANCE_ROOT set.
- Baseline: `go test ./internal/config ./cmd/curator -run 'Posture|Security|Permissive|Hardened|Conformance'` exit 0. `go test ./internal/config ./internal/envprofile -run 'Posture|Security|Hardened|Permissive|TestManagerOwnedAbsenceReadsAreGuarded'` exit 0, so the stateread guard passes.
- Vectors: TestSecurityPostureVectors and TestSecurityPostureVectorsThroughCLI report 11 driven, 0 known-gap, 6 bound, 17 total. The 6 bound cases are:
  - the unreachable-registry pair, owned by TASK-260910-1sapuy;
  - the 4 cases whose vector has rollout_revision=B and no machine posture (revision-B-default-hardened-flip-install, refusal-mcp-allowlist-empty-with-declarations, hardened-contradiction-status-check-non-current, posture-rows-flipped-revisions), owned by TASK-260927-25hk87.
  The hardened MCP refusal and the `--check` contradiction are driven with an explicit hardened posture by dedicated CLI tests, e.g. security_posture_test.go:205-257, where `env status --check` exits 1 with non_current and `mcp_package_allowlist_empty_refused`.
- The effective-posture API is available to the install resolver: `Config.EffectiveSecurityPosture`/`SecurityPostureHardened` in internal/config/security_posture.go:72-85.

### Mutants (exit codes)
| Mutant | Result |
|---|---|
| M2: revision switch set to "B" | killed (cmd/curator exit 1) |
| M-locked: machine value beats the system lock (config.go:528) | killed (config and CLI vectors, exit 1) |
| M-lockedsrc: lock source ignored (security_posture.go:268) | killed (exit 1) |
| M-srcallow: source-allowlist refusal removed | killed (exit 1) |
| M-mcp: MCP allowlist refusal removed | killed (exit 1) |
| M-passenv: `passable_env_names` null refusal removed | killed (exit 1) |
| M-warn-drop: warning dropped | killed (exit 1) |
| M-warn-dup: warning duplicated | killed (exit 1) |

Note: without CURATOR_CONFORMANCE_ROOT, M-locked survives because the vector tests skip. The hosted gate sets the variable (ci.yml:207/328), so this is a local-run caveat, not a defect.

## Hygiene
`git status` lists only product, test and .github/ci paths: no CHANGELOG/LOGBOOK edits, no binaries. The 3 remaining security-posture gap rows are the E1 source_signers schema-case rows owned by STORY-260916-ioemse; this Story does not own them.
