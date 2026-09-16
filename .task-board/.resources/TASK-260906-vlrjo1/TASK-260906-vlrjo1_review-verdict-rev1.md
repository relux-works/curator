# Review verdict — TASK-260906-vlrjo1 rev1 (CR-TASK-260906-vlrjo1-1)

Reviewer run: RUN-260915-cb5667 (claude-fable-5-1). Date: 2026-09-15.
Verdict: **accepted**.

## What was reviewed

- Exact delta `3535d63e..e58a6af2`: one new file `.research/260915_windows-dotfile-manager-locations.md` (89 lines). No changes to `protocol/environments.md`, `profiles/manager.md`, schemas, or conformance vectors — "no spec edits in this task" holds.
- Board outcome resource `TASK-260906-vlrjo1_findings.md` exists (12450 B) with the same content as the repo file.

## Independent verification (my own, not the producer's evidence)

| Claim | Check | Result |
| --- | --- | --- |
| chezmoi Windows default sourceDir is `%USERPROFILE%/.local/share/chezmoi`, not `%LOCALAPPDATA%` | Fetched https://www.chezmoi.io/reference/configuration-file/variables/ | Reproduced: sourceDir default `"$XDG_DATA_HOME/chezmoi" / "$HOME/.local/share/chezmoi" / "%USERPROFILE%/.local/share/chezmoi"`; persistentState `%USERPROFILE%/.config/chezmoi/chezmoi.boltdb`. Task premise (LOCALAPPDATA) is wrong; findings correctly say so and label docs-confidence. |
| Home Manager config at `~/.config/home-manager/home.nix`, no Windows path | Fetched https://nix-community.github.io/home-manager/usage/configuration.html | Reproduced; no Windows path in the docs. "Not applicable on native Windows" is labelled as an inferred bound, not a fact. |
| Spec text and location | `protocol/environments.md:1794-1807` at 3535d63e | Quoted sentences match verbatim; "closed" list ends with "and the like" — ambiguity finding is genuine. |
| Profile citation | `profiles/manager.md:2303` | Phrase "well-known dotfile-manager state location" present at cited line. |
| Curator implementation | `internal/envprofile/managed.go:725-755`, call at 1667 (control root at 4f27ccb2) | Two entries, `os.UserHomeDir` + `filepath.FromSlash` + `os.Stat`, no OS filter; comment repeats the LOCALAPPDATA premise. Findings describe it accurately. |
| Evidence ledger | claude/codex versions, chezmoi/home-manager absent on PATH | Consistent with board notes; ledger honestly reports exit 1 as "unavailable", not passing. |

## Definition of Done

- Per-manager table with POSIX/Windows locations and evidence kind + URL: yes; every location is docs-confidence, none invented.
- Recommendation (platform-specific closed table) with exact 9.5 sentences to replace and the inserted table: yes.
- Both AC branches addressed: what the heuristic reports when no entry applies ("not applicable"), and absence vs. failed inspection distinguished.
- Brief conflation of `.claude`/`.codex` with dotfile-manager markers is corrected rather than propagated.
- Logbook: campaign forbids LOGBOOK.md edits; dated anomaly section kept in the outcome resource and board notes. Acceptable.
- Tests: research deliverable, no code; handoff validation log present. Not applicable beyond that.

## Notes for the follow-up (non-blocking)

- The proposed replacement paragraph is long; the docs follow-up may split it, but the semantics are right.
- Follow-up must fix the misleading comment in `dotfileStateDirs` and gate Home Manager to Linux/macOS, as the findings state.
- Bound: no Windows execution evidence anywhere in this task; all Windows locations remain docs-confidence until reproduced on a pinned release.
