# TASK-260923-3e4i3n: native-policy-conflict-refusal

## Description
Implement the Decision 0018 known-conflict refusal in skill-agents-management: when LaunchRequest.PermissionMode resolves to yolo (and, where 0018 says so, native), known provider policy selectors in NativeArgs that conflict with the mapped mode are refused through the module's existing native-argument grammar (both = and separate-token placement), with exported stable typed errors the launcher can map. Update README Permission-mode section (remove 'later leaf' wording), CHANGELOG, release v0.5.19.

## Scope
(define task scope)

## Acceptance Criteria
1) every known conflicting selector for claude and codex refused in both placements with a typed error; 2) non-conflicting known selectors still forwarded; 3) row per selector x placement x mode executed; 4) one narrowing mutant per refusal family killed; 5) go test ./... exit 0; 6) CHANGELOG + README updated
