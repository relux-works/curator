# THE ONLY CURRENT INSTRUCTION — TASK-261005-22yioq: publish CIP drafts 0002–0006 into curator-spec `cips/` (developer, docs; curator-spec)
The operator asked (2026-10-04) for Curator Improvement Proposals to live in curator-spec `cips/`, with the research evidence staying in each repo's `.research/`. The process is now on spec main (b0caf8db): read `cips/README.md`, `cips/TEMPLATE.md` and `cips/CIP-0001-*.md` first and follow them exactly.
Source drafts: on relux-works/curator main (77fabd45 or later), under `.research/`. Read them with `git -C /Users/administrator/Developer/ReluxWorks/curator/curator show origin/main:.research/<file>`:
- CIP-0002 project context in managed launches;
- CIP-0003 Claude managed-home credential modes;
- CIP-0004 shell hook without sourcing, plus PATH append (K3);
- CIP-0005 audit backends and CLI secret transport;
- CIP-0006 legacy provider settings and MCP opt-outs (curator#105).
For each one, write `cips/CIP-000N-<slug>.md` in the template's shape:
- Status: Draft. The operator has made no acceptance decision.
- Keep the substance and decisions-needed lists. Condense the evidence: link to the curator `.research/` file by repository-relative path and commit instead of copying long evidence tables.
- Keep open questions and alternatives.
Also add the five rows to the index in `cips/README.md`, if it has one.
Rules:
- No normative protocol or schema edits.
- Do not touch frozen or released files, CHANGELOG or LOGBOOK.
- No personal paths, host names or employer names: the repo is public.
- Run the spec's docs checks that apply (links/format, e.g. `make validate` targets for docs or the link checker) and record the exit codes.
Then `task-board handoff TASK-261005-22yioq --role developer` and END YOUR TURN.
