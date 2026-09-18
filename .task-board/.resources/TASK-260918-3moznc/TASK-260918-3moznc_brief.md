# Brief — TASK-260918-3moznc: the absence-vs-read-failure discipline, stated once and referenced everywhere (curator-spec)

Story `STORY-260916-1ll22r` (absence-vs-read-failure-collapse), wave 3 of the
2026-09 security-audit remediation, `EPIC-260910-2hw1xb`; spec first, then the
manager task (created after this lands: inventory of read sites, shared
classification helper, build-failing guard).
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1ll22r/worktree`
(branch `task-board/story/STORY-260916-1ll22r`, already checked out at
curator-spec `main` `5146c7b` — S1+S3 landed; work only inside it). Rules: `remediation-spec-producer-rules.md`
(attached). Role: doc-writer of normative spec text; no implementation.

## Finding (read it first)
Story text: environments §8.4 requires that an absent file and a failed read
stay different facts and that a read or parse error is never downgraded to
absence; the launcher/migration campaign found the two collapsed at five
sites over three review cycles, and the producer stated there is no
structural barrier against the next one. Today §8.4 states the rule for the
marker and recorded surfaces only ("An absent surface file and a failed read
are different facts …"); seeds (§7.4), locks (§1.3), ledgers/backups (§8.3),
inventories (§9.5/§9.6), passthrough entries and `env resolve` inputs
(§10.1) each restate — or omit — it locally.

## Settled decisions (do not reopen)
- **One general rule, stated once**, in §8.4 (or a new §8.4.1 the other
  sections reference): for every file the manager reads as state or
  surface — marker, lock, ledger, backup record, provisioning seed,
  passthrough entry, recorded surface, inventory candidate — the outcomes
  "absent" (`ENOENT`-class on `lstat`/open of the exact path) and "present
  but unreadable or malformed" (permission, I/O, type, encoding, parse,
  schema errors, a component that is not a directory, a symlink where a
  regular file is required) are distinct facts; a read failure MUST NOT be
  reported, persisted or acted upon as absence, and MUST NOT trigger any
  absence-shaped action (provisioning, re-seeding, takeover, re-materialization,
  "unprovisioned" verdicts); an unreadable file makes the affected row
  non-current with currency unknown (existing wording) and the operation
  that needs the file fails closed.
- **Closed table of unreadable outcomes per file class**: marker →
  `environment_marker_unreadable` (exists); recorded surface →
  `environment_surface_unreadable` (exists); lock → the existing lock
  diagnostic if one names an unreadable lock (`environment_lock_invalid`?
  — find the landed spelling, do not invent a parallel one); seed and
  passthrough entry → reuse the existing §7.4 diagnostics if they cover a
  present-but-unreadable file, else add at most ONE new code for the seed
  class and ONE for the passthrough class (working names
  `environment_seed_unreadable`, `environment_passthrough_unreadable`), only
  where the producer shows no existing code applies; ledger/backup → the
  §8.3 discipline's existing codes. Every row: file class, absence outcome
  (what absence legitimately means there), unreadable outcome, the section
  that owns it. Closed sets stay closed: new codes appear in the §9.7 table
  and wherever the document lists diagnostics.
- **Every read-site section references the rule** (one sentence + the
  table row), never a local restatement that could drift.
- **Conformance**: vectors exercising unreadable-but-present (permission
  denied, malformed content, symlink where a regular file is required,
  directory where a file is expected) for markers, seeds, locks and
  passthrough entries, each pinned to the file class, the observed failure
  class and the expected diagnostic, with negatives that would pass under
  an absence-shaped outcome refused (rule 7: the validator gate pins every
  scenario to its typed, present discriminating inputs). Family: the
  environments family that already holds marker/surface drift cases
  (`environments.json` or a sibling — find it; a new
  `environments-read-failure.json` is acceptable if the existing family's
  generator makes extension awkward; state the choice).
- Rollout is direct (correctness of an existing MUST).

## Deliverable
1. §8.4 rule text and table; references from §1.3, §7.4, §8.2, §8.3,
   §9.5, §9.6, §10.1 (and any other read site you find — list them all in
   the evidence with the verdict "already conformant / restated → now
   referenced / silent → now covered").
2. §9.7 diagnostics rows for any new code; §13 conformance surfaces.
3. Vectors + validator gate + tests; manifest/rc.9 pins regenerated;
   existing vectors byte-identical.
4. `CHANGELOG.md` Unreleased entry naming the story and the rule.

## Out of scope
The implementation (helper, analyzer/lint guard) — the manager task; the
launcher repository; any other finding.

## Validation, patch and handoff
`make regenerate`, `make validate`, the regeneration proof (exit codes) with
the repo venv on PATH; evidence `TASK-260918-3moznc_evidence.md`;
`TASK-260918-3moznc_spec-patch_rev1.patch` = `git diff HEAD` of the worktree
(record the base commit) with new files via `git add -N`; EMPTY curator
delta; `task-board handoff TASK-260918-3moznc --role doc-writer`.
