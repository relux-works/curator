# TASK-260720-1nlmvv review verdict — cycle 1

## Verdict: CHANGES REQUESTED

Route: `to-dev`. The candidate is mechanically green but does not satisfy the stable currentness and repair acceptance criteria. No product code was modified during review.

## Blocking findings

1. **Several promised machine-readable states are not reachable through the real CLI.** `planBuilds` treats `corrupt` and `unsupported` as blocking and returns an error before `statusReport`; `cmdStatus` skips the target, so human output has no `state=corrupt-build-cache` or `state=unsupported-build-platform`, while `status --json` emits an empty result. Missing, malformed, or unsupported Go fails during toolchain probing before any build row exists and likewise has only stderr guidance. Separately, `marker.Read` rejects unknown schemas before `scopeStatusDrift` can emit `unsupported-marker`, and rejects non-`go-v1` marker drivers before classification can emit `unsupported-build-driver`. The closed-set unit tests directly construct impossible values and therefore prove enumeration, not production reachability. This fails the requirement that human and JSON diagnostics expose every stable currentness code.

2. **Corrupt compiled cache state is not repaired by install or upgrade.** `BuildCorrupt.blocking()` causes `planBuilds` to return before `stageBuilds`. The CLI prints that corrupt state was refused, although the new README claims missing, corrupt, wrong-target, and untrusted state are rebuilt. The cache publication layer already has quarantine-and-replace support, but this workflow never reaches it. Add end-to-end install and upgrade tests for corrupt receipt bytes and artifact bytes, proving gates run first, the old install remains live until commit, successful repair publishes a protected replacement, and failure preserves the old install/cache.

3. **A logical-key mismatch is over-attributed as target/toolchain drift.** `buildmeta.Input` binds build source, build root, command, source directory, target, toolchain, and policy. After comparing only driver and source, `classifyBuildCommand` labels every remaining opaque key mismatch `build-target-mismatch`. The marker stores no prior complete input that could distinguish target, toolchain, build-root/source-dir/command, or policy causes. Persist/inspect sufficient evidence and classify truthfully, or use a non-attributing stable code plus precise subcodes. Add adversarial cases for every logical input component.

4. **Diagnostic path redaction is not complete.** `RedactDiagnostic` only examines whitespace-delimited tokens, so embedded and URI forms such as `source=/private/cache/x`, `file:///private/cache/x`, and `error=C:\Users\name\cache` survive. In addition, blocking plan errors interpolate the raw cache reason into `Result.Errors` after the redacted plan line. Replace token-only detection with bounded scanning that covers Unix, Windows, UNC, URI, and embedded forms at both status and install/error surfaces; retain invalid-UTF-8, control/format-rune, multiline, and 240-rune guarantees.

## Additional scope adjudication

- `curator global status` remains build-blind even though global install/upgrade can create compiled state. Because the task names the `status` surface and this candidate already changes global install presentation, either include build currentness there or record an explicit contract-level exclusion and tracked follow-up.
- Repeated compiled install remains non-idempotent because `stageNode` does not pass `BuildCurrentness`. It is safe but changes an unchanged install from `up-to-date` to `installed`; retain as tracked rework unless a predecessor owner lands it first.
- Hidden `TestMain` worker dispatch matches production exact-argument dispatch and the worker performs executable identity proof; no user-visible parsed command bypass was found.
- Planned commands are collected across closure nodes and build rows are checked independently of declared-skill drift, so transitive commands are represented by the source path. Add a real transitive compiled-status integration test because current coverage is only indirect.
- Cache-hit classification does compare marker receipt identity, artifact path, and artifact hash with the protected hit. Referenced cache entries remain covered by the existing GC-safe path.

## Independent validation

- `go test ./cmd/curator ./internal/install ./internal/godriver -count=1`: PASS (`196.644s`, `260.263s`, `28.658s`).
- Focused `go test -race` over classification, currentness, redaction, repair notices, real compiled status, and untrusted repair: PASS (`135.109s`, `1.969s`, `1.560s`; godriver had no matching test).
- `go vet ./cmd/curator ./internal/install ./internal/godriver`: PASS.
- Task-file `gofmt -l`: no output. `git diff --check 17804ce`: PASS.
- Producer raw evidence was inspected: full Go tests, full install race, cmd/godriver race, macOS gates, Windows cross/native scoped tests, and Linux crossbuild evidence match the submitted notes. Native Windows classification tests exercise pure classifiers with constructed values; they do not close the production reachability gaps above.

## Required next cycle

Make each stable code reachable and emitted by both human and JSON status, keep ordinary `status` reporting separate from `--check` failure semantics, repair corrupt cache/receipt/artifact state atomically through install and upgrade, classify opaque key drift without overclaiming, harden redaction across every output path, and add end-to-end regressions for these cases. Then rerun normal/race/lint/build/cross-platform gates and attach revised task-scoped evidence.