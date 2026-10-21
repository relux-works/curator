# BUG-261021-2qv4zj — review r1

Verdict: **changes_requested**. One P1 regression: on a case-insensitive filesystem, a differently cased context root bypasses the new Markdown audit filter while its contents are copied into active agent context.

Head: `2c01da83de4007c5d6f6d4edf3a887dc85707795`.
Base: `2c89fd46443d461e25bd377fb94a0c6754282838`.
Both introduced commits (`35f8a18d50795636e342baf87abc491bf7ce1aca`, `2c01da83de4007c5d6f6d4edf3a887dc85707795`) verify locally with `%G? = G`. Detached worktree and independent normal-`.git` validation clone both matched the requested head. No repository source was edited.

Read all three board preconditions: `W-AUDITSCOPE-brief.md`, `reviewer-BUG-261021-2qv4zj.md`, and `W-AUDITSCOPE-review-brief.md` (the latter was initially missing, then attached by the orchestrator). Also read the producer result. Normative comparison used core §3.1 at pinned spec `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`.

| Surface | Held or broken | Attack run | Evidence |
| --- | --- | --- | --- |
| Scope correctness | **Broken — F1** | Independent public `GateReadOnly`/`Gate` and actual `CopyContext` probes for every eligible root, all 19 exclusions at depth, root lookalikes, uppercase extensions, nested build/runtime roots and siblings; then differently cased root names on a case-insensitive filesystem. | Ordinary probes pass. All nine case-alias variants copy the undeclared contact but the new gate admits it; the base detector blocks the same nine fixtures. |
| Single source of truth/context assembly | **Broken — F1** | Compare the materialized files with the audit decision, including actual filesystem lookup rather than only calling the shared predicate with identical strings. | `CopyContext` starts from canonical root strings and `os.Stat` resolves aliases; the audit walks actual directory-entry spellings. Sharing `ContextPathIncluded` does not make these inputs equivalent. |
| Non-Markdown scan | Held | Public-gate probes for scripts, excluded-looking script descendants, canonical/legacy manifests, outside-root text/scripts and fallback manifest; repeat against the exact base detector through a source overlay. | Head and base both exit 0 for the same nine fixtures. Production diff leaves the non-Markdown predicate and detector body unchanged. |
| Ruleset/cache | Held | Seed a real ruleset-5 refusal with the base production detector, then run current read-only and writing gates on the same bytes in hash v1 and v2. Run prior-ruleset-4 allowance test and neighboring cache/mutation tests. | Independent seed and upgrade each exit 0: old block misses, read-only writes nothing, new allow caches, old record remains byte-identical. Prior allowance regression and version-reuse mutants pass their intended assertions. Bump is necessary in both directions. |
| CIP-0021 witness | Held | Run pinned schema, identity, lifecycle, declared-host and narrowing suites; inspect adapter relocation and unchanged corpus. | 56 schema cases and both identity fixtures pass. Lifecycle contact moves only from `notes.md` to `references/notes.md`; selected-package authority, finding and nonpublication expectations remain. No vendored corpus diff. |
| Regression and controls | Held for canonical paths | Public-gate matrix plus CLI noise/visible/declared cases; rerun every-Markdown, dropped-references, scripts and ruleset mutants. | Noise admits. Canonical SKILL/references/assets contacts refuse; declared host admits. Required two mutants produce child exit 1 with the named behavioral assertion; harness exit 0. |
| End to end | **Broken — F1** | Build exact-head executable; invoke `curator profile install <local-profile>` with isolated HOME/native homes, explicit strict/high audit and fixture Git URL rewrites. | Five normal controls pass. `References/guide.md` and `SKILL.MD` attacks both make CLI exit 0, publish the profile lock, activate the profile, and expose the undeclared URL through the native Claude skill path. |
| Public hygiene/signatures/default suite | Held | Inspect complete introduced diff, whitespace and personal-information patterns; verify signatures and clean exact-head clone; run build/vet/default suite with private cache. | No personal paths, operator identities or client/employer material found in the introduced files; public module imports are expected. Both signatures G. Build/vet exit 0. Default suite exit 0: 80 passing packages, 5 packages with no tests; 11627 passing tests/subtests and 180 skipped tests/subtests, no failures. |

## F1 — case-alias root escapes the audit but reaches agent context

Location: `internal/audit/audit.go:847`, using `internal/whitelist/whitelist.go:129–138`; compare `CopyContext` at lines 55–71.

The audit walks the snapshot root and sees the actual name `References/guide.md`. The new exact-string root predicate rejects `References`, so the detector skips the file. Context assembly instead constructs `snapshot/references`, which resolves to that same directory on a case-insensitive filesystem. Its relative walk names the file `references/guide.md`, accepts it and copies the URL. `SKILL.MD` has the same problem: validation and copying resolve `SKILL.md`, while the audit skips the actual `SKILL.MD` entry.

Reproduction through production entry points:

1. On a case-insensitive filesystem, create a valid skill with canonical metadata and an empty network capability declaration. Put `https://review.example.net/endpoint` in `References/guide.md`; a second fixture puts it in a valid `SKILL.MD` instead of `SKILL.md`.
2. Require that skill from a local profile. Use an isolated, valid manager configuration with `audit.enabled=true`, `mode=strict`, `fail_on=high`, `backend=null`; serve the Git requirement from a local fixture using a fixture-only URL rewrite.
3. Run the exact-head executable with `profile install <local-profile>`.
4. Observe exit **0**, a published profile lock, activated profile, and the same URL readable through `skills/scoped-review/references/guide.md` or `skills/scoped-review/SKILL.md` in the native Claude home. The independent expected-refusal test exits **1**, specifically because the CLI succeeded; it separately asserts the published bytes and lock.
5. Public-gate/real-copy probes also reproduce for `ASSETS`, `AGENTS`, `.SKILL_TRIGGERS`, `Templates`, `Examples`, `Data`, and command-less `Scripts`. All nine expected-refusal cases fail at the review head and pass with the base production audit file overlaid. Declaring the host remains an admitted control.

Required fix: make snapshot validation, context enumeration and Markdown auditing agree on the actual files that can reach an agent, including filesystem case aliases. Reject noncanonical aliases consistently before publication or use a shared filesystem-aware enumeration that audits every copied file. Do not simply make protocol paths case-insensitive on every platform: the spec requires case-sensitive comparison. Ensure an already-cached ruleset-6 allowance cannot bypass the correction (invalidate it or reject the alias before cache reuse). Add case-insensitive-filesystem regression coverage for `SKILL.MD` and directory aliases, with canonical/declared-host controls and a case-sensitive-host control where available.

## Validation receipts

Platform: darwin/arm64; Go 1.25.5; pinned Rust 1.91.0; Node 26.12.0. Tests unset `GIT_DIR` and `GIT_WORK_TREE`, used a private per-task `GOCACHE`, lane-local short `TMPDIR`/`GOTMPDIR`, and a PATH containing the verified Go/Node tools, pinned Rust and system tools. The default suite ran in the clean normal-`.git` clone. Supplemental conformance used the exact spec pin above.

| Run | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | All packages build. |
| `go vet ./...` | 0 | No vet findings. |
| `go test -count=1 -timeout=60m -json ./...` | 0 | Default suite exit 0: 80 passing packages, 5 packages with no tests; 11627 passing tests/subtests and 180 skipped tests/subtests, no failures. |
| Affected audit/envprofile/skillcheck/whitelist tests with pinned corpus | 0 | Selected Markdown, prior-cache, network, capability-cache, skill-directory, profile-install, context and validation tests pass; no skipped tests. Expected-red child mutant output is retained, not confused with an outer-suite failure. |
| Pinned `TestSkillDirectory*` in contextpkg/contextlock | 0 | 56 schema vectors, two lock identities and closed-successor cases pass. |
| Independent normal scope/root/non-Markdown probes | 0 | Public gates and real copying agree for canonical fixtures. |
| Same independent non-Markdown probes with base audit overlay | 0 | Historical behavior preserved. |
| Actual base-ruleset-5 cache seed; current upgrade | 0 / 0 | Both hash framings; read-only/no-mutation and retained-old-record assertions. |
| Exact-head CLI canonical/noise/declared controls | 0 | Five expected outcomes, including nonpublication on refusal. |
| Case-alias public-gate/real-copy probes, head / base | 1 / 0 | Nine head regressions; baseline blocks all nine. |
| Case-alias exact-head CLI probes | 1 | Both CLI subprocesses incorrectly exit 0; test proves URL and lock publication before reporting expected-refusal failure. |

Initial CLI fixture setup attempts exited 1 before reaching installation: the temporary system/user config lacked required schema, skills-root and projects fields. These are reviewer fixture failures, not product findings; the final complete config passes all five canonical controls. The separate case-alias failures above occur after successful installation and publication.

Reproducible reviewer sources are retained as `.out/BUG-261021-2qv4zj-probe_test.go`, `...-cache_test.go`, and `...-cli_test.go`. They are injected with Go `-overlay`, never copied into the repository. Run logs, exit receipts, full-suite JSONL and signature output share the `.out/BUG-261021-2qv4zj-` prefix. Local raw logs contain temporary paths and are not public board attachments. Scratch worktrees, clone, cache, fixtures and binary are removed after handover.

```json
{"verdict":"changes_requested","head":"2c01da83de4007c5d6f6d4edf3a887dc85707795","findings":[{"id":"F1","severity":"P1","file_line":"internal/audit/audit.go:847; internal/whitelist/whitelist.go:129-138","mechanism":"On case-insensitive filesystems, audit walks actual root spellings and skips References or SKILL.MD, while CopyContext and metadata validation resolve canonical references or SKILL.md to the same bytes. An undeclared network contact therefore reaches active agent context without a blocking finding.","reproduction":"With strict/high audit, install a valid local profile whose skill has no network declaration and contains https://review.example.net/endpoint in References/guide.md or valid SKILL.MD. Exact-head CLI exits 0, publishes the profile lock and exposes the URL through native Claude skills/scoped-review/references/guide.md or SKILL.md. Nine public-gate/copy variants fail their expected-refusal assertions at head and pass with the base detector.","required_fix":"Align validation, context enumeration and audit over actual filesystem entries; consistently reject noncanonical case aliases or audit every file that context assembly can copy. Preserve case-sensitive protocol semantics, prevent existing ruleset-6 allowances from bypassing the correction, and add case-alias production-entry regression tests with canonical and declared-host controls."}]}
```
