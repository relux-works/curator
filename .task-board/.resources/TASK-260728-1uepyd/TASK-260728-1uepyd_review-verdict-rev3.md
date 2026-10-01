# TASK-260728-1uepyd review verdict — CR rev3 — ACCEPTED

Candidate: base 5432c85f, tree 5a1f7895 (5 paths). Exact-tree proof: temp-index `git add -A -- .github internal` + `write-tree` on the Story worktree = 5a1f7895c9a17d181f021adbbbaae9c13c798fde. Hosted gate (rev3 validation log): every lane success incl. Test (windows-latest), Race, Lint, Interop conformance gate, Gate self-test x3. Spec pin 23435129 = v1.0.0-rc.13; vector from `git archive` in /tmp/spec13 (protocol_version 1.0.0-rc.13; 12 cases / 55 argv / 17 env / 11 forbidden).

## Rev2 finding F1 — fixed
`verifyAcquisitionFetchCallSite` (acquisition_conformance_test.go:557-600) compares the FULL logged fetch argv (reflect.DeepEqual against vector common_fetch_argv resolved per case; `<operation-private>` resolved from the logged --git-dir / core.hooksPath after asserting one `curator-buildrepo-*` private root with repo.git / empty-hooks) and the FULL environment Git received (shim logs os.Environ(); compared with DeepEqual against clean_environment + transport additions + platform allowlist). Applied to every one of the 11 fetching cases; malformed-ref case asserts zero Git invocations. Shim and fixture only; production diff is still rev2's ValidateGitTool reorder (admission.go, 9 lines).

## Measured, on a byte-identical copy (diff -rq internal/ .github/ identical, file restored + cmp after each mutant)
Baseline `go test ./internal/buildrepo -run 'ExternalRepositoryAcquisition|TestLocalConfigAndAdministrationAdversarialBoundaries|TestPackIndexConformanceAndExactSSHWrapper' -count=1 -v` rc=0. Ratios: acquisition/cases 12/12, common-fetch-argv 55/55, clean-environment 17/17, forbidden-fetch-features 11/11, local-config-and-refs 15/15, pack-index 8/8; 0 known-gap / 0 bound / 0 skipped. Count rows 12/17/55/11 in conformance-case-counts.tsv match.
`go vet ./internal/buildrepo` rc=0; `GOOS=windows go vet ./internal/buildrepo` rc=0.

Mutants (real exit codes), all KILLED:
- M1 extra fetch flag `--verbose` in strictFetchArgs: rc=1 (argv[45] + production fetch argv).
- M2 env leak `GIT_SSL_NO_VERIFY=1` in cleanGitEnvironment: rc=1 (clean-environment "unexpected environment entries" + production fetch environment).
- M3a drop `--no-tags` + add `--depth=1`: rc=1 (argv, forbidden-fetch-features "admits forbidden feature depth"); M3b add `--prune`: rc=1 (argv[45], forbidden "prune", production fetch argv).
- M4 call-site `-c fetch.prune=true` prepended at the fetchArgs call site: rc=1 (cases: "production fetch argv") — previously SURVIVED.
- M5 call-site `fetchEnv = append(fetchEnv, "GIT_SSL_NO_VERIFY=1")`: rc=1 (cases: "production fetch environment") — previously SURVIVED.

## Residuals
- R-a fixed: helper-selected-transport row now calls production ParseSource on an `ext::` URL and requires refusal, plus argv checks for protocol.allow=never / explicit selected protocol.
- R-b stated as an owned bound (results + comment at the success branch): the 7 successful raw-object fixtures have no build descriptor, so downstream audit/cache/compiler order is not observed for them; error cases ARE driven through RunPipeline with Audit hook and phase trace == ["exact-source-acquisition"]. The old vacuous vector-data assertion was removed. Acceptable under the rework brief ("or state a bound").
- R-c unchanged (symlink-privilege fallback bound; driven on all hosted lanes).
- Non-blocking note: fetchEnv additionally carries EnvHTTPSBrokerState / EnvSSHWrapperState only when credentials / sshPolicy are set; the consumer's fixtures do not set them, so those two broker variables are outside the DeepEqual. Not required by the vector.
- Non-blocking: only the fetch invocation's env is compared (init / version probes are not) — within the rework scope.

## Hygiene
5 paths only; no CHANGELOG/LOGBOOK edit; no stray files; state reads via internal/stateread; no Windows-reserved names; no Windows skip (fixture-only fix of rev1 stands); no employer names.
