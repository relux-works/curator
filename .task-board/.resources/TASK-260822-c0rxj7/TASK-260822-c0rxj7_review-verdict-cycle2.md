# TASK-260822-c0rxj7 — review cycle 2 verdict (RUN-260823-129cf5)

**Verdict: CHANGES REQUESTED → `to-dev`.**

The engineering is done. The three prose fixes cycle 1 demanded are delivered and
correct, the candidate re-cut is clean, and both lanes are green on the exact new
SHA. What is not delivered is rework item 4 of the cycle-1 packet, and the green
evidence artifact misdescribes its own delta. One substantive prose gap of the
same class as the cycle-1 blocker is also reported below for an explicit decision.

Nothing here requires vector regeneration or a digest change.

## 1. Independently verified — good

Every fact below was measured in this run, not taken from the producer summary.

**Candidate identity.** `e66cb72d9988c614c7232af9195bf829c82d328e` on
`candidate/schema-8-rc.9`, signed (`%G? = G`), pushed. Two commits on `edd0721`:
`586ebd9` merges curator-spec `origin/main` `517a130`, `e66cb72` carries the fixes.
`git log e66cb72..origin/main` is empty — main is fully contained. `dd9c9fc`,
`bac193c`, `ebfed81`, `41cf556`, `e5df43d`, `b92b105`, `edd0721`, `517a130` all
verify as ancestors. The `edd0721` and `859727b` identities are unrewritten.

**Digests.** `edd0721..e66cb72` touches only `conformance/README.md`,
`profiles/manager.md`, `protocol/core.md` — nothing under `conformance/v1`, so the
suite identity is unchanged by construction, and measured:
`conformance/v1/manifest.json` = `803918bf…b44403` at both `edd0721` and `e66cb72`;
692 files under `conformance/v1`. `release/1.0.0-rc.8.json` is
`293f101d…e31ede` on both the candidate and `origin/main` — byte-identical.

**Tree digest and the locale finding — both reproduced.** Over the candidate's
`conformance/v1`, using the exact `candidate-suite.sh record` algorithm:
`LC_ALL=C` → `9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769`;
`LANG=en_US.UTF-8` → `176dc52bdb73bc57ae394e2a063e9bb80dc3cd8f4c51f75b74c4144a8c942f02`.
The producer's finding is real: `.github/ci/candidate-suite.sh:101` runs a bare
`sort` with no locale pin. It is not a gate hole — `ci.yml:365` wires only
`CANDIDATE_EXPECTED_MANIFEST_SHA256`; `CANDIDATE_EXPECTED_TREE_SHA256` is never
supplied, so the tree digest is recorded evidence, never verified. The recorded
canonical value is the `LC_ALL=C` one, which is what the runners produce. Deferring
it mid-qualification was the right call; it should be pinned separately.

**Local gates, re-run from scratch in a clean detached worktree at the candidate,
real exit codes.**

| Gate | Exit | Result |
| --- | ---: | --- |
| `python3 tools/validate.py` | 0 | 53 schemas, 691 vector files |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 0 | 98 tests, OK |
| `go test ./tools/...` | 0 | ok |
| `gofmt -l tools` | 0 | no output |
| generate-vectors pass 1 + `git diff --exit-code` | 0 / 0 | clean |
| generate-vectors pass 2 + `git diff --exit-code` | 0 / 0 | clean |
| `cmp` pass1 vs pass2 inventory (692 entries each) | 0 | byte-identical |
| `shasum -a 256 conformance/v1/manifest.json` | — | `803918bf…b44403` |

**CI.** curator-spec Specification CI
[32654392587](https://github.com/relux-works/curator-spec/actions/runs/32654392587)
= success on exactly `e66cb72`, all six jobs (Formatting, Links, Release target
provenance, Specification on ubuntu/macos/windows).

curator candidate-conformance
[32654422338](https://github.com/relux-works/curator/actions/runs/32654422338),
dispatched from curator `main` `e17b0f1` (verified an ancestor of curator
`origin/main`) = success. Job logs confirm on all three OSes:
`CANDIDATE_REF: e66cb72d9988…`, `candidate-suite: revision accepted (immutable,
full 40-hex)`, `CANDIDATE_EXPECTED_MANIFEST_SHA256: sha256:803918bf…`,
`candidate-suite: manifest digest matches the supplied expectation`, and
`SPEC_PIN: 00b1688a9b2457ca397a0bb550acf47cad8ee967` throughout — the pin is
untouched. Attempt 1 had all three **Candidate suite** jobs green and failed only
on the unrelated `Test (windows-latest)` / `go test + platform-case gate`; attempt
2 re-ran that job and passed. The green-matrix claim on this point is accurate.

**The cycle-1 blocker is closed.** `core.md:1695-1702` now reads: managers
supporting schema 7 MUST read marker schemas 1, 2, 3, **and managers supporting
schema 8 MUST read marker schemas 1, 2, 3, and 4**; they MUST write marker schema 2
for schema 1-6 mutations, 3 for schema 7, **and 4 for schema 8 installation
mutations**. The schema-7 sentence is intact. `core.md:1740-1749` adds the marker-v4
paragraph. Checked against reality, not just against the packet: `diff` of
`install-marker-v3.schema.json` vs `install-marker-v4.schema.json` differs in
exactly `$id`, `title`, `schema_version` 3→4, `skill_schema_version` 7→8 — so
"same object shape … otherwise carries marker-v3 meaning unchanged" is literally
true, and every clause of the new paragraph (receipt schema version and
`execution_policy` per build entry, `go-v1` / `go-repository-v1` entry semantics,
top-level `build_source` / `build_roots`) maps 1:1 onto the marker-v3 paragraphs at
`core.md:1713-1738`. With v4 in the schema-8 read set, `core.md:1790` no longer
strands it.

**Delta (c) second half delivered and matches the implementation.**
`core.md:190-195` now gates `execution_policy` and `interpreter` for schemas 1-7 at
the top level and on every command, naming both enforcement paths. Verified against
`tools/validate.py:571-582` (`validate_wire_semantics`, regex `v([1-7])`, both
top-level and per-command, `SCRIPT_EXECUTION_FIELDS`) and the schema loop at
`tools/validate.py:255-287`. Prose and code now agree.

**Remaining checklist claims re-verified at the new SHA.** Lockstep cell reads
`host-conditional: delegated cgroup v2 pids.max` in `core.md:438` and
`manager.md:784`. Section 4 preamble names v1-v8 including `csk-skill-v8`
(`core.md:156-160`); Added-behavior row 8 present (`core.md:183`).
`conformance/README.md:66-67` now lists the marker-v4 writer golden, and
`conformance/v1/expected/install-marker-v4.json` exists with `schema_version` 4 /
`skill_schema_version` 8. `CHANGELOG.md:6-21` carries the rc.9 entry;
`COMPATIBILITY.md:158-162` carries the required paragraph (schema 8 + marker 4, one
shared bump with decision 0009, module roots takes no sequential version). rc.9 pin
moved in `Makefile`, `ci.yml`, `release.yml`, `README.md`, `protocol/assurance.md`,
`tools/validate.py`, `tools/release_gate.py`, `tools/generate-vectors/main.go`,
decision 0007, with the rc.5-rc.9 immutability guard in all three diff gates. Not
merging to `main` remains correct per the re-scope.

## 2. Blocking

### 2.1 CONFIRMED — cycle-1 rework item 4 was not delivered and was not mentioned

`TASK-260822-c0rxj7_results.md` (mtime 17:37, i.e. untouched by the 21:19 rework)
still opens with:

- Full commit SHA `859727b103ed175ff214cbb64641f4686d8c6a68`
- Suite manifest SHA-256 `sha256:782d6868…f11f`
- Tree SHA-256 `sha256:f88a7626…9f3`
- a "Candidate-conformance result and blocker" section describing a definitively
  red matrix and two external blockers that are now cleared.

Cycle 1 listed this as rework item 4 verbatim ("Rewrite … to the delivered state …
keeping the superseded `859727b1…` identity recorded as history rather than as the
candidate"). The delivery note for RUN-260823-c4cb4f enumerates fixes 1-4 and never
touches it. Its sibling `…_candidate-suite-identity-e66cb72.txt` records
`e66cb72` / `803918bf…` / `9d5a10b6…` / 692.

Failure scenario: this task is accepted, goes `done`, and releases the board edge
that unblocks `TASK-260822-f4qv7w` and `TASK-260822-1so0ym`. Anyone — the later
landing task that owns steps 6-7, or either vector task — opening this task's
canonical results artifact reads "candidate is 859727b, digest 782d6868, matrix
red, blocked". That is exactly the false-evidence class this chain has stopped the
line for twice.

### 2.2 CONFIRMED — the green-evidence artifact misattributes its own delta

`TASK-260822-c0rxj7_e66cb72-green-matrix.md` says: *"core.md section 10 obligations
extended … **matching profiles/manager.md surface**"* and lists
`profiles/manager.md +5/-1` as part of the delta content.

That is not what the `profiles/manager.md` change is. `git diff edd0721 e66cb72 --
profiles/manager.md` is entirely the merged `517a130` prose — "Operator credential
selections — the `build_ssh` scopes among them — are never lockable" at
`manager.md:45-51`. It has nothing to do with marker v4. The board note for the same
run states the opposite and is the correct one: *"profiles/manager.md was checked
and needs no matching obligation surface."*

Failure scenario: the green matrix is the artifact downstream tasks will cite as
proof that the marker-v4 obligation landed on both surfaces. It claims a
manager.md obligation surface that does not exist, which is precisely the gap
raised in 3.1 below.

## 3. Needs an explicit decision before re-cut

### 3.1 PLAUSIBLE — `profiles/manager.md` section 11 still bands on marker v3 only

The delivery note's justification is: *"its only marker-band statements are inside
the section 11 go-repository-v1 profile and are scoped to marker v3 there; the
read/write band obligation lives in core.md section 10 alone."* Being scoped to v3
is the concern, not its resolution.

Schema 8 admits external builds — `agent-skill-v8.schema.json` carries
`build_repositories` and `commandV8`, and the suite ships schema-8 marker-v4
external cases (`schema-cases/install-marker-v4/valid-external-only-substituted`,
`valid-external-only-unsubstituted`, `valid-sha256-external`,
`valid-untagged-external`, plus six `invalid-external-*`). So a schema-8
installation with `go-repository-v1` commands is a vector-covered configuration,
and per the new `core.md:1695-1702` it writes **marker v4**. But:

- `manager.md:1765` — the shim "resolves exactly to the protected artifact selected
  by **marker v3**";
- `manager.md:1786` — read-only status "validates **marker v3**";
- `manager.md:1806` — GC: "Valid **marker v3** records mark referenced local skill
  snapshots, complete external snapshot keys, receipt/artifact keys, and
  manager-generated shim relationships."

Failure scenario: a conforming schema-8 manager installs a skill with a
`go-repository-v1` command, writes marker v4 as core.md now requires, then runs GC.
`manager.md:1806` roots only marker-v3 records; the base GC rule at
`manager.md:1083-1086` roots build-cache entries from "a valid marker v2" (extended
by section 11 to v3, never to v4). The external snapshot, receipt/artifact keys and
shim relationships of that installation are unreferenced, sweep after the grace
period, and the installed shim breaks — while the manager conformed throughout.
Same shape as the cycle-1 blocker: the write band moved, a v-scoped consumer did
not.

Why this is PLAUSIBLE and not CONFIRMED, stated fairly:

- `core.md:1740` says marker v4 "otherwise carries marker-v3 **meaning** unchanged
  … the same local `go-v1` and external `go-repository-v1` entry semantics". A
  charitable reading makes every "marker v3" statement in section 11 apply to v4 by
  inheritance, and section 11 opens "extends, but does not reinterpret, sections 1
  through 10".
- `schemas/v1/README.md:97-101` states it more strongly still: "Marker v4 is
  marker v3 with `schema_version` 4 and `skill_schema_version` 8 and no other
  difference, so every marker-v3 **build-record** rule … applies unchanged." Note
  the scope: build-record rules, not the manager profile's shim-selection, status
  and GC-rooting obligations.
- The `manager.md:1096` "rc.5 manager profile for **schema-7** `go-repository-v1`"
  scope line is *not* by itself evidence of exclusion — `core.md:889` uses the same
  "Schema-7 external repositories" introduction-naming convention. Discount that
  half of the argument.
- `manager.md:1083-1084` already carries a stale "(currently marker v1 or marker
  v2)" parenthetical from before rc.5, so band-lag has precedent in this document —
  though the precedent was *fixed* by section 11 naming v3, which is the pattern
  that is missing now.

Decide it explicitly, either way, and record the reasoning:

- **Fix** — the cheapest form is one sentence (e.g. in section 11 or at
  `manager.md:1083`) stating that a schema-8 installation records marker v4 and
  that every marker-v3 obligation of this profile — shim selection, read-only
  status validation, GC rooting — applies unchanged to marker v4. This is prose
  outside `conformance/v1`, so the digest stays `803918bf…`, but it **does** require
  a superseding candidate commit and one CI loop.
- **Defer** — write the inheritance argument down in the results artifact as an
  accepted reading, and no re-cut is needed.

Do not leave it decided implicitly by a one-line board note that the green matrix
then contradicts.

### 3.2 Minor, carried over, non-blocking

`schemas/v1/README.md:7` still opens "Rc.8 carries `assurance-policy-v1`, …" while
the body below documents the rc.9 additions (marker v4 at
`schemas/v1/README.md:97-101`, the schema 1-7 field gate at `:89-93`). Flagged
non-blocking in cycle 1 and still open. Fold it in if a re-cut happens anyway.

### 3.3 Observation, no action required

`conformance/README.md:61-62` now reads "published in **every** marker role the
suite exercises" and then lists v1, v2, v4. The suite also exercises a marker-v3
golden (`conformance/v1/expected/external-repository/install-marker-v3-mixed.json`),
covered by a different bullet at `:86-87`. "Every" is a slight overreach; "each
normalized marker role" would be exact. Cosmetic.

## 4. Rework packet

Board-only, no repo change, no CI:

1. Rewrite `TASK-260822-c0rxj7_results.md` to the delivered state — candidate
   `e66cb72d9988c614c7232af9195bf829c82d328e`, protocol `1.0.0-rc.9`, manifest
   `sha256:803918bf…b44403`, tree `sha256:9d5a10b6…b769` (note: `LC_ALL=C`), 692
   files, `release/1.0.0-rc.8.json` `293f101d…e31ede` byte-identical to main, spec
   CI `32654392587`, candidate lane `32654422338` (attempt 2 conclusion success;
   all three Candidate suite jobs green on attempt 1). Keep `859727b`/`782d6868`
   and `edd0721`/`803918bf` recorded as superseded history, not as the candidate.
   Carry forward the `candidate-suite.sh` locale finding.
2. Correct `TASK-260822-c0rxj7_e66cb72-green-matrix.md`: the `profiles/manager.md`
   delta is the merged `517a130` operator-credential prose, not a marker-v4
   obligation surface.
3. Refresh the unblock notes on `TASK-260822-f4qv7w` and `TASK-260822-1so0ym` —
   both currently name `edd0721` / run `32651139699` as the green identity, which is
   now superseded by `e66cb72` / `32654422338`. Both of their named external inputs
   are satisfied either way, so this is accuracy, not a gate.

Then decide 3.1:

4a. If **fixing** — add the manager.md sentence, fold in 3.2, regenerate vectors
    twice (expect no change; digest must stay `803918bf…`), cut the superseding
    immutable commit, record its identity without rewriting `e66cb72`, re-run
    curator-spec Specification CI on the new SHA, re-dispatch curator
    candidate-conformance with the **new ref and the unchanged digest**,
    `SPEC_PIN` untouched. Curator `main` needs nothing — one CI loop, not a fix loop.
4b. If **deferring** — record the inheritance rationale in the rewritten results
    artifact. `e66cb72` stays the qualified candidate; no re-cut, no CI.

Do **not** re-cut for items 1-3 — those are board artifacts and cannot change the
candidate.

## 5. Not defects

- `TASK-260822-f4qv7w` and `TASK-260822-1so0ym` are still `blocked` only because
  the board edge points at this task and releases when it reaches `done`.
- Not merging to `main`, and preserving branch and worktree, is correct per the
  TASK-260823-omp8zt re-scope.
- No Implementation conformance run on the candidate branch, by design.
- The `Test (windows-latest)` attempt-1 failure is a non-candidate job and matches
  the previously proven flake pattern; it does not touch the candidate evidence.
