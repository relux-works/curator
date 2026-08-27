# TASK-260823-1l1p8q results — GREEN CANDIDATE MATRIX

Candidate-conformance run 32651139699 (workflow_dispatch 2026-08-23T16:16Z): Candidate suite SUCCESS on ubuntu-latest, macos-latest, windows-latest, against the exact immutable rc.9 identity — revision edd07210d4f3db34fd60238cb14b90f837de03cb, manifest sha256 803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403 (verified in the job log: revision accepted immutable full 40-hex). Dispatched from main including all seven qualification fixes (PR 23, 24, 25, 26, 28, 29, 30, 31, 32).

Final blocker resolved by PR 32 (Give the Windows runner its own per-package test budget, e17b0f1): the last windows failures were per-package test-budget exhaustion, not correctness.

A confirmation re-dispatch 32653068219 was left running by the producer; its result is additive, not gating.

History of the loop: first dispatch red on 2 lanes (5 leaf tests) -> six focused fixes with produce/review cycles -> green. SPEC_PIN unchanged throughout; candidate evidence is candidate-only, not a release claim.