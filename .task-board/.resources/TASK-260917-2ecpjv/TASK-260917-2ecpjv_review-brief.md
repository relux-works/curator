# Review brief — TASK-260917-2ecpjv (read-only qualification of curator-spec v1.0.0-rc.12), review round 1 (Change Request revision 1, empty delta)

You are the independent reviewer of a research/qualification task on the
curator board (story `STORY-260917-3w3lvj`, `EPIC-260910-2hw1xb`). Read the
brief `TASK-260917-2ecpjv_brief.md`, the outcome
`TASK-260917-2ecpjv_qualification.md` (verdict: qualified) and the handoff
note. The Change Request carries an EMPTY repository delta by design (nothing
is implemented here); the reviewed artifact is the qualification evidence.

## What to verify (all read-only; no repository mutation, no tags)
1. Independently re-derive every claim in a disposable checkout under
   `/tmp` (never the main checkouts): `git ls-remote --tags git@github.com:relux-works/curator-spec.git v1.0.0-rc.12`
   and the tag's commit = `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`;
   `git verify-tag v1.0.0-rc.12` against the allowed-signers file of
   `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
   (`git config gpg.ssh.allowedSignersFile`) — Good signature by the bot key
   `SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds`; tagger identity and
   message quoted.
2. At that commit: `make validate` and `make regenerate-check` exit 0 with the
   repo venv (`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin` on PATH), tree clean afterwards; the
   manifest sha256 equals `ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed`
   and equals the `manifest_sha256` inside `release/1.0.0-rc.9.json` at that
   commit.
3. Family inventory: the outcome's list of present and absent vector
   families matches the tree (spot-check `vectors/registry-client.json`
   `page_boundary_cases` count, `manager-config-v2.json`,
   `umbrella-provider-resolution.json`, `environments-env-passthrough.json`,
   `shell-hook-trust.json`; `environments-source-signers.json` absent).
4. The Story branch `task-board/story/STORY-260917-3w3lvj` of
   `/Users/administrator/Developer/ReluxWorks/curator/curator` pins
   `SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b` in
   `.github/workflows/ci.yml` (`git show <branch>:.github/workflows/ci.yml`).
5. The outcome mutated nothing (no leftover `/tmp/qual-rc12`, no changes in
   the main checkouts: `git status --short` clean in both, no new tags
   besides the one the orchestrator created).
Leave NO files anywhere but your disposable `/tmp` copy, and remove it.

## Verdict
Record `TASK-260917-2ecpjv_review-verdict-rev1.md` (task outcome) with the
per-item table and transcripts, then exactly one of
`task-board m 'accept_cr(TASK-260917-2ecpjv, revision=1, evidence=TASK-260917-2ecpjv_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260917-2ecpjv, status=to-dev)`
naming the failed step. Never accept on the outcome's own transcript alone.
