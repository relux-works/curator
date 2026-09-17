# TASK-260916-2bwfli results rev3 — rework 1: named provider admission (story_final)

Producer: developer. Shell: bash in the Story worktree
(`task-board/story/STORY-260910-1bhj0g`). All exit codes below are real.
Rev2 candidate patch applied first (`git apply --binary`, status showed
exactly the 7 rev2 leaf files), then rework 1. Executor
(`internal/buildrepo`) and policy loader (`internal/config/sourcepolicy.go`)
untouched: `git diff HEAD --stat` over both is empty.

## Why a new operator file

Rework 1 requires provider names to resolve "only through operator
configuration" with unknown/missing material refused. No such surface
existed: `CredentialProviders`/`ProviderSet` had no production population,
manager config has only identity/host scopes (not provider-keyed), and the
only provider-keyed operator state was the broker secret namespace. The
leaf therefore adds the minimal missing surface rather than re-pointing
the bypass at a substitute: `source-providers.json` beside the manager
configuration (machine-owned, machine-private, no secrets — HTTPS secrets
stay in the broker, SSH stays admitted paths). Absent file = no configured
providers; present-but-invalid = `repository_policy_invalid` before any
fetch (the §2 malformed-policy class, path in the message).

## Delta (9 files; rev2's 7 + 2 new)

- `internal/config/sourceproviders.go` (new): `SourceProvidersFileName`,
  `SourceProvidersPath/Load/Parse`; closed schema
  (`schema_version:1`, `providers` object, per-entry `https`/`ssh` only;
  anonymous⊕username; ssh needs identity or agent_socket).
- `internal/config/sourceproviders_test.go` (new): shape + 28 reject cases
  + load/missing/invalid.
- `internal/install/drafttransport.go`: `draftTransportAuth` DELETED;
  `acquireDraftNetwork` drops the repo-selection args and passes
  `buildrepo.CredentialProviders{Set, Reader}`; providers file loads only
  when a planned attempt names a provider (`draftPlanNamesProvider`).
- `internal/install/external.go`: `ExternalDeps` gains `DraftProvidersPath`
  + `DraftProviderReader`; call site updated.
- `cmd/curator/main.go`: CLI boundary sets both (`gitcred.Access{}` reader).
- Tests: `TestDraftTransportAuth` rewritten (was: unconfigured named HTTPS
  = anonymous; now: refusal, 0 fetches, plus distinct-material, silent,
  anonymous, invalid-file, providerless cases); selection subtests configure
  `operator-acme`; legacy subtests carry a `not json` providers file.
- `cmd/curator/draft_transport_test.go`: providers-file plumbing, fake-git
  `credential fill` arms + `cred:` request log, new
  `TestDraftTransportProviderAdmission` (unknown https/ssh, distinct
  providers, explicit anonymous); golden runs keep a `not json` providers
  file. Golden bytes UNCHANGED (md5 af83b55eb42e926b1b9b718701ebc500).
- `docs/draft-transport-resolution.md`: caller section rewritten for the
  provider table + broker design (the "maps to the same selection / out of
  scope" paragraph is gone).

## Evidence (narrow; direct-binary runs are the same test binaries `go test`
builds, executed with `-test.run/-test.count/-test.timeout`; see anomaly)

| command | exit |
|---|---|
| `go test -c` + config binary `-test.run 'TestParseSourceProviders\|TestLoadSourceProviders'` | 0 PASS |
| install binary `-test.run 'TestDraft\|TestAcquireDraft\|TestDefaultAcquire'` | 0: 9 PASS, 1 Windows SKIP (darwin) |
| `go test ./cmd/curator/ -run TestDraftTransportLegacyGolden` | 0 PASS 27.9s, golden md5 unchanged |
| CLI binary: `TestDraftTransportResolvedMatrix` (3/3), `TestDraftTransportProviderAdmission` (4/4) | 0 PASS |
| CLI binary from `cmd/curator`: `TestProductionBinaryDispatchesSSHWrapper` | 0 PASS (6/6 incl. passthrough) |
| mutant 1 (drop lookup: every plan-named provider admitted anonymous) vs install unknown-https | 1 FAIL `err = <nil>, want repository_endpoint_unavailable` — killed |
| mutant 1 vs CLI unknown-https | 1 FAIL `1 fetches, want 0` — killed |
| mutant 2 (`if !deps.DraftTransportResolution` → `if false`) vs golden | 1 FAIL switch-off exits 1 — killed |
| mutants reverted (`grep -c MUTANT` = 0), golden re-run | 0 PASS (above) |
| `gofmt -l` on all 7 Go files; `git diff --check` | 0, no output |
| `go vet` incl. tests (3 pkgs); `GOOS=windows go vet` (3 pkgs) | 0 |
| `golangci-lint run ./cmd/curator/ ./internal/install/ ./internal/config/` | 0, 0 issues |
| skip vocabulary | no new `t.Skip` text (POSIX-only + Windows-probe lines reused) |

## Anomalies (environmental, with proof)

- This host intermittently stalls fresh-binary exec (~50s+: `go run`
  hello took 50.5s with 0.3s user+sys; a test process sat at 0 CPU with no
  output past its own timeout). Several `go test` runs were stall-killed
  with zero test output; all were retried to clean PASS/FAIL verdicts.
  No stall signature was ever counted as a result.
- Direct-binary runs used only to dodge the wrapper stall; the golden ran
  via `go test` because its fetch bytes embed the binary path (only
  `$TMPDIR` spellings normalize). A direct golden run from `/tmp`
  mismatching on that path is a harness artifact, not a product finding.

Workspace holds exactly the 9 intended files (3 modified, 6 new); no
stray files, no commits. Executor admission/grammar/bounds unchanged;
switch not widened; frozen v1 untouched.
