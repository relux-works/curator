## Status
done

## Review
light

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- TASK-260823-1l1p8q

## Checklist
- [x] Merged to main (or candidate branch where applicable) with green CI; candidate case verified
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260823-feedc0, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260823-feedc0)
Root cause split: candidate vector hardcoded Darwin/arm64 closed environment; Go conformance adapter rewrote symlink target separators on Windows before hashing. Published signed superseding candidate edd07210d4f3db34fd60238cb14b90f837de03cb (manifest 803918bf..., tree 9d5a10b6..., 692 files) after byte-identical double regeneration; old candidate retained. Curator signed fix fbca88617c3765cfa40c1284035429962bf81bda pushed with draft PR #29. Local focused tests, full candidate gate (41/0/0), lint, build, spec validator/Python/Go tool gates green. One earlier full go test exited 1 honestly due concurrent host-GOROOT hashing timeout plus ENOSPC; later full candidate package gate exited 0 and the formerly timed-out test passed. Remote runs 32642316308, 32642306296, 32642340559 in progress; Windows spec job already green.
agent completed: [implementer] developer (codex) (exit=124)
spawn run completed: codex (run=RUN-260823-feedc0, pid=96330, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-3e7189, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-3e7189)
REVIEW VERDICT: CHANGES REQUESTED (to-dev). AC not met — only one of the two owned Windows cases is green against candidate edd0721.

GREEN and accepted: fixed_environment. The candidate now publishes fixed_environment_cases for darwin/arm64, linux/amd64, windows/amd64 with validate_fixed_environment_cases pinning coverage, arch tuning and the Windows private-process variables; curator selects the native case and tolerates the declared optional SYSTEMROOT/WINDIR. Verified green twice: native Windows host (go1.25.5 windows/amd64) and CI Candidate suite (windows-latest) in run 32642340559.

RED and blocking: TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link. The fix in fbca886 (PR 29, merged 351db49) removed the explicit filepath.FromSlash from materializeToolchain, but Go os.Symlink on Windows applies FromSlash itself before the syscall (os/file_windows.go, first statement of func Symlink) and os.Readlink returns a relative reparse target verbatim. So the removal is a no-op on Windows and the newly added round-trip assertion is guaranteed to fire — builddriver_positive_conformance_test.go line 267 reports the materialized link target as backslash form while the protocol target is ../bin/go. One Windows red was converted into a different Windows red. Reproduced on a native Windows host and confirmed by CI Candidate suite (windows-latest) 32642340559.

Guilty side is the implementation, not the materializer and not the vector: internal/godriver/fingerprint.go lines 198-220 store record.link as the raw host-native Readlink output and hash it into the curator-go-toolchain-v1 preimage, making the toolchain digest host-dependent. The candidate preimage carries the 9 platform-neutral bytes of ../bin/go and is correct. Required rework: normalize the link target to the protocol slash form inside the fingerprint via a build-tagged helper (windows ToSlash, unix verbatim — a blanket ToSlash would corrupt a Unix filename containing a backslash), compare through the same normalizer in the test assertion, and re-verify on a real Windows host (ssh win reproduces in about 3s). No candidate regeneration needed for this half; edd0721 and manifest 803918bf stay valid.

VERIFIED GOOD, keep: candidate edd0721 is signed, pushed, stacked on 859727b without rewriting it; shasum of conformance/v1/manifest.json independently recomputed as 803918bf; release/1.0.0-rc.9.json advances both pins; double regeneration inventories are byte-identical and their build-drivers.json digest f9c77e49 matches the published file; the new identity IS recorded on TASK-260822-c0rxj7 with regeneration proof, so that half of the AC is satisfied.

Evidence gap that let this land: the pinned default root publishes no vectors/build-drivers.json, so the ordinary Test (windows-latest) lane skips both owned tests — PR 29 being all-green proves nothing here, and the producer only ran the focused and full gates on macOS.

Out of scope, do not chase: dispatch 32642340559 ran on fbca886, which predates PR 26 and PR 28; its other reds (buildsource duplicate and invalid-unicode paths, invalid-unicode-toolchain-path, DryRunEffectBindings, install/atomicity platform-case gate) are other tasks and are already fixed on main. Dispatch the rerun from current main.

Nit: fixedEnvironmentForHost calls t.Fatalf on an unpublished host (Intel Mac, linux/arm64) where t.Skipf would be right.

Artifacts: TASK-260823-czs1cx_review-verdict.md, TASK-260823-czs1cx_windows-verification.log.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-3e7189, pid=20185, exit=0)
REWORK CYCLE 2 per review-verdict.md (RUN-260823-3e7189): the PR 29 test edit is a no-op on Windows — os.Symlink performs FromSlash itself before the syscall, so the byte-exact round-trip assertion does not exercise anything. Required: (1) normalize the link target to protocol slash form IN THE TOOLCHAIN FINGERPRINT, before validation and hashing, via a build-tagged helper (windows: filepath.ToSlash; unix: verbatim — blanket ToSlash would corrupt legitimate backslashes in Unix filenames); (2) keep the materializeToolchain round-trip assertion but compare through the same normalizer; (3) verify on the real Windows host: ssh win (go1.25.5 windows/amd64, reproduces in ~3s) — macOS-only evidence is insufficient, the default pinned root publishes no vectors/build-drivers.json so the ordinary windows Test lane skips these; (4) no candidate regeneration needed for this half. Land via PR with EVERY lane verified green pre-merge. Executor policy: claude only.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260823-bd9292, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260823-bd9292)
REWORK CYCLE 2 (RUN-260823-bd9292) implemented. Confirmed the reviewer diagnosis independently: os.Symlink on Windows applies filepathlite.FromSlash as its first statement (verified in GOROOT src/os/file_windows.go), so PR 29 was a no-op there; and the candidate preimage for unsorted-directories-files-and-internal-link decodes to the 9 platform-neutral bytes ../bin/go with content_sha256 baf7c5f3... — the vector is correct, the implementation was guilty. Fix: build-tagged protocolLinkTarget (windows filepath.ToSlash, unix verbatim) applied in internal/godriver/fingerprint.go before validation and hashing, plus the same point in the equivalence gate reference traversal (otherwise the two traversals diverge on Windows), plus the materializeToolchain round-trip assertion now compares through the same normalizer. Test: TestFingerprintImplementationMatchesRC4ToolchainVector moved out from under the unix build tag; it pins exactly the vector published digest baf7c5f3... so running it on every host IS the host-independence claim. Proved load-bearing on native Windows by neutralizing only the normalizer: both cases FAIL without it (digest forks to e9b9a60b...), PASS with it. Windows evidence: focused tests exit 0 and FULL internal/godriver exit 0 against candidate edd0721 on DESKTOP-3PBO632 go1.25.5 windows/amd64. macOS: make candidate-test exit 0, make check-ci exit 0, go vet darwin/linux/windows exit 0, golangci-lint darwin/linux exit 0. GOOS=windows golangci-lint reports 10 findings, all pre-existing in *_windows.go files this change does not touch. No candidate regeneration needed; edd0721 and manifest 803918bf stay valid. Also applied the reviewer nit: fixedEnvironmentForHost now Skipf instead of Fatalf. Landed as signed commit 695c041 on fix/TASK-260823-czs1cx-toolchain-link-normalize, PR 31 open, awaiting all lanes green pre-merge.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260823-bd9292, pid=39782, exit=124)
Rework 2 landed: PR 31 (build-tagged toolchain link-target normalization in the fingerprint, round-trip assertion through the same normalizer, cross-platform TestFingerprintImplementationMatchesRC4ToolchainVector pinning the digest on default lanes) merged as 2671743 with every lane green pre-merge; real-Windows verification log attached (rework2-windows-verification.log via ssh win). Candidate identity unchanged: edd0721 / 803918bf. Producer run timed out post-delivery; orchestrator finished the mechanical merge.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-ee0acd, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-ee0acd)
REVIEW VERDICT CYCLE 2 (RUN-260823-ee0acd): ACCEPTED (done). Both AC halves met and verified independently, not taken on the producer's word.

Owned cases green on native Windows (DESKTOP-3PBO632, go1.25.5 windows/amd64) against candidate edd0721, run by the reviewer: fixed_environment PASS, unsorted-directories-files-and-internal-link PASS, TestFingerprintImplementationMatchesRC4ToolchainVector PASS, and the FULL internal/godriver package exit 0 (123s). CI agrees for the right reason — Candidate suite (windows-latest) in run 32645833559 verifies in-job that it checked out CANDIDATE_REF edd0721 and that the manifest digest matches 803918bf, with suite-plan served=42 deferred=0 under CI_REQUIRE_FULL_ROOT=1. That is the lane that was red on this case last cycle. All lanes on 695c041 green.

Fix is on the guilty side and load-bearing, reproduced by the reviewer: neutralizing ONLY protocolLinkTarget on Windows makes both cases fail and forks the digest to e9b9a60b against the published baf7c5f3; restoring it returns exit 0. The candidate preimage was correct all along.

Adversarially checked the normalization point rather than just the happy path: VolumeName still catches C:/ after ToSlash, HasPrefix "/" catches //?/ and /foo, utf8.ValidString cannot be weakened because 0x5C is never a UTF-8 continuation byte, and escaping/absolute/duplicate/invalid-unicode toolchain cases all PASS on Windows. record.link has exactly two uses so nothing downstream sees the host-native form. The equivalence gate reference traversal had to move too and did.

Fatalf->Skipf nit closed and safe: the candidate's validate_fixed_environment_cases pins the host set to exactly darwin-arm64/linux-amd64/windows-amd64, which is exactly the CI matrix, so a dropped case fails validation instead of skipping.

Candidate identity unchanged and correct: reviewer recomputed shasum of conformance/v1/manifest.json at edd0721 as 803918bf independently; TASK-260822-c0rxj7 still records edd0721 / 803918bf / tree 9d5a10b6 / 692 files with double-regeneration proof, superseded 859727b unrewritten. Merge 2671743 introduces exactly the reviewed content (diff-of-diffs vs 695c041 differs only in hunk offsets from PR 30 landing in between).

Gates: GOOS=windows go vet exit 0; darwin golangci-lint v2.12.2 0 issues.

TWO NON-BLOCKING FINDINGS, do not reopen this task for them:
(1) The claimed regression guard is real but unenforced. TestFingerprintImplementationMatchesRC4ToolchainVector does run and pass on the reviewer's Windows host with no conformance root, but it is absent from .github/ci/platform-cases.tsv and its "symlink unavailable" reason matches skip-classes.tsv:62 (host-capability, allow everywhere), so a runner forbidding unprivileged symlink creation would skip it green — and platform-cases.tsv:124-125 says the Windows runner may do exactly that. Cheap follow-up: add a Tier-1 row with must_run_on including windows plus the host-capability tolerance, so a rename or deletion fails by name.
(2) GOOS=windows golangci-lint reports 10 findings (5 errcheck, 4 gosec, 1 revive) in buildcache/buildrepo/managerlock/transaction/godriver *_windows.go files. Reviewer reproduced all 10; the producer's count and pre-existing classification were accurate and none are in a touched file. Worth noting the CI Lint job is ubuntu-only, so no gate covers Windows-tagged code at all.

Artifacts: TASK-260823-czs1cx_review-verdict-cycle2.md, TASK-260823-czs1cx_reviewer-verification-cycle2.log.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-ee0acd, pid=17145, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260823-czs1cx_spawn-log_-implementer--developer--codex-_RUN-260823-feedc0.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_spawn-log_-implementer--developer--codex-_RUN-260823-feedc0.log) — System spawn log captured by task-board
- [TASK-260823-czs1cx_results.md](file://TASK-260823-czs1cx/TASK-260823-czs1cx_results.md) — Root-cause decisions, candidate identity, delivery commits, PR, and validation evidence
- [TASK-260823-czs1cx_candidate-suite-identity.txt](file://TASK-260823-czs1cx/TASK-260823-czs1cx_candidate-suite-identity.txt) — Superseding candidate SHA, manifest digest, tree digest, and file count
- [TASK-260823-czs1cx_landing-evidence.md](file://TASK-260823-czs1cx/TASK-260823-czs1cx_landing-evidence.md)
- [TASK-260823-czs1cx_spawn-log_-reviewer--reviewer--claude-_RUN-260823-3e7189.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_spawn-log_-reviewer--reviewer--claude-_RUN-260823-3e7189.log) — System spawn log captured by task-board
- [TASK-260823-czs1cx_review-verdict.md](file://TASK-260823-czs1cx/TASK-260823-czs1cx_review-verdict.md) — Reviewer verdict RUN-260823-3e7189: changes requested; toolchain link case still red on Windows, fixed-environment half accepted
- [TASK-260823-czs1cx_windows-verification.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_windows-verification.log) — Native Windows (go1.25.5 windows/amd64) run of both owned conformance tests against candidate edd0721 plus os.Symlink/os.Readlink round-trip probe
- [TASK-260823-czs1cx_spawn-log_-implementer--developer--claude-_RUN-260823-bd9292.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_spawn-log_-implementer--developer--claude-_RUN-260823-bd9292.log) — System spawn log captured by task-board
- [TASK-260823-czs1cx_rework2-evidence.md](file://TASK-260823-czs1cx/TASK-260823-czs1cx_rework2-evidence.md) — Rework cycle 2: diagnosis confirmed from primary sources, build-tagged normalizer fix, load-bearing proof on native Windows, and the full gate matrix with real exit codes
- [TASK-260823-czs1cx_rework2-windows-verification.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_rework2-windows-verification.log) — Native Windows (DESKTOP-3PBO632, go1.25.5 windows/amd64) against candidate edd0721: both owned cases PASS with the fix, both FAIL with the normalizer neutralized, full internal/godriver package exit 0
- [TASK-260823-czs1cx_spawn-log_-reviewer--reviewer--claude-_RUN-260823-ee0acd.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_spawn-log_-reviewer--reviewer--claude-_RUN-260823-ee0acd.log) — System spawn log captured by task-board
- [TASK-260823-czs1cx_review-verdict-cycle2.md](file://TASK-260823-czs1cx/TASK-260823-czs1cx_review-verdict-cycle2.md) — Reviewer verdict RUN-260823-ee0acd: ACCEPTED; both owned Windows cases green vs candidate edd0721, fix reproduced load-bearing on a native Windows host
- [TASK-260823-czs1cx_reviewer-verification-cycle2.log](file://TASK-260823-czs1cx/TASK-260823-czs1cx_reviewer-verification-cycle2.log) — Reviewer's own runs: macOS + native Windows focused, normalizer neutralized/restored, full internal/godriver, GOOS=windows vet, darwin and windows lint

## Created
2026-08-23T12:43:53Z

## Last Update
2026-08-23T15:19:25Z

## Assigned To
[reviewer] reviewer (claude)
