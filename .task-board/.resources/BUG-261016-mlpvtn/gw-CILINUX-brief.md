# Producer brief: BUG-261016-mlpvtn, self-hosted Linux CI isolation

Goal: Linux CI must run reliably on the repository's self-hosted Linux runners
again. The owner decided the hosted Linux image is not an option. Fix the
workflow side; do not change the runner host.

Read the bug description and notes first. Failure classes to remove:

1. Shared Go build cache across concurrent runners and several Go toolchains
   (1.25.5 / 1.26.0 / 1.27.1). Witness: link failure "relocation target
   golang.org/x/sys/unix.Syscall not defined for ABI0 (but is defined for
   ABIInternal)".
2. Probe deadlines under host load. Witness: "go-v1 process_timeout: Go probe
   exceeded its deadline" in TestReviewProbeLegacyLaneIdentityInvalid.

Required changes in `.github/workflows/ci.yml`:

- Every job that can run on the self-hosted Linux runners (go-family-qualification,
  test, lint, gate-selftest, interop, naming-gate, and any other job routed by the
  LINUX_SELF_HOSTED expression) gets a job-private Go state before any Go command,
  including before or as part of actions/setup-go caching: GOCACHE, GOMODCACHE and
  GOPATH under `$RUNNER_TEMP` (note: the `runner` context is not available in
  job-level `env`; use a first step that writes to `$GITHUB_ENV`, or step-level
  env, and make sure setup-go's cache restore/save does not reintroduce a shared
  per-user cache — disable setup-go caching on self-hosted Linux, or key it by Go
  version and point it at the job-private paths).
- Keep GOTOOLCHAIN=local and GOENV=off as they are.
- Add a workflow-level `concurrency` group keyed by workflow and ref, with
  cancel-in-progress for pull_request events only (never cancel main or gate/**
  pushes).
- Do not change the hosted macOS/Windows routing, the fork guard, the Race job
  routing, or the SPEC_PIN/PNPM_PIN contracts.
- For the probe deadline: find where the go-v1 probe deadline is set. If it is a
  fixed short wall-clock deadline, make it tolerant of a loaded host in CI only
  through an existing, documented knob if one exists; otherwise record the
  finding in your handoff and do not invent a new product flag.

Docs: update `docs/self-hosted-runner-setup.md` (the isolation contract and why).
No host names, IPs, personal names or local paths in any committed text.

Negative evidence:
- A fixture or script check (existing gate-selftest style is fine) that fails if
  a self-hosted-routable job runs a Go command without job-private GOCACHE and
  GOMODCACHE. Show it fails on the current main workflow (control plant) and
  passes on yours.
- Show that fork pull requests still route to the hosted image (existing guard
  check stays green).

Validation: run the repository's normal gate (`sh scripts/remote-gate.sh`) and
the workflow lint the repo uses. The curator lane flips LINUX_SELF_HOSTED back
to true for the remote gate run; report the run id and per-job runner names.

Handoff: Change Request with summary, files changed, the control-plant result,
and any probe-deadline finding.

## Revision after the blocked handoff (2026-10-17)

Your blocker is resolved; continue from the preserved changes in the Story workspace.

- Root cause of the failed remote gate: between about 01:40Z and 02:35Z the self-hosted Linux host was saturated by a runaway process tree from a curator test-binary recursion in the Landlock probe (load average above 200, thousands of orphaned probe processes). The `toolchain_timeout` in `TestRealTrustedGoProbe` (Go 1.26 Linux job) happened inside that window. The tree was cleared, the curator runners now have a task-count cap, and the recursion fix (BUG-261017-ygfdal) is in review. The host is idle now.
- Do not change product deadlines or add a timeout knob. Rerun the full remote gate on the current snapshot; if it is green, hand off. If a Linux job fails again, record the job, runner registration name, the time window and the host load you can observe from the job log, and hand off as blocked with that evidence.
- The three consecutive green main runs are a post-integration obligation of the orchestrator, not a handoff prerequisite.

## Revision 2 after the second blocked handoff (2026-10-17)

- The 05:00-05:18Z failure (`go-v1 process_timeout` in `TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only`) fell inside a host-load spike with the same signature as before: load average about 240 with about 43 percent system time. Your Story was based on a main without the Landlock probe recursion fix, so your own gate's Linux test job ran the recursing probe again; the task cap kept the host up, but not the deadline.
- The fix is now on main (BUG-261017-ygfdal, 6e85a91) and your Story workspace has been converged onto it with your changes carried. Rerun the full remote gate on a fresh snapshot of the converged workspace. If it is green, hand off. If a Linux job fails again, collect the same evidence as before and hand off as blocked.
- Keep the job-private Go state. If the evidence shows a cold-cache deadline problem rather than host load, say so explicitly; do not change product deadlines.
