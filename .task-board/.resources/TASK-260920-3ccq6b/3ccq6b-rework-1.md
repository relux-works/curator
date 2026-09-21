# TASK-260920-3ccq6b rework 1 (orchestrator, binding)

Revision 1 gate FAILED (run 35529945079 on gate commit d0b9e056). Two causes, one yours:

1. YOURS — Windows platform-case gate: `FAIL skip with an unrecognised reason on windows:
   internal/gitops :: TestCloneIsolatedIgnoresGitProxyCommand`. Every skip reason must match the
   vocabulary in `.github/ci/skip-classes.tsv` (see `.github/ci/platform-case-gate.sh` tier 2);
   use the exact reason text/class the sibling POSIX-shim rows use (e.g. the reason printed by
   `TestDraftLiteralRefreshIgnoresUserConfig`'s Windows skip), or register the case in
   `.github/ci/platform-cases.tsv` with the same must/skip shape as its siblings. Check every
   other new row's Windows skip reason the same way (rows a–e, g) so the next gate does not
   fail on the next one. Do not widen the ledger vocabulary.
2. NOT YOURS — Race (macos-latest): `TestDraftSourcesSemanticCases/attestation-evidence-wrong-key`
   (`fresh install errors = [every trusted audit registry served a tampered snapshot], want "is
   not audited by any trusted registry"`) — the known race-lane nondeterminism BUG-260920-2d9gfv
   (next leaf of this Story). Do not chase it; note it in results.md as an unrelated flake with
   the run id. The republish reruns the suite; if it flakes again, republish once more.

Continue from the revision-1 tree in the Story workspace (no checkout/clean/stash). Then
republish (revision 2) and hand off when the gate is green; append "Revision 2" to results.md
with the skip-reason fix and the exact ledger match.
