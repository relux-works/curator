# TASK-260924-1nh93t — F-M1e rework 1 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 2 was CHANGES_REQUESTED (`TASK-260924-1nh93t_review-verdict-rev2*`). Everything else verified and KEPT (classifier,
typed errors, rows, mutants). Blocking: the audit missed launcher SPEC §4.3 — every untracked launch prints, BEFORE admission,
`curator-run: permissions=<mode> source=<...> mapped=<flag or none>`, and "the `mapped` value is supplied by agents-management;
this SPEC names no provider flag". No public API returns that mapping today (`ReleaseCapability` has Release/Grammar/
YoloSupported only; `Plan`/`LaunchProvenance` carry no mapping).
1. Add the audit row for §4.3 `mapped=` and a public, release-versioned API, e.g.
   `Registry.PermissionMapping(system, toolRelease, mode) (PermissionMapping{Flag string /* "" = none */, Grammar string}, error)`
   via an optional plugin capability reusing the SAME verified-release rows as the permission grammar; `native` → none;
   `yolo` → the module-owned bypass flag for that verified release; unknown system / unverified release / unsupported mode →
   the existing typed errors (fail closed). Callable before admission (no plan needed).
2. Re-read SPEC §4 ONE more time for any other "supplied by agents-management" / "the module returns" phrase and list each in
   the audit with its API (state explicitly "none further" if so).
3. Rows per system × mode × verified/unverified release; mutant: return a flag for native, or for an unverified release →
   killed. Spelling guard green; README + extend the SAME unreleased CHANGELOG bullet; `go test ./...`, `go vet ./...`
   (bounded — host load is high; if a package exceeds its timeout under load, re-run that package alone and say so).
Append "Revision 3" to results, `task-board resource update` it, then `task-board handoff TASK-260924-1nh93t --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
