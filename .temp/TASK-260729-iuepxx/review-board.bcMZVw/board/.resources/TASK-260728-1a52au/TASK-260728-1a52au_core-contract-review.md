# TASK-260728-1a52au independent review

Verdict: **CHANGES REQUESTED**. Route to `to-dev` for focused normative-document rework, followed by a new reviewer cycle.

## Review basis

- Binding architecture-v6 SHA-256 independently verified as `2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`.
- Accepted review-v6 inspected and treated as binding.
- Producer worktree HEAD matches the curator-spec source HEAD `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Seeded `profiles/manager.md` and Decision 0004 are byte-identical to the accepted rc.4 source checkout.
- Worktree index is clean. The only worktree files are the accepted rc.4 seeds plus `SECURITY.md`, `protocol/core.md`, and new Decision 0005. No commit was created.

## Required changes

### 1. Network-substitution revision uses the wrong object format

Binding architecture-v6 section 3.2 says a structured `revision` is a full lowercase object ID for the **effective repository object format**. Core section 6.3 currently says it is for the **declared object format** at `protocol/core.md:528`.

This is a normative contradiction. A network substitution can have an actual effective object format distinct from the unchanged declaration, and sections 5/6.4 expressly retain actual object format in effective state. The current rule can impose the wrong OID width or reject an otherwise valid effective substitution.

Required resolution: make the structured revision depend on the effective repository object format and ensure related schema/profile ownership language does not collapse effective state into declared state.

### 2. Accepted snapshot-key, deduplication, and GC boundaries are absent from owned normative text

Architecture-v6 section 9.1 fixes the protected snapshot key as effective identity kind/value, effective object format, full effective commit, and external build-source digest; it permits snapshot deduplication only on complete key equality and keeps audit decisions subject-specific. Section 9.4 fixes GC roots, journal roots, unchanged marker-v1/v2 behavior, non-root receipt content, conservative retention for unreadable/unprovable state, and no source/artifact execution or adoption.

The outcome map claims section 9.1 is owned by Core 9.2/9.3/10 and section 9.4 by Core 10/12. The owned documents contain no current-driver snapshot-key/deduplication rule and no current-driver GC rule. The only GC mention in Core 12.3 is a requirement for future drivers. Deferring exact manager CLI/state-machine details is valid; omitting the accepted portable logical boundaries is not.

Required resolution: add MUST/MUST NOT normative text for the section 9.1 logical snapshot/dedup rules and the section 9.4 high-level GC roots and safety behavior, while leaving physical paths and exact manager mechanics downstream. Correct the ownership map to point to the resulting text and keep only genuinely downstream details in exclusions.

### 3. External audit policy gates are not explicitly required

Architecture-v6 section 7 step 6 requires allowlist, revocation, registry, tag-lock, and audit-policy gates for each external subject before artifact-cache lookup or compiler execution. Core 6.5 binds the audit subject and ordering but does not explicitly require those gates.

Required resolution: state that the applicable gates MUST be applied independently to the external subject before cache/compiler work, without duplicating manager-profile rendering details.

## Passing independent evidence

- `tools/validate.py`: exit 0, 30 schemas and 93 vector files; this gate also validates local Markdown links.
- `make validate`: exit 0, 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...`.
- `git diff --check`: exit 0.
- Producer-owned file hashes match the outcome resource exactly.

The validation gates are green, but they do not exercise the new prose semantics. Findings 1 through 3 prevent acceptance against the binding architecture and the task requirement that every accepted boundary be normative and internally consistent.