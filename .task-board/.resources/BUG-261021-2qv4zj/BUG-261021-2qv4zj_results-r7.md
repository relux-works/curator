# BUG-261021-2qv4zj results r7

Branch: `fix/audit-markdown-scope-r3`  
Base: `08fa312e49437893b05ad87a888983601b1899c9`  
Delivered head: `df147f0883027a5dc737b2d2556ec208dde78931`  
Signature: `git log -1 --format='%G?'` returned `G`.  
Publication: normal push succeeded; `git ls-remote --heads origin refs/heads/fix/audit-markdown-scope-r3` returned the delivered head.

## Scope and changes

This delivery addresses the two test-fixture defects in `W-AUDITSCOPE-rework-r7-brief.md`. The task assignment and all 13 precondition resources were verified and read before acknowledgement. Board status was not changed.

- `internal/audit/markdown_kind_test.go`: extract the existing ruleset-7 source transformation into `rulesetSevenSource`. Normalize CRLF for matching, retain the exactly-one requirement for every existing anchor, and restore CRLF when writing the seed overlay. Add LF/CRLF positive controls and missing/duplicate multiline-anchor negatives. The real Gate seed test calls this transform and still checks both hash frames.
- `internal/envprofile/markdown_context_test.go`: replace the manual context-package setup with existing `writePackage` from `internal/envprofile/envprofile_test.go`. That helper writes the same package contract and calls existing `pathboundary.ProtectTree`, whose Windows implementation applies an owner-only mutation DACL recursively. Existing `pinOperatorHome` and `pinHomes` isolation remains. Product permission checks and every refusal/publication assertion remain unchanged.

Changelog line: Fix Windows Markdown audit test fixtures for CRLF source anchors and protected context-package permissions. No production or protocol behavior changes.

## Verification at the delivered head

Platform/toolchain: macOS, darwin/arm64, Go 1.25.5. Tests used a private home, build cache, module cache and GOPATH inside task scratch; inherited Git directory/worktree variables were removed. Nested Go processes used the pinned Go binary directory in PATH. `TMPDIR` and `GOTMPDIR` were `/tmp/cw-dev3`. No real operator configuration or credentials were needed.

| Check | Exit | Claim and limit |
| --- | ---: | --- |
| `TestRulesetSevenSourceLineEndings` | 0 | Six LF/CRLF cases: unique anchor produces ruleset-7 seed with its original endings and no current kind binding; missing and duplicate multiline anchors return the specific anchor error and no seed. |
| `TestMarkdownScopeInvalidatesRulesetSevenAllowance` | 0 | Real production Gate seeds and reuses a ruleset-7 allowance. Current audit rejects it in v1 and v2 hash frames while preserving the old record. |
| `TestNetworkContactNarrowingMutants/markdown-reuse-ruleset-seven` | 0 | The ruleset-reuse mutant exits 1 with the intended `old ruleset-7 record changed` assertion in both hash frames. Setup failure is not accepted as a kill. |
| `TestProfileInstallMarkdownContextCLI` | 0 | All three cases run: undeclared contact specifically refuses without metadata or native URL publication; clean and reserved-host controls install with metadata and expected native bytes. |
| Windows cross-vet, audit and envprofile | 0 | Windows-selected package code and tests pass static analysis. This is not Windows runtime execution. |

Audit JSON evidence: 12 passing test/subtest records, zero failures, zero skips (3 top-level tests; 9 leaf cases). Context CLI JSON evidence: 4 passing records, zero failures, zero skips (1 top-level test; 3 leaf cases).

Exact final-head commands (`$CURATOR_GO` denotes the pinned Go executable):

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestRulesetSevenSourceLineEndings$|^TestMarkdownScopeInvalidatesRulesetSevenAllowance$|^TestNetworkContactNarrowingMutants$/^markdown-reuse-ruleset-seven$' ./internal/audit
# exit 0; 2.98s

env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestProfileInstallMarkdownContextCLI$' ./internal/envprofile
# exit 0; 2.33s

GOOS=windows env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./internal/audit/... ./internal/envprofile/...
# exit 0; 0.54s
```

## Bounded regression controls

Three temporary overlays modify only the test helper; each ran the LF/CRLF regression and was checked for both exit 1 and its intended named assertion. They are not committed product changes.

| Plant | Exit | Required failure observed |
| --- | ---: | --- |
| Remove CRLF normalization | 1 | `TestRulesetSevenSourceLineEndings/CRLF/unique`: `ruleset-7 anchor not unique:` for the multiline kind-binding anchor. |
| Remove CRLF restoration | 1 | `TestRulesetSevenSourceLineEndings/CRLF/unique`: `ruleset-7 seed changed source line endings`. |
| Weaken exactly-one check to missing-only | 1 | LF and CRLF duplicate cases: `ambiguous anchor: seed empty=false, error=<nil>`. |

Each control command was:

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -overlay "$SCRATCH/<plant>.json" -count=1 -timeout=2m -json -run '^TestRulesetSevenSourceLineEndings$' ./internal/audit
# each expected exit 1; the control-output checker exited 0
```

## Delivery commands

All commands below completed with exit 0:

```sh
a2a inbox
task-board q 'get(BUG-261021-2qv4zj) { status assignee preconditionResources }'
# Assignment verified; each listed resource was materialized with resource get and read.
a2a post '#curator' --topic audit-md-scope --kind ack --to <orchestrator> 'ack BUG-261021-2qv4zj'
git fetch origin
git switch -c fix/audit-markdown-scope-r3 --track origin/fix/audit-markdown-scope-r3
git diff --check
git add internal/audit/markdown_kind_test.go internal/envprofile/markdown_context_test.go
git commit -S -m 'test: fix Windows Markdown audit fixtures' <descriptive-body>
git log -1 --format='%G?'
# G
git push -u origin fix/audit-markdown-scope-r3
git ls-remote --heads origin refs/heads/fix/audit-markdown-scope-r3
# df147f0883027a5dc737b2d2556ec208dde78931
```

The configured agent Git transport/signing environment was sourced before fetch, commit and push. The configured human commit identity was preserved. The working tree was clean after commit and validation. No PR was opened and no default branch was pushed.

## Open decisions and limits

No implementation decision is open. Hosted `Test (windows-latest)` on the landing PR is the required Windows runtime check; it has not been claimed as passing here. Local CRLF fixtures exercise anchor behavior, and cross-vet compiles/analyzes Windows-selected code, but neither executes Windows DACL operations. The full suite was not rerun for this narrow test-only rework; the brief requests the focused checks above.

The report is retained outside the repository. Attachment presence will be verified by a fresh `outcomeResources` query before handover. Task scratch, overlays, copied caches and materialized preconditions will be deleted after handover.
