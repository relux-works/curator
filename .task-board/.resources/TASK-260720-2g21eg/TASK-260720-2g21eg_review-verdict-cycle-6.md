# TASK-260720-2g21eg review verdict — cycle 6

## Verdict

Accepted.

Cycle 6 was a narrow post-acceptance portability correction, not new feature
work: GitHub PR #12 strict mypy on ubuntu failed with 13 `attr-defined` errors
for Darwin-only kqueue symbols, and the focused suite could not run on a
non-macOS/Windows CI host. Both are fixed, the fix is provably runtime-neutral,
and the fail-closed guarantee it touches is still enforced by a live test. This
reviewer also closed the standing native-Windows caveat by re-running the suite
natively on Windows against byte-identical source.

## Candidate and provenance

- Worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch `task/TASK-260720-2g21eg-go-v1-compile-driver`, HEAD
  `673a38dc2fac499cbcbfa3ff6e9be84d9bae3ee8`
- Cycle-6 delta is uncommitted and is exactly two files, as claimed:
  `src/csk/builds/go_v1.py` (33 lines) and `tests/test_builds_go_v1.py`
  (124 lines). `git status --porcelain` shows nothing else.
- `src/csk/builds/go_v1.py`
  `sha256:f8433dc1e8eaecca5c4f04e574720e83f5fbf6a403a474576711297ff2cc0203`
- `tests/test_builds_go_v1.py`
  `sha256:5fad24308d5e3ec0640e91e4ea553b1cbbeef101097bb1b173d7a7afc828f254`
- Both match the developer handoff exactly.
- No product or test file was modified during review. All reviewer scratch work
  was done in throwaway copies under `.temp/`.

### Conformance root discrepancy — resolved, benign

The handoff pins rc.5 conformance commit
`f5d7673039226ab81de2f4f87e2155ae995c4df3`; the accepted conformance worktree is
at `0aae5dff11ab90400fc6a0b003a4492767b38043` and `f5d7673` is not an ancestor
of it. This is not a staleness problem: the `conformance/v1` trees of the two
commits are the identical git tree object `0ea6b7166482cfe951fdf62d72dbcbe3b5d8b8e4`,
so every vector byte the developer validated against is the byte the accepted
root carries.

## Product change — verified runtime-neutral, minimal, and complete

The whole `go_v1.py` delta is inside `_IdentityMutationGuard._setup_macos`:
Darwin-only `select`/`os` symbols are reached through `cast(Any, select)` /
`cast(Any, os)` instead of directly. `cast` is a no-op at runtime and the
attribute names, order, and values are unchanged, so the native macOS control
behavior is identical.

Independently confirmed the fix is causally correct and exactly scoped: I
restored HEAD's `go_v1.py` into a scratch tree and ran `mypy --platform linux`,
reproducing **exactly 13 `attr-defined` errors at lines 3731–3758** — the same
count as the CI failure, every one inside `_setup_macos`, and none outside it.
The cast covers precisely those symbols and nothing more; no unrelated typing
was loosened.

`_setup_macos` is reachable only under `platform == PLATFORM_MACOS`; every other
platform raises `CODE_CONTROL_UNAVAILABLE` in the same constructor. The comment
added in the diff is accurate. No Linux control path, host-label bypass, or
weakened fail-closed behavior was introduced — the diff contains nothing else.

## The test seam — the one risky change, adversarially checked

`tests/test_builds_go_v1.py` gains an autouse fixture that, **only when the host
is neither macOS nor Windows**, monkeypatches `go_v1._resolve_interpreter_runtime`
with a synthetic resolver so host-independent protocol/graph/evidence tests can
run on Linux CI. Two ways this could be wrong; both tested, both clean.

1. **Could it mask the real resolver on a supported host?** No. It early-returns
   on `sys.platform == "darwin"` / `os.name == "nt"`. Proven by mutation: I made
   the production `_resolve_macos_interpreter_runtime` raise
   `NotImplementedError` in a scratch copy and the macOS focused suite went
   **53 failed / 88 passed**. The real resolver is genuinely exercised by 53
   tests on macOS.

2. **Could it hide a regression in the unsupported-host fail-closed guarantee?**
   No. The suite captures the production resolver at import time
   (`_NATIVE_INTERPRETER_RUNTIME_RESOLVER`) before any patching, and the new
   `test_unsupported_host_interpreter_runtime_identity_fails_closed` asserts it
   raises `CODE_CONTROL_UNAVAILABLE`. Proven by mutation: I deleted the
   unsupported-host rejection from `_resolve_interpreter_runtime` and the
   forced-Linux run went red on exactly that test
   (`1 failed, 128 passed, 12 skipped`).

The seam is confined to `test_builds_go_v1.py` — it is not in `tests/conftest.py`
(whose only autouse fixture is the unrelated `stable_env`), so the real-Go
fixture module `tests/test_builds_go_v1_fixture.py` uses the production resolver.
That is consistent with its 10/10 native pass with real identity binding.

### Skip sets contain nothing load-bearing

The suite is the same 141 tests on all three hosts. macOS runs all 141 with zero
skips. The 12 Linux and 10 Windows skips are all genuine host-mechanism gates —
kqueue replacement/restore guard, `RLIMIT_NOFILE` retention, `os.killpg`,
macOS-specific native runtime-image layout, POSIX shebang, and the five
native-probe measurements. No mandatory control or fail-closed guarantee is
silently dropped while coverage is claimed; both unsupported-host tests run on
every host.

### Windows test-assumption corrections check out against product code

The changed expectations mirror real product behavior rather than being bent to
fit: `_manager_runtime_roots` returns `(python_home,)` on Windows (so
`runtime_trees[0].path == python_home`), `_manager_windows_stdlib_root` is
`python_home / "Lib"` (so the entry is `Lib/json/__init__.py`), and
`_runtime_archive_slots` returns `python_home / pythonXY.zip` on Windows versus
`stdlib_root.parent / pythonXY.zip` on POSIX — where the new
`startup.archive_slots[0]` equals the previously hardcoded POSIX value, so the
macOS assertion is behavior-preserving. The `-int(signal.SIGKILL)` → `-9` change
in the fake process is value-identical on POSIX, and the tests using it now skip
where `os.killpg` is absent.

## Independent gates run by this reviewer

| Gate | Host | Exit | Result |
| --- | --- | ---: | --- |
| Focused `tests/test_builds_go_v1.py`, accepted rc.5 root | macOS arm64, Py 3.14.4 | 0 | 141 passed, 0 skipped |
| Focused suite, `sys.platform` forced to `linux` | macOS process | 0 | 129 passed, 12 skipped |
| Focused suite, **native**, exact-byte overlay over `ssh win` | Windows amd64, Py 3.14.4 | 0 | **131 passed, 10 skipped** |
| Full `pytest -q` | macOS arm64 | 0 | 1,159 passed, 21 skipped |
| Full `pytest -q`, native | Windows amd64 | 1 | 1,063 passed, 116 skipped, 1 environmental failure (below) |
| `python -m mypy` (project-configured strict) | macOS | 0 | No issues in 65 source files |
| `python -m mypy --platform linux` | macOS | 0 | No issues in 65 source files |
| `mypy --platform linux` on **pre-fix** HEAD source | macOS | 1 | 13 `attr-defined`, lines 3731–3758, all in `_setup_macos` |
| `python -m build --wheel` + fresh-venv install | macOS | 0 | Installed `go_v1.py` byte-identical to source (`f8433dc1…`) |
| Wheel-installed native real-Go fixture | macOS arm64, Go 1.25.5 | 0 | 10 passed in 65.03s |
| Mutation: break macOS runtime resolver | macOS | 1 | 53 failed — seam proven inert on macOS |
| Mutation: remove unsupported-host rejection | forced Linux | 1 | fail-closed test goes red — guarantee proven live |
| `git diff --check`, tabnanny, compileall, conflict-marker and trailing-whitespace scan | macOS | 0 | Clean |

Reviewer JUnit SHA-256:

- macOS focused `96fc7c0eb0399ef4718f40033131a0362187acba98d3b867695f45aef2034739`
- Linux focused `12a09072389915e843e160adaff7c3e93883803b5e31b91cc829c69e4477a6af`
- Windows focused `231a87234ef780617d5ae7fc5498009bdfdc94cebff693c2dc2c9239ab1b43c6`
- macOS full `25d4e87899be5bebad7b07af451c33dcb97f526eaee1ed7745bfe731e627c3eb`
- macOS native fixture `17194c135a89c6659f4f274e2fac537919bdbcc892fd67fe6cd661d31d931572`

### Native Windows re-verification closes the standing caveat

Cycles 5 and 6 both carried native Windows evidence as a caveat because the
reviewer host was macOS. This reviewer ran it independently: `git archive` of
exact HEAD plus overlays of the two changed files and the pinned conformance
tree, uploaded over `ssh win`, remote SHA-256 confirmed equal to local bytes for
all four payloads, then the focused suite run natively. Result: 131 passed, 10
skipped, exit 0 — and the JUnit test set **and** skip set are identical to the
developer's attached `cycle6-windows-focused.xml`. The task-scoped remote root
was removed afterwards (`CLEANUP_OK`).

### The one non-zero gate is environmental, not the candidate

My Windows full run reported `1 failed`:
`tests/test_shell_init.py::test_powershell_hook_activates_and_restores_on_every_prompt`,
a PowerShell `UnauthorizedAccess` execution-policy error. `Get-ExecutionPolicy
-List` showed `Undefined` at every scope on my ssh session, which defaults to
`Restricted` for `-File` invocation. Setting the policy to `RemoteSigned` and
re-running that module gave `8 passed, 14 skipped, exit 0`. `test_shell_init.py`
contains zero references to `go_v1`. So this is a property of my remote session,
outside this task's scope, and the run is equivalent to the developer's
1,064 passed / 116 skipped.

## Evidence integrity

All seven attached cycle-6 artifacts hash exactly to the handoff table:
`cycle6-linux-focused-final.xml` `f190275512…`, `cycle6-macos-focused-final.xml`
`1ba917d6c1…`, `cycle6-macos-full-final.xml` `f506c9adb1…`,
`cycle6-macos-native-final.xml` `157ffc3768…`, `cycle6-windows-focused.xml`
`2e0a7cf9c4…`, `cycle6-windows-full.xml` `c019b91887…`, `cycle6-final-wheel.whl`
`63a45f87e2…`. The handoff also reports its non-zero and superseded diagnostics
as failures rather than passes, which matches what I could re-derive.

## Acceptance criteria

The cycle-5 AC verification stands — the product delta since then is only the
Darwin casts, which change no behavior. I re-confirmed the vector-pinned
elements directly: `test_accepted_rc5_vectors_match_the_implementation` asserts
`execution_policy`, `process_graph`, the 13 `session_states`, the 18
`mandatory_controls`, the native inventory version/platforms/5 controls, the
`capability_evidence_record` examples per platform, and the named
`identity_and_protocol_cases` (14), `package_influence_cases` (8) and
`capability_evidence_cases` (11) against the accepted vector files, plus the two
`source_aware` argv forms byte-for-byte. Output verification, the never-run
invariant, and the graph rejection surface pass on all three hosts.

## Non-blocking finding — record and follow up

**`fixed_environment` is asserted against a transcription, not against the
vector.** `test_worker_environment_matches_the_fixed_darwin_vector` compares
`plan.environment` — which `_worker_fixture` builds as a literal dict in the
same test file — against a second hardcoded literal. Neither side reads
`build-drivers.json`, and `grep -rn fixed_environment src/ tests/` returns
nothing, so it is the only AC-named vector element not pinned by
`test_accepted_rc5_vectors_match_the_implementation`.

Severity is limited, which is why this does not block:

- I verified the transcription directly against the accepted vector: 28/28 keys
  present and every value equal, so the fixture is conformant **today** and
  there is no correctness defect. This discharges the "fixed environment" item
  in the reviewer definition of done.
- Constructing this environment is not this task's scope — `go_v1` consumes
  `request.toolchain_session.environment`, and `_build_environment` lives in
  `csk/builds/toolchain.py` (TASK-260720-3j8pp5). This task's product duty is to
  *reject* a non-conformant environment, and `_validate_worker_environment` does
  enforce all of it (GOENV/GOTOOLCHAIN/locale/GOFLAGS/GOPROXY/GOSUMDB/GOPRIVATE/
  GONOPROXY/GONOSUMDB/GOVCS/GOWORK/CGO/GO_EXTLINK/GOEXPERIMENT/GOROOT, native
  GOOS plus the exact tuning variable, the eight operation-private roots as real
  non-reparse directories, and an empty PATH directory), exercised by the
  poisoned-environment negatives.

The residual risk is silent drift: if a later rc changes `fixed_environment`,
nothing fails. The fix is a near-one-liner because the test already normalizes
to the vector's exact placeholder tokens — replace the hardcoded expected dict
with `drivers["fixed_environment"]` in the existing vector test. Recommend
carrying this as a follow-up rather than a seventh rework cycle.

For the record, the cycle-5 verdict's phrasing that the environment "fixes the
vector `fixed_environment` exactly" was verified against the transcription, not
against the vector file. The conclusion holds; the provenance was overstated.

## Notes for the commit-owning mover

The cycle-6 delta is intentionally uncommitted and no commit authorization was
given, so no commit or push was made and no GitHub rerun was triggered. This
reviewer supplied no `commit_ack`. The commit-owning mover should commit
`src/csk/builds/go_v1.py` and `tests/test_builds_go_v1.py`, update the signed
PR #12 commit, and confirm the ubuntu strict-mypy job is green before the final
transition with `commit_ack=scope_committed`.
