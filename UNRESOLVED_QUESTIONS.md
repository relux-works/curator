# Unresolved questions

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

- TASK-261022-h1ttey: decide the first Claude strict form (proposed fresh
  headless only), required tools/child paths, response to unqualifiable late
  reloads, and whether restricted-form advertisement needs normative carrier
  negotiation. Installed 2.1.309 has no transferred qualification from 2.1.305;
  permissions, memory-off and all reachable exclusion paths must qualify.
  [Contract, negative-first matrix and owner decisions](docs/qualification/strict-mode-claude-plan.md#owner-decisions-to-resolve-before-implementationadmission).
