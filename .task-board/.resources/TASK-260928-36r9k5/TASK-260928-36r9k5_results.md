# TASK-260928-36r9k5 results

## Reapply and merge

Reapplied `refs/campaign/2qmrb8-combined-20260928` (parent `6bd98d49`) onto `origin/main` at `213a53e5c701ef16b961f3398c7f2f0b82c98094`. `git apply --3way` exited 1 because the old carrier overlapped trunk updates; I resolved the conflicts keeping the landed source-signer behavior, E5 no-follow managed writes, and `internal/stateread` reads, while retaining the security-posture and unreachable-registry content. No conflict markers remain.

The accepted `TASK-260927-4pv4au` portion is security-posture revision A: schema-2 posture parsing and status rows, permissive migration warning, hardened defaults/refusals, and the rev3 F1 property that the warning does not leak onto the enforced launch path. The new `TASK-260910-1sapuy` portion adds typed registry-unavailability classification and an operation-level install/update notice: permissive warns on stderr with registries and artifacts lacking evidence; hardened refuses before build/commit work. Read-only status plans do not emit the notice.

Ledgers start from trunk. `conformance-case-counts.tsv` adds only `security-posture/vectors = 17`. Per the gatefix resource, `conformance-gaps.tsv` removes exactly the three now-driven security-posture schema-case rows; every other trunk row remains unchanged. 1sapuy owned no ledger rows. No `CHANGELOG.md` or `LOGBOOK.md` edits were made.

## Scope variance for review

The original carrier delta names 31 paths. The final diff against `origin/main` changes 34 paths: 29 paths from the carrier remain changed, while `internal/config/environments_conformance_test.go` and `internal/config/environments_test.go` have no net diff because their carrier hunks assert that `source_signers` is unsupported, which is stale after E1 landed. Five additional test/golden paths update trunk tests that failed the previous hosted validation: they now assert one permissive warning per invocation and pin the warning in the delta transcripts. These are the exact files named in the diff: `cmd/curator/global_adopt_test.go`, `cmd/curator/global_lock_publication_test.go`, `cmd/curator/profile_delta_confirmation_test.go`, `cmd/curator/testdata/profile-update-mcp-delta.golden`, and `cmd/curator/testdata/profile-update-system-delta.golden`. No unrelated product paths were added. This differs from the requested 31-path count and is called out for review.

## Verification

Commands were run as standalone processes. Exit codes are the process exit codes.

| Command | Exit | Result |
|---|---:|---|
| `go test ./internal/config -run 'Posture|Security|Hardened|Permissive'` | 0 | Focused posture/config tests passed. |
| `go test ./cmd/curator -run 'SecurityPosture|StatusJSON|Enforced'` | 0 | Posture, status JSON, and enforced-path checks passed, including the no-warning enforced launch regression. |
| `go test ./internal/install -run 'Posture|Unreachable|Registry'` | 0 | Focused install cases passed. |
| `go test ./internal/scriptworker -run Enforced` | 0 | Enforced launcher tests passed. |
| `go test ./internal/envprofile -run Guarded` | 0 | Manager-state absence-read guard passed. |
| `go test ./internal/registry -run 'TestRegistryOutageClassificationExcludesOtherFailures|TestHTTPFailuresUseOfflineCacheAndRejectSnapshots'` | 0 | Outage classification and HTTP failure tests passed. |
| `go test ./internal/config` | 0 | Full package passed (cached). |
| `go test ./internal/config -count=1` | 0 | Full package passed uncached after the ledger update. |
| `go test ./cmd/curator -run 'GlobalAdopt|GlobalAdd|GlobalInstall|ProfileUpdate|SecurityPosture|Golden'` (before gatefix) | 1 | Four existing tests still expected empty stderr or old goldens; the required permissive warning was present. |
| Same curator gatefix command (after updates) | 0 | Passed uncached in 223.827s; a subsequent invocation also exited 0 (cached). Tests assert one warning per permissive invocation and goldens include it. |
| `go build -o "$TMPDIR/curator-task-build" ./cmd/curator` | 0 | CLI build passed after test/golden updates. |
| `golangci-lint run` | 0 | 0 issues after updates. |
| `git diff --check HEAD` | 0 | No whitespace errors. |

The previous hosted revision-1 validation was red on the three posture gap rows and four stderr/golden expectations. Those were addressed locally. The hosted gate for this handoff runs after the turn; its result is not yet available and is not claimed green here.

## CHANGELOG entry (for release prep)

Install and update now emit a prominent registry-unreachable gate notice naming trusted registries and artifacts without registry evidence; hardened posture refuses the operation.
