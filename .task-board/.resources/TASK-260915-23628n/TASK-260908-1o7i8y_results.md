# TASK-260908-1o7i8y — production integration blocker

Date: 2026-09-15. Producer run: RUN-260915-bc7df9.
Disposition: blocked before product-code edits; no review-ready candidate.

## Constraint and source evidence

SPEC.md §4.5 (lines 532–537) requires owned environment literals from
`System.ChildEnv(nil, req)` for the **same** system and `LaunchRequest` used
to build the admitted plan. It explicitly forbids environment diffing or a
second `BuildPlan` with an empty parent. `internal/composition/composition.go`
implements this contract by requiring that request from its caller.

The pinned agents-management **v0.5.11** exposes
`vendorplugin.BuildLaunch(ctx, registry, SpawnRequest, mode) (agentic.Plan, error)`
(`pkg/vendorplugin/spawn.go:140`). The admitted `LaunchRequest` is local to that
function: vendor translation at lines 181–204, authoritative Runtime/Vendor
assignment at 211–215, preparation at 219, and `BuildPlan` at 254. It is never
returned. `agentic.Plan` (`pkg/agentic/plan.go:33–58`) retains neither the request
nor an owned-environment snapshot. The internal launcher `plan.Build` likewise
returns only `agentic.Plan`.

Additionally, `agentic.BuildPlan` prepares the request at line 204 and projects
model aliases to their launch identity at lines 263–265 before `ChildEnv` at
280. Recreating a request from the original flags/SpawnRequest is not proof of
the same effective request. Calling Vendor.Spawn again would duplicate only
part of the private admission/preparation path. A system wrapper to intercept
ChildEnv would couple the launcher to internal callback order and optional
system capabilities. Neither is the specified ordinary API composition.

This is an upstream API/ownership boundary outside the assigned launcher
worktree. The task's explicit Stop-The-Line instruction requires stopping
before adding compensating wrappers or synthetic requests. No such workaround
was implemented. An inspection of the locally cached v0.5.12 Plan shape also
showed no request or owned-environment member; this is not a claim about any
later remote release.

## Options and recommendation

1. **Recommended:** agents-management computes an owned-environment snapshot
   via `ChildEnv(nil, effectiveRequest)` at the admitted planning boundary and
   exposes it with the plan (or through an explicit admitted-plan result API).
   The launcher composer consumes that authoritative snapshot. This keeps
   private request preparation and vendor policy upstream, at the cost of an
   upstream API addition, release and scoped composition API update.
2. Expose the effective admitted request and system through a supported result
   API, preserving the current composition signature. This exposes a larger
   upstream contract, including its snapshot/mutation semantics.
3. Revise SPEC §4.5 to permit a documented request projection for the closed
   adapters. This weakens the exact-request contract and creates coupling to
   their current ChildEnv implementation; not recommended without an explicit
   architecture decision.

**External input needed:** the orchestrator/upstream owner selects and lands
the supported ownership API (recommended option 1), publishes an authorized
tag, and provides that tag plus the agreed launcher composition contract.
Then resume this task. No tag creation or upstream repository writes are
authorized in this producer workspace.

## Production-entry coverage inventory

These are inspection bounds, not newly driven integration claims. No new
production-entry test was added; new obligation coverage is **0/8 rows**.

| Obligation | Current main call-site evidence / remaining work |
| --- | --- |
| axconfig.Load before cli.Parse | Missing; run hardcodes AxConfigured=false |
| Fragment resolve, mapping, defaults/Lineup and origin group | Existing run wiring; unchanged in this run |
| Tagged BuildLaunch and separate provider limits | Existing helper API only; run never calls it |
| Prompt selection, PrepareLaunch and warnings | Existing helper API only; run never calls it |
| Composition from the admitted inputs | Blocked by missing effective request/owned env API |
| All three late checks in both modes | execution.Run owns checks; unreachable from run |
| Direct child status / fake-ax handoff / diagnostic byte boundaries | Existing execution API; run never calls it; execution still formats its own diagnostics |
| Remove not_implemented; installed umbrella/repair/model/MCP evidence | not_implemented remains; installation and real launches reserved to orchestrator after landing |

Pi's explicit epic decision remains in force: native Pi has no MCP channel.

## Validation actually executed

Shell: zsh, with `set -o pipefail`. Each command was a standalone process,
without tee or a pipe chain. No validation result was accepted from older
attached evidence.

| Command | Real exit | Scope |
| --- | --- | --- |
| `go test ./internal/composition ./internal/plan -count=1` | 0 | Existing helper tests only; composition 0.359s, plan 0.709s |
| `go build -o .temp/TASK-260908-1o7i8y/curator-run ./cmd/curator-run` | 0 | Existing launcher compiles; output not installed or executed |

No `make check`, production-entry goldens, new late-check/mode mutants, real
provider launches, real ax calls, installation, independent reviewer acceptance,
PR or signed delivery was performed. Those are unverified/pending, not passing.
No checklist items were checked. No product code or tests changed.

## Base and operational evidence

Worktree HEAD: `26baf9777e5a406aeeca343ab5a3b0a25925165e`.
Fresh `git ls-remote --symref origin HEAD` reported `refs/heads/main` at the
same OID; no fetch, branch mutation or managed-base reconfiguration was made.
`git status --short` and `git diff --stat` were empty before this ignored
scratch artifact. The runtime issued a Story-base-preflight reminder; this
read-only inspection does not claim to have satisfied the runtime's full
provisioning preflight.

Read the task briefs, epic original goal and continuation, current SPEC, and
the project-management / Go-testing skills. Public board directives reported
none. Early exploratory board projections using resources/artifacts were
rejected; schema inspection corrected them to preconditionResources and
outcomeResources. These were read-query errors, not validation gates.

### Logbook finding for the orchestrator

A2 exposes a gap between separately accepted API stages: a Plan-only admission
API cannot supply the exact request demanded by composition. Baseline helper
tests remain green despite this missing production seam. Record this finding
in the owning logbook during integration. This worker did not edit LOGBOOK.md
(campaign prohibition); `logbook` is not available as a command on this host.
The finding is persisted here and in task notes instead.
