# TASK-260728-1g0z69 reviewer verdict — cycle 2

Verdict: changes requested. Route: `analysis`.

The rework closes the five cycle-1 findings: it restores the manager-tested
release-family gate, makes the twelve diagnostic reasons total, adds revisioned
guidance identifiers, separates wire fields from source-metadata assertions,
and distinguishes Stage-A incompatibility from a cache identity miss. The
contract is still not implementation-ready because the following cases have no
single conforming outcome.

## Blocking findings

1. **The complete Go entry does not exhaustively classify its own `toolchain`
   directive.** Reference section 4.2 defines only "above resolved" and "at or
   below resolved", and vectors 9 and 34 cover only those two cases. The Go
   toolchain contract also admits `toolchain default` and toolchain names are
   not limited to canonical release literals; the official Go toolchain
   documentation explicitly defines `default` and permits named/custom
   toolchains. The current decision therefore leaves a validated `go.mod`
   carrying `toolchain default`, a custom name, or another non-canonical name
   without a deterministic Stage-B result. Add an exhaustive classifier and
   positive/negative vectors for every admitted Go `ToolchainName` class. This
   cannot be deferred to a reserved-driver decision because the `go` registry
   entry is declared complete.

   Primary evidence:
   - `docs/compiled-build-toolchain-requirements.md` lines 345-369 and vectors
     9/34.
   - https://go.dev/doc/toolchain (`toolchain default` and named toolchains).
   - https://go.dev/ref/mod (`ToolchainName` grammar).

2. **The diagnostic triggers overlap without a validation or Stage-A
   precedence rule.** A wire `version` value containing a path, URL, prefix, or
   track is simultaneously a bad literal
   (`build_toolchain_requirement_invalid`, section 5) and a smuggled forbidden
   selector (`build_toolchain_package_influence_forbidden`, section 3.1 and
   section 5). Vectors 15-18 and 41-42 do not fix the outcome for the same
   overlapping payload. Stage A has a second overlap: step 2 combines
   resolution and executable verification, while `unavailable` includes "not
   resolvable" and `untrusted` includes a missing/non-regular primary executable
   or unprotected configuration. A configured root whose `primary_relpath` is
   absent can satisfy both descriptions. Partition the triggers or define
   explicit sub-step precedence, and add overlap vectors. The existing
   forbidden-before-compared precedence applies only inside Stage B and does
   not resolve either case.

3. **The mandatory diagnostic payload cannot exist for two of the diagnostics.**
   Section 5 says every code carries the effective requirement, but
   `build_toolchain_requirement_invalid` can fire before a requirement can be
   parsed, and `build_toolchain_requirement_unsatisfiable` denotes an empty
   intersection rather than an effective interval. Define a typed payload
   union or explicit optionality for values that are unavailable, including
   which validated source fragments are carried, then add payload-shape vectors.
   Otherwise independent implementations must invent incompatible sentinel or
   omission behavior.

4. **The guidance catalog still has two incompatible readings.** Selection
   prefers an exact OS tuple and falls back to `any`, which naturally permits
   an exact override plus an `any` fallback. The totality gate instead says
   there is either one active `any` entry or one active entry for every
   supported OS, without saying whether mixed coverage is valid or rejected;
   "one active per tuple" does not decide this because those are distinct
   tuples. Separately, retirement changes an old entry's `active` value and adds
   `superseded_by`, while the catalog is said to be append-only within a
   version. Define the allowed coverage modes and cross-version retirement
   mutation precisely, and add hybrid-coverage plus version-transition vectors.

## Independent validation

- Board artifacts match the task worktree copies byte-for-byte.
- Delta versus the accepted predecessor is exactly the new decision, the new
  reference, and one `CHANGELOG.md` bullet.
- `git diff --check`: pass.
- `go test ./tools/...`: pass.
- `go vet ./tools/...`: pass.
- `gofmt -l tools`: no output.
- `validation-venv/bin/python tools/validate.py`: pass, 42 schemas and 422
  vector files.
- `validation-venv/bin/python -B -m unittest discover -s tools -p
  'test_*.py'`: pass, 29 tests.
- `conformance/v1` and `release/1.0.0-rc.5.json` remain byte-identical to the
  accepted predecessor.

No source, schema, vector, release artifact, commit, pin, or platform claim was
changed by this review.

## Re-review gate

Make the Go metadata classifier exhaustive, make the diagnostic taxonomy and
payloads disjoint and representable, make catalog coverage/version transitions
unambiguous, add the focused vectors above, rerun the unchanged-rc.5 and
determinism gates, and hand the revised decision through another reviewer cycle.
