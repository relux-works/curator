# Unresolved questions

- TASK-261022-qwkhwj: Windows Git metadata is limited to 259 UTF-16 units,
  with 126 reserved for fixed SHA-256/maintenance paths and a normal clone-root
  maximum of 133. New over-budget clones refuse before creating cache parents.
  A short private metadata layout is the follow-up option; its identity,
  privacy, no-follow and cleanup contract needs a separate design.
  [Measured bound and supported layout](docs/windows-git-paths.md).

- BUG-261016-mlpvtn — self-hosted Linux CI: workflow/ref concurrency cannot
  bound competing repositories, and the go-v1 probe has no CI timeout knob.
  Host-wide scheduling authority or a separate qualification decision remains
  necessary if loaded-host deadlines recur. Three green self-hosted `main`
  runs remain a rollout condition after integration.
  The earlier candidate gate's two-minute fingerprint timeout coincided with
  a runaway Landlock probe process tree. The incident was cleared and runner
  task counts were capped; the recursion fix is tracked separately. Candidate
  qualification requires a fresh green remote gate with the existing product
  deadlines, followed by the orchestrator's post-integration rollout checks.
  [Isolation contract and probe finding](docs/self-hosted-runner-setup.md#linux-job-isolation-and-scheduling).

- TASK-261016-i5gwaf — Muse 1.4.8 exclusion replay: the exact tested build
  retains a foreign personal skill despite the required interactive boolean.
  A vendor correction or explicit normative-owner resolution is required
  before exclusion qualification and the v4 release gate can advance.
  [Evidence, options and executable witnesses](docs/qualification/muse-1.4.8-R6256.1.md).

- TASK-261021-wt8ve7: Claude Code 2.1.305 project-source exclusion preserves
  previously loaded project text on resume/fork-resume. A strict consumer
  needs proven excluded-origin session state, fresh sessions or refusal.
  Broader late hook/settings/plugin/MCP reloads also remain unqualified.
  [Measured controls and bounds](docs/qualification/project-source-exclusion-claude.md).
