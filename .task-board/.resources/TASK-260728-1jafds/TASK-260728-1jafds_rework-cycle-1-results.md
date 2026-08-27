# TASK-260728-1jafds — rework cycle 1

Status: ready for review. Specification only. Nothing was staged, committed, or
published in `curator-spec`, and no native implementation or platform
qualification is claimed.

Work is in the existing task worktree
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`.

All three blockers in `TASK-260728-1jafds_review-verdict-cycle-1.md` are closed.

---

## Blocker 1 — identity binding did not satisfy the task contract

**Was:** the profile identity and the TCB record were structurally excluded from
the cache key and from receipt bytes, in order to reproduce the cache key
`sha256:13736230…` that rc.5 recorded for the reserved hardened policy slot.

**Now:** one mechanically enforced model, `hardened-identity-binding-v1`.

A hardened build input is the portable input with the `execution_policy` value
replaced **and exactly one closed member added**:

```json
{
  "hardened": {
    "profile": "hardened-profile-v1",
    "tcb": {"algorithm": "curator-hardened-tcb-v1", "content_sha256": "sha256:…"}
  }
}
```

`curator-hardened-tcb-v1` is defined in `protocol/hardened-execution.md` § 2.3:
SHA-256 initialized with ASCII `curator-hardened-tcb-v1` + `0x00`, then
`uint64be(length)` and the CCJ-1 canonical bytes of the closed `hardened-tcb-v1`
record. Domain-separated and length-framed in the same house style as
`curator-build-source-v1`, so the digest can never be confused with a cache key
over the same canonical bytes.

Consequences, all checked against real artifacts rather than asserted:

| Identity | Hashed input | Receipt bytes | Marker | Claim | Binds cache reuse |
|---|---|---|---|---|---|
| `hardened-worker-v1` | yes | yes | yes | yes | yes |
| `hardened-profile-v1` | yes | yes | yes | yes | yes |
| `hardened-tcb-v1` | yes, as digest | yes, complete record | yes | yes | yes |
| `hardened-capability-evidence-v1` | no | no | no | no | no |

- Receipt schemas 3 and 4 now **require** `tcb`. `invalid-tcb-in-receipt` was
  inverted into `invalid-missing-tcb`, plus four new negatives covering a
  missing `hardened` member, a missing profile identity, a missing TCB digest,
  and a wrong digest algorithm.
- A reader must reject a receipt whose `input.hardened.tcb.content_sha256` is
  not the digest of its own `tcb` record, and whose `cache_key` is not the
  digest of its own input.
- A marker's `cache_key` must be reproducible from the identities the marker
  itself publishes.
- The reserved rc.5 key is **not** reproduced. rc.5 marks that input
  `schema_valid: false`; it is retained as a fourth comparison point. The suite
  now proves **five** distinct non-aliasing keys over one source: pre-revision,
  portable, rc.5 policy-slot-only, hardened under TCB A, hardened under TCB B.

The cost is stated, not hidden: hardened cache identity now diverges per trusted
computing base, so a supervisor or worker update forces a rebuild.
`protocol/hardened-execution.md` § 8.3, `decisions/0009`, and the docs page all
argue why that is the fail-closed answer — an artifact carries exactly the
guarantees of the trusted code that produced it — and why this is *not* the case
`decisions/0006` rejected, which was about per-operation evidence, not identity.

## Blocker 2 — the self-test was not executable under its own process graph

**Was:** phase 6 `domain-guarantee-self-test` ran "from inside the domain" but
was ordered before phase 7 `domain-entry`, and the worker was defined as the
first process inside the domain. No actor could perform it.

**Now:** every phase names the one actor that performs it, and one rule makes
the ordering enforceable:

> **No phase performed by a process inside the build domain may precede
> `domain-entry`.**

Order is `domain-establishment` → `domain-entry` → `in-domain-guarantee-self-test`
→ `go-list`. The supervisor creates the domain and launches the worker into it;
the worker — the only contained actor — then self-tests from inside and reports
over the pre-opened session channel. Until the supervisor accepts that result the
worker opens no path below the source view and starts no Go process.

`in_domain_self_test` in the profile vector names the actor, the surrounding
phases, the verifier, the channel, and the failure behavior. The validator
rejects any phase whose actor is `domain-root-worker` at an index ≤ the index of
`domain-entry`, so the old ordering cannot be reintroduced.

The failure boundary is now stated as **two** boundaries instead of one
overclaimed boundary:

- **before domain entry** — phases 1–7. Every capability, qualification, and
  TCB-identity rejection lands here, so an unsupported host creates no domain.
- **before package exposure** — phases 8–9.

Both are strictly before `go list`, `go build`, and any compiler.

## Blocker 3 — two conflicting normative phase lists

**Was:** `protocol/hardened-execution.md` § 7.2 published 15 phases without the
toolchain probe; `profiles/manager-hardened.md` published a different 15-step
list with the probe at 3 and permit+build combined at 11. The vector followed
only the protocol.

**Now:** one authoritative list of **17** phases in `protocol/hardened-execution.md`
§ 7.2, with a `# | Phase | Actor | Rejection diagnostic` table. The missing
package-independent `toolchain-probe-and-snapshot-freeze` is phase 4, and
`build-input-and-cache-lookup` is phase 6 (needed because the TCB now binds the
key, and because an exact hit must skip domain creation).

`profiles/manager-hardened.md` § 2 was rewritten as a **phase-keyed obligation
table** that publishes no ordering of its own, and says so explicitly.

Anti-drift is mechanical: `validate_phase_list_documents()` parses the numbered
table out of protocol § 7.2 and the phase column out of manager profile § 2 and
fails if either drifts from the validator's list or from the vector — and fails
if the manager profile reintroduces a numbered list.

---

## Files changed in this cycle

Rewritten or substantially revised:

- `protocol/hardened-execution.md` — §§ 2, 4.1, 4.3, 6.1, 7.2, 7.3, 8.1–8.5, 9
- `profiles/manager-hardened.md` — §§ 2, 3, 4, 5, 8
- `tools/generate-hardened/main.go`, `tools/generate-hardened/main_test.go`
- `tools/validate_hardened.py`, `tools/test_validate_hardened.py`
- `schemas/hardened/v1/hardened-common.schema.json`,
  `hardened-build-receipt-v3.schema.json`, `hardened-build-receipt-v4.schema.json`
- `conformance/hardened/v1/` — regenerated (49 files, was 43)
- `decisions/0009-hardened-build-execution-profile.md`,
  `docs/hardened-build-execution-profile.md`
- `SECURITY.md`, `CHANGELOG.md`, `COMPATIBILITY.md`

Change set relative to the accepted rc.5 predecessor
(`.temp/TASK-260728-2kp3tv/curator-spec-worktree`), by `diff -rq`:

```
added:    conformance/hardened, schemas/hardened, protocol/hardened-execution.md,
          profiles/manager-hardened.md, decisions/0009…, docs/hardened…,
          release/hardened-1.0.0-rc.1.json, tools/generate-hardened,
          tools/validate_hardened.py, tools/test_validate_hardened.py
modified: README.md, SECURITY.md, CHANGELOG.md, COMPATIBILITY.md, Makefile,
          .github/workflows/ci.yml, .github/workflows/release.yml
```

Untouched: `conformance/v1`, `schemas/v1`, `release/1.0.0-rc.5.json`,
`protocol/core.md`, `protocol/registry.md`, `profiles/manager.md`,
`profiles/registry-service.md`, `cli/curator.md`, `decisions/0001`–`0006`,
`docs/portable-go-execution-policy.md`, `docs/external-build-repositories.md`,
`tools/validate.py`, `tools/generate-vectors/`, `tools/release_gate.py`.

## Verification actually run

Every command was run as a standalone process; the exit codes below are the real
exit codes of the command itself, not of a pipe.

| Command | Exit | Result |
|---|---|---|
| `make validate` | 0 | 42 portable + 6 hardened schemas, 422 + 48 suite files, **87 tests** (was 64) |
| `python3 tools/validate_hardened.py` | 0 | standalone |
| `go test -count=1 ./tools/...` | 0 | both packages |
| `gofmt -l tools` | 0 | no output |
| `git diff --check` | 0 | clean |
| `diff -r <pred>/conformance/v1 conformance/v1` | 0 | identical after regenerating with the portable generator |
| `diff -r <pred>/schemas/v1 schemas/v1` | 0 | identical |
| `diff <pred>/release/1.0.0-rc.5.json …` | 0 | identical |
| `shasum -a 256 conformance/v1/manifest.json` | 0 | `9ba9b8ec…`, equals the rc.5 release pin |
| hardened regenerate ×2, digest compare | 0 | byte stable |

### The cycle-1 expected-red gates are now resolved, not re-excused

In the task worktree the three git-comparing gates are still structurally red,
because the whole rc.5 candidate plus this hardened material is uncommitted
against `57c1f56`:

| Command (task worktree) | Real exit |
|---|---|
| `make regenerate-check` | 2 |
| `make regenerate-hardened-check` | 2 |
| `python3 tools/release_gate.py --version 1.0.0-rc.5 --commit HEAD` | 1 (`requires a clean candidate checkout`) |

To resolve them without staging or committing anything in `curator-spec`, the
tree was copied into a throwaway scratch git repository under
`.temp/clean-probe`, committed there, and the gates re-run. `diff -r` between
the probe and the task worktree was empty, so the evidence applies to exactly
these bytes. The probe was deleted afterwards; the log is
`.temp/rework-clean-probe-gates.log`.

| Command (clean probe) | Real exit |
|---|---|
| `make validate` | **0** |
| `make regenerate-check` | **0** |
| `make regenerate-hardened-check` | **0** |
| `python3 tools/release_gate.py --version 1.0.0-rc.5 --commit HEAD` | **0** |

The last row also answers the reviewer's question about CI: adding the hardened
gates does not alter rc.5 qualification — the rc.5 release gate still passes.

## Test coverage added this cycle

Python (`tools/test_validate_hardened.py`, 58 tests, was 35). New mutants:

- the exact cycle-1 ordering — self-test swapped before `domain-entry` — is
  rejected;
- any in-domain actor placed before `domain-entry` is rejected;
- the self-test block naming an uncontained actor, or moved past package
  exposure, or permitting a partial mode, is rejected;
- an ordering invariant that is dropped, or reversed, is rejected;
- a phase with no graph actor, or a graph node claiming a phase that names
  another actor, is rejected;
- an unbound profile identity, an unbound TCB, capability evidence leaking into
  a reusable output, and permitted cross-TCB reuse are each rejected;
- a receipt whose TCB digest lies, and a receipt whose cache key ignores the
  binding, are detected on real files;
- a marker reporting a different TCB beside its key is detected;
- protocol § 7.2 drift and manager-profile drift are each detected;
- a pre-entry rejection moved past entry, and a rejection moved past package
  exposure, are each detected.

Go (`tools/generate-hardened/main_test.go`), new or rewritten: recorded rc.5
policy-slot keys still reproduce; the hardened input is the portable input plus
exactly one closed member; rotating the TCB and changing the profile identity
each move the cache key and five keys stay distinct; the TCB digest is
domain-separated; receipts bind profile + TCB + key; no in-domain actor precedes
domain entry; every ordering invariant holds; the process graph and the phase
list agree.

## What is still deliberately not claimed

Unchanged from cycle 1, and re-verified:

- **No platform is qualified.** `linux`, `macos`, `windows` are all
  `unqualified` with `native_evidence: "absent"`; `claims_emitted` and
  `qualified_platforms` are both `[]`, and the validator fails if either is
  populated. Every host therefore rejects with `hardened_profile_unsupported`.
- **No native adversarial test was run on any platform.** Every adversarial case
  remains `pending-native-validation`. No Linux host was available; the directive
  states that is non-gating for a specification task.
- The macOS and Windows blocking findings restate `decisions/0006`'s
  `no-private-aggregate-domain` analysis; they are not new claims.
