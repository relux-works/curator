# THE ONLY CURRENT INSTRUCTION — TASK-261002-1ig5ev: restore the rc.13 release record and freeze release history (curator-spec)

## Evidence
Release-readiness review TASK-261002-1pif8m: `git diff --quiet v1.0.0-rc.13 origin/main -- release/1.0.0-rc.13.json` exits 1.
- The tagged record pins the published suite digest be11bb1e… in `candidate_protocol_pin` and `downstream_consumption`.
- Main pins the CANDIDATE digest bd03456… there.
- The rewrite came in via PRs #113/#116/#121: the generator regenerates the rc.13 record from the current corpus.

A published release record must never change.

## Do
1. Restore `release/1.0.0-rc.13.json` byte-for-byte to the v1.0.0-rc.13 tag. Check with `git diff --quiet v1.0.0-rc.13 -- release/1.0.0-rc.13.json`, exit 0.
2. Stop the generator (`tools/generate-vectors`) from writing records of published versions. Candidate digest bookkeeping for the next version must go somewhere that is not a published record. Use the established candidate mechanism if one exists; otherwise propose the minimal one and say so in the results. Do NOT create the rc.14 record yet; that is a separate release-prep leaf.
3. Add a byte-freeze guard. `tools/validate.py` and/or a test fails when any `release/<published>.json` differs from its tag's bytes, in the same spirit as the existing released-schema immutability guard.
   - Negative test: mutate the rc.13 record → the guard fails, with a real exit code.
   - `make regenerate-check` must still pass.
4. Run `make validate` (or the subset that runs locally) and the tools tests. Record real exit codes.
5. CHANGELOG under Unreleased: one line. No LOGBOOK. Never spell any employer name.

The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

## Handoff
Update the results, then run `task-board handoff TASK-261002-1ig5ev --role developer`, then END YOUR TURN.
