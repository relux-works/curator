# BUG-261018-bisu6n review revision 1

Candidate tree: `5116fb26b41e41f5d668f94797f9af1191618aba`.
Base and freshly fetched origin/main: `b9caa20901ee859eec20e649940693033e6297d4`; no overlap or stale revision.
Signed hosted snapshot: `102a56dea6d80480bbd9c928138bd0fad66e5dee`, identical tree.
Coverage reviewed before source assessment: 7 of 7 producer AC rows have named driving tests or validation commands. The review brief supplies the seven surface rows below. No numeric budget was supplied.

| Row | Result | Attack and evidence |
| --- | --- | --- |
| 1 Early pnpm restoration | held | conformance_test.go:1127,1735,1857: registration precedes child creation; TestPNPMTemporaryTreeRemovable pass/Fatal/capture-Fatal runs the real body. Deletion and capture-only bypass overlays fail with non-writable-entry diagnostics. Reviewer pnpm.log records targeted command, exit 0. |
| 2 Sweep and neighboring fixtures | held | Independent rg over Chmod/chmod/GOMODCACHE/modcacherw/readonly literals compared with gaps. Additional hits are writable modes, comments, mode restoration, or the new witnesses. Nested sealed-tree/external-link attack rerun in npm and both Yarn packages: neighbors.log, exit 0. Existing archive/cache owners and exceptions are enumerated in producer gaps. |
| 3 All Linux jobs | held | Walked all nine runs-on declarations: seven Linux-routable job definitions, one macOS-only and one hosted matrix. ci.yml:141,303,551,618,651,695,723 end in exact guarded shared invocation. TestLinuxCIWorkflowTemporaryTreeRestorePlants attacks omission, echo-token, success-only, hosted-only, ordering, error policy and shell; workflow.log exit 0. |
| 4 Self-heal boundary | held | scripts/restore-runner-tmp.sh:5-24. Actual bash entry tested for unset/empty/root/root-alias/double-slash/non-directory/missing/symlink, readonly current and stale children, external link, trailing newline and chmod failure. TestRunnerTemporaryTree* suite including behavioral narrowing controls rerun; workflow.log exit 0. |
| 5 Witness and plant controls | held | Reviewer reran TestPNPMTemporaryTreeCleanupControls via Go overlay (deletion and capture-only bypass) and script narrowing controls. Expected child exit 1 must contain named failure, positive controls pass. pnpm.log and workflow.log, outer exit 0. Producer's 49 external workflow-file plants are reused from digest-verified evidence; these invoke the workflow gate binary, not merely token grep. |
| 6 Scope and HOME | held | Exact 14-path diff contains only test files, CI, script and runner docs; no product file, README or CHANGELOG delta. New helper roots are t.TempDir-owned; real manager uses existing private execution context. Direct HOME/UserHomeDir scan of changed conformance fixtures has no matches. External target assertions in all four helper witnesses and script tests stayed unchanged. |
| 7 Validation and runner evidence | held | Re-fetched Actions run/jobs API: exact signed commit/tree, successful gate. Nine Linux executions have successful restore and Complete runner. Producer evidence manifest digests verified; 78 accepted shard log/zero-exit records present, with explicit failed-shard replacements and retained individual passes. Full local suite, lint/vet/build reused, not rerun wholesale. Signature independently verified G. |

## Verdict

```json
{"verdict":"accepted","revision":1,"findings":[],"notes":["No review budget or separate catalog table was supplied; the seven ordered review-brief rows define this sweep.","POSIX modes only: ACL repair and hostile concurrent pathname replacement remain stated bounds.","Hosted evidence proves successful restore and Complete runner on curator-4, not an inventory of stale contents before restoration."]}
```

All seven rows held for the named attacks; this is not proof of absence. Free hunt produced no findings: direct script invocation with directory and dangling symlinks, leading-dash/space/multiple-newline/dot root names preserved outside target bytes and modes and restored nested trees.

## Reviewer commands

- `go test -count=1 ./tools/ciworkflow` — exit 0 (workflow.log).
- With the producer's pinned runtime/spec environment: `go test -count=1 -tags tmphygienecontrols -timeout=8m -v ./internal/pnpmsource -run 'Test(PNPMTemporaryTree|WritableTempDir)'` — exit 0; both mutant children exit 1 for the intended witness (pnpm.log).
- `go test -count=1 ./internal/npmsource ./internal/yarnclassicsource ./internal/yarnmodernsource -run '^TestWritableTempDirRestoresNestedTrees$'` — exit 0 (neighbors.log).
- `python3 .temp/review-bisu6n/free-hunt.py` — exit 0 (free-hunt.log).
- All 14 candidate paths byte-compared with the immutable tree before review acceptance. No source edits, commits or branch changes made.
- Fresh API run: https://github.com/relux-works/curator/actions/runs/39153364696. A diagnostic aggregation initially encountered null runner_name on skipped jobs; corrected read handles skipped rows explicitly and verifies all nine scheduled Linux jobs. No failed read was treated as absence.

## Linux executions

| Job | Runner | Result |
| --- | --- | --- |
| Lint | relux-rise1-curator-1 | success; restore and Complete runner success |
| Race (ubuntu-latest) | relux-rise1-curator-4 | success; restore and Complete runner success |
| Gate self-test (ubuntu-latest) | relux-rise1-curator-2 | success; restore and Complete runner success |
| Test (ubuntu-latest) | relux-rise1-curator-3 | success; restore and Complete runner success |
| Naming gate | relux-rise1-curator-1 | success; restore and Complete runner success |
| Interop conformance gate | relux-rise1-curator-2 | success; restore and Complete runner success |
| Go 1.25.5 driver (ubuntu-latest) | relux-rise1-curator-3 | success; restore and Complete runner success |
| Go 1.27.1 driver (ubuntu-latest) | relux-rise1-curator-2 | success; restore and Complete runner success |
| Go 1.26.0 driver (ubuntu-latest) | relux-rise1-curator-1 | success; restore and Complete runner success |

## Evidence and logbook

Producer resources reused: `BUG-261018-bisu6n_gaps.md` and `BUG-261018-bisu6n_evidence.json` (digest-checked container; accepted-log-plan, environment, partial-green-reuse and exact signed identity reviewed). Reviewer reproduction scripts/logs and sanitized API projections are attached in `BUG-261018-bisu6n_review-evidence-rev1.json`. No additional defect or product decision arose. This task-scoped outcome records the review logbook without writing the control root.
