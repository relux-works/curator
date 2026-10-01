# TASK-261001-1klixs — carry-second-operator-guide

Restored TASK-260930-o5uu7a rev1 unchanged, accepted base 5ed5c4e1 / tree ca2d4a1a. Left changes uncommitted in the assigned Story worktree.

Steps and directly observed exit codes:
- task-board m set_status(..., development): 0.
- git diff 5ed5c4e1 ca2d4a1a -- docs/second-operator.md README.md > "$TMPDIR/so.patch" && git apply --3way "$TMPDIR/so.patch": 0; both commands succeeded, README applied cleanly, guide used direct application.
- git diff origin/main --stat: 0; README.md 2 additions, docs/second-operator.md 154 additions; exactly 2 files / 156 insertions / 0 deletions.
- Python 3 assertions using git diff --name-only, git show ca2d4a1a:<path>, Path.read_bytes(), and ordered +/- line comparison against git diff 5ed5c4e1 ca2d4a1a: 0. Both files byte-identical to accepted tree; all 156 added lines equal the accepted diff. README relative guide link resolves.
- git diff origin/main --check: 0 (documentation whitespace validation).
- go build -o "$TMPDIR/TASK-261001-1klixs-curator" ./cmd/curator: 0. Binary written outside the repository.
- Final git diff --numstat origin/main / --name-only and git status --short: 0; only README.md and docs/second-operator.md changed.

Scope and verification bounds:
- No runtime behavior changed. No test/code/config paths added or modified, as expressly required by the carrier brief. The carrier assertions above are the relevant tests run in this turn.
- Full Go/race suites and golangci-lint were not run: this task carries accepted documentation verbatim; whitespace validation and compilation were run directly after the patch.
- The accepted review's clean-HOME literal walkthrough is prior evidence; I did not rerun it or the installer, real agent sessions, Linux, or keyring branch. No new claims about those capabilities.
- Generic checklist tests item is not applicable to new behavior, because there is no new behavior and additional paths are forbidden; documentation assertions passed. Generic logbook item is not applicable: no new anomaly or decision; carrier brief explicitly prohibits CHANGELOG/LOGBOOK edits.
- No commits, branch operations, CHANGELOG/LOGBOOK edits, or extra repository paths changed.


## Revision 2 (refresh)

Base refreshed from bab2433ba115a7eafb2298dba6ef16154e2cc62e to bd126a9acdc51b6061917ba8c4d7d26a7abafd41. The guide remains unchanged. README preserves both the accepted second-operator link and trunk's External build repositories link. Candidate remains uncommitted.

Identity proof and per-file statistics:

| Comparison | README.md | docs/second-operator.md |
| --- | --- | --- |
| Refreshed candidate vs new base / origin/main | +2 / -0 | +154 / -0 |
| Carrier rev1 (891d805d) vs its base (bab2433b) | +2 / -0 | +154 / -0 |
| Refreshed candidate vs carrier rev1, on carrier paths | +1 / -1 | unchanged (0 / 0) |

The last row is exactly the incoming credential documentation sentence in README. The guide is byte-identical to both 891d805d and ca2d4a1a. Ordered per-path +/- lines relative to each candidate's own base are identical. All other files inherit trunk unchanged; whole-tree differences from rev1 include upstream work and must not be confused with the two-path carrier delta. `git diff origin/main --stat` lists exactly README.md and docs/second-operator.md, 156 additions and no deletions. `git status --short` agrees; no extra paths changed.

Directly observed commands and real exit codes in this run:

- Initial `task-board m 'set_status(TASK-261001-1klixs, status=development)'`: 0.
- Initial `task-board worktree status STORY-261001-38ijl9`: 0; preserved stale rev1 on bab2433b.
- First `task-board worktree refresh-candidate TASK-261001-1klixs`: 1, refusal: `candidate recovery refuses staged state outside the checkpoint baseline`.
- `git restore --staged -- README.md docs/second-operator.md`: 0; file bytes preserved.
- Retried `task-board worktree refresh-candidate TASK-261001-1klixs`: 0, refresh_advanced to bd126a9a; no replay conflict and no manual checkpoint commit.
- Python README combination: 0, replacing only the accepted credential sentence with trunk's sentence while preserving the guide link.
- First run of temporary refresh regression checks: 1, correctly detecting missing docs/external-build-repositories.md before incoming tree content was replayed. This was a failing check, not passing evidence.
- First incoming patch generation (`git diff bab2433b bd126a9a --binary -- . ':!README.md'`): 0; application: 1 because resource payloads were already materialized by managed refresh. No partial patch applied.
- Incoming patch generation excluding README and already materialized `.task-board/.resources`: 0; `git apply --whitespace=nowarn /tmp/TASK-261001-1klixs_incoming-trunk.patch`: 0. This replays upstream bytes exactly; it does not fix or rewrite whitespace. Only checkout artifacts were replayed; authoritative board writes use task-board CLI.
- Updated temporary regression suite (six tests): `python3 /tmp/TASK-261001-1klixs_refresh_test.py > /tmp/TASK-261001-1klixs_refresh-validation.log 2>&1`: 0. Named checks cover guide identity, exact README combination, exact two-path scope, unchanged carrier +/- delta, all relative links, and a scope narrowing mutant.
- Relative links: README 20/20 and guide 1/1 resolve, including fragments (21/21 total); checked as part of the suite above, exit 0. External URLs and inline code examples are outside this local link-check bound.
- `git diff origin/main --check` after the combined candidate: 0 (documentation whitespace validation).
- `go build -o /tmp/TASK-261001-1klixs-curator ./cmd/curator` after the incoming trunk content was restored: 0. An earlier build also exited 0 but is not used as combined-tree build evidence.
- `git diff origin/main --stat`, `git diff origin/main --numstat`, `git diff 891d805d --numstat -- README.md docs/second-operator.md`, and `git status --short`: 0; stats recorded above.

The named `test_refresh_preserves_both_readme_links` protects this refresh. `test_readme_scope_narrowing_mutant` injects a missing README target in memory: narrowing the checker to the guide admits it, while full README-plus-guide checking rejects it. No product gates or behavior were added. The legacy review-round template refers to a rejection, but the actual attached TASK-261001-1klixs_review-verdict-rev1.md says ACCEPTED; integration_base_moved is the refresh trigger.

Verification bounds: this run executes the documentation checks and build itself. It accepts the prior literal clean-HOME walkthrough as prior evidence and does not claim a new walkthrough. Full Go/race suites, golangci-lint, installer downloads/attestation, real agents, Linux, and keyring login were not rerun because this instruction permits only the accepted documentation and base combination. No CHANGELOG/LOGBOOK or runtime/test/config delta is present. The recovery anomalies are recorded here and on the board; the task explicitly forbids repository logbook edits.

Attached refreshed evidence: TASK-261001-1klixs_refresh_test.py and TASK-261001-1klixs_refresh-validation.log. Developer handoff is the next lifecycle command; its actual exit code is emitted by the CLI and recorded in the run transcript. No acceptance or landing is claimed.
