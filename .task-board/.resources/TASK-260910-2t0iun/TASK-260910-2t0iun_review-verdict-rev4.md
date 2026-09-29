# TASK-260910-2t0iun review verdict rev4 — ACCEPTED

Candidate: base 3f60f7f0, tree 3db884ae (6 paths). Reviewer: claude-opus-5-5, zsh with pipefail, darwin host.

## F1 (rev3) is fixed
- install.sh gh branch now passes exactly one signer pin. If `gh attestation verify --help` lists --signer-workflow, it passes `--repo --signer-workflow relux-works/curator/.github/workflows/release.yml --cert-oidc-issuer`. Otherwise it passes `--repo --cert-identity-regex <anchored> --cert-oidc-issuer`.
- Real gh 2.88.1 on this host, with the exact argv against a dummy file: flag parsing accepted it and the command reached the API (`HTTP 404 ... attestations/sha256:...`, exit=1, verification only). There was no "if any flags in the group" error. The help text gives the value format as `[host/]<owner>/<repo>/<path>/<to>/<workflow>`, and the value matches it.
- Control: passing both --signer-workflow and --cert-identity-regex to real gh gives "if any flags in the group [...] were all set", the rev3 failure mode.
- The fake gh (test:409-425) counts signer pins and rejects more than one with gh's exact message. Rows assert the exact argv for each branch (test:87-218, including the older-gh identity-regex branch).

## Rerun (independent)
`go test ./internal/install -run InstallScript -count=1 -v`: exit=0, 12/12 subtests PASS (valid, tampered, missing/duplicate checksum, unattested, bad cosign sig, wrong identity, no verifier, opt-out warns, cosign fallback, no downgrade after gh failure, older gh).

## Mutants (run on a temp copy, real exit codes)
- both signer flags passed: exit=1, 7 FAIL (killed)
- signature/attestation check skipped: exit=1, 11 FAIL (killed)
- archive checksum check skipped: exit=1, 4 FAIL (killed)
- identity regexp unanchored (leading ^ removed): exit=1, 5 FAIL (killed)

## Still OK from rev3
- Order: gh, then cosign, then refuse. After that the archive sha256 is checked against the verified checksums.txt; the awk exact `$2 ==` match requires exactly one entry. Then extract/install.
- Every failure refuses with a clear message. A failed gh check exits and does not fall through to cosign.
- The opt-out warns loudly and is documented in README and SECURITY.md.
- The regex is anchored with dots escaped, and the issuer is pinned.
- Windows ledger rows are byte-exact with the t.Skip text.
- No CHANGELOG or LOGBOOK changes.

## Residual (non-blocking)
Real gh attestation success is unverified against a real release. The gh/cosign semantics are modelled by the fakes, and flag parsing is proven against real gh.
