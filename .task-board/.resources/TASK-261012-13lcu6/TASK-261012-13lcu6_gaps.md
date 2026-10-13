# TASK-261012-13lcu6 — launcher v0.3.0, revision 2: rev1 re-applied onto fresh main

## Re-application record (revision 2)

- New base OID: `f08f566dd378ea27ce23ca0351ec793e95a696b6` (LR6 landing, docs and
  figures only: README.md, SPEC.md, diagrams/README.md, four `.puml` figures).
- Source: revision-1 candidate patch (25 files), applied with `git apply --3way`.
- Per-file result: 24 files applied cleanly; SPEC.md applied with conflicts in
  exactly one region (the revision-note list at the top). No other file needed
  any resolution.
- SPEC.md resolution: kept BOTH entries, in this order — first the
  `0.5.1-draft` launcher-v0.3.0 entry (pending tag), then LR6's `0.5.0-draft`
  documentation-revision entry. The specification version header stays
  `0.5.1-draft`. No other SPEC.md edit was made.
- README.md applied cleanly with no conflict: the v0.3.0 hunks touch the
  interactive-launcher sections (Muse mapping, pipeline, options, install,
  tracked mode, exit table, mutantguard) while LR6 touched the §10
  target-mode prose. No overlapping passage.
- Delta-identity check for the 23 files the new base did not touch: 23/23
  worktree files are byte-identical to the rev1 post-image blobs, and 23/23
  base blobs are identical between the rev1 base and the new base — so the
  delta against the new base equals the rev1 delta exactly for those files.
- No other change: the worktree holds exactly the 25 candidate files,
  uncommitted, with no untracked files.

## Merged-document coherence read

- Read the merged README.md and SPEC.md once, as a whole. The only passage
  where v0.3.0 text meets LR6-changed text is the SPEC revision-note list,
  resolved above. The v0.3.0 text (pin v0.5.62 with attested Muse builds,
  `muse-wonder`, fragment v3, hosted launch unavailable with
  `session_host_missing`) and LR6's text (§10 wave-5 target contract,
  target-mode README prose) state both without contradiction.
- The §10 target session-host lifecycle stays disabled pending joint
  publication (§§10.2, 10.5); the interactive hosted path (§§4.8–4.9) fails
  closed until the swarma dispatcher path ships. Both are unavailable through
  different mechanisms; neither presents the other as working.
- Figures (LR6's, untouched by rev1 and by this revision): all four depict
  the interactive hosted path inside their CURRENT panels with no v0.3.0
  unavailability notice — `components.puml` (exporter to task-board receiver,
  "Hosted Phase 1" note), `launch-paths.puml` ("receiver available" handoff
  branch), `launch-states.puml` (Compose to Hosted to Exited), and
  `artifact-layout.puml` (Hosted 1.0.0 payload, receiver, FD3/FD4);
  `diagrams/README.md` likewise indexes "hosted Phase 1" without the notice,
  and the CURRENT panels still read "SPEC 0.5.0-draft". Recorded here, not
  redrawn, per the revision brief; a later docs pass owns the figure update.

## Carried revision-1 change (summary, unchanged in revision 2)

- Pin `skill-agents-management v0.5.62`, no replace/workspace override. Muse
  interactive launches admit any well-formed build the selected binary
  attests, including 1.4.6, with no release table; the frozen-copy tuple is
  optional and without it `muse` resolves on the handed PATH; the
  `muse-wonder` alias and fragment v3 are admitted; the module lineup
  defaults Muse to `muse-wonder-1.0-contributor` / `xhigh`.
- API drift adapted: the probed `ToolRelease` claim passes through unchanged
  and the module refuses a mismatch with the binary attestation; Muse yolo's
  no-evidence mapping query defers to the real builder admission over bounded
  binary help, reporting `mapped=module-attested` only then.
- `--hosted` fails closed with the existing `session_host_missing` (exit 1)
  before any handoff, even with `task-board` on PATH; no new refusal code.
  Machine and operator hosted defaults cannot start a hosted launch. Native,
  `--native`, `--untracked` and direct `--network` are unchanged.
- `--help`, README, SPEC `0.5.1-draft` and CHANGELOG `0.3.0` (pending tag)
  state that hosted launch is unavailable until the swarma dispatcher path
  ships. The binary reports `0.3.0`.
- Refusal tests keep their meaning: new Muse malformed/mismatch/help rows,
  new public hosted rows over flag/operator/machine/locked-machine by
  native/yolo, and required hosted mutant M-H22 alongside M-H1–M-H3. No
  refusal test was weakened.

## Carried stale-copy observations (left unchanged in revision 2)

Per "make no other change", these rev1-carried inconsistencies were not
edited; they are for the reviewers to route:

- README's top contract link still reads "SPEC 0.5.0-draft" while SPEC is
  `0.5.1-draft` (untouched by both rev1 and LR6).
- SPEC §10 preamble: "The executable's existing specification version
  remains `0.5.0-draft`" (made stale by rev1's version bump; untouched by LR6).
- SPEC §4.2 Pi paragraph: "this revision pins v0.5.48" (already stale before
  rev1, whose base pinned v0.5.53).

## Validation (revision 2)

- No local `go test` and no local `go build` were run (revision host rules).
  No LOGBOOK.md write. No commit, tag, release, or push to `main`.
- Hosted gate: `sh scripts/remote-gate.sh` exited 0 on attempt 1; snapshot
  `a15a715f95d7d5744fdb9e23df6b19ac76c7919c`; CI run 38363575922
  (https://github.com/relux-works/curator-agent-launcher/actions/runs/38363575922).
  Test and race lanes on Linux/macOS: 4/4 green with every step green; the
  optional self-hosted lane is skipped by repository configuration, not
  passed. The snapshot tree equals the worktree tree exactly, so the gate
  ran on this candidate. No repository change was made after the snapshot.
- `git diff --check` clean; the candidate diff carries no local paths,
  session links, or board-resource links.

## Open

- Hosted activation awaits the swarma dispatcher path; the retained payload
  scaffolding is not a shipped hosted launch.
- Muse `--network` tuples remain unsupported by the independent child-scope
  ceiling; qualification does not expand it.
- BUG-261011-sr6fvi: a scoped board read in this revision again returned
  element-not-found, so its title, status and disposition could not be
  established. This task does not claim to resolve it.
- The LR6 figures' hosted-availability depiction and stale CURRENT labels
  await the later docs pass noted above.
- Tag/release/landing belong to the orchestrator after review. All
  repository changes stay uncommitted in the isolated Story worktree.
