# TASK-260910-2t0iun results — rework 1 (answers review rev3 F1)

## Fix
- install.sh gh branch now passes exactly one signer pin:
  - gh supports --signer-workflow: `gh attestation verify <checksums.txt> --repo relux-works/curator --signer-workflow relux-works/curator/.github/workflows/release.yml --cert-oidc-issuer https://token.actions.githubusercontent.com`
  - older gh: `--repo … --cert-identity-regex '^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v' --cert-oidc-issuer …`
  - Value format confirmed from host `gh attestation verify --help` (gh 2.88.1): `[host/]<owner>/<repo>/<path>/<to>/<workflow>`.
- Fake gh (internal/install/installer_script_test.go) now models real gh: rejects >1 of [cert-identity cert-identity-regex signer-repo signer-workflow] with gh's exact message and exit 1; logs normalised argv; 7 rows assert the exact argv line (`expectedGHCall`), replacing substring checks.

## Evidence (zsh, real exit codes)
- `go test ./internal/install -run InstallScript -count=1` → exit 0, 12/12 rows, 8/8 acceptance rows.
- Mutant F1 (re-add `--cert-identity-regex` beside `--signer-workflow`) → exit 1; FAIL valid_attestation_installs, tampered_archive_refused, missing_archive_checksum_refused, duplicate_archive_checksum_refused, unattested_checksums_refused, failed_attestation_does_not_downgrade. Restored → exit 0.
- Real gh 2.88.1 on a dummy file: new shape → exit 1 at `HTTP 404 … /attestations/sha256:…` (flags parsed, reached API lookup); old combined shape → exit 1 at flag parse: "if any flags in the group [...] are set none of the others can be; [cert-identity-regex signer-workflow] were all set".
- `go vet ./internal/install` exit 0; gofmt clean; `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` exit 0.
- Earlier mutants (signature skipped, checksum skipped, identity unanchored) unchanged from rev3; not rerun this round.
- Hosted gate: not run locally; runner runs it after handoff.

## Files
install.sh, internal/install/installer_script_test.go, README.md, SECURITY.md, .github/ci/platform-cases.tsv, .github/ci/skip-classes.tsv. No CHANGELOG/LOGBOOK edits.

## CHANGELOG entry (for release prep)
- Security: `install.sh` now verifies the release before installing — GitHub artifact attestation of `checksums.txt` (`gh attestation verify`, pinned to relux-works/curator's release.yml signer workflow) or, without gh, the keyless cosign signature of `checksums.txt` pinned to the release workflow identity and GitHub Actions OIDC issuer — then the archive's exact sha256 entry. Any failure, or no verifier available, refuses to install. `CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1` is the only, loudly warned bypass.

## Revision 5 — re-apply on f30c2b34 (SECURITY.md merged with 3i6vod)
Applied `git diff 3f60f7f0 refs/campaign/234vmx-rev4-20260929 -- . ':!.task-board'` with `git apply --3way` on trunk f30c2b34 (= origin/main). Only SECURITY.md conflicted (add/add). Merged: trunk's `## Installed command execution` kept verbatim and first, followed by rev4's `## Release installer verification`. No duplicate headings (only `# Security` from trunk). README.md applied cleanly, so both trunk's and rev4's changes are present.

Verification (zsh):
- `git diff --name-only origin/main` + untracked: .github/ci/platform-cases.tsv, .github/ci/skip-classes.tsv, README.md, SECURITY.md, install.sh, internal/install/installer_script_test.go (6 rev4 paths; the test is a new file).
- The +/- lines of install.sh, platform-cases.tsv and skip-classes.tsv are identical to rev4. installer_script_test.go is byte-identical to the rev4 blob (`cmp` rc=0).
- `git show origin/main:SECURITY.md | /usr/bin/grep -vxFf SECURITY.md` printed nothing (0 lines). Python set check: no trunk line is missing. Note: the harness's `grep` wrapper printed 4 spurious blank lines; the system grep printed none.
- Every non-heading line that rev4's SECURITY.md added is present (grep printed nothing, rc=1).
- `go test ./internal/install -run InstallScript -count=1` → ok (21.2s), rc=0.
