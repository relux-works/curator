# Schema 8 impact analysis

Task: `TASK-260823-omp8zt`  
Date: 2026-08-23  
Scope: curator-spec landing sequence, CocoaSkills implementation impact, and curator-spec residual work.

## Executive findings

1. **Do not land the schema-8 suite on `curator-spec/main` as a vectors-only change.** The current pinned implementation jobs do not fail against the staged schema-8 suite, but this is a false-green result: the Go pin consumes only schema-6 cases and the Python pin does not consume the schema-case index. The landing PR must therefore advance implementation pins and add schema-8 consumption assertions, not rely on the existing green jobs.
2. **Qualify schema 8 as a candidate first.** Curator's only candidate path is the non-default `candidate-conformance` job under explicit `workflow_dispatch`; it accepts exactly one immutable full SHA or pre-materialized root, keeps `SPEC_PIN` unchanged, sets `CI_REQUIRE_FULL_ROOT=1`, and labels its output candidate-only. Default jobs continue to use the one committed `SPEC_PIN`.
3. **Schema 8 is one shared bump.** It carries both `script-worker-v1` and first-party module roots. CocoaSkills needs two implementation stories, with shared manifest/marker/CI foundations. Script-worker containment is the larger body of work; module roots requires replacing the current unconditional `Module.Replace` rejection with a declared-set/effective-set bijection.
4. **The spec work has material residuals.** The in-flight tasks cover manager/core/security/schema/vectors and landing prose, but do not own a closed audit-record wire surface, registry gating/profile changes, implementation-case coverage in `implementations.yml`, or a new immutable rc.9 release surface. Current feature branches modify `release/1.0.0-rc.8.json`; that file belongs to the published rc.8 history and must not be changed.

## 1. Sequencing and CI

### Exact current mechanism

`curator-spec/.github/workflows/implementations.yml` runs on every push to `main`, every pull request, and manual dispatch. It checks out:

- Go Curator at `bd6ba08acda3dc801512c408c759ac0ac6f79f26`;
- CocoaSkills at `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`;
- Curator Skill Registry at `d690bea6fab1c8e6392e05d3a3cdfcf1168bc914`;
- the specification checkout's own `conformance/v1` as `CURATOR_CONFORMANCE_ROOT`.

This is the mechanism introduced by `Pin landed agent-skill implementations (#12)` (`57c1f568...`) and retained on `origin/main` at `be7861c...`.

Curator's `.github/workflows/ci.yml` is a separate consumer-side mechanism:

- one workflow-level `SPEC_PIN`, currently `00b1688a9b2457ca397a0bb550acf47cad8ee967`, is used by every default job;
- `candidate-conformance` runs only when `github.event_name == workflow_dispatch` and one candidate input is non-empty;
- `.github/ci/candidate-suite.sh` requires exactly one candidate source, rejects a non-40-hex revision and a candidate equal to `SPEC_PIN`, records the revision/manifest/tree identities, and stamps the evidence as neither a release nor a conformance claim;
- `CI_REQUIRE_FULL_ROOT=1` makes package deferral fatal in the candidate lane;
- the candidate job never changes `SPEC_PIN`.

The comments still say “schema v6”, but the checks are version-agnostic and can qualify schema 8. The wording should be generalized as maintenance, not forked into a second candidate mechanism.

### Does schema 8 break the pinned implementation jobs?

**Observed answer: no, not with the currently staged suite; the jobs are falsely green.**

I ran the exact Go and Python commands from `implementations.yml` against the current script-worker candidate root (`spec/script-worker-v1-normative` worktree, HEAD `a690d63e...` plus the active vector task's working-tree output, manifest SHA-256 `9dc1c659697f7218dd844863cef08255a4526b4ab04bdcdb6cfc7d21e3784322`):

| Gate | Pinned implementation | Result | Why this is insufficient |
| --- | --- | ---: | --- |
| `go test -count=1 -v ./internal/interop ./internal/closure ./internal/skillspec` | Curator `bd6ba08...` | exit 0 | `TestSchemaV6ConformanceCases` switches only on the two schema-6 names and skips every other indexed schema. |
| `python tools/run_pytest_no_skips.py -q .../tests/test_protocol_conformance.py` | CocoaSkills `6fc2fd9...` | exit 0; 106 passed | This revision checks selected legacy vectors and never reads `schema-cases/index.json`. |

Logs: `.temp/TASK-260823-omp8zt/pinned-curator-go-test.log` and `.temp/TASK-260823-omp8zt/pinned-cocoaskills-pytest.log`.

Therefore:

- landing schema-8 files alone is not expected to make the current pinned jobs red;
- the green result does **not** establish schema-8 conformance;
- adding only new vector files is invisible to the old pins unless their tests explicitly enumerate those files;
- changing an existing vector or shared fixture could still make a future landing red, so a vectors-only merge is not stable even operationally.

The landing gate must assert that every implementation consumes the schema-8 schema cases and both new behavioral vector families. A pin that merely passes while ignoring them is not a qualifying pin.

### Required landing order

1. **Create one immutable candidate commit outside `main`.** Merge the reviewed script prose/profile/schema work and the module-roots work into a schema-8 candidate branch; generate both vector families twice. The candidate must use a new protocol version (`1.0.0-rc.9`), not rewrite rc.8 metadata.
2. **Record candidate identity.** Capture the full commit SHA, suite-manifest SHA-256, whole-tree SHA-256, file count, and proposed `release/1.0.0-rc.9.json` identity.
3. **Qualify Curator through the existing candidate lane.** Run explicit `workflow_dispatch(candidate_ref=<full SHA>, candidate_manifest_sha256=<digest>)`. Keep the committed `SPEC_PIN` unchanged.
4. **Implement and qualify both managers against that exact candidate.** Curator and CocoaSkills changes must consume schema-8 schema cases and the named script-worker/module-roots behavioral vectors. CocoaSkills' current candidate CI is hard-coded to the rc.6 SHA/digest through repository variables and literal assertions; generalize it or add an equally fail-closed explicit candidate input before citing it for schema 8.
5. **Land implementation commits while their default released-suite pins remain unchanged.** This preserves current users and makes candidate evidence non-release evidence, exactly as the candidate contract requires.
6. **Atomically update `curator-spec` implementation pins and coverage assertions in the schema-8 landing PR.** Pin Curator, CocoaSkills, and any affected Registry commit that has green own CI and demonstrably consumes the exact candidate suite. This prevents a `main` interval in which new normative bytes are paired with non-consuming pins.
7. **Merge the spec landing PR only after all required Specification and Implementations jobs are green with those new pins.** Both schema-8 feature families must be present because schema 8 is their single shared bump.
8. **Publish rc.9 through the normal release gate.** Create and validate `release/1.0.0-rc.9.json`, update release tooling, squash-merge through GitHub, verify the main release target, then create the signed annotated `v1.0.0-rc.9` tag/release. Never retag or rewrite rc.8.
9. **Advance consumer released-suite pins only after rc.9 is published and qualified.** Update Curator's single `SPEC_PIN` to the exact immutable rc.9 release commit in a separate consumer PR; run default CI on all three OSes. Advance CocoaSkills' released-suite pin/variables under the same rule. Candidate SHA/digest inputs never become defaults implicitly.
10. **Complete compatibility documentation before publication.** `CHANGELOG.md` must contain the schema migration; `COMPATIBILITY.md` must name manifest schema 8, marker v4, the shared decision-0008/0009 bump, legacy rejection behavior, `script-worker-v1` opt-in/declared-only behavior, module-roots absent/empty compatibility, and rc.9 superseding rc.8 without changing rc.8 bytes.

## 2. CocoaSkills change list and estimates

Estimates are Fibonacci implementation points, including focused tests but excluding cross-platform CI wait time. They are mapped to the two existing CocoaSkills stories rather than proposing a parallel story structure.

### Shared foundation used by both stories

| Change | Concrete scope | Estimate | Story ownership |
| --- | --- | ---: | --- |
| Schema-8 and marker-v4 foundation | Extend `skillspec.SUPPORTED_SCHEMA_VERSIONS`; add immutable fields to `CommandSpec`; preserve schemas 1-7 rejection; add marker v4 read/write/currentness for skill schema 8; update fixtures and serializers without changing marker v1-v3 bytes. Hotspots: `src/csk/skillspec.py`, `src/csk/install_marker.py`, `src/csk/installer.py`, `src/csk/status.py`. | 5 | `STORY-260822-2evh3p`, with module-roots acceptance added in the same schema parser before either story closes. |
| Candidate-suite consumption contract | Add schema-8 index consumption and explicit coverage assertions; authenticate exact candidate SHA/manifest digest; remove rc.6-only literals from the candidate lane while keeping the released lane pinned. | 5 | Shared dependency; place under the first story executed, then reuse from `STORY-260822-27ze8z`. |

### `STORY-260822-2evh3p` — script-worker-v1 implementation

| Change | Concrete scope | Estimate |
| --- | --- | ---: |
| Manifest parsing | On schema-8 script commands accept only co-present `execution_policy: script-worker-v1` and `interpreter: python3-v1|node-v1`; absence means declared-only; reject either field on schemas 1-7, top level, system/build commands, null/unknown/successor/hardened values, and pathless scripts. Store both fields in `CommandSpec`. | 3 |
| Containment-profile derivation | Derive deny-by-default values from the *declared manifest bytes*, not schema defaults: offline/no proxy, interpreter-only `PATH` plus declared `exec`, private temp/config/cache/working roots, declared write targets, no secret injection, filtered `env_read`, and reserved manager-owned environment families. | 8 |
| Manager-owned script worker and launcher | Add a distinct hidden script-worker protocol and manager-owned launcher; reuse only proven generic primitives from `go_v1.py`, while keeping the script protocol, request shape, interpreter identity, session nonce, streams, and result separate. Resolve interpreter before package-controlled directories; reject symlink/reparse/hard-link substitution; re-check identity at launch; never use a shebang, shell, `.cmd`, file association, or inherited `PATH`. | 13 |
| Cross-platform native controls and lifecycle | Probe/apply the eight-row inventory per invocation; macOS/Linux process-domain teardown and file-size controls, Windows Job Object controls, and Linux host-conditional cgroup/Landlock/netns controls; preflight mandatory controls at install/update and invocation; join the complete worker domain. Honest unavailable/host-conditional reporting is required. | 13 |
| Evidence and operator surfaces | Produce exactly one closed `script-capability-evidence-v1` record per invocation; expose it only through plan/status/operator-selected diagnostics; never put it in stdout/stderr, cache keys, receipts, markers, or claims; add all `script_execution_*` diagnostics. | 8 |
| Audit labeling | Carry per-command effective policy identity or absence into the audit result; emit `script-command-declared-only` for every non-opted-in script and `script-command-unfiltered-declared-network` for enforced host-glob declarations. Update detector/model/serialization/cache tests; do not present either warning as enforcement. | 5 |
| Shared-vector and platform coverage | Consume all schema-8 script cases, `script-host-execution-policy.json`, marker-v4 fixture, evidence closure, mandatory-preflight rejection, audit labels, and the three-OS matrix with no skips. | 5 |

**Story estimate: 55 points including the 8-point shared foundation.** The main risk is not parsing; it is creating a second, interpreter-oriented worker without accidentally inheriting build-only assumptions such as closed stdin, bounded output, a policy deadline, toolchain-tree hashing, or build cache identity.

### `STORY-260822-27ze8z` — module-roots implementation

| Change | Concrete scope | Estimate |
| --- | --- | ---: |
| Manifest parsing and containment | Add `modules: tuple[str, ...]` only to schema-8 local `go-v1` commands; validate portable non-`.` paths, uniqueness, real/link-free directories, direct `go.mod`, pairwise disjointness, and disjointness from every build/runtime root under exact and platform folding. | 5 |
| Effective replace parser and bijection | Parse only replacement annotation lines from `<build root>/vendor/modules.txt`; accept exactly one relative directory token; reject redirects/versioned targets/malformed shapes; compare normalized declared and effective sets in both directions; keep absent/empty as an empty replace set. Do not parse `go.mod` and do not trust `Module.Replace.Dir`/`GoMod` as existence evidence. | 8 |
| Fix current `Module.Replace` rejection | Replace `src/csk/builds/go_v1.py`'s current `_validate_module` condition (`module.get("Replace") is not None` => `vendor_metadata_inconsistent`) with classification: declared directory-form replacements are admitted only after the bijection; unreplaced dependencies retain existing vendor/version checks; redirect/undeclared cases receive the new stable diagnostics. | 8 |
| Scan-surface extension | Apply cgo/C/C++/ObjC/Fortran/SWIG, `SysoFiles`, `SFiles`, and exact `//go:cgo_import_dynamic` checks to declared module inputs and their vendor copies. Withhold decision-0005 vendored exceptions from any module carrying a replacement; retain them for `Replace == nil`. | 5 |
| Lifecycle and evidence integration | Thread module sets through build planning/status/currentness; keep build-source/cache/receipt identities unchanged because the whole snapshot is already hashed; add diagnostic mapping and module-root vector adapters. Marker v4 is shared foundation, not a module-specific new marker. | 5 |
| Shared-vector and platform coverage | Consume valid declaration, escape, redirect, undeclared, unused, nested, build/runtime overlap, vendor-copy divergence, and Windows-collision cases; assert every named case is consumed. | 5 |

**Story estimate: 41 points including the 10-point applicable shared foundation.** The highest-risk seam is the transition from “reject every `Module.Replace`” to “admit exactly the declared/effective bijection” without widening ordinary vendored dependency handling.

## 3. Curator-spec residuals beyond in-flight tasks

### Residuals that require explicit ownership

| Residual | Evidence and required change | Estimate |
| --- | --- | ---: |
| Audit-record wire surface | Decision 0008 requires policy identity in the skill audit record, but `audit-record-v1.schema.json` leaves `audit` optional and open-ended; the prose says the record “carries” per-command identity without defining a required field or closed shape. Add a normative, versioned shape (recommended: audit-record v2 preserving v1) with per-command policy identity/absence and warning classes; add positive/negative cases and expected registry records. | 5 |
| Registry profile and implementation | `profiles/registry-service.md` contains no rule for indexing/gating script execution policy or warning classes. Define v1/v2 acceptance/migration, query/gate semantics, signature coverage, and no inferred enforcement. Update the pinned Registry implementation/tests before advancing its `implementations.yml` pin. | 5 spec + implementation work |
| Implementation coverage contract | Current pinned jobs pass while ignoring schema 8. Add explicit required schema/vector families and fail when zero cases are consumed. Pin only commits proving those assertions. This belongs in or alongside both landing tasks; neither current task description names it. | 3 |
| Immutable rc.9 release surface | The staged schema/vector branches update `release/1.0.0-rc.8.json`, while `v1.0.0-rc.8` and its metadata are published historical evidence. Revert that delta and add rc.9: new metadata, `PROTOCOL_VERSION`, README status, generator/regeneration lists, Makefile, CI/release workflow paths, release-gate required files/cases/digests/tests, CHANGELOG, and COMPATIBILITY. Decide claim handling explicitly; claim v4 is pinned to rc.8, so an rc.9 claim needs a new schema rather than rewriting v4. | 8 |
| Candidate wording and CocoaSkills candidate input | Curator's functional candidate gate is generic but comments/input text say schema v6. CocoaSkills is hard-coded to one rc.6 SHA/digest and has no `workflow_dispatch` candidate input. Generalize both descriptions and provide a fail-closed schema-8 candidate run that cannot alter released pins. | 3 |

### In-flight coverage confirmed

- `TASK-260822-1f533i` and `TASK-260822-3fkfmf` cover Protocol Core, manager profile, and SECURITY split for script-worker-v1.
- `TASK-260822-1mwy10` covers schema 8, legacy rejection, and marker v4; `TASK-260822-f4qv7w` covers script behavioral vectors and audit-label oracles.
- `TASK-260822-3nvx91` covers module-roots schema/core/manager prose; `TASK-260822-1so0ym` is assigned the module-roots vectors.
- `TASK-260822-c0rxj7` and `TASK-260822-10udu1` mention CHANGELOG/COMPATIBILITY and PR landing, but their current descriptions do not cover audit-record/registry versioning, implementation coverage/pin atomics, or the rc.9 release-tool migration above.

No additional manager-profile residual was found beyond coordinating the two schema-8 branches and their inventories. The uncovered profile is the Registry profile.

## Recommendation

Treat schema 8 as a coordinated release train, not two independent spec merges:

`candidate suite -> Curator/CocoaSkills/Registry candidate qualification -> landed implementation commits -> atomic spec pin+coverage+schema/vector PR -> rc.9 publication -> consumer SPEC_PIN/released-pin bumps`.

The existing board sequence saying implementation starts “once vectors land” must mean “once an immutable candidate vector commit exists”, not “once vectors are on main”. Otherwise the project either creates a false-green normative release or temporarily pairs main with implementations that do not consume its new requirements.

## Sources and fact-check evidence

- `curator-spec@be7861c:.github/workflows/implementations.yml` — exact three implementation pins and commands.
- `curator/.github/workflows/ci.yml:1-39, 295-391` and `curator/.github/ci/candidate-suite.sh:1-159` — `SPEC_PIN`, explicit candidate dispatch, identity checks, and `CI_REQUIRE_FULL_ROOT=1`.
- `curator-spec@be7861c:COMPATIBILITY.md:1-24, 146-153` and `RELEASE.md` — versioning, frozen historical releases, and release sequence.
- `curator-spec@be7861c:release/1.0.0-rc.8.json` and tag `v1.0.0-rc.8` — rc.8 identity and publication boundary.
- `curator-spec@be7861c:decisions/0008-enforced-script-capabilities.md` and `decisions/0009-first-party-module-roots.md` — accepted requirements.
- `spec/script-worker-v1-normative@a690d63:protocol/core.md`, `profiles/manager.md`, and active `TASK-260822-f4qv7w` vector worktree — current normative and vector shape.
- Active `TASK-260822-3nvx91` worktree — module-roots schema/core/profile shape and diagnostics.
- `cocoaskills/src/csk/skillspec.py:23-327`, `install_marker.py:29,640-770`, `installer.py:2912-3358`, `audit/detectors.py:503-540`, and `builds/go_v1.py:967-1027` — concrete implementation gaps and reuse points.
- CocoaSkills board `STORY-260822-2evh3p` and `STORY-260822-27ze8z`; Curator board tasks cited above — assigned versus residual scope.
- Direct validation logs under `.temp/TASK-260823-omp8zt/`; both gate commands ran standalone and returned their real exit code 0.
