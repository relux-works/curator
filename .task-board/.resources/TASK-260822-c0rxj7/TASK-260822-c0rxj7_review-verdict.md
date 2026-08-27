# TASK-260822-c0rxj7 — reviewer verdict: CHANGES REQUESTED (-> to-dev)

Reviewed independently against the re-scoped deliverable (steps 1-3 of the
TASK-260823-omp8zt landing order): one immutable schema-8 rc.9 candidate outside
main, its recorded identity, and its qualification through the candidate lane.
No merge to main was expected and none happened, which is correct.

## What is genuinely delivered and independently verified

| Claim | Verified how | Result |
| --- | --- | --- |
| Candidate is immutable and pushed | `git rev-parse origin/candidate/schema-8-rc.9` | `edd07210d4f3db34fd60238cb14b90f837de03cb` |
| Combines both families | `git merge-base --is-ancestor` for `dd9c9fc`, `bac193c`, `ebfed81`, `41cf556`, `e5df43d`, `b92b105` | all YES |
| `release/1.0.0-rc.8.json` byte-identical to main | `git show origin/main:… \| shasum -a 256` vs `git show edd0721:…` | both `293f101d…e31ede` |
| rc.9 metadata added, live pin moved | diff of `ci.yml`, `release.yml`, `Makefile`, `README.md`, `protocol/assurance.md`, `tools/assurance.py`, `decisions/0007` | rc.8 -> rc.9; rc.9 added to the immutability guard list |
| Double regeneration byte-identical | `cmp` of the two 693-line checksum inventories | identical; `conformance/v1/manifest.json` = `803918bf…b44403` |
| Identity recorded | `TASK-260822-c0rxj7_candidate-suite-identity.txt` | SHA, `sha256:803918bf…`, tree `sha256:9d5a10b6…`, 692 files, explicit non-release / non-conformance disclaimer |
| Specification CI green on the exact SHA | run `32642316308` | success; provenance, formatting, links, Specification on ubuntu/macos/windows |
| Candidate lane green on the exact SHA + digest | run `32651139699`, job logs `97222819641` / `97222819694` | `manifest digest matches the supplied expectation` on ubuntu and windows; Candidate suite success on all three OSes; `SPEC_PIN` still `00b1688…` |
| Green run used landed curator code | `f073aea` (its head) is an ancestor of curator `origin/main` | YES |
| Lockstep inventory cell (checklist 1) | `protocol/core.md:434` and `profiles/manager.md:781` | both `host-conditional: delegated cgroup v2 pids.max` |
| Green evidence routed to the two vector tasks | notes on `TASK-260822-f4qv7w` and `TASK-260822-1so0ym` | present, naming `edd0721` / `803918bf` / run `32651139699` |

The stop-the-line calls made earlier in this task (refusing to weaken the gate,
to supply a false digest, or to cherry-pick unqualified implementation code) were
correct and are upheld.

## Why this is not accepted

### 1. CONFIRMED — checklist item 2 delta (d) is checked done but is not delivered anywhere (blocking)

`schemas/v1/install-marker-v4.schema.json` ships on the candidate with
`skill_schema_version: {"const": 8}`, the suite ships the writer golden
`conformance/v1/expected/install-marker-v4.json`, and the schema-case family
`conformance/v1/schema-cases/install-marker-v4/` is in the manifest.

`protocol/core.md` never mentions marker v4. Section 10's only obligation
sentence is `protocol/core.md:1691-1694`:

> Managers supporting schema 7 MUST read marker schemas 1, 2, and 3. They MUST
> write marker schema 2 for schema 1 through 6 installation mutations and marker
> schema 3 for schema 7 installation mutations.

Searched the whole candidate tree: `marker.?v4|marker schema(s)? .*4|install-marker-v4`
outside `conformance/` hits only `CHANGELOG.md:12`, `COMPATIBILITY.md:158`, and
`schemas/v1/README.md:97`. `profiles/manager.md` has nothing.
`schemas/v1/README.md:97` states only what marker v4 *is* structurally; it states
no manager read/write obligation. That is a semantic manager-lifecycle rule, and
by that file's own opening line semantic rules belong in the protocol documents.

Failure scenario: a manager implements schema 8, writes marker v4 for a schema-8
install (the only thing `schemas/v1/README.md` implies), then reads it back under
section 10, which obliges it to read schemas 1, 2, and 3 only. Section 10's own
rule "Unsupported or unreadable markers are not current" (`protocol/core.md:1773`)
then makes every schema-8 installation report not-current immediately after a
successful install — and the manager is still conforming, because no MUST says
otherwise. The only thing forcing the correct behaviour is a conformance vector,
which `README.md:55-58` explicitly says is not how this spec assigns semantics.

This is exactly the "main would contradict schema 8" class the checklist item was
written to prevent, and it is the item's own listed delta (d).

### 2. CONFIRMED — checklist item 2 delta (c), second half, is in the wrong document (medium)

`protocol/core.md:187-191` carries the downward-gate list and states the gate for
`build_roots` (1-5), `build_repositories`/`repository`/`target`/`go-repository-v1`
(1-6), and `modules` at the top level and on every command (1-7). It has no
sentence for `execution_policy`/`interpreter`. The rule exists only at
`schemas/v1/README.md:89-93`.

This matters because of the schema-1 carve-out that the same paragraph states:
"Schema 1 preserves its deployed extension behavior; schemas 2 through 8 reject
unknown fields." So for schema 1 the field is not rejected structurally, and
`protocol/core.md:214` ("Schema 7 and earlier script commands keep their exact
meaning above") is descriptive, not a MUST. A reader of core.md alone cannot tell
whether a schema-1 manifest carrying `execution_policy` must be rejected or
silently ignored. `tools/validate.py:255-286` and `validate_wire_semantics`
(`tools/validate.py:568-582`) do reject it, top-level and per-command, for both
`agent-skill` and `csk-skill` v1-v7, and the schema-cases cover it — so behaviour
is right and only the normative prose is asymmetric with its three siblings.

### 3. CONFIRMED — the task's primary outcome artifact contradicts what was delivered (medium)

`TASK-260822-c0rxj7_results.md` still documents the superseded candidate
`859727b1…` with its old manifest digest `782d6868…`, a red candidate matrix, and
a "Handoff" section listing blockers that are now cleared. Its sibling in the same
resource folder, `TASK-260822-c0rxj7_candidate-suite-identity.txt`, records
`edd0721` / `803918bf…`. No task-scoped artifact on this task records the green
qualification; that evidence lives only on `TASK-260823-1l1p8q`. Anyone opening
this task's results reads "blocked, candidate is 859727b".

## Rework packet

1. `protocol/core.md` section 10: extend the obligation to schema 8 — managers
   supporting schema 8 MUST read marker schemas 1, 2, 3, and 4, and MUST write
   marker schema 4 for schema-8 installation mutations; add the marker-v4
   paragraph (permits `skill_schema_version` 8, otherwise carries marker-v3
   meaning unchanged). Keep the existing schema-7 sentence intact.
2. `protocol/core.md` section 4 version gates: add the missing sentence, phrased
   like its `modules` sibling — schemas 1 through 7 MUST reject `execution_policy`
   and `interpreter` at the top level and on every command.
3. `conformance/README.md:61-65`: the install-marker bullet stops at the marker-v2
   writer golden while the suite now ships `expected/install-marker-v4.json`.
   Bring it in line.
4. Rewrite `TASK-260822-c0rxj7_results.md` to the delivered state (`edd0721`,
   `803918bf…`, `9d5a10b6…`, 692 files, spec CI `32642316308`, candidate lane
   `32651139699`), keeping the superseded `859727b1…` identity recorded as history
   rather than as the candidate.

### Requalification is one CI loop, not another fix loop

Items 1-3 are prose-only and live outside `conformance/v1`. The suite manifest
digest stays `803918bf…` — no vector regeneration, no digest change. Re-cut the
immutable candidate with those commits on top, re-run Specification CI on the new
SHA, and re-dispatch the candidate lane with the **new SHA and the unchanged
digest**. Curator `main` already carries all seven qualification fixes
(`f073aea`/`e17b0f1`), so no implementation work is expected.

While re-cutting, also fold in curator-spec `origin/main` commit `517a130`
("State that operator credential selections are never lockable", `profiles/manager.md`
only) — it landed ~12 minutes before the candidate CI and is the single main
commit absent from `edd0721`. Prose-only, no suite-identity impact, but the
landing PR would otherwise have to resolve it late.

Minor, non-blocking: `schemas/v1/README.md:7` still opens "Rc.8 carries …" while
the body below now documents the rc.9 additions.

## Not defects

- `TASK-260822-f4qv7w` and `TASK-260822-1so0ym` are still `blocked` even though
  checklist item 7 is checked. The green evidence *is* on both tasks; the board
  refuses the transition because the dependency edge points at this task. That
  releases when this task reaches `done`. Mechanics, not a delivery gap.
- Branch and worktree are preserved rather than removed. The original AC said to
  clean up; the re-scope reversed that, and preserving them is correct.
- No "Implementation conformance" workflow run on the candidate branch. That
  workflow validates against the pinned suite; the candidate is qualified through
  the curator candidate lane instead, which is the designed path.
