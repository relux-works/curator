# THE ONLY CURRENT INSTRUCTION — TASK-261004-hy8zmn: audit backends + CLI secret transport design (researcher)

DESIGN PENDING. (1) curator audit --publish --token puts the secret in argv (cmd/curator/main.go ~2120). (2) profiles/manager.md §7 audit backends are parsed but never invoked. Design: token transport (env var / --token-file rules), backend invocation for several agent environments (codex, claude, a command backend), the child-environment allowlist (strip *_TOKEN, *_KEY, SSH_AUTH_SOCK, GIT_*; Windows case-insensitive), canary, request cap, egress classification, failure semantics in strict/advisory mode, and what to do with a configured-but-unsupported backend. Reference implementation for comparison: public cocoaskills PR #142 (commit 0da153a). Include the spec sentences needed.

## Output format (operator decision 2026-10-04)
Write the result as a **Curator Improvement Proposal draft** using the attached `cip-template.md`, saved as `.research/261004_CIP-NNNN-<slug>.md` (NNNN given below). CIPs will live in curator-spec `cips/`; the orchestrator publishes the draft there after review. Put raw evidence in a companion `.research/261004_CIP-NNNN-<slug>_evidence.md`.

CIP number: CIP-0005 (slug: audit-backends-and-cli-secret-transport).

## Research rules (binding)
Read-only research: no product code changes. Output ONE document under .research/ (named <YYMMDD>_<slug>.md) plus evidence, attached as task outcome. Cite file:line on curator main, the curator-spec rc.14 text, or a measured probe in a scratch HOME. Never read, print or copy any real credential, token or Keychain secret; never log in or out of the operator account. Never edit LOGBOOK.md. Board ids always with their titles. The public board is visible to anyone: no internal hostnames, personal paths or employer names.
End with: options (2-4), tradeoffs, a recommendation, open questions for the operator, and the spec/implementation leaves the recommendation implies.

Then `task-board handoff TASK-261004-hy8zmn --role researcher` and END YOUR TURN.
