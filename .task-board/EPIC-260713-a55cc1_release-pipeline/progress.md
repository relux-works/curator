## Status
done

## Assigned To
claude

## Created
2026-07-13T12:00:00Z

## Last Update
2026-07-13T13:00:00Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
v0.1.0 released and verified live: archives for six targets, checksums,
SBOMs, cosign signature, deb/rpm, Homebrew cask, Scoop manifest,
install.sh (downloaded, checksum-verified, executed), gh attestation
verify passes, macOS binaries Developer ID signed (262RZ595FP).

Open item: Gatekeeper kills the brew-cask binary because Homebrew
quarantines cask artifacts and the binaries are signed but not yet
notarized. Everything is wired: add ASC_ISSUER_ID, ASC_KEY_ID, and
ASC_KEY_P8 secrets (App Store Connect API key), uncomment the notarize
block in .goreleaser.yml and the env lines in release.yml, retag. The
install.sh and go install channels are unaffected.
