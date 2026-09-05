# Review verdict (recovery run RUN-260905-a02f6f): ACCEPT — CR-TASK-260905-2qojpx-1 rev 1

Subject: draft/snapshot-byte-exactness at 606d9be (worktree
/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact, PR #39).
This run was spawned as autonomous recovery of RUN-260905-961fbf, whose ACCEPT was
recorded but left no verdict branch. Verdict is the same: ACCEPT.

## Empty repository delta
The story branch task-board/story/STORY-260905-5u97yt carries no change (tree == base
b4f29cd). That is the correct outcome: the producer brief puts the whole deliverable on
draft/snapshot-byte-exactness in a separate worktree and forbids touching the story
worktree. The reviewable artifact is 606d9be, not the story branch.

## Independently re-verified at 606d9be by this run
- Signature: Good "git" signature, ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM,
  author Ivan Oparin <oparin@me.com>.
- make validate: Python suite OK; go test ./tools/... ok. make regenerate-check: exit 0.
- release/: only release/1.0.0-rc.9.json changed, and only its two manifest_sha256 pins
  (accepted deviation 1 from cycle 1; regenerate-check diffs that file, precedent cef93fb, f8d7e7a).
  rc.5..rc.8 untouched.
- git ls-files --eol on the fixture: crlf.txt i/crlf w/crlf, mixed.txt i/mixed w/mixed.
- Vector discriminates (scratch repo under the worktree .temp/review-m3-3/, fixture committed
  via hash-object --no-filters + update-index so blobs equal fixture bytes; hash in the
  generator's format: rel\0payload joined by \0):
    expected                       sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0
    ODB extraction, autocrlf=true  sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0
    ODB extraction, autocrlf=false sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0
    git archive, autocrlf=true     sha256:0a2305a685be72e4e3ebceef91a392fb959be382e747d36cc7f503fe6553e21a
    git archive, autocrlf=false    sha256:77306c6ac35426e5e2f1504d49ebe50893209835ceaf3d374cbb32f166e63dea
  git archive expands $Format:...$ in subst.txt under both settings (literal count 0), and
  CRLF conversion moves the hash further under true. A plain working-tree `git add` of the
  fixture under `* text=auto` normalizes crlf/mixed in the index, which is exactly why the
  vector's acquisition contract requires committing the exact bytes.
- Cycle 1 (d85c719) and cycle 2 (606d9be F1/F3) findings stand; no open findings.

## Board state observed
TASK and STORY were moved to `done` by the commit-owning mover seconds after this run set
`reviewing`; this run does not alter that. accept_cr on the already-accepted revision is
expected to refuse with a state conflict; this artifact is the verdict evidence.
