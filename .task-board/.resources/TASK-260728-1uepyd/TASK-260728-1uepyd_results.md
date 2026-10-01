# TASK-260728-1uepyd results — rework 1

## Outcome and revision-2 finding F1

The consumer now checks what `acquireNetworkFormat` hands to Git at the fetch call site (`internal/buildrepo/admission.go`, `fetchArgs` / `fetchEnv` into `runGitCapture`). The portable shim records the unmodified argv and received environment before substituting the fixture's local transport. Every fetch-bearing vector case compares all 55 argv entries, including the trusted executable, to the resolved published vector. Repo and hooks must occupy `repo.git` and `empty-hooks` under one absolute operation-private root. Expected environment values come from the published vector, that root, the pinned tool and the selected transport, independently of the production environment builder. The comparison rejects unexpected, missing and duplicate variables.

Named regressions: `TestExternalRepositoryAcquisitionConformance/cases/sha1-untagged-https` catches both F1 bypasses; `TestExternalRepositoryAcquisitionConformance/cases/sha256-tagged-ssh` catches the transport-narrowed bypass and a missing transport variable. These invoke `AcquireNetwork`, with failure vectors also invoking `RunPipeline` and asserting that only exact-source-acquisition was reached. No production change was added by this rework. The prior revision's pure request validation reorder and local-admission fixtures remain intact; all temporary production mutants were restored from a saved copy and verified with `cmp`.

## Measured coverage

Selected rc.13 suite: `/tmp/spec13/conformance/v1`; SPEC_PIN `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`; manifest SHA-256 `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`. The manifest is asserted by the consumer. Exact count rows remain in `.github/ci/conformance-case-counts.tsv`.

| Family | Driven | Known gap | Bound | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| external-repository/acquisition/cases | 12 | 0 | 0 | 0 | 12 |
| external-repository/acquisition/common-fetch-argv | 55 | 0 | 0 | 0 | 55 |
| external-repository/acquisition/clean-environment | 17 | 0 | 0 | 0 | 17 |
| external-repository/acquisition/forbidden-fetch-features | 11 | 0 | 0 | 0 | 11 |
| external-repository/local-config-and-refs | 15 | 0 | 0 | 0 | 15 |
| external-repository/pack-index | 8 | 0 | 0 | 0 | 8 |

Acquisition cases are driven for acquisition semantics. Each of the 11 observed fetches additionally verifies the full vector argv and common environment plus transport and platform additions. The malformed-ref case proves no Git invocation occurred. The three older bounds are exercised through `AdmitLocal`: gitfile refusal, linked administration refusal and pack-without-index refusal. The link fixture may report an explicit host-permission bound on other hosts; it was driven here.

## Residuals

- R-a fixed: the helper-selected-transport row calls the production `ParseSource` with an `ext::` URL and requires refusal. Its fetch checks also require the default-deny protocol config and the explicit selected protocol. The old constant substring check was removed.
- R-b explicitly bounded, owner TASK-260728-1uepyd: **0/7 successful vector cases observe downstream pipeline audit/cache/compiler order**. These raw-object acquisition fixtures contain no `curator.build.json` descriptor, so an unmodified snapshot cannot reach the pipeline's audit stage. The previous assertion on the vector's boolean data was removed. The seven cases prove acquisition, full argv and environment only. Ordering remains unknown for these vector fixtures; existing pipeline tests are separate evidence, not substitutes for those rows. This bound is stated independently of the acquisition coverage table and in the test beside the success branch. Driving descriptor-bearing success fixtures belongs to the pipeline/lifecycle consumer scope.
- R-c unchanged as requested: the symlink-permission bound is retained. No Windows skip or fixture gap was added.

## Mutation checks (real expected-red exits)

All commands ran directly under zsh, without pipes or tee, with `CURATOR_CONFORMANCE_ROOT=/tmp/spec13/conformance/v1`, package `./internal/buildrepo`, and `-count=1`. Each mutant command returned **exit 1**, an expected failure proving the named gate rejects the altered behavior. These are failing commands, not passing gates.

| Mutant | Temporary production change | Exact `-run` mask | Exit | Kill |
| --- | --- | --- | ---: | --- |
| M1 | strictFetchArgs allows extra --show-forced-updates | ^TestExternalRepositoryAcquisitionConformance/common-fetch-argv$ | 1 | arg-52 differs from vector |
| M2 | cleanGitEnvironment leaks GIT_TRACE=1 | ^TestExternalRepositoryAcquisitionConformance/clean-environment$ | 1 | unexpected environment entry |
| M3 | strictFetchArgs accepts --depth=1 | ^TestExternalRepositoryAcquisitionConformance/forbidden-fetch-features$ | 1 | forbidden depth accepted |
| M4 | real fetch call site prepends -c fetch.prune=true | ^TestExternalRepositoryAcquisitionConformance/cases$ | 1 | sha1-untagged-https full production argv mismatch |
| M5 | real acquisition environment leaks GIT_SSL_NO_VERIFY=1 | ^TestExternalRepositoryAcquisitionConformance/cases$ | 1 | sha1-untagged-https full production environment mismatch |
| M5-ssh-only | same leak only for SSH requests; HTTPS path remains clean | ^TestExternalRepositoryAcquisitionConformance/cases$ | 1 | sha256-tagged-ssh full production environment mismatch |
| M7-ssh-missing | real SSH acquisition removes only GIT_SSH_VARIANT=ssh | ^TestExternalRepositoryAcquisitionConformance/cases$ | 1 | sha256-tagged-ssh detects missing required variable |

Command shape: `CURATOR_CONFORMANCE_ROOT=/tmp/spec13/conformance/v1 go test ./internal/buildrepo -run '<mask above>' -count=1`. Raw outputs are in the task-scoped rework evidence archive; production restoration was verified after every mutant.

## Validation run by this developer

| Command | Exit | Evidence / scope |
| --- | ---: | --- |
| CURATOR_CONFORMANCE_ROOT=/tmp/spec13/conformance/v1 go test ./internal/buildrepo -count=1 -v | 0 | 333.123 s; package-test.log; all package tests, parity, adversarial fixtures and ordering-refusal tests |
| CURATOR_CONFORMANCE_ROOT=/tmp/spec13/conformance/v1 go test ./internal/buildrepo -run ExternalRepositoryAcquisition -count=1 -v | 0 | 23.706 s; acquisition-test.log; rerun against the current tree after comment-only shim lint fixes |
| GOOS=windows go vet ./internal/buildrepo | 0 | Required Windows cross-target check |
| GOOS=windows go vet ./internal/buildrepo ./internal/buildrepo/testdata/acquisitiongitshim | 0 | Rechecked including the current standalone shim |
| go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 | 0 | 15.124 s; manager-state read guard |
| go build ./internal/buildrepo | 0 | Rerun after restoration and after shim comment edits |
| golangci-lint run ./internal/buildrepo | 0 | Initial package-only lint |
| golangci-lint run ./internal/buildrepo ./internal/buildrepo/testdata/acquisitiongitshim | 1, 1, 0 | First found missing package comment and G204; next found G702 at the same fixture-owned Git launch; current rerun reports 0 issues after a package comment and narrowly justified G204/G702 annotations |
| git diff --check | 0 | Current tree whitespace check |
| gofmt -l internal/buildrepo/acquisition_conformance_test.go internal/buildrepo/testdata/acquisitiongitshim/main.go | 0 | No formatting output |

The full package run and mutants preceded only shim comment/annotation edits. No executable statement changed afterward. Separate SHA-256 input manifests identify the full-package/mutant inputs and the current acquisition rerun. I reran all commands above myself; no prior attached green evidence was substituted. The repository-wide landing suite is reserved for the configured handoff path and was not run manually.

Host: darwin/amd64. Native Windows and Linux execution of this revision was not run locally; cross-target vet is not a native Windows test result. No platform skip was added. The package log includes pre-existing platform-specific skips outside the acquisition vector; acquisition coverage reports zero skipped rows.

## Hygiene and lifecycle

Changes remain uncommitted in the assigned Story worktree. This rework changes the acquisition consumer and test shim only, beyond the existing revision-2 candidate. New state reads still go through internal/stateread. No CHANGELOG or LOGBOOK was edited, and no Windows-reserved file name was introduced. Findings and decisions are recorded here and in the task notes, following the binding no-logbook instruction. The original results resource is updated and a new task-scoped raw evidence archive is attached before developer handoff.

## CHANGELOG entry (for release prep)

Verify full production external-repository fetch arguments and clean environments against the pinned rc.13 conformance vector, including unexpected and missing transport variables, and exercise helper-selected transport refusal.
