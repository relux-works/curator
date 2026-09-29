# TASK-260910-2t0iun — orchestrator decision (THE ONLY CURRENT INSTRUCTION, with 2t0iun-brief.md)

Option 1. The release already signs checksums.txt with keyless cosign and attests it with GitHub Actions (.goreleaser.yml, release.yml);
reuse that trust anchor — no static key, no new secret. The earlier "public key pinned" wording is superseded: what is pinned is the signer
IDENTITY and ISSUER.
install.sh, before installing:
1. If `gh` is available: `gh attestation verify <archive or checksums.txt> --repo relux-works/curator` (pin the signer workflow if the gh
   version supports --signer-workflow).
2. Else if `cosign` is available: `cosign verify-blob checksums.txt --signature/--certificate (or --bundle) from the release
   --certificate-identity-regexp '^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v'
   --certificate-oidc-issuer https://token.actions.githubusercontent.com`.
3. Then verify the downloaded archive's sha256 against the verified checksums.txt.
4. If neither gh nor cosign is available, or any verification fails: REFUSE to install with a clear message (what failed; how to install
   gh or cosign). The only bypass is an explicit, loudly warned environment opt-out `CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1` (document it in
   README with the warning); never a silent downgrade.
Rows with a local fixture release (no network): valid → installs; tampered archive → refused; bad signature / wrong identity → refused;
no tool → refused; opt-out → installs with the warning. Mutants: signature check skipped; checksum check skipped — killed (real exit codes).
Update docs (README install section, SECURITY.md). Set status development; update results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
