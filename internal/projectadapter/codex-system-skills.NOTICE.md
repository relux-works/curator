# Pinned Codex native system skills

`codex-system-skills.tar.gz` contains the unchanged system-skill bytes from
Codex CLI 0.162.1 for macOS arm64. Curator prepublishes them so native startup
preserves the immutable receipt instead of adding unreceipted files.

- Native executable SHA-256: `74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0`.
- Archive SHA-256: `dbfd0e484b3dae9a72f197153c60cdf2e6bcf57536294def3d081a663718770f`.
- 49 regular files, 307566 uncompressed content bytes, relative paths only.
- 48 asset files match the Git blob IDs under
  [`codex-rs/skills/src/assets/samples` at `rust-v0.162.1`](https://github.com/openai/codex/tree/rust-v0.162.1/codex-rs/skills/src/assets/samples).
  The release tree is `092d3acd6bec3e3a14bdc7e7a2810ab628ab759d`.
- The remaining file is `.codex-system-skills.marker`, with native bytes
  `4af37f669ff79b4f` followed by a newline. Independent fresh native startup
  fixtures reproduce it. Curator does not synthesize a marker-only cache.
- Native installation behavior is in
  [`codex-rs/skills/src/lib.rs`](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/skills/src/lib.rs).

The archive retains the upstream license files at `imagegen/LICENSE.txt`,
`openai-docs/LICENSE.txt`, `skill-creator/license.txt` and
`skill-installer/LICENSE.txt`. These assets originate in OpenAI's Codex
repository. Curator does not modify their contents or execute their scripts.
The tar metadata is normalized to uid/gid/mtime zero and mode 0600, with no
owner names. Publication uses private directories and receipt-bound modes.

To inspect the archive: `tar -tzf internal/projectadapter/codex-system-skills.tar.gz`.
To check its digest: `shasum -a 256 internal/projectadapter/codex-system-skills.tar.gz`.
The opt-in native regression in `.research/project-exclusion-codex` verifies
that complete production plans survive actual startup and unchanged restart.
An existing native home is never used as a source or recaptured as a baseline.
