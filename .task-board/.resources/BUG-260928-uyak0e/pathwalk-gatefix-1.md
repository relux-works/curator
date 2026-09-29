# BUG-260928-uyak0e — Windows gate fix (THE ONLY CURRENT INSTRUCTION, with pathwalk-bug-brief.md)

Rev1 (tree a970450d) is green on linux/macOS; Windows fails your new row (run 36482327099):
internal/pathboundary TestValidateSkipsEntryRemovedBetweenReadDirAndLstat — `link_safety: … boundary check failed at …\source\removed-directory:
The system cannot find the file specified.` On Windows the per-entry link-safety probe (the reparse-point / handle-based check, not the
Lstat) reports ERROR_FILE_NOT_FOUND / ERROR_PATH_NOT_FOUND for the vanished entry, and only the Lstat path treats not-exist as "vanished".
Apply the same vanished-entry rule to EVERY per-entry probe of the walk (Lstat, reparse/DACL/ownership probes, directory open for
recursion): a not-exist result (errors.Is(err, fs.ErrNotExist), which covers both Windows codes) means the entry is gone → skip; any other
error stays a failure. The symlink-replacement row must still refuse. `GOOS=windows go vet ./internal/pathboundary`; unix tests with a real
exit code. Set status development; update results (`git diff a970450d` non-empty); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
