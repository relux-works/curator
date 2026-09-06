# TASK-260906-vlrjo1: dotfile-heuristic-list-is-posix-only

## Description
environments.md 9.5 gives the dotfile-manager heuristic a closed documented list spelled only in POSIX form, so the heuristic is silently inert on Windows. Decide whether the list is platform-specific and, if it is, what each managers Windows state location is - verified on an installed binary or labelled docs-confidence, never invented.

## Scope
9.5 says: the presence of a well-known dotfile-manager state location (a closed, documented list per manager - ~/.local/share/chezmoi, ~/.config/home-manager, and the like) elevates the notice for plain unmanaged files to environment_foreign_manager_suspected. Both spellings are POSIX. chezmoi on Windows keeps its state under %LOCALAPPDATA%, a different path entirely, and home-manager is Nix-only. The curator implementation resolves the operator home portably through os.UserHomeDir and joins the listed relative paths, so the lookup is correct and the list is what makes it inert: on Windows no listed location can exist. Found by stage (c) review cycle 2 (C2-m1) after the orchestrator raised it from a windows-latest lane failure whose fixture defect was a separate matter.

## Acceptance Criteria
9.5 states whether the closed list is platform-specific. If it is, each added location is either reproduced on an installed binary and recorded as verified, or labelled docs-confidence per the documents own verified/docs-confidence rule, and the manager profile and the implementation follow. If the list stays POSIX-only, 9.5 says so and says what the heuristic does on a platform where no listed location can exist, so an implementation can state a truthful bound instead of appearing to check something it cannot.
