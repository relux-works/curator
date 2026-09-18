# TASK-260918-1hl5f8 — landing review

Verdict: ACCEPTED. The exact 1ca4b3d delivery tree may land. No corrections requested.

Exact tree: 1ca4b3df32baceff99f6be4b136f39addfa4da00, parent 4a2fa3ec428b5021e9e8a80f29fa7fc5a88e4990. Delivery tree was read-only; all generation and verification used a disposable byte copy. No implementation edits.

## Provenance and fidelity

Both attachment SHA-256 digests match the assignment. The attached union patch is byte-identical to git diff 4a2fa3e 1ca4b3d. The accepted patch applies cleanly to e8b53a0. The brief’s isolated 684c9f1 reference is stale: git apply --check fails there (CHANGELOG.md, protocol/environments.md, release/1.0.0-rc.9.json, tools/validate.py). This is a brief typo, not a tree defect.

All 8/8 changed paths equal the candidate path set. Six files have identical added/deleted-line streams; only protocol/environments.md and regenerated release pins differ. tools/validate.py retains both registrations. Independent git merge-file of landed/base/accepted yielded exactly 2 prose conflicts and 1 registration conflict; all 5/5 non-conflict regions are byte-identical and ordered in the delivery file.

## Per-hunk verdict

| Hunk | Evidence and interpretation | Verdict |
|---|---|---|
| protocol/environments.md:2954–2968 (§12) | E6 “entries, and — for a `path` root or overlay — the source directory” and “naming the failing check and the path” precede the full E3 codex-seed revision/native-server rows; the list ends “Both commands follow”. No new condition or lost row. | PASS |
| protocol/environments.md:3214–3273 (§13) | E5 write/nofollow and S5 protected-boundary cases retained; “non-conforming;” joins “the section 7.4 codex-seed” block and “and the section 2.2 and section 4 path-kind admission cases”. Both complete case lists retained. Only list punctuation and line wrapping changed. | PASS |
| tools/validate.py:8155–8156 (main) | `validate_environments_codex_seed_vectors,` followed by `validate_environments_path_kind_admission_vectors,`; main iterates and calls every registered check. | PASS |


### Full union quote: protocol/environments.md

```text
2954: value, the store-trust row per installed profile — the section 4
2955: boundary and pin-hash verdict for the profile's lock, marker, named store
2956: entries, and — for a `path` root or overlay — the source directory,
2957: naming the failing check and the path or, for an enclosing failure, the
2958: boundary (environments root or store root) when the profile is
2959: `environment_store_untrusted`, the active codex-seed revision (`A` or `B`, section 7.4) with its
2960: behaviour, and per managed `codex_cli` home the `codex_seed_record`
2961: revision with its native-server rows (`mcp_native_servers_ungoverned`
2962: for the native names an `A`-record home carries,
2963: `mcp_native_servers_not_inherited` for a `B`-record home's stripped
2964: names, `mcp_seed_unstripped` with the re-provision hint when the marker
2965: predates the rule or when an `A`-record home with a non-empty snapshot
2966: is served by a revision-`B` manager). Both commands follow
2967: the manager §10 discipline exactly: recompute and report, never mutate — no
2968: fetch, no repair, no adoption, no channel application, no onboarding.
```


### Full union quote: protocol/environments.md

```text
3214: write-discipline vectors
3215: (`vectors/environments-write-nofollow.json`) — the symlinked-target
3216: takeover (replaced with backup under authorization, stopped with
3217: `environment_foreign_manager_detected` without), the symlinked-parent
3218: refusal with `environment_write_would_follow_link` under both
3219: authorization states, the post-provisioning planted-link repair, the
3220: manager-owned-link replace, the symlinked-backup-destination refusals (a traversed parent link and a directly symlinked target),
3221: the inside-pointing-link ledger refusal, and the clean-path and
3222: recorded-file positives — every foreign-link case asserting the link's
3223: former target is byte-identical afterwards;
3224: the section 4
3225: protected-boundary cases
3226: (`vectors/environments-store-boundary.json`) — the intact resolve that
3227: emits a fragment; the swapped system-prompt bytes, swapped root-context
3228: bytes, symlinked entry root, wrong ownership, wrong permissions,
3229: containment escape, non-regular component, and pin-hash mismatch cases
3230: that refuse with `environment_store_untrusted` and emit no fragment; the
3231: environments-root and store-root enclosing-boundary cases that refuse with
3232: no rebuild; the intact-updated-store with old marker case that reports
3233: `environment_home_stale` and repairs, and the swapped-updated-store with
3234: old marker case that refuses with `environment_store_untrusted` and is
3235: never adopted; the unprovisioned-home intact and swapped cases; the
3236: unreadable-marker case that reports
3237: `environment_marker_unreadable`; the dry-run `would-rebuild-untrusted-store`
3238: entry-class case and the enclosing no-rebuild case; the repair rebuild,
3239: entry-rebuild, enclosing-refusal, stale-repair, and unprovisioned cases;
3240: the `env status` non-current posture rows naming the failing check and the
3241: boundary; and the negative cases whose fragment-emitting,
3242: current-reporting, or re-applying observation is non-conforming;
3243: the section 7.4 codex-seed
3244: cases (`vectors/environments-codex-seed.json`) — a native `config.toml`
3245: with `mcp_servers` tables under revision A (copied whole,
3246: `mcp_native_servers_ungoverned` names the entries, posture lists them
3247: ungoverned) and under revision B (the seeded file lacks them,
3248: `mcp_native_servers_not_inherited` names them), a native file without
3249: `mcp_servers` entries (no warning under either revision, the record still
3250: written), the pre-rule-home `mcp_seed_unstripped` posture, and the
3251: revision-A home under a revision-B manager (`mcp_seed_unstripped` with
3252: the re-provision hint, the recorded names still ungoverned) — with the
3253: `codex_seed_record` marker shape;
3254: and the section 2.2 and section 4 path-kind admission cases
3255: (`vectors/environments-path-kind-admission.json`) — the `git` MCP
3256: declaration admitted, the `path` root, overlay, and onboarding-import MCP
3257: declarations refused with `mcp_declaration_path_source_refused` naming the
3258: package and the declaration, the directly named `path` overlay with a
3259: system module admitted when its directory passes the contract, the `path`
3260: root, the `path` overlay, and the onboarding import without system
3261: modules admitted, the world-writable, symlinked-component,
3262: wrong-ownership, containment-escape, and non-regular-component
3263: directories refused with `environment_store_untrusted`, no fragment, and
3264: a posture row naming the path and the failing check, the world-writable
3265: no-system overlay and the untrusted no-system onboarding import refused
3266: the same way regardless of content class, the transitive `path` system
3267: module refused with `context_system_module_transitive`, the dry-run
3268: evaluation of a `path` directory failure reporting
3269: `environment_store_untrusted` with no rebuild planned and mutating
3270: nothing, and the negatives whose admitted, current-reporting,
3271: rebuilding, or would-rebuild-reporting observation is non-conforming. The nine
3272: retired `expected/environments/*` sets are regenerated under the v2 type
3273: line. A manager claiming this capability MUST pass the complete vector set;
```


### Full union quote: tools/validate.py

```text
8151:         validate_environment_vectors,
8152:         validate_environments_env_passthrough_vectors,
8153:         validate_environments_source_signers_vectors,
8154:         validate_environments_store_boundary_vectors,
8155:         validate_environments_codex_seed_vectors,
8156:         validate_environments_path_kind_admission_vectors,
8157:         validate_context_version_vectors,
```


### Both source sides: protocol/environments.md

Left = landed 4a2fa3e; right = accepted rev2 applied to e8b53a0.

```text
<<<<<<< /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/landed.tmp
value, the store-trust row per installed profile — the section 4
boundary and pin-hash verdict for the profile's lock, marker, and
named store entries, naming the failing check and, for an enclosing
failure, the boundary (environments root or store root) when the profile
is `environment_store_untrusted`, the active codex-seed revision (`A` or `B`, section 7.4) with its
behaviour, and per managed `codex_cli` home the `codex_seed_record`
revision with its native-server rows (`mcp_native_servers_ungoverned`
for the native names an `A`-record home carries,
`mcp_native_servers_not_inherited` for a `B`-record home's stripped
names, `mcp_seed_unstripped` with the re-provision hint when the marker
predates the rule or when an `A`-record home with a non-empty snapshot
is served by a revision-`B` manager). Both commands follow
=======
value, and the store-trust row per installed profile — the section 4
boundary and pin-hash verdict for the profile's lock, marker, named store
entries, and — for a `path` root or overlay — the source directory,
naming the failing check and the path or, for an enclosing failure, the
boundary (environments root or store root) when the profile is
`environment_store_untrusted`. Both commands follow
>>>>>>> /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/accepted.tmp
<<<<<<< /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/landed.tmp
current-reporting, or re-applying observation is non-conforming.;
and the section 7.4 codex-seed
cases (`vectors/environments-codex-seed.json`) — a native `config.toml`
with `mcp_servers` tables under revision A (copied whole,
`mcp_native_servers_ungoverned` names the entries, posture lists them
ungoverned) and under revision B (the seeded file lacks them,
`mcp_native_servers_not_inherited` names them), a native file without
`mcp_servers` entries (no warning under either revision, the record still
written), the pre-rule-home `mcp_seed_unstripped` posture, and the
revision-A home under a revision-B manager (`mcp_seed_unstripped` with
the re-provision hint, the recorded names still ungoverned) — with the
`codex_seed_record` marker shape. The nine
retired `expected/environments/*` sets are regenerated under the v2 type
line. A manager claiming this capability MUST pass the complete vector set;
=======
current-reporting, or re-applying observation is non-conforming; and the
section 2.2 and section 4 path-kind admission cases
(`vectors/environments-path-kind-admission.json`) — the `git` MCP
declaration admitted, the `path` root, overlay, and onboarding-import MCP
declarations refused with `mcp_declaration_path_source_refused` naming the
package and the declaration, the directly named `path` overlay with a
system module admitted when its directory passes the contract, the `path`
root, the `path` overlay, and the onboarding import without system
modules admitted, the world-writable, symlinked-component,
wrong-ownership, containment-escape, and non-regular-component
directories refused with `environment_store_untrusted`, no fragment, and
a posture row naming the path and the failing check, the world-writable
no-system overlay and the untrusted no-system onboarding import refused
the same way regardless of content class, the transitive `path` system
module refused with `context_system_module_transitive`, the dry-run
evaluation of a `path` directory failure reporting
`environment_store_untrusted` with no rebuild planned and mutating
nothing, and the negatives whose admitted, current-reporting,
rebuilding, or would-rebuild-reporting observation is non-conforming.
The nine retired `expected/environments/*` sets are regenerated under the
v2 type line. A manager claiming this capability MUST pass the complete vector set;
>>>>>>> /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/accepted.tmp
```


### Both source sides: tools/validate.py

Left = landed 4a2fa3e; right = accepted rev2 applied to e8b53a0.

```text
<<<<<<< /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/landed.tmp
        validate_environments_codex_seed_vectors,
=======
        validate_environments_path_kind_admission_vectors,
>>>>>>> /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/accepted.tmp
```


## Scope and architecture

This is a spec/conformance landing, not a manager runtime implementation. E6 extends the existing boundary contract and conformance validator; E3 seed rules remain independent. The tree preserves the accepted contract, including metadata-only source reinspection, no byte adoption, no path-directory rebuild, and content-independent checks. No claim is made about curator runtime implementation coverage. Existing E6 negative tests include five-to-one boundary narrowing, no-system overlay/import narrowing, healed negatives, exact inventories and dry-run refusal shapes.

## File scope

```text
 CHANGELOG.md                                       |  34 ++
 conformance/v1/manifest.json                       |   4 +
 .../vectors/environments-path-kind-admission.json  | 541 +++++++++++++++++++++
 .../0012-context-packages-and-semver-locks.md      |  76 ++-
 protocol/environments.md                           | 151 +++++-
 release/1.0.0-rc.9.json                            |   4 +-
 tools/test_validate.py                             | 385 +++++++++++++++
 tools/validate.py                                  | 350 +++++++++++++
 8 files changed, 1516 insertions(+), 29 deletions(-)
```

## Mechanical edit-stream comparison

```text
Candidate paths: ['CHANGELOG.md', 'conformance/v1/manifest.json', 'conformance/v1/vectors/environments-path-kind-admission.json', 'decisions/0012-context-packages-and-semver-locks.md', 'protocol/environments.md', 'release/1.0.0-rc.9.json', 'tools/test_validate.py', 'tools/validate.py']
Union paths: ['CHANGELOG.md', 'conformance/v1/manifest.json', 'conformance/v1/vectors/environments-path-kind-admission.json', 'decisions/0012-context-packages-and-semver-locks.md', 'protocol/environments.md', 'release/1.0.0-rc.9.json', 'tools/test_validate.py', 'tools/validate.py']
CHANGELOG.md edit stream identical: True
conformance/v1/manifest.json edit stream identical: True
conformance/v1/vectors/environments-path-kind-admission.json edit stream identical: True
decisions/0012-context-packages-and-semver-locks.md edit stream identical: True
protocol/environments.md edit stream identical: False
--- accepted edits protocol/environments.md

+++ union edits protocol/environments.md

@@ -119,17 +119,19 @@

 -boundary and pin-hash verdict for the profile's lock, marker, and
 -named store entries, naming the failing check and, for an enclosing
 -failure, the boundary (environments root or store root) when the profile
--is `environment_store_untrusted`. Both commands follow
+-is `environment_store_untrusted`, the active codex-seed revision (`A` or `B`, section 7.4) with its
 +boundary and pin-hash verdict for the profile's lock, marker, named store
 +entries, and — for a `path` root or overlay — the source directory,
 +naming the failing check and the path or, for an enclosing failure, the
 +boundary (environments root or store root) when the profile is
-+`environment_store_untrusted`. Both commands follow
--current-reporting, or re-applying observation is non-conforming.  The nine
--retired `expected/environments/*` sets are regenerated under the v2 type
--line. A manager claiming this capability MUST pass the complete vector set;
-+current-reporting, or re-applying observation is non-conforming; and the
-+section 2.2 and section 4 path-kind admission cases
++`environment_store_untrusted`, the active codex-seed revision (`A` or `B`, section 7.4) with its
+-current-reporting, or re-applying observation is non-conforming.;
+-and the section 7.4 codex-seed
++current-reporting, or re-applying observation is non-conforming;
++the section 7.4 codex-seed
+-`codex_seed_record` marker shape. The nine
++`codex_seed_record` marker shape;
++and the section 2.2 and section 4 path-kind admission cases
 +(`vectors/environments-path-kind-admission.json`) — the `git` MCP
 +declaration admitted, the `path` root, overlay, and onboarding-import MCP
 +declarations refused with `mcp_declaration_path_source_refused` naming the
@@ -146,6 +148,4 @@

 +evaluation of a `path` directory failure reporting
 +`environment_store_untrusted` with no rebuild planned and mutating
 +nothing, and the negatives whose admitted, current-reporting,
-+rebuilding, or would-rebuild-reporting observation is non-conforming.
-+The nine retired `expected/environments/*` sets are regenerated under the
-+v2 type line. A manager claiming this capability MUST pass the complete vector set;
++rebuilding, or would-rebuild-reporting observation is non-conforming. The nine
release/1.0.0-rc.9.json edit stream identical: False
--- accepted edits release/1.0.0-rc.9.json

+++ union edits release/1.0.0-rc.9.json

@@ -1,4 +1,4 @@

--    "manifest_sha256": "sha256:7342f14dc74ccb56c4f1fac6ffdaa10694d791bd39f6627786c603d304ba3ad5",
-+    "manifest_sha256": "sha256:997b8de46fc5fc80a1814c6b0be8ec69d170f275101b89eba13ed352addba523",
--    "required_manifest_sha256": "sha256:7342f14dc74ccb56c4f1fac6ffdaa10694d791bd39f6627786c603d304ba3ad5"
-+    "required_manifest_sha256": "sha256:997b8de46fc5fc80a1814c6b0be8ec69d170f275101b89eba13ed352addba523"
+-    "manifest_sha256": "sha256:92d1e090065a4027f9aaa753b2826d4a1cdb46153e176501aca481620b0c18d1",
++    "manifest_sha256": "sha256:5ff658fa2889d4bb3d9e3173dfe6a7e6a47e929f29e2cd8f7def9d5c2b06cbdd",
+-    "required_manifest_sha256": "sha256:92d1e090065a4027f9aaa753b2826d4a1cdb46153e176501aca481620b0c18d1"
++    "required_manifest_sha256": "sha256:5ff658fa2889d4bb3d9e3173dfe6a7e6a47e929f29e2cd8f7def9d5c2b06cbdd"
tools/test_validate.py edit stream identical: True
tools/validate.py edit stream identical: True
```

## Verification transcripts and bounds

Reviewer reran both complete gates; no prior attached test result was substituted. Disposable byte copy: `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/e6-review-quag4zqx/delivery`. Temporary Git baseline was the staged index (`git init`, `git add .`, `git write-tree`), sufficient for the Makefile’s working-tree-versus-index diff. No commit or source-worktree mutation was needed. After regeneration, direct byte comparison against delivery found 0 mismatches over all tracked regular files; this additionally checks the regenerated artifacts independently of Git normalization.

Commands used `PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`.

`make regenerate-check`: exit 0.
```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

`make validate`: exit 0.
```text
python3 tools/validate.py
validated 62 schemas and 1104 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...........................................................................................................................................................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 475 tests in 770.129s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.048s
```

The Python suite took 770.129 seconds; the full command exceeded the brief’s approximate ten-minute bound, but remained alive under sequential tool polls and returned its actual exit 0 before this verdict. No background process or unfinished gate is being handed off. Future runs should split the slow production-entry mutation tests into bounded subsets.

Additional reviewer negative probes through `tools/validate.py:8141 main()` used in-memory malformed capability identities for the two vector families, with source files unchanged. 2/2 families refused with main exit 1 (probe harness exit 0):
```text
validation failed: path-kind vector has the wrong capability identity
validation failed: codex-seed vector has the wrong capability identity
environments-path-kind-admission.json main exit 1
environments-codex-seed.json main exit 1
2/2 corrupted vector families refused through main(); production code unchanged
```

Coverage bound: 475/475 discovered Python tests passed, Go generator package passed. E6 contains 36 test methods over 5 MCP, 14 boundary and 3 dry-run cases; these are spec/conformance checks, not evidence that a manager runtime implements the contract. No exhaustive mutation-coverage percentage is claimed.

Final source check: `git status --porcelain` empty; HEAD unchanged at 1ca4b3df32baceff99f6be4b136f39addfa4da00.
