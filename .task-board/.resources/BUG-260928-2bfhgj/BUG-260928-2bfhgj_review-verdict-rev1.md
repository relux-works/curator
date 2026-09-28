# BUG-260928-2bfhgj review verdict — rev1 ACCEPTED (reviewer claude-opus-5-5 low, 2026-09-28)

Candidate: base 2252ebee, tree 3db1bec3; worktree `git diff 3db1bec3 --quiet` = 0 (exact tree). 1 path: internal/scriptworker/worker_test.go.
Host: x86_64 macOS 15.7.4 (NOT arm64 — rose-air arm64 kill cannot be reproduced here; rose-air remains unverified until the main-push lane runs).

## Diagnosis check (note item 1)
Cross-built darwin/arm64 test binary (`GOOS=darwin GOARCH=arm64 go test -c`): linker-signed ad hoc, LC_CODE_SIGNATURE at 8342672 size 65298 = EOF,
superblob holds ONE CodeDirectory (offset 20, size 65278) that ends exactly at EOF. So the old `payload[len-1]^=0xff` flips the last byte of the
last page-hash slot (hash of the final __LINKEDIT page before the signature). The mismatch surfaces only when that page is validated/faulted in
(or under stricter AMFI enforcement), which is why hosted macos-latest arm64 could still launch it: the old fixture worked there by accident.
The producer's comment "kernel kills an arm64 image whose signature no longer matches before main runs" overstates it as universal — minor
wording bound, not blocking. Re-signing ad hoc under a new identifier removes the accident on every darwin host. x86_64 local: unsigned binary,
byte-flip runs (rc 0) — consistent.

## Tests (zsh, set -o pipefail, real exit codes)
- `go test -count=1 -run '^TestScriptWorker(RejectsForgedWorkerIdentity|RejectsSubstitutedManager)$' -v ./internal/scriptworker` → PASS, exit 0.
- `go vet ./internal/scriptworker` → exit 0.

## Mutants on internal/scriptworker/worker.go:130 (restored after; git diff --stat shows only the candidate file)
- M0 worker check neutralised (`MatchesExpectation(identity.Path, identity.SHA256, identity.Size)`) → BOTH tests FAIL ("worker sent ready, want a failure"), exit 1. KILLED.
- M1 digest ignored (expected SHA := own SHA) → PASS. SURVIVES.
- M2 digest+size ignored (path only) → PASS. SURVIVES.
Survivors are PRE-EXISTING (the tamper fixture always lives at a different physical path, so the path half refuses first); the "modified bytes"
arm never isolated the digest on base either. Stated bound / follow-up suggestion: a same-path byte-swap row (replace the installed file in place
after resolving the expectation) would kill M1/M2. Not a regression of this rev.

## Note item 3
Non-darwin path is byte-identical in effect (last-byte flip). darwin requires /usr/bin/codesign (base-OS tool) and fails loudly if absent — acceptable
(no silent skip). Failure messages now carry worker exit status/state + captured stderr (exitReport, worker_test.go:74-86) for receive/send.
No CHANGELOG/LOGBOOK edits.

Verdict: ACCEPTED. Rose-air green on main still to be observed after landing (lane runs only on main pushes).
