# TASK-260916-1zgucp — orchestrator decision on the rev5 block (THE ONLY CURRENT INSTRUCTION)

Your block is correct: the two rev4 hunks were needed because rev4 drives manager-config-v2 vectors that were gap rows on trunk. Decision
(technical, from the rc.13 text — no upstream change needed):
Q1 (unpadded keys). environments.md §12.1 (lines ~3580-3590): an ssh entry's `key` is "an OpenSSH public key line — `<key-type> <base64>
[<comment>]`" with the key-type from the closed list, and "an ssh entry's identity is its key type plus its base64 key material". The spec
defines a GRAMMAR and a STRING identity; it does not require the material to decode to a well-formed key blob. Its own valid vectors
(schema2-every-knob, schema2-source-signers) carry 66-char unpadded material. So config admission must accept base64 material padded or
unpadded (alphabet check, no wire-format decode requirement); identity comparison is type + material string (normalise padding
consistently on both sides, and say how). Cite §12.1 in a code comment and add a negative row (non-base64 characters / unknown key-type →
refused) so the gate is not widened beyond the grammar. Signature VERIFICATION keeps using real keys (a malformed blob simply never matches).
Q2 (comparator). 24 of 56 manager-config-v2 vectors give `expected` = {"environments": …} only — a projection. Keep trunk's strict
full-object comparator and TestManagerEffectiveJSONComparisonRejectsExtraKnobs VERBATIM for full-object vectors. For a vector whose expected
declares a subset of top-level keys, compare exactly those keys, each with the same strict canonical-bytes comparison (an extra or missing
key INSIDE `environments` is a mismatch); add a regression row proving an extra knob inside `environments` fails a projected vector.
Everything else of rev4 stays. `task-board m 'set_status(TASK-260916-1zgucp, status=development)'`; `go test ./internal/config` and the
envprofile focused run with real exit codes; append "Revision 5 — grammar-level signer keys, projected manager vectors" with the §12.1
citation; handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK edit.
