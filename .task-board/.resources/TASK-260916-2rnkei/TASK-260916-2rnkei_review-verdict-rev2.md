# TASK-260916-2rnkei — revision 2 independent review

Verdict: **accepted** (CR-TASK-260916-2rnkei-2, revision 2). F1 and F2 from round 1 are resolved. No blocking findings.

Reviewed candidate: curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1i1gfo/worktree`, HEAD `684c9f1324d46b4938b2e5943f20c89e27971ec8`. Attached `TASK-260916-2rnkei_spec-patch_rev2.patch` and independent `git diff HEAD` have identical bytes and stable patch ID `acd7b7c89665dff8e7c512151f3bc665c3965f9a`. Candidate repositories were not edited. Scratch regeneration and replacement probes ran under `/tmp/e3-review2`.

## Per-item review

File references below are relative to the curator-spec candidate.

| Requirement | Evidence and exact text | Result |
|---|---|---|
| §7.4 seed row and honest confidence | `protocol/environments.md:1273`: “revision A copies it whole”; “revision B copies every top-level member except `mcp_servers`”; “the remaining member shapes are **docs-confidence**”. Existing verified claims remain scoped to the pinned tool/whole-copy behavior. | Pass |
| Two labelled revisions, A first | `protocol/environments.md:1282`: “a manager MUST ship revision A before revision B”; lines 1284 and 1291 explicitly label “Revision A (warning release)” and “Revision B (flip release)”. | Pass |
| A copy, warning, names and migration | `protocol/environments.md:1285`: “provisioning MUST emit `mcp_native_servers_ungoverned` (warning) naming every native `mcp_servers` entry”; lines 1286–1290 state the next revision stops inheriting them, “declare the server in the profile's MCP set, or accept the loss”, and require ungoverned status rows outside lock/allowlist. | Pass |
| B stripping and once-only report | `protocol/environments.md:1292`: “the `mcp_servers` table and every `mcp_servers.*` sub-table removed”; lines 1294–1299 retain every other top-level member, require valid TOML, “report the stripped names once” with `mcp_native_servers_not_inherited`, and list not-inherited rows. | Pass |
| Snapshot and provisioning-only boundary | `protocol/environments.md:1301`: “Both revisions write the `codex_seed_record` marker record”; line 1304: “names only, never server commands or env values”; lines 1305–1309 require no warning for empty snapshots, still record them, and preserve existing home bytes. | Pass |
| F2: shipped versus recorded revision | `protocol/environments.md:1309`: “The manager-shipped seed-rule revision and the home's recorded seed revision are distinct”; lines 1311–1317 specify A-home/B-manager ungoverned names plus `mcp_seed_unstripped`, re-provision hint, empty-snapshot exception, and absent-record handling. Matching diagnostics at 1443 and detailed status predicate at 2645. | Pass |
| §7.7 closed diagnostics | `protocol/environments.md:1441–1443` adds exactly `mcp_native_servers_ungoverned`, `mcp_native_servers_not_inherited`, `mcp_seed_unstripped`, all warnings. Exact spellings agree with §12, CHANGELOG, and vectors/gate. | Pass |
| §7.8 adapter asymmetry | `protocol/environments.md:1466`: “`--strict-mcp-config` disables every other MCP configuration”; 1467: “`-p curator-mcp` layers the profile set over the seeded base”, with B limited to homes provisioned under B; 1468 records opencode merge order and “project-level servers remain”; 1469: “none: no channel, no MCP configuration in a managed launch”. | Pass |
| §8.2 marker shape and immutability | `protocol/environments.md:1631`: “the closed object `{ revision, native_mcp_servers }`”; revision exactly A/B, ascending-byte-order names-only snapshot, absent on other homes; 1639: “The recorded `revision` never changes after provisioning”. Schema at `schemas/v1/agent-environment-marker-v1.schema.json:87` requires both fields, A/B enum, unique nonempty string array, `additionalProperties: false`; line 124 forbids the record outside managed-home mode. | Pass |
| Schema versioning | `schemas/v1/README.md:6` identifies frozen manager/system schema-1 files; the environment marker belongs to the unreleased environment family and is not in those byte-frozen sets. The optional marker extension is brief-authorized and unchanged from accepted round-1 content. Existing schema instances unchanged and validated. | Pass |
| §12 posture and currentness | `protocol/environments.md:2640–2647` distinguishes active revision and per-home record; 2672–2674 names all three diagnostics and says “never make a row non-current”; 2692–2701 defines shipped revision and re-provision reporting, referring to §7.4's empty-snapshot exception. | Pass |
| §12.1/§12.2 knobs and locks | `protocol/environments.md:2701`: “no configuration knob selects the revision”. Closed machine knob table starts at 2722; lockable set remains unchanged. No new knob/lock key is requested or introduced. | Pass |
| §13 conformance | `protocol/environments.md:2885–2895` names the new family and both rollout, server-free, pre-rule, and A-home/B-manager cases; 2906–2911 defines both conformance revisions and “MUST NOT claim revision B while still inheriting native servers”. | Pass |
| Provisioning and posture vectors | Opened all 7 provisioning and 8 posture cases in `conformance/v1/vectors/environments-codex-seed.json`. Inputs distinguish A/B, present/absent/empty tables, subtable-only/inline syntax, retained trust/model/TUI members and values, absent/A/B marker records, and non-codex homes. New mismatch case at line 148 supplies shipped B and recorded A with both warnings and repair hint. | Pass: 15/15 cases reviewed |
| F1 production gate and tests | `tools/validate.py:6004` defines `validate_environments_codex_seed_vectors`; `main()` registers it at 6868 and invokes every check at 6880–6881. It pins exact case inventories and discriminating inputs, recomputes parsed members/values, names, diagnostics/hints, closed marker record, and posture. `tools/test_validate.py:2453` adds 32 tests, including both requested replacement families. Independent production-entry attacks below rejected 3/3 replacements. | Pass |
| Generated marker cases | `tools/generate-vectors/environments.go:1261–1323` emits 2 positive and 5 negative cases: valid/empty record; unknown field, revision C, empty name, non-array names, and linked-home record refusal. Index entries at `conformance/v1/schema-cases/index.json:258`, 263, 478–498; manifest entries at 512–528 and 744–748. Opened generated cases; the main validator checks their declared outcomes. | Pass: 7/7 registered |
| Manifest and prior byte identity | New vector registered at `conformance/v1/manifest.json:4312`. Independently compared HEAD blobs: all 31/31 prior vector files and 974/974 prior non-index schema-case files are byte-identical. Regeneration proof below is green. | Pass |
| CHANGELOG and follow-up | `CHANGELOG.md:10–38`: entry starts “E3”, names both releases, all three diagnostics, names-only marker, mismatch posture, gate, and `TASK-260916-33abdk` manager/README follow-up. | Pass |
| Rev1 preservation and scope | Reconstructed rev1 from HEAD plus attached rev1 patch, compared against rev2. Differences are confined to F1 gate/tests/generated cases/derived hashes/seeded-member expectations and F2 wording/new posture vector/CHANGELOG. Marker schema is byte-identical to rev1. No E5/S5, proposals 0014–0018, or manager implementation edits. | Pass |
| Task outcomes and logbook | Both patches and producer evidence are board outcomes. Producer's task-scoped logbook is attached; no curator `LOGBOOK.md` change. Its historical round-1 statement deferring the validator is superseded by the explicit Revision 2 correction in producer evidence and this review. This verdict is the new task-scoped review outcome. | Pass |

## Empty curator repository delta

Independently inspected the exact CR base `0f0ae61766026cf2389ec4b91c09134cdb8aac62` to candidate `3c30a7f627678f20a8c70ce0f776d7ef7ef7260f`: zero changed paths; curator worktree is clean. This is correct for this leaf because its deliverable is a curator-spec normative revision, published as the attached spec patch and present in the separately named curator-spec Story worktree. The empty manager-repository delta is neither implementation delivery nor evidence that the specification is already merged. Acceptance routes to integration; this reviewer does not mark done, commit, or provide `commit_ack`. Manager implementation remains TASK-260916-33abdk.

## Independent validation

All verification below was rerun by this reviewer; no passing gate was accepted solely from producer evidence. Shell: bash, `set -o pipefail`; Python PATH prefix `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`. The `make validate` recipe was executed as its exact constituent commands, splitting Python discovery by module to keep each command below the headless time bound. The recipe itself has three gates: validator, all Python unittests, and Go tests.

```text
python3 tools/validate.py
validated 62 schemas and 1101 vector files
exit 0

PYTHONPATH=tools python3 -B -m unittest test_validate
Ran 307 tests in 324.897s
OK
exit 0

PYTHONPATH=tools python3 -B -m unittest test_implementation_coverage test_release_gate test_verify_release_commit test_verify_release_merge_policy
Ran 78 tests in 96.948s
OK
exit 0

go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors  3.151s
exit 0
```

Python coverage is 385/385 discovered tests across all 5 `test_*.py` modules, including all 32 new Codex-seed tests. All three `make validate` recipe gates exited 0. The monolithic `make validate` wrapper was not invoked; its full recipe was independently rerun as the bounded calls above. Final `git diff HEAD` is byte-identical to the attached rev2 patch, so the completed validation remains bound to that candidate.

Regeneration: copied candidate tracked files to scratch, `git init` and `git add .` established an index baseline without committing or touching the Story. `make regenerate-check`:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit 0
```

`git diff --check` in the actual candidate: exit 0. Regeneration ran only in scratch because running its writer in the read-only candidate would violate the review role, and comparing the uncommitted candidate to its old index is not an idempotence proof.

## Production-entry replacement probes

Each probe replaced a case body while retaining its required name, then regenerated the manifest with the repository generator and invoked the real `python3 tools/validate.py`. This avoids a manifest-hash failure standing in for semantic refusal.

```text
b-strips-servers-keeps-rest <- b-without-servers-no-warning: exit 1
validation failed: codex-seed case b-strips-servers-keeps-rest: native top-level members are not the pinned set (['mcp_servers', 'model', 'projects', 'tui'])

a-home-unstripped-under-b <- a-home-lists-ungoverned: exit 1
validation failed: codex-seed case a-home-unstripped-under-b: revision_shipped does not match the pinned one (B)

a-home-unstripped-under-b <- b-home-lists-not-inherited: exit 1
validation failed: codex-seed case a-home-unstripped-under-b: recorded revision does not match the pinned one (A)
```

Measured production-entry replacement refusal: **3/3**. Corpus inventory reviewed: **7/7 provisioning + 8/8 posture**. This proves the tested conformance-corpus branches and gate reachability; it does not assert runtime Codex behavior or manager implementation coverage. No manager execution was claimed. No blanket claim of exhaustive clause-mutant coverage is made.

## Findings and disposition

No blocking findings. F1 resolved by the reachable semantic gate, discriminating negative tests, generated schema cases, and independently rejected replacements. F2 resolved by distinguishing manager-shipped and home-recorded revisions and retaining ungoverned reporting with the re-provision warning for A homes under B.

The explicit no-inherited-server exception is stated in §7.4/§7.7 and the detailed §12 predicate; later summary sentences refer back to those sections. Schema-case marker fixtures prove structural shape, while adapter-specific posture belongs to the separate semantic vector family (the marker has no environment-id field). These are the stated bounds of this review, not additional implementation claims.
