# THE ONLY CURRENT INSTRUCTION — TASK-261002-7ajvgs: prepare curator-spec v1.0.0-rc.14 (release-prep change; NO tag)

The operator approved rc.14 on 2026-10-02. Scope is everything merged since v1.0.0-rc.13 up to main 045ceb2:
- #99, #113, #115, #116 (content-hash v2), #118, #120 (Decisions 0019/0021 adopted), #121 (muse) and #122 (rc.13 record restored, release-history freeze).

B3 (#119) is DEFERRED; do not include it. The skillfile-sources content-hash v2 (TASK-260930-3ny11n) is DEFERRED: the source suite stays unchanged.

Authoritative inputs:
- RELEASE.md;
- .github/workflows/release.yml, tools/generate-vectors, tools/release_gate.py, tools/validate.py;
- the readiness checklist `.research/261002_rc14_rc3_readiness.md` on curator main (TASK-261002-1pif8m), section "Exact release procedure and version files".

## Do
1. **Version files.** Advance the ACTIVE version to 1.0.0-rc.14:
   - README Version and release links;
   - CHANGELOG: a dated "1.0.0-rc.14" heading built from Unreleased, accurate per merged PR;
   - COMPATIBILITY.md and SECURITY.md: the compatibility/security notes, including the scoped-implementation-claims boundary;
   - `PROTOCOL_VERSION` in tools/validate.py and tools/release_gate.py;
   - protocolVersion and release date in tools/generate-vectors.

   Adapt the current-record assertions in the tests. Historical rc.13/rc.9 identifiers stay byte-identical.
2. **Corpus and record.**
   - Generate conformance/v1 as rc.14.
   - Add `release/1.0.0-rc.14.json`. It carries the EXACT generated manifest SHA-256 in candidate_protocol_pin and downstream_consumption, the unchanged source-suite digest and baseline, and truthful EMPTY unsupported implementation/platform/claim sets.
   - Preserve the portable default and do not claim the reserved hardened execution policy.
   - The rc.13 record and every released schema stay byte-identical: the #122 guard must pass.
3. **Diff coverage.** Add the rc.14 record to the generated-file diff lists in the Makefile and release.yml.
4. **Checks.** Run `make validate`, `make regenerate-check` and the tools tests, recording real exit codes. Print the final manifest SHA-256 in the results: the curator lockstep needs it.
5. **Implementation pins.** Leave .github/workflows/implementations.yml pins unchanged for now. The curator lockstep commit that accepts the rc.14 digest comes next, and the orchestrator updates the Go pin in the release PR afterwards. State in the results which Implementations rows will need the new digest.

No tag, no release. CHANGELOG is fine; never edit LOGBOOK.md. Never spell any employer name. The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

## Handoff
Update the results, then run `task-board handoff TASK-261002-7ajvgs --role developer`, then END YOUR TURN.
