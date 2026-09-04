# BUG-260901-393lbo disposition

## Diagnosis

The `curator.test` process that drew attention (400% CPU, 430 MB RSS) was the task-board spawn-runner's post-agent validation suite for RUN-260901-c3351b, command 4 of 4 (`go test -count=1 -timeout 30m ./...`). It completed in 630 s, inside the 290 s to 771 s range of the eleven previous runs. Not an orphan, not a memory leak.

The real leak is on disk. `TMPDIR` held 4531 `curator-*` directories (36 MB):

| Prefix | Count | Last created | Cause |
| --- | ---: | --- | --- |
| curator-buildrepo-local- | 3800 | 2026-09-01 | AdmitLocal sealed store defeats RemoveAll; error discarded |
| curator-stubgo- | 488 | 2026-09-01 | godriver TestMain never removes the stub launcher directory |
| curator-install-private-commit- | 187 | 2026-08-28 | historical; not reproduced on current code |
| curator-install-private- | 51 | 2026-08-29 | historical; consistent with killed processes |

Proof for the first cause: `rm -rf` on a leaked tree fails with `Permission denied` on the 0500 object directories.

## Fix

- `internal/buildrepo/local.go`: `releasePrivateRoot` restores owner write on every directory beneath the private root and then removes it. `AdmitLocal` uses named results so a release failure on an otherwise successful admission returns `CodeLocalLayoutUnsafe` instead of being ignored. Sealing and proofs unchanged.
- `internal/godriver/main_test.go`: `stubOnce` records its scratch directory; `TestMain` removes it after `m.Run()`.
- `internal/buildrepo/local_release_test.go`: three regression tests (accepted admission, refused admission, sealed-store release including the negative `RemoveAll` guard).

## Evidence

- Commit 826b5128 on `bug/BUG-260901-393lbo`, SSH-signed, `git verify-commit` good.
- Worktree `.temp/BUG-260901-393lbo/worktree` based on `origin/main` 979fa36e.
- `.temp/BUG-260901-393lbo/logs/validation-02.log`: build, vet, `go test -count=1 -timeout 30m ./...` all green, 58 packages ok, exit 0.
- Leak counters across full package runs: buildrepo-local 3816 -> 3816, stubgo 490 -> 490.
- PR https://github.com/relux-works/curator/pull/52

## Not in scope

The 4531 existing directories in `TMPDIR` are not removed by this change; they are host state and can be cleared by hand (the sealed ones need `chmod -R u+w` first).
