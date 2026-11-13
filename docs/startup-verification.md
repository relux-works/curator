# Startup verification and skill integrity

This contract covers unsigned configuration. Configuration signatures, admin
keys and remote grant-service wire formats are outside this interface.

## Policy

The optional machine carrier11 member is `"startup_verification":{"level":3}`.
Omission selects3 without rewriting existing configuration. The object is closed:
`level` is required and must be an integer token from1 through4. Null, strings,
booleans, fractions, exponents, duplicate keys and unknown fields refuse. System
policy can lock the entire `startup_verification` member.

| Level | Local decision | Centre decision |
| --- | --- | --- |
|1|Failed or unavailable verification warns and permits startup|No centre requirement|
|2|A definite mismatch refuses; unavailable evidence warns|No centre requirement|
|3 (default)|Every required local check must complete and pass|No centre requirement|
|4|The same local requirement as3|A denial or invalid answer refuses; absent/unreachable service falls back to3 with a warning|

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
After structural checks pass, levels3/4 may restore ordinary regular-file bytes
from the same recorded pin through the existing repair transaction. This does
not authorize replacing expected evidence, migrating a lock or changing a pin.

Manifest3 retains manifest2's field and skill-directory grammar and adds the
optional same `startup_verification` object. Each explicit context minimum must
be at least the machine level. The strongest selected root, overlay or transitive
context minimum applies. Lock6 binds that minimum, so editing a source manifest
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

Carrier11 optionally admits
`"extensions":{"grant_service":{"endpoint":"https://verification.example.test/"}}`.
This is location metadata only. The endpoint is an absolute HTTPS locator of at
most4096 UTF-8 bytes, with a valid DNS hostname or literal IP. Credentials,
queries (including a forced empty query), fragments, raw whitespace, controls,
backslashes, dot segments, invalid escapes and invalid/empty ports refuse.
Encoded routing bytes are preserved. Bracket IPv6 literals. An absent member is
omitted; null, an empty object and unknown members refuse.

Parsing or reading this reference performs no network or secret-store access.
A configured location without an implemented provider is unavailable, so level4
uses the documented local3 fallback and warning. This metadata does not enter
`curator.launcher-config/1` or change the launch fragment format.

## Complete skill evidence

Lock6 extends lock4 and does not enable lock5 skill-plan adoption. Its required
`skill_integrity` array contains one entry per selected skill, in lock member
order. An entry contains `manifest` and `manifest_sha256`. The latter is SHA-256
of the manifest's CCJ-1 bytes, spelled `sha256:` plus64 lowercase hex digits.

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
existing portable relative-path grammar, at most4096 Unicode codepoints, with
no duplicates or case collisions. Size is a nonnegative integer. Every regular
file is included, including reference assets and runtime scripts; there are no
content exclusions. Payload symlinks, hardlinks, special files and nested Git
metadata refuse. Adapter links into a verified store are a separate surface.

One shared capture budget covers the closure:16384 regular files,64MiB of file
bytes and30 seconds. Evidence JSON and lock6 JSON each have a64MiB limit.
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

## Migration and diagnostics

Older machine, manifest and lock carriers remain readable. Reading and startup
do not rewrite them. A native-only legacy profile can satisfy default3 using
its existing complete pinned Git tree or admitted full state-hash2 snapshot.
No new cap or executable-metadata claim is retroactively added to that legacy
algorithm. Existing version1 NUL refusals remain mandatory.

A legacy manifestless skill has no recorded synthesized evidence. It is
unavailable:1/2 warn,3/4 refuse. Explicit `curator profile update <profile>` can
create lock6 evidence transactionally. Preserved-pin migration must retain the
exact original pins; a resolving pin change is an ordinary update. Failure
preserves the old selection and lock. Startup never silently migrates a profile.

| Diagnostic | Meaning |
| --- | --- |
|`startup_verification_invalid`|Malformed policy or unsupported outcome|
|`startup_verification_weakening`|A profile minimum is below the machine baseline|
|`startup_verification_failed`|A completed local check contradicts its expectation|
|`startup_verification_unavailable`|A required local check did not complete|
|`verification_centre_unavailable`|Level4 falls back to local3, with warning|
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
