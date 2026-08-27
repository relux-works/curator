# TASK-260728-2kp3tv developer handoff evidence

Replace the implementation-branded repository descriptor with the
manager-neutral `skill-build.json` name throughout the unreleased schema-7 /
rc.5 candidate contract. No alias, no compatibility behavior.

## 1. Workspace and provenance

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`
- Base: fresh detached worktree created from the pinned predecessor base
  `57c1f56846d221ecc55786bd3c2467ec32f11730` (`git worktree add --detach`).
- Import: byte copy of the accepted `TASK-260728-zb2s4z` candidate state
  (`rsync -a --delete --exclude=.git`). Immediately after import,
  `diff -r --exclude=.git` between source and destination returned exit 0, so
  the starting bytes were exactly the accepted predecessor candidate.
- Predecessor worktree
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
  was read only. Re-checked after all work: HEAD still
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, still 127 uncommitted paths, and
  its manifest digest is still
  `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`.
- Provenance note for reviewers: `TASK-260728-zb2s4z_results.md` records the
  predecessor manifest pin as `sha256:bfe49f25...`. That value is from the
  pre-rework cycle. The accepted post-rework predecessor state on disk pins
  `sha256:58f8d229...` in both `candidate_protocol_pin.manifest_sha256` and
  `downstream_consumption.required_manifest_sha256`, and
  `TASK-260728-zb2s4z_rework-1-results.md` names that same value. The
  `58f8d229...` state is the baseline this task started from.
- The assigned worktree HEAD remains `57c1f56846d221ecc55786bd3c2467ec32f11730`
  with zero commits after the pin (`git rev-list --count <pin>..HEAD` = 0) and
  an unstaged index (`git diff --cached --quiet` exit 0). Nothing was staged,
  committed, tagged, pushed, or published.
- The disposable clean release probe
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260728-2kp3tv/release-probe`
  is a scratch repository used only to exercise the real clean-checkout gates.
  Its synthetic commit `bfbfc4be7bdafb8756b3218d2f51f0903d88fb1d` is not a
  protocol, implementation, or downstream pin.

## 2. Exactly what was renamed, and what was deliberately not

Renamed everywhere (content plus paths):

| Retired identifier | Neutral identifier | Occurrences |
| --- | --- | --- |
| `curator-build.json` | `skill-build.json` | 41 |
| `curator-build-v1.schema.json` | `skill-build-v1.schema.json` | 22 |
| `curator-build-v1` (registry id, schema-case directory) | `skill-build-v1` | 26 |
| `curatorBuildTargetV1` (`common.schema.json` `$defs`) | `skillBuildTargetV1` | 4 |
| `validCuratorBuildV1` (generator helper) | `validSkillBuildV1` | 3 |
| `curatorBuildV1SchemaExamples` (generator helper) | `skillBuildV1SchemaExamples` | 2 |

Path renames: `schemas/v1/curator-build-v1.schema.json` ->
`schemas/v1/skill-build-v1.schema.json`, and
`conformance/v1/schema-cases/curator-build-v1/` ->
`conformance/v1/schema-cases/skill-build-v1/` (13 cases: 3 valid, 10 invalid).

Deliberately **not** renamed: `curator-build-source-v1` (113 occurrences) and
its negative-fixture mutation `curator-build-source-v2` (4 occurrences). That
is the schema-6 **build-source snapshot digest algorithm** identifier, a
different concept from the repository descriptor, and it is bound into rc.4
byte-frozen artifacts — notably
`conformance/v1/schema-cases/install-marker-v2/valid.json`, digest-pinned at
`538d12bb89d2d15259bbb378efc9e6496fe7de195af82099da36fd9d7a1e2c73`. Renaming it
would have broken the "schemas 1-6 byte-for-byte" constraint. The rename was
applied with an explicit exclusion for that stem, and both new guards keep the
exclusion explicit rather than incidental.

Surfaces covered: `protocol/core.md`, `profiles/manager.md`, `SECURITY.md`,
`COMPATIBILITY.md`, `CHANGELOG.md`, `decisions/0004`, `decisions/0005`,
`docs/external-build-repositories.md`, `schemas/v1/README.md`,
`schemas/v1/common.schema.json`, the descriptor schema, the suite manifest, the
schema-case index, every generated receipt-v2 and marker-v3 case, the external
expected fixtures, the `go-host-execution-policy` vector, the generator, the
Python validator, the release gate, and `release/1.0.0-rc.5.json`. The CI
workflows carry no descriptor path reference.

## 3. Normative and documentation changes beyond the mechanical rename

- `protocol/core.md` §4.2.2: new paragraph fixing the descriptor filename and
  location — `skill-build.json` at the repository root and nothing else — and
  stating that a manager MUST NOT read, accept, or search for a descriptor
  under any other filename, directory, or implementation-specific name, and
  MUST NOT treat any such file as an alias.
- `decisions/0005` §"Manifest and descriptor ownership": records why the name
  is manager-neutral (the descriptor lives in a source repository the skill
  author usually does not own and is read by any conforming manager) and that
  schema 7 being unreleased means one name is defined with nothing to migrate.
- `CHANGELOG.md` rc.5 "Added": now names `skill-build.json` explicitly and
  states the neutrality rationale. No "renamed" entry was added, because the
  retired name never shipped — claiming a rename would fabricate released
  history.

Deliberately **not** added: a rule that an absent or unreadable root descriptor
is a hard failure. That would introduce a failure mode with no row in the
`profiles/manager.md` §11 diagnostic table, which is beyond a rename. Existing
ordering (`protocol/core.md`: parse the root descriptor before audit success,
artifact-cache lookup, or a compiler child) and the existing
`build_repository_descriptor_invalid` code are unchanged.

## 4. How rejection is enforced (no alias anywhere)

1. **Wire**: `common.schema.json` `$defs.repositoryDescriptorSelectionV1.path`
   is a bare `{"const": "skill-build.json"}`. Any other descriptor path is a
   schema rejection in build receipt v2.
2. **Marker**: `buildRecordV2` exposes `descriptor_target` and no descriptor
   path at all, with `additionalProperties: false`, so the retired name is not
   expressible in install marker v3; the descriptor bytes bind transitively
   through `receipt_sha256`.
3. **Normative**: the new §4.2.2 MUST NOT-alias paragraph.
4. **Validator**: `tools/validate.py` gains
   `validate_repository_descriptor_identity`, wired into `main()`. It checks
   the descriptor schema `$id`/`title`, the path const, that the generated
   receipt-v2 example selects the neutral name and validates, that mutating
   that example's descriptor path to the retired name is rejected by the real
   compiled validator, that marker v3 cannot express a descriptor path, and
   then scans every repository surface for the retired stem.
5. **Release gate**: `tools/release_gate.py` gains
   `validate_repository_descriptor`, wired into `main()`. It requires the
   neutral descriptor schema to exist and rejects a candidate whose schemas,
   suite manifest, `CHANGELOG.md`, `COMPATIBILITY.md`, or release metadata
   names the retired descriptor.
6. **Go guard**: `TestRepositoryDescriptorIsManagerNeutral` walks the whole
   repository for the same condition and asserts the retired schema file and
   case directory are gone.

In the Go test, the Python validator, and the release gate, the retired stem is
assembled from parts (`"curator" + "-build"`) so each guard scans its own
source file without matching itself. Consequently the retired literal appears
**nowhere** in the tree; the final census over the whole worktree is exactly
113 `curator-build-source-v1` and 4 `curator-build-source-v2`, and zero of
anything else.

## 5. Identity effect: a miss, not an alias

The descriptor path is part of the external build input, so the neutral name is
a real cache-identity revision of the unreleased candidate. Each key is
`"sha256:" || lowercase_hex(SHA-256(CCJ-1(input)))`, recomputed independently
from the stored input:

| Identity | Pre-rename | After rename |
| --- | --- | --- |
| external `go-repository-v1` cache key | `sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492` | `sha256:4abc903bde7d8d9f65d32fd276f37dadccc88eb28bbaf693106dcebc4a19107a` |
| external receipt hash in the mixed marker | `sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea` | `sha256:0f8f910a2b6ba9b35531bb232cb2890e11eb55a64ba01bcdd2d93d5ea421d0e0` |
| local `go-v1` receipt example digest | `1a887eb6bb436a3491250b0814dded2a1b1d108640ba67837ba9e89b1183daf3` | `1a887eb6bb436a3491250b0814dded2a1b1d108640ba67837ba9e89b1183daf3` (unchanged) |

Both new cache keys were verified to equal the CCJ-1 digest of their own stored
input. Local `go-v1` builds carry no descriptor, so their identities and bytes
are untouched — which is what proves the change is scoped to the external
driver.

## 6. Preserved invariants

All nine rc.4 byte-frozen artifacts re-verified after every edit and after the
gates, digests unchanged:

| Path | SHA-256 |
| --- | --- |
| `schemas/v1/agent-skill-v6.schema.json` | `982832e410f85e415e16e8f9104c3b9af23f6d846bbfbe5497ff170dde947f6f` |
| `schemas/v1/csk-skill-v6.schema.json` | `2148eafc4fa110311b52f528651424e2f53c69042235338fb2c8b414035eab9c` |
| `schemas/v1/build-receipt-v1.schema.json` | `f673a8815f5a5f752bc5b612f20c4ba63d9e8dcce61f5af6e7afe11b131c7ab9` |
| `schemas/v1/install-marker-v2.schema.json` | `6d7b65dbdf684272815fb0e61cc4eb02103d09dfdd397de948bd836293debeb2` |
| `schemas/v1/conformance-claim-v2.schema.json` | `4c05a97a1aa9f7dafe629a406a853239928413e79e95488ac2b20ebd0c52a38c` |
| `conformance/v1/schema-cases/agent-skill-v6/valid.json` | `cf029927b7032aaad2fb17931133a897fc8183cce3d091df7321912ad152d634` |
| `conformance/v1/schema-cases/csk-skill-v6/valid.json` | `cf029927b7032aaad2fb17931133a897fc8183cce3d091df7321912ad152d634` |
| `conformance/v1/schema-cases/install-marker-v2/valid.json` | `538d12bb89d2d15259bbb378efc9e6496fe7de195af82099da36fd9d7a1e2c73` |
| `conformance/v1/schema-cases/conformance-claim-v2/valid.json` | `f7e7cc86f33ea03ee9bb4d149e1dba29cf34f5ceaf5504df8a9e91c659a1835f` |

The existing `TestRC4CompiledArtifactsRemainByteFrozen` guard was not weakened
or edited. Local `build-receipt-v1` and `install-marker-v2` cases are entirely
absent from the diff against the predecessor. Command, target, output and
driver ownership are unchanged: no manifest, descriptor, or receipt field was
added, removed, or widened, and no package-controlled surface appeared. Suite
inventory is unchanged at 42 schemas and 422 manifest-listed files.

Diff against the accepted predecessor: 71 files changed, one schema file and
one schema-case directory renamed, zero files added or removed.

## 7. Tests added

Go (`tools/generate-vectors/main_test.go`):

- `TestRepositoryDescriptorIsManagerNeutral` — descriptor schema `$id`/`title`,
  the neutral `$defs` target, the bare path const, the generated receipt
  selection, non-empty registry coverage, absence of the retired schema file
  and case directory, and a full-repository scan for the retired stem that
  explicitly permits the frozen build-source namespace and asserts that
  namespace is still present.
- `TestDescriptorRenameIsACacheIdentityRevision` — the external cache key is
  the exact CCJ-1 digest of its own input, differs from the pinned pre-rename
  key, the mixed marker no longer carries the pre-rename receipt hash, and the
  local `go-v1` receipt example is byte-identical to its predecessor digest.

Python (`tools/test_validate.py`, class `RepositoryDescriptorIdentityTests`):

- `test_candidate_names_only_the_neutral_descriptor`
- `test_scanner_keeps_the_frozen_build_source_algorithm` — the offset scanner
  ignores `-source-v1` and `-source-v2` and catches the bare stem
- `test_receipt_v2_rejects_the_retired_descriptor_name` — real jsonschema
  rejection on a mutated copy of the generated example
- `test_absence_guard_fires_when_the_retired_name_returns` — plants the retired
  name in a stand-in surface and requires the guard to raise
- `test_descriptor_rename_misses_the_pre_rename_external_identity` — Python
  mirror of the non-aliasing proof
- `test_marker_v3_cannot_express_any_descriptor_path`

Python (`tools/test_release_gate.py`):

- `test_release_surfaces_name_only_the_neutral_descriptor` — accepts a clean
  candidate, accepts a schema carrying the frozen build-source algorithm,
  rejects release metadata naming the retired descriptor, and rejects a
  candidate missing the neutral descriptor schema.

Test totals: Python 22 -> 29; Go package tests all pass.

## 8. Guards proved to fire (negative probes)

- Planted `docs/negative-probe.md` containing the retired name in the assigned
  worktree. `go test ./tools/... -run TestRepositoryDescriptorIsManagerNeutral`
  failed with `docs/negative-probe.md:1: retired repository descriptor name is
  not an alias and must be absent`; `python tools/validate.py` failed with the
  same message, exit 1. The probe file was then removed and both went green.
- Committed a `descriptor_probe` field carrying the retired name into
  `release/1.0.0-rc.5.json` in the disposable probe repository.
  `python tools/release_gate.py --version 1.0.0-rc.5 --commit HEAD` failed with
  `release/1.0.0-rc.5.json names the retired repository descriptor, which is
  not an alias`, exit 1. The probe was reset with `git reset --hard` back to
  its clean commit and the gate went green again.

## 9. Exact rc.5 candidate identity after the rename

- Protocol version: `1.0.0-rc.5` (unchanged)
- New candidate protocol pin (SHA-256 of `conformance/v1/manifest.json`):
  `sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`
  (predecessor was `sha256:58f8d229...`)
- Release metadata SHA-256: `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`
- Independent recomputation: `shasum -a 256 conformance/v1/manifest.json`
  equals both `candidate_protocol_pin.manifest_sha256` and
  `downstream_consumption.required_manifest_sha256` in
  `release/1.0.0-rc.5.json`.
- Downstream environment remains `CURATOR_CONFORMANCE_ROOT`;
  `committed_release_pin_advanced` remains `false`; claim-v3 candidate claims
  remain empty; macOS and Windows remain pending downstream native evidence;
  Linux remains excluded until `TASK-260728-1skseh`. The execution-policy block
  is untouched.
- Other exact digests: `schemas/v1/skill-build-v1.schema.json`
  `5710e4cb7cb105892fac9ef8eaa40cf965b33e9a8a75eedab62c9c96d9bc4ac6`;
  `schemas/v1/common.schema.json`
  `4c725692979f64104664f51391f7f99b63d84dedc350c34d92a9fa0e3a620190`;
  `conformance/v1/schema-cases/build-receipt-v2/valid.json`
  `1a519db0c46f68e50d72d76d94cb77d403798e005099fcc04091ed1b98b961bd`;
  `conformance/v1/expected/external-repository/install-marker-v3-mixed.json`
  `20a1ef924f54f8c6dab582a78c22be0778b2f08c17f1cedaeb8016efd624ddd0`.

## 10. Gates and exact exit codes

Every command below was run directly as a standalone process, not through
`tee` or a pipe. Green gates in the assigned worktree:

| Command | Exit |
| --- | --- |
| `go run ./tools/generate-vectors -root .` | 0 |
| `python tools/validate.py` (task venv) — `validated 42 schemas and 422 vector files` | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` — 29 tests, OK | 0 |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `test -z "$(gofmt -l tools)"` | 0 |
| `python -m compileall -q tools` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |
| `git rev-list --count 57c1f568...:HEAD` = 0 | 0 |

Deterministic regeneration in the assigned worktree: three consecutive
`go run ./tools/generate-vectors -root .` runs produced the identical aggregate
digest `758c24e88cc1bec27ad5ad6583f917c546559a9cdeab613bb3c8f6216f3ab47f` over
every file under `conformance/v1` and `release/`.

Clean disposable release probe at commit
`bfbfc4be7bdafb8756b3218d2f51f0903d88fb1d`, byte-identical to the assigned
worktree before and after the gates (`diff -r --exclude=.git` exit 0 both
times):

| Command | Exit |
| --- | --- |
| `make regenerate-check`, first run | 0 |
| `make regenerate-check`, second consecutive run | 0 |
| `make release-check VERSION=1.0.0-rc.5` — validation, Python, Go, regeneration, exact metadata pin, execution-policy honesty, descriptor identity, clean checkout, version, candidate gate | 0 |
| `git status --porcelain` after the gates — zero lines | 0 |

Expected-red gates, reported truthfully as failures:

1. `python tools/validate.py` immediately after wiring the first version of the
   absence guard: exit 1, `conformance/v1/schema-cases/build-receipt-v1/
   invalid-build-source-algorithm.json:11: retired repository descriptor name
   is not an alias and must be absent`. Cause: the first draft allowed only the
   exact `-source-v1` algorithm, but the frozen negative fixture mutates the
   version suffix to `-source-v2`. Closed by allowing the whole build-source
   algorithm namespace, not by weakening the descriptor check.
2. `python -B -m unittest discover -s tools` after adding the absence-guard
   mutation test: exit 1, one error — `ValueError: ... is not in the subpath
   of ...` from `Path.relative_to` when the planted surface lived outside the
   repository root. The guard fired correctly; only the failure message
   formatting was wrong. Closed by adding a `display_path` fallback.
3. The three negative probes in section 8, each expected to be red and each
   red for the intended reason.

Not run, and why: no native macOS or Windows manager execution was exercised.
This task is specification-only and adds no manager implementation, so there is
no worker binary to run. Native qualification remains downstream, which is why
the candidate still emits zero claim-v3 tuples.

## 11. Boundaries preserved

- No Curator or csk implementation code, no manager, no worker binary.
- No new, removed, or widened package-controlled field in any manifest,
  descriptor, receipt, marker, or claim; command, target, output and driver
  ownership are byte-identical in shape.
- No generic language driver, no fallback, no package-selected program, argv,
  environment, output, credential, helper, filter, signer, hook, plugin, or
  generator.
- Independent audit-before-cache/compiler ordering, fail-closed offline
  behavior, the portable `manager-worker-v1` execution policy, and
  operator-owned signing are unchanged.
- No alias, no dual-name read path, and no compatibility shim was added.
- No real remote was contacted and no release, pin, commit, or platform claim
  was fabricated.
- No source repository, predecessor worktree, release, ref, tag, or downstream
  pin was modified.
