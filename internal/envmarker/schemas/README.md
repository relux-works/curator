These files are copied without modification from curator-spec commit
`d6c40195962d9fa3ccafa2c2001c2312c6c5e086` (Apache-2.0). They are the complete
marker-7 schema and its transitive reference closure, resolved by canonical IDs.
The production reader validates the original JSON before typed decoding so
missing, null and forbidden members cannot disappear during decoding. Older
adopted marker versions retain their existing reader rules.

`../testdata/seed-c/node.json` is the full node marker control from
`conformance/local-setup-token-v1/controls.json`, with version 7 and a C seed
record, as required by `protocol/codex-seed-c.md`. Tests exercise both authority
branches. Parsing a node carrier preserves its identity; this manager refuses
node activation and credential migration until that separate runtime is adopted.
