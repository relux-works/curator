## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- BUG-260731-33v6zz

## Checklist
- [ ] Reproduce the setup-go Windows GOROOT rejection from native artifact evidence and identify the exact filesystem/path invariant mismatch.
- [ ] Fix trusted GOROOT and go.exe resolution without PATH search, download, provenance relaxation, or platform skips.
- [ ] Add focused godriver and cmd/curator regression tests and prove Windows plus macOS/Linux CI.
- [ ] Publish a signed Curator PR with native windows-latest evidence and hand off to independent Opus review.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-367e40, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-367e40)
Root cause identified from native run 30623699047 artifact test-evidence-windows-latest.

The actions/setup-go GOROOT C:\hostedtoolcache\windows\go\1.25.5\x64 lstats as mode=?rw-rw-rw- size=0 dir=false symlink=false. Per Go 1.23+ os/types_windows.go fileStat.mode(), that exact encoding means FILE_ATTRIBUTE_REPARSE_POINT with a *name-surrogate* tag: isReparseTagNameSurrogate() suppresses ModeDir, and the tag is not IO_REPARSE_TAG_SYMLINK/AF_UNIX/DEDUP so it falls to ModeIrregular. Directory junction (IO_REPARSE_TAG_MOUNT_POINT) is that signature.

Consequence chain in internal/godriver/session.go selectToolchain:
1. filepath.EvalSymlinks only follows ModeSymlink (path/filepath/symlink.go walkSymlinks), so the junction is NOT resolved and comes back unchanged -- matching EvalSymlinks(GOROOT)=same path err=nil in the evidence.
2. os.Lstat(resolvedRoot).IsDir() is false -> go_toolchain_missing trusted GOROOT is not a real directory.
3. Latent second failure behind it: walkSymlinks returns ENOTDIR for a non-symlink non-dir *intermediate* component, so EvalSymlinks(<junction>/bin/go.exe) would also fail.

Fix direction: the boundary already intends full canonicalization; filepath.EvalSymlinks simply is not the host canonicalizer on Windows. Introduce a platform physicalPath() hook -- EvalSymlinks on unix, GetFinalPathNameByHandle(VOLUME_NAME_DOS) on a follow-reparse handle on Windows -- and route every canonicalization in the package through it. After that the *existing* fail-closed predicate (IsDir && !ModeSymlink, SameFile re-verification) is unchanged and now pins the physical directory instead of the junction, which is strictly stronger. No PATH search, no download, no trust relaxation, no platform skip.

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-fs3dht_spawn-log_-implementer--developer--claude-_RUN-260731-367e40.log](file://BUG-260731-fs3dht/BUG-260731-fs3dht_spawn-log_-implementer--developer--claude-_RUN-260731-367e40.log) — System spawn log captured by task-board

## Created
2026-07-31T10:35:09Z

## Last Update
2026-07-31T10:43:53Z

## Assigned To
[implementer] developer (claude)
