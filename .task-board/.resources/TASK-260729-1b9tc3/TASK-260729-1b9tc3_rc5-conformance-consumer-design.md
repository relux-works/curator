# CocoaSkills rc.5 conformance consumer — architecture and executable test blueprint

**Board task:** `TASK-260729-1b9tc3` (parent `STORY-260720-1uv5gi`)
**Date:** 2026-07-29
**Class:** read-only design. No CocoaSkills product or test edits, no Go logic copied, no pin or publication change.
**Primary consumer of this design:** `TASK-260720-12r55p` (shared v6 vector consumer). Secondary: `TASK-260720-2dnqw2`, `TASK-260720-2g21eg`, `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5`, `TASK-260720-z9j4c9`, `TASK-260720-3pemm6`.

---

## 1. Immutable root and digest provenance

### 1.1 Authoritative root

```
/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1
```

| Property | Value |
| --- | --- |
| `manifest.json` SHA-256 | `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` |
| `manifest.protocol_version` | `1.0.0-rc.5` |
| `manifest.generated_at` | `2026-07-13T00:00:00Z` |
| `manifest.generator` | `tools/generate-vectors` |
| Manifest entries | 447 (`schema-cases` 377, `expected` 27, `fixtures` 24, `vectors` 19) |
| Files on disk under the root | 448 (447 + `manifest.json`) |
| Aggregate tree digest | `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` |

Aggregate digest is reproducible with:

```bash
cd "$CURATOR_CONFORMANCE_ROOT" && find . -type f -print0 | sort -z | xargs -0 shasum -a 256 | shasum -a 256
```

The manifest digest equals the `required_manifest_sha256` recorded by the accepted `TASK-260729-3nx97g` candidate, and equals the digest already written into the `TASK-260720-12r55p` and `TASK-260720-2dnqw2` acceptance criteria. **The parity-map §3.3 gap is closed in this root:** `vectors/build-drivers.json` and the 11-file `expected/build-driver/` tree are both present and manifest-covered.

### 1.2 Key cluster digests (recorded before analysis)

| Path | SHA-256 |
| --- | --- |
| `vectors/build-drivers.json` | `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea` |
| `vectors/go-host-execution-policy.json` | `c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de` |
| `vectors/manager-lifecycle.json` | `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf` |
| `vectors/conformance-claim-v3-qualification.json` | `c4b9132ce8344d5b210aa7fbc2715a06586ea2efc2b88ae7f6464468e48de2ee` |
| `vectors/canonical-valid.json` | `129891dd4993622dc698fa305fa6ec31c07f118fa9dbce902201874a41e35370` |
| `vectors/canonical-invalid.json` | `6c290244145528b52de2d70a434ff5916ae196fc3a5f5fb9b93f54277e39168c` |
| `vectors/source-identities.json` | `5a11e96878835217f5fa2bcf08f6644fa2e080df2fb7dd12c85b3a0ab5b44642` |
| `vectors/identifiers.json` | `fe903f7dda4df42d9d5f05aa4d7e5bc0d4bff896b1a756e57cde53590c80147d` |
| `vectors/locale-selectors.json` | `dd9299b5baf009ca31eee52744c6fdfe4c2c4595d7e6fe91b7cf547d8a393521` |
| `vectors/manager-config.json` | `e915c1f76729773cab6578105beba19f425154e1a9e200d0206b3f664eaed16c` |
| `vectors/portable-paths.json` | `6be790547304efa8fc1bafb3122c6aa40f74f699f4e35d34379246ca90c3249b` |
| `vectors/closures.json` | `d8032d222c64609067ea66c00403ccfe702349085c230c013e461145779485a5` |
| `vectors/skill-manifest-resolution.json` | `07c42987cdfcbdb0b78a0149373ad8b49ac2d1d70a5484b4e42a54e20586647c` |
| `schema-cases/index.json` | `2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0` |
| `expected/marker.json` | `80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a` |
| `expected/context_files.json` | `72b219e2013df777b89265e959fa941e3e00519849910535781792b10a4262e3` |
| `expected/context_sha256.txt` | `5ec200cfb97f15ae1a379c071a4be7199db24719e20e47ee1ca54b9ee695d43b` |
| `expected/snapshot_sha256.txt` | `6c67a5b415fa432c0d6118bcc66d878da7dec90b65d6a276e200437c083f7ab4` |

### 1.3 `expected/build-driver/` — the 11 byte goldens

| File | Bytes | SHA-256 of the file | Semantic role |
| --- | ---: | --- | --- |
| `build-input.ccj.json` | 869 | `529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` | **is** the cache key preimage |
| `cache-key.txt` | 72 | `b55decf2abc313876c7f4e86473d480fd53cad234c436921cc1dc2cdceae4734` | `sha256:529370…` + `\n` |
| `receipt.ccj.json` | 1120 | `919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd` | **is** the receipt preimage |
| `receipt-sha256.txt` | 72 | `84286d002d90e89cb70cc71e54ab80c44d2770d593a7f330a143ac7506d16a0f` | `sha256:919fbbad…` + `\n` |
| `build-source.preimage.bin` | 2126 | `27cdcac0734aa3e069e95a10341e89b118a07c60002516e7b401e95477f01332` | `curator-build-source-v1` framing |
| `build-source-sha256.txt` | 72 | `f7155f073664e96cbe66cfe08c33cb87d3ea23ecc80bd889fbbd13567236fbd5` | `sha256:27cdcac0…` + `\n` |
| `toolchain.preimage.bin` | 177 | `baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e` | `curator-go-toolchain-v1` framing |
| `toolchain-sha256.txt` | 72 | `cc773f322a60de7a4321e5fa30a43192181bd0c13d80c84a803dfabaf04ba708` | `sha256:baf7c5f3…` + `\n` |
| `marker.json` | 1339 | `feae3ffbe4e6c9bed17a6f077702c523bf6b0c7783edcef9716fddaa3d62502b` | install-marker v2 golden |
| `context_files.json` | 39 | `a1000f86bd8f28bf7001dca84e94151095013788d6d006260d4e57b7808232b2` | `["SKILL.md","assets/prompt.md"]` |
| `context_sha256.txt` | 72 | `8cdd48625fb1bdc1a4f20f8ba3d2516f04e55a91bdc66a389906cdccccc94634` | `sha256:82c0a18e…` + `\n` |

The four `*-sha256.txt` / `cache-key.txt` / `context_sha256.txt` files are **`sha256:<hex>` plus one trailing `\n`**. Every reader must `.strip()`. The four `*.ccj.json` / `*.preimage.bin` files carry **no** trailing newline; the `.txt` files do. This asymmetry is a designed negative surface (`receipt_not_canonical` covers `noncanonical-receipt-trailing-lf`), so the harness must read them as **bytes**, never as text with universal-newline translation.

### 1.4 Structural constraint — schemas live outside the root

`conformance/v1/manifest.json` covers only `schema-cases/`, `expected/`, `fixtures/`, `vectors/`. The JSON Schema files are at `<root>/../../schemas/v1/` and are **not** digest-covered by the conformance manifest. `schema-cases/index.json` references them by bare filename (`"schema": "agent-skill-v6.schema.json"`), and they `$ref` `common.schema.json`.

Consequence for the design: **the consumer must not depend on JSON Schema validation.** CocoaSkills has zero runtime dependencies and no `jsonschema` dev dependency; adding one would introduce a second, non-independent oracle. See §4.4 for how the schema-derived assertions in the `TASK-260720-12r55p` AC are satisfied without reading the schema files.

---

## 2. Measured baseline — what the current consumer does with this root

### 2.1 Previous default pin (must be preserved)

`.github/workflows/ci.yml` checks out `relux-works/curator-spec` at ref `cbe912d064e06275b0a1aa6762b7c31f687051c5` into `protocol-spec/`, and exports
`CURATOR_CONFORMANCE_ROOT=${{ github.workspace }}/protocol-spec/conformance/v1`.

Measured content of that pin:

| Property | Value |
| --- | --- |
| `protocol_version` | `1.0.0-rc.2` |
| `manifest.json` SHA-256 | `728f772950414b9c3ddf38a8f1e9f2c7d2953bdca1d8c135c7e1a9abf40fff06` |
| Manifest entries | 81 |
| `vectors/` present | 12 (no `build-drivers.json`, no `go-host-execution-policy.json`, no `manager-lifecycle.json`, no `conformance-claim-v3-qualification.json`) |
| `expected/build-driver/` | absent |
| Fixture manifest filename | `csk-skill.json` |

This is the **released** suite the committed default pin must stay on until `TASK-260720-25d05o` qualifies the release and `TASK-260720-1utsx8` audits the promotion. The rc.5 root is an **explicit caller-supplied candidate**, never the default.

### 2.2 Current test results

| Root | Command | Result |
| --- | --- | --- |
| pinned rc.2 | `pytest tests/test_protocol_conformance.py -q` | **98 passed** |
| candidate rc.5 | `pytest tests/test_protocol_conformance.py -q` | **1 failed, 97 passed** |

The single failure is `test_shared_fixture_context_hash_and_marker`:

```
files == expected_files
AssertionError: Left contains one more item: 'scripts/golden-tool'
```

### 2.3 Root cause — exact, and not a protocol semantics change

`fixtures/skill/csk-skill.json` (rc.2) and `fixtures/skill/agent-skill.json` (rc.5) are **semantically identical** — same `schema_version: 5`, same `runtime_roots: ["scripts"]`, same script command, same capabilities. Only the filename changed. `expected/marker.json` is byte-identical between the two roots.

`csk.skillspec.load_skill_spec()` reads `snapshot / "csk-skill.json"`, falls back to `agents/runtime.json`, and otherwise returns `SkillSpec(commands={}, source_file=None)` — **silently**. Against the rc.5 fixture it therefore yields an empty spec, so the existing test computes `include_scripts=True` and `exclude_roots=()`, and `whitelist.copy_context` pulls `scripts/golden-tool` into the agent context.

Verified directly:

```
fixtures/skill           -> source_file=None commands=[] schema=1 runtime_roots=()
fixtures/go-build-skill  -> source_file=None commands=[]
```

**This is a hard precondition for every fixture-driven rc.5 assertion.** `fixtures/go-build-skill/agent-skill.json` is schema 6 with `runtime_roots: ["scripts"]` and `build_roots: ["assets/build-tool"]`; a silent empty spec would make the build-root context-exclusion and build-source assertions vacuous instead of red. It is covered by `TASK-260720-z9j4c9`'s stated scope ("canonical and legacy manifest parity") but is **not** in that task's testable AC list — see §10.2 for the routed recommendation.

### 2.4 Independent reproduction feasibility — proved with zero product edits

Using only today's `csk.audit_registry.canonical_bytes` (the existing CCJ-1 primitive) against the rc.5 root:

| Assertion | Result |
| --- | --- |
| `canonical_bytes(portable_identity.build_input)` equals the 869 golden bytes | **True** |
| `canonical_bytes(portable_identity.stored_receipt)` equals the 1120 golden bytes | **True** |
| `sha256(build-input.ccj.json)` equals `cache-key.txt` | **True** |
| `sha256(receipt.ccj.json)` equals `receipt-sha256.txt` | **True** |
| `expected/build-driver/marker.json` equals `portable_identity.marker` | **True** |
| `cache_identity.aliases` | **False** |
| Three cache keys re-derived from their own `input` objects, all matching, all distinct | **True** |
| `protocol_json.loads` accepts the golden receipt bytes | **True** |

The two framed preimages are also fully recoverable in pure Python — a one-byte record tag (`D` directory, `F` file, `L` symlink, `V` version) followed by uint64 **big-endian** path length, path, uint64 big-endian payload length, payload, after a `curator-build-source-v1\0` / `curator-go-toolchain-v1\0` domain prefix. Decoding the published preimage is enough to specify the writer; **no Go source needs to be read or copied**, which satisfies the independence requirement in §6.2/§6.3 of the accepted parity map.

---

## 3. Consumer architecture

### 3.1 Design rules (binding)

1. **Read-only, root-driven.** Every protocol value comes from `CURATOR_CONFORMANCE_ROOT` at run time. No vector input, cache key, error code, or control name is literal in test source, except the two identity digests already frozen into board AC (`529370…`, `919fbbad…`), which are asserted *against* the root as a tripwire.
   This is the explicit anti-pattern from parity-map §3.3: Curator's `internal/godriver/controls_test.go` hard-codes `portableVectorCacheKey` and reconstructs the input in Go. The Python consumer must not mirror that.
2. **Three verbs only.** Harness modules may (a) read bytes/JSON from the root, (b) call a public `csk.*` entry point, (c) compare. No protocol arithmetic, no canonicalization, no framing, no ordering logic inside `tests/`. If an assertion needs a derived value, the derivation belongs in `src/csk/`, owned by the implementing task.
3. **No second oracle.** No `jsonschema`, no vendored schema copies, no reimplementation of Go behavior in test helpers.
4. **Bytes, not text.** All golden comparisons use `Path.read_bytes()`. No `read_text()` on `*.ccj.json` or `*.preimage.bin`.
5. **Missing cluster is a failure, never a skip** — when the root claims a protocol version that should contain it (§4.3).

### 3.2 Module layout

New test-support package, owned by `TASK-260720-12r55p`:

```
tests/conformance/__init__.py       # re-exports the public helpers below
tests/conformance/root.py           # root resolution, manifest verification, provenance recording
tests/conformance/clusters.py       # cluster registry, protocol-version gating, typed accessors
tests/conformance/goldens.py        # expected/** byte accessors (read_bytes, strip-suffix helpers)
tests/conformance/platform.py       # native-control platform gating and skip/xfail policy
```

Test modules (all under `tests/`, all consuming the package above):

| Module | Owns | Status |
| --- | --- | --- |
| `test_protocol_conformance.py` | Existing portable clusters: CCJ, source identity, identifiers, locales, manager config, portable paths, closures, registry client/records, shared skill fixture. Stays green against the pinned rc.2 root. | exists, extended |
| `test_conformance_root.py` | Root provenance, manifest integrity, cluster presence gating | new |
| `test_conformance_schema_cases.py` | `schema-cases/` for both manifest names + receipt/marker/claim | new |
| `test_conformance_build_driver.py` | `build-drivers.json` positives + 77 rejections | new |
| `test_conformance_build_source.py` | `build_source_cases`, fixture context/build-root exclusion | new |
| `test_conformance_toolchain.py` | `toolchain_cases`, fixed environment, argv forms | new |
| `test_conformance_identity.py` | `cache_identity`, `portable_identity`, all 11 `expected/build-driver/` goldens | new |
| `test_conformance_execution_policy.py` | `go-host-execution-policy.json` in full | new |
| `test_conformance_manager_lifecycle.py` | `manager-lifecycle.json` | new |
| `test_conformance_claim.py` | claim-version separation + `conformance-claim-v3-qualification.json` | new |

**Rationale for splitting rather than growing one file:** `test_protocol_conformance.py` is 390 lines / 98 tests today; the rc.5 surface adds roughly 320 parametrised cases. One module would make `-k` targeting, per-cluster skip policy, and per-task ownership during the 17-task chain impractical, and would couple the legacy rc.2-green module to rc.5-only imports. `TASK-260720-12r55p`'s scope line ("own `tests/test_protocol_conformance.py` plus focused vector adapters and fixtures") is read as permitting sibling modules; this is an explicit design decision, recorded here so review can reject it cheaply if the owner disagrees.

### 3.3 Product boundaries the consumer calls

| Product surface | Module | State today | Owning task |
| --- | --- | --- | --- |
| Manifest / schema-6 build model | `src/csk/skillspec.py`, `src/csk/skillcheck.py` | schemas 1–5, `csk-skill.json` only | `z9j4c9` |
| Build-source identity + context boundary | `src/csk/builds/source.py` (new), `src/csk/hashing.py`, `src/csk/whitelist.py` | `hashing.content_sha256` exists, marker-excluding; no framed digest | `3c0ss2` |
| Toolchain identity | `src/csk/builds/toolchain.py` (new) | absent | `3j8pp5` |
| Canonical input / receipt / marker | `src/csk/builds/metadata.py` (new), `src/csk/protocol_json.py` | CCJ-1 primitive exists in `audit_registry.canonical_bytes` | `2dnqw2` |
| go-v1 driver + worker | `src/csk/builds/go_v1.py` (new), worker entry point | absent | `2g21eg` |
| Protected cache | `src/csk/builds/cache*.py` (new) | absent | `2jfnz6`, `8nxlgx` |
| Currentness / repair / GC | `src/csk/status.py`, `src/csk/gc.py` | pre-v6 | `th0jdi` |
| Lifecycle: bootstrap, dry-run, upgrade, launcher | `src/csk/cli.py`, `installer.py`, `shims.py`, `global_install.py` | exists, pre-v6 | `2x6mjn`, `3t8nr3`, `g7kgox`, `11yhth` |

**`csk.audit_registry.canonical_bytes` is the CCJ-1 primitive the consumer uses today.** `TASK-260720-2dnqw2` scopes "narrowly shared CCJ-1 support in `src/csk/protocol_json.py`". Design note: when that move happens, `audit_registry` must delegate to the shared primitive rather than keep a copy, or the registry and build surfaces will drift. The harness should import the *shared* symbol so the move is a one-line change here.

---

## 4. Root resolution, provenance, and cluster gating

### 4.1 Input contract (preserves the existing surface)

| Variable | Required | Meaning |
| --- | --- | --- |
| `CURATOR_CONFORMANCE_ROOT` | no | Absolute path to a `conformance/v1` directory. **Unchanged semantics:** unset ⇒ whole conformance suite skips, exactly as today (`pytestmark = pytest.mark.skipif(not ROOT_TEXT, …)`). |
| `CURATOR_CONFORMANCE_MANIFEST_SHA256` | no | When set, `test_conformance_root.py` asserts the root's `manifest.json` digest equals it. The rc.5 CI job sets it to `b6f56aac…`. Unset ⇒ digest is recorded, not asserted. |

No other input is added. The committed CI default keeps pointing at the pinned `relux-works/curator-spec@cbe912d0` checkout; the rc.5 candidate arrives only as an explicit caller-supplied root in a separate job (see §9.3).

### 4.2 Provenance recording (non-release evidence)

`tests/conformance/root.py` exposes `provenance()` returning `{root, manifest_sha256, protocol_version, file_count, generated_at, generator}`. A `pytest_report_header` hook in `tests/conftest.py` prints one line:

```
curator-conformance: root=<abs> protocol=1.0.0-rc.5 manifest=sha256:b6f56aac… files=447
```

This satisfies "records the suite digest as non-release evidence" in `12r55p`, `2dnqw2`, and `3pemm6` without writing a new artifact file or touching release metadata. It appears in every CI log, including the pinned-rc.2 job (where it will read `protocol=1.0.0-rc.2 manifest=sha256:728f7729…`).

`test_conformance_root.py::test_manifest_integrity` re-hashes all 447 manifest entries and asserts each matches, and asserts the on-disk file set equals the manifest set plus `manifest.json`. This is ~1 MB of hashing — negligible — and it is the only defence against a partially materialised or mutated candidate root.

### 4.3 Cluster gating — missing means fail, not skip

The failure mode this design exists to prevent is documented in parity-map §3.3: Curator's `TestCandidateBuildMetadataArtifacts` **skipped** when `expected/build-driver` was absent and when the input lacked `execution_policy`, so for the whole rc.5 candidate line "no candidate root exercised the byte-exact build-driver artifact suite" while every gate reported exit 0.

`tests/conformance/clusters.py` therefore declares, per cluster, the protocol version that introduced it:

| Cluster | Path | Introduced |
| --- | --- | --- |
| portable core | `vectors/canonical-*.json`, `identifiers`, `locale-selectors`, `manager-config`, `portable-paths`, `closures`, `source-identities`, `registry-*` | ≤ rc.2 |
| shared skill fixture | `fixtures/skill`, `expected/context_*`, `expected/marker.json` | ≤ rc.2 |
| schema cases v6 | `schema-cases/{agent,csk}-skill-v6/` | rc.4 |
| build driver | `vectors/build-drivers.json`, `expected/build-driver/`, `fixtures/go-build-skill` | rc.5 (this root) |
| execution policy | `vectors/go-host-execution-policy.json` | rc.5 |
| manager lifecycle | `vectors/manager-lifecycle.json` | rc.4 |
| claim v3 | `schema-cases/conformance-claim-v3/`, `vectors/conformance-claim-v3-qualification.json` | rc.5 |

Resolution rule, in this order:

1. Root protocol version **below** the cluster's introduction ⇒ `pytest.skip` with the exact reason, naming both versions and the cluster. Legacy rc.2/rc.3 roots stay green with honest, self-describing skips.
2. Root protocol version **at or above** ⇒ the cluster files **must** exist. Absence raises `AssertionError("cluster 'build driver' missing from a 1.0.0-rc.5 root")`. Never a skip.
3. Present ⇒ run.

Version comparison uses an explicit ordered tuple of known protocol versions declared in `clusters.py`, not string ordering (`1.0.0-rc.10` must not sort below `1.0.0-rc.2`).

### 4.4 Schema-derived assertions without reading schema files

`TASK-260720-12r55p`'s AC requires asserting that "`conformance-claim-v3` pins `protocol_version` `1.0.0-rc.5` and requires `build_drivers`", and claim-version separation across v1/v2/v3. §1.4 rules out reading `schemas/v1/*.schema.json` as a manifest-covered input.

Satisfied entirely from manifest-covered `schema-cases/`:

- `conformance-claim-v3/valid.json` and `valid-macos-only.json` carry `protocol_version: "1.0.0-rc.5"` and a `build_drivers` array ⇒ assert the const and the required member from the positive cases.
- `conformance-claim-v3/invalid-rc4.json` ⇒ the rc.4 protocol version is rejected under claim v3.
- `conformance-claim-v3/invalid-missing-execution-policy.json`, `invalid-hardened-execution-policy.json`, `invalid-generic-driver.json`, `invalid-linux-unqualified.json`, `invalid-duplicate-platform.json`, `invalid-duplicate-driver-assertion.json`, `invalid-driver-platform-outside-claim.json`, `invalid-language-mismatch.json`, `invalid-unknown-field.json`, `invalid.json` ⇒ the 13-case claim-v3 negative surface.
- `conformance-claim-v1/` (2) and `conformance-claim-v2/` (7) ⇒ version separation: a v2 instance must not validate as v3 and vice versa.

The oracle is `csk`'s own claim reader (owned by `12r55p`), not a schema library. **No conformance claim is emitted by any of this work** — assertion only, per the `12r55p` AC.

---

## 5. Cluster → test → product boundary map

Counts are measured from the root at the digests in §1.

### 5.1 `vectors/build-drivers.json` (`f412c107…`)

| Block | Count | Test module :: test | Product entry point | Impl task |
| --- | ---: | --- | --- | --- |
| `positive_cases` | 8 | `test_conformance_build_driver.py::test_positive_case` | see §5.1.1 | mixed |
| `rejection_cases` | 77 | `test_conformance_build_driver.py::test_rejection_case` | see §5.1.2 | mixed |
| `build_source_cases` | 10 | `test_conformance_build_source.py::test_build_source_case` | `csk.builds.source` | `3c0ss2` |
| `toolchain_cases` | 12 | `test_conformance_toolchain.py::test_toolchain_case` | `csk.builds.toolchain` | `3j8pp5` |
| `argv` | 5 | `test_conformance_toolchain.py::test_argv_form` | `csk.builds.go_v1` argv builder | `2g21eg` |
| `fixed_environment` | 28 keys | `test_conformance_toolchain.py::test_fixed_environment` | `csk.builds.go_v1` env builder | `2g21eg` |
| `cache_identity` | 3 + `aliases` | `test_conformance_identity.py::test_cache_identity` | `csk.builds.metadata` | `2dnqw2` |
| `portable_identity` | 1 | `test_conformance_identity.py::test_portable_identity_*` | `csk.builds.metadata`, marker | `2dnqw2` |
| `fixture` | 1 | `test_conformance_build_source.py::test_fixture_context_boundary` | `skillspec`, `whitelist`, `hashing` | `z9j4c9`, `3c0ss2` |

#### 5.1.1 The 8 positive cases

| Case | Asserted behaviour | Product boundary | Task |
| --- | --- | --- | --- |
| `schema-6-mixed-script-and-build-commands` | manifest with `type: build` + `type: script` in one skill parses; `build_roots` accepted | `skillspec.load_skill_spec` | `z9j4c9` |
| `build-root-excluded-from-agent-context` | `expected_context_files` / `expected_excluded_files` hold on real install, cache hit, and dry-run; `dry_run_source_aware_go_commands` and `cache_hit_source_aware_go_commands` are both **0** | `whitelist.copy_context` + planner | `3c0ss2`, `2x6mjn` |
| `valid-standard-library-only-main` | stdlib-only main package accepted | `go_v1` graph validation | `2g21eg` |
| `valid-vendor-only-main-with-transitive-embed` | vendored + transitive embed accepted | `go_v1` graph validation | `2g21eg` |
| `fixed-environment-and-five-direct-argv-forms` | exactly the 5 argv forms; `shell_used` false; `artifact_executed` false | `go_v1` | `2g21eg` |
| `portable-execution-policy-is-required-input` | `policy.execution_policy = manager-worker-v1`; `package_selectable` false; key is `529370…` | `builds.metadata` | `2dnqw2` |
| `protected-cache-hit` | hit yields `source_aware_go_commands == 0`, `artifact_executed` false, `protected_boundary_verified` true, receipt `919fbbad…` | cache + driver | `2jfnz6`/`8nxlgx` |
| `compiler-free-dry-run-miss` | miss plans without any Go invocation; `persistent_effects` empty | planner | `2x6mjn` |

#### 5.1.2 The 77 rejection cases, by boundary

| Boundary | Cases | Test entry | Product boundary | Task |
| --- | ---: | --- | --- | --- |
| `manifest` | 8 | `test_rejection_case[manifest-*]` | `skillspec` / `skillcheck` raise before activation | `z9j4c9` |
| `filesystem` | 14 | `test_rejection_case[filesystem-*]` | build-root/source-dir static validation | `z9j4c9`, `3c0ss2` |
| `module` | 2 | `test_rejection_case[module-*]` | nearest-`go.mod` resolution | `z9j4c9` |
| `dependency-graph` | 14 | `test_rejection_case[dependency-graph-*]` | `go list` JSON stream validation in the parent | `2g21eg` |
| `toolchain` | 5 | `test_rejection_case[toolchain-*]` | toolchain identity + family allowlist | `3j8pp5` |
| `compiler-directive` | 3 | `test_rejection_case[compiler-directive-*]` | directive scan | `2g21eg` |
| `process` | 11 | `test_rejection_case[process-*]` | env/telemetry/graph enforcement | `2g21eg` |
| `cache` | 16 | `test_rejection_case[cache-*]` | protected cache verification | `2jfnz6`, `8nxlgx` |
| `context` | 2 | `test_rejection_case[context-*]` | context/build-source boundary | `3c0ss2` |
| `execution-policy` | 2 | `test_rejection_case[execution-policy-*]` | input validation | `2dnqw2` |

Each case asserts the tuple `(error, result, artifact_executed, reuse)` from `case["expected"]`, and the error string is compared to a **stable csk diagnostic code**, not to a message. Every csk error type raised on these paths must therefore carry a `code` attribute drawn from the vector's vocabulary — **58 distinct codes across the 77 cases**. This is a cross-task contract: `2g21eg`, `2dnqw2`, `3c0ss2`, `3j8pp5`, `2jfnz6`, `8nxlgx` and `z9j4c9` must agree on it. Recommended shape: `class BuildError(Exception): code: str`, one hierarchy under `src/csk/builds/errors.py`, established by the first task to land (`z9j4c9`) and extended by each subsequent one.

Two cases carry named regression semantics rather than a rejection and must not be lumped in:
`build_source_cases[legacy-nul-stream-structural-collision]` and `[root-marker-bytes-are-build-input]` have `result: "regression-proved"` with boolean equality expectations (`legacy_streams_equal`, `framed_hashes_equal`, `build_source_hashes_equal`, `legacy_installed_tree_hashes_equal`). They are asserted as **equality relations between two computed digests**, proving the framing fixes a real collision the legacy NUL stream had — this is exactly the `3c0ss2` AC "root `.csk-install.json` changes alter build-source identity while legacy installed-tree `content_sha256` remains unchanged".

### 5.2 `vectors/go-host-execution-policy.json` (`c3d42f76…`)

| Block | Count | Test :: assertion | Task |
| --- | ---: | --- | --- |
| `mandatory_controls` | 18 | `test_mandatory_controls_are_all_enforced` — each name maps to an enforced control; all `portable: true`, all `hardened_guarantee: false` | `2g21eg` |
| `native_control_inventory` | 5 controls × 2 platforms | `test_native_control_inventory` — `version == rc5-native-control-inventory-v1`, `exhaustive` true, `platforms == ["macos","windows"]`, `probe_scope == per-operation`, `probe_timing == pre-worker-launch`, per-platform availability + mechanism/`unavailable_reason` | `2g21eg` |
| `identity_and_protocol_cases` | 14 | `test_identity_and_protocol_case` — 7 → `build_execution_worker_identity_invalid`, 6 → `build_execution_worker_protocol_invalid`, 1 → `build_execution_control_unavailable`; all assert `published` false and the declared `worker_started`/`compiler_started` | `2g21eg` |
| `package_influence_cases` | 8 | `test_package_influence_case` — all → `build_execution_package_influence_forbidden`, `worker_started` false, `compiler_started` false | `2g21eg` |
| `capability_evidence_cases` | 11 | `test_capability_evidence_case` — 3 accept, 6 → `build_execution_capability_evidence_invalid`, 2 → `build_execution_hardened_claim_forbidden` | `2g21eg` |
| `capability_evidence_record.consistency_rules` | 8 | `test_capability_evidence_consistency_rule` — 6 → invalid, 2 → hardened-claim-forbidden | `2g21eg` |
| `capability_evidence_record` shape | — | `record_fields == [controls, execution_policy, platform, record_version]`; `control_entry_fields == [availability, name, probed_at, status]`; `exposed_in == [dry-run-plan-result, install-result, status-result]`; **`excluded_from == [cache-key, conformance-claim, install-marker, receipt]`** | `2g21eg`, `2dnqw2`, `th0jdi` |
| `deferred_hardened_guarantees` ↔ `deferred_capability_rejection_guards` | 6 ↔ 6 | `test_deferred_guarantee_is_refused` — names pair 1:1; each guard has `rejects_portable_build` false, `portable_profile_claims` false, `in_*` all false | `2g21eg` |
| `failure_boundary` | 3 | `test_failure_boundary` — only `missing_mandatory_portable_control` rejects (`build_execution_control_unavailable`, `fails_before: worker-launch`, `published` false); the other two neither reject nor block publication | `2g21eg` |
| `session_states` | 13 | `test_session_state_order` — the driver traverses exactly this ordered list | `2g21eg` |
| `process_graph` | 4 | `test_process_graph` — exactly `manager-parent`, `identity-verified-manager-owned-worker`, `fingerprinted-goroot-bin-go`, `fingerprinted-goroot-pkg-tool-child` | `2g21eg` |
| `policy_semantics` | 6 | `test_policy_semantics_are_not_overclaimed` — each key's `means` is implemented and its `does_not_mean` is **not** claimed anywhere in csk's evidence or docs output | `2g21eg`, `akf5kh` |
| `cache_identity` | 3 | duplicate of §5.1 — asserted **once**, in `test_conformance_identity.py`, and cross-checked for equality between the two vector files | `2dnqw2` |
| `drivers` | 2 | `test_drivers_vocabulary` — `["go-repository-v1","go-v1"]`; only `go-v1` is in this story's scope | — |

`cache_identity` appears identically in both `build-drivers.json` and `go-host-execution-policy.json`. The harness asserts the two blocks are equal and then evaluates the identity assertions once, so a future divergence between the two vector files fails loudly instead of being tested twice against one of them.

### 5.3 `schema-cases/` — 102 in-scope cases

| Directory | Cases (valid/invalid) | Test | Product oracle | Task |
| --- | ---: | --- | --- | --- |
| `agent-skill-v6/` | 24 (1/23) | `test_conformance_schema_cases.py::test_manifest_case[agent-skill-v6]` | `skillspec.load_skill_spec` on a temp tree named `agent-skill.json` | `z9j4c9` |
| `csk-skill-v6/` | 24 (1/23) | same, named `csk-skill.json` | same | `z9j4c9` |
| `build-receipt-v1/` | 18 (1/17) | `test_receipt_case` | `csk.builds.metadata` receipt reader | `2dnqw2` |
| `install-marker-v2/` | 14 (3/11) | `test_marker_case` | marker model | `2dnqw2` |
| `conformance-claim-v1/` | 2 | `test_claim_case[v1]` | claim reader | `12r55p` |
| `conformance-claim-v2/` | 7 | `test_claim_case[v2]` | claim reader | `12r55p` |
| `conformance-claim-v3/` | 13 (2/11) | `test_claim_case[v3]` | claim reader | `12r55p` |

The two v6 directories are **byte-identical case sets under two filenames** — that is the whole point of the pair, and it is the direct regression test for §2.3. `test_manifest_case` writes the instance into `tmp_path` under the directory's manifest name and asserts `spec.source_file` is that name, so a silent empty-spec fallback can never make a case pass again.

Explicitly **out of scope** here, per the `12r55p` scope line, and belonging to the `go-repository-v1` line: `agent-skill-v7`/`csk-skill-v7` (33 each), `build-receipt-v2` (26), `install-marker-v3` (27), `skill-build-v1` (13, driver `go-repository-v1`), `skillfile-dev-v2` (16), and the registry/external-repository directories. `schema-cases/index.json` is the enumeration source, filtered by directory; the filter list is declared once in `clusters.py` with the owning line named in a comment, so an out-of-scope case is skipped **by name**, never by accident.

### 5.4 `vectors/manager-lifecycle.json` (`2ddbd266…`)

| Block | Count | Test | Product boundary | Task |
| --- | ---: | --- | --- | --- |
| `bootstrap_cases` | 3 | `test_bootstrap_case` — `missing + if_missing → created`, `existing-invalid + if_missing → unchanged-success`, `if_missing + force → usage-error` | `csk.cli` / `config` | `2x6mjn` |
| `dry_run_cases` | 2 | `test_dry_run_purity` — after a dry run, all 9 `forbidden_persistent_effects` are absent for project and global scope | planner | `2x6mjn` |
| `launcher_cases` | 2 | `test_launcher_case` — arguments forwarded, exit status preserved, inherited PATH preserved, the 3 `required_path_roles` present, on both `unix` and `windows` | `csk.shims` | `11yhth` |
| `upgrade_cases` | 3 | `test_upgrade_case` — selected project closure with direct+transitive fetch and exclusion, all-projects dedup, global closure | `closure`, `installer`, `global_install` | `3t8nr3`, `g7kgox` |

`dry_run_cases` is the strongest lifecycle assertion available and is worth wiring as a real filesystem-diff test: snapshot the whole home + project trees before, run the dry run, snapshot after, assert byte equality. That covers all 9 forbidden effects with one mechanism rather than 9 bespoke probes.

### 5.5 `vectors/conformance-claim-v3-qualification.json` (`c4b9132c…`)

| Assertion | Value |
| --- | --- |
| `claim_schema_version` | 3 |
| `protocol_version` | `1.0.0-rc.5` |
| `candidate_claims_emitted` | `[]` — **and csk emits none** |
| `rules[schema-valid-is-not-qualified]` | requires `native-driver-platform-evidence` |
| `rules[driver-platform-subset]` | each driver platform is also top-level evidenced |
| `rules[no-generic-driver]` | `allowed_drivers == ["go-repository-v1","go-v1"]` |
| `rules[no-unevidenced-platform]` | every emitted tuple has immutable passing evidence |
| `platforms[linux]` | `status: excluded`, `until_task: TASK-260728-1skseh` |
| `platforms[macos]`, `platforms[windows]` | `status: pending-downstream-native-evidence` |

`test_conformance_claim.py::test_no_claim_is_emitted` asserts csk produces no claim artifact anywhere under the test home — the guard against a fabricated release claim, which both `12r55p` and the parity map call out.

### 5.6 `expected/` goldens

| Golden | Test | Compared against |
| --- | --- | --- |
| `expected/build-driver/build-input.ccj.json` | `test_portable_identity_input_bytes` | `csk.builds.metadata.canonical_input_bytes(...)` |
| `expected/build-driver/cache-key.txt` | `test_portable_identity_cache_key` | `sha256` of the above, and `portable_identity.cache_key` |
| `expected/build-driver/receipt.ccj.json` | `test_portable_identity_receipt_bytes` | receipt writer output; asserts no BOM, no whitespace, no trailing `\n` |
| `expected/build-driver/receipt-sha256.txt` | `test_portable_identity_receipt_hash` | `sha256` of the above |
| `expected/build-driver/build-source.preimage.bin` | `test_build_source_preimage_bytes` | `csk.builds.source.preimage(fixtures/go-build-skill)` |
| `expected/build-driver/build-source-sha256.txt` | `test_build_source_digest` | `sha256` of the above |
| `expected/build-driver/toolchain.preimage.bin` | `test_toolchain_preimage_bytes` | toolchain framing over `toolchain_cases[0].entries` |
| `expected/build-driver/toolchain-sha256.txt` | `test_toolchain_digest` | `sha256` of the above |
| `expected/build-driver/marker.json` | `test_build_marker_golden` | marker v2 payload; also equals `portable_identity.marker` |
| `expected/build-driver/context_files.json` | `test_go_fixture_context_files` | `whitelist.copy_context(fixtures/go-build-skill, …)` |
| `expected/build-driver/context_sha256.txt` | `test_go_fixture_context_hash` | `hashing.content_sha256(destination)` |
| `expected/context_files.json`, `context_sha256.txt`, `marker.json` | existing `test_shared_fixture_context_hash_and_marker` | `fixtures/skill` — unchanged, but see §2.3 |
| `expected/snapshot_sha256.txt` | `test_snapshot_digest` (**new, wave 1**) | `hashing.content_sha256(fixtures/skill)` |

`expected/snapshot_sha256.txt` is `sha256:f2c21f31f69bb5c9366028ac6e63a701c66803586c8ce12eac1186dcd350e5eb`. It is the whole-snapshot content digest of `fixtures/skill` — verified during this design to be reproduced exactly by today's unmodified `csk.hashing.content_sha256`, with and without the default `.csk-install.json` exclusion (the fixture has no such file, so both agree). It is the same value as `vectors/registry-behavior.json#artifact_hash`, which binds the snapshot identity to the registry artifact key. No existing csk test consumes it; it belongs in wave 1 and needs no product change.

`expected/adapter-ledger.json`, `expected/registry/**`, `expected/external-repository/**` are the registry and `go-repository-v1` lines — out of scope, and named in the `clusters.py` skip list so they are visibly excluded rather than forgotten.

---

## 6. Fixture loading strategy

### 6.1 The two fixtures

| Fixture | Files | Manifest | Purpose |
| --- | ---: | --- | --- |
| `fixtures/skill` | 6 | `agent-skill.json`, schema 5, `runtime_roots: ["scripts"]`, one script command | Legacy shared context/marker golden. Present in rc.2 too (as `csk-skill.json`). |
| `fixtures/go-build-skill` | 13 | `agent-skill.json`, schema 6, `runtime_roots: ["scripts"]`, `build_roots: ["assets/build-tool"]`, one `type: build` + one `type: script` command | The rc.5 build-driver fixture: real vendored Go module, `.csk-install.json` root marker, two eligible context files. |

`fixtures/go-build-skill` snapshot is exactly the 13 files listed in `build-drivers.json#fixture.snapshot_files`; 2 are eligible context (`SKILL.md`, `assets/prompt.md`) and 9 are excluded (7 under the build root, 2 script runtime files). `.csk-install.json` and `agent-skill.json` are in neither list — the marker is a build-source input but not context, and the manifest is not context.

### 6.2 Loading rules

1. **Never mutate the root.** Every test that needs a writable tree copies the fixture into `tmp_path` with `shutil.copytree(..., symlinks=True)`. The root is opened read-only.
2. **Copy, then assert the copy is faithful.** After copying, assert `hashing.content_sha256` (or the build-source digest) over the copy matches the vector's declared value, before running the behaviour under test. A copy that silently dropped a mode bit or normalised a newline would otherwise produce a confusing downstream failure.
3. **Do not `git init` the fixture** for pure-hash tests. `build-drivers.json#rejection_cases[process/vcs-metadata]` expects `ambient_vcs_input_forbidden`; the fixture must stay VCS-free unless the case demands otherwise.
4. **Windows and symlinks.** `fixtures/go-build-skill` has no symlinks, but `toolchain_cases` do (`pkg/tool-link → ../bin/go`, escaping/absolute/dangling variants). `csk` CI sets `core.symlinks=false` on Windows. Toolchain link cases build their trees programmatically from the vector's `entries` and are gated on `os.symlink` succeeding; on a Windows runner without the privilege they must **fail with an explicit unsupported-runner message**, not silently skip, in the rc.5 qualification job (§7.3).
5. **Case-insensitive filesystems.** `hashing._reject_platform_collisions` already handles macOS/Windows normalisation. Manifest cases that differ only by case must be materialised one at a time in separate `tmp_path` directories.
6. **Schema-case materialisation.** `test_manifest_case` writes exactly one instance file into an otherwise empty `tmp_path` directory under the correct manifest name, then calls `skillspec.load_skill_spec(dir)`. It asserts `spec.source_file == "<manifest name>"` for valid cases and a `SkillSpecError` for invalid ones. No other file is created, so path-validation cases are not perturbed.

---

## 7. Platform gating and skip policy

### 7.1 The three tiers

| Tier | What it covers | Runs on | Gate |
| --- | --- | --- | --- |
| **A — pure data** | vector shape, closed vocabularies, cache identities, CCJ bytes, all 11 `expected/build-driver` goldens, schema cases, framing preimages, claim rules, manifest integrity | **every** platform incl. Linux, every supported Python | cluster gating only (§4.3) |
| **B — host behaviour** | native-control probe, worker launch/identity, session-state traversal, process-graph enforcement, protected-cache filesystem primitives | macOS + Windows | `sys.platform in {"darwin","win32"}` |
| **C — real Go build** | end-to-end `go list`/`go build` through the worker with the vendored fixture, shim launch, rebuild/rollback/recovery | macOS + Windows **with a resolvable trusted GOROOT** | tier B **and** toolchain resolution |

Tier A is the reason Linux CI stays valuable under rc.5. `protocol/core.md` §4.2.1 requires `manager-worker-v1` on macOS and Windows only, and `rc5-native-control-inventory-v1` is `exhaustive: true` over exactly `["macos","windows"]` — but none of that makes the byte identities platform-dependent. Roughly 85% of the ~320 new cases are tier A.

### 7.2 Linux is asserted, not skipped

The single most important platform test is **positive**:

```
test_conformance_execution_policy.py::test_non_supported_host_fails_closed
```

On any host outside macOS/Windows, the source-aware path must raise `build_execution_control_unavailable`, before worker launch, publishing nothing. This is `failure_boundary.missing_mandatory_portable_control` and mirrors accepted Curator `internal/godriver/controls_other.go`. Linux CI therefore proves fail-closed behaviour rather than skipping the driver, which is precisely the retarget `TASK-260720-3pemm6` needs (its current AC still says "CI runs the real fixture on ubuntu, macOS, and Windows" — see §10.1).

### 7.3 Skip policy

| Situation | Behaviour |
| --- | --- |
| `CURATOR_CONFORMANCE_ROOT` unset | skip whole suite (**unchanged** from today) |
| root protocol < cluster introduction | skip, reason names cluster + both versions |
| root protocol ≥ introduction but cluster absent | **fail** |
| tier B on Linux | skip the host-behaviour case, **and** run `test_non_supported_host_fails_closed` |
| tier C, no trusted GOROOT, developer run | skip with the resolution attempt in the reason |
| tier C, no trusted GOROOT, rc.5 qualification job | **fail** — the job declares `CSK_CONFORMANCE_REQUIRE_GO=1` |
| tier C, Windows without symlink privilege | fail in the qualification job, skip locally, same mechanism |

`3pemm6`'s "no unexpected skip or xfail" AC is met by making every skip reason machine-checkable: the qualification job runs `pytest -rs` and a small assertion over the short summary that every skip line matches an allow-list of expected reasons. `xfail` is not used anywhere in the conformance suite.

Pytest markers (`build_driver`, `execution_policy`, `host_behaviour`, `real_go`) must be registered under `[tool.pytest.ini_options] markers = [...]` in `pyproject.toml`; there is no marker configuration today, so unregistered markers would warn.

---

## 8. First-wave literal implementation gates

The first wave is everything that is **runnable today against unmodified `csk`**. It has been executed during this design and passes. It gives `TASK-260720-12r55p` a landable first slice that does not wait on the 17-task chain, and it locks the identity contract before any implementation exists to drift from it.

### 8.1 Scope of wave 1

`tests/conformance/{__init__,root,clusters,goldens}.py`, `tests/test_conformance_root.py`, and the identity half of `tests/test_conformance_identity.py`:

- root resolution, manifest integrity over all 447 entries, provenance header
- cluster registry + version gating, including the fail-on-missing rule
- `cache_identity`: `aliases is False`; three keys distinct; each re-derived from its own `input` via the shared CCJ-1 primitive; `schema_valid` flags `True/False/False`
- `portable_identity`: `canonical_bytes(build_input)` equals the 869 golden bytes and hashes to `cache-key.txt`; `canonical_bytes(stored_receipt)` equals the 1120 golden bytes and hashes to `receipt-sha256.txt`; `marker.json` equals `portable_identity.marker`
- the two `cache_identity` blocks in `build-drivers.json` and `go-host-execution-policy.json` are equal
- the closed vocabularies from `go-host-execution-policy.json` (18 controls, 5-control × 2-platform inventory, 6↔6 deferred pairs, 3 failure boundaries, 13 session states, 4 process-graph nodes, 8 consistency rules) as data-shape assertions
- `conformance-claim-v3-qualification.json` rules and platform statuses
- `expected/snapshot_sha256.txt` reproduced by `hashing.content_sha256(fixtures/skill)`

Wave 1 needs **no** new product code. It uses `csk.audit_registry.canonical_bytes` and `csk.protocol_json.loads` as they exist.

### 8.2 Literal commands

Baseline, against the released default pin — must stay green (measured: **98 passed**):

```bash
cd /Users/iv/Developer/intranet/cocoaskills
TMPDIR="${TMPDIR:-/tmp}" \
CURATOR_CONFORMANCE_ROOT="$PWD/protocol-spec/conformance/v1" \
  python -m pytest tests/test_protocol_conformance.py -q
```

Wave 1, against the immutable rc.5 candidate:

```bash
cd /Users/iv/Developer/intranet/cocoaskills
export CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1
export CURATOR_CONFORMANCE_MANIFEST_SHA256=sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
export TMPDIR=/tmp/csk-conformance

mkdir -p "$TMPDIR"
python -m pytest tests/test_conformance_root.py -v
python -m pytest tests/test_conformance_identity.py -v
python -m pytest tests/test_conformance_execution_policy.py -v -k "vocabulary or inventory or deferred or failure_boundary or session_state or process_graph or consistency"
python -m pytest tests/test_conformance_claim.py -v
```

Regression guard — the legacy suite must stay green against **both** roots:

```bash
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-1b9tc3/pin-cbe912d0/conformance/v1 \
  python -m pytest tests/test_protocol_conformance.py -q     # expect 98 passed
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
  python -m pytest tests/test_protocol_conformance.py -q     # currently 1 failed, 97 passed — see §2.3
```

Skip audit and typing gate:

```bash
python -m pytest tests/ -q -rs        # every skip reason must match the allow-list
python -m mypy                        # strict, files = ["src/csk"]
```

> `TMPDIR` must be outside any Git worktree. The accepted `TASK-260729-1t1z2l` review recorded one full-run failure caused solely by a `TMPDIR` inside a Git checkout; the focused rerun outside one was green. Set it explicitly in CI.

### 8.3 Wave ordering

| Wave | Content | Unblocked by |
| --- | --- | --- |
| 1 | §8.1 — harness, provenance, identity, closed vocabularies | nothing (runs today) |
| 2 | schema cases for `agent-skill-v6` + `csk-skill-v6`, manifest-name parity regression | `z9j4c9` |
| 3 | build-source framing, fixture context boundary, `expected/build-driver` context goldens | `3c0ss2` |
| 4 | toolchain cases, argv forms, fixed environment | `3j8pp5` |
| 5 | receipt/marker schema cases + byte goldens | `2dnqw2` |
| 6 | execution-policy behavioural cases, 14 identity/protocol, 8 package-influence, 11 capability-evidence, tiers B/C | `2g21eg` |
| 7 | cache rejection cluster (16) | `2jfnz6`, `8nxlgx` |
| 8 | manager-lifecycle, dry-run purity, launcher | `2x6mjn`, `3t8nr3`, `g7kgox`, `11yhth` |
| 9 | claim-version separation, qualification rules, no-claim-emitted | `th0jdi` |

Waves 2–9 are ordered by the existing dependency DAG; no new board dependency is required.

---

## 9. CI integration

### 9.1 What must not change

The `test` job keeps checking out `relux-works/curator-spec@cbe912d064e06275b0a1aa6762b7c31f687051c5` and exporting `CURATOR_CONFORMANCE_ROOT` to it. That pin advances only via `TASK-260720-25d05o` → `TASK-260720-1utsx8`. Nothing in this design touches it.

### 9.2 What is added

A separate `conformance-candidate` job, non-default and explicitly parameterised:

- matrix `os: [macos-latest, windows-latest, ubuntu-latest]`, one Python version
- checks out the candidate suite at a caller-supplied revision **or** consumes a pre-materialised `CURATOR_CONFORMANCE_ROOT`
- sets `CURATOR_CONFORMANCE_MANIFEST_SHA256` so a wrong root fails loudly
- sets `CSK_CONFORMANCE_REQUIRE_GO=1` on macOS/Windows only
- ubuntu runs tier A plus `test_non_supported_host_fails_closed`, and is **not** expected to produce a green `go-v1` build
- Go setup on macOS/Windows uses the accepted release family (Curator's accepted allowlist is Go family 1.25; 1.23 is only the protocol floor)

### 9.3 Known runner gap

`TASK-260729-1bf72u` recorded that `ssh win` is reachable but has no Go on PATH. Wave 6+ tier C on Windows is blocked on that task's operator-safe Go setup recommendation. Tiers A and B are not.

---

## 10. Routed recommendations (not applied here)

This task is read-only design; brief edits belong to `TASK-260729-v5hqnv` and platform work to `TASK-260729-1bf72u`. One item needs routing, and one converged during this run.

### 10.1 `TASK-260720-3pemm6` — converged, no action

At the start of this analysis `3pemm6` still read "CI runs the real fixture on **ubuntu**, macOS, and Windows". `TASK-260729-v5hqnv` handed off to review during this run and retargeted it. Its current AC now reads:

> "Ubuntu jobs run portable non-driver coverage across the same Python matrix, prove the source-aware go-v1 path fails closed with an unavailable-control error and no worker launch, and neither require nor claim a green go-v1 build; Linux driver success is not asserted anywhere."

That is the same boundary §7 of this design derives independently: tier A everywhere, tiers B/C on macOS/Windows, and Linux asserted fail-closed rather than skipped. **No further retarget is needed; §7 is the implementation of the retargeted AC.**

### 10.2 `TASK-260720-z9j4c9` AC does not name manifest-name parity — needs routing

`z9j4c9` is **not** one of the seven briefs `v5hqnv` owns, so nothing in flight covers this. Its description says "canonical and legacy manifest parity", but its testable AC list never mentions `agent-skill.json`.

§2.3 shows this is not cosmetic. It is the single cause of the current rc.5 red, and the failure mode is silent: `load_skill_spec` returns `SkillSpec(commands={}, source_file=None)` for a directory whose manifest it does not recognise, which turns fixture-driven build-root and context assertions vacuous instead of red.

Exact recommended AC clause, to append to `TASK-260720-z9j4c9`:

> `load_skill_spec` reads `agent-skill.json` and `csk-skill.json` with identical semantics, sets `source_file` to the filename it actually read, and never returns an empty spec for a directory containing either name; the shared `fixtures/skill` and `fixtures/go-build-skill` conformance fixtures load with their declared `runtime_roots`, commands and `build_roots`.

Not applied here: this task's scope is read-only design, and `z9j4c9` is `backlog` and outside it. **Route as a one-field board edit.** Until it lands, `test_protocol_conformance.py::test_shared_fixture_context_hash_and_marker` stays red against any rc.5 root.

### 10.3 No new tasks are proposed

Verification performed before concluding this, against the spec surfaces available to this task:

| Candidate gap considered | Checked against | Result |
| --- | --- | --- |
| rc.5 build-driver goldens missing (parity-map §3.3) | root manifest at `b6f56aac…`; `TASK-260729-3nx97g` status `done`, reviewer ACCEPTED | **Closed.** `vectors/build-drivers.json` + 11-file `expected/build-driver/` present and manifest-covered. No task needed. |
| Schemas outside the manifest ⇒ needs a schema-provenance task | `12r55p` AC; `schema-cases/conformance-claim-v*/` contents | **Rejected.** Every schema-derived AC assertion is satisfiable from manifest-covered `schema-cases/` (§4.4). Adding a dependency and a task would be invented scope. |
| `skill-build-v1` (13 cases) unowned | `skill-build-v1.schema.json` uses driver `go-repository-v1`; `12r55p` scope explicitly excludes the external-repository line | **Rejected.** Belongs to the `go-repository-v1` line, already out of scope by name. |
| Dual manifest-name support unowned | `z9j4c9` description "canonical and legacy manifest parity"; `z9j4c9` scope owns `src/csk/skillspec.py` | **Rejected as a new task.** The implementation work is already inside an existing task; only its AC wording needs sharpening (§10.2). A separate task would duplicate `z9j4c9`'s scope. |
| Windows Go absence needs a task | `TASK-260729-1bf72u` (status `development`) AC: "operator-safe Go setup recommendation for Windows if absent" | **Rejected.** Already owned. |
| Shared CCJ-1 primitive duplicated between `audit_registry` and `protocol_json` | `2dnqw2` scope: "narrowly shared CCJ-1 support in `src/csk/protocol_json.py`" | **Rejected.** Owned; recorded as a design constraint in §3.3 instead. |

No element in this design is created beyond the literal spec, so no `Justified gap` record is required.

---

## 11. Verification record

Everything below was executed during this design, read-only, on macOS (darwin 25.5.0).

| # | Check | Result |
| --- | --- | --- |
| 1 | rc.5 root manifest digest | `b6f56aac…`, matches the accepted `TASK-260729-3nx97g` pin |
| 2 | rc.5 aggregate tree digest, file count | `e6a13215…`, 448 files (447 manifest + `manifest.json`) |
| 3 | 18 key vector/expected digests recorded | §1.2, §1.3 |
| 4 | Pinned CI suite resolved and exported at `cbe912d0` | protocol `1.0.0-rc.2`, manifest `728f7729…`, 81 files |
| 5 | `pytest tests/test_protocol_conformance.py -q` vs rc.2 pin | **98 passed** |
| 6 | same vs rc.5 candidate | **1 failed, 97 passed** |
| 7 | Failure root-caused to the `csk-skill.json` → `agent-skill.json` rename | fixture bodies semantically identical; `load_skill_spec` returns `source_file=None` for both rc.5 fixtures |
| 8 | `canonical_bytes(build_input)` == 869 golden bytes | **True** |
| 9 | `canonical_bytes(stored_receipt)` == 1120 golden bytes | **True** |
| 10 | `sha256` of both goldens == `cache-key.txt` / `receipt-sha256.txt` | **True** |
| 11 | `expected/build-driver/marker.json` == `portable_identity.marker` | **True** |
| 12 | `cache_identity.aliases` false; 3 keys distinct; all 3 re-derived | **True** |
| 13 | `protocol_json.loads` accepts the golden receipt bytes | **True** |
| 14 | Framed preimages decoded in pure Python (`D`/`F`/`L`/`V` + uint64 BE lengths) | 13 build-source records, 5 toolchain records — all parsed |
| 15 | Schemas outside the conformance manifest | `<root>/../../schemas/v1`; manifest covers only `schema-cases`, `expected`, `fixtures`, `vectors` |
| 16 | Case counts per in-scope schema directory | §5.3, from `schema-cases/index.json` (376 cases total, 102 in scope) |
| 17 | Rejection-case boundary and error-code census | 77 cases, 10 boundaries, 58 distinct error codes |
| 18 | Execution-policy closed vocabularies enumerated | 18 controls, 5×2 inventory, 14 + 8 + 11 cases, 8 rules, 6↔6 deferred, 3 boundaries, 13 states, 4 graph nodes |
| 19 | `expected/snapshot_sha256.txt` reproduced by unmodified `hashing.content_sha256(fixtures/skill)` | **True** (`f2c21f31…`, identical with and without the default exclusion) |

No CocoaSkills or Curator file was created, edited, staged, or committed. No pin, release, or publication was touched. The only artifacts written are this document and the read-only export of the pinned rc.2 suite under `.temp/TASK-260729-1b9tc3/pin-cbe912d0/`.

---

## 12. References

- Board: `TASK-260729-1b9tc3`, parent `STORY-260720-1uv5gi`, epic `EPIC-260720-21aq1i`
- Accepted parity map: `TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md` (revision 2) + `TASK-260729-1t1z2l_review-verdict-cycle-2.md`
- Accepted rc.5 golden regeneration: `TASK-260729-3nx97g` (`done`, reviewer ACCEPTED)
- Siblings in flight: `TASK-260729-v5hqnv` (brief retargets), `TASK-260729-1bf72u` (runner readiness), `TASK-260729-35tb37` (baseline file map)
- Immutable root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Pinned released suite export: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-1b9tc3/pin-cbe912d0/conformance/v1`
- CocoaSkills: `/Users/iv/Developer/intranet/cocoaskills` at local `main` `edce881`
