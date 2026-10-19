

## Publication after converge

Current r5 publication handoff is based on `892dab098aef1497df9ccacdb173faaf4f8b8418`; uncommitted candidate tree `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. The orchestrator's converge resolves the previous board/checkpoint refresh blocker. `task-board worktree status STORY-261018-ogw2f1` exited 0: current lease and expected dirty candidate, no refresh-specific refusal. No `refresh-candidate` was run.

The implementation, both permission test files and `docs/environment-config.md` are byte-identical to signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`. The only producer delta removes the old four-line permission note from root README to honor the explicit rc.7 wave rule; README and CHANGELOG now equal trunk exactly. The full mode table remains in docs. No other source, test or documentation edit. HEAD and the real index remain at the checkpoint; no producer commit was created. The candidate above is a tree identity, not a claim of a new signed commit; the configured gate owns that snapshot.

| Command | Real exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m -json -run '^Test(Managed\|Resolve)' ./internal/envprofile` | 0 | 33 top-level tests / 63 named passes, no skips; all six narrowing controls killed |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -timeout=8m -json -run '^Test(CodexSeed\|EnvironmentsCodexSeedVectors$\|EnvironmentWriteNofollowVectors$\|DotfileManagerVectors$\|UnreadableDotfileStateKeepsTakeoverQuiet$)' ./internal/envprofile` | 0 | 8 top-level tests / 40 named passes, no skips |
| `go test -count=1 -timeout=8m -json -run '^Test(Env\|ResolveTakeover\|UmbrellaSymlinkIntoManaged)' ./cmd/curator` | 1 | 59 top-level passes; six status failures from ambient untrusted `curator-engines` providers, not called green |
| Isolated previous-validation PATH: `go test -count=1 -timeout=3m -json -run '^TestEnvStatus(Matrix\|RegistryBoundaryPostureAndCheck)$' ./cmd/curator` | 0 | First two failing tests pass; removes ambient provider discovery, no product/test edits |
| Same isolated PATH: `go test -count=1 -timeout=4m -json -run '^TestEnvStatus(CheckCurrentScopeOnly\|MissingAndUnreadableKeepRecord\|UnreadableApprovalStateSurfaced\|ReportsShellHookTrustPosture)$' ./cmd/curator` | 0 | Remaining four failing tests pass |
| `go build ./...` | 0 | Build passes |
| `go vet ./internal/envprofile ./cmd/curator` | 0 | Vet passes |
| `git diff --exit-code HEAD -- README.md CHANGELOG.md`; `git diff --check`; `git diff --cached --exit-code` | 0 each | Wave docs unchanged, no whitespace errors, real index unchanged |
| `gofmt -l` on four changed Go files; explicit empty-output assertion | 0 each | Formatting clean |
| Initial `go test -count=1 -timeout=8m -json ./internal/envprofile` | 1 | Deliberately stopped owned test PID before timeout: prior package inventory measured 1329.734s. Partial stream is not passing evidence |

Current evidence is in `BUG-261018-dbeghc_publication-after-converge.json`, including exact base/blob/tree identities, test verdicts, stream digests, mutant child outputs and exits. Local macOS arm64 / Go 1.25.5. Conformance tests use the unchanged `43bf0a2506d5c354a73bbc3ea4623d4653db10c7` fixture root. Prior full-suite results remain historical; none is represented as a current full run. Per r5, full configured validation runs after handoff; AC5/checklist row 5 remains unchecked. No post-turn-validation checklist item was added. A runner infrastructure failure remains a blocker, never a pass.

Current product coverage remains **4 of 5 AC rows driven**: AC1 `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan`; AC2 `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/managedDirectory`; AC3 `TestManagedAuthoredFilesArePrivate` and `TestManagedFallbackPermissions` -> `Resolve/claudeSeed/applyPlan/copyLinkFallback/copyTree`; AC4 `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`, plus the controls below. `TestEnvResolveCredentialRecordProvisionAndRepair` additionally drives `cli.cmdEnvResolve` -> `envprofile.Resolve`. AC5 is deferred configured validation, not product behavior. Current r5 has no surface table (brief gap); prior product bounds and out-of-contract declarations above remain applicable. No reviewer test was removed/renamed; prior reviewer launches were freshness refusals, not a repeat-of verdict. Full platform matrix is unverified on this tree until the configured gate.

| Mutant | Gate narrowed to | Named failing test | Result / survivor bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only 0600 seeds widen to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed; mode 0644, expected 0600 |
| claude-trust-0644 | Only Claude trust output widens | `TestManagedAuthoredFilesArePrivate` | Killed; .claude.json 0644, expected 0600 |
| single-fallback-0644 | Only single-file fallback widens | `TestManagedFallbackPermissions/tree-false` | Killed; 0644, expected 0600 |
| executable-tree-fallback-0755 | Only 0755 sources evade owner mask, mask token preserved | `TestManagedFallbackPermissions/tree-true` | Killed; 0755, expected 0700 |
| codex-directory-drift-admitted | Only Codex drift is admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; wide home admitted |
| codex-directory-not-tightened | Only Codex protection is skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; repair still reports permission drift |

All six behavioral wrappers pass and require child `go test` exit 1 with the named failure and intended permission reason. No survivors. These bounds cover the named permission classes, not filesystem races. Out-of-contract scope remains spec changes (leaf Scope), refreshed tool-owned seed bytes/modes (section 7.4), native passthrough/link targets and takeover backups (ownership contracts). Windows has no new POSIX-bit preservation claim.

Changelog text for rc.7: preserve native seed modes, make managed routes owner-only, and tighten directory modes through repair without discarding contents.

Logbook: converge recovered the prior refresh refusal; root README now obeys the wave rule. Ambient user providers explain six diagnostic status failures, all resolved by rerunning under the previously documented isolated PATH. No shared process, runner, service, login or credential was changed. Evidence precedes role handoff; review acceptance and integration are not claimed.
