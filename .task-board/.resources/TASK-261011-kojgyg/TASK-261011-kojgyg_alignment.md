# TASK-261011-kojgyg — R-LR1-D21 contract alignment

Docs-only developer handoff for review. The sole repository change is
`decisions/0021-sessions-enter-through-curator-run.md`, left uncommitted.

Sources read: [contract table at 177cb554](https://github.com/relux-works/curator/blob/177cb55417fc1d0c1630842c07aca714ebb3db9e/.research/261010_platform-contract-table.md),
[audit at 48789217](https://github.com/relux-works/curator/blob/48789217c815ebc45edb7d1d0c282879c83450bf/.research/261010_platform-docs-audit.md), and the supplied
`owner-decisions-20261010-audit.md` precondition. D-EXECUTOR (confirmed
2026-10-11), D-PATH and D-WIRE-NAMES override their corresponding table
questions. No owner decision is inferred from a recommendation.

| R-LR1 audit row | Decision 0021 section changed | Canonical value / recorded owner decision / explicit open question |
|---|---|---|
| A09.1 | Amendment v1: Authority and scope; Launch, transport and evidence owners; original Decision §6/§8 | N6/P1 + D-EXECUTOR: caller composes, dispatcher admits/provisions/tracks, root `swarma-user-launcher` verifies the approved executor, agent-account `curator-run` verifies and executes. Human interactive composition/direct or ax execution remains unchanged. No agent-account profile resolution or composition; one shared launch library with `swarma-session-runner`; executor network only to broker. |
| A09.6 | Amendment v1: Launch, transport and evidence owners; Activation gate and remaining owner questions | L1/L2/P2/I2/I3/P6: immutable plan on FD3, no inherited broker connection; decoded 48 KiB / base64 64 KiB plan; privileged launch request 80 KiB/5 s; dispatcher v0 96 KiB/5 s, v1 128 KiB; broker 64 KiB; smallest applicable bound before provisioning. Root start receipt precedes privilege drop and is not ready proof; existing receipt refuses replay. Separate account generation/execution/board/correlation IDs. D-PATH binds the resolved protected-tools → admitted base → protected manager PATH and digest. Launcher spec owns publication of closed input/FD carrier, credential intake, signals/exit, outcome receipts and typed refusals; qualification is required before activation. W3 and I4 retain explicit owner questions and options; no signed grammar, receipt fields or new error codes invented. |
| A09.10 | Status; Context; original Decision §6/§8; Amendment v1; Compatibility and security impact | P1 + D-EXECUTOR: versioned Decision amendment v1 permits only the local headless agent-account one-shot executor. Dedicated intake example `curator-run exec --plan-fd N` is distinct from the human command; exact CLI/plan version remains launcher-owned. The mode stays disabled pending published intake and qualification. The historical blanket prohibition is qualified consistently; no general human headless command or runtime capability is claimed. |
| A10.9 | Amendment v1: Authority and scope; credential and provider boundary rows; activation gate | C9 + D-EXECUTOR Q2: first broker consumer is `curator-run` under the agent account; board runner/session host embed no client and ask the dispatcher. C3/C9 slice 0: cross-user peer/binding proof plus one protected Claude setup-token launch via `CLAUDE_CODE_OAUTH_TOKEN`, no credential file/native fallback/refresh. Later Codex lane is not first-slice support; C4's adoption/qualification question and its two options are explicit. |

Coverage: **4/4 R-LR1 A-rows mapped**. This ratio measures the documentary
mapping, not runtime qualification. N6/P1/C9 owner questions are resolved
by D-EXECUTOR. W3 (component-specific format or shared versioned adapter),
I4 (joint ticket/host-role versioning and activation or SH2 disabled), and
C4 (single auth owner with named supported-release evidence or defer)
remain explicitly open in the decision. D-WIRE-NAMES settles repository-based
wire naming only; no unpublished signed bytes are inferred.

The shared-library rail explicitly amends the launcher's old
single-module/no-shared-library boundary. Its spec and implementation are
separate consuming work; this change does not silently claim that edge is
already available.

Diagram accounting: Decision 0021 contains no diagrams. R-LR1's existing
launch figures belong to other repository owners (F-L); launcher M-LR1/M-LR2
are proposed P1 diagrams. No diagram under this task's sole permitted file
scope needed correction, and no new diagram was created or rendered.

| Validation run directly | Real exit code | Evidence and limit |
|---|---:|---|
| `git diff --check` after the last document edit | 0 | Tracked-diff whitespace lint. |
| Task-scoped standalone Python document lint after the last document edit | 0 | 4/4 audit row references, 4/4 relative file links, 8/8 amendment boundary rows; consistent Markdown table columns, no trailing whitespace/NUL or selected local-home/session-link forms; exactly 1/1 allowed tracked file. Syntax/packaging only, not semantic or runtime proof; external links not remotely checked. |

Semantic alignment was reviewed manually against the pinned table and
supplied decisions. No tests, builds, provider login, renderer, deployment,
commit or other source repository mutation ran. Tests/builds are explicitly
forbidden by the brief and not claimed as passing. No earlier gate evidence
is accepted as proof of this change. Findings are also persisted in the
separate task-scoped logbook outcome. The generic code/tests/build checklist
was replaced with documentation authoring and validation entries.
