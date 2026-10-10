# THE ONLY CURRENT INSTRUCTION — carrier TASK-261010-232rsr: re-apply two accepted studies byte-identical (researcher, mechanical)

Route decided by tb-keeper on 2026-10-10: STORY-261010-bujd60 cannot land, because its checkpoint does not descend from trunk (BUG-261010-3hzbcb). This carrier puts both accepted studies onto fresh trunk.

Do exactly this in your Story worktree:
1. Write `.research/261010_modular-instructions-design.md` with the exact bytes of `git show 1d7eb18c24c4f156730f5c14aa4790a8c86ed891:.research/261010_modular-instructions-design.md`.
2. Write `.research/261010_project-surfaces-coverage.md` with the exact bytes of `git show 2793eb6e0b393d8c6495a6e694f5d47f5cf7281b:.research/261010_project-surfaces-coverage.md`. Its sha256 must be 2dabff8504032b0c68439cd7b3a6990874684449a48a7f525a6db69fae363239.
3. Record both sha256 values next to their sources in a short resource `carrier-identity.md`. Confirm that `git status --short` shows only these two files.
4. Run `task-board handoff TASK-261010-232rsr --role researcher` and END YOUR TURN.

No edits to the content, no `LOGBOOK.md` change, no tests or builds on this host, no commits by hand. If a source object is missing, or the handoff refuses, attach the exact output as a resource and stop.
