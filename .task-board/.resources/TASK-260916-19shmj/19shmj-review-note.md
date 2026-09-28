# Review note — TASK-260916-19shmj E5 managed-write nofollow (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev5 (tree 9c4311a5, hosted gate GREEN on every lane incl. Windows — the first green after 4 red revisions). Its recorded base is still
55b94af2, but the orchestrator verified the tree == trunk eca2bf27 + exactly 9 own paths (`git diff --name-only eca2bf27 9c4311a5` →
conformance-case-counts.tsv, root-artifacts.tsv, internal/envprofile/{managed.go, nofollow.go, nofollow_open_unix.go,
nofollow_open_windows.go, state_read_guard_test.go, switch.go, write_nofollow_conformance_test.go}). REVIEW THAT DELTA
(`git diff eca2bf27 9c4311a5 -- . ':!.task-board'`), not the 39-path diff vs 55b94af2.
Against `19shmj-sec-brief.md`, gatefix-1..4 and curator-spec rc.13 §8.3.1 (managed writes never follow a link) / §9.5 (cite): planted
links at managed write targets are replaced, never followed or written through; parent-component links refused; the manager's recorded
links (credential modes 0017, passthrough, drift repair, takeover) keep their specified behaviour incl. environment_credential_conflict
refusals; Windows: entry replacement of directory links/junctions without traversal (the Windows non-atomic window, if any, is stated as a
platform bound); copy fallback only when symlink creation itself fails. rc.13 write-nofollow vectors driven (11: 10 driven + 1 bound —
judge the bound); mutants (Lstat→Stat on the parent walk; rename-over replaced by open-through; fallback on rename failure) killed with
real exit codes. No CHANGELOG/LOGBOOK, no stray files. Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
