# Round 3 handoff recovery

The predecessor implemented round 3 and attached the full validation evidence, but Change Request revision 2 validation exited 1 before hosted execution: the required signing environment was absent and SSH signing requested a passphrase. That failed validation is not a green gate.

Recovery introduced no repository changes. Sourcing the host's prescribed agent-git-env restored the configured agent socket and signing key. A detached exact-tree signing probe verified G; the Story branch was neither committed nor moved.

| Check rerun by successor | Exit | Evidence |
|---|---:|---|
| Four takeover regression tests via production run(), including all 67 unrelated registered command cases | 0 | resume/regression.log |
| list-only-consent-bypass narrowing mutant | 0 parent, 1 child | TestNativeSkillsTakeoverEveryRegisteredCommand/list and TestNativeSkillsTakeoverArgumentDiscardRegression/list fail as intended |
| go build -o task-local artifact ./cmd/curator | 0 | resume/build.log |
| git diff --check | 0 | No whitespace diagnostics |
| Fresh origin/main fetch and changed-path intersection | 0 | resume/freshness.json; no overlapping paths, root README unchanged |
| Detached signing probe | 0 | Signature G; exact candidate tree |
| Provider read of prior hosted gate | 0 | Run 39117060646, attempt 2, completed/success on 3e00fcc8abe75e411a88773c65c4042ae2448eb0 |

Full local coverage is reused from the already attached round-3 evidence on the identical source, test, configuration and environment identity: 85 packages, 3494 top-level names, explicit unchanged-evidence bounds in validation-r3 and its full logs. It was not replayed wholesale by this successor. Accepted rows 1-11 remain unchanged. Coverage remains 12 of 12 surface rows and 5 of 5 AC rows, with help-control limitations explicitly preserved in the coverage map.

| Mutant | Narrowed gate | Named failing tests | Survivor bound |
|---|---|---|---|
| list-only-consent-bypass | Only list silently accepts takeover; all other command rejections remain | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Killed by both named failures; no survivor |

The successor invoked handoff from a shell with the prescribed signing environment. Handoff returned exit 0, status to-review, checklist 23/23. No fresh hosted workflow was observed during that command. The runtime publishes the Change Request after the producer turn ends, so its own validation result is not yet established by this note. The runtime must also inherit the prescribed agent signing/transport environment; shell-local sourcing proves signing readiness only in this successor shell. The earlier signing failure is retained. All changes remain uncommitted at the recorded Story checkpoint. No real HOME was used as a test write target. No control-root files were edited directly.


## Handoff audit warning

The successful role handoff reported run_wrote_outside_worktree under warn policy. Its unattributed rows name unrelated tasks BUG-261018-b2jnyg, BUG-261018-dbeghc and STORY-261018-m1ak35, with samples 2026-10-18T11:23:42.155598Z through 2026-10-18T11:26:57.971668Z; two other rows are explicitly attributed to other runs. This successor made no direct control-root writes: all task evidence and state writes used task-board CLI. The six rows attributed to this run's own chain are task/parent activity, progress and outcome resources. The warning is preserved, not cleared or treated as proof that the unrelated writes belong to this producer.
