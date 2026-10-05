# THE ONLY CURRENT INSTRUCTION — TASK-261004-2iewnz: curator#105 design and implementation plan (researcher)

curator#105: preserve legacy provider settings and explicit MCP opt-outs in existing profile/overlay surfaces. It blocks infra-exit TASK-260927-367jb7. Read the issue (gh issue view 105 -R relux-works/curator), Decision 0014 (proposed, not adopted) and 0018, environments §§7-10, and the current profile/overlay code. Produce: the minimal design that solves #105 WITHOUT adopting all of 0014, the exact surfaces and fields, migration of existing installs, tests, and the implementation leaves with sizes. Flag anything that needs a spec decision.

## Output format (operator decision 2026-10-04)
Write the result as a **Curator Improvement Proposal draft** using the attached `cip-template.md`, saved as `.research/261004_CIP-NNNN-<slug>.md` (NNNN given below). CIPs will live in curator-spec `cips/`; the orchestrator publishes the draft there after review. Put raw evidence in a companion `.research/261004_CIP-NNNN-<slug>_evidence.md`.

CIP number: CIP-0006 (slug: legacy-provider-settings-and-mcp-optouts).

## Research rules (binding)
Read-only research: no product code changes. Output ONE document under .research/ (named <YYMMDD>_<slug>.md) plus evidence, attached as task outcome. Cite file:line on curator main, the curator-spec rc.14 text, or a measured probe in a scratch HOME. Never read, print or copy any real credential, token or Keychain secret; never log in or out of the operator account. Never edit LOGBOOK.md. Board ids always with their titles. The public board is visible to anyone: no internal hostnames, personal paths or employer names.
End with: options (2-4), tradeoffs, a recommendation, open questions for the operator, and the spec/implementation leaves the recommendation implies.

Then `task-board handoff TASK-261004-2iewnz --role researcher` and END YOUR TURN.
