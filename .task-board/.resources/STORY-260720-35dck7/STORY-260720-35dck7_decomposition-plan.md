# Protocol schema v6 decomposition

**Board story:** `STORY-260720-35dck7` — Protocol schema v6  
**Target repository:** `/Users/iv/Developer/ReluxWorks/curator-spec`  
**Verified baseline:** `origin/main` = `57c1f56846d221ecc55786bd3c2467ec32f11730` locally and remotely  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`

## Development plan

| Phase | Task | Single deliverable | Primary owned artifacts |
|---:|---|---|---|
| 1 | `TASK-260720-1nvomm` — Specify protocol v6 core and security contract | Normative protocol and security boundary | `protocol/core.md`, `SECURITY.md`, decision 0004 |
| 2 | `TASK-260720-17llva` — Specify the normative go-v1 manager lifecycle | Normative manager process and transaction model | `profiles/manager.md` |
| 2 | `TASK-260720-wajgn8` — Add canonical and legacy manifest schema v6 | Strict v6 manifest pair plus generated cases | common schema, agent/csk v6 schemas, generator cases |
| 3 | `TASK-260720-12iigs` — Add build receipt v1 and install marker v2 schemas | Strict compiled-artifact metadata plus generated cases | receipt v1, marker v2, shared definitions, generator cases |
| 4 | `TASK-260720-2zc6k1` — Add conformance claim v2 for protocol rc.4 | Explicit rc.3-to-rc.4 claim transition | claim v2, split generator constants, suite version assertion |
| 5 | `TASK-260720-37ei85` — Add legacy schema compatibility guards | Regression lock for historical wire contracts | generator tests for manifests 1–5, marker v1, claim v1 |
| 6 | `TASK-260720-1s1vr6` — Generate go-v1 build-driver conformance vectors | Independent fixture and build-driver vectors | separate Go fixture, expected identities/context, `build-drivers.json` |
| 7 | `TASK-260720-cw39jh` — Generate compiled-build manager lifecycle vectors | Shared lifecycle, concurrency, and rollback evidence | compiled-build additions to `manager-lifecycle.json` |
| 8 | `TASK-260720-1u7hes` — Enforce schema v6 validation and release gates | Fail-closed validator and release inventory | validator, release gate, negative gate tests |
| 8 | `TASK-260720-3lo9jc` — Document schema v6 authoring and CLI behavior | Author and implementer documentation | schema index, conformance guide, CLI guide |
| 9 | `TASK-260720-q5oy3o` — Publish protocol 1.0.0-rc.4 release metadata | Accurate compatibility and release metadata | README, compatibility, changelog, release checklist |
| 10 | `TASK-260720-3ag6pi` — Verify integrated protocol v6 conformance | Reproducible integration evidence | validation logs and AC/vector coverage matrix |

The canonical board plan contains 12 tasks in 10 phases. After phase 1, the manager profile and manifest-schema work can proceed in parallel. After lifecycle vectors stabilize, validation gates and documentation can proceed in parallel. Shared edits to `tools/generate-vectors/main.go` are intentionally serialized.

Critical path:

```text
TASK-260720-1nvomm
  -> TASK-260720-17llva
  -> TASK-260720-12iigs
  -> TASK-260720-2zc6k1
  -> TASK-260720-37ei85
  -> TASK-260720-1s1vr6
  -> TASK-260720-cw39jh
  -> TASK-260720-1u7hes
  -> TASK-260720-q5oy3o
  -> TASK-260720-3ag6pi
```

## Completeness matrix

| Story requirement | Owning tasks |
|---|---|
| Strict schema 6 declarations for canonical and legacy names | `TASK-260720-wajgn8`, `TASK-260720-37ei85` |
| No package shell, arbitrary argv/environment/output, hooks, or plugins | `TASK-260720-1nvomm`, `TASK-260720-17llva`, `TASK-260720-1s1vr6` |
| Normative fixed Go driver and toolchain semantics | `TASK-260720-17llva`, `TASK-260720-1s1vr6` |
| Build sources excluded from agent context and runtime copying | `TASK-260720-1nvomm`, `TASK-260720-wajgn8`, `TASK-260720-1s1vr6` |
| Audit-before-build and compiler-free dry-run | `TASK-260720-17llva`, `TASK-260720-cw39jh` |
| Build-source, cache, receipt, marker, currentness, repair, rollback, and GC | `TASK-260720-1nvomm`, `TASK-260720-17llva`, `TASK-260720-12iigs`, `TASK-260720-cw39jh` |
| Preserve manifest schemas 1–5, marker v1, and claim v1 | `TASK-260720-wajgn8`, `TASK-260720-12iigs`, `TASK-260720-2zc6k1`, `TASK-260720-37ei85` |
| Positive fixture and all key rejection vectors | `TASK-260720-1s1vr6`, `TASK-260720-cw39jh` |
| Deterministic vector generator and schema cases | `TASK-260720-wajgn8`, `TASK-260720-12iigs`, `TASK-260720-2zc6k1`, `TASK-260720-1s1vr6`, `TASK-260720-cw39jh` |
| Validation and release gates | `TASK-260720-1u7hes`, `TASK-260720-3ag6pi` |
| Decision, documentation, compatibility, security, changelog, and version metadata | `TASK-260720-1nvomm`, `TASK-260720-3lo9jc`, `TASK-260720-q5oy3o` |
| Protocol rc.4 conformance claim transition | `TASK-260720-2zc6k1`, `TASK-260720-1u7hes`, `TASK-260720-q5oy3o` |

## Gap and risk audit

- No unresolved product or architecture choice remains: the accepted research contract fixes schema 6, `go-v1`, protocol `1.0.0-rc.4`, marker v2, receipt v1, claim v2, static build-root exclusion, protected cache provenance, and manager-home transaction isolation.
- No new blocking research or clarification task is required. The upstream security story is accepted and already `done`.
- Generator tasks are serialized because they share `tools/generate-vectors/main.go`; documentation and validation are parallel because their owned files do not overlap materially.
- Schema tasks generate their own cases so `make validate` remains green at each handoff rather than relying on a later task to repair an intentionally broken intermediate state.
- Release metadata must not pin unreleased implementation commits or claim nonexistent reviews, tags, checksums, signatures, or attestations.
- Existing unrelated board validation anomalies reported by the accepted research task remain out of scope: 12 legacy broken dependency references and one orphan resource do not belong to this story.
- Local diagram tooling anomaly: system PlantUML is absent and Homebrew Graphviz is missing `libltdl.7.dylib`. A task-local PlantUML 1.2026.6 JAR with Smetana validated both diagram sources and rendered the linked SVG/PNG artifacts; this is not a product blocker.

## Planning diagrams

- `diagrams/plantuml/component/STORY-260720-35dck7_artifact-map.puml` — task-to-artifact ownership and dependency flow.
- `diagrams/plantuml/activity/STORY-260720-35dck7_install-lifecycle.puml` — normative install and dry-run ordering.
- Rendered SVG and PNG files are under `diagrams/artefacts/plantuml/`.

## Integration gates

The final verification task must run and attach evidence for:

```bash
make validate
make regenerate
make regenerate-check
make release-check VERSION=1.0.0-rc.4
```

It must also perform a second regeneration with byte-identical `conformance/v1` output and map every story acceptance criterion plus every minimum rejection cluster to a schema case or vector.
