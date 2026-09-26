# TASK-260922-1t551d results — F-C1 envprofile credential-link repairs

Producer: developer. Worktree: `.temp/STORY-260922-1cenbr/worktree` (uncommitted).
Normative source: curator-spec 05053cd7 (Decision 0017 + operator corrections):
environments §7.4/§7.7/§8.4.1/§10.1/§10.4, manager §12.4/§12.5.

## Outcome

Implemented fix-first credential-link repairs. Repair unlinks only a recorded
symlink still targeting the declared native store (native bytes preserved);
a regular file, an unexpected-target symlink, or a foreign link at a link
path refuses with `environment_credential_conflict` naming the path, removing
nothing. Shared→isolated removes the stale recorded link. `env status` and
`env resolve` report a dangling or mis-targeted recorded link as detached
with conflict-class wording, never silence. `codex_cli` isolated is admitted
under effective `file` storage only (absent key = file); `keyring`/`auto`
refuse with `environment_isolated_unsupported`; any other selector fails
closed with `environment_credential_unsupported`. New Pi homes link
`~/.pi/agent/auth.json`. The frozen v1 marker still records path+strategy
only. CHANGELOG (`Fixed` + `Changed`) and two troubleshooting entries added.

## Hazard analysis (the two 0017 hazards)

H1 — shared→isolated leaves the old credential link behind, so sharing
survives behind an `isolated` configuration. Pre-change chain in
`internal/envprofile/managed.go`: `effectivePassthrough` returned empty for
isolated (~:500), `finalizeMarker` looped over wanted links only (old
:934-942), the removal set came from marker surfaces only in
`repairUnderLock` (old :1671-1678), and `checkPassthrough` only compared
set shapes (old :1342). Fix: `finalizeMarker` (:979) now ensures wanted
links via `ensureCredentialLink` (:990 → :1063) and then unlinks stale
recorded links via `removeStaleCredentialLinks` (:994 → :1115), which
unlinks only a recorded symlink whose target equals the reconstructed
declared target (`declaredPassthroughTarget`, :1160) and refuses anything
else with `environment_credential_conflict` naming the path.

H2 — `finalizeMarker` unlinked the wanted-link path before symlinking
(`_ = os.Remove(full)`, old :936) without credential-specific
preservation, contrary to the manager §12.4 leave-bytes-untouched promise.
Fix: `ensureCredentialLink` (:1063) links an absent path, leaves an intact
recorded symlink alone, replaces an empty directory, and refuses a regular
file / unexpected-target symlink / non-empty directory / unreadable link
with `environment_credential_conflict` naming the path, removing nothing.
The `:871` (plan.links) and `:975` (XDG seeds) unconditional removes are
manager-owned surfaces/seeds with backup discipline, not credential links:
deliberately unchanged.

## Deliberate contract change (one existing test)

`TestResolvePassthroughLiveness` pinned the H2 behavior: repair re-linked
over a regular file. Under 0017 §7.4/§10.1 that state is the
`environment_credential_conflict` refusal, so the test keeps the
absent-link re-link half and the regular-file half moved to
`TestCredentialLinkRegularFileRefuses` with the new contract (refusal
naming the path, managed + native bytes untouched). The test comment
records the change; the failure observed before the update was the new
refusal firing, not a regression.

## Row table (all driven through production entries on temp stores)

| Brief row | Test (file `internal/envprofile/credential_link_test.go`) | Entry | Asserts |
|---|---|---|---|
| (a) shared→isolated, no stale link | `TestSharedToIsolatedRemovesStaleLink` | `Resolve` | stale turn; repair ok; link path absent; native bytes identical; marker record empty; isolated bare resolve current |
| (b) regular file → conflict, bytes untouched | `TestCredentialLinkRegularFileRefuses` (non-empty + empty) | `Resolve` | stale + conflict wording in reasons; repair refuses `environment_credential_conflict` naming the full path, not wrapped as `environment_repair_failed`; managed + native bytes identical; still a regular file |
| (c) dangling Pi link → detached + conflict wording | `TestDanglingPiLinkReportedDetached` (operator reproduction: managed link → nonexistent `…/pi/auth.json`, real credential at `…/pi/agent/auth.json`) | `StatusOf` + `Resolve` | status row non-current; finding carries `environment_passthrough_detached` + `environment_credential_conflict` + path + expected agent target; stale reasons same; repair refuses naming the path; link untouched |
| (d) codex isolated admission | `TestCodexIsolatedAdmission` (absent file / absent key / file admit; keyring / auto → `environment_isolated_unsupported`; unknown → `environment_credential_unsupported`) + `TestCodexUnknownStoreSharedRefuses` (shared-side fail-closed before first write; fix-and-retry provisions) | `Resolve` | per-cell diagnostic; admitted homes link nothing and record empty passthrough; refused provisioning writes nothing |
| (e) Pi agent root for new provisioning | `TestPiProvisionTargetsAgentRoot` | `Resolve` | exact link target `…/pi/agent/auth.json`; bare resolve current while native target absent (dangling-but-correct is current); marker entry is `{auth.json, file-link}` and marshals with exactly `path`+`strategy` keys (frozen v1 pin); present credential reads through |
| extra: stale-side refusals | `TestStaleCredentialLinkRefusals` (regular file / foreign same-basename symlink / directory at the stale path) | `Resolve` | each refuses with conflict naming the path; bytes/link/contents untouched |
| extra: directory rule §7.4 | `TestCredentialLinkDirectory` (empty re-linked; non-empty refuses) | `Resolve` | re-link target exact; refusal names path; contents untouched |

## Mutant table (production call sites named; all killed, shell `bash`)

| Mutant | Site | Type | Killer |
|---|---|---|---|
| M1 skip stale unlink | `removeStaleCredentialLinks` (`managed.go:1115`) | deleting | (a): link still on disk |
| M2 restore unconditional Remove over regular file | `ensureCredentialLink` (`managed.go:1063`) | deleting (brief row b) | (b) both subtests: repair succeeds instead of refusing |
| M3 refuse only non-empty files | same | narrowing | (b) empty subtest admits |
| M4 silence mis-targeted link | `checkPassthrough` (`managed.go:1516`) | deleting/silence (brief row c) | (c): status reports current (reproduces the operator's silent-status bug) |
| M5 basename-only target compare | same | narrowing | (c): both basenames are `auth.json` → current |
| M6 Pi target back to `auth.json` | registry (`envregistry.go:293`) | deleting (silence) | (c) silent + (e) wrong target |
| M7 admit `auto` under isolated | `checkIsolatedCredentialStore` (`managed.go:575`) | deleting an arm | (d) `auto_refuses` admits |
| M8a reader maps unknown → file | `codexCredentialStore` (`managed.go:544`) | deleting | (d) unknown admitted; shared-side unknown provisions |
| M8b reader passes unknown through | same | narrowing (partition) | shared-side unknown links as file instead of failing closed (isolated side still refuses via the gate default, proving the two-layer bound) |
| M9 unlink any symlink at stale path | `removeStaleCredentialLinks` | deleting | stale foreign-symlink subtest: link removed, no refusal |
| M10 basename-only stale compare | same | narrowing | stale foreign-symlink subtest (same basename, other dir) unlinked |
| M11 skip directory emptiness check | `ensureCredentialLink` | deleting | non-empty-dir subtest gets `passthrough …: file exists`, not the conflict diagnostic |
| M12 plain wording for regular file | `checkPassthrough` | narrowing (wording) | (b): stale reasons lack the conflict class |
| M13 drop pre-write credential check | `repairUnderLock` (`managed.go:1853`) | deleting | shared-side unknown leaves a markerless partial home; "writes nothing" fails |

Mutant script (scratch, not committed): `/tmp/fc1_mutants.py`; M4/M6/M11
re-run individually for genuine assertion kills (first pass hit unused
variables / output filtering). Tree verified restored after every mutant
(`git diff --stat` + `go build ./...`).

## Bounds (not owned here)

- F-C2 owns the explicit inspect→plan→apply credential migration: Pi
  old-target → agent-root moves, isolated→shared displacement, and any
  re-pointing. This task reports and refuses those states; repair moves
  no credential bytes.
- F-S1 owns the extended credential record (`isolation`, `source_role`,
  `backend`, `backend_version`, `provenance`). The marker still records
  path+strategy only (pinned by row e).
- F-C3 (per the operator relay) owns the remaining dangling-link
  follow-through; the `environment_passthrough_unreadable` row (§8.4.1:
  stat/read failures are unreadable, never detached) is not implemented
  here — permission/I/O failures keep the pre-existing detached mapping
  and repair refuses them fail-closed without touching the path.
- `resolveIsolation` (registry static matrix) is unchanged: the
  `codex_cli` store gate is dynamic (reads native `config.toml`) and
  lives in `assembleHome` via `checkIsolatedCredentialStore`, so the
  registry `TestIsolationMatrix` pin (codex isolated available at the
  static layer) still holds.
- Windows symlink privilege: new rows use `os.Symlink` directly, matching
  the existing envprofile rows (no skip pattern exists in this package);
  verified on darwin; other platforms unverified.

## Validation (shell `bash`, real exit codes)

- `go build ./...` → 0
- `go vet ./...` → 0
- `gofmt -l cmd internal` → clean
- `golangci-lint run ./internal/envprofile/... ./internal/envregistry/...` → 0 issues
- `go test -count=1 ./internal/envregistry/ ./internal/envmarker/` → ok
- `go test -count=1 ./internal/envprofile/` → ok (262s, full package incl. 8 new tests)
- `go test -count=1 -run 'TestEnv|TestCmdEnv|TestEnvalias|TestEnvconfig' ./cmd/curator/` → ok (146s)
- 14 mutants above → all killed; tree restored (final `git diff --stat`:
  6 files + 1 new test file; `go build` green after restore)

## Files changed

- `internal/envregistry/envregistry.go`: `environment_credential_conflict`
  / `environment_credential_unsupported` diagnostics; Pi `FileLinkTarget`
  `agent/auth.json`.
- `internal/envprofile/managed.go`: `codexCredentialStore` (replaces
  `codexKeyring`); `checkIsolatedCredentialStore` gate in `assembleHome`;
  `ensureCredentialLink` / `removeStaleCredentialLinks` /
  `declaredPassthroughTarget`; conflict-class liveness wording in
  `checkPassthrough`; pre-write credential check + refusal-preserving
  error surfacing in `repairUnderLock`.
- `internal/envprofile/status.go`: detached prefix covers conflict-worded
  liveness reasons.
- `internal/envprofile/managed_test.go`: liveness contract update (above).
- `internal/envprofile/credential_link_test.go` (new): rows (a)–(e) +
  stale-refusal + directory rows.
- `CHANGELOG.md` (`Fixed` + `Changed`), `docs/troubleshooting.md` (two
  diagnostics).

## Revision 2 (rework-1: reviewer R1 + R2)

Verdict rev1 was CHANGES_REQUESTED
(`TASK-260922-1t551d_review-verdict-rev1.md`): dangling
expected-target links stayed silently current (R1), and valid TOML
literal selectors bypassed the Codex refusal gate (R2). Both are fixed
below, on top of the revision-1 tree, with the reviewer's probes
committed as rows.

### R1 — expected-target dangling is detached (fix)

`checkPassthrough` (`internal/envprofile/managed.go:1541`) compared the
recorded target string but never established that the native target
exists. After a target-string match it now stats the native target
(`:1592`, `os.Stat`, never a read of credential bytes):

- target absent (`os.IsNotExist`) ⇒ `passthrough entry <path> is
  detached: link target <native> does not exist
  (environment_credential_conflict)` — detached with conflict-class
  wording in `StatusOf` and bare `Resolve`, never silence;
- target stat fails otherwise (permission, I/O, not-a-directory) ⇒
  `passthrough entry <path> target <native> cannot be inspected: <err>
  (environment_credential_conflict)` — a conflict/inspection
  diagnostic, never absence, never silence. It deliberately carries no
  `is detached` label, so `env status` does not prefix it
  `environment_passthrough_detached`: unreadable is not detached
  (§8.4.1), and F-C3 owns the `environment_passthrough_unreadable`
  row for link-level failures.

Both reasons are also collected in `verification.dangling`
(`managed.go:1244`). `verifyHome` stays a pure function of state; the
provision transition owns the one policy carve-out:
`repairUnderLock` (`:1948`) calls
`tolerateDanglingAtProvision` (`:1977`) on the post-apply verdict when
(and only when) it provisioned, downgrading exactly the dangling-slice
reasons to warnings. Rationale, forced by the reviewer's probe (which
provisions with an empty native home and requires success): at
provision the manager establishes the correct link — its whole job for
the entry — while the native target is the operator's (a first login
heals a dangling link through the verified in-place writer). So a fresh
provision with no native credential yet succeeds loudly (fragment +
dangling warning on stderr via the existing `warning:` print), and
every later `StatusOf`/bare `Resolve` reports the home stale until the
native side heals. Repair of a dangling home (not a provision) leaves
the already-correct link untouched and fails with
`environment_repair_failed` carrying the dangling reason: there is
nothing repair can re-link, and the missing native file is the
operator's to restore.

The contrary assertion is fixed, not deleted:
`TestPiProvisionTargetsAgentRoot`
(`internal/envprofile/credential_link_test.go:359`) now provisions with
an empty native home (success + dangling warning asserted), asserts the
agent-root target and the frozen v1 pin, asserts bare resolve is stale
with dangling conflict-class wording (was: current), then writes the
native credential and asserts read-through + current.

### R2 — real TOML parsing for the Codex selector (fix)

`codexCredentialStore` (`internal/envprofile/managed.go:550`) used a
regexp matching only `key = "value"`; single-quoted and other valid
spellings read as absent ⇒ file ⇒ isolated admitted under the
operator-global store. The reader now parses the native `config.toml`
with the repo's existing `github.com/BurntSushi/toml` dependency
(import `:27`, already direct in `go.mod`; the `keyringSetting` regexp
is deleted) and reads the top-level key only:

- absent file or absent top-level key ⇒ `file` (isolated admitted);
  a same-named key nested in a `[table]` or behind a dotted prefix is
  not the selector and stays absent;
- `file`/`keyring`/`auto` in any valid spelling (double-quoted,
  single-quoted, multi-line basic/literal, indented, trailing comment)
  ⇒ that store (`keyring`/`auto` refuse isolated with
  `environment_isolated_unsupported` via the unchanged gate);
- any other parsed value — another string or any non-string TOML value
  (int, bool, …) — ⇒ `environment_credential_unsupported`;
- unreadable file ⇒ the existing `codex native config.toml is
  unreadable` refusal (unchanged); invalid TOML ⇒ `codex native
  config.toml is not valid TOML: <parse error>`, refusing fail-closed
  and naming the file, never absent. File-level failures name the
  file; value-level failures name the key, matching the existing
  unknown-selector diagnostic they share the class with.

### New rows (all driven through production entries on temp stores)

| Rework item | Test (`internal/envprofile/credential_link_test.go`) | Entry | Asserts |
|---|---|---|---|
| R1 probe, verbatim | `TestReviewerDanglingExpectedTarget` (:488) | `StatusOf` + `Resolve` | provision with empty native succeeds; status row non-current with conflict finding; bare resolve errors |
| R2 probe, verbatim | `TestReviewerLiteralCodexSelector` (:511) | `Resolve` (isolated) | single-quoted `keyring`/`auto` ⇒ isolated-unsupported; `ephemeral` ⇒ credential-unsupported |
| R1 repair half | `TestDanglingExpectedTargetRepairFails` (:542) | `Resolve` | live home current; native removed ⇒ stale with dangling wording; repair fails `environment_repair_failed` carrying the conflict reason, link untouched; native restored ⇒ repair converges, home current |
| R1 absent vs inspection | `TestCredentialLinkTargetInspectionFailure` (:600) | `Resolve` + `StatusOf` | file-as-agent-dir (ENOTDIR, deterministic Unix, root-proof) ⇒ stale/non-current with `cannot be inspected` + conflict, and NOT `does not exist` |
| R2 spellings | `TestCodexCredentialStoreTOMLSpellings` (:661) | `Resolve` (isolated) | single-quoted file admits; multi-line/indented/commented keyring/auto refuse; nested-table + dotted keys stay absent (admit); invalid TOML + int + bool fail closed |
| R2 shared side | `TestCodexMalformedStoreSharedRefuses` (:706) | `Resolve` (shared) | invalid TOML + int fail closed naming file/key before the first write; fix-and-retry provisions |
| R1 contrary fix | `TestPiProvisionTargetsAgentRoot` (:359, updated) | `Resolve` | provision-ok + dangling warning; agent-root target; frozen v1 pin; stale-while-dangling; live ⇒ current |

### New mutants (driver `/tmp/fc1_rev2_mutants.py`, shell `bash`)

| Mutant | Site | Type | Killer (must fail) | Witness (must pass) |
|---|---|---|---|---|
| N1 drop target-liveness stat | `checkPassthrough` (`managed.go:1592`) | deleting/silence | reviewer dangling probe, repair-fails row, row (e) | wrong-target Pi row still passes — the required expected-vs-wrong distinguisher |
| N2 inspection mapped to absence (`if true`) | same | narrowing | inspection row (negative `does not exist` assert) | absence row still passes |
| N3 drop provision carve-out | `repairUnderLock` (`:1948`) | deleting | reviewer dangling probe + row (e) (provision fails) | live-provision row still passes |
| N4 nested-table key counts as selector | `codexCredentialStore` (`:550`) | narrowing | spelling table (table/dotted admit cases refuse) | literal probe still passes |
| N5 non-string defaults to file | same | narrowing | spelling table + malformed-shared (int/bool admit) | double-quoted admission table still passes |
| N6 invalid TOML defaults to file | same | narrowing | spelling table + malformed-shared (invalid admits) | double-quoted admission table still passes |
| N7 restore rev1 regexp reader | same | deleting (R2 regression) | literal probe + spelling table | double-quoted admission table still passes |

All 7 killed by assertion failures (no build-failure kills); all 7
witnesses green; tree verified restored after every mutant (`git diff`
clean of mutant edits + `go build ./...` green). Two driver iterations
were needed and are disclosed: the first N2 anchor matched an earlier
identical `if` line (survived spuriously; re-anchored on the unique
`var reason string` context), and the first N7 replacement over-escaped
the regexp (failed to compile, then fixed to the faithful rev1
restore); both re-ran green.

### Existing-test triage (intent-preserving setup only)

No assertion was weakened. Tests about other behaviors that provision
linked homes now seed live native targets, or they would (correctly)
go stale under the new liveness row:

- `internal/envprofile/managed_test.go`: new `seedLiveNativeCredentials`
  helper (`:99`, codex + pi-agent + Linux-claude targets, harmless
  where unlinked); used by `TestResolveProvisionRepair`,
  `TestResolveDriftRepair` (fixture + per-env probe loop),
  `TestResolveFormats`; inline native-auth writes in
  `TestCodexKeyringAmbient` (ambient-vs-linked, not dangling) and
  `TestRepairNeverRefreshesSeeds` (repair must converge).
- `internal/envprofile/status_test.go`:
  `TestStatusShadowAcknowledgment` seeds the pi agent credential (the
  shadow row needs an otherwise-current home).
- `internal/envprofile/credential_link_test.go`:
  `TestCredentialLinkDirectory` subtests seed live targets (empty-dir
  re-link must converge post-repair).
- `cmd/curator/profile_test.go`: new `writeNativeCredentials` helper
  (`:58`) writing all three native targets under the redirected
  `CODEX_HOME`/`PI_CODING_AGENT_DIR`/`CLAUDE_CONFIG_DIR`; called by
  `TestEnvResolveRepairEmitsFragment`, `TestEnvStatusMatrix`,
  `TestEnvResolveAcceptsAliases`, `provisionedEnvMatrix`
  (`hook_posture_test.go`), and
  `TestEnvStatusReportsShellHookTrustPosture` (`hook_test.go`).

### Corrected claims (CHANGELOG + troubleshooting)

Revision 1 claimed dangling coverage it only implemented for the
mis-targeted string case. The entries now state the implemented
contract: expected-target dangling is reported detached;
provisioning with no native credential yet succeeds carrying the
dangling state as a warning; later resolves report stale until the
native side heals (CHANGELOG `Fixed` + troubleshooting
`environment_credential_conflict` cause/remedy, incl. the heal remedy:
log in natively or in the managed home and re-run). The Codex entries
now state TOML parsing, top-level-only key, nested-key-is-absent, and
fail-closed malformed config (CHANGELOG `Changed` + troubleshooting
`environment_credential_unsupported`).

### Revision-2 validation (shell `bash`, real exit codes)

- `go build ./...` → 0; `GOOS=linux go build ./...` → 0;
  `GOOS=windows go build ./...` → 0
- `go vet ./...` → 0
- `gofmt -l cmd internal` → clean; `git diff --check` → clean
- `golangci-lint run ./internal/envprofile/... ./internal/envregistry/...
  ./cmd/curator/...` → 0 issues
- `go test -count=1 ./internal/envprofile/` → ok (316s, full package
  incl. 5 new tests + 2 committed probes + updated row (e))
- `go test -count=1 ./internal/envregistry/ ./internal/envmarker/` → ok
- `go test -count=1 -run
  'TestEnv|TestCmdEnv|TestEnvalias|TestEnvconfig' ./cmd/curator/` → ok
  (170s; covers the fragment/matrix/alias/hook-posture CLI rows)
- 7/7 rev2 mutants killed, 7/7 witnesses green (above)
- No new credential-byte reads: the liveness stat is `os.Stat` on the
  target path; the TOML parse reads only native `config.toml` (as in
  rev1); `grep ReadFile.*auth` over the production diff → 0 hits

### Anomaly (shared-host contention, no code impact)

Mid-validation, `cmd/curator` runs blocked pre-`TestMain` with zero CPU
and empty logs, then died with SIGKILL at ~85s (three occurrences), and
one full-mask run hit the 600s `go test` timeout under load average
~6. Cause: `cmd/curator/status_test.go:43` `TestMain` acquires the
host-wide `curator-host-goroot-test-lock-v1` file lock
(`internal/testtoolchain/lock.go`, polite 10ms poll, no kill logic)
before any test runs; a concurrent agent's long test run on this
shared host held it. After the holder finished, the single test passed
in 3s and the full mask went green in 170s. Nothing in this change can
block (no new locks, loops, or subprocesses). No action needed beyond
this note.

### Revision-2 bounds (amendments to rev1 bounds)

- The inspection-failure row (`TestCredentialLinkTargetInspectionFailure`)
  is Unix-only (`t.Skip` on Windows with reason): Windows maps a stat
  through a file to path-not-found, which the code correctly reports
  with the absence wording there. Enforcement on Windows is the
  absence branch, covered by the other dangling rows.
- Linux behavior verified by code-path identity (the liveness check is
  platform-agnostic; only `PassthroughFor` differs) plus `GOOS=linux`
  build/vet and the Linux-claude seeding in fixtures — not by a Linux
  test run from here. The ubuntu lane owns that.
- F-C3 still owns `environment_passthrough_unreadable` for link-level
  `lstat`/`readlink` failures (unchanged); the interim target-level
  inspection diagnostic here is conflict-class without the detached
  label, ready to be re-homed. F-C2 (migration) and F-S1 (extended
  record) bounds unchanged; the marker schema is untouched (no
  `envmarker` changes; row (e) still pins path+strategy-only).
- Provisioning emits a fragment while carrying a dangling warning —
  the only fragment-emitting path for a dangling home; every steady
  resolve (bare or `--repair`) refuses until the native side heals.
  If the re-reviewer wants provisioning itself to fail closed here,
  that contradicts the committed reviewer probe (provision with empty
  native must succeed) and needs an orchestrator ruling, not a silent
  change.

### Revision-2 files changed (delta on rev1)

- `internal/envprofile/managed.go`: TOML selector reader (BurntSushi,
  top-level key; regexp deleted); target-liveness stat + absent-vs-
  inspection reasons in `checkPassthrough`; `verification.dangling`
  slice; `tolerateDanglingAtProvision` + provision-only call.
- `internal/envprofile/credential_link_test.go`: 2 committed reviewer
  probes + 4 new tests + row (e) contrary fix + directory-test seeding.
- `internal/envprofile/managed_test.go`: `seedLiveNativeCredentials` +
  5 call sites + 2 inline seedings.
- `internal/envprofile/status_test.go`: shadow-test seeding.
- `cmd/curator/profile_test.go`: `writeNativeCredentials`; call sites
  in `env_test.go`, `envalias_test.go`, `hook_posture_test.go`,
  `hook_test.go`.
- `CHANGELOG.md`, `docs/troubleshooting.md`: corrected dangling/TOML
  contract (above).

## Revision 3 (rework-2: mis-targeted vs dangling-to-declared split)

Revision 2 gate FAILED on both ubuntu lanes (run 35685549418):
`TestResolveClaudeProjectEntry` (`managed_test.go:481`) and
`TestClaudeSeedMergePreservesToolState` (`status_test.go:412`) failed
with `environment_repair_failed: passthrough entry .credentials.json
is detached: link target … does not exist
(environment_credential_conflict)`. Cause: on Linux `claude_code` is
file-link and the native `~/.claude/.credentials.json` appears only
after the first login, so the rev2 liveness stat made every repair of
a correctly provisioned pre-login home fail. Rework-2 refines brief
R2: the operator's Pi case (recorded target ≠ declared native path)
stays a stale detached-with-conflict state, while dangling-to-declared
(recorded target == declared path, target absent) is the normal
provisioning shape and succeeds loudly. Both stay reported, never
silent; they differ in failure vs warning.

### Fix 1 — detached-pending warning, never stale

`checkPassthrough` (`internal/envprofile/managed.go:1541`) now splits
the post-match target stat:

- target absent (`os.IsNotExist`) ⇒ warning `passthrough entry
  <path> is detached-pending: link target <native> does not exist yet
  — log in to <adapter> to populate it`. No conflict diagnostic, no
  stale reason: provisioning, bare resolve, and `--repair` all
  succeed, carrying the finding as a warning. The wording is
  stat-only (never a credential-byte read).
- target stat fails otherwise ⇒ unchanged stale reason `… cannot be
  inspected: <err> (environment_credential_conflict)`; absent and
  inspection-failed stay distinct facts (EACCES ⇒ conflict).

`verification.dangling` and `tolerateDanglingAtProvision` are deleted:
there is no longer a dangling reason to downgrade anywhere, at
provision or steady state. `StatusOf` surfaces the pending state in
the row's warnings (current with a loud finding); `Resolve` surfaces
the identical text in result warnings with a fragment.

### Fix 2 — repair re-points a recorded mis-targeted symlink

`ensureCredentialLink` (`managed.go:1083`) takes a `recorded` flag
(path present in the prior marker's passthrough set, computed in
`finalizeMarker` `:1001`). A symlink to an unexpected target at a
RECORDED path is now removed and re-linked at the declared store —
link-only, native bytes untouched — so the operator's Pi home heals
under `--repair`. A symlink to an unexpected target at an UNRECORDED
path stays a foreign link and refuses with
`environment_credential_conflict` naming the path, as do a regular
file, a non-empty directory, and unreadable link state.
`removeStaleCredentialLinks` (shared→isolated) is unchanged: it
unlinks only a recorded symlink still targeting the declared store.

Spec tension disclosed: environments §7.4/§10.1 say repair MUST NOT
re-point a symlink to an unexpected target. Rework-2 (binding)
explicitly refines this for recorded credential links — the path is
ours, only the pointer moves, bytes are never touched — while
foreign (unrecorded) links keep the MUST-NOT refusal. The
troubleshooting entry documents the recorded/foreign split.

### Row updates (all driven through production entries on temp stores)

| Rework item | Test (`internal/envprofile/credential_link_test.go`) | Entry | Asserts |
|---|---|---|---|
| mis-targeted repair | `TestDanglingPiLinkReportedDetached` (updated: repair half inverted) | `StatusOf` + `Resolve` | status non-current with detached+conflict finding naming path + agent target; bare resolve stale same; repair SUCCEEDS and the link now targets the agent root; native bytes identical; nothing created at the old target; home current |
| pending provision | `TestPiProvisionTargetsAgentRoot` (updated) | `Resolve` | provision succeeds; warnings carry `detached-pending` + `does not exist yet` and NOT the conflict diagnostic; bare resolve succeeds with the same finding; frozen v1 pin unchanged; native present ⇒ finding clears, link reads through |
| reviewer probe (revised, not verbatim) | `TestReviewerDanglingExpectedTarget` | `StatusOf` + `Resolve` | pending finding REPORTED in status (findings+warnings) without conflict; bare resolve succeeds carrying it as a warning. The probe's original stale-plus-conflict expectation now belongs to the mis-targeted state only |
| pending repair | `TestDanglingExpectedTargetRepairSucceeds` (renamed from `…RepairFails`) | `Resolve` | live home current; native removed ⇒ bare resolve succeeds with pending warning (no conflict); repair succeeds with the warning, link untouched; native restored ⇒ finding clears |
| recorded-only bound | `TestCredentialLinkUnrecordedSymlinkRefuses` (new) | `Resolve` (keyring→file turn) | foreign symlink at a newly-wanted unrecorded path refuses with conflict naming the path; link untouched |
| inspection still stale | `TestCredentialLinkTargetInspectionFailure` (unchanged) | `Resolve` + `StatusOf` | `cannot be inspected` + conflict, still stale/non-current; NOT absence wording |

Narrowing pair: `TestDanglingPiLinkReportedDetached` ×
`TestDanglingExpectedTargetRepairSucceeds` — a mutant collapsing the
two states fails one of them (stale-plus-relink on dangling breaks
the pending success; pending-warning on mis-targeted breaks the Pi
staleness). A mutant dropping the `recorded` check fails
`TestCredentialLinkUnrecordedSymlinkRefuses`; one refusing every
re-point fails the Pi repair.

### New mutants (driver `/tmp/fc1_rev3_mutants.py`, shell `bash`)

| Mutant | Site | Type | Killer (must fail) | Witness (must pass) |
|---|---|---|---|---|
| P1 pending reported stale (warning→reason) | `checkPassthrough` (`managed.go:1586`) | narrowing (collapse) | pending repair/provision/probe rows | Pi mis-targeted row still stale |
| P2 pending silenced (absent branch gated off) | same | deleting/silence | pending rows (warning assertions) | inspection row still passes |
| P3 always relink (recorded refusal deleted) | `ensureCredentialLink` (`managed.go:1104`) | narrowing (bound) | unrecorded-symlink refusal | Pi relink still passes |
| P4 always refuse re-point (`!recorded`→`true`) | same | narrowing (bound) | Pi relink repair | unrecorded refusal still passes |
| P5 inspection mapped to pending (reason→warning) | `checkPassthrough` (`managed.go:1591`) | narrowing | inspection row (stale expected) | pending rows still pass |

All 5 killed by assertion failures (exit 1 on killers, exit 0 on
witnesses, single driver run, no iterations); tree verified restored
after every mutant (sha256 digest match + `go build ./...` green).

### Corrected claims (CHANGELOG + troubleshooting)

The rev2 claim "later resolves report it stale until the native side
heals" is superseded: dangling-to-declared is never stale. CHANGELOG
`Fixed` now states the three-state contract (mis-targeted stale +
re-point; pending warning + success; inspection-failed stale
conflict) and the recorded-only re-point bound; `Changed` notes Pi
repair re-points. Troubleshooting
`environment_credential_conflict` lists only refusing states
(regular file, unrecorded symlink, non-empty directory), documents
the recorded re-point, and adds a "Not this error"
detached-pending paragraph. Test-seed comments that said the
liveness row "keeps the home stale" now say the seeds keep the
pending finding out of unrelated tests (seeds themselves unchanged).

### Revision-3 validation (shell `bash`, real exit codes)

- `go build ./...` → 0; `GOOS=linux go build` (envprofile+registry) →
  0; `GOOS=windows go build` (envprofile+registry) → 0
- `go vet ./internal/envprofile/ ./internal/envregistry/` → 0
- `gofmt -l` → clean; `git diff --check` → clean
- `golangci-lint run ./internal/envprofile/... ./internal/envregistry/...`
  → 0 issues
- `go test -count=1 ./internal/envprofile/ ./internal/envregistry/` →
  ok (envprofile 371s incl. 1 new + 4 updated rows; envregistry ok)
- `go test -count=1 -run 'TestEnv|TestCmdEnv|TestEnvalias|TestEnvconfig'
  ./cmd/curator/` → ok (149s)
- focused credential rows (9 tests incl. both narrowing partners) → ok
- 5/5 rev3 mutants killed, 5/5 witnesses green (above)
- `grep ReadFile.*auth` over the production diff → 0 hits
  (stat-only liveness; TOML parse reads only native `config.toml`)

### Revision-3 bounds

- The ubuntu-lane failures are addressed by construction (pending is
  a warning, repair succeeds); the Linux run itself is owned by the
  gate lanes — no Linux host here. Darwin full-package run is the
  local evidence, plus `GOOS=linux` build and the platform-agnostic
  liveness path (only `PassthroughFor` differs per GOOS).
- The mis-targeted re-point is the one intentional deviation from the
  spec's MUST-NOT-re-point sentence, per binding rework-2; flagged
  for the reviewer to confirm against the spec revision that adopts
  it.
- F-C2 (explicit migration of native-side moves), F-S1 (extended
  record), F-C3 (`environment_passthrough_unreadable` for
  link-level failures) bounds unchanged; marker schema untouched (no
  `envmarker` changes; row (e) still pins path+strategy-only).

### Revision-3 files changed (delta on rev2)

- `internal/envprofile/managed.go`: pending-warning split in
  `checkPassthrough`; `dangling` slice + `tolerateDanglingAtProvision`
  deleted; `recorded` flag + re-point in `ensureCredentialLink`;
  recorded-set computation in `finalizeMarker`.
- `internal/envprofile/credential_link_test.go`: Pi repair-half
  inversion; row (e) pending rewrite; reviewer probe revision;
  `…RepairFails`→`…RepairSucceeds` rename+inversion; new
  unrecorded-symlink refusal row.
- `internal/envprofile/managed_test.go`,
  `internal/envprofile/status_test.go`, `cmd/curator/profile_test.go`:
  seed-comment corrections only (no behavior change).
- `CHANGELOG.md`, `docs/troubleshooting.md`: three-state contract
  (above).

## Revision 4 (rework-3: Windows skip-class gate)

Revision 3 gate FAILED only on `Test (windows-latest)` (run
35688466491), tier 2 of `.github/ci/platform-case-gate.sh`:

`FAIL skip with an unrecognised reason on windows: internal/envprofile ::
TestCredentialLinkTargetInspectionFailure`

Linux/macOS lanes were green. Cause: the row's Windows skip reason
("Windows maps a stat through a file to path-not-found, not an
inspection error") matched no entry in
`.github/ci/skip-classes.tsv`. Reproduced locally with the real gate
script against a synthetic Windows `go test -json` stream (empty
tier-1 ledger to isolate tier 2): old reason → exit 1 with the
byte-identical FAIL line; see validation below.

### Fix (one line, test-only, no vocabulary change)

`internal/envprofile/credential_link_test.go:692` — the skip reason is
now the sibling envprofile host-capability phrasing:

`this host cannot create an uninspectable link target: Windows maps a
stat through a file to path-not-found, so the inspection diagnostic
runs on the unix runners`

It classifies `host-capability` (policy `allow`) via the existing
`this host cannot create` regex — the same regex and class as the
sibling rows' `this host cannot create symlinks: …`
(`pathkind_test.go:105`, `takeover_test.go:145`) — with the Windows
stat-mapping tail kept so the reason stays truthful: on Windows the
fixture (a file where the agent directory was) stats as
path-not-found, so the inspection-error artefact the row needs cannot
be produced there. No `.tsv` file was touched (no new vocabulary, no
ledger row: like every other new credential-link row, this case is
not ledger-tracked and needs no `must_run_on` entry — the unix lanes
run it, Windows records the classified skip).

The literal sibling text (`… symlinks: …`) was deliberately NOT
copied: this row fails to produce no symlink, and a false reason
would misreport the skip. Same class, same matched entry, truthful
specifics.

### Other new rows checked

The fixed skip is the only Windows skip this task adds: `grep
Skip|GOOS` over `credential_link_test.go` hits only lines 691–692,
and the tracked-file diffs add no `Skip`/`GOOS` line. All other new
rows run (no skip) on all three platforms; the Windows runner set
creates symlinks (per the ledger's own link-capability note), so no
new row can hit an unclassified skip there.

### Revision-4 validation (shell `bash`, real exit codes)

- Real `platform-case-gate.sh`, `CI_GATE_GOOS=windows`, synthetic
  stream, empty tier-1 ledger: old reason → exit 1
  (`FATAL-unclassified`, byte-identical FAIL line to run 35688466491);
  new reason → exit 0 (`allowed-host-capability` in
  `skips-observed.tsv`).
- `gofmt -l internal/envprofile/` → clean; `git diff --check` → clean.
- `go vet ./internal/envprofile/` → 0; `GOOS=windows go vet
  ./internal/envprofile/` → 0 (also compiles the edited test for
  Windows).
- `go test -count=1 -run TestCredentialLinkTargetInspectionFailure
  ./internal/envprofile/` → PASS (1.8s, darwin runs the row, no skip).
- `go test -count=1 ./internal/envprofile/ ./internal/envregistry/`
  → ok (envprofile 324.4s incl. the fixed row; envregistry 0.9s).
- Lint: rev3 tree was `golangci-lint` clean on these packages; this
  revision edits one string literal in one test, so no new lint
  surface exists (gofmt+vet above are the applicable gates).

### Revision-4 bounds

- Windows/CI-lane ownership unchanged from rev3: the gate pass is
  proven locally against the real gate script and committed
  `.tsv` tables; the hosted `windows-latest` run owns the final
  confirmation.
- No production, CHANGELOG, troubleshooting, or ledger change in this
  revision — the one-line skip reword is the complete delta.
