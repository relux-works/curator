# TASK-260908-2kqa77 — goreleaser config value gate (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned Story
worktree `.temp/STORY-260908-g7o5zw/worktree`. Read `campaign-producer-rules.md` first. Landing gate =
hosted CI (runtime runs it once at handoff).

Why (cycle-3 verdict §5 of TASK-260908-1jv1h3,
`.task-board/.resources/TASK-260908-1jv1h3/TASK-260908-1jv1h3_review-verdict-3.md`): `.goreleaser.yml`
keeps an rc out of the install channels through three fields — `brews[].skip_upload: "auto"` (two
entries, lines ≈75 and ≈88) and `release.prerelease: "auto"` (≈106). `goreleaser check` validates field
names/types only: it discriminated 0 of 6 wrong-value mutants (`Auto`, `sometimes`, `true`, field removed,
`"ato"`), and no in-repo gate reads `.goreleaser.yml` at all. A future typo silently restores the August
incident (rc published as the current tap release) with every lane green.

Deliverable: a gate that PARSES `.goreleaser.yml` (not a grep — row E `"Auto"` defeats a token grep) and
asserts, case-sensitively, that every `brews[*].skip_upload` is exactly the string `auto` and
`release.prerelease` is exactly the string `auto`, failing with a message naming the field, the entry and
the observed value. Implementation choice is yours, but it must run in the existing CI lanes on every push
and be pinned by `.github/ci/gate-selftest.sh` like the other gates: either a Go test under `tools/` (the
module already depends on `gopkg.in/yaml.v3`) that reads the repository file, or a `.github/ci/*.sh` gate
with a small parser — no new external binary on the runners. Negative rows (the executable evidence the
verdict asks for): row C (field absent — the pre-change state that published the rc), row E (`"Auto"`),
`sometimes`, `true`, and `prerelease: "ato"` must each FAIL the gate; the committed file must PASS. Add
them as a table-driven test / self-test so the mutants stay executed on every run, not only in results.md.

Boundaries: no change to `.goreleaser.yml` semantics, release.yml, or the ancestry gate; no goreleaser
runs (the cycle-3 live drivers already proved the semantics). CHANGELOG (unreleased) entry. Attach
`TASK-260908-2kqa77_results.md` (gate design, mutant table with exit codes, gate-selftest pin, narrow
test command + exit code) and hand off with `task-board handoff TASK-260908-2kqa77 --role developer`.
