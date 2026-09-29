# TASK-260910-32gki6 — S5 Git pin verification blocker

## Disposition

Blocked pending a protocol decision. The current Story worktree contains the earlier rev1 implementation and tests, but its gate-fix requires resolve to be independent of the source repository. I made no product-code changes and ran no tests in this continuation because the pinned lock model does not provide enough trusted information to implement that requirement for Git members.

## Constraint and evidence

The rc.13 environments spec requires resolve to recompute each store entry's tree hash and compare it to the pin (§4, lines 724–732; §10.1, lines 2874–2889). The same spec defines a Git pin as the resolved commit object ID (§1, lines 53–65), while schema 1 records only commit or state_sha256, never both (§1.3, lines 163–180). It carries no Git snapshot tree hash. The curator lock implementation enforces this: Member has only Commit and StateHash (internal/contextlock/contextlock.go:36–47), rejects both together (:246–264), and its strict parser rejects unknown fields (:364–389).

A commit object ID is not its root tree object ID. On this checkout, HEAD is 213a53e5c701ef16b961f3398c7f2f0b82c98094 while HEAD^{tree} is 9e6d659fd4303c23e3616969edc4749e989e3325. Recomputing the Git tree identity from store bytes therefore cannot be compared directly to the lock's commit ID. Obtaining the expected tree ID requires either the pinned commit object database or another trusted record. The current gate-fix explicitly rejects reopening the source repository during resolve; adding a new lock field or changing what commit means would alter the frozen rc.13 schema/contract.

The earlier rev1 implementation derived an expected hash by extracting the pinned snapshot from the local profile repository. That detects byte swaps when the cache is available, but the gate-fix says resolve must not need that repository or network. Treating a recomputed tree ID as equal to the commit ID would be incorrect; writing an unbound sidecar would not make the expected tree hash lock-authenticated.

## Decision needed

Choose the contract that makes the Git expected tree identity available during resolve:

1. Recommended: revise/version the protocol lock to carry a Git snapshot tree hash bound by the lock, then implement source-independent recomputation against that field. This needs a spec/schema revision before this leaf can conform.
2. Amend the gate-fix to allow local pinned-object-database lookup during resolve (with no network), and define the fail-closed outcome when that cache is missing or unreadable. This makes resolve depend on the retained Git object database.

Please select which contract S5 should implement. No lock-schema extension or cache-dependent workaround was added in this leaf.

## Rules, rows, and vectors

| Rule | rc.13 clause | Production row/vector state |
|---|---|---|
| Verify enclosing roots, lock, marker, and named entries with all five boundary checks | §4; §10.1 | The existing rev1 candidate wires checks into env resolve; its 20 resolve_cases and 4 status_cases were reported in the prior results resource. The gate-fix is not accepted until Git pin verification is source-independent. |
| Recompute each entry pin, including Git snapshots | §4; §10.1 | Blocked for Git members because schema 1 carries no tree hash and the gate-fix forbids resolving the commit object from the source repository. Prior rev1 used local snapshot extraction. |
| Repair eligible entries from a revalidated snapshot; never rebuild from an unreadable lock | §4; §8.4.1; §10.1 | Existing rev1 reported 10 repair_cases; no new evidence was run in this continuation. |
| Dry-run reports an entry rebuild without writes | §4; §10.1 | Existing rev1 reported 4 dry_run_cases; no new evidence was run in this continuation. |
| Verify recorded system-prompt and root-context surface hashes | §8.4; §10.1 | Existing rev1 reported a focused surface-drift test; it was not rerun in this continuation. |

The rc.13 conformance ledger currently has 0 rows owned by STORY-260910-148pj1 or TASK-260910-32gki6; before and after counts are 0 → 0. The ledger has 10 total data rows and was not edited. No gap row was removed in this continuation. The prior results resource reported 38 vectors driven across the four boundary families, but that evidence does not resolve the source-independent Git pin requirement.

## Validation in this continuation

No test, build, or lint command was run after the gate-fix because the mismatch is in the normative input needed to determine valid implementation behavior. The prior results resource contains rev1 local run evidence; the gate-fix note identifies why that candidate needs revision. Hosted validation has not run for this continuation.

No CHANGELOG.md or LOGBOOK.md file was edited. The earlier release-prep entry remains in the prior results resource and should be finalized after the pin contract is decided.
