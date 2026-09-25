# TASK-260922-3bbvrs — rework 1 (orchestrator, binding): platform-case gate rejects the new skip reason

Revision 1's hosted gate (run 35889012099) failed on ubuntu, macOS and Windows (test + race) for one reason — the
platform-case gate does not recognise the skip reason your gap ledger now emits for 8 buildsource vector rows:

    FAIL  skip with an unrecognised reason on linux|darwin|windows: internal/buildsource :: TestBuildSourceConformanceVectors/<case>
    cases: build-source-mutation-during-use, build-source-special-file, build-source-symbolic-link,
           duplicate-build-source-path, invalid-unicode-build-source-path, legacy-nul-stream-structural-collision,
           mode-and-timestamp-are-non-inputs, root-marker-bytes-are-build-input
    reason text: "this case is outside the accepted non-empty record identity assertion"

(These same cases PASS under `TestBuildSourceIdentityVectors`.) The ledger's whole point is that every
non-driven row carries a DECLARED classification the gate understands. Fix it at the source:
- either drive these rows in `TestBuildSourceConformanceVectors` (if the identity assertion now covers them), or
- classify them as `bound`/`known-gap` in the ledger with a skip reason that matches a registered class in
  `.github/ci/skip-classes.tsv` — add ONE narrow class for the ledger's bound/known-gap classifications if none
  fits, with a `gate-selftest.sh` row proving an unregistered reason still fails; never broaden an existing
  class (`no-broad-suppression.sh` must stay green). The ratchet must still refuse a driven row regressing to skip.
Run locally (bounded calls): `sh .github/ci/gate-selftest.sh`, `sh .github/ci/ledger-consistency.sh`,
`go test ./internal/buildsource/... ./internal/crossconformance/...` and emulate the platform-case gate over the
skips you produce (`.github/ci/platform-case-gate.sh` against your local go-test json). Append "Revision 2" to
results, then `task-board handoff TASK-260922-3bbvrs --role developer`. A `run_wrote_outside_worktree … policy
warn` block is a warning — verify the status moved to `to-review`.
