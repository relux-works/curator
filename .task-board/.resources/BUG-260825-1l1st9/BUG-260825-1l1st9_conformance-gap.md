# Conformance-suite gap: no status vector varies the install-marker schema

Reported by BUG-260825-1l1st9 (curator). For the suite owner of
`curator-spec`, pin `0ed5c691e9208eea52f21db2fc05e226ce3516fd`
(v1.0.0-rc.9), `conformance/v1/manifest.json`
`sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`.

## What escaped

A schema-8 skill installed successfully, wrote marker v4, and published its
shims. `curator global status` then reported every one of its compiled
commands `needs-install`, with the remedy
`install marker schema 4 cannot describe a compiled command; reinstall to
record marker schema 2` -- naming a schema the manager would never write for
that band. The manager was conformant on every published vector while doing
this.

## Why the suite did not catch it

The suite proves the write side and the currentness side separately and never
crosses them with the schema as the variable.

- `conformance/v1/expected/install-marker-v4.json` is a write-side golden. It
  pins what a schema-8 installation must record. It says nothing about what a
  reader must then accept.
- `vectors/manager-lifecycle.json` -> `status_cases` has exactly two cases:
  - `compiled-installation-current` lists `marker-schema` among its
    `validated` steps, but does not parameterise which marker schema the
    installation carries. A manager that hard-codes a single schema passes it.
  - `compiled-currentness-failure-matrix` enumerates 14
    `independent_conditions` -- `missing-raw-snapshot`,
    `context-visible-build-root`, `runtime-copied-build-root`,
    `untrusted-cache-boundary`, `unsupported-driver`, `unsupported-toolchain`,
    `corrupt-receipt`, `corrupt-artifact`, `wrong-native-target`,
    `build-source-mismatch`, `cache-key-mismatch`, `receipt-hash-mismatch`,
    `artifact-path-mismatch`, `artifact-hash-mismatch`. None of them is a
    marker-schema condition, in either direction.
- The fixture `fixtures/go-build-skill/.csk-install.json` carries no
  `schema_version` field at all, so a case built on it cannot vary the schema.

Net effect: the suite has no vector that would fail a manager whose status
reader admits exactly one build-bearing marker schema. Every "not current
after a successful install" escape of this shape is invisible to it.

## Requested vectors

1. **Positive, parameterised by band.** Extend `status_cases` with a case
   family that installs the same compiled skill at manifest bands 6, 7 and 8 --
   producing marker schemas 2, 3 and 4 -- with cache, receipt and artifact
   identical, and requires `result: current` for each. This is the vector that
   fails the escape.
2. **Negative, to keep the bound.** Add a marker-schema entry to
   `compiled-currentness-failure-matrix`: an installation whose marker schema
   predates the build record (schema 1) while the closure activates a compiled
   command must be non-current. Without it, a manager could pass (1) by
   deleting its schema band entirely.
3. **Readability band, presentation.** A marker document at a readable schema
   that is nonetheless invalid must not be reported as "from a newer manager".
   Only a schema outside the readable band may be. The escape's sibling defect
   was exactly this: schemas 3 and 4 were reported unreadable.

## Confirmation that (1) and (2) are both load-bearing

The curator fix carries mutation evidence for the equivalent in-repo
regression tests. Narrowing the band back to a single schema turns the
positive cases red; deleting the band turns the negative cases red. A suite
vector family with only one of the two halves would admit one of those
mutants.

## Second reader with the same defect

Beyond status, `internal/scopes/gc.go` marked live build-cache keys only from
marker schemas 2 and 3. A marker v4 contributed no reference, so a garbage
collection pass could delete cache entries a live schema-8 installation was
still running from. `gc_cases` in `manager-lifecycle.json` has no
marker-schema-parameterised liveness vector either; the same case family
should cover the mark phase, not just the status reader.
