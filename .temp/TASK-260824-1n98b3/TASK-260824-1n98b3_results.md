# TASK-260824-1n98b3 — advance consumer pins to rc.9

Landing-order step 9. Both consumer managers move their committed released
conformance-suite pin to the immutable `v1.0.0-rc.9` release commit of
`relux-works/curator-spec`.

## Release identity, re-derived rather than inherited

`v1.0.0-rc.9` is an **annotated** tag, object `b67966449220d42218bd50420e74dac673431464`.
Its peeled commit — and therefore the only thing a pin may name — is:

```
0ed5c691e9208eea52f21db2fc05e226ce3516fd
```

which is also `refs/heads/main` on `curator-spec`. Both consumer pins carry that
40-character commit, not the tag object and not a branch name. Suite manifest
digest `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`,
protocol `1.0.0-rc.9`.

**rc.9 is a strict superset of rc.6**, verified file-by-file against two local
checkouts of `conformance/v1` rather than taken from the commit message:

| | rc.6 (`0c81c1f8`) | rc.9 (`0ed5c691`) |
| --- | ---: | ---: |
| files under `conformance/v1` | 449 | 692 |
| added | — | 243 |
| **removed** | — | **0** |
| changed | — | 7 |

The 7 changed are `manifest.json`, `schema-cases/index.json`,
`vectors/build-drivers.json`, `vectors/conformance-claim-v3-qualification.json`,
`vectors/external-repository-acquisition.json`,
`vectors/external-repository-lifecycle.json`,
`vectors/go-host-execution-policy.json`. (The commit message says "6 changed";
that count excludes `manifest.json`, which necessarily moves. Nothing is
removed either way — that is the load-bearing number.)

## curator — LANDED AND GREEN

PR 38 → merge commit `272b2034d765920672adc4d32c5b1d54a3755ca1` on `main`.
`SPEC_PIN` `00b1688a` (rc.3) → `0ed5c691` (rc.9). Exactly one occurrence of a
pin literal survives anywhere in the repo outside board/logbook prose:

```
.github/workflows/ci.yml:44:  SPEC_PIN: 0ed5c691e9208eea52f21db2fc05e226ce3516fd
```

**Post-merge `main` run `32770177743` — completed / success**, all 11 jobs:
Test, Race and Gate self-test on ubuntu/macos/windows, plus Lint, Naming gate
and Interop conformance gate. The non-default `Candidate suite` lane is
correctly `skipped`.

The pin bump is not cosmetic, and the uploaded evidence proves it. Before, the
suite plan deferred nine packages for want of a publishing root; at rc.9
nothing is deferred:

| Runner | GOOS | served | deferred | excluded |
| --- | --- | ---: | ---: | ---: |
| ubuntu-latest | linux | 42 | **0** | 1 |
| macos-latest | darwin | 43 | **0** | 0 |
| windows-latest | windows | 43 | **0** | 0 |

The single linux exclusion is `internal/godriver`, excluded by the root's *own*
`vectors/conformance-claim-v3-qualification.json`, with the exclusion asserted
by `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`, which still runs and
must pass. Read from `test-evidence-*/test/suite-plan.txt` and
`plan-deferred.txt` (0 lines on all three) in run `32770177743`.

## cocoaskills — pin surface and fail-closed bindings

`RELEASED_SUITE_PIN` `0c81c1f8` (rc.6) → `0ed5c691` (rc.9) in
`.github/workflows/ci.yml`, referenced by 4 `ref:` sites, gated by
`test_the_released_suite_pin_is_declared_once_and_never_inlined`.

The pin cannot move alone — the repo authenticates the released identity in
several independent places, and every one advances in the same commit:

| Surface | rc.6 → rc.9 |
| --- | --- |
| `ci.yml` `RELEASED_SUITE_PIN` | `0c81c1f8` → `0ed5c691` |
| `test_protocol_conformance.py` manifest digest | `sha256:12e58b82…` → `sha256:803918bf…` |
| `test_protocol_conformance.py` protocol version | `1.0.0-rc.6` → `1.0.0-rc.9` |
| release record read | `release/1.0.0-rc.6.json` → `release/1.0.0-rc.9.json` |
| release claim block | `claim_v3` → `claim_v5` |
| in-scope inventory | `agent-skill-v6` 24→28, `csk-skill-v6` 24→28, `SCHEMA_CASES` 102→110 |
| `test_build_metadata.py` manifest digest | `sha256:12e58b82…` → `sha256:803918bf…` |
| `test_ci_workflow.py` pin literal | `0c81c1f8` → `0ed5c691` |
| `test_protocol_shards.py` baseline | 1045 → 1053, p00 565 → 573 |
| shard manifest `.research/…_protocol-shards.json` | `protocol_commit`, `baseline_count` 1045 → 1053, p00 `node_count` 565 → 573 |
| shard verifier `.research/…_verify-protocol-shards.py` | `EXPECTED_PROTOCOL`, plus a named `EXPECTED_BASELINE_COUNT` replacing the inline 1045 |
| isolation classification `.research/…_protocol-isolation-classification.json` | `protocol_commit`, `audited_files_sha256`, 1045 → 1053 nodes, new audit revision |

Two claims were machine-checked instead of read:

- all four `audited_files_sha256` digests recompute **exactly** against the
  working tree (`test_protocol_conformance.py`, `protocol_conformance_adapters.py`,
  `protocol_lifecycle_observations.py`, `conftest.py`);
- the only `atomic_clusters` delta is the 8 new `invalid-v8-*` rows joining the
  **existing** `function::test_rc6_generated_schema_case_is_consumed` cluster —
  no cluster added, removed, or re-membered.

### One defect found and fixed during review

`tests/test_build_metadata.py` advanced its digest constant but kept the comment
above it, which still read *"The 1.0.0-rc.6 candidate suite that publishes
expected/marker-v2.json…"*. That stopped being true the moment the digest below
it moved — a fail-closed identity binding described by the wrong revision is
precisely the drift that misleads whoever advances this pin next. The sibling
file `test_protocol_conformance.py` got its comment updated in the same commit;
this one was missed. Corrected to name the released pin the same way, and the
commit message paragraph explaining it was added. Commit amended, so `main`
receives one coherent commit rather than a follow-up patch.

