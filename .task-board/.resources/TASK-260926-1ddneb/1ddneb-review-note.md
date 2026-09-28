# Review note — TASK-260926-1ddneb trust the Relux Bot release signer (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Operator decision 2026-09-26 (option b). This commit will be the v1.0.0-rc.13 release TARGET (squash-merged through GitHub, then tagged by
Relux Bot). Verify in a disposable clone:
1. maintainers.allowed_signers = the unchanged oparin@me.com line + `bot@relux.works ssh-ed25519 AAAAC3…DTID` (the key of
   ~/.ssh/relux-git-bot.pub; principal = the email Relux Bot signs with); trailing newline; tools/test_allowed_signers.py sound.
2. With the candidate file: verify-commit a21905d (Relux Bot) exits 0; verify-tag v1.0.0-rc.9 (maintainer) exits 0; a commit signed by an
   unknown key still FAILS (show it).
3. `python3 tools/release_gate.py --version 1.0.0-rc.13 --commit <candidate>` exits 0 on the candidate tree (the rc.13 release metadata stays
   valid at this target). CHANGELOG: no new Unreleased section above 1.0.0-rc.13 if that would break the release gate — judge whether the
   missing CHANGELOG entry is correct for a release-target commit and say so (waive AC item or require it).
4. GOVERNANCE.md / RELEASE.md wording accurate; nothing else changed.
accept_cr or changes requested with file:line. No LOGBOOK.md.
