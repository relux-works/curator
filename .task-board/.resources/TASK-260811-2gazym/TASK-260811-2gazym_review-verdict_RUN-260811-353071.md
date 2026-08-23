# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

## Goal and reviewed scope

- Reviewer run: `RUN-260811-353071`
- Authoritative run goal at the final review checkpoint: `GOAL-260811-cfb057` revision 1
- Resolved scope: `TASK-260811-2gazym`
- Reviewed task state: `reviewing`
- Reviewed HEAD: `8ff4a238f7725bada3cfb8aa7c9c135698483caa`
- Reviewed artifact-policy Go-source aggregate: `sha256:9b74167c600559008d47565af78c93727f9094c61eda6c5b222281907c268af3`
- Reviewed conformance corpus: `sha256:49f796cd44fa89428b4b572d390bb49765119b12fedb428f7ae87ee45e61dd24`
- Reviewed producer evidence: `TASK-260811-2gazym_implementation-evidence.md`, `sha256:4e515a4f8187ad0a4c09741b01c93480cdde743f68485626c46392ddcd557444`

This artifact records only the changes-requested branch. The findings are
ordinary, autonomously repairable implementation and conformance work. They do
not require a human-only product/architecture decision and therefore do not
meet the board's Stop-The-Line boundary.

## Acceptance blockers

### R1 — `canonical_tree` manifests do not bind their nodes back to the raw tree identity

`canonicalTreeIdentity` derives the captured tree identity from every entry's
relative path, kind, executable bit, size, and SHA-256 under the
`curator-artifact-tree-v1` domain (`internal/artifactpolicy/tree.go:351-376`).
The decoder checks root size/hash equality only when
`RawPayload.Kind == "file"` (`internal/artifactpolicy/codec.go:434-445`). It
never rederives the domain-separated identity for `canonical_tree`.

Consequently, a self-rehashed admitted tree manifest can replace
`RawPayload.SHA256`, `Origin.ChecksumSHA256`, and the matching
`origin_checksum_sha256` role fact with another syntactically valid digest
while leaving every manifested node unchanged. The existing role checks see
three mutually equal strings and accept them; the encoded manifest no longer
proves the exact captured tree represented by its nodes. This violates the
exact raw-payload/origin binding required by `artifact-manifest-v1` and defeats
the semantic-forgery protection added during prior rework.

The same decoder also never requires nonempty `AdapterID`, `Manager`,
`PackageName`, or `PackageVersion`; runtime descriptor admission does require
all four (`internal/artifactpolicy/policy.go:386-389`). The normative manifest
requires adapter/profile, manager, and package identity.

Required rework:

1. Recompute and compare the canonical-tree identity and total byte size from
   the canonical nodes using exactly the capture identity rules.
2. Validate all required descriptor identity fields during decode.
3. Add self-consistent rehash tests for tree-root digest/origin drift,
   executable-bit drift, entry kind/path/hash drift, and empty descriptor
   identity fields.

### R2 — PAX/GNU tar metadata members bypass entry accounting and manifest evidence

`walkTar` calls `archive/tar.Reader.Next()` and only then increments
`max_entry_count` (`internal/artifactpolicy/containers.go:290-316`). Go's
reader explicitly consumes PAX extended headers and GNU long-name/long-link
headers internally and continues without returning them (`archive/tar/reader.go`
in Go 1.25.5, lines 69-131). Their data is parsed before the artifact service
sees a header or charges an entry.

This means physical PAX/GNU metadata files are neither counted nor manifested.
A tar can make the reader consume more than 100,000 such header files while the
artifact accounting records only a later logical member. Normal PAX metadata
is also reduced to `format`, `mode`, `size`, and `typeflag`; the required
presence evidence for timestamps, ownership, xattrs, and other metadata is not
recorded. This contradicts policy-v1's PAX/GNU support, actual-streamed-count
rule, 100,000-entry limit, and complete archive metadata evidence.

Required rework:

1. Enumerate and bound physical tar headers before any library layer can hide
   or merge them, or use an equivalently complete raw tar parser.
2. Count every PAX/GNU metadata header and payload, retain normalized metadata
   presence, and reject unsupported metadata that changes resolution or
   execution.
3. Add accepted ordinary PAX/GNU-long-name vectors plus negative 100,001
   metadata-header, malformed metadata, repeated metadata, xattr, and
   path/link-resolution vectors.

### R3 — no production caller can obtain a positive toolchain or output authorization

The exported `ToolchainRequest` and `LocalOutputRequest` require sealed
interfaces whose methods and return records are package-private
(`internal/artifactpolicy/types.go:325-339`). Both record types, both sealed
implementations, the manager seal, and both issuer functions are private
(`types.go:603-665`, `authorization.go:14-33`). A repository-wide production
search finds no issuer call site; only same-package tests call the issuers.

The non-mintability goal is correct, but the resulting production API has no
manager-owned path that can ever produce `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT`.
The published conformance corpus is therefore not executable end-to-end by an
adapter/executor consumer: its A07/A08 harness works only because
`conformance_test.go` uses package-private test helpers. As shipped,
`AuthorizeAdapterExecution` and `AuthorizeCachePublication` can verify positive
tokens that no production subsystem can create.

Required rework:

1. Add a usable manager-owned issuer/verifier integration boundary without
   exporting caller-populatable trust facts or allowing adapters to mint trust.
2. Exercise A07/A08 and the pre-execution/publication gates from an external
   test package using production APIs and real checkpoint/receipt verifier
   outputs, not same-package constructors.

### R4 — the 182-case corpus is self-consistent but does not contain all accepted exact byte vectors

The focused conformance test passes and pins 182 manifest digests, but at least
two normative byte families are substitutes for the accepted vectors:

- C01a-C01c require pinned GNU `-fPIE -pie`, `-fPIE -static-pie`, and
  `-fPIC -shared -Wl,-soname,...` outputs. The corpus instead synthesizes a
  minimal ELF byte array in `ELF64` (`conformance/corpus.go:747-825`); there are
  no pinned compiler outputs or toolchain identity in the package.
- C07 requires a valid JVM class. `JVMClass()` declares a constant-pool count
  of 1 and then uses `this_class == 0` (`conformance/corpus.go:899-904`). A
  local JDK 26 `ClassLoader.defineClass` probe rejects those exact 24 bytes with
  `java.lang.ClassFormatError: Invalid this class index 0 in constant pool`.
  The detector accepts them because it skips the six access/this/super bytes
  without validating their constant-pool references
  (`internal/artifactpolicy/native.go:664-726`).

The shared `0xCAFEBABE` magic also enters both Mach-O and JVM detectors
(`native.go:423-465`, `644-661`). The Mach-O candidate error then marks a JVM
match incomplete (`detect.go:118-167`), a boundary the invalid negative-only
fixture does not expose for role-positive handling.

Required rework:

1. Publish the actual pinned GNU byte fixtures and their immutable generation
   provenance/digests for C01a-C01c.
2. Replace C07 with a genuinely valid pinned JVM class (direct and nested), and
   validate class constant-pool references and required structural indexes.
3. Resolve the shared Mach-O/JVM magic without allowing one invalid candidate
   interpretation to make a structurally valid alternative perpetually
   incomplete; retain deny-dominant behavior for dependency input.
4. Keep 182 (or a larger explicitly mapped set) only after each case is shown
   to be the accepted semantic byte vector, not merely a labeled digest.

### R5 — duplicate `link_or_load` evidence can bypass the intended ELF ambiguity branch

Both runtime ELF resolution and manifest semantic derivation select
`elf.shared_object` when `DT_SONAME` is present using
`hasSoname || linkEdges == 1` before checking `linkEdges > 1`
(`internal/artifactpolicy/native.go:243-257` and
`semantics.go:427-438`). Thus an ET_DYN node with `DT_SONAME` and two or more
resolved link/load edges is classified as a dynamic library; the existing
`duplicate_use_edges` branch is unreachable for that combination. The accepted
rule permits one resolved link/load edge and sends unresolved/conflicting facts
to the compiled ambiguity class.

Required rework: apply duplicate/conflict checks before the shared-object rule
in both runtime and decoder semantics, and add exact runtime plus self-rehashed
codec tests for duplicate execute/link evidence with and without SONAME.

## Independent validation

The implementation is mechanically healthy but the green tests do not cover
the blockers above:

| Command | Result |
| --- | --- |
| `go test -short -count=1 ./internal/artifactpolicy/...` | pass; artifactpolicy 9.031s |
| `go test -short -race -count=1 ./internal/artifactpolicy/...` | pass; artifactpolicy 82.847s |
| focused 182-case conformance command | pass; artifactpolicy 1.036s |
| `go vet ./internal/artifactpolicy/...` | pass |
| `go build ./internal/artifactpolicy/...` | pass |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` | pass; `0 issues.` |
| `gofmt -l internal/artifactpolicy` | pass; no files listed |
| `git diff --check` | pass |
| `task-board validate` | pass; board valid |
| JDK 26 `ClassLoader.defineClass` against the exact `JVMClass()` bytes | rejects with `ClassFormatError: Invalid this class index 0 in constant pool` |

Per the active directive, this reviewer did not duplicate a repository-wide
lane. The producer's durable current-source evidence records green full
`go test -count=1 ./...`, full vet/build/lint, and stable pre/post source
fingerprints. Those results remain useful regression evidence but cannot
satisfy the uncovered acceptance semantics above. An exploratory `staticcheck`
invocation under the active goenv was unavailable (exit 127); the repository's
pinned authoritative `golangci-lint` command passed.

## Preserved properties

- Dependency compiled bytes remain fail-closed in the reviewed paths.
- `verified-binary-v1` remains unavailable.
- Kotlin remains excluded.
- The resolved F14 option-1 amendment is implemented: physical payloads keep
  distinct exact identities while the logical projection is order-independent.
- No product code was modified by this reviewer.

## Routing

Route to `to-dev`. A new producer should repair R1-R5, refresh the task-scoped
implementation evidence and exact corpus digests, rerun current focused/race/
regression/full vet/build/lint gates, and hand the task to another independent
review cycle.
