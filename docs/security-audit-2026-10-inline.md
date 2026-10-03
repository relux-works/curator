# In-depth Curator Review: Confirmed New Defects

> **Tracking.** Issue: [relux-works/curator#106](https://github.com/relux-works/curator/issues/106). Board Story: `STORY-261004-3oognx` (epic `EPIC-260910-2hw1xb`). Tasks: N1 — `BUG-261004-bknio5`, N2 — `BUG-261004-2v9pbz`, N3 — `BUG-261004-33fnzw`, N4 — `BUG-261004-13ptlq`.
> Russian version (original): [security-audit-2026-10-inline.ru.md](security-audit-2026-10-inline.ru.md).
> **Commit identity.** The audited commit is given by its current identity on main. Every file cited by the findings is unchanged up to and including `68210ecc`.

Report date: 2026-10-03. The review was performed by the primary agent in a single session, **without sub-agents**. It is a report of targeted analysis and reproductions, not a certification of the whole code base.

## Identity, method and boundaries

- Audited Curator commit: `aeab533469d5647c2cbe5c8bbc126d6ea2573147`. HEAD was the same at the end of the review.
- A disposable copy was produced with `git archive HEAD` in `/tmp/curator-inline-audit.t8ljBs`. Only two test files were added to the copy: `cmd/curator/audit_inline_test.go` and `internal/buildrepo/audit_inline_test.go`. The production Go files are identical to the source; a tree comparison differed only in these tests and in local .DS_Store files that git archive does not carry.
- Tested on macOS arm64, Go 1.27.1. Commands used `env -u GOROOT GOMAXPROCS=2 go test -p 2`: the inherited GOROOT points at a different Go installation.
- All mutable data were temporary manager homes, projects and local Git repositories. No real user credentials were read for the reproductions, and no payload was sent to production services.
- No production code was fixed; there were no commits, pushes, tool installations or changes to earlier done statuses.
- The new findings were cross-checked against `docs/security-audit-2026-09.md`, its E1–E7 addendum and the three reports of the previous review. They are new, concrete defects, not repeats of the A-warning, the revocation-hiding issue or the inaccurate security_posture.
- This is not a full audit of all platforms and subsystems. There was no independent second reviewer in this pass: the user explicitly asked for the work to be done without sub-agents.

## N1. GC deletes a runtime still in use when the reference set is incomplete

**Confirmed. Priority: fix first; risk of breaking installed commands.**

Source: `internal/scopes/gc.go:90`, `:97`, `:103`; public CLI: `cmd/curator/main.go:2274`, `:2328`.

`Collect` first obtains the reference set and any errors in establishing it, then calls `sweepRuntime`, and checks `marked.uncertain` only after the deletion. The guard protects the build cache but not the runtime store. With a corrupted consumers.json the set of runtime references can be empty; with an unreadable or corrupted install marker the reference of one specific live installation is lost.

The probe `TestInlineAuditGCPreservesLiveRuntimeOnUncertainty` calls the real `run([gc], ...)`, takes the real production locks and creates a working shim with `runtimestore.WriteBinShim`. Before GC the command prints `runtime-ok`.

| State | Result |
|---|---|
| Valid registry and marker | CLI exit 0, runtime kept, the command works |
| consumers.json holds truncated JSON | CLI exit 0, warning about incomplete references, runtime deleted, running the shim: no such file or directory |
| Install marker holds truncated JSON | CLI exit 0, warning install_marker_invalid, runtime deleted, running the shim: no such file or directory |

No malicious package needs to run: an unavailable or corrupted state is enough. Corruption of the consumer registry can affect the runtimes of several projects. The test did not delete any real installation.

Why the existing tests missed it: `internal/scopes/gc_conservative_test.go:55` and `gc_test.go:292` check that the consumer registry is preserved and that no **build-cache** sweep happens, not that runtime files in use survive.

First fix: stop any runtime sweep while the marking is incomplete; keep the existing conservative treatment of registry/consumer state. Add CLI regressions for a corrupted registry, an invalid/unreadable marker and a repeated GC. Do not weaken the normal removal of genuinely unused objects.

## N2. Repeated references to one Git blob bypass the expanded-snapshot budget

**Confirmed on a small, safe example. Priority: fix first; risk of excessive memory use when admitting an external repository.**

Source: `internal/buildrepo/admission.go:716`, `:738`, `:1061`, `:1072`. Public entry of the local check: `internal/buildrepo/local.go:44`. Shared network admission path: `internal/buildrepo/admission.go:429`.

`objectReader.read` returns a repeated object from its cache before the limit is accounted. That bounds the sum of unique objects, but `walkTree` copies the blob separately for every file path. `frameSnapshot` then assembles the whole expanded snapshot in memory again. There is no separate aggregate budget for these copies.

`TestInlineAuditExpandedSnapshotBudget` creates an ordinary real Git repository and calls `AdmitLocal`, not just an internal helper:

| State with MaxExpandedBytes=16384 | Result |
|---|---|
| One 4096-byte blob | Admitted: 4113 content bytes, 4236 bytes of canonical snapshot |
| 64 distinct 4096-byte blobs | Refused build_repository_incomplete_source: object size limit exceeded |
| One 4096-byte blob at 64 different paths | **Admitted:** 262161 content bytes, 263922 bytes of canonical snapshot |

The third case confirms consumption beyond the declared budget without large data and without attempting an OOM. Potential memory exhaustion is an inference from the repeated byte allocation in production code, not an observed machine failure.

The network and local entries converge in `proveRepository`; applicability to a network source follows from the code, and no separate network scenario with this input was run. The manager §11.5 contract in the pinned rc.13 specification requires bounding aggregate expanded bytes and memory. A file-count limit alone does not hold a realistic total volume of data.

First fix: account separately for the total emitted files, bytes and canonical framing size, checking the limit **before** copying/appending. Keep the unique-object limits. Also bound the number of visited expanded tree entries and check the context during the walk; repeated tree DAGs were not tested dynamically.

## N3. Restoring a private file widens its access permissions

**Confirmed. Priority: medium; potential disclosure of private context when parent directories are accessible to other users.**

Source: `internal/envprofile/unmanage.go:273`, `internal/envprofile/unmanage.go:378`. Backup creation preserves the mode: `internal/envprofile/switch.go:1059`.

Restore loads the backup as `map[string][]byte`, losing the mode, and always creates the file with 0644. A real `profile use --takeover` saves an original 0600 file into the backup as 0600, but a subsequent `env unmanage --restore-backups --env claude_code` restores the same bytes with mode **0644**, exit 0. An ordinary 0644 file is the passing control.

`TestInlineAuditUnmanageRestoresOriginalMetadata/private-file` uses only a synthetic CLAUDE.md and the real CLI install/use/unmanage. No actual read by another user was performed; whether disclosure is possible depends on the parent directory permissions. It is not claimed that auth.json files were restored or leaked: credential paths are excluded separately.

The environments §8.3 specification explicitly acknowledges that a backup of hand-maintained context may contain secrets. Preserving only the bytes without the original access restrictions is unsafe.

First fix: keep the type and mode of each backup entry in the restore plan, and use a protected atomic write that never widens the original permissions. Regressions: 0600, ordinary 0644, executable mode where the type supports it; check permissions on Unix, and Windows ACLs separately.

## N4. Curator cannot restore a link backup it created itself

**Confirmed. Priority: medium; the return to the previous context management is broken.**

Source: creating the link in the backup — `internal/envprofile/switch.go:1032`; the unconditional refusal to read such a backup — `internal/envprofile/unmanage.go:260`.

On takeover of a foreign symlink, Curator correctly saves **the link itself**, without reading or overwriting its target. However, `readBackupTree` admits only regular files. After a successful takeover, the ordinary `env unmanage --restore-backups` cannot read its own backup.

`TestInlineAuditUnmanageRestoresOriginalMetadata/foreign-symlink`: the original CLAUDE.md is a link to a temporary external file; profile install and profile use --takeover succeed; the backup has type symlink; the restore ends with exit 1 and `environment_backup_record_unreadable: backup entry is not a regular file`.

This is **not a repeat of E5** from the September audit: E5 concerned writing through a symlink and damaging the external target. Here the takeover is protected, but the reverse operation does not support the backup format the takeover produced.

First fix: align the entry types supported by backup and restore, and restore the saved link as an entry, atomically and without following its target. Do not weaken the refusal of links in the parent route. Add a full CLI round-trip foreign symlink → takeover → unmanage restore.

## Checks and limitations

The new regressions deliberately state the expected safe behaviour. On the current code they are red: **9 scenarios, 4 controls pass, 5 demonstrate defects**. These are not fixes, and not "tests green".

Commands:
```sh
env -u GOROOT GOMAXPROCS=2 go test -p 2 ./cmd/curator ./internal/buildrepo -run '^TestInlineAudit' -count=1 -v
```

The results and the exact probe files were attached next to the report as task-scoped resources.

Targeted existing tests:
- `internal/scopes`: the selected `TestCollect*` passed.
- `internal/buildrepo`: the selected raw-object, local/network parity, tree-path and local administration conformance checks passed with the rc.13 vectors.
- `cmd/curator`: the selected ordinary GC and `TestEnvUnmanage*` tests passed.
- The first, wider CLI expression also included `TestGCRetainsAndReportsReferencedCompiledState`; it failed because of Go 1.27, while the driver allowlist admits 1.25. This is an environment limitation, not a newly confirmed GC defect. No allowlist bypass or production configuration change was made. The repeated narrow pass excluded exactly this build/toolchain-dependent test; the result of the first pass is kept.

### Actual coverage

| Surface | What was done / what is not proven |
|---|---|
| GC and runtime deletion | Production CLI, working command before/after, corruption controls; N1 |
| Git raw-object admission | Production AdmitLocal with real Git and small budgets, unique/repeated controls; N2. A heavy OOM and all DAG variants were not run |
| Takeover/restore | Full CLI round-trip for regular/private/symlink; N3/N4 |
| Transaction namespace and snapshots | Path checks, match/alias targets and immutable-cache authentication reviewed statically; no new confirmed bypass, exhaustive races not checked |
| Global shim ownership and managed file writes | Ownership ledger, staging, atomic publication and refusal of links in parents reviewed statically; no full new dynamic sweep |
| Credential helper / worker framing | Response limit, prompting ban and the bounded-frame protocol were read; no full check of secrets/inter-process authorization |
| npm/tar input | The missing local limit in the helper is not declared a separate hole: artifact-policy admission precedes it, and the whole chain needs a separate analysis |
| Linux/Windows, published releases and neighbouring services | Not checked in this inline pass |

Earlier investigations interrupted by the child-run filter are not evidence for these conclusions. All the reproductions listed were performed after the user explicitly asked for the work to be done in the main session. Overall conclusion: **four new defects in three mechanisms** were found, but the absence of other problems cannot be claimed.
