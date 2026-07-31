# TASK-260728-1yhuqi — review verdict, cycle 1

## Verdict

**CHANGES REQUESTED → `analysis`**

The central architecture is viable: SwiftPM cannot supply the required package graph without executing `Package.swift`, total manager enumeration of `Sources/**/*.swift` is deterministic and non-discovering, and the direct `swiftc` pipeline replayed successfully. The submitted contract is not yet implementation-ready because several load-bearing rules contradict their probe or remain underspecified.

## Findings

### 1. The claimed first-line-only metadata rule contradicts the classifier

The decision says only the first line of `Package.swift` is read and the body is never parsed (`TASK-260728-1yhuqi_decision-0011-swift-driver-pair.md`, lines 341–346, 506–507, 600–601). The reference repeats that contract at lines 229–231 and 432–436.

The ordered classifier nevertheless distinguishes:

- `rejected-absent-header`: no specification on line 1 **and no other line**; and
- `rejected-non-canonical-header`: a specification exists below line 1.

That requires reading later lines. The attached probe implements the contradiction in `classifier.go`: `CuratorClassify` calls `scanForHeader(manifest)` when line 1 has no header (lines 30–39), and `scanForHeader` scans every later line (lines 134–145). Consequently a line-2 header changes the manager verdict through bytes the normative contract says are not metadata input.

Required rework: choose and specify one rule. Either classify solely from line 1 and update the line-2 case/security partition, or explicitly admit a bounded whole-file byte scan and revise every statement that says the body is never parsed or is not a metadata input. The decision, reference, probe, disposition table, and vectors must agree.

### 2. The `swiftc -###` verifier is not fail-closed and does not cover symlink/TOCTOU boundaries

The reference requires every job line to have an absolute executable inside the toolchain root and every plugin-bearing flag value to resolve inside a fingerprinted root or be absent (lines 286–307). The reviewer directive specifically requires fail-closed parsing plus symlink, TOCTOU, path-derivation, plugin/load/server coverage.

The probe does not establish those properties:

- `splitJobLine` toggles single-quote state but returns tokens for an unmatched quote and has no invalid result (`structural.go`, lines 126–149).
- `jobExecutables` silently ignores every non-empty line whose first token is not absolute (`structural.go`, lines 173–182). A relative executable or unknown wrapper line therefore disappears instead of failing.
- `pluginPaths` silently ignores a plugin flag with no following value and does not reject an unrecognized line shape (`structural.go`, lines 151–171).
- `insideAny` is a lexical prefix check (`structural.go`, lines 184–191), not resolved containment. A path lexically below the root but symlinked outside passes this probe.
- S3/S4 exercise only the current well-formed Apple 6.3.2 output. There are no negative parser vectors for an unknown line, relative executable, unmatched quoting, missing plugin value, extra load/server spelling, symlink escape, or an absent plugin path that becomes present before the build permit.

The replayed green result therefore proves the current observed plan is acceptable, not that the proposed manager parser fails closed.

Required rework: define a closed `-###` output grammar and mandatory rejection for every malformed or unknown non-empty line; define quote/escape handling and the complete plugin flag/value grammar; resolve and boundary-check every executable and existing plugin component; specify the symlink and mutation-race rule at the graph-to-permit boundary; add runnable negative controls for each failure family. Graph and compile argv equality should be asserted token-for-token, with `-###` the sole difference.

### 3. Module-name derivation is not total for protocol-valid command keys

Protocol command keys allow `^[A-Za-z0-9][A-Za-z0-9._-]*$`. The Swift reference says only that the module name is manager-derived through `^[A-Za-z_][A-Za-z0-9_]{0,63}$` (lines 225–235); it gives no mapping or rejection rule.

Valid keys such as `my-tool`, `9.tool`, and keys longer than 64 scalars cannot be passed unchanged. A simple replacement rule can collide (`my-tool` and `my.tool`), which would break exact argv identity and can merge Swift module identities even though artifact/cache identities remain distinct.

Required rework: define one total, deterministic, collision-safe derivation for every valid command key, or define an explicit driver-specific rejection and diagnostic. Bind the exact result into local/external conformance vectors, including punctuation, leading-digit, length-boundary, and collision cases.

### 4. The matrix and Windows closure still have explicit-scope gaps

Decision 0008 requires the matrix to decide response files explicitly. This task also names scripts, configurations and `unsafeFlags` in scope. The submitted matrix has no response-file row and does not name scripts or `unsafeFlags`; it relies on broad statements that the manifest body is not an input and the fixed vectors admit no flags. That is directionally sound but not the requested exhaustive audit matrix. It also leaves the disposition of otherwise inert files outside `Sources` implicit.

The closure vocabulary is inconsistent:

- the decision’s Windows contract lists `swiftc.exe`, `swift-frontend.exe`, `clang.exe`, and the linker;
- reference section 2.1 requires five named closure members by additionally requiring `swift`; and
- reference acceptance test 12.1 again requires all five.

`swift` is used by the conformance probe as an upstream oracle but is forbidden from the manager pipeline. The contract must say whether it is a required fingerprinted registry member, a probe-only member, or absent from the runtime closure.

Required rework: add explicit allow/reject/ignored-with-proof rows for response files, scripts, products/targets/configurations, `unsafeFlags`, plugin/macro/binary/system-library declarations, and non-compiler-visible files; keep the distinctions between build rejection, inert bytes, and unreachable channels. Reconcile the exact macOS/Windows named closure and qualification test.

## Evidence that passed

- Extracted `TASK-260728-1yhuqi_probe.tar.gz` to a temporary directory; `go test ./... -count=1` exited 0.
- Replayed `go run .` on macOS arm64 / Apple Swift 6.3.2:
  - 20/20 cases matched;
  - P1 and P2 held with a non-empty security partition;
  - 32 closure checks yielded no verdict;
  - 6/6 expected-red controls failed as required;
  - 18/18 structural checks matched;
  - `green: true`, exit 0.
- Independently exercised SwiftPM metadata commands against the marker-writing manifest. `describe`, `show-dependencies`, `show-executables`, `show-traits`, and `dump-package` each exited 0 and emitted `PROBE-MANIFEST-EXECUTED`. This supports rejecting SwiftPM; no alternative graph query was found among the relevant installed subcommands.
- Total enumeration of every regular `Sources/**/*.swift` file, with every other entry rejected, is a deterministic mapping to one module and is compatible with decision 0008’s non-discovering requirement.
- The submitted macOS measurement supports the manager-owned SDK presentation, the current executable/plugin closure, native target admission distinction, and deferred Windows/Linux qualification. No platform claim is accepted beyond the measured macOS-arm64 evidence.

## Route

This is decision/research rework, not an external blocker and not a stop-the-line human decision. Route to `analysis`, preserve this verdict, revise the decision/reference/probe, attach a new task-scoped outcome, and submit a new reviewer cycle.
