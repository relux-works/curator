# TASK-260916-1zgucp — rev5: drop two unaccepted hunks (THE ONLY CURRENT INSTRUCTION)

Delta review of rev4 = CHANGES REQUESTED (`TASK-260916-1zgucp_review-verdict-rev4.md`). The re-apply introduced content that is in NEITHER the
accepted rev3 NOR trunk:
F1. internal/config/environments.go:489-491 (parseOpenSSHPublicKey) — the `base64.RawStdEncoding` fallback. Remove it (restore rev3's
    behaviour; the signer key gate must not be widened without an accepted spec citation).
F2. internal/config/environments_conformance_test.go:414-440 — restore trunk d41da0fb's `exactManagerEffectiveJSON` and
    `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` VERBATIM (full normalized-object comparison; extra keys are a mismatch).
Change nothing else. `task-board m 'set_status(TASK-260916-1zgucp, status=development)'` first. Verify: `git diff 3a680a1e-tree` touches only
those two hunks; `go test ./internal/config` and `go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'` with real exit
codes. Append "Revision 5 — unaccepted re-apply hunks removed"; handoff and WAIT for the gate (never interrupt); hand off only green.
No CHANGELOG/LOGBOOK edit. Never add content during a re-apply that is in neither side.
