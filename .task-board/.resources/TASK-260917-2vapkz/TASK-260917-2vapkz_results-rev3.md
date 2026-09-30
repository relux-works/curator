# TASK-260917-2vapkz results — rev3 (rework 2)

## Blocking finding from rev2: generator deleted hand-authored released fixtures
- tools/generate-vectors/main.go writeSchemaCases: removed `must(os.RemoveAll(caseRoot))`. The generator now only writes the files it emits; it never prunes conformance/v1/schema-cases.
- Restored all 25 fixtures (agent-environment-marker-v1/* 12, launch-env-fragment-v1/* 13) with `git checkout 4ad8042b -- <paths>`.
- `git diff --diff-filter=D --name-only 4ad8042b -- conformance` → prints nothing (0 deletions).
- Regenerated with the generator only (`make regenerate`, exit 0); conformance/v1/manifest.json picked the restored fixtures back up.

## Regression test + mutants
- New `tools/generate-vectors/schema_cases_preservation_test.go` TestGeneratorPreservesHandAuthoredSchemaCases: copies the repo to a temp dir, adds a hand-authored sentinel schema case, runs the real generator entry point (`go run . -root <copy>`, i.e. main → writeSchemaCases), and asserts two released fixtures and the sentinel survive byte-identical. Green (exit 0).
- Mutant A (delete-only, reinstate full RemoveAll(caseRoot)): test FAILS ("regeneration deleted hand-authored fixture agent-environment-marker-v1/valid-no-composition.json").
- Mutant B (narrowing, RemoveAll only caseRoot/launch-env-fragment-v1): test FAILS ("... launch-env-fragment-v1/valid-file-channels.json"). Source restored after each mutant.

## Gates (real exit codes)
- `go test ./tools/... -count=1` → 0
- `python tools/validate.py` (venv with jsonschema; system python3 lacks jsonschema → exit 1 ModuleNotFoundError, env issue) → 0, "validated 72 schemas and 1253 vector files"
- `make regenerate-check` → 0
- `python -m unittest test_validate -k released_schema` (byte-freeze guard) → 0
- NOT rerun this round: the full python unittest discovery (~20 min) — accepted from rev2 evidence; only generator + guard changed scope here.

## Follow-up (not done here)
- The 25 agent-environment-marker-v1 / launch-env-fragment-v1 schema-case fixtures are not produced by the generator; consider wiring them into index.json (or retiring them explicitly) in a separate task.
