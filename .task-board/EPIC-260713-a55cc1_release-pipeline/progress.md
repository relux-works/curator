## Status
done

## Assigned To
claude

## Created
2026-07-13T12:00:00Z

## Last Update
2026-07-13T13:40:00Z

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

Resolved: notarization is live as of v0.1.1 (App Store Connect API key
barycenter-ci, base64-encoded p8 in ASC_KEY_P8). Verified on a real
brew upgrade: the cask binary runs under quarantine, Gatekeeper clear.
Remaining hygiene item: replace TAP_GITHUB_TOKEN with a fine-grained PAT.
