# TASK-260916-19shmj — gate fix 3 (THE ONLY CURRENT INSTRUCTION, with 19shmj-sec-brief.md) — orchestrator diagnosis, apply it

Revision 3 fails on Windows with the IDENTICAL 22 failures as revision 2 (run 36308915452): `link skills/myskill:
environment_write_would_follow_link: …\skills\myskill`. Root cause (orchestrator read your code):
1. managed.go link loop → `atomicManagedLink(root, rel, target)` creates `.curator-link-<nonce>` and calls `renameManagedEntry(temp, full)`.
   On Windows that is SetFileInformationByHandle(FileRenameInfo, ReplaceIfExists=TRUE). When `full` already exists as a DIRECTORY entry
   (the manager's own earlier directory symlink `skills/myskill`, or a directory/junction a test planted), Windows refuses to replace a
   directory entry by rename → error.
2. The error sends the loop to `copyLinkFallback` → `copyTree` → `managedDirectory(root, "…/skills/myskill")`, whose final component is
   that existing symlink → `environment_write_would_follow_link`. The pre-change code did `_ = os.Remove(full)` first.
Fix (keep the §8.3.1 intent — never open/traverse a link; replacing the ENTRY is allowed):
a. In `atomicManagedLink` (and `atomicManagedFile` for file targets): when the target entry exists and is a symlink / reparse point
   (Lstat: ModeSymlink or ModeIrregular on Windows) or a directory that the replacement must supersede, remove the ENTRY itself without
   following it (os.Remove on a directory symlink/junction removes the link; for a real non-empty directory keep the existing refusal /
   credential-conflict semantics the tests expect), then rename the temp into place. On unix keep rename-over (atomic). Document the Windows
   non-atomic window as a platform bound in the results.
b. `copyLinkFallback` must never be reached because of a rename failure over an existing link: only fall back when os.Symlink itself fails
   (the platform takes no symlink). Distinguish the two errors.
c. Make sure the credential-link refusal checks (environment_credential_conflict for regular file / directory / foreign link at a recorded
   credential path) run BEFORE any replacement, as on trunk — the failing tests expect those diagnostics.
Verify locally: `GOOS=windows go vet ./internal/envprofile`, unix `go test ./internal/envprofile -run 'Credential|Link|Drift|Passthrough|Nofollow|Seed|Repair'`
(real exit codes). Also: base is still 55b94af2 — `git fetch origin main` (trunk eca2bf27), combine keeping both sides, VERIFY `git diff
--name-only origin/main -- . ':!.task-board'` lists only your paths. Set status development; append "Revision 4 — Windows entry replacement";
handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK edit.
