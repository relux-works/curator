# TASK-260910-3ungjy independent review transcripts

See verdict for commands and exit codes. Go tests use the campaign conformance root; bash pipefail enabled.

## base

```text
build_exit=0
vet_exit=0
gofmt_exit=0
ok  	github.com/relux-works/curator/internal/hookapproval	2.590s
ok  	github.com/relux-works/curator/internal/envfiles	1.471s
tests_exit=0

```

## command

```text
ok  	github.com/relux-works/curator/cmd/curator	52.638s
cmd_exit=0
ok  	github.com/relux-works/curator/internal/envprofile	7.068s
envprofile_exit=0

```

## shell

```text
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
--- PASS: TestShellHookTrustVectors (35.33s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced (0.44s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/sh (0.10s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/dash (0.12s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/bash (0.09s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-A-warning-sourced/zsh (0.10s)
    --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced (0.47s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/sh (0.13s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/dash (0.11s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/bash (0.09s)
        --- PASS: TestShellHookTrustVectors/approved-env-sh-B-enforcing-sourced/zsh (0.09s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning (0.35s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/sh (0.09s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/dash (0.07s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/bash (0.08s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-A-warning-sourced-with-warning/zsh (0.08s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced (0.79s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/sh (0.15s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/dash (0.17s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/bash (0.14s)
        --- PASS: TestShellHookTrustVectors/unapproved-env-sh-B-enforcing-not-sourced/zsh (0.25s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning (0.90s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/sh (0.17s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/dash (0.16s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/bash (0.19s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-A-warning-sourced-with-warning/zsh (0.21s)
    --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced (0.86s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/sh (0.22s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/dash (0.17s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/bash (0.20s)
        --- PASS: TestShellHookTrustVectors/changed-env-sh-B-enforcing-not-sourced/zsh (0.23s)
    --- PASS: TestShellHookTrustVectors/approved-env-ps1-A-warning-sourced (8.54s)
    --- PASS: TestShellHookTrustVectors/approved-env-ps1-B-enforcing-sourced (5.91s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-ps1-A-warning-sourced-with-warning (4.98s)
    --- PASS: TestShellHookTrustVectors/unapproved-env-ps1-B-enforcing-not-sourced (3.56s)
    --- PASS: TestShellHookTrustVectors/changed-env-ps1-A-warning-sourced-with-warning (4.51s)
    --- PASS: TestShellHookTrustVectors/changed-env-ps1-B-enforcing-not-sourced (3.25s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning (0.27s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/sh (0.08s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/dash (0.06s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/bash (0.06s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-A-warning-sourced-with-warning/zsh (0.07s)
    --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced (0.49s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/sh (0.12s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/dash (0.10s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/bash (0.13s)
        --- PASS: TestShellHookTrustVectors/forged-project-record-env-sh-B-enforcing-not-sourced/zsh (0.14s)
=== RUN   TestShellHookRefusesHostileCheckout
=== RUN   TestShellHookRefusesHostileCheckout/posix
=== RUN   TestShellHookRefusesHostileCheckout/posix/sh
=== RUN   TestShellHookRefusesHostileCheckout/posix/dash
=== RUN   TestShellHookRefusesHostileCheckout/posix/bash
=== RUN   TestShellHookRefusesHostileCheckout/posix/zsh
=== RUN   TestShellHookRefusesHostileCheckout/powershell
--- PASS: TestShellHookRefusesHostileCheckout (5.63s)
    --- PASS: TestShellHookRefusesHostileCheckout/posix (0.89s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/sh (0.24s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/dash (0.17s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/bash (0.25s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/zsh (0.24s)
    --- PASS: TestShellHookRefusesHostileCheckout/powershell (4.74s)
=== RUN   TestGeneratedHookParsesUnderPOSIXShells
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/zsh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/sh
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/dash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/bash
=== RUN   TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/zsh
--- PASS: TestGeneratedHookParsesUnderPOSIXShells (0.35s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-no/zsh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/A-warning/global-yes/zsh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-no/zsh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/bash/B-enforcing/global-yes/zsh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-no/zsh (0.02s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/A-warning/global-yes/zsh (0.02s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-no/zsh (0.02s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/sh (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/dash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/bash (0.01s)
    --- PASS: TestGeneratedHookParsesUnderPOSIXShells/zsh/B-enforcing/global-yes/zsh (0.02s)
=== RUN   TestHookEmbedsClosedRecordGrammar
--- PASS: TestHookEmbedsClosedRecordGrammar (0.00s)
=== RUN   TestShellHookRejectsMalformedRecords
=== RUN   TestShellHookRejectsMalformedRecords/two-fields
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/posix
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/two-fields/powershell
=== RUN   TestShellHookRejectsMalformedRecords/five-fields
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/posix
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/five-fields/powershell
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/posix
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/forged-approver/powershell
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/posix
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/empty-approver/powershell
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/posix
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-approver/powershell
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/posix
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/uppercase-digest/powershell
=== RUN   TestShellHookRejectsMalformedRecords/short-digest
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/posix
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/short-digest/powershell
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/posix
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/nonhex-digest/powershell
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/posix
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/bad-timestamp/powershell
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/posix
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/empty-timestamp/powershell
=== RUN   TestShellHookRejectsMalformedRecords/month-13
=== RUN   TestShellHookRejectsMalformedRecords/month-13/posix
=== RUN   TestShellHookRejectsMalformedRecords/month-13/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/month-13/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/month-13/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/month-13/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/month-13/powershell
=== RUN   TestShellHookRejectsMalformedRecords/february-30
=== RUN   TestShellHookRejectsMalformedRecords/february-30/posix
=== RUN   TestShellHookRejectsMalformedRecords/february-30/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/february-30/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/february-30/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/february-30/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/february-30/powershell
=== RUN   TestShellHookRejectsMalformedRecords/second-60
=== RUN   TestShellHookRejectsMalformedRecords/second-60/posix
=== RUN   TestShellHookRejectsMalformedRecords/second-60/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/second-60/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/second-60/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/second-60/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/second-60/powershell
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/sh
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/dash
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/bash
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/zsh
=== RUN   TestShellHookRejectsMalformedRecords/lowercase-timestamp/powershell
--- PASS: TestShellHookRejectsMalformedRecords (102.72s)
    --- PASS: TestShellHookRejectsMalformedRecords/two-fields (5.36s)
        --- PASS: TestShellHookRejectsMalformedRecords/two-fields/posix (1.00s)
            --- PASS: TestShellHookRejectsMalformedRecords/two-fields/posix/sh (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/two-fields/posix/dash (0.24s)
            --- PASS: TestShellHookRejectsMalformedRecords/two-fields/posix/bash (0.26s)
            --- PASS: TestShellHookRejectsMalformedRecords/two-fields/posix/zsh (0.24s)
        --- PASS: TestShellHookRejectsMalformedRecords/two-fields/powershell (4.36s)
    --- PASS: TestShellHookRejectsMalformedRecords/five-fields (6.77s)
        --- PASS: TestShellHookRejectsMalformedRecords/five-fields/posix (1.01s)
            --- PASS: TestShellHookRejectsMalformedRecords/five-fields/posix/sh (0.25s)
            --- PASS: TestShellHookRejectsMalformedRecords/five-fields/posix/dash (0.22s)
            --- PASS: TestShellHookRejectsMalformedRecords/five-fields/posix/bash (0.22s)
            --- PASS: TestShellHookRejectsMalformedRecords/five-fields/posix/zsh (0.31s)
        --- PASS: TestShellHookRejectsMalformedRecords/five-fields/powershell (5.76s)
    --- PASS: TestShellHookRejectsMalformedRecords/forged-approver (7.41s)
        --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/posix (1.49s)
            --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/posix/sh (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/posix/dash (0.42s)
            --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/posix/bash (0.35s)
            --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/posix/zsh (0.46s)
        --- PASS: TestShellHookRejectsMalformedRecords/forged-approver/powershell (5.92s)
    --- PASS: TestShellHookRejectsMalformedRecords/empty-approver (6.53s)
        --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/posix (1.68s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/posix/sh (0.30s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/posix/dash (0.35s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/posix/bash (0.47s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/posix/zsh (0.55s)
        --- PASS: TestShellHookRejectsMalformedRecords/empty-approver/powershell (4.85s)
    --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver (8.73s)
        --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/posix (1.39s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/posix/sh (0.44s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/posix/dash (0.35s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/posix/bash (0.29s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/posix/zsh (0.32s)
        --- PASS: TestShellHookRejectsMalformedRecords/uppercase-approver/powershell (7.34s)
    --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest (5.20s)
        --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/posix (1.14s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/posix/sh (0.31s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/posix/dash (0.26s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/posix/bash (0.33s)
            --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/posix/zsh (0.25s)
        --- PASS: TestShellHookRejectsMalformedRecords/uppercase-digest/powershell (4.06s)
    --- PASS: TestShellHookRejectsMalformedRecords/short-digest (6.02s)
        --- PASS: TestShellHookRejectsMalformedRecords/short-digest/posix (0.92s)
            --- PASS: TestShellHookRejectsMalformedRecords/short-digest/posix/sh (0.24s)
            --- PASS: TestShellHookRejectsMalformedRecords/short-digest/posix/dash (0.25s)
            --- PASS: TestShellHookRejectsMalformedRecords/short-digest/posix/bash (0.24s)
            --- PASS: TestShellHookRejectsMalformedRecords/short-digest/posix/zsh (0.19s)
        --- PASS: TestShellHookRejectsMalformedRecords/short-digest/powershell (5.10s)
    --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest (6.36s)
        --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/posix (0.92s)
            --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/posix/sh (0.23s)
            --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/posix/dash (0.22s)
            --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/posix/bash (0.28s)
            --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/posix/zsh (0.19s)
        --- PASS: TestShellHookRejectsMalformedRecords/nonhex-digest/powershell (5.44s)
    --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp (8.89s)
        --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/posix (2.22s)
            --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/posix/sh (0.58s)
            --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/posix/dash (0.54s)
            --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/posix/bash (0.44s)
            --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/posix/zsh (0.66s)
        --- PASS: TestShellHookRejectsMalformedRecords/bad-timestamp/powershell (6.67s)
    --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp (7.28s)
        --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/posix (1.07s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/posix/sh (0.25s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/posix/dash (0.23s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/posix/bash (0.30s)
            --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/posix/zsh (0.29s)
        --- PASS: TestShellHookRejectsMalformedRecords/empty-timestamp/powershell (6.21s)
    --- PASS: TestShellHookRejectsMalformedRecords/month-13 (8.56s)
        --- PASS: TestShellHookRejectsMalformedRecords/month-13/posix (1.32s)
            --- PASS: TestShellHookRejectsMalformedRecords/month-13/posix/sh (0.38s)
            --- PASS: TestShellHookRejectsMalformedRecords/month-13/posix/dash (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/month-13/posix/bash (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/month-13/posix/zsh (0.39s)
        --- PASS: TestShellHookRejectsMalformedRecords/month-13/powershell (7.24s)
    --- PASS: TestShellHookRejectsMalformedRecords/february-30 (9.12s)
        --- PASS: TestShellHookRejectsMalformedRecords/february-30/posix (1.37s)
            --- PASS: TestShellHookRejectsMalformedRecords/february-30/posix/sh (0.36s)
            --- PASS: TestShellHookRejectsMalformedRecords/february-30/posix/dash (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/february-30/posix/bash (0.30s)
            --- PASS: TestShellHookRejectsMalformedRecords/february-30/posix/zsh (0.44s)
        --- PASS: TestShellHookRejectsMalformedRecords/february-30/powershell (7.75s)
    --- PASS: TestShellHookRejectsMalformedRecords/second-60 (6.16s)
        --- PASS: TestShellHookRejectsMalformedRecords/second-60/posix (1.26s)
            --- PASS: TestShellHookRejectsMalformedRecords/second-60/posix/sh (0.37s)
            --- PASS: TestShellHookRejectsMalformedRecords/second-60/posix/dash (0.27s)
            --- PASS: TestShellHookRejectsMalformedRecords/second-60/posix/bash (0.30s)
            --- PASS: TestShellHookRejectsMalformedRecords/second-60/posix/zsh (0.32s)
        --- PASS: TestShellHookRejectsMalformedRecords/second-60/powershell (4.90s)
    --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp (10.33s)
        --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix (1.26s)
            --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/sh (0.38s)
            --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/dash (0.22s)
            --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/bash (0.34s)
            --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/posix/zsh (0.32s)
        --- PASS: TestShellHookRejectsMalformedRecords/lowercase-timestamp/powershell (9.07s)
=== RUN   TestShellHookTrustResolvesSymlinkedProject
=== RUN   TestShellHookTrustResolvesSymlinkedProject/posix
=== RUN   TestShellHookTrustResolvesSymlinkedProject/posix/sh
=== RUN   TestShellHookTrustResolvesSymlinkedProject/posix/dash
=== RUN   TestShellHookTrustResolvesSymlinkedProject/posix/bash
=== RUN   TestShellHookTrustResolvesSymlinkedProject/posix/zsh
=== RUN   TestShellHookTrustResolvesSymlinkedProject/powershell
--- PASS: TestShellHookTrustResolvesSymlinkedProject (15.97s)
    --- PASS: TestShellHookTrustResolvesSymlinkedProject/posix (1.63s)
        --- PASS: TestShellHookTrustResolvesSymlinkedProject/posix/sh (0.37s)
        --- PASS: TestShellHookTrustResolvesSymlinkedProject/posix/dash (0.32s)
        --- PASS: TestShellHookTrustResolvesSymlinkedProject/posix/bash (0.37s)
        --- PASS: TestShellHookTrustResolvesSymlinkedProject/posix/zsh (0.56s)
    --- PASS: TestShellHookTrustResolvesSymlinkedProject/powershell (14.35s)
=== RUN   TestShellHookTrustNativeRecordAuthorizesMSYSSpelling
    shell_hook_trust_test.go:1070: native/MSYS spelling identity is exercised on Windows
--- SKIP: TestShellHookTrustNativeRecordAuthorizesMSYSSpelling (0.00s)
=== RUN   TestWindowsWarningNormalization
--- PASS: TestWindowsWarningNormalization (0.00s)
=== RUN   TestPosixHookEmitsWindowsIdentity
--- PASS: TestPosixHookEmitsWindowsIdentity (0.00s)
=== RUN   TestDetectIsCrossPlatform
=== RUN   TestDetectIsCrossPlatform/zsh
=== RUN   TestDetectIsCrossPlatform/bash
=== RUN   TestDetectIsCrossPlatform/git_bash
=== RUN   TestDetectIsCrossPlatform/windows
=== RUN   TestDetectIsCrossPlatform/PowerShell_environment
=== RUN   TestDetectIsCrossPlatform/portable_fallback
--- PASS: TestDetectIsCrossPlatform (0.00s)
    --- PASS: TestDetectIsCrossPlatform/zsh (0.00s)
    --- PASS: TestDetectIsCrossPlatform/bash (0.00s)
    --- PASS: TestDetectIsCrossPlatform/git_bash (0.00s)
    --- PASS: TestDetectIsCrossPlatform/windows (0.00s)
    --- PASS: TestDetectIsCrossPlatform/PowerShell_environment (0.00s)
    --- PASS: TestDetectIsCrossPlatform/portable_fallback (0.00s)
=== RUN   TestPosixHookEntersNestedSwitchesAndLeavesProjects
=== RUN   TestPosixHookEntersNestedSwitchesAndLeavesProjects/bash
=== RUN   TestPosixHookEntersNestedSwitchesAndLeavesProjects/zsh
--- PASS: TestPosixHookEntersNestedSwitchesAndLeavesProjects (1.24s)
    --- PASS: TestPosixHookEntersNestedSwitchesAndLeavesProjects/bash (0.55s)
    --- PASS: TestPosixHookEntersNestedSwitchesAndLeavesProjects/zsh (0.69s)
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/#00
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/.
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/relative-path
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/#00
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/.
=== RUN   TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/relative-path
--- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging (0.16s)
    --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash (0.05s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/#00 (0.01s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/. (0.01s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/bash/relative-path (0.01s)
    --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh (0.11s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/#00 (0.01s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/. (0.01s)
        --- PASS: TestPosixHookRejectsNonAbsolutePWDWithoutHanging/zsh/relative-path (0.01s)
=== RUN   TestPosixHookSupportsNounsetProfiles
=== RUN   TestPosixHookSupportsNounsetProfiles/bash
=== RUN   TestPosixHookSupportsNounsetProfiles/zsh
--- PASS: TestPosixHookSupportsNounsetProfiles (0.22s)
    --- PASS: TestPosixHookSupportsNounsetProfiles/bash (0.15s)
    --- PASS: TestPosixHookSupportsNounsetProfiles/zsh (0.07s)
=== RUN   TestBashHookDoesNotDuplicatePromptCommand
--- PASS: TestBashHookDoesNotDuplicatePromptCommand (0.05s)
=== RUN   TestBashHookPreservesPromptCommandArray
--- PASS: TestBashHookPreservesPromptCommandArray (0.08s)
=== RUN   TestBashHookToleratesEmptyPromptCommandArrayUnderNounset
--- PASS: TestBashHookToleratesEmptyPromptCommandArrayUnderNounset (0.15s)
=== RUN   TestZshHookIsIdempotentAndDoesNotReenter
--- PASS: TestZshHookIsIdempotentAndDoesNotReenter (0.29s)
=== RUN   TestPosixHookDoesNotReenterWhileSourcingGlobalEnv
=== RUN   TestPosixHookDoesNotReenterWhileSourcingGlobalEnv/bash
=== RUN   TestPosixHookDoesNotReenterWhileSourcingGlobalEnv/zsh
--- PASS: TestPosixHookDoesNotReenterWhileSourcingGlobalEnv (0.14s)
    --- PASS: TestPosixHookDoesNotReenterWhileSourcingGlobalEnv/bash (0.05s)
    --- PASS: TestPosixHookDoesNotReenterWhileSourcingGlobalEnv/zsh (0.09s)
=== RUN   TestPosixHookNormalizesGitBashConfigAndAvoidsDirname
--- PASS: TestPosixHookNormalizesGitBashConfigAndAvoidsDirname (0.11s)
=== RUN   TestInstallHookAndSourceCommand
--- PASS: TestInstallHookAndSourceCommand (0.11s)
=== RUN   TestHookVariants
--- PASS: TestHookVariants (0.00s)
=== RUN   TestHookTrustGateShape
--- PASS: TestHookTrustGateShape (0.00s)
=== RUN   TestPowerShellHookRunsOnEveryPrompt
    shell_test.go:415: PowerShell prompt integration is exercised on Windows
--- SKIP: TestPowerShellHookRunsOnEveryPrompt (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/shell	163.413s
shell_exit=0
0 issues.
lint_exit=0

```

## e2e

```text
sh approve exit 0 approved /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh 
sh activation after approve exit 0 stdout 'marker=yes\n' stderr ''
sh revoke exit 0 revoked /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh 
sh activation after revoke exit 0 stdout 'marker=yes\n' stderr 'curator: shell_hook_env_unapproved: /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh is not approved; run curator hook approve /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh to approve it\n'
bash approve exit 0 approved /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh 
bash activation after approve exit 0 stdout 'marker=yes\n' stderr ''
bash revoke exit 0 revoked /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh 
bash activation after revoke exit 0 stdout 'marker=yes\n' stderr 'curator: shell_hook_env_unapproved: /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh is not approved; run curator hook approve /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.sh to approve it\n'
powershell approve exit 0 approved /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.ps1 
powershell activation after approve exit 0 stdout 'marker=yes\n' stderr ''
powershell revoke exit 0 revoked /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.ps1 
powershell activation after revoke exit 0 stdout 'marker=yes\n' stderr 'curator: shell_hook_env_unapproved: /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.ps1 is not approved; run curator hook approve /private/tmp/TASK-260910-3ungjy-e2e-2sk3a9of/project/.agents/env.ps1 to approve it\n'
E2E passed 6/6 activation checks

```

## probes

```text
=== RUN   TestReviewRecordedMissingAndUnreadable
=== RUN   TestReviewRecordedMissingAndUnreadable/missing
    review_probe_test.go:18: [status app --check] exit=0 stdout= stderr=
    review_probe_test.go:19: recorded path omitted by [status app --check]
    review_probe_test.go:20: recorded approved_by omitted by [status app --check]
    review_probe_test.go:18: [status app --check --json] exit=0 stdout={
          "alias": "app",
          "path": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadablemissing748726226/003",
          "shell_hook_trust": null,
          "skills": {}
        }
         stderr=
    review_probe_test.go:19: recorded path omitted by [status app --check --json]
    review_probe_test.go:20: recorded approved_by omitted by [status app --check --json]
    review_probe_test.go:18: [env status --check] exit=1 stdout=default claude_code: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default codex_cli: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default opencode: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default pi: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        tool claude_code: recorded 2.1.261 detected 2.1.274
        tool codex_cli: recorded 0.153.2 detected 0.153.4
        tool opencode: recorded unrecorded detected unknown
        tool pi: recorded 0.84.2 detected 0.84.2
        target xcode-coding-assistant (claude_code): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        target xcode-coding-assistant (codex_cli): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        profile default: lock sha256:726310f80f44…, precedence winner=higher-weight placement=winner-last
          member context default weight 0
        note: opencode skills come from the machine-current profile, split-brain by construction (§7.1)
         stderr=
    review_probe_test.go:19: recorded path omitted by [env status --check]
    review_probe_test.go:20: recorded approved_by omitted by [env status --check]
    review_probe_test.go:18: [env status --check --json] exit=1 stdout={
          "homes": [
            {
              "profile": "default",
              "environment": "claude_code",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadablemissing748726226/001/environments/default/claude_code",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "codex_cli",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadablemissing748726226/001/environments/default/codex_cli",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "opencode",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadablemissing748726226/001/environments/default/opencode/opencode",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "pi",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadablemissing748726226/001/environments/default/pi",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            }
          ],
          "scopes": null,
          "adapters": [
            {
              "id": "claude_code",
              "recorded": "2.1.261",
              "detected": "2.1.274",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "codex_cli",
              "recorded": "0.153.2",
              "detected": "0.153.4",
              "supported": [
                "monolithic"
              ]
            },
            {
              "id": "opencode",
              "recorded": "unrecorded",
              "detected": "unknown",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "pi",
              "recorded": "0.84.2",
              "detected": "0.84.2",
              "supported": [
                "monolithic"
              ]
            }
          ],
          "targets": [
            {
              "id": "xcode-coding-assistant",
              "adapter": "claude_code",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            },
            {
              "id": "xcode-coding-assistant",
              "adapter": "codex_cli",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            }
          ],
          "profiles": [
            {
              "profile": "default",
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "members": [
                {
                  "kind": "context",
                  "name": "default",
                  "weight": 0
                }
              ],
              "precedence": {
                "winner": "higher-weight",
                "placement": "winner-last"
              }
            }
          ],
          "unregistered_environments": null,
          "orphans": null,
          "notes": [
            "opencode skills come from the machine-current profile, split-brain by construction (§7.1)"
          ],
          "non_current": true,
          "shell_hook_trust": null
        }
         stderr=
    review_probe_test.go:19: recorded path omitted by [env status --check --json]
    review_probe_test.go:20: recorded approved_by omitted by [env status --check --json]
=== RUN   TestReviewRecordedMissingAndUnreadable/unreadable
    review_probe_test.go:18: [status app --check] exit=0 stdout=shell-hook-trust: /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh: shell_hook_env_unapproved (warning; run curator hook approve /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh)
         stderr=
    review_probe_test.go:20: recorded approved_by omitted by [status app --check]
    review_probe_test.go:18: [status app --check --json] exit=0 stdout={
          "alias": "app",
          "path": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003",
          "shell_hook_trust": [
            {
              "path": "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh",
              "state": "unapproved",
              "diagnostic": "shell_hook_env_unapproved"
            }
          ],
          "skills": {}
        }
         stderr=
    review_probe_test.go:20: recorded approved_by omitted by [status app --check --json]
    review_probe_test.go:18: [env status --check] exit=1 stdout=shell-hook-trust: /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh: shell_hook_env_unapproved (warning; run curator hook approve /private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh)
        default claude_code: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default codex_cli: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default opencode: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        default pi: non-current, unprovisioned, mode managed-home, form , lock sha256:726310f80f44…
          finding: home unprovisioned
        tool claude_code: recorded 2.1.261 detected 2.1.274
        tool codex_cli: recorded 0.153.2 detected 0.153.4
        tool opencode: recorded unrecorded detected unknown
        tool pi: recorded 0.84.2 detected 0.84.2
        target xcode-coding-assistant (claude_code): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        target xcode-coding-assistant (codex_cli): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        profile default: lock sha256:726310f80f44…, precedence winner=higher-weight placement=winner-last
          member context default weight 0
        note: opencode skills come from the machine-current profile, split-brain by construction (§7.1)
         stderr=
    review_probe_test.go:20: recorded approved_by omitted by [env status --check]
    review_probe_test.go:18: [env status --check --json] exit=1 stdout={
          "homes": [
            {
              "profile": "default",
              "environment": "claude_code",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/001/environments/default/claude_code",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "codex_cli",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/001/environments/default/codex_cli",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "opencode",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/001/environments/default/opencode/opencode",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "pi",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/001/environments/default/pi",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            }
          ],
          "scopes": null,
          "adapters": [
            {
              "id": "claude_code",
              "recorded": "2.1.261",
              "detected": "2.1.274",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "codex_cli",
              "recorded": "0.153.2",
              "detected": "0.153.4",
              "supported": [
                "monolithic"
              ]
            },
            {
              "id": "opencode",
              "recorded": "unrecorded",
              "detected": "unknown",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "pi",
              "recorded": "0.84.2",
              "detected": "0.84.2",
              "supported": [
                "monolithic"
              ]
            }
          ],
          "targets": [
            {
              "id": "xcode-coding-assistant",
              "adapter": "claude_code",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            },
            {
              "id": "xcode-coding-assistant",
              "adapter": "codex_cli",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            }
          ],
          "profiles": [
            {
              "profile": "default",
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "members": [
                {
                  "kind": "context",
                  "name": "default",
                  "weight": 0
                }
              ],
              "precedence": {
                "winner": "higher-weight",
                "placement": "winner-last"
              }
            }
          ],
          "unregistered_environments": null,
          "orphans": null,
          "notes": [
            "opencode skills come from the machine-current profile, split-brain by construction (§7.1)"
          ],
          "non_current": true,
          "shell_hook_trust": [
            {
              "path": "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewRecordedMissingAndUnreadableunreadable841847021/003/.agents/env.sh",
              "state": "unapproved",
              "diagnostic": "shell_hook_env_unapproved"
            }
          ]
        }
         stderr=
    review_probe_test.go:20: recorded approved_by omitted by [env status --check --json]
--- FAIL: TestReviewRecordedMissingAndUnreadable (17.57s)
    --- FAIL: TestReviewRecordedMissingAndUnreadable/missing (9.32s)
    --- FAIL: TestReviewRecordedMissingAndUnreadable/unreadable (8.25s)
=== RUN   TestReviewUnreadableStateJSON
    review_probe_test.go:31: [status app --check --json] exit=0 stdout={
          "alias": "app",
          "path": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/003",
          "shell_hook_trust": [
            {
              "path": "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/003/.agents/env.sh",
              "state": "unapproved",
              "diagnostic": "shell_hook_env_unapproved"
            }
          ],
          "skills": {}
        }
         stderr=
    review_probe_test.go:32: read error disappeared from [status app --check --json]
    review_probe_test.go:31: [env status --check --json] exit=1 stdout={
          "homes": [
            {
              "profile": "default",
              "environment": "claude_code",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/001/environments/default/claude_code",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "codex_cli",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/001/environments/default/codex_cli",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "opencode",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/001/environments/default/opencode/opencode",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "pi",
              "mode": "managed-home",
              "form": "",
              "home": "/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestReviewUnreadableStateJSON485876049/001/environments/default/pi",
              "provisioned": false,
              "current": false,
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "marker_hash": "",
              "surfaces": null,
              "passthrough": null,
              "seeds": null,
              "seed_links": null,
              "seeded_projects": null,
              "backups": 0,
              "backups_oldest": "",
              "backups_newest": "",
              "findings": [
                "home unprovisioned"
              ],
              "warnings": null
            }
          ],
          "scopes": null,
          "adapters": [
            {
              "id": "claude_code",
              "recorded": "2.1.261",
              "detected": "2.1.274",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "codex_cli",
              "recorded": "0.153.2",
              "detected": "0.153.4",
              "supported": [
                "monolithic"
              ]
            },
            {
              "id": "opencode",
              "recorded": "unrecorded",
              "detected": "unknown",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "pi",
              "recorded": "0.84.2",
              "detected": "0.84.2",
              "supported": [
                "monolithic"
              ]
            }
          ],
          "targets": [
            {
              "id": "xcode-coding-assistant",
              "adapter": "claude_code",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            },
            {
              "id": "xcode-coding-assistant",
              "adapter": "codex_cli",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            }
          ],
          "profiles": [
            {
              "profile": "default",
              "lock_hash": "sha256:726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19",
              "members": [
                {
                  "kind": "context",
                  "name": "default",
                  "weight": 0
                }
              ],
              "precedence": {
                "winner": "higher-weight",
                "placement": "winner-last"
              }
            }
          ],
          "unregistered_environments": null,
          "orphans": null,
          "notes": [
            "opencode skills come from the machine-current profile, split-brain by construction (§7.1)"
          ],
          "non_current": true,
          "shell_hook_trust": null
        }
         stderr=
    review_probe_test.go:32: read error disappeared from [env status --check --json]
--- FAIL: TestReviewUnreadableStateJSON (5.11s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	23.813s
FAIL

```

## mutants

```text
M1 reapproval only reuses old digest for operator records
--- FAIL: TestHookApproveReRecordsAfterChange (0.41s)
    hook_test.go:150: re-approval kept the stale digest
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.020s
FAIL
 exit=1
M2 changed-check only covers manager approvals
--- FAIL: TestStatusReportsShellHookTrustPosture (1.69s)
    hook_test.go:488: status --check over a changed file = 0, want 1
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	2.249s
FAIL
 exit=1
restored: ok  	github.com/relux-works/curator/cmd/curator	3.852s
 exit=0
2/2 narrowing mutants killed; byte-identical restore

```
