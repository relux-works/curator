## Rules for every rc.7 code-wave producer

- **Source of truth.** The normative text on curator-spec `main` (CIP-0014, CIP-0015, CIP-0017 and CIP-0018 are Accepted) and the owner decisions D1–D14 in `rc7-decisions.md`. The attached leaf comes from the rc.7 implementation plan and is your contract.
- **Negative evidence.**
  - Every forbidden outcome named by the leaf has an executable negative test.
  - Each test asserts the intended refusal code and the unchanged forbidden state, beside a valid positive control.
  - Plant the leaf's checker control and show the test fails on it.
  - Write a brief comment beside each test stating the claim it proves.
- **No forced fits.** If the platform or the normative text makes the leaf impossible or contradictory, stop and report the exact constraint. Do not add flags, stubs or mock-only behaviour.
- **Gates.**
  - Focused package tests first.
  - The full gate is `sh scripts/remote-gate.sh` (curator), or `make validate` plus the repository checks (curator-spec).
  - List exact commands and exit codes in the handoff.
- **Hygiene.** Both repositories and the curator board are public: English only; no host names, IPs, seat or lane names, personal names, local paths or incident references.
- **Delivery.** Leave the change uncommitted for the Change Request flow, as the board requires.
