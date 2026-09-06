# Review findings — TASK-260906-1hn93j, cycle 1 (cli/curator.md takeover and import rows)

Verdict: **CHANGES REQUESTED**. One blocking finding, two major, one minor.

Subject reviewed: `f39f4a9..f013e0c` on `task-board/story/STORY-260905-2z9pw4`.
Authority: `protocol/environments.md` at `f39f4a9` (§7.6, §8.3, §8.5, §9.1, §9.2,
§9.3, §9.5, §9.6, §9.7), `decisions/0010-agent-environment-profiles.md`,
`profiles/manager.md`, and the `cli/curator.md` preamble and flag family.

repeat-of: none (first review cycle on this element).

## 0. On the empty Change Request delta

`CR-TASK-260906-1hn93j-1` reports `repository_delta: empty`. That is a snapshot
artifact, not an absence of work: the CR base OID is `f013e0c`, and
`git rev-parse f013e0c^{tree}` = `6c9e7ca9156fdd168b68ebf7281b392fa4960f03` =
the candidate tree. The producer committed before the CR was cut, so the CR
records "nothing beyond the already-committed head". The reviewable content is
`git diff f39f4a9..f013e0c` — 2 files, +17 lines — and that is what this review
attacks. The emptiness is therefore neither the deliverable nor the defect; the
defects below are in the committed content.

## 1. BLOCKING — `curator env takeover --takeover` publishes an operation shape no sentence in `environments.md` supports, and the shape defeats its own gate

**File/section:** `cli/curator.md:43` (Commands table), `cli/curator.md:175-177`
(examples), `CHANGELOG.md:89-90`.

**Quoted row:**

> `curator env takeover --takeover [--env <env-id>] [--target <target-id>]` |
> Take over a specific unmanaged file outside onboarding under the explicit
> takeover flag, with the same notice and section 8.3 backup generation as
> onboarding; **without the flag the operation fails with
> `environment_surface_unmanaged_conflict` rather than overwrite**

The producer's Q5 asks whether §9.5's takeover is a standalone command or a
modifier on the mutating operations, then publishes the standalone reading
anyway. I attacked both readings against the text. The standalone reading is not
merely unsourced — it is excluded, and the row it produces is self-defeating.

**(i) The refusal clause is unreachable under the published shape.** The row's
own gate clause describes what happens "without the flag". For a subcommand
named `takeover` that requires `--takeover`, the flagless invocation is
`curator env takeover` — a command with no behaviour at all. There is no
operation left to fail with `environment_surface_unmanaged_conflict`, so the
only normatively interesting half of the row documents a branch the published
command cannot reach. This is the index-level form of a gate no production path
invokes.

**(ii) The published grammar cannot express what the row says the command does.**
The description says "a specific unmanaged file". The signature offers
`--env <env-id>` (an adapter scope, §9.3) and `--target <target-id>` (a
secondary fixed home, §7.6) and no path operand. Either the row means
scope-wide takeover — which no sentence states, and which is a new normative
claim in a surface index (see finding 2) — or the command as published cannot
name its own operand. The example inherits it verbatim: the comment says "Take
over an unmanaged file" above `curator env takeover --takeover --env claude_code`.

**(iii) Every sentence that touches the takeover presupposes an operation that
exists without the flag.** The evidence, in the spec's own words:

| Sentence | What it entails |
|---|---|
| §9.5: "Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag ...; without the flag, section 8.3 applies and **the operation** fails rather than overwrite." | "The operation" needs an antecedent that exists flagless. Under a standalone `env takeover` there is none; under a modifier on `profile install` / `use` / `sync` / `update` / `env resolve --repair` there is exactly one. |
| §9.5 trigger list: "`profile install`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`, and **an explicit takeover**" | Five triggers are code-spanned command names; the sixth is plain prose. The document spells a command as a command everywhere else it names one. |
| §9.5 step 4: "Onboarding without an import ends after step 3 and **the takeover writes** the operator chose." | Takeover is a class of writes an operation performs, not a command. |
| §9.6 closing: "replacing native files remains the section 9.5 **takeover path** with its notice and backup." | "Path", not "command". |
| §8.3: "never wedges a second **takeover of the same path**." | Takeover is per-path — which is the write set of an ordinary mutating operation, not a scope selection. |
| `decisions/0010`:405: "the **takeover flag** remains the manual equivalent." | The ADR calls it a flag, and calls it the manual equivalent of the onboarding an ordinary operation triggers. |
| `profiles/manager.md`:2295-2297: "without the flag the section 12.2 ledger discipline fails **the operation** rather than overwrite." | The normative manager profile repeats the same antecedent problem. |

Against that, no sentence anywhere in the repository names a takeover
subcommand. `grep -rn takeover` over `protocol/`, `profiles/`, `decisions/`
returns nine hits and not one of them is a command name. The spec's own
convention is to publish an operand grammar when it defines an operation —
§9.1 gives `profile install` a full signature in a fenced block, §9.2 gives
`env unmanage [--restore-backups] [--env <env-id>] [--target <target-id>]` with
its default ("default: every scope") inline. The takeover has neither, because
it is not an operation of its own.

**Conclusion.** §9.5 does not fix the standalone reading; it forecloses it. It
indicates the modifier reading — an operation refused by §8.3 with
`environment_surface_unmanaged_conflict` proceeds on a re-run under the takeover
flag, with §9.5's notice and §8.3 backup — under which every sentence has a
referent and no new operand grammar is needed. What §9.5 still does **not** fix
is the enumeration: no sentence states which operations accept the flag (the
trigger list is the only candidate) or whether the flag is accepted where the
operation is not itself an onboarding trigger.

**Fix — either:**

- (a) Publish the takeover as `[--takeover]` on the rows that already exist —
  `curator profile install|use|sync|update` and `curator env resolve` — with one
  clause sourced to §9.5 ("takes over the unmanaged files the operation would
  write, under the §9.5 notice and the §8.3 backup generation; without it the
  write fails with `environment_surface_unmanaged_conflict`"), and delete the
  standalone row and its example. Note that `cli/curator.md` currently mentions
  onboarding nowhere, so this is also the first place those five rows state why
  they can meet unmanaged state at all. The producer's "deliberately not
  published" list dismissed those five rows as "no edit needed"; under this
  reading they are precisely where the flag belongs, and that dismissal is what
  has to be revisited.
- (b) Or hold the row entirely and raise the missing sentence as the spec
  follow-up, per the producer brief's own rule that a row you cannot source is a
  defect and not a deliverable. Q5 is a real open question and was correctly
  identified; publishing a row on it anyway is the defect.

Do **not** re-publish the standalone row with a path operand invented to fix
(ii). That trades an unsourced flag shape for an unsourced operand grammar.

## 2. MAJOR — the takeover row's scope flags assert scope-wide takeover, and state no default

**File/section:** `cli/curator.md:43`.

**Quoted clause:** `[--env <env-id>] [--target <target-id>]`

`environments.md` scopes the takeover per file — "Takeover of **a specific
unmanaged file** outside onboarding" (§9.5), "a second takeover of **the same
path**" (§8.3). The published grammar scopes it per adapter/target. Substituting
a scope for a file is not a spelling choice; it changes what the operation does,
and it makes `cli/curator.md` state a rule `environments.md` does not — a
surface-index discipline breach, and the exact thing the AC forbids ("No
normative rule is added to cli/curator.md that environments.md does not already
state").

The drafting report marks this CHOICE and raises it as Q2, which is honest as
far as it goes. What it does not notice is that the chosen grammar **contradicts
the row's own description in the same cell**: "a specific unmanaged file" and
`--env <env-id>` cannot both be true. A clause marked as a free choice that in
fact conflicts with the sourced clause beside it is a defect, not a choice.

Second, unlike the §9.2 signature it mirrors, the row states no default scope.
`env unmanage` publishes "(default: every scope)" because §9.2 states it. The
takeover row publishes optional scope flags with no stated behaviour when both
are omitted — undefined behaviour published as operator surface.

**Fix:** falls out of finding 1. Under fix (a) the scope of the takeover is the
scope of the operation performing it and no scope flags are needed. Under fix
(b) the row does not ship. In neither case is `--env`/`--target` published for a
takeover.

## 3. MAJOR — the import row omits `[--use]`, the one activation control §9.6 explicitly imports from §9.1

**File/section:** `cli/curator.md:31`.

**Quoted row signature:** `curator profile import [--as <name>] [--allow-lossy]`

§9.6: "The assembled directory then installs through section 9.1 exactly as an
operator-supplied `path` source ... **Activation follows the section 9.1 rules
without magic.**"

§9.1: "Activation on install follows operator intent without magic: `install`
sets the machine current profile only when the machine has none — first install,
and the activation is reported, never silent — **or when the operator passes
`--use`**. `--use` takes no name."

So §9.6 pulls in, by explicit cross-reference, an operator control whose
spelling is already published one row above on `profile install`. This is not a
Q1/Q3/Q4-class spelling the informative document is free to invent — it is a
clause the spec fixes and the row drops. As published, the surface index says
the import cannot be activated in the same operation, and §9.6's activation
sentence has no operator surface at all.

The omission is also unjudged: `[--use]` appears nowhere in the drafting
report — not in the sourcing table, not in the "deliberately not published"
list, not in Q1–Q5. An unnoticed omission, not a decided one.

**Fix:** publish `curator profile import [--as <name>] [--allow-lossy] [--use]`
and extend the description with the §9.1 activation rule the install row already
carries (`--use` takes no name; first install activates and says so).

## 4. MINOR — the report's "no coverage" bound is a proxy-derived absence, and one sub-claim is false

**File/section:** `TASK-260906-1hn93j_drafting-report.md`, "Verification bounds".

**Quoted claim:**

> `tools/validate.py` has no coverage of `cli/curator.md` (verified: no
> reference to `curator.md` or `CHANGELOG` anywhere under `tools/`)

Two problems.

- The `CHANGELOG` half is false. `tools/release_gate.py:650` reads
  `CHANGELOG.md` and fails with "CHANGELOG has no {version} release heading";
  `tools/release_gate.py:722` lists it among the release surfaces scanned for a
  retired descriptor name; `tools/test_release_gate.py:29,133,215` exercises it.
  Consequence for this batch is nil — `release_gate.py` runs only under
  `make release-check VERSION=...`, and an entry under `## Unreleased` moves no
  version heading — but the claim as written is falsifiable by one grep and is
  false.
- The method cannot establish the `curator.md` half. A grep for a filename is
  blind to a glob walk, and `tools/validate.py:3311` has one:
  `validate_local_links()` does `for path in sorted(ROOT.rglob("*.md"))` and
  therefore *does* cover `cli/curator.md`, for link integrity. The new rows add
  no markdown links, so nothing was going to fire — but "the grep found nothing"
  and "nothing is there" are different facts, and the report presents the first
  as the second.

The conclusion the bound supports — that no committed check asserts
`cli/curator.md` row text — is correct; I verified it independently (no `cli/`
path is read anywhere under `tools/`, and the only `*.md` walk checks links).
Restate the bound as measured: `cli/curator.md` is covered by
`validate_local_links()` for link integrity only; no check asserts row text.

## Clause → source-sentence table (both rows, in full)

### Takeover row — `cli/curator.md:43`

| Clause | Sourced? | Sentence |
|---|---|---|
| `curator env takeover` (standalone subcommand) | **NO** | No sentence in `protocol/`, `profiles/` or `decisions/` names a takeover subcommand. §9.5's trigger list carries "an explicit takeover" in prose beside five code-spanned command names. Finding 1. |
| `--takeover` (flag spelling) | Choice, legitimate | §9.5 "requires the explicit takeover flag" names a flag but no spelling. `cli/curator.md:3` — "This document is informative ... Other managers may use different command names, flags" — licenses the spelling. Bare boolean matches `--purge`, `--repair`, `--use`. Q1 is answered by the preamble, not open. |
| `--takeover` **required** by a subcommand named `takeover` | **NO** | Tautological; makes the row's refusal clause unreachable. Finding 1(i). |
| `[--env <env-id>] [--target <target-id>]` | **NO** | §9.5 scopes takeover per file ("a specific unmanaged file"); §8.3 per path ("the same path"). §9.3/§7.6 define these flags for `profile use` and `env unmanage`, not for a takeover. Finding 2. |
| no stated default scope | **NO** | §9.2 states `env unmanage`'s default explicitly ("default: every scope"); no counterpart exists for takeover, so the omission publishes undefined behaviour. Finding 2. |
| "Take over a specific unmanaged file outside onboarding" | YES | §9.5: "Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag..." — but contradicted by the row's own grammar. Finding 2. |
| "under the explicit takeover flag" | YES | §9.5, same sentence. |
| "with the same notice ... as onboarding" | YES | §9.5: "performs the same notice and backup"; notice content at §9.5 step 2: "before any write, the operator is told that native global context files are being replaced by managed ones and where the backup lands." |
| "and section 8.3 backup generation" | YES | §8.3: "Takeover and onboarding backups (section 9.5) land in **versioned** backup sets `.agent-environment-backup/<n>/` ... incremented per operation that backs anything up." |
| "without the flag the operation fails with `environment_surface_unmanaged_conflict` rather than overwrite" | YES as text, **unreachable as published** | §9.5: "without the flag, section 8.3 applies and the operation fails rather than overwrite" composed with §8.3: "MUST fail with `environment_surface_unmanaged_conflict` rather than overwrite an unmanaged file." Diagnostic spelled exactly as the §8.5 table spells it. Finding 1(i). |
| example `curator env takeover --takeover --env claude_code` | Inherits findings 1 and 2 | Comment says "an unmanaged file", command names a scope. `claude_code` is a §7.1 adapter id — correct as an id. |

### Import row — `cli/curator.md:31`

| Clause | Sourced? | Sentence |
|---|---|---|
| `curator profile` placement | YES | §9.6: "its output is one installed, audited, locked profile whose environment markers record `imported_from_native`." |
| `import` (subcommand spelling) | Choice, legitimate | §9.6 names no command; `cli/curator.md:3` licenses the spelling. Q4 answered by the preamble. |
| "Reassemble the section 9.5 inventory into a context-package-shaped directory" | YES | §9.6: "Its input is the section 9.5 inventory"; "The manager assembles a context-package-shaped directory inside the machine home (physical location implementation-specific, manager §1)". |
| "and install it through the ordinary `path` pipeline" | YES | §9.6: "The assembled directory then installs through section 9.1 exactly as an operator-supplied `path` source — snapshot copy, state-hash pin, resolution of the pinned skills, always-strict audit". |
| "the profile is named `imported` unless `--as` supplies a name" | YES | §9.6: "`agent-context.json`, `schema_version` 1, `name` `imported` unless the operator supplies a name under the core §2 grammar". `--as <name>` carrier reuses §9.1's published spelling ("the profile name is the root package's `name` unless `--as <name>` is given"), already on the install row. |
| `[--allow-lossy]` (flag spelling) | Choice, legitimate | §9.6 names "an explicit per-operation consent flag" with no spelling; the preamble licenses it, and the verb-led shape matches `--restore-backups` / `--allow <hash>`. Q3 answered by the preamble. Best available spelling: it names the class it admits, which `--force`/`--yes` would not. |
| "a lossy import stops with `environment_import_lossy` and the loss list" | YES | §9.6: "A lossy import stops with `environment_import_lossy` and the loss list". Diagnostic spelled exactly as the §9.7 table spells it (both rows of it). |
| "unless the per-operation consent flag re-reports the list as warnings" | YES | §9.6: "it proceeds only under an explicit per-operation consent flag, which re-reports the loss list as warnings under the same diagnostic." |
| "and machine configuration never pre-records consent" | YES | §9.6: "Machine configuration MUST NOT pre-record consent." (MUST NOT → "never" is a fair informative compression; it does not weaken the rule.) |
| `[--use]` | **MISSING** | §9.6: "Activation follows the section 9.1 rules without magic", and §9.1: "or when the operator passes `--use`". Finding 3. |
| example `curator profile import --as legacy --allow-lossy` | Matches the row | Carries both published flags; comment restates the `imported` default and the consent gate. |

### Producer's "deliberately not published" list, judged

| Item | Verdict |
|---|---|
| lossless/lossy classification rules, closed detected-surface list, reassembly normalization | Genuinely out of surface scope. The doc publishes no comparable normative depth on any row. |
| `profile_import_name_taken` | Consistent: the `profile install` row publishes no `profile_name_taken` either. |
| `environment_import_skill_foreign` | Consistent: the install row publishes no `context-system-module-present` warning either. |
| the onboarding trigger rows ("already behave per §9.5; no edit needed") | **Convenient omission.** Under the reading finding 1 establishes, those five rows are exactly where the takeover flag belongs. Revisit. |

## Mechanics — all green

| Check | Result |
|---|---|
| Commits past `f39f4a9` | Exactly one: `f013e0c` |
| Signature | `git log --show-signature -1 f013e0c` → "Good \"git\" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM". `%G?` is `U` only because the repo's `allowed_signers` path (`/private/tmp/curator-spec-rc8-verify.*`) is gone — identical for `f39f4a9` and every prior commit, so not a regression. |
| Identity | `Ivan Oparin <oparin@me.com>`, author and committer, matching `git config` and every prior commit |
| Files touched | `CHANGELOG.md` (+6), `cli/curator.md` (+11). No `LOGBOOK.md`, no stray file. |
| `CHANGELOG.md` convention | Entry under `## Unreleased` → `### Added`. `f61ee9a` filed its CLI *rewrite* under `### Changed`; new rows under `### Added` is the correct counterpart. |
| `make validate` | **exit 0**, re-run by me (venv `.temp/venv`): `validated 60 schemas and 1017 vector files`; `Ran 227 tests ... OK`; `ok github.com/relux-works/curator-spec/tools/generate-vectors 0.511s`. Log: `.temp/review-rev1/validate.log`. |
| `make regenerate-check` | **exit 0**, re-run by me, no diff. Log: `.temp/review-rev1/regenerate-check.log`. |
| Worktree after gates | clean (I removed the `tools/__pycache__/` my own run produced) |
| Examples ↔ rows | Both example lines match their rows verbatim and sit next to the rows' table neighbours. They also inherit findings 1 and 2. |

## Coverage of the AC rows

| AC row | Driving check | Result |
|---|---|---|
| takeover row, every clause traceable to §9.5/§8.3/§7.6 | clause table above, sentence by sentence | **FAIL** — subcommand shape and scope grammar unsourced |
| import row with optional profile name and lossy-consent flag, traceable to §9.6 | clause table above | **PARTIAL** — every published clause sourced; `[--use]` missing |
| flag spellings consistent with published env/profile rows | compared against `--env`, `--target`, `--as`, `--purge`, `--repair`, `--use`, `--restore-backups`, `--allow <hash>` | PASS |
| examples block gains one line per new row | `cli/curator.md:148-151` and `:175-177` | PASS (inherits findings 1–2) |
| `make validate` and `make regenerate-check` green | re-run by me | PASS |
| exactly one signed commit, human identity, no stray files | `git show --stat`, `--show-signature`, `git log --format` | PASS |
| report carries row→sourcing table with quoted sentences plus open questions | read in full | PASS in form; finding 4 on the verification bound |

**4 of 7 AC rows pass; 1 partial; 2 fail.**

Stated bound on the evidence: this batch changes prose only, and there is no
committed check anywhere under `tools/` that asserts `cli/curator.md` row text —
`validate_local_links()` (`tools/validate.py:3311`) walks every `*.md` for link
integrity and nothing else reads `cli/`. There is consequently no gate to
narrow and no mutant to run for these rows; the review's evidence is the
clause-by-clause sourcing attack above, run against `environments.md` at
`f39f4a9`. That is a stated bound, not a claim of coverage.

## What to do next

1. Resolve finding 1 — either publish the takeover as `[--takeover]` on the five
   §9.5 trigger rows, or drop the row and raise the missing `environments.md`
   sentence as the spec follow-up. Finding 2 dissolves with either.
2. Add `[--use]` to the import row with the §9.1 activation clause (finding 3).
3. Correct the report's verification bound (finding 4).
4. Q1, Q3, Q4 can be closed as answered by `cli/curator.md`'s own informative
   preamble; Q2 and Q5 are the real open questions and Q2 is subsumed by Q5.
