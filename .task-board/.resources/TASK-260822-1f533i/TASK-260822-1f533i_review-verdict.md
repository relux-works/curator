# TASK-260822-1f533i — review verdict

**Verdict:** changes requested → `to-dev`.
**Reviewed:** `spec/sw-core-prose` @ `41cf556` (= `origin/spec/sw-core-prose`), `protocol/core.md` +388/-2,
parent `3dc9ca6` = `origin/main`. Worktree
`curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree`.
**Siblings at review time:** `origin/spec/sw-manager-security` = `c2371d3`, `origin/spec/sw-schema` = `ebfed81`.

One blocking defect. The prose itself is strong and the surrounding process is clean — the fix is
one table cell plus two wording changes, not a rewrite.

## What holds

- Committed and pushed to `spec/sw-core-prose`, one clean commit on `origin/main`, no PR, no AI
  attribution, signed. Branch/DoD reconciliation is explained in notes and logbook.
- Every DoD structural element is present in 4.1.1: opt-in surface, skill-wide capability note,
  fixed process graph, interpreter identity and resolution, launcher carve-out, deny-by-default
  derivation of all six section 4.3 values, 11-item mandatory portable set, mechanism-vs-deferred
  table, three-platform inventory, `script-capability-evidence-v1`, single failure boundary,
  diagnostics, audit warning classes, section 12.3 admission rule.
- The 4.2.1 mirror is faithful: manager mechanism and kernel guarantee are named separately
  throughout, the seven deferred `script-` guarantees are disjoint from 4.2.1's frozen six in both
  directions, and the three deliberate divergences (stdin open, streaming output, no policy
  deadline) are written as positive controls rather than silently dropped.
- The three-sentence edit to 4.3 is the right call: it keeps "audit surface, not a runtime sandbox"
  true by scoping it to commands that have not opted in, instead of deleting it.
- Identifier reconciliation with the sibling is **resolved**. `c2371d3` adopted core.md's spellings.
  Verified by extraction: deferred-guarantee names, warning classes, inventory version, record
  version, availability values, and unavailable-reason vocabulary are identical across
  `core.md@41cf556` and `manager.md@c2371d3`. The stale name list in the board coordination note no
  longer describes either file.
- Section-number handoff is satisfied: `manager.md@c2371d3` already cites section 4.1.1 in five
  places and `SECURITY.md` in one. Nothing left for the landing task there.
- Outcome resource and logbook entry are thorough, and the forward reference to the not-yet-existing
  `conformance/v1/vectors/script-host-execution-policy.json` is disclosed rather than hidden.

## Gates — re-run by the reviewer at `41cf556`, not taken on trust

| Gate | Exit | Note |
|---|---|---|
| `python tools/validate.py` | 0 | 49 schemas, 471 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | |
| `go test ./tools/...` | 0 | |
| `go run ./tools/generate-vectors -root .` | 0 | |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | 0 | deterministic regeneration |
| `git diff --check` (formatting job) | 0 | |
| `lychee --offline '**/*.md'` (links job) | 0 | 36 OK, 0 errors, 6 excluded |
| `gofmt -l tools` | n/a | no Go file changed |

The AC clause "formatting and link gates pass" is independently confirmed.

## Blocking — 1. The Linux `active-process-count-limit` cell mandates a false `applied` claim

`protocol/core.md` (script native-control inventory, `active-process-count-limit` row):

```
| `active-process-count-limit` | unavailable: `no-private-aggregate-domain` | available: `RLIMIT_NPROC` | available: Job Object active-process limit |
```

This is wrong on the merits and it is not a recordable style preference, because 4.1.1 itself says
in the paragraph below the table:

> An `available` control MUST report `applied` [...]

So every conforming Linux manager is **required** to emit
`active-process-count-limit / available / applied` in each `script-capability-evidence-v1` record,
on the strength of a primitive that does not do what the control name says.

Why `RLIMIT_NPROC` cannot back this control:

1. `RLIMIT_NPROC` bounds processes/threads for the **real UID of the calling process**, shared
   across every process that UID already owns. It is not a domain private to the invocation. An
   unrelated process of the same user consumes the budget, and conversely the enforced command's
   descendants are not privately bounded — which is the entire content of
   `no-private-aggregate-domain`.
2. The same table's macOS cell marks the same control `unavailable: no-private-aggregate-domain`,
   and macOS also has `RLIMIT_NPROC` with the same per-UID semantics. The table therefore applies
   two different standards to one primitive, in one row.
3. It is incoherent with the row directly beneath it. `aggregate-memory-limit` on Linux is
   `host-conditional: delegated cgroup v2 memory.max` — `RLIMIT_AS`/`RLIMIT_DATA` were correctly
   **not** claimed as an aggregate memory limit. If cgroup v2 delegation is host-conditional for the
   memory controller, the `pids` controller is host-conditional by exactly the same argument, and
   an rlimit is no more sufficient for process count than it was for memory.
4. It breaks the analysis's own stated honesty rule. `TASK-260822-1l4r4f_analysis.md` line ~792:
   "both are mapped to `unavailable`/`host-conditional`, so neither can produce a false `applied`
   claim." This cell is the one place that rule was not applied.
5. It collides with this policy's own deferred guarantee. Hard aggregate descendant bounds are
   deferred as `script-hard-aggregate-descendant-resource-bounds`; an `applied` process-count claim
   asserts a piece of exactly that.

This was not missed silently — it was handed to this task explicitly. The `TASK-260822-3fkfmf`
coordination note on the board says, verbatim, under "Two things you own that 3fkfmf could not":

> (2) The inventory Linux cell for `active-process-count-limit`. The analysis says
> `available: RLIMIT_NPROC`. That looks wrong [...] An available entry MUST report applied, so the
> cell as written would produce a false applied claim. 3fkfmf mirrored it rather than diverging
> unilaterally in one of three copies. Recommendation: `host-conditional: delegated cgroup v2
> pids.max`, changed in core.md, manager.md, and the vectors together.

It was neither applied nor recorded as a justified deviation: the cell is unchanged, and neither the
outcome resource's "Calls made beyond the analysis" / "Divergences" sections nor the LOGBOOK entry
mentions it. The AC line "no false hardened claims" is the clause this fails.

**Required fix.** Change the Linux cell to:

```
| `active-process-count-limit` | unavailable: `no-private-aggregate-domain` | host-conditional: delegated cgroup v2 `pids.max` | available: Job Object active-process limit |
```

`host-conditional` already exists in this inventory with the exact semantics needed ("the platform
MAY provide the control and the per-invocation probe decides"), and 4.1.1 already states that a
`host-conditional` control that probes unavailable MUST NOT reject the invocation. No new machinery.

The same cell must change in lockstep in `profiles/manager.md@c2371d3` line 710 (identical row) and
in `conformance/v1/vectors/script-host-execution-policy.json` when `TASK-260822-f4qv7w` writes it.
If the fix is not made here, coordinate it explicitly — do not leave three copies of a cell that the
sibling already flagged as wrong. If instead there is a defensible argument for keeping
`available: RLIMIT_NPROC`, it must be written into the divergence list with the reason, not
inherited from the analysis in silence.

## Should fix in the same pass — 2. "The complete diagnostic set" contradicts the sibling profile

`protocol/core.md`:

> The complete diagnostic set of this policy is:

followed by four rows. `profiles/manager.md@c2371d3` line ~822 defines **seven** `script_execution_*`
codes and pre-emptively reconciles the gap:

> The first four are the policy-level set of Protocol Core section 4.1.1. The last three are
> worker-session codes of this profile [...]

Section 4.2.1, the section this one is required to mirror, makes no completeness claim at all:
`core.md` names three `build_execution_*` codes and `manager.md` adds
`build_execution_worker_identity_invalid`, `build_execution_worker_protocol_invalid`, and
`build_execution_package_influence_forbidden` on top. The word "complete" is a divergence from the
mirror that manufactures a reader-level contradiction the moment the two branches land together.

**Fix:** "The policy-level diagnostic set of this policy is:" — matching the sibling's own wording.
One word.

## Should fix in the same pass — 3. 4.1.1 never names the manifest fields it normalizes

The opening paragraph says "Selection is an OPTIONAL per-command field on a script command of
manifest schema 8 or later" and the interpreter paragraph says "An enforced script command names a
closed interpreter identifier" — neither field is named anywhere in the subsection. The wire names
exist in `schemas/v1/common.schema.json@ebfed81` (`execution_policy`, `interpreter`, both fixed by
`dependentRequired` in each direction) and in `manager.md@c2371d3`, but not in the normative core.

Section 4.1 immediately above names `unix_path` and `win_path`; section 4.3 names all six capability
fields. Protocol core is a document that names manifest fields, so leaving these two to a profile
and a schema is out of pattern, and it means the section defining the opt-in surface cannot be read
standalone. The `dependentRequired` co-requirement (opt-in without an interpreter, or an interpreter
without opt-in, is invalid) also has no normative statement in core.md.

**Fix:** name `execution_policy` and `interpreter` in the opening and interpreter paragraphs, and
state the co-requirement in one sentence.

## Nit — 4. Wrong prefix in the `network-isolation-domain` non-aliasing sentence

> `network-isolation-domain` is not, and MUST NOT be spelled as, `total-network-denial`.

`total-network-denial` is section 4.2.1's build-policy guarantee. This policy's deferred guarantee is
`script-total-network-denial`, and that is the name a script surface would plausibly be tempted to
misuse. Name both, or name the `script-` one.

## Not blocking, but the story needs it — schema 8 is unreachable from section 4

`4.1.1` says "manifest schema 8 or later", and `spec/sw-schema@ebfed81` adds
`agent-skill-v8.schema.json` and `csk-skill-v8.schema.json`. But `core.md` section 4 still says
manifests conform to "`agent-skill-v1.schema.json` through `agent-skill-v7.schema.json`", its
"Schema | Added behavior" table stops at row 7, and the downward version-gate paragraph says
"schemas 2 through 7 reject unknown fields" with no schema-8 rule. `spec/sw-schema` touches no
`.md` outside `schemas/v1/README.md`, so no branch in this story closes it.

This reads as `TASK-260822-1mwy10`'s scope ("parse and validation rules **in prose**"), which is in
`reviewing` now — so it is raised here as coordination, not charged against this task. It does need
an owner before `TASK-260822-c0rxj7` merges, or the merged core.md will reference a manifest schema
its own section 4 does not admit.

## Route

`to-dev` for items 1–4 on `spec/sw-core-prose`, then another reviewer cycle. Item 1 alone is the
blocking one; 2–4 are cheap and belong in the same commit. Nothing here is a stop-the-line boundary:
the fix is local, the mechanism (`host-conditional`) already exists in the text, and no external or
human-only decision is required.

No acceptance evidence is handed to a commit-owning mover, because this review does not accept.
