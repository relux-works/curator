# TASK-260901-2pho68 — env-manager notes

Manager-profile and CLI sections for the agent-environments capability.

## Where the work lives

- Repo: `~/Developer/ReluxWorks/curator-spec`
- Worktree: `~/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`
- Branch: `draft/environments-manager-profile`, base `origin/main` at
  `c3b29b1f7f37829fd4d0c50b2023efa2feb4c615` (exactly the required base).
- Commit: `6697c1e1a7218cd0b7fc270226bba8e59d53c7b9`, SSH-signed, verifies `G`
  against `maintainers.allowed_signers` (principal `oparin@me.com`).
- Not pushed, not tagged, not merged (per brief).

## Sections added

`profiles/manager.md` — new top-level **§12 "Agent-environments manager
profile"**, following the §11 precedent (extends §1–10, restates no byte rule;
`protocol/environments.md` cited as the sole normative source throughout):

- 12.1 Environment adapter registry — four adapters, forms/defaults +
  fallback, opencode XDG seed-link maintenance, secondary fixed-home targets
  with auto|off|explicit participation and probing, shadowing-path warnings.
  Diagnostics table referencing environments §5.7/§7.7 codes.
- 12.2 Materialization modes and the environment marker — three modes and
  mode defaults, marker-as-ledger, backups, drift, absence-vs-unreadable.
  Diagnostics table referencing environments §8.5.
- 12.3 Profile lifecycle — install/activation-without-magic, `use` under the
  manager-home lock (journaled per §2.5), scoped switching/split-brain
  visibility, profile-scoped skills + `default` migration + `profile sync`,
  revision-1 onboarding (detect / foreign-manager stop / replace notice /
  backup-always / takeover), `path`-kind rejection. Diagnostics table.
- 12.4 Credential passthrough — per-platform closed sets by reference,
  shared/isolated, exclusions from hashes/store/audit, no credential writes.
- 12.5 `env resolve` — pure resolution, verify-and-repair with the
  no-adoption rule tied to §2.6/§11.9, profile-influence boundary.
  Diagnostics table.
- 12.6 Profile audit — always-strict, `context-secret-material` detector
  class over the unchanged §7 pipeline, prompt-material review note.
- 12.7 Status and GC — `profile list`/`env status` under the §10 read-only
  discipline, row currency, GC live roots extended with profile store
  entries, marker-referenced homes/surface sets, and journal references.

Also a short pointer paragraph appended to **§7 Source audit policy** naming
the always-strict profile audit and the `context-secret-material` class, so
§7 readers land on §12.6 and environments §9.1.

`cli/curator.md`:

- Command-table rows: `profile install|list|use|sync`, `env resolve`,
  `env status`; the `global ...` row now carries
  `[--profile <name>|--all-profiles]`.
- New "Environment profiles" section: umbrella note (`curator run` /
  `curator session` dispatch to `curator-<name>` on `PATH`, nothing installed
  implicitly) plus a usage example block.

## Validation

All three `make validate` components run directly, real exit codes:

- `tools/validate.py` (schemas, vectors, markdown links): exit 0 —
  "validated 53 schemas and 691 vector files".
- `python -B -m unittest discover -s tools -p 'test_*.py'`: exit 0, 134 tests.
- `go test ./tools/...`: exit 0.

System `python3` lacks `jsonschema`; a task-scoped venv at
`<worktree>/.temp/venv` from `requirements-dev.txt` was used for the Python
components (make itself would need that venv on PATH).

## Normative gap observed (filed, not fixed)

`protocol/environments.md` §1 says a `git` profile declaration "carries
exactly one of `tag`, `branch`, or `revision`", and a directly installed
profile may track a branch — but §9.1 gives `profile install <git-url>` with
no statement of how the operator expresses that ref choice at install time
(flag, default branch, config?), nor what the default is when none is given.
The declaration shape exists; the operator-facing declaration mechanism for
direct installs is unspecified. Left untouched per the brief (protocol prose
belongs to the sibling story); the CLI row was kept generic
(`profile install <git-url> [--use [name]]`) so it does not invent one.

## Out of scope, untouched

`schemas/`, `conformance/`, `protocol/`, `CHANGELOG.md`; no push, no tag.
