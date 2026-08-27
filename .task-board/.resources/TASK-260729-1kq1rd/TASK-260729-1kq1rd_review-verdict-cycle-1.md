# TASK-260729-1kq1rd reviewer verdict — cycle 1

## ACCEPTED

Route: `done`. The research outcome satisfies the task acceptance criteria and
the reviewer found no rework issue. This verdict accepts the read-only
publication-gate analysis; it does not claim that rc.5 has been landed,
published, or pinned.

The review was read-only. It did not edit product files, stage or commit
changes, create or move refs, publish artifacts, mint a protocol version, or
advance a manager pin.

The tracked run is not goal-bound:

```text
task-board spawn goal RUN-260728-f4b349
Active Goal: none (run is not goal-bound)
```

## Independently reproduced authority and candidate identity

`git ls-remote --heads --tags origin` in the authoritative `curator-spec`
clone exited 0 and resolved:

- `refs/heads/main` to
  `57c1f56846d221ecc55786bd3c2467ec32f11730`;
- `refs/tags/v1.0.0-rc.3^{}` to the same commit;
- no `v1.0.0-rc.4` or `v1.0.0-rc.5` tag.

The three retained candidates all have base `HEAD` `57c1f568...`, zero staged
paths, and the following independently recomputed identities:

| Candidate | Status entries | Normalized snapshot identity |
| --- | ---: | --- |
| rc.4 `TASK-260720-q5oy3o` | 34 | 269 files, SHA-256 `86b8028a0c848d7be5be247fb0c427d89d01a7a628dfc8c80e3a95981972fbf0` |
| releasable rc.5 `TASK-260728-2kp3tv` | 127 | 514 files, SHA-256 `3e4fd26acd9cafd1a76b2b5312da49ee35d234738263beb17a42be971d9dc582` |
| later rc.5-adjacent `TASK-260728-2jaw7h` | 146 | 705 files, SHA-256 `4983e90887be3efebe1bf81469ab107a7b4f8b0ee83bb683681b4e1766db161a` |

The normalized identity covers sorted relative path, file/symlink kind, mode,
and content/link-target SHA-256 while excluding only `.git`, `.temp`, Python
bytecode caches, and pytest caches.

The exact accepted rc.5 publication input is the `2kp3tv` snapshot:

- `release/1.0.0-rc.5.json`:
  `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`;
- `conformance/v1/manifest.json`:
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`;
- `conformance/v1/schema-cases/index.json`:
  `2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0`.

The later `2jaw7h` snapshot preserves all three hashes byte-for-byte, but its
`conformance/next/manifest.json` has no protocol version, records
`released: false`, and names `TASK-260728-251p01` as `release_pin_owner`.
Its release workflow copies the complete `conformance`, `schemas`,
`decisions`, and `docs` directories into release archives. It therefore cannot
be tagged as rc.5 without also publishing deferred schema-8 and
`conformance/next` content.

The five schema-6 wire artifacts named in the producer packet were independently
hashed in rc.4 and rc.5 and are byte-identical:

- `agent-skill-v6.schema.json`: `982832e410f85e415e16e8f9104c3b9af23f6d846bbfbe5497ff170dde947f6f`;
- `csk-skill-v6.schema.json`: `2148eafc4fa110311b52f528651424e2f53c69042235338fb2c8b414035eab9c`;
- `build-receipt-v1.schema.json`: `f673a8815f5a5f752bc5b612f20c4ba63d9e8dcce61f5af6e7afe11b131c7ab9`;
- `install-marker-v2.schema.json`: `6d7b65dbdf684272815fb0e61cc4eb02103d09dfdd397de948bd836293debeb2`;
- `conformance-claim-v2.schema.json`: `4c05a97a1aa9f7dafe629a406a853239928413e79e95488ac2b20ebd0c52a38c`.

## Release-gate and test replay

The five release gates were run as standalone processes with the retained
pinned Python environments and `PYTHONDONTWRITEBYTECODE=1`. No output pipeline,
alternate index, Git wrapper, or clean-status shim was used.

| Checkout | Version | Exit | Exact diagnostic |
| --- | --- | ---: | --- |
| rc.4 `q5oy3o` | `1.0.0-rc.4` | 1 | `release gate requires a clean candidate checkout` |
| rc.5 `2kp3tv` | `1.0.0-rc.5` | 1 | `release gate requires a clean candidate checkout` |
| later `2jaw7h` | `1.0.0-rc.5` | 1 | `release gate requires a clean candidate checkout` |
| clean authoritative main | `1.0.0-rc.4` | 1 | `README version is not 1.0.0-rc.4` |
| clean authoritative main | `1.0.0-rc.5` | 1 | `README version is not 1.0.0-rc.5` |

These expected-red results prove the exact blocker: no authoritative landed
commit contains either accepted candidate. They are not reported as passing
release evidence.

All internal candidate validation is green:

- rc.4: exit 0, 35 schemas, 189 vector files, 27 Python tests, Go tests;
- rc.5 `2kp3tv`: exit 0, 42 schemas, 422 vector files, 29 Python tests, Go tests;
- later `2jaw7h`: exit 0, 48 schemas, 592 vector files
  (422 released, 170 candidate), 169 Python tests, Go tests.

Candidate status/staging counts remained unchanged after validation.

## Alternatives and recommendation

The producer compared the compliant alternatives correctly:

1. Landing retained rc.4 bytes is technically possible only by creating a real
   current candidate, but the retained snapshot predates accepted
   execution-policy/cache-identity corrections. Publishing it would discard
   accepted work; rebuilding it as current requires a new composite, security
   review, and downstream retesting.
2. Landing the exact accepted `2kp3tv` rc.5 snapshot is recommended. The local
   compatibility policy permits a prerelease candidate to tighten or correct
   behavior before 1.0, states that neither rc.4 nor rc.5 was released or
   pinned, preserves schema-6 wire bytes, and requires candidate consumers to
   use an explicit root/digest without advancing committed release pins.
3. Tagging `2jaw7h` as rc.5 is noncompliant because it packages deferred
   next-version content whose mint/pin owner is `TASK-260728-251p01`.
4. Weakening the release gate, using a disposable commit or mutable branch, or
   guessing a future SHA violates the no-fabrication boundary and still does
   not create an immutable released ref.

No release-policy change is required. The one task-contract decision is for
the board/product owner to approve rc.5 as superseding literal rc.4 wording in
`TASK-260720-3ag6pi` and downstream rc.4-named briefs.

## Dependency and pin-order verification

An independent reverse walk over the live board found exactly 47 downstream
tasks from `TASK-260720-3ag6pi`, all currently `backlog`. The two direct
dependents are:

- `TASK-260720-12r55p` (`shared-v6-vector-consumer`), which also waits for
  `TASK-260720-th0jdi`;
- `TASK-260720-3pvihp` (`qualify-manager-release-evidence`), which also has five
  other blockers.

Clearing the publication gate removes one hard edge from each; it does not
override their remaining blockers. The producer packet lists all 47 tasks by
shortest dependency distance, and the independently computed distance groups
and count matched exactly.

The verified critical unblock/pin order is:

1. board approval of rc.5 supersession;
2. authorized landing of exact `2kp3tv` bytes and clean unwrapped validation,
   regeneration, and rc.5 release gates;
3. accept `TASK-260720-3ag6pi`;
4. candidate-root CocoaSkills/Curator consumers and manager integration, with
   committed suite pins held at the prior public release;
5. real Curator and CocoaSkills releases, then
   `TASK-260720-3pvihp`;
6. `TASK-260720-vs6den` promotes curator-spec implementation pins;
7. authorized signed/public `v1.0.0-rc.5`;
8. `TASK-260720-25d05o` qualifies the public protocol release;
9. `TASK-260720-38l1sy` and `TASK-260720-1utsx8` audit the two manager
   released-suite pins;
10. `TASK-260720-22ynoi` performs final cross-manager Go interoperability
    acceptance.

This matches the task ownership boundaries: `vs6den` owns implementation refs,
`25d05o` owns public protocol-release qualification, `38l1sy` and `1utsx8` own
the manager released-suite pins, and `251p01` owns the later version mint.

## Exact human/external inputs

1. Board/product authority must approve rc.5 substitution for the literal
   rc.4 task contract.
2. A curator-spec integrator must authorize and merge the exact accepted
   `2kp3tv` snapshot to the protected default branch.
3. A curator-spec release maintainer must sign the qualified landed commit and
   authorize publication of `v1.0.0-rc.5`.
4. Curator and CocoaSkills release owners must later publish real manager
   releases before manager-release qualification.

These are real future authorization boundaries for executing the runbook, not
blockers to accepting this completed read-only decision task.

## Sources

- Live `curator-spec` `git ls-remote`, `show-ref`, worktree status, SHA-256,
  normalized-tree, validation, and release-gate commands replayed on
  2026-07-29.
- Accepted `curator-spec` `COMPATIBILITY.md`, especially the prerelease rule,
  rc.4/rc.5 compatibility section, and candidate-consumption boundary.
- Accepted `curator-spec` `GOVERNANCE.md` release process and immutable signed
  target requirements.
- Accepted `curator-spec` `release/1.0.0-rc.5.json` and
  `.github/workflows/release.yml`.
- Board outcomes `TASK-260728-2kp3tv_review-verdict.md`,
  `TASK-260728-2jaw7h_review-verdict-cycle-2.md`, and
  `TASK-260720-3ag6pi_reviewer-verdict.md`.
- Live task-specific board projections and reverse dependency walk.
- Producer outcome
  `TASK-260729-1kq1rd_protocol-publication-gate-for-parity.md`.

