# TASK-260910-1952mz — producer logbook rev6 (2026-09-17)

Repair of the CR rev5 hosted-gate failure (run 35201254365, Windows lane
only). Full transcripts in `TASK-260910-1952mz_results.md` (rev6 appendix).

## Finding: the rev5 identity repair works; the new assertion was wrong

The Windows CI stderr/stdout pair for the failing cross-spelling subtests
shows the hook behaving exactly per spec (`shell_hook_env_changed`,
native path, one warning, new bytes sourced under A). The approved-phase
activations (silent sourcing under A and B through MSYS spelling) passed
on Windows. Only the assertion's hardcoded `sourced1=1` marker was wrong
for the `=2` rotated bytes. No production change in rev6.

## Ops: task-board CLI hangs without a closed stdin

Two `task-board` invocations (one `resource update`, one trivial `m`
query) sat at 0 CPU for 6–9 minutes in this session's managed shell;
both were terminated and the retry with `</dev/null` completed
instantly. Until this is understood, run every board command with stdin
closed.

## Ops: warm-binary validation under host contention

With the shared host heavily contended, a plain `go test
./internal/shell/...` wrapper hit the 600s package timeout with zero
assertion failures, and the ~2s hookapproval/envfiles suites took
~162s. Same mitigation as rev3: `go test -c` once, then the warm
same-inode binary with `-test.run` masks — every chunk green. The
change itself is three string-comparison lines and cannot affect timing.
