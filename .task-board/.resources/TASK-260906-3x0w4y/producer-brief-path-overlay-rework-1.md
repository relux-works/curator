# Producer brief: path overlay rework 1 — fix the discriminator, and give it killing evidence

## Verdict

Review cycle 1 returned CHANGES REQUESTED: one blocking, two major, one minor. Read
`TASK-260906-3x0w4y_review-findings-1.md` in full.

Two things in your first revision were right and are not to be redone: the POSIX matrix is exactly
faithful to §1 and §6 (all five forbidden members refused on a `path` source, `weight` legal on both
kinds), §1 and §6 prose is untouched, the pin-consumption check was verified from source rather than
assumed, and both surviving mutants were disclosed rather than hidden. Keep all of that.

Also on the record: you were right to correct my brief. I wrote that `tools/__pycache__` is ignored;
in your workspace it is not, because the story branch is based on `c25d78e`, which predates the
`.gitignore` change now on main. Thank you — say so again if I hand you a false premise.

## F1 (blocking) — a Windows absolute path classifies as `git`

Measured by the reviewer against your committed schema:

- `C:\Users\operator\context` bare → **INVALID** — the schema demands a form, and §1 then refuses the
  declaration for carrying one. That is precisely the undeclarability this leaf exists to remove,
  reproduced on a platform `ci.yml` runs a lane for.
- `C:\Users\operator\context` + `revision` → **VALID** — the schema admits exactly the shape §1 calls
  `profile_source_invalid`.

It fails in both directions, and it holds under either reading of the spec, so it is not a bound.
Neither justification you offered survives: "the spec's path spellings are POSIX" was ruled out in the
brief you were given — examples are not a restriction — and `portablePath` governs `directory`, not
`source`, which is a bare `nonEmptyString` with no pattern.

The reviewer reports the repair is a one-character-class change: exclude a backslash after the SCP
colon, and use the core §6.1 host grammar rather than a permissive `[^/:\s]+`. It verified that repair
leaves all 43 schema cases and 15 vectors green. Apply it or something demonstrably better, and drive
the full classification matrix yourself afterwards — do not take the reviewer's repair on trust any
more than it took yours.

## F2 (major) — the discriminator has no killing evidence

The reviewer reproduced your M3 survivor independently, and its own repair pattern is *also* green
against the corpus. That is the finding: the published corpus contains only four overlay `source`
spellings — one POSIX absolute path and three `https` URLs — so no case distinguishes a correct
discriminator from a wrong one. A gate with no case that can fail it is not a gate.

Publish cases that kill it. At minimum, one per classification arm the pattern claims to handle: a
Windows absolute path (`C:\…` and `C:/…`), a project-relative spelling, an SCP `[user@]host:path`, a
`file:` URL, and each of the four schemes in a non-lowercase spelling. Each must be a positive or
negative case whose verdict flips if the discriminator is wrong. Then re-run both your surviving
mutants and the reviewer's, and report which are now killed and which still survive with the reason.

## F3 (major) — the manager profile still states the universal form

`profiles/manager.md:2189-2190` still describes the overlay as
`{ source, range | tag | revision, directory?, weight? }`, the universal shape this leaf removes. You
disclosed it and correctly noted it was outside the file scope I gave you.

**Scope call: it is in scope now.** Bring that sentence into agreement with the amended §12.1 row, in
the manager's own voice — the manager states manager-side obligations and cites the environments
section, it does not restate the rule normatively. Leaving the normative manager profile contradicting
the protocol is not an acceptable end state for this leaf.

## F4 (minor) — `svn://host/x` classifies as `path` and is admitted form-free

A scheme outside core §6.1's four is treated as a path and let through with no form. Decide from core
§6.1 and §1 whether that is correct — an `svn://` spelling is not a legal `path` (it names no
directory) and is not a legal `git` source either — and either refuse it or state plainly why the
schema is not the layer that catches it. A case either way.

## Delivery

**Exactly one signed commit past the story base** — amend `535fda6`, do not stack. Human identity, no
stray file; and note `make validate` generates `tools/__pycache__`, which is **not** ignored on your
base, so do not stage it.

Gates: `make validate` and `make regenerate-check`, each a standalone process with its observed exit
code quoted. Confirm again, from source rather than assumption, that no file the pinned Go manager
consumes gained a case it cannot pass.

Attach `TASK-260906-3x0w4y_rework-report-1.md`: the finding → resolution table for F1–F4; the full
classification matrix driven against the committed schema, both directions, including every spelling
named in F2; the mutant table re-run with the reviewer's mutant included; the manager sentence quoted
before and after; the gate tails. Then `task-board handoff TASK-260906-3x0w4y --role developer`.
