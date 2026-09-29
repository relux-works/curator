# Review verdict — TASK-260929-1kze6z CR rev1 — ACCEPTED

Base 7f2fb6b8, candidate tree e5318a75 (checked out with git archive into a throwaway directory). 3 paths changed: naming-gate.sh (new), gate-selftest.sh (+27), ci.yml (the inline step is replaced by `bash .github/ci/naming-gate.sh`).

## Checks (real exit codes)
- `naming-gate.sh` on the candidate tree, which still contains the 31gaka binary patch: **rc=0**. So no committed file spells either name. Both names are built at runtime with printf '%s%s'.
- Short word on a normal line of a temp .md file: **rc=1**. Full name in .md: **rc=1**.
- Full `gate-selftest.sh` on the clean candidate: **rc=0, 273 passed, 0 failed**. Rows a–e are all ok. The fail rows b, c and d each pass their assert_contains check on the naming message, so they fail for the naming reason.

## Mutants (re-run by the reviewer)
| Mutant | Effect | Result |
|---|---|---|
| M1: no skip (`elif False:`) | real tree rc=1; selftest 272/1: row (a) FAIL | killed |
| M2: no base85 shape check (`or True`) | selftest 271/2: row (c) and its reason row FAIL | killed |

## Bypass probes (temp .md, not .patch)
- "GIT binary patch" followed by the bare short word, or the short word glued into base85 characters: rc=0, so the exemption applies. **Accepted residual.** Base85-shaped lines cannot contain spaces or dots, so a real sentence cannot hide there. The most that can hide is a single token. A line with a space or dot inside the block: rc=1.
- File name: the exemption does not depend on the extension. It applies to any file with a `GIT binary patch` line. This is acceptable and arguably correct: resources can be stored under other names, and the shape bound above holds whatever the extension.
- CRLF: `\r` is stripped before the checks. A CRLF bare token is exempt (same residual). A CRLF sentence: rc=1.

## Notes (not blocking)
- A file that fails to read or a python error exits 2, so the gate fails closed. A grep exit above 1 exits 2.
- Small nit: `$hits.g` is not covered by the EXIT trap if python fails, so it can be left behind in TMPDIR. Cosmetic only.
- As the brief required, CHANGELOG, LOGBOOK and .task-board are untouched.
