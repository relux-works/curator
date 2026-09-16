# STORY-260916-73a5zg: managed-write-nofollow-rule

## Description
Finding E5 (Medium): environments §9.5 detects a foreign-manager symlink and allows takeover with backup, but no sentence forbids writing through an existing link at a managed-surface path. The reference implementation had exactly this defect during EPIC-260905 (takeover wrote through a dotfile-manager symlink into the foreign source of truth) and fixed it locally; an implementation written from the text alone reproduces it.

Implementation verification (TASK-260916-dv7xv5 rev2, curator main 80483355, launcher main b34e1e27, static): mitigated in code for the inspected case, spec rule absent. The copied root-context write removes the target first (switch.go:523-530, fallback :539-545) and replaceLink removes before os.Symlink (:694-696). Limits: remove-then-create is not O_NOFOLLOW and not race-free; writeStoreDocument still uses plain os.WriteFile (:686); other managed writes were not inspected. The manager task is a vector plus an atomicity review covering those writes.

## Scope
curator-spec environments §8.3/§9.5; curator materialization and takeover writes

## Acceptance Criteria
One normative rule: a takeover or repair write replaces the directory entry (unlink then create) and never follows a symlink at the target path, O_NOFOLLOW-class semantics on every managed-surface write; a conformance vector with a symlinked target proves the link is replaced and its target untouched
