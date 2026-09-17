# Brief — TASK-260916-1x0ogh: spec rule for provider resolution trust roots (E4)

Story `STORY-260916-2otjbn` (umbrella-provider-trust-roots), wave 1. This story
blocks proposal 0016 / `path_prepend`.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`.
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer (technical writer) of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E4 and Appendix B (E4 confirmed:
`cmd/curator/umbrella.go:30-63` resolves `curator-<name>` with `exec.LookPath` on
the ambient `PATH`; only manager-published directories are refused). Composed
with S6, a project `.agents/env.sh` can plant `curator-run` on `PATH`.
`protocol/environments.md` §11 "Umbrella subcommand discovery" and §11.1 are the
text to revise; §12.1 is the knob table.

## Settled decisions (do not reopen)
- Trust roots = (a) the directory of the running manager executable (its
  install directory, resolved after symlinks) and (b) an explicit machine-config
  list of provider directories — a new closed §12.1 knob (working name
  `provider_directories`, list of absolute paths, default empty). Ambient
  `PATH` is no longer a trust root. Keep every existing §11 rule (identifier
  grammar, implemented-subcommand-wins, no provider registry, no implicit
  install, profile/marker/fragment data never influence dispatch, refusal of
  manager-published directories).
- Name the S6-injected `PATH` case explicitly as the attack the rule closes.
- Posture: `env status` names the resolved provider path for each discovered
  `curator-<name>` (or that it is missing / untrusted).
- Warn-first rollout (impact row "E4 trust roots"): revision A keeps resolving
  on `PATH` but warns when the resolved executable lies outside the trust
  roots, with the migration hint to list its directory in `provider_directories`
  (e.g. a `curator-run` in `/usr/local/bin`); revision B refuses it with
  `subcommand_provider_untrusted` (existing diagnostic, extended condition) —
  specify both explicitly.

## Deliverable
1. §11 rewritten: resolution order (install directory first, then the listed
   directories in order; first match wins; a match in a manager-published or
   managed directory is refused as today), the ambient-`PATH` exclusion, the
   printed resolved path, the two rollout profiles.
2. §11.1 diagnostics table: extend `subcommand_provider_untrusted` (outside
   trust roots, naming the path and the trust roots consulted); add the
   revision-A warning (`subcommand_provider_outside_trust_roots` or the
   spelling the tables favour); keep `subcommand_provider_missing`.
3. §12.1 knob row `provider_directories` (values, default, section); decide and
   state whether it is §12.2-lockable (recommendation: lockable — a locked list
   is a fleet policy) and add it to the lockable set if so.
4. Conformance vectors (§13 "conformance surfaces" and `conformance/v1/`):
   provider in install dir → resolved; in a listed dir → resolved; only on
   `PATH` → revision A warning / revision B refusal; in a manager-published
   dir → refused; `provider_directories` schema case in the machine-config
   schema if one exists under `schemas/v1/` (follow the existing pattern).
5. `CHANGELOG.md` Unreleased entry "E4: …" naming the two rollout steps and
   the 0016 dependency.

## Out of scope
Implementation (`TASK-260916-3oh0u8` manager, `TASK-260916-16ys92` launcher
prints the resolved path), proposal 0016 itself, the launcher SPEC (lives in
curator-agent-launcher).

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-1x0ogh_change-request_rev1.patch` and `TASK-260916-1x0ogh_evidence.md`
(with the `make validate` transcript), then
`task-board handoff TASK-260916-1x0ogh --role doc-writer`.
