# Handoff (operator, 2026-09-24): make Skillfile schema 2 default-on — binding for TASK-260924-1aa9wb

Context: skillfile-sources-v1 (Skillfile schema 2) and repository-transport revisions 1-2 move from draft to accepted. The spec change
lands in relux-works/curator-spec from the other manager's side: "unreleased" removed, draft schemas and `conformance/draft-sources-v1`
vectors move to the released namespace, and a global-scope rule is added (skillfile-sources covers the PROJECT scope; global/profile
scope follows environments.md profile locks §9.4).
Curator's work (checked on origin/main 1511b345):
1. REMOVE the operator switch `CURATOR_DRAFT_SOURCES_V1` entirely (`internal/install/draftsources.go:15-31`, `EnvDraftSourcesV1`,
   `DraftSourcesEnabled`) and make the schema 2 lane the default reader path. Call sites: cmd/curator/draft_help.go:90,
   draft_status.go:41, main.go:1181, project_resolve.go:67,85; internal/install/*; internal/envprofile (`DraftSourcesV1` policy fields in
   envprofile.go and overlays.go).
2. Schema 1 stays byte-identical: a `schema_version: 1` Skillfile NEVER enters the schema 2 path. Only the refusal of
   `schema_version: 2` goes away.
3. Remove draft wording from diagnostics, help and docs: docs/draft-source-expansion.md (its header "not an enabled schema-2
   installation path" is stale), docs/draft-transport-resolution.md, docs/cli.md, README, docs/troubleshooting.md, CHANGELOG.
4. (Later leaf) move the conformance pin from the draft vectors to the released suite once curator-spec publishes it.
5. Global scope stays OUT of scope: global install keeps "no draft lane" (`internal/install/global.go:395`) behaviour.
