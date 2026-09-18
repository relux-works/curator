# TASK-260918-24eazm: spec-dotfile-manager-platform-table

## Description
curator-spec: replace the §9.5 by-example heuristic list with one closed table per dotfile manager (chezmoi, home-manager, yadm, stow, dotbot at least) giving the well-known state location on macOS, Linux and Windows (or an explicit none), each row labelled verified or docs-confidence with its source, and the rule that the implementation reads the same table and that a table change is a spec revision; vectors per platform.

## Scope
(define task scope)

## Acceptance Criteria
§9.5 carries the closed per-manager, per-platform table with confidence labels and sources; the heuristic rule references the table and states platform resolution (home directory, %LOCALAPPDATA%, XDG variables) precisely; vectors exercise the heuristic per platform (present location → environment_foreign_manager_suspected; absent → no notice; a manager with none on a platform → inert) with a rule-7 validator gate; existing vectors byte-identical; make validate and regenerate-check exit 0; CHANGELOG entry.
