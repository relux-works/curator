# TASK-260720-1nvomm rework 1

## Contract input

- Accepted contract SHA-256: `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`.
- Repository base: `57c1f56846d221ecc55786bd3c2467ec32f11730`.

## Corrections verified

1. `protocol/core.md` now requires the three package-independent bootstrap
   probes to run from a manager-owned empty working directory with an
   environment that starts empty except for indispensable operating-system
   process variables, uses operation-private roots, sets `GOENV=off`,
   `GOTOOLCHAIN=local`, `LC_ALL=C`, and `LANG=C`, and inherits neither `GOROOT`
   nor a target.
2. `protocol/core.md` now classifies a `go list` result as a trusted toolchain
   package only when both `Standard == true` and `Goroot == true`; its directory
   and every listed input must remain below the fingerprinted `GOROOT`.
3. During the recovery audit, dependency command narrowing was corrected to
   admit exported build commands only from schema 6 providers while retaining
   script-only narrowing for schemas 1 through 5. This matches the accepted
   requirement that build commands participate in dependency-command selection
   like script commands without changing legacy behavior.

No schemas, vectors, generator, manager profile, CLI guide, or release metadata
were edited.

## Validation evidence

The pinned task-local virtual environment was activated so the requested
command ran with `jsonschema==4.25.1`:

```text
$ python3 tools/validate.py
validated 30 schemas and 93 vector files
```

The repository's combined validation gate also passed:

```text
$ make validate
validated 30 schemas and 93 vector files
Ran 8 tests in 0.812s
OK
ok github.com/relux-works/curator-spec/tools/generate-vectors (cached)
```

`git diff --check` passed with no output. The only project documents changed
are the owned `protocol/core.md`, `SECURITY.md`, and
`decisions/0004-compile-only-build-drivers.md`; `.temp/` contains task-local
validation support and this outcome source.
