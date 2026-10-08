# TASK-260918-bi6ouz — implement the §9.5 dotfile-manager table (curator)

Control root: curator; work only in your Story worktree (STORY-260916-12lbww). Read `campaign-producer-rules.md`.
Landing gate = hosted CI, once, at handoff. Read this task's description and AC (they are exact).

Source of truth: curator-spec `802caee` (environments §9.5 closed table: chezmoi, home-manager, yadm, stow,
dotbot × macOS/Linux/Windows with explicit `none` cells; per-platform resolution rules; lstat presence;
`environment_foreign_manager_suspected` naming the manager) and `conformance/v1/vectors/environments-dotfile-managers.json`.
Today `internal/envprofile/managed.go` `dotfileStateDirs` is a POSIX-only list — replace it with ONE table
constant in the spec order.

IMPORTANT about the conformance root: the CI pin is `dced9b8` (rc.12), which PREDATES `802caee`, so the vector
file is absent from the pinned root. Therefore: (1) pin the table against the spec with a unit test that reads
the spec table from a fixture copied byte-for-byte from `802caee` (cite the path and hash); (2) the
vector-consuming test reads `CURATOR_CONFORMANCE_ROOT` and, when the vector file is absent from the supplied
root, skips with an EXISTING declared skip class (check `.github/ci/skip-classes.tsv`; if none fits, add the
narrowest class with a gate-selftest row and say why); (3) run it yourself against a local checkout of
curator-spec main (`/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, now ≥ 802caee) and report
the real pass counts. STORY-260922-2goxjs will move the pin later and turn that skip into a driven row.

Per-platform resolution tested on every OS via injected env/home; Windows cells exercised on the Windows lane;
symlinked directory does not count; failed inspection is never "absent" (§8.4.1). Docs/troubleshooting +
CHANGELOG. Attach `TASK-260918-bi6ouz_results.md` and hand off.
