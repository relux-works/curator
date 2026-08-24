# TASK-260811-tkurtl validation gate log bundle

Every command ran as a standalone process from the repository root.

## focused-tests-01.log

Command: `go test -timeout 20m -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/ ./internal/artifactpolicy/...`

Exit code: 0

```
ok  	github.com/relux-works/curator/internal/swiftpminterop	2.584s
ok  	github.com/relux-works/curator/internal/swiftpmsource	11.254s
ok  	github.com/relux-works/curator/internal/closureexec	5.018s
ok  	github.com/relux-works/curator/internal/closuregraph	11.222s
ok  	github.com/relux-works/curator/internal/artifactpolicy	37.884s
?   	github.com/relux-works/curator/internal/artifactpolicy/conformance	[no test files]
```

## race-tests-01.log

Command: `go test -race -timeout 20m -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/`

Exit code: 0

```
ok  	github.com/relux-works/curator/internal/swiftpminterop	17.687s
ok  	github.com/relux-works/curator/internal/swiftpmsource	13.213s
```

## lint-01.log

Command: `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... (v2.12.2)`

Exit code: 0

```
0 issues.
```

## lint-full-01.log

Command: `golangci-lint run ./... (v2.12.2)`

Exit code: 0

```
0 issues.
```

## suite-no-cmd-01.log

Command: `go test -timeout 30m -count=1 $(go list ./... | grep -v cmd/curator)`

Exit code: 0

```
ok  	github.com/relux-works/curator/internal/adapters	0.463s
ok  	github.com/relux-works/curator/internal/artifactpolicy	127.147s
?   	github.com/relux-works/curator/internal/artifactpolicy/conformance	[no test files]
ok  	github.com/relux-works/curator/internal/audit	1.034s
ok  	github.com/relux-works/curator/internal/buildcache	3.817s
ok  	github.com/relux-works/curator/internal/buildmeta	1.350s
ok  	github.com/relux-works/curator/internal/buildrepo	44.872s
ok  	github.com/relux-works/curator/internal/buildsource	2.401s
ok  	github.com/relux-works/curator/internal/capabilities	2.737s
ok  	github.com/relux-works/curator/internal/closure	6.996s
ok  	github.com/relux-works/curator/internal/closureexec	24.757s
ok  	github.com/relux-works/curator/internal/closuregraph	14.558s
ok  	github.com/relux-works/curator/internal/config	5.004s
ok  	github.com/relux-works/curator/internal/devsub	4.607s
ok  	github.com/relux-works/curator/internal/envfiles	4.276s
ok  	github.com/relux-works/curator/internal/gitignore	4.023s
ok  	github.com/relux-works/curator/internal/gitops	5.961s
ok  	github.com/relux-works/curator/internal/globalbins	5.129s
ok  	github.com/relux-works/curator/internal/godriver	87.954s
ok  	github.com/relux-works/curator/internal/hashing	4.959s
ok  	github.com/relux-works/curator/internal/identifiers	4.590s
ok  	github.com/relux-works/curator/internal/identity	4.574s
ok  	github.com/relux-works/curator/internal/install	110.360s
ok  	github.com/relux-works/curator/internal/install/atomicity	110.464s
ok  	github.com/relux-works/curator/internal/interop	4.502s
ok  	github.com/relux-works/curator/internal/locale	4.730s
ok  	github.com/relux-works/curator/internal/managerlock	5.683s
ok  	github.com/relux-works/curator/internal/manifest	4.324s
ok  	github.com/relux-works/curator/internal/marker	4.664s
ok  	github.com/relux-works/curator/internal/mcp	4.721s
ok  	github.com/relux-works/curator/internal/nodesource	5.456s
ok  	github.com/relux-works/curator/internal/npmsource	82.478s
ok  	github.com/relux-works/curator/internal/pnpmsource	51.453s
ok  	github.com/relux-works/curator/internal/protocoljson	4.220s
ok  	github.com/relux-works/curator/internal/registry	5.708s
ok  	github.com/relux-works/curator/internal/runtimestore	18.200s
ok  	github.com/relux-works/curator/internal/rustsource	137.248s
ok  	github.com/relux-works/curator/internal/scopes	5.219s
ok  	github.com/relux-works/curator/internal/shell	4.814s
ok  	github.com/relux-works/curator/internal/skillcheck	3.250s
ok  	github.com/relux-works/curator/internal/skillspec	4.324s
ok  	github.com/relux-works/curator/internal/snapshot	3.988s
ok  	github.com/relux-works/curator/internal/staging	2.172s
ok  	github.com/relux-works/curator/internal/swiftpminterop	20.255s
ok  	github.com/relux-works/curator/internal/swiftpmsource	33.799s
ok  	github.com/relux-works/curator/internal/transaction	69.906s
ok  	github.com/relux-works/curator/internal/ui	1.749s
ok  	github.com/relux-works/curator/internal/verr	2.078s
ok  	github.com/relux-works/curator/internal/version	2.328s
ok  	github.com/relux-works/curator/internal/whitelist	2.996s
ok  	github.com/relux-works/curator/internal/yarnclassicsource	11.987s
ok  	github.com/relux-works/curator/internal/yarnmodernsource	11.570s
```

## cmd-curator-subset-01.log

Command: `go test -timeout 9m -count=1 -run 'TestStatus|TestGlobalStatus|TestGC|TestToolchainHost|TestLifecycle' ./cmd/curator/`

Exit code: 0

```
ok  	github.com/relux-works/curator/cmd/curator	81.501s
```

## canonical-verifier-01.log

Command: `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb`

Exit code: 0

```
canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2
canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true
```

