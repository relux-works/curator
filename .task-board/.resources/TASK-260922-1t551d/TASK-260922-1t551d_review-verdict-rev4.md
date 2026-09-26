# TASK-260922-1t551d revision 4 — independent review

Verdict: ACCEPTED under the binding rework-2/review-rev4 rulings, which supersede the original blanket dangling/conflict and recorded-mistarget refusal wording. Acceptance is for integration, not landed delivery.

## Identity and gate

Base: 09b25ef6629b41455d91dcb252ab4e4034e12750.
Candidate: 520a182a4336af6d24b1c177f513d2613b881dbd.
GitHub run 35690797842 succeeded; its head fd98c94b321e0df7023ce32061118717efdfabeb resolves to this exact tree. Independently queried run/job conclusions with gh and inspected the board validation log. Hosted Ubuntu/macOS tests and race, Windows tests, lint, and gate self-tests passed. Candidate-suite and rose-air jobs were skipped, not passing. No local full landing-suite replay.

Review used a disposable git-archive copy under the assigned worktree's ignored .temp directory. All 13 changed paths byte-match the candidate in both the assigned worktree and restored copy. No candidate code edits, commits, schema changes, or host credential changes were made.

## Findings and production coverage

- H1: finalizeMarker at managed.go:1012 now removes prior recorded passthrough links even when effectivePassthrough is empty for isolated. removeStaleCredentialLinks at :1144 checks both symlink shape and full declared target before unlinking. Shared-to-isolated and regular-file/foreign-target/directory refusal rows drive Resolve.
- H2: wanted credentials go through ensureCredentialLink at :1008/:1083 instead of unconditional removal. Empty and nonempty regular files refuse with the named path, native and managed bytes preserved. Unrecorded foreign symlinks refuse. Under the newer ruling, recorded mis-targeted wanted links can be re-pointed; stale unwanted links retain stricter removal checks.
- checkPassthrough at :1552, called by verifyHome at :1342, distinguishes mis-targeted links (stale conflict), declared-but-absent targets (detached-pending warning, success), and target inspection failures (stale conflict). StatusOf surfaces the same findings/warnings. Both former reviewer probes and the pending-repair row pass. The operator Pi reproduction uses the old missing target alongside the actual agent-root file.
- buildPlan's dynamic isolated gate (:278/:587) reads native config via existing github.com/BurntSushi/toml (:550), considers only the top-level key, admits absence/file, refuses keyring/auto and unknown/non-string values with their respective classes. Literal/multiline/table/malformed rows pass. Reviewer-added unreadable config-directory probes refuse before creating a home in both shared and isolated modes.
- Pi auth strategy points to agent/auth.json below the existing native base. Frozen v1 marker remains path+strategy only. New production reads concern config.toml; credentials are inspected with metadata calls, not read/copied.
- CHANGELOG and troubleshooting describe the revised three-state behavior. The registry's older comment about F-C2-only re-pointing is stale explanatory prose; executable behavior and user-facing documentation follow the explicit newer ruling. No delivery-blocking defect found.

## Independent verification and bounds

Shell: zsh; Go commands executed with -count=1 and bounded masks/timeouts. Final logs are attached separately.

1. Focused credential/refusal/TOML/Pi rows plus TestResolveClaudeProjectEntry and TestClaudeSeedMergePreservesToolState: PASS, 4.770s. Those Claude rows ran locally on Darwin; their Linux file-link behavior is covered by the exact-tree hosted Ubuntu gate, not claimed from Darwin.
2. Explicit mis-targeted Pi, pending-repair, and both prior reviewer probes: PASS, 9.036s.
3. Corrected independent unreadable-config probe plus prior reviewer probes: PASS, 10.661s. Initial shared-mode expectation was too specific: gatherSeeds reports environment_seed_unreadable before the selector reader. Corrected assertion requires unreadable + config.toml and no managed-home creation; no production fix was needed.
4. envregistry package completed independently: PASS, 0.369s.
5. Independent CLI probe alone: PASS, 11.994s. It installs a temporary profile through the real CLI, drives env resolve pi --repair, requires the pending warning with success, and requires env status --json to display detached-pending. Earlier broad CLI attempts are explicitly incomplete below.
6. git diff --check: PASS. Lint/vet/build coverage otherwise reused from verified exact-tree CI.

Initial broad envprofile and CLI masks were stopped with SIGQUIT after roughly 449s/422s in unrelated overlay/profile-install work; exit 1, not counted as passing. The initial extra CLI attempt timed out at 150s waiting in TestMain for AcquireHostGOROOT; its stack is attached. A narrower TestEnv.* attempt also timed out at 120s in the pre-existing TestEnvStatusMissingAndUnreadableKeepRecord, not a credential assertion failure. These incomplete runs are not claimed green. Targeted API reruns replace them for credential review scope. No Windows/macOS/Linux equivalence is inferred beyond named hosted evidence; rose-air is unverified.

## Mutation evidence

11/11 selected mutations produced named behavioral assertion failures, zero compilation-error kills. Scripts and complete logs attached; source restored and hashed afterward.

| Mutation | Killer |
|---|---|
| Omit recorded stale removal | TestSharedToIsolatedRemovesStaleLink |
| Restore regular-file removal | TestCredentialLinkRegularFileRefuses, empty and nonempty |
| Silence expected-target pending warning | TestReviewerDanglingExpectedTarget |
| Collapse pending into stale conflict | TestDanglingExpectedTargetRepairSucceeds |
| Admit auto, retain keyring refusal | TestReviewerLiteralCodexSelector/auto |
| Admit ephemeral only | TestReviewerLiteralCodexSelector/ephemeral |
| Invalid TOML becomes absent/file | TestCodexCredentialStoreTOMLSpellings/invalid_TOML_fails_closed |
| Re-point unrecorded foreign link | TestCredentialLinkUnrecordedSymlinkRefuses |
| Exempt only empty regular files | TestCredentialLinkRegularFileRefuses/empty |
| Compare wanted-link basename only | TestDanglingPiLinkReportedDetached |
| Compare stale-link basename only | TestStaleCredentialLinkRefusals/foreign_symlink |

This is 11 selected attacks, not exhaustive clause coverage. Inspection-failure committed row uses ENOTDIR on Unix, not a direct EACCES fixture; Windows skips that shape with the existing host-capability vocabulary. F-C2 credential-byte migration, F-C3 broader unreadability classification, and F-S1 extended records remain out of scope. Producer results contain historical sections superseded by their Revision 3/4 addenda; reviewed final behavior against the latest rulings.

Run goal queried: not goal-bound. Record exactly one accept_cr for revision 4; producer integration owns subsequent closure.
