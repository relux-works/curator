# Review findings cycle 2: byte-exactness F1/F3 edit at 606d9be — ACCEPT

Subject: worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact,
head 606d9be (on top of cycle-1-accepted d85c719), PR relux-works/curator-spec#39.
Scope reviewed: `git diff d85c719..606d9be` = CHANGELOG.md + tools/test_validate.py only.

## Verdict: ACCEPT. No blocking, major, or minor findings.

Change Request CR-TASK-260905-2qojpx-1 rev 1 has `repository_delta=empty` on the story
branch. That is the correct outcome: the producer brief places the deliverable on
`draft/snapshot-byte-exactness` in a separate worktree, not on the story branch. The
orchestrator integrates 606d9be.

## F1 (CHANGELOG wording) — resolved, and the new claim is true

New text claims: fixture blobs committed through plumbing survive every checkout
unconverted under the repository's `eol=lf` policy; no attribute rule can protect them
because the nested `* text=auto` outranks the root file.

Verified:
- `git check-attr -a` on lf.txt/crlf.txt: `text: auto`, `eol: lf` (root `eol=lf` still
  applies because attributes compose per-attribute; nested file only sets `text`).
  So a root `-text` rule would indeed be overridden for `text`. Comment in root
  `.gitattributes` is accurate.
- `eol=lf` overrides `core.autocrlf` on checkout and never adds CR; git also does not
  strip CRLF that is already in the index for `text=auto`.
- Fresh `git clone --config core.autocrlf=true` of the branch: `ls-files --eol` shows
  `i/crlf w/crlf` for crlf.txt and `i/mixed w/mixed` for mixed.txt; sha256 of each
  working-tree file equals sha256 of its HEAD blob (5/5 SAME).
- `touch` + `git add` of the fixture dir in that clone: index unchanged (no
  renormalization on add).
- `python3 tools/validate.py` in the autocrlf=true clone: `validated 57 schemas and 780
  vector files`, exit 0.

Negative shape (gate must reject): normalized crlf.txt to LF in the clone ->
`validation failed: vector digest mismatch for fixtures/byte-exact/crlf.txt`, exit 1.
Reverted -> exit 0. The gate discriminates.

## F3 (test class below `__main__` guard) — resolved

`python3 tools/test_validate.py -v` now runs all 5 `SnapshotAcquisitionVectorTests`
(published passes; crlf-normalized, expanded export-subst, hash omitting .gitattributes,
stale expected file all fail closed). Suite: `Ran 74 tests ... OK`.

## Gates rerun by reviewer at 606d9be

- `make validate`: `Ran 152 tests in 51.750s OK`; `go test ./tools/...` ok.
- `make regenerate-check`: `git diff --exit-code` over conformance/v1 and release/*.json
  clean (release files untouched).
- `git log --show-signature -1 606d9be`: Good "git" signature, ECDSA
  SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, author Ivan Oparin. (The
  "allowed keys file" warning is a stale local allowed_signers path, not a signature
  problem.)
- `git status` in the worktree: only untracked `tools/__pycache__/`.

No edits, commits, or pushes were made; scratch under the worktree's `.temp/review-m3-2/`.
