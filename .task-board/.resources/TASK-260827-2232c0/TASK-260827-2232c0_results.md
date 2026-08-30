# TASK-260827-2232c0 Rework Results (Round 4)

## Summary of Rework Changes

### 1. Protocol-Suite Pin Block (R1)
- Restored the protocol-suite pin block to `docs/ci-gates.md` under `## Protocol-suite pin and verification`.
- Includes module release pin (`internal/buildrepo/release_pin.go`), `curator-spec-pin` verification against curator-spec `v1.0.0-rc.8`, promotion tracking to `rc.9` in `../.github/workflows/ci.yml`, pinned commit `f8c405aa3ad0a39d260c2ed93684e55c5a346359`, SHA-256 digests (`d14e3a16...` manifest, `293f101d...` metadata), and local verification command `make verify-spec-pin ...`.

### 2. Complete 17-Row Gate Table (R2)
- Restored all missing language profile harness rows (`python_protocol_golden.py`, `npm`, `pnpm 10.33.0`, `Yarn Classic 1.22.22`, `Modern Yarn 4.9.2`, `Swift / SwiftPM`) and `Cross-adapter source-closure conformance` row to `docs/ci-gates.md`.
- `docs/ci-gates.md` now carries the complete 17-row gate catalog with exact invocations and repository-relative links (`../.github/...`).

### 3. Execution Assurance Fail-Closed Guarantee (R3) & Full Adapter List (R4)
- Restored `README.md` `## Execution assurance` section:
  - Explicit non-fallback verified execution (`execution.mode: verified`).
  - Fail-closed guarantee (release ships no platform provider; missing/unhealthy/drifted provider fails closed rather than using portable execution).
  - Disjoint cache key identities for portable, verified, legacy assurance-blind, cross-provider, and capability-drifted cache entries.
  - Receipt tracking for adopted/published artifacts.
  - Complete adapter enumeration: `Rust, SwiftPM, npm, pnpm, Yarn Classic, and Modern Yarn` (satisfying R4).

### 4. LOGBOOK Accuracy (R5)
- Updated `LOGBOOK.md` narrative entry for `TASK-260827-2232c0` to record the complete 17-row gate catalog, protocol-suite pin verification, accurate shim PATH claims, and verified execution assurance details.

### 5. Historical Header Reconciliation (R6)
- Updated line 1 of `docs/implementation-plan.md` to `Historical plan of record for Curator v0.1 against protocol 1.0.0-rc.8.`, matching line 6 of the same document.

## Line Count and Verification

- `README.md`: 122 lines (ceiling < 260 lines).
- `docs/ci-gates.md`: 39 lines with complete 17-row table and all required sections.
- `docs/compiled-commands.md`: 104 lines preserving all 10 load-bearing facts (B1-B10).
- `docs/implementation-plan.md`: header reconciled, body untouched.
- `go run ./cmd/curator ... --help` and `/tmp/curator-rev` tested cleanly.
