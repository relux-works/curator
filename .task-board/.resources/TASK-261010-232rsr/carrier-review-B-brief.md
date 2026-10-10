# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of Change Request rev1 of carrier TASK-261010-232rsr (read-only)

The carrier re-applies two studies that were already accepted, onto fresh trunk. tb-keeper chose this route on 2026-10-10 because STORY-261010-bujd60 cannot land (BUG-261010-3hzbcb).

Check:
1. **Delta.** From the CR base to the candidate tree, the change is exactly two added files: `.research/261010_modular-instructions-design.md` and `.research/261010_project-surfaces-coverage.md`. Nothing else, and no `LOGBOOK.md` change.
2. **Byte identity.** Each file equals its source: `git show 1d7eb18c24c4f156730f5c14aa4790a8c86ed891:.research/261010_modular-instructions-design.md` and `git show 2793eb6e0b393d8c6495a6e694f5d47f5cf7281b:.research/261010_project-surfaces-coverage.md` (sha256 2dabff8504032b0c68439cd7b3a6990874684449a48a7f525a6db69fae363239). Compare the blob ids or sha256 values yourself.
3. **Provenance.** The sources were reviewed. The modular study was accepted for TASK-261010-2uqd3t (review run RUN-261010-e69e1a). Review rev1 of TASK-261010-1992si accepted the project-surfaces content and rejected only the candidate scope (`TASK-261010-1992si_review-verdict-rev1.md`).
4. **Base.** The CR base is the protected trunk tip, or an ancestor of it whose later commits touch only `.task-board/`.
5. **Hygiene.** No secrets, personal or local paths, or session links in the two files. They were reviewed before; confirm with a scan.

No tests or builds on this host. Do checks 1–5 yourself and write your verdict BEFORE you read reviewer A's record `carrier-review-A.md` on this task. Then read it and reconcile. Accept (`accept_cr`) if 1–5 hold; otherwise request changes with the exact defect. Then END YOUR TURN.
