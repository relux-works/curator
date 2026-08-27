# TASK-260728-2spy93 — rework 01, additional-driver version and artifact boundary

Developer rework after the independent review requested changes. Status: ready
for review.

## Execution base

- Task worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2spy93/curator-spec-worktree`,
  unchanged since the first cycle, still branched at `57c1f56` and carrying the
  uncommitted rc.5 candidate.
- Versus the accepted `TASK-260728-2kp3tv` base, `diff -rq --exclude=.git` reports
  exactly three differences: the new decision record and the two tool files. No
  schema, conformance case, vector, fixture, release file, profile, or normative
  document differs.
- One stray `.venv/` left inside the worktree by the previous cycle was removed;
  the task venv now lives outside the worktree at
  `.temp/TASK-260728-2spy93/validation-venv`.
- Nothing staged, committed, published, pinned, or advanced in `curator-spec`.

## Blocking findings and what changed

### 1. No legal wire location for the decision-0007 toolchain object

Accepted. Decision 0007 §2.2 places the requirement as REQUIRED on both schema-8
manifest build commands and OPTIONAL on the descriptor schema-2 target; decision
0008 previously said the local command is exactly three members and the
descriptor "changes exactly one thing, adds no member". `TASK-260728-2jaw7h` and
`TASK-260728-251p01` could not have satisfied both.

Reconciled by fixing the exact shapes and naming the definitions that carry them,
so the placement is now checkable rather than prose:

| Definition | Members | Toolchain |
|---|---|---|
| `buildCommandV6` (frozen) | `type`, `driver`, `source_dir` | must never gain it |
| `repositoryBuildCommandV1` (frozen) | `type`, `driver`, `repository`, `target` | must never gain it |
| `skillBuildTargetV1` (frozen) | `driver`, `build_root`, `source_dir` | must never gain it |
| `buildCommandV8` (reserved) | + `toolchain` | REQUIRED |
| `repositoryBuildCommandV2` (reserved) | + `toolchain` | REQUIRED |
| `skillBuildTargetV2` (reserved) | same three REQUIRED | OPTIONAL |

The decision states why the one exception is admissible — a
`toolchain-requirement-v1` object can only intersect an interval against the
manager-trusted set, cannot name a location or a candidate, and cannot reach the
registry `compatibility` set — and keeps every other ban, now extended with
toolchain root, mirror, channel/track, version-manager reference, install and
package-manager command, and trust root. Section 8 records why the REQUIRED
fourth member changes no `go-v1` identity: the requirement is a gate, not a build
input, so a schema-8 Go command produces a byte-identical canonical build input.
Rejected alternatives now cover the top-level-object and manager-side homes, and
why the descriptor member is OPTIONAL rather than REQUIRED or absent.

### 2. `manager-worker-v1` semantically widened while claimed unchanged

Accepted, and the previous rejected-alternative claim was false as written. A new
policy identity for new drivers only need not touch any Go cache key or rc.5
byte; only migrating Go would.

Taken: the reviewer's first branch. `manager-worker-v1` is left frozen and
Go-only, with its decision-0006 process graph and its exactly-one-`go list`-plus-
one-`go build` session intact and unre-read. A second identity,
`manager-worker-v2`, is minted for the six reserved drivers, is declared a
concurrent sibling rather than a successor, and carries the identical portable
containment contract — same mandatory controls, same
`rc5-native-control-inventory-v1`, same single failure boundary, same six
deferred guarantees. Exactly two things differ and both are named: the two lower
graph nodes are bound to the driver's fingerprinted toolchain closure, and the
v2 session admits *at most* one graph phase where v1 requires exactly one. The
driver-to-policy binding is a closed table with no third combination, and claim
schema 4 makes it structural by selecting the `execution_policy` `const` from the
assertion's own `driver` `const`.

The knock-on that the previous draft hid: `capability-evidence-v1` is normatively
closed to `manager-worker-v1` and to one record per operation, and section 9
admits a manifest mixing Go and additional-driver commands. It is therefore
re-versioned to `capability-evidence-v2` — same closed member set, same
probe-once-per-operation rule, one record per distinct active policy. It is
result-only, so no frozen byte moves.

The rejected-alternatives list now records the true cost of each option and adds
the naming trade-off honestly: a name outside the `manager-worker-` family would
remove the successor reading but would falsely imply a different containment
contract, so the family is continued and the non-ordering rule is normative.

### 3. Admitted-versus-reserved contradiction

Accepted. Section 2 is rewritten as two disjoint, separately labelled closed
sets — **Admitted wire driver set** (exactly `go-v1` and `go-repository-v1`, the
only values any schema, validator, manager, or claim reader may accept today) and
**Reserved driver namespace** (exactly the six, admitted by no schema). Their
union is declared the complete Protocol 1.0 identifier space; the transition
between them is a single atomic move owned by `TASK-260728-251p01`, with no
partial state, and the receipt-schema and execution-policy columns of the reserved
table are explicitly allocations rather than wire facts. Both labels are required
terms in the gate, so the split cannot be silently collapsed again.

### 4. The boundary gate did not enforce the boundary

Accepted; the demonstrated false accepts were real. The gate previously inspected
only `driver` plus a short artifact-member deny-list, so anything nobody had
thought to forbid passed.

Replaced with closed exact member-set tables:

- every driver-bearing `common.schema.json` definition must match its table entry
  exactly — property set, required set, and `additionalProperties: false`;
- a driver-bearing definition **missing from the table is itself a failure**, so a
  new shape cannot be smuggled in;
- a table entry that disappears from the schema is a failure;
- the three reserved schema-8 and descriptor schema-2 shapes are held in the same
  table, so their toolchain placement is enforced the moment they are minted;
- a residual deny-list of names that can never carry protocol meaning (generic
  language and build-system selectors, package-controlled installation and trust
  fields, runtime-bundle members) is applied to every `$defs` entry as defense in
  depth, and a test asserts it cannot collide with any deployed or reserved
  member.

The reserved-name absence scan now covers the reserved policy identity as well.
It deliberately does **not** cover the reserved evidence-record version: that
literal already appears in the frozen rc.5 corpus and release gate as the example
of a version that MUST be rejected, so absence would be the wrong test. Its
non-admission is instead proved positively — the gate requires the frozen corpus
to keep rejecting it. That collision is recorded in the decision with the exact
two literals `TASK-260728-251p01` must move when rc.6 mints the record, and with
the constraint that the rc.5 corpus bytes and pin must not change to do it.

## Code

`tools/validate.py` (+460 lines vs the accepted base): reserved policy and
evidence-record identifiers assembled from parts, the toolchain-requirement
object name, the two set labels, `DEPLOYED_WIRE_SHAPES` /
`RESERVED_WIRE_SHAPES` / `MANAGER_DRIVER_SHAPES` member-set tables,
`FORBIDDEN_BUILD_MEMBERS`, `check_closed_member_set()`,
`reserved_boundary_identifiers()`, `check_reserved_evidence_record_is_rejected()`,
and the rewritten `validate_additional_driver_boundary()`.

`tools/test_validate.py` (+491 lines vs the accepted base): 31 boundary tests,
16 of them new, including one test per demonstrated false accept.

Files differing from the accepted base — exactly three:

```text
decisions/0008-additional-language-driver-boundary.md   (new, 912 lines)
tools/validate.py
tools/test_validate.py
```

## Gate evidence (real exit codes, each command run standalone)

In the task worktree, with the task venv:

| Command | Exit | Result |
|---|---|---|
| `python tools/validate.py` | 0 | validated 42 schemas and 422 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | Ran 60 tests, OK (44 before, +16) |
| `go test ./tools/...` | 0 | ok generate-vectors |
| `go vet ./tools/...` | 0 | clean |
| `gofmt -l tools/` | 0 | no output |
| `python -m compileall -q tools` | 0 | clean |
| `git diff --check` | 0 | clean |

Determinism and candidate preservation, in the task worktree:

| Probe | Exit | Result |
|---|---|---|
| `go run ./tools/generate-vectors -root .` (run 1) | 0 | `diff -r conformance/v1` vs pre-run snapshot: 0 |
| `go run ./tools/generate-vectors -root .` (run 2) | 0 | `diff -r conformance/v1` vs snapshot: 0 |
| `diff release/1.0.0-rc.5.json` vs snapshot | 0 | unchanged |
| independent `sha256(conformance/v1/manifest.json)` | — | `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`, identical to the accepted predecessor pin |

Clean-checkout probe (isolated scratch git repo at
`.temp/TASK-260728-2spy93/clean-probe`, commit `92cf7a5`, never the real
repository), rebuilt from the final worktree state:

| Command | Exit |
|---|---|
| `make validate` | 0 |
| `make regenerate-check` (run 1) | 0 |
| `make regenerate-check` (run 2) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

`make regenerate-check` still cannot pass inside the task worktree, because it
runs `git diff --exit-code` against `57c1f56` while the worktree carries the
uncommitted rc.5 candidate. That is why it is reported from the clean probe.

Evidence-hygiene note, recorded because it produced one invalid reading before
it was caught: prepending the venv to `PATH` and invoking bare `python` does not
select the venv in this shell — an interactive alias binds `python` to
`/opt/homebrew/bin/python3`, which has no `jsonschema`, so `tools/validate.py`
exits 1 with `ModuleNotFoundError` and looks like a boundary-gate regression. A
confirmation run taken that way was discarded and re-run with the interpreter's
absolute path. Every exit code in the tables above comes from an explicit
interpreter path, and `make` targets from the clean probe where `make` invokes
`python3` directly with the venv first on `PATH`.

## Negative evidence — the three demonstrated false accepts

Replicated with the reviewer's own method, a read-only monkey-patch of
`load_json` (`logs/probe-false-accepts.py`, output
`logs/probe-false-accepts-final.log`), exit 0 because every probe was rejected:

| Probe | Previous gate | This gate |
|---|---|---|
| optional `language` on `buildCommandV6` | accepted | rejected: `$defs.buildCommandV6 is not its closed member set: added ['language']` |
| arbitrary `command` on `skillBuildTargetV1` | accepted | rejected: `... added ['command']` |
| `runtime_files` on `buildRecordV2` | accepted | rejected: `... added ['runtime_files']` |
| REQUIRED `language` on `buildCommandV6` | not probed | rejected: `... added ['language']` |
| unmodified schema | accepted | accepted |

CLI-level plant-and-revert probes against real files, each reverted immediately
and verified byte-identical afterwards:

| Probe | Exit | Diagnostic |
|---|---|---|
| `manager-worker-v2` planted in `protocol/core.md` | 1 | `protocol/core.md:1313: reserved identifier 'manager-worker-v2' is not admitted by any schema version` |
| `toolchain` added to the frozen `skillBuildTargetV1` | 1 | `$defs.skillBuildTargetV1 is not its closed member set: added ['toolchain']` |
| restore, re-run validator | 0 | validated 42 schemas and 422 vector files |

Unit-level negative coverage additionally exercises: relaxing a required member,
opening `additionalProperties`, removing a closed definition, adding an unlisted
driver-bearing definition, minting each reserved shape correctly and then wrongly
(missing `toolchain`, `toolchain` REQUIRED where it must be OPTIONAL, one extra
member), forbidden selectors on a non-driver definition, the reserved policy
identity on a surface, three mutations of the evidence-record rejection case, and
stripping each required term or set label from the decision.

## Deliberately not done

- No schema file, conformance case, vector, fixture, or release file added or
  changed. `1.0.0-rc.6`, `manager-worker-v2`, `capability-evidence-v2`, and the
  three reserved definition names exist only inside the decision record and the
  gate's reservation tables.
- No `CHANGELOG.md` or `COMPATIBILITY.md` entry: both record shipped wire
  changes, and this decision ships none.
- No `docs/` companion; nothing here is usable by an author yet.
- No commit, stage, publish, pin advance, native validation claim, or platform
  claim.
- The decision number stays 0008. Decision 0007 lives in the parallel
  `TASK-260728-1g0z69` worktree and is referenced by number and task ID only; if
  review lands the two in the other order, the references to renumber are the
  filename, the title, the `ADDITIONAL_DRIVER_BOUNDARY_DECISION` constant, and
  one docstring.

## Reviewer focus

1. Whether `manager-worker-v2` as a concurrent sibling with an explicit
   non-ordering rule is the right resolution, versus a composition model, given
   the naming trade-off is acknowledged rather than solved.
2. Whether re-versioning the capability-evidence record is justified by the
   mixed-operation case, and whether one record per active policy is the right
   cardinality.
3. Whether the reserved definition names (`buildCommandV8`,
   `repositoryBuildCommandV2`, `skillBuildTargetV2`) and their exact member sets
   match what `TASK-260728-2jaw7h` will actually need to land.
4. Whether proving the reserved evidence-record version un-admitted positively,
   rather than by absence, is acceptable given that literal is already in the
   frozen corpus.
5. That every `go-v1` and `go-repository-v1` identity is still byte-unchanged:
   the regeneration and pin evidence above is the intended proof.
