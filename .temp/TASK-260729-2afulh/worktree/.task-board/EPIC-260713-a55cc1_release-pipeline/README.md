# EPIC-260713-a55cc1: release-pipeline

## Description
Production distribution for the curator binary. Level 1: GoReleaser on tag
(darwin/linux/windows x amd64/arm64, archives, checksums, Homebrew tap,
Scoop bucket, deb/rpm via nfpm) plus an install script served over GitHub
Pages. Level 2: supply chain trust: cosign keyless signatures, SBOMs,
GitHub build provenance attestations, and macOS Developer ID signing with
notarization wiring (Relux Works, LLC).

## Scope
.goreleaser.yml, the release workflow, relux-works/homebrew-tap and
relux-works/scoop-bucket repositories, repository secrets, install.sh.

## Acceptance Criteria
- git tag vX.Y.Z produces a GitHub Release with archives, checksums, SBOMs, and cosign signatures
- brew install relux-works/tap/curator works
- gh attestation verify passes against the released artifacts
- macOS binaries are Developer ID signed; notarization submission wired (needs an App Store Connect API key to activate)
- install script performs os/arch detection and checksum verification
