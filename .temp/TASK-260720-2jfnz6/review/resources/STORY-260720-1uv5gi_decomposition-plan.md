# csk Go build-driver decomposition

**Board story:** `STORY-260720-1uv5gi` — csk Go build driver  
**Target repository:** `/Users/iv/Developer/Wildberries/cocoaskills`  
**Verified remote baseline:** `origin/main` = `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` locally and through `git ls-remote`  
**Local-clone finding:** clean local `main` at `edce8816dda44bb121d661b7c4dea942558ce408`, two commits behind `origin/main`  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`

## Development plan

| Phase | Task | Atomic deliverable | Primary ownership |
|---:|---|---|---|
| 1 | `TASK-260720-z9j4c9` — Implement schema v6 build manifest model | Strict Python schema-6 domain and static validation | `skillspec.py`, validation-focused `skillcheck.py`, parser tests |
| 1 | `TASK-260720-z2z795` — Implement install transaction engine | Reusable locks, durable journals, recovery, and reverse rollback | `locking.py`, new `transactions.py`, concurrency tests |
| 2 | `TASK-260720-3c0ss2` — Implement build-source identity and context isolation | Raw-snapshot identity plus prompt/runtime exclusion | new build-source module, additive hashing, `whitelist.py` |
| 2 | `TASK-260720-3j8pp5` — Implement trusted Go toolchain identity | Private telemetry-off toolchain session and framed fingerprint | new `builds/toolchain.py` and tests |
| 3 | `TASK-260720-2dnqw2` — Implement canonical build metadata models | CCJ-1 input/key, receipt v1, marker v2 | build metadata modules, `install_marker.py`, protocol JSON helper |
| 3 | `TASK-260720-2g21eg` — Implement fixed go-v1 compile driver | Fixed dependency preflight and compiler pipeline | new `builds/go_v1.py`, executor boundary, fixture tests |
| 4 | `TASK-260720-2jfnz6` — Implement protected POSIX build cache | Backend interface and owner-protected POSIX store | `builds/cache.py`, `builds/cache_posix.py` |
| 5 | `TASK-260720-8nxlgx` — Implement protected Windows build cache | DACL- and handle-safe Windows store | `builds/cache_windows.py`, Windows tests |
| 6 | `TASK-260720-11yhth` — Harden runtime reuse and activate built commands | Required-path runtime reuse plus Unix/Windows shims | `shims.py`, narrow global-bin helpers, platform tests |
| 6 | `TASK-260720-2x6mjn` — Implement pure compiled-build planning | Audit-first project/global plans and pure dry-run | `builds/planner.py`, planning paths, registry read-only API, CLI routing |
| 7 | `TASK-260720-3t8nr3` — Integrate atomic project and hybrid builds | Build-before-mutation, cache publication, marker/shim/context commit, consumer last | project/hybrid `installer.py` materialization |
| 8 | `TASK-260720-g7kgox` — Integrate atomic global builds | Global scope parity through the shared transaction engine | `global_install.py`, global bins/adapters/env wiring |
| 9 | `TASK-260720-th0jdi` — Implement build currentness, repair, and GC | Status, reinstall repair, and locked cache lifecycle | `status.py`, `gc.py`, global status and CLI tests |
| 10 | `TASK-260720-12r55p` — Consume shared schema v6 vectors | Independent Python rc.4 vector consumer | `test_protocol_conformance.py`, vector adapters |
| 10 | `TASK-260720-akf5kh` — Document schema v6 compiled commands | English source docs and maintained Russian mirrors | README, architecture, security, authoring guide, changelog |
| 11 | `TASK-260720-3pemm6` — Add cross-platform Go build end-to-end tests | Real Go build, launch, cache, rollback, concurrency, and CI evidence | test fixtures, E2E/process tests, CI workflow |
| 12 | `TASK-260720-3s27te` — Verify integrated csk schema v6 implementation | Full acceptance and quality-gate evidence | logs, SHA record, AC/vector coverage matrix |

The canonical board plan has 17 tasks in 12 phases. Parallel tasks own separate modules or documentation surfaces. Shared edits to installer and global-install code are serialized behind the planner and transaction foundations.

Critical path:

```text
TASK-260720-z9j4c9
  -> TASK-260720-3c0ss2
  -> TASK-260720-2dnqw2
  -> TASK-260720-2jfnz6
  -> TASK-260720-8nxlgx
  -> TASK-260720-11yhth
  -> TASK-260720-3t8nr3
  -> TASK-260720-g7kgox
  -> TASK-260720-th0jdi
  -> TASK-260720-12r55p
  -> TASK-260720-3pemm6
  -> TASK-260720-3s27te
```

## Architecture decisions

1. The portable contract ends at logical cache key, canonical receipt bytes, artifact-relative path, validation outcomes, and marker identity. Physical csk storage remains implementation-specific.
2. The plan reserves a dedicated top-level `<csk-home>/builds/go-v1/<logical-key>/...` namespace. Reusing `<csk-home>/cache/...` would collide with the existing source-snapshot namespace, where the first path segment is package-controlled source identity.
3. POSIX and Windows cache protection are separate tasks behind one backend interface. A platform that cannot prove protected ownership, ACL or mode, containment, and link safety disables reuse or fails closed; it never admits candidate bytes optimistically.
4. `csk install` and `csk global install` are the repair operations. No package-controlled repair hook or generic build-policy surface is added.
5. Existing script runtimes remain commit-keyed, but reuse now validates all required active paths. Compiled outputs never use the script runtime identity.
6. Independent project builds may run outside the home lock. Publication, every mutable project/global/hybrid/adapter/consumer target, recovery, rollback, and GC serialize under one manager-home mutation lock.
7. The machine-wide consumer ledger commits last. A journaled rollback never restores a stale preimage over another project success.
8. Dry-run may use operation-private package-independent Go probes and read existing state. It acquires no mutation lock, executes no source-aware Go command, and leaves no persistent manager or project mutation.
9. CocoaSkills remains an independent consumer. Tests may share protocol vectors and fixtures but do not import or copy Curator implementation code.

## Completeness matrix

| Story or protocol requirement | Owning tasks |
|---|---|
| Schema 6 model, `build_roots`, closed `go-v1` declarations, legacy compatibility | `TASK-260720-z9j4c9`, `TASK-260720-12r55p` |
| Skill check and fail-closed static path/module validation | `TASK-260720-z9j4c9`, `TASK-260720-3c0ss2` |
| Build sources excluded from agent context and runtime | `TASK-260720-3c0ss2`, `TASK-260720-3t8nr3` |
| Exact build-source, toolchain, CCJ-1, cache-key, receipt, and marker algorithms | `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5`, `TASK-260720-2dnqw2` |
| Trusted Go 1.23+ session, private telemetry, native target, fixed environment | `TASK-260720-3j8pp5`, `TASK-260720-2g21eg` |
| Vendor-only graph validation, cgo/native/directive rejection, fixed build argv | `TASK-260720-2g21eg`, `TASK-260720-12r55p` |
| csk manager-home cache identity and protected POSIX/Windows storage | `TASK-260720-2jfnz6`, `TASK-260720-8nxlgx` |
| Existing runtime required-path verification | `TASK-260720-11yhth` |
| Unix and Windows project/global activation, argv and exit propagation | `TASK-260720-11yhth`, `TASK-260720-3pemm6` |
| Audit and trust gates before build, provider/lexical ordering, pure dry-run | `TASK-260720-2x6mjn`, `TASK-260720-3t8nr3` |
| Project, hybrid, global transaction, cache publication, consumer-last rollback | `TASK-260720-z2z795`, `TASK-260720-3t8nr3`, `TASK-260720-g7kgox` |
| Marker v1/v2, currentness, repair, status JSON/check, protected GC | `TASK-260720-2dnqw2`, `TASK-260720-th0jdi` |
| Shared schema/build/lifecycle vectors | `TASK-260720-12r55p` |
| Real Go build and launch on Linux, macOS, and Windows | `TASK-260720-3pemm6` |
| Authoring, architecture, security, lifecycle, cache, and activation docs | `TASK-260720-akf5kh` |
| Existing pytest, strict typing, packaging, and any declared lint gates | Every implementation task locally; integrated by `TASK-260720-3s27te` |

## Gap and risk audit

- No unresolved product or architecture choice remains in this story. The accepted research fixes schema 6, `go-v1`, build roots, all byte algorithms, protected-state provenance, dry-run, manager-home isolation, marker v2, receipt v1, and the claim transition.
- Executable upstream gates remain explicit at task granularity: shared-vector consumption is blocked by `TASK-260720-3ag6pi` — Verify integrated protocol v6 conformance, and downstream documentation waits for `TASK-260720-3lo9jc` — Document schema v6 authoring and CLI behavior. Coarse story links to the protocol and Curator implementations are intentionally omitted at handoff: the former duplicates these exact gates, while the latter would couple an explicitly independent Python implementation to reference-manager code. `STORY-260720-21bsr2` remains blocked by both manager stories and owns cross-manager sequencing. No duplicate clarification task is needed.
- The local cocoaskills clone is clean but behind `origin/main` by two commits. Every implementation task checklist requires fast-forwarding the clean clone before its task worktree and then basing on all dependency handoffs.
- Windows DACL and reparse-safe access are implementation-heavy but not a human decision boundary. The platform task has a fail-closed backend contract and dedicated Windows CI acceptance criteria.
- The protocol rc.4 suite is not landed yet. The conformance and CI tasks explicitly forbid pinning an unreleased or fabricated commit and wait for the immutable upstream verification task.
- The current cocoaskills CI has pytest, strict mypy, build, and twine gates but no separately configured lint command. Final verification runs any lint gate that exists by execution time and records the current absence instead of inventing release evidence.
- Existing implementation anomalies are tracked in the owning tasks: consumer state is written before materialization, runtime reuse trusts an existing directory without required-path validation, materialization is not transaction-wide, and project/global dry-run currently enters a file-creating global lock while the registry path can persist cache/state.
- Existing unrelated board validation anomalies remain out of scope. Planning validation reports them separately and does not alter unrelated board items.

## Linked diagrams

- `TASK-260720-z9j4c9_csk-build-components.puml` and `.svg` — module ownership and data/control dependencies, attached to `TASK-260720-z9j4c9`.
- `TASK-260720-z2z795_csk-build-lifecycle.puml` and `.svg` — audit, dry-run, build, commit, rollback, and GC ordering, attached to `TASK-260720-z2z795`.

Both sources pass PlantUML 1.2026.6 `-checkonly` and render with Smetana. The system Graphviz binary is unusable because its Homebrew `libltdl.7.dylib` dependency is missing, so rendering deliberately does not depend on Graphviz.

## Integration gates

The final verification task must attach evidence for the repository-supported equivalents of:

```bash
CURATOR_CONFORMANCE_ROOT=/absolute/path/to/curator-spec/conformance/v1 python -m pytest -v
python -m mypy
python -m build
python -m twine check dist/*
git diff --check
```

It must also attach the Linux, macOS, and Windows CI matrix result for the real Go fixture and map every story criterion plus every required negative vector cluster to passing evidence.
