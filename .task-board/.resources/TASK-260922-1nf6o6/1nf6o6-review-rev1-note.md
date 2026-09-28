# Review note for TASK-260922-1nf6o6 revision 1 (orchestrator, binding) — F-S1 marker credential record

Read the brief `1nf6o6-brief.md` and `TASK-260922-1nf6o6_results.md`. Source: Decision 0017 choice 5
and the environments §7.4 "Credential record" paragraph. The deliverable is a NEW marker schema
revision (`agent-environment-marker-v2`) that admits the credential record on each passthrough entry
(`isolation`, `strategy`, `source_role`, `backend`, `backend_version`, `provenance`), with `path`
required only for linkable strategies (per-home-keychain under isolated and ambient record without a
path), schema 1 left byte-untouched, the new case family, the §7.4/§8.4.1 text, and the named curator
follow-up leaf.

Judge: (1) schema fidelity — v2 = v1 + the record, `additionalProperties:false` everywhere, and the
path conditionality really enforced (craft a linkless strategy WITH a path and a file-link WITHOUT
one; both must be invalid); (2) schema-1 cases unchanged and a v1 marker carrying the record is
invalid; (3) the case family is generator-produced (regenerate in a disposable copy and diff — a
hand-edited case is a finding); (4) the normative text states the upgrade rule (a schema-1 marker is
never rewritten just to add the record) consistently with §12.4/§12.5; (5) the curator follow-up leaf
is named with the exact fields; (6) `make validate` green (reuse the handoff evidence).

Record exactly one verdict: `accept_cr(TASK-260922-1nf6o6, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
