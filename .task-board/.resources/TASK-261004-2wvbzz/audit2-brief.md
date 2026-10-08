# THE ONLY CURRENT INSTRUCTION — TASK-261004-2wvbzz: robustness and invariant review of curator, wave 2 (researcher)
This is defensive quality work on our own open-source code: we write regression tests for invariants the code must keep, and report where an invariant does not hold. No attack tooling and nothing aimed at third parties. Everything runs on temporary local fixtures only.
Context: docs/security-audit-2026-10-inline.md (wave 1; N1/N2/N3/N5 are fixed on main). Work in a disposable copy (`git archive HEAD` into $TMPDIR).
For each area below, list the invariants the spec or code promises. Then write a Go test through the production entry point that checks each invariant on local fixtures, and record pass or fail with exit codes:
(a) transaction namespace and snapshot consistency under concurrent installs;
(b) global shim ownership: only manager-owned shims are replaced;
(c) credential helper and worker message framing stay within their size and format bounds;
(d) package artifact intake: the declared limits are enforced end to end;
(e) curator-run launcher: discovery honours the trust roots, and fragments and permissions pass through unchanged;
(f) audit pin and revocation decisions match the spec table;
(g) Windows-specific paths, by reading.
Output: .research/261004_inline-audit-wave-2.md (EN). Each invariant that does not hold becomes a finding (N6+) with its failing test and exit code. Include a coverage table. Attach the tests as resources. Change no product code.
Full codex pace (tb-R181). Use GOFLAGS=-work. Never edit LOGBOOK.md. Before the handoff, update the results resource.
Then `task-board handoff TASK-261004-2wvbzz --role researcher` and END YOUR TURN.
