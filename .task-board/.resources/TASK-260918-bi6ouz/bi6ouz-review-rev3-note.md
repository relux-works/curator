# Review note — TASK-260918-bi6ouz revision 3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 1 was CHANGES_REQUESTED (`TASK-260918-bi6ouz_review-verdict-rev1*`): F1 `os.Lstat`→`os.Stat` in the
production call survived every test; F2 a non-ENOENT lstat failure aborted the whole §9.5 scan. Revision 2 addressed both
(`bi6ouz-rework-1.md`); its gate failed only on a macOS internal/install git flake, and revision 3 republished it.
Verify, read-only in a disposable clone:
1. F1: a row through the PRODUCTION entry (repair/use takeover path) with a symlink-to-directory at a resolved table path
   asserts no `environment_foreign_manager_suspected`; re-apply the Stat mutant yourself → killed.
2. F2: EACCES on an earlier row + a real later row → the notice names the later manager; a lone unreadable row → no notice
   and nothing recorded as absent; restoring the abort → killed.
3. Nothing else changed vs revision 1 beyond these fixes (diff the rev1 and rev3 patches).
4. Runtime validation log for revision 3 green.
accept_cr on revision 3, or changes requested with file:line. No LOGBOOK.md.
