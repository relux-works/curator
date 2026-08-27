# STORY-260728-19nx3g: external-build-repository-interop

## Description
Qualify curator-spec, Curator, and cocoaskills as independent rc.5 consumers of the external build repository contract. Compare canonical bytes and black-box behavior for declared and substituted sources, mixed local/external commands, failures, caches, transactions, and release evidence.

## Scope
Shared fixtures and vectors, cross-manager consumers, black-box interop runner, authoring guidance, language-driver admission guidance, rc.5 release qualification and implementation pins. Linux is excluded from claim v3 until its separate native-validation story passes.

## Acceptance Criteria
Both managers consume the same accepted spec release and independently produce identical canonical identities, digests, receipts, markers, and typed outcomes where the protocol requires equality; positive and adversarial Git/source/cache/lifecycle vectors pass; macOS and Windows native evidence is recorded; release pins and claim v3 cannot overstate Linux or future-language support.
