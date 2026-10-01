# TASK-260917-2tx81l — manager-content-hash-v2


## Revision 4 (refresh)

This is a base refresh of accepted revision 3, with no new product change. The revision 3 verdict was read and says ACCEPTED, with no blocking findings; the generic review-round rejection wording does not describe that verdict. Its original acceptance was released because trunk moved on the count table.

`task-board worktree refresh-candidate TASK-260917-2tx81l` returned exit **0**, `refresh_advanced`, onto `bd126a9acdc51b6061917ba8c4d7d26a7abafd41`. A follow-up invocation returned exit **0**, `refresh_already_current`. No checkpoint replay conflict occurred, so no `--replay-resolutions` packet was needed. No commit, branch switch, or hand-committed replay was made.

### Identity and resolution

The incoming candidate matched accepted tree `92a072c4e1b03678ea64a2801069ea471628e8e4` on **45/45** owned paths before refresh. All four overlapping paths merged without conflict. The refresh command preserves the supplied working tree, so all 23 trunk-only non-board paths were also explicitly carried forward and byte-checked against trunk. No board files were edited manually.

After convergence, **41/45** owned paths remain byte-identical to revision 3, including every test file and all five added files. The other four differ only by incoming trunk changes:

- Counts: keep the candidate's 131→121 correction under the candidate digest and trunk's four acquisition rows under the rc.13 digest.
- Config: retain trunk's typed `ErrConfigNotFound` and wrapped absence error beside the existing v3 reader.
- Managed environment: retain trunk's `--repair` hint for stale Resolve failures beside the existing version checks.
- Install targets: retain trunk's artifact basename selection beside the existing versioned hash staging.

No assertion or product behavior was changed by this refresh. The source identity check was repeated after mutation testing, and the index is empty. The complete non-board diff against accepted revision 3 is **27 paths, 1,880 insertions, 33 deletions**, entirely the incoming trunk delta. Per-file statistics follow; untracked candidate additions were independently blob-checked so Git's tracked-only diff cannot misclassify them as deletions.

| Path | Added | Removed | Source |
|---|---:|---:|---|
| `.github/ci/platform-cases.tsv` | 1 | 0 | trunk-only |
| `README.md` | 1 | 1 | trunk-only |
| `cmd/curator/env.go` | 4 | 0 | trunk-only |
| `cmd/curator/firstrun_ux_test.go` | 285 | 0 | trunk-only |
| `cmd/curator/main.go` | 68 | 1 | trunk-only |
| `cmd/curator/native_blackbox_test.go` | 254 | 0 | trunk-only |
| `cmd/curator/profile.go` | 4 | 0 | trunk-only |
| `cmd/curator/umbrella.go` | 1 | 0 | trunk-only |
| `docs/cli.md` | 13 | 4 | trunk-only |
| `docs/external-build-repositories.md` | 75 | 0 | trunk-only |
| `internal/buildrepo/acquisition_conformance_test.go` | 775 | 0 | trunk-only |
| `internal/buildrepo/admission.go` | 6 | 3 | trunk-only |
| `internal/buildrepo/admission_test.go` | 68 | 9 | trunk-only |
| `internal/buildrepo/adopt_test.go` | 2 | 2 | trunk-only |
| `internal/buildrepo/protected.go` | 29 | 3 | trunk-only |
| `internal/buildrepo/protected_artifact_test.go` | 67 | 0 | trunk-only |
| `internal/buildrepo/protection_windows_test.go` | 2 | 2 | trunk-only |
| `internal/buildrepo/testdata/acquisitiongitshim/main.go` | 77 | 0 | trunk-only |
| `internal/crossconformance/draftsources_semantic_external_test.go` | 3 | 1 | trunk-only |
| `internal/envprofile/named_absence_boundary_test.go` | 56 | 0 | trunk-only |
| `internal/envprofile/status.go` | 9 | 2 | trunk-only |
| `internal/install/external.go` | 2 | 2 | trunk-only |
| `internal/pathboundary/named_absence_test.go` | 64 | 0 | trunk-only |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 | merged |
| `internal/config/config.go` | 5 | 1 | merged |
| `internal/envprofile/managed.go` | 4 | 1 | merged |
| `internal/install/targets.go` | 1 | 1 | merged |

### Exact count recomputation

Ran `python3 /tmp/TASK-260917-2tx81l-refresh4/counts.py` directly: exit **0**. It reads the two suite manifests, schema indexes, sibling skillfile corpus, vector arrays and external-repository fixture arrays, and applies the production consumers' operation/phase splits and CRC32 semantic batching. It asserts every declared pin is resolved and equal; no counts were guessed.

| Suite | Manifest digest | Exact declared pins | Sum of cases |
|---|---|---:|---:|
| rc.13, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` | `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca` | 92/92 | 1,731 |
| b1a2efb, commit `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28` | `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60` | 97/97 | 1,722 |

The rc.13 acquisition arrays recompute to 12 cases, 17 clean-environment entries, 55 fetch arguments and 11 forbidden features: 95 additional declared cases. The historical rc.13 88/1,636 total above is superseded for revision 4 by 92/1,731. Candidate counts remain unchanged. Both sibling skillfile schema indexes still contain exactly 121 cases. This exactness claim covers the declared pins, not every possible unpinned suite family.

### Verification rerun here

All validation commands are standalone processes with output redirected directly to individual log files, preserving their real exit status; no gate was piped through tee. Commands and log SHA-256 digests are in the attached verification JSON.

| Command / scope | rc.13 exit | Candidate exit |
|---|---:|---:|
| `go test ./internal/conformancecoverage -count=1` | 0 | 0 |
| Content-hash v2, carrier schema and mismatch regression selectors | 0 | 0 |
| `go test ./internal/crossconformance -count=1 -run '^(TestDraftSourcesSchemaCases|TestSkillfileSourcesCorpusCounts)$' -v` | 0 | 0 |
| Full hashing, registry, marker, contextlock, envmarker, config and conformancecoverage packages, `-count=1 -v`, after all trunk paths were carried | 0 | 0 |
| `bash .github/ci/ledger-consistency.sh <evidence-dir>` across linux/darwin/windows compile inventories | 0 | 0 |

The current candidate log records **80/80** owned v2 cases as driven across all nine families, with zero known gaps, bounds or skips for these rows. The 80 core rows and seven absent deferred rows remain removed; the gap file is byte-identical to accepted revision 3.

Additional current-tree checks: `go build ./...` exit **0**; `go vet ./...` exit **0**; `golangci-lint run` exit **0**, **0 issues**; `git diff --check -- . ':!.task-board'` exit **0**. The newly carried acquisition production-entry conformance test and `TestResolveStaleUnprovisioned` each ran under rc.13 with `-count=1 -timeout=4m -v` and exit **0**. No process was left running.

### Regression and narrowing evidence

No new regression test was added because the binding refresh brief forbids product changes and all existing tests are byte-identical to revision 3. The existing named regression `TestWritePreservesContentHashInRC13Mode` was rerun, along with its narrowing mutant: preserve caller-supplied hashes only for skill schema 6 while recomputing schemas 7 and 8. The authoritative compiled-marker control still passes; the named regression fails in both schemas 7 and 8. Its unmutated coverage remains **3/3** schema bands, with the narrowed regression detected in **2/3**.

All four mutants were applied through canonical-path Go overlays outside the repository; source files were never changed.

| Mutant | Separate legacy-control exit | Negative exit | Unmodified regression exit |
|---|---:|---:|---:|
| M1: remove v2 length writes; adjacent-record collision test | 0 | 1 | 0 |
| M2: omit both registry version comparisons; mismatch and ResolveVersioned admission tests | 0 | 1 | 0 |
| M3: admit a v1 marker for a v2 expectation; Current regression | 0 | 1 | 0 |
| F1 narrowing: preserve supplied hashes only for schema 6; rc.13 Write regression | 0 | 1 | 0 |

The nonzero mutant commands are expected failures: each admits the invalid behavior and its negative test detects it. The unmodified regressions were rerun after all overlay commands with exit **0**. Coverage is **3/3** required mutants plus the prior F1 narrowing regression. Legacy survivor controls describe current mutated code, not a new claim that these v2-only sites existed in the original base.

### Evidence carried from revision 3 and limits

Accepted from the attached revision 3 verdict and evidence, not rerun here: the independent base-test overlay (21 modified files, 76/76 individually passing environment/install selectors), the four reader-version exceptions, and the revision 3 hosted platform gate. Every existing test file remains identical to that accepted candidate. This refresh makes no claim that the full unreleased b1a2efb suite is green; the previously reproduced snapshot-acquisition vector/driver mismatch in the rev3 verdict remains outside this task's 80 core rows. The hosted gate for revision 4 belongs to the runner after handoff.

No CHANGELOG or LOGBOOK was edited. The release-prep migration text remains under `## CHANGELOG entry (for release prep)` above. New task-scoped verification and log artifacts are attached before the developer handoff.
