## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260728-17sclp
- TASK-260728-wy3dsw

## Blocks
- TASK-260728-pwbr32
- TASK-260728-zb2s4z

## Checklist
- [x] Run clean regeneration plus the complete curator-spec validation suite
- [x] Attach rc.5 release evidence and exact downstream protocol pin
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
EXECUTION DIRECTIVE (2026-07-28): Target curator-spec in a new isolated worktree /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree. Create it from pinned HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730, then seed the complete accepted uncommitted state from /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-wy3dsw/curator-spec-worktree byte-for-byte while excluding .git, .temp, caches, scratch probes, and build artifacts. Do not mutate any source/prior worktree; do not commit or stage. Binding inputs are architecture-v6 at /Users/iv/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260720-1nvomm/TASK-260720-1nvomm_external-build-repositories-architecture-v6.md, accepted Core/Decision 0005, accepted schema-7/descriptor-v1/Skillfile.dev-v2/receipt-v2/marker-v3/claim-v3 wire corpus, and accepted manager profile/CLI through review cycle 3. Implement the shared rc.5 conformance and release layer only: executable positive/adversarial fixtures and vectors for SHA-1/SHA-256 tagged/untagged acquisition, HTTPS/SSH/local substitution, exact refs/raw objects/config/packs/index/LFS/special files, whole-snapshot audit ordering, protected cache hit/miss/corruption/offline, mixed receipt-v1/v2 and marker-v2/v3 planning, transaction/rollback/status/repair/GC/shim/PATH/signing boundaries; harness/validator/generator support and deterministic clean regeneration; frozen legacy byte/behavior guards; authoring/operator guidance; release metadata and exact downstream protocol pin; claim-v3 qualification rules that include only evidenced driver/platform tuples and explicitly exclude Linux until TASK-260728-1skseh later passes. Do not implement Curator/csk managers, contact real remotes, fabricate macOS/Windows/Linux evidence, add a generic language driver, reopen schemas 1-6/go-v1/receipt-v1/marker-v1-v2/claim-v2/rc.4, or weaken independent audit-before-cache/compiler and fail-closed signing/offline semantics. Prefer hermetic local fixtures with exact expected bytes/errors and stable platform-independent test oracles; platform-native implementation evidence belongs downstream. Inspect existing conformance/release conventions before editing. Run clean generation twice and prove a zero second diff, complete schema/link/Python/Go/make validation, legacy hash/compatibility checks, fixture limit and negative-case checks, diff/no-stage/no-commit and accepted-baseline comparisons. Attach a task-scoped outcome with files, fixture matrix, exact release/downstream pin, platform claim exclusions, hashes and commands; route to to-review, never self-accept.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260727-1a699b, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260727-1a699b)
Logbook 2026-07-28 — Implemented the rc.5 shared external-repository corpus in the isolated curator-spec worktree. Added exact SHA-1/SHA-256 acquisition, local Git/config/ref/pack/index/raw-object/LFS fixtures, lifecycle/cache/mixed receipt-marker/rollback/status-repair-GC/shim-PATH/signing vectors, claim-v3 evidence gating, frozen rc.4 guards, author/operator docs, and release metadata. Exact candidate protocol pin is sha256:30a64ed0da6e4e68abb5f46e8807f7bc57a4545c7c582e644c9d09c9406c9324. Candidate emits no platform claim; macOS/Windows remain pending downstream evidence and Linux is excluded until TASK-260728-1skseh. Full make validate and clean-probe make release-check passed; two consecutive clean regenerations produced zero diff. Assigned worktree remains at HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730 with no staged changes or commit. Initial expected-red checks exposed and fixed an rc.4 claim constant bug; the system Python dependency gap was resolved with a task-local venv.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-1a699b, pid=90587, exit=0)
INDEPENDENT REVIEW DIRECTIVE (cycle 1): Review /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree read-only against architecture-v6 and all accepted rc.5 predecessor work, especially accepted TASK-260728-wy3dsw review 3. First prove pinned HEAD, clean index/no commits, and content-only delta against the accepted manager-profile worktree; distinguish harmless regenerated mtimes from byte changes and prove frozen schemas 1-6/go-v1/receipt-v1/marker-v1-v2/claim-v2/rc.4 bytes remain exact. Adversarially inspect and execute the new corpus: recompute SHA-1/SHA-256 raw object IDs; validate commit/tag/tree/blob bytes; independently parse/check pack v2/v3 plus index-v2 hashes/trailers/fanout and negative mutations; verify exact LFS 1023/1024 and near-miss semantics against the pinned detector; inspect local config/ref/alternate/replace/graft/promisor/filter/helper/special-file cases; verify exact tagged/untagged HTTPS/SSH/local substitution and stable failure mappings; verify whole-snapshot audit precedes cache/compiler on every hit/miss/dry-run/repair/audit path; inspect mixed receipt-v1/v2 and marker-v2/v3 expected bytes, transaction/rollback/status/repair/GC/shim/PATH/signing cases and fixture size limits. Prove required AC/threat cases are not merely named but have meaningful exact oracles. Validate manifest inventory exactly once per generated file, no self-referential/cyclic hash, release/1.0.0-rc.5.json exact manifest pin and candidate-only semantics, committed_release_pin_advanced=false, empty claim-v3 candidate claims, no fabricated macOS/Windows evidence, and explicit Linux exclusion until TASK-260728-1skseh. Re-run generation twice and prove the second content diff is empty; run full validate/Python/Go/vet/gofmt/compile/release-check in a clean disposable copy and independently recompute reported manifest/release hashes. Check CI/release/docs consistency and that no manager implementation, generic driver, real remote contact, or platform claim entered scope. Attach TASK-260728-3b8qym_review-1.md with exact ACCEPTED or CHANGES REQUESTED findings and evidence. Set done only on complete implementation-ready acceptance; otherwise route to to-dev. Do not modify the task worktree.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-946f60, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-946f60)
Review cycle 1 CHANGES REQUESTED: exact receipt cache_key and mixed marker receipt_sha256 do not match normative CCJ-1 hashes; external source-covering dry-run/audit-only ordering vectors are absent; two pack-negative cases are descriptive and not cryptographically enforced. Full clean suite otherwise passed. See TASK-260728-3b8qym_review-1.md.
Logbook 2026-07-28 — Independent review cycle 1 requested changes. Normative CCJ-1 recomputation proves the published receipt cache key and mixed-marker receipt hash are false exact oracles; source-covering dry-run/audit-only ordering vectors are absent; pack checksum/hash-family negatives are not executable or cryptographically enforced. Clean release gates, legacy-byte guards, manifest inventory, positive raw-object IDs, positive pack/index structure, and claim-v3 exclusions otherwise passed.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-946f60, pid=9723, exit=0)
REWORK DIRECTIVE (review cycle 1): Work only in /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree at pinned HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730. Read TASK-260728-3b8qym_review-1.md and preserve all independently verified work while closing all three blockers. (1) Generate build-receipt-v2 cache_key as SHA-256(CCJ-1(receipt.input)) and mixed marker external receipt_sha256 as SHA-256(CCJ-1(full generated receipt)); propagate these exact values to mixed-build-plan and any references. Add independent Python and Go semantic assertions that canonicalize with the project CCJ-1 implementation and fail on either mismatch, not string constants. (2) Add executable external-repository source-covering dry-run and coverage-claiming audit-only cases with explicit ordered phases/outcomes/no-mutation properties; enforce independent whole-snapshot audit before artifact-cache lookup/compiler for cache hit, cache miss, dry-run, repair, and audit paths, while keeping syntax-only check disjoint. (3) Replace descriptive reject-index-checksum-mismatch and reject-pack-hash-family-mismatch with self-contained concrete bytes or exact base_fixture+deterministic mutation references. The harness must materialize mutations, parse index-v2 fanout/trailers, recompute pack/index checksums under declared hash family, prove the intended single fault, and assert exact stable errors. Strengthen positive tests beyond magic. Regenerate all outputs, manifest, release/README/docs pin references and release metadata so the old sha256:30a64e... pin is nowhere presented as current. Preserve empty candidate claims, committed_release_pin_advanced=false, macOS/Windows pending and Linux excluded until TASK-260728-1skseh. Run full clean release-check, two consecutive regenerate-checks with zero diff, Python/Go/vet/gofmt/compile/diff/index/pinned-head and content-only legacy guards. Attach TASK-260728-3b8qym_rework-1.md with revised hashes, exact failures fixed and evidence; route to to-review, do not self-accept or commit/stage.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260727-ea6326, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260727-ea6326)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-ea6326, pid=20161, exit=0)
INDEPENDENT REVIEW CYCLE 2 DIRECTIVE: Review /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree read-only after rework 1. Treat TASK-260728-3b8qym_review-1.md as the blocking baseline, but independently reproduce every closure. (1) Using the repository CCJ-1 canonicalizer and a separate implementation/check, recompute SHA-256(CCJ-1(receipt.input)) and SHA-256(CCJ-1(full receipt)); require exact values in receipt.cache_key, marker external cache_key/receipt_sha256, mixed plan and regenerated schema examples. Mutate each relationship and prove both Python and Go gates fail for the intended reason. (2) Inspect exact source-covering dry-run and audit-only vector bytes/fields and independently verify no-mutation plus ordered acquisition -> snapshot validation -> external audit; verify audit precedes cache/compiler separately for hit, miss, dry-run, repair and audit, and syntax-only claims none of source/audit/cache/mutation. (3) Materialize both pack negatives from base+structured mutation, compare published concrete bytes, parse full PACK/index-v2 structure and checksums, prove checksum negative flips exactly one final byte, prove family negative is valid SHA-1 and invalid as declared SHA-256, and verify exact stable error. Ensure malformed mutations/base refs would fail. Reconfirm all previously accepted corpus coverage, fixture ceiling, frozen legacy content, manifest uniqueness/non-self-reference, empty candidate claims, pending macOS/Windows, Linux exclusion, committed_release_pin_advanced=false and no real remote/manager implementation. Independently recompute new manifest pin sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8 and release metadata/file hashes; verify obsolete sha256:30a64ed0... absent from current project files. Run two clean regenerations, full validate/Python/Go/vet/gofmt/compile/release-check in disposable clean repo, diff/index/pinned-head/content-only baseline checks. Attach TASK-260728-3b8qym_review-2.md with ACCEPTED or exact CHANGES REQUESTED. Set done only if complete corpus and candidate release layer are implementation-ready; otherwise route to to-dev. Do not modify task worktree.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-9d00b4, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-9d00b4)
Logbook 2026-07-28 — Independent review cycle 2 ACCEPTED. Recomputed CCJ-1 receipt oracles and rc.5 manifest pin, cryptographically materialized both pack negatives, verified dry-run/audit/cache/repair ordering, matched 196 frozen legacy artifacts to the accepted predecessor, and passed two clean regeneration runs plus the complete release-check. Exact downstream pin: sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8. Candidate claims remain empty; Linux remains excluded until TASK-260728-1skseh.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-9d00b4, pid=43081, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260728-3b8qym_spawn-log_-implementer--developer--codex-_RUN-260727-1a699b.log](file://TASK-260728-3b8qym/TASK-260728-3b8qym_spawn-log_-implementer--developer--codex-_RUN-260727-1a699b.log) — System spawn log captured by task-board
- [TASK-260728-3b8qym_results.md](file://TASK-260728-3b8qym/TASK-260728-3b8qym_results.md) — rc.5 conformance corpus, exact downstream pin, platform exclusions, file matrix, hashes, and validation evidence
- [TASK-260728-3b8qym_spawn-log_-reviewer--reviewer--codex-_RUN-260727-946f60.log](file://TASK-260728-3b8qym/TASK-260728-3b8qym_spawn-log_-reviewer--reviewer--codex-_RUN-260727-946f60.log) — System spawn log captured by task-board
- [TASK-260728-3b8qym_review-1.md](file://TASK-260728-3b8qym/TASK-260728-3b8qym_review-1.md) — Independent rc.5 reviewer verdict and rework evidence
- [TASK-260728-3b8qym_spawn-log_-implementer--developer--codex-_RUN-260727-ea6326.log](file://TASK-260728-3b8qym/TASK-260728-3b8qym_spawn-log_-implementer--developer--codex-_RUN-260727-ea6326.log) — System spawn log captured by task-board
- [TASK-260728-3b8qym_rework-1.md](file://TASK-260728-3b8qym/TASK-260728-3b8qym_rework-1.md) — Review-cycle-1 rework: corrected CCJ-1 hashes, executable pack negatives, source-covering ordering, revised pin, and full gate evidence
- [TASK-260728-3b8qym_spawn-log_-reviewer--reviewer--codex-_RUN-260728-9d00b4.log](file://TASK-260728-3b8qym/TASK-260728-3b8qym_spawn-log_-reviewer--reviewer--codex-_RUN-260728-9d00b4.log) — System spawn log captured by task-board
- [TASK-260728-3b8qym_review-2.md](file://TASK-260728-3b8qym/TASK-260728-3b8qym_review-2.md) — Independent rc.5 review cycle 2 acceptance verdict and release evidence

## Created
2026-07-27T20:20:00Z

## Last Update
2026-07-28T00:34:53Z

## Assigned To
[reviewer] reviewer (codex)
