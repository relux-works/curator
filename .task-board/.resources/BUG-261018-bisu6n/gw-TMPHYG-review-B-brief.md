# BUG-261018-bisu6n: review B, read-only tmp trees and the Linux runner self-heal (deciding reviewer)

Review the published Change Request revision against `gw-TMPHYG-brief.md`. You are the deciding reviewer: accept only when every point holds on the exact revision.

Sweep and record in the verdict table with evidence (file and line, test name, command and exit code):
1. the pnpm conformance test registers the write-permission restore before creating the tree and restores on pass and failure paths;
2. sweep complete: grep yourself for read-only tree creators (pnpm/npm stores, module caches, explicit chmod to read-only modes, archive extraction) and compare with the producer's list;
3. every job in `.github/workflows/ci.yml` that can run on the self-hosted Linux runners ends with the guarded `if: always()` step; none is missed (walk every `runs-on` expression);
4. the self-heal script changes permissions only, never deletes or moves, stays inside the root, refuses unset/empty/`/`/non-directory roots without changes, and never fails the job;
5. regression witness, script tests and workflow structure tests pass; the plant control fails for the stated reason;
6. no product behaviour changed; no test touches the real HOME; `CHANGELOG.md` and root `README.md` unchanged;
7. full local suite and hosted gate pass on the exact candidate, with the runners each Linux job used. A failure only in "Complete runner" is a blocker, never a pass.

Before you start, compare the CR base with fresh protected trunk. A path overlap means the revision is stale: say so, and do not accept a stale revision.

Record the verdict with `accept_cr` or `request_changes`, attach `BUG-261018-bisu6n_review-verdict-rev<N>.md`, then END YOUR TURN.
