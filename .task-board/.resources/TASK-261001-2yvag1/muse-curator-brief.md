# TASK-261001-2yvag1 — curator: the muse environment (curator#100) (THE ONLY CURRENT INSTRUCTION; priority, Muse root-session critical path)

The normative source is the curator-spec candidate on branch `spec-muse-environment` (commit d373078a, base add50233; accepted after
an astra-high review). Read its protocol/environments.md muse sections, Decision 0018 additions and launch-env-fragment-v3, plus
relux-works/curator#100. It is NOT on spec main yet. The spec PR lands in lockstep with your commit, the way #116 did with
TASK-260930-12i5zr.
1. Implement the `muse` environment in the adapter registry (internal/envregistry) and the managed-home/profile code
   (internal/envprofile):
   - XDG_CONFIG/DATA/STATE/CACHE_HOME set to `<home>/{config,data,state,cache}`;
   - NEVER replace HOME;
   - seeds config/muse/settings.json and trust.json;
   - the skills and root-context targets as specified.
2. Passthrough `shared`: `config/muse/auth.json` is a file-link to the native auth.json. At env resolve, detect a replaced or forked link
   (a regular file instead of the link, a link to the wrong target, a dangling link) exactly as the spec's 16 link-state × resolve/repair
   rows define. Repair or refuse per the spec, and never silently use a forked credential. New state reads go through
   internal/stateread; link writes and repairs use the E5 nofollow helpers.
3. `curator env resolve muse --repair --format json` and `curator run muse` produce the spec's launch-env-fragment-v3. Keep v1/v2
   fragments unchanged for the other environments unless the spec says otherwise.
4. Conformance: drive the new muse vectors with CURATOR_CONFORMANCE_ROOT pointed at a clone of the candidate (d373078a). Under SPEC_PIN
   rc.13 (curator CI), everything stays green and unchanged. Key exact counts by suite digest (like 12i5zr). Anything not driven is an
   owned known gap; nothing skips silently.
5. Tests use a FAKE muse binary that prints its env/argv. No real Muse session and no real credentials. Add rows for:
   - the XDG layout;
   - HOME unchanged;
   - each link state;
   - the fragment shape.
6. Mutants, with real exit codes:
   - HOME replaced;
   - a forked auth.json accepted;
   - XDG_DATA_HOME missing.

   Each must be killed.
7. No Windows-reserved names. No CHANGELOG/LOGBOOK: put the entry text in the results. Never spell any employer name.
Update the results, then run `task-board handoff TASK-261001-2yvag1 --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate.

## Decision (binding, 2026-10-01)
Scope split: hand off the curator resolver side now. `curator run muse` through curator-run is a separate launcher leaf: do not touch the launcher repository. Mark the run row as an explicit bound in the results (it needs a v3-capable curator-run). Finish the checklist, then run `task-board handoff TASK-261001-2yvag1 --role developer` and END YOUR TURN.
