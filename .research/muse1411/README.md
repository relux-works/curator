# Muse 1.4.11 offline loader probe

The canonical [qualification record](../../docs/qualification/muse-1.4.11-R6667.1.md)
separates exec and interactive loader results from vendor qualification. Green
witness tests reproduce observations; they do not grant consumer admission.

## Run the opt-in probe

Set `MUSE1411_BINARY` privately to the exact vendor executable from the
task-board Muse pin store, never the updating launcher. The harness accepts
only SHA-256 `6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`.
Do not modify that manager's pin store. A task-local copy with the same digest
is permitted. Every vendor call sets `MUSE_NO_AUTO_UPDATE=1`, disposable
HOME/XDG, file credentials and provider echo under Apple sandbox-exec.
Without `MUSE1411_BINARY`, every vendor test skips before fixture creation.
Ordinary `go test ./...` excludes this hidden evidence package; explicitly
run the ungated package to verify the skips and non-vendor controls.

```sh
mkdir -p .temp/TASK-261021-mw4laf/tmp
go build -o .temp/TASK-261021-mw4laf/curator ./cmd/curator
export MUSE1411_CURATOR="$PWD/.temp/TASK-261021-mw4laf/curator"
export MUSE1411_RUN_ROOT="$PWD/.temp/TASK-261021-mw4laf/replay-new"
export TMPDIR="$PWD/.temp/TASK-261021-mw4laf/tmp"
# Set MUSE1411_BINARY privately to the frozen executable.
# Set MUSE1411_SPEC to curator-spec e41f96387287f5e76604627f9e63662f7d6d86d4.
# Set MUSE1411_SPEC_PYTHON to its prepared Python executable.
go test ./.research/muse1411 -count=1 -v
go vet ./.research/muse1411
golangci-lint run ./.research/muse1411/...
```

Always use a fresh run root, including controls. Private raw argv, loader
logs, output, exits and tmux PIDs stay there. Tmux uses an isolated background
socket and session. The TUI receives a local echo sentinel, `/rules`, then
`/exit`; no active UI is used. No login, logout or auth command runs, and no
remote inference or real Keychain item is accessed. The offline test pairs
an accessible loopback endpoint with a denied request, and readable ordinary
bytes with a denied synthetic Keychains file.

## Fixtures and reference checker

`testdata` is byte-identical to the 1.4.10/1.4.8 fixture. Keeping its synthetic
`muse148-probe` name preserves root, lock and snapshot identities. The current
Curator binary still drives real `profile install` and `env resolve muse
--profile muse148-probe --repair --format json`; its link must render the
same frozen bytes. Snapshot metadata describes the original frozen fixture,
not a new export from this checkpoint. This proves representation and loader
behavior, not export provenance or future materialize integration.

`TestFrozenSnapshot` checks physical source inputs before staging private
0600 copies. Reviewer test `TestReviewContradictoryObservation` retains its
original name and cases. `TestFrozenSnapshotSourceInputs` and
`TestReferenceSnapshotPhysicalClosure` retain both symlink attack classes.
The reference wrapper classifies entries without following links before
calling the published `managed_home_snapshots.check_bundle`.

```sh
python3 -m venv .temp/TASK-261021-mw4laf/reference-venv
.temp/TASK-261021-mw4laf/reference-venv/bin/python -m pip install -r "$MUSE1411_SPEC/requirements-dev.txt"
export MUSE1411_SPEC_PYTHON="$PWD/.temp/TASK-261021-mw4laf/reference-venv/bin/python"
go test ./.research/muse1411 -run '^TestReferenceSnapshotPhysicalClosure$' -count=1 -v
```

The Python dependency is the spec's existing `jsonschema==4.25.1`; no new
host-test language or library is introduced. Reference tests skip without
both locators and then provide no reference-checker evidence.

## Narrowing controls

See [mutants](mutants/README.md). To plant an exec foreign-skill leak:

```sh
MUSE1411_OMIT_EXCLUSION=exec go test ./.research/muse1411 -run '^TestMuse1411LoaderMatrix$/exec/linked/skills' -count=1 -v
```

The no-exclusion positive control passes; the required exclusion row must
fail with `user_skills=1; want 0`. Interactive omission is a declared survivor
if the vendor's exclusion already retains the skill. Do not run Muse under
the offline-policy mutant: that control runs only synthetic curl/cat probes.

No source-token scanner attests this module. Controls execute behavioral
suites. Model-visible removal, credentials/refresh, workspace precedence and
native platforms other than macOS arm64 remain outside this loader record.

The [task gaps record](TASK-261021-mw4laf_gaps.md) contains per-form verdicts,
actual command exits and mutant results. [Observations](observations.json)
bind each vendor case to source digests; the
[1.4.10 comparison](comparison-1.4.10.json) reports the unchanged measured
loader/parser behavior. Green witnesses do not qualify the broken interactive
exclusion.
