# TASK-260907-2as5sx — close the §8.4 absence-vs-failed-read collapse as a class (THE ONLY CURRENT INSTRUCTION)

Control root: curator; work only in your Story worktree (STORY-260906-1a2i5a; it already carries the checkpointed
TASK-260907-187z6x). Read `campaign-producer-rules.md`. Landing gate = hosted CI once at handoff. This is the
Story's FINAL leaf.

## Normative source
curator-spec environments §8.4 / §8.4.1 (local checkout /Users/administrator/Developer/ReluxWorks/curator/curator-spec,
main): a read of manager-owned state that FAILS (EACCES, EIO, not-a-directory, decode error…) is never treated as
ABSENT; absence is only a proven not-exist.

## Do
1. Inventory every read of manager-owned state (config, machine policy, profile/lock/marker/ledger files,
   environment surfaces, dotfile-manager table rows, compiled caches…) — `rg -n 'os\.IsNotExist|errors\.Is\(.*fs\.ErrNotExist|os\.(ReadFile|Stat|Lstat|Open)'`
   over `internal/` and `cmd/`. Find the five known collapse sites from past review cycles (search the board:
   `task-board grep "8.4"` and the review verdicts under `.task-board/.resources/` for "absent"/"unreadable"; e.g.
   TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault, profile path operand diagnostics, dotfile table scan) and
   list every other site in a table (file:line, what it reads, current behaviour on failure).
2. Close it structurally: one shared read seam (e.g. `readManagerState(path) (data, absent bool, err error)` /
   a typed `ErrUnreadable`) that every listed site uses, OR — if a seam is infeasible for some sites — a test that
   enumerates the read sites (AST/grep based, failing when a new `IsNotExist`-then-default pattern appears outside
   the seam) and asserts each distinguishes the two. Say which and why, with alternatives weighed.
3. Per site: rows proving absent and unreadable produce DIFFERENT outcomes with typed diagnostics, through the
   production entry where one exists. Unreadability via chmod 000 on POSIX; on Windows use the repository's
   existing skip class for permission-only rows (check `.github/ci/skip-classes.tsv`), never a new broad class.
4. Narrowing mutants (disposable copy only): collapse each of the five known sites back to "failure = absent";
   each killed by a named test. Table in results.
5. Coordinate, don't duplicate: security leaves TASK-1ll22r/ryh3kw (§8.4.1) are NOT yours; if your seam makes one
   of their sites trivially correct, note it in results.
6. Bounded local runs (≤10 min per call; per-package), `go vet`, CHANGELOG. Attach `TASK-260907-2as5sx_results.md`,
   check the two DoD items citing it, `task-board handoff TASK-260907-2as5sx --role developer`. A
   `run_wrote_outside_worktree … policy warn` block is a warning — verify the status moved to `to-review`.
