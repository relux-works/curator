## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- BUG-260731-2rhy74

## Checklist
- [ ] Reproduce and isolate each Windows failure family on ssh win or a faithful Windows harness.
- [ ] Fix command-shim digest, replacement, and provenance behavior without weakening corruption guards.
- [ ] Add focused Windows regression tests and keep Linux/macOS behavior green.
- [ ] Publish a signed commit to the CocoaSkills PR 16 branch only on ivanopcode/cocoaskills.
- [ ] Attach evidence, hand off to independent Opus review, and require the full PR 16 matrix green.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-2dec82, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-2dec82)
Root cause analysis (in progress).

Failure classification of run 30594273278 windows py3.11 (45 failed):
- 28 digest-mode: TransactionCorruptionError transaction target changed while digesting <...>.cmd
- 12 winerror5: PermissionError WinError 5 on project/.claude|.codex/skills/<name> (8 direct + 4 test_e2e CLI exit 1)
- 4 publication-owner: cache_publication_invalid manager home owner does not match the current manager principal
- 1 test_audit_cli exit 1, same digest-mode family (skill ships scripts/tool.cmd)

CONFIRMED CAUSE 1 (digest-mode). CPython Modules/posixmodule.c update_st_mode_from_path() ORs 0o111 into st_mode for .bat/.cmd/.com/.exe, and it is only reachable from path-based stat (win32_xstat_impl). Python/fileutils.c _Py_fstat_noraise has no path so it cannot apply it. transactions._digest_file compares stat.S_IMODE(os.fstat(fd)) against stat.S_IMODE(path.lstat()) and therefore always sees 0o666 vs 0o777 for a .cmd shim. Deterministic, fires for every command shim, never for .csk-install.json. Same defect in _entry_content_digest and _staging_prefix_digest. The digest payload also embedded the synthesized mode, so a staged sidecar name and its live .cmd name digest differently on Windows.

CONFIRMED CAUSE 2 (winerror5). adapters.stage_project_adapter_targets stages the adapter entry as a relative symlink whose destination is computed for the live location, so it is dangling while staged. transactions._staging_tree_entry recorded link_is_directory=path.is_dir(), which resolves the dangling destination and yields False. _create_staging_entry then rebuilt it with target_is_directory=False, producing a Windows file-type symlink onto a directory, which cannot be traversed: os.stat raises WinError 5. POSIX ignores the flag, which is why only Windows breaks.

CAUSE 3 (publication-owner) still under diagnosis on a windows-latest probe job. The affected tests reach cache publication only under PR16 because c4131bd added the private_base build operation root and stubbed go_v1.build.

Fixes applied so far on the PR16 branch content: permission-identity based mode handling in transactions digest/guards, link type read from the reparse point instead of the resolved destination, adapters symlink probe now creates a directory link. New regression tests in tests/test_transactions.py.
Fix published to PR 16 as signed commit 7a66c73 on task/TASK-260720-3t8nr3-transactional-project-hybrid (parent 8a02e17). Four independent Windows platform causes, none of them transaction-engine logic.

CAUSE 1 (28 failures) - synthesized execute bits. CPython update_st_mode_from_path ORs 0o111 into st_mode for .bat/.cmd/.com/.exe and is reachable only from path-based stat; os.fstat has no path. Probe on windows-latest: tool.cmd lstat=0o777 fstat=0o666 same inode; tool.txt 0o666/0o666. transactions._digest_file compared the two directly. Same defect in _entry_content_digest and _staging_prefix_digest, plus a latent one in the digest payload (a staged sidecar name and a live .cmd name digested differently). Fix: _permission_identity, reusing the existing _permission_mode_identity; no-op on POSIX.

CAUSE 2 (12 failures) - Windows symlink type. A staged adapter link is dangling by construction, and link_is_directory was derived by resolving that destination, so the link was rebuilt as a file link onto a directory, which Windows cannot traverse (WinError 5). Probe: dangling dir link has st_file_attributes=0x410 while is_dir() is False. Fix: read the type from the reparse point; also compared during staging validation, and the adapters probe now creates the link kind it will actually create. The 3.14 cell differing by exactly these 12 (assert False instead of WinError 5, because Path.exists() swallows it on 3.14) is independent confirmation.

CAUSE 3 (4 failures) - Windows never grants manager ownership. New objects belong to the token owner; on the runner _current_user_sid() is RID 500 while every created object is owned by S-1-5-32-544 BUILTIN\Administrators. The manager owned neither the home it created nor the artifact its compiler produced, so the protected build cache rejected both. POSIX gets that state free from st_uid plus mkdir(0o700). Not reachable on main, whose installer only plans builds; c4131bd is the first commit where install compiles and publishes. Fix on the producer side, guards untouched: locking.provision_new_manager_home provisions a home it creates (never repairs an existing one, so the drift guard and its test still hold), and builds.cache.make_publication_source_private stamps a fresh artifact before publication.

CAUSE 4 - _stub_trusted_toolchain hard-coded goos=linux and bin/<cmd>, so it could only ever publish on a non-Windows host. Now derives target and artifact path from the host.

EVIDENCE. Local macOS: mypy strict exit 0; full suite 1140 passed 100 skipped exit 0 (baseline before fixes 1136/99, delta is exactly the five new tests). windows-latest staged probes each removed one signature and exposed the next, no signature reappearing; final probe run 30620126160 = 253 passed, 55 skipped, success across test_install, test_status, test_gc, test_build_cache_windows, test_build_cache_posix, test_transactions, test_locking, test_adapters, test_config. Probe branches deleted. PR 16 full matrix run 30624304158 in flight on 7a66c73; mypy strict already green.

OPEN PRODUCT DECISION (not blocking this fix): a Windows home created by an elevated shell before this change stays Administrators-owned and fails closed. Adopting it would blunt the drift guard, so no automatic repair was added. Explicit opt-in repair vs documented manual re-provisioning is unresolved.
Host migration resumed on macbook-iv.local. CocoaSkills PR 16 head 7a66c73 currently has Ubuntu/macOS/mypy and Windows Python 3.11, 3.13, 3.14 green; Windows Python 3.12 remains in progress. Monitor that exact job, collect evidence, make focused rework only if it fails, then hand off to independent Opus review. Push only to ivanopcode/cocoaskills; no tags/releases.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-d6e3c5, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-d6e3c5)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-2dec82.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-2dec82.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_root-cause-and-fix.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_root-cause-and-fix.md) — Root cause analysis of all four Windows failure families and the fixes applied
- [BUG-260731-1rldqv_evidence.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_evidence.md) — Commands run, exit codes, local and windows-latest CI evidence
- [BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-d6e3c5.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-d6e3c5.log) — System spawn log captured by task-board

## Created
2026-07-31T01:59:51Z

## Last Update
2026-07-31T13:22:53Z

## Assigned To
[implementer] developer (claude)
