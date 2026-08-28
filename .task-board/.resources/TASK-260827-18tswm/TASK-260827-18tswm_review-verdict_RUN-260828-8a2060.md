# TASK-260827-18tswm review verdict — RUN-260828-8a2060

## Verdict

**Changes requested.** Route to `to-dev`, then publish a fresh Change Request
revision and run another reviewer cycle. Revision 1 must not be accepted.

## Blocking findings

### 1. The handed Change Request is stale and does not identify the CI-green tree

`CR-TASK-260827-18tswm-1` revision 1 is recorded as `stale`. Its exact candidate
tree is `87510b97c9498588fc1dfdd51f98ff28fee80d62`, the 17-path patch whose
SHA-256 is `cf15ca8d04dd811c7646125e3fd37329541fac8a1acd20f32d0f00fc9937158a`.
The successful GitHub run instead tested PR #47 head
`c2215f9b929e11a32d75bff1205d296c135ddd7f`. Later task-owned corrections,
including host-GOROOT isolation (`fd5911fb`) and subsequent Windows delivery
work, are not represented by revision 1. A green run over `c2215f9b` therefore
cannot attest the stale candidate revision.

Required correction: publish a fresh Change Request revision from the complete
task-owned delivered tree, with a candidate identity that can be tied directly
to the CI-tested commit. Do not reuse revision 1 evidence for a different tree.

### 2. The required Windows Race lane never ran

The task acceptance criterion and checked checklist item 6 require Test **and
Race** on Ubuntu, macOS, and Windows. GitHub Actions run
[`33130874599`](https://github.com/relux-works/curator/actions/runs/33130874599)
is successful for the jobs it actually ran, but its jobs contain only
`Race (ubuntu-latest)` and `Race (macos-latest)`. There is no
`Race (windows-latest)` job and no `race-evidence-windows-latest` artifact.
The tested workflow explicitly sets the race matrix to
`[ubuntu-latest, macos-latest]` and says Windows is deliberately absent.

Required correction: either add and pass the Windows Race lane with its
evidence artifact, or obtain an explicit human change to the acceptance
criterion. The current green workflow cannot satisfy the criterion as written,
and checklist item 6 must not be represented as proved until one of those paths
is complete.

### 3. The new Swift host-skip classifier has no negative regression

The candidate turns a compound Swift diagnostic into an allowed
`host-capability` skip in both `internal/swiftpmsource/swift_integration_test.go`
and `internal/swiftpmbuild/swift_integration_test.go`. The implementation is
narrow by inspection (it requires both `Invalid manifest` and the exact clang
`posix_spawn` diagnostic), but no test drives the classifier with either token
alone or an adjacent Swift failure and proves that those shapes remain fatal.
The only occurrences are the two classifiers and the skip-class row. This is
positive-path-only evidence for a gate that suppresses failures.

Required correction: add a testable predicate/seam and negative tests proving
that `Invalid manifest` alone, the spawn diagnostic alone, and an unrelated
Swift/permit failure do not classify as host-unavailable; retain a positive
case for the exact conjunction. Exercise both call sites or share one tested
classifier between them.

## Accepted evidence and reviewer reruns

- The GitHub run is tied to PR #47 head `c2215f9b`; Test on Ubuntu/macOS/Windows,
  Race on Ubuntu/macOS, Lint, Naming, Interop, and all three Gate self-tests are
  completed/success.
- Downloaded Test evidence for Ubuntu, macOS, and Windows reports
  `platform-case gate: ok`; searches over `platform-cases.txt` and
  `skips-observed.tsv` found zero `UNCLASSIFIED` or `FATAL` rows.
- Observed classified rows contain pnpm x3 and Yarn Modern x4 on each runner;
  Yarn Classic has one classified host limitation per runner.
- Reviewer reran and passed:
  - `TestCargoHostCapabilityReasonClassifiesOnlyAbsence`
  - both closed Rust-gap negative tests
  - `TestYarnClassicPackageRootRecognizesPinnedInstallLayouts`
  - the two verified-provider CLI cases plus the native-inventory boundary
  - full `internal/crossconformance`
  - full `internal/swiftpmsource` and `internal/swiftpmbuild`
  - `.github/ci/gate-selftest.sh` (81 passed, 0 failed)
  - `git diff --check` on the exact candidate and `gofmt -l cmd internal`

These passes support the accepted portions of the patch but do not repair the
three findings above. No source file was modified, staged, committed, pushed,
or merged by this reviewer.
