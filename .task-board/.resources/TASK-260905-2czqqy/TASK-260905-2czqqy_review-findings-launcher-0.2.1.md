# Review findings: launcher SPEC `0.2.1-draft` (cycle 1)

Reviewer run `RUN-260905-f284fe`, read-only.

Subject: `curator-agent-launcher`, worktree `.worktrees/curator-agent-launcher-spec-0.2.1`,
branch `draft/spec-0.2.1`, head **`484933b`**, one commit past `main` `e19eb9f`.
Scope reviewed: `git diff e19eb9f..484933b` — `SPEC.md` (+179/−41), `README.md` (+2/−2),
`cmd/curator-run/main.go` (+1/−1), `cmd/curator-run/main_test.go` (+1/−1). `git diff --check` clean.

Authority read at curator-spec `f39f4a9` (current `main`). `protocol/environments.md` is
**byte-identical** between the producer's cited `fcdb9ba` and `f39f4a9` (`git diff fcdb9ba..f39f4a9
-- protocol/environments.md` is empty), and `profiles/manager.md` changed in exactly one hunk at
§1 (`system-config-v2` locked keys) — §12.5 untouched. The draft's normative-reference bullet
citing `fcdb9ba` therefore names the exact bytes this review checked; no citation is stale.

## Verdict: ACCEPT (no blocking or major findings; four minors below, fold into the next touch)

## Provenance: the committed tree equals the producer's staged tree

The producer could not sign (passphrase-protected key, headless session) and left the change
staged; the orchestrator committed it. Verified rather than taken on trust:

| Check | Result |
|---|---|
| `git diff --binary e19eb9f..484933b` vs. attached `TASK-260905-2czqqy_spec-0.2.1.patch` | **byte-identical**, 28352 bytes each (`diff` exit 0) |
| Commit count past `e19eb9f` | exactly 1 |
| Signature | `git verify-commit 484933b` → `Good "git" signature for ivan@relux.works`, ED25519 `SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0`; `%G?` = `G` |
| Author / committer | `Ivan Oparin <ivan@relux.works>` for both — the repository's human identity, not an agent identity |
| Scope | exactly the four files the brief allows; no LOGBOOK, no control-root write |
| curator-spec story worktree | `task-board worktree status`: tree clean at base `fcdb9ba`, `repository_delta=empty` — empty by design, as briefed |

Note, not judged (cycle-2 precedent): the branch **is** on `origin` at `484933b` and **PR #3 is
OPEN** (created 23:02:34Z, ~9 s before this run started). The drafting report's "Nothing pushed,
tagged, or PR'd" was true when the producer wrote it — the producer never made the commit. Push
and PR are the orchestrator's delivery step. No tag points at `484933b`. PR head `484933b` = the
reviewed head; PR file list = the same four files; `statusCheckRollup` empty (this repository
configures no hosted checks).

## The four brief items: present, exact, and authority-consistent

| # | Item | Where | Authority checked against | Verdict |
|---|---|---|---|---|
| 1 | `--repair` on every resolve | §4.1, §1 non-goal, §6 resolve row, README line 10 | environments §9.2 step 5 (line 1643: "`curator run` always passes `--repair` and repairs the home instead of surfacing it"), §10.1 (read-only default, lock-free verification first, current home emits with **no** lock, only a stale home takes the mutation lock with a bounded wait, `environment_home_stale` fail-closed with **no fragment**), §10.4, manager §12.5 | **exact**; no clause contradicts §10.1 |
| 2 | cycle-2 residual minors 1–5 | §4.6, §6, §9 | `TASK-260905-3ewdq0_review-verdict-cycle2.md` | **all five applied** (see table below) |
| 3 | codex layer stat + single `-p` | §4.5, §4.6, §6 `mcp` family, §9 | environments §7.8 codex_cli row ("`-p` accepts **exactly one** value … a **missing layer file is silently ignored** (exit 0) — under `--strict-config` too … the launcher MUST stat the layer file immediately before exec and fail rather than launch without the set"), §5.8 (`<home>/curator-mcp.config.toml`, marker-recorded), §10.2 (`mcp.path` shape), §7.1 (`CODEX_HOME=<home>`) | **exact**, including the recorded consequence that closes Decision 0012 OQ3 |
| 4 | file family vs. environments §12.1 | new §4.7, §4.3 pointer, §9 | environments §12.1 ("carried by `manager-config` schema 2 … under one `environments` object"), §10.3, §12.2 | **holds**; checked knob by knob |

Cycle-2 residual minors: (1) `enabled: false` = not configured — applied verbatim; (2) the
machine-over-operator inversion now carries its rationale ("whether sessions on this machine are
tracked is machine policy"); (3) §6 `defaults` row now names `ax.json` and §4.6/§4.7 — and its
wording is improved beyond the ask, from "a machine-configuration file" to "a launcher-owned
configuration file", which also fixes the machine/operator conflation; (4) §9 bullet now covers
both files; (5) the ordering note is pinned ("because that read precedes argument validation it is
the diagnostic that fires when the command line would also have been a usage error").

Item 4 attacked knob by knob against the §12.1 table: no launcher knob (model default, effort
default, `ax` switch) appears in it, and every §12.1 knob that shapes a launch reaches the launcher
only through the fragment — `passable_env_names` → `mcp.env_names` bounded before the launcher sees
it (§10.3, confirmed at environments line 2094: "bounded twice before the composer sees them");
`precedence` → the fragment member §4.1 accepts and does not consume; `system_prompt_files.pi` →
detected on disk by the §5.1 probe; `current_profile`/`scoped_current` → profile selection;
`overlays*`/`mcp_package_allowlist` → the lock and the resolved MCP set;
`isolation`/`forms`/`in_place_mode` → the managed home the `env` map points at;
`xdg_seed_allowlist`/`targets`/`shadow_acknowledged`/`secret_material_waivers`/`backup_retention`/
`require_current_profile` → manager-side, never launch-shaping for `curator run`. No knob is
duplicated into a launcher file. `f39f4a9`'s new `system-config-v2` does not change this: it locks
the same §12.2 `environments.*` keys, which still reach the launcher only through the fragment.

## Attacks run (behaviour, not reading)

- **Is `--repair` a bypass of the fail-closed rule?** No. §4.1 states there is no read-only launch
  mode, gives the reason ("a launch into a home Curator knows is wrong is the failure `env resolve`
  exists to prevent"), and names the operator's hand path (`curator env resolve` / `curator env
  status`, both real surfaces — environments §7.1, §12.2). No flag re-opens it.
- **Can `environment_home_stale` reach a `--repair` caller?** environments §10.4 makes it the
  no-`--repair` diagnostic. §4.1 declares it unreachable **and still handles it** as
  `resolve_invocation_failed` — fail-closed on a Curator that misbehaves. Correct direction.
- **Does the lock claim overreach?** No. §4.1 says the lock is taken *only* when verification finds
  the home stale — matching §10.1 exactly. `resolve_lock_unavailable` states its remedy (retry
  later) and that the launcher does **not** retry on its own.
- **Can the codex stat be bypassed?** The one path is an operator `-p` after `--` when the fragment
  has **no** `mcp` section: the launcher spells no `-p`, does not stat, and codex layers whatever
  the operator named. §4.5 states this explicitly and grounds it in §3 (everything after `--` is
  uninspected). A recorded consequence, not a hole.
- **Is `env_names`/seed reconciliation newly reachable under `--repair`?** `environment_seed_shadowed`
  and the XDG seed path are `opencode`-only (environments §7.1), and `opencode` is `env_unsupported`
  for `curator run` (environments §10.1, launcher §4.2/§9). Unreachable. Dissolved.
- **Does the tracked handoff invalidate the local pre-launch checks?** Checked against
  `agent-session-manager-spec` §14.1: `ax start NAME --provider ID [--profile] [--workspace PATH]`
  takes no `--to HOST`; `--to` is permitted only with `takeover`/`fork`. `ax start` is local, so
  the binary check, the §4.5 stat and the §5.1 probe test the filesystem the child will use.
  Dissolved — no finding.
- **Diagnostic-code collision / orphan sweep:** every code used anywhere in `SPEC.md` appears in the
  §6 table and vice versa (18 codes, set difference empty, computed not eyeballed). The three new
  codes collide with nothing. Exit class: §6's preamble ("usage errors exit 2; every operational
  failure exits 1") puts all three at exit 1. §6's resolve row lists six codes and six condition
  clauses **in matching order**; the `mcp` row likewise.
- **Cross-reference sweep:** every `§N.M` in `SPEC.md` resolves — self-references against the
  document's own section set `{1,2,3,4,4.1–4.7,5,5.1,5.2,6,7,8,8.1,9}`, foreign ones against
  environments.md / manager.md / ax. No dangling reference; §4.7 is reachable from §4.3, §4.6 and §9.
- **Version fact:** `0.2.1-draft` agrees across `SPEC.md` line 3, §8, the §8.1 row, `README.md`
  line 24, `main.go:18` and `main_test.go:52`. No `0.2.0-draft` string survives outside the §8.1
  history row.

## Gates rerun by this reviewer (scratch `git archive` of `484933b`; subject worktree never written)

```text
go build ./... ; go vet ./... ; go test ./... -count=1
ok  github.com/relux-works/curator-agent-launcher/cmd/curator-run  0.233s
make check exit=0
```

Mutant table — three narrowing mutants, one reversion, two blind-spot probes:

| # | Mutant | Narrows / probes | `go test` | Named failing test |
|---|---|---|---|---|
| M1 | `main.go` `specVersion` → `0.2.0-draft` | the version pin (reversion) | **exit 1** | `TestSpecVersionPinned` |
| M2 | `README.md` `0.2.1-draft` → `0.2.0-draft` | does the pin cover README? | **exit 0** | *none — blind spot* |
| M3 | `SPEC.md` line 3 → `0.2.0-draft` | does the pin cover SPEC? | **exit 0** | *none — blind spot* |
| M4 | `if len(args) == 1` → `>= 1` | the arg-shape gate narrowed from "exactly one" to "at least one", so `--version extra` returns 0 | **exit 1** | `TestRunRejectsEverythingElse` |
| M5 | refusal `return 2` → `return 1` | the refusal's exit class | **exit 1** | `TestRunRejectsEverythingElse` |
| M6 | refusal text loses `not implemented` | the refusal's message contract | **exit 1** | `TestRunRejectsEverythingElse` |

M4–M6 are narrowings, not deletions: the gate still exists and still refuses, it just refuses a
smaller class or with a different class/message. All three are caught. The stub's negative tests
bind.

**AC-row coverage, measured.** Five AC rows. One (`stub/README version bumped`) has a named driving
test, and M2/M3 measure its reach at **1 of the 3 files** it claims. One (`make check green`) is
rerun above. One (`SPEC.md 0.2.1-draft with the four items`) has **no executable driving test** and
cannot have one in a prose-specification repository — verified instead against the authority
documents, quoted above, line by line. One (`independent review ACCEPT`) is this run. One (`landed
by fast-forward`) is the orchestrator's step and is out of this review's scope. The producer stated
this bound honestly in the drafting report; M2/M3 turn the stated bound into a measured one.

## Minor findings (non-blocking; fold into 0.2.2 or the next touch)

### 1 — minor / §4.1, §6 resolve row: `--repair` widened resolve's failure surface, but not the pass-through

> "any other non-zero exit is `resolve_invocation_failed`" (§4.1)
> "`resolve_invocation_failed` … `curator` not startable or an unexpected non-zero exit" (§6)

Making `--repair` unconditional turns every resolve into a potential journaled mutation
transaction, so resolve can now fail with diagnostics that environments §10.4 does not list and
§4.1 therefore does not map: `environment_marker_invalid`, `environment_surface_unmanaged_conflict`,
`environment_backup_exists` (§8.5), `environment_seed_unreadable` (§7.7). All of them collapse into
`resolve_invocation_failed`, whose §6 gloss leads with "`curator` not startable" — so an operator
whose managed home has an unmanaged-file conflict is told the launcher could not start `curator`,
and the actual remedy is lost. The behaviour is safe (terminal, fail-closed, exit 1); the
diagnosability is not.

The asymmetry is visible inside §6 itself: the `plan` family surfaces the verdict's "structure and
evidence … verbatim" and the `ax` family passes "its Structured Error … through" — `resolve` is the
only plane that states no pass-through, and it is the plane this revision gave a mutation to.

**Fix:** one clause in §4.1's failure-mode list — Curator's own diagnostic code and message are
printed verbatim beneath the launcher's code line, exactly as `ax_handoff_failed` does.

### 2 — minor / §4.6: three simultaneous pre-launch checks, one required diagnostic line, no order

> "The three pre-launch checks — this binary check, the §4.5 codex layer stat, and the §5.1
> file-kind probe — all run immediately before the handoff or exec, in both modes" (§4.6, new)
> "A failing launch prints exactly one diagnostic code line to stderr" (§6)

The new sentence is right to group them, but it makes a gap explicit that 0.2.0 only implied with
two checks: a codex launch into a home with a missing binary *and* a missing layer file *and* an
unreadable `pi` file has three eligible terminal codes and no stated precedence, while §6 admits
exactly one line. Two conforming implementations can report different codes for the same state,
and a code is the contract here, not the prose.

**Fix:** state the order in that sentence — "…run in this order, and the first failure is the one
reported" — with an order chosen deliberately (the binary check first is the natural one: it is the
cheapest and the only one that is not about the managed home).

### 3 — minor / §4.5, §9: the silent-swallow class is closed for codex and unrecorded for the others

The stat rule exists because of a **verified** codex fact — a missing `-p` layer is silently
ignored, exit 0, under `--strict-config` too. The symmetric question for the other argv-carried MCP
channel is neither verified nor recorded: what does `claude_code` do with `--mcp-config <missing
path> --strict-mcp-config`? environments §7.8 verifies only that both flags exist and that strict
mode ignores every other MCP configuration — which, if a missing file is also tolerated, is the
*same* silent failure with a worse blast radius (zero servers, and nothing else to fall back to).
`opencode`'s `OPENCODE_CONFIG` variable has the same open question.

§9's new bullet does bound the revision, but it bounds the wrong axis: it says the stat "closes the
environments.md §10.1 residual window for one file only" and that other surfaces are "verified at
resolve time and not re-verified before exec". That is the residual-window framing. The reason
codex needs a stat is not the residual window — it is that the tool swallows the absence. An
unknown about the sibling adapters should be reported as unknown rather than left to be inferred
from the residual-window sentence.

**Fix:** a §9 item — whether `claude_code` and `opencode` also proceed silently when their MCP
configuration file is absent is unverified at the pinned releases; if either does, §4.5's stat rule
generalizes to it. Arguably this belongs upstream in environments §7.8 as well, whose per-adapter
rows are where such facts are recorded.

### 4 — minor / `cmd/curator-run/main_test.go:49–50`: the one executable gate overclaims its coverage

> "// TestSpecVersionPinned fails when the stub's reported specification version
> // drifts from the version SPEC.md and README.md state; the three are one fact."

Measured false. M2 and M3 above set `README.md` and `SPEC.md` back to `0.2.0-draft` and `go test`
exits **0** both times: the test compares `specVersion` to a literal in the test file and never
opens either document. It covers 1 of the 3 files it names.

This is pre-existing text (the diff only changed the `want` literal), so it is not a defect this
change introduced — but every revision of this repository is a three-file version bump, which makes
this the gate most worth making true, and this task is exactly the failure mode it claims to catch.

**Fix:** read `SPEC.md` and `README.md` in the test and assert `specVersion` appears in both — a
few lines, no new dependency, and the comment becomes accurate.

## Nits (no action required)

- §4.1 calls the verification-to-first-read window "environments.md §10.1's recorded residual".
  §10.1 records it for "the window between a **completed repair** and the child process's first
  read". The draft's version is broader and errs safe (it also covers a current home that took no
  lock), so it is an acceptable restatement — but it attributes to §10.1 slightly more than §10.1
  says.
- The same sentence lists "the §4.6 binary check" among what the launcher re-verifies in that
  window; resolve never verified the tool binary, so it is not a *re*-verification.
- §4.5 classifies a dangling symlink as `mcp_layer_unreadable` while "no file at the path" is
  `mcp_layer_missing` — under a following `stat(2)` those are the same `ENOENT`, so an implementer
  needs `lstat` to tell them apart. This mirrors the landed §5.1 convention word for word, so
  applying it here was the right call; the ambiguity is the house convention's, and both outcomes
  are terminal at exit 1, so nothing observable turns on it.
- The `defaults` family name now covers `ax.json`, which is not about defaults. The cycle-2 review
  already accepted reusing `defaults_config_invalid` for it; noted only so the naming is a
  deliberate carry-over rather than a drift.

## Deviations the producer flagged: reviewed and accepted

Three diagnostic codes beyond the brief's list — `resolve_lock_unavailable`, `mcp_layer_missing`,
`mcp_layer_unreadable`. Keep all three. environments §10.1 makes lock timeout and repair failure
"distinct" with different remedies in the same sentence, so folding `environment_lock_unavailable`
into `resolve_repair_failed` would erase a distinction the authority draws explicitly and tell an
operator to fix a store that is fine. Splitting the stat's two outcomes is §6 invariant 1 applied
correctly: absence and read failure are different facts, and the draft extends the invariant to
say so where absence is *itself* the failure. Fewer codes would be a worse specification.
