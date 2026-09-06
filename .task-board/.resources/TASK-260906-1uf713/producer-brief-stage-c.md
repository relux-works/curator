# Producer brief: implementation stage (c) — composition, the `path` kind, onboarding import, config schema 2 and its CLI

## Where and what

- Repository `~/Developer/ReluxWorks/curator` (Go 1.25.5). Worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch `feat/agent-environments-stage-c`,
  base = curator main `b056e5dae73be4dc92f2a948992f0941d7283f89` (stage (b) landed). First run
  `git submodule update --init --recursive`.
- Authority: curator-spec main `550579d12ea6d5daeaeabdae69933de87ba28e4b` — `protocol/environments.md` revision 1.1: §1 (the `path`
  source kind in full, including every `profile_source_invalid` condition), §6 and §6.1 (overlays as
  closure members, the four effective-weight rules in order, the two precedence primitives,
  `overlays_allowed`), §9.1 (installation through the `path` pipeline), §9.5 (onboarding inventory,
  the foreign-manager stop and its heuristic, the notice, the always-backup step, the classification
  offer), §9.6 (detected surfaces, lossless/lossy classification, the loss list, the consent gate,
  reassembly and its normalization, `environment_import_skill_foreign`), §9.7, §12.1 (every knob) and
  §12.2 (the lockable subset); `schemas/v1/manager-config-v2.schema.json` and
  `system-config-v2.schema.json` with their `conformance/v1/schema-cases/` families and the
  `vectors/manager-config-v2.json` family; `profiles/manager.md` §1 (`locked` keys and the
  system-file warning) and §12; `cli/curator.md` rows for `profile compose`, `env config`,
  `profile install <path>`, and the takeover and import rows that TASK-260906-1hn93j adds.
- Curator's own code: everything stages (a) and (b) added, plus `internal/config`, `marker`,
  `transaction`, `managerlock`, `scopes`, `adapters`, `audit`, `interop`.

## Scope

1. **Machine configuration schema 2**: read `manager-config` schema 2 and `system-config` schema 2,
   every §12.1 knob with its exact name, value grammar and default; the §12.2 lockable subset
   enforced through the manager §1 `locked` machinery, including `require_current_profile` making
   `profile use` of another profile a configuration error, `overlays_allowed: false` emptying every
   overlay list with the manager §1 warning, and `isolation` lockable only toward `shared`. A schema-1
   file stays valid; an unknown `schema_version` is rejected explicitly.
2. **Composition**: overlays declared per profile in machine configuration join the closure beside the
   root and resolve jointly with it; the lock flags them `overlay`; a repeated closure name is
   `environment_composition_invalid`; an unreadable or uninstallable overlay source fails with that
   source's §1.1 diagnostic. Effective weight by the four §6 rules in order, with the
   `context_weight_conflict` error-versus-warning split the third rule defines, `context_weights_not_root`,
   `context_weights_duplicate` and `context_weight_unknown`. The two precedence primitives drive §5
   emission order independently of each other.
3. **The `path` source kind**: install copies the directory tree into the profile store as an
   immutable snapshot, never read again; symbolic link, hard link, special file and platform path
   collision are `profile_source_invalid`; a root-level `.git` is excluded and a `.git` below the root
   is `profile_source_invalid`; the pin is the core §8 content hash of the snapshot, a state hash;
   the manifest `version` is authoritative and no tag check applies; `range`, `tag`, `branch`,
   `revision` or `directory` on a `path` declaration is `profile_source_invalid`; a `path` package is a
   root or an overlay and never appears in a `requires`.
4. **Onboarding and import**: the §9.5 inventory per adapter and participating target; the
   foreign-manager stop with its explicit abort-or-take-over choice; the dotfile-manager heuristic as a
   non-blocking `environment_foreign_manager_suspected` warning over the closed documented location
   list; the notice before any write; the backup of every file the operation will replace, before the
   first write, subject to `environment_backup_exists`; the §9.6 classification with its loss list,
   where an absent surface is never a loss and a failed read always is; the consent gate, with machine
   configuration unable to pre-record consent; reassembly exactly as §9.6 spells it, including the
   CRLF/CR-to-LF normalization with exactly one trailing LF applied only at reassembly, the ascending
   environment-identifier module order, `profile_import_name_taken` before any write, and one
   `requires.skills` entry per mapping entry pinned by `revision` with `environment_import_skill_foreign`;
   then installation through the ordinary `path` pipeline with always-strict audit. Read-only commands
   never begin onboarding, never write a backup and never prompt. The import writes nothing into any
   native home.
5. **CLI**: `profile compose add|remove|list`, `env config show|set|unset`, `profile install <path>`,
   and the takeover and import rows, exactly as `cli/curator.md` spells them at `550579d12ea6d5daeaeabdae69933de87ba28e4b`.

Out of scope: ax integration (stage (d)).

## Conformance subset

Every `manager-config-v2` and `system-config-v2` schema-case family, and the composition, `path`-kind
and import cases the environments vectors publish, through `CURATOR_CONFORMANCE_ROOT`. Register the
root artefacts your new packages read unguarded in `.github/ci/root-artifacts.tsv` so the default lane
defers them against the rc.9 pin and the candidate lane fails closed on a root that stops serving them
— a tolerated `root-content` skip is NOT acceptable for a family this stage introduces, because that
class is `allow` in every lane. Every new required case goes into `.github/ci/platform-cases.tsv` with
`root-unset` tolerance where the package is deferrable.

## Delivery

Small signed commits, each building and testing green. Gates: `go build ./...`, `go vet ./...`,
`gofmt -l`, `golangci-lint run ./...`, `go test -count=1 -race` on the touched packages, the vector
families through `CURATOR_CONFORMANCE_ROOT`, `bash .github/ci/gate-selftest.sh`, the platform-case gate
for the three GOOS values as `ci.yml` runs it, `bash .github/ci/ledger-consistency.sh`, and
`go test -count=1 -timeout 30m ./cmd/curator` once at the end. Additionally reproduce both CI lanes
locally: `test-gate.sh` against a materialized rc.9 root (`git archive` the `SPEC_PIN` commit's
`conformance/v1`) and `test-gate.sh` with `CI_REQUIRE_FULL_ROOT=1` against a curator-spec main root.
Quote both. Do not push, tag, or open a PR. Attach `TASK-260906-1uf713_drafting-report.md`; `task-board handoff
TASK-260906-1uf713 --role developer`. Never write LOGBOOK.md or anything into the control root.

## Facts you must not re-derive

The verification sprint (board resource `TASK-260905-3jq1so_verification-sprint.md`) and the stage (a)
and (b) reports are settled. Do not re-probe adapter behaviour; label anything new as docs-confidence.

## Lessons stages (a) and (b) paid for — do not re-pay them

- **The pinned root is not curator-spec main.** `ci.yml` pins `SPEC_PIN` at a released revision that
  predates the environments artifacts. A test that reads a conformance family unguarded fails every
  hosted lane while your local run against curator-spec main is green. Register every root artifact
  your new packages read in `.github/ci/root-artifacts.tsv` and give the ledger rows a `root-unset`
  tolerance. Reproduce both lanes locally before you hand off: materialize the pin with
  `git -C <curator-spec> archive <SPEC_PIN commit> conformance/v1 | tar -x -C <scratch>` and run
  `test-gate.sh` against it, then against curator-spec main with `CI_REQUIRE_FULL_ROOT=1`.
- **No Unix lane sees a Windows path bug.** Stage (a) shipped native Windows paths into git config
  values, where a backslash is an escape; stage (b) fed POSIX literals to `filepath.IsAbs`, which
  needs a volume name on Windows. Both were fixture defects invisible to every local run. Build every
  path fixture with `filepath.Join` over a platform-absolute base, and say in the report what you
  swept for and what you found.
- **A gate that no production path calls is not a gate.** Stage (a) F9 and stage (b) M1 were both
  helpers that existed, compiled, returned the right value when called directly by a test — and were
  never reached from `run()`. Drive every new refusal through the production entry point, and prove
  it with a mutant that *narrows* the gate rather than deleting it.
- **The report's gate table is itself reviewed.** Every row: the exact command as a standalone
  process and its observed exit code. Root-sensitive gates appear twice, once per root. A claim about
  hosted CI carries `gh pr checks` output verbatim or says plainly that CI was not consulted. A gate
  you did not run is listed as not run — that is an acceptable answer; an unread gate reported green
  is not.
