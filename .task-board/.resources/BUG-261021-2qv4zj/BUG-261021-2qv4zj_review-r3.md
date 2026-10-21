# BUG-261021-2qv4zj review r3

Head: `d20c4363cf2dea88ce7418fa594664a466c116c8`  
Base: `2c89fd46443d461e25bd377fb94a0c6754282838`  
Verdict: **changes_requested** — one P1 publication-coverage finding (F3). F1 and F2 remain closed in the exercised cases.

The task, reviewer assignment, seven brief/precondition resources, previous reviews and producer r3 results were checked before acknowledgment. The exact head was fetched and checked out detached. All four introduced commits verify `G`. Source trees remained unchanged; reviewer tests and control implementations were supplied through Go overlays.

| Surface | Held or broken | Production exercise | Evidence |
| --- | --- | --- | --- |
| F2 context publication | Held | Exact-source CLI installs undeclared, clean and reserved-host context modules; checks diagnostic, source/lock absence on refusal, and native root-document bytes on allowance. | `focused.log`, TestProfileInstallMarkdownContextCLI; undeclared contact refused before publication. |
| Skill publication coverage | **Broken, F3** | Independent exact-source CLI profile installs with implicit-none skill capabilities and contacts in `.research/notes.md`, `references/README.md`, or `assets/tests/a.md`; then read both native skill surfaces; additionally provision and inspect both managed homes through Resolve. | `independent-native.log`: all three installs succeed, lock exists, and excluded contact bytes remain readable in both native skills directories. Declared-host `references/guide.md` control passes. `independent-managed-provisioned.log` repeats the same result through Resolve(Repair:true). |
| Skill roots, agent definitions, templates and triggers | Held for CopyContext; native exception above | Both gates, copier and independent path/byte checks for agents, references, templates, examples, data and triggers; scripts with/without exports and build/runtime roots in focused suite. | `focused.log`, `independent-matrix.log`. Native profile publication must also satisfy this projection. |
| Context classes, selectors and extensions | Held | Both public gates scan declared root/system modules, selected modules, README and non-Markdown names; independent additional context path cases. | `focused.log`, TestMarkdownScopeContextModules; `independent-matrix.log`. Missing module specifically refuses with scan-incomplete. |
| Shared module reader and path validation | Held | Traced ReadModule from ValidateModules, source audit and loadMaterial. Production module validation and malformed path cases rerun through contextpkg tests; independent probes reject absolute/parent-component declarations, read a case alias through ReadModule, and refuse a linked module through ValidateModules. | `independent-module-paths.log`: same filepath.Join/ReadFile path on audit/materialization; parser PortablePath and validation reject invalid shapes. A rejected malformed fixture is not counted as published-module coverage. |
| MCP, unknown kinds, invalid context fallback | Held within changed Markdown contract | MCP profile install plus Gate/GateReadOnly with out-of-root contact, clean and reserved/declared controls. | Focused other-kind/MCP tests. MCP publishes server configuration, not package Markdown; no new MCP description/template Markdown publisher found. |
| Ruleset/cache | Held | Real production ruleset-7 empty allowance seeded under old semantics, both hash frames, current read-only/writable refusal, retained old record, new cache reuse; cross-kind records cannot lend allowance. | Focused ruleset-7 and kind tests; rulesets 4/5/6 regressions also pass. |
| Case, exclusions and filesystem neighbors | Held on this filesystem | Nine root aliases, all 19 fixed exclusions and case neighbors, nested mixed case, NFC/NFD, sharp-s, dotless-i and root symlink negative control. | `focused.log`, `independent-matrix.log`. The host filesystem is case-insensitive; duplicate roots on a case-sensitive volume were not exercised. |
| Canonical context bytes | Held | Independent multi-root fixture run with current copier and baseline CopyContext implementation supplied by overlay; scripts enabled/disabled, excluded names and build root. | `canonical.log` / `canonical-old.log`: identical path lists and SHA-256 digests in both modes. |
| Mutation controls | Held | Drop context loop carried into child CLI build; old case matcher; all-Markdown scan; old namespace; kind equality and module-read weakening. | TestNetworkContactNarrowingMutants passes only after each control yields its named behavioral failure, exit 1. |
| CIP-0021 and pinned spec | Held | Selected-directory lifecycle, metadata and narrowing tests; pinned corpus `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; core audit/selection/context tests and all 45 producer-named envprofile conformance entry points. | Conformance, conformance-directory, conformance-core and conformance-env-rest logs, exit 0. |
| Documentation and producer claims | **Broken in per-kind skill row** | Re-derived publication table against profile native/managed publishers, not only copier. Other selection/cache statements traced and exercised. | Producer claim that every skill publisher uses the core section 3 projection is contradicted by F3. Both native and managed-home publication were independently exercised for F3. |
| Public repository hygiene | Held | Complete added-line review plus bounded personal path/host/email search. | No personal paths, host/user names or client/employer disclosures found in introduced diff. |
| Build, vet and default suite | Held | Ordinary `.git` validation clone at exact head, Go 1.25.5 darwin/arm64, private caches/home, short TMPDIR, inherited Git overrides unset. | Final execution ledger follows. |

## F3: narrowed skill audit disagrees with native publication

`internal/audit/audit.go:905-910` limits skill Markdown to WalkContext, and lines 930-931 omit all other Markdown from the fallback. However, `internal/envprofile/managed.go:279-290` assigns the raw selected package root to the skill; `switch.go:744` uses it for native skill links, and `managed.go:3065` uses it for managed-home links. Neither publisher applies CopyContext. A copy fallback also copies that raw tree.

The independent probe uses a real local Git skill and a valid context profile, explicit strict/high/null audit configuration, schema-3 skill capabilities `{}`, and the actual `curator profile install` command. For each excluded file, the command exits 0 and creates the lock; reading the installed Claude and Codex `skills/scoped-guide/<file>` returns `Read https://api.example.net/guide`. The nearby declared-host visible file installs and is readable, proving the fixture does not simply reject everything. The native probe fails only on the publication assertion, not setup or compilation. A second independent probe calls production Resolve with Repair:true after installation and reproduces the same three mismatches in both managed homes; its visible control passes. The first managed attempt omitted provisioning and returned environment_home_stale for every case; that setup-only result is not finding evidence.

This demonstrates publication, not a claim that a provider automatically reads every file. It violates the review brief's requirement that every agent-facing publisher be covered by the audit contract. These contacts were scanned by the base all-Markdown implementation; the narrowed detector now admits them while the existing publisher still exposes them.

Required fix: reconcile profile native and managed skill publication with the intended audited projection, including copy fallback, status/repair and surface identity. Preserve the requirement that repository-only Markdown does not create false refusals. Keep actual CLI regressions for excluded repository files, fixed excluded names/directories, runtime/build roots and exported-script Markdown, with positive visible/declared controls. If raw-tree publication is a deliberate contract requirement, resolve that incompatibility explicitly rather than describing raw publication as CopyContext projection.

## Reproduction and execution evidence

The companion `BUG-261021-2qv4zj_native_projection-r3_test.go` is an overlay test in package envprofile. Map it to a virtual `internal/envprofile/reviewer_native_projection_test.go` using a Go overlay JSON, then run `env -u GIT_DIR -u GIT_WORK_TREE go test -overlay "$OVERLAY" -count=1 -v -timeout=10m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'` in the exact-head validation clone. It builds the production CLI itself. Expected current result: exit 1, three named publication failures and one passing visible/declared control. Use a private HOME/GOCACHE/GOPATH/GOMODCACHE and short TMPDIR; the test supplies isolated native homes and fixture-only Git URL rewrites.

The companion managed probe is `BUG-261021-2qv4zj_managed_projection-r3_test.go`; overlay it as `internal/envprofile/reviewer_managed_projection_test.go` and select `^TestReviewerManagedSkillProjection$`. It calls Resolve with Repair:true to provision the isolated managed homes before reading them.

Default suite: **85 packages**, 80 passed, 5 without test files; test/subtest results: 11713 pass, 199 skip, 0 fail. Every package from an independent go-list inventory completed. Conditional/platform/corpus skips are not counted as passing checks. Pinned conformance supplements ran separately. The 56 structural vectors and two identity vectors documented in the changed guide passed.

All test commands used `env -u GIT_DIR -u GIT_WORK_TREE`; the full command was `go test -json -p=2 -count=1 -timeout=90m ./...`. No overlay was supplied to the default suite.

| Run | Exit | Seconds |
| --- | ---: | ---: |
| `build` | 0 | 1.51 |
| `build-cli` | 0 | 0.77 |
| `canonical` | 0 | 0.70 |
| `canonical-old` | 0 | 0.79 |
| `conformance` | 0 | 176.78 |
| `conformance-core` | 0 | 58.16 |
| `conformance-directory` | 0 | 82.24 |
| `conformance-env-rest` | 0 | 64.86 |
| `focused` | 0 | 31.34 |
| `full-suite` | 0 | 1372.03 |
| `independent-managed` | 1 | 7.38 |
| `independent-managed-provisioned` | 1 | 7.43 |
| `independent-matrix` | 0 | 1.06 |
| `independent-module-paths` | 0 | 2.33 |
| `independent-native` | 1 | 5.59 |
| `vet` | 0 | 5.64 |

The two publication probes intentionally exit 1 on the reproduced F3 assertion. `independent-managed` was the initial unprovisioned fixture attempt and is not evidence; `independent-managed-provisioned` is the valid managed reproduction. All other listed commands exit 0. Exact argv, exit codes, JSON test output and reviewer overlay sources are retained locally under `.out/BUG-261021-2qv4zj-review-r3-evidence/`. The report and the two self-contained publication probe sources are board outcome resources.


```json
{"verdict":"changes_requested","head":"d20c4363cf2dea88ce7418fa594664a466c116c8","findings":[{"id":"F3","severity":"P1","file_line":"internal/audit/audit.go:905-910,930-931","mechanism":"Skill Markdown is restricted to WalkContext, but profile native and managed skill links publish the raw selected package root. Excluded Markdown is admitted and remains available on native agent surfaces.","reproduction":"Run the companion exact-source CLI overlay probe. A schema-3 skill with capabilities {} and https://api.example.net/guide in .research/notes.md, references/README.md or assets/tests/a.md installs with exit 0 and a lock; both native skills/scoped-guide paths retain the contact bytes. The declared visible-file control passes.","required_fix":"Align every profile skill publisher and copy fallback with the audited context projection, including surface identity/status/repair, while retaining repository-noise admission. Add real CLI publication regressions for all excluded classes and positive visible controls; resolve any intended raw-tree publication contract explicitly."}]}
```
