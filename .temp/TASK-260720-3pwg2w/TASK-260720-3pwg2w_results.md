# TASK-260720-3pwg2w implementation evidence

## Provenance and integration boundary

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3pwg2w/worktree`
- Exact base: `origin/main` at `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Accepted predecessor: the complete reviewed product diff from `TASK-260720-3mrm4z` was imported before task work.
- Byte-for-byte provenance comparison: `go.mod`, `internal/protocoljson`, `internal/registry`, `internal/snapshot`, `internal/buildmeta`, and `internal/buildsource` remain identical to `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3mrm4z/worktree`.
- Candidate conformance root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1`
- Candidate manifest SHA-256: `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`
- The candidate suite was consumed only as explicit vector input, never as release or pin evidence.
- No task-board/config, planning/research, diagrams, binaries, caches, alternate indexes, or unrelated shared-checkout files were imported. Nothing was committed or staged.

## Task-only implementation

Added `internal/buildcache` with:

- Curator-local physical layout `cache/build/go-v1/<64-lowercase-hex-key>/curator-receipt.ccj.json` and `bin/<command>[.exe]`, while public inspection addresses portable `buildmeta.Input`, logical keys, receipt hashes, and relative artifact paths.
- Explicit `hit`, `miss`, `corrupt`, `untrusted-provenance`, and `unsupported` outcomes plus stable dry-run mapping (`cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`, `corrupt`, `unsupported`). Inspection is strictly read-only.
- Exact canonical receipt decoding, complete expected-input and logical-key comparison, optional exact receipt-hash comparison, manager-derived artifact path checks, and artifact size/SHA-256 verification.
- A caller-held manager-home lock witness required by every publication and quarantine. A missing, typed-nil, released, or rejecting witness fails before persistent mutation.
- Private protected staging with synced receipt/artifact bytes and one atomic directory winner. Identical concurrent losers are discarded; a different protected winner for the same logical key returns `ConflictError` without modifying the winner.
- Locked quarantine/replacement for corrupt or untrusted live entries. Existing entries are never merged, adopted, or permission-repaired.
- POSIX no-follow handle traversal from the manager-home boundary, effective-UID ownership, group/other non-writability, regular singly linked receipt/artifact files, exact directory contents, and owner-executable artifact enforcement.
- Windows reparse-point rejection, pinned no-delete-share handles, regular singly linked files, effective-user owner equality, protected DACL enforcement, owner mutation rights, and rejection of mutation grants to any other principal. Manager-created cache paths receive a protected owner-only DACL.
- Unsupported platforms disable persistent reuse/publication/quarantine fail closed.

## Protected-state and publication evidence

- Exact protected receipt/artifact: reusable `hit`.
- Noncanonical receipt, receipt-hash drift, input/key/path mismatch, artifact hash/size drift, partial entries, and unexpected contents: `corrupt`, never reused.
- POSIX writable components, non-owner state, symlinks, special files, multiply linked files, non-executable artifacts, and a symlinked manager-home/root boundary: `untrusted-provenance`, never reused.
- Windows tests cover DACL mutation grants, reparse points, special files, and multiply linked artifacts; the package cross-compiles with these platform tests.
- The self-consistent forged entry is built with matching receipt, key, receipt hash, artifact hash, and size outside protected state. Read-only inspection leaves the tree byte-for-byte unchanged and reports `would-rebuild-untrusted-cache`; a subsequent real locked publication quarantines it and publishes freshly verified protected bytes rather than adopting it.
- Concurrent identical and conflicting publication races were run under the race detector and repeated 20 times. Exactly one new directory winner is selected; identical losers reuse it and different bytes produce one success plus one conflict.

## Verification

- `make check` — PASS (`go vet ./...`, `go test ./...`, repository gofmt gate).
- `go build ./...` — PASS.
- `go test -race ./...` — PASS.
- `go test -race -count=1 ./internal/buildcache` — PASS.
- Candidate-focused `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./internal/buildcache ./internal/buildmeta ./internal/buildsource ./internal/protocoljson ./internal/registry` — PASS.
- `go test -count=20 ./internal/buildcache -run 'TestAtomicPublication|TestUnixForgedSelfConsistentEntryIsNeverAdopted'` — PASS.
- Native `internal/buildcache` statement coverage — 81.6%.
- `GOOS=linux GOARCH=amd64 go test ./... -run '^$' -exec=true` — PASS.
- `GOOS=windows GOARCH=amd64 go test ./... -run '^$' -exec=true` — PASS (compile/link gate).
- Focused buildcache compile matrix — PASS for FreeBSD, NetBSD, OpenBSD, AIX, DragonFly BSD, illumos, Solaris, and unsupported Plan 9.
- `git diff --check` — PASS; `gofmt -l internal/buildcache` — empty.
- Accepted predecessor product files compared byte-for-byte — PASS.

Windows runtime tests were not executed because the available host is Darwin and has neither a Windows runner nor Wine. Windows production and test sources compile/link successfully; native Darwin exercises the POSIX protection tests. This is the only unexecuted platform runtime gate.

Logs are retained under `.temp/TASK-260720-3pwg2w/`: `make-check-final.log`, `full-race-final.log`, `go-build-final.log`, `candidate-final.log`, `coverage-final.log`, `race-stress-final.log`, `windows-compile-final.log`, `linux-compile-final.log`, and `platform-matrix-final.log`.
