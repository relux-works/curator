# TASK-260910-hwxr26 results — rev10 (rework-8: unreadable binding vs first issuance)

## Change (leaf scope only: internal/audit + install regression test)

`internal/audit/sourceaudit.go`
- Deleted `sourceAuditObjectExists` (bool over `os.Stat`: followed symlinks,
  collapsed every error to absent).
- Added tri-state `sourceAuditEntryPresence` (absent / present / read-failure)
  via `probeSourceAuditEntry` using `os.Lstat` (never follows symlinks):
  Lstat ok (any kind, incl. any symlink) = present; ENOENT = absent; any other
  inspection error = read-failure.
- Added `sourceAuditBindingVacant`: true only on positive proof of absence
  (Lstat ENOENT for the object entry AND the report entry); any path error
  keeps it non-vacant.
- `CheckSourceAudit` unavailable+mutating branch: establish (first issuance)
  only when vacant; otherwise refuse before either record is written with the
  existing typed non-renewable outcome (`rejectAuditf`, `source_audit_rejected:
  evidence: existing binding is missing or unreadable; ...`, wrapped with %w).
  Read-only path behavior unchanged (all unavailable refuse).

`internal/install/draftaudit_test.go`
- Added `TestDraftAuditUnreadableBindingRefuses` at install.Project:
  first-issuance-establishes control, dangling-binding-symlink,
  self-loop-binding, loop-plus-corrupt-report (`{}`), present-valid control.
  Each breakage row asserts source_audit refusal on dry-run AND mutating
  paths, no link-target creation, unchanged link, unchanged report bytes.
  Symlink-creation failure skips with `symlinks unavailable: ...` (matches
  host-capability class in .github/ci/skip-classes.tsv; no new skip class).

## Evidence (zsh, direct redirection, real exit codes)

- `gofmt -l internal/audit/sourceaudit.go internal/install/draftaudit_test.go` — exit 0, no output.
- `go vet ./internal/audit/ ./internal/install/` — exit 0.
- `git diff --check` — exit 0.
- `go test -p 1 ./internal/install -run '^TestDraftAuditUnreadableBindingRefuses$' -count=1 -timeout=100s -v` — exit 0 (2.56s test; 5/5 subtests pass).
- `go test -p 1 ./internal/audit -count=1 -timeout=100s` — exit 0.
- `go test -p 1 ./internal/install -run '^TestDraftAudit(DuplicateKeysRefuse|MalformedShapeRefuses|ClosedPackageShapeRefuses|UnreadableBindingRefuses)$' -count=1 -timeout=100s` — exit 0 (20.654s).
- `go test -p 1 ./internal/install -run '^TestDraftAudit(RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable|PolicyRenewalBrokenEvidence|BrokenReportRefuses|FirstIssuanceVsBrokenRecord|BindingLifecycle|PinAdmits|RevocationBlocks|StrictLocalRequiresAttestation)$' -count=1 -timeout=110s` — exit 0 (21.728s).
- Narrowing mutant (temporarily replaced vacant body with `os.Stat` error-to-absence collapse `return statErr != nil`): new test — exit 1, all three link rows fail on the mutating path (dangling passes the audit gate and fails only downstream at the sibling marker boundary; loop rows fail with `audit blocked: ... too many levels of symbolic links` instead of a source_audit refusal), reproducing the rev9 verdict; controls still pass. Mutant reverted; revert verified by grep (no MUTANT marker, Lstat body restored) and the green runs above.

## Checklist

- [x] F1 fixed exactly as specified; no wider change (siblings' files untouched)
- [x] Production regression rows at install.Project, dry-run + mutating
- [x] Genuine first-issuance and present/valid controls kept
- [x] Narrowing mutant kills recorded with command + exit code
- [x] Narrow tests only, every tool call bounded; no full-suite run
- [x] No spec/LOGBOOK/registry/artifactpolicy/scriptpolicy changes

Ready for review (non-final leaf handoff rev10).
