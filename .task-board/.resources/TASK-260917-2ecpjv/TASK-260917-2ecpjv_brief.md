# Brief — TASK-260917-2ecpjv: qualify curator-spec v1.0.0-rc.12 (read-only)

Story `STORY-260917-3w3lvj` (conformance-pin rc.12 promotion), `EPIC-260910-2hw1xb`.
The sibling leaf `TASK-260917-16l2md` (the `SPEC_PIN` bump to `dced9b8` plus the
E2/E4/S4 manager union) is accepted and checkpointed; this task is the
release-qualification evidence gate that must exist before the pin lands.
Role: researcher. Read-only: no repository, board resource tree or checkout
outside your disposable directory may be mutated; no pushes, no tags.

## Do exactly this, quoting every command and its output in the outcome
1. Disposable checkout: `git clone -q git@github.com:relux-works/curator-spec.git /tmp/qual-rc12`
   (or `git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree add --detach /tmp/qual-rc12 v1.0.0-rc.12`
   if cloning is refused); never touch the main checkout's worktree.
2. Tag resolution: `git -C /tmp/qual-rc12 fetch -q origin tag v1.0.0-rc.12`;
   `git rev-parse v1.0.0-rc.12` (tag object) and `git rev-parse v1.0.0-rc.12^{commit}`
   — the commit MUST be `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`; `git cat-file -p v1.0.0-rc.12`
   (tagger identity, message); `git verify-tag v1.0.0-rc.12` — the signature must
   verify against a key in the repository's allowed signers file
   (`git config gpg.ssh.allowedSignersFile` of the main checkout; the
   release bot key `SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds`
   and the maintainer key are listed there). Quote the verification line verbatim.
3. Checkout `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` in the disposable
   directory and run, with `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`:
   `make validate` (exit code), `make regenerate-check` (exit code; the tree
   must stay clean: `git status --short` empty afterwards).
4. Record `sha256sum conformance/v1/manifest.json` (or `shasum -a 256`), the
   full content of `release/1.0.0-rc.9.json`, and the existence plus the
   named case families: `conformance/v1/vectors/manager-config-v2.json`,
   `vectors/environments*.json` (list them), `vectors/umbrella-provider-resolution.json`,
   `vectors/environments-env-passthrough.json`, `vectors/shell-hook-trust.json`,
   `vectors/registry-client.json` with a `page_boundary_cases` member (count
   the cases). Note explicitly which families the pin does NOT publish
   (e.g. `environments-source-signers.json` lands later at `684c9f1`).
5. Cross-check the curator side read-only: `git -C /Users/administrator/Developer/ReluxWorks/curator/curator show task-board/story/STORY-260917-3w3lvj:.github/workflows/ci.yml | grep -n SPEC_PIN`
   — the pinned value must be the same full commit.
6. Remove the disposable checkout/worktree.

## Outcome and handoff
Attach `TASK-260917-2ecpjv_qualification.md` with the transcripts, the
verdict (`qualified` / `not qualified` with the failing step), the manifest
digest, the pin content and the family inventory. Tick every checklist item
you satisfied (for the role-baseline items that do not apply to a read-only
qualification — tests, implementation — the orchestrator has checked them
with a basis note; leave them), then `task-board handoff TASK-260917-2ecpjv --role researcher`.
If the tag is absent or resolves elsewhere, record that as `not qualified`
and hand off anyway — do not wait, do not create the tag.
