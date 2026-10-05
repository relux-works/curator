# TASK-261004-hy8zmn: audit-token-argv-and-backend-env-allowlist-design

## Description
DESIGN PENDING — NOT ACCEPTED FOR EXECUTION. Operator decision 2026-10-04: decide the design first, then prioritise. Do not spawn producers and do not schedule until the operator accepts a design. K24 class. (1) `curator audit --publish --token <value>` takes the secret in argv (visible in ps and shell history, cmd/curator/main.go:2120): refuse --token; take an env var or --token-file (regular file, not group/other readable, opened no-follow, 64 KiB cap). (2) Implement the audit backends of profiles/manager.md §7 (audit.backend/audit.backends are parsed but unused) for several agent environments, with environment protection from day one: one allowlist for every backend child (PATH, HOME, USER, LOGNAME, LANG, LC_*, TMPDIR/TEMP/TMP, SYSTEMROOT/COMSPEC on Windows, plus documented per-backend variables); *_TOKEN, *_KEY, SSH_AUTH_SOCK, GIT_* always stripped, case-insensitive on Windows; real child-process observation tests and narrowing mutants. Git transport children keep their auth env (separate hygiene item). Reference implementation: public cocoaskills PR #142 (commit 0da153a).

## Scope
(define task scope)

## Acceptance Criteria
Research document (CIP draft per cip-template.md where the brief says so) under .research/ with options, tradeoffs, recommendation, evidence (file:line or measured probes) and decision-ready open questions; no product code changes; no secrets read or printed; LOGBOOK.md untouched.
