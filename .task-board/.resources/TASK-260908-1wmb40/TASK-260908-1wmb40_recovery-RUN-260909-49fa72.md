# TASK-260908-1wmb40 recovery — fresh landed export + authorized rework (RUN-260909-49fa72)

Role: developer (implementer), Muse Spark xhigh. Story STORY-260908-v16gn5.
Old CR1 acceptance (rev1, tree 0361a3dfe25405da63ed9e70f886ba25880701db) stays immutable history; this run makes no source redevelopment and transfers no acceptance.

## Fresh authority (observed, not assumed)

- Protected default freshly observed via prepare/start commands: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`, tree `ff61be4a8bd43fa4ffb179d31aa38e41891d4313`.
- `remote_url ssh://git@github.com/relux-works/curator-agent-launcher`, `protected_ref refs/heads/main`, `author Ivan Oparin <ivan@relux.works>`, `signature_status G`.
- Historical PR9 SHA `84747c3` is NOT used; current main also carries systemprompt + execution.
- Installed CLI: `task-board version 0.24.3-332-gac72ed99 (commit ac72ed99)`. Spec: `.specs/separate-owner-completion.md` from `/Users/iv/Developer/ReluxWorks/.worktrees/STORY-260909-3ue5iq-delivery` (changed-tree recovery sequence).

## Exact sequence (this run)

1. `task-board m 'set_status(TASK-260908-1wmb40, status=integrating)'` — exit 0 (already integrating).
2. `task-board --no-update-check worktree prepare-landed-review STORY-260908-v16gn5 --cr TASK-260908-1wmb40 --revision 1 --landed-commit 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 > .../RUN-260909-49fa72/TASK-260908-1wmb40_prepare-landed-review.json` — exit 0, stderr 0 bytes.
3. Attach exact export: `task-board --no-update-check resource add TASK-260908-1wmb40 <json> --name TASK-260908-1wmb40_prepare-landed-review-3ff66a9.json --type outcome` — exit 0.
4. `task-board --no-update-check worktree start-landed-rework STORY-260908-v16gn5 --cr TASK-260908-1wmb40 --revision 1 --landed-commit 3ff66a9... --export <json>` — exit 0, stderr empty; task `integrating` → `to-dev` via privileged path. No refusal to record.
5. JSON-decode `update_patch` (declared `update_sha256 e7d9b0af01ea7937c74356fdd7e268a2cae2edd8d19208f752fb866dac5add01`, recomputed match true), wrote `TASK-260908-1wmb40_update.patch` (110511 bytes).
6. `git apply --check <patch>` in managed worktree — exit 0. `git apply <patch>` — exit 0. Branch NOT moved/committed (HEAD stays `18aeaed9af7dc5ffbe6cc79a4731a852fbb716da`).
7. Candidate-tree verify via temp index (`GIT_INDEX_FILE` + `read-tree HEAD` + `add -A` + `write-tree`, temp index removed): `ff61be4a8bd43fa4ffb179d31aa38e41891d4313` — MATCHES `observed_landing.tree_oid`. Real index/branch untouched.

## Export provenance

- Export JSON bytes 269100, sha256 `08e93a55606f3dcaf6d11e731458d298a474cc063c740abada291f1d2709d72d`.
- `comparison_base_oid 18aeaed9af7dc5ffbe6cc79a4731a852fbb716da`, `previous_tree_oid 0361a3d...` (same as CR1 record).
- `review_sha256 071f8bd5bc8a9f93d2a811fc4215fc9d3a46f46fda337ad7d387eec1af487b53` (137338 bytes, recomputed match), `update_sha256 e7d9b0af...` (110511 bytes, recomputed match).
- `changed_paths` (21): `.scripts/composition-mutants.py`, `.scripts/execution-mutants.py`, `.scripts/systemprompt-mutants.py`, `README.md`, `SPEC.md`, `go.mod`, `go.sum`, `internal/axconfig/*` (3), `internal/composition/*` (3), `internal/execution/*` (6 incl testdata), `internal/systemprompt/*` (2).
- `update_patch` file list (15): execution-mutants, systemprompt-mutants, README, SPEC, axconfig (3), execution (6), systemprompt (2). It does NOT touch `go.mod`, `go.sum`, `.scripts/composition-mutants.py`, `internal/composition/*`.

## Byte-identical composition preservation

Pre/post sha256 unchanged (no source redevelopment; upstream bytes only carried):

- `.scripts/composition-mutants.py d44ad2197018850a608b0c34f26bef852c2a6ac994eec9ad51c7f74147003c78`
- `internal/composition/composition.go e53003ce5aa27b8412d9c7323a43aa7ac7808fa313bc0db7aaefa2a898dfe1a2`
- `internal/composition/composition_test.go ef28f326adef93519c127095dd2168d944eea195b11a4040aecbbe6b9ee1069f`
- `internal/composition/probe.go 0f45362923c8efcfe9308e4858421d306ae6d62599db06471d999ae64593187a`
- `go.mod f5a91ac64b7485254be556e06f35839a7a57f8bd4dae05ba7e9943944a8891a8`, `go.sum 61692218a85029aa4c33b388e1d816ff643c35d843d7432c8c9e5892ddca9807`
- Post-apply `git status`: `M README.md`, `M SPEC.md`, `M go.mod`, untracked `.scripts/composition-mutants.py`, `.scripts/execution-mutants.py`, `.scripts/systemprompt-mutants.py`, `go.sum`, `internal/axconfig/`, `internal/composition/`, `internal/execution/`, `internal/systemprompt/`. `.temp/` scratch ignored, excluded from candidate tree.

## Bounds and next step

- No manual `make check`/mutants run this turn per override; runtime handoff validation owns new-tree validation.
- No installs/restarts, hosted CI, tags/releases, ax/model launches, LOGBOOK/control-root/private-record writes, PR edits, or board-history synthesis.
- Retained obligations (unchanged): full pipeline wiring = execution/plan stories; no native Pi admission claimed before its operator tag; execution must invoke `Value.CheckLaunchBoundary` immediately before BOTH direct and ax process creation; both modes must print `Warnings` names-only to stderr; TOCTOU window remains.
- Workspace: `/Users/iv/Developer/ReluxWorks/.worktrees/launcher-control/.temp/STORY-260908-v16gn5/worktree`, common dir `/Users/iv/Developer/ReluxWorks/curator-agent-launcher/.git`. Candidate left UNCOMMITTED for handoff snapshot; new independent Astra review follows; no Complete until new revision accepted.
