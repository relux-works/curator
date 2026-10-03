# TASK-261002-7ajvgs — prepare-spec-rc14

Ready for review. Release preparation only; no commit, tag or publication.

## Candidate identity

- Protocol: `1.0.0-rc.14`, dated `2026-10-02`.
- Core manifest: `sha256:6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`.
- Source manifest: `sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`.
- Source baseline: `v1.0.0-rc.10`, core manifest `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`.

## Scope

Advance active versions, date, changelog and compatibility/security notes;
prepare the rc.14 corpus and exact release record; include the record in Make,
Specification CI and Release workflow regeneration diff lists. Keep released
rc.13/rc.9 identities, schemas, source-suite inputs and implementation pins.
Portable remains the default; unsupported implementation/platform/claim sets
are empty. B3 (#119), source-suite hash v2 and adopted-decision implementation
amendments remain deferred.

The current candidate version is separate from the latest published rc.13
history anchor, so preparing rc.14 does not require creating its tag. Nine
hand-authored vectors require explicit version-field advancement because the
generator preserves their content. Historical claim schemas retain their
original protocol identifiers.

## Observed validation

| Command | Real exit code | Result |
| --- | --- | --- |
| Initial `make validate` | 2 | Host Python lacked `jsonschema`; validator exited 1. |
| Venv `make validate`, attempt 2 | 2 | Validator rejected the still-rc.13 hand-authored env-passthrough vector; validator exited 1. |
| Venv `make validate`, attempt 3 | 2 | Validator passed; 671 Python tests ran, with two assertion failures. Python exited 1; Go tests were not reached. |
| History test subset | 0 | 8 tests passed. |
| Workflow test subset | 0 | 2 tests passed. |
| `make regenerate-check` | 0 | Published record history and generation/diff checks passed. |
| Corrected full `make validate` | 0 | 73 schemas and 1,294 vector files validated; all 671 Python tests passed (2083.107 s); Go tools tests passed (23.299 s). |
| Go formatting check | 0 | `test -z "$(gofmt -l tools)"` produced no unformatted files. |
| `git diff --check HEAD` | 0 | Staged and unstaged whitespace clean. |
| Independent release identity check | 0 | Exact core/source pins; 73/73 released schemas and rc.13 record match tagged bytes; implementation pins unchanged. |

The two assertion failures were corrected: the missing-tag negative must
expect the release-record refusal first through `validate.main()`, and the
generated-file inventory must include rc.14 consistently in Specification CI
as well as Make and the Release workflow. The full rerun uses the corrected
tree and passed. Its elapsed time includes two host execution stalls; the
existing session stayed live and was polled with bounded waits. No existing board evidence is substituted for executing these checks.

The pinned `requirements-dev.txt` dependency was installed in a task-local
virtual environment. Generated files are staged so regeneration compares
against the prepared candidate; all work remains uncommitted.

## Downstream handoff

Leave `.github/workflows/implementations.yml` pins unchanged. All three
`Implementations (ubuntu-latest/macos-latest/windows-latest)` matrix rows feed
the candidate core to Go. The next curator lockstep must admit the exact
rc.14 digest above, then the orchestrator advances the Go commit pin and
requalifies those rows. Registry candidate consumption also needs qualification
against the new core. The Python manager remains on rc.10 core and the unchanged
source suite; neither digest advances for those scoped lanes.

Linux/Windows validation, downstream implementation runs, native provider
qualification, signed release-target gates and publication are not performed
by this macOS release-preparation task. Scoped CI is not full rc.14 conformance
or native-platform evidence.

## Task logbook

- Used the task-scoped readiness outcome from TASK-261002-1pif8m because the
  named `.research/261002_rc14_rc3_readiness.md` was absent from the local curator
  checkout/main ref. The later #122 freeze fix is already present at base
  `045ceb2`; the release brief's scope takes precedence over older readiness
  observations.
- Kept the published rc.13 history anchor while advancing candidate metadata;
  retained fail-closed behavior for missing/unreadable historical evidence.
- Added the rc.14 record to Specification CI after its existing inventory
  consistency test exposed the omitted workflow scope.
- Host execution stalled during the corrected full suite. Waited for existing
  processes rather than changing service configuration. A recovered probe
  reported syspolicyd running with 357 successive crashes, after earlier
  running probes reported 355 and 356. A second stalled interval ended with
  a running probe reporting 358 successive crashes. Validation continued in
  the existing process across these delays and subsequently exited 0.
- Repository `LOGBOOK.md` is untouched; this embedded logbook records the
  task's findings and decisions.

Tagged rc.13 record identity: `sha256:09b5bcfcb4e03df3bd10c6c328704fe3d387ac35beca3b832afb82801445f3dc`.
The generated-path diff is empty against the staged candidate after tests,
and all frozen/schema/source-suite/pin/logbook paths are unchanged against HEAD.
Machine-readable input digests and command outcomes are in
`TASK-261002-7ajvgs_evidence.json`; full successful and failed-attempt logs are
attached separately. No validation evidence from another run was accepted.
