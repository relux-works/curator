# TASK-260910-1xph2y: write-and-review-approved-source-contracts

## Description
Update curator-spec for the approved detailed Skillfile source design: backwards compatibility, local and Git sources, individual and collection selection, source/output separation, complete local runtime/build behavior and transport-neutral repository identity with machine policy. Use Codex gpt-6-astra medium producer and independent reviewer. Leave the reviewed result uncommitted.

## Scope
curator-spec normative prose, versioned schemas, author/operator examples, compatibility and relevant conformance validation. Exclude manager implementation, broad rules/knowledge/root-context/MCP/plugin and prebuilt binary-distribution work. No commits or publication.

## Acceptance Criteria
The approved contract is implementable and consistent across prose, schema and examples. v1 semantics remain unchanged; new forms reject ambiguous mixes. Local snapshots cover context/runtime/build changes and protect managed output boundaries. Repository identity is stable across permitted SSH/HTTPS resolution and credential/integrity failures remain fail-closed. Relevant positive and negative validation passes. Reviewer accepts the current CR; no commit/checkpoint/integration commands run.
