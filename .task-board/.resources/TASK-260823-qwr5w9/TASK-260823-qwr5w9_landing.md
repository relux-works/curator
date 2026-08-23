# land-credential-scopes-composite — landing evidence

PR: https://github.com/relux-works/curator/pull/22 — MERGED, merge commit 1f55f1b4e.
Composite: 6 signed commits assembled from the accepted patch artifacts, plus two
surgical fixes the Windows CI lane earned:

- 87e9906 — vendored-audit fixture got its .cmd counterpart (the POSIX-only
  script path failed portability before the audit assertion ran on Windows)
- bd80197 — a timed-out ssh-add probe no longer reads as "agent holds nothing"
  (a killed process reports exit 1 on Windows; the deadline is now checked
  before the exit code on every platform)

CI on the merged head: fmt/lint/vet, gate self-tests, ledger, naming gate,
tests and races on macos/ubuntu/windows, interop conformance gate — all green.
Two managerlock failures across earlier rounds were in tests untouched by this
composite (pre-existing Windows contention flakes) and passed on the final round.
