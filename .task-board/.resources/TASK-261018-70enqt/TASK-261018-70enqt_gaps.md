# TASK-261018-70enqt: qualification handoff and gaps

Candidate: Codex `0.162.1` qualification; source checkpoint
`402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`, uncommitted Story candidate.
The full per-field registry/evidence map and pilot observations are in
[README.md](README.md). Raw vendor evidence is in `evidence/`; source/test/
configuration digests are in `evidence/source-sha256.txt`.

No registry fact changed within the recorded qualification bounds, so
`VerifiedRelease` is `0.162.1`. The executable digest is
`74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0`,
macOS/arm64. No product behavior, spec text, or root README changed.

## AC coverage: 11 of 12 AC rows driven

The brief has no attack-surface table. This is a **brief gap**, not an implicit
waiver or a claim of catalog coverage. The following expanded AC rows map the
brief's commands/states to actual production entry points. Row 12 is an
explicit preservation/review bound, not a named behavioral test. Test sources
travel uncommitted in the mandated handoff snapshot; no producer commit was made.

| AC row | Named test / production call site | Evidence / limit |
| --- | --- | --- |
| 1. CODEX_HOME selects home | `TestCodex162HomeRootSkills` / `codex debug prompt-input` | `evidence/TestCodex162HomeRootSkills/{root-skills,empty-home-control}.log` |
| 2. AGENTS.md, size advisory, replace channel | `TestCodex162RootSize`, `TestCodex162ManagedAdvisory`, `TestCodex162ManagedChannels` / `envprofile.Install -> Resolve`, `codex debug prompt-input`, `codex exec -c model_instructions_file=...` | root tail, real managed symlink, 33666-byte warning, exact replacement developer message |
| 3. skills directory | `TestCodex162HomeRootSkills` / `codex debug prompt-input` | regular and linked skill discovery; empty-home negative |
| 4. accepted config.toml seed | `TestCodex162ExecSeedProfile`, `TestCodex162ManagedAdvisory` / `codex exec`, production `envprofile.Resolve` seed | loopback model selection; real seeded managed home loads |
| 5. linked auth, home scope, in-place write | `TestCodex162LinkedAuth` / `codex login status`, `codex login --with-api-key` | linked read, same inode/link after write, other home and missing link both exit 1 |
| 6. outside-git refusal | `TestCodex162ExecSeedProfile` / `codex exec` | confirmed non-Git cwd; without flag exit 1; with flag exit 0 |
| 7. single/missing -p and MCP channel | `TestCodex162ExecSeedProfile`, `TestCodex162ManagedChannels` / `codex exec -p`, `contextmaterialize.MCPFile -> codex -p curator-mcp mcp list --json` | layer changes model/server; missing layer silent; repeated flag exits 2 |
| 8. project-hook trust prompt | `TestCodex162ProjectHookTrust` / real `codex --no-daemon --no-alt-screen` TUI | first project prompts, consent persists hash, fresh same-project run skips review, second project prompts |
| 9. .git sandbox / separate layout | `TestCodex162GitSandbox` / `codex sandbox`, real `git init --separate-git-dir`, `git config` | primary/additional root .git denied; ordinary and descendant control writable; separate metadata writable with explicit root |
| 10. conditional version pin | `TestCodexCLI162VerifiedRelease` / production `envregistry.ByID(CodexCLI)` | `evidence/registry-02.log`; neighboring-release overlay fails the pin |
| 11. opt-in replay skipped by default | all eight `TestCodex162*` replay tests / gate before fixture creation or vendor launch | `evidence/default-04.log`: eight SKIPs, command exit 0; `replay-06.log`: eight PASSes with opt-in |
| 12. operator home and root README preserved | manual scope review and explicit environment allowlist, not a whole-home filesystem attestation | vendor homes/auth/config are synthetic; root README has no diff; normal host Go tool caches are not an asserted surface |

The primary-root and additional-root `.git` pilot facts reproduce. The stronger
recursive interpretation does **not**: the real descendant Git repository is a
counterexample, explicitly recorded in row 9 and the research record. That
recursive interpretation is not a registry field or a protection this change
introduces; it is not silently promoted to a qualification guarantee.

## Narrowing controls

Each individual expected-red `go test` below actually exits **1**. The runner
requires the named test failure and the intended diagnostic; it does not count
an unrelated error. `sh .research/codex162/mutants.sh` exits **0** only after all
seven controls are detected. Vendor wrappers alter one CLI contract case;
they are sensitivity controls, not mutations of proprietary executable internals.
The release control uses Go's source overlay and leaves the candidate untouched.

| Mutant | Gate/contract narrowed to admit | Named failing test | Evidence / survival bound |
| --- | --- | --- | --- |
| `skip-git` | one unflagged non-Git `exec --ephemeral probe` | `TestCodex162ExecSeedProfile` (`exit=0 want=1`) | `evidence/mutants/skip-git.log` |
| `repeated-profile` | exactly `-p one -p two`, retaining a profile flag | `TestCodex162ExecSeedProfile` (`exit=0 want=2`) | `evidence/mutants/repeated-profile.log` |
| `auth-other-home` | synthetic auth minted only in the otherwise empty second home | `TestCodex162LinkedAuth` (`exit=0 want=1`) | `evidence/mutants/auth-other-home.log` |
| `root-git` | workspace-derived permission profile permits only `.git/sentinel`; sandbox remains active | `TestCodex162GitSandbox` (`exit=0 want=1`) | `evidence/mutants/root-git.log` |
| `trust-second` | trust bypass only for the second project, keeping first-project review | `TestCodex162ProjectHookTrust` (`second entered main UI without hook consent`) | `evidence/mutants/trust-second.log` |
| `root-git-extra-root` | explicit writable-root exception for the primary `.git` only | `TestCodex162GitSandbox` (`exit=0 want=1`) | `evidence/mutants/root-git-extra-root.log` |
| `release` | attest the unqualified neighbor `0.162.0` while retaining the registry and other facts | `TestCodexCLI162VerifiedRelease` | `evidence/mutants/release.log` |
| historical top-level root override | attempted `.git` exception supplied at the wrong CLI level | **none — survivor**, suite exit 0 | `evidence/mutants/unapplied-root-override-survivor.log`; the override was not applied. Survival establishes no resistance to an effective exception; the corrected control above is killed |

An intermediate named-profile attempt failed because its config table was not
loaded, not because the prohibited write was admitted. It is **not** a killed
control (`evidence/mutants/profile-setup-error.log`). Both diagnostics were
corrected in the task-local runner; no product workaround was introduced.
There is no new source-text security checker, so the source-token-preserving
checker obligation is inapplicable. All controls execute the behavioral tests.

## Commands actually run

The command environment and binary identity are in the research README. All
gates ran directly, without `tee` or a pipeline hiding their exit status.

| Command | Actual exit | Evidence |
| --- | ---: | --- |
| `go test ./.research/codex162 -count=1 -v` (default) | 0 | `evidence/default-04.log`; eight vendor skips |
| same command with CODEX162_BINARY / EVIDENCE / TMPDIR | 0 | `evidence/replay-06.log`; eight vendor tests pass |
| `sh .research/codex162/mutants.sh` (opt-in environment) | 0 | `evidence/mutants/results.tsv`; each of seven child tests exits 1 as required |
| `go test ./internal/envregistry -count=1 -v` | 0 | `evidence/registry-02.log` |
| `go test ./internal/envprofile -run 'Test(Codex\\|EnvironmentsCodexSeedVectors\\|UnreadableCodexSeedDoesNotSelectAbsentDefault\\|ReviewerLiteralCodexSelector\\|PermissionsForProfileProperties)' -count=1 -v` | 0 | `evidence/managed-01.log` |
| `go test ./internal/envprofile -run '^TestReviewer' -count=1 -v` | 0 | `evidence/reviewer-regressions-01.log`; all four existing reviewer tests retain their original names |
| `go test ./internal/contextmaterialize -count=1 -v` | 0 | `evidence/materialize-01.log`; external conformance vectors skip without CURATOR_CONFORMANCE_ROOT |
| `go vet ./internal/envregistry ./.research/codex162` | 0 | `evidence/vet-04.log` |
| `golangci-lint run ./internal/envregistry/... ./.research/codex162/...` | 0 | `evidence/lint-04.log`: 0 issues |
| `go build -o .temp/codex162/curator ./cmd/curator` | 0 | `evidence/build-01.log` (no output) |
| `sh -n .research/codex162/mutants.sh` | 0 | `evidence/shell-syntax-03.log` (no output) |
| `git diff --check` | 0 | narrow working-tree whitespace validation |

The build, registry, managed-home and materializer evidence is reused after
changes confined to the research tests/docs: their production source, tests,
configuration and host/toolchain identities did not change. Replay, controls,
default skip, vet and lint were rerun after the last replay source change.
Early fixture/readiness errors and exploratory live unauthenticated CLI failure
are not green qualification evidence. Earlier attempts remain in ignored
`.temp/codex162/`; only the two relevant mutant diagnostics are published.

The landing-gate property suite `TestPermissionsForProfileProperties` is green
in `managed-01.log` and rerun green in `successor-credentials-after-01.log`.
The configured full hosted `sh scripts/remote-gate.sh` ran during revision-1
Change Request construction and **failed** (exit 1), as detailed below. It has
not been rerun on the reworked candidate by this producer; automated Change
Request validation and exact-head landing remain downstream. No hosted-CI
success is claimed.

## Out-of-contract rows and spec gaps

- Live OAuth refresh, real API success and native Keychain qualification:
  the brief asks for linked `auth.json` inheritance and in-place writes.
  These are measured with the file backend and synthetic API keys only.
- Model compliance with root instructions: "root context is loaded" is
  measured in the actual prompt/request, not by a live model's answer.
- Trusted-hook execution, every hook event, and arbitrary subdirectory trust:
  the pilot clause asks to record the startup trust prompt. This probe
  measures that screen, persisted trust and two distinct project sources.
- Linux/Windows sandbox behavior: the pilot is a macOS observation. A
  non-macOS skip is explicitly not a sandbox qualification.
- General HOME tree/cache immutability: the AC's operator HOME restriction
  is applied to vendor home/auth/config writes. This is not a whole-host
  filesystem monitor; root README preservation is reviewed separately.
- No `repeat-of:` class was supplied. Recovery from the revision-1 validation
  failure requires five dependent expected-value updates outside the brief's
  original module scope, listed below. They preserve all existing test names
  and assertions; no product behavior or spec text is changed.

CIP-0020/spec should record hook consent as separate from project trust,
root-level `.git` protection and its descendant/explicit-root limits, and the
separate-metadata layout's required writable metadata root and Git variables.
No spec text is edited by this task.

## Task-scoped logbook entry

2026-10-18: qualified the existing registry facts on the exact installed
0.162.1 executable in disposable homes; retained the advisory as a warning,
not a vendor global-root cap. Reproduced project-hook consent and separate Git
metadata. The descendant `.git` counterexample and explicit-root override limit
are recorded instead of a recursive protection claim. Seven controls are killed;
one historical unapplied override survives with no protection claim. Review and
full hosted integration checks remain downstream. This entry travels with the
task resource; the control-root LOGBOOK was not edited.

## Revision-1 validation recovery

The hosted gate for snapshot `c09b2f83e2525f8a3cfc85de44de7dd3d0395e68`
failed in [run 39101892700](https://github.com/relux-works/curator/actions/runs/39101892700).
The authoritative board resource is
`TASK-261018-70enqt_change-request_rev1-validation.log`.
The successor downloaded the complete `test-evidence-macos-latest` artifact
and extracted every failed named case into
`evidence/successor-hosted-failed-cases-01.log`: three internal marker tests
and two CLI tests all still expected `BackendVersion == "0.153.2"`.
Production `credentialPolicy` derives that field from the registry's
`VerifiedRelease`; the observed `0.162.1` was the intended qualification result.

The bounded scope exceptions are test expectations only:

| File outside the original scope | Change and reason |
| --- | --- |
| `internal/envprofile/credential_link_test.go` | One expected version updated in `TestCredentialLinkUnrecordedSymlinkRefuses`; the foreign-link refusal and unchanged-state assertions remain intact. |
| `internal/envprofile/credential_record_test.go` | Two expected versions updated in `TestResolvePublishesCredentialMarkerThroughLockedJournal` and its `V2WriterMode` counterpart; journal/lock/schema checks remain intact. |
| `cmd/curator/env_credential_marker_test.go` | Two expected versions updated for `TestEnvResolveCredentialRecordProvisionAndRepair` and `TestEnvResolveCredentialRecordPathlessKeyring`; production `env resolve --repair`, repair provenance and omitted-path checks remain intact. |

Remaining `0.153.2` strings in the version-comparison table and historical
marker round-trip fixture are independent sample data and intentionally kept.
No tests were removed or renamed. All four existing `TestReviewer*` tests
were rerun green. `TestPermissionsForProfileProperties` was rerun green.

The successor ran these standalone commands directly; paths in the evidence
column are relative to `evidence/`. Every listed log is present and untruncated.

| Command | Real exit | Evidence |
| --- | ---: | --- |
| `go test ./internal/envprofile -run '^TestCredentialLinkUnrecordedSymlinkRefuses$' -count=1 -v` before the expected-value update | 1 | `successor-regression-before-01.log`; expected stale-version assertion failure |
| `go test ./internal/envprofile -run '^TestResolvePublishesCredentialMarkerThroughLockedJournal(V2WriterMode)?$' -count=1 -v` before the update | 1 | `successor-record-before-01.log`; both stale-version assertions fail |
| `go test ./cmd/curator -run '^TestEnvResolveCredentialRecord' -count=1 -v` before the update | 1 | `successor-cli-before-01.log`; two stale-version assertions fail, adjacent keychain case passes |
| `go test ./internal/envprofile -run 'Test(Credential\\|Resolve.*CredentialMarker\\|PermissionsForProfileProperties\\|Reviewer)' -count=1 -v` | 0 | `successor-credentials-after-01.log`; includes all three corrected tests, negative credential cases, properties and all four reviewer tests |
| `go test ./cmd/curator -run '^TestEnvResolveCredentialRecord' -count=1 -v` | 0 | `successor-cli-after-01.log`; both corrected tests and adjacent keychain case pass |
| `go test ./internal/envregistry ./.research/codex162 -count=1 -v` without opt-in | 0 | `successor-registry-default-01.log`; registry passes and all eight vendor probes skip |
| `go vet ./internal/envprofile ./internal/envregistry ./.research/codex162` | 0 | `successor-vet-01.log` |
| `go vet ./cmd/curator` | 0 | `successor-cli-vet-01.log` |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/... ./internal/envregistry/... ./.research/codex162/...` | 0 | `successor-lint-02.log` |
| `go build -o .temp/codex162/curator ./cmd/curator` | 0 | `successor-build-01.log` |
| `git diff --exit-code -- README.md` | 0 | `successor-root-readme-01.log` |
| `go test -overlay .temp/codex162/successor-stale-release-overlay.json ./internal/envprofile ./internal/envregistry -run '^Test(CredentialLinkUnrecordedSymlinkRefuses\\|ResolvePublishesCredentialMarkerThroughLockedJournal(V2WriterMode)?\\|CodexCLI162VerifiedRelease)$' -count=1 -v` | 1 | `successor-stale-release-mutant-01.log`; intentional narrowing control, not a green suite |

| Mutant | What it narrows the gate to admit | Named failing tests | Survival bound |
| --- | --- | --- | --- |
| `stale-release` Go source overlay | Changes only the Codex registry release from `0.162.1` back to historical `0.153.2`, keeping the registry, credential policy and all other gates present | `TestCodexCLI162VerifiedRelease`, `TestCredentialLinkUnrecordedSymlinkRefuses`, `TestResolvePublishesCredentialMarkerThroughLockedJournal`, `TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode` | Killed with exit 1 and explicit wrong-version diagnostics; not a claim about arbitrary registry corruption. Original controls and historical survivor bounds above remain unchanged. |

To reproduce the extra overlay, copy `internal/envregistry/envregistry.go` to
an ignored scratch file with only `VerifiedRelease:   "0.162.1"` replaced by
`VerifiedRelease:   "0.153.2"`. A Go overlay JSON `Replace` entry maps the
absolute original filename to that scratch filename; run the exact command
above. The working-tree registry is never changed by this control.

Accepted unchanged evidence: the eight opt-in vendor probes in `replay-06.log`,
the original seven mutant controls and the materializer tests. Their replay
source, vendor executable digest, product sources, Go 1.25.5 darwin/arm64 and
configuration identity are unchanged. The successor verified the recorded
SHA-256 set before editing; the five changed expected values are now included
in the extended `source-sha256.txt`. No live vendor replay is claimed for this
successor. Full hosted CI for the reworked candidate remains unrun here and
must legitimately pass the automatic Change Request gate.

2026-10-18 recovery logbook: rev1 was red because dependent expected versions
were stale. Five expectations were updated, all named failures were reproduced
and rerun green, and a stale-release narrowing control fails the intended four
assertions. CLI suites waited for the repository's shared host GOROOT lock;
the lock was respected and no other session's processes were changed.
The Story HEAD remains at the recorded checkpoint, with the candidate
uncommitted. AC coverage remains **11 of 12 rows driven**, with row 12's
preservation bound unchanged. No control-root files were directly edited.

## Refresh onto 9c1605a7

2026-10-18 refresh attempt: **blocked before preparation**. Revision 2 remains
`ready`; changing the task status to `development` does not make the Change
Request a rework revision. No CHANGELOG resolution or checkpoint movement was
performed. The qualification source and existing tests were not changed.

| Command | Actual exit | Result |
| --- | ---: | --- |
| `task-board m 'set_status(TASK-261018-70enqt, status=development)'` | 0 | Task moved from to-review to development. |
| `task-board worktree refresh-candidate TASK-261018-70enqt` | 1 | `candidate refresh requires a rework revision; TASK-261018-70enqt is ready` |
| `task-board worktree status STORY-261018-evihyu` | 0 | Revision 2 ready, 118 changed paths; checkpoint remains `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`; live producer owns the lease. |
| `task-board worktree converge --help` | 0 | Documents an orchestrator-only command; tracked runs refuse with `transaction_disposition_not_operator`. |
| `git diff --exit-code -- README.md` | 0 | Root README unchanged. |
| `git diff --check` | 0 | Current candidate whitespace check passes. |

The refresh entry point admits only `changes_requested`, `stale`, `conflicted`,
`aborted`, `draft`, or a provably absent initial CR. Its refusal precedes Git
preparation. The installed project-management recovery contract likewise
requires a current rework revision. The production `withdraw_cr` entry point
requires orchestrator authority; the named source test
`TestMWithdrawCRIsRefusedFromInsideATrackedRun` explicitly covers a producer
bound to this element. This source inspection is not a test execution claim.
Neither operator-only action was attempted and the run identity was not unset.

**External input required:** the orchestrator must put revision 2 into a
supported rework-admissible state using the authorized CR lifecycle, then
resume the producer. If no supported transition exists for an unreviewed
ready revision beside a conflicting converge, repair the shared lifecycle
contract in its owning source repository before retrying. Recommended path:
resolve that lifecycle admission first, preserve the existing candidate, then
run refresh, resolve only CHANGELOG, and revalidate the resulting exact tree.
Manual branch/ledger edits, a fabricated reviewer verdict, and validation on
the old checkpoint cannot establish the requested refreshed candidate.

The full local suite, hosted gate, build, vet, lint and opt-in replay were
**not run in this refresh attempt**: no refreshed candidate exists. Historical
qualification evidence and its stated bounds above are retained; none is
claimed as validation of a refreshed tree. No new signed hosted snapshot was
created and no handoff was attempted. This is a lifecycle blocker, not a
qualification failure or a runner-infrastructure verdict.

Task-scoped logbook: r3 could not reach conflict resolution because the brief
assumed a rework CR while the authoritative record remains ready. Evidence is
attached before blocking; the Story candidate stays uncommitted and intact.


## Publication retry

2026-10-19, producer RUN-261018-77aa5f. No source, test or documentation
file was edited in the managed worktree in this retry. CHANGELOG.md and the
root README were not edited. The existing changelog delta belongs to the
previous candidate; this run neither removes nor republishes it as a new edit.

Last signed candidate located from the authoritative rev2 validation resource:
`da16449862e45250ab62168c0a1085109d768650`, tree
`5ea9873c866ef558a0f0bc07f67887b69ae06071`. Git signature verification reports
`G`. All 118 candidate-changed paths were compared directly against that
snapshot: **117 of 118 are byte-identical**. The only difference is the existing
`.research/codex162/TASK-261018-70enqt_gaps.md` r3 refresh-blocker appendix.
All six tracked candidate paths, including CHANGELOG.md, match the signed
snapshot (diff exit 0). There is no source/test/configuration refresh delta.
This resource adds only the publication-retry record; it does not alter the
worktree's research record or qualification artifacts.

Fresh authority resolved by `git ls-remote --symref origin HEAD` is `main` at
`b9caa20901ee859eec20e649940693033e6297d4`; the subsequent exact main fetch
agrees. CHANGELOG.md is the sole overlapping candidate path. Accordingly r4
requires refresh before publication. The production refresh entry point
refuses the current CR state before preparation: revision 2 remains `ready`.
Task status `development` does not grant a rework CR lifecycle transition.

| Command/check actually run | Real exit | Result |
| --- | ---: | --- |
| `task-board m 'set_status(TASK-261018-70enqt, status=development)'` | 0 | Task entered development; CR lifecycle unchanged. |
| `git ls-remote --symref origin HEAD` | 0 | Fresh main authority recorded above. |
| `git fetch origin main` | 0 | Remote-tracking main equals advertised authority. |
| `git show -s --format='%H %T %G? %s' da16449862e45250ab62168c0a1085109d768650` | 0 | Exact signed snapshot identified; signature G. |
| Python direct byte comparison of all 118 snapshot-changed paths | 0 | 117 identical; only prior r3 gaps appendix differs. |
| `git diff --exit-code da16449862e45250ab62168c0a1085109d768650 -- CHANGELOG.md cmd/curator/env_credential_marker_test.go internal/envprofile/credential_link_test.go internal/envprofile/credential_record_test.go internal/envregistry/envregistry.go internal/envregistry/envregistry_test.go` | 0 | All tracked candidate files match. |
| `task-board worktree refresh-candidate TASK-261018-70enqt` | **1** | `candidate refresh requires a rework revision; TASK-261018-70enqt is ready` |
| `git diff --exit-code -- README.md` | 0 | Root README unchanged. |
| `git diff --check` | 0 | Existing candidate whitespace valid. |

Evidence discrepancy: the attached rev2 validation log reports hosted run
39108873281 successful with command exit 0. It does not contain the curator-4
cleanup failure described by r4. This run does not infer another candidate or
claim a new hosted result from that mismatch. No full local suite, vendor
replay, build, lint or hosted gate ran in this publication retry: refresh was
refused and source/test/configuration identity is unchanged. Historical green
qualification evidence and all original negative/control bounds remain as
recorded above. No post-turn validation checklist items were added.

### Publication blocker and recovery

Constraint: a stale ready CR cannot be refreshed by its tracked producer.
The same production refusal was reproduced in r3 and now r4. No branch switch,
rebase, merge, commit, ledger edit, operator-identity bypass or control-root
file edit was attempted. The signed candidate and uncommitted work are preserved.

Required external input: the orchestrator must use the authorized CR lifecycle
to move revision 2 to a supported rework-admissible state, then resume the
producer for `refresh-candidate`. Recommended option: repair that admission,
resolve only the overlapping combination under the rc.7 ownership rule, and
publish through normal handoff with automatic exact-candidate validation.
Alternative: an operator-owned supported convergence before retry; any conflict
must respect the rc.7 prohibition on producer CHANGELOG/root README edits.
Manual Git or ledger rewriting would bypass ownership and is not a viable
producer workaround. If the supported lifecycle cannot express this transition,
its owning tool needs repair before publication. This is an ownership/lifecycle
blocker, not a qualification or runner failure.

Task-scoped logbook, 2026-10-19: publication retry preserved all candidate files,
verified the signed snapshot and fresh trunk overlap, and reproduced the
ready-CR refresh refusal (exit 1). No refreshed candidate or hosted-green claim
was made. Evidence was attached before blocking. Coverage remains **11 of 12 AC
rows driven**, with the existing preservation bound, brief surface-table gap,
out-of-contract rows, reviewer-test evidence and mutant table unchanged.


## Changelog entry

- Qualify the `codex_cli` adapter on Codex 0.162.1; disposable-home evidence
  and the opt-in replay are in `.research/codex162/` (TASK-261018-70enqt).

This is the text from the attached CODEX162-changelog-entry.diff. The rc.7
wave keeps CHANGELOG.md and root README equal to trunk; neither is edited here.


## Publication after converge

2026-10-19, r6, RUN-261018-b61901. The current managed checkpoint is
`59746a4545db06607446dcfe5fc35f28118a9125`, descended from the brief's
`892dab098aef1497df9ccacdb173faaf4f8b8418` through the native-skills landing
and its board-state commit. Local HEAD..main distance is 0; this is not a
fresh remote-authority claim. Worktree status reports revision 2 stale and
the current run's lease; it does not direct a producer refresh. No
refresh-candidate, source/test/documentation edit, branch operation or producer
commit was performed. The candidate remains uncommitted for handoff.

All five tracked candidate files match the last signed candidate
`da16449862e45250ab62168c0a1085109d768650` (git diff exit 0).
Across its 118 changed paths, 116 remain byte-identical. The two differences
are the pre-existing r3 research-gaps appendix and CHANGELOG.md, which the
orchestrator restored to trunk before convergence. README.md and CHANGELOG.md
both equal the current checkpoint (exit 0). The changelog entry is preserved
in this outcome resource above. The tracked diff is unchanged throughout this
run; the replay source and seven vendor-control definitions are unchanged.

Current dependencies differ from the historical qualification checksum set:
upstream changed go.mod/go.sum. The source identities in the new evidence
bundle supersede those old dependency hashes for these reruns. Upstream also
changed envprofile and CLI native-skills paths, so affected checks and the real
vendor replay were rerun rather than claiming the historical suites apply.
Codex still has the recorded 0.162.1 executable SHA-256. All eight probes pass
using disposable homes; no real Codex home/auth/config is imported or written.

Plain-text evidence bundle: `TASK-261018-70enqt_publication-r6-evidence.log`.
Each command ran directly with file redirection and its actual exit captured.
Paths below are bundle entries; private home prefixes in diagnostics are
redacted. Full vendor command logs are in `r6-replay-evidence/`.

| Command/check | Real exit | Evidence / result |
| --- | ---: | --- |
| `go test ./internal/envregistry ./.research/codex162 -count=1 -v` | 0 | r6-registry-default-01.log; pin passes, eight vendor skips |
| `go test ./internal/envprofile ./cmd/curator -count=1` | **1** | r6-affected-tests-01.log; both package suites hit 10m timeout; four CLI named failures also recorded. This is not a green full suite. |
| `go test ./internal/envprofile -run '^Test(Codex\|EnvironmentsCodexSeedVectors\|UnreadableCodexSeedDoesNotSelectAbsentDefault\|Reviewer\|PermissionsForProfileProperties\|CredentialLinkUnrecordedSymlinkRefuses\|ResolvePublishesCredentialMarkerThroughLockedJournal)' -count=1 -timeout 4m -v` | 0 | r6-managed-focused-01.log; affected managed-home scope, permission properties and all four original reviewer tests |
| `go test ./cmd/curator -run '^TestEnvResolveCredentialRecord(ProvisionAndRepair\|PathlessKeyring\|IsolatedKeychain)$' -count=1 -timeout 4m -v` | 0 | r6-cli-focused-01.log; all three relevant credential scenarios |
| `go test ./cmd/curator -run '^Test(DraftStatusReportsUpToDateThroughCLI\|DraftStatusNeedsInstallAfterRefreshWithoutInstall\|DraftStatusLiveMutationWithoutRefreshStaysCurrent\|DraftTransportLegacyGolden)$' -count=1 -timeout 4m -v` with isolated PATH | 0 | r6-cli-host-isolation-01.log; all four previously failed named cases green; no source edits |
| `CODEX162_BINARY=<installed executable> CODEX162_EVIDENCE=<task scratch>/r6-replay-evidence TMPDIR=<task scratch> go test ./.research/codex162 -count=1 -v` | 0 | r6-replay-01.log; eight real probes pass, exact Codex identity in each fixture |
| `go build ./...` | 0 | r6-build-01.log |
| `go vet ./internal/envregistry ./internal/envprofile ./cmd/curator ./.research/codex162` | 0 | r6-vet-01.log |
| `golangci-lint run ./internal/envregistry/... ./internal/envprofile/... ./cmd/curator/... ./.research/codex162/...` | 0 | r6-lint-01.log; 0 issues |
| `git diff --check` | 0 | r6-whitespace-01.log |
| `git diff --exit-code HEAD -- README.md CHANGELOG.md` | 0 | r6-preservation-01.log |
| `go test -overlay <task scratch>/r6-stale-overlay.json ./internal/envregistry -run '^TestCodexCLI162VerifiedRelease$' -count=1 -v` | **1** | r6-mutant-01.log; expected-red narrowing control |
| `rg -n '/Users/\|Bearer \|sk-proj-' <task scratch>/r6-replay-evidence` | **1** | r6-publication-scan-01.log; bounded no-match scan, not a green test command |

The isolated PATH contains the actual Go toolchain bin, /usr/bin, /bin,
/usr/sbin, /sbin and /opt/homebrew/bin. The initial draft-status failures
reported ambient operator provider binaries as untrusted; all four named
CLI failures pass in the bounded isolated rerun. The golden case's initial
failure is retained; no more specific root-cause assertion is made for it.
The full two-package command remains failed/incomplete, not superseded by a
claim of full-suite success. Current r6 requires checks covering changed paths
and build; those bounded checks are green. No full repository suite or hosted
gate was run by this producer. Configured post-turn validation remains the
authoritative automatic gate; no checklist items were added for it.

| Mutant | Gate narrowed to admit | Named failing test | Survival bound |
| --- | --- | --- | --- |
| stale-release source overlay | Only historical 0.153.2 registry attestation; registry and other gates stay present | TestCodexCLI162VerifiedRelease | Killed, exit 1; no arbitrary-registry-corruption claim |

Historical vendor controls and their survivor table above are retained as
previous evidence, not rerun controls. AC coverage remains **11 of 12 rows
driven**, with row 12's preservation bound unchanged. No surface table was
provided; that brief gap and all out-of-contract clauses above remain declared.
No new gate, source-text checker or repeat-of class is introduced.
The existing three outside-original-scope test files remain justified by the
revision-1 expected-version recovery; r6 adds no repository delta.

Task-scoped logbook: convergence has removed the old ready-CR refresh blocker.
Recorded current checkpoint, unchanged candidate bytes, fresh default and
opt-in probe evidence, green bounded affected checks/build/vet/lint, the
expected-red release mutant, and the full-package timeout/host-path diagnostics.
The control-root LOGBOOK and board files were not edited directly. Normal
review routing is requested through handoff; this does not claim review
acceptance, hosted success, landing or task completion.

Resource publication: the archive attachment was refused (exit 1, resource_binary_refused); all selected raw outputs are instead attached in the plain-text evidence bundle above. No binary resource was persisted.
