# Muse 1.4.8 offline loader probe

The canonical [qualification record](../../docs/qualification/muse-1.4.8-R6256.1.md)
records exec success and interactive skill-exclusion **FAILURE** on the exact
build. A green regression suite means those observations reproduce; it is
not successful vendor qualification. The §7.9 interactive release gate stays
pending. [Task handoff and command results](TASK-261016-i5gwaf_outcome.md).

## Run the opt-in probe

The existing environment gate is `MUSE148_BINARY`; without it, vendor tests
skip. Ordinary `go test ./...` excludes this hidden evidence directory. The
harness rejects any digest other than the recorded exact vendor build.
Set the binary locator privately; never use the auto-updating launcher or
an active home. All vendor calls set `MUSE_NO_AUTO_UPDATE=1`, disposable
HOME/XDG, file backend and offline echo, under a tested network/Keychain deny.

```sh
mkdir -p .temp/TASK-261016-i5gwaf/tmp
go build -o .temp/TASK-261016-i5gwaf/curator ./cmd/curator
export MUSE148_CURATOR="$PWD/.temp/TASK-261016-i5gwaf/curator"
export MUSE148_RUN_ROOT="$PWD/.temp/TASK-261016-i5gwaf/replay-new"
export TMPDIR="$PWD/.temp/TASK-261016-i5gwaf/tmp"
# Set MUSE148_BINARY to the private, frozen vendor executable.
go test ./.research/muse148 -count=1 -v
go vet ./.research/muse148
golangci-lint run ./.research/muse148/...
```

To include the independent reference CLI regressions, also set
`MUSE148_SPEC` to the pinned spec checkout and `MUSE148_SPEC_PYTHON` to its
prepared Python executable. `TestReferenceSnapshotPhysicalClosure` drives the
real CLI with regular, listed-symlink and unlisted-dangling-link inputs.
Without these two locators that test skips; it is not reference evidence.

Review revision 1's `TestReviewContradictoryObservation` is retained by name.
It rejects duplicate count/outcome/event fields, both contradictory orderings,
quoted decoys, malformed counts and integer overflow beside a valid control.
`TestFrozenSnapshot` checks the actual physical fixture before staging any
copy. `TestFrozenSnapshotSourceInputs` drives that entrypoint in child test
processes against malformed source fixtures, so restaging clean bytes cannot
hide a linked source. Source mode bits alone are waived because Git does not
preserve 0600; private runtime copies must still be 0600. The reference CLI
classifies the full physical home inventory without following links before
calling the published `check_bundle`; its bytes-only API cannot attest physical
file types. These checks cover frozen disposable fixtures, not concurrent
hostile mutation or materialize provenance.

Always choose a fresh run root, including mutants. Go drives assertions;
Curator builds the real linked fixture; tmux supplies a private background
terminal and records owned PIDs; sandbox-exec blocks network and native
stores. The TUI receives the local echo sentinel, `/rules`, then `/exit`.
Raw logs, argv, exits and fixture state stay in the task-local ignored run
root. Only synthetic sanitized observations are published here.

`TestMuse148InteractiveExclusionRegression` asserts the observed retained
skill with linked and regular-file roots. It fails loudly on changed loader
counts or startup banner and points to the qualification record. Existing
matrix test names remain; their expectations pin observed behavior. No
runtime qualification constants or gates change.

To plant missing exec exclusion, set `MUSE148_OMIT_EXCLUSION=exec` and run
`go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/linked/skills' -count=1 -v`.
The no-exclusion control passes; required exclusion must fail with one
forbidden skill. The interactive omission **survives** because the vendor
boolean already fails to exclude; it establishes no interactive protection.

## Independently check the frozen snapshot

The source profile and rendered root are synthetic; real Curator install and
resolve must produce a symlink matching the frozen bytes. Snapshot runtime
copies are independent `0600` regular files with the declared XDG placement.
Use the existing reference checker from the exact normative spec revision,
not a replacement snapshot implementation. The only Python dependency is
that spec's pinned `requirements-dev.txt` (`jsonschema==4.25.1`).

```sh
python3 -m venv .temp/TASK-261016-i5gwaf/reference-venv
.temp/TASK-261016-i5gwaf/reference-venv/bin/python -m pip install -r "$PINNED_SPEC/requirements-dev.txt"
export SPEC_PYTHON="$PWD/.temp/TASK-261016-i5gwaf/reference-venv/bin/python"
chmod 600 .research/muse148/testdata/snapshot/home/config/muse/AGENTS.md
"$SPEC_PYTHON" -B .research/muse148/reference-snapshot-check.py "$PINNED_SPEC" .research/muse148/testdata/snapshot
MUSE148_SNAPSHOT_PLANT=extra "$SPEC_PYTHON" -B .research/muse148/reference-snapshot-check.py "$PINNED_SPEC" .research/muse148/testdata/snapshot
```

The first check must exit 0; the planted extra member must exit **1** with
`ValueError: entry-set`. This tests frozen representation/loader behavior,
not materialize provenance, capture locks, publication, destination admission,
runtime-B or actual `env materialize -> Muse` integration.

Historical `replay-03.log` records the first expected-zero assertions failing.
`observations.json` is its sanitized vendor evidence. Rework observations and
logs are separately named; a changed assertion is never mislabeled as a
changed vendor result. Model-visible behavior, refresh and credential backend
qualification remain unmeasured. Full repository gate: `sh scripts/remote-gate.sh`.
