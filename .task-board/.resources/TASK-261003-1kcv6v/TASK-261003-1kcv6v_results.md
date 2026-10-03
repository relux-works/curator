# TASK-261003-1kcv6v — carry-rc14-lockstep

Ready for review. The accepted campaign delta was applied as a diff onto rewritten main `f17ea7331c258b59db84fc6ea83872cf638b51a0`, without conflicts. All 16 paths are byte-identical to the accepted campaign blobs; CHANGELOG.md needs no exception. Exactly those 16 tracked paths changed. HEAD remains the requested base; all changes remain uncommitted. No history from the source campaign was grafted, cherry-picked, or pushed.

The supplied delta includes its existing test changes; no extra code, tests, or workarounds were introduced. SPEC_PIN, writer activation, and LOGBOOK.md remain unchanged. Findings were recorded in board notes and this outcome because the binding host rules prohibit LOGBOOK edits. The generic logbook checklist was replaced with this explicit evidence requirement.

## Fresh local validation

Go version: 1.25.5, darwin/amd64. Every Go build/test/vet invocation and the Go-consuming gate self-test inherited GOFLAGS=-work; build/test execution was serialized under the shared build lock. Every WORK directory was retained. Commands ran directly without tee or pipelines. Syspolicyd was running before and after every command; successive crashes stayed 374 → 374. The initial bounded lock acquisition returned 75 while another operator held it, and started no builds; ownership was acquired normally after that operator released it.

| Command | Exit | Seconds | Crashes |
| --- | ---: | ---: | --- |
| `build: go build -o .temp/TASK-261003-1kcv6v/curator ./cmd/curator` | 0 | 1.46 | 374 → 374 |
| `candidate-conformance: go test -count=1 -json -timeout 8m ./internal/conformancecoverage ./internal/buildrepo ./internal/config ./internal/contextlock ./internal/envmarker ./internal/envprofile ./internal/interop/environments ./internal/marker ./internal/registry ./internal/scriptpolicy ./cmd/curator -run Conformance\|Coverage\|SchemaCases\|Muse\|ReadFailureVectors\|DeferredContentHashV2\|ContentHashV2Vectors\|ScriptWorkerProtocolVersionAcceptanceIsClosed\|ScriptExecutionPolicyIdentityMatchesTheSuite` | 0 | 113.39 | 374 → 374 |
| `default-conformance: go test -count=1 -json -timeout 8m ./internal/conformancecoverage ./internal/buildrepo ./internal/config ./internal/contextlock ./internal/envmarker ./internal/envprofile ./internal/interop/environments ./internal/marker ./internal/registry ./internal/scriptpolicy ./cmd/curator -run Conformance\|Coverage\|SchemaCases\|Muse\|ReadFailureVectors\|DeferredContentHashV2\|ContentHashV2Vectors\|ScriptWorkerProtocolVersionAcceptanceIsClosed\|ScriptExecutionPolicyIdentityMatchesTheSuite` | 0 | 69.818 | 374 → 374 |
| `formatting: python3 -c import subprocess; paths=subprocess.check_output(["git","diff","--cached","--name-only"],text=True).splitlines(); r=subprocess.run(["gofmt","-l",*[p for p in paths if p.endswith(".go")]],capture_output=True,text=True); print(r.stdout,end=""); assert r.returncode==0 and not r.stdout` | 0 | 0.138 | 374 → 374 |
| `gate-selftest: bash .github/ci/gate-selftest.sh` | 0 | 139.289 | 374 → 374 |
| `lint: golangci-lint run --timeout 8m` | 0 | 34.684 | 374 → 374 |
| `naming-gate: bash .github/ci/naming-gate.sh` | 0 | 26.087 | 374 → 374 |
| `no-broad-suppression: bash .github/ci/no-broad-suppression.sh` | 0 | 0.432 | 374 → 374 |
| `protected-paths: git diff --exit-code HEAD -- LOGBOOK.md .github/workflows/ci.yml internal/hashing/hashing.go` | 0 | 0.05 | 374 → 374 |
| `spec-consumption-gate: python3 <rc14-spec>/tools/implementation_coverage.py go --stream .temp/TASK-261003-1kcv6v/spec-implementations.log` | 0 | 0.149 | 374 → 374 |
| `spec-implementations: go test -count=1 -json -timeout 8m ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy` | 0 | 22.395 | 374 → 374 |
| `spec-presence-gate: python3 <rc14-spec>/tools/implementation_coverage.py families --implementation go --root <rc14-spec>/conformance/v1` | 0 | 0.139 | 374 → 374 |
| `vet: go vet ./...` | 0 | 8.422 | 374 → 374 |
| `whitespace: git diff --cached --check` | 0 | 0.071 | 374 → 374 |

Candidate-conformance used the authenticated rc.14 root (manifest SHA-256 `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`). Default-conformance used the authenticated rc.13 root (manifest SHA-256 `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`). Both passed 11/11 selected packages. Candidate: 929 passing test events; default: 793 (including nested subtests). Each skipped exactly one release-pin test because these roots publish a release other than the separately pinned rc.8. The selected -run expression bounds this evidence; it is not a full package-suite replay.

The exact six-package implementation command passed. Its real JSON stream upheld 8/8 declared consumption claims; the separate presence gate upheld 8/8 claims. The gate self-test passed 301 assertions with 0 failures, including its expected-red refusal cases. Real top-level exits are recorded in the table and attached command metadata.

## Reused evidence and limits

[Accepted hosted run 37017428885](https://github.com/relux-works/curator/actions/runs/37017428885) was freshly queried with gh, exit 0. Its overall conclusion and all three Candidate suite + platform-case gate steps are success (3/3). This is reused acceptance evidence for the unchanged carrier files, not a fresh three-OS run of the rewritten-base carrier. The accepted review also records default, race, lint, interop, and self-test success. No full local ./... test run, race suite, or newly dispatched hosted matrix was executed in this run: the carrier task preserves the already accepted delta; fresh validation covers affected conformance paths, the exact six-package implementation gate, full-module vet/lint, and the application build. No new mutant execution is claimed.

Fresh whitespace, formatting, protected-path, and byte-identity checks all exited 0. The initial identity check and the closing identity check measured 16/16. Only the diff was applied, and no new commit references pre-rewrite history.

## Byte identity (closing check exit 0)

| Path | Accepted blob | Worktree blob | Match |
| --- | --- | --- | --- |
| .github/ci/conformance-case-counts.tsv | f8f1eb69a7d4e8f1632ed844c8aefdafaf3d6b7c | f8f1eb69a7d4e8f1632ed844c8aefdafaf3d6b7c | True |
| .github/ci/conformance-gaps.tsv | d66d43ca540980269e64530f1713602f2dbe518e | d66d43ca540980269e64530f1713602f2dbe518e | True |
| CHANGELOG.md | e3d440076ae48ea805b08d763f78e88cf960796a | e3d440076ae48ea805b08d763f78e88cf960796a | True |
| cmd/curator/muse_test.go | 95412d346794dc43747388405563771d99d44d26 | 95412d346794dc43747388405563771d99d44d26 | True |
| internal/buildrepo/acquisition_conformance_test.go | 752408f33abda26424f2c1b8cd7a46e094339348 | 752408f33abda26424f2c1b8cd7a46e094339348 | True |
| internal/config/environments_conformance_test.go | 1668897bcdda9f5d2c8933cf799f2c4eec9faeb0 | 1668897bcdda9f5d2c8933cf799f2c4eec9faeb0 | True |
| internal/conformancecoverage/content_hash_v2_gaps_test.go | 3a1c7453d5a20b960c53839aeab065e0ba2f0a89 | 3a1c7453d5a20b960c53839aeab065e0ba2f0a89 | True |
| internal/conformancecoverage/coverage.go | 4d7ca1eda8e61740a68b934e4cc3edec709c085b | 4d7ca1eda8e61740a68b934e4cc3edec709c085b | True |
| internal/contextlock/schema_conformance_test.go | b8355867fffed12aac3a07d885b453590aab08b5 | b8355867fffed12aac3a07d885b453590aab08b5 | True |
| internal/envmarker/marker_env_schema_test.go | f41e75061d7b463e006da63b88ec1c65d6ee1fc5 | f41e75061d7b463e006da63b88ec1c65d6ee1fc5 | True |
| internal/envprofile/muse_test.go | d2fc911797f3e36ef7270781d5482b6a3921558c | d2fc911797f3e36ef7270781d5482b6a3921558c | True |
| internal/envprofile/read_failure_conformance_test.go | d85c33a1d5cdb56d305b8aba3073f0d9487a2a0d | d85c33a1d5cdb56d305b8aba3073f0d9487a2a0d | True |
| internal/interop/environments/snapshot_acquisition_test.go | 2ad8a7ad5bb11311f8045d6ad5fa185df0dbd52a | 2ad8a7ad5bb11311f8045d6ad5fa185df0dbd52a | True |
| internal/marker/schema_coverage_test.go | a2eb1c84333ecd8cdc5cff4784302120f496e929 | a2eb1c84333ecd8cdc5cff4784302120f496e929 | True |
| internal/registry/schema_conformance_test.go | 3a3d2e5b20307e65ece37ae7e58792955d2220f0 | 3a3d2e5b20307e65ece37ae7e58792955d2220f0 | True |
| internal/scriptpolicy/conformance_test.go | f3f2726d635f31e3bcefc1ae7ea9923c2471a29a | f3f2726d635f31e3bcefc1ae7ea9923c2471a29a | True |


Attached validation evidence contains copies of command logs and metadata with local workspace/home/temporary paths replaced by neutral placeholders. Private raw originals remain in ignored task-local storage; the accepted source review was read privately and is not republished here.
