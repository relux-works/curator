# TASK-261011-2lwcrk — reviewer B verdict, CR-TASK-261011-2lwcrk-2 revision 2

**Verdict: ACCEPT.** No P0/P1 remains. Docs-only delta of 8 paths (base a60fd34, candidate tree ce2b96e). Checks 1–5 were done before reading reviewer A's record (`p0-R-CS1-review-A.md`); then A's single finding was reconciled.

## Checks (done independently, then compared with A)
1. **Rows vs contract table (177cb554) and audit (48789217).** All 10 A-rows (A04.4, A04.5, A11.2, A11.5–A11.11) verified against the changed text, plus the five B2 diagram rows.
   - A04.4/D1: three separate keys; A/C Secure Enclave P-256 on supported Macs, files elsewhere; B bridge-owned 0600 Ed25519, no Keychain; K_enc is X25519. Diagram C4 Ed25519 text sits in the B-only view and is scoped.
   - A04.5/D3: host CA plus user CA, TXT `_swarma-ssh-ca`, `@cert-authority`, `TrustedUserCAKeys`, KRL, overlap rotation. Durations are "policy examples". No SSHFP, `AuthorizedKeysFile` or `restrict,command=` recipe remains (grep clean in the 5 diagrams, CIP-0008/0009 and the protocol file).
   - A11.2/D1/D5: A first with no admin and no OS boundary; B with one admin request and no Keychain, disabled until the separate provisioner exists; C sandbox fallback. CIP-0008 R3/R5/R7/R8/threat model/layers are variant-scoped. Open question 7 (A selectable after B/C) is preserved, not decided.
   - A11.5/D2: DNSSEC validated in the client with DoH transport fallback. Invitation pins only for a declared no-DNSSEC environment, chosen before lookup. Pins must agree. Read, validation or mismatch failure refuses. Records are per environment.
   - A11.6–A11.11: schemas registered as pending; D9 limits kept as the owner's numbers; the `init` heuristic stated as policy and not proof of humanity; lifecycle as independent axes; offline purge unknown; identity preserved on retire (D-IDENTITY); the one-shot dispatcher exception is not extended to donors.
2. **Owner decisions applied exactly, nothing invented.**
   - D-DONOR-PROVISIONING: separate component, no user-manager operation, binary name left pending, B disabled.
   - D-WIRE-NAMES: `swarma-bridge invitation/1`, `join/1` and `credential-package/1`. `swarma-rb/1` and `swarma-workplace-exec/1` stay design shorthand with an explicit open owner question on wire ownership. No migration is claimed.
   - The contract table's open items (D5 helper extension, D1 A availability) are resolved or listed as open without a recommendation being turned into a decision.
   - No change was needed for D-GOALS, D-PATH, D-EXECUTOR or D-RESUME, and none was made.
3. **Diagrams.** All five show the final names and flows: the separate donor provisioner (not user-manager), the host/user CA flow, the A/C/B branches, the DNSSEC/pin branch with no fallback on failure, independent lifecycle axes. Diagram text was read, not rendered (no renderer on this host). Syntax looked sound on reading.
4. **Alignment outcome** `TASK-261011-2lwcrk_alignment.md`: 15 rows (10 A + 5 B2). Spot-checked rows are true against the diff. Its validation paragraph describes the pre-handoff state; the rev2 validation log supersedes it.
5. **Hygiene.** No Cyrillic in added lines; no local paths, board-resource links or session links; the TASK ids in CIP-0008 are pre-existing. `git diff --check` base..candidate: exit 0 (run directly). Rev2 validation log: 73 schemas, 1294 vectors, 689 tests OK, Go compile-only (R223), exit 0.

## Reconciliation with reviewer A
A-N1 (note, P3): schemas and ownership pending, diagrams not rendered, no runtime qualification. Agreed; it is a disclosed limit and not a defect. A raised no P0/P1.

## Notes (P2/P3, not blocking; tb-R226)
- N1 (P2): CIP-0008 dropped the "Owner, verbatim" attributions and the original-language quotes from the admission/strict-qualification/Pi-OpenCode bullets (English-only rule). The decisions stay in the text, but their owner provenance is now only in git history.
- N2 (P3): `protocol/remote-worker.md` is not yet listed in the README documents table. Add it when the schemas are published.
- N3 (P3): "swarma-bridge" is assigned as the owner of the join/admission objects. This follows D9 ("the only public part of the bridge") but it is a naming assumption; the protocol file flags the rest as pending.
- N4 (P3): "Open applications require DNSSEC" is an inference from "pins only for a declared no-DNSSEC invitation". It is consistent with the table and stated in the CIP, the protocol file and both diagrams.

Not run on this host: renderers or diagram syntax tools, provider qualification, the full unit/draft-corpus gate beyond the attached validation log.
