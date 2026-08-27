# Audit of the compiled-language story dependency cycle

Board task: `TASK-260729-iuepxx` (`audit-language-story-dependency-cycle`)  
Scope: blocker graph for `EPIC-260720-21aq1i` only; no live dependency, parent, product, release, worktree, or implementation mutation was applied.

## Executive finding

The cycle is real at the **story projection** and absent from the underlying **task execution order**. The five stories form one strongly connected component with:

- 37 cross-story task `blockedBy` links;
- 14 derived story edges;
- seven reciprocal two-story cycles.

The root cause is phase mixing inside the story containers:

1. Rust, Swift, and Kotlin **design tasks** live in their language implementation stories.
2. Protocol integration correctly waits for those designs.
3. Toolchain implementation correctly waits for the qualified protocol.
4. Language implementation correctly waits for toolchain implementation.
5. Deferred Linux toolchain qualification correctly waits for the language qualification tasks.

For each language `L`, the mandatory phase edges therefore collapse to:

```text
protocol P --waits for--> language L design
language L --waits for--> toolchain T implementation
toolchain T --waits for--> protocol P qualification
```

That is `P → L → T → P` in the board's dependent-to-prerequisite direction. It is impossible to make the current five-story quotient acyclic using only `unlink`/`link` while retaining all three real orderings.

An exhaustive weighted feedback-edge search proves that the smallest purely acyclic unlink set has seven task links. That set is **not semantically acceptable**: it removes four hard specification prerequisites and three real Linux qualification-ordering links. It is included below as exact evidence, not as the recommendation.

The smallest phase-aligned repair found that preserves every one of the 37 blocker links is six `set_parent` mutations: move the three language design tasks and two toolchain wire-spec tasks into the protocol story, and move the shared Linux toolchain qualification task into the existing Linux qualification story. An isolated board-copy plan succeeds after those moves. This exceeds the assigned unlink/link scope and therefore requires orchestrator approval.

## Direction and aliases

An edge `A → B` means “A is blocked by B”; prerequisite execution runs in the reverse direction.

| Alias | Story |
|---|---|
| `K` | `STORY-260728-16spsm` (`kotlin-compiled-build-drivers`) |
| `S` | `STORY-260728-2abmzr` (`swift-compiled-build-drivers`) |
| `R` | `STORY-260728-3tymqm` (`rust-compiled-build-drivers`) |
| `T` | `STORY-260728-2fsqtv` (`compiled-build-toolchain-preflight`) |
| `P` | `STORY-260728-2mnlp0` (`additional-language-driver-protocol`) |
| `Q` | `STORY-260728-1eye8p` (`linux-native-external-build-validation`), used only in the structural alternative |

All five cycle stories have empty direct `blockedBy` arrays. Every edge below is a derived cross-container edge originating in task blockers. [S1]

## Exact cycle edges and originating links

Every one of the 14 edges participates in a reciprocal two-story cycle. The seven reciprocal pairs are `K↔P`, `S↔P`, `R↔P`, `K↔T`, `S↔T`, `R↔T`, and `P↔T`.

Semantic classification describes the rationale. Mechanically, every listed relation is currently a hard `blockedBy` link because the CLI has no advisory dependency type.

### Kotlin edges

| Story edge | Exact originating task blocker | Semantic class | Rationale/source |
|---|---|---|---|
| `K → P` | `TASK-260728-168smo` (`select-kotlin-artifact-model-and-driver-pair`) → `TASK-260728-2spy93` (`define-additional-driver-version-and-artifact-boundary`) | hard prerequisite | The Kotlin contract must select identities/artifact semantics within the accepted version boundary. [S2] |
| `K → P` | `TASK-260728-1koh5v` (`implement-local-kotlin-driver-in-curator`) → `TASK-260728-2bu2q6` (`qualify-additional-driver-spec-candidate`) | implementation ordering | Curator implementation consumes the qualified protocol candidate. [S2] |
| `K → P` | `TASK-260728-gmfxdg` (`implement-external-kotlin-driver-in-curator`) → `TASK-260728-2bu2q6` | implementation ordering | Same qualified-candidate gate for the external driver. [S2] |
| `K → T` | `TASK-260728-168smo` → `TASK-260728-1g0z69` (`define-toolchain-requirement-and-guidance-contract`) | hard prerequisite | The Kotlin decision must define the selected trusted compiler/runtime requirement against the shared contract. [S2] |
| `K → T` | `TASK-260728-1koh5v` → `TASK-260728-2gbtb9` (`implement-trusted-toolchain-preflight-in-curator`) | implementation ordering | The local Curator driver requires Curator preflight. [S2] |
| `K → T` | `TASK-260728-gmfxdg` → `TASK-260728-2gbtb9` | implementation ordering | The external Curator driver requires Curator preflight. [S2] |
| `K → T` | `TASK-260728-3ar1qp` (`implement-local-kotlin-driver-in-csk`) → `TASK-260728-1j72zq` (`implement-trusted-toolchain-preflight-in-csk`) | implementation ordering | CocoaSkills local parity requires its accepted preflight implementation. [S2] |
| `K → T` | `TASK-260728-1uj0bc` (`implement-external-kotlin-driver-in-csk`) → `TASK-260728-1j72zq` | implementation ordering | CocoaSkills external parity requires its accepted preflight implementation. [S2] |
| `K → T` | `TASK-260728-2uh7em` (`document-kotlin-driver-pair-authoring-and-operations`) → `TASK-260728-ypbuav` (`maintain-official-toolchain-install-guidance`) | advisory sequencing | Documentation should match reviewed official guidance; it does not gate the implementation itself. [S2] |
| `T → K` | `TASK-260728-1e6811` (`qualify-linux-toolchain-preflight`) → `TASK-260728-3u1nho` (`run-linux-kotlin-driver-pair-qualification`) | qualification ordering | Shared Linux preflight qualification is explicitly deferred until language-driver Linux qualification exists. [S3] |

### Swift edges

| Story edge | Exact originating task blocker | Semantic class | Rationale/source |
|---|---|---|---|
| `S → P` | `TASK-260728-1yhuqi` (`design-swift-driver-pair-security-contract`) → `TASK-260728-2spy93` | hard prerequisite | Swift identities and artifact rules are constrained by the accepted version boundary. [S4] |
| `S → P` | `TASK-260728-21x3yc` (`implement-swift-v1-local-builds-in-curator`) → `TASK-260728-2bu2q6` | implementation ordering | Local Curator implementation consumes the qualified protocol candidate. [S4] |
| `S → P` | `TASK-260728-2lnhci` (`implement-swift-repository-v1-in-curator`) → `TASK-260728-2bu2q6` | implementation ordering | External Curator implementation consumes the qualified protocol candidate. [S4] |
| `S → T` | `TASK-260728-1yhuqi` → `TASK-260728-1g0z69` | hard prerequisite | The Swift design needs the shared toolchain/SDK requirement contract. [S4] |
| `S → T` | `TASK-260728-21x3yc` → `TASK-260728-2gbtb9` | implementation ordering | Local Curator Swift requires Curator preflight. [S4] |
| `S → T` | `TASK-260728-2lnhci` → `TASK-260728-2gbtb9` | implementation ordering | External Curator Swift requires Curator preflight. [S4] |
| `S → T` | `TASK-260728-3j60e3` (`implement-swift-v1-local-builds-in-csk`) → `TASK-260728-1j72zq` | implementation ordering | CocoaSkills local parity requires its preflight. [S4] |
| `S → T` | `TASK-260728-2ztr3c` (`implement-swift-repository-v1-in-csk`) → `TASK-260728-1j72zq` | implementation ordering | CocoaSkills external parity requires its preflight. [S4] |
| `S → T` | `TASK-260728-1egim2` (`document-swift-driver-pair-authoring-and-operations`) → `TASK-260728-ypbuav` | advisory sequencing | Documentation should match the official guidance catalog. [S4] |
| `T → S` | `TASK-260728-1e6811` → `TASK-260728-1y8u4m` (`run-linux-swift-driver-pair-qualification`) | qualification ordering | Shared Linux preflight qualification is intentionally after language Linux qualification. [S3] |

### Rust edges

| Story edge | Exact originating task blocker | Semantic class | Rationale/source |
|---|---|---|---|
| `R → P` | `TASK-260728-12pnm1` (`design-rust-driver-pair-security-contract`) → `TASK-260728-2spy93` | hard prerequisite | Rust identities/artifact behavior must fit the accepted version boundary. [S5] |
| `R → P` | `TASK-260728-q283m8` (`implement-rust-v1-local-builds-in-curator`) → `TASK-260728-2bu2q6` | implementation ordering | Local Curator implementation consumes the qualified candidate. [S5] |
| `R → P` | `TASK-260728-13ioo0` (`implement-rust-repository-v1-in-curator`) → `TASK-260728-2bu2q6` | implementation ordering | External Curator implementation consumes the qualified candidate. [S5] |
| `R → T` | `TASK-260728-12pnm1` → `TASK-260728-1g0z69` | hard prerequisite | The Rust design needs the shared trusted-toolchain contract. [S5] |
| `R → T` | `TASK-260728-q283m8` → `TASK-260728-2gbtb9` | implementation ordering | Local Curator Rust requires Curator preflight. [S5] |
| `R → T` | `TASK-260728-13ioo0` → `TASK-260728-2gbtb9` | implementation ordering | External Curator Rust requires Curator preflight. [S5] |
| `R → T` | `TASK-260728-2yxdo7` (`implement-rust-v1-local-builds-in-csk`) → `TASK-260728-1j72zq` | implementation ordering | CocoaSkills local parity requires its preflight. [S5] |
| `R → T` | `TASK-260728-gjxj1v` (`implement-rust-repository-v1-in-csk`) → `TASK-260728-1j72zq` | implementation ordering | CocoaSkills external parity requires its preflight. [S5] |
| `R → T` | `TASK-260728-1t59zp` (`document-rust-driver-pair-authoring-and-operations`) → `TASK-260728-ypbuav` | advisory sequencing | Documentation should match the official guidance catalog. [S5] |
| `T → R` | `TASK-260728-1e6811` → `TASK-260728-26e3n2` (`run-linux-rust-driver-pair-qualification`) | qualification ordering | Shared Linux preflight qualification is intentionally after language Linux qualification. [S3] |

### Protocol/toolchain edges

| Story edge | Exact originating task blocker | Semantic class | Rationale/source |
|---|---|---|---|
| `P → K` | `TASK-260728-251p01` (`integrate-rust-swift-kotlin-wire-contracts`) → `TASK-260728-168smo` | hard prerequisite | Protocol integration admits only the selected, reviewed Kotlin contract. [S3] |
| `P → S` | `TASK-260728-251p01` → `TASK-260728-1yhuqi` | hard prerequisite | Protocol integration admits only the reviewed Swift contract. [S3] |
| `P → R` | `TASK-260728-251p01` → `TASK-260728-12pnm1` | hard prerequisite | Protocol integration admits only the reviewed Rust contract. [S3] |
| `P → T` | `TASK-260728-251p01` → `TASK-260728-2jaw7h` (`add-toolchain-requirements-to-curator-spec`) | hard prerequisite | Integration consumes the accepted toolchain wire placement and conformance contract. [S3] |
| `T → P` | `TASK-260728-2jaw7h` → `TASK-260728-2spy93` | hard prerequisite | The toolchain schema landing consumes the accepted protocol boundary. This prerequisite is already accepted, but remains structurally present. [S3] |
| `T → P` | `TASK-260728-2gbtb9` → `TASK-260728-2bu2q6` | implementation ordering | Curator preflight implementation consumes the independently qualified candidate. [S3] |
| `T → P` | `TASK-260728-1j72zq` → `TASK-260728-2bu2q6` | implementation ordering | CocoaSkills preflight also names the qualified candidate directly; this link is transitively redundant through `TASK-260728-2gbtb9` but is not the cause of the irreducible phase cycle. [S3] |

Counts reconcile mechanically: `3×(3 language→protocol) + 3×(6 language→toolchain) + 3 toolchain→language + 3 toolchain→protocol + 3 protocol→language + 1 protocol→toolchain = 37` source links and 14 distinct story edges. [V2]

## Why the task DAG is valid but the story graph is cyclic

For each language, the intended task execution is:

```text
accepted protocol boundary + toolchain contract
  → language design
  → protocol integration
  → protocol candidate qualification
  → Curator toolchain preflight
  → Curator language implementation
  → CocoaSkills preflight/parity
  → cross-manager and Linux qualification
```

This is acyclic. The quotient becomes cyclic because the first and fifth-plus phases are placed in the same language story, while the toolchain story similarly combines early wire specification, middle implementation, and late Linux qualification.

The irreducible proof does not rely on documentation or Linux advisory edges. Retain only:

- `P → L`: protocol integration waits for each accepted language design;
- `T → P`: Curator preflight implementation waits for protocol qualification;
- `L → T`: each language implementation waits for Curator preflight.

For each of `K`, `S`, and `R`, those three real edges form `P → L → T → P`. Therefore an acyclic link-only graph must discard at least one real specification or implementation ordering per language, or discard the shared `T → P` gate for all languages. [V2]

## Mathematically minimal unlink cut — acyclic but rejected

The verifier enumerated all `5! = 120` story orders, weighted each story edge by the number of originating task links that must be unlinked to remove it, and found a minimum weight of seven. The six language permutations tie; the cut edge set is unique:

```text
P→K, P→S, P→R, P→T, T→K, T→S, T→R
```

Exact dry-run-safe command:

```bash
task-board m --dry-run 'unlink(TASK-260728-251p01, blocked_by=TASK-260728-12pnm1); unlink(TASK-260728-251p01, blocked_by=TASK-260728-1yhuqi); unlink(TASK-260728-251p01, blocked_by=TASK-260728-168smo); unlink(TASK-260728-251p01, blocked_by=TASK-260728-2jaw7h); unlink(TASK-260728-1e6811, blocked_by=TASK-260728-26e3n2); unlink(TASK-260728-1e6811, blocked_by=TASK-260728-1y8u4m); unlink(TASK-260728-1e6811, blocked_by=TASK-260728-3u1nho)'
```

An isolated board copy with those seven unlinks produces a valid epic plan. Its prerequisite-first ordering is `P`, then `T`, then the three language stories. [V3]

This cut must **not** be applied:

- the four `TASK-260728-251p01` links are hard inputs to protocol integration;
- the three `TASK-260728-1e6811` links encode the stated deferred Linux qualification order;
- none of the seven is merely the documentation/guidance advisory edge;
- consequently the cut achieves acyclicity by weakening the task contract.

No compensating `link` can restore those orderings without restoring the same story directions and the cycle. The `link` mutation supports only hard `blocked_by`; there is no advisory dependency kind. [S6]

## Recommended phase-aligned board proposal

Do not mutate blocker links. Align the task parents with the phases instead:

```bash
task-board m --dry-run 'set_parent(TASK-260728-12pnm1, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1yhuqi, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-168smo, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1g0z69, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-2jaw7h, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1e6811, parent=STORY-260728-1eye8p)'
```

Rationale:

- `TASK-260728-12pnm1`, `TASK-260728-1yhuqi`, and `TASK-260728-168smo` are specification/design tasks and become siblings of protocol integration.
- `TASK-260728-1g0z69` and `TASK-260728-2jaw7h` own the shared toolchain wire contract and its curator-spec landing, so they belong with the protocol specification phase rather than manager implementation.
- `TASK-260728-1e6811` is an explicitly deferred Linux native qualification task and fits the existing Linux qualification container better than the implementation/preflight container.

The resulting focused story graph, with every existing blocker retained, is acyclic:

```text
Q → {K,S,R,T}
{K,S,R} → T
{K,S,R,T} → P
```

Read prerequisites first, it orders:

```text
P protocol/language/toolchain specifications
  → T Curator then CocoaSkills preflight implementations
  → K/S/R Curator implementations then CocoaSkills parity
  → Q deferred Linux qualification
```

This preserves the product priority:

- the language specifications are in `P` before manager implementations;
- the Curator tasks remain prerequisites of their CocoaSkills counterparts;
- the Rust/Swift/Kotlin implementation stories contain backlog implementation/qualification work after their design tasks move;
- shared Linux qualification remains last.

An isolated copy of the full board accepted all six moves and `plan(EPIC-260720-21aq1i, mode=children, active=true)` exited 0. No dependency was removed. [V4]

This is a structural proposal, not an in-scope unlink/link correction. Applying `set_parent` also recomputes parent statuses, and moving `TASK-260728-1e6811` may warrant broadening `STORY-260728-1eye8p` wording from rc.5 external-only Linux validation to shared compiled-build Linux qualification. The orchestrator should review those ownership/status side effects before applying anything.

## Exact recommendation to the orchestrator

1. **Reject the seven-unlink cut** despite its mathematical minimality.
2. **Apply no live link/unlink mutation** under this task.
3. Choose one of:
   - approve the six phase-aligned `set_parent` mutations above and review the Linux story wording/status side effects; or
   - change planning semantics to plan the task DAG before story projection / tolerate derived cross-container cycles caused solely by phase mixing.
4. Re-run the epic plan after the chosen structural correction.

The first option is board-only, preserves all current blocker evidence, and is already mechanically demonstrated on a board copy. The second is a task-board product change outside this audit.

## Verification record

| Validation | Exit | Result |
|---|---:|---|
| `task-board --help` | 0 | CLI readiness confirmed. |
| Live `plan(EPIC-260720-21aq1i, mode=children, active=true)` | 1 | Expected red: named exactly the five audited stories in a dependency cycle. [V1] |
| `.temp/TASK-260729-iuepxx/verify-cycle.js` | 0 | Verified 37 links, 14 story edges, current cyclicity, seven-link minimum, acyclicity of the seven-cut graph, impossibility of retaining the mandatory phase edges, and acyclicity of the phase-aligned proposal. [V2] |
| Seven-unlink batch on live board with `--dry-run` | 0 | All exact mutations parsed and resolved; no write. |
| Follow-up live blocker projection | 0 | All seven blockers remained present. |
| Seven-unlink batch on isolated board copy | 0 | Copy only; seven links removed there. |
| Epic plan on seven-cut isolated copy | 0 | Acyclic plan produced, proving the mathematical cut. [V3] |
| Six-`set_parent` batch on live board with `--dry-run` | 0 | All exact mutations parsed and resolved; no write. |
| Follow-up live parent projection | 0 | All six live parents remained unchanged. |
| Six-`set_parent` batch on isolated board copy | 0 | Copy only; all blockers retained. |
| Epic plan on phase-aligned isolated copy | 0 | Acyclic plan produced for the full epic. [V4] |
| `task-board validate` | 0 | Command exited 0 but reported 43 pre-existing board issues outside this task, including old broken links/status mismatches, one unsupported unrelated container link, and orphan resources. It did not report a target-scoped structural defect. |

No green claim is made for the live epic plan; its real exit code remains 1 until an approved correction is applied.

## Sources

- **[S1]** Compact board projections of the five stories: IDs, names, direct `blockedBy`, `derivedBlockedBy`, descriptions, scope, acceptance criteria, and child IDs; rechecked after all dry runs.
- **[S2]** Compact blocker and requirement projections for all Kotlin tasks under `STORY-260728-16spsm`.
- **[S3]** Compact blocker and requirement projections for toolchain tasks under `STORY-260728-2fsqtv` and protocol tasks under `STORY-260728-2mnlp0`, especially `TASK-260728-251p01`, `TASK-260728-2bu2q6`, `TASK-260728-2gbtb9`, `TASK-260728-1j72zq`, and `TASK-260728-1e6811`.
- **[S4]** Compact blocker and requirement projections for all Swift tasks under `STORY-260728-2abmzr`.
- **[S5]** Compact blocker and requirement projections for all Rust tasks under `STORY-260728-3tymqm`.
- **[S6]** `task-board q --format compact 'schema(mutation=link)'`: `link(id:string, blocked_by:string)` only; no advisory kind.
- **[V1]** Live plan command output: `dependency cycle detected involving: STORY-260728-16spsm, STORY-260728-2abmzr, STORY-260728-2fsqtv, STORY-260728-2mnlp0, STORY-260728-3tymqm`.
- **[V2]** Mechanical verifier: `.temp/TASK-260729-iuepxx/verify-cycle.js`.
- **[V3]** Isolated copy: `.temp/TASK-260729-iuepxx/board-seven-cut.k3WewS`.
- **[V4]** Isolated copy: `.temp/TASK-260729-iuepxx/board-reparent.xFjMXM`.

