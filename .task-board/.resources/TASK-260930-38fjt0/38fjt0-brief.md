# TASK-260930-38fjt0 — pin the Test (rose-air) lane to the rose-air runner (THE ONLY CURRENT INSTRUCTION)

`.github/workflows/ci.yml` job `test-self-hosted` (name "Test (rose-air)") has `runs-on: [self-hosted, macOS, ARM64]`. Any org
self-hosted Apple Silicon runner can take it, and run 36626413154 landed on the runner "macbook-iv". An org admin confirmed that the
intended runner carries the label `rose-air`, and that macbook-iv carries its own label `macbook-iv` and not `rose-air`.
1. Change it to `runs-on: [self-hosted, macOS, ARM64, rose-air]`. Update the job's comment block to say the lane is pinned to the
   rose-air runner by label.
2. Add a gate-selftest.sh row, in the existing style, asserting that the test-self-hosted job's runs-on includes `rose-air`. Mutant:
   remove the label → the row fails. Give real exit codes.
3. If docs/self-hosted-runner-setup.md describes runner labels, add the `rose-air` label requirement there.

No CHANGELOG/LOGBOOK edit. Never spell any employer name. Update the results, then run `task-board handoff TASK-260930-38fjt0 --role developer`, then
END YOUR TURN.
