# BUG-261001-2772iz — askpass-pipe-refusal-epipe: reviewer command evidence, revision 2

Run RUN-261001-6e13a7. Native environment: Go 1.26.0 darwin/amd64. Each exit below is the actual shell process exit observed by the reviewer. No product-code or index edits; base and mutants were exported to /tmp.

## Base + candidate deterministic test

Command: `go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$' -count=1 -v`

Exit code: **1**.

```text
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
--- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe (0.17s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user (0.05s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username (0.02s)
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	0.667s
FAIL
```

## Candidate deterministic and security rows

Command: `go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$|^TestHTTPSBrokerSecretTransportReportsWriteFailureAfterRequest$|^TestHTTPSCredentialBrokerRejectsClosedSecretSocket$|^TestHTTPSBrokerSecretTransportRejectsInvalidRequest$' -count=1 -v`

Exit code: **0**.

```text
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username
--- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe (0.33s)
    --- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user (0.20s)
    --- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt (0.02s)
    --- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments (0.02s)
    --- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument (0.05s)
    --- PASS: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username (0.02s)
=== RUN   TestHTTPSBrokerSecretTransportReportsWriteFailureAfterRequest
--- PASS: TestHTTPSBrokerSecretTransportReportsWriteFailureAfterRequest (0.00s)
=== RUN   TestHTTPSBrokerSecretTransportRejectsInvalidRequest
--- PASS: TestHTTPSBrokerSecretTransportRejectsInvalidRequest (0.00s)
=== RUN   TestHTTPSCredentialBrokerRejectsClosedSecretSocket
--- PASS: TestHTTPSCredentialBrokerRejectsClosedSecretSocket (0.22s)
PASS
ok  	github.com/relux-works/curator/internal/buildrepo	1.521s
```

## Targeted mutant: write on EOF without request

Command: `go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$' -count=1 -v`

Exit code: **1**.

```text
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write https-askpass-server: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write https-askpass-server: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write https-askpass-server: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write https-askpass-server: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write https-askpass-server: broken pipe
--- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe (0.09s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username (0.02s)
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	0.524s
FAIL
```

## Full transport revert mutant

Command: `go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$' -count=1 -v`

Exit code: **1**.

```text
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
=== RUN   TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username
    httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
--- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe (0.20s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user (0.13s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument (0.02s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username (0.02s)
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	0.715s
FAIL
```

## Darwin broker 20 repetitions

Command: `go test ./internal/buildrepo -run 'HTTPSCredentialBroker|HTTPSBrokerSecretTransport' -count=20`

Exit code: **0**.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	8.501s
```

## Darwin dispatch race 50 repetitions

Command: `go test -race ./internal/crossconformance -run TestDraftSourcesBrokerAskpassDispatch -count=50`

Exit code: **0**.

```text
ok  	github.com/relux-works/curator/internal/crossconformance	57.384s
```

## Darwin full relevant packages

Command: `go test -race ./internal/buildrepo ./internal/testcli -count=1`

Exit code: **0**.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	101.817s
?   	github.com/relux-works/curator/internal/testcli	[no test files]
```

## Linux broker 20 repetitions

Command: `docker run --rm -v "$PWD:/src:ro" -w /src golang:1.26.0 go test ./internal/buildrepo -run HTTPSCredentialBroker -count=20`

Exit code: **0**.

```text
go: downloading golang.org/x/text v0.14.0
go: downloading golang.org/x/sys v0.36.0
ok  	github.com/relux-works/curator/internal/buildrepo	3.769s
```

## Lint

Command: `golangci-lint run ./internal/buildrepo/... ./internal/testcli/...`

Exit code: **0**.

```text
0 issues.
```

## Other reviewer checks

`go vet ./internal/buildrepo ./internal/testcli`: exit 0. `git diff --check <base>` and gofmt of all five changed Go files: exit 0, no output. Hosted run metadata and logs fetched successfully.

An early base invocation started before archive preparation completed and reported no tests; it is discarded, not counted. The completed base rerun above executes all five cases. An initial all-blob hash walk could not follow a repository symlink; identity was verified instead by candidate diff (empty apart from the new test) and exact hashes for all six changed files including that test.

## Candidate source identity

```json
{
  "base": "54d4aface89e22bf4a4116505eba3ecfe668efa8",
  "candidate_tree": "4189bda212dd1bb79f2e17f7ec00881ab7b4e2ed",
  "changed_files": [
    {
      "path": "CHANGELOG.md",
      "match": true,
      "sha256": "6fc83450332200b8dc2ac718f32eca171a0099fdd7c6f8f7fd3710bd6e86786a"
    },
    {
      "path": "internal/buildrepo/httpsbroker_pipe_unix.go",
      "match": true,
      "sha256": "5b922ad8362fdbeab1a03958ad76c6478a27c78d58aa74395d3fb4f368384cfd"
    },
    {
      "path": "internal/buildrepo/httpsbroker_pipe_unix_test.go",
      "match": true,
      "sha256": "38f4bfc5c8dd332109e0ff1869d15e9e6b3a3afc947817b9f374bbac32d2d971"
    },
    {
      "path": "internal/buildrepo/httpsbroker_test.go",
      "match": true,
      "sha256": "444ffd3cfd98dee7426cbe99746b2d9aab91e8365494f1701d127a0978d045ea"
    },
    {
      "path": "internal/buildrepo/httpsbroker_test_pipe_unix_test.go",
      "match": true,
      "sha256": "c6f19bca01e6dec754823fb4f124afe81c91ab02f4b697f85bd43725370a1b2d"
    },
    {
      "path": "internal/buildrepo/httpsbroker_test_pipe_windows_test.go",
      "match": true,
      "sha256": "626f9002e2a34b0a90701185e34685fb1d7faef747e651a45531c76578b63316"
    }
  ],
  "upstream": "e87d488b8fd892bc86c037ae5e325e596df8e745",
  "overlap": []
}
```

## Reused exact-tree hosted evidence

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing 144a79ee0a1622d282d8b630a385bf34b7995d83 as gate/STORY-261001-17ali8/261001-101557-95917-1
remote: 
remote: Create a pull request for 'gate/STORY-261001-17ali8/261001-101557-95917-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261001-17ali8/261001-101557-95917-1        
remote: 
remote gate: run 36848051591 (https://github.com/relux-works/curator/actions/runs/36848051591)
remote gate: run 36848051591 finished: success
  Lint: success  
  Test (windows-latest): success  
  Race (macos-latest): success  
  Gate self-test (macos-latest): success  
  Test (macos-latest): success  
  Interop conformance gate: success  
  Gate self-test (ubuntu-latest): success  
  Naming gate: success  
  Race (ubuntu-latest): success  
  Gate self-test (windows-latest): success  
  Test (ubuntu-latest): success  
  Test (rose-air): skipped  
  Candidate suite (${{ matrix.os }}): skipped  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown
```

Hosted run 36848051591 head 144a79ee0a1622d282d8b630a385bf34b7995d83 has Git tree 4189bda212dd1bb79f2e17f7ec00881ab7b4e2ed, equal to the CR candidate. Independently queried GitHub: conclusion success. Windows artifact `test-evidence-windows-latest/test/go-test.json` selected terminal records below are all pass; zero selected skips/failures. Raw stream SHA256: 867e5de5dc23b332e15a700ba458c6a977b8d4edd9698f99914aed9e541730d6.

```json
[
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSBrokerNamedPipeHasCurrentUserOnlyDACL",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSBrokerNamedPipeServesOneClientAndThenRejectsAnother",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSBrokerNamedPipeServerStopsWhenFetchExitsWithoutAClient",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/username",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/foreign_host",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/foreign_prompt",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/extra_argument",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/absent_secret_transport",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/malformed_secret_transport",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/absent_state",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/unreadable_state_shape",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSCredentialBrokerRejectsEmptyPasswordPipe",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSBrokerStateContainsHostAndUsernameOnly",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestHTTPSBrokerPipeSurvivesFetchAndAskpassExecOnEveryPlatform",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/buildrepo",
    "Test": "TestPrivateHTTPSBrokerAuthenticatesRealGitRepository",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/username_prompt",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/password_prompt",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/missing_secret_handle",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/legacy_env_secret_is_not_answered",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/plain_binary_answers_no_prompt",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/foreign_host",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/foreign_user",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/bare_prompt",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/no_arguments",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/extra_argument",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/missing_state",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/relative_state",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch/malformed_state",
    "Action": "pass"
  },
  {
    "Package": "github.com/relux-works/curator/internal/crossconformance",
    "Test": "TestDraftSourcesBrokerAskpassDispatch",
    "Action": "pass"
  }
]
```
