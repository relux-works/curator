# TASK-260822-3fkfmf review cycle 2 — verdict: ACCEPTED

**Verdict: accepted → `done`.** All four blocking findings of cycle 1 (R1–R4) are
genuinely fixed and were re-verified mechanically, not taken on the implementer's
report. Gates re-run standalone by this reviewer with real exit codes. Nothing new
and blocking was found. Remaining items are hand-forwards to `TASK-260822-c0rxj7`
and `TASK-260822-f4qv7w`.

## What was reviewed

Repo `curator-spec`, branch `spec/sw-manager-security`, HEAD `e5df43d`, worktree
`curator-spec/.temp/STORY-260822-3k3hbs/manager-security-worktree`.

- Four commits `b9ca2ad` → `c2371d3` → `110e1f6` → `e5df43d`, every one signed `G`
  (`oparin@me.com`).
- `git ls-remote origin` confirms `refs/heads/spec/sw-manager-security` = `e5df43d`;
  local HEAD matches, branch is in sync with origin.
- Base `b92b105`. Cumulative diff vs base: **2 files, +502/-4**, exactly
  `SECURITY.md` (+175) and `profiles/manager.md` (+331). No other path touched.
- All four deleted lines are the single `SECURITY.md` security-model paragraph the
  change deliberately qualifies; nothing else was removed anywhere on the branch.
- Worktree clean apart from untracked `.temp/` and `tools/__pycache__/`.

Cross-checked against sibling branches at their current heads, verified by
`ls-remote`: `spec/sw-core-prose` @ `78d544d`, `spec/sw-schema` @ `ebfed81`.

## Gates — re-run standalone by the reviewer on `e5df43d`

Note: the first pass was piped through `${PIPESTATUS[0]}` under zsh, which returns
empty; every gate below was re-run under bash with a direct `$?` so the exit codes
are real.

| Gate | Exit |
|---|---|
| `.temp/venv/bin/python tools/validate.py` | **0** — 49 schemas, 471 vector files |
| `.temp/venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'` | **0** — 91 tests OK |
| `go test ./tools/...` | **0** |
| `gofmt -l tools` | **0**, empty listing |
| `go run ./tools/generate-vectors -root .` | **0** |
| `git diff --exit-code -- conformance/v1 release/` | **0** — regeneration deterministic |
| `git diff --check` | **0** |
| `lychee --no-progress SECURITY.md profiles/manager.md` | **0** — 1 link, 1 OK, 0 errors |

`validate.py` covers local links, so internal cross-references are gate-verified.

## Blocking findings of cycle 1 — each re-verified

### R1 — reserved environment-name set narrower than core.md's criterion → **fixed**

`profiles/manager.md` 3.1. Both axes the cycle-1 review offered were taken, and both
were needed:

1. **Enumeration widened, loader family closed by prefix.** `LD_PRELOAD` /
   `LD_LIBRARY_PATH` / `LD_AUDIT` → *every `LD_`-prefixed name*, those three as
   examples. This reaches `LD_ORIGIN_PATH`, `LD_TRACE_LOADED_OBJECTS`,
   `LD_PROFILE_OUTPUT`, and `LD_DEBUG_OUTPUT`, which cycle 1 named by hand, plus the
   rest of the glibc family. `SYSTEMROOT` added on Windows, with a new normative
   sentence that Windows environment names match case-insensitively (so `SystemRoot`
   and `windir` are the same reserved names) while macOS and Linux match exactly —
   which is also what justifies the separate lowercase-proxy listing. Two further
   names added on the implementer's own initiative and both check out: `LOCALDOMAIN`
   (glibc resolver search-list override, same family as the already-reserved
   `RES_OPTIONS` / `HOSTALIASES`) and `__PYVENV_LAUNCHER__` (selects the interpreter
   on CPython framework builds, not `PYTHON`-prefixed, so the prefix rule misses it).
2. **"Closed" → "reserved minimum" with core's criterion still governing.** The new
   paragraph states normatively that a name meeting section 4.1.1's criterion is
   manager-owned even when unlisted, that a manager MUST reserve it, and that a
   manager MUST NOT read an unlisted name as licence to pass an inherited value
   through. Omission is declared a defect of the profile, corrected by revising the
   enumeration, and stays visible because every withheld `env_read` entry MUST be
   reported.

`grep -n "reserved" protocol/core.md` confirms core.md 4.1.1 still states the open
criterion (line 351-357) and delegates the enumeration. The two documents now agree
in substance: enumeration is the interop floor, criterion is the backstop, and the
residual divergence is fail-closed (variable absent) and reported, rather than
fail-open and silent as before. The implementer's consequence note for f4qv7w — that
a "name outside the set passes through" negative vector is no longer sound — is
correct and is carried forward below.

### R2 — broken nested code span in SECURITY.md → **fixed**

The inline span was replaced with the same fenced ```` ```text ```` block the other
two files use. Verified by extracting the graph from all three files and diffing:
`protocol/core.md`, `profiles/manager.md`, and `SECURITY.md` are now **byte-identical**
across all four lines. The `exec: "none"` three-node case moved to its own sentence
after the block, matching core.md's structure.

### R3 — §3 carve-out over-scoped → **fixed**

`profiles/manager.md`: "The launcher rules of this section govern **every command
other than an enforced script command**." This is the wording cycle 1 proposed and it
no longer contradicts 3.1's own "build commands are outside this policy entirely;
their install-time execution boundary is section 2.2.1 and their launcher is
unchanged".

### R4 — over-long lines → **fixed**

Mechanically re-checked over every line the branch adds vs `b92b105`: **zero**
non-table added lines exceed 81 columns; exactly two sit at 81, which are the two
cycle 1 explicitly cleared as within existing tolerance. The 22 added lines over 81
are all Markdown table rows, which the file already wraps that way throughout.

## Re-mirror against core.md @ `78d544d` — verified

`spec/sw-core-prose` moved from `41cf556` to `78d544d` during the review cycle, so
the mirror had to be refreshed. Checked mechanically:

- **Native-control inventory** — the 8-row × 3-platform table in `profiles/manager.md`
  (lines 725-734) is **byte-identical** to core.md's (lines 429-438), including the
  H1 correction: Linux `active-process-count-limit` is now
  `host-conditional: delegated cgroup v2 pids.max`, not `available: RLIMIT_NPROC`.
- **rc5-overlap sentence** — now scoped to "the same **macOS and Windows** verdicts".
  Verified row by row against `profiles/manager.md`'s own
  `rc5-native-control-inventory-v1` table (lines 253-257): all five macOS and Windows
  cells match. rc5 has no Linux column, so the previously unqualified claim was false
  after the H1 correction and is now true. core.md carries the identical qualification.
- **Per-user resource limit rule** — `e5df43d` removes manager.md's compact copy of
  core.md's new "a per-user resource limit is not a private aggregate domain /
  MUST NOT back `active-process-count-limit` or `aggregate-memory-limit` with
  `RLIMIT_NPROC` or `RLIMIT_AS`" rule, on 1f533i's request as owner of 4.1.1. I checked
  this leaves no gap and no dangling reference: the rule survives in core.md (lines
  446-454), and a manager reading manager.md alone still cannot reach `RLIMIT_NPROC`
  because the inventory cell names the mechanism (`delegated cgroup v2 pids.max`) and
  the inventory is declared exhaustive and normative. The deleted paragraph had no
  inbound references. Not duplicating a normative MUST across two documents is the
  right call.
- **H2** — core.md now reads "The **policy-level** diagnostic set of this policy is:"
  (line 558), not "complete". The seven-row table in manager.md 3.1 with its
  four-policy-level + three-worker-session framing is now consistent with core.md
  without needing a change here. Confirmed the `build_execution_*` precedent the
  framing leans on is real: manager.md 2.2.1 carries six build codes including
  `build_execution_worker_identity_invalid`, `_worker_protocol_invalid`, and
  `_package_influence_forbidden`.

Additional identifier checks run this cycle, all matching core.md `78d544d`:

- seven deferred guarantee names — identical sets in core.md, manager.md, and
  SECURITY.md, and disjoint from the build policy's six (verified against
  SECURITY.md's build table at lines 126-133);
- `script-capability-evidence-v1` record fields (`record_version`,
  `execution_policy`, `platform`, `controls`) and per-entry fields
  (`name`, `availability`, `status`, `probed_at` = `pre-worker-launch`) — match;
- closed unavailable-reason vocabulary — 5 entries, match;
- three availability values and the applied/unavailable rules — match;
- SECURITY.md's 7-row mechanism/deferred-guarantee table — same rows, same order,
  same guarantee per row as core.md's 4-column version;
- `execution_policy` const `script-worker-v1` and `interpreter` enum
  `{node-v1, python3-v1}` — match `spec/sw-schema` @ `ebfed81`
  `schemas/v1/common.schema.json` (`scriptExecutionPolicyV1` /
  `scriptInterpreterV1`), including the `dependentRequired` co-requirement;
- `decision-0006` resolves to `0006-portable-manager-worker-execution.md`; sections
  2.2.1 (line 202), 3.1 (line 550), and 7 (line 926) all exist, and 2.2.1 was already
  referenced by pre-existing prose, so the new references are not inventing a target.

## AC — met

- **Both files committed on the story branch** — yes, `spec/sw-manager-security` @
  `e5df43d`, signed and pushed, exactly the two files.
- **Gates pass** — yes, re-verified standalone above.
- **No kernel-guarantee claim for portable enforcement** — re-audited mechanically
  over the added lines only. All 6 `kernel` occurrences are negations or exclusions
  ("Neither form is a kernel sandbox", "stops short of a kernel-enforced guarantee",
  "Kernel confinement is not part of it", "no kernel confinement primitive at all"
  ×2, "no supported host offers host-granular filtering as a kernel control"). All 15
  `guarantee` occurrences are deferrals, prohibitions, or table headings. No sentence
  asserts the portable policy denies, prevents, or confines at kernel level on any
  platform. The macOS/Windows "no kernel confinement primitive at all" statements and
  the "applying a host-conditional control never upgrades the policy-level guarantee"
  rule are present in both files.

Decision 0008's two consequences scoped to this task are both delivered:
`profiles/manager.md` gets manager obligations for worker launch (8 ordered
obligations) plus the per-platform control inventory; `SECURITY.md` gets the
enforcement/guarantee split mirroring the build-policy prose, as a clean structural
peer of `## Compile-only build boundary` and `## External build repository boundary`.
No stale contradicting prose survives elsewhere in the repo — checked every `.md`
outside the three policy documents.

## Non-blocking — hand forward

### N1. core.md still says the profile defines the "exact" reserved set → `TASK-260822-1f533i` scope, route via `c0rxj7`

core.md 4.1.1 line 356-357: "The **exact** reserved set per platform and per
interpreter identifier is defined by the manager profile." After R1's fix the profile
supplies a *minimum* and defers the remainder back to core's criterion, so "exact"
now slightly overclaims what the profile provides. This is a wording seam, not a
behavioural divergence: under either document a manager reserves the listed names and
also any unlisted name meeting the criterion, and there is no input on which the two
lead to different behaviour. Cheapest fix is one word in core.md ("exact" → "minimum
reserved set"), in the file this task does not own.

### N2. Two list-nesting ambiguities in the reserved enumeration — cosmetic

- Base set: "…`LOCALDOMAIN`, every `LD_`-prefixed name, including `LD_PRELOAD`,
  `LD_LIBRARY_PATH`, and `LD_AUDIT`, `IFS`, and `CSK_PROJECT_ROOT`." Grammatically the
  trailing `IFS` / `CSK_PROJECT_ROOT` continue the inner "including" list.
- `python3-v1` set: "…and `PYTHONSAFEPATH`, and `__PYVENV_LAUNCHER__`" reads as if
  `__PYVENV_LAUNCHER__` were an example of a `PYTHON`-prefixed name, which it is not.

Both parses yield the identical reserved set — "including X" makes X a member either
way — so nothing normative turns on it. Worth one sentence-split if `c0rxj7` is
touching these paragraphs anyway.

### N3. "A manager that cannot apply all of them" antecedent is now distant — cosmetic

`profiles/manager.md` 3.1: the sentence sits ~45 lines below the mandatory-control
bullet list it refers to, separated by the environment-reservation prose that R1's fix
made two paragraphs longer. core.md places the same sentence directly after its list.
The diagnostic name `script_execution_control_unavailable` disambiguates, so this is
readability only.

### H3 (carried, still owned by nobody). Audit-record extension for the policy identity

`decisions/0008-enforced-script-capabilities.md:117` requires it under the "Schemas"
bullet. `profiles/manager.md` §7 now states normatively that the record carries the
effective execution-policy identity or its absence, but
`schemas/v1/audit-record-v1.schema.json` is `additionalProperties: false` and
`spec/sw-schema` @ `ebfed81` did not touch it. Implementable today only through the
open `audit` object, so no schema version pins it and no vector can assert it.
**Route to `TASK-260822-f4qv7w` or `c0rxj7`** — it is not this task's defect and it is
currently unassigned.

### H4 (carried). `host-capability` platform-case class

`platform-cases.tsv` defines it as a filesystem limitation, not a missing kernel
feature, and host-conditional controls route to it. Needs a widened or new class in
`TASK-260822-f4qv7w`.

### H5 (carried). Windows Job Object nesting

Unconfirmed, inherited from rc5 where the identical claim already ships. No new
exposure from this change.

### F1 (new, from R1). Vector consequence for `TASK-260822-f4qv7w`

A negative vector asserting "a name outside the reserved set passes through" is no
longer sound under the reserved-minimum framing and MUST NOT be written. Positive
vectors over the enumerated minimum and over the withheld-entry reporting rule are
unaffected.

### F2 (new). `script-host-execution-policy.json` does not exist on any branch yet

core.md 4.1.1 line 425-427 names
`conformance/v1/vectors/script-host-execution-policy.json` as the authority for the
inventory. `git ls-tree` over `origin/main`, `spec/sw-core-prose`, `spec/sw-schema`,
and this branch finds no such path. Neither of *this task's* two files cites the
vector file — both reference only the `script-worker-v1-native-control-inventory-v1`
version string — so nothing here dangles and `validate.py` / `lychee` are both clean.
Flagged so `c0rxj7` does not land core.md with a forward reference `f4qv7w` has not
yet satisfied.

## Assessment

The rework is disciplined. Each of R1–R4 was fixed at the level the finding was
raised at rather than patched at the symptom, and R1 in particular was fixed on both
axes when either alone would have left a seam — the reasoning given for that
(enumeration as interop floor, criterion as backstop, divergence fail-closed and
reported) is correct and is the argument I would have made. The re-mirror against a
core.md that moved mid-review was caught proactively rather than left to rot, and the
one place the implementer chose *not* to mirror (`e5df43d`'s removal of the duplicated
per-user-limit MUST) is the right normative-layering call, which I verified leaves no
behavioural gap. The rework write-up's factual claims all held up under independent
check, including the byte-identity claims and the line-length claim.

No forced fit anywhere in this change: every place the portable policy cannot deliver
a guarantee, the prose says so and names the deferred guarantee rather than smoothing
it over.

**Accepted.** Work is already committed and pushed on `spec/sw-manager-security`; no
uncommitted scope remains, so no separate commit-owning mover step is needed.
