# Review note — TASK-260924-291k0q skillfile-sources independence gate (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `291k0q-brief.md`. Candidate touches: ci.yml, CHANGELOG, conformance/skillfile-sources-v1/manifest.json,
protocol/skillfile-sources.md, release/1.0.0-rc.13.json, schemas/skillfile-sources-v1/README.md, tools/{verify,test}_skillfile_sources_independence.py.
1. BLOCKING unless justified: release/1.0.0-rc.13.json is the metadata of a RELEASED version (tag v1.0.0-rc.13) — released records must stay
   byte-identical after release (compare with the tag). Any change there is a finding.
2. protocol/skillfile-sources.md: normative text of the accepted suite — only an informative-only marker for the post-rc.10 citation (#96
   review N1) is acceptable; any normative change is a finding. Name the exact diff. The suite manifest.json change must follow only from
   that and be generator output.
3. The gate: $refs from schemas/skillfile-sources-v1 into schemas/v1 resolve to definitions byte-identical at v1.0.0-rc.10; prose citations
   of post-rc.10 clauses rejected unless informative-marked; wired into Specification CI; both mutants fail the gate (reproduce); passes on main.
4. CHANGELOG under Unreleased. accept_cr or changes requested with file:line. No LOGBOOK.md.
