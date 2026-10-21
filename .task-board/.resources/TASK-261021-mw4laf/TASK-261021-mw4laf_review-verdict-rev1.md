# Review verdict — TASK-261021-mw4laf: muse-1-4-11-exclusion-replay

Verdict: **accepted**. CR-TASK-261021-mw4laf-1 revision 1.
Candidate tree: `5bb43ed306e560c96a47fc50f3fbd37b777ff2f3`.
Base: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`.
No code was changed. This accepts the qualification record, not interactive exclusion or runtime admission. Landing is not claimed.

```json
{
  "verdict": "accepted",
  "findings": [],
  "notes": [
    "The deciding brief names R6427.1; the task and binding SHA require R6667.1, which was exercised.",
    "Spec main advanced to 3a3e041b3fbf76292e69b6abb4a0eca855524007; all three normative input files are byte-identical to the recorded e41f96387287f5e76604627f9e63662f7d6d86d4.",
    "Evidence is pinned to the CR base, not current trunk. Upstream advanced by two commits with no candidate-path overlap. Any combined landing validation belongs to delivery.",
    "Exec accepts the interactive key spelling; exclusion semantics on that alternative remain explicitly unmeasured. Credentials, refresh and model-visible removal remain outside this loader qualification."
  ]
}
```

## Coverage and surface sweep

Coverage map was inspected before source: **6 of 6 task AC rows driven**. AC6 maps to command-level build/test/hygiene evidence, not a vendor observation. Production calls are current Curator `profile install` / `env resolve`, sandboxed vendor exec/TUI parsers and loaders, sandbox-exec curl/cat, and the published snapshot checker. No runtime gates changed.

# Review B sweep — revision 1
Candidate tree: 5bb43ed306e560c96a47fc50f3fbd37b777ff2f3.
Coverage inspected before source: 6 of 6 AC rows mapped; AC6 uses command evidence and unchanged-input default-suite reuse, not a new vendor test.
No review/free-hunt duration was supplied. Reviewer uses a bounded 20-minute round and a 2-minute free hunt.

| Surface | Result | Attack and evidence |
|---|---|---|
| Binary identity | held | Own TestVendorIntakeBoundary rejects synthetic wrong digest and confirms no fixture on unset gate; ungated package passes with vendor skips. Own TestMuse1411Identity reports exact R6667.1 and exit 0 under digest pin. Review brief R6427.1 is a typo relative to binding SHA and task. |
| Isolation | held | Own TestMuse1411OfflineBoundary passes paired loopback and synthetic Keychains controls; candidate invocation is sandbox-exec -> env -i -> echo with native-store denial. No auth invocation. |
| P1–P4 per form | held | Own 36-case matrix and two parser refusals completed successfully. Each matrix case reports exit 0 and actual counts; both parser refusals exit 2. Exec skill exclusion yields zero, interactive yields one, on linked and snapshot roots. Observation match is bounded to offline loader evidence. |
| §7.9 status line | held | Tried to contradict the status with the linked/snapshot interactive rows: retained skill reproduces and record explicitly refuses interactive. Exec status is explicitly restricted to loader rows, not credentials/model-visible isolation/runtime admission. |
| Narrowing control | held | Own exec omission plant exit 1, exclude-true fails user_skills=1; want 0, positive exclude-false passes. Own checker-one-skill mutant exit 1 at TestObservationChecker: checker admitted one foreign skill. |
| Spec citations | held | Fresh spec main 3a3e041b3fbf76292e69b6abb4a0eca855524007 checked. environments.md, CIP-0017, Muse v4 README are byte-identical to recorded e41f96387287f5e76604627f9e63662f7d6d86d4. No changed requirement or misnumbered section found. |
| Scope and hygiene | held | All 62 changed files match exact candidate blobs and allowed prefixes; root README/CHANGELOG unchanged. Negative personal-locator scan passes. Fresh upstream e8d906828354dfa90e8e244922e710213cec4671 has no intersecting candidate paths; worktree remains cdcba3bd and is two commits behind remote main. |

## Validation ownership and bounds

Own runs: ungated package, full gated vendor package with published reference checker, exec omission plant, checker narrowing mutant, Curator build, harness vet, exact-blob scope check, current-spec comparison, 1.4.10 observation comparison and default-evidence identity audit. All use `-count=1` where applicable. The opt-in harness used the producer's exact-base Curator executable; its Go build information names cdcba3bd. The additional reviewer build exited 0; its VCS metadata differed and it was not substituted into the replay.

Accepted producer evidence rather than rerunning: full `go build ./...`, `go vet ./...`, lint and remaining nine patch mutants. Default-suite reuse was independently verified: 1,671/1,671 manifest entries equal candidate blobs, manifest digest baad451f52d27e33920e1aee5d1b91a53686416fa0fcd5c0c94af280ea4594a9; complete original hosted log digest b96fe7f9635782fe8193cd8e1e968c7a7c942c9e8c3f4b7a67df0c1fa402b7af; full Go stream digest 3592b90d813b35f565b63fd1d0e5a71586f79b8b9a40418c4d148bb48381aca8 with zero failure events. This reuses its original hosted environment, not an equivalence claim about this machine. Hosted source: https://github.com/relux-works/curator/actions/runs/39193852508/job/117877582459.

Additional adversarial sweep: contradictory count/outcome records, missing records, altered/extra/missing snapshot entries, listed symlinks and unlisted dangling links all rejected with nearby valid controls, including the published reference CLI. Free hunt found no additional blocking mechanism. Held means only the named attacks did not reproduce, never proof of absence.

Own 1.4.10 comparison independently reproduces 45/45 equal normalized argv, exits, event/outcome/count records and visible-marker sets. A first ad-hoc comparator assertion was discarded after inspecting its extraction; the corrected comparison succeeds. No failed diagnostic is presented as a successful test.

The attached review evidence log includes complete gated/ungated outputs and the intended-red plant/mutant outputs. Private raw loader logs remain in ignored task-local storage. New review notes and findings are persisted here as the task-scoped review logbook; no control-root file was edited directly.

Final own replay: exit **0**, 296.505 seconds, **45 vendor cases: 43 exit 0 and two exit 2**, plus version inspection exit 0. All 36 matrix cases, two interactive failure witnesses and four workspace residuals passed their observation assertions. Ungated suite exit 0; expected-red omission plant exit 1; narrowing checker mutant exit 1. Companion resource: `TASK-261021-mw4laf_review-evidence-rev1.log`.
