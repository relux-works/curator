# Producer brief: split the manager-config schema-2 cases into their own vector family

## Why
curator-spec PR #41 (`draft/environments-manager-cli-1-1`, head `9af8af8`) fails the hosted
`Implementations` lane: the pinned Go manager's `internal/interop TestManagerConfigVectors` reads
`conformance/v1/vectors/manager-config.json` and cannot pass the new `schema2-*` cases (it implements
manager-config schema 1 only). The suite's contract is that a pinned implementation is observed passing the
cases it consumes, so schema-1 consumers must keep passing the schema-1 file byte for byte until stage (c)
lands manager-config schema 2 in curator.

## Where
Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, branch
`draft/environments-manager-cli-1-1`, head `9af8af8` (one signed commit past main `a68559b`). Work HERE, not
in the managed story worktree (its Change Request rev 1 is accepted and must stay untouched). Amend/squash so
the branch remains exactly ONE signed commit past `a68559b` (`git commit --amend -S`, or reset --soft + commit).

## The change
1. `conformance/v1/vectors/manager-config.json` MUST be byte-identical to `git show a68559b:conformance/v1/vectors/manager-config.json`
   (verify with `git diff a68559b -- conformance/v1/vectors/manager-config.json` → empty).
2. Every schema-2 case moves to a new file `conformance/v1/vectors/manager-config-v2.json` with the same identity
   header shape (`schema_version`, `protocol_version`, plus whatever the generator writes for schema 2), generated
   by `tools/generate-vectors` (adjust `manager_config.go`/`main.go`); `tools/validate.py` validates both files
   (schema-1 cases against schema 1, schema-2 cases against schema 2, the §12.1 knob/default cross-check against
   the v2 file), with `tools/test_validate.py` coverage; `conformance/README.md` names the new family;
   `conformance/v1/manifest.json` and the rc.9 pin regenerate.
3. Check `.github/ci/implementation-coverage.tsv` and `tools/implementation_coverage.py`: if a new family needs a
   ledger row marked as not yet consumed by the pins, add exactly that; if the ledger only lists consumed cases,
   leave it and say so.
4. Everything else in `9af8af8` (manager §12, cli rows, schema 2 + cases, COMPATIBILITY, CHANGELOG) stays; update the
   CHANGELOG entry sentence to name the new vector file.
5. Gates: `make validate`, `make regenerate-check` (venv `.temp/venv`), and reproduce the hosted lane locally:
   clone or reuse the pinned Go manager at the commit `release/1.0.0-rc.9.json` / `.github/workflows/implementations.yml`
   names (read the workflow for the exact pin and env var `CURATOR_CONFORMANCE_ROOT`), run
   `go test -count=1 ./internal/interop` against your worktree's `conformance/v1` root, and paste the result.
6. Push with `--force-with-lease` to `draft/environments-manager-cli-1-1` (this branch is not yet reviewed-landed;
   the PR updates in place), then `gh pr checks 41 --watch` until `Implementations` is green on all three OSes.
   Attach `TASK-260905-2tvae4_drafting-report.md` (commit + signature, file table, gate outputs, PR check summary) and
   `task-board handoff TASK-260905-2tvae4 --role developer`. Never write LOGBOOK.md or anything into the control root or the
   repository.
