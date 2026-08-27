## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260728-1a52au

## Blocks
- TASK-260728-wy3dsw
- TASK-260728-3b8qym

## Checklist
- [x] Add deterministic generated valid and invalid cases for every new schema branch
- [x] Prove schemas 1-6 remain compatible and reject schema-7-only fields
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
Execution directive (2026-07-28): Target /Users/iv/Developer/ReluxWorks/curator-spec. Use accepted TASK-260728-1a52au as the sole schema-7 prose base. Create /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree from HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730 and seed the accepted uncommitted files from /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1a52au/curator-spec-worktree, including SECURITY.md, profiles/manager.md, protocol/core.md, decisions/0004-compile-only-build-drivers.md, and decisions/0005-external-build-repositories.md, preserving accepted hashes. Do not mutate the source or prior worktree; do not copy .temp/.DS_Store; do not commit or stage. Implement only this task wire ownership: manifest schema 7, curator-build.json descriptor schema 1, Skillfile.dev schema 2, receipt schema 2, marker schema 3, claim schema 3, schema registry/version selection, deterministic generated valid/invalid cases, and schema 1-6 compatibility guards. Derive exact branches and examples from architecture-v6 plus accepted Core/Decision 0005; reject unknown/package-controlled argv, env, output/name, credentials, signing, hooks/plugins/generic drivers. Preserve all legacy bytes/meanings and do not add exact Git CLI profile text, manager code, shared runtime vectors, or release promotion. Follow existing generator/schema naming and canonicalization patterns after inspecting the repo. Run clean regeneration idempotence, targeted schema tests, full tools/validate.py and make validate, diff/format/no-stage checks. Attach a new task-scoped outcome with changed-file map and exact validation evidence and route to to-review.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260727-7f08cc, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260727-7f08cc)
Implementation logbook 2026-07-28: The pinned HEAD is rc.3, so schema-7 work was layered on the independently accepted TASK-260720-37ei85 composite rc.4 schema baseline and then the accepted TASK-260728-1a52au five-file schema-7 prose baseline. This preserves legacy wire bytes while avoiding reconstruction. Added seven closed external-repository schemas, semantic cross-field/version-selection validation, deterministic cases, and Go/Python guards. Final gates are green: 42 schemas and 263 vector files, 13 Python tests, Go tests/vet/format, two alternate-index regenerate-check passes, accepted-seed and rc.4 legacy byte comparisons, diff/no-stage/same-HEAD checks. One first legacy-comparison harness attempt exited 1 because zsh treats variable path as PATH; rerun with rel exited 0. Host Python import check exited 1 because jsonschema was absent; task-local pinned venv installation exited 0 and all actual validation used it.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-7f08cc, pid=72731, exit=0)
Independent wire-schema review directive (2026-07-28): Review /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree read-only against architecture-v6, accepted TASK-260728-1a52au core review, and the accepted rc.4 composite baseline identified in the producer outcome. First prove baseline integrity: accepted core/security/Decisions/profile bytes and all schema 1-6, receipt-v1, marker-v1/v2, claim-v1/v2, Skillfile.dev-v1 schemas/cases must match their accepted inputs; no unrelated deletion/reinterpretation, staging, or commit. Then adversarially review every new schema and semantic validator: manifest-v7 build_repositories and go-repository-v1 references; exact HTTPS/SSH/SCP URL, object-format/full-lock and optional-tag grammar; closed curator-build-v1 logical targets/build_root/source_dir with no output/argv/env/name; Skillfile.dev-v2 local/network substitutions with effective object format/ref and no package policy changes; receipt-v2 declared/effective/substitution/cache identity; marker-v3 explicit receipt-v1/v2 mixed branches and conditional top-level build_source; claim-v3 truthful driver/platform evidence; exact version registry selection; unknown/package-controlled field rejection; semantic containment/equality rules; deterministic generated valid/invalid cases and legacy rejection of schema-7-only fields. Probe JSON Schema conditional pitfalls, pattern false accepts/rejects, reference closure, aliases, additionalProperties/unevaluatedProperties behavior, integer/Unicode/path edge cases, and whether any runtime-only semantic was incorrectly claimed by schema. Independently run generator twice and compare, tools/validate.py, Python tests, make validate, Go tests/vet/gofmt, legacy byte comparisons, diff/no-stage/no-commit checks. Attach TASK-260728-17sclp_review.md with concrete ACCEPTED or CHANGES REQUESTED evidence and set done only on full acceptance.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-ae11ce, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-ae11ce)
Review logbook 2026-07-28 — CHANGES REQUESTED. Independent green gates: accepted prose/frozen legacy hashes, Go test/vet/format, 13 Python tests, 42-schema/263-vector validation, make validate, two byte-idempotent regenerations, diff/no-stage/pinned-HEAD. Adversarial validation found actual false accepts for non-ASCII and shell-metacharacter SSH paths, dot URL components, >255-byte tags, noncanonical identities, and SHA-1 effective sources with 64-hex revision refs; schema 1 also accepts reserved repository/target/go-repository-v1 top-level surface. Skillfile.dev v2 incorrectly requires the optional external map, and generated cases miss several semantic branches. Exact evidence and rework gate: TASK-260728-17sclp_review.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-ae11ce, pid=92016, exit=0)
Rework directive 1 (2026-07-28): Consume TASK-260728-17sclp_review.md and edit only /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree. Close all four findings with schema plus deterministic semantic/generator tests, preserving accepted baselines and scope. (1) Enforce raw transport/canonical identity/ref grammar exactly: HTTPS path may use valid Unicode scalars but no empty/dot/dot-dot components or forbidden whitespace/%/?/#/backslash; raw SSH URI/SCP repository path for go-repository-v1 is ASCII [A-Za-z0-9._/-] with nonempty no-dot components and rejects quotes, shell metacharacters, escapes, whitespace, and non-ASCII; host/identity is canonical lowercase host plus normalized path with one trailing lowercase .git removed and no dot components; tag/branch names obey safe-ref rules and 1..255 UTF-8 bytes; structured revision length/case must match the effective repository object format. Add 255/256-byte, SHA-1/SHA-256, Unicode HTTPS, ASCII SSH, metacharacter, dot-component, canonical identity/.git/case boundary vectors. (2) Keep frozen legacy schema bytes but add deterministic semantic version guards for every reserved schema-7 surface at relevant top-level/command locations for both manifest filenames and versions 1..6, including build_repositories, repository, target, and go-repository-v1/driver forms; avoid unrelated recursive extension rejection. (3) Make Skillfile.dev-v2 build_repository_substitutions optional; add ordinary-only, empty-map, and populated-map valid cases plus invalid ownership fields. (4) Add generated valid/invalid cases for every implemented structural/semantic branch, explicitly receipt containment and declared/effective mismatch, effective format/revision width, marker declared/effective mismatch and mixed build_source conditions, duplicate driver assertions, and driver-platform assertions outside top-level platforms; make Go inventory tests require the named cases. Independently rerun generator idempotence, all schema/Python/Go/vet/gofmt/full gates, baseline byte comparisons, and no-stage/no-commit checks. Update main results and attach a distinct TASK-260728-17sclp_rework-1.md outcome with exact failure-to-test mapping and evidence, then route to to-review.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260727-26b743, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260727-26b743)
Rework logbook 2026-07-28: Closed all four reviewer findings without changing accepted prose or frozen legacy bytes. Shared structural plus semantic validation now enforces exact HTTPS/SSH/SCP and canonical identity grammar, UTF-8 ref byte limits, effective object-format/ref width, every v1-v6 reserved schema-7 surface, optional Skillfile.dev external maps, and missing receipt/marker/claim branches. Final gates: 42 schemas/389 vector files, 15 Python tests, Go tests/vet/format/build, make validate, two byte-idempotent regenerations, accepted baseline comparisons, diff/no-stage/pinned-HEAD/artifact checks all exit 0. Expected-red and exact failure-to-test mapping are in TASK-260728-17sclp_rework-1.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-26b743, pid=2996, exit=0)
Fresh review cycle 2 directive (2026-07-28): Review current wire worktree read-only against the first verdict and all accepted inputs. Replay every original false accept and require rejection: Unicode/metacharacter SSH, HTTPS/SSH dot components, >255-byte tag/ref, wrong effective revision width, noncanonical host/path/.git identity, and every schema-1 reserved field example. Verify boundary positives remain valid: Unicode HTTPS, safe ASCII SSH URI/SCP, exact 255-byte ref, SHA-1/SHA-256 full revisions, canonical identities. Confirm Skillfile.dev-v2 ordinary-only/empty/populated maps and ownership negatives. Inspect location/version-aware legacy guards for both manifest filenames and all versions 1..6 without over-rejecting unrelated extensions. Verify all required receipt/marker/claim semantic branches have generated named fixtures and Go inventory enforcement, including containment, declared/effective and format/width mismatch, mixed build_source conditions, duplicate driver and platform subsets. Regression-review every original schema/semantic branch, JSON Schema closure, baseline byte equality, version registry, generated index/manifest, and task scope. Independently run 42-schema/389-vector validation, 15 Python tests, Go tests/vet/gofmt/build, make validate, two-pass byte-idempotence, baseline cmp, diff/no-stage/no-commit/artifact-cleanliness checks. Attach TASK-260728-17sclp_review-2.md. Mark done only if accepted; otherwise route exact defects to to-dev.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-33cb02, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-33cb02)
Review cycle 2 logbook 2026-07-28 — CHANGES REQUESTED. Prior correctness findings and all independent validation/baseline gates are green, but marker-v3 generation omits live empty-build, network-substitution, SHA-256, untagged, structured-ref-width, and substitution-identity branches. Exact evidence and rework gate: TASK-260728-17sclp_review-2.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-33cb02, pid=21845, exit=0)
Rework directive 2 (2026-07-28): Consume TASK-260728-17sclp_review-2.md and change only generator/tests/generated corpus in the existing worktree unless a demonstrable marker-only validator defect requires a minimal fix. Add deterministic marker-v3 generated cases and Go inventory requirements for every listed live branch: valid empty builds with no top-level build_source; valid network-git substitution using safe tag and branch; valid network substitution revisions at SHA-1 and SHA-256 widths; valid SHA-256 external record; valid untagged external record; invalid marker-local identity-kind mismatch; invalid marker-network identity-kind mismatch; invalid SHA-1 effective format with 64-hex revision and SHA-256 with 40-hex revision. Ensure cases exercise marker-v3/buildRecordV2 integration rather than merely duplicating receipt or Skillfile fixtures. Preserve all accepted schemas/validators unless a failing case proves a defect, and preserve all prior generated names. Run generation from an empty output tree, compare to checked corpus, run a second pass and byte digest comparison, then all 42-schema/Python/Go/vet/gofmt/build/make validate, baseline cmp, diff/no-stage/no-commit/artifact-cleanliness gates. Update results and attach TASK-260728-17sclp_rework-2.md with exact new case names and evidence, then route to to-review.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260727-5539a3, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260727-5539a3)
Rework 2 logbook 2026-07-28: Closed the cycle-2 marker-v3 generated-coverage finding with 11 marker-native cases covering empty builds, network tag/branch and SHA-1/SHA-256 revision substitutions, SHA-256 and untagged external records, local/network identity-kind mismatches, and both structured revision-width mismatches. Go inventory now requires every case. Final gates all exit 0: 42 schemas/400 vectors, 15 Python tests, Go focused/full tests, vet, format, build, make validate, empty-root generation, checked-corpus comparisons, second-pass digest 631d1555ecfa7d4e7223b1e1789dd60c26e8a8e320b1e6db5e51c7b758a7296f, accepted baselines, diff/no-stage/pinned-HEAD/artifact checks. Expected pre-regeneration inventory test exited 1 and is recorded in TASK-260728-17sclp_rework-2.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-5539a3, pid=28494, exit=0)
Fresh review cycle 3 directive (2026-07-28): Review rework 2 read-only. Verify the eleven exact marker-v3 case files exist, are generated rather than hand-authored, have the stated validity, and individually exercise marker-v3/buildRecordV2 integration for empty builds, tag/branch/revision network substitutions, SHA-1/SHA-256 boundaries, untagged records, identity-kind mismatches, and opposite revision-width failures. Confirm the Go inventory test requires every name/classification. Re-run direct validator probes for those branches and a regression sweep over all cycle-1 closures. Build the generator, populate a truly empty alternate output, compare all generated schema-case/index/manifest bytes, rerun second-pass digest, 42-schema/400-vector validation, 15 Python tests, all Go tests/vet/gofmt/build, make validate, baseline byte comparisons, diff/no-stage/no-commit/artifact-cleanliness. Attach TASK-260728-17sclp_review-3.md and mark done only if no correctness, coverage, baseline, scope, or validation finding remains.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-c17962, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-c17962)
Review cycle 3 2026-07-28 — ACCEPTED. The eleven marker-v3 cases are generated, indexed, inventory-required, and directly validate through the intended marker/buildRecordV2 branches. Independent 42-schema/400-vector validation, 15 Python tests, Go tests/vet/format/build, make validate, two-pass clean regeneration, 103-case adversarial replay, accepted prose comparison, legacy frozen-byte guards, clean index, pinned HEAD, and artifact checks all pass. Verdict evidence: TASK-260728-17sclp_review-3.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-c17962, pid=33230, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-7f08cc.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-7f08cc.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_results.md](file://TASK-260728-17sclp/TASK-260728-17sclp_results.md) — Schema-7 implementation map, both review rework summaries, and exact validation evidence
- [TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-ae11ce.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-ae11ce.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_review.md](file://TASK-260728-17sclp/TASK-260728-17sclp_review.md) — Changes-requested reviewer verdict with adversarial schema evidence and exact rework gate
- [TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-26b743.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-26b743.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_rework-1.md](file://TASK-260728-17sclp/TASK-260728-17sclp_rework-1.md) — Review finding closure map, deterministic cases, and validation evidence
- [TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-33cb02.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-33cb02.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_review-2.md](file://TASK-260728-17sclp/TASK-260728-17sclp_review-2.md) — Cycle-2 changes-requested verdict: missing marker-v3 generated branch coverage with independent validation evidence
- [TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-5539a3.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-implementer--developer--codex-_RUN-260727-5539a3.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_rework-2.md](file://TASK-260728-17sclp/TASK-260728-17sclp_rework-2.md) — Marker-v3 branch inventory closure and exact clean-regeneration evidence
- [TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-c17962.log](file://TASK-260728-17sclp/TASK-260728-17sclp_spawn-log_-reviewer--reviewer--codex-_RUN-260727-c17962.log) — System spawn log captured by task-board
- [TASK-260728-17sclp_review-3.md](file://TASK-260728-17sclp/TASK-260728-17sclp_review-3.md) — Cycle-3 accepted reviewer verdict with marker-v3 branch closure and independent validation evidence

## Created
2026-07-27T20:20:00Z

## Last Update
2026-07-27T22:20:22Z

## Assigned To
[reviewer] reviewer (codex)
