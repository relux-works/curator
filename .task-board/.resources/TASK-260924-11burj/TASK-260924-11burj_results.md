# TASK-260924-11burj results

## Delivered

Added a production-CLI acceptance scenario using two local bare Git repositories and temp homes. One schema-2 collection entry selects `skills/*` from the playbook repository; the `developer` manifest declares `qa` from `roles/qa` in the second repository. The scenario covers resolve/install, advisory source audit, status, refresh after adding `writer`, committed-lock replay on a new home, a moved tag, snapshot tampering, unavailable remotes, and the three required selector/dependency refusals.

Fresh-home replay needed to recover a transitive network Git source from the verified frozen manifest of its locked requirer. Replay fetches the locked commit, checks repository identity and selected directory, and validates the resulting snapshot against the lock. It does not resolve the moved tag. A non-root network Git directory also shares its canonical repository cache identity.

## Production CLI scenario

The test builds/uses the production `curator` binary through the existing CLI test harness; Git URLs are mapped by a local wrapper to local bare repositories. All fixture data is local.

| Scenario command | Exit | Asserted outcome |
|---|---:|---|
| `curator project resolve app` | 0 | Schema-2 collection resolved from `v1.0.0` |
| `curator install app --audit advisory` | 0 | Three selected playbook skills plus transitive `qa` installed; audit records bind all four lock members |
| `curator status app` | 0 | Each installed skill reports up-to-date |
| `curator project refresh app` | 0 | Refresh adopts playbook `v1.1.0` |
| `curator install app --audit advisory` | 0 | New `writer` installed; lock changed; audits bind all five members |
| `curator status app` | 0 | All five installed skills report up-to-date |
| fresh-home `curator install app --audit advisory` | 0 | Replays committed lock and source content; lock remains byte-identical; status up-to-date; five audit bindings present |
| fresh-home install after cached snapshot tamper | 1 | Refused with `source_snapshot_changed`; lock remains byte-identical |
| fresh-home install with unavailable local remotes | 1 | Refused with `source_snapshot_unavailable` |
| resolve with include `missing` | 1 | Refused with `source_member_missing` |
| resolve with directory `../outside` | 1 | Refused with `source_selection_invalid` |
| resolve dependency directory `roles/no-skill` | 1 | Refused because it has no `SKILL.md` |

The fresh-machine case cloned the project after committing `Skillfile.json` and `Skillfile.lock.json`, then moved the remote `v1.1.0` tag to a commit containing an additional `unlocked` skill. Replay retained the original lock commit and did not install `unlocked`.

### Refreshed lock excerpt

All four playbook collection members share commit `b95ce8bcb00c0b39a02aec7fb0d41c219289d3dc`; the transitive dependency uses roles commit `b7a996097286f589f5951974e469d13f4a5c2874`. The same values were asserted after fresh-home replay.

| Member | Repository | Directory | Content SHA-256 | Collection root |
|---|---|---|---|---|
| `orchestrator` | `fixture.test/playbook` | `skills/orchestrator` | `7079dc673834f8c6181ae7fcd5b11ccd79b3640cccf32d05ae5b49ca9f0af680` | yes |
| `developer` | `fixture.test/playbook` | `skills/developer` | `4b80f5f60c2e12c1eb79f0da90399a97a911cc8e059a76c9f7fce3e2587bd544` | yes |
| `reviewer` | `fixture.test/playbook` | `skills/reviewer` | `535c73bcd70b15046dfc893659532207d1ab7dcbee443ad6f162ef30c650dd36` | yes |
| `writer` | `fixture.test/playbook` | `skills/writer` | `ea19978fff208728f12fb41f431ccd5cba6714c0f6391fd27ae1aa00f003d5eb` | yes |
| `qa` | `fixture.test/roles` | `roles/qa` | `0023a8e45ab2f219fe7828d0dde29efbe5bee155bacf29e00a525d4f7334f2a4` | no; transitive manifest dependency |

The final CLI test parsed eight accumulated audit records for five refreshed members, and five records for the fresh-home replay.

## Validation commands and exit codes

| Command | Exit |
|---|---:|
| `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1 -v` (final run; 55.70s test, 57.139s package) | 0 |
| `go test ./internal/install -run '^(TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock)$' -count=1 -v` | 0 |
| `go build -o /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.hyQ1o918ka/curator ./cmd/curator` | 0 |
| `go vet ./internal/install ./internal/crossconformance` | 0 |
| `golangci-lint run ./internal/install ./internal/crossconformance` | 0 (`0 issues`) |
| `git diff --check` | 0 |
| `gofmt -l internal/install/draftsources.go internal/crossconformance/draftsources_playbook_acceptance_test.go` | 0 (no files listed) |

Two earlier development runs of the acceptance test exited 1: the first tried install before generating the required lock (the documented lifecycle now resolves first); the second used an unsupported glob pattern for a negative include row (the final row uses valid selector `missing` and asserts its refusal). The final acceptance run above exits 0.

Windows was not exercised: the fixture Git URL mapper uses a POSIX shell wrapper and the test skips Windows.

## CHANGELOG entry (for release prep)

- Skillfile schema 2 collections can select every skill from a repository subdirectory with `directory`, `include`, and `exclude` selectors.
- Skill manifest dependencies can select a skill from a repository subdirectory with `dependencies.skills[].directory`.

No `CHANGELOG.md` edit was made. This result records the relevant replay behavior and validation findings for the logbook checklist item; `LOGBOOK.md` was not edited. Product and test changes remain uncommitted in the assigned Story worktree for handoff.

## Producer rerun (2026-09-26)

| Command | Exit | Result |
|---|---:|---|
| `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1 -v` | 0 | Production CLI acceptance passed in 50.19s; all scenario commands and lock/audit assertions passed. |
| `go test ./internal/install -run '^(TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock)$' -count=1 -v` | 0 | Four focused lock replay tests passed in 21.438s. |
| `go build -o /tmp/curator-11burj-20260926 ./cmd/curator` | 0 | CLI builds. |
| `go vet ./internal/install ./internal/crossconformance` | 0 | No vet findings. |
| `golangci-lint run ./internal/install ./internal/crossconformance` | 0 | 0 issues. |
| `gofmt -l internal/install/draftsources.go internal/crossconformance/draftsources_playbook_acceptance_test.go` | 0 | No files listed. |
| `git diff --check` | 0 | No whitespace errors. |
| `go test ./internal/install ./internal/crossconformance -count=1` | 1 | The combined unfiltered package run hit Go's 10 minute test timeout. The timeout dump showed `internal/install.TestDraftBuildFinalRootParentsAreStoreCreated` in transaction journal `Fsync`, and `internal/crossconformance.TestDraftSourcesSemanticCases` waiting in the attested CLI fixture. The new focused acceptance test and the focused source replay tests above passed. |

The worktree remains uncommitted and contains only `internal/install/draftsources.go` plus the new acceptance test `internal/crossconformance/draftsources_playbook_acceptance_test.go`.

## Platform skip-class correction (2026-09-26)

The first handoff's attached revision 1 validation log showed Windows CI rejected the new acceptance test's skip reason (`the local fixture URL mapper is POSIX-only`) as unclassified. The Windows Go tests had exited 0; the subsequent platform-case gate failed on that reason. The test skip now uses the existing `.github/ci/skip-classes.tsv` reason `test transport wrapper is POSIX-only; native admission remains covered on Windows`, which is already classified as allowed `platform-control`.

| Command | Exit | Result |
|---|---:|---|
| `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1 -v` (after skip-class correction) | 0 | Acceptance scenario passed in 96.13s; install, refresh, audit, status, replay, tamper, unreachable and negative rows passed. |
| `CI_PLATFORM_CASES=/tmp/TASK-260924-11burj-empty-ledger.tsv CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh /tmp/TASK-260924-11burj-windows-skip.json /tmp/TASK-260924-11burj-windows-gate-evidence` | 0 | The actual platform-case gate classified the synthetic Windows skip as allowed. This verifies the reason matcher; it is not a Windows execution result. |
| `go build -o /tmp/curator-11burj-20260926 ./cmd/curator` | 0 | CLI builds after the correction. |
| `golangci-lint run ./internal/install ./internal/crossconformance` | 0 | 0 issues after the correction. |
| `gofmt -l internal/install/draftsources.go internal/crossconformance/draftsources_playbook_acceptance_test.go` | 0 | No files listed. |
| `git diff --check` | 0 | No whitespace errors. |

Windows runtime execution remains unverified locally. The corrected skip reason is covered by the CI classifier; the remote Windows lane must provide the platform execution evidence.
