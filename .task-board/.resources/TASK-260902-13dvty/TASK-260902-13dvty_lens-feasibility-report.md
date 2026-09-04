# Lens C report — implementation feasibility and cross-spec contract realism

Task: TASK-260902-13dvty. Corpus reviewed at curator-spec main `4d55698` (worktree == main),
curator-agent-launcher SPEC 0.1.2-draft, agent-session-manager-spec v0.5.0 + open PR #1,
skill-agents-management SKILL.md + pkg/agentic source (GitHub main), curator reference
implementation (read-only). Binaries verified on this machine: claude 2.1.258,
codex-cli 0.151.0, pi 0.84.2, gemini 0.54.4; opencode not installed; no Xcode
CodingAssistant probe path present. Spec repo validation ran green here:
`tools/validate.py` — 57 schemas, 773 vector files; `make regenerate-check` — no diff;
`go test ./...` — ok.

Findings are ranked by cost-of-being-wrong before expensive implementation.
Severity | area | quote/section | why | proposal | MUST-before-implementation?

---

## C1 — CRITICAL — ax handoff: the launcher's mandatory route does not exist

**Quote.** Launcher SPEC §4.5: "With the `ax` integration configured on the machine, the
launcher ALWAYS routes the composed launch through `ax`'s instrumentation … A handoff that
fails is `ax_handoff_failed`; the launcher MUST NOT fall back to an untracked direct exec."
Decision 0010 §10: "An environment expressed as literal variables is therefore directly
injectable into `ax`-managed sessions without any new `ax` mechanism."

**Why it is a gap.** Verified against ax SPEC v0.5.0: the only session-creating operation is
`ax start NAME --provider ID [--profile standard|yolo] [--workspace PATH]` (§14.1). It accepts
no argv, no env literal, no extension value, no model, no effort. The Session Record's Launch
Plan is built inside ax at start (§13.1 step 2), and provider `launch` is required to receive a
plan that "MUST exactly equal their Session Record values" (§7.5) — so injection can only
happen at record creation, where no input surface exists. The open PR #1 adds only (a) three
Session Record extension keys and (b) prose that "the launcher merges the `env` variables of
the resolved Curator `launch-env-fragment-v1` object into the Launch Plan and resulting
`SpawnPlan` `env_literals` map" — but "the launcher" is an actor with no interface anywhere in
the ax spec: no CLI flag, no RPC operation, no bridge op accepts a caller-supplied Launch Plan.
Decision 0010's "without any new ax mechanism" claim confuses *representability* (env_literals
can hold the values) with *reachability* (no operation lets an external caller put them there).
Consequence: on every machine with ax configured, every `curator run` launch MUST fail
(`ax_handoff_failed`, fallback forbidden) — the flagship integration is unimplementable as
specified.

Additionally, Launch Plan / SpawnPlan carry **no stdin member**. agents-management plans carry
`Stdin` bytes and effort/prompt transport is per-system `argv | stdin | none`; a plan whose
transport is stdin cannot traverse ax at all even after the input surface exists.

**Proposal (minimal concrete contract).** One ax spec change, delivered as a revision of PR #1:

- Extend `ax start` (direct form) with `--launch-plan FILE|-`: a caller-supplied closed object
  `{argv_suffix or argv, env_literals, extensions}` validated under the existing §5.1 grammar,
  count, byte, and secret rules. ax embeds it verbatim into the Session Record Launch Plan at
  creation; provider plugins carry argv/env_literals through into the SpawnPlan (already
  implied by "plan MUST equal record"). The `works.relux.curator.*` keys ride in `extensions`.
- Ownership of the merge: **curator-run owns composition** (agents-management plan + fragment
  env + operator native args → the launch-plan payload); **ax owns validation and record
  immutability**; **the provider plugin owns translation to SpawnPlan** and MUST NOT rebuild
  argv it was handed. This keeps ax free of Curator and agents-management knowledge.
- Stdin: revision 1 of the launcher refuses the ax route for a plan with non-empty stdin
  (`ax_handoff_failed` with a distinct detail), or SpawnPlan grows an optional
  `stdin:bytes|null` member — the latter is a larger ax change and should be decided
  explicitly, not discovered mid-implementation.
- Launcher SPEC §9 already admits the shape is unspecified; it must additionally state that
  the behavioral contract is unimplementable until the ax change lands, so nobody builds
  against the current PR believing it sufficient.

**MUST be resolved before implementation starts** (of the launcher's ax path; curator-side
work is unblocked).

---

## C2 — CRITICAL — spawn plane: agents-management has no interactive launch mode

**Quote.** Launcher SPEC §1: "The operator types `curator run codex_cli --profile companyA --
resume --last` and gets exactly the codex they know." §4.1/non-goals: "The launcher never
reconstructs argv from its own knowledge of a tool."

**Why it is a gap.** Verified against skill-agents-management source (pkg/agentic): the closed
LaunchMode set is `LaunchModeExec`, `LaunchModeDryRun`, `LaunchModeManagedSession` — all
shapes for running a tracked *assignment*. The claude system's single argv construction site
(`systems/claude/args.go`) emits, for every allowed mode:
`-p --output-format json --model <id> [--effort <e>] [--max-budget-usd <x>]
--dangerously-skip-permissions [goal pair]`. There is no mode that produces a plain
interactive invocation. Consequences for `curator run`:

- Consuming BuildPlan as a value yields a **headless print-mode run with permission
  bypass**, not "the claude they know" — and `--dangerously-skip-permissions` injected into an
  operator's interactive terminal session is a safety-relevant wrong default, not just a UX
  miss.
- Appending operator native args after that argv (`resume --last` after
  `-p --output-format json …`) is incoherent.
- The two launcher non-goals ("no plan rebuilding" + "exactly the tool they know") are jointly
  unsatisfiable against the module as it exists.

**Proposal.** Add `LaunchModeInteractive` to agents-management: per-system argv containing
only model selection and effort transport (no print mode, no output format, no permission
bypass, no goal/assignment machinery; stdin empty), admission verdicts unchanged. This keeps
the module's invariant 5 (one argv construction site per plugin) and keeps the launcher free
of tool flag knowledge. The alternative — launcher-owned interactive argv — breaks the
launcher's own §1 non-goal and creates a second flag-spelling site; reject it. The launcher
SPEC §4.1 should then name the mode it requests.

**MUST be resolved before implementation starts** (it defines what BuildPlan is asked for).

---

## C3 — CRITICAL — determinism: snapshot acquisition is not byte-exact, and Windows default git config kills every profile

**Quote.** environments §3: "There is no normalization path: a violating module fails the
snapshot, it is never rewritten." §5: "Identical inputs MUST yield byte-identical output on
every platform and in every mode." Core §6.2: "Snapshots are immutable regular-file trees
produced from that commit."

**Why it is a gap.** Core §6.2 never requires the snapshot bytes to equal the committed blob
bytes, and the reference implementation demonstrably does not produce them. Verified
empirically on this machine (git, scratch repo):

- `git -c core.autocrlf=true archive --format=tar HEAD` emits `line1\r\nline2\r\n` for a
  committed-LF blob — **git archive applies autocrlf conversion**. `core.autocrlf=true` is the
  Git-for-Windows installer default, so on a stock Windows machine every LF-validated context
  module extracts as CRLF → `profile_module_bytes_invalid` → **every git-sourced profile fails
  installation on that platform**. The lens hypothesis "the snapshot path may avoid checkout"
  is false: curator `internal/gitops.Archive` runs plain `git -C repo archive --format=tar
  <commit>` with no config neutralization.
- `git archive` expands `export-subst` attributes: a committed `$Format:%H$` extracts as the
  commit hex (verified). Installed bytes ≠ committed bytes, keyed to attributes the profile
  author controls.
- Audit hashes the **extracted** snapshot (`hashing.ContentSHA256(subject.Snapshot)` in
  `internal/audit`), so audit and install are locally coherent — but the content hash, `path`/
  `local` state hashes, and therefore effective pins and revocation identities become
  git-config- and attribute-dependent: the same commit yields different pins on different
  machines, breaking §5.6 cross-platform hash equality and cross-machine revocation matching.

**Proposal.** Normative text in core §6.2 (or environments §1, scoped to this capability if
core is frozen): "Snapshot production MUST reproduce the exact committed blob bytes of every
entry of the resolved tree. Working-tree conversion (EOL, smudge/clean filters) and
attribute-driven processing (`export-subst`; `export-ignore` beyond git's archive-format
necessity) MUST NOT alter or omit entry bytes." Add a conformance vector: a fixture tree with
`* text=auto` and an `export-subst` entry whose snapshot must hash to the raw-blob value.
Implementation note for the reference: `-c core.autocrlf=false -c core.eol=lf` does not
disable in-tree attribute processing; the reliable path is object-database extraction
(`git ls-tree -r` + `cat-file`), which also removes the tar intermediary.

**MUST be resolved before implementation starts** (it invalidates every hash identity the
capability introduces; also silently affects the already-shipped skills pipeline).

---

## H1 — HIGH — launcher composes in the wrong order: limit state and plan are keyed to the wrong home

**Quote.** Launcher SPEC §4: plan first (§4.1), fragment second (§4.3), env merge last
(§4.4): "The fragment wins on exactly its own closed names … the override can only re-aim the
tool's home, never alter how the process is launched."

**Why it is a weakness.** agents-management `LaunchRequest.Home` is documented load-bearing:
"On-disk limit state is keyed by (provider, home), so this value is load-bearing beyond the
launch" (pkg/agentic/system.go). The launcher builds the plan before it has the fragment, so
Home is empty → the system's DeclaredHome (`~/.claude`) → admission verdicts are read from,
and limit observations written to, the **native** home while the process actually runs in the
**managed** home. With per-profile `isolated` accounts this is materially wrong: profile A's
rate-limit evidence gates profile B's launches, and vice versa. The module's own invariant
warns that `IdentityKey(provider, home)` "must never move without a demonstrated migration on
real state files" — getting this wrong at first ship bakes in a migration.

**Proposal.** Reorder launcher §4: resolve the fragment (current §4.3) before building the
plan (current §4.1), and pass the fragment's managed-home path as `LaunchRequest.Home`
normatively. The §4.4 merge rule stays as the env-layer statement; add: "the plan request's
configuration home MUST equal the fragment's managed-home path."

**MUST be resolved before implementation starts** (cheap now; a keyed-state migration later).

---

## H2 — HIGH — bare `curator run <env>` cannot launch: nobody owns model/effort defaults

**Quote.** Launcher SPEC §3: `--model` "passed through … The launcher does not validate model
names"; `--effort` "Effort is per-model and required by the spawn plane, which injects no
default." Decision 0010 §6 flagship example: `curator run codex_cli --profile companyA --
resume --last` (no model, no effort).

**Why it is a gap.** Verified: claude `Args` requires `req.Model.ID` unconditionally;
`ErrEffortMissing` refuses a required-effort model with no effort; "No default is injected at
any call site" is a module invariant. The launcher spec defines no default-selection surface
and no configuration file. So the canonical one-liner exits with a plan refusal every time —
the launcher's whole reason to exist ("just run it") is unreachable without two extra flags on
every invocation.

**Proposal.** The launcher spec must own default resolution explicitly: a launcher
machine-config file (its own closed surface) mapping env-id → {model, effort}, with
`vendorplugin.Lineup`'s top admitted pair plus the vendor's recommended effort as the
documented fallback when unconfigured; refusals still surface verbatim. Name the config file,
its schema, and its precedence against the flags. (Note this interacts with C2: an interactive
mode may make model optional for some tools — claude launches without `--model` natively — and
that choice belongs in the same section.)

**MUST be resolved before implementation starts** (of the launcher; it is its primary UX).

---

## H3 — HIGH — pi channel flag spellings are wrong against pi 0.84.2

**Quote.** environments §7.3, pi row: "`flag`/`append`: `--append-system-prompt-file`;
`flag`/`replace`: `--system-prompt-file`". Same spellings reach `launch-env-fragment-v1`
consumers as data a launcher applies blindly.

**Why it is a weakness.** Verified on this machine: pi 0.84.2 rejects both —
`Error: Unknown option: --system-prompt-file` / `--append-system-prompt-file`. The real flags
are `--system-prompt <text>` and `--append-system-prompt <text>`, the latter documented as
"Append text or file contents … (can be used multiple times)" — **polymorphic**: a path whose
file is missing is silently sent as literal prompt text, which is exactly the
read-failure-treated-as-absence class the protocol bans elsewhere (§8.4). The spec's §7.3
hedge ("spellings verify against the pinned tool releases before the conformance vectors
freeze") anticipates drift, but the recorded values are wrong today, and Decision 0010's "pi
takes the same flags … (verified in 0.84.2)" is contradicted by the binary. (claude 2.1.258
checks out: `--system-prompt-file <file>` and `--append-system-prompt-file <file>` are both
recognized. codex 0.151.0 checks out: `model_instructions_file` exists in the config surface,
`-c` override mechanism present.)

**Proposal.** Correct the pi rows to the real spellings and record the polymorphism
normatively: the flag value is a path; the launcher MUST verify the file exists and is
readable immediately before exec and fail (`sysprompt_file_unreadable`-class) otherwise, never
letting the tool interpret a dead path as prompt text. Re-verify against the pinned pi release
at vectors freeze as §7.3 already requires.

**MUST be fixed before the fragment/channel conformance vectors freeze**; blocks launcher §5
implementation, not curator-side materialization.

---

## H4 — HIGH — claude_code `isolated` credentials are not implementable on macOS as specified

**Quote.** environments §7.4: "A profile × environment pair MAY be configured `isolated`: no
passthrough, the tool authenticates fresh inside the managed home — the supported shape for
genuinely separate accounts." And the claude_code row: "macOS: none (Keychain is ambient)".

**Why it is a weakness.** Verified on this machine: the credential is a Keychain generic
password with `service="Claude Code-credentials"`, `account="iv"` — keyed by macOS user, not
by config dir. `CLAUDE_CONFIG_DIR` isolation therefore cannot isolate auth: every managed home
reads the same item ("ambient" is correct for `shared`), and a fresh login inside an
`isolated` home writes the **same** Keychain item, clobbering the operator's ambient
credential for every other home including the native one. An operator using `isolated` for
company/personal separation silently ends up with both profiles on one account — the exact
failure the knob exists to prevent, with billing/data-boundary consequences. (Observed: an
`oauth.claude.profile.<64-hex>` account also exists in the keychain, suggesting a newer
per-profile scheme worth investigating; its keying did not match obvious config-dir hashes.)

**Proposal.** Registry text: claude_code on macOS declares `isolated` unsupported in
revision 1 — configuring it is a configuration error with a dedicated diagnostic (pattern of
`environment_form_unsupported`), not a silently-shared home. Add the platform verification of
the `oauth.claude.profile.*` scheme to the §7.6-style pinned-release checklist; lift the
restriction only on positive evidence.

**MUST be resolved before implementation starts** (of the `isolated` knob; declaring the
limitation is one paragraph now versus a credential-clobbering incident later).

---

## M1 — MEDIUM — auth.json symlink passthrough can be silently severed by token refresh

**Quote.** environments §7.4: passthrough entries shared "by symlink or seeding"; codex/pi:
`auth.json`.

**Why it is a weakness.** OAuth-refreshing tools rewrite auth.json; a tool that replaces the
file by atomic rename replaces the *symlink itself* with a regular file — from that moment the
managed and native homes hold diverging credentials with no diagnostic (passthrough entries
are deliberately outside drift detection). Not verified against codex/pi write paths (unknown,
and it can change per release); the failure shape is common enough to demand evidence.

**Proposal.** The registry's per-entry declaration should record the pinned release's verified
write behavior (in-place rewrite vs rename-over). For rename-over tools, passthrough must be
declared at the directory level or `env status` must include a passthrough-link liveness row
(link still a link, still targeting the native entry) despite passthrough being outside
surface hashes. Verify before vectors freeze; can follow implementation start.

---

## M2 — MEDIUM — repair-on-resolve hot path: unbounded verification and unspecified lock behavior

**Quote.** environments §10.1: "Resolution verifies that the profile's managed home for the
environment is materialized and current under section 8, and repairs it from the store when it
is not." §8.4 makes linked-surface currency require hashing link targets; manager §2.5 makes
repair a journaled manager-home transaction under the exclusive mutation lock.

**Why it is a weakness.** `env resolve` sits on every launch. As written, currency requires
re-hashing every recorded surface (for a profile with a large skills tree, tens of MB per
launch), and any needed repair contends on the single exclusive manager-home lock
(`internal/managerlock` is one file lock) — a running `profile sync`/install blocks every
`curator run` with no specified outcome (wait forever? fail? which diagnostic?). The
transaction engine itself fits the per-entry re-materialization fine; the unspecified part is
resolve's read/lock contract.

**Proposal.** Two clauses in §10.1: (a) the currency check covers exactly the surfaces the
home's marker records, and for symlinked surfaces whose target is an immutable, previously
validated store entry an implementation MAY treat link-target identity as sufficient currency
(the store entry's own integrity is the store's invariant, §4); (b) verification is lock-free;
repair acquires the mutation lock with bounded wait, and lock-acquisition failure is a
distinct diagnostic (not `environment_repair_failed`, which should keep meaning "store cannot
restore this home"). Can follow implementation start; settle before vectors freeze resolve
semantics.

---

## M3 — MEDIUM — codex byte cap can silently truncate a composed monolithic AGENTS.md

**Quote.** environments §5.2/§7.2: monolithic is codex_cli's only form; composition appends
every chain member's modules into one `<home>/AGENTS.md`.

**Why it is a weakness.** codex 0.151.0 reads a global AGENTS.md from CODEX_HOME (string
"Failed to read global AGENTS.md instructions from `" in the binary) and exposes
`project_doc_max_bytes` (present in its config surface; documented default 32 KiB). A composed
profile chain can exceed the cap; the tool then reads a truncated prefix while drift detection
reports the surface current — the determinism story ends at the file and the operator has no
signal. Whether the cap applies to the CODEX_HOME global doc (vs project docs only) is
unverified.

**Proposal.** Adapter registry records each tool's root-context read cap per pinned release;
materialization and `env status` warn when the materialized root context exceeds it (new
warning diagnostic). Verify the codex global-doc cap semantics empirically before vectors
freeze. Can follow.

---

## M4 — MEDIUM — opencode skills stay machine-global: a managed opencode home is split-brain by design

**Quote.** environments §7.1, opencode skills target: "the manager §5 native surface
(`~/.agents/skills`), unchanged in revision 1"; §9.4: resolved skills "materialize into that
profile's managed homes".

**Why it is a weakness.** For opencode the skills surface lives outside the managed home, so a
session launched into profile A's managed home reads profile A's root context but the
*machine-current* profile's skills. That is precisely the split-brain §9.3 promises to make
"always visible, never implicit" — but no rule surfaces this one, because it is not a scoped
switch; it is structural.

**Proposal.** One sentence in §7.1 (and the §9.4 story): a managed opencode home's skills
follow the current profile for the machine scope, not the launched profile, until the skills
surface moves into `<home>/skills/` (open question 3's pinned-release verification); and `env
status` includes this as a standing per-adapter note rather than leaving it implicit. Can
follow.

---

## M5 — MEDIUM — ax fragment-digest is defined over non-canonical bytes

**Quote.** ax PR #1: `works.relux.curator.fragment-digest` is "the `sha256:`-prefixed
lowercase-hex digest of the exact `launch-env-fragment-v1` JSON bytes consumed at launch".

**Why it is a weakness.** environments §10.2 does not canonicalize `--format json` output;
whitespace/key-order are implementation- and version-dependent, so the digest is incomparable
across managers and even across releases of one manager — a resume drift check comparing
digests false-positives after a pretty-printer change while the profile pin is unchanged. (The
pin keys are unaffected and sufficient for the drift check itself.)

**Proposal.** Either environments §10.2 declares the emitted JSON canonical (the CCJ-1 bytes
already defined in registry.md §1 and used for the managed opencode.json), or the ax PR keys
the digest as "sha256 over the CCJ-1 canonicalization of the fragment object". Fix in the PR
before it merges; can follow curator implementation.

---

## L1 — LOW — the environments machine-configuration surface has no model at all

Forms per environment, per-profile overlay lists + precedence, `system_prompt_files`, target
participation, `isolated` pairs, per-scope current profiles — every knob lives in "machine
configuration" (implementation-specific per manager §1), and `manager-config-v1.schema.json`
contains none of it. Legitimate as protocol freedom; noted because the implementation must
design the config schema growth (curator `internal/config` is a strict parser — mechanical but
non-trivial surface area), and config-driven conformance behaviors are testable only through
direct-input vectors (which is what the shipped vectors correctly do). No spec change needed.

## L2 — LOW — `local` profile pin churn

`default`'s store key and effective pin recompute on every state change (§9.4): each global
skill mutation re-keys the store entry and rewrites every marker/home referencing the old pin,
with the old entry left to GC. The transaction engine handles it, but this is the
highest-churn path in the design; implementation should make one operation = one transaction
covering the re-key. No spec change needed.

## L3 — LOW — claude referenced-form approval claim: supportive but bundle-level evidence only

§5.3: "References stay inside the home, so `claude_code` referenced output never requires the
tool's external-include approval." Inspection of the claude 2.1.258 bundle shows the
external-includes approval list is built from loaded memory files filtered by
`type !== "User"` — the global (config-dir) CLAUDE.md chain is User-type and thus exempt from
the approval prompt, which supports the claim and even suggests it holds regardless of where
the reference points. This is minified-bundle reading, not a behavioral test; keep the claim
behind the §7.6 pinned-release verification gate before the referenced-form vectors freeze
(the spec already phrases it that way for §7.3; extend explicitly to §5.3). Claude's
documented 5-hop import limit is comfortably above the referenced form's 1 hop.

---

## Feasibility matrix — environments.md + manager §12 onto the curator Go codebase

| Spec area | Existing base (internal/…) | Fit | Rel. size |
|---|---|---|---|
| Profilefile/context.json validation, snapshot shape rules | `manifest`, `skillspec`, `protocoljson` strict-parse patterns | clean new package | M |
| Profile store (pin-keyed immutable entries, §4) | `runtimestore` (skill×commit-keyed, atomic once-per-key install) | pattern reuse, new package | S–M |
| `git`/`path`/`local` acquisition | `gitops` + `snapshot` (git archive path — **C3 fix required**), `path` snapshot copy is new | mostly reuse | S + C3 |
| Always-strict audit + `context-secret-material` detector | `audit` detector pipeline, verdict cache | clean extension; detector quality is the real work | S–M |
| Deterministic materialization (parts, header, forms, §5.6 hashes) | none — new pure emitter, driven by the shipped vectors | new package | M |
| Adapter registry generalization (§7) | `adapters` (skills-only table, `.csk-managed.json` ledger) | generalize; env marker and skill ledger stay separate records as specified | M |
| Environment marker (§8.2) | `marker` (install-marker schema bands v1–v4) | sibling marker type, same discipline | S |
| Modes + per-entry atomic switch (§8.1, §9.2) | `transaction` engine: ordered targets, `KindEntry` symlink targets, journal, backup/rollback, durability tests | **strong fit** — `profile use` across N adapters + targets is one Plan of per-entry targets under `managerlock`; managed-home provisioning is per-surface targets (a home tree with tool-owned state can never be one tree target, and doesn't need to be) | M |
| Onboarding: detect, foreign-manager stop, backup, takeover, import (§9.5–9.6) | nothing comparable | new, stateful, UX-heavy; the largest greenfield chunk | L |
| `env resolve` + fragment (§10) | `config`, `marker` reads; repair via `transaction` | S, once M2 semantics are settled | S–M |
| `env status` matrix (§12) | status/recompute patterns | extension | S–M |
| GC live roots (§12) | `scopes/gc.go` conservative mark-sweep, fail-safe retention already the house pattern | clean extension: add profile-store pins, marker-referenced homes, journal refs to the mark set | S |
| Umbrella subcommand discovery (§11) | `cmd/curator` main dispatch | trivial | S |
| opencode XDG seed links (§7.1) | none | new but small; refresh rules are precise | S |

Overall: the curator-side revision-1 surface fits the existing architecture without any
structural change — the transaction/journal model, marker discipline, lock, and GC were built
for exactly this shape. Aggregate size is comparable to the manager §11 external-repository
capability. Nothing curator-side is blocked except C3 (acquisition byte-exactness). The
blocked planes are external: C1 (ax input surface) and C2 (interactive launch mode), plus the
launcher-spec fixes H1/H2.

## Contract holes that MUST close before implementation starts

1. **C3** — byte-exact snapshot acquisition rule + vector (blocks all curator-side hashing
   identities; silently affects shipped skills pipeline too).
2. **C1** — the ax launch-plan input surface (blocks the launcher's ax route; the PR #1 delta
   is necessary but nowhere near sufficient).
3. **C2** — `LaunchModeInteractive` in agents-management (blocks the launcher's spawn plane;
   without it `curator run` ships headless `--dangerously-skip-permissions` runs).
4. **H1** — fragment-before-plan ordering + `LaunchRequest.Home` (avoids a keyed-state
   migration).
5. **H2** — model/effort default ownership in the launcher spec (the flagship one-liner).
6. **H4** — declare claude_code/macOS `isolated` unsupported (one paragraph now vs credential
   clobbering later).
7. **H3** — pi channel spellings (before fragment/channel vectors freeze and any launcher §5
   implementation).

Everything else (M1–M5, L1–L3) can follow implementation start but should land before the
conformance-vector freeze the spec itself gates on.

## Claims verified clean (no finding)

- claude 2.1.258 recognizes `--system-prompt-file` / `--append-system-prompt-file` (§7.3
  claude row correct).
- codex 0.151.0 carries `model_instructions_file` and the `-c key=value` override mechanism
  (§7.3 codex row plausible; per-invocation injection still needs the §7.3 pinned-release
  test at freeze).
- Schemas and vectors: 57 schemas / 773 vector files validate; `regenerate-check` clean; the
  `monolithic-claude-code` expected bytes match the §5/§5.1 byte rules exactly (header lines,
  LF discipline, one empty line between parts, no compose/precedence lines without
  composition).
- `launch-env-fragment-v1` env values fit ax §5.1 `env_literals` name grammar and byte
  limits.
- ax PR #1 extension keys are well-shaped (reverse-DNS, non-secret, immutable-at-creation)
  and consistent with ax §1.6.
- Transaction engine (`KindEntry`, ordered targets, journal recovery) supports per-entry
  atomic multi-adapter re-materialization under one lock as manager §12.3 requires.
- GC extension is a conservative mark-set addition matching the existing fail-safe pattern.
