# Unreleased draft source-policy fixtures

Copied without edits from curator-spec at 871d11bcdfd240a6260d0722503bdd1642a8fce8
(the files are unchanged since the source contract landed in a4fcaf02):

- `conformance/draft-sources-v1/schema-cases/source-policy-v1`: all 5 cases.

`TestDraftPolicyPublishedSchemaCases` drives every published case through
`config.ParseSourcePolicy`. Filename labels are the published schema oracle.
Go tests do not execute a JSON Schema engine. Semantic cases (exact key
identity, endpoint canonicalization, distinct URLs, provider refs, pin and
fallback selection, resolution without network) additionally drive the
loader and `ResolveRepositoryEndpoints` in sourcepolicy_test.go.

This evidence covers machine-policy loading and attempt planning only. It
does not establish fetch behavior, failure classification, credential
resolution, installation, or cross-platform filesystem behavior.
