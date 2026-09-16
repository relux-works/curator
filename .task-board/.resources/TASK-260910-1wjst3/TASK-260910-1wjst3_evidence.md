# TASK-260910-1wjst3 evidence (rev2 — answers review-verdict-rev1)

Finding: S6 (spec `docs/security-audit-2026-09.md` §S6; manager
`curator/docs/security-audit-2026-09.md` S6/I1).
Story worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2awkzu/worktree`
(branch `task-board/story/STORY-260910-2awkzu`, diffed against `origin/main`).
Role: doc-writer. No implementation code touched; no commits, pushes,
branches, or PRs. No writes outside the story worktree (patch and this
evidence live in `/tmp` and are attached as board outcome resources).

Revision note: rev1 items the reviewer marked "Pass" are unchanged
(normative §§8.1–8.6, CLI command rows, CHANGELOG rule text, manifest
registration pattern). Rev2 closes exactly R1 and R2 below. Two
pass-adjacent values moved as a consequence: the CHANGELOG Vectors
sentence now describes the reproducible vector, and `example_record`
now uses the real `env-sh-v1` fixture digest instead of the empty-byte
placeholder.

## R1 closure — CLI guidance carries both rollout revisions (medium)

`cli/curator.md` Developer-shell paragraph rewritten to qualify the
trust behavior by the two labelled revisions with the exact profile
spellings from §8.5:

- Revision A (`A-warning`, warning release, shipped first): keeps
  sourcing unknown/changed files, warns once per shell session with
  `shell_hook_env_unapproved` / `shell_hook_env_changed`, migration
  hint naming `curator hook approve <path>`.
- Revision B (`B-enforcing`, flip release): sources only trusted
  bytes; unknown/changed files are not sourced, warn once per shell
  session naming path + approval command, and continue.
- Re-approval required after a change under both revisions.

## R2 closure — emitted-hook conformance surface is reproducible (high)

Choice: the second option of the rework brief — downstream execution
binding plus structural validation in `tools/validate.py`. Reason: the
repository's gates are JSON-structural (`tools/validate.py`) and
vector-emitting (Go `tools/generate-vectors`); neither can execute a
POSIX or PowerShell hook, so an in-repo execution gate is not
available. The normative text still carries the full execution recipe
so the downstream owners can run it.

1. `profiles/manager.md`, new §8.7 "Conformance surface and execution
   binding": declares the emitted POSIX/PowerShell hook code a
   conformance-vector surface that MUST satisfy every case of
   `conformance/v1/vectors/shell-hook-trust.json`; states the forged
   project-local record MUST be ignored; gives the 5-step execution
   recipe (materialize fixture bytes at the absolute candidate path,
   seed manager-home state with exactly the case record, keep the
   forged record in project data only, select the profile, run two
   same-session activations and compare sourced/diagnostic/warning
   count); and carries the labelled downstream execution binding:
   spec gates do NOT execute hooks, execution is owned by
   `TASK-260910-1952mz` (manager-hook-digest-pin, the hook trust gate)
   and `TASK-260910-3ungjy` (hook-approval-command).
2. `conformance/v1/vectors/shell-hook-trust.json` rebuilt as a real
   input/expected-output contract (14 cases: the rev1 12-case matrix
   plus a forged-record case under each profile):
   - `fixtures`: 4 entries (`env-sh-v1`, `env-sh-v2-changed`,
     `env-ps1-v1`, `env-ps1-v2-changed`), each with `bytes_base64` and
     the `sha256` computed over exactly those bytes (digests computed
     by script, not transcribed).
   - Each case carries `candidate_path` (absolute),
     `candidate_fixture`, `observed_sha256`,
     `manager_approval_record` (exact `{ path, sha256, approved_by,
     approved_at }` or null), `project_supplied_record` (null except
     the forged family, where a `project:`-sourced record matches the
     observed bytes yet the case stays unapproved), `rollout_profile`,
     and expected `sourced` / `diagnostic` / first-activation warning
     / second-activation silence / total warning count across the
     two-activation session pair.
   - `activation_sequence`, `execution_recipe` (7 steps), and
     `execution_binding` (surface, spec-gate honesty note, downstream
     owner tasks, downstream contract) blocks.
3. `tools/validate.py`: new `validate_shell_hook_trust_vectors`,
   registered in `main()` so it runs on every `make validate`. It
   recomputes every fixture digest from the fixture bytes, checks the
   closed record shape / files / diagnostics / profiles / case
   inventory, and derives each case's expected `sourced`, diagnostic,
   and 7 warning fields from trust state + rollout profile. It never
   executes a hook (stated in its docstring).
4. `tools/test_validate.py`: new `ShellHookTrustVectorTests` (8
   tests: published vector passes + 7 narrowing mutants).
5. Reviewer's mutant re-run: flipping `sourced` to true in
   `changed-env-sh-B-enforcing-not-sourced` is now detected:
   `shell-hook-trust case changed-env-sh-B-enforcing-not-sourced
   sourcing does not follow its trust state and profile`.
   A stale fixture digest is likewise detected. (Transcript below.)

## What changed per file and why

1. `profiles/manager.md` — new §8.7 only (§§8.1–8.6 untouched).
2. `cli/curator.md` — Developer-shell paragraph qualified by
   Revisions A/B (R1); command rows untouched.
3. `conformance/v1/vectors/shell-hook-trust.json` — rebuilt as the
   reproducible contract (R2); same 12 rev1 case names kept, 2 forged
   cases added.
4. `tools/validate.py` — structural gate for the vector (R2; spec-repo
   validation, not implementation code).
5. `tools/test_validate.py` — 8 unit tests for the gate.
6. `conformance/v1/manifest.json` + `release/1.0.0-rc.9.json` —
   regenerated via `make regenerate` (new digest pin only; no other
   vector bytes changed).
7. `CHANGELOG.md` — Vectors sentence extended (fixtures, exact
   records, forged record, §8.7 binding); rule text untouched.

`git diff origin/main --stat` lists only these files (7 tracked +
the new vector); `git status --short` shows nothing else.

## Closed-set spelling ledger (identical in text, tables, vectors, gate)

- Diagnostics: `shell_hook_env_unapproved`, `shell_hook_env_changed`
- Commands: `curator hook approve <path>`, `curator hook approvals`,
  `curator hook revoke <path>`
- Files: `.agents/env.sh`, `.agents/env.ps1`
- Record members: `path`, `sha256`, `approved_by`, `approved_at`;
  values `manager`, `operator`
- Rollout profiles: `A-warning`, `B-enforcing`
- Fixtures: `env-sh-v1`, `env-sh-v2-changed`, `env-ps1-v1`,
  `env-ps1-v2-changed`
- Cases: 14 exact names (12 rev1 + 2 `forged-project-record-…`)
- Downstream owners: `TASK-260910-1952mz`, `TASK-260910-3ungjy`
  (profile §8.7, vector `execution_binding`, validator constant)
- No new §12.1 knob and no new lockable key (the brief does not ask
  for one); no new schema member.

## Schema note (unchanged from rev1)

`schemas/v1/` carries wire objects only; the approval record is
manager-home state, defined in text only (§8.2). No schema added.

## Out of scope (deliberately untouched)

Implementation (`TASK-260910-1952mz`, `TASK-260910-3ungjy` — named
only as downstream execution owners), S4/E4/E2 rules (parallel
stories), proposals 0014–0018, `.agents/env.ps1` semantics beyond the
identical rule, `protocol/environments.md` §12 knobs, `schemas/v1/*`,
tags/releases, `ax`. The validator addition is spec-repo structural
validation required by the rework brief, not implementation code.

## Validation transcript (all commands run from the story worktree)

Shell `/bin/sh` throughout (`set -o pipefail` where a pipe is used;
exit codes below are the gate command's own status). The three
`make validate` gates were run as their constituent commands across
bounded calls (one full-suite call exceeds a single bounded window);
each command, shell, and exit code is quoted.

1. `make regenerate` — exit code 0. Output:
   `go run ./tools/generate-vectors -root .`
   (manifest + rc.9 pin updated for the rewritten vector only; no
   other vector bytes changed — `git diff --stat` shows only the
   manifest and rc.9 pin deltas under conformance/v1 and release/.)
2. `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" python3 tools/validate.py`
   — exit code 0. Output quoted verbatim:
   `validated 60 schemas and 1048 vector files`
   (includes the new `validate_shell_hook_trust_vectors` gate; a
   pre-regenerate run exited 1 with `vector digest mismatch for
   vectors/shell-hook-trust.json`, the expected stale pin, fixed by
   step 1 and re-run green on the final tree.)
3. `PATH="…/.temp/venv/bin:$PATH" python3 -B -m unittest discover -s tools -p 'test_*.py'`
   — exit code 0. Output tail quoted verbatim:
   ```
   Ran 235 tests in 375.493s

   OK
   ```
   (227 pre-existing + 8 new `ShellHookTrustVectorTests`; the new
   class alone also runs green: `Ran 8 tests … OK`.)
4. `go test ./tools/...` — exit code 0. Output:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.647s`
5. `git diff --check` (with `git add -N` intent-to-add, so the new
   vector is covered) — exit code 0, no output.
6. Reviewer-mutant re-run (in-memory, worktree untouched) — both
   detected, positive control passes:
   ```
   MUTANT-SOURCED: detected -> shell-hook-trust case changed-env-sh-B-enforcing-not-sourced sourcing does not follow its trust state and profile
   MUTANT-DIGEST: detected -> shell-hook-trust fixture env-ps1-v1 sha256 does not match its bytes
   POSITIVE: published vector passes
   ```
7. Patch identity: `git diff origin/main | git patch-id --stable`
   and `git patch-id --stable` on the attached
   `TASK-260910-1wjst3_spec-patch_rev2.patch` both give
   `d8a2632131680ddc196f67d810a88fd1468faabb`.

## Checklist mapping

- Normative rule + closed sets (§8.7, ledger above): satisfied.
- Warn-first two steps + posture (§§8.5–8.6, R1 CLI qualification):
  satisfied.
- Vectors + manifest + `make validate` exit 0: satisfied (see
  transcript); reviewer mutant now detected.
- CHANGELOG + spec patch + evidence, no implementation code:
  satisfied (`TASK-260910-1wjst3_spec-patch_rev2.patch` attached;
  `tools/` edits are spec-repo validation per the rework brief).
- Docs consistent: CLI, profile, CHANGELOG, vector, and gate use the
  identical spellings (ledger + grep counts).
- No code/description discrepancies: gate derives exactly the rule
  arms in §8.1/§8.5/§8.7; 14 cases cover approved / unapproved /
  changed / forged × A / B (forged: sh file under both profiles —
  record-source rejection is file-agnostic, and both sourcing
  behaviors are shown).
- Outcome resources: this evidence file and the rev2 spec patch.
- Logbook: nothing anomalous (validation green, no regressions, no
  forced fits); campaign rules forbid LOGBOOK.md edits from spec
  tasks, so no logbook write was made.
