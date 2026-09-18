# Review brief — TASK-260910-3u9t1e (R7 service: verify-backup implicit-key warning, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-3u9t1e` (story `STORY-260910-2xe3n2`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-3u9t1e_brief.md`, the producer results `TASK-260910-3u9t1e_results.md`,
the published Change Request patch `TASK-260910-3u9t1e_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and `profiles/registry-service.md` §6 (backup verification, operator checkpoint) at the pinned `47c3c8c`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-2xe3n2/worktree`
on branch `task-board/story/STORY-260910-2xe3n2` (HEAD = the checkpointed R8 leaf) holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Behaviour** (settled decision in the brief): `verify-backup` without `--public-key` still verifies but prints a prominent stderr warning AND records it in the structured output/exit summary ("keys resolved from the live home <path>; supply --public-key from an out-of-band copy for an independent check"); with `--public-key` nothing changes; the verdict and exit code are unchanged in both cases (warning, not refusal). Quote file:line; replay both invocations through the installed CLI in your disposable copy.
2. **Tests**: implicit resolution warns (stderr text + structured field), explicit key does not; verdict unchanged in both. Probe: a compromised-home scenario (rotate the live home key after the checkpoint) — the warning must appear and the verdict must reflect the live-home keys, exactly as documented.
3. **Story branch state**: the candidate sits on the Story branch carrying the checkpointed R5 (`d7f424c`) and R8 (`bd27d32`) leaves; the Change Request delta is only this leaf's change (compare the published patch with `git diff HEAD` in the worktree); hygiene: no results/logbook/build files anywhere in the candidate.
4. **Docs**: README/SECURITY document the out-of-band key expectation; CHANGELOG "R7" entry.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `47c3c8c` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-3u9t1e_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-3u9t1e, revision=1, evidence=TASK-260910-3u9t1e_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-3u9t1e, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
