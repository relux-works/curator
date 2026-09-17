# TASK-260916-hxr6qv results rev2 — rework 1 (P1 provenance sink)

Role: developer. Worktree: `.temp/STORY-260916-v58b5y/worktree`, branch
`task-board/story/STORY-260916-v58b5y`, base checkpoint `c89473c`
(TASK-260916-27cv45 schema-2 loader). No commits made; all changes
uncommitted for orchestrator integration. Rev1 candidate content
unchanged except the P1 wiring below; strict port/alias refusals and
the frozen grammar untouched.

## P1 fix: production provenance sink

- `internal/install/drafttransport.go`: new production sink
  `DraftTransportProvenanceTrace(home)` / `DraftTransportProvenancePath(home)` /
  `DraftTransportProvenanceFileName` (`draft-transport-provenance.jsonl`).
  Each sanitized `AttemptRecord` appends one JSON line with a fixed
  allowlisted field set only (canonical identity, listed URL, resolved
  host/port, alias and `mirror_of` when used, lane transport, provider
  identifier, outcome class). The record type carries no secrets, and the
  allowlist keeps it that way structurally. File mode 0600 directly under
  the manager home (already exists; no new directory); best-effort writes
  never fail an acquisition; empty home returns nil (fail-closed).
- `cmd/curator/main.go` `productionExternalDeps` now assigns
  `deps.DraftTransportTrace = install.DraftTransportProvenanceTrace(cfg.Home())`.
  The legacy lane never invokes it, so switch-off runs create no file.
- `docs/draft-transport-resolution.md` production-caller section updated
  (nil-sink wording replaced with the manager-home sink).

## Tests

- `internal/install/drafttransport_provenance_test.go` (new):
  `TestAcquireDraftNetworkRevision2ProvenanceSink` through the production
  entry `acquireDraftNetwork` with the production sink construction and the
  fake-git lane: mirror success records canonical identity + listed URL +
  resolved mirror host + `mirror_of` in the sink; exhaustion with
  secret-bearing broker keeps secrets out of the sink while errors stay
  closed (canonical identity + class only); snapshot carries no endpoint
  properties; manager home holds only the sink file; nil trace records
  nothing without changing the verdict; empty home is nil; path stays
  under the manager home.
- `cmd/curator/draft_transport_provenance_test.go` (new, fast, no fixtures):
  `TestProductionExternalDepsAssignsTransportProvenanceSink` calls the real
  `productionExternalDeps`, asserts a non-nil trace, invokes it, and asserts
  the manager-home sink holds the canonical identity with listed/resolved
  mirror provenance under exactly the allowlisted JSON keys. Removing the
  production assignment fails this test by construction.

## Verification (real exit codes, sh with pipefail where piped)

Shared host was contended (parallel agents); slow first invocations were
cold-binary startup, so suites ran via precompiled test binaries
(`go test -c`) with `-p 1 -count=1 -run` filters:

- `/tmp/install_v2.test -run TestAcquireDraftNetworkRevision2ProvenanceSink` → PASS (exit 0)
- install draft set `Test(AcquireDraftNetworkRevision2|Revision1ReaderRejectsV2Shapes|UserConfigIgnoredByResolvedLane|DraftTransportPlanV2|DraftPlanNeedsSSHRevision2|DraftTransport|DraftPlan|DraftSSH|DraftManager|DefaultAcquireFetchesTheDeclaredURL)` → 16 top-level PASS, no FAIL (exit 0)
- buildrepo `Test(ValidateTransportPlan|PortAlias|ParseTransportEndpointURL|ExhaustionV2)` → 7 PASS (exit 0)
- config `Test.*(V2|Revision2|SourcePolicy|ResolveRepository)` → 18 PASS (exit 0)
- cmd `TestProductionExternalDepsAssignsTransportProvenanceSink` → PASS (exit 0)
- `go vet` on install/buildrepo/config/cmd → clean (exit 0)
- `gofmt -l` on touched trees → clean; `git diff --check` → clean
- `golangci-lint run` on the four touched trees → 0 issues (exit 0)
- `sh scripts/remote-gate.sh` → exit 0, run
  https://github.com/relux-works/curator/actions/runs/35188951833 finished
  success (Race ubuntu/macos, Test ubuntu/macos/windows, Lint, Interop,
  Naming, Gate self-tests; rose-air and Candidate suite skipped)

## Executed mutant evidence (temporary overlays, source restored after each)

- M1 attempt bound 2→1 (`transport.go:448` `> 2` → `> 1`): killed by
  `TestAcquireDraftNetworkRevision2Positives/mirror_first_with_availability_fallback_attempts_second` → exit 1,
  `--- FAIL` (2.86s, assertion failure, no timeout). Restored (`> 2` verified).
- M2 mirror class collapse (`transport.go:558`
  `CodeRepositoryMirrorUndeclared` → `CodeRepositoryPolicyInvalid`):
  killed by `TestValidateTransportPlanRevision2Refusals/undeclared_mirror`
  → exit 1, `--- FAIL`. Restored (identifier verified).
- M3 alias class collapse (`sourcepolicy.go:651`
  `CodeRepositoryAliasUnknown` → `CodeRepositoryPolicyInvalid`): killed by
  `TestParseSourcePolicyV2Refusals/unknown_alias_with_empty_table` → exit 1,
  `--- FAIL`. Restored (identifier verified).
- M4 nil sink (`main.go:1503` production assignment → `nil`): killed by
  `TestProductionExternalDepsAssignsTransportProvenanceSink` → exit 1,
  `--- FAIL` (0.12s). Restored (assignment verified).
- Nil-trace counterfactual is also pinned inside the end-to-end test
  (`nil trace records nothing and changes no verdict`): the positive sink
  assertions fail without a wired sink, since the sink file is their only
  provenance source.

## Findings and decisions

- Sink placement: no pre-existing operation-diagnostics log file exists in
  the curator manager home (operation output otherwise goes to
  stdout/stderr result reporting). The sink is a new machine-private file
  directly under the manager home beside the machine policy — never a
  portable artifact (locks, receipts, markers, manifests untouched; receipt
  transport enum unchanged). CLI stdout/stderr output is byte-identical for
  the legacy lane (trace never invoked there).
- `references/negative-evidence.md` still absent from this tree (noted in
  rev1); refusal rows keep the existing negative-row test shape.
- Story final leaf (story_final): workspace = 27cv45 checkpoint plus this
  leaf's delta (6 modified files, 4 untracked test files), uncommitted.
