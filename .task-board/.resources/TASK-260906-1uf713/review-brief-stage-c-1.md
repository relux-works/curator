# Review brief: stage (c) cycle 1

## Subject

- Repository `~/Developer/ReluxWorks/curator`, worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `833918d2`, 10 signed commits past `origin/main`, rebased onto
  current main with `git range-diff` showing all ten content-identical. PR
  https://github.com/relux-works/curator/pull/61 — pushed, so hosted lanes are running on the exact
  head you review. Diff: `git diff origin/main..HEAD` — 27 files, +5398/−151.
- Read `producer-brief-stage-c.md`, `TASK-260906-1uf713_drafting-report.md` (increment 1) and
  `TASK-260906-1uf713_drafting-report-inc2.md` (increment 2, the one carrying the gate table, the
  mutant table and the stated bounds).
- Authority: curator-spec `550579d` — `protocol/environments.md` §1 (the `path` kind in full), §6 and
  §6.1, §9.1, §9.5, §9.6, §9.7, §12.1 and §12.2; `schemas/v1/manager-config-v2.schema.json` and
  `system-config-v2.schema.json` with their schema-case families and `vectors/manager-config-v2.json`;
  `profiles/manager.md` §1 and §12; `cli/curator.md` rows for `profile compose`, `env config`,
  `profile install <path>`, the takeover flag and `profile import`.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Method

Drive every claim through the production entry point — `run()` in `cmd/curator`, not a helper. Apply
your own narrowing mutants; do not accept the producer's mutant table on trust. A mutant that
*deletes* a gate proves less than one that *weakens* it to admit exactly one member, and stages (a)
and (b) both found delete-only negatives hiding real holes.

## Review dimensions

1. **Schema 2 and the locked subset.** Every §12.1 knob present with its exact name, value grammar and
   default. Then attack the §12.2 enforcement: `require_current_profile` refusing another profile in
   the machine scope, a locked `overlays_allowed: false` emptying every overlay list with the manager
   §1 warning, `isolation` lockable only toward `shared`, and the manager §1 credential rule that no
   key selecting or constraining credential material is lockable. Verify a schema-1 file stays valid
   and an unknown `schema_version` is rejected explicitly.

2. **Composition, the part most likely to be subtly wrong.** The four effective-weight rules must
   apply *in order*, each overriding the previous. Construct a closure where all four disagree and
   check the winner. Check `context_weight_conflict` is an error when direct requirers disagree and a
   *warning* when the root's `weights` map names the member; `context_weights_not_root`,
   `context_weights_duplicate` and `context_weight_unknown`; that an overlay repeating a closure name
   is `environment_composition_invalid`; that an overlay needing a skill version the root forbids is
   `context_range_conflict` and never a silent second copy; and that a machine overlay's weight
   outranks a manifest weight for a package that is both.

3. **The two precedence primitives.** `winner` and `placement` must be independent — changing one
   without the other must produce the emission order §5 fixes. Test all four combinations against the
   materialized bytes, not against an internal ordering function.

4. **The `path` kind.** Every `profile_source_invalid` condition, driven: symlink, hard link, special
   file, platform path collision, a `.git` below the root (and a root-level `.git` correctly
   *excluded* rather than refused), and a declaration carrying `range`, `tag`, `branch`, `revision` or
   `directory`. Verify the snapshot is genuinely immutable — edit the source directory after install
   and prove nothing changes. Verify the pin is the core §8 content hash and the manifest `version` is
   authoritative with no tag check.

5. **Onboarding and import.** The foreign-manager stop must be a real stop with an explicit choice,
   never a silent absorption. The heuristic must never block. The backup must happen before the *first*
   write, whether or not an import was requested. In the loss list, an absent surface is never a loss
   and a failed read always is — drive both. The consent gate must not be satisfiable from machine
   configuration. Reassembly: `name` `imported` unless supplied, `version` `1.0.0`, `weight` `0`, no
   `weights`; modules in ascending environment-identifier order; the CRLF/CR-to-LF normalization with
   exactly one trailing LF applied *only* at reassembly; `profile_import_name_taken` before any write;
   one `requires.skills` entry per mapping entry pinned by `revision` with
   `environment_import_skill_foreign`. Confirm the import writes nothing into any native home by
   itself. Confirm read-only commands never begin onboarding, never write a backup and never prompt.

6. **The surviving mutant.** The report declares M3 a survivor: collapsing absence into unreadable in
   `EnsureState` kills no test because `pathManifestDiag` precedes the store. Verify that reasoning
   rather than accept it — is the store branch genuinely unreachable except on a race, or is there an
   addressing mode that reaches it? Stage (a)'s F1 was exactly this shape: a gate bypassed by one
   addressing mode nobody tested.

7. **The gate table and the hosted lanes.** The report's table is itself under review: standalone
   commands, observed exit codes, both roots, and it states plainly that CI was not consulted because
   nothing was pushed. That is now stale — the head *is* pushed. Read `gh pr checks 61`; for any lane
   not green, extract failing cases from the run's uploaded evidence artifacts (`test-evidence-<os>`,
   `race-evidence-<os>`), not from `--log-failed`, which returns nothing useful here. A red lane on
   this head is blocking. **Run the two `test-gate.sh` lanes sequentially, never concurrently and
   never alongside a `-race` suite** — the producer lost two attempts to that contention over the
   machine-wide Go test lock and the resulting red is indistinguishable from a real regression.

8. **The stated bounds.** The report lists bounds including `requires` never naming a `path` source
   structurally, and the secondary fixed-home targets' divergent-file comparison having no probe.
   Judge whether each bound is honest and whether any hides a missing gate.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-1.md`, plus a probes archive if you build one.
Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at `to-review` with
`accept_cr` on the recorded revision, and state plainly whether every hosted lane was green on the
exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
