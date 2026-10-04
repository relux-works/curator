# Review verdict: accepted

BUG-261004-33fnzw — unmanage-restore-widens-private-file-mode.
CR-BUG-261004-33fnzw-1 revision 1. Base ca1b776fb580ec0cee0173bf150daf063023aeaa; candidate tree 7d4cb634db059c154dade477f9d9985c9a23e8c8.

No blocking findings. Reviewed all four changed paths and the unchanged production backup, managed writer, and platform-private staging implementations. No product code modified by reviewer. No LOGBOOK.md or CHANGELOG.md edits. The platform ledger additions are directly related to these regressions.

## Hosted evidence and identity

https://github.com/relux-works/curator/actions/runs/37171121482 completed successfully. GitHub commit API identifies its head 9fad4f6f544f1d6da9922c56ddd451527e463ec2 with tree 7d4cb634db059c154dade477f9d9985c9a23e8c8: exact candidate match, not an unrelated green run. Downloaded the run artifacts and parsed each lane's go-test.json. Test jobs on Linux, macOS, Windows; race jobs on Linux and macOS; lint; interop; naming; and gate self-tests succeeded. Candidate-suite and optional self-hosted jobs were skipped.

All three new top-level regressions passed in all four Unix test/race lanes (12/12 expected executions). The CLI's 0600, 0644, 0751, 0400 subcases passed in every Unix lane (16/16). The two touched packages have terminal package pass events in all five lanes (10/10), with no fail events in the downloaded JSON. Existing five CLI restore controls passed in all five lanes (25/25): BackupRecordVectors, DoesNotRestoreWithoutAnObservedRecord, RestoresNewestBackupGeneration, RestoresRecordedTakeoverBackup, LeavesBackupsWithoutRestoreFlag (all prefixed TestEnvUnmanage).

The workflow pins curator-spec rc.14 to 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. TestEnvironmentsReadFailureVectors, TestEnvironmentWriteNofollowVectors, path-kind, store-boundary, passthrough, and system-delta conformance top-level tests passed on all five lanes. Linux suite plan reports served=82, deferred=0, excluded=1 (godriver, outside this change). This is not a claim that every subcase executes on every OS: scoped cmd/curator plus internal/envprofile pass/skip event counts were Linux 1310/15, macOS 1354/1, Windows 1268/68, with race matching each Unix lane. macOS's envprofile skip is TestFoldedPathsAreSourceInvalid; Windows skips include permission-unreadability fixtures. New Unix regressions are intentionally not compiled on Windows.

Normal test JSON SHA-256:
- Linux: 22e1c854251d39b3fb51397301da687963ccb3a86ad6fadfc55736965cec25a4
- macOS: 6de2e7352f6f2463c8f6386960425a74d4ee13ec86cdcfcfecf3ed78d199c6a3
- Windows: 9060f74190372dd07003844ce3e0e5c2f88697a54eb86bfc388ce8b94b629989

## Swept surfaces / acceptance criteria

| Surface | Assessment |
| --- | --- |
| CLI production wiring | runProfile calls run; regression drives profile install, initial unmanage to remove auto-activation, profile use --takeover, then env unmanage --restore-backups --env claude_code. It asserts command success, backup and restored regular-file modes, retained backup mode, and original bytes. |
| Backup metadata / AC1 | Existing copyManagedEntry already saves regular-file permission bits. Restore entries now retain lstat type, permission bits, and copied bytes. Nonregular entries fail closed; symlink support remains the separate N4 task. |
| Atomic restore / AC1 | applyUnmanagePlan calls atomicManagedFile with saved mode. Unix temp starts 0600, is chmodded to saved bits before payload write, synced, closed and renamed in the same directory. Final symlinks are replaced rather than followed. Parent checks remain active. Supported Unix regular-file permissions are not widened. |
| Modes / AC2 | All required 0600, 0644 and executable 0751 rows plus 0400 pass at the CLI entry point in hosted Unix lanes. |
| Windows / AC3 | Explicitly bounded: FileMode is not a Windows DACL. The existing CreateTemp backend supplies a protected owner-only DACL at creation, rather than reproducing the original ACL. Windows applicable package tests pass, but the new tests do not attest original DACL preservation. Existing Windows replacement fallback may remove a link before retrying rename; no new universal Windows atomicity guarantee is inferred. |
| Compatibility | Existing restore controls and rc.14 conformance entry points remain green; no unrelated changes or weakened assertions in the diff. |

## Red-first and mutation reasoning (not executed)

Binding HOSTED-EVIDENCE MODE prohibits local Go build/test/vet/lint and replaces red/mutant reruns with reading. No such commands were run by this reviewer.

Base applyUnmanagePlan removes the managed entry and writes payload with 0644, losing saved mode. The new CLI exact-mode assertion therefore catches widening for 0600 and 0400 and loss of execute bits for 0751 under the ordinary hosted umask; 0644 is the unchanged control. The plan metadata test is candidate-structure-specific and cannot simply compile against the old map of bytes; the CLI regression alone independently exposes the base defect. The linked-parent case guards an existing safety property and is not represented as a newly failing base test.

Required mutant: pass hard-coded 0644 to the restore writer. Expected killed by exact restored-mode assertions for 0600, 0751 and 0400 (0644 remains the control).

Reviewer-added mutant: narrow permissions to entry.mode & 0600 when restoring, a likely overcorrection that keeps private files safe but loses supported modes. Expected killed by 0644 and 0751 equality assertions; 0600 and 0400 remain controls. Both mutant conclusions are static reasoning, not measured mutation runs.

## Verification boundary and lifecycle

Local read-only git diff, GitHub API/run queries, artifact download, JSON parsing, and git diff --check completed with exit 0. No local test exit codes are claimed; green execution evidence comes from the exact hosted run. No pending hosted validation for this candidate remains; original Windows ACL reproduction remains outside the stated bound. Queried spawn goal: this run is not goal-bound. Acceptance routes to integrating via accept_cr, never done; producer integration remains outstanding. The non-acceptance checklist branch is not applicable because this verdict accepts the revision.
