# THE ONLY CURRENT INSTRUCTION — TASK-261004-3pvg2k: introduce Curator Improvement Proposals (CIPs) in curator-spec (developer, docs)

Operator decision 2026-10-04: design proposals for the protocol and its implementations are written as CIPs and kept in curator-spec `cips/`. Research evidence stays in the implementation repo's `.research/`.

Create, docs only:
1. `cips/README.md`:
   - purpose;
   - how a CIP relates to issues, `decisions/` records and normative changes. A CIP proposes and argues; adoption produces Decision record(s) plus normative prose, schemas and vectors under GOVERNANCE.md;
   - numbering: `CIP-NNNN-<slug>.md`, sequential, never reused;
   - statuses: Draft, Review, Accepted, Rejected, Withdrawn, Superseded. The operator or maintainer decides;
   - lifecycle and required sections;
   - an index table.
2. `cips/TEMPLATE.md`: copy the attached `cip-template.md`, which is the binding section list.
3. `cips/CIP-0001-curator-improvement-proposals.md`: the process itself, with Status: Accepted (operator, 2026-10-04).
4. Index rows. CIP-0001 is Accepted. These are "In preparation":
   - CIP-0002 project context in managed launches;
   - CIP-0003 Claude managed-home credential modes;
   - CIP-0004 shell hook without sourcing and PATH append;
   - CIP-0005 audit backends and CLI secret transport;
   - CIP-0006 legacy provider settings and MCP opt-outs.
5. A short "Proposals" section in GOVERNANCE.md and a link in README.md.

No change to protocol/, schemas/, conformance/ or release/. The manifest digest stays 6f832d81…. Run `make validate`, or at least the link and format checks, and record real exit codes; jsonschema requires a venv. Never edit LOGBOOK.md. No CHANGELOG entry is needed, unless the repo convention requires docs entries; if it does, add one Unreleased line.
Then `task-board handoff TASK-261004-3pvg2k --role developer` and END YOUR TURN.
