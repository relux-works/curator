# Review verdict: launcher SPEC 0.1.1-draft (pi file-kind channels) — CHANGES REQUESTED

Reviewer run: RUN-260901-7354f1 on TASK-260901-25zqel, CR-TASK-260901-25zqel-1 rev 1.
Subject: curator-agent-launcher `draft/spec-pi-file-channels`, head `0029b36` (signed, verifies G), base `origin/main` @ `dae0c35`; delta confined to SPEC.md (110+/27-). Curator-side CR delta: LOGBOOK.md only, matches the SPEC facts.

## What was verified and holds

1. File-kind semantics (brief item 1): correct. §5 states the launcher never places/removes/edits `APPEND_SYSTEM.md`/`SYSTEM.md`; ownership by Curator materialization under `system_prompt_files` (default `off`) matches environments.md §5.5 verbatim; "tool applies unconditionally when present" matches §7.3.
2. Warning rule (brief item 2, positive path): §5.1 probe + §5.2 make the warning mandatory for an active file-kind channel without `--system-prompt`; both-kinds case enumerates every active channel; opt-in engages non-`file` channels only, orthogonal and additive. Matches Decision 0010 D6's warning triple.
3. Diagnostic (brief item 3): `sysprompt_file_unreadable` defined in §5.1, coded in the §6 table, and correctly folded into the §6 absence-vs-read-failure invariant. The producer's deliberate narrowing of the brief's "absent/unreadable" to unreadable-only is CORRECT and I endorse it: the fragment reproduces pi's full §7.3 descriptor list whenever any system module applies (environments.md §10.2), and `system_prompt_files` defaults `off`, so absence is the channel's legitimate inactive state — an absence diagnostic would fail every plain pi launch. Dangling symlink / non-regular file routed to unreadable, not absence: correct per the invariant.
4. Hygiene (brief item 4): revision history §8.1 present, version 0.1.1-draft in header, §8, and history; commit signed (G, ivan@relux.works ED25519); no code changes.

## Finding 1 (MAJOR, blocking): §5.1's stray-file clause contradicts environments.md and leaves an unwarned bypass of the warning contract

§5.1 closes with: a fragment with no `system_prompt` section has nothing to probe, and "hygiene of a managed home that nevertheless contains a stray `SYSTEM.md` is Curator's drift-and-repair surface (§4.3 resolution repairs the home before the fragment is returned), not the launcher's."

That claim is false against the cited protocol:
- environments.md §10.1: repair re-materializes managed surfaces and passthrough links "while leaving environment-owned mutable state, unmanaged files, and backups untouched", and "MUST NOT adopt candidate bytes found in the home".
- environments.md §8.4/§8.5: drift is defined only over marker-RECORDED surfaces. Under `system_prompt_files=off`, `SYSTEM.md` is never materialized and never recorded, so a stray one is an unmanaged file: drift detection does not flag it and repair does not remove it.

Consequence (the negative shape: the guard the SPEC relies on never runs on this path): resolve a pi profile whose composition carries no applicable system modules into a managed home where someone or something has written `SYSTEM.md`. The fragment has no `system_prompt` section, so §5.1 probes nothing; §4.3 repair leaves the file in place; the launcher execs; pi applies `SYSTEM.md` unconditionally with full-replacement semantics. Result: a silent, replace-semantics customized run with zero warning — exactly the outcome §5.2's own rationale says MUST NOT happen ("the cost of a silent customized run is an operator misattributing behavior to the tool"), and exactly the scenario the machine-setting rationale invokes (the operator at the keyboard is not the one who mutated the home). This fails brief items 2 (warning bypass path exists) and 5 (contradiction with a cited normative reference).

Required rework (either closes the finding):
- Preferred: extend the §5.1 probe to be keyed on the environment, not the fragment — for an env-id whose §7.3 adapter registry declares `file`-kind channels (revision 1: exactly `pi`), probe the closed registry filenames in the managed home on every launch, fragment `system_prompt` section or not. The registry is closed and the launcher already owns the §4.2 env mapping, so this adds no new knowledge edge. A present-and-readable file with no corresponding fragment descriptor is an active channel: warn (and state which contract owns removing it). Present-but-unreadable stays `sysprompt_file_unreadable`.
- Or: delete the false Curator-repairs-it claim and state the residual honestly — a stray file-kind file in a home with no applicable system modules is applied by the tool unwarned, outside both the launcher's probe and Curator's repair as currently specified — and record it in §9 Open items as a known gap. (Weaker: the warning contract keeps a named hole; acceptable only if deliberately chosen.)
- Not acceptable: keeping the paragraph as written, since it asserts a protection that no component provides.

## Non-blocking notes

- CR rev 1 on the curator side bundles the TASK-260901-32j97g bootstrap logbook entry into the same commit `411d3184` — a consequence of the documented provenance rewrite after `change_request_base_authority_mismatch`; content is accurate, no action needed.
- §5.2 retains the pre-existing "before the exec" wording, which in the ax-configured shape means "before the §4.5 handoff"; carried from 0.1.0-draft, fine to leave.

## Verdict

CHANGES REQUESTED → to-dev. Fix Finding 1 in SPEC.md (0.1.2-draft with a revision-history row), re-run build/vet, new signed commit on the same branch, new CR revision.
