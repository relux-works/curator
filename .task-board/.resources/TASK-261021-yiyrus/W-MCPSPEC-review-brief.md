# Review brief: TASK-261021-yiyrus, Draft CIP-0022 "MCP server declarations in the project Skillfile" (curator-spec)

Review the branch `cip-mcp-in-project-skillfile` at the exact head in the request, against `W-MCPSPEC-brief.md` and `mcp-design-agreed-v1.md` (board resources).

| Surface | Attack |
| --- | --- |
| Owner decisions | Independence: no other implementation is named and no name is reserved for one (grep the diff). `install` is defined neutrally. The environments rule: absent means every environment the manager can write; an unwritable environment gives `skipped: unsupported` per pair; an unknown id is invalid. Two stages: `package` first, then inline as a CIP-0002 contribution with digest-bound approval. Status Draft. |
| Grammar | Schema 3 is a strict superset of 2, project scope only, global rejects 3. NAME grammar. The shorthand expands exactly to the object. `server` is the closed `agent-mcp.json` object: check member by member against environments §2.2, nothing more and nothing less. URL rules and the secret detector. Digest equality of the shorthand and its expansion. |
| Security | No process starts at install. A token in a URL is a blocking finding. Approval binds the normalized digest, and an edit means re-approval. The allowlist key is identity plus NAME. No path from repository bytes to authority without approval. |
| Consistency | Agrees with CIP-0002 (contribution, ceilings, the agent approval limits), manager §12 and the Skillfile spec. Refusal codes are unique and follow the naming conventions. |
| Completeness | Compatibility (rc.7 refuses 3), conformance vectors, open owner questions, each with a recommendation. |
| Repo rules | Prose rules, links, the validator (rerun it), CIP numbering and index, no personal paths or names. |
