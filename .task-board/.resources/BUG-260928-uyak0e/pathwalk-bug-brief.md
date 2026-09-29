# BUG-260928-uyak0e — path boundary walk races a vanishing entry (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. Flake: internal/envprofile TestPathInstallCapturesDirtyUntrackedInsideGit on macOS (run 36467710280):
`profile_source_invalid: regular_types: … boundary check failed at …/root/.git/objects/maintenance.lock: lstat …: no such file or directory`.
The E6 path-source walk (internal/pathboundary) lists a directory, then lstat's each entry; git background maintenance removed the lock file
in between. Decide the semantics from environments.md §4 ("path source directories": the checks verify the directory; "later edits to the
source directory change nothing until the operator reinstalls") and implement them: an entry that no longer exists when examined was not
part of the directory at examination time — skip it (it cannot be a symlink/special file if it does not exist), while ANY other lstat error
stays a failure (fail closed), and an entry that is replaced by a symlink/special file is still caught. Add rows: entry removed between
readdir and lstat → passes; entry replaced by a symlink in that window → refused (use a test hook). Test fixtures that create git repos set
`gc.auto=0` and `maintenance.auto=false`. Show 20 repeated runs of the test green (real exit codes). Mutant: treat every lstat error as
skip → the "other error" row fails. No CHANGELOG/LOGBOOK edit. Update results, handoff, END YOUR TURN.
