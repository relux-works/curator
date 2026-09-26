# BUG-260916-2f3xbf — results (developer)

## Outcome

Windows junction spelling of pnpm directory links declared in the
writable-store registry and the materialized-link validators; ledger deferral
for the two real-pnpm cases removed. Ready for review; Windows/macOS/Linux
execution evidence must come from the hosted lanes (this host SIGKILLs every
freshly built Go binary — pre-existing environment condition, see §8).

## 1. Root cause (evidence chain)

1. Hosted gate run 35098955988, artifact `test-evidence-windows-latest`
   (`go-test.json`): both install cases fail with
   `closure_input_undeclared: pnpm writable store registry contains an
   undeclared member`, entries `c8805e97311ac3a64fd7c507d26b47dc` and
   `4990f83e93c63b5b58405c5ac8c8ca6a` (32 lowercase hex). No other
   `internal/pnpmsource` failure on that lane.
2. pnpm 10.33.0 bundle (`projectRegistry.js`): `registerProject` links the
   project with `symlink-dir`, whose `symlinkType` is `"junction"` on win32
   and `"dir"` elsewhere; the link name is `createShortHash(projectDir)` =
   first 32 chars of sha256 hex. The observed entries match that grammar.
3. Go semantics (toolchain 1.26.0 source): on Windows `os.Lstat` reports a
   junction as `ModeIrregular` (only `IO_REPARSE_TAG_SYMLINK` maps to
   `ModeSymlink`), and `filepath.EvalSymlinks`/`walkSymlinks` does not follow
   junctions; `os.Readlink` resolves both symlinks and junctions
   (`IO_REPARSE_TAG_MOUNT_POINT`). Hence the old symlink-only check rejected
   the member, and the old resolve-then-compare could never have succeeded
   past it.
4. The same symlink-only assumption guards two validators behind the registry
   check that both install cases traverse on Windows (importer dep links in
   `validateDirectNodeModules`, snapshot dep links in
   `validateSnapshotInstance` — pnpm links those directories through the same
   `symlink-dir`). All three call sites fixed with one shared helper pair.
5. Ruled out as further Windows gaps (by code + fixture inspection):
   `inventoryPackage` executable bit (fixtures are all tar mode 0644;
   Windows synthesizes 0666 → non-executable both sides),
   `scanAmbientSideEffects` (name/content checks only),
   `copyContainedNode` (junctions resolve via `os.Stat`-follows; fixtures
   are link-acyclic — proven by Unix green), lockfile byte check
   (`--frozen-lockfile` never rewrites), `.bin`/side-effects presence
   (identical set on Unix, which passes).

## 2. Fix (declaration, not refusal)

The member is a legitimate pnpm store artifact, so the registry now declares
it — closed and platform-conditional, mirroring
`internal/npmsource/linkentry_*.go`:

- `internal/pnpmsource/linkentry_windows.go` (new, `//go:build windows`):
  `admittedLink` = symlink OR (`ModeIrregular` + reparse-point-and-directory
  attributes via `golang.org/x/sys/windows`, already a dependency);
  `normalizeTreeLink` = `os.Readlink` for junctions (identity otherwise).
  Any other irregular node stays refused; the target comparison downstream
  still decides admissibility.
- `internal/pnpmsource/linkentry_unix.go` (new, `//go:build unix`):
  symlink-only admission, identity normalize — Unix behavior byte-identical
  (same conditions, same codes, same messages at all three call sites).
- `internal/pnpmsource/materialize.go`: the three link checks use the
  helpers; no other production change.

## 3. Deferral removal

- `internal/pnpmsource/conformance_test.go`: deleted
  `skipOnWindowsForStoreRegistryGap` and both call sites.
- `.github/ci/platform-cases.tsv`: both rows now
  `linux,darwin,windows - -`; comment block rewritten (no deferral language).
- `.github/ci/skip-classes.tsv`: bug-specific `stage-deferred` row removed;
  section header kept with a placeholder comment.
- `.github/ci/gate-selftest.sh`: deferral pins replaced with no-deferral
  assertions (no bug text in Go suite; no bug data-row in class table; both
  ledger rows require-all/tolerate-none) plus behavioural any-skip-on-windows
  and same-skip-on-linux fatal cases; rust wrong-class probe re-pointed from
  the removed reason to a live `root-unset` reason.

## 4. Tests

- Retained: `TestWritableStoreOverlayAllowsOnlyExactProjectRegistration`
  file-member negative subtest; extended to pin the kind-gate diagnostic.
- Added (portable): `plain directory member refused` subtest — a plain
  directory (even with a pnpm-shaped 32-hex name) is refused as an
  undeclared member.
- Added (Windows-only, `linkentry_windows_test.go`, `mklink /J` idiom per
  repo precedent): junction registry registration admitted end-to-end
  through `reconcileWritableStoreOverlay`; misdirected junction refused;
  `admittedLink`/`normalizeTreeLink` shape checks; junction dep link through
  `validateDirectNodeModules`.

## 5. Narrowing-mutant analysis

- `admittedLink` → always true: rogue file/dir members still fail closed at
  the target check, but with `targets an undeclared project` instead of
  `contains an undeclared member` — the strengthened negative tests assert
  the kind-gate diagnostic, so the mutant fails. (Reviewed statically; the
  mutant run itself needs a host that can execute test binaries.)
- Deleting the registry loop: `err` becomes nil → `assertCode` fails.
  Deleting a ledger row / re-adding tolerance: the self-test row-shape
  assertions fail (proven by the self-test run in §6, which caught my own
  placeholder comment mentioning the bug ID).

## 6. Verification (real exit codes, this host)

| check | exit |
|---|---|
| `gofmt -l internal/pnpmsource/` (empty) | 0 |
| `go vet ./internal/pnpmsource/` (darwin) | 0 |
| `GOOS=windows go vet ./internal/pnpmsource/` | 0 |
| `GOOS=linux go vet ./internal/pnpmsource/` | 0 |
| `GOOS=windows go test -c -o /dev/null ./internal/pnpmsource/` (Windows test file compiles) | 0 |
| `golangci-lint run ./internal/pnpmsource/...` (0 issues) | 0 |
| `bash .github/ci/ledger-consistency.sh` (241 rows; both pnpm rows `must=linux,darwin,windows skip=-`) | 0 |
| `bash .github/ci/no-broad-suppression.sh` | 0 |
| `bash .github/ci/gate-selftest.sh` (185 cases incl. new no-deferral + skip-fatal rows) | 0 |
| `bash -n .github/ci/gate-selftest.sh` | 0 |

## 7. For the reviewer / hosted lanes

- `windows-latest` must show both install cases `ok` (not deferred, not
  skipped) plus the four new `linkentry_windows_test.go` cases passing;
  `ubuntu`/`macos` lanes must stay green (Unix path unchanged).
- Suggested hosted proof: `go-test.json` pass rows for
  `TestRealPinnedPNPMLockSupersetSnapshotDependencies`,
  `TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization`,
  `TestWritableStoreOverlayAdmitsWindowsJunctionRegistration`,
  `TestWritableStoreOverlayRefusesMisdirectedWindowsJunction`,
  `TestAdmittedLinkCoversWindowsJunctionShape`,
  `TestDirectNodeModulesAdmitWindowsJunctionLinks` on windows-latest.

## 8. Environment blocker (pre-existing)

This host (e11-1) SIGKILLs (signal 9, exit 137, zero output) every freshly
built Go binary at startup: reproduced with the pnpmsource test binary, an
untouched package's test binary (`internal/hashing`), and a hello-world
binary. `go build`/`go vet` succeed; only execution dies. Therefore no `go
test` could run here and the mutant was reviewed statically rather than
executed. No workaround was attempted beyond relocating binaries out of
/tmp (same kill from `$HOME`). Test execution is deferred to the hosted
lanes, which the campaign assigns to the integration path.
