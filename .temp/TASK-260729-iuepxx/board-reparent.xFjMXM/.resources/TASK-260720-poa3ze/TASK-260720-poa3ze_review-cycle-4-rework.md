# TASK-260720-poa3ze review-cycle-4 rework evidence

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Status:** Ready for review  
**Primary outcome SHA-256:** `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`

## Key takeaways

- The recommendation now uses one coherent v1 cache trust model: reusable build entries live in manager-created, owner-protected machine state that is part of the TCB. Receipt and marker hashes are deterministic corruption/currentness metadata, not signatures, MACs, or independent provenance.
- If ownership, permission/DACL, containment, file type, or link safety cannot be established, persistent reuse fails closed. Dry-run reports `would-rebuild-untrusted-cache`; a real operation rebuilds from the revalidated snapshot into newly protected state; marker currentness, repair, publication, rollback, and GC use the same rule.
- The exact reviewer counterexample is now a conformance vector: its input, key, 24-byte artifact, artifact digest, size, and receipt digest all agree, but it is rejected solely because the cache boundary is untrusted.
- Protocol `1.0.0-rc.4` adds `conformance-claim-v2` while freezing claim v1 as rc.3 evidence. The generator uses separate current-rc.4 and historical-v1-rc.3 constants so regeneration cannot silently rewrite old claims.
- Only the logical key, canonical receipt bytes, artifact-relative path, and validation semantics are portable. Manager-home paths, cache directory names, receipt filenames, quarantine/lock paths, and storage backends remain implementation-specific.

## Returned findings and resolution

| Reviewer finding | Resolution in the primary outcome |
|---|---|
| High: self-consistent untrusted receipt does not prove build provenance | Sections 2.1 and 5.3–5.6 now define the protected-state TCB, same-principal/admin scope boundary, fail-closed fallback, non-authenticating receipt semantics, exact forged-hit vector, marker currentness, and GC behavior. Section 6 applies the rule to dry-run, publication, commit revalidation, repair, and rollback. |
| Medium: protocol-versioned conformance claim migration missing | Section 8 adds the exact `conformance-claim-v2` JSON Schema and valid case, rc.3-v1/rc.4-v2 reader-writer transition, generator constant split, schema/semantic negative cases, manifest/release changes, and the exact affected-artifact inventory. |
| Medium: physical cache namespace conflicts with portability boundary | Section 5.3 makes the layout explicitly illustrative and implementation-specific. Section 8 and the recommendation preserve portable logical identity while retaining each manager's local layout. |

## Verification evidence

- All 24 fenced JSON documents parse. Independent Python recomputation confirms cache key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`, primary receipt `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`, forged artifact `sha256:a4f06a1304c926ed7f2326c8fd90cabc5c5bd2981e690a4351c852d91c079d88`, and forged receipt `sha256:9a23f5b77e6173b0f10e7ed43cd2b21aa3b99f3a34945ec432fbb31338a6186d`.
- Markdown fences are balanced, and the classification gate finds Go, Rust, Zig, Swift, C/C++, Java/Kotlin, .NET, Node/TypeScript, Deno, and Python.
- The `.research` source and primary board outcome are byte-identical at the SHA-256 above.
- All nine newly cited primary URLs for protected filesystem state, cache portability, and conformance-claim migration returned HTTP 200.
- Detached exact-ref worktrees are clean and still equal remote `main`:
  - curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`;
  - Curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`;
  - cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Exact-ref gates are green:
  - curator-spec: 30 schemas, 93 vector files, 8 Python unit tests, and Go tool tests (`spec-validate-05.log`);
  - Curator: `go test ./...` (`curator-go-test-04.log`);
  - cocoaskills: 488 passed, 18 skipped (`cocoaskills-pytest-04.log`).
- No tracked product or specification source file was modified. Changes are limited to the research mirror and task-board outcomes.
- `task-board validate` still reports the same 12 unrelated legacy EPIC-260712 dependency references and one unrelated orphan TASK-260713-7a9c1e resource. No issue belongs to `TASK-260720-poa3ze`.

## Reproducibility notes

- The first spec-gate attempt (`spec-validate-04.log`) stopped before validation because system Python lacked `jsonschema`. The successful run selected the already verified repository Python environment (`jsonschema 4.25.1`) and executed the unmodified Makefile gate.
- An initial document coverage assertion used two stale display labels for Java/Kotlin and .NET. The report already contained both required rows; correcting the checker labels produced a green document gate without changing those classifications.

## Outcomes handed off

- `.research/260720_compile-only-build-drivers.md`
- `TASK-260720-poa3ze_compile-only-build-drivers.md` (updated board outcome)
- `TASK-260720-poa3ze_review-cycle-4-rework.md` (this task-scoped evidence)

