# TASK-260926-1ddneb review verdict — CR rev1: ACCEPTED

Reviewer: claude-opus-5-5. Disposable clone at $TMPDIR; candidate tree reconstructed and confirmed == c0b9b24f18a75e4f49d56bdb99cc25bd1709af08.

1. maintainers.allowed_signers: line 1 oparin@me.com byte-identical to base; line 2 `bot@relux.works ssh-ed25519 AAAAC3...DTID`; key matches ~/.ssh/relux-git-bot.pub; ends with `
` (od). 
2. With candidate file (zsh, real exit codes):
   - verify-commit a21905d → Good signature bot@relux.works, rc=0 (with BASE file: rc=1 — change is load-bearing)
   - verify-tag v1.0.0-rc.9 → Good signature oparin@me.com, rc=0
   - commit signed by freshly generated unknown ed25519 key → "No principal matched", rc=1
3. `python3 tools/release_gate.py --version 1.0.0-rc.13 --commit HEAD` (venv w/ jsonschema) on candidate → "release gate passed", rc=0.
   tools/test_allowed_signers.py: 7 tests OK, rc=0. Sound for pinning exact file content (content-pin test; parser negatives are helper-level only — bound, acceptable for a data file).
4. Docs: GOVERNANCE.md:32-34 and RELEASE.md:74-76 enumerate both principals with the operator authorization date; nothing else changed.
   Minor non-blocking: GOVERNANCE.md:37 still says "The maintainer-signed tag explicitly authorizes the exact target" — slightly stale if Relux Bot signs the tag; follow-up wording only.
5. CHANGELOG: no Unreleased entry. This commit is the rc.13 release target and CHANGELOG top section is `## 1.0.0-rc.13`; adding an Unreleased section above it at the release target is inappropriate. AC item waived by reviewer; entry belongs in next release notes.