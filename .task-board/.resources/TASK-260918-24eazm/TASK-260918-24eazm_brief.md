# Brief — TASK-260918-24eazm: one closed per-platform table for the §9.5 dotfile-manager heuristic (curator-spec)

Story `STORY-260916-12lbww` (onboarding-heuristic-platform-parity), wave 3 of
the 2026-09 security-audit remediation, `EPIC-260910-2hw1xb`; spec first,
then the manager task (created after this lands: implement from the table).
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-12lbww/worktree`
(branch `task-board/story/STORY-260916-12lbww`, already checked out at curator-spec `main` `5146c7b` — S1+S3 landed; work only inside it). Rules:
`remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
Story text: the §9.5 heuristic ("presence of a well-known dotfile-manager
state location (a closed, documented list per manager — `~/.local/share/chezmoi`,
`~/.config/home-manager`, and the like)") is by-example and POSIX-only; the
implementation (`internal/envprofile/managed.go` `dotfileStateDirs`,
TASK-260906-vlrjo1) carries two POSIX entries and explicitly notes that
Windows locations are a spec question §9.5 does not decide. Platforms must
stay in sync from one table.

## Settled decisions (do not reopen)
- §9.5 carries ONE closed table: one row per manager — at least `chezmoi`,
  `home-manager`, `yadm`, `stow`, `dotbot` — with the well-known state
  location on macOS, Linux and Windows, or an explicit `none` when the
  manager has no fixed state location on that platform (stow and dotbot
  keep no state directory of their own: say so, do not invent one), each
  cell labelled `verified` (you verified it against the manager's own
  documentation or source and cite the URL and the version/date) or
  `docs-confidence` (documented but not verified end-to-end), with the
  citation in the evidence and a short source note in the table.
- Path resolution is stated precisely per platform: the operator home
  directory (`$HOME` / `%USERPROFILE%`), XDG variables where the manager
  honours them (`$XDG_DATA_HOME`, `$XDG_CONFIG_HOME`, with their defaults),
  `%LOCALAPPDATA%` / `%APPDATA%` on Windows; a location is "present" when
  the resolved path is an existing directory (not followed through a
  symlink outside the home? — decide and state, matching §8.4/§4 link
  discipline).
- The implementation MUST read the same table (the spec says a conforming
  manager's list is exactly this table, in this order); a table change is a
  spec revision, never an implementation-private list. The heuristic still
  never blocks (`environment_foreign_manager_suspected`, warning).
- Conformance: vectors per platform — for each manager row: location
  present → `environment_foreign_manager_suspected` naming the manager;
  absent → no notice; `none` on that platform → inert; XDG override cases;
  pinned by a rule-7 validator gate (platform, manager, resolved path,
  expected outcome typed and present; name-preserving replacements refused).
  Family: a new `environments-dotfile-managers.json` (or the family that
  holds §9.5 cases — find it and state the choice).
- Rollout is direct (a finer closed list; no behaviour change for existing
  POSIX entries).

## Deliverable
1. §9.5 table + resolution rules + the implementation-reads-the-table
   sentence; §13 conformance surfaces; no new diagnostic.
2. Vectors + validator gate + tests; manifest/rc.9 pins regenerated;
   existing vectors byte-identical.
3. `CHANGELOG.md` Unreleased entry naming the story.

## Out of scope
The implementation (manager task), other heuristics, takeover semantics.

## Validation, patch and handoff
`make regenerate`, `make validate`, the regeneration proof (exit codes) with
the repo venv on PATH; evidence `TASK-260918-24eazm_evidence.md` (with the
citations); `TASK-260918-24eazm_spec-patch_rev1.patch` = `git diff HEAD` of
the worktree (record the base commit) with new files via `git add -N`;
EMPTY curator delta; `task-board handoff TASK-260918-24eazm --role doc-writer`.
