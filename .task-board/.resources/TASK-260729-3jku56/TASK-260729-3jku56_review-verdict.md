# TASK-260729-3jku56 review verdict

## Verdict

ACCEPTED. No correctness, safety, architecture, or task-scope defect found.

## Scope and architecture review

The task-only delta was isolated against accepted predecessor TASK-260720-1ljev5 and is limited to internal/install/targets.go, internal/install/install.go, internal/install/global.go, and internal/install/stage_test.go. stageNode now receives marker.BuildCurrentness only for nodes whose complete compiled plan is exact cache-hit. Build misses, rebuilds, corrupt or unsupported outcomes, absent inputs, and derivation failures pass no proof or fail before staging, preserving marker.Current fail-closed semantics.

The proof does not trust the installed marker. It revalidates node.Snapshot with buildsource.Validate, performs a fresh CacheInspector.Inspect using the planned input and recorded receipt identity, clones all planned build inputs including target tuning, and independently enumerates the same static context include/exclude boundary and active script runtime source boundary used by staging. marker.Current still verifies build source identity, command set, cache key, receipt bytes and hash, artifact metadata, context files, runtime/build-root separation, and snapshot stability around cache inspection.

Project and global call sites use the same helper. Exact cache hits can return up-to-date without a context replacement; source, target, toolchain, cache, context, or unavailable state cannot be accepted as current.

## Independent validation

- Focused regression and drift matrix: go test ./internal/install -run focused-regex -count=1 — PASS, 34.553s.
- Focused race gate: go test -race ./internal/install -run focused-regex -count=1 -timeout=10m — PASS, 102.698s.
- Full install package: go test ./internal/install -count=1 — PASS, 287.388s.
- Marker package: go test ./internal/marker -count=1 — PASS, 0.486s.
- Repository build: go build ./... — PASS.
- Repository vet: go vet ./... — PASS.
- Lint: /Users/iv/go/bin/golangci-lint run — PASS, 0 issues.
- Formatting: gofmt -l on all four task files — PASS, no output.
- Diff hygiene: git diff --check — PASS, no output.

The first reviewer go test ./... attempt exhausted the host temp volume and produced cascading no-space failures after many packages passed. The isolated godriver timeout from that loaded run passed on rerun in 1.348s. A serialized go test -p 1 ./... rerun passed through internal/install in 310.551s, then an external process removed shared Go cache objects and later packages failed with cache files missing. These are host/shared-cache interruptions, not code failures. Producer evidence records a prior clean go test ./... -count=1 exit 0, and every task-relevant reviewer gate is clean.

## Concurrent-task reconciliation

Concurrent TASK-260720-1nlmvv changes diagnostics before install staging. Its production hunks do not overlap or alter this currentness path. The only composition detail is a mechanical combination of independent imports in stage_test.go; no priority rule, compatibility shim, or behavioral tradeoff is required.
