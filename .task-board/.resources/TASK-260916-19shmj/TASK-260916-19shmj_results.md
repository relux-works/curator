# TASK-260916-19shmj results — manager nofollow conformance

## Rules and production paths

The pinned source is curator-spec v1.0.0-rc.13 at commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. Its `protocol/environments.md` §8.3.1 requires materialization, takeover, repair, and backup writes to replace the directory entry via an operation-private same-directory entry and rename, without following a symlink at the target or at a path component below the managed root. Section 9.5 says takeover authorization covers replacing the entry after backup, never opening the link target for writing.

Production paths now route managed file and link replacement through `atomicManagedFile` / `atomicManagedLink`, which validate managed path components, stage beside the destination, and rename the entry. Materialization uses these from `materializeOne` (`internal/envprofile/switch.go`); repair and provisioning use them from `applyPlan`, `repairUnderLock`, passthrough/seed handling, and rendered-document writes (`internal/envprofile/managed.go`). Backup copies preserve symlink text and use nofollow reads for regular files. Takeover comments now describe the §8.3.1 directory-entry rule directly.

## Vector and ledger evidence

The vector bytes match the pinned rc.13 blob exactly: SHA-256 `f29c25c0f25f763a37718eff0f4b1d8b065d41ec1eb290a03961f3bb854e89b0`.

`TestEnvironmentWriteNofollowVectors` drives the published family through `UseWithPolicy` and `Resolve`: **10 driven, 0 known-gap, 1 bound, 0 skipped, 11 total**. The bound row is `backup-symlinked-target-refused`: no production entry requests backup of an unauthorized, unowned symlink; ordinary takeover refuses it before backup, while authorized takeover preserves the symlink itself. The authorized symlinked-target takeover, unauthorized refusal, symlinked-parent materialization/takeover refusals, and planted-link repair are driven. An additional `Resolve` test covers a symlink planted at the rendered store-document destination.

- Published case count: absent/0 before → 11 after.
- `.github/ci/conformance-gaps.tsv`: 70 total rows before → 70 after; rows owned by this task or `STORY-260916-73a5zg`: 0 → 0. No passing owned gap row existed to remove, and the ledger was not edited.
- `.github/ci/root-artifacts.tsv` now declares the vector consumed by `TestEnvironmentWriteNofollowVectors`.

## Validation evidence

- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^TestEnvironmentWriteNofollowVectors$' -count=1 -v` — exit 0; 10 driven, 1 bound.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^Test(EnvironmentWriteNofollowVectors|ResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$' -count=1 -v` — exit 0.
- Mutant 1 replaced staged `atomicManagedFile` writes with `os.WriteFile`, narrowing away final-target nofollow semantics. The vector gate failed on authorized takeover and planted-link repair — exit 1 (expected red; mutant killed).
- Mutant 2 changed parent inspection in `managedPath` from `Lstat` to following `Stat`. The vector gate failed on both symlinked-parent cases — exit 1 (expected red; mutant killed).
- `golangci-lint run ./internal/envprofile/...` — exit 0, 0 issues.
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/TASK-260916-19shmj_envprofile_windows.test.exe ./internal/envprofile` — exit 0 (Windows compile only; tests were not run on Windows).
- `git diff --check` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=... go test ./internal/envprofile -count=1` — interrupted after 479 seconds to stay within the single-shell bound; real exit 1 (`signal: interrupt`). This broad package run is unverified, not passing. The task-specific conformance tests above completed independently.

## CHANGELOG entry (for release prep)

Managed environment materialization, takeover, and repair now replace managed path entries without following unmanaged symlinks; takeover backups preserve symlink entries and link text.

## Handoff notes

No CHANGELOG.md or LOGBOOK.md file was edited. This results resource records the validation, mutation evidence, and the package-suite timeout for the task handoff.

## Revision 5 — Windows entry replacement

Root cause (orchestrator diagnosis, confirmed): on Windows `renameManagedEntry` (FileRenameInfo, ReplaceIfExists) refuses to replace an existing DIRECTORY entry (the manager's own directory symlink `skills/myskill`, or a planted junction); the error sent the link loop into `copyLinkFallback` → `copyTree` → `managedDirectory`, which refused the link as `environment_write_would_follow_link`.

Changes (rc.13 environments §8.3.1: never open/traverse a link; replacing the entry is allowed; manager §5 copy fallback only where the platform takes no symlink):
- `nofollow.go`: new `replaceManagedEntry(source, target)` used by `atomicManagedFile` and `atomicManagedLink`. If the rename fails and `os.Lstat(target)` shows a symlink / reparse point (`ModeSymlink|ModeIrregular`) or a directory, the ENTRY is removed with `os.Remove` (removes a link/junction itself, never its target; a non-empty real directory is not removed and the original error is returned) and the rename is retried. On unix the first rename replaces atomically; the fallback is not reached.
- `nofollow.go`: `atomicManagedLink` wraps only an `os.Symlink` failure in `*errSymlinkUnavailable`.
- `managed.go` link loop: falls back to `copyLinkFallback` ONLY for `errSymlinkUnavailable`; any rename/replace failure is returned as `link <path>: <err>` (never re-routed through copyTree over an existing link).
- Credential-link refusal checks (`ensurePassthroughLink`: regular file / non-empty directory / foreign link → environment_credential_conflict) are unchanged and still run before any replacement.
- Trunk combine: `DiagWriteWouldFollowLink` is now declared on trunk in switch.go; the duplicate in managed.go was dropped.

Platform bound: on Windows the remove-then-rename of an existing link/directory entry is not atomic (a window in which the entry is absent). A concurrent writer planting a link in that window makes the retried rename fail (no follow). Windows behaviour is proven only by the hosted windows-latest lane; locally unverifiable (darwin host).
Mutant bound: "fallback on any atomicManagedLink error" (pre-fix shape) is unkillable on unix because rename-over never fails there; its killer is the hosted Windows lane (the 22 tests of runs 36308915452/36313796489).

Trunk: combined with origin/main eca2bf27 (base 55b94af2 kept; trunk changes applied into the working tree; conflict in conformance-case-counts.tsv resolved keeping both rows). `git diff --name-only origin/main -- . ':!.task-board'` (temp-index incl. untracked) lists only: .github/ci/conformance-case-counts.tsv, .github/ci/root-artifacts.tsv, internal/envprofile/{managed.go,nofollow.go,nofollow_open_unix.go,nofollow_open_windows.go,state_read_guard_test.go,switch.go,write_nofollow_conformance_test.go}.

Local evidence (zsh, darwin, output redirected to $TMPDIR, real exit codes):
- `go build ./...` exit 0; `go vet ./internal/envprofile` exit 0; `GOOS=windows go vet ./internal/envprofile` exit 0.
- Pre-combine: `go test ./internal/envprofile -run 'Credential|Link|Drift|Passthrough|Nofollow|Seed|Repair|ManagerOwnedAbsence'` exit 0; `-run '^Test[A-H]' -skip <that>` exit 0; `-run '^Test[I-Z]' -skip <that>` exit 0 (whole package covered in slices; an unsplit run exceeds the 10 m default timeout on this host — timeouts, no FAIL rows).
- Post-combine: `go test ./internal/envprofile -run '…|Unmanage'` exit 0; `go test ./internal/crossconformance ./cmd/curator -run 'Count|Conformance|Unmanage'` exit 0.
- Hosted gate on the published revision is the arbiter.

## Revision 6 — re-apply on 97e85642
Accepted rev5 content (`refs/campaign/73a5zg-rev5-20260927`, diff vs eca2bf27) applied with `git apply --3way` on trunk 97e85642. Conflicts resolved keeping both sides:
- `.github/ci/root-artifacts.tsv`: trunk's envprofile row (codex-seed, env-passthrough, source-signers + system-delta) merged with rev5's `environments-write-nofollow.json` into one row; trunk's read-failure row kept.
- `internal/envprofile/state_read_guard_test.go`: trunk (ryh3kw) already removed every managed.go allow-list entry incl. `claudeSeed` (rev5's only change here) → file equals trunk; no longer a changed path.
- `internal/envprofile/managed.go`:
  - `claudeSeed` (ryh3kw stateread routing ↔ rev5 nofollow): routed through `readManagedRegular(root, rel)` (stateread.Lstat absent/unreadable split + no-follow open; a link at `.claude.json` → `environment_write_would_follow_link`; read failure → `environment_seed_unreadable`, never absence); non-regular entry → seed_unreadable; write stays `atomicManagedFile`.
  - Resolve entry: rev5 `checkManagedPrivateTarget(marker)` kept, combined with trunk's `loadResolveInputsWith(req.Home, profile, req.readRegularFile)`.
- Trunk writes added since eca2bf27, routing: E3 codex seed files → `atomicManagedFile` (seed loop); E3 codex_seed_record / seeds record → marker via `op.publish` behind `checkManagedPrivateTarget` (provision + resolve entry); XDG seed links → `atomicManagedLink`/`removeManagedEntry`; E1 signer/delta writes are in gitsource.go/profiledelta.go under `os.MkdirTemp` snapshots (not managed surface — out of scope). No raw os.* write remains in managed.go except `os.MkdirAll(root)` of the env root itself (as in rev5).
- `git diff --name-only origin/main -- . ':!.task-board'` = 8 own paths (conformance-case-counts.tsv, root-artifacts.tsv, managed.go, nofollow.go, nofollow_open_unix.go, nofollow_open_windows.go, switch.go, write_nofollow_conformance_test.go).

Local evidence (zsh, darwin, output to $TMPDIR, real exit codes):
- `go build ./...` 0; `GOOS=windows go vet ./internal/envprofile` 0; gofmt clean.
- `go test ./internal/envprofile -run 'Nofollow|Credential|Link' -count=1` 0.
- `-run 'Drift|Passthrough|Seed|Repair|Guarded|ReadFailure|SourceSigner|CodexSeed'` 0.
- Whole package in slices: `^Test[A-H]` 0, `^Test[I-R]` 0, `^Test[S-Z]` 0. (Unsplit run hit the 9 m bound in TestSCPOverlayResolvesAsGit — timeout, no FAIL rows.)
- Mutants: not re-run in rev6; rev5 mutant results (accepted) stand for unchanged nofollow.go/rules — stated bound. Hosted gate on the published revision is the arbiter.
