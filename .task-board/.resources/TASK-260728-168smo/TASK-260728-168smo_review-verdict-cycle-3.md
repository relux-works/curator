# TASK-260728-168smo — review verdict cycle 3

## Verdict

**CHANGES REQUESTED.** Route to `analysis` for decision-artifact rework and another reviewer cycle. This is ordinary recoverable research rework, not a stop-the-line boundary. The run is not goal-bound: `task-board spawn goal RUN-260729-a93356` returned `Active Goal: none`.

## Findings

1. **The normative Windows dynamic-library boundary has two incompatible values.** Reference section 10.1 states that the measured `windows/amd64` allow-list is exactly `{KERNEL32.dll, msvcrt.dll}`, explicitly says `ADVAPI32.dll` and `USER32.dll` remain excluded, and requires a new A6 qualification before widening. Section 11.6 K-9 instead records `{KERNEL32.dll, msvcrt.dll, ADVAPI32.dll, USER32.dll}`. The retained W14 evidence and cycle-3 results support only the two-entry set. An implementer cannot know which closed set to enforce, and choosing the four-entry row would silently admit unmeasured behavior. Required change: select the measured two-entry set everywhere unless a new A6 run proves a wider set; synchronize decision 0010, the reference, platform matrix, obligation table, and conformance-vector inputs.

2. **The artifacts claim completed A1-A9 qualification while their own gate says A8/A9 are still future/conditional.** Reference lines 10-13 say `windows/amd64` passed A1-A9 on a real host. Section 11.1 says a tuple is admitted only after all A1-A9, but its table assigns A8 to future task `TASK-260728-251p01` and makes A9 effective only if those conformance vectors pass. Sections 11.3 and 11.5 nevertheless call the tuple qualified and declare the platform retirement branch closed. Decision 0010 has the same split: section 12 says a tuple enters `platforms` only after all A1-A9, then records only A1-A7 as passed and delegates A8/A9. The gate log likewise labels the host transcript A1-A9 although it contains only the host-side checks. Required change: choose one honest state. Either run and attach the A8 classifier corpus and the A9 2.4-family conformance result now, then retain the qualified claim; or call Windows an A1-A7 host-qualified candidate, defer the final `platforms`/`compatibility` admission to `TASK-260728-251p01`, and keep the retirement branch active until that evidence exists.

## Evidence that passed review

- Exactly one v1 artifact model is selected: Kotlin/Native `native-executable-v1`; all Kotlin/JVM runtime-bundle variants and GraalVM are explicitly deferred/rejected.
- Paired identifiers are closed as `kotlin-native-v1` and `kotlin-native-repository-v1`; local `build_roots` and external `skill-build.json` semantics are aligned.
- The source/module allow-list rejects Gradle, Maven, KSP, annotation-processor, compiler-plugin, script, response-file, prebuilt-library, native-interop, arbitrary argv, launcher, network, and package-selected toolchain escape surfaces without a generic task or command.
- The trusted `curator-kotlin-bundle-v1` layout, direct fingerprinted JDK launch, P1/P2 probes, operation-private overlay, cache/receipt/marker binding, signing boundary, macOS exclusion, and Linux deferral are specified. All four cycle-2 findings are structurally addressed.
- Retained ETW XML reports `EventsLost=0` and `BuffersLost=0`; the external `where.exe` control is captured; the compiler-JVM descendant closure contains only in-bundle `clang++.exe` and `ld.lld.exe`. Bundle digest is unchanged across the consolidated run.
- Independent read-only checks: producer resource copies equal the revision-3 worktree documents; scoped link sweep 6 checked / 0 broken; curator-spec Python tests 8/8 pass; curator-spec `go test ./tools/...` passes; Curator `go test ./...` passes across 31 packages.

## Re-review gate

1. One identical Windows DLL allow-list appears in decision 0010, reference sections 10.1 and 11.6, the platform table, and vectors; any widening has fresh A6 evidence.
2. Every use of qualified, passed A1-A9, admission, platform set, compatibility set, and retirement branch describes the same persisted evidence state.
3. If qualification is retained now, attach passing A8 and A9 artifacts. If it remains deferred, remove the completed-qualification claims and preserve the exact downstream admission/retirement condition.
4. Re-run the scoped link sweep and focused specification tests; preserve the already-green A1-A7 Windows evidence and corrected macOS record.

The reviewer modified no producer decision/reference artifact, product code, schema, vector, release file, staging state, commit, publication, pin, toolchain installation, or platform host.