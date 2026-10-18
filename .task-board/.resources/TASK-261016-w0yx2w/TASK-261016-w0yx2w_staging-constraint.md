# SNAP-PUBLISH: resumed admission candidate and staging authority constraint

Task: TASK-261016-w0yx2w. State: blocked, not a developer review handoff.
Checkpoint: fcd785ba (full identity is in the attached source-identity resource).
Normative curator-spec main read directly: 9dfc557d247243e0e8ba60fb877afe6d1facfd06,
managed-home-snapshots sections 2/4/5, environments section 5 collision rules,
and CIP-0014 DestinationBoundary / Idempotence rows.

## Implemented independently of the constraint

- `OpenDestination` / `inspectDestination` now refuse existing empty targets
  with `environment_materialize_destination_invalid: empty-target`.
- The immediate retained publication parent is caller-owned mode 0700 on both
  initial admission and `Revalidate`; mode 0755 refuses without writes.
- Absent destinations and intact exact prior snapshots remain admitted;
  `MatchPrior` still verifies the actual tree and refuses different generations.
- All original test function names and committed reviewer tests are retained.
  Replacement tests now start from intact snapshots instead of admitted empty
  targets. The identical-prior replacement test distinguishes inode identity
  from equal manifest/content. Permission/protected-root fixtures avoid masking
  one refusal with a second independently rejecting condition.
- Candidate remains uncommitted. No publication or deletion implementation was
  added. Materialize CLI integration belongs to SNAP-CLI and is absent here.

Resume edits stay under `internal/envsnapshot`, except
`UNRESOLVED_QUESTIONS.md`, required by repository instructions to persist the
new owner decision. The carried admission candidate is preserved. Scratch
probes and logs stay in task-scoped ignored scratch; no control-root file was
edited directly. Board persistence uses the CLI only.

## Concrete constraint and failed assumptions

The revised brief says a substituted or linked staging entry is never followed
or removed. Retaining the parent and original object, opening without following
links, and successfully checking device/inode immediately before deletion do
not make `unlinkat(parent, name, flags)` conditional on that object. A same-user
process can substitute the name after the final check. The kernel deletes the
replacement entry. For a symlink the referent survives, but the substituted
link itself is removed, violating the brief's stronger requirement.

The deterministic schedule uses real syscalls, not mocked filesystem results:

1. Create invocation-owned stage entry and retain its handle and parent.
2. `fstatat(..., AT_SYMLINK_NOFOLLOW)` matches the retained `fstat` identity.
3. Move the original away and create an unrelated file, empty directory, or
   symlink at the same name, under the same unprivileged user's authority.
4. `unlinkat` succeeds and removes the replacement. Original moved entry and
   retained handle survive. The link referent sentinel is unchanged.

`TestRequestedCleanupInvariant/{file,directory,link}` asserts the requested
invariant against this proposed check-then-unlink algorithm. All three fail
with the intended VIOLATION message; the standalone command exits **1**.
This is expected-red evidence against that algorithm, not a product test pass,
not a planted mutation of implemented cleanup, and not proof that publication
has been implemented.

No-replace rename resolves hostile **destination-name** competition: the actual
macOS `RENAME_EXCL` call succeeds for absence and returns EEXIST for a competing
empty destination, preserving its inode and contents. It does not bind its
**source name** to a retained staging inode. A source substitution after the
last check is moved to the absent target by the same real syscall. Verifying
afterward can prevent success emission but cannot claim that no unrelated
staging entry was moved. No post-failure destructive rollback was attempted.

Handle-only deletion with an empty name or `.` refuses (ENOENT / EINVAL on the
exercised macOS runtime). The ordinary unlink interface offers a parent fd and
name, not an expected child inode. Repeating the last check just moves the
race; private 0700 mode excludes different users, not another process with the
same user's authority. Advisory coordination is not enforced namespace
exclusivity. No escalating flags, stubs, or same-user mutation exceptions were
introduced as product implementation.

Primary API references:

- [Linux unlinkat](https://man7.org/linux/man-pages/man2/unlinkat.2.html):
  parent-relative name deletion; no expected-child identity argument.
- [Linux renameat2](https://man7.org/linux/man-pages/man2/rename.2.html):
  no-replace applies to the destination name.
- [Apple unlink](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/unlink.2.html)
  and [Apple syscall declarations](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/syscalls.master).

These references support API semantics. Runtime evidence in this run is
**macOS arm64 only**, Go 1.25.5, OS 27.0. The supplied Linux probe backend was
not executed or qualified. There is no Windows runtime qualification.

## Owner decision needed

Specify who can mutate the staging namespace while this invocation owns it.

| Option | Meaning | Tradeoff |
|---|---|---|
| Enforced staging isolation | Stage inside a namespace which competing same-user processes cannot mutate; design its principal/ownership and publication handoff explicitly | Retains the strong invariant, but needs an owning architecture beyond ordinary caller-owned 0700 directories |
| Explicit staging threat bound | Exclude adversarial same-user writes inside private invocation staging, while retaining no-replace protection against hostile target-name creation | Fits ordinary retained-handle Unix operations, but narrows the current substituted-staging requirement and needs explicit owner/spec approval |
| Keep the current unconditional contract and backend refusal | Admit read-only observations but do not expose publication on an unqualified backend | Preserves the stated safety contract; materialization remains unavailable |

Recommendation: settle this authority boundary explicitly before publication
implementation. If hostile same-user staging mutation remains in contract,
choose enforced isolation; otherwise record the narrower staging bound in both
normative text and brief. Fail-closed retention/no automatic cleanup alone can
avoid destructive unlink, but it does not solve source-stage substitution and
must not be presented as satisfying the whole leaf. This is a new staging
decision; the absent-only destination decision is already implemented.

## Validation ledger

All commands below were executed directly, without tee or a status-hiding
pipeline. All logs are attached under task-prefixed names.

| Exact command | Exit | Evidence / qualification |
|---|---:|---|
| `go test -count=1 -run '^TestDestination' -v ./internal/envsnapshot` | 1 | admission-01.log: private-parent rule exposed the old containing-target fixture's nonprivate immediate parent; fixture repaired |
| `go test -count=1 -race ./internal/envsnapshot` | 0 | package-race-04.log: latest source/test identity, all package tests including committed reviewer regression |
| `go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot` | 1 | destination-controls-01.log: mode ancestor mutant survived because the same fixture also hit private-parent refusal |
| Same destination-control command after fixture repair | 0 | destination-controls-03.log: latest identity, all 12 children fail by named tests with exit 1 |
| `golangci-lint run ./internal/envsnapshot/...` | 0 | lint-03.log: latest source/test identity |
| `go build -o .temp/TASK-261016-w0yx2w/curator-check ./cmd/curator` | 0 | build-01.log: unchanged product source since this build; later changes affect tests and documentation only |
| `git diff --check` | 0 | diff-check-01.log; attached candidate patch also includes untracked admission files |
| `go test -count=1 -run '^TestProbe' -v ./.temp/TASK-261016-w0yx2w/staging-probe` | 0 | staging-probe-01.log: four syscall-witness test functions, with paired no-replace destination control |
| `go test -count=1 -run '^TestRequestedCleanupInvariant$' -v ./.temp/TASK-261016-w0yx2w/staging-probe` | 1 | staging-invariant-red-01.log: expected failure in all three named substitution cases |
| `sh scripts/remote-gate.sh` | unrun | Stopped before publication implementation under Stop-The-Line; no full-gate or Linux runtime pass is claimed |

Intermediate package-race-02/03, lint-01/02 and destination-controls-02 also
exited 0; latest evidence above supersedes their older test identities.
No evidence from the prior producer run is used as this run's full gate.

## Narrowing evidence

| Mutant | Gate narrowed to | Named failing test (child exit 1) | Bound |
|---|---|---|---|
| M2-linked-ancestor | Follow only the directory component named target; retain other no-follow checks | `TestDestinationBoundary/linked-ancestor` | Library admission only |
| M6-saved-manifest-bytes | Trust saved metadata only when actual-file identity fails | `TestDestinationPriorIdentity/bytes` | Library prior-tree bytes only |
| owner-root-parent | Permit a root-owned caller parent at the owner gate | `TestDestinationOwnershipAndReadFailure` | Diagnostic ordering only after private-parent tightening: mode still refuses `/`; no end-to-end unsafe admission claim |
| mode-group-writable-ancestor | Permit mode 0770 at the ancestor gate, with an independently private immediate parent | `TestDestinationBoundary/shared-ancestor` | Actual ancestor permission boundary |
| protected-equal-identity | Ignore equality only at the protected root's own inode | `TestDestinationProtectedIdentity` | Intact snapshot avoids empty-target masking |
| ancestor-directory-replacement | Ignore identity change only for component parent | `TestDestinationReplacementRaces/ancestor-directory` | Revalidation only |
| target-identical-replacement | Permit a different target inode when mode is private 0700 | `TestDestinationReplacementRaces/target-identical` | Equal bytes do not make inode replacement valid |
| empty-target | Admit only the otherwise-valid empty-directory class | `TestDestinationReadOnlyAdmission` | Current absent-only owner decision |
| private-parent | Additionally permit immediate parent mode 0755 | `TestDestinationPrivateParent` | Initial admission and retained revalidation |
| unrelated-one-entry | Admit unrelated directories with exactly one entry | `TestDestinationBoundary/unrelated` | Other unrelated targets remain gated |
| acl-one-permit-kind | Permit the exercised macOS ACE kind and permission mask | `TestDestinationKernelACL` | Exercised macOS ACL ABI only |
| prior-acl-one-member | Skip prior member ACL check only for AGENTS.md | `TestDestinationPriorKernelACL` | Exercised macOS member only |
| no-replace-publication / staging-only-cleanup | Unimplemented | No named product test; **not killed / unrun** | Publication and cleanup protection are unproven |

The initial mode-group-writable-ancestor mutant is a recorded survivor at the
old test identity, bounded to a fixture whose private-parent gate masked the
weakened ancestor gate. The corrected fixture kills it in controls-03. Owner
diagnosis evidence is not upgraded to end-to-end ownership admission evidence.
No source-text checker is used as behavioral evidence.

## Coverage and remaining obligations

The brief has a paired-test table, but no catalog-keyed surface table. This is
a brief gap, not a coverage waiver. All four paired rows are mapped explicitly:

| Paired row | Production library call sites / tests | Narrowing evidence | Missing coverage |
|---|---|---|---|
| Absent admission vs links/owner/ACL/protected aliases, empty refusal applied | `OpenDestination`, `destinationDirectory`, `excludeProtected`; `TestDestinationReadOnlyAdmission`, `TestDestinationBoundary`, `TestDestinationPrivateParent`, ownership/kernel ACL/protected identity tests | M2, private-parent, empty-target, mode, protected, ACL | Full supported-platform runtime qualification; complete protected inventory supplied by future coordinator |
| Retained parent publication vs replacement/competition | `Destination.Revalidate`; `TestDestinationReplacementRaces` | ancestor-directory / target-identical replacement | Publication not implemented; probe is not a product call site |
| Verified identical prior vs actual-tree/generation changes | `Destination.MatchPrior`, `inspectDestination`, `verifyDirectory`; `TestDestinationPriorIdentity`, `TestDestinationPriorMissingMetadata`, prior ACL test | M6, prior-acl-one-member | Materialize integration separate |
| Staging failure / cleanup substitution / interruption | No product call site; isolated syscall witnesses only | None; required cleanup mutant unrun | Entire product staging/failure/cleanup/interruption row |

**0 of 5 AC rows driven through production materialize** (SNAP-CLI is explicitly
separate). **3 of 5 AC rows exercised at the existing production library
boundary on macOS**: admission, protected identity, prior verification. This is
not five-row leaf completion: publication and mode-preserving staging/cleanup
remain absent. No publication, cleanup, interruption, Windows support, or Linux
runtime success is claimed. Unqualified platforms continue explicit refusal,
as permitted by the revised brief. Capture authorization, inventory discovery
and CLI integration remain their owning leaves' responsibilities.

Full gate, publication properties, required publication/cleanup mutants,
all-platform qualification and full leaf AC/DoD remain unchecked. Developer
handoff is deliberately not invoked. Existing committed reviewer tests were
unchanged and passed in the current package race suite.

## Logbook

Resumed after the absent-only owner decision, updated the admission branch and
tests, and established a separate final-component staging authority gap with
real macOS syscalls. Ordinary implementation retries cannot supply an atomic
expected-child parameter missing from the delete/rename interface. The next
owner decision must settle staging authority without relitigating empty-target
publication. Evidence and the candidate patch are persisted before lifecycle
end, with no pending background process.
