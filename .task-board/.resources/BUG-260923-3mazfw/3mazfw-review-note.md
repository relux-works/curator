# Review note — BUG-260923-3mazfw (macOS git fork/exec EACCES), CR revision 2 (orchestrator, binding)

Revision 1 failed its gate only on the Windows snapshot flake (fixed separately by BUG-260923-11jgkt); revision 2 is
the same fix republished. Review it on content, read-only (disposable clone if you run anything):
1. Root cause evidenced (hosted artefacts cited in results), fix addresses it: retry ONLY `*os.PathError` op
   `fork/exec` errno EACCES, once, after 100 ms; git exit statuses, EPERM, ENOENT and EACCES from other ops are NOT
   retried; a persistent second failure is surfaced (fails closed) — check rows exist for each and execute.
2. No blanket retry, no widened timeout, no skip. Is the retry at the right seam (production `internal/gitignore`
   call site) and not hiding a real permission problem on a user machine (one retry, then a clear error)?
3. One narrowing mutant of your own in a disposable copy (retry any error, or retry forever) → which row kills it.
4. Hosted gate green on revision 2 (runtime validation log).
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md.
