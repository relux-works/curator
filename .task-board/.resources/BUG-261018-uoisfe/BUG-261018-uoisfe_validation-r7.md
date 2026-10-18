# BUG-261018-uoisfe: round-7 validation

Candidate `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`, signed tree `769581956b4fd7bfd02c2567d7465b60335f5b4b`, signature verdict **G**.
Story checkpoint `5d8e6b07f7f485557d9bbb065c9692348ca065c2` and branch are unchanged; all producer changes remain uncommitted. The detached signed snapshot is published only on its gate branch. Current source, tests and configuration match the tested tree. Fresh protected trunk `b9caa20901ee859eec20e649940693033e6297d4` has no overlap with candidate paths. Root README and CHANGELOG are unchanged.

## Scope and behavior

Finding `unrelated-takeover-argument-discard`, surface row 12, is addressed in shared `rejectsTakeover` at `cmd/curator/command_help.go:128`, called by `cli.run` at `cmd/curator/main.go:283` before all help/dispatch paths. Resolve requires true repair and false dry-run, including help. The redundant late checks in `cmd/curator/env.go` are removed. Rework delta: command_help.go, env.go, native_skills_flags_test.go, and the original row-12 mutant anchor in internal/envprofile/native_skills_mutants_test.go. Native rows 1-11, managed homes, Muse, opencode and specification text are unchanged. No surface-table row is treated as out of contract; the explicit round-2 product exclusions remain.

Coverage is **12 of 12 surface rows driven**, **5 of 5 AC rows driven**; production call sites and all gate-row narrowing controls are in BUG-261018-uoisfe_coverage-map.md. Help cross-product: **595 matrix rows**, **595 no-consent positive controls**, **2975 consent invocations**, 72 registry topics. The full table is in BUG-261018-uoisfe_help-sweep-r7.md and the gaps file. No live UI/login control is claimed. Product calls use temporary manager/native/operator homes; fixture state comparisons and the compiled-binary smoke remain unchanged after help/refusal.

## Actual validation receipts

| Command or suite | Real exit / provider result | Evidence |
|---|---:|---|
| Fresh CLI suite, 15 bounded `go test -json -count=1 -timeout=6m -run <mask> ./cmd/curator` processes | 0 each | curator-01.json through curator-15.json; exact argv and candidate in local-exits.jsonl |
| Fresh envprofile groups 8 and 11, including all native cases, twelve current plants and CLI-source scanner | 0 each | envprofile-08-final.json, envprofile-11-final.json |
| `go test -json -count=1 -timeout=6m ./internal/crossconformance ./internal/scriptpolicy ./internal/scriptworker` | 0 | validation-consumers.json |
| Three install tests that build the current CLI binary (exact names in local-summary.json) | 0 | install-binary-consumers.json |
| `python3 <scratch>/merge-local.py` | 0 | local-summary.json, full-local-merged.json |
| Missing curator-15 receipt control, `merge-local.py --control-missing-group` | 1 expected red | merge-missing-control-final.log: intended `Missing green groups: ['curator-15']` refusal |
| `.github/ci/platform-case-gate.sh <merged-json> <evidence-dir>` | 0 | platform-case.log |
| Same gate on the sanitized published JSON stream | 0 | platform-case-sanitized.log |
| `go build -o <scratch>/curator ./cmd/curator` | 0 | build-04.log |
| `go vet ./...` | 0 | vet-04.log |
| `golangci-lint run` with a private scratch cache | 0 | lint-05.log |
| `git diff --check`; empty `gofmt -l cmd internal` | 0 each | diff-check-03.log, gofmt-02.log |
| Compiled production-binary smoke, 11 invocations | 0 harness | binary-smoke.json: each actual expected exit 0 or 2, fixture tree unchanged |
| Signed-tree/fresh-trunk/index integrity check | 0 | final-integrity.json |
| Full hosted CI gate on this exact signed commit | success (provider); status-read command exit 0 | [run 39158440563](https://github.com/relux-works/curator/actions/runs/39158440563); hosted-current.json |

Local suite evidence covers **85 of 85 served packages**, **3497 unique top-level test names**. This is a composed full suite, not a claim that all packages were freshly replayed: all 434 CLI tests were rerun, as were the complete scriptworker/scriptpolicy/crossconformance packages, envprofile groups 8/11 and the three install binary consumers. The remaining 79 packages, 11 envprofile groups, and other install tests reuse already-green evidence from unchanged source/test/configuration/environment identity. The prior signed candidate is `3e00fcc8abe75e411a88773c65c4042ae2448eb0`; prior full-stream SHA-256 is `3d454853090a56e9c3482b92ebeb62c6004c737a055bbaa905f9e337e9be0b85`. Source/configuration comparison and consumer searches are recorded in reuse-identity.json. The full composed JSON and its per-file raw/sanitized hashes are published in the evidence bundle. The local gate explicitly admits 36 classified case skips: 20 platform-control, 1 root-content, 6 host-capability, 8 opt-in and 1 helper-process; every reason is attached in skips-observed.tsv. Top-level names are pass/skip coverage, not a claim that skipped cases executed. Toolchain: Go 1.25.5, Darwin arm64; clean conformance root commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`.

Hosted gate: **20 successful jobs**, with the workflow's two conditional lanes skipped (Candidate suite (${{ matrix.os }}), Test (rose-air)). Linux/macOS/Windows test, Linux/macOS race, lint, conformance, naming, driver matrix and gate-selftest jobs succeeded. These skips are workflow conditions, not failed jobs called passing. Hosted status is read from the provider, not synthesized. Earlier superseded candidate runs were cancelled and are not acceptance evidence. Configured post-turn validation has not run yet and is not represented as passing.

## Narrowing controls and honest bounds

| Mutant | Narrowed permission | Named failing test | Result / survival bound |
|---|---|---|---|
| resolve-help-missing-repair | Admit only invalid repair with help, retaining dispatch rejection | TestReviewNativeTakeoverHelpBoundary | Child 1 for forbidden actual help 0; parent 0 |
| resolve-help-dry-run | Admit only true dry-run takeover with help | TestReviewNativeTakeoverHelpBoundary | Child 1 for forbidden actual help 0; parent 0 |
| list-only-consent-bypass | Exempt only list from the otherwise present gate | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Child 1 for actual successful list admission; parent 0 |
| claude-takeover-without-repair, relocated | Exempt only claude_code resolve from true-repair gate | TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code | Child 1, parent 0. Bound: actual downstream profile_unknown 1 rather than usage 2; not a successful repair witness |
| Historical late-handler-only plant | Exempt claude_code only in the now-retired handler check | None: survivor | Exploratory suite exit 1, child 0. Bound: the new earlier guard still protected admission. It was retargeted to the shared production guard; no kill claimed for this historical survivor |

The permanent reviewer witness first failed with real exit 1 before the repair. The additional dry-run witness also first failed with exit 1. Exploratory lint and mutant-harness failures remain recorded as failures in checks.json; they are not current green receipts. Both new help plants preserve searched repair tokens and execute the behavioral witness. The twelve native plants in envprofile group 8 also require named behavioral failures, not compilation/timeouts. The evidence merger was independently weakened by removing one green receipt: it refused with the intended assertion, exit 1.

## Open specification questions

1. Muse native ownership across independent XDG roots.
2. Legacy adapter-ledger arbitration with the current profile; ledger entries remain untouched and non-current.
3. Opencode machine-global discovery versus profile-local ownership; current behavior is preserved.

The developer handoff is ready for review. No independent acceptance or integration is claimed.
