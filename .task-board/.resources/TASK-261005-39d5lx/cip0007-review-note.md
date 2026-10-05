# Reviewer instruction — TASK-261005-39d5lx: Draft CIP-0007 (manager-provisioned CLI tools, curator-spec #108)
Cross-provider review: the producer was muse max; the reviewer is codex sol high. This is a security-relevant design (trust in downloaded binaries), so tb-R195 applies.
Check:
1. The template shape and process rules are followed; Status is Draft; there is a README index row.
2. The options A/B/C are fairly compared, with the trade-offs stated. The recommendation (if any) is argued, not assumed.
3. The trust and threat analysis is sound: author-controlled downloads, registry compromise, signature and key rotation, mirror integrity, offline and air-gapped behaviour, and downgrade and rollback attacks. Every security claim cites a spec § or is marked as a proposal.
4. Interactions with the skillfile schema, lock, install transaction, conformance and CIP-0005 are correct against the current spec text (cite §).
5. The decisions-needed list is concrete and operator-answerable.
6. No normative edits, no CHANGELOG or LOGBOOK edits. No organisation, team, client or personal names; count them with grep and report the count. No personal paths or hosts.
7. `python -B tools/validate.py` exits 0.
accept_cr, or request changes with numbered findings. Do not edit files.
