# TASK-260728-17sclp review cycle 3

Verdict: **ACCEPTED**

Route: `done`

## Review basis

- Reviewed `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree` read-only at pinned HEAD `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Run `RUN-260727-c17962` is not goal-bound and had no operator directives at the verdict checkpoint.
- The five accepted schema-7 prose inputs match `TASK-260728-1a52au` byte-for-byte: `SECURITY.md`, `profiles/manager.md`, `protocol/core.md`, Decision 0004, and Decision 0005.
- Relative to the retained cycle-2 clean corpus, rework 2 changes only the generated manifest/index and adds the eleven requested `install-marker-v3` cases.

## Cycle-2 finding closure

The eleven requested marker-v3 cases exist, are emitted by `installMarkerV3SchemaExamples`, are indexed against `install-marker-v3.schema.json`, and are required with the correct valid/invalid classification by `TestGeneratedSchemaV7CasesCoverEveryWireBranch`.

Valid marker-native coverage:

- empty builds with no top-level `build_source`;
- network tag and branch substitutions;
- SHA-1 and SHA-256 structured revision substitutions;
- an unsubstituted SHA-256 external record;
- an untagged external record.

Invalid marker-native coverage:

- local substitution with a network effective identity;
- network substitution with an operator-local effective identity;
- SHA-1 effective format with a 64-hex structured revision;
- SHA-256 effective format with a 40-hex structured revision.

Direct schema-plus-semantic validation classified all eleven as expected. The four invalid cases reached the intended marker semantic errors rather than failing through an unrelated fixture.

## Independent validation

- Tool readiness: task-board, Git 2.50.1, ripgrep 15.2.0, Go 1.25.5, Make, and the task-local Python environment executed successfully.
- `python -B tools/validate.py`: exit 0; 42 schemas and 400 vector files.
- Python unittest discovery: exit 0; 15 tests.
- `go test -count=1 ./tools/...`: exit 0.
- Focused schema-7 inventory, legacy frozen-byte, receipt-v1, marker-v1/v2, claim-v1/v2, and manifest-v1-v6 compatibility tests: exit 0.
- `go vet ./tools/...`: exit 0.
- `gofmt -l tools/generate-vectors`: no output.
- `go build` for the generator to task-local temporary storage: exit 0.
- `make validate` with the pinned Python environment and bytecode disabled: exit 0.
- Clean-from-empty alternate generation matched checked `fixtures`, `expected`, `vectors`, `schema-cases`, and `manifest.json` bytes. A second pass retained digest `bb6fa3ea89a2f4722158ffb99e76b9e543a58f90c0bd6f97d924c5a78ec6d2dd` across 401 conformance files and still matched the checked corpus.
- Adversarial replay: 103 selected cases matched expected classification, including all 84 schema-7-only rejection guards across both manifest filenames and versions 1 through 6 plus 19 URL/ref/identity/substitution/claim boundary probes.
- `git diff --check`, clean real index, pinned HEAD, and repository artifact-cleanliness assertions: exit 0.

No correctness, compatibility, deterministic-generation, coverage, architecture-fit, scope, lint, or validation finding remains. No product code was modified during review.
