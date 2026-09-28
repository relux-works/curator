# TASK-260921-3qcjsy gate rerun on the rev2 tree (2026-09-21, rework-1 verification run)

Role: developer. No tree changes in this run: the worktree is byte-identical
to the published rev2 change request, and the gate is green on my own reruns.

## Tree identity (no drift from published rev2)

- `git status` path set (160 paths: 155 modified + 5 new schema cases) is
  exactly the rev2 patch path set (diff clean).
- Whole-tree diff of the worktree against HEAD + rev2 patch applied in a
  scratch dir: clean except ignored build outputs (`.venv`, `__pycache__`)
  and the `subst.txt` git smudge artifact (git-clean, pre-existing).
- F3 check: `grep -n "not an adoption|adopting revision amends|stays open
  question" decisions/001[78]*.md` returns nothing (exit 1).
- N2 check: no file under `conformance/`, `release/`, or `schemas/` names
  `environment_credential_conflict` / `environment_credential_unsupported`.
- Spot-verified in-tree: 0017/0018 Status adopted with dated notes and full
  Adoption choices tables; §7.4 Credential record worded as the F-S1
  follow-up content with the schema-1 sentence; §8.2 strategy-only;
  codex `isolated` file-only (`auto` in `environment_isolated_unsupported`)
  in §7.4, manager §12.4, 0017 choice 4; non-empty-directory repair names
  `environment_credential_conflict`; 0018 choice 6 encoding owned by the
  launcher SPEC (F-L1), choice 7 marker set {CI, GITHUB_ACTIONS}.

## Gate reruns (all exit 0, own runs, sequential — never concurrent)

- `.venv/bin/python tools/validate.py` → `validated 62 schemas and
  1124 vector files`, exit 0.
- `python -B -m unittest` full 579-test set in 6 sequential shards, all OK:
  shard A 78 (implementation_coverage 36 + release_gate 32 + verify 10),
  shard B 119, shard C 112, shard D 122, shard E1 98, shard E2 50.
  All 27 test_validate classes plus the 4 small files, each exactly once.
  (One shard-B invocation without PYTHONPATH failed with loader errors only;
  rerun correctly, 119 OK — no test failed.)
- `go test -count=1 ./tools/...` → ok 1.031s, exit 0. First attempt was
  `signal: killed` at 111s and `gofmt` was SIGKILLed twice in the same
  window (transient macOS exec-stall under host load); green on retry.
- `go vet ./tools/...` clean; `gofmt -l tools/` empty on retry.
- Post-run `git status` path set unchanged (TREE-STABLE): the suite's
  transient corpus rewrites were restored byte-identical.

## Conclusion

Revision 2 (F1–F3 + N1–N4, no schema/vector/generator change) verified
unchanged in the worktree with a fully green gate. Ready for review.
