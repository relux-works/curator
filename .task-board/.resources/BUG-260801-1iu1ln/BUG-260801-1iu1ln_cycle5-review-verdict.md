# BUG-260801-1iu1ln cycle-5 reviewer verdict

## Verdict

**Changes requested; route to `to-dev`.** This is ordinary implementation rework, not an external blocker.

The signed cycle-5 commit closes the seven cycle-4 regressions and preserves the earlier ten guards, but it still does not meet the core acceptance criterion that all 32 lifecycle cases and every normative field be bound to the same observed CocoaSkills operation/state. Five independently established causal/proxy gaps remain. Each can let a product-seam violation produce the complete normative case unchanged.

The reviewer did not modify CocoaSkills source or tests.

## Reviewed artifact and provenance

- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`
- Branch: `task/BUG-260801-1iu1ln-lifecycle-observed-traces`
- Required signed base and merge base: `ba250bfc4dfe104a160eadd5b5f4e340693bf892`
- Reviewed signed tip: `27ace8b6d1b7d6a1b54acac30e670098ca3b5110`
- Tip parent: signed `963d224c9bb3d6fc274b9dbfbac4bdcafd243c2b`
- `git verify-commit HEAD`: good ECDSA signature for `oparin@me.com`
- Worktree clean; no tag points at the tip.
- Exact conformance root: `/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`
- Exact root `manifest.json` SHA-256: `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`
- The exact-base diff changes only `LOGBOOK.md`, five CocoaSkills source modules, and five focused test/harness modules. It changes no pin, schema-v7, tag, release, claim, CI, changelog, or pyproject surface.

## Positive validation evidence

Reviewer reruns:

- Seven cycle-5 regressions for the cycle-4 findings: **7 passed in 299.39s**.
- All ten inherited adversarial regressions: **10 passed in 421.61s**.
- Exact-root canonical 32 cases, exhaustive scalar-leaf mutation audit, fail-closed literal/proxy classification, unknown-field rejection, process argv, exact project identity, and rollback sensitivity: **417 passed in 43.51s**.
- Strict mypy: **Success: no issues found in 68 source files**.
- `python -m compileall -q src tests`: exit 0.
- `git diff --check ba250bf..HEAD`: exit 0.
- Signed-base, clean-tree, merge-base, no-tag, and restricted-diff checks pass.

The producer outcome also records the full exact-root suite at 852 passed and broader package/related-suite gates. These green gates establish regression and expectation-side coverage, but the surviving probes below show that they do not establish product-seam causality.

## Acceptance-blocking findings

### 1. The cross-project “success” case does not observe two successful installs through publication and commit

In `tests/protocol_lifecycle_observations.py:1110`, each real concurrent install is deliberately stopped by `InstallError("private-build observation stop")` before publication. Lines 1175-1187 define the private probe as successful only when both CocoaSkills results are actually `failed`, the consumer registry remains empty, and the publish handoff was never reached. A separate synthetic consumer transaction helper at line 1201 writes the alpha/beta ledger, and line 1214 then projects the combined result as protocol `"success"`.

Consequently, the complete normative case can remain equal if the real install handoff/publication/commit integration is broken: that integration is explicitly required not to run. `shared_transactions_serialized` and the consumer ledger come from the separate synthetic transaction, while private-build overlap comes from two intentionally failed installs. This is an aggregate proxy, not the corresponding observed cross-project success transition.

The same case also returns the normative fixture key at lines 1218-1219 whenever the two actual plans are merely equal to each other and self-derived; it never proves that their actual key equals the normative key.

Required rework: drive two actual concurrent CocoaSkills installs to successful publication/transaction completion in one scenario; observe both consumer records and the shared cache entry from that scenario; prove overlap before the shared publication boundary and serialization at the actual shared transaction boundary; compare the actual `BuildInput`/key with the normative field. Add a sabotage proving that bypassing or failing the real publish/commit handoff changes the complete case.

### 2. Several normative cache/receipt identities are copied from the authenticated fixture, not derived from the case operation

Direct fixture projections remain at:

- cross-project success `shared_cache_key`: lines 1218-1219;
- cross-project rollback `shared_cache_key`: line 1252, although that synthetic transaction has no cache observation;
- dry-run `logical_cache_key`: line 1598;
- private-build `cache_key` and `receipt_sha256`: lines 2595 and 2597;
- status `cache_key` and `receipt_sha256`: lines 3650 and 3653;
- repair `cache_key`: line 4096;
- deterministic transaction-order `cache_key`: line 4518, although that transaction scenario has no cache seam.

Independent instrumentation recorded the actual plan/cache identities produced by those scenarios:

- normative fixture key: `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
- cross-project shared-plan key: `sha256:5795b6970d07eef737803bbbca15d328808ef3a3e590b7cf1b6c78c14dc49259`;
- private-build golden-tool plan key: `sha256:fd1bff17d590a1a5cf71ce572f6d5bc5bcf0c0c8c97f7f733dc7ec20d7a081c3`;
- dry-run and status plan key: `sha256:0bacaaef0284709007a76815faec96a85e4e2459d33d75ab2cc67f3fdf665d30`;
- actual status hit receipt: `sha256:8d589d0cd98137424aaaf8da23d0f5bd08a5392a79b0134f274adba3b6196b73`;
- returned fixture receipt: `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`.

Repair similarly produced distinct actual keys/receipts for its independent conditions while returning the same fixture key. The complete cases still equal the normative vector because the output comes from the separate authenticated identity fixture.

This is exactly the distinction already acknowledged by the LOGBOOK: the leaf mutation audit proves expectation-side equality sensitivity, not product-seam sensitivity.

Required rework: make each scenario use and observe the normative build input/cache/receipt, or explicitly compare its actual identity to the normative identity and fail closed. A normative field with no corresponding seam in that scenario must be bound through the scenario rather than copied from `identities`. Add regression probes that perturb the operation-side key/receipt while leaving the authenticated fixture untouched and require the complete case to differ.

### 3. Relative artifact execution under `cwd` is invisible

`_record_process_paths` at lines 59-89 intentionally records only absolute argv paths. Status, repair, GC, dry-run, and private-build subprocess wrappers pass the argv and protected roots but do not pass or resolve the subprocess `cwd`.

Independent repair sabotage executed the untrusted candidate as argv `./bin/tool` with `cwd=<candidate_entry>` at the observed repair seam. The process returned 0. The complete `repair-rebuilds-invalid-compiled-entry` object remained byte-for-byte equal to the normative case, including `result="rebuilt-and-journaled"` and all four `forbidden_shortcuts`.

Observed result:

```text
relative_argv=./bin/tool
cwd_is_candidate=true
returncode=0
complete_case_equal_after_execution=true
```

Required rework: resolve every path-like argv element against the effective subprocess `cwd` (and normal inherited cwd), cover each process-launch form CocoaSkills uses on supported platforms, and add at least one cwd-relative protected-artifact execution regression for repair plus the adjacent lifecycle surfaces that report `artifact_executed`.

### 4. Transient low-level persistent writes evade the mutation observer

`_install_persistent_mutation_observer` at lines 121-209 wraps selected `Path` methods plus only `os.chmod` and `os.fchmod`. It does not observe low-level create/write/delete/rename operations such as `os.open`, `os.write`, `os.unlink`, `os.rename`, and `os.replace`. Final `_tree_state` comparisons cannot see a mutation that is restored before return.

Two independent survivors were reproduced:

1. Currentness: a wrapper around `status._collect_resolved_scope` created, wrote, fsynced, and unlinked a file below manager home using `os.open/os.write/os.unlink` before calling the real function. This happened 15 times (clean case plus the 14 matrix rows). Both complete cases remained equal and returned `mutations=[]`.
2. Dry-run: after the real `planner.plan_builds` returned, a wrapper created/wrote/unlinked `home/builds/transient-dry-run-write` with low-level calls and restored parent timestamps. The complete `compiled-cache-miss-is-read-only` case remained exactly equal; the observed differences object was empty.

Observed result:

```text
currentness_low_level_mutations=15
current_case_equal_after_mutations=true
matrix_case_equal_after_mutations=true
current_mutations=[]
matrix_mutations=[]
dry_run_low_level_mutations=1
dry_run_case_equal_after_mutation=true
dry_run_differences={}
```

Required rework: replace the narrow API list with a fail-closed mutation witness for protected roots, or comprehensively intercept the filesystem mutation boundaries used by CocoaSkills, including descriptor-relative low-level operations. Regress low-level transient create/write/unlink on dry-run and both currentness paths, then audit adjacent GC, private-failure, repair, and rollback projections that rely on the same observer/final-state pattern.

### 5. Atomic-publication evidence is tied to one helper and misses a transient live destination created by another namespace operation

The POSIX observer at lines 435-505 records `os.mkdir` of the exact destination and wraps `cache_posix._rename_noreplace`. It does not observe a live destination exposed through `os.rename` (or equivalent alternate namespace operations) inside that boundary.

Independent sabotage inserted a partial directory at the exact live cache hash by `os.rename(..., src_dir_fd=..., dst_dir_fd=...)`, exposed it, removed it, and then called the real no-replace helper. The observer recorded 38 transient live exposures, yet the complete publication case stayed equal and projected `publication="atomic-complete-directory"`.

Observed result:

```text
transient_live_exposures_via_os_rename=38
complete_case_equal_after_exposure=true
projected_publication=atomic-complete-directory
```

Required rework: make the atomic-publication witness independent of one named CocoaSkills helper—observe all namespace transitions targeting the exact live destination or use a synchronization-based reader/witness that detects any partial exposure. Add the `os.rename` survivor as a POSIX regression and an equivalent Windows move/rename parity regression.

## Required completion package for the next reviewer

1. Fix all five causal gaps together; do not weaken the 417 checked canonical/scalar/classification surface or the existing seventeen sabotage guards.
2. Correct the LOGBOOK assertions that publication “rejects any partial destination,” the cross-project case observes success/shared normative identity, and transient persistent writes/process executions are exhaustively detected.
3. Add operation-side sabotage tests for exact identity mismatch, real cross-project handoff failure, cwd-relative execution, low-level transient writes, and alternate live-destination exposure.
4. Rerun the authenticated exact-root suite, focused lifecycle/adversarial tests, strict mypy, compileall/build validation, exact-base diff/release guards, signature/merge-base/clean-tree/no-tag checks.
5. Produce a new signed commit atop the reviewed branch and attach a new task-scoped rework artifact.

No human/product/architecture decision or external input is needed.