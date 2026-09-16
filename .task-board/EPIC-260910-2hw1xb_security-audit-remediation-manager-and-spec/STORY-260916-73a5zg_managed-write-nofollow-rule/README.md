# STORY-260916-73a5zg: managed-write-nofollow-rule

## Description
Finding E5 (Medium): environments §9.5 detects a foreign-manager symlink and allows takeover with backup, but no sentence forbids writing through an existing link at a managed-surface path. The reference implementation had exactly this defect during EPIC-260905 (takeover wrote through a dotfile-manager symlink into the foreign source of truth) and fixed it locally; an implementation written from the text alone reproduces it.

## Scope
curator-spec environments §8.3/§9.5; curator materialization and takeover writes

## Acceptance Criteria
One normative rule: a takeover or repair write replaces the directory entry (unlink then create) and never follows a symlink at the target path, O_NOFOLLOW-class semantics on every managed-surface write; a conformance vector with a symlinked target proves the link is replaced and its target untouched
