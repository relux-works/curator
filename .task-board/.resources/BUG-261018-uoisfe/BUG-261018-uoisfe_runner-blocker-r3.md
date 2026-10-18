# Hosted runner cleanup failure — round 3

Current signed candidate: 3e00fcc8abe75e411a88773c65c4042ae2448eb0 (G).
Hosted run: https://github.com/relux-works/curator/actions/runs/39117060646

Verified cause: Lint job 117644882722 and Go 1.26.0 driver (Ubuntu) job
117644882976 passed all repository steps, then failed Complete runner with
exit 1. Provider logs show Permission denied while removing pre-existing
read-only pnpm fixture-store directories in a shared runner temporary root.
Raw logs are preserved privately; sanitized versions will accompany validation.
This is hosted failure, not a pass.

The preceding superseded candidate also encountered the same cleanup class.
A job retry while the workflow is active was refused (CLI exit 1, job cannot
be rerun), including a direct retry of the current Lint job.

Current recovery: let the remaining Windows Test report, then use the hosting
platform failed-job retry on the exact same signed head. No source, workflow,
required check, runner variable, permission or shared process was changed.

If the provider retry repeats this verified failure, clean options are runner
owner repair of temporary-root cleanup, or rerun on a healthy supported runner.
An unsupported workflow rewrite, removal of required checks or a local pass
would not satisfy the hosted-gate requirement. Local validation is green, but
no local mirror is presented as a replacement for required hosted checks.

Recovery: **resolved by legitimate provider retry on the same signed head**.
Attempt 1 remained failure (watch exit 1). The accepted --failed retry (CLI exit 0)
queued behind four online/busy Linux runners, then both jobs passed including
cleanup. Final gate watch exit 0; 20 successful jobs and two conditional skips.
No source, runner state, protection, check or provider status was modified.
The historical failure remains visible in the attempt-1 provider record and
sanitized logs; no local mirror was used as a required-check substitute.
Status: hosted gate green; runner-owner intervention was not needed.
