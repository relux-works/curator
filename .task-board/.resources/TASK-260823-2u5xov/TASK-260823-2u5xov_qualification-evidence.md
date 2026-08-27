# Curator schema-8 qualification — green with proven consumption

- PR 37 (Consume the schema-8 families instead of merely publishing them) merged as a3abcf3, all lanes green pre-merge; family-removal proof attached earlier (removing a schema-8 family from the root fails the suite).
- Qualification dispatch: run 32689488293, workflow_dispatch with candidate_ref 6001dc33281b94a4ec7442ab15278550dd0f51d9 and manifest sha256 803918bf..., on main at a3abcf3. Conclusion SUCCESS, zero non-green jobs — Candidate suite green on ubuntu, macos, windows WITH the consumption assertions in place.
- This is the curator qualification evidence the impact-analysis landing order requires (step 4) for the module-roots family; the script-worker family remains consumption-covered but behaviorally fail-closed (script_execution_policy_unsupported) pending STORY-260822-2h0v9j.
- Dispatch executed by the orchestrator after the producer run timed out post-PR; producer authored the coverage and the removal proof.