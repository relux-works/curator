# Review note for TASK-260916-1l44nd revision 6 (orchestrator, binding)

Revision 6 = rework 5 for your revision-5 verdict (TASK-260916-1l44nd_review-verdict-rev5.md;
brief 1l44nd-rework-5.md): F1 Landlock access-right constants taken from `golang.org/x/sys/unix`
(TRUNCATE = 1<<14, MAKE_SYM = 1<<12) with correct object typing (TRUNCATE valid on regular files;
MAKE_*/REMOVE_*/REFER directory-only; `/dev/null` file-typed READ|WRITE rule); F2 the handled mask
covers every filesystem right the probed ABI supports (REMOVE_FILE/DIR, MAKE_DIR/REG/FIFO/SOCK/
CHAR/BLOCK/SYM, REFER on ABI≥2, TRUNCATE on ABI≥3, IOCTL_DEV on ABI≥5) with directory-only rights
granted only beneath the derived writable directories; rows on ubuntu Test+Race through the real
worker: truncate/O_TRUNC outside denied with content unchanged, granted file truncation allowed;
unlink/rmdir/mkdir/mkfifo/same-dir rename outside denied with state preserved, inside a granted
directory allowed; narrowing mutants per omitted right; docs/script-interpreters.md and results.md
"Revision 6" with the ABI→mask table. Gate green on all lanes: run 35690194789 — verify the gate
commit resolves to the exact revision-6 tree and that rev5→rev6 is exactly this scope.

Rerun your three contract tests against rev6 (they must pass now); extract the new Linux rows from
`test-evidence-ubuntu-latest` and `race-evidence-ubuntu-latest` (names, PASS, timings); compare
the mask builder output per ABI against the `unix` constants; confirm no writable root was
broadened. Everything else was verified at rev5 — do not re-open it unless rev6 changed those
bytes. Record exactly one verdict: accept_cr(TASK-260916-1l44nd, revision=6, evidence=<your
outcome resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write
into the control root's LOGBOOK.md.
