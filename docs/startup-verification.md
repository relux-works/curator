# Startup verification and skill integrity

This contract covers unsigned configuration. Configuration signatures, admin
keys and remote grant-service wire formats are outside this interface.

## Policy

The optional machine carrier 11 member is `"startup_verification":{"level":3}`.
Omission selects 3 without rewriting existing configuration. The object is closed:
`level` is required and must be an integer token from 1 through 4. Null, strings,
booleans, fractions, exponents, duplicate keys and unknown fields refuse. System
policy can lock the entire `startup_verification` member.

| Level | Local decision | Centre decision |
| --- | --- | --- |
|1|Failed or unavailable verification warns and permits startup|No centre requirement|
|2|A definite mismatch refuses; unavailable evidence warns|No centre requirement|
|3 (default)|Every required local check must complete and pass|No centre requirement|
|4|The same local requirement as 3|A denial or invalid answer refuses; absent/unreachable service falls back to 3 with a warning|

Every applicable check runs before aggregation. A later unavailable observation
cannot erase an earlier mismatch. Permitted degraded startup does not claim
verified success. Missing executable/input, invalid policy/lock structure,
unsafe paths, source admission, credential authority, permission policy and
runtime qualification remain mandatory at every level.

Structural and path failures in the immutable store refuse before any repair,
including dry-run repair. Diagnostics name the path and violated rule; the
violating entry remains in place as evidence and no child starts. A payload
link or special file cannot be silently replaced from a valid Git snapshot.
Context and MCP snapshot regular files must have exactly one link, including
links whose other name is outside the snapshot. Failure names `single_link`
under `environment_store_untrusted`; unavailable link metadata also refuses.
Unix uses native link metadata; Windows uses a no-follow attributes handle bound
to the observed file. Directory link counts, operator-source admission and
managed adapter links retain their separate contracts. Skill payload hardlinks
retain `skill_integrity_boundary_invalid`.
After structural checks pass, levels 3/4 may restore ordinary regular-file bytes
from the same recorded pin through the existing repair transaction. This does
not authorize replacing expected evidence, migrating a lock or changing a pin.

Manifest 3 retains manifest 2's field and skill-directory grammar and adds the
optional same `startup_verification` object. Each explicit context minimum must
be at least the machine level. The strongest selected root, overlay or transitive
context minimum applies. Lock 6 binds that minimum, so editing a source manifest
requires explicit profile recomposition. No per-launch environment or argv
switch can lower the baseline.

```sh
curator config startup-verification
curator config startup-verification --json
curator config startup-verification --level 3
```

The editor validates the full existing configuration, preserves unrelated fields
and file mode, and publishes one staged replacement after source and system
policy readback. It refuses linked or unsafe source files, locked policy and a
changed source. Cancellation before publication preserves the original file.
A durability/readback failure after publication reports that the edit applied
but remains unverified.

The safe editor uses the existing Unix owner/mode and no-follow admission.
Other platforms retain policy reads and refuse CLI file edits until equivalent
descriptor and ownership protections are qualified.

The closed read projection has schema `curator.startup-verification/1`, required
`level` and `origin` (`default`, `machine` or `system`), plus optional
`grant_service`. It contains no managed-home paths.

## Verification centre location

Carrier 11 optionally admits
`"extensions":{"grant_service":{"endpoint":"https://verification.example.test/"}}`.
This is location metadata only. The endpoint is an absolute HTTPS locator of at
most 4096 UTF-8 bytes, with a valid DNS hostname or literal IP. Credentials,
queries (including a forced empty query), fragments, raw whitespace, controls,
backslashes, dot segments, invalid escapes and invalid/empty ports refuse.
Encoded routing bytes are preserved. IPv6 literals must be bracketed. An absent member is
omitted; null, an empty object and unknown members refuse.

Parsing or reading this reference performs no network or secret-store access.
A configured location without an implemented provider is unavailable, so level 4
uses the documented local 3 fallback and warning. This metadata does not enter
`curator.launcher-config/1` or change the launch fragment format.

## Complete skill evidence

Lock 6 extends lock 4 and does not enable lock 5 skill-plan adoption. Its required
`skill_integrity` array contains one entry per selected skill, in lock member
order. An entry contains `manifest` and `manifest_sha256`. The latter is SHA-256
of the manifest's CCJ-1 bytes, spelled `sha256:` plus 64 lowercase hex digits.

Legacy root-addressed skills retain the existing declaration validation without
retroactive frontmatter requirements. Selected-directory skills retain their
existing stricter metadata validation. Neither path infers runtime authority
from prose or frontmatter.

Every manifest field is required:

| Fields | Meaning |
| --- | --- |
|`schema`|`curator.skill-integrity/1`|
|`name`, `source`, `commit`, `state_sha256`, `directory`|Exact existing lock-member identity; unused identity strings stay empty|
|`manifest_kind`|`native` or `synthesized`|
|`declaration_path`, `declaration_sha256`, `declaration`|Selected native file and exact-byte digest with null declaration, or empty path/digest with the fixed synthetic declaration|
|`hash_version`, `content_sha256`|Frozen content-v2 algorithm and complete selected-tree content hash|
|`files`|Complete bytewise-sorted regular-file roster|

Each file binds required `path`, `size`, `executable` and `sha256`. Paths use the
existing portable relative-path grammar, at most 4096 Unicode codepoints, with
no duplicates or case collisions. Size is a nonnegative integer. Every regular
file is included, including reference assets and runtime scripts; there are no
content exclusions. Payload symlinks, hardlinks, special files and nested Git
metadata refuse. Adapter links into a verified store are a separate surface.

One shared capture budget covers the closure:16384 regular files, 64 MiB of file
bytes and 30 seconds. Evidence JSON and lock 6 JSON each have a 64 MiB limit.
Structural budget violations refuse. Interrupted/unreadable checks are
unavailable. Captured file identity and metadata are rechecked before accepting
the observation.

Native precedence and dual-manifest agreement retain the existing parser:
`agent-skill.json`, `csk-skill.json`, then `agents/runtime.json`. Only genuine
absence of all three permits synthesis. The synthetic declaration is exactly:

```json
{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}
```

It grants no commands, runtime roots or capabilities. Synthesis records evidence
in the profile lock and never writes a manifest into the skill source. Startup
compares the frozen expected evidence; it never regenerates an expectation from
drifted installed content.

## Whole-entry availability

A wholly absent selected entry is an availability condition only after an exact
no-follow leaf-not-found result and safe existing ancestors. Every other
selected member is checked before repair. Missing intermediate parents,
dangling links, unknown metadata and a present entry with a missing recorded
payload remain structural refusals at all levels.

Strict repair at levels 3/4 may recreate an absent Git entry from the same local
recorded pin under the existing credential and mutation-lock guards. It checks
all selected structure before recovery, staging and publication; late unsafe
objects are preserved with a named refusal. Read-only and levels 1/2 never
recreate the entry. Strict absent-Git dry-run reports
`would-rebuild-untrusted-store` without writes; strict state/path repair without
a Git snapshot reports `environment_repair_failed` without writes. The low-level
missing-named-target validator retains its fatal behavior.

Recorded Git payload presence uses the same selected-tree route as pin hashing,
including the materialized directory prefix, exact local commit and isolated
Git environment. It fetches nothing. Unsafe types or portable-path collisions,
repository alternates/grafts, more than 16 MiB of path metadata or more than
16,384 paths refuse without a partial roster. Skill payload presence uses the
verified Git or recorded integrity roster. Existing package admission, including
the required `SKILL.md`, is unchanged.
Ordinary regular-file byte restoration remains permitted after structural
checks; missing recorded files cannot be converted to byte repair.

## Migration and diagnostics

Older machine, manifest and lock carriers remain readable. Reading and startup
do not rewrite them. A native-only legacy profile can satisfy default 3 using
its existing complete pinned Git tree or admitted full state-hash 2 snapshot.
No new cap or executable-metadata claim is retroactively added to that legacy
algorithm. Existing version 1 NUL refusals remain mandatory.

A legacy manifestless skill has no recorded synthesized evidence. It is
unavailable: 1/2 warn, 3/4 refuse. Explicit `curator profile update <profile>` can
create lock 6 evidence transactionally. Preserved-pin migration must retain the
exact original pins; a resolving pin change is an ordinary update. Failure
preserves the old selection and lock. Startup never silently migrates a profile.

| Diagnostic | Meaning |
| --- | --- |
|`startup_verification_invalid`|Malformed policy or unsupported outcome|
|`startup_verification_weakening`|A profile minimum is below the machine baseline|
|`startup_verification_failed`|A completed local check contradicts its expectation|
|`startup_verification_unavailable`|A required local check did not complete|
|`verification_centre_unavailable`|Level 4 falls back to local 3, with warning|
|`verification_centre_denied`, `verification_centre_invalid`|Centre refusal; never a fallback|
|`skill_integrity_evidence_missing`|No frozen expected skill evidence|
|`skill_integrity_manifest_invalid`, `skill_integrity_unsupported_version`|Malformed or unsupported supplied evidence|
|`environment_store_untrusted`|Existing immutable-store boundary refusal; path and violated rule are named, including during repair|
|`skill_integrity_boundary_invalid`|Forbidden skill payload/path|
|`skill_integrity_mismatch`, `skill_integrity_unavailable`|Typed skill observation supplied to the level policy|

Independent gates retain their existing diagnostics. The launcher forwards Core
stderr before executing the provider and refuses a nonzero Core resolution.

The structural-before-repair behavior is incompatible with six row expectations
in the released revision-1 store corpus, covering rebuild planning, replacement
of structural evidence, repair-path diagnostic wrapping and associated negative
controls. Those negative controls remain negative; their historical expected
diagnostics do not redefine the new refusal contract.
The released corpus remains unchanged and its replay must report those
incompatibilities. This implementation does not claim full revision-1 store
conformance. Revision-2 qualification requires its separately accepted, landed
specification pin and version-aware consumer; a draft proposal is insufficient.

Permissive startup preserves mismatching source evidence even when the launcher
requests repair. If provisioning would require a legacy identity/lock migration,
perform that migration explicitly first; startup does not silently rewrite the
lock. Composition also requires its observed source generation to remain stable
through the operation, independently of the selected verification level.

## Store verification inputs and runtime coverage

The test-only reader pins the landed curator-spec corpus. Its single pin and
roster block is in `internal/envprofile/store_verification_reader_test.go`:
`storeVerificationSpecCommit`, manifest/case digests and declared row counts.
The embedded closed case schema is copied byte-for-byte from that revision.
It validates all 185 rows (179 implementation, 6 checker controls), including
policy inputs, ordered structural failures, optional observation flags/paths,
and disjoint level-scoped migration successors. These represent 519 and 22
declared level slots; reading them does not execute those operations.

Canonical CI selects the sibling `conformance/environment-store-verification-v2`
from the checkout named by `SPEC_PIN`. The four existing `TestStoreBoundary*VectorsDrive*`
families execute the 38 predecessor successors and their level-scoped siblings,
at every declared level, using revision-2 expectations. Fixture shapes preserve
the predecessor's enclosing object, surface, source and home state. The special-file
row uses a real socket at its original payload path, exercising `regular_types`
on Windows as well as Unix; the historical replay keeps its FIFO platform bound.
The runtime roster is precisely the 38 `legacy-*` IDs named by the manifest's
migration table plus its 17 `level_successors` IDs: 55 rows / 152 level slots
(49 implementation rows / 130 slots; 6 checker rows / 22 slots). The four
families contain 31 resolve, 4 dry-run, 16 repair and 4 status rows respectively.
The remaining 130 implementation rows / 389 slots are metadata-validated only,
not runtime-qualified. Checker rows
in these families run real product behavior with their planted violation; this
is distinct from an exhaustive checker-mutant qualification campaign.

The frozen revision-1 runtime replay lives in `store-v1-compatibility`. It
retains the immutable corpus and historical expectations, and admits only the
exact 32 compatible and 6 incompatible row IDs with their intended assertion
witnesses. Missing, skipped, swapped-ID or different-head evidence refuses.
Canonical Test/Race verifies that same-head proof, then runs the four families
against revision 2; it no longer excludes those families from the inventory.
Other revision-1 families retain the unchanged `conformance/v1` root.

For a separate reader rehearsal, explicitly set the absolute
`CURATOR_STORE_VERIFICATION_ROOT` and its matching
`CURATOR_STORE_VERIFICATION_MANIFEST_SHA256` from the central pin block, then run
`go test -p=1 -parallel=1 -count=1 ./internal/envprofile -run '^TestStoreVerification'`.
Unset selectors are inert; partial, empty or mismatched selectors refuse with
`store_verification_corpus_unsupported`. Production never reads them. The
canonical gate refuses either selector, including an empty value.

Run the historical checker with
`go run ./tools/check-store-compatibility -input .temp/store-replay.jsonl`.
Keep complete streams and reports under `.temp/`. Its success proves migration
compatibility, not full revision-1 or revision-2 runtime conformance.

Derived managed-home drift remains separate from profile/store evidence.
Level 1 without `--repair` starts an existing renderable home, emits the
`environment_home_stale` warning and `--repair` hint, and retains stale reasons.
Levels 2–4 refuse with leading `environment_home_stale`; capture remains strict.
Dry-run preserves drift. Explicit repair restores verified home bytes at every
level when source/store checks pass. A warning never marks drifted output current.
Raising the machine above an explicit profile minimum retains
`startup_verification_weakening`: edit the selected context manifests' minimum,
then run `curator profile update`.
