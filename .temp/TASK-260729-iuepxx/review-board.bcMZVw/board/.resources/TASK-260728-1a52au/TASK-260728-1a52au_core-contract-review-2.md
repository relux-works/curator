# TASK-260728-1a52au core-contract review 2

Verdict: **ACCEPTED**. Route TASK-260728-1a52au to done.

## Review basis

- Binding architecture-v6 independently hashed to `2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`; accepted review-v6 was inspected.
- Run `RUN-260727-177dd8` is not goal-bound and had no operator directives at the verdict checkpoint.
- Producer worktree and accepted curator-spec source both remain at HEAD `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- `profiles/manager.md` and `decisions/0004-compile-only-build-drivers.md` are byte-identical to the accepted rc.4 source checkout.
- The index is clean, no task commit exists, and the isolated task delta is limited to `SECURITY.md`, `protocol/core.md`, and new `decisions/0005-external-build-repositories.md`.

## Rework verification

1. Core 6.3 now makes a network-substitution `revision` a full lowercase object ID for the effective repository object format. Declared state remains unchanged, and effective state records actual format and commit.
2. Core 6.5 independently applies applicable allowlist, revocation, registry, tag-lock, and audit-policy gates to every external subject before artifact-cache lookup or compiler work; decisions from other subjects cannot be reused.
3. Core 9.4 defines the complete external snapshot key, complete-key-only snapshot deduplication, subject-specific audit decisions, marker-v3 and in-flight journal roots, unchanged marker-v1/v2 behavior, receipt non-root behavior, conservative retention for unreadable or unprovable state, and no execution, adoption, or permission-repair trust during GC.

## Acceptance and regression sweep

- Schemas 1 through 6, `go-v1`, receipt v1, marker v1/v2, claim v1/v2, and rc.4 are explicitly frozen.
- Schema 7, first-class `build_repositories`, declared/effective identity, SHA-1/SHA-256 full commit locks, exact optional-tag assertion, descriptor trust boundary, monorepo target selection, manager-derived command/output, clean raw-object source proof, protected offline snapshot handling, audit-before-cache/compiler ordering, typed failures, credential/signing ownership, and closed future-driver admission are normative and internally consistent.
- Status, repair, publication, rollback, shim/PATH, protected cache, receipt-v2/marker-v3 mixed-command, deduplication, and GC boundaries preserve the accepted architecture without reopening rc.4.
- The producer outcome maps every architecture-v6 section 1 through 17.1, including sub-sections, to owned spec text or a justified downstream exclusion. The corrected section 9.1 and 9.4 ownership points to Core 9.4; wire schemas, exact manager mechanics, vectors, implementations, and release metadata remain correctly excluded.

## Independent validation

| Command | Exit | Result |
|---|---:|---|
| task-local Python `tools/validate.py` | 0 | Validated 30 schemas and 93 vector files, including local Markdown links |
| `PATH=<task-venv>:$PATH make validate` | 0 | 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...` passed |
| `git diff --check` | 0 | No whitespace errors |
| `git diff --cached --quiet` | 0 | No staged changes |
| rc.4 seed `cmp -s` checks | 0 each | Manager profile and Decision 0004 remain byte-identical |

Reviewed owned-file SHA-256 values match the producer outcomes:

- `SECURITY.md`: `3b233a2af5fc1cac33f9af75079aeede7df3c37f0b94a91e8352c6df425483a7`
- `protocol/core.md`: `e35f9a076fb7ad21b859e04b0ba88a8ae7bdbc544b3799db751dd6f6a0ea9384`
- `decisions/0005-external-build-repositories.md`: `fa9ff8119350652052b29b462d5dab71af5dbd9201a9c23d25065605b72623fa`

No blocking, rework, scope, or validation finding remains.