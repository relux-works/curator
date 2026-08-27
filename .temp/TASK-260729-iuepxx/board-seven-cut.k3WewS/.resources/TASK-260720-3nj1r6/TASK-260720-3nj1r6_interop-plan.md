# TASK-260720-3nj1r6 interoperability runner contract

## Purpose

Run Curator and csk as independent consumers of one curator-spec suite and
detect semantic divergence without importing either implementation or copying
expected values into the runner.

## Ownership boundary

| Owner | Owns | Must not own |
|---|---|---|
| curator-spec | Fixture bytes, stable case IDs, expected normalized outcomes, suite digest, runner | Manager-specific expected values or machine-home layouts |
| Curator | Go adapter from shared cases to Curator behavior | csk code, Python adapter, copied vectors |
| csk | Python adapter from shared cases to csk behavior | Curator code, Go adapter, copied vectors |
| Runner | Isolation, result collection, strict case-set comparison, report | Protocol decisions, broad error-message rewriting, shared manager cache |

## Consumer result envelope

Each consumer emits one record per shared case through the test-only transport
and deterministic JSON Lines contract owned by TASK-260720-2g7avf. The two
native consumers implement that contract; TASK-260720-3nj1r6 validates and
compares it rather than redefining a runner-local envelope.

- `case_id`
- `implementation`
- `suite_sha256`
- `outcome_class`
- `compiler_process_class`
- `persistent_mutation_class`
- `launch_argv`, `launch_stdout`, `launch_stderr`, and `launch_exit_code` when the case launches a command
- diagnostic detail for humans, excluded from parity unless the shared case explicitly makes exact text normative

The runner may normalize temporary absolute paths and platform executable
suffixes declared by the suite. It must not normalize acceptance versus
rejection, rejection phase, build versus no-build, mutation versus no-mutation,
stdout, stderr, argument order, or exit status.

## Required isolation

- One read-only suite root and verified suite digest.
- Separate temporary project, manager home, build cache, Go cache, config root,
  and fixture copy for each implementation and case.
- No cache or built artifact crosses from one manager to the other.
- Timeouts and process failures become explicit failed results.
- A missing, extra, duplicate, skipped, or xfailed case is a gate failure.

## Scenario families

1. Valid schema v6 fixture: cache miss, verified hit, compiler-free dry-run,
   build-root context exclusion, and explicit command launch.
2. Launch behavior: ordered argument forwarding, exact stdout and stderr, zero
   and nonzero exit propagation.
3. Declaration and filesystem rejection: legacy schema, unknown driver,
   forbidden fields, invalid or overlapping roots, path escapes, links, special
   files, invalid module or package shapes.
4. Toolchain and process rejection: unsupported family, toolchain switching,
   downloads, workspaces, cgo, PGO, generators, native inputs, dynamic import,
   external linking, poisoned environment, and telemetry setup failure.
5. Cache and lifecycle rejection: corrupt or untrusted entries, receipt or
   target mismatch, concurrent publisher, build failure, rollback, recovery,
   currentness, repair, consumer ordering, and locked GC.

The case file may reference normative build-driver or manager-lifecycle case
IDs instead of duplicating their payloads.

## Dependency and release order

1. TASK-260720-2g7avf defines the shared executable cases.
2. TASK-260720-1673lr and TASK-260720-31zeo2 implement independent consumers after their native manager conformance handoffs; the csk consumer also waits for TASK-260720-3pemm6 real-Go cross-platform E2E.
3. TASK-260720-3nj1r6 implements this runner after both consumers.
4. TASK-260720-3pvihp waits for TASK-260720-1pvfj5 and TASK-260720-3s27te, then qualifies real manager releases against the exact candidate suite.
5. TASK-260720-vs6den pins those manager releases in curator-spec and runs the cross-platform release gate.
6. TASK-260720-25d05o qualifies the actual protocol release.
7. TASK-260720-38l1sy and TASK-260720-1utsx8 audit the manager handoffs and ensure suite pins advance only to the qualified release.
8. TASK-260720-22ynoi performs the independent acceptance audit.

No branch, mutable tag, guessed hash, local-only pass, or board status may
substitute for either immutable release qualification.

The Curator consumer extends TASK-260720-jrrgw9, and the csk consumer extends
TASK-260720-12r55p plus TASK-260720-3pemm6. Manager release qualification
consumes the candidate integration handoffs from TASK-260720-1pvfj5 and
TASK-260720-3s27te. The later manager pin audits consume those same handoffs
only after TASK-260720-25d05o. Candidate suites are caller-supplied inputs;
committed release pins remain on the previous protocol until immutable protocol
release evidence exists.
