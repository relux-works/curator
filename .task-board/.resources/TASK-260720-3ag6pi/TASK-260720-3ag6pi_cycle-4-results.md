# TASK-260720-3ag6pi review cycle 4 outcome

## Outcome

The cycle-3 workflow-drift finding is closed with a narrow three-file rework:

- `.github/workflows/ci.yml` now checks generated rc.6 release metadata;
- `.github/workflows/release.yml` now checks generated rc.6 release metadata;
- `tools/test_validate.py` requires both workflow scopes to equal the Makefile's
  complete generated-file inventory.

All previously accepted protocol, vector, generated, documentation, and release
metadata bytes are preserved.

## Old-green/new-red proof

On the retained rc.6 metadata-drift fixture:

| Command | Real exit | Interpretation |
| --- | ---: | --- |
| Old workflow scope without rc.6 metadata | 0 | Expected-green bug reproduced |
| Corrected scope including rc.6 metadata | 1 | Expected-red drift detection |
| Focused workflow-parity regression after parser correction | 0 | Guard passes on both corrected workflows |

The expected-red output identifies only the changed
`release/1.0.0-rc.6.json` `created_at` value.

## Integrated command ledger

| Command | Tree | Exit | Evidence |
| --- | --- | ---: | --- |
| `make validate` | reworked source | 0 | 42 schemas, 447 vector files, 60 Python tests, Go tests |
| `make regenerate` | independent copy 1 | 0 | raw log attached |
| `make regenerate` | independent copy 2 | 0 | raw log attached |
| Recursive/cmp byte comparisons | both copies and source | 0 | all five comparisons |
| `make regenerate-check` | byte-identical gate candidate carrying the 3-file delta | 0 | complete rc.5/rc.6 scope |
| `make release-check VERSION=1.0.0-rc.6` | retained clean product commit | 0 | release gate passed at `ddb181ca…` |
| `go vet ./...` | reworked source | 0 | no findings |
| `go test ./... -count=1` | reworked source | 0 | uncached Go tests passed |
| Go formatting assertion | reworked source | 0 | no unformatted files |
| `git diff --check` | reworked source | 0 | no whitespace errors |

## Release-gate provenance

`release_gate.py` rejects every dirty checkout. The operator directive forbade
staging and committing, so the rc.6 release gate ran on retained clean commit
`ddb181ca3b8e243f212e90ff26fcabe2234fb669`, whose protocol/product bytes are
byte-identical to the reworked source. The three uncommitted workflow/test
changes were validated separately by the 60-test validation gate, the focused
scope assertion, and regenerate-check. The clean-check was not bypassed.

## Compatibility and inventory

Cycle-4 changes no schema, case, fixture, vector, manifest, release metadata,
normative document, or legacy surface. A recursive comparison against the
reviewed rc.6 candidate exits 0 after overlaying only the three intended files.
Therefore the independently reviewed inventory and frozen compatibility proof
remain valid:

- 447 exact manifest entries and hashes;
- agent-skill/csk-skill schemas 1 through 5;
- install-marker-v1 and conformance-claim-v1;
- 24 legacy baseline cases and their index/manifest semantics;
- rc.5 metadata frozen at
  `sha256:75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`.

## Diagnostics

- A first focused-test invocation used system Python and exited 1 because
  `jsonschema` was absent. No test ran. The task's pinned venv was then used.
- The first pinned focused test exited 1 because the new extractor did not
  recognize CI's inline `run:` command. The extractor was corrected, after
  which the focused test and full 60-test validation passed.
- An attempted detached-worktree lookup exited 128 because the review commit
  belongs to a retained standalone candidate repository, not the active
  repository object store. No worktree was created; the retained candidate was
  used directly.

## Safety and publication boundary

The change and tests parse workflow text and execute the repository generator,
validator, Go tests, and Git diff only. They execute no package-provided code.
No index was modified and no commit, push, tag, release, conformance claim,
checksum publication, signature, attestation, or implementation pin was
created. Rc.6 still emits no claim and advances no committed downstream pin.

The developer handoff is ready for review.
