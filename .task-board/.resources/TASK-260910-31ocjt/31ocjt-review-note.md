# Review note — TASK-260910-31ocjt HTTPS askpass secret off the environment (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev4 (base 3f60f7f0, tree 9874753f, 11 paths, gate green) against `31ocjt-brief.md` and `31ocjt-gatefix-2.md`. Verify:
1. The askpass secret reaches the fetch child only through a pipe / inherited fd (Windows: inherited handle or named pipe with a private
   ACL); it never appears in argv, env or on disk. CURATOR_BUILD_HTTPS_ASKPASS_SECRET is gone everywhere (grep the whole tree incl. docs).
2. The fd/handle is closed in the parent after start and is not inherited by grandchildren (CLOEXEC / non-inheritable after the helper reads);
   a helper that reads twice or a missing fd fails closed, not with an empty password.
3. Rows at the production entry: child receives the secret; a grandchild spawned by the askpass helper does not see it in env; the env var is
   absent. The e2e test goes through the shared instrumented process seam (no guard exemption added).
4. Mutants (put the secret back in env; leave the fd inheritable) — run them yourself with real exit codes and confirm each is killed.
5. New state reads via internal/stateread, managed writes via the E5 nofollow helpers; no CHANGELOG/LOGBOOK; no stray files.
accept_cr or changes requested with file:line. No LOGBOOK.md.
