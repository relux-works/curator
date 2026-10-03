# THE ONLY CURRENT INSTRUCTION — TASK-261002-1foyf3: curator rc.14 pin and curator-content-v2 writer cut-over

curator-spec **v1.0.0-rc.14** is published:
- annotated tag object c13ab1bd4e6f7751e873a905ee545366e45fcf77;
- peeled commit **daf15ec8e78c29148063f14905fafc0b8b682786**;
- core manifest sha256 **6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5**.

Curator main 68210ecc already recognises that digest with exact counts (TASK-261003-1kcv6v). It keeps exactly one rc.14 known-gap row, snapshot-acquisition/cases/byte-exact-snapshot, owned by THIS task, because the writer is off.

Do (a coupled change, reviewed as one):
1. **SPEC_PIN.** Set `.github/workflows/ci.yml` SPEC_PIN to daf15ec8e78c29148063f14905fafc0b8b682786, the full peeled rc.14 commit. Verify that the checked-out manifest digest equals 6f832d81… and record it. Update `docs/ci-gates.md` and any digest or ledger that names the default pin.
2. **The v2 writer.** Set `internal/hashing/hashing.go` EnableV2Writers to true; this makes WriteVersion return curator-content-v2. Exercise the REAL entry points that write hashes: install targets and publication, markers, context resolution/materialization/store/locks, and environment marker/migration publication. New writes are v2. Readers still accept rc.13 (v1) artefacts. Keep the negatives: legacy-reader, NUL guard, mismatched version, and "new identities never silently alias old ones".
3. **Gap row.** Remove the byte-exact-snapshot row once that exact case passes under rc.14, and keep exact counts. Remove no other row unless its exact production case now passes; name any you remove.
4. **Unchanged.** Keep CodexSeedRevision and SecurityPostureRevision at "A". Keep the separate immutable module release pin in internal/buildrepo/release_pin.go (rc.8) unchanged.
5. **Checks.** Run, with real exit codes and following host-rules (-work):
   - `go test ./internal/hashing ./internal/marker ./internal/install/... ./internal/contextlock/... ./internal/envmarker/... ./internal/conformancecoverage ./internal/interop/... -count=1`;
   - plus the rows that pin v1 behaviour.
   The hosted gate is the arbiter for the full matrix.
6. **CHANGELOG.** One Unreleased line ("hash writes now use curator-content-v2; SPEC_PIN rc.14"). Never edit LOGBOOK.md.

Then run `task-board handoff TASK-261002-1foyf3 --role developer` and END YOUR TURN.

## Update (orchestrator, ~18:20Z): the rc.14 tag is being re-created
The first v1.0.0-rc.14 tag (on daf15ec8) was withdrawn before any release was published. Its release workflow failed on test-only freeze-guard tests, which are now being fixed. rc.14 will be re-tagged on a NEW spec main commit. The core manifest digest stays **6f832d81…**: the fix touches tools/ tests only.

- For now, keep SPEC_PIN = daf15ec8 as a placeholder; every check you run against it remains valid.
- The orchestrator will give you the final peeled commit as a short follow-up. Then only the SPEC_PIN line and docs/ci-gates.md change.
- Do not wait for it before handing off.
