# TASK-260916-hxr6qv results rev3 — rework 2 (P2 production-composition proof)

Role: developer. Worktree: `.temp/STORY-260916-v58b5y/worktree`, branch
`task-board/story/STORY-260916-v58b5y`, base checkpoint `c89473c`
(TASK-260916-27cv45 schema-2 loader). No commits made; all changes
uncommitted for orchestrator integration. Product code unchanged from
rev2; this delta is test-only (one new test function in an untracked
test file). Strict port/alias refusals and the frozen grammar untouched.

## P2 fix: production-composition end-to-end test

- `cmd/curator/draft_transport_provenance_test.go`: new
  `TestProductionExternalDepsFalseDrivesMirrorFetchToSink` (non-parallel,
  POSIX-only, bounded via `-timeout` and the lane deadline). It drives an
  ACTUAL external operation through `productionExternalDeps(cfg, false)`:
  a real `install app` (dryRun=false) over a temporary v2 machine policy
  with a listed mirror (`fixture.test/https-tools` primary
  `https://fixture.test/https-tools.git` fails DNS, mirror
  `https://mirror.fixture.test/https-tools.git` with
  `mirror_of=fixture.test/https-tools` succeeds via file rewrite;
  `fixture.test/ssh-tools` single success), fake transport on PATH (fails
  primary, rewrites mirror/ssh to fixture bare repos) and a
  secret-bearing provider (`prov-a` via broker `credential fill` returning
  `zz-secret-marker-260916-prod`), keeping the constructed sink.
- Asserts in the actual machine-private sink
  (`home/draft-transport-provenance.jsonl`): 6 records (plan
  primary/mirror/ssh + stage primary/mirror/ssh); 2 mirror success
  records carry canonical `identity=fixture.test/https-tools`,
  `url=mirrorURL`, `resolved_host=mirror.fixture.test`,
  `mirror_of=fixture.test/https-tools`, `succeeded=true`; every record
  names a canonical identity (https/ssh keys, never alias/mirror); fixed
  allowlisted keys only; no broker secret.
- Asserts portable artifacts (lock/receipt/marker/manifest family) carry
  no endpoint provenance and no secret: CLI stdout/stderr contain no
  mirror host/URL and no secret; protected cache receipts
  (`home/cache/build/go-v1/*`) contain no mirror host/URL and no secret;
  installed project markers/outputs under `project/` (excluding `.git`)
  contain none; `Skillfile.json` manifest contains none. Machine policy
  and the machine-private sink legitimately contain the mirror address
  and are excluded by construction (walk covers only portable cache and
  install outputs, never home root).
- Explicitly asserts `productionExternalDeps(cfg, false)` returns a
  non-nil trace with the switch on and no sink file before the run, so
  both sink mutants fail fast at that assert as well as via the missing
  sink file after the run.

## Verification (real exit codes, sh; precompiled test binaries under contention)

Unmutated, `-p 1 -count=1` via `go test -c` + binary where noted:

- New test `TestProductionExternalDepsFalseDrivesMirrorFetchToSink`
  (`go test -p 1 ./cmd/curator -run ... -timeout=150s`): PASS, exit 0
  (65.720s; plan+stage fetch pattern verified: 6 fetches
  primary/mirror/ssh ×2).
- `TestProductionExternalDepsAssignsTransportProvenanceSink` (cmd binary):
  PASS, exit 0.
- install draft set via binary
  `Test(AcquireDraftNetworkRevision2|Revision1ReaderRejectsV2Shapes|UserConfigIgnoredByResolvedLane|DraftTransportPlanV2|DraftPlanNeedsSSHRevision2|DraftTransport|DraftPlan|DraftSSH|DraftManager|DefaultAcquireFetchesTheDeclaredURL)`:
  PASS, exit 0.
- buildrepo via binary
  `Test(ValidateTransportPlan|PortAliasRefusalNarrows|ParseTransportEndpointURL|ExhaustionV2)`:
  PASS, exit 0.
- config via binary `Test.*(V2|Revision2|SourcePolicy|ResolveRepository)`:
  PASS, exit 0; `TestSchema1GoldenUnchanged`: PASS, exit 0.
- `go vet ./cmd/curator ./internal/install ./internal/buildrepo ./internal/config`:
  clean, exit 0. `gofmt -l` on touched trees: clean. `git diff --check`:
  clean, exit 0.
- Hosted gate: product code unchanged from rev2; rev2 validation log run
  35192522852 exit 0 (11 jobs, accepted in rev2 verdict) remains the
  product evidence. This rev adds only the cmd test above (PASS locally);
  the landing suite runs once on handoff per campaign rules (no manual
  full-suite run in addition).

## Executed mutant evidence (temporary overlays under /tmp, source restored; precompiled binaries)

- M1 attempt bound 2→1 (`transport.go:448` `> 2` → `> 1`): killed by
  `TestAcquireDraftNetworkRevision2Positives/mirror_first_with_availability_fallback_attempts_second`
  (install binary, overlay_m1.json): FAIL, exit 1 (0.28s,
  `repository_policy_invalid: transport plan requires one or two attempts`).
- M2 mirror class collapse (`transport.go:558`
  `CodeRepositoryMirrorUndeclared` → `CodeRepositoryPolicyInvalid`):
  killed by `TestValidateTransportPlanRevision2Refusals/undeclared_mirror`
  (buildrepo binary, overlay_m2.json): FAIL, exit 1 (0.00s, got
  `repository_policy_invalid`, want `repository_mirror_undeclared`).
- M3 alias class collapse (`sourcepolicy.go:651`
  `CodeRepositoryAliasUnknown` → `CodeRepositoryPolicyInvalid`): killed by
  `TestParseSourcePolicyV2Refusals/unknown_alias_with_empty_table`
  (config binary, overlay_m3.json): FAIL, exit 1 (0.00s, lacks class
  `repository_alias_unknown`).
- M4 removed sink (`main.go:1503` production assignment → `nil`):
  killed by the new test (cmd binary, overlay_nil.json): FAIL, exit 1
  (3.67s, `productionExternalDeps(cfg, false) left DraftTransportTrace nil`).
- M5 dry-run-only sink (`main.go:1503` → `if dryRun { ... }`): killed by
  the new test (cmd binary, overlay_dryrun.json): FAIL, exit 1 (2.12s,
  same assert). This is the mutant that survived both rev2 tests; it now
  fails.

## Findings and decisions

- Real `install app` fetches twice per repository (read-only plan + staging);
  both phases use the resolved lane with the production sink, so the
  primary/mirror/ssh pattern repeats (6 fetches, 6 sink records, 2 mirror
  successes). The test pins that doubled pattern explicitly.
- Sink placement unchanged from rev2 (manager-home
  `draft-transport-provenance.jsonl`, mode 0600, allowlisted fields only).
  Legacy lane never invokes it; switch-off runs create no file.
- `references/negative-evidence.md` still absent (noted in rev1/rev2);
  refusal rows keep the existing negative-row test shape.
- Story final leaf (story_final): workspace = 27cv45 checkpoint plus this
  leaf's delta (6 modified files, 4 untracked test files, only the new
  test function added in rev3), uncommitted. A base_authority_mismatch
  refusal at handoff is handled by the orchestrator — do not retry.

## Checklist

- [x] §6 resolution over ports/mirrors/aliases with attempt bounds and the two new failure classes, at the production entry; revision 1 and legacy goldens unchanged
- [x] §7 secrets/provenance/compatibility rules: canonical identity in provenance, no secrets in errors, negative rows for every refusal; mutants killed (M1-M5 exit 1 above)
- [x] Docs updated (rev1/rev2); narrow tests + remote gate green (narrow exit 0 above; hosted rev2 run cited, landing suite on handoff); story_final handoff from a clean workspace
- [x] Implementation matches AC; solution fits project architecture
- [x] Tests green (new test + narrow sets exit 0)
- [x] Relevant tests written for new/changed behavior and passing (new P2 test)
- [x] Lint clean (vet/gofmt/diff-check exit 0)
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name (this file)
- [x] Important findings/decisions/anomalies recorded here
