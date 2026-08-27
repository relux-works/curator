# TASK-260728-1jafds — specify-hardened-build-execution-profile

Status: ready for review. Specification only. Nothing was staged, committed, or
published, and no native implementation or platform qualification is claimed.

## Where the work is

Task-owned worktree:
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`

It was created from `curator-spec` commit `57c1f56` and then made byte-identical
to the accepted rc.5 candidate at
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`
(`diff -r` empty at start). The predecessor was never written to.

## Design decision that shaped everything else

`conformance/v1/manifest.json` is hashed into `release/1.0.0-rc.5.json` as
`sha256:9ba9b8ec…`, and that pin is consumed downstream. Any hardened vector or
schema case inside `conformance/v1` would move it. `validate.py` also requires
every schema in `schemas/v1` to have cases inside `conformance/v1`, so a
hardened schema could not go there either.

So the hardened profile is authored as an **additive, separately versioned
candidate**, `hardened-1.0.0-rc.1`, with its own suite root, schemas, validator,
generator, and release pin. That matches the epic ("separately versioned") and
decision 0006's follow-up clause ("its own execution-policy identity, its own
claim schema version, its own adversarial vectors; not enabled by widening the
closed `manager-worker-v1` constant").

## What was specified

**Identities.** The execution policy is `hardened-worker-v1` — the exact value
rc.5 reserved, not a new one. It occupies exactly the slot the portable identity
occupies inside the canonical build policy object, and nothing is added to,
removed from, or reordered within the hashed input. That reproduces the hardened
cache key `conformance/v1` already reserved:

| Input | Execution policy | Key |
|---|---|---|
| pre-revision rc.4 | absent | `sha256:3fcd714a…` |
| portable | `manager-worker-v1` | `sha256:52937012…` |
| hardened | `hardened-worker-v1` | `sha256:13736230…` |

Three keys, no aliasing, and the hardened one is recomputed by both the Go
generator and the Python validator rather than copied.

Profile identity `hardened-profile-v1` is bound one-to-one to the execution
policy and appears in markers and claims. `hardened-tcb-v1` names the supervisor
bytes, worker bytes, enforcement backend, and toolchain. Neither, nor the
platform, nor the capability evidence, is hashed into a build identity —
deliberately, for the reason decision 0006 gave about portable evidence.

**Six guarantees**, each stated as something the kernel or hypervisor refuses to
do, each with an explicit "not sufficient" list naming the manager-side
mechanism that must not be presented as establishing it.

**Process graph** gains one uncontained node (the hardened supervisor) between
the manager parent and the worker, because a process cannot be both the creator
of its own confinement and confined by it from its first instruction. The worker
is the first process inside the build domain.

**Capability inventory** `hardened-capability-inventory-v1`: eleven exhaustive
classes mapped many-to-one onto the six guarantees, probed on the host in the
operation before domain entry. A host label, a version comparison, a build-time
constant, a config file, and a cached result are explicitly not probes.

**Fifteen ordered phases.** Phases 1–7 (selection, qualification, probe, TCB
verification, domain establishment, in-domain guarantee self-test, entry) all
complete before domain entry, and domain entry precedes `go list`. So every
capability rejection is strictly earlier than the portable boundary: no package
byte reaches any Go process. No partial mode, no silent fallback to portable.

**Cache/receipt/marker/claim separation.** Hardened builds write receipt
schema 3 (`go-v1`) or 4 (`go-repository-v1`), install marker schema 4, and claim
schema 4. Portable schemas keep their exact bytes. Claim 3 admits only
`manager-worker-v1` and claim 4 only `hardened-worker-v1`, so they are
structurally disjoint in both directions — proved by running the real validators
across families, not asserted.

**Nine stable `hardened_*` `phase: execution` diagnostics**, disjoint from the
six portable `build_execution_*` codes.

**Adversarial vectors**: 26 in-domain escape attempts across the five
containment guarantees, 13 forced-unavailable preflight cases (one per class
plus unprobed and unqualified), 15 package-influence surfaces, 17 identity and
protocol negatives, 16 evidence negatives, 4 no-fallback cases. All marked
`pending-native-validation`.

## What is deliberately not claimed

**No platform is qualified.** Linux, macOS, and Windows each carry a declaration
naming an enforcement backend and its candidate public primitives, all
`unqualified`, `native_evidence: "absent"`, bound to their owning tasks. The
observable consequence today is that every host rejects the hardened profile in
phase 2 with `hardened_profile_unsupported`, and no claim-4 document can be
emitted. `claims_emitted` and `qualified_platforms` are both `[]`, and the
validator fails if either is populated.

Blocking findings, all restating decision 0006's `no-private-aggregate-domain`
analysis rather than contradicting it:

- **macOS** — `domain-membership-enforcement`, `domain-atomic-termination`,
  `aggregate-resource-bounds`: no unescapable per-operation process domain (a
  contained process can leave the process group or session) and no aggregate
  private storage, memory, or process-count accounting.
- **Windows** — `exec-path-allowlist` (child-process policy is all-or-none, no
  supported per-path allowlist for a contained token) and
  `aggregate-resource-bounds` (no supported facility bounds bytes written below
  the private build root).
- **Linux** — nothing identified as unreachable, but a binding is a design, not
  a proof. No Linux host was available to this task, which the directive states
  is non-gating for specification work.

## Files

Added:

- `protocol/hardened-execution.md`, `profiles/manager-hardened.md`
- `decisions/0009-hardened-build-execution-profile.md` (0007 and 0008 are
  reserved by sibling candidate tasks not in this change set)
- `docs/hardened-build-execution-profile.md`
- `schemas/hardened/v1/` — hardened common defs, receipt v3, receipt v4, install
  marker v4, conformance claim v4, capability evidence v1
- `conformance/hardened/v1/` — manifest, three vectors, 33 schema cases
- `release/hardened-1.0.0-rc.1.json`
- `tools/generate-hardened/` (+ tests), `tools/validate_hardened.py`,
  `tools/test_validate_hardened.py`

Modified (prose, CI, Makefile only — no pinned or generated protocol byte):
`README.md`, `SECURITY.md`, `CHANGELOG.md`, `COMPATIBILITY.md`, `Makefile`,
`.github/workflows/ci.yml`, `.github/workflows/release.yml`.

Untouched: `conformance/v1`, `schemas/v1`, `release/1.0.0-rc.5.json`,
`protocol/core.md`, `protocol/registry.md`, `profiles/manager.md`,
`profiles/registry-service.md`, `cli/curator.md`, `decisions/0001`–`0006`,
`docs/portable-go-execution-policy.md`, `docs/external-build-repositories.md`,
`tools/validate.py`, `tools/generate-vectors/`, `tools/release_gate.py`.

## Verification actually run

All commands were run as standalone processes with their real exit codes.

| Command | Exit | Result |
|---|---|---|
| `make validate` (baseline, before any change) | 0 | 42 schemas, 422 vector files; 29 tests |
| `make validate` (final) | 0 | 42 portable + 6 hardened schemas, 422 + 42 suite files; **64 tests** |
| `python3 tools/validate_hardened.py` | 0 | 6 hardened schemas, 42 hardened suite files |
| `go test ./tools/...` | 0 | `generate-hardened` and `generate-vectors` both ok |
| `gofmt -l tools` | 0 | empty |
| `git diff --check` | 0 | clean |
| `diff -r <predecessor>/conformance/v1 conformance/v1` | 0 | identical after regenerating with the portable generator |
| `diff <predecessor>/release/1.0.0-rc.5.json …` | 0 | identical |
| `diff -r <predecessor>/schemas/v1 schemas/v1` | 0 | identical |
| `shasum -a 256 conformance/v1/manifest.json` | 0 | `9ba9b8ec…`, equals the rc.5 release pin |
| hardened regenerate ×2, digest compare | 0 | byte stable |

Test coverage added: 35 Python tests (`test_validate_hardened.py`) and 8 Go
tests (`tools/generate-hardened/main_test.go`). The Python tests are negative:
each mutates a real artifact — in a temp copy of the tree for file-level checks —
and asserts the validator rejects it. Covered: fabricating a qualified platform,
fabricating a claim, misreporting the portable baseline, claiming native
evidence, claiming a guarantee established, a hardened diagnostic colliding with
a portable code, a phase exposing package bytes before domain entry, a rejection
that publishes, a preflight case falling back to portable, a capability class
with no forced-unavailable case, a tampered hardened vector, a stray hardened
file, a hardened leak into the rc.5 manifest, and dropped negative schema cases.

### Gates that are expected-red, reported honestly

| Command | Real exit | Why |
|---|---|---|
| `make regenerate-check` | **2** | Its `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json` compares against committed `57c1f56`, and the entire rc.5 candidate is uncommitted. Verified structural, not caused by this change: `git diff` produces byte-identical output in the read-only predecessor worktree (`diff` of the two diffs is empty, both exit 1). |
| `make regenerate-hardened-check` | **2** | Same rc.5 line. Its own hardened line passes vacuously, because `git diff` does not see untracked files. Real hardened byte stability was proved by the digest comparison above and by `TestSuiteGenerationIsByteStable`. |
| `python3 tools/release_gate.py --version 1.0.0-rc.5 --commit HEAD` | **1** | `release gate requires a clean candidate checkout`. A clean checkout is impossible without committing, which the directive forbids. |

No native adversarial test was run for any guarantee on any platform. That is
the point of `pending-native-validation`, and it is why every platform
declaration is `unqualified`.

## Notes for review

- The `hardened-1.0.0-rc.1` CHANGELOG section sits above `1.0.0-rc.5`. It is a
  separate version line for a separately pinned candidate, not a supersession.
- `python3 tools/validate_hardened.py` was wired into `make validate` and into
  both CI workflows, so the hardened suite is now a gate on every run — including
  the rc.5 release path, which proves the addition does not disturb rc.5.
- A `.venv/` and a `.temp/` exist in the worktree for tooling and logs. Both are
  in `validate.py`'s non-surface directory list and are not part of the change.
