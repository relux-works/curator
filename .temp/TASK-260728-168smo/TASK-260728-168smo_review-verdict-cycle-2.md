# TASK-260728-168smo — review cycle 2 verdict

## Verdict

**CHANGES REQUESTED → `analysis`.**

The rework resolves the cycle-1 trusted-root defect: Kotlin/Native is the only
selected v1 artifact class, every Kotlin/JVM shape remains explicitly deferred,
the operator-curated bundle puts `jdk/bin/java`, the compiler distribution, and
the prehydrated native dependency closure under one tree digest, and the local
and repository drivers share a closed command/source/cache contract. The
macOS evidence also establishes a valid exclusion: this Kotlin/Native release
cannot meet the accepted fingerprinted-process-closure rule on Apple targets.

The pair is still not an implementable accepted contract for the reasons below.

## Blocking findings

### 1. No qualified tuple means the task has not supplied the required registry entry

Accepted `docs/compiled-build-toolchain-requirements.md` section 1.3 requires
this task to supply every Kotlin registry field **on a qualified host**,
including an initial tested `compatibility` set. The submitted record instead
sets both `platforms` and `compatibility` to empty
(`decisions/0010-kotlin-native-driver-pair.md` lines 495–524;
`docs/kotlin-native-build-drivers.md` lines 104–118 and 654–671), so every host
fails before the compiler and every release is untested.

That is a fail-closed reservation, not the task acceptance criterion's
implementable paired driver. Moving qualification into
`TASK-260729-2vfvgi` does not complete this task's inherited obligation:
the new task is blocked by this task and implementation tasks, while schema-8
integration is blocked only by this task. The proposed retirement trigger is
therefore prose sequencing, not a dependency-enforced prerequisite, and it
cannot supply the fields required before this decision is accepted.

Required change: choose one branch now.

1. Qualify at least one exact native tuple through A1–A9, populate the initial
   `platforms`, `compatibility`, per-OS relpath/probe, artifact allow-list, and
   platform overlay semantics in this task's decision/reference; or
2. retire both identifiers now and revise the outcome so it does not claim an
   implementable paired driver.

The evidence supports retaining Kotlin/Native as the selected artifact model
and permanently excluding macOS. It does not support a third state in which an
unqualified pair is accepted for possible later admission.

### 2. The default native target is used but is absent from the registry preflight

The fixed compile vector and cache identity require a “resolved default native
target” (`docs/kotlin-native-build-drivers.md` lines 356–374), and A7 says it is
read from `konanc -list-targets`. But the registry's exact `probe` contains only
`konanc -version` (`decisions/0010...` line 498; reference sections 1.4–1.5).
No normative stage runs `-list-targets`, no bounded parser or failure rule is
specified, and the worker session is otherwise closed to one compile command.

Required change: make target discovery an explicit package-independent Stage A
probe vector (or define another accepted source), with exact argv, stream,
bound, grammar, exactly-one-default rule, token-to-claim mapping, ordering, and
typed failures. Bind its normalized result to the host-pair check, compile argv,
cache input, receipt, and qualification evidence without adding an undeclared
worker command.

### 3. The rejection matrix contradicts its own allow-list and vector plan

The directory rule admits any name matching
`^[A-Za-z0-9_][A-Za-z0-9_.-]*$` (reference line 394). That admits `gradle/`,
`kapt/`, `KSP/`, `META-INF/`, and `services/`. Section 7.2 nevertheless lists
those directories as rejected build-system/plugin inputs (lines 440–441), and
A8 plus the vector inventory require a rejected case for every row (lines 670
and 811). The documented rejection and its planned tests cannot be implemented
from the normative classifier.

The file-level closure still prevents those directories from carrying Gradle
scripts, plugin service files, or binaries, so the generic execution escape
hatch remains closed. The contract must nevertheless say one deterministic
thing. Either reject the named directory paths with an ordered rule and
diagnostic, or state that inert directories are admitted and move only the
actually rejected entry shapes into section 7.2/A8.

### 4. The macOS “exactly seven” evidence inventory is internally inconsistent

The reproduction archive's final `evidence/allowed-externals.txt` contains six
paths and omits Xcode's `usr/bin/xcodebuild`. `run7-enumeration.log` observes
the denied `xcodebuild` attempt, while `run8-enumeration.log` reaches exit 0
without adding it to the final allow file. The decision and logbook call the
set “exactly seven.”

This does not weaken the macOS exclusion: one unavoidable external executable
already violates the accepted closure, and the raw logs prove several. It does
mean the iterative-denial method does not by itself prove a complete child
process inventory. Correct the artifact/prose and use a trace that records every
spawned executable by resolved path before claiming exact completeness.

## Verification

- `go test ./...` — exit 0
- `go vet ./...` — exit 0
- `go build ./...` — exit 0
- curator-spec Python unit tests — exit 0, 8 tests
- curator-spec `go test ./tools/...` — exit 0
- curator-spec `tools/validate.py` — exit 1 on the previously attributed broken
  link in sibling `docs/external-build-repositories.md`; no new attribution to
  this task was found
- Board run `RUN-260729-4fc5ea` is not goal-bound and had no operator directive

No producer document, repository code, schema, vector, release file, staging,
commit, publication, pin, or toolchain installation was modified by review.
