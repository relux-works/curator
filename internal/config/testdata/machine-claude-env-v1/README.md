# Machine Claude environment vectors

The four JSON files are unmodified copies from curator-spec commit
`bd2e8252990872612cf76a358a1f874281a4711b`:

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
diagnostic. Mapping that file-level result to the pinned member verdict is
pending source specification clarification; no fixture bytes are changed.
