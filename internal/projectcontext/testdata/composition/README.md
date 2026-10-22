# Composition grammar vectors

These 51 complete-record fixtures are copied without modification from
`relux-works/curator-spec` commit `8900129d`,
`conformance/project-context-leaf1-v1`. The manifest retains upstream byte hashes,
verdicts and individual claims. `TestCompositionPublishedVectors` exercises the
production Go reader. These are closed grammar fixtures; root document joining
is separately tested against protocol/project-context section 3.

The upstream corpus is Apache-2.0, like this repository. No native runtime or
adapter qualification is claimed by these vectors.

The six resources in `schemas/` are copied from that same commit, with byte
hashes retained in `schemas-manifest.json`. The Go schema checker uses a closed
loader over those resources and validates all four emitted record kinds for
Claude Code, Codex and Muse. An invalid-memory control must fail specifically
at the `memory` constant, then the restored valid record must pass.
