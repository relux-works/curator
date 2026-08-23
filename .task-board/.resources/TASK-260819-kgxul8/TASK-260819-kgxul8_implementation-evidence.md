# TASK-260819-kgxul8 implementation and validation evidence

## Pinned published identity

- Release: `v1.0.0-rc.8`
- Immutable commit: `f8c405aa3ad0a39d260c2ed93684e55c5a346359`
- Signed annotated tag object: `ad247840292487d5d88ac44331798b6b4182a79f`
- Conformance manifest SHA-256: `d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1`
- Release metadata SHA-256: `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`
- Preserved rc.7 metadata SHA-256: `e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d`
- Published verified implementation, platform, and conformance claim sets: empty.

`curator-spec-pin` verifies the immutable revision shape and equality, exact
manifest and metadata bytes, protocol and claim versions, conformance-suite
cross-digests, portable/verified policy identifiers, empty claim sets, and the
immutable rc.7 compatibility mapping. CI invokes it for every job that checks
out the released suite.

## Compatibility mappings

- Rc.8 receipt cache keys remain the canonical logical build-input identity;
  the additive assurance-bound protected-cache address stays separate.
- Publication-conflict conformance now compares the protected winner key,
  while stored receipt validation continues to compare the logical key.
- The compiled dry-run lifecycle vector is executed with real no-process,
  no-session, no-artifact, no-persistent-mutation, no-lock, and temporary-state
  cleanup assertions. Its `logical_cache_key` is cross-checked exactly against
  `expected/build-driver/cache-key.txt`, and the protected address is proven
  not to alias the logical receipt key.
- Garbage-collection integration fixtures publish an explicit portable
  assurance binding and build-session receipt, and markers retain the returned
  assurance-bound protected key. This eliminated six platform-gate skips.
- Existing rc.5 repository-source grammar tests remain intentionally unchanged;
  advancing the release pin did not reinterpret preserved older wire semantics.

The exact rc.8 tree was regenerated with `go run ./tools/generate-vectors` in a
separate materialization. `diff -qr` against the published tree exited 0, so no
Curator-owned fixture copy or semantic rebaseline was needed.

## Validation results

All listed green commands were executed directly and exited 0:

- Exact release verifier:
  `go run ./cmd/curator-spec-pin --root .temp/TASK-260819-kgxul8/curator-spec-rc8 --revision f8c405aa3ad0a39d260c2ed93684e55c5a346359`.
- Focused pin and real-suite tests:
  `CURATOR_CONFORMANCE_ROOT=... go test ./internal/buildrepo ./cmd/curator-spec-pin`.
- Focused cache rejection mapping:
  `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -run TestCacheRejectionClustersMapToStableCuratorOutcomes ./internal/buildcache`.
- Focused install compatibility mapping:
  `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -run 'TestAuthoritative(CacheRejectionsAreRebuiltNeverAdopted|DryRunCasesMutateNothingPersistent)' ./internal/install`.
- Coordinator-requested compiled dry-run cross-vector proof:
  `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -run 'TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only' ./internal/install`.
- Final scopes package and post-change race:
  `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./internal/scopes` and
  `CURATOR_CONFORMANCE_ROOT=... go test -race -count=1 ./internal/scopes`.
- Full Go against exact rc.8:
  `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -timeout 30m ./...`.
- Full repository race against exact rc.8:
  `CURATOR_CONFORMANCE_ROOT=... go test -race -count=1 -timeout 30m ./...`.
- CI protocol/platform gate:
  `CURATOR_CONFORMANCE_ROOT=... EVIDENCE=.../ci-evidence-rerun make ci-test`;
  46 served packages, 0 deferred, 0 excluded, served tests exit 0, platform-case
  gate exit 0.
- Gate self-tests: `bash .github/ci/gate-selftest.sh`; 75 passed, 0 failed.
- Vet: `go vet ./...`.
- Build: `go build ./...`.
- Formatting: `test -z "$(gofmt -l cmd internal)"`.
- Pinned lint:
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run`;
  `0 issues.`
- Canonical verifier:
  `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md`;
  53 labeled records passed.
- Binary-deny gate:
  `go test -count=1 ./internal/artifactpolicy -run '^(TestCompiledVectorsC01ELFResolution|TestCompiledVectorsC02ThroughC12|TestCompiledDetectorErrorsNeverAdmitOpaqueBytes|TestPinnedCompiledFixtureProvenanceAndIdentity|TestNativeArchiveMetadataIsCanonicalAndRoleBound|TestDependencyDirectoryRejectsCompiledLinkAndBundleNodes)$'`.
- Deterministic regeneration: `go run ./tools/generate-vectors`, followed by
  `diff -qr` against the untouched exact release tree.
- Whitespace: `git diff --check`.

Development-red evidence is retained honestly:

- The first focused test command exited 1 because its synthetic uppercase
  revision used only digits; the fixture was corrected and the rerun passed.
- The first full rc.8 run exited 1 and exposed the logical/protected cache-key
  and compiled dry-run compatibility mappings above; the full rerun passed.
- The first pinned-lint run exited 1 for one obsolete helper and one missing
  narrow G304 justification; both were corrected and the rerun reported zero
  issues.
- The first `make ci-test` exited 2 because six scopes fixtures skipped on a
  missing assurance receipt; the fixture was corrected, its focused test
  passed, and the full gate rerun passed.

The worktree contained substantial pre-existing changes from prerequisite and
parallel board work. This task preserved them and did not stage or commit any
file.
