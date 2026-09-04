# Review findings 1: Decision 0010 draft (agent environment profiles)

Reviewer run: RUN-260831 (TASK-260831-1rjz6j), 2026-08-31.
Subject: `decisions/0010-agent-environment-profiles.md` at `3fd5617` on
`draft/agent-environment-profiles` (base `c9ea2ff` == `origin/main`, verified).
Commits reviewed: `b6b4ef9`, `3fd5617`.

**Verdict: CHANGES REQUESTED — 3 major findings. Status set back to
`development`.** Minor and nit findings may be batched into the same rework
pass. No blocking findings; the overall design is sound and the document is
close.

## Verification evidence (what was checked, not trusted)

- **Spec cross-references**: every cited section exists and says what the
  decision claims — core §1.1, §2, §5, §6.1–6.3, §8, §9.4, §11, §12.3;
  manager §2.4, §2.5, §4.2, §5, §6, §7, §10. Adapter table checked against
  manager §5, MCP against §6, ledger rule against core §11 (wording matches),
  branch allowance against core §5 ("Branches are permitted only for direct
  project declarations…"), `--strict-tags` exists in `cli/curator.md`,
  `curator global …` command family exists. Exceptions are findings 1, 9, 10.
- **Local binary verification** (machine of record, 2026-08-31): claude
  2.1.251, codex-cli 0.151.0, pi 0.84.2, gemini-cli 0.54.4 — all four exactly
  match the versions the document and research resource claim.
  `CLAUDE_CONFIG_DIR` present in the claude 2.1.251 binary; `CODEX_HOME` in
  codex 0.151.0; `PI_CODING_AGENT_DIR` in pi-coding-agent dist;
  `GEMINI_CLI_HOME` in gemini bundle. `~/.curator`, `~/.claude`, `~/.codex`,
  `~/.pi/agent` all exist. opencode is not installed here — its
  `XDG_CONFIG_HOME/opencode` claim rests on the research resource (docs
  confidence) and is not locally verifiable.
- **Xcode secondary targets**: `~/Library/Developer/Xcode/CodingAssistant/`
  does not exist on this machine (consistent with the document's `auto` probe
  story — the assistant has not been used here). Mechanism verified from
  Xcode 26.6 binaries instead: `IDEIntelligenceAgents` contains
  `ClaudeAgentConfig`, "Set CLAUDE_CONFIG_DIR to Claude Agent config
  directory", and "Override CODEX_HOME: %s" — confirming the embedded homes
  are named as claimed and are driven by exactly the primary variables, set
  by the host. The full parent path and the "Xcode 26.3" attribution are
  docs-confidence only (finding 12).
- **ax composition**: agent-session-manager-spec SPEC.md v0.5.0 (VERSION file
  0.5.0) verified read-only: `SpawnPlan` carries `argv`, `cwd`, `env_names`,
  `env_literals` (map, 0..64, non-secret literals ≤4096 bytes) — fragment
  values are paths, compatible. Session Record `extensions` is "Required; may
  be empty; reverse-DNS keys only", so Decision 10's extension-recording
  suggestion is representable in v0.5.0. Spawn is argv+cwd+allowlist, no
  shell, per the DIR-INV rules. All ax claims in the draft hold.
- **Story AC coverage**: profile model + pins (D1), context IR + materializer
  (D2–D3), three modes (D4), launcher + naming (D6, OQ1), profile-scoped
  skills (D8), inventory CLI (D9), ax composition (D10), phased rev-1 scope
  (D11), open questions (7) — all present. Gaps in "with recommendation" are
  findings 8 and 11. Draft branch + English + house style: satisfied
  (structure matches decision 0009's section set; tone and density fit).

## Major findings

### 1. Security claim not delivered by the cited machinery: the "secret canary" does not exist and credential material is not unconditionally blocked

- **Severity**: major
- **Section**: Decision 2 (last paragraph) and Security impact (bullets 1
  and 5)
- **Quote**: "the existing source-audit pipeline (manager §7) applies to
  profile snapshots as it does to skill snapshots, including the secret
  canary; a profile that carries credential-like material is blocked, not
  installed" and "Credentials: never profile content (audit-blocked)".
- **What is wrong**: manager §7 defines a "static canary" whose failure
  always blocks — but the canary is a detector self-test (the reference
  implementation `internal/audit/audit.go` plants a known-bad fixture and
  checks detectors fire), not a secret detector. Nothing in manager §7 names
  a secret or credential detector; the only always-block events are canary
  failure and hash/source revocation. A verifiable finding blocks only in
  strict mode at/above `fail_on`; advisory mode warns and installs. Worse,
  the current deterministic detector set covers only undeclared network
  hosts and undeclared exec names, scanned only in `scripts/` and the
  manifest — a profile snapshot full of AWS keys in `context/*.md` today
  produces zero findings in any mode. The decision leans its credential
  security story on a gate that does not exist. This is the same class of
  overstated security claim decision 0009 had to amend post-review.
- **Suggested fix**: (a) drop the phrase "secret canary"; (b) replace the
  unconditional claim with what the pipeline actually provides, and state
  explicitly that credential-blocking for profiles requires a new
  secret-detection detector class over context modules (name it as
  authorized normative work in the Status/Compatibility sections), and/or
  make profile installation always-strict as an explicit new rule; (c)
  reword the Security bullet to "audit-surfaced" or scope "audit-blocked" to
  the mode and detector set that delivers it.

### 2. The builtin `default` profile contradicts the profile model it migrates into

- **Severity**: major
- **Section**: Decision 8 (Migration) vs Decisions 1, 4, 9
- **Quote**: "the existing machine-local global scope is renamed into a
  builtin profile `default` with its current Skillfile and no root context"
  vs Decision 1: "An **environment profile** is a named, versioned set of
  global agent context installed from a git source."
- **What is wrong**: the builtin profile has no git source, no §6.1 canonical
  identity, no pinned ref, and no effective commit. That leaves undefined:
  Decision 9's `profile list` columns ("source identity, pinned ref,
  effective commit") for `default`; Decision 4's materialization "from the
  same commit-keyed profile store" (there is no commit to key by);
  `profile sync` re-materializing "every installed profile"; and whether
  `use default` participates in the atomic switch like any other profile.
  Every rev-1 CLI surface touches this hole on day one of migration.
- **Suggested fix**: define the local profile shape explicitly — e.g. a
  profile source is either canonical git or `local` (builtin); the store key
  for a local profile is the §8 content hash of its state instead of a
  commit; `profile list` reports `builtin`/`local` in the source column and
  `-` for ref; everything else (switching, sync, status) treats it
  uniformly. One paragraph in Decision 8 plus a sentence in Decisions 4
  and 9 closes it.

### 3. Context IR determinism: "opaque bytes" and "LF line endings, exactly one trailing LF" cannot both hold

- **Severity**: major
- **Section**: Decision 2
- **Quote**: "A module is UTF-8 markdown and is treated as opaque bytes."
  followed by "the applicable modules in manifest order, joined with exactly
  one blank line, encoded with LF line endings and exactly one trailing LF."
- **What is wrong**: for arbitrary opaque module bytes the stated output
  properties are unsatisfiable without transforming the bytes: a module
  containing CRLF line endings either survives verbatim (output is not
  LF-encoded) or is normalized (bytes were not opaque); a module ending in
  zero or several newlines makes "joined with exactly one blank line" and
  "exactly one trailing LF" underdetermined (trim? pad? reject?). This
  ambiguity sits directly on the surface the decision declares
  byte-exact-conformance-vector territory — two conforming implementations
  can disagree today.
- **Suggested fix**: pick one and say it. Cleanest, preserving byte-opacity:
  validate at install time that each module is UTF-8 with LF-only line
  endings and ends with exactly one LF (reject otherwise, like other strict
  schema surfaces), then define output = modules joined by one empty line
  ("\n") with no other transformation. Alternatively define the exact
  normalization algorithm — but validation-and-reject fits the project's
  fail-closed style better than silent rewriting.

## Minor findings

### 4. The launcher fragment boundary is stated as absolute but the managed-home path embeds repo-chosen bytes

- **Severity**: minor
- **Section**: Decision 6 (boundary paragraph) and Security impact bullet 3
- **Quote**: "Profile bytes MUST NOT select or contribute an
  environment-variable name or value" / "values from Curator's machine
  configuration. Profile data cannot add, rename, or retarget a variable."
- **What is wrong**: the fragment value is the managed-home path, and
  Decision 4's illustrative layout is
  `<curator-home>/environments/<profile>/<env-id>/` — `<profile>` is a name
  declared by `Profilefile.json`, i.e. profile bytes, so profile data does
  contribute to the value. The §2 identifier grammar bounds it (no
  separators, no traversal), so there is no escape, but the absolute
  statement is false as written, and this is a security-boundary sentence.
- **Suggested fix**: either weaken the claim to what is true ("profile data
  cannot select the variable name, nor move the value outside the
  manager-owned environments root; the profile-name path component is
  bounded by the §2 identifier grammar") or key the managed-home directory
  by a manager-derived encoding of the profile name.

### 5. claude_code credential passthrough claim is macOS-only

- **Severity**: minor
- **Section**: Decision 7
- **Quote**: "Claude Code's macOS Keychain entries are ambient and need
  nothing".
- **What is wrong**: true on macOS, but the research resource (§4) records
  that on Linux Claude Code keeps `.credentials.json` inside the home — a
  fresh managed home there means unauthenticated sessions, and the
  claude_code passthrough set is undefined for that platform. The document
  otherwise treats non-macOS in scope (open question 6 covers Windows).
- **Suggested fix**: declare the claude_code passthrough set per platform:
  Keychain-ambient on macOS, `.credentials.json` passthrough on Linux (and
  note Windows as needing the same OQ6 verification).

### 6. pi: an unmanaged `AGENTS.override.md` silently masks the managed root context

- **Severity**: minor
- **Section**: Decision 3 (adapter table, pi row) / Decision 9
- **Quote**: pi row: "`<home>/AGENTS.md` (also honors `APPEND_SYSTEM.md`;
  not managed in revision 1)".
- **What is wrong**: pi's discovery chain (research §2, verified from dist
  source) is `AGENTS.override.md` → `AGENTS.md` → … first match. A
  pre-existing unmanaged `AGENTS.override.md` in the agent dir makes the
  materialized `AGENTS.md` inert, the §11 ledger guard does not fire (it
  protects only the managed path itself), and `env status` as specified
  reports the surface current. Same class of hazard for other adapters'
  higher-precedence files, but pi's is the one the research already proved.
- **Suggested fix**: have the pi adapter declare known shadowing paths and
  have materialization/`env status` warn when one exists; or at minimum
  record this as an open question alongside OQ7.

### 7. opencode's `XDG_CONFIG_HOME` fragment leaks into every XDG-conforming child process

- **Severity**: minor
- **Section**: Decision 3 (opencode row), Decision 6, Rejected alternatives
  (`HOME` substitution)
- **What is wrong**: the launcher merges the fragment into the inherited
  environment and execs the tool; children of opencode (git, editors,
  anything XDG-conforming) inherit the redirected `XDG_CONFIG_HOME` and
  lose the operator's real `~/.config` state for the whole process tree.
  This is a milder version of exactly the side-effect class for which the
  `HOME` substitution alternative is rejected ("drags `~/.gitconfig`,
  `~/.ssh` … with it"), but the accepted opencode mechanism's variant is
  nowhere disclosed.
- **Suggested fix**: acknowledge the scope of `XDG_CONFIG_HOME` in Decision
  3 or Security impact (e.g. recommend seeding the managed parent with
  symlinks to the operator's other `~/.config` entries, or note the blast
  radius as an accepted tradeoff pending a dedicated opencode variable).

### 8. Decision 3 promises a per-adapter "materialization-mode default of Decision 4" that Decision 4 never defines for the primary adapters

- **Severity**: minor
- **Section**: Decision 3 ("Each adapter normatively declares: … the
  materialization-mode default of Decision 4") vs Decision 4
- **What is wrong**: Decision 4 names a default only for secondary
  fixed-home targets (`copied`). For the four native-home adapters no
  default is stated anywhere — the AC's "symlink vs copy vs managed-home
  modes with recommendation" is only implicitly answered (`linked` is
  described as "today's global-scope skill mechanism").
- **Suggested fix**: one sentence in Decision 4: default `linked` for
  native-home in-place surfaces (matching manager §5's
  symlink-with-copy-fallback), `copied` for secondary targets, managed
  homes always symlink-from-store; per-adapter overrides where declared.

### 9. Bare "§5" violates the document's own citation convention

- **Severity**: minor
- **Section**: Decision 3
- **Quote**: "Unknown environment identifiers keep their §5 behavior: a
  warning and no output."
- **What is wrong**: the Status section fixes the convention "Section
  numbers cited without a document name refer to `protocol/core.md`" —
  under it this resolves to core §5 (Project manifests), which has no such
  behavior. The intended target is manager §5, which the same paragraph and
  three other places cite correctly with the prefix.
- **Suggested fix**: "manager §5 behavior".

### 10. Read-only status anchored to the wrong manager section

- **Severity**: minor
- **Section**: Decision 9
- **Quote**: "Both are read-only and honor the manager §2.4 dry-run
  constraints."
- **What is wrong**: manager §2.4 is build-cache and dry-run *planning*;
  the read-only status discipline (recompute hashes, report drift,
  non-zero `--check`, "Status MUST NOT … mutate") lives in manager §10.
  Status is not a dry-run.
- **Suggested fix**: cite manager §10 (optionally alongside §2.4 if the
  dry-run no-mutation list is intended to bind `env status` too — then say
  so explicitly).

## Nit findings

### 11. Open questions 4 and 6 carry no recommendation

- **Severity**: nit
- **Section**: Open questions 4, 6
- **What is wrong**: the story AC asks for "open questions with
  recommendations"; OQ1–3, 5, 7 each state one, OQ4 and OQ6 are phrased
  purely as verification items.
- **Suggested fix**: add the working assumption as the recommendation (OQ4:
  keep the manager §5 native surface until a pinned opencode release proves
  `<home>/skills/`; OQ6: hold conformance-vector freeze on the platform
  evidence).

### 12. "Xcode 26.3" attribution and the full CodingAssistant parent path are docs-confidence presented as fact

- **Severity**: nit
- **Section**: Context; Decision 3 (secondary targets table); Open
  question 6
- **What is wrong**: local evidence on the review machine (Xcode 26.5/26.6
  installed; no `~/Library/Developer/Xcode/CodingAssistant/` present)
  confirms the mechanism from binaries (`ClaudeAgentConfig`,
  `CLAUDE_CONFIG_DIR`, `CODEX_HOME` strings in `IDEIntelligenceAgents`) but
  neither the 26.3 version attribution nor the exact parent path. The
  research resource marks the row docs-confidence; the decision states it
  flatly. OQ6 already hedges the semantics, not the path/version.
- **Suggested fix**: fold the path/version attribution into OQ6's
  implementation-time verification, or cite the research resource's
  confidence level at the claim site.

### 13. The environment marker filename is unnamed while every sibling surface is named

- **Severity**: nit
- **Section**: Decision 4 / Compatibility impact
- **What is wrong**: Compatibility lists `Profilefile.json`, `context.json`,
  and `launch-env-fragment-v1` by name and says "The new filenames join the
  §1.1 compatibility identifier list", but the per-home environment marker
  never gets a filename. If it is a §1.1 identifier it needs one; if the
  name is deliberately deferred, say so.
- **Suggested fix**: name it (e.g. `.csk-environment.json`) or mark the
  name as deferred to the normative change.

## Disposition

Majors 1–3 must be fixed before acceptance; the remaining items can ride the
same pass. None of the findings undermine the design direction: the adapter
registry, mode split, fragment boundary concept, and ax composition all
verified clean against the specs and local binaries.
