# Protocol schema v6 decomposition audit

**Story:** `STORY-260720-35dck7` — Protocol schema v6  
**Audit date:** 2026-07-20  
**Target:** `/Users/iv/Developer/ReluxWorks/curator-spec` at fetched local and remote `origin/main` `57c1f56846d221ecc55786bd3c2467ec32f11730`  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`

## Audit result

The existing decomposition remains 12 atomic tasks in 10 dependency phases. No task was added, removed, merged, or duplicated. Every task has a concrete title, description, owned scope, executable acceptance criteria, and task-specific checklist. The dependency graph is acyclic, has one unblocked entry task, preserves two deliberate parallel phases, serializes shared generator ownership, and ends in one integrated verification gate.

Two evidenced brief defects were corrected in place:

1. `TASK-260720-1s1vr6` — Generate go-v1 build-driver conformance vectors now names the byte-level `curator-build-source-v1` and `curator-go-toolchain-v1` edge cases plus the legacy NUL-stream structural-collision regression required by the accepted contract. Its new task-scoped vector coverage plan makes those cases executable without requiring the developer to reconstruct them from research prose.
2. `TASK-260720-37ei85` — Add legacy schema compatibility guards now explicitly protects the fixed manager/system configuration surfaces and the registry/audit source-evidence boundary from accidental build-policy or receipt-provenance expansion. Its new task-scoped guard plan distinguishes legitimate rc.4 inventory/hash changes from frozen legacy wire semantics.

The accepted contract was also attached directly to `TASK-260720-1nvomm` — Specify protocol v6 core and security contract as a verified precondition, so the first implementation owner receives the authoritative input automatically.

## Atomic ownership audit

| Task | Single deliverable | Primary ownership |
|---|---|---|
| `TASK-260720-1nvomm` — Specify protocol v6 core and security contract | Normative protocol/security decision | `protocol/core.md`, `SECURITY.md`, decision 0004 |
| `TASK-260720-17llva` — Specify the normative go-v1 manager lifecycle | Normative execution and transaction profile | `profiles/manager.md` |
| `TASK-260720-wajgn8` — Add canonical and legacy manifest schema v6 | Strict manifest v6 pair and its generated cases | common manifest definitions, both v6 schemas, generator cases |
| `TASK-260720-12iigs` — Add build receipt v1 and install marker v2 schemas | Portable compiled-artifact metadata and cases | receipt v1, marker v2, shared definitions, generator cases |
| `TASK-260720-2zc6k1` — Add conformance claim v2 for protocol rc.4 | Explicit claim-version transition and suite identity | claim v2, split generator constants, rc.4 validator assertion |
| `TASK-260720-37ei85` — Add legacy schema compatibility guards | Frozen-contract regression guard | generator tests and baseline comparison only |
| `TASK-260720-1s1vr6` — Generate go-v1 build-driver conformance vectors | Go fixture and build-driver evidence | separate fixture, expected identities/context, `build-drivers.json` |
| `TASK-260720-cw39jh` — Generate compiled-build manager lifecycle vectors | Shared lifecycle/concurrency evidence | compiled-build additions to `manager-lifecycle.json` |
| `TASK-260720-1u7hes` — Enforce schema v6 validation and release gates | Fail-closed inventory and release enforcement | validator, release gate, negative gate tests |
| `TASK-260720-3lo9jc` — Document schema v6 authoring and CLI behavior | Author/implementer documentation | schema index, conformance guide, CLI guide |
| `TASK-260720-q5oy3o` — Publish protocol 1.0.0-rc.4 release metadata | Accurate compatibility and release metadata | top-level README, compatibility, changelog, release checklist |
| `TASK-260720-3ag6pi` — Verify integrated protocol v6 conformance | Reproducible integrated evidence | logs and AC/rejection coverage matrix; no feature design |

Generator edits are intentionally serialized through manifest schemas, compiled-artifact schemas, claim transition, compatibility guards, build-driver vectors, and lifecycle vectors. The manager-profile task and manifest-schema task can run in parallel because their owned files do not overlap. Validation and documentation can run in parallel after vector names stabilize.

## Accepted-contract coverage

| Contract surface | Owning tasks |
|---|---|
| Closed schema 6 build declaration for canonical and legacy names | `TASK-260720-wajgn8`, `TASK-260720-37ei85` |
| No shell, package argv/environment/output, hooks, plugins, generators, or unsafe build frontends | `TASK-260720-1nvomm`, `TASK-260720-17llva`, `TASK-260720-1s1vr6` |
| Fixed Go toolchain, environment, process graph, dependency inspection, and internal linking | `TASK-260720-17llva`, `TASK-260720-1s1vr6` |
| Static build-root exclusion from context and runtime on real, hit, and dry-run paths | `TASK-260720-1nvomm`, `TASK-260720-wajgn8`, `TASK-260720-1s1vr6` |
| Byte-exact build-source, toolchain, CCJ-1 input, cache key, receipt, artifact, and marker identities | `TASK-260720-1nvomm`, `TASK-260720-12iigs`, `TASK-260720-1s1vr6` |
| Protected cache provenance boundary, currentness, status, repair, and GC | `TASK-260720-1nvomm`, `TASK-260720-17llva`, `TASK-260720-12iigs`, `TASK-260720-cw39jh` |
| Audit-before-build, compiler-free dry-run, private staging, serialized publication, consumer-last commit, rollback, and recovery | `TASK-260720-17llva`, `TASK-260720-cw39jh` |
| Manifest schemas 1–5, marker v1, claim v1, config surfaces, and registry/audit compatibility | `TASK-260720-wajgn8`, `TASK-260720-12iigs`, `TASK-260720-2zc6k1`, `TASK-260720-37ei85` |
| Positive fixture and every minimum build-driver/lifecycle rejection cluster | `TASK-260720-1s1vr6`, `TASK-260720-cw39jh` |
| Protocol rc.4 claim transition and deterministic generated suite identity | `TASK-260720-2zc6k1`, `TASK-260720-1u7hes`, `TASK-260720-q5oy3o` |
| Decision record, authoring/CLI docs, security/compatibility impact, changelog, and release metadata | `TASK-260720-1nvomm`, `TASK-260720-3lo9jc`, `TASK-260720-q5oy3o` |
| Fail-closed validation, deterministic regeneration, release check, and integrated coverage matrix | `TASK-260720-1u7hes`, `TASK-260720-3ag6pi` |

No new research or clarification task is justified: the upstream security story is accepted, its authoritative outcome hash matches, and it fixes schema version, driver semantics, cache/receipt model, install ordering, compatibility transition, conformance identity, and release version. There is no remaining human-only product or architecture decision.

## Dependency audit

The canonical board plan remains:

```text
TASK-260720-1nvomm
  -> {TASK-260720-17llva, TASK-260720-wajgn8}
  -> TASK-260720-12iigs
  -> TASK-260720-2zc6k1
  -> TASK-260720-37ei85
  -> TASK-260720-1s1vr6
  -> TASK-260720-cw39jh
  -> {TASK-260720-1u7hes, TASK-260720-3lo9jc}
  -> TASK-260720-q5oy3o
  -> TASK-260720-3ag6pi
```

The existing cross-story links are also correct: lifecycle vectors unblock the shared interop cases; specification documentation unblocks the manager/interop authoring tasks; and integrated protocol verification unblocks the first schema/identity/locking consumers plus release qualification. No dependency was added or removed during this audit.

## Planning resources

- `TASK-260720-1s1vr6_vector-coverage-plan.md` — exact positive, negative, and byte-level vector obligations.
- `TASK-260720-37ei85_compatibility-guard-plan.md` — frozen compatibility surfaces and baseline comparison rules.
- `TASK-260720-3ag6pi_verification-gates.puml` and rendered SVG — focused four-gate activity model for integrated verification.
- `.planning/260720_065107_story-260720-35dck7.md` — canonical task-board phase snapshot.

## Verification record

- The accepted research outcome hash matches the accepted SHA-256.
- A fresh fetch and `git ls-remote` both resolve `curator-spec origin/main` to `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- PlantUML 1.2026.6 syntax validation and Smetana SVG/PNG rendering pass; the rendered PNG was visually inspected and is readable.
- Homebrew Graphviz remains unavailable because `libltdl.7.dylib` is missing; the verified PlantUML/Smetana path avoids that local tooling defect.
- Post-mutation JSON checks report 12 tasks, no missing title/description/scope/AC/checklist field, exactly one entry task, and the expected new task-scoped resources.
- The canonical plan query reports 12 tasks, 10 phases, the expected two parallel phases, and the unchanged 10-task critical path without a cycle.
- `task-board validate` exits 0 and reports only the same 13 pre-existing unrelated findings: 12 missing `EPIC-260712-*` dependency references and one orphan `TASK-260713-7a9c1e/review.md` resource. No finding references this story, its tasks, or its resources.
- No product or specification implementation file was changed.
