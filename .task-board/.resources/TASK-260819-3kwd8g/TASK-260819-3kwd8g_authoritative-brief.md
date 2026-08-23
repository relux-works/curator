# Authoritative implementation brief

Work in /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-agent-skill on branch codex/portable-verified-assurance. The board lives at /Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-skill/.task-board.

The current published candidate is 1.0.0-rc.6. It already defines portable manager-worker-v1, explicitly does not claim six hardened guarantees, and says a future hardened profile requires a new execution-policy identity and claim schema. Preserve every historical release byte and meaning.

Use /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-draft-hardened and the draft snapshot in main only as research input. Reconcile it with this binding decision:
- portable is the default CLI-only assurance mode and emits only capabilities actually established;
- verified is explicit, provider-backed, and fail-closed with no silent downgrade;
- the common provider contract must be platform-neutral for macOS, Linux, and Windows;
- mode, provider contract, capabilities, permits, receipts, claims, cache, and checkpoints cannot alias;
- signed host provider binaries are separately installed trusted components and never skill-vendored artifacts;
- all vendored compiled artifacts remain prohibited;
- no platform verified implementation or conformance claim is released by this task.

Determine the smallest coherent normative schema and conformance change. Because rc.6 cannot express verified claims and identity separation, prepare the next release-candidate surfaces rather than mutating rc.6. Include decision record, normative protocol and profile text, schemas, valid and invalid vectors, generated manifest, release metadata, changelog, operator guidance, migration compatibility, validators and release gates. Run all project validation and regeneration checks. Do not tag or publish; TASK-260819-2tr2rh owns release publication.
