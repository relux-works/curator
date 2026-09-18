# Review brief — TASK-260910-s9jz1g (R6 service: passphrase-protected signing key, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-s9jz1g` (story `STORY-260910-9484i4`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-s9jz1g_brief.md`, the producer results `TASK-260910-s9jz1g_results.md`,
the published Change Request patch `TASK-260910-s9jz1g_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and `profiles/registry-service.md` §7 (key custody; the external key provider allowance) at the pinned `dced9b8`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-9484i4/worktree`
on branch `task-board/story/STORY-260910-9484i4` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Behaviour** (settled decisions in the brief): the single variable `CSK_REGISTRY_KEY_PASSPHRASE` drives encrypted PKCS8 (`BestAvailableEncryption`) on `genkey` and every rotation write, and decryption on every `load_key` path; unset → unchanged plain-PEM behaviour; encrypted PEM with the variable missing or wrong → fails closed at command/startup with ONE diagnostic naming the variable, no traceback, no partial start; plain PEM under a set variable → loads with an unencrypted-key warning; public key material, `key_id`, rotation semantics and the protocol unchanged. Quote file:line; confirm the passphrase never appears in argv, logs or error text.
2. **Provider seam**: key loading funnels through one seam in `keys.py` (file-backed default), documented in SECURITY.md as the KMS hook point, no external provider implemented.
3. **Tests**: encrypted round-trip (same public key), wrong/missing passphrase fail closed with the named diagnostic, plain PEM still loads, rotation with an encrypted key, `serve` startup with an encrypted key. Probe: an encrypted key with an empty-string passphrase in the variable; a PEM encrypted with a different passphrase; confirm the failure text carries no key material.
4. **Docs**: README (setup, passphrase rotation, encrypting an existing key), SECURITY.md (scope of protection, hook point), compose secret wiring, CHANGELOG "R6" entry.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `dced9b8` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-s9jz1g_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-s9jz1g, revision=1, evidence=TASK-260910-s9jz1g_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-s9jz1g, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
