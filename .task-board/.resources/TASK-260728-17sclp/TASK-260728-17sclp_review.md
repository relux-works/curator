# TASK-260728-17sclp review verdict

Verdict: **CHANGES REQUESTED**

Route: `to-dev`

## Findings

### 1. Security-sensitive Git URL, identity, tag, and ref grammar has false accepts

The normative contract requires no `.`/`..` URL components, ASCII-only SSH
repository paths, rejection of SSH shell metacharacters/non-ASCII bytes,
canonical lowercased identities with one trailing `.git` removed, tags bounded
to 255 UTF-8 bytes, and revision width matching the effective object format.

The current `repositoryGit`, `gitRefName`, `networkSourceIdentity`, and
`repositoryStructuredRef` definitions plus `validate_effective_source` do not
enforce those rules. The same Draft 2020-12 registry and semantic validator used
by `tools/validate.py` accepted all of these adversarial examples:

- `ssh://git@example.com/répo.git`
- `git@example.com:repo;touch`
- `https://example.com/org/../repo.git`
- `ssh://git@example.com/./repo.git`
- a tag of 100 `界` characters (300 UTF-8 bytes)
- a SHA-1 effective source with a 64-hex structured revision
- canonical identities `example.com/org/../repo` and
  `example.com/org/repo.git`

This conflicts with `protocol/core.md` section 6.3 and the task AC's exact
URL/ref/tag and object-format requirements.

Required rework: close the raw transport grammar, canonical identity grammar,
UTF-8 byte bound, and effective-object-format/ref-width relationship in the
schema and/or deterministic semantic layer. Add generated valid and invalid
cases for HTTPS Unicode, SSH ASCII, dot components, SSH metacharacters and
non-ASCII, canonical host/path spelling, the 255/256 UTF-8 byte boundary, and
SHA-1/SHA-256 structured-ref width matching.

### 2. Legacy schema-1 guard does not reject all reserved schema-7 surface

`protocol/core.md` states that schemas 1 through 6 reject
`build_repositories`, `repository`, `target`, and `go-repository-v1`.
`validate_wire_semantics` checks only `build_repositories` and a
`go-repository-v1` driver inside `commands`.

Because manifest schema 1 intentionally retains
`additionalProperties: true`, the validator accepts each of these top-level
additions to the frozen valid schema-1 fixture:

- `"repository": "repo"`
- `"target": "tool"`
- `"driver": "go-repository-v1"`

Required rework: preserve frozen schema-1 bytes while adding semantic
version-selection guards for every reserved schema-7 field/identifier at the
relevant locations. Add deterministic regression coverage for both manifest
filenames and versions 1 through 6.

### 3. Skillfile.dev schema 2 makes an optional extension mandatory

`protocol/core.md` says schema 2 retains `substitutions` and **may add**
`build_repository_substitutions`. The schema currently requires
`build_repository_substitutions`, so the schema-1 valid fixture with only
`schema_version` changed to 2 is rejected.

Required rework: make the external-repository substitution map optional unless
the normative prose is deliberately changed through its owning contract task.
Add a valid schema-2 case with only ordinary substitutions and cases for empty
and populated external substitution maps.

### 4. Generated cases do not cover every new semantic branch

The generated index lacks invalid cases for several semantic paths implemented
in `tools/validate.py`, including receipt containment, unsubstituted
declared/effective mismatch, effective object-format versus structured revision
width, marker declared/effective mismatch, duplicate driver assertions, and
driver-platform assertions outside the top-level platform set. Some paths have
unit-only coverage and duplicate-driver rejection has no direct regression.

Required rework: generate deterministic valid/invalid vectors for every new
structural and semantic branch and make the Go inventory test require them.

## Independent validation

Green checks:

- accepted schema-7 prose files match their accepted source byte-for-byte;
- frozen manifest schemas/cases 1 through 5 match pinned HEAD
  `57c1f56846d221ecc55786bd3c2467ec32f11730`;
- accepted hashes match for manifest v6, receipt v1, marker v2, and claim v2;
- `go test -count=1 ./tools/generate-vectors`;
- `go vet ./tools/...`;
- `gofmt -l tools/generate-vectors` (no output);
- Python unittest discovery: 13 tests;
- `tools/validate.py`: 42 schemas and 263 vector files;
- `make validate`;
- two direct generator passes were byte-idempotent:
  conformance digest stayed
  `a5fb6ef4880959e4cbd7683b473b878208acdfa4834ccdf1e816886aa97e140f`;
- Git status digest stayed unchanged across regeneration;
- `git diff --check`, no staging, and pinned HEAD checks passed.

The green suite therefore demonstrates deterministic generation and general
integrity, but it does not detect the acceptance-criteria failures above.
No product code was modified during review.
