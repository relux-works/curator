# Muse 1.4.13 offline loader probe

The canonical [qualification record](../../docs/qualification/muse-1.4.13-R6996.1.md)
separates exec and interactive loader results from vendor qualification. Green
witness tests reproduce observations; they do not grant consumer admission.

## Run the opt-in probe

Set `MUSE1413_BINARY` privately to an ignored read-only copy of the installed
executable, never the updating launcher. The harness accepts
only SHA-256 `9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0`.
Do not change the installed source or launcher. Every vendor call sets `MUSE_NO_AUTO_UPDATE=1`, disposable
HOME/XDG, file credentials and provider echo under Apple sandbox-exec.
Without `MUSE1413_BINARY`, every vendor test skips before fixture creation.
Ordinary `go test ./...` excludes this hidden evidence package; explicitly
run the ungated package to verify the skips and non-vendor controls.

```sh
mkdir -p .temp/TASK-261026-jk47fn/tmp
go build -o .temp/TASK-261026-jk47fn/curator ./cmd/curator
export MUSE1413_CURATOR="$PWD/.temp/TASK-261026-jk47fn/curator"
export MUSE1413_RUN_ROOT="$PWD/.temp/TASK-261026-jk47fn/replay-new"
export TMPDIR="$PWD/.temp/TASK-261026-jk47fn/tmp"
# Set MUSE1413_BINARY privately to the frozen executable.
# Set MUSE1413_SPEC to curator-spec 88b0d8615a5df8728e0e9540a9edd708c0ca5660.
# Set MUSE1413_SPEC_PYTHON to its prepared Python executable.
go test ./.research/muse1413 -count=1 -v
go vet ./.research/muse1413
golangci-lint run ./.research/muse1413/...
```

Always use a fresh run root, including controls. Private raw argv, loader
logs, output, exits and tmux PIDs stay there. Tmux uses an isolated background
socket and session. The TUI receives a local echo sentinel, `/rules`, then
`/exit`; no active UI is used. No login, logout or auth command runs, and no
remote inference or real Keychain item is accessed. The offline test pairs
an accessible loopback endpoint with a denied request, and readable ordinary
bytes with a denied synthetic Keychains file.

## Fixtures and reference checker

`testdata` is byte-identical to the 1.4.11/1.4.10/1.4.8 fixture. Keeping its synthetic
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
python3 -m venv .temp/TASK-261026-jk47fn/reference-venv
.temp/TASK-261026-jk47fn/reference-venv/bin/python -m pip install -r "$MUSE1413_SPEC/requirements-dev.txt"
export MUSE1413_SPEC_PYTHON="$PWD/.temp/TASK-261026-jk47fn/reference-venv/bin/python"
go test ./.research/muse1413 -run '^TestReferenceSnapshotPhysicalClosure$' -count=1 -v
```

The Python dependency is the spec's existing `jsonschema==4.25.1`; no new
host-test language or library is introduced. Reference tests skip without
both locators and then provide no reference-checker evidence.

## Narrowing controls

See [mutants](mutants/README.md). To plant an exec foreign-skill leak:

```sh
MUSE1413_OMIT_EXCLUSION=exec go test ./.research/muse1413 -run '^TestMuse1413LoaderMatrix$/exec/linked/skills' -count=1 -v
```

The no-exclusion positive control passes; the required exclusion row must
fail with `user_skills=1; want 0`. Interactive omission is a declared survivor
if the vendor's exclusion already retains the skill. Do not run Muse under
the offline-policy mutant: that control runs only synthetic curl/cat probes.

No source-token scanner attests this module. Controls execute behavioral
suites. Model-visible removal, credentials/refresh, workspace precedence and
native platforms other than macOS arm64 remain outside this loader record.

The [task gaps record](TASK-261026-jk47fn_gaps.md) contains per-form verdicts,
actual command exits and mutant results. [Observations](observations.json)
bind each vendor case to source digests; the
[1.4.11 comparison](comparison-1.4.11.json) reports the unchanged measured
loader/parser behavior. Green witnesses do not qualify the broken interactive
exclusion.

## Resume/trust and superseded build

See [the isolated P7 replay](resume-trust/README.md) for the still-open
CIP-0002 finding, directly reproduced on 1.4.13. The main P1–P4 replay
measures foreign personal sources; P7 measures project sources on resume.
1.4.12-R6909.1 is skipped because task intake reports it superseded on all
known hosts; no evidence is promoted from that untested build.
