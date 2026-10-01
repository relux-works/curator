# BUG-261001-2772iz — recovery validation evidence

All exit codes below were observed from the command processes in RUN-261001-8b3ce4. Logs used direct stdout/stderr redirection, without pipes. The pre-change fixture failure log has its dummy expected secret redacted; other output is unchanged. The original-transport mutant is an expected failing gate, not a passing gate. Initial Linux full-package failure is an environment failure and is retained as exit 1.

## focused-darwin.log

Command:

```sh
go test ./internal/buildrepo -run '^TestHTTPS' -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	0.937s

```

## repro-linux-harness.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test ./internal/buildrepo -run '^TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password$' -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	0.020s

```

## repro-linux-harness-x100.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test ./internal/buildrepo -run '^TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password$' -count=100
```

Real exit code: 1.

```text
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.01s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:146: test HTTPS broker transport: write https-askpass-server: broken pipe
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.01s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:146: test HTTPS broker transport: write https-askpass-server: broken pipe
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
--- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts (0.00s)
    --- FAIL: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password (0.00s)
        httpsbroker_test.go:148: code=1 output="", want code=0 output="<fixture-secret-redacted>\n"
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	0.295s
FAIL

```

## fixed-linux-harness-x100.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test ./internal/buildrepo -run '^TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password$' -count=100
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	3.147s

```

## fixed-focused-darwin.log

Command:

```sh
go test ./internal/buildrepo -run '^TestHTTPS' -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	3.493s

```

## mutant-darwin.log

Command:

```sh
go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$' -count=1 # original transport source
```

Real exit code: 1.

```text
--- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe (0.33s)
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/foreign_user (0.02s)
        httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/bare_prompt (0.02s)
        httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/no_arguments (0.02s)
        httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/extra_argument (0.02s)
        httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
    --- FAIL: TestHTTPSBrokerSecretTransportChildExitsBeforeServe/username (0.02s)
        httpsbroker_pipe_unix_test.go:66: HTTPS broker secret transport after non-reading child exit: write |1: broken pipe
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	1.158s
FAIL

```

## restored-repro-darwin.log

Command:

```sh
go test ./internal/buildrepo -run '^TestHTTPSBrokerSecretTransportChildExitsBeforeServe$' -count=1 # restored fixed transport
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	0.708s

```

## race-crossconformance-darwin-x50.log

Command:

```sh
go test -race ./internal/crossconformance -run TestDraftSourcesBrokerAskpassDispatch -count=50
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/crossconformance	92.877s

```

## race-broker-packages-darwin.log

Command:

```sh
go test -race ./internal/buildrepo ./internal/testcli -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	284.240s
?   	github.com/relux-works/curator/internal/testcli	[no test files]

```

## race-broker-packages-linux.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test -race ./internal/buildrepo ./internal/testcli -count=1
```

Real exit code: 1.

```text
--- FAIL: TestPrivateHTTPSBrokerAuthenticatesRealGitRepository (0.14s)
    httpsbroker_test.go:340: private HTTPS ls-remote: exit status 128
        fatal: not a git repository: /Users/administrator/Developer/ReluxWorks/curator/curator/.git/worktrees/worktree51
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	41.802s
?   	github.com/relux-works/curator/internal/testcli	[no test files]
FAIL

```

## race-private-https-linux.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v /Users/administrator/Developer/ReluxWorks/curator/curator/.git:/Users/administrator/Developer/ReluxWorks/curator/curator/.git:ro -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test -race ./internal/buildrepo -run '^TestPrivateHTTPSBrokerAuthenticatesRealGitRepository$' -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	5.877s

```

## race-broker-packages-linux-mounted-git.log

Command:

```sh
docker run --rm -v "$PWD:/workspace" -v /Users/administrator/Developer/ReluxWorks/curator/curator/.git:/Users/administrator/Developer/ReluxWorks/curator/curator/.git:ro -v "$(go env GOMODCACHE):/go/pkg/mod" -v BUG-261001-2772iz-go-cache:/root/.cache/go-build -w /workspace golang:1.26.0 go test -race ./internal/buildrepo ./internal/testcli -count=1
```

Real exit code: 0.

```text
ok  	github.com/relux-works/curator/internal/buildrepo	45.437s
?   	github.com/relux-works/curator/internal/testcli	[no test files]

```

## lint.log

Command:

```sh
golangci-lint run
```

Real exit code: 0.

```text
0 issues.

```

## vet.log

Command:

```sh
go vet ./...
```

Real exit code: 0.

```text
(no output)
```

## build-darwin.log

Command:

```sh
go build -o /tmp/BUG-261001-2772iz-recovery/curator-darwin ./cmd/curator
```

Real exit code: 0.

```text
(no output)
```

## test-compile-windows.log

Command:

```sh
env GOOS=windows GOARCH=amd64 go test -c -o /tmp/BUG-261001-2772iz-recovery/buildrepo-windows.test.exe ./internal/buildrepo
```

Real exit code: 0.

```text
(no output)
```

## Deterministic Linux descriptor diagnostic

The first diagnostic attempt used a /tmp bind mount unavailable inside Colima: exit 1, `stat /evidence/double-wrap.go: no such file or directory`; Go did not execute. Reran by supplying the same source through stdin to the container, exit 0:

```sh
docker run --rm -i -v BUG-261001-2772iz-go-cache:/root/.cache/go-build golang:1.26.0 sh -c 'cat > /tmp/double-wrap.go; go run /tmp/double-wrap.go' < /tmp/BUG-261001-2772iz-recovery/double-wrap.go
```

Output:

```text
second wrapper read: n=0 err=read second-owner: resource temporarily unavailable
```

Diagnostic source (no secret material):

```go
package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil { panic(err) }
	defer syscall.Close(fds[1])
	if err := syscall.SetNonblock(fds[0], true); err != nil { panic(err) }
	owner := os.NewFile(uintptr(fds[0]), "owner")
	defer owner.Close()
	second := os.NewFile(owner.Fd(), "second-owner")
	var buf [1]byte
	// No server has written. An independently owned, pollable File would wait.
	n, err := second.Read(buf[:])
	fmt.Printf("second wrapper read: n=%d err=%v\n", n, err)
	if n != 0 || err == nil { os.Exit(1) }
}
```

## Standalone checks

`git diff --check`: exit 0, no output.

`git diff --exit-code -- LOGBOOK.md`: exit 0, no output.

`test -z "$(gofmt -l internal/buildrepo/httpsbroker_pipe_unix.go internal/buildrepo/httpsbroker_pipe_unix_test.go internal/buildrepo/httpsbroker_test.go internal/buildrepo/httpsbroker_test_pipe_unix_test.go internal/buildrepo/httpsbroker_test_pipe_windows_test.go)"`: exit 0, no output.

Restored-source `cmp`: exit 0, byte-identical to the saved fixed source. No mutant remains.
