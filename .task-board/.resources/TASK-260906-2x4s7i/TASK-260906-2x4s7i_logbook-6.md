# Logbook — cycle-6 review of the path-overlay discriminator (PR #47)

## The review subject moved under the reviewer, and only checking made that visible

The brief named `407424e`. Ten minutes into the review, `gh pr view 47` returned
`state: MERGED` with `headRefOid: 87a0d00` — a commit that did not exist when the brief was
written. Five adversarial cycles had all stopped one commit short of what is now `main`.

The habit that caught it was cheap: read the PR's *current* head OID before trusting the
brief's, rather than running `gh pr checks <n>` and reading only the lane colours. `gh pr
checks` would have shown eight green lanes and said nothing about which commit they belonged
to; `gh api .../commits/<sha>/check-runs` ties each run to a SHA and is the only form that
answers "did CI actually run on what landed".

**Generalisable:** a review verdict that does not name an OID is not falsifiable later. A
brief's subject is a claim about the world at authoring time, and long reviews outlive it.

## Identical blobs are a behaviour-invariance proof, and they turn a re-review into minutes

Rather than re-deriving the whole matrix at the new head from scratch, the first question was
whether the new commit *could* have changed behaviour: `git rev-parse <rev>:<path>` on the
schema, `core.md`, `environments.md` and `manager.md` at both heads. All four identical →
no classification verdict can have moved, and the entire `407424e` analysis transfers. Then
the diff of the two independently-driven matrices confirmed it empirically instead of taking
the inference on faith.

Cycle 5 found this technique too. It is worth keeping: for a change whose behaviour lives in
one artefact, blob equality is a stronger and cheaper invariance statement than any test run.

## A mutant sweep needs its harness falsified before its results are believed

Three controls, all of which could have silently invalidated the sweep:

1. **The unmutated corpus must exit 0.** Otherwise every mutant "dies" for free.
2. **A no-op mutant must survive.** Otherwise the harness is killing on something incidental.
3. **The conformance manifest must not digest the mutated file.** This one nearly bit: the
   repository digests 1,046 files under `conformance/v1`, and had `schemas/v1` been among
   them, *every* mutant would have died on a digest mismatch and the sweep would have
   reported perfect coverage while proving nothing. Checked: 0 manifest paths under
   `schemas/`.

Also: mutants were applied *structurally* (parse → mutate the named node → re-serialise) with
an assertion that the serialised document actually changed. A string-substitution mutant that
silently fails to match reads as "SURVIVES" and looks like a coverage finding.

## The mutant family that finds real gaps is case-narrowing, not deletion

The kill that mattered most this cycle — and the one that closed cycle 5's F19 — came from
narrowing a character class by *case*: `^[A-Za-z]:[\\/]` → `^[A-Z]:[\\/]`. Deleting the
carve-out entirely dies instantly on any Windows case; narrowing it to uppercase-only landed
green across the whole corpus at `407424e`, because all three published Windows positives
happened to spell the drive `C`.

**Generalisable:** when a corpus pins a class by example, check whether every example shares
an incidental property (letter case, one separator, one length). Narrow the gate to exactly
that property. If it survives, the class is pinned by coincidence.

## Hunt the bypass as a measured invariant, not as an argument

Core §6.1's mandate is *"rejected, not treated as local"*. The tempting write-up is a
paragraph explaining why the three arms compose to satisfy it. The better artefact is one
line of code over the whole corpus: *of every spelling classified `path`, how many are
colon-shaped without being a Windows drive?* Zero of 120. That converts "I reasoned about the
arms" into "I looked for the counterexample and there isn't one", which is the difference
between a review and a reading.

## Prose drifts from behaviour four times in a row when nothing checks it

Cycle 3's F11 (stale PR body), cycle 4's F16 (stale CHANGELOG), cycle 5's F20/F21, and this
cycle's F23/F24 are all the same defect: the schema is machine-checked from five directions
and every sentence *about* the schema is checked by a human reading it once. The count word
"eight" in front of a nine-item list survived three review cycles. No gate in this repository
reads the CHANGELOG or the PR body against the schema, and none plausibly could.

Corollary already visible in the same repository: 25 schema-case files ship inside the
digest-covered conformance manifest but are named by no `index.json` entry, so `validate.py`
never drives them. Two inventories, nothing reconciling them — the same shape as cycle 4's
"no lane inspects the tracked file set", which is how a 5.4 MB binary passed nine green lanes.
