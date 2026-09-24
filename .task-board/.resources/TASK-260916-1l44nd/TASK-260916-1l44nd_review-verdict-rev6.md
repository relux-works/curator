# Revision 6 review — changes requested

Candidate tree: `e33bfbf7aa455dbbd9fdbd92f5233eae69d4775f`. Base: `07453544e3defe726f5951af9023a0343c4402b7`. Previous reviewed tree: `f86bf9e434fa2b6aa07f044829f953ad43e5a0a5`.

Reviewed an exact git-archive extraction under the assigned Story worktree. Candidate/product source was not changed. Extra contract tests and temporary mutants were confined to the disposable extraction. Restored the mutated file and compared its bytes with the candidate Git object.

## F1 — HIGH: ABI 1 still leaves supported directory mutation rights unrestricted

`internal/scriptworker/landlock.go:71–75` adds the entire directory-only set only for ABI >= 2. REMOVE_DIR, REMOVE_FILE and all seven MAKE_* rights already exist in ABI 1; only REFER starts in ABI 2. Therefore ABI-1 write confinement currently handles only WRITE_FILE (0x2), instead of WRITE_FILE plus those nine rights (0x1ff2). The probe accepts ABI 1 (`landlock_linux.go:96–99`) and constructs this reduced mask; successful application can still attest `applied` while directory mutations outside the derived grants remain unrestricted. This is the same incomplete-control class as revision-5 F2, on a supported older ABI.

Independent sources: the [Linux v5.13 UAPI header](https://raw.githubusercontent.com/torvalds/linux/v5.13/include/uapi/linux/landlock.h) already defines all nine rights. The [kernel compatibility example](https://docs.kernel.org/userspace-api/landlock.html#defining-and-enforcing-a-security-policy) removes only REFER for ABI < 2, TRUNCATE for ABI < 3, and IOCTL_DEV for ABI < 5.

The regression is actively encoded in tests: `landlock_test.go:67–75` expects ABI 1 to handle only WRITE_FILE, and `preflight_test.go:436–467` expects outside-grant directory mutation to succeed on ABI 1. Passing these tests protects the defect. The production row must require mutation denial on ABI 1 too.

Reproduction: copy attached `TASK-260916-1l44nd_review-rev6-abi1_test.go` into `internal/scriptworker/reviewer_abi1_test.go` of a disposable candidate extraction, then run:

```sh
go test -count=1 -v -timeout=3m ./internal/scriptworker -run '^TestReviewerABI1MutationRights$'
```

Observed exit 1, 0.366s: all 9/9 rights fail; actual mask 0x2, expected 0x1ff2. This is executable proof against the production mask helper plus kernel semantics, **not** a local ABI-1 kernel exploit execution. The macOS host cannot supply that runtime proof.

Required correction: handle REMOVE_* and MAKE_* from ABI 1; gate only REFER at ABI 2, TRUNCATE at ABI 3, and IOCTL_DEV at ABI 5. Keep directory-only rights off file grants and retain the existing writable roots. Correct the source comments, unit expectations, production mutation-row ABI branch, and results.md ABI table. Expected write masks: ABI 1 = 0x1ff2; ABI 2 = 0x3ff2; ABI 3–4 = 0x7ff2; ABI 5+ for this build's known filesystem set = 0xfff2. Exec-denial adds 0x1. Do not suppress the control on ABI 1 to hide the missing rights.

## Scope and completed repairs

Revision 5 → 6 changes 16 files, confined to Landlock masks/typing/constants, their tests and stub operations, two ledger rows, and related documentation. `inventory.go` changes comments only. The derived writable roots remain PrivateBase plus WritePaths; no root broadened. Other revision-5 surfaces are not reopened by this review.

The prior three reviewer contract tests all PASS now (exit 0, 0.377s): `TestReviewerTruncateUAPIIdentity`, `TestReviewerWriteConfinementMask`, `TestReviewerRegularFileTruncationGrant`. Linux constants now use x/sys/unix; real TRUNCATE is retained on file grants. These tests cover the previous ABI-4 findings, not the newly identified ABI-1 gap.

## Exact hosted evidence

[Gate run 35690194789](https://github.com/relux-works/curator/actions/runs/35690194789) reports success. GitHub git-commit API resolves head `c7cb213618bbecb77b2fad39e1576b9e6060bb4a` to tree `e33bfbf7aa455dbbd9fdbd92f5233eae69d4775f`, exactly the assigned candidate.

Downloaded and parsed `test-evidence-ubuntu-latest` (artifact 10679065490) and `race-evidence-ubuntu-latest` (10679346058). The attached extracted JSON preserves names, actions and elapsed times. Aggregate and served files duplicate the same events; these are not counted as separate runs.

| Worker-boundary row | Ubuntu Test | Ubuntu Race |
|---|---|---|
| TestLinuxLandlockConfinementMatchesProbe | PASS 0.08s | PASS 6.41s |
| TestLinuxWriteConfinementTruncateMatchesProbe | PASS 0.08s | PASS 6.98s |
| TestLinuxWriteConfinementDirectoryMutationMatchesProbe | PASS 0.07s | PASS 7.22s |

The new row coverage is 2/2 on each hosted Linux lane. These are accepted hosted executions, not independent local Linux reruns. The mask/object-typing rows also pass there. Two non-JSON module-download lines in Race artifacts were explicitly preserved in the extraction metadata; no test event parse error was silently treated as absence. The newer hosted kernel does not establish ABI-1 correctness.

## Independent local checks and mutation bounds

Host Darwin 24.6.0 x86_64, Go 1.26.0, shell zsh. No full landing suite replay.

- `go build -o ../curator ./cmd/curator`: exit 0, actual manager binary built.
- `go test -count=1 -timeout=3m ./internal/scriptworker -run '^Test(LandlockHandledMaskFollowsABI|LandlockRuleRightsFollowObjectType)$'`: exit 0, 0.510s, exact candidate baseline.
- Prior reviewer contracts: 3/3 PASS, above.
- New ABI-1 contract: 9/9 missing-right assertions FAIL, above.
- Six independent narrowing mutants each omit one handled right: TRUNCATE, REMOVE_FILE, REMOVE_DIR, MAKE_DIR, MAKE_REG, MAKE_FIFO. Each reruns `TestLandlockHandledMaskFollowsABI` with `-count=1 -timeout=90s`; 6/6 killed (exit 1). These are pure mask-test kills; **0/6 Linux worker-boundary mutant runs** were executed locally. Raw logs attached. Original bytes restored afterwards.

Producer N10–N16 worker-boundary kills remain predictions, not executed evidence. In particular, N14's predicted outside-rename success when MAKE_REG alone is omitted is not established: REMOVE_FILE still gates that rename. Add an independently isolating creation attempt (for example O_RDONLY|O_CREAT on an absent outside file) and execute the Linux narrowing mutants to discharge the requested production-boundary proof. This is a coverage bound, not a claimed observed survivor.

R-e streaming and newer-UAPI bounds remain carried forward. No fresh Windows/ARM64 enforcement claim, exhaustive filesystem guarantee, or acceptance of unrelated obligations is made here.

## Lifecycle and anomaly record

Verdict: **changes_requested**, route `to-dev` for ordinary implementation rework. No human decision or external blocker. No accept_cr, commit_ack or done transition. `task-board spawn goal` reports this run is not goal-bound; no directives were present.

No logbook executable is available; campaign rules prohibit LOGBOOK.md edits. This verdict and the task note preserve the ABI-assumption regression. Attachments are persisted before the status transition.
