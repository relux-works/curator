# TASK-260822-c0rxj7 — Schema 8 rc.9 candidate evidence

This is the canonical results artifact for the shared Schema 8 / protocol
`1.0.0-rc.9` conformance candidate. It replaces the earlier revision that
still described `859727b1…` as the candidate; that identity and `edd07210…`
and `e66cb72d…` are recorded below as superseded history, never rewritten.

Nothing here was merged to `main`. Per the TASK-260823-omp8zt re-scope this
task only produces and qualifies an immutable candidate outside `main`; the
landing PR that advances pins atomically is a later task.

## 1. Current candidate

- Repository: `relux-works/curator-spec`
- Branch: `candidate/schema-8-rc.9`
- Full commit SHA: `6001dc33281b94a4ec7442ab15278550dd0f51d9` (signed, pushed)
- Protocol version: `1.0.0-rc.9`
- Suite manifest SHA-256: `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`
- Tree SHA-256: `sha256:9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769` (measured under `LC_ALL=C`; see §6)
- File count under `conformance/v1`: `692`
- `release/1.0.0-rc.8.json` SHA-256: `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`, byte-identical to `origin/main`
- `SPEC_PIN` untouched at `00b1688a9b2457ca397a0bb550acf47cad8ee967`

Verified ancestors of the candidate: `dd9c9fc` (script-worker normative),
`bac193c` (module-roots prose), `ebfed81` (schema), `41cf556` (core prose),
`e5df43d` (manager + SECURITY), `b92b105`, `517a130` (curator-spec `main`),
`edd0721`, `e66cb72`. `git log HEAD..origin/main` is empty — `main` is fully
contained.

The suite digest is unchanged from `edd0721` onward by construction: every
commit after it touches only prose outside `conformance/v1`.

## 2. Superseded candidate identities (history, not the candidate)

| Identity | Manifest SHA-256 | Tree SHA-256 | Files | Outcome |
| --- | --- | --- | ---: | --- |
| `859727b103ed175ff214cbb64641f4686d8c6a68` | `sha256:782d6868…f11f` | `sha256:f88a7626…19f3` | 692 | Candidate lane red (curator multi-project dry-run binding unlanded; Windows `shasum` path escaping). Both causes have since been fixed and merged in curator. |
| `edd07210d4f3db34fd60238cb14b90f837de03cb` | `sha256:803918bf…b44403` | `sha256:9d5a10b6…b769` | 692 | Fully green (spec CI 32642316308, candidate lane 32651139699), superseded by review cycle 1 rework. |
| `e66cb72d9988c614c7232af9195bf829c82d328e` | `sha256:803918bf…b44403` | `sha256:9d5a10b6…b769` | 692 | Fully green (spec CI 32654392587, candidate lane 32654422338), superseded by review cycle 2 rework. |

None of these commits were rewritten; all remain reachable on
`candidate/schema-8-rc.9`.

## 3. What the candidate carries

- Manifest schema 8 (`agent-skill-v8`, `csk-skill-v8`) as the single shared
  bump for the `script-worker-v1` execution policy and first-party module
  roots — module roots takes no sequential version of its own.
- Install marker schema 4, plus its writer golden and schema-case family.
- Conformance claim v5, rc.9 release metadata, and the live candidate-suite
  pin moved to `release/1.0.0-rc.9.json`. rc.5–rc.8 stay byte-frozen and are
  guarded by the immutability check in all three diff gates.
- Normative prose closing the review-cycle findings:
  - `protocol/core.md` §10 — schema-8 managers MUST read marker schemas 1–4
    and MUST write marker schema 4 for schema-8 installation mutations, plus
    the marker-v4 semantics paragraph.
  - `protocol/core.md` §4 — schemas 1 through 7 MUST reject `execution_policy`
    and `interpreter` at the top level and on every command, with both
    enforcement paths named.
  - `profiles/manager.md` §11 — the external repository profile is banded on
    marker v4 for schema-8 installations (see §5).
  - `conformance/README.md` — the marker-v4 writer golden is listed.
  - `schemas/v1/README.md` — the opening paragraph names rc.9's new closed
    objects instead of stopping at rc.8.
- Lockstep cell: `active-process-count-limit` on Linux reads
  `host-conditional: delegated cgroup v2 pids.max` in both `protocol/core.md`
  and `profiles/manager.md`.

## 4. Local gates at `6001dc3`

Each command was run directly as a standalone process; the exit codes below
are the real process results.

| Gate | Exit | Result |
| --- | ---: | --- |
| `python3 tools/validate.py` | 0 | 53 schemas, 691 vector files |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 0 | 98 tests, OK |
| `go test ./tools/...` | 0 | ok |
| `gofmt -l tools` | 0 | no output |
| `git diff --check` | 0 | clean |
| `go run ./tools/generate-vectors -root .` pass 1 | 0 | — |
| `git diff --exit-code` over `conformance/v1` + rc.5–rc.9 after pass 1 | 0 | clean |
| `go run ./tools/generate-vectors -root .` pass 2 | 0 | — |
| `git diff --exit-code` over `conformance/v1` + rc.5–rc.9 after pass 2 | 0 | clean |
| `cmp` of the two 692-entry checksum inventories | 0 | byte-identical |
| `python3 -B tools/release_gate.py --version 1.0.0-rc.9 --commit HEAD` | 0 | release gate passed for 1.0.0-rc.9 |
| `shasum -a 256 conformance/v1/manifest.json` | — | `803918bf…b44403` |
| `shasum -a 256 release/1.0.0-rc.8.json` | — | `293f101d…e31ede`, equal to `origin/main` |

## 5. Decision recorded: `profiles/manager.md` §11 marker band (review cycle 2, item 3.1)

**Decision: fixed, not deferred.**

`protocol/core.md` §10 makes a schema-8 installation record marker v4. The
external repository manager profile still named marker v3 in all three of its
consumer bands — shim selection to the protected artifact
(`profiles/manager.md` §11.8), read-only status validation and GC rooting of
local snapshots, external snapshot keys, receipt/artifact keys and shim
relationships (§11.9). Schema 8 admits `go-repository-v1` commands
(`agent-skill-v8` carries `build_repositories` and `commandV8`, and the suite
ships schema-8 marker-v4 external cases), so a conforming schema-8 manager
could install an external command, write marker v4 exactly as required, and
then find that installation rooted by nothing at GC time.

The counter-argument — that `core.md` "marker v4 … otherwise carries
marker-v3 meaning unchanged" makes every §11 marker-v3 statement apply by
inheritance — was considered and rejected as the basis for shipping. The
strongest written form of it, `schemas/v1/README.md` §"Install markers",
scopes the inheritance to marker-v3 *build-record* rules, not to this
profile's shim-selection, status-validation and GC-rooting obligations. More
decisively, §11 itself set the precedent: when schema 7 arrived it re-banded
explicitly from v2 to v3 rather than relying on inheritance. Doing the same
for v4 costs one paragraph and removes the ambiguity entirely; leaving it
implicit repeats the exact class of gap that blocked review cycle 1.

The fix is prose outside `conformance/v1`, so the suite digest is unchanged.
The carried-over minor from cycle 2 item 3.2 (`schemas/v1/README.md` opening
on rc.8 while its body documented rc.9) was folded into the same commit.
Cycle 2 item 3.3 (`conformance/README.md` "every marker role") was flagged
cosmetic with no action required and was deliberately left alone.

## 6. Finding carried forward and sharpened: locale-dependent tree digest

`.github/ci/candidate-suite.sh:101` (curator repo) sorts the enumerated
candidate paths with a bare `sort` and no locale pin, so `tree_sha256` is a
function of the recording host's collation. Over this candidate's
`conformance/v1` the same bytes measure:

| Locale | `tree_sha256` |
| --- | --- |
| `LC_ALL=C` | `9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769` |
| `LANG=en_US.UTF-8` | `176dc52bdb73bc57ae394e2a063e9bb80dc3cd8f4c51f75b74c4144a8c942f02` |

Both were reproduced locally in this run.

**Correction to the earlier characterisation.** Review cycle 2 recorded that
"the recorded canonical value is the `LC_ALL=C` one, which is what the runners
produce". That is true for only two of the three runners. Reading the job
logs of candidate lane 32659157687 directly:

| Runner | Recorded `tree_sha256` | `manifest_sha256` | Files |
| --- | --- | --- | ---: |
| ubuntu-latest | `sha256:9d5a10b6…b769` | `sha256:803918bf…b44403` | 692 |
| windows-latest | `sha256:9d5a10b6…b769` | `sha256:803918bf…b44403` | 692 |
| macos-latest | `sha256:176dc52b…2f02` | `sha256:803918bf…b44403` | 692 |

The macOS runner records the UTF-8 collation variant. This is not new to this
candidate: the macOS job of the earlier green lane 32654422338 recorded the
same `176dc52b…` for `e66cb72`.

**Status.** Today this is still recorded-only evidence and not a live gate
failure, because `.github/workflows/ci.yml:365` wires only
`CANDIDATE_EXPECTED_MANIFEST_SHA256`; `CANDIDATE_EXPECTED_TREE_SHA256` is
supported by the script but never supplied, and the manifest digest — which is
locale-independent and *is* verified — matched the supplied expectation on all
three runners. But the moment anyone wires the tree expectation, the macOS
lane fails against a Linux-recorded expectation for byte-identical content.
It is a latent gate hole, not merely a cosmetic evidence inconsistency.

**Routing.** The fix belongs in the curator repo (pin `LC_ALL=C` around the
enumeration in `candidate-suite.sh`), not in this spec candidate, and was
deliberately not patched mid-qualification — changing the recording algorithm
during a qualification loop would invalidate the identity being qualified.
It needs its own tracked curator task before `CANDIDATE_EXPECTED_TREE_SHA256`
is ever supplied.

## 7. CI evidence

**curator-spec Specification CI** — run
[32659168954](https://github.com/relux-works/curator-spec/actions/runs/32659168954),
success on exactly `6001dc33281b94a4ec7442ab15278550dd0f51d9`, all six jobs:
Formatting, Links, Release target provenance, and Specification on
ubuntu-latest / macos-latest / windows-latest.

**curator candidate-conformance** — run
[32659157687](https://github.com/relux-works/curator/actions/runs/32659157687),
dispatched from curator `main` (`e17b0f1`) with
`candidate_ref=6001dc33281b94a4ec7442ab15278550dd0f51d9` and
`candidate_manifest_sha256=sha256:803918bf…b44403`. `SPEC_PIN` was not
touched. Workflow conclusion: **success**, all fourteen jobs green, including
**Candidate suite** on ubuntu-latest, macos-latest and windows-latest. The
job logs confirm on all three runners: `CANDIDATE_REF:
6001dc33281b94a4ec7442ab15278550dd0f51d9`, `candidate-suite: revision accepted
(immutable, full 40-hex)`, `CANDIDATE_EXPECTED_MANIFEST_SHA256:
sha256:803918bf…b44403`, `candidate-suite: manifest digest matches the
supplied expectation`, `candidate_revision 6001dc33…`, `file_count 692`, and
`SPEC_PIN: 00b1688a9b2457ca397a0bb550acf47cad8ee967` throughout.

Earlier green runs on superseded identities: 32651139699 (`edd0721`) and
32654422338 (`e66cb72`), both success on all three OSes. The two red runs on
`859727b` (32633572039, 32638424105) exposed seven curator-side defects; all
seven were fixed and merged to curator `main` with green lanes, and that
qualification loop is documented on TASK-260823-1l1p8q.

## 8. Handoff

- `candidate/schema-8-rc.9` and the task-scoped worktree at
  `.temp/TASK-260822-c0rxj7/curator-spec-candidate` are preserved for the
  downstream implementation and landing tasks.
- `SPEC_PIN` stays at `00b1688a9b2457ca397a0bb550acf47cad8ee967`. Advancing
  the pins, merging the candidate and publishing rc.9 belongs to the later
  landing task (steps 6–7 of the TASK-260823-omp8zt order), not to this one.
- Green candidate evidence for `6001dc3` is routed to `TASK-260822-f4qv7w`
  and `TASK-260822-1so0ym`; both remain `blocked` only because the board
  dependency edge points at this task and releases when it reaches `done`.
- Unblock notes were previously recorded on curator `STORY-260822-2h0v9j`
  and CocoaSkills `STORY-260822-2evh3p`.
