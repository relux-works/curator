# TASK-260728-zb2s4z independent review cycle 2

## Verdict

**ACCEPTED.** Review findings R1 and R2 are closed. The amended rc.5 candidate
matches the task acceptance criteria and the curator-spec architecture. No
remaining product or specification finding requires rework.

## R1 — portable-versus-hardened boundary

The rework removes the contradictory implication that portable execution
provides kernel read-only presentation, total network denial, or an exact
executable allowlist.

- `protocol/core.md` sections 4.2 and 4.2.1, `profiles/manager.md` sections 2.2
  and 2.2.1, `SECURITY.md`, decisions 0004/0006, the author guide, CLI,
  compatibility text, changelog, release text, and the executable vector now
  consistently describe manager-selected mechanisms.
- `network: "none"` has one meaning: fixed offline Go module, proxy,
  checksum-database, and VCS configuration with no manager- or Go-initiated
  dependency/build networking. It explicitly does not claim kernel network
  containment.
- Frozen source/toolchain integrity is enforced by manager non-write rules and
  pre/post identity/currentness checks, not by claiming read-only presentation
  to descendants.
- The process graph is the fixed manager-selected four-node graph with
  per-program identity verification. Package bytes cannot select or add an
  executable, argv, environment value, path, flag, hook, plugin, generator, or
  build permit. The stronger kernel executable-path allowlist remains deferred.
- The failure boundary is identical in the normative and executable surfaces:
  an unavailable mandatory portable control rejects with
  `build_execution_control_unavailable` before worker launch and publishes
  nothing; an inventory control normatively unavailable for the platform, or
  any of the six deferred hardened guarantees, never rejects, warns, or blocks
  portable publication.

Exhaustive searches found no residual absolute “source is read-only to
children,” total-network-denial, or kernel executable-allowlist implication.
The remaining “start no program other than” wording is expressly scoped to
manager/worker selection and immediately distinguished from kernel allowlisting.

## R2 — native-control inventory and capability evidence

`rc5-native-control-inventory-v1` is an exhaustive, versioned authority over
exactly macOS and Windows and exactly five controls. Every control has a closed
per-platform `{availability, mechanism, unavailable_reason}` record. Inventory
membership, availability states, the one unavailable-reason vocabulary,
per-operation probe scope, and pre-worker-launch timing are pinned independently
by both Python and Go validators.

`capability-evidence-v1` is closed:

- record fields are exactly `record_version`, `execution_policy`, `platform`,
  and `controls`;
- entry fields are exactly `name`, `availability`, `status`, and `probed_at`;
- availability, status, timing, cardinality, exposure, exclusion, and
  consistency-rule vocabularies are closed;
- generated macOS and Windows examples are cross-checked against every
  per-platform inventory record;
- unknown, missing, duplicate, contradictory, wrong-version, wrong-policy, and
  deferred-guarantee evidence paths have stable errors and executable negative
  guards; and
- reporting remains result-only in dry-run plan, install, and status results,
  and is excluded from cache keys, receipts, markers, and claims. The portable
  execution-policy identity itself remains bound into all four identity
  surfaces.

The per-platform mechanisms are honest native primitives, not qualification
claims. Apple documents `RLIMIT_FSIZE` as a per-file limit inherited by created
processes, `killpg` as process-group signalling, and `setsid` as session/process
group creation. Microsoft documents Job Object active-process and job-wide
memory limits, kill-on-close termination, and explicit inherited-handle lists.
Those primitives support the inventory entries but do not imply any of the six
deferred hardened guarantees:

- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setrlimit.2.html
- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/killpg.2.html
- https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/setsid.2.html
- https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_limit_information
- https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects
- https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute

Candidate claim-v3 tuples remain empty. macOS and Windows remain pending
downstream native evidence, Linux remains excluded until
`TASK-260728-1skseh`, and `committed_release_pin_advanced` remains false.

## Identity, compatibility, and provenance

- Assigned and predecessor worktrees remain detached at
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, with zero commits after the pin
  and clean indexes.
- Accepted predecessor manifest:
  `sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8`.
- Accepted downstream candidate manifest:
  `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`.
- Release metadata SHA-256:
  `b163f445f206c17dc618cc10b3957ca2f6f1b28607288162c0cfc5de02d83ee6`.
- Host-execution vector SHA-256:
  `c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de`.
- The manifest has 422 unique, complete, non-self-referential entries and all
  independently recomputed file hashes match.
- Exact delta against the accepted predecessor is **98 modified, 13 added, and
  0 removed files**. The producer outcome's “15 added” count was stale after
  its documented removal of two scratch artifacts; this reviewer artifact
  records the corrected independently recomputed count. The 13 added files are
  ten execution-policy negative schema cases, the host-execution vector,
  decision 0006, and the portable policy author guide.
- Manifest schemas 1-7, receipt v1/v2, marker v1/v2/v3, claim v1/v2, and
  `curator-build-v1` promised protected surfaces compare byte-for-byte with the
  accepted predecessor. Under `schemas/v1`, only `README.md`,
  `common.schema.json`, and `conformance-claim-v3.schema.json` differ as
  intended.
- External acquisition/lifecycle, pack/index, and claim-v3 qualification
  artifacts compare byte-for-byte. Audit-before-cache/compiler ordering,
  offline errors, operator-owned signing, and empty candidate claims remain
  intact.

Independent CCJ-1 recomputation matched every policy-separation oracle:

- portable cache key:
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
- reserved hardened key:
  `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037`;
- pre-revision key:
  `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`;
- receipt-v2 cache key:
  `sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492`;
- full receipt-v2 hash:
  `sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea`.

Portable, reserved-hardened, and pre-revision identities are distinct. Receipt
and marker values agree, and frozen marker v2 remains transitively bound through
its cache key and receipt hash.

## Independent gates

Assigned worktree, read-only:

- `python -B tools/validate.py`: 42 schemas and 422 vector files, exit 0;
- Python unit suite: 22 tests, exit 0;
- focused dishonest-evidence mutation suite: exit 0;
- `go test ./tools/...`: exit 0;
- focused portable-policy and identity-binding Go tests: exit 0;
- `go vet ./tools/...`: exit 0;
- gofmt check: exit 0;
- `git diff --check`, clean index, and zero-commit checks: exit 0.

The system `python3` preflight lacked the pinned `jsonschema` dependency. The
Python gates were therefore rerun with the existing validation environment
containing the exact `requirements-dev.txt` dependency
`jsonschema==4.25.1`; they passed. This was an environment/tool-entrypoint
issue, not a candidate failure.

Disposable byte-identical clean probe:

- two consecutive `make regenerate-check` runs: exit 0;
- `make release-check VERSION=1.0.0-rc.5`: exit 0;
- Python compileall and Go generator build: exit 0;
- post-gate Git status: clean;
- post-gate recursive comparison with the assigned candidate: byte-identical.

The probe's local synthetic commit was used only to exercise the clean-checkout
release gate. No candidate or predecessor file was staged or committed, and no
ref, tag, release, downstream pin, platform claim, remote, or publication was
created.

## Reviewer boundary

No candidate, predecessor, schema, vector, tool, release, or product-code file
was modified during review.
