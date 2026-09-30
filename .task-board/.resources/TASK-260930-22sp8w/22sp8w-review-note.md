# Review note — TASK-260930-22sp8w tests must not inherit ambient git config (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base bdb77413, tree 6fc60498, 35 paths, gate green on every lane; rev1 failed only on Windows backslash paths in the
hostile config) against `gitiso-brief.md`. Verify:
1. test-gate.sh exports GIT_CONFIG_NOSYSTEM=1 and GIT_CONFIG_GLOBAL, pointing to an empty file it creates, for every go test stage, and
   logs them. The gate-selftest row asserts this.
2. The shared TestMain helper is called in every package whose tests shell out to git. List the packages; grep for exec of "git" in
   *_test.go and compare with the list. Tests that set GIT_CONFIG_GLOBAL on purpose, such as insteadOf rewrites, still work
   (t.Setenv overrides).
3. The hostile-config regression row (commit/tag signing on, ssh format, a throwaway signing key) passes on every OS, and the Windows
   path is written in a form git accepts (forward slashes, or escaped). Re-run the helper-removed mutant yourself and confirm the row
   fails (real exit code).
4. Look through the 35 paths: no production code changes beyond test support, no weakened assertions, no CHANGELOG/LOGBOOK, no stray
   files.
accept_cr, or changes requested with file:line. Never spell any employer name.
