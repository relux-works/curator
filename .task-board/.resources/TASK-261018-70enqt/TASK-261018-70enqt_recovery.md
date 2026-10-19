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
| `go test ./internal/envprofile -run 'Test(Codex\|EnvironmentsCodexSeedVectors\|UnreadableCodexSeedDoesNotSelectAbsentDefault\|ReviewerLiteralCodexSelector\|PermissionsForProfileProperties)' -count=1 -v` | 0 | `evidence/managed-01.log` |
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
| `go test ./internal/envprofile -run 'Test(Credential\|Resolve.*CredentialMarker\|PermissionsForProfileProperties\|Reviewer)' -count=1 -v` | 0 | `successor-credentials-after-01.log`; includes all three corrected tests, negative credential cases, properties and all four reviewer tests |
| `go test ./cmd/curator -run '^TestEnvResolveCredentialRecord' -count=1 -v` | 0 | `successor-cli-after-01.log`; both corrected tests and adjacent keychain case pass |
| `go test ./internal/envregistry ./.research/codex162 -count=1 -v` without opt-in | 0 | `successor-registry-default-01.log`; registry passes and all eight vendor probes skip |
| `go vet ./internal/envprofile ./internal/envregistry ./.research/codex162` | 0 | `successor-vet-01.log` |
| `go vet ./cmd/curator` | 0 | `successor-cli-vet-01.log` |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/... ./internal/envregistry/... ./.research/codex162/...` | 0 | `successor-lint-02.log` |
| `go build -o .temp/codex162/curator ./cmd/curator` | 0 | `successor-build-01.log` |
| `git diff --exit-code -- README.md` | 0 | `successor-root-readme-01.log` |
| `go test -overlay .temp/codex162/successor-stale-release-overlay.json ./internal/envprofile ./internal/envregistry -run '^Test(CredentialLinkUnrecordedSymlinkRefuses\|ResolvePublishesCredentialMarkerThroughLockedJournal(V2WriterMode)?\|CodexCLI162VerifiedRelease)$' -count=1 -v` | 1 | `successor-stale-release-mutant-01.log`; intentional narrowing control, not a green suite |

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
