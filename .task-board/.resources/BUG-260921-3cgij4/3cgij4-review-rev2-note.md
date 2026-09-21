# Review note for BUG-260921-3cgij4 revision 2 (orchestrator, binding) — curator-spec corpus fixture

Repository curator-spec. Brief 3cgij4-brief.md; revision 2 = revision 1 unchanged (revision 1's
local gate was killed by a host exec-stall window; rev2 gate green: `make validate` — 62 schemas,
1119 vectors, tools tests). Patch: `conformance/draft-sources-v1/index.json`,
`schema-cases/install-marker-v5/valid.json`, `schema-cases/install-marker-v5/valid-no-builds.json`.

Judge: (1) the normative rule (`protocol/skillfile-sources.md` ~268: top-level `build_source`
required exactly for active local go-v1 commands and absent otherwise) is untouched and both
fixtures now satisfy it — `valid.json` carries a real active local go-v1 build whose shape matches
the marker v5 closed schema (package/lock_sha256 etc.; compare with an existing valid v5 marker
with a local go-v1 command, and with `schemas/…install-marker-v5` — the fixture must be
schema-valid AND prose-valid); `valid-no-builds.json` has no builds and no build_source; (2) the
index entries (ids, `valid` flags, comments citing the rule) are consistent and the case count
change is intentional; (3) was a NEGATIVE fixture for the exact defect (empty builds + top-level
build_source → invalid) feasible in this corpus's pattern? If the corpus pins invalidity only by
JSON Schema, say so and record it as a bound (the cross-field rule is not schema-expressible);
(4) the curator follow-up named in results.md (pin promotion + bound→driven row) is precise;
(5) nothing else in the patch. Rerun `make validate` yourself in a disposable clone (retry once
on a host stall). Record exactly one verdict: accept_cr(BUG-260921-3cgij4, revision=2,
evidence=<your outcome resource>) on ACCEPT, or changes_requested with file:line.
