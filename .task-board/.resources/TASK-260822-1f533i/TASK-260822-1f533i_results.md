# TASK-260822-1f533i — core.md `script-worker-v1` normative prose

**Deliverable:** protocol core section 4.1.1, "Portable `script-worker-v1` execution policy".
**Branch:** `spec/sw-core-prose`, pushed to `origin`. No PR opened, per the spawn note.
**Worktree:** `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree`
**Commits:** `41cf556` "Define the portable script-worker-v1 execution policy (decision 0008)"
and `78d544d` "Correct the script inventory Linux process-count cell (review rework)", both
signed, base `3dc9ca6` (= `origin/main`).
**Diff:** `protocol/core.md` only, +388/-2 then +26/-11. Head `78d544d` = `origin/spec/sw-core-prose`.

## Branch and baseline note

The task description points at the prepared worktree on `spec/script-worker-v1-normative`, while
the board notes and the DoD name `spec/sw-core-prose` and ask for a push. Both were reconciled:
the prepared worktree was kept, its local-only branch was renamed to `spec/sw-core-prose`, and
that branch is what was pushed. The sibling prose task is on `spec/sw-manager-security`, so no
worktree or branch is shared.

The worktree was handed over at `a2d44eb`, behind `origin/main`. It was moved to `origin/main`
(`3dc9ca6`, decision 0009 and its correction, `decisions/` only, no `core.md` change), so the
branch is one clean commit ahead of `main` and `TASK-260822-c0rxj7` can land it without a rebase.

## What was written

Section 4.1.1 sits directly under section 4.1 (script and system commands) and mirrors the
structure and honesty discipline of section 4.2.1, in this order:

1. **Opt-in surface.** Declared-only unless the command selects `script-worker-v1`; OPTIONAL
   per-command field on manifest schema 8 or later; absent means declared-only on every schema;
   schema 7 and earlier keep their exact meaning. One paragraph explains why a one-value package
   field is not a package-visible policy choice: the influence is monotonically restrictive, and
   section 4.2.1's rule is restated as governing the *build* policy identity rather than
   reinterpreted.
2. **Skill-wide capabilities.** One capability set per skill, so every enforced command of one
   skill derives the same profile, and no other manifest value may vary it.
3. **Fixed process graph.** `manager parent -> identity-verified manager-owned script worker ->
   identity-verified interpreter -> manager-resolved executables named by `exec``. Fourth node
   exists only when `exec` names executables. Worker paragraph reuses 4.2.1's boundary language,
   including the trusted-computing-base sentence for an implementation that cannot ship an
   equivalent worker.
4. **Interpreter identity.** Closed identifier; manager-only, package-independent resolution;
   canonical regular file; substitution rejected; bytes hashed; identity re-checked at the launch
   boundary; no version-string execution; explicit refusal to apply the section 8.2 tree
   fingerprint, with the reason. Library and installed-package trees are named as trusted computing
   base. Shebang, extension, and Windows file association are inert.
5. **Launcher carve-out.** The manager-profile launcher rules govern declared-only commands; an
   enforced shim may not be a symlink, POSIX-shell wrapper, or `.cmd` wrapper, and the worker starts
   the interpreter by resolved path, never via `cmd.exe`, PowerShell, `sh -c`, `PATHEXT`, or a file
   association.
6. **Deny-by-default capability derivation.** Derivation reads the *declared manifest bytes*; an
   absent field derives the deny-by-default meaning and the schema default MUST NOT widen a derived
   control. Every value of section 4.3 is given an exact meaning under this policy: `network`
   (`none` / host globs), `exec`, `filesystem` (`repo`, `home-config`, path set, plus the always-on
   private runtime area), `secrets`, `env_read`, `prompt_scope`. Two honesty statements added
   here: derived `filesystem` values bound *writes only*, and `env_read` may not reintroduce a
   manager-owned name.
7. **Mandatory portable set** (11 bullets), single rejection sentence, and the install/update
   preflight with the same diagnostic.
8. **Three deliberate divergences from 4.2.1**, each stated as a positive control: standard input
   stays open under explicit manager binding; output may stream when the manager binds the streams;
   there is no policy wall-clock deadline.
9. **Mechanism vs deferred-guarantee table**, seven rows, `what it means` / `what it does not mean`
   / deferred name, mirroring 4.2.1's four-column shape.
10. **Script native-control inventory**, three platform columns, eight controls, `host-conditional`
    as a third availability value with its consistency rule, closed unavailable-reason vocabulary
    of five entries, probe rules, and an explicit "install-generation result replayed at invocation
    time is a cached result".
11. **`script-capability-evidence-v1`**, per-invocation cadence, no mid-session re-probe, closed
    field set, an eleven-row error table, result-only exposure with the explicit prohibition on the
    command's own stdout/stderr, bounded default retention, and non-aliasing against
    `capability-evidence-v1`.
12. **Deferred guarantees**, seven `script-`-prefixed names, disjoint from 4.2.1's frozen six in
    both directions.
13. **Failure boundary**, one boundary for control application, plus the complete four-entry
    diagnostic table and the "a manager that does not implement this policy MUST reject" rule.
14. **Package bytes remain interpreter input only**; audit surface and the two warning classes
    (`script-command-declared-only`, `script-command-unfiltered-declared-network`); the migration
    note for skills whose declarations were written as documentation; and identity non-aliasing
    with the section 12.3 admission rule for any further policy, interpreter, control, or
    availability value.

A second, three-sentence edit to section 4.3 keeps "capabilities are an audit surface, not a
runtime sandbox" true: it is now scoped to commands that have not opted in, and a pointer to
4.1.1 says the same declaration becomes the containment profile under enforcement.

## Names this subsection fixes (downstream tasks must match)

| Kind | Value |
|---|---|
| execution policy | `script-worker-v1` |
| inventory version | `script-worker-v1-native-control-inventory-v1` |
| evidence record | `script-capability-evidence-v1` |
| vector file | `conformance/v1/vectors/script-host-execution-policy.json`, section `native_control_inventory` |
| interpreter identifiers | `python3-v1`, `node-v1` |
| new controls | `descendant-exec-denial`, `filesystem-write-confinement`, `network-isolation-domain` |
| availability values | `available`, `host-conditional`, `unavailable` |
| unavailable reasons | `no-private-aggregate-domain`, `no-unprivileged-per-process-exec-policy`, `child-process-policy-requires-appcontainer`, `no-unprivileged-filesystem-domain`, `no-unprivileged-network-domain` |
| deferred guarantees | `script-total-network-denial`, `script-network-host-allowlisting`, `script-exact-executable-allowlisting`, `script-private-runtime-area-only-writes`, `script-read-only-runtime-tree`, `script-hard-aggregate-descendant-resource-bounds`, `script-fail-closed-capability-preflight` |
| diagnostics | `script_execution_control_unavailable`, `script_execution_capability_evidence_invalid`, `script_execution_hardened_claim_forbidden`, `script_execution_policy_unsupported` |
| audit warning classes | `script-command-declared-only`, `script-command-unfiltered-declared-network` |

## Calls made beyond the analysis

The analysis of `TASK-260822-1l4r4f` was followed on all five questions and all seven findings.
Four things it left open were decided here, each recorded so a reviewer can overturn one cheaply:

1. **Placement and number.** Section 4.1.1, under section 4.1, per decision 0008's first
   consequence option. Section 4.2 numbering is untouched.
2. **The interpreter set is exactly `python3-v1` and `node-v1`.** The analysis left "exact set =
   prose task's call". `bash-v1` and `powershell-v1` are named in the text and deferred, with the
   stated reason that neither a POSIX shell nor PowerShell resolves on all three supported
   platforms, so an enforced command declaring one would be uninstallable on at least one lane.
   Adding either later is a section 12.3 revision and costs one enum value.
3. **The seven deferred-guarantee names.** The analysis required script-specific names disjoint
   from the frozen build six and named only `script-network-host-allowlisting`. The other six are
   the build six under a `script-` prefix, except `script-read-only-runtime-tree` (the build name
   is `read-only-source-and-toolchain`, which has no script referent) and
   `script-private-runtime-area-only-writes` (build: `private-build-root-only-writes`).
4. **Two rules the analysis did not raise.** `env_read` naming a manager-owned variable MUST NOT
   pass the inherited value through, or `env_read: ["PATH"]` would defeat mandatory control 5; and
   derived `filesystem` values bound writes only, with an explicit MUST NOT report otherwise.

## Divergences from decision 0008's text, all deliberate

The reviewer of `TASK-260822-1l4r4f` asked (finding 2) that the prose task record these rather
than inherit them silently. Five points where 4.1.1 does not do what 0008 literally says:

1. **§4 "one closed `capability-evidence-v1` record".** 4.1.1 defines a new closed record version,
   `script-capability-evidence-v1`. 0008 as written is not implementable: `core.md:817` (was
   `:437`) makes an `execution_policy` other than `manager-worker-v1` a
   `build_execution_hardened_claim_forbidden`. Analysis F2.
2. **§3 last bullet vs the schema default.** 0008 says absent fields mean no writes outside the
   private runtime area; `common.schema.json` defaults `filesystem` to `"repo"`. 4.1.1 derives from
   the declared manifest bytes and states that the schema default MUST NOT widen a derived control.
   Analysis F1. `TASK-260822-1mwy10` may prefer to drop the default in a schema-8 capability shape;
   either way this text holds.
3. **§3 "kernel-grade host filtering enters the native inventory per platform".** 4.1.1 does the
   opposite: no host-filtering entry may appear in the inventory or a record, and host-granular
   filtering is deferred as `script-network-host-allowlisting`. Reason: no supported host provides
   host-granular filtering unprivileged in any form (Landlock filters ports, never hosts), and the
   existing suite already treats `host-firewall-profile` as an out-of-inventory rejection. Analysis
   Q3; this is the divergence the reviewer flagged as unlisted.
4. **§4 "per invocation-policy identity".** Read as "exactly one record per enforced-command
   invocation", and written that way. Analysis Q4.
5. **Platform scope.** 0008 offers "unshare/netns on Linux" as an example while no
   execution-policy ledger has ever had a Linux cell. 4.1.1 ships three platform columns and a
   `host-conditional` availability value. Analysis F5/Q5, the largest single call inherited from
   the analysis; undoing it means dropping the column and moving the three new controls to the
   deferred list.

## Handoff to the sibling tasks

- **`TASK-260822-1mwy10` (schema 8).** Needs `execution_policy: {"const": "script-worker-v1"}` and
  the closed `interpreter` enum on a schema-8 script command shape, reachable only from a v8
  command; schema 7 bytes stay frozen. `scriptCommand` is `additionalProperties: false`, so a new
  `$defs` sibling is the only legal shape. Consider whether the schema-8 capability shape keeps
  `filesystem.default = "repo"`.
- **`TASK-260822-3fkfmf` (manager profile, SECURITY.md).** 4.1.1 defers exactly two things to the
  manager profile: the reserved environment-name set per platform and per interpreter identifier,
  and the surfaces on which the two audit warning classes are emitted. It also requires a carve-out
  in the manager profile's launcher rules (`manager.md:503-509`) saying they govern declared-only
  commands. SECURITY.md needs the mechanism/guarantee split for scripts and the TCB sentence for
  the interpreter's installed package tree, which applies to the Python manager itself.
- **`TASK-260822-f4qv7w` (vectors).** 4.1.1 names
  `conformance/v1/vectors/script-host-execution-policy.json` and its `native_control_inventory`
  section as the machine-readable authority; that file does not exist yet, so the two branches must
  land together. The reviewer's finding 3 stands: `host-conditional` controls do not fit the
  existing `host-capability` platform-case class, which `.github/ci/platform-cases.tsv` defines in
  terms of the runner *filesystem*; a widened definition or a new `host-kernel-feature` class is
  needed. Each `host-conditional` control also needs the companion case "probe reports unavailable,
  evidence reports unavailable, invocation still succeeds" on all three lanes.
- **`TASK-260822-c0rxj7` (land).** CHANGELOG.md has no Unreleased section and was deliberately not
  touched here; the release entry belongs to the landing change so the three prose branches do not
  conflict. It must also land `spec/sw-core-prose` together with `TASK-260822-f4qv7w`'s vectors,
  because 4.1.1 cites `conformance/v1/vectors/script-host-execution-policy.json`, and it must not
  merge while the schema-8 reachability gap below is open.
- **Schema 8 is unreachable from section 4 — open, unowned.** 4.1.1 says "manifest schema 8 or
  later", and `spec/sw-schema` adds `agent-skill-v8.schema.json` / `csk-skill-v8.schema.json`, but
  `core.md` section 4 still says manifests conform to v1 through v7, its "Schema | Added behavior"
  table stops at row 7, and the downward version-gate paragraph says "schemas 2 through 7 reject
  unknown fields". `spec/sw-schema` touches no `.md` outside `schemas/v1/README.md`, and this
  branch is scoped to 4.1.1, so no branch in the story closes it. Raised by the `41cf556` reviewer
  as `TASK-260822-1mwy10` prose scope; still open at `78d544d`. Needs an owner before `c0rxj7`
  merges, or merged `core.md` references a manifest schema its own section 4 does not admit.

## Gate evidence

Run in the worktree at commit `41cf556`; logs under
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260822-1f533i/`.

| Command | Exit | Log |
|---|---|---|
| `python tools/validate.py` (schemas, vectors, local links) | 0 | `validate-rebased.log` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | `unittest-01.log` |
| `go test ./tools/...` | 0 | `gotest-01.log` |
| `go run ./tools/generate-vectors -root .` then `git diff --exit-code -- conformance/v1 release/*.json` | 0 and 0 | `regenerate-01.log` |
| `git diff --check` (whitespace formatting job) | 0 | `whitespace-01.log`, re-run at `41cf556` |
| `lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429 '**/*.md'` (links job) | 0, 41 OK / 0 errors | `lychee-rebased.log` |
| `gofmt -l tools` | not run: no Go file changed by this task | - |

`tools/validate.py` needs `jsonschema==4.25.1`, which is not in the system Python here. A venv was
created at `.temp/TASK-260822-1f533i/venv` from `requirements-dev.txt` and used for every Python
gate. `lychee` was installed via Homebrew for the link job.

## Review rework — commit `78d544d`

The `41cf556` review (`TASK-260822-1f533i_review-verdict.md`, RUN-260822-ecf868) returned changes
requested with one blocking item and three cheap ones. All four are fixed in a single commit; no
gate regressed.

### 1. Blocking — the Linux `active-process-count-limit` cell mandated a false `applied` claim

The cell said `available: RLIMIT_NPROC`. Section 4.1.1's own rule that "an `available` control MUST
report `applied`" therefore *required* every conforming Linux manager to emit
`active-process-count-limit / available / applied` in each `script-capability-evidence-v1` record,
on the strength of a limit that bounds processes for the real UID of the caller — shared with every
other process that user owns — and so provides no domain private to the invocation, which is the
literal content of `no-private-aggregate-domain`. The same row's macOS cell says exactly that about
the same rlimit.

Fixed as recommended: the Linux cell becomes `host-conditional: delegated cgroup v2 pids.max`, matching
what `aggregate-memory-limit` already does with the memory controller. No new machinery —
`host-conditional` already exists in the inventory and 4.1.1 already says a `host-conditional`
control that probes unavailable MUST NOT reject.

One addition beyond the reviewer's ask: a new normative paragraph after the inventory states that a
per-user resource limit is not a private aggregate domain, that a manager MUST NOT back
`active-process-count-limit` or `aggregate-memory-limit` with `RLIMIT_NPROC` or `RLIMIT_AS` on any
platform, and that `RLIMIT_FSIZE` legitimately backs `per-file-size-limit` because that control
bounds one file write rather than an aggregate. Without it the cell reads as an oversight and the
next editor "corrects" it back.

**Divergence from the analysis, now recorded** (this is the record the reviewer found missing):
`TASK-260822-1l4r4f_analysis.md:560` proposes `available: RLIMIT_NPROC` for this cell. 4.1.1
deliberately does not follow it, for the reasons above and because the analysis's own honesty rule
(line ~792: "both are mapped to `unavailable`/`host-conditional`, so neither can produce a false
`applied` claim") is not satisfiable with an `available` verdict here. It also collides with
`script-hard-aggregate-descendant-resource-bounds` being deferred. `TASK-260822-3fkfmf` flagged
this by name in its coordination note and mirrored the analysis rather than diverging unilaterally
in one of three copies.

**Lockstep obligation — not done here, this branch owns only `core.md`:**

- `profiles/manager.md` line 710 on `spec/sw-manager-security` (`c2371d3`) carries the identical
  row and still says `available: RLIMIT_NPROC`. `TASK-260822-3fkfmf` is in `reviewing`; the cell
  must change there, or `TASK-260822-c0rxj7` must change it at landing.
- `conformance/v1/vectors/script-host-execution-policy.json` must be written with
  `host-conditional` for this cell when `TASK-260822-f4qv7w` creates it, plus the companion
  "probe reports unavailable, evidence reports unavailable, invocation still succeeds" case that
  every `host-conditional` control needs on the Linux lane.

### 2. "The complete diagnostic set" → "The policy-level diagnostic set"

`profiles/manager.md@c2371d3` defines seven `script_execution_*` codes and pre-reconciles the gap
("the first four are the policy-level set of Protocol Core section 4.1.1. The last three are
worker-session codes of this profile"). Section 4.2.1, the mirror target, claims completeness
nowhere. The word "complete" manufactured a contradiction the moment the branches merge. One word
changed; the four-row table is unchanged.

### 3. The manifest fields are now named

4.1.1 previously said only "an OPTIONAL per-command field" and "a closed interpreter identifier".
It now names `execution_policy` (with its single admitted value) and `interpreter` where each is
introduced, and adds a sentence stating the co-requirement that `common.schema.json@ebfed81`
expresses as `dependentRequired` in both directions: a script command declaring one without the
other is an invalid manifest, MUST be rejected by manifest validation, MUST NOT get a default, and
MUST NOT be installed declared-only. Section 4.1 above names `unix_path`/`win_path` and 4.3 names
all six capability fields, so this restores the pattern and lets the section be read standalone.

### 4. `network-isolation-domain` non-aliasing sentence — right prefix

It previously forbade spelling the control as `total-network-denial`, which is section 4.2.1's
*build* guarantee. The sentence now names both: `script-total-network-denial` as this policy's
deferred guarantee (the name a script surface would actually be tempted to misuse) and
`total-network-denial` as a build-policy name that may not appear on a script surface at all.

### Not addressed here, by the reviewer's own routing

The schema-8 reachability gap in section 4 is recorded under "Handoff to the sibling tasks" above.
The reviewer raised it as coordination and explicitly did not charge it against this task.

### Gates at `78d544d`

Re-run in full, logs under `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260822-1f533i/`.

| Command | Exit | Log |
|---|---|---|
| `python tools/validate.py` | 0 — 49 schemas, 471 vector files | `rework-validate.log` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 — 91 tests OK | `rework-unittest.log` |
| `go test ./tools/...` | 0 | `rework-gotest.log` |
| `go run ./tools/generate-vectors -root .` | 0 | `rework-genvectors.log` |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | 0 — deterministic regeneration | (inline) |
| `git diff --check` (formatting job) | 0 | (inline) |
| `lychee --offline '**/*.md'` (links job) | 0 — 36 OK, 0 errors, 6 excluded | `rework-lychee.log` |
| `gofmt -l tools` | 0 — clean; no Go file changed by this task | (inline) |

Same venv as before (`.temp/TASK-260822-1f533i/venv`, `jsonschema==4.25.1` from
`requirements-dev.txt`), because system Python here has no `jsonschema`.
