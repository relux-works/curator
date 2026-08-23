# TASK-260728-1a52au core-contract outcome

## Binding input

- Architecture: `TASK-260720-1nvomm_external-build-repositories-architecture-v6.md`
- Verified SHA-256:
  `2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`
- Independent review: accepted, with no remaining blocking or rework finding.

## Documentation changed

- `decisions/0005-external-build-repositories.md` — new schema-7/rc.5
  decision, compatibility freeze, ownership boundaries, stable failure classes,
  rejected alternatives, and security impact.
- `protocol/core.md` — schema-7 versioning, first-class
  `build_repositories`, `curator-build.json`, `go-repository-v1`, exact source
  lock and tag assertion, declared/effective identity, substitution rules,
  raw-object/audit contract, receipt v2, marker v3, protected currentness,
  manager-derived publication, credential/signing ownership, and future closed
  drivers.
- `SECURITY.md` — external repository adversary/TCB boundary, immutable
  acquisition, raw-object proof, audit-before-cache/compiler, descriptor/output
  isolation, offline/status behavior, credential/signing ownership, and closed
  future-driver admission.

The accepted rc.4 baseline files `profiles/manager.md` and
`decisions/0004-compile-only-build-drivers.md` were seeded byte-for-byte and
were not changed by this task.

## Architecture-v6 ownership map

| Architecture-v6 section | Resulting owned spec text | Exclusion or downstream owner |
|---|---|---|
| 1. Decision summary | Decision 0005 “Version boundary”, “Manifest and descriptor ownership”, “Raw snapshot and audit boundary”; Core 1, 4.2.1, 6.5, 9.2, 10, 12; Security “External build repository boundary” | Wire encoding remains TASK-260728-17sclp |
| 1.1 Fifth-review exact-tag correction | Decision 0005 “Declared and effective source”; Core 6.3; Security “Immutable acquisition and source proof” | Six-case executable fetch vectors remain TASK-260728-3b8qym |
| 1.2 Gap/self-verification | Decision 0005 “Context” and “Decision” | Architecture review history is evidence, not normative spec text |
| 2. Manifest schema 7 and compatibility | Core 1, 4, 4.2, 4.2.1; Decision 0005 “Version boundary” | JSON Schemas/generated cases/claim v3 remain TASK-260728-17sclp |
| 3.1 Canonical network identity | Core 6.3; Decision 0005 “Declared and effective source” | Exact manager URL/transport rendering remains TASK-260728-wy3dsw |
| 3.2 Safe tag/development-ref grammar | Core 6.3; Core 5 | Schema regex/negative cases remain TASK-260728-17sclp |
| 3.3 Lock/optional tag assertion | Core 6.3; Decision 0005 “Declared and effective source”; Security “Immutable acquisition” | Exact fetch argv remains TASK-260728-wy3dsw |
| 4. Descriptor/output boundary | Core 4.2.1; Core 12.1; Decision 0005 “Manifest and descriptor ownership”; Security “Descriptor, output, and audit isolation” | Descriptor JSON Schema remains TASK-260728-17sclp |
| 5.1 Declared/effective source states | Core 6.4, 9.2, 10; Decision 0005 “Declared and effective source” | Receipt/marker JSON branches remain TASK-260728-17sclp |
| 5.2 Skillfile.dev schema 2 | Core 5 and 6.4; Decision 0005 “Declared and effective source” | Skillfile.dev-v2 JSON Schema/cases remain TASK-260728-17sclp |
| 6.1 Trusted Git/process/environment | Core 6.5; Security “Immutable acquisition and source proof”; Decision 0005 “Raw snapshot and audit boundary” | Exact process/env/child allowlist belongs to TASK-260728-wy3dsw |
| 6.2 Private repository initialization | Core 6.5 requires manager-owned private state and fail-closed raw proof | Exact init argv and repository postconditions belong to TASK-260728-wy3dsw |
| 6.3 Network fetch/ref flows | Core 6.3 fixes tagged versus untagged sole acquisition and typed results; Security “Immutable acquisition” | Full fetch argv/ordering/CLI diagnostics belong to TASK-260728-wy3dsw; vectors to TASK-260728-3b8qym |
| 6.4 SSH wrapper/OpenSSH | Core 6.5 and 12.2 fix operator-owned transport/credentials and prohibit repository selection | Exact wrapper, policy FD, OpenSSH argv, and authentication tails belong to TASK-260728-wy3dsw |
| 6.5 Local substitution admission | Core 5, 6.4, 6.5; Decision 0005 stable local failure classes | Byte-level config/ref/parser lifecycle belongs to TASK-260728-wy3dsw |
| 6.5.1 Config admission | Core 6.5 requires bounded data parsing and fail-closed unsupported formats | Exact grammar belongs to TASK-260728-wy3dsw; cases to TASK-260728-3b8qym |
| 6.5.2 Files refs/HEAD | Core 5 fixes committed HEAD; Core 6.5 fixes ordinary files-ref admission | Exact loose/packed ref grammar belongs to TASK-260728-wy3dsw |
| 6.5.3 Pack admission | Core 6.5 fixes pack 2/3 plus index 2 and typed unsupported-object-format failure | Exact byte/checksum tables belong to TASK-260728-wy3dsw; vectors to TASK-260728-3b8qym |
| 6.6 Common raw-object proof | Core 6.5; Security “Immutable acquisition”; Decision 0005 “Raw snapshot and audit boundary” | Manager process implementation contract belongs to TASK-260728-wy3dsw |
| 6.6.1 cat-file batch | Core 6.5 requires one bounded manager-owned raw-object reader and prohibits transformations/lazy reads | Exact argv/framing/limits belong to TASK-260728-wy3dsw |
| 6.6.2 Object semantics/graph | Core 6.5 requires ID recomputation, common commit/tag parsing, full graph proof, and stable semantic/incomplete failures | Exact byte grammar and cross-language fixtures belong to TASK-260728-wy3dsw and TASK-260728-3b8qym |
| 6.6.3 Git LFS rejection | Core 6.5 pins `git-lfs-pointer-parser-v3.7.1` and stable LFS failure; Security prohibits hydration | Exact parser-family algorithm/fixtures belong to TASK-260728-wy3dsw and TASK-260728-3b8qym |
| 7. Snapshot identity/audit equivalence | Core 6.5, 8.1, 9.3, and 9.4; Security “Descriptor, output, and audit isolation”; Decision 0005 “Raw snapshot and audit boundary” | Registry/profile CLI presentation remains TASK-260728-wy3dsw |
| 8.1 Unsubstituted receipt v2 | Core 9.2; Core 6.3/6.4; Decision 0005 “Receipt, marker, status, and lifecycle” | Receipt-v2 schema/canonical cases remain TASK-260728-17sclp |
| 8.2 Substituted receipt v2 | Core 6.4 and 9.2 | Receipt-v2 schema/cases remain TASK-260728-17sclp |
| 8.3 Mixed commands | Core 4.2, 9.1/9.2, 10; Decision 0005 compatibility/lifecycle | Marker-v3 conditional branches remain TASK-260728-17sclp |
| 9. Marker/status/repair/dedup/GC | Core 9.4, 10, and 12.1; Decision 0005 lifecycle | Exact manager status/repair/dedup/GC state machine and diagnostics belong to TASK-260728-wy3dsw; marker schema to TASK-260728-17sclp |
| 9.1 Snapshot/artifact keys | Core 9.2, 9.3, and 9.4 | Physical paths remain intentionally implementation-specific |
| 9.2 Read-only status | Core 10; Security “Offline, status, credential, and signing ownership” | CLI output contract belongs to TASK-260728-wy3dsw |
| 9.3 Repair | Core 10 exact-tag-only repair and no-adoption rule | Transaction details belong to TASK-260728-wy3dsw |
| 9.4 GC roots | Core 9.4 portable live roots, conservative retention, and no-execution/no-adoption safety | Exact manager GC traversal, storage paths, and diagnostics belong to TASK-260728-wy3dsw |
| 10. Deterministic build/install/PATH/rollback | Core 4.2.1 and 12.1; Security descriptor/output isolation; Decision 0005 lifecycle | Exact lifecycle state machine belongs to TASK-260728-wy3dsw |
| 11. Signing/notarization | Core 12.2; Security “Offline, status, credential, and signing ownership”; Decision 0005 credential/signing ownership | Future signer and release-pipeline implementation are out of rc.5 scope |
| 12. Future closed drivers | Core 12.3; Security “Closed future-driver admission”; Decision 0005 future-driver ownership | No future driver admitted by this task |
| 13. Threat model | Security “External build repository boundary”; Decision 0005 “Security impact” | Platform-native enforcement evidence remains conformance/implementation work |
| 14. Board impact | Core 1 and Decision 0005 freeze rc.4; this outcome records task ownership | Curator/cocoaskills/interoperability implementation tasks are explicitly out of curator-spec core scope |
| 15. Rejected alternatives | Decision 0005 “Rejected alternatives”; Core and Security MUST NOT rules | None |
| 16. Primary-source register | Binding architecture digest retained in Decision 0005 and this outcome | Source bibliography is architecture evidence; exact operational citations may be carried into manager-profile/conformance docs |
| 17. Validation/review checks | This outcome’s validation evidence and ownership map | Schema/example/vector gates belong to TASK-260728-17sclp and TASK-260728-3b8qym |
| 17.1 Producer evidence | Binding digest verification and gates below | Historical architecture smoke details are evidence, not copied into normative prose |

## Recorded exclusions

- No JSON Schema, generated case, vector, release metadata, CLI, manager
  profile, Curator, or cocoaskills file was changed.
- Exact Git init/fetch/SSH/cat-file argv, byte-level local Git parsers, exact
  status/repair/GC traversal and CLI mechanics, physical storage paths, and
  stable diagnostic rendering are reserved for `TASK-260728-wy3dsw`. The
  portable snapshot key, complete-key deduplication rule, GC roots, conservative
  retention, and no-execution/no-adoption safety are owned by Core 9.4.
- Schema-7, descriptor-v1, Skillfile.dev-v2, receipt-v2, marker-v3, and claim-v3
  wire enforcement is reserved for `TASK-260728-17sclp`.
- Cross-language/adversarial vectors and rc.5 release qualification are
  reserved for `TASK-260728-3b8qym`.
- The normative prose intentionally names those future wire versions before
  their schema files land; this task is their accepted contract prerequisite,
  not an assertion that the downstream artifacts already exist.

## Validation evidence

All commands ran directly without `tee` or a status-hiding pipeline.

| Command | Exit | Evidence |
|---|---:|---|
| `shasum -a 256 <architecture-v6> <review-v6>` | 0 | Architecture matched required `2abae77d...67e`; review file was readable and accepted that digest |
| task-local venv install from `requirements-dev.txt` | 0 | Installed pinned `jsonschema==4.25.1` |
| `<task-venv>/bin/python tools/validate.py` | 0 | `validated 30 schemas and 93 vector files` |
| `PATH=<task-venv>/bin:$PATH make validate` | 0 | 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...` passed |
| `git diff --check` | 0 | No whitespace errors |

Resulting owned-file SHA-256 values:

- `SECURITY.md`:
  `3b233a2af5fc1cac33f9af75079aeede7df3c37f0b94a91e8352c6df425483a7`
- `protocol/core.md`:
  `e35f9a076fb7ad21b859e04b0ba88a8ae7bdbc544b3799db751dd6f6a0ea9384`
- `decisions/0005-external-build-repositories.md`:
  `fa9ff8119350652052b29b462d5dab71af5dbd9201a9c23d25065605b72623fa`

## Rework 1 addendum

The focused rework after `TASK-260728-1a52au_core-contract-review.md`:

- makes a network-substitution `revision` a full lowercase object ID for the
  effective repository object format while keeping declared state immutable;
- requires independent allowlist, revocation, registry, tag-lock, and
  audit-policy gates for every applicable external subject before
  artifact-cache lookup or compiler work; and
- adds Core 9.4 for the complete protected external snapshot key,
  complete-key-only deduplication, subject-specific audit decisions, portable
  GC roots, conservative retention, and no-execution/no-adoption GC safety.

The architecture-v6 ownership map and recorded exclusions above were corrected
to assign those portable section 9.1 and 9.4 boundaries to Core 9.4 while
leaving only exact manager traversal, paths, CLI mechanics, and diagnostics to
`TASK-260728-wy3dsw`.

Rework validation commands ran directly:

| Command | Exit | Evidence |
|---|---:|---|
| `<task-venv>/bin/python tools/validate.py` | 0 | `validated 30 schemas and 93 vector files` |
| `env PATH=<task-venv>:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin make validate` | 2 | Python/schema checks passed, but the deliberately narrowed harness path hid the installed `go` binary; this attempt is failing evidence and does not satisfy the full gate |
| `env PATH=<task-venv>:$PATH make validate` | 0 | 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...` passed |
| `git diff --check` | 0 | No whitespace errors |
| seed `cmp -s` checks for `profiles/manager.md` and Decision 0004 | 0 each | Accepted rc.4 seed files remain byte-identical to the source checkout |
| `git diff --cached --quiet` | 0 | No staged changes |
