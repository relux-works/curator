# TASK-260720-th0jdi — Build currentness, repair, and GC evidence

## Baseline and scope

- Base SHA: `07655553cebcf867bbe58629de98e77644606c85`
- Branch: `task/TASK-260720-th0jdi-build-currentness-repair-gc`
- Worktree: `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-th0jdi/worktree`
- Owned implementation: project/global status, marker-v2 build currentness, read-only managed-shim validation, normal-install repair coverage, locked protected-build GC, live-journal roots, and focused CLI/maintenance tests.

## Key outcomes

1. Project and global status report current only when the selected ref, installed content, closure/activation fields, static build-root exclusion, persistent raw source identity, current toolchain and native target, complete fixed-policy cache key, marker-v2 build fields, protected canonical receipt, artifact path/hash/size, and exact managed shim all agree. Text, JSON, and `--check` share the same verdict.
2. The complete cache-key and receipt comparison is the execution-policy mechanism. The implementation explicitly reports `policy.execution_policy=manager-worker-v1` as part of that boundary; it does not add a parallel policy flag. The exact legacy key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` and reserved-hardened key `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` are non-current.
3. Capability evidence appears once in status results and is result-only. Evidence availability or failure never changes `ProjectStatus.clean`, cache identity, receipts, or markers.
4. Repair remains the normal install path. A regression test corrupts a protected artifact, observes non-current status, and proves reinstall recompiles from the revalidated snapshot without adopting the planted bytes.
5. Marker-v1 currentness remains supported for skill schemas 1–5, including after snapshot GC. Read-only status may use an operation-private Git archive for legacy resolution, but marker-v2 build status requires the selected persistent raw snapshot and never recreates it.
6. GC acquires or reuses the manager-home lock, marks valid marker-v1/v2 roots across project, global, hybrid, registered consumers, and active journals, and preserves runtime/snapshot behavior. An incomplete mark retains all runtime, snapshot, and build entries. Native POSIX/Windows collectors sweep only unreferenced, older-than-grace, protected entries whose receipt, complete key, artifact path, hash, and size validate; uncertain state is retained with warnings.

## Fact-checking and sources

Claims above were checked against Curator Spec commit `f5d7673039226ab81de2f4f87e2155ae995c4df3`:

- [Status and GC requirements](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L657-L684) define raw-snapshot/current-build validation, non-current cases, lock/mark/sweep behavior, journals, grace, and fail-safe uncertainty.
- [Cache acceptance and rebuild rules](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L329-L340) require full identity/receipt/artifact recomputation, reject absent or mismatched execution policy, exclude capability evidence, and rebuild untrusted state.
- [Recovery and repair](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L444-L463) specifies that repair is the install sequence and forbids permission repair, marker recalculation, or adoption of candidate bytes.
- [Capability evidence](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L251-L265) makes the record result-only and excludes it from keys, receipts, markers, and conformance claims.
- [Build identity and marker compatibility decision](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/decisions/0004-compile-only-build-drivers.md#L53-L75) separates installed-content and raw-build identities, binds toolchain/target/policy to the logical key, requires protected provenance, and preserves valid marker v1 for schemas 1–5.
- [Legacy policy-less key vector](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/conformance/v1/vectors/build-drivers.json#L308) and [reserved hardened-policy key vector](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/conformance/v1/vectors/build-drivers.json#L393) verify the exact regression identities.

The accepted Curator protected-cache collector was also inspected locally as an implementation cross-check. It decodes and validates the receipt before sweeping; therefore unsupported legacy/reserved-policy receipts remain uncertain and are retained, matching the fail-safe requirement rather than being force-deleted.

## Validation evidence

All reported gates were run as standalone processes without output pipes.

- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m pytest -q` — exit `0`; `1197 passed, 100 skipped in 262.52s`.
- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m pytest tests/test_build_currentness.py tests/test_build_gc.py tests/test_status.py tests/test_gc.py tests/test_installer_transactions.py tests/test_global_install_transactions.py -q` — exit `0`; `61 passed in 83.80s`.
- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m mypy` — exit `0`; strict mode found no issues in 68 source files.
- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m compileall -q src tests` — exit `0`.
- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m build --outdir /tmp/csk-TASK-260720-th0jdi-final-build.UgyLdr` — exit `0`; sdist and wheel built.
- `/tmp/csk-TASK-260720-th0jdi-venv/bin/python -m twine check /tmp/csk-TASK-260720-th0jdi-final-build.UgyLdr/*` — exit `0`; both artifacts passed.
- `git diff --check` — exit `0`.

Earlier red-to-green evidence is retained honestly:

- The first broad pytest run exited `1` with `1193 passed, 100 skipped, 1 failed`; the sole failure was an existing post-install-GC mock whose old two-argument signature did not accept the new lock witness. The test now accepts and asserts the same held witness, and its focused rerun plus the later full run exited `0`.
- The first run after adding explicit receipt/marker/global-status vectors exited `1` with `21 passed, 2 failed`; both failures were fixture expectations (the marker decoder rejects a non-`go-v1` driver before build classification, and the protected receipt has the canonical filename `csk-receipt.ccj.json`). Corrected fixtures then passed `23/23`, and the expanded focused and full gates above exited `0`.
- An early changed-module mypy pass exited `1` on a `BuildStatus.to_json` inferred-dictionary type; the return construction was corrected, and both subsequent strict full passes exited `0`.
- A bare `python -m mypy` shell attempt exited `127` because this environment intentionally has no `python` on `PATH`; it was not treated as a passing gate. The explicit task-venv command above is the actual green gate.

No dedicated formatter/linter is declared in `pyproject.toml`; the repository's declared strict static gate is mypy. Static typing, bytecode compilation, and diff-whitespace validation all exited `0`.

## Review handoff

- Signed commit: `f3c1254e4fe7958c720cbef096a4ef00103d43a2` (`Good "git" signature`, ECDSA key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`).
- Pull request: [ivanopcode/cocoaskills#18](https://github.com/ivanopcode/cocoaskills/pull/18).
- GitHub Actions: [run 30676739989](https://github.com/ivanopcode/cocoaskills/actions/runs/30676739989) — watcher exit `0`; all 12 Python 3.11–3.14 tests across Ubuntu, macOS, and Windows, strict mypy, and the dependent artifact build passed.
- Branch was pushed without tags or releases; the PR remains unmerged for review.
