# TASK-260822-c0rxj7 — Schema 8 / rc.9 candidate evidence

Supersedes the earlier revisions of this artifact. Two prior candidate
identities exist and are deliberately left unrewritten in history:
`859727b103ed175ff214cbb64641f4686d8c6a68` (manifest `782d6868…`) and
`edd07210d4f3db34fd60238cb14b90f837de03cb` (manifest `803918bf…`). This
document records the current candidate.

## 1. Candidate identity

- Repository: `relux-works/curator-spec`
- Branch: `candidate/schema-8-rc.9`
- Full commit SHA: `e66cb72d9988c614c7232af9195bf829c82d328e` (signed)
- Protocol version: `1.0.0-rc.9`
- Suite manifest SHA-256: `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`
- Tree SHA-256: `sha256:9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769`
- File count: `692`
- Historical `release/1.0.0-rc.8.json`: `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`, byte-identical to `curator-spec` `origin/main`
- `SPEC_PIN` in `curator`: unchanged at `00b1688a9b2457ca397a0bb550acf47cad8ee967`

The manifest and tree digests are deliberately identical to `edd0721`: this
revision is prose only and touches nothing under `conformance/v1`.

Ancestry (each verified with `git merge-base --is-ancestor`, exit 0):

| Input | Commit |
| --- | --- |
| script-worker vectors | `dd9c9fc079470f03f247b71efb52d4de6b204e78` |
| module-roots vectors | `bac193cadb7d26aabf006c92924b4a05f6574e31` |
| schema branch | `ebfed81` |
| core prose | `41cf556` |
| manager/security prose | `e5df43d` |
| curator-spec `main` | `517a130` |

## 2. What this revision adds over `edd0721`

Two commits:

- `586ebd9` — merge of `curator-spec` `origin/main` `517a130` (operator
  credential selections are never lockable; `profiles/manager.md` prose only).
  It was the single `main` commit absent from `edd0721`.
- `e66cb72` — the three prose fixes below.

### 2.1 Install marker v4 is now bound in normative prose (review blocker)

The suite already shipped `install-marker-v4.schema.json`
(`skill_schema_version` const 8), the `expected/install-marker-v4.json` writer
golden, and the `schema-cases/install-marker-v4` family, while
`protocol/core.md` never mentioned marker v4. Section 10's only obligation
sentence covered managers supporting schema 7 and required reading marker
schemas 1, 2, and 3. A conforming schema-8 manager would therefore write
marker v4 and then, under the "unsupported or unreadable markers are not
current" rule, report every schema-8 installation as not current while still
passing every stated obligation.

Section 10's first paragraph now reads: managers supporting schema 7 MUST read
marker schemas 1, 2, and 3, and managers supporting schema 8 MUST read marker
schemas 1, 2, 3, and 4; they MUST write marker schema 2 for schema 1 through 6
installation mutations, marker schema 3 for schema 7 installation mutations,
and marker schema 4 for schema 8 installation mutations.

Section 10 also gains the marker-v4 paragraph: v4 permits
`skill_schema_version` 8 and otherwise carries marker-v3 meaning unchanged —
same object shape, same requirement that every build entry record its receipt
schema version and `execution_policy` explicitly, same local `go-v1` and
external `go-repository-v1` entry semantics, same top-level `build_source` and
`build_roots` rules. It states that an enforced `script-worker-v1` script
command produces no build entry and adds no marker member, and that markers
v1, v2, and v3 keep their frozen shapes and manifest-version bands, so a
schema-8 installation is recorded by marker v4 alone.

`profiles/manager.md` was checked for a matching obligation surface and needs
none: its only marker-band statements sit inside the section 11
`go-repository-v1` profile and are correctly scoped to marker v3 there. The
read/write band obligation lives in `protocol/core.md` section 10 alone.

### 2.2 Section 4 version gate for the script-execution fields

`protocol/core.md` section 4 gated `build_roots`, `build_repositories`, and
`modules` but said nothing about `execution_policy` and `interpreter`. Because
the same paragraph exempts schema 1 from unknown-field rejection, core.md alone
could not tell a reader whether a schema-1 manifest carrying `execution_policy`
is rejected or ignored; the rule existed only in `schemas/v1/README.md`.

Added: schema 1 through 7 MUST also reject `execution_policy` and `interpreter`
at the top level and on every command — schemas 2 through 7 as unknown fields,
and schema 1, which keeps its deployed extension behavior, through this
document's semantic checks. This matches the implemented behavior:
`tools/validate.py` `validate_wire_semantics` rejects both field families for
manifest schemas 1 through 7, and `validate.py` asserts every one of those
schemas refuses a schema-8-only command field.

### 2.3 Conformance suite documentation

`conformance/README.md`'s install-marker bullet stopped at the marker-v2 writer
golden while the suite ships the v4 golden. It now lists
`expected/install-marker-v4.json` as the marker-v4 writer golden a manager
writes for every schema-8 installation mutation.

## 3. Local gates

Each command was run as a standalone process; these are real exit codes.

| Gate | Exit | Result |
| --- | ---: | --- |
| `python tools/validate.py` | 0 | 53 schemas, 691 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | 98 tests |
| `go test ./tools/...` | 0 | generate-vectors tool tests |
| `gofmt -l tools` | 0 | no output |
| `go run ./tools/generate-vectors -root .` pass 1 | 0 | both vector families |
| `git diff --exit-code` over `conformance/v1` + rc.5–rc.9 after pass 1 | 0 | no drift |
| `go run ./tools/generate-vectors -root .` pass 2 | 0 | both vector families |
| `git diff --exit-code` over `conformance/v1` + rc.5–rc.9 after pass 2 | 0 | no drift |
| `cmp` of the two 692-entry checksum inventories | 0 | byte-identical |
| `cmp` of the two 692-entry path inventories | 0 | byte-identical |
| rc.8 comparison against `curator-spec` `origin/main` | 0 | byte-identical |

Both inventories are attached as
`TASK-260822-c0rxj7_regeneration-pass1-e66cb72.sha256` and
`TASK-260822-c0rxj7_regeneration-pass2-e66cb72.sha256`.

## 4. Remote qualification

Specification CI run
[32654392587](https://github.com/relux-works/curator-spec/actions/runs/32654392587),
dispatched on the exact candidate SHA: **success**. Formatting, Links, Release
target provenance, and Specification on ubuntu-latest, macos-latest, and
windows-latest all passed.

Curator candidate-conformance run
[32654422338](https://github.com/relux-works/curator/actions/runs/32654422338),
dispatched from `curator` `main` `e17b0f1` with
`candidate_ref=e66cb72d9988c614c7232af9195bf829c82d328e` and
`candidate_manifest_sha256=sha256:803918bf…b44403`, `SPEC_PIN` untouched,
`CI_REQUIRE_FULL_ROOT=1`:

CANDIDATE_MATRIX_PLACEHOLDER

## 5. Finding recorded, not fixed here

`.github/ci/candidate-suite.sh record` computes `tree_sha256` over a `sort`
whose collation is not pinned. The same 692 bytes-identical files measure
`9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769` under
`LC_ALL=C` and
`176dc52bdb73bc57ae394e2a063e9bb80dc3cd8f4c51f75b74c4144a8c942f02` under
`LANG=en_US.UTF-8`. The canonical recorded value is the `LC_ALL=C` one, which
is what the runners produce.

`ci.yml` wires only `CANDIDATE_EXPECTED_MANIFEST_SHA256`, never
`CANDIDATE_EXPECTED_TREE_SHA256`, so this is a non-reproducible evidence field
rather than a live gate hole: a developer reproducing the digest on a macOS
shell gets a false mismatch. It was deliberately not patched during an active
qualification run — changing an evidence gate mid-qualification is worse than
recording the defect. It belongs in a `curator`-side task.

## 6. Scope boundary

This task does not merge to `main` and did not open a PR, per the
`TASK-260823-omp8zt` re-scope: the atomic spec landing with pins advanced is a
later task that runs only after both implementations qualify against this
candidate. The candidate branch and its worktree at
`.temp/TASK-260822-c0rxj7/curator-spec-candidate` are preserved on purpose.
