# TASK-260822-3fkfmf rework cycle 2 — review findings R1-R4 addressed

Branch `spec/sw-manager-security` in `curator-spec`, rework commits `110e1f6`
then `e5df43d`, both signed `G` and pushed, branch in sync with origin. Base
`b92b105`. Cumulative diff vs base: 2 files, +502/-4. The two rework commits
touch nothing outside `SECURITY.md` and `profiles/manager.md`.

Worktree: `curator-spec/.temp/STORY-260822-3k3hbs/manager-security-worktree`.
No PR by design — TASK-260822-c0rxj7 merges the sibling branches.

## Blocking findings — all four fixed

### R1 — reserved environment-name set narrower than core.md's criterion

`profiles/manager.md` 3.1. The set was declared **closed** while core.md 4.1.1
states an **open criterion** and delegates only the enumeration, so the two
documents asserted different things and an `env_read` entry naming a name in
the gap passed the inherited value through.

Fixed on both axes the reviewer offered, because either alone leaves a seam:

1. **The enumeration got wider**, and the loader families are now closed by
   prefix the way macOS already was:
   - `LD_PRELOAD`, `LD_LIBRARY_PATH`, `LD_AUDIT` → **every `LD_`-prefixed
     name**, listing those three as examples. This covers `LD_ORIGIN_PATH`,
     `LD_TRACE_LOADED_OBJECTS`, `LD_PROFILE_OUTPUT`, and `LD_DEBUG_OUTPUT`,
     which the reviewer named, plus the rest of the glibc family.
   - `LOCALDOMAIN` added to the base set. Same glibc resolver-override family
     as the already-reserved `RES_OPTIONS` and `HOSTALIASES`; it plainly meets
     the "resolver configuration" clause of the criterion and was missing.
   - `SYSTEMROOT` added on Windows, and a sentence stating that Windows
     environment-variable names match **case-insensitively**, so `SystemRoot`
     and `windir` are the same reserved names as `SYSTEMROOT` and `WINDIR`,
     while macOS and Linux match exactly — which is why the lowercase proxy
     spellings still need their own listing. The case rule was previously
     unstated and the `WINDIR`/`SystemRoot` split was the symptom.
   - `__PYVENV_LAUNCHER__` added for `python3-v1`: it selects the interpreter
     on CPython framework builds and is not `PYTHON`-prefixed, so the prefix
     rule does not reach it.

2. **The "closed" claim became "reserved minimum"** with the core criterion
   still governing. New paragraph states normatively that a name meeting
   section 4.1.1's criterion is manager-owned even when unlisted, that a
   manager MUST reserve it, and that a manager MUST NOT treat an `env_read`
   entry as licensed to pass an inherited value through merely because the
   name is unlisted. An omission is a defect of this profile, corrected by
   revising the enumeration rather than left to per-manager judgement, and it
   stays visible because every withheld entry MUST be reported.

Why both: prefix closure alone still leaves an unprovable completeness claim
(a genuinely complete enumeration over two interpreters and three platforms is
not something this profile can demonstrate), and the minimum-plus-criterion
framing alone leaves per-manager divergence on names anyone could have listed.
Together, the enumeration is the interop floor and the criterion is the
backstop. The divergence that remains is fail-closed (a variable is absent
rather than passed through) and reported, which is the safe direction; the
previous shape diverged fail-open and silently.

Consequence for TASK-260822-f4qv7w: a negative vector asserting "a name outside
the set passes through" is no longer sound and MUST NOT be written. Positive
vectors over the enumerated minimum, and over the withheld-entry reporting
rule, are unaffected.

### R2 — broken nested code span in SECURITY.md

`SECURITY.md`. The process graph was wrapped in one backtick span with a second
backtick pair around `exec` inside it; six backticks paired left to right and
the graph rendered as broken fragments in a public-facing file. Replaced with
the same fenced ```text block `protocol/core.md` and `profiles/manager.md`
already use, so all three copies of the graph are now identical and render.
The `exec: "none"` three-node case moved to its own sentence after the block.

### R3 — section 3 carve-out over-scoped

`profiles/manager.md` section 3. "The launcher rules of this section govern
declared-only commands" read as removing build and system commands from those
rules, contradicting the rest of section 3 and 3.1's own "build launchers are
unchanged". Now: "govern every command other than an enforced script command."

### R4 — two over-long prose lines

`SECURITY.md` (120 cols) and `profiles/manager.md` (85 cols) rewrapped, plus
the two paragraph tails the rewrap left ragged. Every non-table line added by
this branch vs `b92b105` is now ≤81 columns; the two lines at 81 are the
pre-existing ones the reviewer explicitly cleared as within tolerance.

## Re-mirrored against core.md as it now stands

`spec/sw-core-prose` moved from `41cf556` to `78d544d` while this task was in
review, and that commit fixed two of the findings this task handed forward.
Since 3.1 mirrors 4.1.1, the mirror had to be refreshed or the two documents
would have contradicted each other again:

- **H1 resolved upstream and mirrored.** Inventory Linux
  `active-process-count-limit` cell: `available: RLIMIT_NPROC` →
  `host-conditional: delegated cgroup v2 pids.max`. Verified mechanically:
  the 8-row inventory in `profiles/manager.md` is now byte-identical to
  core.md's.
- The rc5 overlap sentence is scoped to **macOS and Windows** verdicts, since
  `rc5-native-control-inventory-v1` has no Linux column.
- Core.md's new normative rule that no per-user resource limit may back either
  aggregate control is **not** restated here. `110e1f6` mirrored it in compact
  form; `e5df43d` removed that copy on the explicit request of
  TASK-260822-1f533i, which owns core.md 4.1.1 — the core section owns the rule
  and the profile inherits it, and a second copy of a normative MUST is only a
  place for the two documents to drift. The corrected inventory cell stays,
  and so does the rc5-overlap sentence, because that sentence is this profile's
  own and was made false by the correction rather than duplicated from it.
- **H2 resolved upstream, no change needed here.** Core.md now says
  "policy-level diagnostic set" instead of "complete diagnostic set", which is
  exactly the framing 3.1's seven-row table already used ("the first four are
  the policy-level set … the last three are worker-session codes of this
  profile"). The tension is gone without touching this file.

## Minor polish taken

Two mandatory bullets aligned on the core.md spellings the reviewer flagged:
"when the derived `network` capability is `none`" (was: when `network` is
`"none"`), and "operation-private temporary, configuration, and cache roots"
(was: private per-command configuration and cache roots — a different
lifetime). The third polish item was flagged for exactness only, with the
broader statement agreed true; left as-is.

## Gates — each run standalone, real exit codes, on the committed content

| Gate | Exit |
|---|---|
| `.temp/venv/bin/python tools/validate.py` | **0** — 49 schemas, 471 vector files |
| `.temp/venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'` | **0** — 91 tests |
| `go test ./tools/...` | **0** |
| `go run ./tools/generate-vectors -root .` | **0** |
| `git diff --exit-code` over `conformance/v1` + `release/1.0.0-rc.5/6/7/8.json` | **0** — deterministic |
| `gofmt -l tools` | **0**, empty listing |
| `git diff --check` | **0** |
| `lychee --no-progress SECURITY.md profiles/manager.md` | **0** — 1 link, 1 OK, 0 errors |

Run in full twice: on `110e1f6`'s content and again on `e5df43d`'s. No gate was
expected-red and none was piped through `tee`.

## No-kernel-guarantee criterion — re-audited after the rework

All 6 case-insensitive `kernel` occurrences and all 15 `guarantee` occurrences
in the lines this branch adds vs `b92b105` are negations, deferrals, or
prohibitions. No sentence asserts that the portable policy denies, prevents, or
confines at the kernel level on any platform. The explicit "no kernel
confinement primitive at all" statements for macOS and Windows and the
"applying a host-conditional control never upgrades the policy-level
guarantee" rule are both still present in both files. The rework added no new
`kernel` or `guarantee` sentence.

## Still open — carried forward, not fixable in this task

- **H3** — decision 0008 line 117 requires an audit-record extension for the
  policy identity. `profiles/manager.md` section 7 states normatively that the
  record carries the effective execution-policy identity, but
  `audit-record-v1.schema.json` is `additionalProperties: false` and no task
  owns the change. Implementable today only through the open `audit` object,
  so no schema version pins it and no vector can assert it. **Owned by
  nobody** — needs routing to the schema/vectors track or the landing task.
- **H4** — `host-capability` platform-case class still defined in
  `platform-cases.tsv` as a filesystem limitation, not a missing kernel
  feature. Host-conditional controls route to it. For TASK-260822-f4qv7w.
- **H5** — Windows Job Object nesting unconfirmed, carried from rc5 where the
  identical claim already ships. No new exposure.
- **New, from R1** — f4qv7w must not write a "name outside the reserved set
  passes through" negative vector; see the R1 consequence note above.
