# Gate note 2 — TASK-261001-2yvag1 (orchestrator; binding)

Rev3 (hosted run 36827469593) fixed the earlier six failures. Every lane is green except Race (macos-latest). There the one failure is `internal/crossconformance TestDraftSourcesBrokerAskpassDispatch`: subtests foreign_user, bare_prompt, no_arguments and extra_argument fail with `HTTPS broker secret transport: write |1: broken pipe`.

That failure is NOT caused by this task. It is a race in the askpass secret pipe on refusal paths and is tracked as a separate bug. Do NOT modify askpass, broker or crossconformance code in this task.

If your tree is otherwise final:
1. State in the results that the gate failure is this unrelated, separately tracked flake.
2. Re-hand off; the gate reruns.
