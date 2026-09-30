# TASK-260930-38fjt0 review verdict rev1 — ACCEPTED
Base 0e3169bb, candidate tree 8311603e (worktree identical, `git diff 8311603e` empty), 3 paths only.
- ci.yml:241 `runs-on: [self-hosted, macOS, ARM64, rose-air]` exact. Only self-hosted job in ci.yml; all others use ubuntu-latest or matrix.os over GitHub-hosted images (lines 93/364/500/557).
- Comment ci.yml:231-233 accurate; docs/self-hosted-runner-setup.md:3-9 states rose-air label requirement.
- gate-selftest.sh on candidate: exit=0; rows 170/171 ok.
- Mutant (ci.yml label removed via sed): gate-selftest exit=1; exactly 2 FAIL rows: "test-self-hosted runs-on is pinned to the rose-air label" and in-script mutant row "could not produce the unpinned mutant". Killed. Worktree restored.
- Row parser scopes to the test-self-hosted job and requires exactly one runs-on line; exact-token match on `rose-air`.
No findings.
