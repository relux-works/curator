# TASK-260910-1sapuy results

## Implementation

Curator now aggregates unreachable trusted registries from snapshot checks and artifact record resolution into one operation-level diagnostic. The CLI renders `registry_unreachable_during_install` on stderr with every unavailable trusted registry URL and every artifact that ended without registry evidence. Under the effective permissive posture it is a warning; under hardened posture it is a blocking error returned before build and commit work. The existing per-query registry warnings remain unchanged. Read-only status plans do not emit the install/update notice.

Transport failures and HTTP 5xx service failures carry a typed unavailability signal. Malformed responses, HTTP 4xx responses, and registry state failures do not enter the outage gate. Tests cover those distinctions.

## Normative rules, rows, and vectors

Spec source: curator-spec commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

| Rule | Spec clause | Story/task ledger rows (before → after) | rc.13 vector | Production entry and result |
|---|---|---:|---|---|
| Permissive posture emits a prominent warning once per install/update, naming trusted registries and artifacts without evidence | `profiles/manager.md` §7.1; `protocol/registry.md` §4 | 0 → 0 | `unreachable-registry-permissive-warns` | Pinned vector driven through the CLI install consumer; direct `curator install` test confirms stderr warning, registry URL, and artifact identity. |
| Hardened posture makes the same notice a blocking error | `profiles/manager.md` §7.1; `protocol/registry.md` §4 | 0 → 0 | `unreachable-registry-hardened-refuses` | Pinned vector driven through the CLI consumer; direct `curator upgrade` test confirms hardened refusal even with advisory registry policy. |

The pinned security-posture suite changed from 11 driven / 6 bound to 13 driven / 4 bound (17 total; 0 known gaps and 0 skipped after the change). The two target vectors are now driven through production entry points.

The baseline ledger at `HEAD` had 67 data rows and zero rows owned by `STORY-260910-2qmrb8` or `TASK-260910-1sapuy`. The current ledger has 67 data rows and zero Story/task-owned rows. Therefore no owned rows were available to remove; total and owned counts are 67 → 67 and 0 → 0. The gap ledger was not edited.

## Mutant evidence

Before the implementation, both target cases were bound, so the suite did not exercise either rule (11 driven / 6 bound). On the implementation:

| Mutant | Killing test | Real exit |
|---|---|---:|
| Remove appending the operation diagnostic to the install result | `go test ./cmd/curator -run '^TestPermissiveRegistryOutageWarnsWithArtifactIdentity$' -count=1` | 1 (expected failure: the required notice was absent) |
| Soften hardened severity to warning | `go test ./cmd/curator -run '^TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy$' -count=1` | 1 (expected failure: upgrade proceeded instead of refusing) |

Both temporary mutations were restored. The final source passed the corresponding focused CLI tests.

## Verification

| Command | Exit | Result |
|---|---:|---|
| `env CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13-vectors.lXxll0 go test ./cmd/curator -run '^TestSecurityPostureVectorsThroughCLI$' -count=1 -v` | 0 | 13 driven, 0 known-gap, 4 bound, 0 skipped, 17 total; both target vectors passed. |
| `go test ./cmd/curator -run '^(TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureWarningIsNotEmittedByEnforcedShim)$' -count=1` | 0 | Final focused CLI tests passed. |
| `go test ./internal/registry -run '^TestRegistryOutageClassificationExcludesOtherFailures$' -count=1` | 0 | Final classification tests passed, including HTTP 503 vs HTTP 400. |
| `go test ./internal/registry -count=1` | 0 | Registry package suite passed before the final two status-classification assertions were added; the final targeted classification test above passed after they were added. |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` | 0 | Manager-owned absence-read guard passed. |
| `make lint` | 0 | golangci-lint reported 0 issues. |
| `go build -o "$TMPDIR/curator-build" ./cmd/curator` | 0 | CLI build passed. |
| `git diff --check` | 0 | No whitespace errors. |
| `go test ./internal/install -count=1` | 1 | Broad package run hit Go's 10-minute timeout in existing `TestDraftDeclaredTagBumpIsNotAMovedTag`, while the transaction journal was syncing via `Engine.saveJournal`. The focused install CLI tests above passed. |

No CHANGELOG or LOGBOOK file was edited. The stateread guard passed. `git status --short` contained only the eight intended Go implementation/test files.

## CHANGELOG entry (for release prep)

Install and update now emit a prominent registry-unreachable gate notice naming trusted registries and artifacts without registry evidence; hardened posture refuses the operation.
