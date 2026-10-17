# Frozen fragment schema and successor corpus

The v3 schema is retained verbatim from curator-spec `d373078a`.
V1, v2 and v4 schemas and the v4 cases were copied verbatim from curator-spec
main `1059b94b4135089d842bf017070b8ab47d79a536`:

- `schemas/v1/launch-env-fragment-v{1,2,4}.schema.json`
- `conformance/environments-muse-v4/schema-cases/`

`TestMuseFragmentV4PublishedCases` selects each indexed frozen schema and
also checks the one-parent/suffix relation through the production boundary.
`TestMuseCLIFragmentAndFakeLaunch`, `TestMuseCLIStorePolicy` and
`TestMuseRootContextVectors` drive current production emission. The inert
child records environment/argv only; it does not qualify Muse authentication,
refresh, hosted composition or context exclusion on a real vendor release.
