# Producer brief: implementation stage (a) — context packages, lock, store, monolithic materialization

## Where and what

- Repository `~/Developer/ReluxWorks/curator` (Go 1.25.5). Worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch `feat/agent-environments-stage-a`,
  base = curator main `__BASE__` (includes the byte-exact acquisition). First run
  `git submodule update --init --recursive`.
- Authority: curator-spec main `fd237ba` — `protocol/environments.md` revision 1.1 (§1 sources and
  §1.2 byte-exactness, §1.3 lock, §1.4 versions and ranges, §2 `agent-context.json`, §2.2 `agent-mcp.json`,
  §3 modules, §4 store, §5–§5.2/§5.4–§5.6 monolithic materialization and hash binding, §6 composition and
  weights, §8.1 modes, §8.2 marker, §9.1 install, §9.2 switching, §9.4 skills/migration, §12 status/GC),
  `schemas/v1/agent-context-v1.schema.json`, `agent-mcp-v1.schema.json`, `context-lock-v1.schema.json`,
  `agent-environment-marker-v1.schema.json`, `conformance/v1/vectors/` (`context-versions.json` or as named
  in §13, `environments.json` (v2 sets: monolithic-*, weights-winner-{higher,lower}-placement-{first,last}, referenced-*, system-prompt-composed, mcp-*), `context-detectors.json`, `snapshot-acquisition.json`) and
  `conformance/v1/expected/environments/*`; `profiles/manager.md` §12; Decision 0012 D1–D5, D7–D8.
  Curator's own architecture: `internal/manifest`, `skillspec`, `protocoljson`, `runtimestore`, `gitops`
  (object-database extraction), `snapshot`, `audit`, `hashing`, `marker`, `transaction`, `managerlock`,
  `scopes`, `adapters`, `config`, `interop` (conformance harness patterns), `crossconformance` (CCJ-1).
- Scope of this stage — exactly the epic's (a) list minus the landed acquisition fix:
  1. **Package kinds and manifests**: parse and validate `agent-context.json` (strict schema 1: name,
     version, `requires.{contexts,skills,mcp}` with ranges, `weights` root-only, modules with `path`,
     `environments`, `class`) and `agent-mcp.json` (schema 1: `server.transport` stdio|http, bare
     `command`, `args`, `https` `url` grammar, `env_names` grammar + manager-reserved exclusion,
     `environments`); diagnostics exactly as §2.1/§2.2 name them.
  2. **Versions and ranges** (§1.4): strict SemVer 2.0 `v`-tags without build metadata; npm-shaped ranges
     (caret incl. 0.x/0.0.x, tilde, comparators, x-ranges, partial coercion, `-0` upper bound, `||`),
     `latest` = `*`, hyphen ranges and `v` inside ranges rejected; prerelease rule; total order. Pass the
     version/range vectors.
  3. **Resolution and lock** (0012 D2/D3, §1.3): joint resolution of root + overlays + skills + MCP with
     downward re-selection and termination; `context_range_conflict`, `context_version_mismatch`;
     weights (manifest → agreeing direct-requirer edges → root map; `context_weight_conflict`,
     `context_weights_not_root`); lock canonical order (kind, name), pins `commit` | `state_sha256`,
     effective weights, requirer chains, overlay flag; CCJ-1 lock bytes and `lock_sha256`. Pass the
     resolution and lock vectors.
  4. **Package store** (§4): per-package commit-/state-keyed immutable entries under the manager home,
     reusing the `runtimestore` pattern; `git` kind through `gitops.Extract` (exact bytes), `local` kind for
     the builtin `default` profile; content hash per core §8.
  5. **Always-strict audit** on every member new to the lock (§9.1): existing `audit` pipeline plus the
     `context-secret-material` detector over context modules, `agent-context.json`, `agent-mcp.json`
     (`args`, `url` in scope), `CONTEXT.md`; the detector is unpinnable; scoped waivers `{pin, file, span,
     reason}` from machine config; `context-system-module-present` always-warn. Pass the detector vectors.
  6. **Monolithic materialization** (§5, §5.1, §5.2, §5.4, §5.6): the `curator-root-context-v2` header,
     chapter parts `## Context: <name> <version>` in weight order under both precedence primitives, the
     no-chapter case, zero-module output, part joining, platform-collision rule, hash binding; pass every
     v2 `expected/environments/*` monolithic set byte for byte (referenced form, system prompt, MCP files
     and managed homes are stage (b)).
  7. **`linked` switching with the M11 transactional shape** (§8.1, §9.2): `profile use` as one
     `transaction` Plan of per-entry targets (attempt every entry, per-adapter results, record the new
     current only when the whole scope materialized), versioned backups `.agent-environment-backup/<n>/`
     with retention, the environment marker (`agent-environment-marker-v1`) with `members`, `precedence`,
     `surfaces` (copied-surface reason for the claude_code root context), `linked`/`copied` modes for
     in-place homes; `profile install <url> --range|--tag|--revision [--use]`, `profile list`, `profile
     update`, `profile remove [--purge]`, `profile use [--clear]`, `profile sync`.
  8. **Global-scope migration** (§9.4): the builtin `local` `default` profile whose lock carries the
     migrated global skills; direct machine declarations write into the lock; skill-scope `update`/`upgrade`
     never move profile pins; `global update|upgrade` fetch-only under this capability.
  Out of scope here (stage b/c/d): managed homes and seeds, passthrough, read-only `env resolve` and the
  fragment, MCP channel files, `path` kind and onboarding import, composition CLI, manager-config schema 2,
  `curator run`, ax.

## Delivery shape

- Work in small signed commits on the branch (this is a code PR; landing is fast-forward of the reviewed
  head, so a linear series is fine — but each commit builds and tests green).
- Conformance subset: a `CURATOR_CONFORMANCE_ROOT`-driven test per vector family (in the
  `internal/interop` style with registered skip classes — read `.github/ci/skip-classes.tsv`,
  `platform-cases.tsv`, `ledger-consistency.sh` first; every new required case goes into the ledger; a
  skip needs a registered class), plus unit tests per package; the pinned rc.9 root lacks these vector
  files, so the tests skip there and run in the candidate lane — say so in the report.
- Gates: `go build ./...`, `go vet ./...`, `gofmt -l`, `go test -count=1 -timeout 30m ./...` (cmd/curator
  ~8 min), `-race` on the new packages; reproduce the platform-case gate locally as `ci.yml` does; the
  adapter suites' hosted x86 redness is a known, separate class.
- CLI surface exactly as `cli/curator.md` rows at `fd237ba`; README tools section updated.
- Do not push, tag, or open a PR; attach `TASK-260905-30zs8t_drafting-report.md` (package map, vector families
  passed with counts, gate outputs, skipped items with reasons); `task-board handoff TASK-260905-30zs8t --role
  developer`. Never write LOGBOOK.md or anything into the control root.
