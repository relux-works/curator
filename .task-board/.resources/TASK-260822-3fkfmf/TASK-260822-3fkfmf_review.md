# TASK-260822-3fkfmf review — manager-profile-and-security-prose

**Verdict: changes requested → `to-dev`.** Four fixes, all inside the two files this task
owns, all cheap. Everything substantive that landed is sound; the branch is close.

## What was reviewed

Repo `curator-spec`, branch `spec/sw-manager-security`, commits `b9ca2ad` then `c2371d3`,
both signed (`G`) and pushed to `origin`. Base `b92b105`. Diff is exactly 2 files,
+476/-4. Worktree clean apart from untracked `.temp/`.

Cross-checked against the sibling branches as they stand now:
`spec/sw-core-prose` @ `41cf556` (protocol/core.md §4.1.1) and `spec/sw-schema` @ `ebfed81`
(manifest schema 8).

## Gates — re-run standalone by the reviewer on `c2371d3`, not taken on report

| Gate | Result |
|---|---|
| `python tools/validate.py` | exit 0 — 49 schemas, 471 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | exit 0 — 91 tests |
| `go test ./tools/...` | exit 0 |
| `go run ./tools/generate-vectors -root .` + `git diff --exit-code` over `conformance/v1` and `release/1.0.0-rc.5/6/7/8.json` | exit 0 — deterministic |
| `gofmt -l tools` | empty |
| `git diff --check` | exit 0 |

`validate.py` covers local links, so internal cross-references are gate-verified. The added
prose introduces no new external links, so the lychee job is unaffected.

## AC — met

- Both files committed on the story branch: yes.
- Gates pass: yes, re-verified above.
- No kernel-guarantee claim for portable enforcement: **verified mechanically**. Every
  occurrence of `kernel` in the added lines (6 total) and of `guarantee` (15 total) is a
  negation, a deferral, or a prohibition. No sentence asserts the portable policy denies,
  prevents, or confines at the kernel level on any platform. The explicit macOS/Windows
  "no kernel confinement primitive at all" statements and the "applying a host-conditional
  control never upgrades the policy-level guarantee" rule are both present in both files.

## Cross-document consistency — verified against core.md §4.1.1

Checked identifier by identifier; the `c2371d3` reconciliation is accurate:

- seven deferred guarantee names — exact match, same order, same spellings;
- `script-worker-v1-native-control-inventory-v1` table — all 8 rows × 3 platforms match
  core.md character for character;
- closed unavailable-reason vocabulary — 5 entries, exact match;
- three availability values and their applied/unavailable rules — match;
- `script-capability-evidence-v1` field set and per-entry field set — match;
- audit warning classes `script-command-declared-only` /
  `script-command-unfiltered-declared-network` — kebab-case, match;
- four-node graph with the first three under `exec: "none"` — match;
- `operation-private`, `script_execution_policy_unsupported`,
  `script_execution_hardened_claim_forbidden` — match;
- manifest field `execution_policy` const `script-worker-v1` and `interpreter` enum
  `{node-v1, python3-v1}` — matches `spec/sw-schema` `common.schema.json`, and the reserved
  sets cover exactly those two interpreter identifiers.

`decision-0006` is `0006-portable-manager-worker-execution.md` — the SECURITY.md citation is
correct. §2.2.1 and §7 exist and 3.1/§7 are placed correctly. SECURITY.md's new section is a
clean structural peer of `## Compile-only build boundary`; there is no TOC to update.

The analysis finding F3 from TASK-260822-1l4r4f is genuinely resolved: the §3 carve-out sits
immediately after the "preserves the inherited `PATH`" paragraph and before the symlink /
POSIX-shell / `.cmd` launcher forms it excludes.

---

## Blocking findings — fix on this branch

### R1. The closed reserved environment-name set is narrower than the criterion core.md makes normative

`profiles/manager.md:654-676`.

core.md §4.1.1 states an **open criterion** and delegates only the enumeration:

> The manager owns every name it sets under this policy and every name that selects a
> program, a library or module search path, an interpreter startup file or option, a
> temporary or configuration root, or proxy or resolver configuration. […] The exact
> reserved set per platform and per interpreter identifier is defined by the manager profile.

manager.md then declares the set **closed**. Two classes of name satisfy core.md's criterion
and fall outside the closed set, so the two documents contradict each other and an `env_read`
entry naming one passes the inherited value through under manager.md's own rule:

- **Linux dynamic-loader variables.** macOS gets prefix closure — "every `DYLD_`-prefixed
  name". Linux gets three enumerated names: `LD_PRELOAD`, `LD_LIBRARY_PATH`, `LD_AUDIT`.
  glibc honours more for non-setuid processes, and at least `LD_ORIGIN_PATH` (`$ORIGIN`
  expansion, i.e. library search path) and `LD_TRACE_LOADED_OBJECTS` (the loader does not run
  the program at all) plainly match core.md's criterion. `LD_PROFILE_OUTPUT` and
  `LD_DEBUG_OUTPUT` additionally direct loader writes at a caller-chosen path, which the
  operation-private-roots mechanism is otherwise supposed to own.
- **Windows `SystemRoot`.** `WINDIR` is reserved; its twin `SystemRoot` is not, and
  `SystemRoot` is the spelling the loader and the Win32/WinSock/crypto APIs actually consume.
  Reserving one and not the other is not defensible on the stated criterion.

This is also inconsistent with the implementer's own stated design principle — the set was
made prefix-based (`DYLD_`, `PYTHON`, `NODE_`, `NPM_CONFIG_`) "so the closed set survives
interpreter evolution" — and `LD_` is the one loader family left enumerated.

Severity is defence-in-depth, not a package-controlled escape: the package chooses the *name*
via `env_read`, but the *value* still comes from the caller's environment, so exploiting it
needs a hostile manifest and a poisoned caller env together. It is still a hole in the exact
artifact this task was delegated to close, and the fix is two lines.

**Fix:** make the loader entry prefix-based on Linux the way macOS already is — "every
`LD_`-prefixed name, including `LD_PRELOAD`, `LD_LIBRARY_PATH`, and `LD_AUDIT`" — and add
`SYSTEMROOT` to the Windows list. If a genuinely closed enumeration is wanted instead, say
explicitly that the set is the closed *minimum* and that core.md §4.1.1's criterion continues
to govern names outside it; do not leave the two documents asserting different things.

### R2. Broken nested code span mangles the process graph in SECURITY.md

`SECURITY.md:350-353`. The sentence wraps the whole graph in one backtick span and then puts
a second backtick pair around `exec` inside it:

```
The process graph is `manager parent -> ... -> manager-resolved executables named by
the `exec` capability`, and exactly the first three nodes under
```

Six backticks pair left to right, so it renders as: a code span ending at "named by the ",
then bare unstyled `exec`, then a stray code span containing just " capability". The graph
that this whole section turns on is unreadable in any Markdown renderer, in a public-facing
SECURITY.md. Both `protocol/core.md` and `profiles/manager.md` render the identical graph
correctly in a fenced ```` ```text ```` block; SECURITY.md is the only copy that inlines it.

**Fix:** either drop the inner backticks around `exec`, or use the same fenced block the other
two files use.

### R3. The §3 carve-out sentence over-scopes "declared-only commands"

`profiles/manager.md:511`: "The launcher rules of this section govern declared-only commands."

`declared-only` is defined in 3.1 as a *script command* category. Section 3 also carries the
build command launcher rule and the system command / shim rules, and 3.1 itself says "Build
commands are outside this policy entirely; […] their launcher is unchanged". As written, the
carve-out reads as removing build and system commands from section 3's launcher rules
entirely, which contradicts both the rest of section 3 and 3.1.

**Fix:** "The launcher rules of this section govern every command other than an enforced
script command." (or "…govern declared-only script commands, and are unchanged for system and
build commands").

### R4. Two added lines break the file's ~80-column wrap

- `SECURITY.md:421` — 120 characters. Nothing in the prose of any of these files is close to
  that; the longest pre-existing prose lines in SECURITY.md are 81-82.
- `profiles/manager.md:942` — 85 characters.

Not gated (there is no Markdown linter in CI), but plainly out of family for a 476-line prose
change that is otherwise wrapped cleanly. Two other added lines sit at 81, which is within the
existing tolerance and needs no change.

---

## Non-blocking — hand forward, not fixable inside this task

### H1. Linux `active-process-count-limit` = `available: RLIMIT_NPROC` looks wrong

Raised by the implementer and I agree it is a real problem, but it is **correctly out of scope
here**. `RLIMIT_NPROC` bounds processes per real UID across the whole session, not a private
aggregate domain — which is exactly the reason string (`no-private-aggregate-domain`) the same
row uses to mark macOS unavailable, and macOS has `RLIMIT_NPROC` too. Since the inventory
requires an `available` control to be applied *and reported applied*, that cell manufactures an
applied claim for a control that does not bound the invocation's domain.

Note the implementer's write-up says the rc5 ledger marks macOS unavailable "for that same
primitive": true, but rc5-native-control-inventory-v1 has **no Linux column at all**
(macOS/Windows only), so this is new content originating in core.md §4.1.1, not a carried-over
rc5 verdict. Mirroring core.md verbatim rather than diverging unilaterally in one of three
copies was the right call.

**Route to:** the landing task, changing core.md §4.1.1, manager.md 3.1, and
`conformance/v1/vectors/script-host-execution-policy.json` in one commit. Recommended
replacement: `host-conditional: delegated cgroup v2 pids.max`, which is what actually gives a
private aggregate domain and matches how the adjacent `aggregate-memory-limit` Linux cell is
already spelled.

### H2. "The complete diagnostic set" vs seven codes

core.md §4.1.1 says "The complete diagnostic set of this policy is:" and lists four.
manager.md 3.1 carries seven — the four plus `script_execution_worker_identity_invalid`,
`script_execution_worker_protocol_invalid`, `script_execution_package_influence_forbidden`.

I checked the build-side precedent and it does not license this: core.md §4.2.1 never claims
completeness, which is why manager.md 2.2.1 can add the three equivalent `build_execution_*`
session codes without contradiction. §4.1.1 does claim it, so the two documents are in genuine
tension. manager.md's framing ("the first four are the policy-level set […] the last three are
worker-session codes of this profile") acknowledges the split but does not dissolve it: all
seven surface on the same `phase: execution` invocation results.

Nothing enumerates these as a closed set yet — I confirmed `spec/sw-schema` @ `ebfed81` adds no
`script_execution_*` enum — so this lands on TASK-260822-f4qv7w, which will have to pick one.
Cheapest correct fix is softening §4.1.1 to "the policy-level diagnostic set" and letting the
profile add session codes, exactly as 4.2.1/2.2.1 already do.

### H3. Decision 0008's audit-record extension is unaddressed anywhere

`decisions/0008-enforced-script-capabilities.md:117` lists "audit-record extension for the
policy identity" as a required consequence. manager.md §7 now states normatively that "for
every script command the record carries the effective execution-policy identity or its
absence" — but `schemas/v1/audit-record-v1.schema.json` is `additionalProperties: false` and
was not touched by the schema task.

It is implementable today only by stuffing the identity into the open `audit` object, so no
schema version pins it down and no vector can assert it. Not this task's defect — the schema
change belongs to the schema/vectors track — but it is currently owned by nobody.

### H4. `host-capability` platform-case class (carried from TASK-260822-1l4r4f finding 3)

Still open. Host-conditional controls route to a platform-case class that
`platform-cases.tsv` defines as a filesystem limitation, not a missing kernel feature.
f4qv7w needs a widened or new class.

### H5. Windows Job Object nesting

Unconfirmed, carried from rc5 where the identical claim already ships. No new exposure from
this change. Noted only so it is not lost.

## Minor polish — optional, fold into the R1-R4 cycle if convenient

- `profiles/manager.md` mandatory bullet says "when `network` is `"none"`" where core.md and
  SECURITY.md both say "when the **derived** network capability is `none`". Not wrong — the
  absent-field case is covered by obligation 4's deny-by-default derivation — just looser than
  the two documents it sits between.
- `profiles/manager.md` mandatory bullet 6 says "private **per-command** configuration and cache
  roots" where core.md §4.1.1 and this change's own SECURITY.md text both say
  "**operation-private**". Those are different lifetimes. Pick one spelling; core.md wins.
- SECURITY.md generalises core.md's TCB rule — core conditions it on "an implementation that
  cannot distribute an equivalent identity-verified worker", SECURITY.md states it for "any
  implementation whose worker graph contains an interpreter and an installed package tree".
  The substance is the same and the broader statement is true; flagging only for exactness.

## Assessment of the work itself

The substance is strong and the reviewer-facing write-up was honest and accurate — every claim
in it that I checked held up, including the self-reported findings, and the one place it
overstated something (H1's rc5 provenance) it overstated *against* itself. The 2.2.1 mirroring
is real rather than cosmetic: the three deliberately-not-carried build controls, the single
failure boundary, the package-influence boundary, and the diagnostics table all have genuine
counterparts. The enforcement-vs-guarantee split in SECURITY.md is the right shape and the
qualification of the blanket "not runtime sandboxes" sentence is exactly the surgical edit that
paragraph needed.

R1 is the only finding with teeth, and it is teeth precisely because core.md delegated that set
here and nowhere else. R2-R4 are mechanical.
