# TASK-261016-hzbo83 — native managed-home policy projection

Developer candidate: narrowed next coherent part under gw-L151B. Role handoff is recorded by the task-board handoff command.

## Scope and implementation

Baseline/checkpoint: 782823d849841ae74efa160243cbac56721aa733. Slice 1 efe0ae9a is an ancestor of freshly fetched origin/main. HEAD remains at the checkpoint and all changes remain uncommitted. Gate snapshot: c82d1cd07ba9d550b6c10c0fe9910ec27e28d592. A separate temporary index verified the current candidate tree equals that snapshot tree (cmp exit 0).

The candidate replaces independent native-store/isolation/link/record derivations with effectiveCredentialPolicy. Resolve repair and verification, status verification and migration consume it. Successful and failed observations are retained only inside the request operation; a fresh public operation observes native state again. Explicit unsupported mappings keep their activation refusals. Native failures retain stale-home reporting; migration retains credential-strategy guidance. A copied identity-migration request clears the caller's projection before adopting recorded sibling isolation.

Changed paths: internal/envprofile/credential_policy.go, managed.go, credential_projection_test.go, migrate.go, identity_migration.go; docs/environment-config.md; README.md. The migration files are necessary consumer seams beyond the leaf's likely-path list: the former removes the independent isolation computation; the latter updates the record call and invalidates cached policy when cloning changed inputs. All source changes stay inside internal/envprofile. Documentation describes the current configuration/application boundary and test command.

**0 of 5 original remainder workstreams fully delivered; first workstream advanced. 2 of 3 scoped AC rows fully driven through public production calls; one internal-seam bound.** See TASK-261016-hzbo83_coverage-map.md for every leaf table row, production calls, named tests, controls and bounds. The brief lacks a catalog surface table; this remains an explicit brief gap.

Remaining requirements: external/adopted-source policy; Muse effective-file literal and successor shared/isolated admission; successor fragment composition per spec 15d9701; readiness/login consumption; qualified Codex file/keychain application. The original hosted-override-drop plant is deferred with fragment/application delivery. No owed requirement is claimed out of contract. Unsupported forced mappings refuse under the leaf's explicit refusal clause. Native auto/shared retains the baseline file-link behavior; this refactor adds no new backend qualification or vendor-state atomicity claim. The provisioning seed is still a separate read.

## Verification run directly by this producer

Local runtime: Go 1.25.5, darwin/arm64; golangci-lint 2.12.2. No prior attached test evidence was used as passing evidence for this candidate. Existing reviewer tests remain unchanged. The full hosted gate succeeded on the exact candidate; every executed job succeeded. Optional self-hosted and candidate-suite jobs were skipped by workflow conditions, not executed. See TASK-261016-hzbo83_hosted-gate.md.

| Command | Exit | Evidence |
|---|---:|---|
| go test ./internal/envprofile ./internal/envregistry ./internal/config ./cmd/curator -run 'Test(Credential\|Codex\|Muse\|.*Isolation\|Migrate\|Resolve.*Credential)' -count=1 | 0 | focused-03.log |
| go test ./internal/envprofile -run 'TestCredentialProjection\|TestCredentialStorePolicy\|TestRC14.*Identity' -count=1 | 0 | projection-green-06.log |
| go test ./internal/envprofile -run 'TestCredential\|TestCodex\|TestMuse\|TestSharedToIsolatedRemovesStaleLink\|TestRC14.*Identity' -coverprofile=.temp/TASK-261016-hzbo83/policy.coverage -count=1 | 0 | coverage-01.log; policy module 63/77 statements, 81.8%; focused package coverage 33.7% is not a whole-package coverage claim |
| go build ./... | 0 | build-02.log (empty output is expected for a successful build) |
| GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-hzbo83/lint-cache" make lint | 0 | lint-03.log; 0 issues |
| git diff --check | 0 | diff-check-06.log |
| python3 .temp/TASK-261016-hzbo83/mutants/run-controls.py | 0 | controls-01.log; all seven child Go tests exit 1 with their named failures |
| python3 .temp/TASK-261016-hzbo83/TASK-261016-hzbo83_controls.py | 0 | controls-03.log; current portable script, all child exits 1 with named failures |
| sh scripts/remote-gate.sh | 0 | remote-gate-02.log; https://github.com/relux-works/curator/actions/runs/38904812207 |

The first regression runs exit 1 on the old code (projection-red-01/02): native policy was re-read during resolution/repair. The first broader focused run exit 1 exposed diagnostic regressions; these were fixed while retaining the original tests. A later red run exposed repeated failed observations; the operation now retains failures as well. Initial make lint exit 2 read stale shared-cache results referring to removed worktrees; the task-local lint cache run passes without changing global cache. The first hosted gate was deliberately cancelled during self-review after the identity-subrequest correction; its supervising script exited 129. That cancelled run is not passing evidence. The replacement gate is bound to the current exact candidate.

## Narrowing mutants

Each gate stays present. A control counts as killed only for the real exit 1 plus its named behavioral test failure. There are no survivors in this bounded set; this does not claim exhaustive mutation coverage.

| Mutant | Narrows the gate to | Named failing test | Child exit / survivor bound |
|---|---|---|---|
| M1-auto-isolated | Reject keyring isolation but admit native auto | TestCredentialProjectionIsolationRefusals/auto | 1, killed; no survivor |
| M2-muse-file-admitted | Admit only Muse forced file before successor activation; retain all other refusals | TestCredentialStorePolicyActivation/muse/file | 1, killed; pins the activation diagnostic, not a successful Muse launch |
| M3-permission-defaults-file | Treat only permission-denied native config as absent/file | TestCredentialStorePolicyNativeInspection/permission | 1, killed; other read classes are outside this mutant |
| M4-keyring-resurrects-file | Withhold ambient/no-file treatment only for native shared keyring | TestCodexKeyringAmbient/keyring | 1, killed; tests file-link/ambient projection, not live Keychain operations |
| M5-shared-second-observation | Bypass reuse only for successful shared projections; retain failed-observation caching, allowing shared re-derivation | TestCredentialProjectionSingleObservation/shared | 1, killed; real Resolve/repair behavioral control |
| M6-error-reobservation | Cache successful observations only, re-read failed observations | TestCredentialProjectionSingleObservation | 1, killed; repairs cannot replace the first diagnostic with later permission errors |
| M7-pi-subrequest-reuses-policy | Clear copied policy for every adapter except Pi | TestCredentialProjectionIdentitySubrequest/isolated | 1, killed; direct internal identity-staging seam bound |

Reproduce from the repository root: python3 <downloaded TASK-261016-hzbo83_controls.py>. It recreates local overlay mappings and executes each Go test as a standalone subprocess with the real exit captured. Production source and tests remain unchanged during plants. The attached plain-text script contains every exact narrowing replacement and reconstructs overlays against this candidate. TASK-261016-hzbo83_candidate.patch preserves all source and test changes. Binary archive attachment was refused by board policy; plain-text resources replace it.

## Logbook

- The baseline repeated native Codex config reads for isolation, links and marker records, so one operation could disagree with itself after a later failed observation. The named regression was red before consolidation.
- Existing refusal diagnostics are part of the contract: the refactor initially surfaced unreadable native config earlier than stale-home verification and changed migration guidance. Both regressions were corrected; original tests were retained.
- Self-review found that identity staging copies a request and changes its isolation. Inheriting an earlier projection can incorrectly apply caller isolation to a recorded shared sibling. The subrequest now clears it; M7 kills the narrowed failure class.
- A stale shared lint cache returned results from removed checkouts. A task-local cache provides clean evidence without mutating another session's cache.
- Binary resource upload refused a gzip archive because public board resources must be text. Switched to a portable plain-text controls script and candidate patch.
- This resource is the task-scoped logbook permitted by the run write boundary; no control-root LOGBOOK.md was edited.
