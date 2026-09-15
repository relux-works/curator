# TASK-260915-23628n — review verdict, CR-TASK-260915-23628n-1 revision 1

Date: 2026-09-16. Reviewer run: independent (claude-fable-5-1). Verdict: **accepted**.

## Candidate identity

- Worktree HEAD = base OID `da59d8b1fe0070bfac803fa646da0779c7985511`.
- Working tree written through a temporary index → tree OID
  `abbeba4f36d6aa40ed607068eb0e17b5a6647157`, equal to the CR candidate tree.
  Re-verified after every mutant restore (same OID).
- 8 changed paths, +271/-1. No golden or testdata bytes changed
  (`git diff --stat base..candidate -- pkg/agentic/parity/testdata` is empty).

## API shape chosen by the producer (and why it is acceptable)

The producer took the brief's fallback shape: `agentic.BuildPlanWithEnvironment`
returning `PlanWithEnvironment{Plan, OwnedEnv}` and
`vendorplugin.BuildLaunchWithEnvironment` returning `BuildLaunchResult`
(a type alias of `agentic.PlanWithEnvironment`). `BuildPlan` and `BuildLaunch`
are thin wrappers over the shared private `buildPlan`/`buildLaunch` and pass a
nil capture pointer, so `agentic.Plan` and its JSON contract are byte-unchanged.
This satisfies "keep BuildLaunch byte-compatible" and avoids touching goldens.

Snapshot placement verified in `pkg/agentic/plan.go`: `sys.ChildEnv(nil, req)`
runs after `PrepareLaunchRequest` and after the single alias-substitution site
(`req.Model.ID = identity.Launched; req.Model.AliasOf = ""`), on the same `req`
used for ResolveBinary/Argv/ChildEnv/Stdin. Result is copied
(`append([]string(nil), ...)`) and sorted; an empty snapshot is nil.
In `pkg/vendorplugin/spawn.go` the capture uses the same `launch` value after
vendor Spawn, fidelity check, authoritative Runtime/Vendor assignment,
PrepareLaunchRequest, engine observation and preflight — i.e. the admitted
effective request.

Nil-parent semantics: all system `ChildEnv` implementations are pure over the
`parent` argument (claude filters then `WithRunContext`; pi-native/pi/codex
`WithRunContext`); no production code reads `os.Environ`. So `ChildEnv(nil, req)`
yields owned literals only, as SPEC §4.5 requires.

## Independent validation (zsh, `set -o pipefail`, real exit codes)

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./pkg/agentic/... ./pkg/vendorplugin/...` | 0 |
| `go vet ./pkg/agentic/systems/codex` | 0 |
| `go test ./pkg/agentic ./pkg/vendorplugin ./pkg/agentic/systems/... -count=1` | 0 (all 10 packages ok, incl. codex goldens on this amd64 host) |

The full landing suite was not rerun manually (runtime runs it once at handoff).

## Gate attacks (mutants applied temporarily, then restored; tree OID re-checked)

| Mutant | Result |
| --- | --- |
| A: move owned snapshot before `PrepareLaunchRequest`/alias projection | KILLED by `TestBuildPlanWithEnvironmentPreparedAlias` (exit 1: got `A_MODEL=pangolin-large B_ALIAS=launch-identity Z_PROMPT=do the thing`) |
| B: hard-code `aarch64`/`arm64` triple in the Codex golden expectation on this x86_64 host | KILLED: `TestPlansMatchTheCodexGoldens` and `TestCodexGoldenRejectsWrongArchitecture` both fail (exit 1) |
| C: drop `sort.Strings` on the snapshot | KILLED by `TestBuildPlanWithEnvironmentPreparedAlias` (exit 1) |

Test-embedded checks also cover: snapshot independence from plugin storage,
zero result on ChildEnv error, legacy `BuildPlan` not invoking the owned
snapshot, and byte-equal JSON between `BuildLaunch` and
`BuildLaunchWithEnvironment(...).Plan` for claude / codex / pi-native
(pi-anthropic → `pi-native` per `pkg/vendorplugin/runtime.go`) across exec,
dry-run, managed-session and interactive modes (unsupported modes asserted as
`ErrUnsupportedLaunchMode`).

## Bounds (recorded, not blocking)

- Mutant A survived at the `pkg/vendorplugin` level alone: the real systems'
  owned literals depend only on RunContext, so the end-to-end test cannot see
  preparation/alias ordering. The ordering gate is proven at the `agentic`
  entry point, which the vendorplugin production path delegates to unchanged.
- The vendorplugin test reconstructs the effective request in the test body
  (Spawn → Runtime/Vendor → Prepare → alias projection). It mirrors the
  production sequence as read in `spawn.go`; it is a parallel derivation, not
  the returned request itself, which is the intended design of option 1.
- arm64 execution remains unverified on this repository (no arm64 host or CI);
  the golden test's arm64 branch is inspection-only. Linux/Windows branches in
  `prepareParityCase` are likewise unexercised here.
- No tag was created by the producer or the reviewer (verified: no new refs).

## Definition of Done mapping

Snapshot API exposed, deterministic and documented (doc.go both packages);
CHANGELOG Unreleased v0.5.13 entry present; existing plans/goldens/JSON
unchanged; Codex golden made architecture-neutral without loosening the exact
path comparison; narrow tests green with exit codes above; narrowing mutants
killed. Verdict: accepted → `accept_cr(TASK-260915-23628n, revision=1)`.
