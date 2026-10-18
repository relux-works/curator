# Round 2 validation evidence

Candidate: `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b` (SSH signature verified `G`). Story HEAD remains `5d8e6b07f7f485557d9bbb065c9692348ca065c2`; all task changes remain uncommitted. Hosted workflow: https://github.com/relux-works/curator/actions/runs/39094506341.

## Execution identity and evidence reuse

Local host: darwin/arm64, Go 1.25.5. Protocol checkout: `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, complete checkout with byte-exact conformance fixtures and sibling skillfile corpus. Suite plan: 85 served, 0 deferred, 0 excluded; `CI_REQUIRE_FULL_ROOT=1` plan exit 0.

The first signed snapshot `ab277c78bbd3b8d60c92756603c2f4e35385a918` differs in `cmd/curator/global_lock_publication_test.go` and the native mutation harness. The second snapshot `646fd6a1cdf1fb5b988ff0e51f947792c21c9ebf` differs only in the mutation harness. All production source is byte-identical across these snapshots. Unrelated tests and repository configuration remain identical. Per standing order 10, green package evidence for those unchanged identities is reused. The complete CLI suite is run with the updated test source; native controls are rerun to include the freshly compiled CLI flag mutant. This report distinguishes reuse from replay.

Commands are direct processes, with stdout/stderr captured without tee pipelines. Long local verification is split into bounded package/test-name groups, with each command's real exit stored in the attached command ledger. No background process is left after handoff.

## Setup and expected-red history

| Command / attempt | Real exit | Meaning and recovery |
|---|---:|---|
| Initial broad local group 2 | 1 | Incomplete `git archive` fixture setup: substituted bytes and missing sibling corpus. The two affected packages were rerun against the exact checkout, exit 0; the other 18 packages' inputs were unchanged. |
| Initial pnpm integration cases | 0, cases skipped | Not acceptance evidence. Task-local pnpm 10.33.0 installed; all three mandatory real-pnpm cases rerun, exit 0, no skips. |
| First CLI group 1 | 1 | Ambient untrusted optional provider made project status non-current; unchanged-checkpoint control also exited 1. CLI child PATH now excludes that provider directory. No user installation was altered. |
| Baseline CLI control and output-only overlay | 1 / 1 | Confirmed the same status failure before this change and identified the provider row; not passing evidence. |
| First hosted race gate | 1 | Four reviewer tests still targeted the retired staging path. Original test names and refusal/recovery assertions retained; destinations updated to native paths and marker assertion strengthened. First run cancelled after failure evidence capture; never treated as passing. |
| Narrowing mutant children | 1 each (12) | Intended red: named behavioral failures prove each narrowed exception is rejected by the real entry-point suite. Parent harness must exit 0. |

Task-local tool install: `npm install --global --prefix .temp/BUG-261018-uoisfe/pnpm-prefix pnpm@10.33.0` exited 0; `pnpm --version` exited 0 and returned 10.33.0. No tool was installed globally outside this task directory. Optional Yarn and platform skips are judged by the unchanged platform-case gate, not silently ignored.

Production links and file paths never target the real HOME in added tests. Fixture HOME/USERPROFILE and all native home variables point to temporary directories.

Additional diagnostic attempts: a native-suite replay ran concurrently with a long CLI group and exited 1 after its 4-minute parent timeout while the nested CLI mutant waited for the shared host-GOROOT test lock. It is not accepted evidence. The subsequent standalone native replay exited 0, killing all twelve controls. A lint replay with the shared cache exited 1 with positions from unrelated deleted worktrees; an isolated `GOLANGCI_LINT_CACHE=.temp/BUG-261018-uoisfe/lint-cache` replay exited 0 with zero issues. The shared cache was not cleared.

The Story checkpoint is two commits behind local main observed during validation (`402fdbc9`); the task explicitly forbids rebasing/merging the Story branch. Evidence qualifies the named checkpoint plus this candidate, not the newer trunk. The newer changes concern snapshot publication and CI routing; they are not absorbed into this leaf.

Fixture identity check: comparing the initial export with the exact checkout found only `fixtures/byte-exact/subst.txt` changed inside conformance/v1; all other files and names were identical. Missing sibling skillfile artifacts were consumed by the two replayed packages. Unaffected package evidence is retained explicitly, rather than reported as a new replay.

The first local gate-selftest exited 1 because its Git ancestry fixtures inherited signing configuration without the required agent Git environment; the commits were not created, so the ancestry controls were not accepted. After sourcing the configured agent Git helper, the same standalone script exited 0: 267 passed, 0 failed, 2 explicitly reported skips. The four previously committed global lock publication/conflict/recovery tests passed under their original names with native destinations.

Hosted queue observation: the official GitHub status endpoint reported no unresolved incident; no infrastructure failure was inferred. macOS jobs later began executing. No local mirror or forged check status was used.

CLI group 8 initially exited 1 because inherited command-level `GIT_CONFIG_*` settings overrode the hostile-config test's temporary signing key. Its four global publication/recovery tests passed. The group is replayed with command-level Git overrides removed from the child test environment, matching the hosted fixture configuration; no host Git config is edited. Earlier overwritten failing logs are not accepted as passing evidence.

The concurrent native-control timeout exposed a real harness scheduling flaw: a flag-only CLI child tried to acquire the whole-package GOROOT lock held by the full CLI suite. The harness now uses the CLI's existing `CURATOR_TEST_CLI_HELPER=1` test lane only for those flag cases; these do not inspect GOROOT. Actual CLI routing and behavioral assertions remain enabled. The corrected controls are replayed while a full CLI group is running. The second hosted run was cancelled after this finding and is not a full passing verdict; a new signed snapshot runs the complete hosted gate.

Flag-boundary review: the named CLI tests prove the five specified operations and enumerated profile/env/status refusals. They do not prove generic argument rejection by legacy top-level/global wrappers that discard trailing arguments. This pre-existing CLI grammar is outside the leaf's switch/status/repair scope and is explicitly bounded in the gaps artifact; no global operation gains takeover authority.

## Local acceptance results

All 85 served packages have passing package evidence; none is deferred or excluded. The four large packages cover all 1303 inventoried top-level names (381 envprofile, 363 install, 132 godriver, 427 CLI), across 46 bounded groups. The full merged stream contains 3490 unique top-level test names across the repository. All 36 recorded skips were accepted by the unchanged platform-case gate, which exited 0. Its per-case skip reasons and ledger classifications are in the attached text metadata and hash-bound local archive.

| Standalone check | Real exit | Result |
|---|---:|---|
| Suite plan, fully serving exact root | 0 | 85 served, 0 deferred, 0 excluded |
| Fresh native suite and narrowing controls | 0 | 12 controls killed; each child really exited 1 with its named behavioral failure |
| Platform-case gate on merged stream | 0 | All required cases observed; 36 skips classified |
| Gate self-test | 0 | 267 passed, 0 failed, 2 explicit skips |
| golangci-lint with task-local cache | 0 | 0 issues |
| go vet ./... | 0 | No diagnostics |
| go build ./cmd/curator | 0 | Task-local binary produced |
| gofmt inspection / git diff --check | 0 / 0 | No unformatted files or whitespace errors |
| Stream assembly and inventory check | 0 | No missing package or top-level name |

Raw and selected test streams, the stream-assembly script, real command-exit ledger, package lists, run masks, tool identities, and gate reports are preserved in the attached plaintext event stream, text metadata bundle, and hash-bound local archive (reference attached). The failed initial group-2 command is preserved with its actual exit 1; its two repaired package streams are explicit. A historical log that was overwritten is not used as green evidence. All previously committed reviewer tests are present and green under their original names.

## Windows fixture correction and fourth signed snapshot

The third complete hosted run, `39091127316`, concluded failure: 19 jobs succeeded, Windows Test failed (gate exit 1), and two conditional jobs skipped. It is not passing evidence. The attempted cancellation returned exit 1 because the run had already completed; it did not cancel any job.

Actual Windows failures were the new native CLI operations, takeover/foreign, unnamed-link preservation, unrecorded-ledger preservation, and the foreign-consent mutant harness. Windows creates the relative symlink with native separators; assertions compared Readlink with a POSIX-spelled input literal. The corrected tests capture the original Readlink text before the operation and compare refusal/backup/preservation with those exact bytes, without normalizing the outcome. The mutant matches its retired-manager prefix after filepath.ToSlash, so its exception applies on Windows too. No production source, shared test helper, configuration or protocol fixture changed. Only the three new native test files differ from the third signed snapshot.

| Historical mutant | What it narrowed on Windows | Named intended failure | Survival bound |
|---|---|---|---|
| dangling-foreign-consent at eb218081 | Nothing: its slash-only prefix did not match Windows Readlink text | None; backup-literal mismatch was unrelated | Survivor/invalid control on Windows. The prior kill was bounded to POSIX link spelling. The fourth snapshot makes the matcher portable; current local replay kills it by TestNativeSkillsTakeover/foreign. Fresh hosted Windows validation is required. |

The fourth snapshot is `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b`, signature G, tree `97af0e0f6eb16747393683be175b7e297d3b60bf`. Its full hosted run is `39094506341`. All unchanged production and per-test input identities reuse the previously green local evidence. Every new native test and all twelve controls were rerun directly: native-candidate-05.json and native-cli-05.json each exited 0. Lint-07, vet-03, build-05, diff-check-06 and gofmt-05 each exited 0; no unformatted file was listed. The merged stream replaces all envprofile and CLI events with the 28 fresh full-package groups; every group exited 0. The other 83 packages retain unchanged source/test/configuration/environment evidence. The platform-case gate on this refreshed stream exited 0.

Additional replay rationale: CLI integration fixtures use the compiled test binary as an executable provenance input. Its identity changes when test source changes even though production source does not. The two affected packages (envprofile and CLI) are therefore refreshed in full, in bounded groups, on the fourth candidate. The other 83 packages keep their unchanged source/test/configuration/environment evidence. This is a recorded identity reason, not an unconditional full-suite replay. The standalone native/CLI reruns remain valid additional evidence.

Archive preparation: the first sanitized-stream gate exited 2 because it was invoked before the sanitizer had produced its input. This is a failed setup check, not a product verdict. Its diagnostic is retained; the gate is rerun as a separate process after successful stream creation. The sanitizer preserves parsed JSON Action/Test fields, and a decoded POSIX/Windows-path redaction control exited 0.

Current Windows proof: artifact download exited 0. Its raw Go stream records 48 native pass events, zero native fail events, and all 12 narrowing control passes, each reporting the intended child exit 1 and named failure. windows-native-current.json preserves these names and messages with the raw artifact SHA-256. This replaces the historical Windows survivor bound only for the current portable matcher, not retroactively.

## Hosted acceptance result

The full hosted gate concluded **success** on the exact signed candidate b32823ac920d8b0485c4a7a7111b48fff8e3ca2b. [Run 39094506341](https://github.com/relux-works/curator/actions/runs/39094506341) has 20 successful jobs, including all Test/Race lanes and all three platform gate self-tests. Two jobs were conditionally skipped: main-only rose-air and optional candidate-spec workflow-dispatch; neither is claimed as executed. gh run watch 39094506341 --exit-status exited 0. No local mirror or synthetic status was used.

All 28 refreshed envprofile/CLI groups exited 0. Stream inventory/merge, platform gate on the raw stream and platform gate on the sanitized archive stream each exited 0. The refreshed merged stream retains 85 served packages, 3490 unique top-level test names including explicit skips, and 36 classified skips. The other 83 packages reuse the unchanged source/test/configuration/environment identities described above. All four original reviewer test names are present and green in the fresh CLI stream. Current Windows evidence contains all 12 named expected-red control failures (children exit 1; parent suites exit 0).

Coverage: **12 of 12 attack-surface rows driven; 5 of 5 AC rows driven**, with production call sites, narrowing mutants and the explicit legacy parser bound in the coverage/gaps resources. The source integrity check exits 0: worktree tree equals the signed candidate tree, signature G, remote gate head equals candidate, and Story HEAD remains at its recorded checkpoint. The orchestrator owns review and trunk integration.

Archive packaging and audit each exited 0. The archive contains 191 files (2,086,483 bytes); all 94 JSON files/streams parse, and the accepted merged stream is byte-identical to the sanitized stream already checked by the platform gate. No concrete private home path or email was found. Archive SHA-256: ca67a410e3e171a6180fe6d03f2d7ef54c1042c77004fb19295f52fb8f0c2209. The separately attached manifest and audit bind these bytes.

Board publication: binary resource upload returned exit 1 with resource_binary_refused (public board permits text only). No archive was uploaded. The supported recovery publishes the full accepted event stream as validation-stream.jsonl (12,680,379 bytes), 107 metadata/evidence files in validation-metadata.json, and a text archive reference containing the ignored relative path, size and SHA-256. Each text attachment succeeded with exit 0. The full local raw archive remains present outside Git; its manifest and audit are attached separately. No binary encoding bypass was used.
