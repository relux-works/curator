# TASK-260822-c0rxj7 — review cycle 3 verdict: ACCEPTED

Run: `RUN-260823-94bb58`. Reviewer-independent verification in a fresh detached
worktree at `.temp/TASK-260822-c0rxj7/review-cycle3/wt`; every figure below was
measured in this run, not copied from the producer summary.

## Verdict

**ACCEPTED → `done`.** All three cycle-2 blocking/decision items are closed,
each verified against primary evidence rather than against the delivery note.
No new blocking finding.

## Cycle-2 items, re-verified

**Cycle-2 BLOCKING 1 (stale results artifact) — CLOSED.**
`TASK-260822-c0rxj7_results.md` is fully rewritten to the delivered state. It
names `6001dc33281b94a4ec7442ab15278550dd0f51d9` as the current candidate,
carries a superseded-identity table for `859727b` / `edd0721` / `e66cb72` with
their outcomes, and the stale red-matrix "blocked, candidate is 859727b"
framing is gone. §8 correctly states nothing was merged to `main`.

**Cycle-2 BLOCKING 2 (`manager.md` misattribution) — CLOSED.**
`TASK-260822-c0rxj7_e66cb72-green-matrix.md` carries a `SUPERSEDED` header and
an explicit correction naming the `+5/-1` hunk as the merged `517a130`
credential-scopes prose. The new `TASK-260822-c0rxj7_6001dc3-green-matrix.md`
describes its delta accurately — I diffed it and the description matches the
bytes exactly.

**Cycle-2 DECISION OWED (§11 marker band) — DECIDED AND FIXED, correctly.**
`git diff e66cb72 6001dc3` is two prose files, `profiles/manager.md` (+6) and
`schemas/v1/README.md` (+10/-4), nothing under `conformance/v1`. The added
`profiles/manager.md` paragraph sits at lines 1102-1106, inside the §11
preamble (§11 spans 1094→end, first subsection §11.1 at 1130), so it governs
every band statement below it. It names exactly the three obligations cycle 2
identified as unrooted, and each maps to a real line:

| Obligation named in the new paragraph | Band statement it reaches |
| --- | --- |
| shim selection | `profiles/manager.md:1771` (§11.8) |
| read-only status validation | `profiles/manager.md:1792` (§11.9) |
| GC rooting | `profiles/manager.md:1812` (§11.9) |

It rests on a real normative rule, not an assertion: `protocol/core.md:1747`
reads "Marker v4 permits `skill_schema_version` 8 and otherwise carries
marker-v3 meaning unchanged." The premise that schema 8 admits external builds
also holds — `agent-skill-v8.schema.json` carries `build_repositories`, and
`go-repository-v1` is the `driver` const in `common.schema.json` reached by
both v7 and v8. The gap that blocked cycle 1 and was flagged in cycle 2 is
genuinely closed: a schema-8 external installation writing marker v4 is now
rooted at GC time by profile text, not by inference.

## Candidate identity, measured

- `origin/candidate/schema-8-rc.9` = `6001dc33281b94a4ec7442ab15278550dd0f51d9`;
  `git log --format=%G?` = `G` (good signature); pushed and immutable.
- Ancestors verified with `git merge-base --is-ancestor`: `dd9c9fc`, `bac193c`,
  `ebfed81`, `41cf556`, `e5df43d`, `b92b105`, `edd0721`, `517a130`, `e66cb72`
  — all yes. `git log 6001dc3..origin/main` is empty; `main` is contained.
- `conformance/v1/manifest.json` = `803918bf…b44403`, 692 files.
- `release/1.0.0-rc.8.json` = `293f101d…e31ede` on the candidate and on
  `origin/main` alike — byte-frozen.
- `SPEC_PIN` untouched at `00b1688a…`.

## Local gates, re-run standalone in a clean worktree

| Gate | Exit | Measured result |
| --- | ---: | --- |
| `tools/validate.py` | 0 | 53 schemas, 691 vector files |
| `unittest discover -s tools` | 0 | 98 tests, OK |
| `go test ./tools/...` | 0 | ok |
| `gofmt -l tools` | 0 | empty |
| `git diff --check` | 0 | clean |
| `release_gate.py --version 1.0.0-rc.9 --commit HEAD` | 0 | passed at `6001dc3` |
| `generate-vectors` pass 1 + scoped `git diff --exit-code` | 0 / 0 | clean |
| `generate-vectors` pass 2 + scoped `git diff --exit-code` | 0 / 0 | clean |
| `cmp` pass1 vs pass2 inventories (692 entries each) | 0 | byte-identical |

Stronger than a self-consistency check: my independently generated 692-entry
inventory is `cmp`-identical to the producer's
`TASK-260822-c0rxj7_regeneration-pass2-6001dc3.sha256`. The python gates need
`jsonschema`; run them with an env that has it (I used
`curator-spec-production/.venv`), and clear `tools/__pycache__` before
`release_gate.py` — it requires a clean checkout and an untracked `__pycache__`
alone fails it.

## CI evidence, read from the API and the job logs

- curator-spec **Specification CI 32659168954**: conclusion `success` on
  `headSha` exactly `6001dc33281b94a4ec7442ab15278550dd0f51d9`, all six jobs
  (Formatting, Links, Release target provenance, Specification ×3 OS).
- curator **candidate-conformance 32659157687**: `workflow_dispatch`,
  `attempt: 1`, conclusion `success`, all **14** jobs green, no reruns;
  dispatched from curator `main` `e17b0f1` (confirmed ancestor of `origin/main`).
- Candidate suite job logs on all three runners
  (`97242559072` / `97242559109` / `97242559132`) each show `CANDIDATE_REF:
  6001dc33…`, `revision accepted (immutable, full 40-hex)`,
  `CANDIDATE_EXPECTED_MANIFEST_SHA256: sha256:803918bf…b44403`, `manifest digest
  matches the supplied expectation`, `file_count 692`, `SPEC_PIN 00b1688a…`.

## Downstream routing — verified on the board

`TASK-260822-f4qv7w` and `TASK-260822-1so0ym` both carry a refreshed unblock
note naming `6001dc3` / `32659168954` / `32659157687`, superseding the
`edd0721` and `e66cb72` notes. Both remain `blocked` only because the
dependency edge points at this task; it releases on `done`.

## Non-blocking observations (do not gate acceptance)

1. **`results.md` §5 rationale prose is loose about the mechanism.** It says
   the inheritance counter-argument was "rejected" and that §11 "set the
   precedent … doing the same for v4", but the delivered paragraph *is* an
   inheritance bridge — stated normatively and scoped to this profile's
   obligations — rather than a per-band re-naming. The substance is right and
   arguably the better fix (one governing sentence beats three edits), the
   factual claim about what the delta contains is true, and the
   `6001dc3-green-matrix.md` description is precise. Worth a one-line
   tightening if the artifact is touched again; not rework.
2. **The task's AC free-text field is stale and contradicts the governing
   scope.** It still reads "PR squash-merged to main … branch and worktree
   cleaned", while the board-recorded TASK-260823-omp8zt re-scope explicitly
   forbids merging to `main` from this task and requires preserving the branch
   and worktree. The developer flagged this rather than silently reinterpreting
   it, which is the correct call; all three review cycles agree not merging is
   right. This needs a board-owner edit to the AC field so a future reader does
   not act on the stale text. Not a developer action, not a rework loop.
3. **The locale-dependent `tree_sha256` is a real latent gate hole and is still
   untracked.** Independently reproduced: `.github/ci/candidate-suite.sh:101`
   sorts with a bare `sort` and no locale pin; `.github/workflows/ci.yml:365`
   wires only `CANDIDATE_EXPECTED_MANIFEST_SHA256`, while the script supports
   `CANDIDATE_EXPECTED_TREE_SHA256` at line 129 and it is wired nowhere. The
   three green job logs confirm the divergence for byte-identical content:
   ubuntu `9d5a10b6…b769`, windows `9d5a10b6…b769`, macos `176dc52b…2f02`, all
   with the same verified `manifest_sha256` and `file_count 692`. Not a defect
   of this candidate and correctly not patched mid-qualification, but it needs
   its own curator task before anyone supplies the tree expectation — and until
   then no downstream consumer may compare `tree_sha256` across hosts. The
   recorded `9d5a10b6…` is the `LC_ALL=C` value.

## Not defects

- Not merging to `main`, and preserving `candidate/schema-8-rc.9` plus the
  worktree at `.temp/TASK-260822-c0rxj7/curator-spec-candidate`, is required by
  the re-scope, not an omission.
- The suite digest being unchanged across `edd0721` → `e66cb72` → `6001dc3` is
  correct by construction: every commit after `edd0721` touches only prose
  outside `conformance/v1`. I verified the delta path set, not just the claim.
- No Implementation-conformance run on the candidate branch is by design.
