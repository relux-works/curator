# Approved source contract: producer and reviewer assignment

Task: TASK-260910-1xph2y. Story: STORY-260910-8fv3s5.
Repository control root: /Users/iv/Developer/ReluxWorks/curator-spec.
Initial verified main/origin/main: d019f0e7179520b5c8dcde321c4fe51e04552f58.

## Authorization and delivery boundary

The user approved the detailed source design and now requests specification
changes through a producer/reviewer/rework cycle. BOTH roles must use Codex
gpt-6-astra with medium reasoning. The user explicitly says NO COMMITS YET.
This overrides generic automatic commit/PR/landing instructions for this task.
Do not commit, tag, push, open a PR, checkpoint, integrate or run any command
that creates a commit. CR publication through a temporary index/tree snapshot
and independent reviewer acceptance are permitted. Leave the accepted candidate
uncommitted in the managed Story worktree. Acceptance may remain `integrating`
on the board because commit-based delivery is expressly deferred; do not forge
a `done` transition or commit acknowledgement.

Work only in the execution root assigned by task-board. Preserve the control
checkout, adjacent implementation checkout and all unrelated task state. Do not
modify task-board.config.json or other workflow configuration to bypass gates.
Never replace or install the task-board runtime from this session.

## Role-specific work

Producer (`doc-writer`): inspect current normative conventions and relevant
validation entrypoints, then write the approved specification changes, schemas,
examples and proportionate positive/negative conformance coverage. Choose
routine syntactic details consistently. Do not expand unresolved product scope.
Attach a task-scoped outcome with exact changed paths, validation commands,
results and any deferred decisions; hand off with `task-board handoff
TASK-260910-1xph2y --role doc-writer`. Publish an uncommitted Change Request.

Reviewer: independently inspect the actual current CR diff and affected
contracts, run relevant checks and adversarial cases, and record precise
findings. Accept only the current candidate after resolving all material
contradictions, then stop. On findings, use the normal request-changes lifecycle
with actionable path-specific evidence; the parent will route producer rework
and a new independent review. Never repair normative files yourself during
review. Never checkpoint or integrate an accepted CR.

## Design inputs (read these; previous proposal status is superseded here)

All paths below are read-only design inputs. They live in the existing board
flow and were checked against current specs/implementation before approval.

- /Users/iv/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260910-16vtxi/skillfile-compatibility-clarification.md
- /Users/iv/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260910-16vtxi/skillfile-source-examples.md
- /Users/iv/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260910-16vtxi/local-source-runtime-boundaries.md
- /Users/iv/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260910-3du5nd/transport-neutral-repository-resolution.md

The broader research and UNRESOLVED_QUESTIONS resources on TASK-260910-16vtxi
are background only. Do not implement their unrelated topics.

## In-scope contract

### Skillfile evolution and source selection

- Keep Skillfile v1 unchanged, including the meaning of its legacy `source`
  field (a path under manager-configured source root, not a new local path).
- Add a versioned project-manifest extension, with an optional `sources` table.
  The skill remains the dependency, audit, command-owner and installation unit.
  A source declaration is acquisition configuration; it is not a new bundle
  package type and does not require rewriting each skill manifest.
- Retain disjoint legacy individual entries inside the new version with their
  old semantics. New individual selectors use `name`, `from`, `directory`;
  collections use `from`, `directory`, `include` and optional `exclude`.
  Reject ambiguous mixing of acquisition/ref and source-selector forms.
- A source is a filesystem path (relative to the declaring Skillfile directory
  or absolute) or a Git source with exactly one tag/branch/revision. Preserve
  root-only branch admission; do not silently enable floating transitive refs.
- `directory: "."` explicitly selects the source root. Other selectors are
  portable source-contained paths. Do not blindly inherit a schema pattern
  that accidentally rejects `.` or admits escaping paths.
- Collections select immediate child skill directories using literal folder
  names or `*`, with explicit exclusions. No recursive `**` in the initial
  contract. Validate each SKILL.md name, required metadata and collisions.
  Do not silently rename skills or support simultaneous versions under one
  installed name. Define errors for missing explicit members and invalid
  discovered candidates, deterministic ordering and frozen membership.
- Define how resolution freezes Git commit, local snapshot content and expanded
  member set; mutations happen through explicit update/refresh, not implicit
  per-launch rescans. Keep machine-specific path state separate where needed
  for portable lock identity. Do not invent a general semver solver.

### Local package inputs and complete installation behavior

- Authored `agents/skills/<name>` is distinct from installed
  `.agents/skills/<name>`. Adapter directories and `.agents/bin` are managed
  outputs too. A broad source alias `path: "."` is legal; evaluate selected
  packages/effective inputs instead of rejecting the whole project root.
- Resolve physical path equivalence including symlinks and filesystem casing;
  reject sources inside managed outputs and destination writes that overwrite
  inputs. Recheck at the write boundary. Deterministically prune generated
  outputs before snapshots/traversal. Root packages require explicitly disjoint
  effective inputs; reject when separation cannot be established. Preserve
  existing unmanaged-path conflict protections.
- Local acquisition must feed the complete existing package pipeline, not just
  copy SKILL.md. Preserve declared runtime_roots, commands, supported builds,
  capabilities, dependency closure, readiness, runtime store and shims.
- Script runtime goes to the protected store and compiled artifacts to the
  immutable build cache. `.agents/bin` exposes declared commands. Existing
  script/context eligibility and runtime/build projection exclusions continue.
- Local directories use admitted filesystem bytes, including dirty/untracked
  inputs, even if `.git` is present. Freeze those bytes so mutation cannot
  switch source between audit, build and installation.
- Local package identity must cover every admitted context, runtime and build
  input and be distinct from a Git commit identity. A script-only edit must
  invalidate installed runtime even when SKILL.md is unchanged. Reconcile all
  affected marker, receipt, audit and cache references with a real versioned
  contract; never fake a Git commit or weaken existing audit/assurance gates.
- Preserve closed build drivers, toolchain requirements, dependency trust and
  no arbitrary package-provided install-hook execution. This does not add a
  prebuilt CLI distribution route or compiler bootstrap feature.

### Separately scoped repository transport amendment

- Separate stable repository identity and pinned content from connection URL.
  The existing host/path canonicalization is the starting point.
- Support supported URL declarations and an explicitly distinguishable logical
  identity form; avoid ambiguity with relative local paths. Machine policy
  selects permitted SSH/HTTPS endpoints and operator-owned authentication.
- The same checked-in declaration/locked content can use SSH on a workstation
  and HTTPS in CI without changing identity. No credentials in shared manifests,
  lockfiles, receipts, logs or compiler environments.
- Define deterministic bounded endpoint preference/fallback and explicit
  endpoint pinning. Alternate transport requires operator policy and an
  eligible availability/authentication failure. TLS/host-key validation,
  identity mismatch, integrity, revision and audit failures fail closed.
- Preserve the stricter external-build transport graph, SSH wrapper and
  credential broker. User Git/SSH configuration is usable only through admitted
  machine policy; no arbitrary inheritance, helpers, ProxyCommand or untrusted
  remapping. Do not broaden current endpoint grammar (ports/mirrors/aliases)
  merely because a generalized resolver could conceivably support it.
- Keep this amendment easy to identify separately from local-source changes.
  Record genuinely undecided advanced mappings in UNRESOLVED_QUESTIONS.md.

## Explicit exclusions

No manager implementation changes. No broad new rules/knowledge type system,
AGENTS.md/CLAUDE.md composition, MCP endpoint import/wiring, plugin system,
binary CLI distribution, registry/signature redesign, native notarization or
toolchain-family reconciliation. Link these as future work only if needed to
prevent the new contract from falsely claiming support.

## Repository integration and validation

- English normative prose and docs. Update canonical contracts rather than
  leaving the approved behavior only in an informal proposal appendix.
- Keep prose, versioned JSON Schemas, examples, compatibility documentation and
  relevant conformance vectors consistent. Follow existing namespace/release
  conventions; do not rewrite frozen historical release artifacts or claim a
  published release from an uncommitted candidate.
- Add a concise author/operator guide with the agreed source examples and
  authored-versus-managed directory tree. Keep README navigation current.
- Inspect tools/validate.py, tools/test_validate.py and conformance/README.md
  for existing checks. Use proportionate meaningful validation, including
  negative ambiguous forms, selection/path escape/overlap, snapshot mutation,
  runtime-only changes and forbidden transport fallback where representable.
  State clearly which checks validate schemas/spec vectors versus a manager
  implementation. Do not claim local v2 installation was executed.
- Store scratch and logs under `.temp/` and attach task-scoped reviewable
  outcomes through task-board resource CRUD. Do not directly edit board files.
- Preserve the no-commit instruction through every rework cycle and handoff.
