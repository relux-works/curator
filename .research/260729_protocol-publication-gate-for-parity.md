# Protocol publication gate for Curator/CocoaSkills Go parity

Task: `TASK-260729-1kq1rd`  
Date: 2026-07-29  
Mode: read-only audit; no product file, index, commit, tag, release, or pin was
changed.

## Executive finding

The protocol-v6 verification is blocked by provenance, not by the accepted
candidate bytes.

- The canonical remote still publishes only `v1.0.0-rc.1`,
  `v1.0.0-rc.2`, and `v1.0.0-rc.3`. `origin/main` and
  `v1.0.0-rc.3^{}` both resolve to
  `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- The accepted rc.4 and rc.5 candidates are unstaged, uncommitted worktree
  snapshots on top of that rc.3 commit. An unwrapped release gate therefore
  rejects each with `release gate requires a clean candidate checkout`
  (real exit code 1).
- The current clean `origin/main` checkout rejects both requested versions
  because it is rc.3 (real exit code 1 for rc.4 and for rc.5).
- The accepted specification policy does not require publication of every
  prerelease identifier. It explicitly permits a release candidate to tighten
  or correct behavior before 1.0, and the accepted rc.5 compatibility text says
  neither rc.4 nor rc.5 has been released or pinned, so the corrected execution
  policy “lands in place.”

**Recommendation:** do not publish or land the stale rc.4 candidate. With board
owner approval to treat rc.5 as the superseding evidence for the rc.4-named
verification task, an authorized curator-spec integrator should land the exact
accepted `TASK-260728-2kp3tv` rc.5 snapshot as a real protected-default-branch
commit. Re-run the gates unwrapped at that clean commit. Keep the later
`TASK-260728-2jaw7h` `conformance/next`/schema-8 snapshot out of the rc.5 tag;
its mint/pin owner remains `TASK-260728-251p01`.

This needs no release-policy change. It does require:

1. a board/product-owner decision that rc.5 supersedes the literal rc.4
   acceptance wording in `TASK-260720-3ag6pi` and its rc.4-named downstream
   briefs; and
2. curator-spec maintainer authorization to merge the exact rc.5 snapshot,
   later sign `v1.0.0-rc.5`, and allow release CI to publish it.

## Authoritative refs and candidate bytes

### Public refs as observed

Read-only command:

```text
git -C /Users/iv/Developer/ReluxWorks/curator-spec \
  ls-remote --heads --tags origin
```

Exit code: **0**.

Relevant result:

| Ref | Peeled commit / observation |
| --- | --- |
| `refs/heads/main` | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| `refs/tags/v1.0.0-rc.3^{}` | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| `refs/tags/v1.0.0-rc.4` | absent |
| `refs/tags/v1.0.0-rc.5` | absent |

Every advertised remote head inspected identifies protocol rc.1, rc.2, or
rc.3 in `README.md`; none identifies rc.4 or rc.5. A local `show-ref` audit
agrees.

### Accepted rc.4 snapshot

Path:
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-q5oy3o/curator-spec-worktree`

| Property | Value |
| --- | --- |
| Base `HEAD` | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| Git state | 34 modified/untracked status entries; 0 staged paths |
| README version | `1.0.0-rc.4` |
| `conformance/v1/manifest.json` | `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae` |
| `conformance/v1/schema-cases/index.json` | `c437f5530693868ad11c852e0e083cbcbe3986ac3a5754c9882fe8f74f63020f` |
| Normalized snapshot | 269 files; SHA-256 `86b8028a0c848d7be5be247fb0c427d89d01a7a628dfc8c80e3a95981972fbf0` |

The normalized snapshot digest is over sorted relative paths, file/symlink
kind, mode, and content/link-target SHA-256, excluding `.git`, `.temp`,
`__pycache__`, and `.pytest_cache`.

The schema-6 wire bytes remain present unchanged in the accepted rc.5
snapshot:

| Artifact | SHA-256 in rc.4 and rc.5 |
| --- | --- |
| `agent-skill-v6.schema.json` | `982832e410f85e415e16e8f9104c3b9af23f6d846bbfbe5497ff170dde947f6f` |
| `csk-skill-v6.schema.json` | `2148eafc4fa110311b52f528651424e2f53c69042235338fb2c8b414035eab9c` |
| `build-receipt-v1.schema.json` | `f673a8815f5a5f752bc5b612f20c4ba63d9e8dcce61f5af6e7afe11b131c7ab9` |
| `install-marker-v2.schema.json` | `6d7b65dbdf684272815fb0e61cc4eb02103d09dfdd397de948bd836293debeb2` |
| `conformance-claim-v2.schema.json` | `4c05a97a1aa9f7dafe629a406a853239928413e79e95488ac2b20ebd0c52a38c` |

The rc.4 snapshot is nevertheless stale as a publication target: later
accepted decisions amend its direct-Go process graph and cache identity before
publication. Publishing it now would discard accepted security-policy work or
require producing and reviewing a different rc.4 candidate.

### Accepted releasable rc.5 snapshot

Path:
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`

This is the accepted predecessor whose rc.5 identity the later toolchain task
was required to preserve byte-for-byte.

| Property | Value |
| --- | --- |
| Base `HEAD` | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| Git state | 127 modified/untracked status entries; 0 staged paths |
| README version | `1.0.0-rc.5` |
| `release/1.0.0-rc.5.json` | `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441` |
| `conformance/v1/manifest.json` | `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` |
| `conformance/v1/schema-cases/index.json` | `2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0` |
| Normalized snapshot | 514 files; SHA-256 `3e4fd26acd9cafd1a76b2b5312da49ee35d234738263beb17a42be971d9dc582` |

`release/1.0.0-rc.5.json` names the manifest digest above as both the candidate
pin and required downstream digest, records
`committed_release_pin_advanced: false`, and requires candidate consumers to use
caller-supplied `CURATOR_CONFORMANCE_ROOT`.

### Later accepted but unreleased rc.5-adjacent snapshot

Path:
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2jaw7h/curator-spec-worktree`

| Property | Value |
| --- | --- |
| Base `HEAD` | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| Git state | 146 modified/untracked status entries; 0 staged paths |
| Frozen rc.5 release/manifest/index hashes | byte-identical to `2kp3tv` |
| Normalized snapshot | 705 files; SHA-256 `4983e90887be3efebe1bf81469ab107a7b4f8b0ee83bb683681b4e1766db161a` |
| New candidate surface | versionless `conformance/next`, schema 8/toolchain files |
| Mint/pin owner | `TASK-260728-251p01` |

This is not the rc.5 publication target. Its release workflow packages the
entire `conformance`, `schemas`, `decisions`, and `docs` directories. Tagging
that tree as rc.5 would therefore put versionless `conformance/next` and schema
8/toolchain artifacts in the rc.5 archives even though their manifest says
`released: false` and assigns mint ownership to `TASK-260728-251p01`.

## Gate reproduction and fact checks

All commands below ran as standalone processes. No output was piped through
`tee`, no Git wrapper or alternate-index clean-status shim was used, and the
real exit code is reported.

### Candidate validation

| Candidate | Command | Exit | Result |
| --- | --- | ---: | --- |
| rc.4 `q5oy3o` | `make validate` with the retained pinned venv on `PATH` | 0 | 35 schemas, 189 vectors, 27 Python tests, Go tools tests |
| rc.5 `2kp3tv` | `make validate` with the retained pinned venv on `PATH` | 0 | 42 schemas, 422 vectors, 29 Python tests, Go tools tests |
| later `2jaw7h` | `make validate` with the retained pinned venv on `PATH` | 0 | 48 schemas, 592 vectors (422 released + 170 candidate), 169 Python tests, Go tools tests |

These green validations establish internal candidate consistency. They do not
create a candidate commit or publication identity.

### Real unwrapped release-gate results

```text
PYTHONDONTWRITEBYTECODE=1 <pinned-python> \
  tools/release_gate.py --version <version> --commit HEAD
```

| Checkout | Requested version | Exit | Diagnostic |
| --- | --- | ---: | --- |
| accepted rc.4 `q5oy3o` worktree | `1.0.0-rc.4` | **1** | `release gate requires a clean candidate checkout` |
| accepted rc.5 `2kp3tv` worktree | `1.0.0-rc.5` | **1** | `release gate requires a clean candidate checkout` |
| later `2jaw7h` worktree | `1.0.0-rc.5` | **1** | `release gate requires a clean candidate checkout` |
| clean authoritative `origin/main` worktree | `1.0.0-rc.4` | **1** | `README version is not 1.0.0-rc.4` |
| clean authoritative `origin/main` worktree | `1.0.0-rc.5` | **1** | `README version is not 1.0.0-rc.5` |

These are expected-red results and remain failures. The exact blocker is that
no public/landed commit contains either accepted candidate snapshot.

The retained rc.5 reviewer evidence reports a clean disposable-probe
`make release-check VERSION=1.0.0-rc.5` exit 0. That proves the byte snapshot
can satisfy the gate after committing, but the scratch commit is deliberately
not treated as an authoritative ref, a release, or a permissible pin.

## Alternatives under the no-fabrication constraint

| Alternative | Compliant? | Tradeoff / decision |
| --- | --- | --- |
| Land accepted rc.4 `q5oy3o`, then verify rc.4 | Technically creates the missing ref, but **not recommended** against current accepted policy | It publishes a stale pre-amendment process graph/cache identity. Making rc.4 current would require another accepted composite, security review, and likely downstream retesting; the existing exact bytes should not be represented as current. |
| Land accepted rc.5 `2kp3tv`, approve rc.5 supersession, verify rc.5 | **Recommended** | Preserves schema-6 wire bytes, includes the accepted pre-publication execution-policy corrections, and follows the existing prerelease compatibility policy. Requires a task-contract approval because the old board briefs literally say rc.4. |
| Tag/publish later `2jaw7h` as rc.5 | **No** | Would package versionless `conformance/next` and schema-8/toolchain work under the rc.5 archive despite deferred mint ownership. Land/publish it only through its next-version owner after rc.5. |
| Use a disposable commit, alternate index, Git wrapper, guessed SHA, or mutable branch as release evidence | **No** | Repeats the rejected provenance fabrication: the reported commit would not authoritatively identify the candidate bytes. |
| Change `release_gate.py` to accept dirty worktrees | Possible only through an explicit governance/release-policy change; **not recommended** | Weakens the candidate-commit invariant and still does not create a signed public release or a pin-safe immutable ref. |
| Reclassify `3ag6pi` as pre-publication-only and remove release acceptance | Compliant as a scope reduction, but **insufficient** for final parity | It could accept conformance content but cannot unlock the published-release and manager-pin gates. |

## Recommended approval and command sequence

The commands below are a runbook for the authorized owners. They were not
executed by this research task.

### 1. Board decision before any repository mutation

The board/product owner records:

> Protocol `1.0.0-rc.5` supersedes the never-published rc.4 candidate for
> `TASK-260720-3ag6pi` and all rc.4-named candidate-consumption briefs.
> Evidence must identify the exact landed rc.5 commit and released-suite
> manifest SHA-256
> `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.
> No committed manager pin advances until `TASK-260720-25d05o`.

This is a task-contract/acceptance decision, not a release-policy amendment.

### 2. Land exactly the releasable rc.5 snapshot

An authorized curator-spec integrator creates a fresh isolated worktree from
exact `origin/main` `57c1f568…`, copies the complete `2kp3tv` product snapshot
while excluding only `.git`, `.temp`, caches, and scratch artifacts, and then
verifies:

```text
git rev-parse HEAD
git status --porcelain=v1
git diff --cached --check
shasum -a 256 release/1.0.0-rc.5.json \
  conformance/v1/manifest.json \
  conformance/v1/schema-cases/index.json
```

Expected hashes are the three `b32ee9d…`, `9ba9b8e…`, and `2faa2ba…` values
above. The integrator then creates the real integration commit, sends it through
protected-default-branch review/checks, and merges it. The landed full SHA,
not the worktree base or a scratch commit, becomes the only candidate ref.

### 3. Verify the landed commit unwrapped

From a clean detached checkout at the landed SHA:

```text
make validate
make regenerate-check
make release-check VERSION=1.0.0-rc.5
git status --porcelain=v1
```

Required exits: 0, 0, 0, 0; the final status output must be empty. Record the
landed SHA, suite-manifest SHA-256, command logs, and exact clean checkout.

The reviewer then returns `TASK-260720-3ag6pi` from `blocked` to `reviewing` and
accepts it only if the rc.5 supersession approval and all unwrapped evidence are
present.

### 4. Candidate parity and manager-release sequence

Keep manager committed suite pins on the previous published release. Supply the
landed rc.5 root/digest explicitly while completing:

1. `TASK-260720-12r55p` → `TASK-260720-3pemm6` →
   `TASK-260720-3s27te` for the independent CocoaSkills schema-6/Go consumer
   and cross-platform integration.
2. `TASK-260720-2g7avf`; then both consumers
   `TASK-260720-1673lr` and `TASK-260720-31zeo2`; then
   `TASK-260720-3nj1r6`.
3. Complete the remaining manager integration/docs/admission blockers and
   publish real Curator and CocoaSkills manager releases against the same
   candidate digest.
4. `TASK-260720-3pvihp` qualifies those real manager releases.
5. `TASK-260720-vs6den` pins those manager release commits in curator-spec and
   runs the in-tree parity matrix.

### 5. Protocol publication and released pins

After the protected default branch and implementation-conformance checks are
green, an authorized curator-spec maintainer:

```text
git tag -s -a v1.0.0-rc.5 <qualified-landed-sha>
git verify-tag v1.0.0-rc.5
git push origin v1.0.0-rc.5
```

Release CI must verify the tag/default-branch containment, rerun the full gate,
publish the normative archives/checksums, and create provenance attestations.

Then:

1. `TASK-260720-25d05o` qualifies the public tag, commit, archive, checksum,
   manifest digest, signature, and cross-platform specification CI.
2. `TASK-260720-38l1sy` and `TASK-260720-1utsx8` audit the Curator and
   CocoaSkills committed released-suite pins.
3. `TASK-260720-22ynoi` performs final cross-manager Go interop acceptance.

Only after rc.5 publication should the separately owned `2jaw7h`
`conformance/next` work proceed through `TASK-260728-251p01` to a newly minted
protocol version.

## Board dependency impact

### Immediate hard edges

`TASK-260720-3ag6pi` currently has two direct downstream dependents:

- `TASK-260720-12r55p` — also waits for `TASK-260720-th0jdi`;
- `TASK-260720-3pvihp` — also waits for the runner, documents/admission matrix,
  Curator integration, and CocoaSkills integration.

Clearing the publication gate removes one hard edge from each; it does not
override their other blockers.

### Full reverse transitive closure

The board query

```text
task-board q --format json \
  'list(type=task) { id name status parent blockedBy }'
```

exited 0. A shortest-edge reverse dependency walk found **47 downstream tasks**
in addition to `TASK-260720-3ag6pi`. All were `backlog` at audit time.

Grouped by shortest dependency distance:

- Distance 1: `TASK-260720-12r55p`, `TASK-260720-3pvihp`.
- Distance 2: `TASK-260720-31zeo2`, `TASK-260720-3pemm6`,
  `TASK-260720-vs6den`.
- Distance 3: `TASK-260720-22ynoi`, `TASK-260720-25d05o`,
  `TASK-260720-3nj1r6`, `TASK-260720-3s27te`,
  `TASK-260728-1e6811`.
- Distance 4: `TASK-260720-1utsx8`, `TASK-260720-2g7avf`,
  `TASK-260720-38l1sy`, `TASK-260720-p7sdhg`,
  `TASK-260728-1j72zq`, `TASK-260728-1ph8rs`,
  `TASK-260728-2yxdo7`, `TASK-260728-3ar1qp`,
  `TASK-260728-3j60e3`, `TASK-260728-3lqm4z`,
  `TASK-260728-asbuoo`, `TASK-260728-r3j8ef`.
- Distance 5: `TASK-260720-14jjgt`, `TASK-260720-1673lr`,
  `TASK-260728-1egim2`, `TASK-260728-1t59zp`,
  `TASK-260728-1uj0bc`, `TASK-260728-1y8u4m`,
  `TASK-260728-26e3n2`, `TASK-260728-2mfeje`,
  `TASK-260728-2uh7em`, `TASK-260728-2ztr3c`,
  `TASK-260728-3u1nho`, `TASK-260728-3vtl57`,
  `TASK-260728-gjxj1v`.
- Distance 6: `TASK-260728-16kefa`, `TASK-260728-1aveb2`,
  `TASK-260728-1xviwb`, `TASK-260728-2uxmut`,
  `TASK-260728-ypbuav`.
- Distance 7: `TASK-260728-3jaa57`.
- Distance 8: `TASK-260728-3kuxg7`.
- Distance 9: `TASK-260728-2u5u14`.
- Distance 10: `TASK-260728-1hwq5b`, `TASK-260728-1t4cyb`.
- Distance 11: `TASK-260728-d8ktna`.
- Distance 12: `TASK-260728-1skseh`.

The long tail includes future schema-7 and additional-driver work because those
tasks depend on the CocoaSkills v6 integration chain. They become less blocked,
not automatically runnable; every other listed blocker remains binding.

## Irreducible human authorization

1. **Board/product authority:** approve rc.5 as superseding the literal rc.4
   task contract. Without this, a reviewer cannot honestly check an rc.4 command
   item based on rc.5 evidence.
2. **curator-spec integration authority:** authorize the exact `2kp3tv` bytes
   for a real commit/merge on the protected default branch. The current
   candidate has no commit identity.
3. **curator-spec release maintainer:** sign the exact qualified landed commit
   as `v1.0.0-rc.5` and authorize publication. Governance makes this
   authorization cryptographic and cannot be inferred from a board status.
4. **External release owners later in the chain:** publish real Curator and
   CocoaSkills releases before `TASK-260720-3pvihp`; local worktrees and future
   SHAs cannot substitute for those artifacts.

If item 1 is denied, the honest alternatives are either to commission and
review a current rc.4 candidate under the later security policy or to make an
explicit release/task-policy change. The retained stale rc.4 snapshot must not
be silently presented as current.

## Sources

- `curator-spec/GOVERNANCE.md`, release process and immutable-tag policy.
- Accepted rc.5 `COMPATIBILITY.md`, prerelease correction rule and explicit
  “neither rc.4 nor rc.5 has been released or pinned” statement.
- Accepted rc.5 `release/1.0.0-rc.5.json`, candidate pin and downstream pin
  ownership boundary.
- Accepted rc.5 `.github/workflows/release.yml`, complete release archive
  contents and tag verification.
- Board outcome
  `.task-board/.resources/TASK-260720-3ag6pi/TASK-260720-3ag6pi_reviewer-verdict.md`.
- Board outcome
  `.task-board/.resources/TASK-260728-2kp3tv/TASK-260728-2kp3tv_review-verdict.md`.
- Board outcomes
  `.task-board/.resources/TASK-260728-2jaw7h/TASK-260728-2jaw7h_rework-cycle-2.md`
  and `TASK-260728-2jaw7h_review-verdict-cycle-2.md`.
- Live read-only `git ls-remote`, `show-ref`, worktree, SHA-256, release-gate,
  validation, and task-board dependency queries recorded above.

