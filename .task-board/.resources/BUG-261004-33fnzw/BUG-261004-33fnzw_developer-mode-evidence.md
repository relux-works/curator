# BUG-261004-33fnzw — unmanage restore preserves private regular-file modes

Developer handoff evidence, 2026-10-04. Repository changes remain uncommitted.

## Implementation and scope

`internal/envprofile/unmanage.go` now loads each supported backup file into a typed restore entry containing its lstat type, permission bits, and bytes. Restoration calls the existing `atomicManagedFile` writer: it stages in an owner-only temporary file, applies the saved mode, syncs, and replaces the destination entry. It does not unlink the live entry before staging or open through its final managed symlink. Parent-link refusal remains enforced. `switch.go` already preserves regular-file backup permission bits, so it needed no change. Backup symlink restoration remains the following N4 task.

New CLI regression invokes the production `run()` entry point: profile install, removal of the initially activated managed context, creation of synthetic operator context, profile use --takeover --env claude_code, then env unmanage --restore-backups --env claude_code. It checks regular type and exact mode on both the takeover backup and restored file, plus restored bytes. Unix cases: 0600, 0644, 0751 executable, and 0400 read-only. New envprofile tests inspect the actual planned entry metadata and drive production Unmanage to prove a linked parent refuses before changing external context. Three platform-ledger rows require these tests on Linux and macOS.

## Direct local command evidence

All commands ran directly, without tee or pipelines. Go test commands used GOFLAGS=-work and -count=1.

| Command | Real exit | Result |
| --- | --- | --- |
| `GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanageTakeoverPreservesFileMode$' -count=1 -v` before production fix | 1 | Expected red: 0600, 0751, 0400 became 0644; 0644 control passed. |
| Same command after production fix | 0 | 4/4 mode subtests passed. |
| Same command with only the production restore mode argument mutated to hard-coded 0644 | 1 | Mutant killed: 0600, 0751, 0400 failed; 0644 control passed. |
| Restore corrected source from saved copy, then `cmp` against that copy | 0 | Exact corrected source restored before subsequent validation. |
| `GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanage' -count=1 -v` | 0 | 5 top-level tests passed, including 4/4 new mode cases. `TestEnvUnmanageBackupRecordVectors` skipped because CURATOR_CONFORMANCE_ROOT is unset. |
| `GOFLAGS=-work go test ./internal/envprofile -run '^(TestUnmanageRestore\|TestUseTakeoverBacksUpAndNotifies$\|TestUseReplacesWithBackup$)' -count=1 -v` | 0 | 4 selected tests passed: plan metadata, parent-link refusal, takeover backup, replacement backup. The backslashes in this Markdown table escape the regex alternation characters; the executed expression used ordinary `|` characters. |
| `GOFLAGS=-work golangci-lint run ./internal/envprofile ./cmd/curator` | 0 | 0 issues. |
| `GOFLAGS=-work go build -o /tmp/BUG-261004-33fnzw-curator ./cmd/curator` | 0 | Production CLI builds. |
| `gofmt -l internal/envprofile/unmanage.go internal/envprofile/unmanage_mode_unix_test.go cmd/curator/env_unmanage_mode_unix_test.go` | 0 | No unformatted files reported. |
| `git diff --check` | 0 | No whitespace errors. |

No already-attached test evidence was substituted for these local runs.

## Platform and hosted bounds

Local runtime: macOS. Unix permission-bit coverage is 4/4 selected regular-file modes, verified at the CLI entry point. Linux runtime and Windows runtime were not run locally. The Unix tests are compiled for Unix and required on Linux/macOS in the CI platform ledger.

Windows FileMode does not represent DACLs. The reused atomic writer obtains its temporary file from `privatedir.CreateTemp`, whose Windows implementation attaches an owner-only protected DACL at creation; applying FileMode changes file attributes, not the original ACL. Exact original Windows ACL reconstruction is not provided or claimed. Windows ACL runtime behavior was not tested on this macOS host and remains subject to the hosted Windows gate. Permission modes cannot establish that ACL claim.

The full local suite, published backup-record vectors, race gate, full platform ledger validation, and hosted Change Request gate were not run. The binding produce-mode instruction prohibits the full local suite and leaves hosted validation as the arbiter. Hosted gate status is pending, not green. The producer checklist is scoped accordingly. Findings are recorded in task notes and this outcome; LOGBOOK.md and CHANGELOG.md are untouched as instructed.
