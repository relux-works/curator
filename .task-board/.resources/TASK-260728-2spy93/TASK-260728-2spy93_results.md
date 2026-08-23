# TASK-260728-2spy93 — additional-driver version and artifact boundary

Developer handoff evidence. Status: ready for review (rework cycle 5).

**Current state.** This document records the original cycle-1 decision content,
which still stands. Five review cycles have since hardened the enforcement gate
only; each is recorded in its own artifact, and the newest one supersedes the
figures in this file where they differ:

| Cycle | Verdict artifact | Rework artifact | Subject |
| --- | --- | --- | --- |
| 1 | `..._review-verdict.md` | `..._rework-01.md` | wire/toolchain contradiction, policy identity drift, admission ambiguity, incomplete gate |
| 2 | `..._review-verdict-cycle-2.md` | `..._rework-02.md` | unconstrained reserved toolchain schemas, claim-v4 admission |
| 3 | `..._review-verdict-cycle-3.md` | `..._rework-03.md` | object-keyword and array-applicator escapes |
| 4 | `..._review-verdict-cycle-4.md` | `..._rework-04.md` | `buildArtifactV1` closed as an exact object schema |
| 5 | `..._review-verdict-cycle-5.md` | `..._rework-05.md` | the three shared artifact identity definitions pinned structurally |

Current figures: decision 0008 is 1082 lines; `tools/validate.py` is +1002 lines
over the accepted base and `tools/test_validate.py` +1236; the suite is 91 tests.
Scope is unchanged at exactly three files.

## Execution base

- Task worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2spy93/curator-spec-worktree`,
  branched independently at `57c1f56` and re-materialized from the accepted
  `TASK-260728-2kp3tv` candidate.
- Base fidelity verified before any edit: `diff -r --exclude=.git` between the
  predecessor worktree and this one produced no output (exit 0).
- The predecessor worktree was read only. Nothing was staged, committed,
  published, pinned, or advanced in `curator-spec`.

## What was decided

New file `decisions/0008-additional-language-driver-boundary.md`
(561 lines at cycle 1, 1082 lines now).
It is a boundary decision: it adds no schema, regenerates no vector, mints no
release metadata, and makes no platform claim.

Numbering note: this record was drafted as 0007 and renumbered to 0008 after the
parallel `TASK-260728-1g0z69` worktree was found to already hold
`decisions/0007-compiled-build-toolchain-preflight.md`. The two worktrees are
independent so nothing conflicted on disk, but the number would have collided at
merge. The renumber is mechanical: filename, title, the
`ADDITIONAL_DRIVER_BOUNDARY_DECISION` constant in `tools/validate.py`, and one
docstring. Section 6 of this decision now cites decision 0007 by number as well
as by task ID. If review lands the two decisions in the other order, only these
four references move.

### 1. Version boundary

Next protocol surface is named `1.0.0-rc.6`, reserved here and minted only by
`TASK-260728-251p01`. Reserved wire versions: manifest schema 8
(`agent-skill-v8` / `csk-skill-v8`), descriptor `skill-build.json` schema 2,
build receipt schema 3 (local source mode) and schema 4 (external source mode),
install marker schema 4, conformance claim schema 4.

Deliberately **not** re-versioned, with reasons recorded: `Skillfile.dev.json`
stays at schema 2 (substitution selects acquisition only), the execution policy
stays `manager-worker-v1`, the native-control inventory stays
`rc5-native-control-inventory-v1`, the capability-evidence record stays
`capability-evidence-v1`, and `curator-build-source-v1` is reused unchanged for
every driver in both source modes.

Receipt numbering rule fixed generally: one receipt schema version per source
mode per admitting protocol version. `go-v1` keeps receipt 1 and
`go-repository-v1` keeps receipt 2 inside a schema-8 manifest.

### 2. Six closed driver identities

`rust-v1`, `rust-repository-v1`, `swift-v1`, `swift-repository-v1`,
`kotlin-native-v1`, `kotlin-native-repository-v1`, alongside the admitted
`go-v1` and `go-repository-v1`. The identifier space is closed by enumeration,
never by grammar or detection; every schema must express `driver` as a `const`
inside a `oneOf`; no `language`, `toolchain_family`, `build_system`, or
`backend` selector may exist anywhere.

Reservation is explicitly not admission. A reserved identifier whose contract is
rejected is retired unused and may never be reassigned to another language,
backend, artifact class, or source mode.

### 3. Artifact class — the substantive call

Exactly one class is admitted, `native-executable-v1`: one bounded regular file,
manager-named `bin/<command>` or `bin/<command>.exe`, runnable with only the
target platform's base-installation libraries, never executed by the manager.
Debug info, PDBs, module/interface files, import libraries, and incremental
state are staging by-products and are discarded.

The `runtime-bundle` class is **rejected** for this version, with four recorded
reasons: it would redefine the single-artifact identity that receipts, markers,
shims, currentness, and GC all rest on; a manager-generated launcher is
manager-authored executable content derived from package data, which reopens the
install-time execution surface; an unfingerprintable execution-time runtime
cannot enter cache identity, so a marker could report current for something that
no longer runs; and `protocol/core.md` 12.1 requires the shim to point exactly at
the marker-selected artifact.

Consequences stated plainly: a driver/platform pair that needs a published
sidecar runtime is not admitted for that platform and fails with
`build_artifact_class_unsupported`; Kotlin is admissible only through a native
backend, which is why the family segment carries `native`; and this version does
not require bit-reproducible artifacts, since identity is input-keyed.

### 4/5. Local and external source ownership

Local drivers reuse the schema-6/7 context-excluded `build_roots` model
unchanged. The local command object stays exactly `{type, driver, source_dir}`;
no driver may add a package-controlled member, and each local driver must bind
one closed project-metadata file at the build root with the nearest-ancestor
rule and a deterministic non-discovering `source_dir` to single-program mapping,
or be rejected rather than widen the command.

`skill-build.json` stays the sole neutral descriptor. Schema 2 changes exactly
one thing — the target's `driver` becomes a `oneOf` over the four external
identifiers — and adds no member. The repository owns the descriptor version;
managers read schemas 1 and 2; a non-Go external command against a schema-1
descriptor fails `build_descriptor_driver_unsupported`, an unknown version fails
`build_descriptor_schema_unsupported`, and neither falls back.

### 6. Toolchain boundary (explicit dependency, not duplicated)

The requirement, resolution, version grammar, two-stage preflight, diagnostics,
and guidance catalog belong to `TASK-260728-1g0z69` and are integrated by
`TASK-260728-2jaw7h`; decision 0008 does not restate them. It fixes only the
five boundary properties the version and artifact model depends on, of which the
load-bearing new one is: every executable started below the worker must be a
fingerprinted member of the driver's declared trusted toolchain closure,
including any platform linker, SDK, sysroot, runtime library, or archiver, or
the driver must reject that platform.

### 7. Execution policy

All eight drivers keep the portable `manager-worker-v1` policy. The process
graph is restated generically (worker → driver's fingerprinted launcher →
fingerprinted executables in that driver's closure); the two lower nodes were
always per-operation values bound by the toolchain identity inside the build
input, so every `go-v1` and `go-repository-v1` input, cache key, receipt, marker,
and claim is unchanged. Minting `manager-worker-v2` was considered and rejected
because it would change every `go-v1` cache key and invalidate the frozen rc.5
candidate for no security gain.

The worker session shape is generalized to exactly one read-only graph phase of
at most one command and exactly one compile phase of exactly one command. A
driver that cannot map onto it is not admitted; widening the session requires a
new execution-policy identity.

Each driver must apply an exhaustive deterministic pre-compile rejection matrix
in the position where `go-v1` rejects `SysoFiles` and
`//go:cgo_import_dynamic`, covering build scripts, procedural and compiler
macros, plugins, annotation processors, generators, manifest programs, tasks,
recipes, response files, package-selected linkers and native libraries, and
network/registry access. Shared class:
`build_package_code_execution_forbidden`.

### 9. Security, platform, signing, deferred containment

No hardened guarantee is added, implied, or claimed; the six deferred guarantees
stay with `STORY-260728-327soo` and the decision does not even name them. The
honest security delta is recorded: three more compiler front ends under the same
portable, non-hardened controls widen compiler-input exposure, and all three
languages ship a mainstream build path whose normal operation executes
package-selected code — which `SECURITY.md` already forbids a manager to invoke.
That is why the pre-compile rejection matrix is mandatory rather than advisory.

No platform claim is made. Each reserved identifier starts with an empty
qualified-platform set; macOS and Windows remain the portable platforms and
Linux stays excluded until `TASK-260728-1skseh`. No driver signs, timestamps, or
notarizes; a platform requiring local signing rejects the build until the
separately reviewed signer profile exists. A linker-applied ad-hoc signature is
classified as compiler output, not a manager signing step.

### 10. Downstream obligations

Named per task: `1g0z69`/`2jaw7h` (toolchain contract), `12pnm1`/`1yhuqi`/`168smo`
(one accepted contract per pair, including the per-platform proof that the
artifact class is met, and for Kotlin the backend choice inside the class),
`251p01` (integrate accepted contracts only, keeping schemas 1-7 byte-stable),
`2bu2q6` (evidence-backed claims only), `327soo` (hardened guarantees).

## Code written

`tools/validate.py` (+216 lines): new `validate_additional_driver_boundary()`
gate wired into `main()`, plus helpers `is_decision_record`,
`reserved_schema_slot_paths`, `driver_bearing_definitions`, `set_at`. It:

1. requires decision 0008 to fix every admitted and reserved identifier, both
   artifact classes, the four boundary failure classes, the portable policy, and
   the hardened-profile owner;
2. forbids decision 0008 from naming any of the six deferred hardened
   guarantees;
3. rejects any reserved identifier appearing on a surface file outside
   `decisions/` (decision records are where identifiers are proposed and
   retired; every other surface is admission);
4. requires every driver-bearing `common.schema.json` definition to close
   `driver` with a `const` over the admitted set and to require it;
5. requires `buildArtifactV1` to stay exactly `{path, sha256, size}` and every
   build record carrying an artifact to carry no bundle member
   (`artifacts`, `bundle`, `classpath`, `interpreter`, `launcher`, `runtime`,
   `sidecar`);
6. keeps the seven reserved schema slots unallocated; and
7. proves against the compiled validators and the generated positive cases that
   `agent-skill-v6/v7`, `csk-skill-v6/v7`, `skill-build-v1`, `build-receipt-v1/v2`,
   `install-marker-v2/v3`, and `conformance-claim-v3` each reject each of the six
   reserved identifiers — 66 rejection assertions.

Reserved names are assembled from a family tuple and a source-mode suffix so
neither the guard nor its tests match their own source; a test asserts that.

`tools/test_validate.py` (+224 lines): `AdditionalDriverBoundaryTests`, 14 tests
including negative probes for the absence guard, the const guard, a generic
string driver, a bundle member on a build record, a multi-file artifact, an
early-minted schema slot, a decision that stops fixing a closed term, and a
decision that claims a hardened guarantee.

Files changed versus the accepted base — exactly three:

```text
decisions/0008-additional-language-driver-boundary.md   (new, 1082 lines)
tools/validate.py                                       (+1002)
tools/test_validate.py                                  (+1236)
```

Nothing under `conformance/`, `schemas/`, `release/`, `protocol/`, `profiles/`,
`docs/`, `SECURITY.md`, `COMPATIBILITY.md`, or `CHANGELOG.md` was touched.

## Gate evidence (real exit codes, each run standalone)

In the task worktree:

| Command | Exit | Result |
|---|---|---|
| `python tools/validate.py` | 0 | validated 42 schemas and 422 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | Ran 91 tests, OK (29 before; 44 at cycle 1) |
| `go test ./tools/...` | 0 | ok generate-vectors |
| `go vet ./tools/...` | 0 | clean |
| `gofmt -l tools/` | 0 | no output |
| `python -m compileall -q tools` | 0 | clean |
| `git diff --check` | 0 | clean |

Determinism and candidate preservation:

| Probe | Exit | Result |
|---|---|---|
| `go run ./tools/generate-vectors -root .` (run 1) | 0 | `diff -r conformance/v1` vs pre-run snapshot: 0 |
| `go run ./tools/generate-vectors -root .` (run 2) | 0 | `diff -r conformance/v1` vs snapshot: 0 |
| `diff release/1.0.0-rc.5.json` vs snapshot | 0 | unchanged |
| independent `sha256(conformance/v1/manifest.json)` | — | `sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`, identical to the accepted predecessor pin |

Clean-checkout probe (isolated scratch git repo under
`.temp/TASK-260728-2spy93/clean-probe`, never the real repository):

| Command | Exit |
|---|---|
| `make validate` | 0 (91 tests, go test ok) |
| `make regenerate-check` (run 1) | 0 |
| `make regenerate-check` (run 2) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

`make regenerate-check` cannot pass inside the task worktree, because it runs
`git diff --exit-code` against `57c1f56` while the worktree carries the
uncommitted rc.5 candidate. That is why it is reported from the clean probe,
where it exits 0 twice, and why determinism is additionally shown in the task
worktree by diffing regenerated output against a pre-run snapshot.

Negative probes, each reverted immediately afterwards, each confirming the guard
fires with the right message:

| Probe | Exit | Diagnostic |
|---|---|---|
| reserved driver added to a `common.schema.json` enum | 1 | `schemas/v1/common.schema.json:210: reserved driver 'rust-v1' is not admitted by any schema version` |
| `buildCommandV6.driver` widened to an enum over admitted drivers only | 1 | `$defs.buildCommandV6.driver is not a const over the admitted drivers` |
| reserved driver planted in `protocol/core.md` | 1 | `protocol/core.md:547: reserved driver 'swift-v1' is not admitted by any schema version` |
| `launcher` member added to `buildRecordV2` | 1 | `$defs.buildRecordV2 admits runtime-bundle members ['launcher']` |
| restore, re-run validator | 0 | validated 42 schemas and 422 vector files |

Honest note on one probe: planting a reserved schema slot file
(`schemas/v1/agent-skill-v8.schema.json`) makes `tools/validate.py` exit 1, but
the diagnostic comes from `validate_schemas`, which runs first and rejects a
schema with no positive/negative cases. The slot guard itself is therefore
defense in depth on the CLI path; it is exercised directly by
`test_reserved_slot_guard_fires_when_a_slot_is_minted_early`, which patches
`reserved_schema_slot_paths` and asserts the boundary function raises.

## Deliberately not done

- No schema file, conformance case, vector, fixture, or release file was added
  or changed. `1.0.0-rc.6` exists only as a reserved name in a decision record.
- No `CHANGELOG.md` or `COMPATIBILITY.md` entry: both record shipped wire
  changes, and this decision ships none. They belong to `TASK-260728-251p01`.
- No `docs/` companion. The 0005 and 0006 companions are author/operator
  guidance for usable features; nothing here is usable by an author yet.
- No commit, stage, publish, pin advance, native validation claim, or platform
  claim.

## Reviewer focus

1. Whether rejecting the `runtime-bundle` class is the right call given it
   constrains `TASK-260728-168smo` to a Kotlin native backend or to retiring both
   Kotlin identifiers unused.
2. Whether keeping `manager-worker-v1` while restating its process graph and
   session shape generically is acceptable, or whether decision 0006's
   "different execution contract requires a different identity" rule should be
   read as forcing a new policy identity despite the frozen rc.5 cost.
3. Whether receipt numbering per source mode (3 local / 4 external) beats one
   receipt version per driver.
4. Whether the "local command object stays exactly `{type, driver, source_dir}`"
   constraint is implementable for Rust and Swift, or whether it will force a
   driver rejection that the story does not want.
5. That every `go-v1` and `go-repository-v1` identity is byte-unchanged: the
   regeneration and pin evidence above is the intended proof.
