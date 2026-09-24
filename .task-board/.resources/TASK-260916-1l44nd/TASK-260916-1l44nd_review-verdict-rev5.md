# Revision 5 review — changes requested

Candidate tree: `f86bf9e434fa2b6aa07f044829f953ad43e5a0a5`; base `07453544e3defe726f5951af9023a0343c4402b7`. Reviewed from an exact `git archive` extraction under the assigned worktree. No candidate source changed. All mutation/extra test work was confined to that disposable extraction.

## Findings

### F1 — high: wrong UAPI bit leaves truncation unrestricted

`internal/scriptworker/landlock.go:30` defines `landlockAccessFSTruncate = 1 << 12`. Linux defines TRUNCATE as `1 << 14` (0x4000); bit 12 (0x1000) is MAKE_SYM. The actual ABI-4 write mask is therefore WRITE_FILE | MAKE_SYM, and real truncation is not handled at all. An enforced script can call truncate(path,0), or open(path,O_RDONLY|O_TRUNC), on an existing writable file outside every derived grant while evidence claims filesystem-write-confinement applied. These operations do not require WRITE_FILE, so the current os.WriteFile denial row cannot catch the bypass.

This also explains the earlier EINVAL on non-directory grants: the purported truncate right was a directory-only symlink-creation right. `landlockRuleRights` at lines 79–83 then compensates by stripping the mislabeled bit from all non-directories. Real TRUNCATE is valid on regular files; correcting the constant alone would leave that typing rule incorrectly denying explicitly granted file overwrites. Correct both the UAPI identity and object typing, including the special handling of the null device, and replace tests that derive expected masks from the same wrong constants.

Evidence: installed x/sys/unix constants and Linux v6.8 UAPI agree on TRUNCATE=0x4000 and MAKE_SYM=0x1000. Attached reviewer tests fail on the wrong constant, the missing real truncate handled bit, and the intended regular-file truncation bit being stripped. Primary reference: https://raw.githubusercontent.com/torvalds/linux/v6.8/include/uapi/linux/landlock.h (definitions and file-right documentation); operation semantics: https://docs.kernel.org/userspace-api/landlock.html#truncating-files .

Required production reproduction on Ubuntu (including race): create an existing outside file with content, launch through the real worker with grants excluding it, attempt truncate and O_RDONLY|O_TRUNC, and require denial plus unchanged content. Grant a separate existing individual file and require its overwrite/truncation to succeed. Check applied evidence in the same invocation. Add a narrowing mutant that omits the real truncate bit.

### F2 — high: write-confinement omits directory mutation rights

`internal/scriptworker/landlock.go:46–50` also omits REMOVE_FILE, REMOVE_DIR, MAKE_DIR, MAKE_REG, MAKE_FIFO and other directory mutation rights. Landlock permits unhandled actions (REFER is the special exception). A script can unlink user-owned files, remove empty directories, create directories/FIFOs, and rename regular files within the same directory outside its derived grants while reporting the control applied. This contradicts the applied-control requirement and `docs/script-interpreters.md:50–53` (“everything else stays denied”). It is an incomplete application of the named host control, not a request for a deferred cross-platform guarantee. Symlink creation on ABI >= 3 is incidentally handled by F1's wrong bit and is NOT claimed as an observed bypass here.

Source-level reproduction: the attached reviewer test calls the production mask builder for ABI 4 and fails for five independently sourced mutation rights: mask 0x1002 omits 0x20, 0x10, 0x80, 0x100, 0x400. Existing TestLinuxLandlockConfinementMatchesProbe only attempts file-content writes and exec. Kernel reference: the same v6.8 UAPI header explicitly states unhandled access rights are allowed and defines these mutation flags.

Required repair: handle ABI-supported filesystem mutation rights and grant directory-only rights only over derived writable directories. At the worker boundary test unlink, rmdir, mkdir, FIFO creation and same-directory rename outside the grant (deny and preserve filesystem state) and equivalent operations inside a granted directory (allow), with matching evidence. Attack individual omitted rights with narrowing mutants. Do not broaden writable roots.

Both findings are established by production source, executable source-contract checks and primary kernel semantics. The Linux runtime exploit scenarios above were NOT executed locally; they are explicit rework regressions to add, not claimed local kernel observations.

## Independent verification

Shell: zsh; local host Darwin. No full landing suite replay.

- GitHub run 35685793652 is success. Its head `7f74ba107dafa57cf46417d0923c34802d3c01e0` resolves via GitHub git-commit API to the exact candidate tree above.
- Downloaded and parsed `test-evidence-ubuntu-latest`, `race-evidence-ubuntu-latest`, and `test-evidence-windows-latest`. Attached extracted rows include the Landlock probe/confinement/missing-path rows, descendant termination, all named evidence mutation rows, invocation preflight, installed CLI launch, and Windows Job Object confirmation. Linux confinement/probe rows PASS in Test and Race; Windows Job Object row PASS on Windows. Platform-specific skips are present on other lanes. These are accepted hosted results, not local Linux/Windows reruns. Rose-air job was skipped.
- Local exact-candidate portable subset: `go test -count=1 ./internal/scriptworker -run 'Test(ScriptEvidence|FixedUnavailable|PreflightRefuses|RunShim|LandlockHandled|LandlockRule)'` exit 0, 5.277s.
- Built actual manager: `go build -o ../curator ./cmd/curator`, exit 0.
- Narrowing mutant: timing validator admits `at-install` while retaining its other checks. `go test -count=1 ./internal/scriptworker -run '^TestScriptEvidenceMutationsRefuseBeforePermit$/cached-probe-result$'` exit 1; named row fails because it reaches the identity backstop instead of the required probed_at refusal. 1/1 reviewer narrowing mutants killed. Restored original bytes; targeted row then exit 0.
- Three added reviewer contract tests exit 1 as expected, exposing F1/F2. They are helper-level source proofs, not production-entry runtime coverage. Attached source makes them reproducible.
- A first narrow install/CLI run passed (122.752s), but overlapped the disposable timing mutant, so it is NOT counted as exact-candidate baseline evidence. Clean rerun on restored candidate: `go test -count=1 -timeout=5m ./internal/install -run '^Test(EnforcedInstallAndLaunchAtCLIEntry|EnforcedScriptCommandIsRefusedAtInstall|ScriptOptInCasesAtInstallEntry)$'` exit 0; timing is included in the attached logs.
- Docker info returned Linux, but image listing stalled and was terminated. No local Linux execution is claimed. No full matrix or exhaustive review completion is claimed after finding these blockers.

## Coverage and bounds

Hosted artifacts confirm the named evidence/preflight paths passing; the producer maps 14/14 evidence and 5/5 preflight vectors. Those counts do not prove filesystem mutation completeness: current confinement row checks file-content writes and exec, and has 0 tests for the five outside-directory mutation actions above. Reviewer attack ratio: 1/1 narrowing mutant killed; 3/3 additional source-contract tests fail. Linux enforcement mutants in producer results remain predictions where no executed mutant evidence exists; ordinary green hosted runs do not turn those predictions into kills.

Parent validates the evidence before permit (`client.go:209–220`); result exposes the record and DerivationReport through the operator-selected diagnostics path. Thread locking and no_new_privs are on the worker spawn path. R-e streaming remains explicitly deferred to R5. No acceptance of remaining unexamined obligations is implied.

## Lifecycle / logbook

Verdict: changes_requested; route to `to-dev`. Ordinary implementation rework; no human decision or external blocker. No accept_cr, commit_ack, or done transition. Run goal query reports not goal-bound. No logbook executable is available and campaign rules prohibit LOGBOOK.md edits; this task-scoped verdict and board note preserve the findings.

## Correction during review

An initial draft misidentified bit 12 as real truncation, mirroring the implementation. Independent comparison with the kernel UAPI exposed that mistake before the verdict transition. This updated artifact supersedes that draft: the regular-file regression is latent after fixing the constant; the current defect is unrestricted real truncation. The corrected attached test uses independently sourced UAPI values.
