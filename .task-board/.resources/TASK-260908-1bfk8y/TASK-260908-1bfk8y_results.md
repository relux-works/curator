# TASK-260908-1bfk8y — stale pin rationale in platform-exclusions.tsv: results

## Outcome

Kept the `default_excluded_on` column (live consumer exists) and corrected the
TSV rationale prose. Comment-only change; no exclusion row changed, no
behaviour change, no CHANGELOG entry.

Changed file: `.github/ci/platform-exclusions.tsv` (comment lines only).

## Before / after

BEFORE (`.github/ci/platform-exclusions.tsv:10-13`):

```text
# `default_excluded_on` is the fallback for a root that predates the vector.
# The committed released pin is such a root: it publishes no qualification
# vector at all, so without a fallback the default lane would lose an exclusion
# that is a property of the implementation rather than of the root.
```

AFTER (`.github/ci/platform-exclusions.tsv:10-16`):

```text
# `default_excluded_on` is the fallback for a SUPPLIED root that predates the
# vector (e.g. an older CURATOR_CONFORMANCE_ROOT passed explicitly). The
# committed released pin is NOT such a root: it publishes
# `vectors/conformance-claim-v3-qualification.json`, so on the default lane the
# vector is authoritative and this fallback stays dormant. Without it, a
# pre-vector root would lose an exclusion that is a property of the
# implementation rather than of the root.
```

Data-row integrity: non-comment lines of the TSV are byte-identical before and
after (`cmp` of `git show HEAD:...` vs worktree filters: identical). `git
status` shows only `M .github/ci/platform-exclusions.tsv`.

## Pin verification (the old claim is false)

- `.github/workflows/ci.yml:52`: `SPEC_PIN: 87a0d0060bad64ab883d007dcdf35df7485368bf`
- `git -C <curator-spec> ls-tree -r --name-only <pin> | grep qualification`
  → `conformance/v1/vectors/conformance-claim-v3-qualification.json` (exists)
- Vector content at that pin: `platforms[linux].status = "excluded"` with
  `until_task: TASK-260728-1skseh`, `protocol_version: 1.0.0-rc.9` — matching
  what the TSV already states normatively at lines 17-20. So on the default
  lane the vector excludes linux authoritatively and the fallback is dormant
  but consistent.

## Consumer analysis (file:line) — fallback is live, column kept

Reader — `.github/ci/excluded-packages.sh`:

- `:61` — `while IFS=... read -r pkg assertcase defaults _note` reads TSV
  column 3 into `defaults`.
- `:41`, `:44-59` — detects `$ROOT/vectors/conformance-claim-v3-qualification.json`
  and computes `vector_excludes_this_goos`.
- `:65-74` — vector present → the vector decides, `defaults` ignored
  (`source_of_truth="the root's own ..."`); vector absent → `defaults` matched
  against the wanted GOOS (`source_of_truth="default_excluded_on (this root
  publishes no ...)"`).

Callers of the reader:

- `.github/ci/suite-plan.sh:89` — resolves the exclusion set via
  `excluded-packages.sh "$GOOS" "$ROOT"`; `:96-99` reports which source
  applied; `:113-129` writes `plan-excluded.txt` / `plan-assert.txt`.
- `.github/ci/ledger-consistency.sh:90` — same helper for the per-platform
  exclusion set, so the gate and the run cannot disagree.

Non-reader (checked): `.github/ci/platform-case-gate.sh` never reads
`default_excluded_on`; it consumes the already-resolved `CI_EXCLUDED_PKGS`
(`:39`, `:266-269`).

Self-test cover for the fallback — `.github/ci/gate-selftest.sh:577-581`:
a synthetic pre-vector root must still exclude godriver on linux and the
report must name `default_excluded_on`.

Functional proof (this session, `bash`, real exits):

- `excluded-packages.sh linux <root-WITH-vector>` → godriver excluded by
  `the root's own vectors/...-qualification.json`, exit 0
- `excluded-packages.sh darwin <root-WITH-vector>` → empty, exit 0
- `excluded-packages.sh linux <root-WITHOUT-vector>` → godriver excluded by
  `default_excluded_on (...)`, exit 0
- `excluded-packages.sh darwin <root-WITHOUT-vector>` → empty, exit 0

Decision: a real consumer exists, so per the task scope the column and reader
stay and only the prose is fixed.

## Validation (real exit codes, `bash`, `set -o pipefail`)

- `bash .github/ci/gate-selftest.sh` → exit 0, `185 passed, 0 failed` (run
  twice; full log at `/tmp/1bfk8y-selftest.log` on host e11-1). Includes:
  `ok a pre-vector root still excludes godriver on linux` and
  `ok the fallback exclusion is recorded as such`; no `FAIL` lines.
- `bash .github/ci/ledger-consistency.sh /tmp/1bfk8y-lc` → exit 0,
  `241 rows checked`, `ledger-consistency: ok`.
- Lint/build: no `.go`, shell, or workflow file changed (TSV comments only),
  so gofmt/vet/lint have no delta; nothing to rebuild.

CHANGELOG: untouched — comment-only change, no behaviour change (per AC 3).

## Finding for follow-up (left untouched per scope)

`.github/ci/excluded-packages.sh:12-14` header repeats the same stale sentence
("...applies only to a root that predates that vector -- the committed
released pin is such a root."). Task scope limits this change to the TSV prose
when a consumer exists, so it was deliberately not edited. Recommend a
one-line follow-up to reword it the same way. `.github/ci/root-artifacts.tsv:19`
("The committed SPEC_PIN is now such a root...") is a different, unaffected
claim about artefact coverage, not the qualification vector.
