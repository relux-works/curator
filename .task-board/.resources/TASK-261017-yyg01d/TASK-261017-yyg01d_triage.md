# Install and takeover preservation triage

Task: TASK-261017-yyg01d — preservation-triage. Source: a field report.

Research only. No repository code, tests, documentation, configuration, branch or commits changed. Scratch harnesses and outcomes are task-local temporary artifacts. The requested deliverable is ready for review, not an implementation of the proposed fixes.

## Identities and safety

- Implementation: fresh `origin/main`, `fcd785ba93c5d00e2e80615f8add2baa93ccd73d`; Story workspace HEAD equals that commit. `git fetch origin main` exited 0; no checkout, commit, rebase or integration performed.
- Spec: fresh `origin/main`, `1059b94b4135089d842bf017070b8ab47d79a536`. Read `protocol/core.md`, `protocol/environments.md`, `protocol/skillfile-sources.md` and the relevant `profiles/manager.md` clauses from that exact Git object. Spec fetch exited 0. This is the requested current-main comparison; the implementation's immutable release pin remains rc.8 (`f8c405aa3ad0a39d260c2ed93684e55c5a346359`).
- Build: `go build -o .temp/TASK-261017-yyg01d/curator ./cmd/curator`, exit 0. Toolchain: Go 1.25.5, darwin/arm64. Curator help readiness exited 0. Build used the host toolchain; every reproduction used the locally built binary in a private HOME.
- Every fixture has a private mode-0700 HOME and scratch Git project under the task temp directory. The driver checks HOME containment and mode before **every** subprocess. It starts each subprocess with a newly constructed environment, not inherited runtime homes or credentials. All four native home variables and XDG paths point inside scratch. Global/system Git configuration is disabled; Git permits only `file` transport; HTTP/HTTPS/all proxies point at a closed loopback port. All sources are local path packages. No network source, authentication or agent executable is used. This is not a claim of OS-level network isolation.
- The diagnostic build-cache sweep attempted process-executable enumeration and could not enumerate one executable; it explicitly retained all builds. This warning is incidental and does not establish any deletion protection.
- Public attachments replace machine-local absolute prefixes with `$WORKTREE`, `$TOOLCHAIN`, `$MODULE_CACHE`, `$OPERATOR_HOME` or `$SYSTEM_TEMP`. These are redactions, not runnable literal paths. No field-report product material or real credentials are included.

## Results

| Item | Current-main reproduction and exits | Verdict |
| --- | --- | --- |
| A | Resolve 0, initial install 0, reinstall 0. All 3 unmanaged entries disappear: skill folder, notes folder, loose file. Native unmanaged symlink remains but becomes dangling. Managed sample remains. No deletion notice or named conflict. | Implementation bug; first priority. The exact field-report output was not identical: this build printed `installed` plus permissive-posture/cache-sweep warnings, rather than only `up-to-date`. Silent user-content deletion reproduces. |
| B | Without manifest: `skill check` 0; resolve 0; install **1**, `sample: install marker is invalid for schema 5`. Adding minimal schema-4 manifest: check 0, refresh 0, install 0. | Spec gap with an observable installation defect. Optional-manifest package admission and mandatory marker representation contradict each other. `skill check` acceptance is correct under core §4; the source marker validator matches its schema. A coordinated spec/implementation fix is required. |
| C | `.agents/.instructions/hand.md` and `.agents/README.md` survive install 0. A pre-existing **selected** `.agents/skills/sample/` without marker loses its manual notes and has SKILL.md replaced, install 0. Native adapter skill collision refuses, install **1**, unchanged manual bytes and no canonical publication. Unmanaged native CLAUDE.md refuses without takeover; authorized takeover 0 preserves its exact bytes in backup. | Mixed: canonical selected-directory overwrite is an implementation bug, separate from A's stale sibling deletion. Wholesale `.agents/` removal and loss at the tested native takeover path did not reproduce. Native refusal/backup are specified behavior. |
| D | First `profile install ./alpha` without `--use`: 0, activates alpha and materializes all 4 native homes; `profile list` marks alpha current. Installing beta without `--use`: 0, alpha stays current, beta remains idle. | Specified behavior. Post-operation diagnostic is clear; operator-facing pre-operation docs/help omit the first-install exception. |
| E | With Claude marker removed in scratch and manual CLAUDE.md seeded, `profile use beta`: **1**, Claude unchanged, Codex/OpenCode/Pi switch to beta; recorded current remains alpha. `profile use alpha`: **1**, restores the other 3 to alpha, Claude still refuses. `profile use beta --takeover`: 0, all 4 converge, beta current, manual Claude bytes backed up. | Specified per-entry partial success, not a missing global rollback. CLI names the conflict and `profile_use_partial`, and states recorded current is unchanged. Operator docs do not plainly explain the persisted partial state and recovery. |

The native roots in these results are scratch roots. No real home or real project was passed to Curator.

## Exact fixture and command protocol

`TASK-261017-yyg01d_reproduce.py` contains the complete setup, environment, command arguments, mutation steps, and before/after SHA-256 inventories for the authoritative run (`sandbox-05`). `TASK-261017-yyg01d_supplement.py` adds the two C collision fixtures (`sandbox-06`). `TASK-261017-yyg01d_verify-observations.py` checks captured inventories without rerunning Curator. The combined sandbox log preserves every subprocess command, stdout, stderr and actual exit code, including expected non-zero refusals.

Common setup, performed separately for each case:

```text
mkdir scratch HOME (0700), scratch project, scratch TMPDIR
write .gitignore: .agents/, .claude/skills/, .codex/skills/
git init -q -b main                                      # 0
curator bootstrap -non-interactive --skills-root <scratch>/sources  # 0
```

The schema-2 Skillfile is exactly:

```json
{"schema_version":2,"agents":["codex_cli","claude_code"],"sources":{"local":{"path":"./agents"}},"skills":[{"from":"local","directory":"skills","include":["*"]}]}
```

`skills/*` in the brief is collection shorthand; a literal glob in `directory` is invalid. The source package is `agents/skills/sample/`, with valid name/description SKILL.md frontmatter. Except for B's negative fixture it has:

```json
{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{}}}
```

Commands after fixture creation:

```text
A: curator project resolve .                             # 0
   curator install .                                     # 0
   seed .agents/skills/unmanaged-skill/SKILL.md,
        .agents/skills/unmanaged-notes/notes.md,
        .agents/skills/loose.txt
   link .codex/skills/manual-link -> ../../.agents/skills/unmanaged-skill
   capture BEFORE; curator install .; capture AFTER      # 0 (destructive witness)

B: curator skill check ./agents/skills/sample             # 0, no manifest
   curator project resolve .                             # 0
   curator install .                                     # 1 (reported defect)
   write minimal manifest above
   curator skill check ./agents/skills/sample             # 0
   curator project refresh .                             # 0, explicit new source identity
   curator install .                                     # 0 (positive control)

C1: seed .agents/.instructions/hand.md and .agents/README.md
    curator project resolve .; curator install .          # 0, 0; preserved
C2: seed .agents/skills/sample/SKILL.md and private-note.md, no marker
    curator project resolve .; curator install .          # 0, 0; manual content lost
C3: seed .codex/skills/sample/SKILL.md, no adapter ledger
    curator project resolve .; curator install .          # 0, 1; native bytes preserved

D: curator profile install ./alpha                       # 0, no --use, first install
   curator profile list                                  # 0, alpha current
   curator profile install ./beta                        # 0, no --use, later install
   curator profile list                                  # 0, still alpha current

E/C-native: remove only scratch .claude/.agent-environment.json
   write scratch .claude/CLAUDE.md with unique manual bytes
   capture BEFORE
   curator profile use beta                              # 1, expected unmanaged refusal
   curator profile list                                  # 0, alpha current
   capture failed-use state
   curator profile use alpha                             # 1, other 3 restored; conflict persists
   capture recovery state
   curator profile use beta --takeover                    # 0, backup + convergence
   capture takeover state; curator profile list           # 0, beta current
```

Each context package alpha/beta has `agent-context.json` schema 1, package name alpha/beta, version `1.0.0`, and `context.modules=[{"path":"a.md"}]`; `context/a.md` contains the package-specific text. The driver captures file hashes and symlink text/resolution, so E is not inferred solely from printed `switched` lines. Native linked targets are resolved against the **captured** rendered-file hashes. Observed switching ratio: 3/4 successful, 1/4 refused, 4/4 native root states inspected. A checks 3/3 seeded unmanaged entries plus one unmanaged link; this is not universal filesystem-type coverage.

## Governing clauses and diagnostics

Spec links below are immutable at the spec revision above:

- [Skillfile-sources §2](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/skillfile-sources.md#2-physical-boundaries-and-admitted-local-inputs): authored `agents/skills` versus generated `.agents/skills`, preservation of unmanaged-path conflict protection, no implicit takeover from source-policy admission. Being inside a managed-output root is not evidence of individual ownership. This governs A and C's canonical destination collision. The conservative supported fix is refusal naming unmanaged paths, before publication. Generated-output pruning in this section concerns **source enumeration**, not permission to erase arbitrary destination content.
- [Core §10–§11](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/core.md#11-adapter-ledger): each installed node has an install marker; adapter roots remove only preceding-ledger entries and refuse unmanaged collisions. Core §11 directly governs C3 and the native unmanaged link in A; it is not a grant to delete unknown canonical-store siblings. Manager §3 locates canonical contexts under `.agents/skills/<name>`; manager §5 adopts the core adapter ledger.
- [Manager §2.5](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/profiles/manager.md#25-private-build-and-serialized-publication): serialized publication, stale **managed** removals, consumer ledger last, rollback. C3 refused before any canonical publication. A/C2 should refuse before destructive publication too. Do not apply the special `global add/install` rollback rule to `profile use`.
- [Core §4](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/core.md#4-skill-manifest) explicitly makes the manifest optional for pure context skills. [Skillfile-sources §1](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/skillfile-sources.md#1-acquisition-and-selection) preserves ordinary package admission. However §4 requires source marker schema 5 and its migration table says actual manifest version 1–8. The [normative source marker schema](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/schemas/skillfile-sources-v1/install-marker-v5.schema.json) has `skill_schema_version.minimum=1, maximum=8`. There is no no-manifest representation. B is a real spec conflict, not authority to reject otherwise valid pure context packages or silently invent a manifest.
- [Environments §8.3 and §8.3.1](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md#83-ledger-discipline-and-backups): environment markers delimit ownership, unmanaged writes refuse, takeover backups are versioned, and writes replace entries rather than follow links. [§9.5](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md#95-onboarding) requires notice and backup before authorized takeover, only for touched files. These support native C/E results; they do not authorize takeover of every file under `.agents/`.
- [Environments §9.1](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md#91-installation-resolution-and-audit): first profile activates when machine current is absent, even without `--use`, and activation must be reported. Later installs require `--use` to switch. D matches; observed `installed and activated profile alpha` is clear after the mutation. README's environment command section and `profile install` flag help (`--use: activate the installed profile`) do not explain the exception beforehand.
- [Environments §9.2](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md#92-current-profile-switching-update-and-removal), §9.7: switching attempts every entry, is atomic per entry, leaves successful entries changed on partial failure and current unchanged, reports `profile_use_partial`; explicit retry completes the scope. E matches. The CLI says partially switched/current unchanged and names the unmanaged conflict; it does not print a recovery command in that failure. README and the inspected `docs/cli.md`, `docs/environment-config.md`, and `docs/second-operator.md` do not plainly explain partial persistence/no fleet-wide rollback. Retrying either profile still needs the unmanaged conflict resolved or explicitly backed-up takeover, as the experiment demonstrates.

## Fix leaves and order

These are proposed bounded implementation briefs, not executed fixes or newly created board tasks.

1. **A — P1: preserve unowned canonical-store siblings during stale cleanup** (size 3, medium risk).
   - Required behavior: under the existing serialized publication path, identify stale entries by valid preceding manager ownership, not basename/type or presence of SKILL.md. Refuse the operation naming every discovered unowned entry that the plan would delete; preserve its bytes and foreign native links. Refuse unreadable/invalid ownership evidence without treating it as absence or permission. The same helper serves project/global/hybrid stores: confirm each caller's contract before broadening the change.
   - Negative controls: CLI regression `TestInstallPreservesUnownedStaleStoreEntries` covers unmanaged skill folder, notes-only folder, loose file; verify rejection diagnostic, exact bytes/link resolution, unchanged preexisting managed state and ledger. Neighbor tests cover malformed/unreadable marker, link/special entry and mixed owned/unowned candidates. Test a selected managed node and an actual owned stale node together, so reject-everything cannot pass.
   - Positive controls: owned stale node is removed, selected node installs/remains readable, ordinary clean reinstall succeeds; foreign unrelated native link remains valid when its unowned target is preserved.
   - Production sites/packages: `internal/install/install.go:stageProjectTargets`, `internal/install/targets.go:stageStaleSkillRemovals`; review `global.go:stageGlobalTargets`, marker reads, `internal/staging` / `internal/transaction`, and adapter behavior. Prefer not to change native-link ownership.
   - Risk: accidental deletion through invalid ownership classification, over-refusal of valid stale cleanup, cross-scope callers. Preserve transaction and GC invariants.

2. **C2 — P1: refuse an unowned selected canonical destination** (size 3, medium risk).
   - Required behavior: before `stageNode` can plan a directory replacement, prove preceding ownership of an existing canonical destination; absent/unreadable/invalid ownership cannot authorize replacement. Refuse with the conflicting path and no partial changes. Do not add a generic recursive `.agents` takeover flag. Keep native backup semantics intact.
   - Negative controls: `TestInstallRefusesUnownedSelectedCanonicalDestination` seeds manual SKILL.md and unrelated note at the selected name without a marker; assert named refusal, byte-for-byte preservation, unchanged adapters/ledger. Add invalid/unreadable marker and canonical link variants. Keep `TestDraftUnmanagedAdapterDestinationRefuses` under its original name.
   - Positive controls: absent destination installs, a valid preceding owned node updates; unrelated `.agents/.instructions/hand.md` survives. Preserve native CLI refusal and authorized-backup control.
   - Production sites/packages: `internal/install/targets.go:stageNode`, `internal/marker` ownership/read-state API, `internal/install` caller planning; `internal/adapters` only if needed to share existing refusal semantics.
   - Risk: confusing currency/drift with ownership; refusing legitimate managed updates, or trusting mere marker filename existence. Ownership definition should be explicit, and marker is not cryptographic authorization.

3. **B — P2: define no-manifest source-marker representation, then align writers/readers** (spec size 2; implementation size 3, medium risk).
   - Required behavior: pure context skill without manifest passes check, resolve, install and reinstall; emitted marker must match a documented normative schema. First settle the no-manifest representation in curator-spec (recommended: explicit documented absent-manifest version sentinel with narrowly validated semantics). A fabricated schema-1 manifest or silently weakening every marker version is not a clean fix. The choice belongs to the format owner; this research task does not need to make that change.
   - Negative controls: `TestSourceMarkerAbsentManifestIsContextOnly` rejects that representation when commands/runtime/build declarations imply a manifest, rejects out-of-band versions, and preserves malformed-modern-manifest refusal instead of fallback. `TestInstallManifestlessContextThroughCLI` checks actual install, not only marker helpers, with no manifest as positive admission and malformed present manifest as negative control; retain minimal-manifest positive control and explicit-refresh binding.
   - Production sites/packages: spec core §4, skillfile-sources §4 and source marker schema/fixtures; `internal/skillspec`, `internal/install/draftruntime.go:buildDraftMarker`, `internal/marker/marker.go:valid`, source-lock/snapshot admission tests, `internal/skillcheck` and CLI integration. No edit to `skill check` to reject all manifestless packages.
   - Risk: wire compatibility, marker-version bands, synthetic identities and lock binding. Schema/conformance decision precedes implementation; this is recorded as a follow-up dependency, not a blocker to delivering triage.

4. **D/E — P3 documentation/diagnostic leaf** (size 1, low risk): document first-install activation, idle later install, 4-runtime partial persistence/current unchanged, and conflict resolution before retry/explicit takeover. Add examples matching observed CLI outputs; do not change specified activation/partial-success semantics. Touch README, docs/cli.md and profile help if useful. A plain-language example is the positive control; a misleading promise that absence of `--use` always means no writes, or that failed `profile use` rolls back all runtimes, is the negative documentation counterexample. No runtime product fix is justified by these observations.

## Verification and evidence limits

All commands were standalone gate processes with redirected logs, without `tee` or pipelines. Actual exit codes:

| Command | Exit | What it establishes |
| --- | ---: | --- |
| Local build above | 0 | Current-main binary compiles on darwin/arm64 |
| `python3 .temp/TASK-261017-yyg01d/reproduce.py` (sandbox-05) | 0 | Driver ran all A–E commands and captured their individual exits; does not turn expected-red install/use into passes |
| `python3 .temp/TASK-261017-yyg01d/supplement.py` | 0 | C collision commands plus A loss/link, unrelated-file preservation and exact native-backup assertions |
| `python3 .temp/TASK-261017-yyg01d/verify-observations.py` (corrected verifier) | 0 | Captured content/hash checks for 4 native roots, retry behavior, canonical loss and native refusal |
| `go test ./internal/install -run '^(TestDraftUnmanagedAdapterDestinationRefuses)$' -count=1 -v` | 0 | Existing native adapter refusal test, 1 test |
| `go test ./internal/envprofile -run '^(TestUseWithoutTakeoverRefusesUnmanaged\|TestUseTakeoverBacksUpAndNotifies\|TestUsePartialRecordsNothing)$' -count=1 -v` | 0 | Existing refusal, takeover and partial-current tests, 3 tests; actual shell regex uses unescaped `\|` alternatives as shown in the commands attachment |
| `go test ./cmd/curator -run '^(TestProfileInstallListUse\|TestProfileInstallUsePartialLeavesCurrent)$' -count=1 -v` | 0 | Existing CLI first activation and partial-current tests, 2 tests; see exact invocation attachment |
| `git diff --exit-code` | 0 | No tracked worktree delta; reviewer tests remain untouched |

The test invocations used `env -i`, private mode-0700 test HOME/TMPDIR/GOCACHE, the existing Go toolchain/module cache, `GOPROXY=off`, `GOSUMDB=off`, and `GIT_ALLOW_PROTOCOL=file`. Their exact full environment commands are in `TASK-261017-yyg01d_commands.txt`; complete stdout/stderr logs are attached. Existing host-test language is Go. Python is only a task-local subprocess/inventory harness, using the standard library for filesystem bytes, SHA-256 and redaction; no new product language, test library or persistent tooling was introduced.

Diagnostic attempts are retained, not counted as green evidence: sandbox-01 used an invalid literal `directory: skills/*`; sandbox-02 hit the gitignore skip (CLI 0 but no installation); sandbox-03 lacked an explicit source lock (install 1). Their drivers exited 1 on fixture setup afterward. Sandbox-04's driver exited 0 but its E fixture removed no ownership marker, so `profile use` succeeded and is **not** evidence of an unmanaged-conflict path. The first inventory verifier exited 1 because linked entries have link metadata rather than a direct hash; the corrected verifier resolves links inside the captured snapshot. The attempted optional read of nonexistent `protocol/manager.md` failed; the successful governing file is `profiles/manager.md`. These are setup/inspection failures, not additional product defects.

No source or test candidate changed, so no newly selected landing-property suite or full CI was required or claimed. Full repository suites, conformance and mutation sweeps were **not run**; six existing targeted tests above were rerun directly, not accepted from another operator's evidence. No missing/truncated log is credited as a passing command.

## Coverage map / mutant bounds

Brief gap: the brief supplies A–E scenarios but **no attack-surface-catalog surface table**. No catalog-complete coverage is claimed. The scenario table and full inventories cover every supplied scenario; research does not ship a new gate. No mutants were executed. The proposed narrowing controls below are leaf requirements, **not killed-mutant evidence**; therefore this task cannot claim a measured killed-mutant coverage ratio.

| Proposed mutant | Gate narrowed to | Named proposed failing test | Evidence bound |
| --- | --- | --- | --- |
| A: preserve only directories containing SKILL.md | Recognized skills, admits deletion of notes-only directories and loose files | TestInstallPreservesUnownedStaleStoreEntries | Not executed; required future leaf control |
| A: preserve directories but delete unknown regular files | Directory-only protection | TestInstallPreservesUnownedStaleStoreEntries | Not executed; loose-file row remains a required future control |
| C: refuse only when unowned destination contains SKILL.md | Skill-shaped content only, misses notes-only collisions | TestInstallRefusesUnownedSelectedCanonicalDestination (notes-only variant) | Not executed; required future leaf control |
| C: invalid/unreadable marker counts as owned | Any marker filename instead of valid readable ownership | TestInstallRefusesUnownedSelectedCanonicalDestination (invalid/read-failure variants) | Not executed; required future leaf control |
| B: admit absent-manifest sentinel with runtime declarations | Sentinel without context-only bound | TestSourceMarkerAbsentManifestIsContextOnly | Not executed; representation must first be specified |

Out of contract / unverified, with scope basis: product changes and actual fix implementation are excluded by task Scope, “read-only research ... no product change”; exhaustive alternate takeover/activation paths, other platforms, runtime execution, concurrent/racy mutation, unusual filesystem types, source transport and destructive recovery are outside AC1's five private-sandbox reproductions. C is a bounded path sweep, not a proof that no other path can delete data. Native homes tested are all four revision-1 roots on darwin/arm64; secondary targets and managed-session homes were not exercised. Full landing/reviewer acceptance is the orchestrator's lifecycle, not this research handoff. All prior reviewer tests are kept byte-identical because the tracked diff is empty.

No independent implementation or mutant acceptance is implied. Review should assess the spec-gap classification, the canonical ownership reading, and the proposed leaf boundaries before scheduling fixes.
