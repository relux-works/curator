## 2026-09-05
### 1400 — Git snapshots now byte-exact via object-database extraction
- ROOT CAUSE: `gitops.Archive` used `git archive`, which applies `core.autocrlf`, `text`/`eol`, and `export-subst`; snapshot content hashes depended on the acquiring machine's git config and the repo's `.gitattributes` (review M3, environments §1.2).
- FIX: `gitops.Extract` = `rev-parse <commit>^{tree}` + `ls-tree -r -z --full-tree` + one `cat-file --batch`; raw blob bytes, no filters. Callers `internal/snapshot/snapshot.go:48`, `internal/closure/closure.go:413`. Commits f855a34c, 5beced46 on `feat/byte-exact-acquisition` (signed, not pushed).
- FINDING: the old tar path never detected duplicate platform paths (opened with `O_TRUNC`, silent overwrite). New path refuses via `Lstat` + `O_EXCL` at write time, exact on the acquiring filesystem.
- FINDING: copying the spec fixture `.gitattributes` (`* text=auto`) into `testdata/` under its real name normalizes its siblings in curator's index even with root `**/testdata/** -text` (nested attributes win). Stored as `gitattributes.fixture`; tests map it back.
- FINDING: on git 2.50.1 `git archive` under `core.autocrlf=true` + `* text=auto` rewrites `lf.txt` to CRLF and expands `$Format:%H$`; that is the live negative control for `TestExtractIgnoresWorkingTreeConversion`.
- DECISION: interop conformance test skips with a reason on a root lacking `vectors/snapshot-acquisition.json` (pinned rc.9 `SPEC_PIN 0ed5c691`), passes on curator-spec `ec695ba`.
- STATUS: all gates exit 0 (build, vet, gofmt, full `go test` split cmd/curator + 57 packages, focused `-race`). Not verified: Windows/Git-for-Windows behavior. Entry kept on the board because the brief forbids writing LOGBOOK.md.
