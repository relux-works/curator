# TASK-260811-3ksxig pnpm source-closure rework report

Date: 2026-08-23

## Outcome

Implemented all four required changes from reviewer run `RUN-260823-872c60`.

1. Declared patches now bind the normalized-LF SHA-256 manager hash used by the
   pinned pnpm v10 profile. Every patched package snapshot must carry the exact
   `(patch_hash=...)` context. Curator independently applies a closed unified
   diff to admitted tarball bytes, derives the expected file inventory, and
   records `pnpm-patch-transform-receipt-v1` evidence before materialization.
2. Materialization owns the complete virtual-store layout. Every selected
   snapshot maps to its pinned pnpm directory name; missing, malformed, or
   unclaimed entries fail closed. Package identity and full file inventory are
   reconciled, including the expected patched inventory.
3. Both `pnpm-lock.yaml` and `pnpm-workspace.yaml` require one YAML document and
   decoder EOF. Trailing documents/content return
   `closure_lock_format_unsupported`.
4. Each dependency/peer symlink is checked against the exact selected snapshot
   target. Two simultaneous peer contexts materialize successfully; missing or
   swapped peer links fail with `closure_graph_incomplete`.

Self-review also found and fixed a workspace selection defect: non-root
importer edges used `importer:<path>` while traversal entered
`local:<path>`, leaving workspace-only dependencies unreachable. Non-root
workspace edges now share the `local:<path>` identity used by selection.

## Files

- `internal/pnpmsource/lock.go`
- `internal/pnpmsource/capture.go`
- `internal/pnpmsource/materialize.go`
- `internal/pnpmsource/patch.go`
- `internal/pnpmsource/conformance_test.go`
- `internal/pnpmsource/patch_test.go`
- `README.md`

## Final validation

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 -cover ./internal/pnpmsource` | 0 | 80.3% statements; `go-test-cover-final.log` |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/pnpmsource` | 0 | `go-test-scoped-01.log` |
| `go vet ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | 0 | `go-vet-scoped-final.log` |
| `golangci-lint run ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | 0 | `golangci-lint-scoped-02.log` |
| `go build -o .temp/TASK-260811-3ksxig/curator ./cmd/curator` | 0 | `go-build-curator-final.log` |
| `go test -count=1 ./...` | 0 | `go-test-all-final.log` |
| `git diff --check` | 0 | post-README handoff check; `git-diff-check-handoff.log` |
| `task-board validate` | 0 | post-resource/checklist check; `task-board-validate-handoff.log` |

## Iteration evidence

No red command is represented as passing:

- Baseline `go test -count=1 ./internal/pnpmsource`: exit 0.
- Development test 01: exit 1 because the old placeholder manager hash was
  correctly rejected; the fixture was replaced with a real hash/context.
- Development test 02: exit 1 and exposed the workspace importer identity bug.
- Development test 03: exit 0 after the workspace fix.
- First scoped `golangci-lint`: exit 1 for one G304 finding and one obsolete
  test helper; both were fixed. The second lint run exited 0.
- All later coverage/test/vet/build/diff/board gates listed above exited 0.

Research and tool-readiness evidence is retained under
`.temp/TASK-260811-3ksxig/`, including `research-pnpm-patch.md`. The initial
unquoted GitHub-tree `curl` probe exited 1 due to zsh glob expansion; its quoted
retry exited 0.
