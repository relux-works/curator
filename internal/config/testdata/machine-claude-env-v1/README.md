# Machine Claude environment vectors

The four JSON files are unmodified copies from curator-spec commit
`aaa843965c4caaf4ededf45a0c269dddc50e0472`:

- `conformance/machine-claude-env-v1/cases.json`
- `conformance/machine-claude-env-v1/plants.json`
- `schemas/machine-claude-env-v1/claude-env.schema.json`
- `schemas/machine-claude-env-v1/projection.schema.json`

The declaration driver uses the selected machine-file loader. Contribution
vectors use the strict contribution transport parser. These tests do not start
Claude, exercise a launcher consumer, or qualify native output-limit behavior.

The driver validates each containing file with a valid replacement member
before exercising the vector. An unrelated carrier defect is classified as
`invalid-containing-file`, never as a detected member refusal. Malformed and
duplicate declaration JSON is classified separately from semantic member
errors: the file loader must report a neutral, path-bearing malformed-JSON
diagnostic. The landed proposal §§4 and 6 map that file-level refusal to the
member verdict for these vectors; the five Q1 claim edits are copied verbatim.
The containing-carrier token regressions separately exercise the integer-token
rule through the real loader and auth editor. These copies do not run the
specification's separate containing-cases.json corpus.
