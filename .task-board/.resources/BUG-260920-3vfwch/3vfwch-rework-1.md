# BUG-260920-3vfwch rework 1 (orchestrator, binding)

Verdict rev1: CHANGES REQUESTED (BUG-260920-3vfwch_review-verdict-rev1.md, RUN-260920-fa9cc7).
Continue in the Story workspace from the revision-1 tree (no checkout/clean/stash). Fix
exactly the verdict, then republish (revision 2) when the gate is green.

F1 (must) — the wired helpers retry silently. `gitFixtureWithBinary` builds
`gitFixtureConfig{binary: binary}` without `logf` (install gitfixture_test.go:83, atomicity
copy :87). Wire `logf: t.Logf` in both copies and add a committed row that drives the OUTER
entry (`gitFixture`/`env.git`) through a real retry and asserts the `retrying once` /
`succeeded on retry` lines: either accept `testing.TB` and pass a recording fake TB, or
re-exec `os.Args[0] -test.run=^X$ -test.v` (precedent commit_test.go:736) and grep the child
output. The reviewer's probe shape is the reproduction: PATH-resolvable `git` script whose
shebang interpreter is mode 0644 (kernel EACCES at exec), chmod +x after 50 ms from a
goroutine — reuse it (probe source attached to the verdict).

F2 (must) — darwin rlimits line renders garbage (`unix.Sysctl("kern.maxproc")` raw bytes) and
the premise "Darwin has no RLIMIT_NPROC" is false (`x/sys/unix` zerrors_darwin_*: RLIMIT_NPROC =
0x7; Getrlimit returns cur=5568 max=8352 on the host). Replace the darwin special case with one
`//go:build linux || darwin` file using `unix.Getrlimit` for NOFILE and NPROC (drop the sysctl
or use `SysctlUint32`), delete the darwin-only file, and pin the rendered content on unix with a
regexp like `NOFILE cur=\d+ max=\d+; NPROC cur=\d+ max=\d+` in
`TestGitFixturePersistentSpawnFailureCarriesDiagnostic` so "n/a"/garbage fails the row. Keep
the `other` GOOS file reporting unavailable; both package copies in sync (diff them).

N1–N3 (cheap, do them): add `{"EPERM", spawnErr(syscall.EPERM), false}` to the predicate table;
pin the 200 ms backoff with a literal comparison (not against the constant); fix the row count
in results.md (10 per package).

results.md: append "Revision 2" with F1/F2/N1–N3 resolution, the outer-entry retry row output
(the exact logged lines), the rlimits line as rendered on darwin, mutants re-run (A, B-full, C
backoff, E EPERM) with observed kills. Everything else (rulings R1–R4, bounds B1–B3) unchanged.
Publish revision 2 only when the configured gate is green.
