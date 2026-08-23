# TASK-260811-33ukne developer rework outcome

Status: ready for independent review

## Delivered

- Added the production `swiftpm-source-v1` package under `internal/swiftpmsource` with a non-test manager entry point, executor-backed manifest evaluation, brokered lock generation, exact Git acquisition, mirror verification, and C0-C4 source closure.
- Bound Swift, SwiftPM, PackageDescription, and Git at C0 and rechecked each exact tool before affected use.
- Removed unsupported `dump-package --manifest-path`; the production permit and observed launch both use `swift package --disable-experimental-prebuilts dump-package` from the admitted package root.
- Added production offline metadata replay through the shared protected executor. It runs exact `swift package ... --force-resolved-versions show-dependencies --format json`, with network `none`, a frozen lock, isolated home/cache/config/security/scratch state, and admitted read-only kind-preserving mirrors. The executor-issued receipt and observed dependency graph are validated against the capture.
- Added exact source-control acquisition for exact version, range, branch, and revision requirements; complete snapshot extraction; bare same-kind mirror capture; pinned revision/tree/object verification; and immediate byte/content rechecks before replay.
- Added generated-lock resolution permit, causal receipt, intake/permit/receipt journal entries, and C1/C3 evidence.
- Made lock-pin acquisition discovery ordered and rejects missing, duplicate, stale, dangling, wrong-kind, drifted, submodule/LFS/filter/hook, compiled payload, binary target, plugin, macro, and unsafe-setting cases before affected process starts.
- Added semantic and real-process coverage for R01-R13, P01-P08, selection-neutral graph/binding records, deterministic replay, stale lock, transitive mirrors, mirror drift, and exact production SwiftPM/Git paths.
- Updated `README.md` with the production SwiftPM closure and tool workflow.

## Validation evidence

Every command below ran directly as a standalone gate; exit codes are the real process results.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/swiftpmsource` | 0 | Focused package green, including real transitive SwiftPM mirror replay. |
| `go test -count=1 -race ./internal/swiftpmsource` | 0 | Race gate green. |
| `go test -count=1 -coverprofile=.temp/TASK-260811-33ukne/swiftpmsource-cover-02.out ./internal/swiftpmsource` | 0 | 80.2% statement coverage. |
| `go vet ./internal/swiftpmsource` | 0 | Vet green. |
| `golangci-lint run ./internal/swiftpmsource` (first run) | 1 | Truthful red gate: unchecked cleanup and deprecated `tar.TypeRegA`; both corrected. |
| `golangci-lint run ./internal/swiftpmsource` (rerun) | 0 | `0 issues.` |
| `go build ./...` | 0 | Repository build green. |
| `make no-broad-suppression` | 0 | Suppression policy green. |
| `git diff --check -- README.md internal/swiftpmsource` | 0 | Patch whitespace green. |
| accepted Ruby canonical verifier | 0 | 53 canonical records and references green. |
| `go test -count=1 ./...` | 0 | Full uncached repository suite green. |

Full-suite log: `go-test-all-02.log`, 53 lines, SHA-256 `05d75d722ac6077f5f64ce2e1ff71dd79d338f2981794a77dc8622635d5304f7`.

Coverage profile: `swiftpmsource-cover-02.out`, SHA-256 `b60a9b0ea3304bfa62434f41dffa1ce396b349a25f8fcc39fe311e7b2a2a4cb1`.

## Review focus

- Verify that `ExecutorSwiftPM.Replay` commits byte-identical supported argv before the observed `show-dependencies` launch and that all replay inputs are executor-admitted.
- Verify remote and local mirror kind mapping, package-identity-preserving mount names, pinned revision/tree verification, and no original-origin acquisition during replay.
- Verify generated-lock resolution evidence placement in C1/C3 and zero-start behavior for pre-process negative vectors.

No files were staged or committed.
