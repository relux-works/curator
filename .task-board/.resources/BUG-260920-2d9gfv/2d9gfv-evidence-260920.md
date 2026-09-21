# BUG-260920-2d9gfv — additional occurrence (orchestrator evidence note, 2026-09-20)

Gate run 35529945079 (TASK-260920-3ccq6b rev1, a gitops/env-isolation change unrelated to
attestation evidence), job `Race (macos-latest)` 106128653400, artifact
`race-evidence-macos-latest` 10611258201: `internal/crossconformance ::
TestDraftSourcesSemanticCases/attestation-evidence-wrong-key` failed with
`fresh install errors = [every trusted audit registry served a tampered snapshot], want "is not
audited by any trusted registry"` (0.24 s). Same refusal-class flip as the original
`attestation-evidence-unreadable` occurrence (run 35486280770, Race ubuntu) — so the
nondeterminism spans at least two evidence rows (unreadable, wrong-key) and both race lanes.
The Test lanes of the same run passed the row. Treat both rows as the reproduction set.
