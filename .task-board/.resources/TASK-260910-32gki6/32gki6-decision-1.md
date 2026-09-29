# TASK-260910-32gki6 — orchestrator decision (THE ONLY CURRENT INSTRUCTION, with 32gki6-sec-brief.md)

No protocol decision is needed; the rc.13 text settles it. environments.md §9 (the env resolve steps, ~L2880): "recomputes every named store
entry's tree hash from its bytes and requires equality with the pin — for a `git` member the tree object identity of the pinned commit, for
a `path` or `local` member the `state_sha256` … missing hashes never count as passed". §4 (~L727) says the same. So: OPTION 2 —
resolve the pinned commit's tree id from the LOCAL git object database Curator already keeps for that source (its checkout/snapshot cache —
read-only, no fetch, no network) and compare it with the store entry's recomputed tree hash; when the commit object is not available
locally, the hash is missing → the entry is environment_store_untrusted (fail closed, diagnostic names the member and "pinned commit object
unavailable"), never a pass and never a network fetch during resolve. The orchestrator's gatefix-1 phrase "must not need the source repo"
was too strong: it meant no re-extraction into the store and no network — reading the local object DB for the commit→tree mapping is what
the spec prescribes. Do NOT version the lock.
Test fixtures that used fake pins without any object DB (e.g. TestReservedSkillName's dddd… pin) must now provide a real local commit whose
tree matches the store entry, because the spec runs pin verification BEFORE the checks those tests target — adjust fixtures, keep their
assertions (reserved-name, unreadable-manifest …). Add rows: tampered store entry → untrusted; missing commit object → untrusted; match →
passes. Mutants: pin check skipped; missing object treated as pass — killed, real exit codes.
Set status development; update results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
