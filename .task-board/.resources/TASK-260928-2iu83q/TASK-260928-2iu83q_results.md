# TASK-260928-2iu83q — Results

## Revision 1 — carrier re-apply of 31gaka rev2

Re-applied accepted source `refs/campaign/3qf8er-rev2-20260928` (tree `31e7a36f7b7263f824533e2ae028a74efd5d821b`) on Story base `main` at `d8e87bacda3bb4cd9010801156646f75de9354ce`. Fresh remote `HEAD` advertised `main` at that OID; the exact-ref fetch and local `refs/remotes/origin/main` matched it. `task-board worktree status` also reported base `main` and tip `d8e87bac...`.

The three-way apply conflicted only in `.github/ci/conformance-gaps.tsv`. Kept the current trunk rows and the nine accepted revision-B rows owned by `TASK-260927-1e5qqm`. The merged code retains trunk's nofollow/stateread and atomic managed-write helpers alongside revision-A seed behavior and the E1/E6 status changes.

`git diff --name-only origin/main -- . ':!.task-board'` exited 0 and listed exactly these seven paths:

- `.github/ci/conformance-gaps.tsv`
- `cmd/curator/env_credential_marker_test.go`
- `cmd/curator/envstatus_test.go`
- `internal/envprofile/codex_seed_test.go`
- `internal/envprofile/managed.go`
- `internal/envprofile/status.go`
- `internal/envregistry/envregistry.go`

## Local verification

- `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Nofollow|Guarded'` — exit 0, 169.910s.
- `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp'` — exit 0, 570.153s.
- `go build -o "$TMPDIR/TASK-260928-2iu83q-curator" ./cmd/curator` — exit 0.
- `golangci-lint run ./cmd/curator ./internal/envprofile ./internal/envregistry` — exit 0, 0 issues.
- `git diff --check origin/main` — exit 0.
- `git diff --name-only origin/main -- CHANGELOG.md LOGBOOK.md` — exit 0, no paths listed.

The hosted gate is run by the task-board runner after handoff; its result is not available in this producer turn. Delta review remains with the reviewer.

No CHANGELOG or LOGBOOK edits were made. This outcome records the merge decision and verification evidence.

## Revision 2 — hosted gate ledger correction

The attached `carrier-31gaka-gatefix-1.md` reports that hosted run 36406667044 failed because revision 1 restored stale manager-config conformance-gap rows from before E1/E6 landed. Rebuilt `.github/ci/conformance-gaps.tsv` from `origin/main` and reapplied only the nine Codex seed revision-B rows owned by `TASK-260927-1e5qqm`. Its diff against `origin/main` contains only those nine additions; the stale E1/E6 rows are absent.

The task snapshot remains `d8e87bacda3bb4cd9010801156646f75de9354ce`. The diff against that snapshot lists the same seven 31gaka paths. Since that snapshot, `origin/main` advanced to `56392800b8adc9ae2ffa49d5ae01101241fe6fe9` and includes an unrelated `internal/scriptworker/worker_test.go` change; this carrier did not incorporate or modify that path.

### Verification for this revision

- `go test ./internal/config ./internal/envprofile -run 'Conformance|Schema|Seed|Codex'` — exit 0 (`internal/config` 1.592s; `internal/envprofile` 15.349s).
- `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Nofollow|Guarded'` — exit 0 (52.472s).
- `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp'` — exit 0 (254.519s).
- `golangci-lint run ./internal/envprofile ./cmd/curator ./internal/envregistry` — exit 0, 0 issues.
- `go build ./internal/envprofile ./internal/envregistry ./cmd/curator` — exit 0.
- `git diff --check HEAD` — exit 0.
- `git diff --name-only d8e87bac -- . ':!.task-board'` — exit 0, exactly the seven 31gaka paths.
- `git diff --name-only origin/main -- .github/ci/conformance-gaps.tsv` — exit 0, only the ledger path; its content diff has the nine intended Codex rows.

Hosted run 36406667044 remains the recorded failed revision-1 gate. The runner will publish the next hosted-gate result after this handoff; no hosted result for this ledger correction is available in this producer turn. Delta review is pending. No CHANGELOG or LOGBOOK edits were made.
