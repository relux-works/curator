# TASK-260730-2kfjm8 landing evidence

Date: 2026-07-30
Role: developer
Disposition: ready for independent review

## Candidate commit

- Branch: `release/curator-v0.13.0-candidate`
- Commit: `cfffd7cd7be33aba9ca44d26993ce5ab19b5fa4d`
- Parent/base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Tree: `2ce3f14440c5ae8104ef2d9c1fa73908a84553fc`
- Subject: `Add declarative compiled skill builds`
- Remote comparison: `origin/main...HEAD` is `0 1`
- Accepted manifest SHA-256: `c4c1ef8f0238c2cad18e2d3ab898889396035b8be1ed628d694d41bd1e724240`
- Canonical staged/commit path-list SHA-256: `9d47ac9d02f59ac4bb1c934d91d2d779ad4a92bbee1a32b87ee00a71f6fd5a89`

No push, tag, GitHub release, protocol-pin promotion, source-byte edit, or
heavy-suite rerun was performed.

## Verification ledger

Each gate below was run directly, and its real exit code is reported.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `git fetch --no-tags origin main` | 0 | Fresh fetch succeeded. |
| Exact base assertion | 0 | `HEAD`, `origin/main`, and required base all equal `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. |
| First live-manifest verifier | 1 | Verifier-model error: it treated the accepted gitlink `dir` row as unknown and enumerated intentionally excluded ignored/gitlink-inner files. No data mismatch was established. |
| Canonical accepted manifest generation/comparison | 0 | 374 generated entries equal the attached 374-entry accepted manifest byte-for-byte. |
| Pre-stage status set | 0 | Exactly 230 unique modified/new paths. |
| `git add -A` | 0 | Staged the accepted working-tree delta. |
| First staged path hash comparison | 1 | Order-sensitive verifier error: `git status` and `git diff --cached` emitted the same 230-path set in different orders. |
| Canonical staged path-set comparison | 0 | 230 status paths, 230 staged paths, zero unstaged paths, equal sets; canonical hash shown above. |
| `git diff --quiet` | 0 | No unstaged tracked delta after staging. |
| `git diff --cached --check` | 0 | No whitespace errors in the staged delta. |
| Index pin/candidate identity assertion | 0 | Exactly one `SPEC_PIN`, still rc.3 commit `00b1688a9b2457ca397a0bb550acf47cad8ee967`; one each of `candidate-only`, `release_claim none`, and `conformance_claim none`. |
| `git write-tree` | 0 | Prospective tree `2ce3f14440c5ae8104ef2d9c1fa73908a84553fc`. |
| First prospective-tree verifier | 1 | Verifier-model error: it assumed the accepted materialized gitlink directory carried nested `.git` metadata; it does not, so `git -C` resolved to the parent repository. |
| Corrected prospective-tree verifier | 0 | Gitlink mode plus its exact 34-inner-file digest and every other tree entry match all 374 accepted manifest entries. |
| `git commit -m 'Add declarative compiled skill builds'` | 0 | Created exact candidate commit `cfffd7cd7be33aba9ca44d26993ce5ab19b5fa4d`. |
| Post-commit identity assertion | 0 | Exact commit parent, tree, subject, branch, and one-commit-ahead relationship. |
| Post-commit `git status --porcelain=v1 -uall` | 0 | Empty output; accepted worktree is clean. |
| Commit-tree manifest verifier | 0 | Commit tree has 374 entries and matches all 374 accepted manifest entries exactly. |
| Commit pin/candidate identity assertion | 0 | Released pin is unchanged and candidate evidence remains candidate-only with no release or conformance claim. |
| Commit delta path assertion | 0 | Exactly 230 unique paths; canonical path-list hash matches the staged hash. |
| `git diff --check HEAD^ HEAD` | 0 | No whitespace errors in the committed delta. |
| `logbook --help` readiness check | 127 | Standalone `logbook` CLI is not installed; anomalies are preserved here and in board notes instead. |

The explicit landing brief forbids rerunning heavy suites. The accepted
composite's existing test, build, lint, and race evidence remains applicable
because the commit-tree verifier proves zero byte drift across the accepted
374-entry manifest.

## Handoff blocker

`task-board handoff TASK-260730-2kfjm8 --role developer` exited 1 because
checklist items 3, 4, and 10 remain unchecked.

- Item 3 requires an independent reviewer to accept the exact commit before
  main is pushed. That is downstream of this developer handoff.
- Item 4 requires the reviewed commit to be pushed to main and a GitHub release
  to be created. The attached developer brief forbids push, tag, and release,
  and the existing human authority note explicitly defers the tag/release.
- Item 10 requires findings to be recorded in `logbook`, but no standalone
  `logbook` command or repository logbook exists. The readiness check exited
  127. The findings are preserved in this artifact and board notes.

The handoff CLI requires every checklist item and offers no phase-aware
exception. Falsely checking future work or bypassing the handoff command with a
raw status mutation would violate the evidence and ownership contracts.

Recommended resolution: move items 3 and 4 to reviewer/publisher follow-on
work, or make the checklist phase-aware, and explicitly accept board
notes/outcome evidence for item 10 when no logbook integration is installed.
Then rerun the developer handoff without changing the candidate commit.

## Canonical staged path list (230)

```text
.github/ci/candidate-suite.sh
.github/ci/excluded-packages.sh
.github/ci/gate-selftest.sh
.github/ci/ledger-consistency.sh
.github/ci/no-broad-suppression.sh
.github/ci/platform-case-gate.sh
.github/ci/platform-cases.tsv
.github/ci/platform-exclusions.tsv
.github/ci/root-artifacts.tsv
.github/ci/skip-classes.tsv
.github/ci/suite-plan.sh
.github/ci/test-gate.sh
.github/ci/toolchain-identity.sh
.github/workflows/ci.yml
Makefile
README.md
cmd/curator/builds.go
cmd/curator/builds_test.go
cmd/curator/gc_test.go
cmd/curator/global_status_test.go
cmd/curator/lifecycle_conformance_test.go
cmd/curator/main.go
cmd/curator/main_test.go
cmd/curator/status_test.go
go.mod
internal/adapters/adapters.go
internal/adapters/adapters_test.go
internal/adapters/fifo_other_test.go
internal/adapters/fifo_unix_test.go
internal/adapters/stage.go
internal/buildcache/builddriver_conformance_other_test.go
internal/buildcache/builddriver_conformance_unix_test.go
internal/buildcache/builddriver_positive_conformance_test.go
internal/buildcache/builddriver_rejection_conformance_test.go
internal/buildcache/cache.go
internal/buildcache/cache_test.go
internal/buildcache/collect.go
internal/buildcache/collect_test.go
internal/buildcache/collect_unix_test.go
internal/buildcache/collect_windows_test.go
internal/buildcache/compensation_test.go
internal/buildcache/conformance_test.go
internal/buildcache/helpers_test.go
internal/buildcache/protection_unix.go
internal/buildcache/protection_unix_test.go
internal/buildcache/protection_unsupported.go
internal/buildcache/protection_unsupported_test.go
internal/buildcache/protection_windows.go
internal/buildcache/protection_windows_test.go
internal/buildcache/publish.go
internal/buildcache/validation_test.go
internal/buildmeta/builddriver_policy_conformance_test.go
internal/buildmeta/buildmeta_test.go
internal/buildmeta/codec.go
internal/buildmeta/models.go
internal/buildsource/builddriver_conformance_other_test.go
internal/buildsource/builddriver_conformance_test.go
internal/buildsource/builddriver_conformance_unix_test.go
internal/buildsource/buildsource.go
internal/buildsource/buildsource_special_unix_test.go
internal/buildsource/buildsource_test.go
internal/buildsource/conformance_test.go
internal/closure/closure.go
internal/closure/closure_test.go
internal/closure/conformance_test.go
internal/devsub/devsub.go
internal/envfiles/envfiles.go
internal/envfiles/stage.go
internal/gitignore/gitignore.go
internal/globalbins/globalbins.go
internal/globalbins/globalbins_test.go
internal/globalbins/stage.go
internal/godriver/boundary_test.go
internal/godriver/build.go
internal/godriver/build_conformance_test.go
internal/godriver/build_test.go
internal/godriver/builddriver_positive_conformance_test.go
internal/godriver/builddriver_rejection_conformance_test.go
internal/godriver/controls.go
internal/godriver/controls_darwin.go
internal/godriver/controls_other.go
internal/godriver/controls_test.go
internal/godriver/controls_windows.go
internal/godriver/errors.go
internal/godriver/executor.go
internal/godriver/executor_test.go
internal/godriver/fingerprint.go
internal/godriver/fingerprint_equivalence_test.go
internal/godriver/fingerprint_test.go
internal/godriver/fingerprint_unix_test.go
internal/godriver/graph.go
internal/godriver/graph_test.go
internal/godriver/guards_test.go
internal/godriver/identity.go
internal/godriver/main_test.go
internal/godriver/platform_unix.go
internal/godriver/platform_windows.go
internal/godriver/process_alive_unix_test.go
internal/godriver/process_alive_windows_test.go
internal/godriver/session.go
internal/godriver/session_test.go
internal/godriver/testdata/realbuild/build/cmd/golden-tool/main.go
internal/godriver/testdata/realbuild/build/cmd/golden-tool/message.txt
internal/godriver/testdata/realbuild/build/go.mod
internal/godriver/testdata/realbuild/build/vendor/modules.txt
internal/godriver/testdata/stubgo/main.go
internal/godriver/testhelper_unix_test.go
internal/godriver/testhelper_windows_test.go
internal/godriver/worker_test.go
internal/godriver/workerclient.go
internal/godriver/workerproto.go
internal/godriver/workerserver.go
internal/install/aba_test.go
internal/install/atomicity/activation_test.go
internal/install/atomicity/commit_atomicity_test.go
internal/install/atomicity/doc.go
internal/install/atomicity/fixture_test.go
internal/install/builddeps.go
internal/install/cache_conformance_test.go
internal/install/commit.go
internal/install/commit_test.go
internal/install/diagnostics.go
internal/install/diagnostics_test.go
internal/install/dryrun_conformance_test.go
internal/install/generation.go
internal/install/generation_test.go
internal/install/global.go
internal/install/install.go
internal/install/install_test.go
internal/install/maintenance_test.go
internal/install/plan.go
internal/install/private.go
internal/install/private_test.go
internal/install/registry_e2e_test.go
internal/install/revalidation_test.go
internal/install/stage.go
internal/install/stage_test.go
internal/install/targets.go
internal/interop/golden_test.go
internal/managerlock/filelock.go
internal/managerlock/filelock_unix.go
internal/managerlock/filelock_windows.go
internal/managerlock/identity.go
internal/managerlock/identity_unix.go
internal/managerlock/identity_windows.go
internal/managerlock/identity_windows_test.go
internal/managerlock/managerlock.go
internal/managerlock/managerlock_test.go
internal/manifest/manifest.go
internal/marker/marker.go
internal/marker/marker_test.go
internal/marker/marker_v2_test.go
internal/protocoljson/ccj.go
internal/protocoljson/ccj_test.go
internal/protocoljson/json_test.go
internal/registry/registry.go
internal/runtimestore/conformance_test.go
internal/runtimestore/launcher_conformance_test.go
internal/runtimestore/runtimestore.go
internal/runtimestore/scripts.go
internal/runtimestore/targets.go
internal/runtimestore/targets_test.go
internal/runtimestore/targets_unix_test.go
internal/runtimestore/targets_windows_test.go
internal/scopes/consumers.go
internal/scopes/gc.go
internal/scopes/gc_conformance_test.go
internal/scopes/gc_conservative_test.go
internal/scopes/gc_conservative_unix_test.go
internal/scopes/gc_conservative_windows_test.go
internal/scopes/gc_integration_other_test.go
internal/scopes/gc_integration_test.go
internal/scopes/gc_integration_windows_test.go
internal/scopes/gc_test.go
internal/scopes/hybrid.go
internal/scopes/redirect_other.go
internal/scopes/redirect_windows.go
internal/scopes/stage.go
internal/skillcheck/builddriver_context_conformance_test.go
internal/skillcheck/skillcheck.go
internal/skillcheck/skillcheck_test.go
internal/skillspec/build_test.go
internal/skillspec/builddriver_conformance_other_test.go
internal/skillspec/builddriver_conformance_test.go
internal/skillspec/builddriver_conformance_unix_test.go
internal/skillspec/conformance_test.go
internal/skillspec/parse.go
internal/skillspec/parse_test.go
internal/skillspec/types.go
internal/snapshot/lock.go
internal/snapshot/lock_unix.go
internal/snapshot/lock_windows.go
internal/snapshot/snapshot.go
internal/snapshot/snapshot_test.go
internal/staging/staging.go
internal/staging/staging_test.go
internal/transaction/digest.go
internal/transaction/durability_unix.go
internal/transaction/durability_windows.go
internal/transaction/engine.go
internal/transaction/engine_test.go
internal/transaction/entry_test.go
internal/transaction/files.go
internal/transaction/journal.go
internal/transaction/journal_order_test.go
internal/transaction/namespace.go
internal/transaction/namespace_case_darwin.go
internal/transaction/namespace_case_other.go
internal/transaction/namespace_case_windows.go
internal/transaction/namespace_pass_test.go
internal/transaction/preparation_durability_test.go
internal/transaction/recovery_corruption_test.go
internal/transaction/rename_noreplace_darwin.go
internal/transaction/rename_noreplace_linux.go
internal/transaction/rename_noreplace_other_unix.go
internal/transaction/rename_noreplace_windows.go
internal/transaction/replace_unix.go
internal/transaction/replace_windows.go
internal/transaction/root_metadata_unix_test.go
internal/transaction/staging.go
internal/transaction/subprocess_test.go
internal/transaction/types.go
internal/transaction/validation_darwin_test.go
internal/transaction/validation_test.go
internal/transaction/validation_windows_test.go
internal/whitelist/builddriver_context_conformance_test.go
internal/whitelist/conformance_test.go
internal/whitelist/whitelist.go
internal/whitelist/whitelist_test.go
task-board.config.json
```
