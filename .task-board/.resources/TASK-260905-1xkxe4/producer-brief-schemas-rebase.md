# Producer brief: rebase the schemas/vectors head onto main, gofmt, one signed commit

## Where
Draft worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-schemas`, branch
`draft/environments-schemas-1-1`, head `794c7bd` (the cycle-1-accepted `401b665` minus `LOGBOOK.md`),
PR https://github.com/relux-works/curator-spec/pull/42. Work HERE; the managed story worktree of
STORY-260905-1z93ju stays at `401b665` untouched. Do not run the validator inside the story worktree.

## Why
curator-spec main advanced to `f61ee9a` (batch 3: manager §12, cli rows, manager-config schema 2 with a new
`vectors/manager-config-v2.json`, regenerated manifest and rc.9 pin, CHANGELOG/README entries). `git rebase
origin/main` of `794c7bd` conflicts in `CHANGELOG.md`, `conformance/README.md` (prose merges) and in
`conformance/v1/manifest.json`, `conformance/v1/schema-cases/index.json`, `release/1.0.0-rc.9.json` (generated
files). PR #42's `Formatting` check also fails: `gofmt -l tools/` lists `context_detectors.go`,
`context_versions.go`, `environments.go`.

## Do
1. `git fetch origin && git rebase -S origin/main`; resolve `CHANGELOG.md` and `conformance/README.md` by keeping
   BOTH batches' entries (batch 3 already on main; add this batch's bullets under `## Unreleased`, no duplicate
   headings); for the three generated files take either side, then regenerate: `make regenerate` (venv
   `.temp/venv`), `gofmt -w tools/generate-vectors/*.go`, `make validate`, `make regenerate-check` — all green.
2. Keep exactly ONE signed commit past `origin/main` (amend into the rebased commit; `git log --oneline
   origin/main..HEAD` shows one line). Prove identity to the reviewed content: `git range-diff origin/main
   794c7bd HEAD` — paste it; the only non-`=` hunks must be the conflict resolutions, regeneration, and gofmt.
3. Confirm no consumed-by-pin file changes except identical regeneration: `git diff origin/main --stat --
   conformance/v1/vectors/manager-config.json conformance/v1/vectors/manager-config-v2.json
   conformance/v1/vectors/canonical-valid.json conformance/v1/vectors/canonical-invalid.json
   conformance/v1/vectors/source-identities.json conformance/v1/vectors/identifiers.json
   conformance/v1/vectors/locale-selectors.json conformance/v1/expected/snapshot_sha256.txt` — paste it (expected: empty or manifest-only).
4. `git push --force-with-lease origin HEAD:refs/heads/draft/environments-schemas-1-1`; then `gh pr checks 42 --watch`
   until every check is green (Formatting, Specification ×3, Implementations ×3); paste the summary.
5. Attach `TASK-260905-1xkxe4_rebase-report.md` (new head, range-diff, conflict resolutions, gate tails, check summary)
   and `task-board handoff TASK-260905-1xkxe4 --role developer`. Never write LOGBOOK.md or anything into the control
   root or the repository.
