# BUG-260920-2d9gfv brief (orchestrator, binding)

Story STORY-260919-37szes (Story tip e0f52c95 carries 3ukdk4, 2wyzde, 3v7x6j, 2eg8nv, 1sbj7o,
3ccq6b). Reproduction set: `TestDraftSourcesSemanticCases/attestation-evidence-unreadable`
(gate 35486280770, Race ubuntu) and `/attestation-evidence-wrong-key` (gate 35529945079, Race
macos, evidence note 2d9gfv-evidence-260920.md): Layer 1 fresh install returns
`every trusted audit registry served a tampered snapshot` instead of `is not audited by any
trusted registry`; the Test lanes of the same runs pass the rows.

## Root-cause hypothesis to confirm first (orchestrator reading of the code — verify, do not assume)

- The harness Go-API config (`draftInstallConfig`, draftsources_fixtures_test.go:126) is a
  bare `config.Config{}` literal: `Audit.SnapshotClockSkewSeconds = 0` and
  `SnapshotMaxAgeSeconds = 0` (loader defaults are 300 s / 604800 s; `checkSnapshotsWithPolicy`
  substitutes the max-age default but NOT the skew).
- `install.go:1347/1353` captures `now := time.Now()` BEFORE the snapshot fetch; the stub
  signs `created_at = time.Now().UTC().Truncate(time.Second)` AT serve time
  (draftsources_semantic_evidence_test.go:137). Whenever a whole-second boundary falls between
  the two clock reads, `created_at > now` by up to 1 s, and with zero skew
  `parsed.CreatedAt.After(now.Add(clockSkew))` (snapshot.go:~152) classifies the registry as
  tampered ("snapshot timestamp is too far in the future"). The window is the fetch latency
  (ms on the Test lanes, tens–hundreds of ms under `-race`), which matches the lane pattern
  and both rows (any evidence row that runs Layer 1 with the live stub can flip).
- Confirm by driving it deterministically at the production entry: make the stub (in a
  throwaway or committed row) sleep until just past the next second boundary before signing,
  run `install.Project` with the zero-skew config, observe the flip; then show the same
  request with the fix no longer flips. If the evidence contradicts this hypothesis, stop and
  report the actual cause with the same rigour before changing anything.

## Rulings

R1 Product fix at the root (this is a product defect: a refusal class must not depend on the
fetch's own duration). The future-timestamp check must be evaluated against a clock reading
taken after the snapshot fetch — preferred shape: keep the exported signatures and semantics
(`now` = the check's reference time, deterministic for callers that inject a fixed `now`) and
add the elapsed monotonic time since function entry to the comparison bound
(`parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(start)))`), or an equivalent clock
seam used by the install caller — say why. A timestamp genuinely ahead of the client's clock
after the fetch must still be refused with the same class and text (negative row). No change
to the stale check semantics.
R2 The legacy v1 lane shares `registry.CheckSnapshotsWithPolicy*`; this is a declared bug fix
that only removes the false "future" refusal — CHANGELOG `## Unreleased` → `### Fixed`, and a
legacy-lane row (registry evidence through the v1 install entry with zero skew and an
on-demand-signed snapshot) proving the fix and that a real future timestamp still refuses.
Legacy goldens (`TestDraftTransportLegacyGolden`, `TestProjectResolveLegacyUntouched`) stay
green.
R3 Harness: keep the zero-skew Go-API config as the strict case (it exposed the defect) but
document it in a comment; do NOT paper over the flake with a larger skew or a fixed
`created_at` in the stub. Both reproduction rows plus every other evidence row must be
deterministic: 20 consecutive `go test -race -count=20 -run
'TestDraftSourcesSemanticCases/attestation-evidence-'` passes locally (record the exact
command and elapsed time; the host has stall windows — rerun once if a stall kills it).
R4 Evidence: production-entry row that drives the boundary crossing deterministically (stub
delay until the next second) and passes only with the fix; negative row for a genuinely
future timestamp (created_at = now + skew + 2 s → still refused); narrowing mutants (drop the
elapsed term → the deterministic row fails; drop the future check entirely → the negative row
fails); unit rows in `internal/registry` for the elapsed-time bound with an injected `now`.
results.md: root cause with the driven proof, rulings restated, the 20-run determinism
evidence, mutant table, ratio line (`registerDraftSemantic` unchanged unless you add a corpus
row), Windows proof status. Publish the Change Request only when the configured gate is green.
