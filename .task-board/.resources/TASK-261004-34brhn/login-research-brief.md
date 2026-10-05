# THE ONLY CURRENT INSTRUCTION — TASK-261004-34brhn: can a Claude Code login be passed into a macOS managed home? (researcher)

Operator question 2026-10-04: a managed home (CLAUDE_CONFIG_DIR=~/.curator/environments/<profile>/claude_code) currently needs its own login on macOS. If there is a way to pass the operator's login, define modes: pass the login vs use a new per-home login.
Investigate, from Claude Code's shipped bundle/docs and safe scratch probes (no real credentials):
- the long-lived token path (`claude setup-token` → CLAUDE_CODE_OAUTH_TOKEN), apiKeyHelper, ANTHROPIC_API_KEY precedence;
- where a CLAUDE_CONFIG_DIR login is stored on the current release (Keychain service suffix vs .credentials.json under the config dir; Decision 0017 Q1 says 2.1.273 wrote a file);
- refresh behaviour for each path, and what survives across launches;
- how a launcher could inject a token without it entering argv, logs or the fragment (env from a protected store; macOS Keychain read by curator-run).
Spec constraints: environments §7.4, Decision 0017 (Q1 gate needs a disposable-account experiment; Q7 copy refused). Propose the mode names, config shape and the spec revision needed.
## Operator decision (2026-10-04, binding)
The operator likes ALL THREE mechanisms: (1) the long-lived `claude setup-token` passed as CLAUDE_CODE_OAUTH_TOKEN, (2) apiKeyHelper, (3) the per-home Keychain item / .credentials.json behaviour of a separate home. The ideal design gives the operator CONTROL: a per-profile (and per-environment) credential mode with all of these available, e.g. `token` (from a protected store, never argv/logs/fragment), `helper` (apiKeyHelper command), `per-home` (separate login, persists in the home), plus `shared-file` only where verified safe. Define the config shape, defaults, precedence, refresh/expiry handling, rotation, what curator-run does at launch, how each mode is verified without exporting secrets, and which Decision 0017 answers must be revised (Q1, Q7).

## Output format (operator decision 2026-10-04)
Write the result as a **Curator Improvement Proposal draft** using the attached `cip-template.md`, saved as `.research/261004_CIP-NNNN-<slug>.md` (NNNN given below). CIPs will live in curator-spec `cips/`; the orchestrator publishes the draft there after review. Put raw evidence in a companion `.research/261004_CIP-NNNN-<slug>_evidence.md`.

CIP number: CIP-0003 (slug: claude-managed-home-credential-modes).

## Research rules (binding)
Read-only research: no product code changes. Output ONE document under .research/ (named <YYMMDD>_<slug>.md) plus evidence, attached as task outcome. Cite file:line on curator main, the curator-spec rc.14 text, or a measured probe in a scratch HOME. Never read, print or copy any real credential, token or Keychain secret; never log in or out of the operator account. Never edit LOGBOOK.md. Board ids always with their titles. The public board is visible to anyone: no internal hostnames, personal paths or employer names.
End with: options (2-4), tradeoffs, a recommendation, open questions for the operator, and the spec/implementation leaves the recommendation implies.

Then `task-board handoff TASK-261004-34brhn --role researcher` and END YOUR TURN.
