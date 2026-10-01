# TASK-260728-20ao7p rev2 review verdict: ACCEPTED

Candidate: base bab2433b, tree 0d59fab2. The worktree tree was recomputed through a temp index and is byte-identical (0d59fab2…).

## Production fix (buildrepo.CacheArtifactName)
- Producers: `StoreArtifact` writes `CacheArtifactName(target.goos)`, taken from the receipt input (protected.go). Consumers: `LookupArtifact` reads the same name, from the same input. `stageExternalBuilds` (external.go, two sites) uses the adapter GOOS. targets.go derives the shim target from `filepath.Base(entry.artifactPath)`, which is the staged name. external.go:369 uses `Dir`, so it is unaffected. The only remaining `"artifact"` literal in production is the Unix branch of the helper.
- Cache key and receipt bytes are unchanged. The key derives from the logical input. The receipt's `artifact.path` stays `bin/<cmd>[.exe]` from `buildmeta.ArtifactPath`, which already suffixed `.exe` on Windows. Only the physical file name inside the private entry changed. Unix entries are untouched (`artifact`).
- Windows entries written before this fix (extensionless `artifact`): `LookupArtifact` fails with `build_repository_artifact_invalid`. With `mutate` it quarantines the entry, and the pipeline then falls through to a rebuild. A dry run reports `corrupt` with the code. Nothing breaks silently, and pre-fix Windows external shims never ran anyway. The unit test pins this (extensionless Windows artifact is refused).

## Unit test and mutant
- `TestProtectedArtifactNameMatchesTargetPlatform` drives the real store writer and reader for darwin, linux and windows, in both receipt namespaces (`artifacts`, `artifacts-receipt-3`), on any host. It checks the physical file, the receipt path and key, and the cache-hit round trip.
- Mutant: `CacheArtifactName("windows")` returns "artifact". Both windows subtests FAIL with `open …/artifact.exe: no such file`, exit=1. The file was restored byte-identical.
- With the fix: `go test ./internal/buildrepo -run 'ProtectedArtifactName|AdoptArtifact'` exit=0.
- `GOOS=windows go vet` on buildrepo, install, crossconformance and cmd/curator: exit=0. Host `go vet` on buildrepo, install and cmd/curator: exit=0.

## Black-box
- `go test ./cmd/curator -run NativeBlackbox -count=1 -v`: exit=0, PASS (335 s on darwin).
- Drives the real binary through: a refused install over an ordinary checkout (`build_repository_source_unavailable`, nothing published) → install (outcome=would-preflight-and-build, shim, artifact and receipt) → shim run (exit 0, `blackbox-tool-ok`) → reinstall (outcome=cache-hit, same key, artifact mtime and receipt unchanged) → remove (shim, skill and build root gone).
- The test HOME is created via `privatedir.Make` and validated. The test fails on a "build cache sweep skipped" warning.
- No Windows skip. The ledger row is `darwin,windows` must-run, with a linux skip tolerated as platform-control.
- Gate (hosted, per the orchestrator) was green on every lane including Windows.

## Docs and hygiene
- `docs/external-build-repositories.md` matches the code. It covers schema 7 `build_repositories`, the locked commit and tag, `Skillfile.dev.json` substitution, and the admission → audit → cache → compiler order. Its link to the spec's guide is present, and the README line links it. Sibling doc links (build-ssh, build-https, authoring-cli-commands, compiled-commands, troubleshooting) resolve.
- No CHANGELOG or LOGBOOK, no stray files, no Windows-reserved names. 11 paths, as declared.

## Residual (non-blocking)
- Hosted Windows is the arbiter for the black-box itself. It was green per the gate, and I did not re-run it.