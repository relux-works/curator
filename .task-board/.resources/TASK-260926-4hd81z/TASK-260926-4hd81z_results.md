# TASK-260926-4hd81z results

## Change

Updated [internal/scriptpolicy/conformance_test.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260926-nn2j3l/worktree/internal/scriptpolicy/conformance_test.go:500). The suite identity check accepts exactly protocol labels 1.0.0-rc.9 and 1.0.0-rc.13, with a comment documenting that rc.13 only relabels the unchanged script-worker-v1 identity. Added a closed-set test covering rc.9, rc.13, and rejection of rc.14. The existing schema_version, execution_policy, and interpreter-set assertions are unchanged.

## Suite roots

- Curator pinned suite: curator-spec commit dcc7f015e2d97edf2d52928afb6fd79ec8129e8b; vector label 1.0.0-rc.9; root `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.uNjnWGKdD2/pinned-suite/conformance/v1`.
- Candidate: curator-spec release/v1.0.0-rc.13 at f6bd748c59e015125b475428269ccdd367420d30; vector label 1.0.0-rc.13; root `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.uNjnWGKdD2/conformance/v1`.
- Candidate ref was freshly advertised, fetched into TMPDIR, and checked out detached at the requested head. Pinned commit was fetched and materialized separately in TMPDIR.

## Validation

Shell: /bin/zsh. Commands ran as standalone processes without tee or output pipes.

- `env CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.uNjnWGKdD2/pinned-suite/conformance/v1 go test ./internal/scriptpolicy -count=1` — exit 0. Ran twice, including after restoring the mutation probe.
- `env CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.uNjnWGKdD2/conformance/v1 go test ./internal/scriptpolicy -count=1` — exit 0. Ran twice, including after restoring the mutation probe.
- `go test ./internal/scriptpolicy -run '^TestScriptWorkerProtocolVersionAcceptanceIsClosed$' -count=1` — exit 0 on the restored implementation.
- `golangci-lint run ./internal/scriptpolicy` — exit 0, 0 issues.
- `go vet ./internal/scriptpolicy` — exit 0.
- `go build ./internal/scriptpolicy` — exit 0.
- `gofmt -w internal/scriptpolicy/conformance_test.go` — exit 0; `git diff --check` — exit 0.

## Negative evidence

- Crafted a valid rc.14-labelled vector from the rc.13 candidate root. `env CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.uNjnWGKdD2/rc14-mutant/conformance/v1 go test ./internal/scriptpolicy -run '^TestScriptExecutionPolicyIdentityMatchesTheSuite$' -count=1` — exit 1 as expected; the identity check reported rc.14 outside the closed set.
- Mutated `scriptWorkerProtocolVersionSupported` temporarily to return true for every version. `go test ./internal/scriptpolicy -run '^TestScriptWorkerProtocolVersionAcceptanceIsClosed$' -count=1` — exit 1 as expected: the rc.14 row returned true where false was required. Restored the closed-set implementation, then reran the package against both roots with exit 0.
- An initial rc.14 probe exited 1 because the temporary fixture writer produced invalid JSON. That attempt was discarded; after regenerating and parsing the vector as valid JSON with protocol_version rc.14, the probe failed at the intended version check as recorded above.

## Worktree and scope

Preflight confirmed the selected base is main at 60498052a1833f7511bd16086d953c8099fe7eed. Fresh `git ls-remote --symref origin HEAD refs/heads/main` resolved both remote HEAD and main to the same OID; local main and the Story worktree HEAD match it. Final worktree status contains only the intended conformance test file. No CHANGELOG.md or LOGBOOK.md edit was made.

## CHANGELOG entry (for release prep)

Accept the rc.13 suite label for the unchanged script-worker-v1 execution-policy identity while retaining a closed protocol-version set.
