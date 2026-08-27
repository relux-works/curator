# TASK-260728-1a52au core-contract rework 1

## Binding review

This rework consumes
`TASK-260728-1a52au_core-contract-review.md` against architecture-v6 SHA-256
`2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`.
It changes only task-owned normative documentation and outcome evidence.

## Findings closed

1. **Effective object-format revision width.** Core 6.3 now requires an
   operator network-substitution `revision` to be a full lowercase object ID
   for the effective repository object format. The adjacent rule now limits
   declared/effective equality to an unsubstituted source and explicitly keeps
   declared state unchanged while substituted effective state records the
   actual object format and full commit.
2. **Independent policy gates.** Core 6.5 now requires the manager to apply
   applicable allowlist, revocation, registry, tag-lock, and audit-policy gates
   independently for each external repository subject. Every applicable gate
   must succeed before artifact-cache lookup or compiler work, and another
   subject's decision cannot be reused.
3. **Snapshot key, deduplication, and GC.** New Core 9.4 makes the protected
   snapshot key the complete tuple of effective identity kind/value, effective
   object format, full effective commit, and external build-source digest.
   Snapshot bytes may deduplicate only on complete-key equality; audit
   decisions remain subject-specific. It also fixes marker-v3 and journal
   roots, preserves marker-v1/v2 behavior, excludes receipt content alone as a
   root, requires conservative retention for unreadable or unprovable state,
   and prohibits execution, adoption, or permission-repair trust during GC.

The main outcome ownership map now points architecture-v6 sections 9, 9.1, and
9.4 to Core 9.4. Exact manager traversal, storage paths, CLI mechanics, and
diagnostic rendering remain downstream in `TASK-260728-wy3dsw`; schemas and
generated cases remain downstream in `TASK-260728-17sclp`.

## Scope and integrity

- `profiles/manager.md` and
  `decisions/0004-compile-only-build-drivers.md` each compare byte-identical to
  the accepted rc.4 source checkout (`cmp -s`, exit 0 for each).
- Worktree and source checkout HEAD are both
  `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- `git diff --cached --quiet` exited 0; nothing is staged and no commit was
  created.
- Current task-owned hashes:
  - `SECURITY.md`:
    `3b233a2af5fc1cac33f9af75079aeede7df3c37f0b94a91e8352c6df425483a7`
  - `protocol/core.md`:
    `e35f9a076fb7ad21b859e04b0ba88a8ae7bdbc544b3799db751dd6f6a0ea9384`
  - `decisions/0005-external-build-repositories.md`:
    `fa9ff8119350652052b29b462d5dab71af5dbd9201a9c23d25065605b72623fa`

## Validation evidence

Every gate ran directly without `tee` or a status-hiding pipeline.

| Command | Exit | Result |
|---|---:|---|
| `<task-venv>/bin/python tools/validate.py` | 0 | Validated 30 schemas and 93 vector files, including documentation-link checks |
| `env PATH=<task-venv>:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin make validate` | 2 | Failing attempt: schema/vector and 8 Python tests passed, but the narrowed harness path hid the installed `go` binary; this is not counted as a green full gate |
| `env PATH=<task-venv>:$PATH make validate` | 0 | Validated 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...` |
| `git diff --check` | 0 | No whitespace errors |
| `cmp -s profiles/manager.md <accepted-source>/profiles/manager.md` | 0 | rc.4 manager seed unchanged |
| `cmp -s decisions/0004-compile-only-build-drivers.md <accepted-source>/decisions/0004-compile-only-build-drivers.md` | 0 | rc.4 decision seed unchanged |
| `git diff --cached --quiet` | 0 | Index clean |

The focused rework is ready for independent review.
