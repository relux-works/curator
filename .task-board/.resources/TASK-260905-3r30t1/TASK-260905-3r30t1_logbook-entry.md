## 2026-09-05
### 1756 — Hosted platform-case gate satisfied; Windows has no executable bit
- REGRESSION: PR #58 at 5beced46 failed Test/Race on every OS from `.github/ci/platform-case-gate.sh`: the rework had renamed the ledger-required `internal/gitops TestArchiveRejectsLinks`, and `TestConformanceSnapshotAcquisition` skipped with a reason outside `skip-classes.tsv`.
- FIX: 5abec244 restores `TestArchiveRejectsLinks` via `git mktree -z` with a `120000` entry (no `ln`, runs on all three GOOS); skip text reworded to the existing `root-content` pattern ("publishes no "); ledger rows added for the byte-exact cases (`.github/ci/platform-cases.tsv:82-86`, `:171`).
- ROOT CAUSE: Windows `Test` then failed `TestExtractPreservesExecutableBit`: Go on Windows synthesizes `-rw-rw-rw-` for every regular file, so a 100755 blob can never show an execute bit. Local darwin gates could not catch it.
- FIX: a46abc80 skips the case on windows with the existing `platform-control` reason `Windows does not expose portable executable permission bits` (same text as `internal/artifactpolicy`), ledger row `linux,darwin | windows | platform-control`. Gate probe: skip tolerated on windows, fatal by name on linux.
- FINDING: the platform-case gate's awk JSON reader needs compact `go test -json` framing (`"Action":"skip"`); a hand-built stream with `": "` spacing is silently ignored.
- ANOMALY: `Race (macos-latest)` at a46abc80 failed once in `internal/install TestStrictRegistryPolicyFailsUnknown` ("snapshot timestamp is too far in the future", `internal/registry/snapshot.go:158`, wall-clock vs a `time.Now()` fixture); untouched by this branch, passed at 5abec244; job rerun.
- STATUS: Windows Test green at a46abc80; monitoring the macOS race rerun. Entry kept on the board because the brief forbids writing LOGBOOK.md.

## 2026-09-05
### 1400 — Git snapshots now byte-exact via object-database extraction
- ROOT CAUSE: `gitops.Archive` used `git archive`, which applies `core.autocrlf`, `text`/`eol`, and `export-subst`; snapshot content hashes depended on the acquiring machine's git config and the repo's `.gitattributes` (review M3, environments §1.2).
- FIX: `gitops.Extract` = `rev-parse <commit>^{tree}` + `ls-tree -r -z --full-tree` + one `cat-file --batch`; raw blob bytes, no filters. Callers `internal/snapshot/snapshot.go:48`, `internal/closure/closure.go:413`. Commits f855a34c, 5beced46 on `feat/byte-exact-acquisition` (signed, not pushed).
- FINDING: the old tar path never detected duplicate platform paths (opened with `O_TRUNC`, silent overwrite). New path refuses via `Lstat` + `O_EXCL` at write time, exact on the acquiring filesystem.
- FINDING: copying the spec fixture `.gitattributes` (`* text=auto`) into `testdata/` under its real name normalizes its siblings in curator's index even with root `**/testdata/** -text` (nested attributes win). Stored as `gitattributes.fixture`; tests map it back.
- FINDING: on git 2.50.1 `git archive` under `core.autocrlf=true` + `* text=auto` rewrites `lf.txt` to CRLF and expands `$Format:%H$`; that is the live negative control for `TestExtractIgnoresWorkingTreeConversion`.
- DECISION: interop conformance test skips with a reason on a root lacking `vectors/snapshot-acquisition.json` (pinned rc.9 `SPEC_PIN 0ed5c691`), passes on curator-spec `ec695ba`.
- STATUS: all gates exit 0 (build, vet, gofmt, full `go test` split cmd/curator + 57 packages, focused `-race`). Not verified: Windows/Git-for-Windows behavior. Entry kept on the board because the brief forbids writing LOGBOOK.md.
