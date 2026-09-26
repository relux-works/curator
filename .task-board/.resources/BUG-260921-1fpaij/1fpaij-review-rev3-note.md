# Review note for BUG-260921-1fpaij revision 3 (orchestrator, binding)

Review REVISION 3 only. History: rev1 (brief 1fpaij-brief.md) made every non-zero
`git check-ignore` outcome an error and broke the CLI dry-run flows in non-git projects
(gate 35558711655); an auto-successor then published rev2 that kept the strict classifier
and "repaired" the CLI fixtures (`cmd/curator/main_test.go`) — that revision is superseded
and must NOT be accepted; the orchestrator's rework-1 ruling (1fpaij-rework-1.md) narrowed
the rule and rev3 implements it: an error is returned iff `cmd.Run()` fails with a
non-`*exec.ExitError` (spawn/exec failure: `*exec.Error`, `exec.ErrNotFound`, `*fs.PathError`
like `fork/exec …: permission denied`); EVERY git exit status (1 = not ignored, 128 = not a
repository, other) keeps the pre-existing "not ignored" policy outcome; no retry; success
path and the `generated paths are not ignored by git; …` message byte-identical;
`main_test.go` untouched. Gate green: run 35575924994 — verify the gate commit resolves to
the exact revision-3 tree (patch `BUG-260921-1fpaij_change-request_rev3.patch`, 5 paths:
CHANGELOG.md, internal/gitignore/{gitignore.go,gitignore_test.go},
internal/install/gitignore_spawn_test.go, internal/install/install.go).

Judge with your own reruns (disposable clone; bounded commands; retry once on host stalls):
1. Classifier sits exactly on the ruling boundary: rows (a) injected non-executable git
   (0755 script with a 0644 shebang interpreter) → wrapped error carrying the spawn
   diagnostic; (b) non-repository root → NOT an error, "not ignored" as before; (c) exit 1
   unchanged; (d) production entry `install.Project` with (a) → `failed` with the spawn
   diagnostic, never the not-ignored message; mutants: restore the conflation → (a)/(d)
   fail; over-narrow (treat exit 128 as error) → (b) fails. Rerun the two CLI tests the rev1
   gate broke (`TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed`,
   `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot`) on rev3.
2. Sanitization of the wrapped error (entry name only; no probe path, environment or root);
   `install.go` caller keeps fail-closed mapping (what changed there and why).
3. Legacy goldens green; CHANGELOG entry present and accurate to the narrowed rule; Windows
   skips only via the ledger vocabulary.
4. The producer reports a "supplementary-run blocker (environmental)" in results_rev3.md §—
   read it and judge whether anything it left unrun is required evidence (run it yourself
   if cheap).
Record exactly one verdict: accept_cr(BUG-260921-1fpaij, revision=3, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
