# Rework report 1 — TASK-260906-1hn93j (cli takeover and import rows)

Subject: story branch `task-board/story/STORY-260905-2z9pw4` at `e01de3f`,
exactly one signed commit past curator-spec main `f39f4a9`. The cycle-1
commit `f013e0c` was superseded by `git reset --soft f39f4a9` plus a fresh
single commit, so the branch carries one commit, not two. Reviewable
content is `git diff f39f4a9..e01de3f` — 4 files, +49/−13.

Authority: `protocol/environments.md` §§7.6, 8.3, 9.1, 9.2, 9.5, 9.6,
`profiles/manager.md` §12.3, the `cli/curator.md` preamble and flag family,
and `TASK-260906-1hn93j_review-findings-cli-1.md` (cycle 1, CHANGES
REQUESTED: 1 blocking, 2 major, 1 minor). The reviewer's shape argument is
accepted as instructed and is not re-litigated below.

## 1. Finding → resolution

| # | Severity | Finding | Resolution |
|---|---|---|---|
| 1 | BLOCKING | `curator env takeover --takeover` publishes a standalone-subcommand shape no sentence supports; its refusal clause is unreachable, its scope grammar contradicts its own "specific unmanaged file" description, and every takeover sentence presupposes an operation that exists without the flag. | Accepted. The standalone row and its example are deleted. The spec gap is closed in `environments.md` §9.5 (amended sentence below: flag carried by a mutating operation, closed enumeration, per-file scope), and `cli/curator.md` publishes `[--takeover]` on the six table rows for the five enumerated operations instead. |
| 2 | MAJOR | The takeover row's `[--env]/[--target]` grammar asserts scope-wide takeover contradicting the per-file sourcing, and states no default scope. | Accepted. No scope grammar is published for any takeover: under the amended sentence the takeover's scope is the carrying operation's scope. The finding dissolves with fix (a); no `--env`/`--target` and no path operand were invented. |
| 3 | MAJOR | The import row omits `[--use]`: §9.6 imports the §9.1 activation rules by explicit cross-reference, and `--use` is that control's published spelling. | Accepted. The import row is now `curator profile import [--as <name>] [--allow-lossy] [--use]`, with the activation clause the install row already carries. Source chain is §9.6 → §9.1, quoted in the table below. |
| 4 | MINOR | The cycle-1 report's "no coverage" bound was proxy-derived, and its `CHANGELOG` sub-claim was false. | Accepted. Bounds in §7 below are stated only as established by the named method (grep token + gate run), and the false sub-claim is not repeated. |

Q1, Q3, Q4 (`--takeover`, `--allow-lossy`, `profile import`, `--as`
spellings) are kept as published per the rework brief: `cli/curator.md:3`
declares the document informative and licenses divergent command/flag
names, so these remain legitimate preamble-licensed choices, not defects.
Q2 is subsumed by Q5, and Q5 is closed by the §9.5 amendment, not by an
index-side choice.

## 2. Amended sentences, before and after (quoted in full)

### 2a. `protocol/environments.md` §9.5

BEFORE:

> Takeover of a specific unmanaged file outside onboarding requires the
> explicit takeover flag and performs the same notice and backup; without the
> flag, section 8.3 applies and the operation fails rather than overwrite.

AFTER:

> Takeover is not an operation of its own: the explicit takeover flag is
> carried by a mutating operation and covers only the specific unmanaged
> files that carrying operation would write — it never selects a scope of
> its own. The flag is accepted on exactly the mutating operations named
> above as onboarding triggers — `profile install`, `profile use`, `profile
> sync`, `profile update`, and `env resolve --repair` — and on no other
> operation. A carrying operation that meets unmanaged files outside
> onboarding performs the same notice and backup as onboarding when the flag
> is given; without the flag, section 8.3 applies and the operation fails
> with `environment_surface_unmanaged_conflict` rather than overwrite.

What the replacement fixes, clause by clause: (i) shape — "not an
operation of its own ... carried by a mutating operation", giving "the
operation" in the refusal clause its antecedent ("a carrying operation");
(ii) enumeration — the closed set is exactly the five triggers the same
section already names, "and on no other operation"; (iii) per-file
semantics — "covers only the specific unmanaged files that carrying
operation would write" and "never selects a scope of its own", so §8.3's
"a second takeover of the same path" and the old "a specific unmanaged
file" both still read true. Surviving clauses kept intact: the same notice
and backup as onboarding, and the without-flag failure under §8.3. The
diagnostic token `environment_surface_unmanaged_conflict` is the existing
§8.3/§8.5 name written out where the old sentence said "section 8.3
applies" — no name changed, none added. The spelling `--takeover` is
deliberately NOT fixed in the normative text (that would be a fourth
change); it stays a preamble-licensed index spelling. No other sentence in
the section was touched; the document stays revision 1.

### 2b. `profiles/manager.md` §12.3

BEFORE:

> operation. Takeover of a specific unmanaged file outside onboarding
> requires the explicit takeover flag and performs the same notice and
> backup; without the flag the section 12.2 ledger discipline fails the
> operation rather than overwrite.

AFTER:

> operation. Takeover of a specific unmanaged file outside onboarding is a
> flag carried by the mutating operation performing it, never an operation
> of its own: the manager takes over only the files the carrying operation
> would write, under the environments section 9.5 notice and backup;
> without the flag the section 12.2 ledger discipline fails the operation
> rather than overwrite.

The manager states the manager-side obligation ("the manager takes over
only ...") and cites environments §9.5 for the rule; it does not restate
the closed enumeration normatively. Shape and per-file scope agree with
the amended §9.5; the §12.2 failure clause is byte-identical in meaning.

## 3. Clause → source-sentence table (every published row clause)

Amended-§9.5 shorthands: S1 = "the explicit takeover flag is carried by a
mutating operation and covers only the specific unmanaged files that
carrying operation would write — it never selects a scope of its own";
S2 = "accepted on exactly ... `profile install`, `profile use`, `profile
sync`, `profile update`, and `env resolve --repair` — and on no other
operation"; S3 = "performs the same notice and backup as onboarding when
the flag is given; without the flag, section 8.3 applies and the operation
fails with `environment_surface_unmanaged_conflict` rather than overwrite".

### 3a. The six `[--takeover]` rows

| Published clause | Source sentence |
|---|---|
| `[--takeover]` on install / use / use --clear / sync / update / resolve signatures | S1 (flag carried by a mutating operation) + S2 (closed enumeration naming exactly these operations). Spelling `--takeover` stays preamble-licensed (`cli/curator.md:3`: "Other managers may use different command names, flags"); bare-boolean shape matches `--purge`, `--repair`, `--use`. |
| "takes over the unmanaged files the install / switch / re-materialization / update / sync / repair would write" | S1 ("covers only the specific unmanaged files that carrying operation would write"). The operation noun per row names the carrying operation; no scope of its own is stated anywhere, per S1's "never selects a scope of its own". |
| "(section 9.5 notice, section 8.3 backup)" | S3 ("the same notice and backup as onboarding") composed with §9.5 step 2 (notice content: "the operator is told that native global context files are being replaced by managed ones and where the backup lands") and §8.3 ("Takeover and onboarding backups (section 9.5) land in **versioned** backup sets `.agent-environment-backup/<n>/` ... incremented per operation"). |
| "without it the write fails with `environment_surface_unmanaged_conflict`" | S3 ("without the flag, section 8.3 applies and the operation fails with `environment_surface_unmanaged_conflict` rather than overwrite") composed with §8.3 ("MUST fail with `environment_surface_unmanaged_conflict` rather than overwrite an unmanaged file"). Spelled exactly as the §8.5 table spells it. Reachable now: every carrying row names an operation that exists without the flag. |
| resolve-only: "`[--takeover]` applies only with `--repair`" | S2 enumerates the trigger as `env resolve --repair`, not bare `env resolve`; §9.5's read-only list ("`env resolve` without `--repair` — report unmanaged state and never begin onboarding, never write a backup") entails a flagless bare resolve never writes, so there is nothing to take over. Signature keeps `[--takeover]` adjacent to `[--repair]` to mirror the enumerated pair. |
| takeover example `curator profile use companyA --env claude_code --takeover` + comment | Illustrates the `profile use` row verbatim (flags carried on a real operation). `companyA`/`claude_code` reuse the neighbouring use example's operands; the comment restates S1/S3 in index voice and adds no rule. |

### 3b. Judgement recorded: `profile use --clear` carries the flag

§9.5 enumerates `profile use` without distinguishing forms, and
`cli/curator.md` publishes two rows for its two forms. §9.2: "Both remove
the scope record, re-materialize the scope from the machine default" — a
re-materialization writes in-place surfaces, so `--clear` can meet
unmanaged state exactly like the `<name>` form. Both rows therefore carry
`[--takeover]`; withholding it from `--clear` would have been the unsourced
choice.

### 3c. Import row — `curator profile import [--as <name>] [--allow-lossy] [--use]`

| Published clause | Source sentence |
|---|---|
| `curator profile` placement | §9.6: "its output is one installed, audited, locked profile whose environment markers record `imported_from_native`." |
| `import` spelling | Choice, preamble-licensed (§9.6 names no command). Kept per brief. |
| "Reassemble the section 9.5 inventory into a context-package-shaped directory" | §9.6: "Its input is the section 9.5 inventory"; "The manager assembles a context-package-shaped directory inside the machine home". |
| "install it through the ordinary `path` pipeline" | §9.6: "The assembled directory then installs through section 9.1 exactly as an operator-supplied `path` source — snapshot copy, state-hash pin, resolution of the pinned skills, always-strict audit". |
| "named `imported` unless `--as` supplies a name" | §9.6: "`name` `imported` unless the operator supplies a name under the core §2 grammar"; `--as <name>` reuses §9.1's published carrier ("the profile name is the root package's `name` unless `--as <name>` is given"). |
| `[--allow-lossy]` spelling | Choice, preamble-licensed (§9.6 names "an explicit per-operation consent flag" with no spelling); verb-led shape matches `--restore-backups` / `--allow <hash>`. Kept per brief. |
| "stops with `environment_import_lossy` and the loss list" | §9.6: "A lossy import stops with `environment_import_lossy` and the loss list". Spelled exactly as the §9.7 table spells it. |
| "unless the per-operation consent flag re-reports the list as warnings" | §9.6: "it proceeds only under an explicit per-operation consent flag, which re-reports the loss list as warnings under the same diagnostic." |
| "machine configuration never pre-records consent" | §9.6: "Machine configuration MUST NOT pre-record consent." ("never" is the install row's register for MUST NOT; it does not weaken the rule.) |
| `[--use]` + "activation follows the section 9.1 rules — `--use` takes no name, and a first install activates and says so" (NEW, finding 3) | §9.6: "Activation follows the section 9.1 rules without magic." §9.1: "`install` sets the machine current profile only when the machine has none — first install, and the activation is reported, never silent — or when the operator passes `--use`. `--use` takes no name". The import installs "through section 9.1 exactly as an operator-supplied `path` source", so both halves apply without invention. |
| example `curator profile import --as legacy --allow-lossy` | Unchanged; carries the name and consent flags and matches the row (both flags optional, so the `[--use]`-less form remains valid). |

Deleted with no replacement: the standalone `curator env takeover`
table row, which findings 1–2 exclude. No path operand and no
`--env`/`--target` were invented in its place.

## 4. Spec-surface impact statement

The §9.5 amendment touches no schema, no vector, no diagnostic table, and
no conformance case. Established by: (i) `grep -n "9\.5\|9\.6\|9\.7\|
8\.5\|takeover" tools/validate.py` returns nothing — no validator reads
the amended sentences or the §§8.5/9.7 tables; the validator's
`environments.md` reads are the §12.1 knob table, the §12.2 sentence, and
the §§1.4/6 vectors, all untouched; (ii) `grep -n "profiles/manager"
tools/validate.py tools/test_validate.py` returns nothing — the manager
edit is validator-invisible; (iii) both gates below are green on the
amended tree, and `regenerate-check` reports no diff in
`conformance/v1` or any release pin. Had any of these fired, it would be
raised here as a finding; none did.

## 5. Gates (standalone processes, real exit codes, no pipes)

`make validate` — exit 0 (venv `.temp/venv` on PATH; system python3 lacks
`jsonschema`, so a bare `make validate` fails at import — environment
fact, not a tree defect; re-ran correctly scoped):

```text
python3 tools/validate.py
validated 60 schemas and 1017 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...
Ran 227 tests in 61.088s
OK
go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors 0.524s
```

`make regenerate-check` — exit 0:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

(no diff output — byte-clean). Gates ran on the exact tree committed as
`e01de3f` (`git status` clean after the runs apart from the removed
`tools/__pycache__/`, which the gate run produced and which is untracked).

## 6. Verification bounds (only what the named method establishes)

- `cli/curator.md` row text: no committed check asserts it. Method: grep
  for `cli/` path reads under `tools/*.py`, `tools/*.go`,
  `tools/*/*.go` returns only the unrelated `tools/cli` Go-module path
  exclusions in `generate-vectors/main.go`; the sole `*.md` walk is
  `validate_local_links()` (`tools/validate.py:3310`, `ROOT.rglob("*.md")`),
  which checks link integrity only — and this batch adds no markdown
  links. So `cli/curator.md` is covered for links, not for row text.
- `CHANGELOG.md`: `tools/release_gate.py:650,722` do read it, but that
  gate runs only under `make release-check VERSION=...`, not under the two
  required gates; an entry under `## Unreleased` moves no version heading.
  This corrects the cycle-1 false sub-claim without repeating it.
- Coverage ratio: the rework brief's concrete surface is 6 amended trigger
  rows + 1 fixed import row + 1 swapped example + 2 amended spec sentences
  + the split CHANGELOG entry — 11 of 11 present in the deliverable,
  verified by reading the committed diff. There is no production entry
  point in this batch (prose-only: spec + surface index + changelog; the
  stage-(c) Go implementation is the downstream consumer), so no committed
  test drives these rows, and there is no gate to narrow: narrowing-mutant
  and token-preserving-mutant evidence is not applicable, stated as a
  bound, not as coverage. The clause-by-clause sourcing attack in §3 is
  the evidence for these rows.
- The one bare-`make validate` failure observed (system python3,
  `ModuleNotFoundError: No module named 'jsonschema'`, exit 2) is an
  environment scoping fact; the scoped runs above are the evidence.

## 7. Mechanics

- Commits past `f39f4a9`: exactly one (`e01de3f`).
- Signature: `Good "git" signature with ECDSA key
  SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` (`%G?` shows `U`
  only because the `allowed_signers` path baked into the repo config
  (`/private/tmp/curator-spec-rc8-verify.*`) is gone — identical for
  `f39f4a9` and every prior commit, not a regression).
- Identity: `Ivan Oparin <oparin@me.com>`, author and committer.
- Files: `CHANGELOG.md`, `cli/curator.md`, `profiles/manager.md`,
  `protocol/environments.md` (+49/−13). No `LOGBOOK.md`, no stray file;
  `git status` clean.
- CHANGELOG convention: new import row under `### Added` (new surface, as
  `f61ee9a` put new schema/vectors under Added); the §9.5 + manager
  amendment and the `[--takeover]` additions to existing rows under
  `### Changed` (rewrites of existing text, as `f61ee9a` put the manager
  §12 and CLI row rewrites under Changed). Not pushed, tagged, or PR'd.

