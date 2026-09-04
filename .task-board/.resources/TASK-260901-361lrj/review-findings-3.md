# Review findings 3: Decision 0010 draft (operator-direction delta)

Reviewer run RUN-260901-675cad for TASK-260901-361lrj, review cycle 3,
2026-09-01. Subject: `decisions/0010-agent-environment-profiles.md` at
`0289328` (`0289328621e86b8bd27d4030230019feff1725e7`) on
`draft/agent-environment-profiles`; delta under review `fe21fb0..0289328`
(`f188173`, `5a782af`, `4365a7d`, `0289328`), 382 insertions / 106 deletions,
markdown-only under `decisions/`. Inputs: `review-brief-cycle-3.md`,
`review-brief-0010.md`, `review-findings-2.md`.

**Verdict: ACCEPT.** No blocking or major findings. Four minors and three
nits are recorded below as rework guidance for the next editing pass or the
normative phase; none invalidates the direction, the extraction, or the
phasing, and the cycle-3 checks otherwise pass on verified evidence.

## Cycle-3 check results

### 1. Numbering integrity — PASS with one minor (finding M1)

- Open questions are exactly 1..7, no gaps, no duplicates (lines 796–841).
- Every "open question N" cross-reference points at matching content,
  verified quote-by-quote: line 43 (Xcode parent path/version → OQ5), line
  152 (import-semantics survey → OQ6), line 193 (embedded-host override
  channel → OQ7), line 253 (Xcode path/version → OQ5), line 499
  (cache/billing research → OQ7), line 534 (Windows claude_code credential
  shape → OQ5), line 633 (phasing rev-3 embedded hosts → OQ7). No stale
  reference to the pre-renumbering targets survives (cycle 2's "Windows →
  OQ6" is now correctly OQ5 in both directions).
- Onboarding list is exactly 1..6 (Inventory, Classification, Consent gate,
  Backup, Import, Authentication); the step-5 cross-reference at line 429
  is correct. The "steps 1–4" scope claim is finding M1.
- All "Decision N" cross-references checked against the section they name
  (1↔8 local kind, 2↔5 header/chapters, 2↔6 system-prompt split, 3↔4 mode
  defaults, 5↔1 path kind, 7↔5 credentials, 9↔3/5/8) — all correct.

### 2. New factual claims — PASS (verified on this machine)

- `claude_code` `--system-prompt` / `--append-system-prompt` and their
  `-file` variants: present in local Claude Code help (local binary is
  2.1.252; the document's "verified in 2.1.251" is a version-stamped record
  and the flags persist).
- `hasClaudeMdExternalIncludesApproved`: real per-project key in the live
  `~/.claude.json` (dozens of project entries, paired with
  `hasClaudeMdExternalIncludesWarningShown`) and present in the claude
  2.1.252 binary strings (8 hits). The Decision 2 gating description is
  accurate.
- `codex_cli` 0.151.0: `codex --version` confirms 0.151.0;
  `model_instructions_file` present in the binary strings (15 hits).
- `pi` 0.84.2: `pi --version` confirms; `--system-prompt` and
  `--append-system-prompt` in help; `APPEND_SYSTEM.md` verified in the
  installed dist source (`dist/core/resource-loader.js:819–828`,
  `discoverAppendSystemPromptFile`): the global agent-dir file is discovered
  and applied unconditionally when present and no CLI flag overrides it; the
  project-level variant is gated on project trust. The document's
  "applies unconditionally when present" claim is exactly right for the
  agent-dir path. See finding M3 for what the same source file shows beyond
  the recorded channel.
- `GEMINI_SYSTEM_MD`: presented as docs-recorded in the document, which is
  the correct confidence label; incidentally the string is present in the
  local gemini-cli 0.54.4 bundle chunks, so the docs claim is corroborated,
  not contradicted.
- `opencode` is not installed on this machine; the `instructions`-list and
  XDG claims remain docs-confidence. The document mostly treats them as
  such — except OQ6's "verified" wording, finding M2.

### 3. Four-plane map vs skill-agents-management — PASS (one nit, N1)

Read SKILL.md from `relux-works/skill-agents-management` (via gh api).
The decision's spawn-plane characterization matches the skill's own
contract: "which agent runs, on which model, and whether it may launch
right now"; "It builds plans; it does not execute — every LAUNCH-plane
entry point ends at a value"; effort is per-model and required; provider
limits produce structured availability verdicts. The declared only-import
edge (launcher → `agents-management` Go module) is not merely consistent
with the skill — it is forced by it: the shipped `agents-management` binary
has no `spawn`/`availability`/`models` command ("the consumer links the
packages; the binary is a listing surface"), so consuming plans requires
the module import, and Curator/`ax` as CLI contracts stay import-free
exactly as Decision 10 draws it. "Consumed as a built launch plan, never
rebuilt" matches `BuildPlan`/`BuildLaunch` returning values. One shape
mismatch in the parenthetical is nit N1 (stdin).

### 4. Extraction coherence — PASS

- `curator launch` appears exactly once, inside Rejected alternatives
  (line 698), as the rejected earlier draft. Every operative reference is
  `curator run`, `curator-run`, `curator-agent-launcher`, or `curator
  session` and they agree with each other (Decision 6, Decision 10 table,
  phasing rows 637–638, Consequences 786–789).
- One story across Decision 2 / Decision 6 / Security / phasing:
  materialization = Curator (system modules into managed homes only, rev 1
  claude_code + codex_cli + pi), application = launcher opt-in with warnings
  (deferred to the launcher's own specification), native in-place homes
  never receive system-prompt files, `ax` engagement is
  always-when-configured (stated identically at Decision 6 item 3,
  Decision 10 degradation paragraph, and Consequences). No contradiction
  found.

### 5. Determinism story — PASS with one minor (M4)

- Generation header: content fixed to project URL, profile identity,
  effective commit/state hash, composition chain, precedence, drift notice;
  explicitly no timestamp, machine path, or operator identity. Deterministic.
- Monolithic output: header + applicable modules in manifest order + single
  empty-line joins, byte-identical for identical (commit, adapter set, form,
  composition) — well-defined and vectorable. Chapter separators under
  composition are deterministic in order and attribution; their exact bytes
  are normative-phase work and Compatibility correctly claims byte-exact
  vectors for header, chapters, and both forms.
- Referenced form: the output *set* is not pinned — finding M4.
- Cycle-2 carry-forward nit (zero-applicable-modules output shape) remains
  open for the normative phase, unchanged, as agreed in cycle 2.

## Findings

### M1 — minor — Decision 5 (First-run onboarding), line 427

Quote: "Revision 1 ships steps 1–4 — detection, foreign-manager stop,
backup, and takeover…"

The ordinal range contradicts the list it cites and the rest of the
document. Step 2 is Classification (lossless/lossy) and step 3's operative
branch is the lossy-import consent stop — but the very next sentence
assigns "lossless/lossy classification" to the deferred step-5 import
machinery, and the phasing table (line 640) likewise defers
"lossless/lossy" to the revision-2 import story. The gloss itself names
only step-1 content (detection, foreign-manager stop, takeover choice) and
step 4 (backup). Rev 1 cannot ship a lossless/lossy classifier whose
machinery is deferred.

Fix: restate the shipped subset accurately — e.g. "Revision 1 ships steps
1 and 4, plus the consent-gate notice that native contexts are being
replaced (step 3's replace-notice half); the classification of step 2 and
the lossy-consent branch of step 3 travel with the step-5 import story" —
or renumber the list so classification and lossy consent sit inside the
deferred import step.

### M2 — minor — Open question 6, line 829

Quote: "ship revision 1 with the two verified referenced targets".

The document's own confidence vocabulary distinguishes "verified" (binary
or source evidence: claude 2.1.251 flags, pi 0.84.2 loader source, codex
0.151.0 key) from "documented"/"recorded"/"docs-confidence". Both
referenced-form targets are the latter: claude_code `@path` is "documented
up to five hops" (line 137, line 824), opencode `instructions` "lists are
recorded" (line 824) and opencode is not even installed locally to check.
Calling them "verified" three lines later overclaims by the document's own
scale. (The `hasClaudeMdExternalIncludesApproved` gate is genuinely
verified; the import traversal itself is not.)

Fix: "the two documented referenced targets" (or "recorded"), keeping the
re-survey recommendation unchanged.

### M3 — minor — Decision 3 table / Decision 6 — pi's SYSTEM.md channel unrecorded

Verified in the same pi 0.84.2 source the document cites for
`APPEND_SYSTEM.md`: `discoverSystemPromptFile()`
(`dist/core/resource-loader.js:808–817`) discovers `SYSTEM.md` from the
agent dir under identical semantics — applied unconditionally when present,
no flag needed — and it is a **full replacement** of the system prompt,
i.e. strictly sharper than the recorded append channel. The document
records only `AGENTS.md` + `APPEND_SYSTEM.md` for pi (Decision 3 table,
Decision 6). The generic guard "managed homes carry no active
system-prompt file" covers SYSTEM.md in spirit, but the adapter's declared
channel record, the shadowing/hazard awareness, and OQ7's research list do
not name it — and a materialization or takeover path that lets a
`SYSTEM.md` land in a pi managed home silently replaces the entire system
prompt with no warning surface defined for it.

Fix: record `SYSTEM.md` (replacement) beside `APPEND_SYSTEM.md` (append)
wherever pi's channel is described — Decision 3 table parenthetical,
Decision 6's pi sentence, OQ7 — and make the "no active system-prompt
file" guard enumerate both for pi.

### M4 — minor — Decision 2 (Materialization forms) — referenced-form output set undefined

The determinism paragraph defines the monolithic byte stream precisely;
for `referenced` the document says only that modules "materialize as
individual files beside the root file". The file layout is unstated: the
materialized filenames, whether they collide with unmanaged files beside
the root, and — the sharp case — the collision rule under Decision 5
composition, where two composed profiles can both carry `00-base.md`.
Composition and the referenced form are both revision 1
(phasing rows 630–632), so the intersection is in scope, and Compatibility
promises byte-exact determinism vectors "for both forms" over an output
set the draft never defines.

Fix: either pin the layout rule in Decision 2 (e.g. per-source-profile
subdirectory or profile-prefixed filenames under a managed subdirectory,
all entries ledgered) or explicitly flag referenced-form layout and its
composition collision rule as normative-phase work in Decision 2 / OQ6,
the way the zero-modules case was flagged in cycle 2.

### N1 — nit — Decision 6 item 1, line 477

"consumed as a built launch plan (binary/argv/environment), never rebuilt"
— the agents-management plan value is "a binary, an argv, an environment,
stdin bytes" (SKILL.md), and effort transport is per-system argv/stdin/none,
so a stdin-transport system's plan is under-consumed by the parenthetical
as written. The launcher spec inherits this sentence; add "/stdin".

### N2 — nit — Decision 2 vs Decision 6 — pi activation channel taxonomy

Decision 2: system modules "activate only through the launcher's explicit
opt-in (Decision 6) or through natively typed commands". Decision 6 adds a
third shape for pi: the per-profile×environment machine setting that
materializes the auto-applied `APPEND_SYSTEM.md`. It is best read as the
persistent form of the opt-in, but the document never says so; one clause
in Decision 2 ("…opt-in, including its persistent per-profile×environment
form for pi's file channel…") closes the gap.

### N3 — nit — Security impact — umbrella dispatch surface unacknowledged

The delta introduces PATH-based external-subcommand execution
(`curator-<name>`), the one place Curator now executes a binary it does
not ship. The trust model is exactly git/kubectl plugin discovery and is
fine, but the Security impact section — which carefully bounds every other
new surface of this delta — is silent on it. One bullet acknowledging that
umbrella dispatch trusts PATH precisely as the git/kubectl convention does
(and that profile data cannot influence the dispatched name) would keep
the section complete.

## Validation evidence

- Worktree verified: branch `draft/agent-environment-profiles`, head
  `0289328`, four commits since the cycle-2 ACCEPT point `fe21fb0`, good
  ECDSA signature on head (the principal-mapping warning is the same stale
  verifier config recorded in cycle 2, not a signature failure). Delta
  touches only `decisions/0010-agent-environment-profiles.md`; tree clean
  except the excluded `tools/__pycache__/` residue.
- Re-ran myself: `go test ./tools/...` in the draft worktree — ok.
- Accepted from cycle-2 evidence (same basis, unchanged since): the
  `validate.py`/unittest gates — `tools/validate.py` has zero references to
  `decisions/` and the delta is markdown-only under `decisions/`, so it
  cannot affect schema/vector validation.
- Binary/source evidence gathered this cycle: claude 2.1.252 help + binary
  strings + live `~/.claude.json`; codex 0.151.0 `--version` + binary
  strings; pi 0.84.2 `--version` + help + installed dist source
  (`resource-loader.js` discovery and application sites read, not just
  grepped); gemini 0.54.4 bundle grep; opencode absent (docs-confidence
  claims left as such). skill-agents-management SKILL.md fetched from the
  repository head via gh api.

## Verdict

ACCEPT. Minor findings M1–M4 and nits N1–N3 are rework guidance for the
next operator/editing pass (M1, M2, N1, N2, N3 are one-sentence fixes;
M3 and M4 can also be discharged into OQ7/OQ6 as flagged normative-phase
work). None of them is a condition on accepting the operator-direction
delta.
