# STORY-260916-12lbww: onboarding-heuristic-platform-parity

## Description
The §9.5 dotfile-manager heuristic (well-known state locations that raise environment_foreign_manager_suspected) is POSIX-only today. Platforms must stay in sync: one closed table per manager with macOS, Linux and Windows locations, declared in the spec and implemented from the same table.

## Scope
curator-spec protocol/environments.md §9.5 (closed heuristic table with per-platform paths), schemas/vectors; curator onboarding heuristic implementation

## Acceptance Criteria
§9.5 carries one closed table with macOS, Linux and Windows locations per manager (chezmoi, home-manager, yadm, stow, dotbot at least), each row labeled verified or docs-confidence; the implementation reads the same table; vectors per platform; a table change is a spec revision, never an implementation-private list
