# THE ONLY CURRENT INSTRUCTION — TASK-261016-w0yx2w: route the Linux race job, then rerun the full gate (developer)

Your gate retry found that the GitHub-hosted Linux image has a default ACL above the workspace. The snapshot destination admission refuses that ancestry by design, so the valid controls in the hosted Linux `race` job fail. The regular Linux `test` job runs on the self-hosted Linux runners and passed.

CI-owner decision (orchestrator): the `race` job follows the same Linux routing as the `test` job. This routing change is part of this candidate, because this candidate introduces the ancestry requirement that the hosted Linux image does not meet. Do not broaden the ACL admission policy to fit an image.

## Do this

1. In `.github/workflows/ci.yml`, give the `race` job the same `runs-on` expression that `test` and `go-family-qualification` use:
   - `ubuntu-latest` goes to `["self-hosted","relux-linux","curator"]` when `vars.LINUX_SELF_HOSTED == 'true'`, for pushes, dispatches and same-repository pull requests;
   - a fork pull request stays on the hosted image.

   Update the job's comment block to say why.
2. Add one short paragraph to `docs/self-hosted-runner-setup.md`. It must say:
   - snapshot publication refuses a destination whose retained ancestry carries an access or default ACL;
   - the qualified self-hosted Linux runners meet that requirement;
   - the hosted Linux image does not, so a fork pull request's hosted Linux snapshot controls are expected to refuse. This is a known limit, not a pass.
3. If the repository has a CI-routing test or ciguard that enumerates job routing, extend it so that it covers the `race` job. If no such test exists, say so in the outcome.
4. Re-sign the candidate and rerun the full hosted gate on the new exact candidate. Then inspect every job.
5. If the gate is green, complete the checklist rows with the run URL and job results. Update the round-2 outcome with a "Round 3" section that names the routing change and the run. Then hand off with `task-board handoff TASK-261016-w0yx2w --role developer`.

A runner infrastructure failure is a blocker to report, never a pass.

End your turn after the handoff, or after recording a blocker.
