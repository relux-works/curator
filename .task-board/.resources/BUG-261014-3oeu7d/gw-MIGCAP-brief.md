# Producer brief: BUG-261014-3oeu7d, migrated global skill capabilities

Read the bug description. Field report: on another operator's machine the default-profile migration of a legacy global skill is blocked by `audit.capability.network-undeclared` for URL literals in its scripts, because `strictAuditMember` (internal/envprofile/envprofile.go) builds the audit subject with `capabilities.ImplicitNone()`. A declaration in the skill's own `agent-skill.json` is never seen on this path. Every spawn that needs a Curator child profile is blocked there.

**Normative source:** curator-spec `protocol/environments.md` §9.1 (strict audit: a skill member is audited with its own package manifest capabilities) and §9.4 (migrated global skills; the refusal names skill, host and file and points to `capabilities.network` in `agent-skill.json` or `csk-skill.json`). Read both on curator-spec main yourself. Where this brief and the spec disagree, the spec wins: stop and report.

**Input (untrusted, use critically):** `BUG-261014-3oeu7d_candidate.patch`, a candidate fix written outside this lane against base 34a89308 (sha256 3a121d0d1317d5e29d5808af38fccdf5eaa752ae730152a1e48cdec5d9c5a27e, 2 files). Review it as you would any external code: check that it implements §9.1/§9.4 exactly, that it does not widen admission elsewhere (other callers of `strictAuditMember`), that a schema-2 manifest without capabilities and a malformed or unreadable manifest refuse, and that the audit subject's schema version matches what the manifest actually declares. Adopt, fix or replace it.

**Required behaviour and controls, all through the real migration entry:**
- Admitted: an exact host and a glob declared in the skill's manifest let the migration publish, and the skill is in the lock.
- Refused with no profile publication: no manifest; `network: none`; another host declared; schema without capabilities; malformed `capabilities.network`; an undeclared neighbouring host next to a declared one; a revoked source with a covering declaration; a failed static canary with a covering declaration.
- The refusal text names the skill, the host, the file, and the remedy (§9.4). The finding is never downgraded or declared on the skill's behalf.
- A narrowing mutant (back to `ImplicitNone`) must redden exactly the admitted cases; a mutant that drops the manifest-error refusal must redden the malformed case.
- Non-skill members and other audit paths keep their current behaviour; add a positive control that shows it.

**Validation:** the full gate, `sh scripts/remote-gate.sh` (Linux CI runs on the self-hosted runners with job-private Go state). If an unrelated test fails, record it with evidence and whether it reproduces on clean main.

**Public repository rules:** no host names, people, local paths or other operators' details in code, tests, commit text, LOGBOOK or board resources. Describe the field report as "a legacy global skill whose scripts reference hosts".

**Handoff:** Change Request with summary, a mapping of every control above to its test, the mutant results, the gate run, and how the candidate patch was used.

## Decision on the blocked question (orchestrator, 2026-10-17)

Your probe is right, and the brief's "actual manifest schema" requirement was wrong; the spec wins. Environments §9.4 requires the network-undeclared finding with skill, host, file and remedy even for a missing manifest or declaration. Forwarding schema 0 or 2 to `audit.Gate` would turn that into `require_pin` and hide the finding, contradicting §9.4. So:

- Keep the profile audit's existing schema floor 3 for skill members with no manifest or a pre-capability manifest; forward a modern manifest's schema version exactly. Do not change `internal/audit` and do not widen the generic skill pipeline.
- Load the skill's own capabilities from its manifest (§9.1); a malformed or unreadable manifest blocks.
- State the floor and its reason in a code comment and in the LOGBOOK entry, and add a test that shows a manifestless legacy skill with a URL gets the network-undeclared finding with the remedy (not `require_pin`).
- Everything else in the brief stands: controls through the real migration entry, no publication on refusal, non-skill members unchanged, narrowing mutants, full remote gate.

Continue from your regression tests in the Story workspace.
