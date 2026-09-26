# BUG-260921-1fpaij rework 1 (orchestrator, binding)

Revision 1 gate FAILED (run 35558711655; Test ubuntu/windows/macos + Race lanes):
`cmd/curator :: TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed/portable`
(`main_test.go:345: portable install dry-run exit=1`) and
`TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot` (`main_test.go:703: upgrade --dry-run = 1`):
these CLI flows run in a project that is NOT a git repository (or where `git check-ignore`
exits 128), which the product legitimately treats as "nothing is ignored" (a policy
outcome), and revision 1 turned that into a hard error.

Orchestrator ruling correcting brief R1: ONLY a spawn/exec failure (a non-`*exec.ExitError`
error from `cmd.Run()`: `*exec.Error`/`exec.ErrNotFound`, `*fs.PathError` such as `fork/exec
…: permission denied`) is returned as the wrapped error. EVERY exit status — 1 (not ignored)
AND 128/other (not a repository, fatal git error) — keeps the pre-existing behaviour: the
entry is reported as not ignored. Adjust the classification accordingly, keep the spawn rows
(a) and (d) and the mutant, and convert row (b) (non-repository root) into a row asserting the
PRE-EXISTING outcome (not ignored, no error). Update results.md, CHANGELOG wording and the
docs/comments to the narrowed rule ("a git that cannot be executed is an error; git's own
verdicts, including 'not a repository', are policy outcomes as before"). Rerun the two failing
CLI tests locally plus `go test ./internal/gitignore/ ./cmd/curator/ -run
'Gitignore|DryRun|ExecutionAssurance' -count=1`. Continue from the revision-1 tree in the
Story workspace (no checkout/clean/stash); republish (revision 2) only when the gate is green.
