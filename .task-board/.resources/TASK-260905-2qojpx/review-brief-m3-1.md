# Review brief: snapshot acquisition byte-exactness (review M3)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact`,
  branch `draft/snapshot-byte-exactness`, head `d85c719`, base `b4f29cd` (= main).
- Read the producer brief `producer-brief-m3.md` (precondition on TASK-260905-2qojpx)
  and the producer's drafting report (outcome resource). The brief is the acceptance bar.
- Diff to review: `git diff b4f29cd..d85c719` in the worktree.

## Review dimensions
1. **The rule**: environments.md states that a snapshot from a commit reproduces exact
   committed blob bytes; working-tree conversion (`core.autocrlf`, `text`/`eol`,
   clean/smudge, `ident`) and attribute-driven archive processing (`export-subst`,
   `export-ignore`) MUST NOT alter, add, or omit entries; `path` snapshots are not
   normalized either; the consequence for hash-bound identities is stated; the text does
   not amend frozen `protocol/core.md` and does not contradict core §6.2/§6.5/§8. Attack
   the wording for a reading under which `git archive` with `core.autocrlf=false` and no
   attributes would be considered conforming "in practice" — the rule must be about the
   function of the commit graph alone.
2. **The vector is real**: reproduce it. In a scratch repo under the worktree's `.temp/`,
   commit the fixture tree with its exact bytes, run an object-database extraction
   (`git ls-tree -r` + `git cat-file`) under `core.autocrlf=true` and `false`, compute the
   core §8 content hash (write a 20-line Python or use the generator's function) and
   confirm it equals the expected file; then show `git archive` under `autocrlf=true`
   does NOT match (so the vector actually discriminates). Paste commands and hashes.
3. **Fixture survives checkout**: the repo `.gitattributes` marks the fixture `-text`;
   `git ls-files --eol` shows `i/crlf` and `i/mixed` for the CRLF/mixed files; a fresh
   `git clone` + `git -c core.autocrlf=true checkout` of the branch leaves the fixture
   bytes unchanged (do it in `.temp/`).
4. **Generator and gates**: the new case is tested like its siblings; `make validate`
   and `make regenerate-check` pass in the worktree (run them); `release/*.json`
   untouched; `conformance/v1/manifest.json` regenerated, not hand-edited.
5. **Docs**: CHANGELOG `Unreleased` entry accurate; conformance/README bullet; §13
   surface list updated; nothing else changed (`git diff --stat`).
6. **Signed commit**: exactly one commit, `git log --show-signature -1` verifies with
   the repository's human identity.

## Constraints
Read-only on the worktree and every checkout: no edits, commits, pushes. Scratch work
only under the worktree's `.temp/`. Never write LOGBOOK.md or anything into the control root.

## Verdict contract
Attach `TASK-260905-2qojpx_review-findings-m3-1.md` (outcome resource): per finding —
severity (blocking|major|minor|nit), file/section, quote, what is wrong, fix; plus your
reproduction evidence. Blocking or major → route to `development`; otherwise an explicit
ACCEPT and leave at `to-review`. Do not mark done. Then
`task-board handoff TASK-260905-2qojpx --role reviewer`.
