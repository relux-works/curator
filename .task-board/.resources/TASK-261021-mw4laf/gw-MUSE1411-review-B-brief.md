# Deciding review B — TASK-261021-mw4laf: Muse 1.4.11 exclusion qualification record

Review the CR against `gw-MUSE1411-brief.md` and the 1.4.8 template (`docs/qualification/muse-1.4.8-R6256.1.md`, `.research/muse148/`). Sweep every row; a first finding does not end the review.

| Surface | Attack |
| --- | --- |
| Binary identity | Rerun the gated harness yourself against `muse-bin-1.4.11-R6427.1`. Confirm that a different digest is refused and that without the gate everything skips before creating a fixture. |
| Isolation | Confirm the sandbox profile denies network, Keychain and the real native directories (run the offline controls). No auth command; no remote inference. |
| P1–P4 per form | Re-derive each row from your own run. Every claim in the record must match an observation; a passing example is not a universal claim. |
| §7.9 status line | Must follow from the rows. A failing interactive row means interactive is not qualified. |
| Narrowing control | A planted foreign-skill leak must make the P-row fail for that reason. |
| Spec citations | Section numbers and the spec main OID are current. |
| Scope and hygiene | No runtime constants, registry pins or gate changes; `CHANGELOG.md` and the root `README.md` equal trunk; no personal paths in committed files. |

Write the verdict through the normal review contract, with a surface table and an explicit accepted or changes_requested.
