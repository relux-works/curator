# TASK-260926-1ddneb — trust the Relux Bot release signer (THE ONLY CURRENT INSTRUCTION)

Operator decision 2026-09-26 (option b). Read the task description and AC. Work in your Story worktree only.
1. Append to maintainers.allowed_signers exactly: `bot@relux.works ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPG7xTX05HL1XaD4XLUk0/TTeqRNHbMj5HdnqNQdDTID` (one line; keep the oparin@me.com line byte-identical; keep the
   file's trailing-newline convention).
2. Prove it: in a disposable clone, `git -c gpg.format=ssh -c gpg.ssh.allowedSignersFile=<candidate file> verify-commit a21905d` (a Relux
   Bot-signed commit on main) exits 0, and verify-tag/verify-commit for an oparin@me.com-signed object (e.g. tag v1.0.0-rc.9) still exits 0.
   Real exit codes.
3. Docs: GOVERNANCE.md / RELEASE.md only where they enumerate trusted signers; CHANGELOG entry under Unreleased.
No LOGBOOK.md. Attach results, check DoD, `task-board handoff TASK-260926-1ddneb --role developer`. A write-boundary `policy warn` block is a warning.
