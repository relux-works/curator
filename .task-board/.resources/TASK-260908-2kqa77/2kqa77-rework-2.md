# TASK-260908-2kqa77 rework 2 → revision 4 (orchestrator brief, binding)

Verdict on revision 3: CHANGES_REQUESTED (`TASK-260908-2kqa77_review-verdict-rev3.md`) — three
findings, one High. The reviewer's conclusion is the one to act on: **stop extending the text walk**.

## Ruling (orchestrator): replace the awk walk with the repository's YAML parser
`go.mod` already depends on `gopkg.in/yaml.v3`, and the reviewer used it as the oracle that exposed
both bypasses. Move the check into Go:

- a small program/test under `tools/` (follow the existing `tools/` layout) that unmarshals
  `.goreleaser.yml` with `gopkg.in/yaml.v3` into an explicit struct/`yaml.Node` walk and asserts
  every `homebrew_casks[*].skip_upload`, every `scoops[*].skip_upload` and `release.prerelease`
  equals the exact string `auto`, failing with field, entry index and observed value;
- yaml.v3 rejects duplicate mapping keys, which fixes R2 for free — assert that a duplicate-key
  document FAILS (R2's fixture: `skip_upload: true` then `skip_upload: auto`);
- keep the failure wording contract from revision 1 (field + entry + observed value).

Wiring, given `Gate self-test (${{ matrix.os }})` has **no `setup-go` step** (it runs
`bash .github/ci/gate-selftest.sh` only) while `Lint` does have Go:
- run the Go check in a job that has Go (the `lint` job's step, or the interop/naming job pattern);
- keep a thin `.github/ci/goreleaser-config-gate.sh` ONLY if it merely shells out to the Go check
  (no parsing of its own), or drop the shell gate entirely and delete its self-test section;
- the self-test must still pin the WIRING. Fix R3 while you are there: pin the executable workflow
  STRUCTURE, not a substring — a commented-out `run:` line and a step carrying `if: false` must
  both fail the pin (the reviewer showed both currently pass 24/24). Parse `ci.yml` as YAML for the
  pin (the self-test may use `python3`, which every runner has, or the Go check itself can assert
  its own wiring);
- every negative must be an EXECUTED row, not prose.

## The three findings, restated for your test table
- **R1 (High)**: a first-position `- description: |` block scalar whose content line reads
  `skip_upload: auto`, and a first-position `repository:` nested map, both make the gate return 0
  while the entry has NO real `skip_upload`. Reviewer fixtures are attached to the task
  (`committed_block_bypass.yml` and the inline fixture in the verdict) — commit BOTH as negatives.
- **R2 (Medium)**: duplicate `skip_upload` keys in one cask return 0; yaml.v3 refuses the document.
- **R3 (Medium)**: the wiring pin accepts a commented invocation and an `if: false` step.

## Bound to record honestly
The reviewer's portability assessment stands: the source-level greps are a historical-regression
tripwire, not a general portability proof. If the shell gate goes away, say so and drop the claim;
if any shell remains, keep the greps but describe them as a tripwire only.

Keep everything else from revision 1–3 that the reviewer did not fault (the value semantics, the
stanza mapping to `homebrew_casks`/`scoops`/`release`, row C/E negatives, the CHANGELOG entry).
Continue from the current tree (no checkout/clean/stash), append a "Revision 4" section to
`TASK-260908-2kqa77_results.md` (design change, the three findings' rows with exit codes, the wiring
pin's new shape), then `task-board handoff TASK-260908-2kqa77 --role developer`. Publish only on a
green gate — and note trunk moved to 48da2690 (rc.12 pin), so let the base refresh happen.
