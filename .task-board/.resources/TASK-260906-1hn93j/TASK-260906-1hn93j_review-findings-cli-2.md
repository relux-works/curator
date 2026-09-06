# Review findings — TASK-260906-1hn93j, cycle 2 (the §9.5 amendment and the published rows)

Verdict: **ACCEPT**. No blocking finding, no major finding. Three minors and two
recorded spec follow-ups, none of which makes a published clause wrong.

repeat-of: **none blocking or major.** Minor 1 below is the *residue* of the
cycle-1 blocking finding's class (takeover shape stated inconsistently), now
confined to one stale list item in two normative documents that the amendment's
own operative sentence overrides two paragraphs later. It is not a recurrence of
the defect: cycle 1 rejected a *published operator surface* that documented an
unreachable refusal branch; that surface is gone and every replacement clause is
sourced. Minor 3 touches cycle-1 finding 2's "undefined behaviour published"
class and is argued down to a non-defect below.

Subject: `git diff f39f4a9..e01de3f`, 4 files, +49/−13, on
`task-board/story/STORY-260905-2z9pw4`.
Authority: `protocol/environments.md` and `profiles/manager.md` **as amended by
this commit** for the rows, and at `f39f4a9` for judging the amendment.

## 0. On the empty Change Request delta — the work exists and was reviewed

`CR-TASK-260906-1hn93j-2` reports `repository_delta: empty`, 0 changed paths,
patch sha256 `e3b0c44…855` (the SHA-256 of the empty string). That is a
base-recording artifact, not an absence of work:

```
git rev-parse e01de3f^{tree}  →  f3b96cd32716842b02eb09b5c0f627010cfa3678
candidate tree per the CR     →  f3b96cd32716842b02eb09b5c0f627010cfa3678
CR base OID                   →  e01de3f5731555457c8d3c7de6bec58b8e768f32
```

The CR base OID *is* the producer's own commit, so the snapshot records "nothing
past the already-committed head" by construction. The same shape appears on
cycle 1 (`CR-…-1` base `f013e0c`) and on six other accepted CRs on this board.

The leaf's repository change is real, is committed, and is what this review
attacks: `git diff f39f4a9..e01de3f` — `CHANGELOG.md` (+16), `cli/curator.md`
(+23/−6), `profiles/manager.md` (+6/−4), `protocol/environments.md` (+10/−3).
So this is **not** an "accepting an empty delta" decision; it is an ordinary
acceptance of a committed 4-file change whose CR snapshot was cut one commit too
late. Acceptance below is of that change, and my evidence is clause-level, not
snapshot-level.

## 1. Is the §9.5 amendment faithful and minimal? — YES

The rework brief demanded exactly three fixes. All three land, in the document's
own voice, and nothing else in the section moved.

| Required fix | Amended text carrying it | Verdict |
|---|---|---|
| Shape: a flag carried by a mutating operation, never an operation of its own, so "the operation" in the refusal clause has an antecedent | "Takeover is not an operation of its own: the explicit takeover flag is carried by a mutating operation …" and later "A carrying operation that meets unmanaged files outside onboarding …" | **FIXED.** The refusal clause's antecedent is now "a carrying operation", named in the same sentence. |
| Enumeration: exactly the five named onboarding triggers, closed | "The flag is accepted on exactly the mutating operations named above as onboarding triggers — `profile install`, `profile use`, `profile sync`, `profile update`, and `env resolve --repair` — and on no other operation." | **FIXED**, and the five are spelled and ordered identically to the ¶2 trigger list. |
| Per-file semantics preserved | "covers only the specific unmanaged files that carrying operation would write — it never selects a scope of its own" | **FIXED.** §8.3's "never wedges a second takeover of the same path" reads true (a path taken over twice by two carrying operations). §9.5's "a specific unmanaged file" reading survives verbatim in the manager sentence and in substance here. |

Minimality, checked rather than assumed:

- `git diff -U0` shows exactly **two** hunks across both normative documents:
  `protocol/environments.md @@ -1805,3 +1805,10` and
  `profiles/manager.md @@ -2294,4 +2294,6`. No diagnostics table, no other
  section, no other paragraph of §9.5 was touched. The §8.5 and §9.7 tables and
  the manager's own table are outside both hunks.
- Diagnostic tokens introduced by the whole diff, extracted mechanically:
  `environment_surface_unmanaged_conflict` and `environment_import_lossy`. Both
  pre-exist — the first in the §8.5 table ("write would touch a file the marker
  does not record"), the second in the §9.7 table (both of its rows). **No name
  added, none renamed.** Writing `environment_surface_unmanaged_conflict` where
  the old sentence said "section 8.3 applies" is a composition of §8.3's own
  MUST, not a fourth change.
- The document stays revision 1; no revision line exists to bump.

Cross-section consistency, `grep -rn takeover` over `protocol/`, `profiles/`,
`decisions/`, `cli/`, `core/` — 26 hits, every one judged:

| Hit | Consistent with the new shape? |
|---|---|
| `environments.md:1463` §8.3 "never wedges a second takeover of the same path" | Yes — per-path, which the amendment preserves. |
| `environments.md:1769` "detection, the foreign-manager stop, the replace notice, backup, takeover, and the §9.6 import" | Yes — a capability in a list of capabilities. |
| `environments.md:1774` trigger list "…, and an explicit takeover" | **Minor 1.** See below. |
| `environments.md:1803` "the takeover writes the operator chose" | Yes — writes, not a command. |
| `environments.md:1815`, `:1910` | Yes — "takeover, or import" as a class; "the §9.5 takeover path". |
| `decisions/0010:405` "the takeover **flag** remains the manual equivalent" | Yes, and it is corroborating evidence: the ADR already called it a flag. |
| `decisions/0010:500, 505, 693, 720, 841` | Yes — capability/domain prose, no command shape claimed. |
| `manager.md:2285` trigger list "…, and an explicit takeover" | **Minor 1.** |
| `manager.md:2310` | Yes. |
| `cli/curator.md:30,33,34,35,37,40,174,177` | Judged in §3 below. |

Nothing under `conformance/`, `schemas/` or `release/` mentions a takeover at
all — verified by the same repo-wide grep, and independently by
`release/*.json` containing no reference to `environments.md`, `manager.md` or
`curator.md`.

§7.6 is worth one positive note: it declares the embedded hosts' own files
(`.claude.json`, `commands/`, `config.toml`, caches) "unmanaged in revision 1
and MUST NOT be written". The amendment's "covers only the … files that
carrying operation would write" **protects** that MUST NOT — a takeover can
never reach a file the carrying operation would not have written anyway. A
scope-selecting takeover (the shape cycle 1 rejected) could have.

## 2. Does the manager sentence agree without restating? — YES

> Takeover of a specific unmanaged file outside onboarding is a flag carried by
> the mutating operation performing it, never an operation of its own: the
> manager takes over only the files the carrying operation would write, under
> the environments section 9.5 notice and backup; without the flag the section
> 12.2 ledger discipline fails the operation rather than overwrite.

Agrees with §9.5 exactly on all three points, and correctly does **not** restate
the closed enumeration — that stays in §9.5, cited. The voice is right for this
document: the operative half is a manager obligation ("the manager takes over
only …") and the rule is cited, not re-legislated. The scoping half ("is a flag
carried by … never an operation of its own") is a restatement, but the
surrounding paragraph already restates §9.5's trigger list, inventory, notice
and backup in exactly that condensed register, so it matches the document rather
than breaking it. The §12.2 failure clause is unchanged.

## 3. The six carrying rows — correct, and the `--clear` judgement is right

Six rows for five enumerated operations, because `profile use` has two published
forms. `grep -c` confirms exactly six table rows carry the flag (lines 30, 33,
34, 35, 37, 40) and no seventh.

- **`profile use --clear` is covered.** §9.3: "A scoped current is cleared in
  either of two ways … a scoped `profile use --clear`, or a scoped `profile use`
  naming the profile that is the machine default. **Both** remove the scope
  record, **re-materialize the scope from the machine default**." So `--clear`
  is a published *form* of the enumerated operation `profile use`, and it
  writes in-place surfaces, so it can meet unmanaged files. The enumeration is
  over operations, not over table rows. Withholding the flag from `--clear`
  would have been the unsourced choice, and the producer says so explicitly
  (rework report §3b). Correct.
- **`env resolve` states the `--repair` relationship exactly as the amendment
  does.** Row: "`[--takeover]` applies only with `--repair`". Amendment
  enumerates `env resolve --repair`, not bare `env resolve`; §9.5's read-only
  paragraph independently confirms bare `env resolve` "report[s] unmanaged
  state and never begin[s] onboarding, never write[s] a backup". Correct.
- **No row published the flag outside the enumeration.** `profile remove`,
  `profile compose`, `env status`, `env unmanage`, `env backups scrub`,
  `env config`, `profile list`, `profile import` — none carries it. `env
  unmanage` and `profile remove --purge` write only ledgered surfaces and
  managed homes, so they cannot meet unmanaged files; the rest are read-only or
  configuration-only.

Clause → source table for the repeated takeover clause:

| Published clause | Source sentence (amended §9.5 unless noted) |
|---|---|
| `[--takeover]` on the six rows | "the explicit takeover flag is carried by a mutating operation" + "accepted on exactly … `profile install`, `profile use`, `profile sync`, `profile update`, and `env resolve --repair` — and on no other operation". Spelling `--takeover` is preamble-licensed (`cli/curator.md:3`: "Other managers may use different command names, flags"); bare-boolean shape matches `--purge`, `--repair`, `--use`. |
| "takes over the unmanaged files the *install / switch / re-materialization / update / sync / repair* would write" | "covers only the specific unmanaged files that carrying operation would write — it never selects a scope of its own". Each row names its own carrying operation; no row states a scope. |
| "(section 9.5 notice, section 8.3 backup)" | "performs the same notice and backup as onboarding" + §9.5 step 2 (notice content) + §8.3 ("Takeover and onboarding backups (section 9.5) land in **versioned** backup sets `.agent-environment-backup/<n>/`"). |
| "without it the write fails with `environment_surface_unmanaged_conflict`" | "without the flag, section 8.3 applies and the operation fails with `environment_surface_unmanaged_conflict` rather than overwrite", composed with §8.3's MUST. Spelled exactly as the §8.5 table spells it. **Reachable now** — every carrying row names an operation that exists without the flag, which is precisely what cycle-1 finding 1(i) demanded. |
| example `curator profile use companyA --env claude_code --takeover` | Illustrates the `profile use` row verbatim. `companyA` is created by an earlier example (`--as companyA`); `claude_code` is a §7.1 adapter id. |

## 4. The import row — `[--use]` is right, and no `[--takeover]` is right

`[--use]` and its clause ("activation follows the section 9.1 rules — `--use`
takes no name, and a first install activates and says so") match §9.1
word-for-word in substance: "`install` sets the machine current profile only
when the machine has none — first install, and the activation is reported, never
silent — or when the operator passes `--use`. `--use` takes no name." And §9.6
imports it explicitly: "Activation follows the section 9.1 rules without magic."
It is the same clause the install row one line above already publishes. Cycle-1
finding 3 is resolved exactly.

**Does §9.6's "The import writes nothing into any native home by itself" settle
why `profile import` carries no `[--takeover]`? Yes for the import proper — and
it leaves a seam at the activation half.** Stated either way, as asked:

- *Settled part.* The import reads the §9.5 inventory, assembles a directory
  inside the machine home, and installs it into the store. It never writes a
  native home, so it can never meet an unmanaged file, so it needs no takeover.
  The very next clause names the alternative: "replacing native files remains
  the section 9.5 takeover path with its notice and backup." Publishing
  `[--takeover]` on the import row would have contradicted §9.6 and widened a
  closed set. **Correct as published.**
- *Seam.* "by itself" is load-bearing. The row also publishes `[--use]`, and
  §9.1's activation fires *without* `--use` when the machine has no current
  profile — the exact machine class the import exists for. Activation is
  `profile use` semantics (§9.2 step 1: "re-materializes every in-place surface
  …"), which does write native homes. So `curator profile import` on a
  hand-maintained machine installs, then activates, then meets unmanaged files,
  then refuses with `environment_surface_unmanaged_conflict` — with no flag on
  that command to proceed. See Minor 2.

## 5. Surface-index discipline — no row states anything `environments.md` does not

Checked clause by clause above; every published clause maps to a sentence.
Two specific discipline checks:

- The rows state **no default scope, no path operand, no scope flags** for the
  takeover — exactly the silence cycle-1 finding 2 demanded, and the silence is
  now *sourced* ("it never selects a scope of its own") rather than merely
  absent.
- The import row's account of the consent gate compresses "MUST NOT pre-record
  consent" to "never pre-records consent". Register change, not a weakening;
  the install row uses the same register.

## Minor 1 — §9.5 ¶2 and `manager.md` still enumerate "an explicit takeover" among the *mutating operations*

**File/section:** `protocol/environments.md:1772-1774`; `profiles/manager.md:2283-2285`.

**Quoted, environments.md:**

> Onboarding is triggered only by a **mutating** profile operation that meets
> unmanaged state — `profile install`, `profile use`, `profile sync`, `profile
> update`, `env resolve --repair`, **and an explicit takeover**.

**Quoted, manager.md:**

> Onboarding runs only on a mutating profile operation that meets unmanaged
> state — `profile install`, `profile use`, `profile sync`, `profile update`,
> `env resolve --repair`, **and an explicit takeover** — never on a read-only
> command…

**What is wrong.** Grammatically the dash-list enumerates *mutating profile
operations*, and its sixth member is "an explicit takeover". Three paragraphs
later (and one paragraph later, in the manager) the amendment says "Takeover is
not an operation of its own". After the amendment those two sentences
categorise the same thing two ways, in both normative documents. Before the
amendment they were ambiguous in the same direction and so did not clash; the
amendment makes one precise and leaves the other stale.

**Why it is minor and not major.** The operative sentence is unambiguous and
self-closing — it re-enumerates five members *and* excludes everything else
("and on no other operation") — so no implementer can build a standalone command
against it, and nothing downstream (schema, vector, diagnostic, CLI row) is
affected. The stale item is redundant under the charitable reading ("a mutating
operation carrying the explicit takeover flag") and merely miscategorised under
the strict one.

**Fix (one clause, either document).** Drop the sixth list member, or make it
subordinate: "… `env resolve --repair` — including when one of them carries the
explicit takeover flag." Recommended for the follow-up batch that touches §9.5
next, not for a rework cycle of its own.

## Minor 2 — the closed set excludes two operations that §9.1/§9.4/§9.6 say do write in place

**Where:** amended `protocol/environments.md` §9.5, closed enumeration.

Two operations outside the five can reach an in-place write and therefore an
unmanaged file, with no flag available:

1. **`profile import`'s activation half** (§4 above). Note the asymmetry: the
   amendment grants `profile install` the flag, and §9.6 says the import
   "installs through section 9.1 **exactly as** an operator-supplied `path`
   source" with the same activation rules. The import is an install in every
   respect but its source assembly, and it is excluded.
2. **`global add` / `global install` and kin** (§9.4: "the resolved skills
   materialize into that profile's managed homes and — for the current profile
   of each scope — **the in-place adapter surfaces**"), which can meet an
   unmanaged, non-ledgered global skills entry (§9.6 detects exactly those).

Neither is a dead end: `profile sync --takeover` and `profile use <name>
--takeover` both re-materialize the same surfaces and both carry the flag, so
the operator always has a second step. The set is workable.

**Why this is not a rework item.** The producer brief closed the set on the
orchestrator's explicit instruction ("That set is closed … do not widen it and
do not leave it open"), from §9.5's own candidate list — the only candidate list
the document supplies. Widening it here would have been the producer inventing
normative scope, which the AC forbids. Closing the seam means either widening
the enumeration or adding a sentence to §9.6, and both are the orchestrator's
call, not this leaf's. **Recorded as a spec follow-up for STORY-260905-2z9pw4.**

## Minor 3 — two small editorial residues (argued, not blocking)

- **Six-fold verbatim repetition.** The same ~25-word takeover clause now
  appears in six rows. Judged, as the brief asks, rather than passed: the
  repetition is **acceptable but not the best shape**. For it: every row in this
  table is self-contained and semicolon-dense today, and a reader scanning for
  `profile sync` gets the whole contract without a cross-reference. Against it:
  no existing clause is repeated verbatim across rows (the `--env`/`--target`
  rows word it differently each time), the document already carries cross-cutting
  rules as prose after the table (exit codes, `bootstrap --if-missing`, the
  `curator run` dispatch paragraph), and — the one substantive point — six
  per-row copies carry the enumeration's *members* but not its *closure*: from
  the table alone a reader cannot tell whether the flag's absence on `profile
  remove` is meaningful. Nothing unsourced is stated either way, so this is a
  shape preference, not a defect. A single note under the table
  ("`--takeover` is accepted on exactly these operations — environments §9.5 —
  and on no other") plus a shortened per-row clause would be strictly better.
- **`env resolve --takeover` without `--repair` is behaviourally determined but
  syntactically unstated.** The row says the flag "applies only with `--repair`";
  whether a bare `env resolve --takeover` is exit-2 invalid flag use or silently
  inert is not stated. This *touches* cycle-1 finding 2's "undefined behaviour
  published as operator surface" class but is **not** a recurrence: §9.5's
  read-only paragraph fixes the behavioural outcome (nothing is written, nothing
  is taken over), and §9.5 states no diagnostic for a flag given where it does
  not apply — so publishing one would have been an invention. Correctly left
  unstated.
- **Example placement.** The new takeover example sits at the end of the block,
  after `env unmanage`, inherited from the deleted standalone row's position,
  although it now illustrates `profile use`, whose own example group is ~20
  lines above. Purely cosmetic.

## Mechanics — all green, re-run by me

| Check | Result |
|---|---|
| Commits past `f39f4a9` | Exactly one, single-parent: `git log -1 --format=%P e01de3f` → `f39f4a9…`; `git rev-list --count f39f4a9..e01de3f` → `1`. The cycle-1 commit `f013e0c` is gone as required. |
| Signature | `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`. `%G?` = `U` because the repo's `allowed_signers` path (`/private/tmp/curator-spec-rc8-verify.*`) no longer exists — **identical for `f39f4a9`** (`%G? %GK` → `U SHA256:V6JiKG…`), so not a regression. |
| Identity | `Ivan Oparin <oparin@me.com>`, author **and** committer. |
| Files | `CHANGELOG.md`, `cli/curator.md`, `profiles/manager.md`, `protocol/environments.md`. No `LOGBOOK.md`, no stray file. Tree clean. |
| `CHANGELOG.md` placement | Verified by section offsets, not by the report's claim: `## Unreleased` → `### Changed` spans lines 8–29 and holds the takeover amendment (line 20); `### Added` spans 31–104 and holds the import row (line 98). Matches the `f61ee9a` convention (rewrites of existing text under Changed, new surface under Added). |
| `make validate` | **exit 0**, re-run by me with `.temp/venv/bin` on `PATH`: `validated 60 schemas and 1017 vector files`; `Ran 227 tests in 64.238s / OK`; `ok github.com/relux-works/curator-spec/tools/generate-vectors 0.581s`. Log `.temp/review-rev2/validate2.log`. (A bare `make validate` without the venv fails at `ModuleNotFoundError: No module named 'jsonschema'`, exit 2 — environment fact, reproduced and logged at `.temp/review-rev2/validate.log`, matching the producer's note.) |
| `make regenerate-check` | **exit 0**, re-run by me, `git diff --exit-code` silent — byte-clean. Log `.temp/review-rev2/regenerate-check.log`. |
| Worktree after gates | Clean; I removed the `tools/__pycache__/` my own run produced. |
| Amendment touches schema/vector/diagnostic table/conformance case? | **No** — established by check, not by reading the claim: `git diff -U0` shows two hunks, neither inside a table; the only diagnostic tokens in the diff are two pre-existing names; no `takeover` string exists under `conformance/`, `schemas/` or `release/`; `release/*.json` reference no spec document; `regenerate-check` is byte-clean. |

## The cycle-1 minor (finding 4) — resolved, and the method now establishes the claim

The rework report's §4 and §6 bounds were re-run by me verbatim:

- `grep -nE "9\.5|9\.6|9\.7|8\.5|takeover" tools/validate.py` → no output. ✓
- `grep -rn "profiles/manager\|manager\.md" tools/` → no output. ✓
- `grep -rn "cli/" tools/ --include='*.py' --include='*.go'` → only
  `tools/generate-vectors/main.go:254-296`, the unrelated `tools/cli` Go-module
  *fixture path* strings. ✓

The false `CHANGELOG` sub-claim is not repeated, and `release_gate.py:650,722`
is now correctly cited as real-but-out-of-gate. Bounds established by the method
that establishes them. **Resolved.**

## Evidence: I attacked the gates rather than reading them

Because this batch ships no product code, the risk is the inverse of the usual
one: that the green gates get read as evidence *for* the rows. They are not. I
established that by mutation, with a live control, not by grep.

| Mutant | Shape | `tools/validate.py` | 227 unittests | `go test` | Meaning |
|---|---|---|---|---|---|
| **A — widen the closed enumeration** | Token-preserving *narrowing-inverse*: adds `` `profile import` `` and `` `profile remove` `` to the amended §9.5 closed set while keeping every token, including "and on no other operation" | exit 0 | exit 0 | exit 0 | **No committed check constrains §9.5's enumeration.** A gate that admitted a widened closed set would have to fail here; none exists. |
| **B — rename the published flag** | Token-preserving: `--takeover` → `--seize-takeover` on all 8 `cli/curator.md` occurrences; the searched-for substring `takeover` survives in every line | exit 0 | exit 0 | exit 0 | **No committed check constrains the CLI row flag spellings.** A source-text checker keyed on the token would pass this mutant too — which is why the mutant preserves the token. |
| **C — control, broken local link** | `../profiles/manager.md#12-…` → `../profiles/NO_SUCH_FILE.md#x` in `cli/curator.md` | **exit 1**: `validation failed: …/cli/curator.md: broken local link: ../profiles/NO_SUCH_FILE.md#x` | exit 0 | exit 0 | **The harness is live and does read `cli/curator.md`.** So A's and B's green are a measured absence of coverage, not a dead harness. |

Each mutant was applied to a full tree copy under `.temp/review-rev2/mutant-*`
and confirmed applied before the run (mutant A's first attempt did **not** apply
— an em-dash the regex missed — and was rebuilt and re-verified by diff before
being run; a mutant that does not apply proves nothing). The harness ran the
**behavioural** suite (227 unittests + `go test`), not the static checker alone.

**Stated bound.** `cli/curator.md`, `profiles/manager.md` and §9.5's prose are
covered by `validate_local_links()` (`tools/validate.py:3311`,
`ROOT.rglob("*.md")`) **for link integrity only**. No committed test asserts a
row's text, a flag spelling, or the enumeration. There is no gate to narrow for
this batch and none is missing — the deliverable is normative prose. The
evidence for correctness is the clause-by-clause sourcing attack above, run
against the amended text and against `environments.md` at `f39f4a9`.

## AC coverage — 7 of 7 rows pass

| AC row | Driving check | Result |
|---|---|---|
| Takeover surface traceable to §9.5/§8.3/§7.6 | clause table §3, sentence by sentence, against the amended text | PASS |
| Import row with optional profile name + lossy consent flag traceable to §9.6 | clause table §4 and cycle-1's, re-verified | PASS (plus `[--use]`, sourced §9.6 → §9.1) |
| Flag spellings consistent with the published family | compared against `--env`, `--target`, `--as`, `--purge`, `--repair`, `--use`, `--restore-backups`, `--allow <hash>` | PASS |
| Examples block gains one line per new row | `cli/curator.md:147-150` (import) and `:174-177` (takeover) | PASS |
| `make validate` and `make regenerate-check` green | both re-run by me, exit 0/0 | PASS |
| Exactly one signed commit past `f39f4a9`, human identity, no stray files | `git log -1 --format=%P`, `--show-signature`, `--name-only` | PASS |
| Report carries row → sourcing table with quoted sentences plus open questions | read in full; its three bounds re-run and confirmed | PASS |

Rework-brief rows: §9.5 amendment faithful and minimal — PASS (§1). Manager
sentence agrees without restating — PASS (§2). Standalone row and its example
deleted with no replacement grammar — PASS. CHANGELOG split under the right
headings — PASS.

**No production entry point exists for any of this** — the deliverable is the
specification itself; the stage-(c) Go implementation is the downstream
consumer. That is a stated bound, not a coverage claim.

## Verdict

**ACCEPT.** The blocking cycle-1 finding is resolved at its root rather than
papered over: the shape gap is closed in the normative text, the enumeration is
closed and sourced, the standalone row and its unsourced scope grammar are gone,
and the refusal clause every row now publishes is reachable. Both majors and the
minor are resolved, and the resolutions were re-verified rather than accepted on
the report's word. The three minors above are one-clause editorial items and one
spec follow-up; none makes a published clause wrong, and none is worth another
producer/reviewer cycle. They belong on the Story's follow-up list.
