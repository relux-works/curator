# Round 2 implementation logbook

The round-2 decision excludes Muse; the three single-home adapters now use registry-native skills directories. Effective-current status/repair, ledger exclusion/reporting, old staging convergence, takeover backups and no-follow checks are implemented. An unchanged-lock update previously skipped materialization; it now reconciles the effective-current native surfaces. All new behavior tests use temporary homes.

Twelve surface rows have named production-entry tests; twelve narrowing controls were killed. The coverage map and behavioral control output are attached separately.

Validation exposed two setup issues: an initial conformance export changed export-subst bytes and omitted the sibling skillfile corpus; the affected packages were rerun against the exact complete pinned checkout. Mandatory pnpm tests initially skipped; task-local pnpm 10.33.0 was installed with `npm install --global --prefix .temp/BUG-261018-uoisfe/pnpm-prefix pnpm@10.33.0`, and all three real-pnpm tests then passed. No global tool installation was changed.

A project-status test failed on the unchanged checkpoint because an ambient optional provider was present but untrusted. A baseline run and a bounded output-only overlay confirmed this. CLI validation now excludes that provider directory from the child PATH; no operator files were changed. This was an environment finding, not a product regression.

The first hosted race gate returned exit 1 because four existing reviewer tests still targeted retired staging. Their original names and rejection/recovery assertions are retained; destination/backup paths now address native skills, and the marker-path assertion is stronger. The failed run was not treated as green and was cancelled after evidence capture.

Current signed gate snapshot: `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b`. Hosted run: https://github.com/relux-works/curator/actions/runs/39094506341. Full validation is still running. Story HEAD remains its checkpoint and all deliverable changes remain uncommitted.

## Harness scheduling correction

Concurrent replay exposed the whole-package CLI GOROOT lock blocking a flag-only mutant child. The existing CURATOR_TEST_CLI_HELPER lane now skips only that unrelated lock for the child, preserving actual run() calls and assertions. A fresh native replay ran alongside a full CLI group and exited 0; all twelve controls exited 1 with their named behavioral failures. The prior hosted snapshot was cancelled and not accepted as a full pass. The replacement signed candidate above carries this correction.

## Flag scope bound

The CLI tests cover all five named operations and enumerated profile/env/status refusals. They do not claim generic argument rejection by legacy top-level/global wrappers that discard trailing arguments. That pre-existing grammar is outside this leaf's switch/status/repair scope; no legacy-global operation receives takeover authority. The same bound appears in the gaps and coverage map.

## Local acceptance

All 85 served packages have green package evidence, with every inventoried top-level name observed. The assembled stream and platform-case gate pass (exit 0), recording all 36 allowed skips. Standalone lint, vet, build, gofmt inspection and diff-check pass. The previous hosted run ultimately had 19 successful jobs and a Windows Test failure. Its overall result is failure. No queued/running result is counted as green.

## Windows fixture and control correction

The Windows gate returned 1 because new tests compared Readlink against POSIX input literals. Windows normalizes separators during link creation. The tests now capture actual original link text and compare it byte-for-byte after refusal, backup and unnamed/ledger preservation; results are never normalized. The foreign-consent mutant now matches after filepath.ToSlash so the narrowing exception actually applies on Windows. Its old Windows result is recorded as a survivor/invalid control, not a kill. Only the three new native test files changed, with no production or shared-helper change. Fresh native and CLI native runs exit 0; all twelve controls are killed locally. Lint, vet, build, formatting and the updated platform gate exit 0. The fourth full hosted gate is pending on the current signed snapshot; unchanged per-test identities reuse the exact prior green local evidence.

The fourth Windows Test job is green. The downloaded test-evidence-windows-latest artifact contains 48 passing native test/subtest events, no native failure, and 12 passing narrowing controls, each reporting its named expected-red child exit 1. The complete hosted workflow is still pending. All 28 refreshed envprofile/CLI groups exited 0; the other 83 package identities reuse exact green evidence. The refreshed full stream and its sanitized archive equivalent each pass the unchanged platform-case gate (exit 0).

## Round-2 handoff evidence

Full hosted run 39094506341 concluded success on exact signed candidate b32823ac920d8b0485c4a7a7111b48fff8e3ca2b: 20 successful jobs, 2 documented conditional skips (main-only rose-air and candidate-spec workflow-dispatch). The final queued macOS gate self-test ran and passed. gh run watch --exit-status returned 0. The source tree, signature G, remote gate head and preserved Story checkpoint were verified again. All changes remain uncommitted for orchestrator review/integration; local main remains two commits ahead of the checkpoint. No newer trunk qualification is claimed. The coverage map reports 12 of 12 attack rows and 5 of 5 AC rows driven, with the legacy parser bound and three open spec questions explicit. Sanitized logs and command-exit evidence accompany the handoff.

The public board refused binary archive publication with resource_binary_refused (upload exit 1). Its supported plain-text path was used: the full accepted event stream, 107-file metadata bundle, manifest/audit, and local archive path/size/SHA-256 reference are attached. Raw archive retention is explicit; no binary archive is claimed as uploaded. No product or candidate change resulted.
