# TASK-260918-bi6ouz: manager-dotfile-manager-table

## Description
curator: implement the landed environments §9.5 dotfile-manager table (curator-spec 802caee) in internal/envprofile (managed.go dotfileStateDirs → the closed per-manager × per-platform table read from one source in the spec order: chezmoi, home-manager, yadm, stow, dotbot × macOS/Linux/Windows with explicit none cells), the per-platform resolution rule (home, XDG_DATA_HOME/XDG_CONFIG_HOME defaults and the deliberate relative/empty-value policy, USERPROFILE), lstat presence discipline (symlinked directory does not count; failed inspection never reported as absent per §8.4.1), naming the manager in environment_foreign_manager_suspected; vectors/environments-dotfile-managers.json executed per platform through the real inventory path.

## Scope
(define task scope)

## Acceptance Criteria
1. One table constant mirrors the spec table exactly (order, cells, none), with a test that pins it against the vector file. 2. Resolution and presence rules implemented per spec with unit tests per platform (path resolution tested on every OS via injected env/home; Windows cells exercised on the Windows lane). 3. environments-dotfile-managers.json cases executed from CURATOR_CONFORMANCE_ROOT at or after curator-spec 802caee (skip only when unset; ledger rows as needed). 4. Docs/troubleshooting note and CHANGELOG entry stating the detection-semantics change; rule 8 hygiene.
