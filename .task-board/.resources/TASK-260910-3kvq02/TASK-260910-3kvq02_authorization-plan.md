# TASK-260910-3kvq02 planning and authorization boundary

Inspected revision: 65c6f1ec4b18c2374a9381028d12b78149114e7a.
No production code or tests changed. Existing untracked worktree board artifacts left untouched.

## Blocking instruction
TASK-260910-3kvq02_source-contract.md explicitly states: "Planning is authorized now; implementation dispatch and commits are not."
The task description independently says this task is not authorization to start work. Campaign rules supersede old host paths/policy, but do not explicitly lift this task-specific implementation restriction. Board notes show a developer dispatch, not an explicit override. No runtime directives were recorded.

## Inspection
Read accepted skillfile-sources protocol from current curator-spec checkout.
internal/manifest/sources.go already admits draft parsing behind ParseOptions.DraftSourcesV1, validates aliases, individual names, contained directory syntax, unique literal includes or *, and literal excludes.
internal/closure/closure.go Build currently queues only legacy Name/Git/Ref/Source; it does not consume Selector or expand collections. Existing nodes and unification are Git-commit based. Do not encode a local snapshot as a Git commit to bridge this difference.
Prior implementation outcomes were not retrieved: resources is not a supported get projection; resource list is not a CLI subcommand. Existing production implementation was inspected directly. No claim of prior-outcome verification.

## Proposed implementation after authorization
1. Retrieve prior accepted outcome resources using supported board resource query and preserve delivered parser behavior.
2. Connect draft selectors to the production resolution entry point with explicit source acquisition/package identity ownership.
3. Enumerate immediate children in UTF-8 byte order, prune managed outputs, validate explicit literal existence/type before exclusions, reject empty results and invalid discovered members.
4. Validate SKILL.md name/description and manifest identity; reject physical traversal and repeated/equivalent destination names.
5. Keep each expanded member as a skill-level closure node, retain diamond unification and reject identity/version conflicts.
6. Add production-entry positive, negative and legacy regression tests, including excluded missing literals, nested/nonrecursive discovery, invalid discovered metadata, links escaping roots, duplicate direct selections and conflicting dependencies.
7. Run narrow tests/build/lint and narrowing mutants; record exact exits and candidate tree before independent review. Do not run the full landing suite manually.

## Decision needed
Explicitly authorize implementation of TASK-260910-3kvq02 under the accepted draft contracts, superseding the planning-only sentence. Recommended option: authorize this bounded implementation while retaining no-commit and all campaign exclusions. Alternative: keep planning-only and defer dispatch.

## Verification bounds
No test, build, lint or mutation gate ran: this was read-only planning under an explicit authorization restriction. No checklist validation items checked. No production-path coverage claimed.
No LOGBOOK.md write: campaign rules explicitly prohibit it; this outcome and board notes preserve the finding instead.
