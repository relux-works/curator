# Review brief — TASK-260917-16l2md (SPEC_PIN → v1.0.0-rc.12 with the E2/E4/S4 manager union), review round 1 (Change Request revision 3)

You are the independent reviewer of a curator change produced for
`TASK-260917-16l2md` (story `STORY-260917-3w3lvj`, conformance-pin promotion of
the 2026-09 security-audit remediation). Read, in this order:
`remediation-manager-producer-rules.md` (the SPEC_PIN exception applies to
this task only), the producer brief `TASK-260917-16l2md_brief.md`, the
producer results `TASK-260917-16l2md_results.md` (merge notes per candidate),
the gate-failure notes `TASK-260917-16l2md_gate-failure-rev1.md` / `-rev2.md`,
the published Change Request patch `TASK-260917-16l2md_change-request_rev3.patch`
(revisions 1–2 failed the hosted gate on Windows/Linux harness issues in the
E4 identity tests; revision 3 passed on all lanes at the new pin — see its
validation log), the three input candidates (`E2-candidate.patch`,
`E4-candidate.patch`, `S4-candidate.patch`, each with its own task's brief
and results on `TASK-260916-55g9dg`, `TASK-260916-3oh0u8`, `TASK-260910-gocke2`),
the E2 rev-2 review corrections (`TASK-260916-55g9dg_review-verdict-rev2.md`,
`_rework-rev3.md`), and the landed spec at the pin: curator-spec
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` — check it out read-only
(`git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree add /tmp/spec-rc12-review dced9b8317e0e8af79edf2d0539b32bd22b6c85b`;
`environments.md` §3/§5.5 (E2), §11/§11.1 (E4), §2.2/§2.3/§10.3 (S4), §12.1/§12.2,
the vector families `environments.json`, `manager-config-v2.json`,
`umbrella-provider-resolution.json`, `environments-env-passthrough.json`).

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260917-3w3lvj/worktree`
on branch `task-board/story/STORY-260917-3w3lvj` holds the exact candidate
(uncommitted revision-3 delta over `main`). Do not edit it and leave NO
files in it; build, test and probe in a disposable copy under `/tmp` with
`CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12-review/conformance/v1`.

## What to verify
1. **The pin move.** `SPEC_PIN` equals the rc.12 commit in every suite
   checkout of `.github/workflows/ci.yml` (no second pin, `release.yml`
   untouched); no vendored spec bytes; every `root-content` skip row the
   candidates added for a family the rc.12 root publishes was removed, and
   no skip class or ledger row was added to hide a vector family; the gate
   self-test still enforces the ledger.
2. **Union fidelity, per candidate.** For each of E2, E4, S4: every hunk of
   the input candidate is present in the union or is one of the merge
   rewrites the results name (`internal/config` knob tables, `EnvLockKey`
   arms, §12.2 transcription maps, CHANGELOG); nothing was dropped, nothing
   invented; the producer's own union checker is not a substitute — sample
   at least the config parser, the lockable-key set, the status rows and
   each candidate's vector-driving test yourself.
3. **Spec conformance, item by item, at the rc.12 text**:
   - E2: `transitive_system_modules` `drop|error` default `drop` with the
     warning, direct naming or a `system_module_waivers` entry admits,
     `context_system_module_transitive` under `error`, lockable only towards
     `error`, waivers not lockable, a `null` waiver list refused, posture row;
     the rev-2 review corrections (no `prunePostRevisionKnobs` workaround, root-
     content driver) stay closed.
   - E4: resolution from the manager install directory then
     `provider_directories`, never ambient `PATH`; revision A (shipped
     default) WARNS `subcommand_provider_outside_trust_roots` with the
     migration hint and keeps resolving, revision B refuses
     `subcommand_provider_untrusted`; manager-published/managed directories
     refused as before; resolved path in `env status`; Windows identity
     (8.3 short spelling, drive case, symlinks) compared by identity — run the
     identity tests' logic on this host and read the Windows-only ones.
   - S4: `passable_env_names` default `[]`, explicit `null` = unbounded,
     `mcp_package_allowlist_empty` warning at install/update/status,
     `s4-warn` shipped and `s4-enforce` implemented behind one option, the
     §2.3 surfacing rows.
4. **Vectors are driven at the new root**: `TestManagerConfigV2Vectors` and
   every family above execute with zero root-content skips; run
   `go build ./... && go vet ./... && gofmt -l internal cmd` and the narrow
   packages (`./internal/config/... ./internal/contextresolve/...
   ./internal/contextmaterialize/... ./internal/contextaudit/... ./internal/shell/...
   ./internal/envprofile/... ./cmd/curator/...`), `set -o pipefail`, exit codes
   quoted; at least three narrowing mutants, one per candidate (e.g. admit a
   transitive system module under `error`; resolve a provider from `PATH`;
   pass an unlisted env name), each caught by a committed test.
5. **Scope and hygiene**: no E1/R1-client/S6 content, no unrelated edits, no
   writes outside the worktree; CHANGELOG carries the three entries with
   their warn-first steps and the pin note.

## Verdict
Record `TASK-260917-16l2md_review-verdict-rev3.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260917-16l2md, revision=3, evidence=TASK-260917-16l2md_review-verdict-rev3.md)'`
or a changes-requested verdict routed with `set_status(TASK-260917-16l2md, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate. Remove `/tmp/spec-rc12-review` when done
(`git -C …/curator-spec worktree remove --force /tmp/spec-rc12-review`).
