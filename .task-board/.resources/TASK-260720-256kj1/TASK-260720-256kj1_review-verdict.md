# Review verdict: changes requested

Reviewed detached worktree at base 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8. Production code was not modified by the reviewer.

## Passing evidence

- Candidate build-source conformance plus focused snapshot and legacy hashing tests pass.
- Focused race tests pass.
- make check passes, including go vet, full go test, and gofmt verification.
- Windows amd64 compile checks for internal/buildsource and internal/snapshot pass.
- Exact framing, unsigned UTF-8 ordering, root marker inclusion, legacy marker-excluding hash compatibility, invalid path/link/special-file rejection, and frozen-token mutation checks match the task acceptance criteria.

## Required change

Snapshot repair is not atomic and is unsafe under concurrent repair. internal/snapshot/snapshot.go lines 100-103 first rename the live target to a backup and only afterward rename the staged directory into place. The live path is absent between those operations. Concurrent repairers can race across that gap; the winner fallback at lines 62-72 cannot recover reliably while another repair is between renames.

A black-box probe using only snapshot.Get ran 100 tampered-cache rounds with 12 concurrent repairers. Result: get_errors=19 and missing_target_observations=7198. Observed errors included ENOENT renaming the target, validation of a missing published snapshot, directory-not-empty publication, and failed restore because another repair had recreated the target. This directly fails the AC requiring corrupt or incomplete commit-keyed snapshots to be recreated atomically.

The existing concurrency test covers only a cold miss, not concurrent replacement of a present tampered cache entry.

## Rework acceptance

Implement a cross-platform publication/coordination strategy that makes concurrent repair of a present invalid snapshot deterministic: no caller errors, no observable missing or partial live target, and no backup/staging leak. Add a regression test that starts from a present tampered snapshot, runs concurrent Get/GetValidated repairers, proves every result validates against the repository commit, and exercises the race path under go test -race. Preserve the passing build-source framing, marker compatibility, mutation, full-suite, and Windows gates.