# Public signature fixtures

Unmodified public Git tag objects from `swarma-memory` v0.2.2 (platform release key), `curator` v0.15.0-rc.8 (Curator release key), and historical `curator` v0.15.0-rc.7 (agent commit key, forbidden for new source admission). Each file is the exact `git cat-file tag` payload. Tests verify real SSH tag signatures offline with source-specific public keys. Target commits are not needed for standalone signature checks and are deliberately not copied. The historical commit-key tag is negative evidence only. No private keys are included or needed.
