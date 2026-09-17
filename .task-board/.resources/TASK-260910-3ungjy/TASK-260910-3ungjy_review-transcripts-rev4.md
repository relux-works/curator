# TASK-260910-3ungjy review transcripts, revision 4

Shell: bash; set -o pipefail. CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1. Frozen candidate 6987eaa1d028317b2d27e51b2b5fb00561a3fbf7.

## review-build.log

```text
$ go build ./... && go vet ./... && gofmt -l internal cmd; EXIT=0
```

## review-libraries.log

```text
$ go test -count=1 -timeout 8m ./internal/hookapproval/... ./internal/shell/... ./internal/envfiles/...; EXIT=0
ok  	github.com/relux-works/curator/internal/hookapproval	1.734s
ok  	github.com/relux-works/curator/internal/shell	41.190s
ok  	github.com/relux-works/curator/internal/envfiles	1.283s
```

## review-envprofile.log

```text
$ go test -count=1 -timeout 8m ./internal/envprofile/ -run TestStatus; EXIT=0
ok  	github.com/relux-works/curator/internal/envprofile	29.880s
```

## review-lint.log

```text
$ golangci-lint run ./internal/hookapproval/... ./internal/envprofile/... ./cmd/curator/...; EXIT=0
0 issues.
```

## review-cli.log

```text
$ go test -count=1 -timeout 8m ./cmd/curator/ -run 'TestHook|TestStatusReportsShellHook|TestStatusJSONCarriesShellHook|TestStatusReportsRecordedPaths|TestStatusJSONKeepsTheLegacyShape|TestStatusRecordedButMissing|TestStatusUnreadable|TestEnvStatusReportsShellHook|TestEnvStatusMissingAndUnreadable|TestEnvStatusUnreadable'; EXIT=1 (lock acquisition)
acquire package host GOROOT test lock: context deadline exceeded
FAIL	github.com/relux-works/curator/cmd/curator	301.014s
FAIL
```

## review-cli-retry.log

```text
$ Same exact CLI command retried; EXIT=0
ok  	github.com/relux-works/curator/cmd/curator	256.322s
```

## review-vectors.log

```text
$ go test -count=1 -v -timeout 2m ./internal/shell -run '^TestShellHookTrustVectors$'; EXIT=0 (6 PowerShell capability skips)
=== RUN   TestShellHookTrustVectors
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/sh
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/dash
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/bash
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/zsh
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/sh
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/dash
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/bash
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/zsh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/zsh
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/zsh
=== RUN   TestShellHookTrustVectors/approved-env-ps1-A-warning-sourced
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/approved-env-ps1-B-enforcing-sourced
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/unapproved-env-ps1-A-warning-sourced-with-warning
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/unapproved-env-ps1-B-enforcing-not-sourced
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/changed-env-ps1-A-warning-sourced-with-warning
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/changed-env-ps1-B-enforcing-not-sourced
    shell_hook_trust_test.go:97: PowerShell is unavailable (no pwsh on this runner)
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/zsh
--- PASS: TestShellHookTrustVectors (4.84s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced (0.62s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/sh (0.11s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/dash (0.09s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/bash (0.15s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/zsh (0.14s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced (0.70s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/sh (0.11s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/dash (0.10s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/bash (0.09s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/zsh (0.13s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning (0.42s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/sh (0.09s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/dash (0.09s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/bash (0.12s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/zsh (0.10s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced (0.55s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/sh (0.11s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/dash (0.11s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/bash (0.15s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/zsh (0.18s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning (0.59s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/sh (0.12s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/dash (0.10s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/bash (0.15s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/zsh (0.13s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced (0.76s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/sh (0.17s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/dash (0.15s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/bash (0.17s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/zsh (0.15s)
    --- SKIP: TestShellHookTrustVectors/approved-env-ps1-A-warning-sourced (0.11s)
    --- SKIP: TestShellHookTrustVectors/approved-env-ps1-B-enforcing-sourced (0.10s)
    --- SKIP: TestShellHookTrustVectors/unapproved-env-ps1-A-warning-sourced-with-warning (0.00s)
    --- SKIP: TestShellHookTrustVectors/unapproved-env-ps1-B-enforcing-not-sourced (0.00s)
    --- SKIP: TestShellHookTrustVectors/changed-env-ps1-A-warning-sourced-with-warning (0.02s)
    --- SKIP: TestShellHookTrustVectors/changed-env-ps1-B-enforcing-not-sourced (0.02s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning (0.28s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/sh (0.08s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/dash (0.06s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/bash (0.07s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/zsh (0.07s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced (0.65s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/sh (0.13s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/dash (0.10s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/bash (0.16s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/zsh (0.20s)
PASS
ok  	github.com/relux-works/curator/internal/shell	5.407s
```

## review-vectors-pwsh.log

```text
$ PATH=/tmp/pwsh/app:$PATH go test -count=1 -v -timeout 3m ./internal/shell -run '^TestShellHookTrustVectors$'; EXIT=0 (14/14 cases, no skips)
=== RUN   TestShellHookTrustVectors
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/sh
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/dash
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/bash
=== RUN   TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/zsh
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/sh
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/dash
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/bash
=== RUN   TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/zsh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/zsh
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/zsh
=== RUN   TestShellHookTrustVectors/approved-env-ps1-A-warning-sourced
=== RUN   TestShellHookTrustVectors/approved-env-ps1-B-enforcing-sourced
=== RUN   TestShellHookTrustVectors/unapproved-env-ps1-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/unapproved-env-ps1-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/changed-env-ps1-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/changed-env-ps1-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/sh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/dash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/bash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/zsh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/sh
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/dash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/bash
=== RUN   TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/zsh
--- PASS: TestShellHookTrustVectors (61.77s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced (2.24s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/sh (0.38s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/dash (0.39s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/bash (0.50s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/zsh (0.34s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced (1.64s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/sh (0.48s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/dash (0.36s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/bash (0.31s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/zsh (0.17s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning (0.90s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/sh (0.15s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/dash (0.09s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/bash (0.18s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/zsh (0.22s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced (2.19s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/sh (0.45s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/dash (0.54s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/bash (0.37s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/zsh (0.38s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning (1.53s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/sh (0.32s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/dash (0.24s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/bash (0.43s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/zsh (0.19s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced (1.40s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/sh (0.25s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/dash (0.33s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/bash (0.46s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/zsh (0.18s)
    --- PASS: TestShellHookTrustVectors/approved-env-ps1-A-warning-sourced (11.35s)
    --- PASS: TestShellHookTrustVectors/approved-env-ps1-B-enforcing-sourced (10.72s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-ps1-A-warning-sourced-with-warning (4.39s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-ps1-B-enforcing-not-sourced (5.41s)
    --- PASS: TestShellHookTrustVectors/changed-env-ps1-A-warning-sourced-with-warning (6.09s)
    --- PASS: TestShellHookTrustVectors/changed-env-ps1-B-enforcing-not-sourced (12.20s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning (0.68s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/sh (0.13s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/dash (0.16s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/bash (0.18s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/zsh (0.14s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced (1.01s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/sh (0.26s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/dash (0.26s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/bash (0.22s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/zsh (0.15s)
PASS
ok  	github.com/relux-works/curator/internal/shell	62.257s
```

## review-e2e.log

```text
$ bash e2e-review.sh; EXIT=0
approved /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh
revoked /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh
curator: shell_hook_env_unapproved: /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh is not approved; run curator hook approve /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh to approve it
sh approve=silent revoke=warn sources=yes EXIT=0
approved /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh
revoked /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh
curator: shell_hook_env_unapproved: /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh is not approved; run curator hook approve /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.sh to approve it
bash approve=silent revoke=warn sources=yes EXIT=0
```

## Fresh binary E2E script

```bash
#!/bin/bash
set -euo pipefail
cd /tmp/curator-review4.ibq8gp
go build -o /tmp/curator-review4.ibq8gp/curator-review ./cmd/curator
mkdir -p e2e/project/.agents e2e/home
export CURATOR_CONFIG=/tmp/curator-review4.ibq8gp/e2e/home/config.json
export PROJ=/tmp/curator-review4.ibq8gp/e2e/project
export HOOK=/tmp/curator-review4.ibq8gp/e2e/hook
printf 'export REVIEW_MARKER=ok\n' > "$PROJ/.agents/env.sh"
./curator-review shell-init bash > "$HOOK"
for interpreter in sh bash; do
 ./curator-review hook approve "$PROJ/.agents/env.sh"
 "$interpreter" -c 'cd "$PROJ"; . "$HOOK"; test "$REVIEW_MARKER" = ok' > "e2e/$interpreter-approved.out" 2> "e2e/$interpreter-approved.err"
 test ! -s "e2e/$interpreter-approved.err"
 ./curator-review hook revoke "$PROJ/.agents/env.sh"
 "$interpreter" -c 'cd "$PROJ"; . "$HOOK"; test "$REVIEW_MARKER" = ok' > "e2e/$interpreter-revoked.out" 2> "e2e/$interpreter-revoked.err"
 rg shell_hook_env_unapproved "e2e/$interpreter-revoked.err"
 echo "$interpreter approve=silent revoke=warn sources=yes EXIT=0"
done
```

## PowerShell binary E2E

Used same fixture/config as script, wrote `$env:REVIEW_MARKER="ok"` to env.ps1; `curator-review shell-init powershell > hook.ps1`; real CLI approve then revoke. Each activation ran `/tmp/pwsh/app/pwsh -NoProfile -Command 'Set-Location $env:PROJ; . $env:HOOK; if ($env:REVIEW_MARKER -ne "ok") { exit 1 }'`.

```text
APPROVED_EXIT=0 SILENT_EXIT=0
REVOKED_EXIT=0 WARNING_EXIT=0
curator: shell_hook_env_unapproved: /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.ps1 is not approved; run curator hook approve /private/tmp/curator-review4.ibq8gp/e2e/project/.agents/env.ps1 to approve it
```

## Selected CLI tests (24)

```text
TestEnvStatusMissingAndUnreadableKeepRecord
TestEnvStatusReportsShellHookTrustPosture
TestEnvStatusUnreadableApprovalStateSurfaced
TestHookApprovalsListsReadOnly
TestHookApprovalsOnAbsentStateIsEmpty
TestHookApprovalsToleratesMalformedRecord
TestHookApprovalsUnreadableStateNamesReadFailure
TestHookApproveEndToEndWithGeneratedHook
TestHookApproveFailsDistinctlyOnAbsentAndUnreadable
TestHookApproveIsIdempotentWhenRecordMatches
TestHookApproveReRecordsAfterChange
TestHookApproveRecordsOperatorApproval
TestHookApproveResolvesSymlinkedOperand
TestHookDiagnosticsAgreeWithShell
TestHookRevokeAbsentLeavesStateByteIdentical
TestHookRevokeRemovesRecord
TestHookUsageErrors
TestStatusJSONCarriesShellHookTrust
TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands
TestStatusRecordedButMissingStaysInInventory
TestStatusReportsRecordedPathsBeyondTheCurrentProject
TestStatusReportsShellHookTrustPosture
TestStatusUnreadableApprovalStateSurfaced
TestStatusUnreadableCandidatesKeepRecord
```

## Narrowing mutants — 2/2 detected, 0 survivors

M1 retains missing-file handling only for manager records, dropping operator records. M2 retains unreadable-file record association only for manager records. These narrow record classes rather than removing the whole gate. Each runs committed production `run()` tests. Byte-identical original restored in finally and verified against frozen copy.

```python
import os, subprocess
from pathlib import Path
p=Path('internal/hookapproval/hookapproval.go')
original=p.read_bytes()
mutants=[
('missing', 'if !found {\n\t\t\t\t\tcontinue', 'if !found || record.ApprovedBy == ApprovedByOperator {\n\t\t\t\t\tcontinue', 'TestStatusRecordedButMissingStaysInInventory'),
('unreadable', 'if found {\n\t\treturn Posture{Path: canonical, State: PostureApproved', 'if found && record.ApprovedBy == ApprovedByManager {\n\t\treturn Posture{Path: canonical, State: PostureApproved', 'TestStatusUnreadableCandidatesKeepRecord'),
]
for name, before, after, test in mutants:
    text=original.decode()
    assert text.count(before)==1, (name, text.count(before))
    try:
        p.write_text(text.replace(before,after))
        result=subprocess.run(['go','test','-count=1','-timeout','8m','./cmd/curator/','-run','^'+test+'$'],stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
        Path('review-mutant-'+name+'.log').write_bytes(result.stdout+f'\nEXIT={result.returncode}\n'.encode())
        print(name,result.returncode,flush=True)
        assert result.returncode==1 and b'--- FAIL: '+test.encode() in result.stdout
    finally:
        p.write_bytes(original)
        assert p.read_bytes()==original
result=subprocess.run(['go','test','-count=1','-timeout','8m','./cmd/curator/','-run','TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord'],stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
Path('review-mutant-restored.log').write_bytes(result.stdout+f'\nEXIT={result.returncode}\n'.encode())
print('restored',result.returncode,flush=True)
assert result.returncode==0
```

### missing

```text
--- FAIL: TestStatusRecordedButMissingStaysInInventory (0.64s)
    hook_posture_test.go:107: status lacks "shell-hook-trust: /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestStatusRecordedButMissingStaysInInventory2031641019/003/.agents/env.sh":
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	282.918s
FAIL

EXIT=1
```

### unreadable

```text
--- FAIL: TestStatusUnreadableCandidatesKeepRecord (0.00s)
    --- FAIL: TestStatusUnreadableCandidatesKeepRecord/recorded (0.20s)
        hook_posture_test.go:168: status lacks "approved_by=operator":
            shell-hook-trust: /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestStatusUnreadableCandidatesKeepRecordrecorded4259535366/003/.agents/env.sh: shell_hook_env_unapproved (file is unreadable; make it readable, then run curator hook approve /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestStatusUnreadableCandidatesKeepRecordrecorded4259535366/003/.agents/env.sh)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.297s
FAIL

EXIT=1
```

### restored

```text
ok  	github.com/relux-works/curator/cmd/curator	3.314s

EXIT=0
```

The two earlier mutant attempts were unexecuted lock waits, not mutant detections: wrapper killed each at 180 seconds; stacks include TestMain → AcquireHostGOROOT → AcquireHomeOnly. The successful M1 run spent 282.918s including lock wait; the actual assertion failed in 0.64s.
