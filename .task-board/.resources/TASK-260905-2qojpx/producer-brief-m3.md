# Producer brief: snapshot acquisition byte-exactness (review M3)

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact`,
  branch `draft/snapshot-byte-exactness`, base `b4f29cd` (= curator-spec main).
- Scope: `protocol/environments.md` (rule text), a new fixture tree under
  `conformance/v1/fixtures/`, `tools/generate-vectors/` (one new case), the
  generated `conformance/v1/vectors/*.json` + `conformance/v1/expected/*` +
  `conformance/v1/manifest.json` (regenerated, never hand-edited), the repo
  `.gitattributes` (see below), `conformance/README.md` (one bullet),
  `CHANGELOG.md` (an entry under a new `## Unreleased` heading — mirror the
  rc.9 entry style; do not bump any version). Nothing in `protocol/core.md`
  (frozen). The story worktree task-board provisions for this run stays
  untouched; the work lives only in the worktree above.
- Deliverable: one signed commit (`git commit -S`; paste the
  `git log --show-signature -1` line into your report). `make validate` and
  `make regenerate-check` MUST pass in the worktree (record the exact output
  tails). Do not push, tag, open a PR, or mark the task done. Attach
  `TASK-260905-2qojpx_drafting-report.md` as an outcome resource. Then hand off.

## The defect (verified by the pre-implementation review, lens C)

Board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8
(`~/Developer/ReluxWorks/curator/.task-board/.resources/STORY-260901-zddtn8/`), item M3:
`git -c core.autocrlf=true archive` emits CRLF for a committed-LF blob and
`git archive` expands `export-subst`; the reference implementation
(`~/Developer/ReluxWorks/curator/internal/gitops/gitops.go` `Archive`, called from
`internal/snapshot/snapshot.go` and `internal/closure/closure.go`) runs plain
`git archive`, and audit hashes the *extracted* tree. So content hashes, `path`/`local`
state hashes, effective pins, and revocation identities depend on git config and
attributes; on stock Git-for-Windows every git-sourced profile fails; the shipped
skills pipeline carries the same defect. Reproduce it yourself on a scratch repo
under `.temp/` of the worktree before writing (record the commands and outputs in
the report): commit an LF file, a CRLF file, a `.gitattributes` with `* text=auto`
and `<name> export-subst`, and a file containing `$Format:%H$`; compare
`git archive` output bytes under `core.autocrlf=true|false` with `git cat-file -p`
blob bytes.

## The rule (normative; place it in environments.md §1 right after the source-kind list, as a new paragraph or a §1.2 "Snapshot bytes" subsection — decide, keep numbering stable for §1.1 Diagnostics)

- A snapshot produced from a commit MUST contain, for every regular-file entry of
  the commit's tree, exactly the committed blob bytes. Working-tree conversion
  (`core.autocrlf`, `text`/`eol` attributes, clean/smudge filters, `ident`) and
  attribute-driven archive processing (`export-subst`, `export-ignore`) MUST NOT
  alter, add, or omit any entry: the snapshot is a function of the commit object
  graph alone, never of the acquiring machine's git configuration or of the
  repository's attributes. State that this is what core §6.2's "snapshots are
  immutable regular-file trees produced from that commit" and core §6.5's
  "materialize exact blob bytes" already require for external repositories, now
  stated for profile, context, and skill snapshots alike — environments.md may
  not amend core, so phrase it as the environments capability's requirement and
  note that a manager whose skill snapshots come through the same acquisition
  path satisfies it for both (say this in a non-normative note).
- Consequence, stated: the core §8 content hash of a snapshot, the `path`/`local`
  state hash, the effective pin, and every hash-bound identity are platform- and
  configuration-independent (the §5.6 cross-platform hash equality claim now has
  its premise).
- The `path` kind: snapshotting a working directory copies the directory's bytes
  as they are (there is no commit); say explicitly that `path` snapshots are not
  normalized either, and cross-reference review N7 as future work only if you
  can cite it from the board resource — otherwise omit.
- Diagnostics: none new; a manager that cannot produce exact bytes has no
  conforming acquisition path — say so.
- Informative implementation note (one sentence): extraction from the object
  database (`git ls-tree -r` + `git cat-file --batch` or a raw-object reader
  under core §6.5) satisfies the rule; `git archive` does not.

## The vector

- New fixture tree `conformance/v1/fixtures/byte-exact/` (or a name you justify)
  containing at least: `.gitattributes` with exactly the lines `* text=auto` and
  `subst.txt export-subst`; `lf.txt` (LF line endings, several lines); `crlf.txt`
  (CRLF line endings); `subst.txt` containing the literal `$Format:%H$` and one
  more `$Format:...$` placeholder; `mixed.txt` (a file with both LF and CRLF
  lines). Keep every file small and ASCII.
- The spec repository must itself preserve those bytes on every checkout: add to
  the repo root `.gitattributes` a rule marking `conformance/v1/fixtures/byte-exact/**`
  as `-text` (verify the existing file's conventions first and follow them).
  Verify after committing: `git ls-files --eol conformance/v1/fixtures/byte-exact`
  shows `i/crlf` for `crlf.txt` and `i/mixed` for `mixed.txt`, and paste that
  output into the report.
- Generator: add one case to `tools/generate-vectors` that computes the core §8
  content hash over the fixture's regular files (reuse `contentHash`/`regularFiles`
  from `main.go`; check whether `.gitattributes` must be part of the hashed set —
  it is a regular file of the tree, so it is) and writes it to
  `conformance/v1/expected/byte-exact-snapshot_sha256.txt` (or a name matching the
  existing `snapshot_sha256.txt` convention) plus a vector entry in a new
  `conformance/v1/vectors/snapshot-acquisition.json` describing the case: the
  fixture path, the acquisition contract ("commit the fixture tree with these
  exact bytes and attributes in a repository; acquire a snapshot of that commit
  with `core.autocrlf=true` and again with `false`; both snapshots MUST hash to
  the expected value; an `export-subst` entry MUST still contain the literal
  `$Format:` text"), and the expected hash. Extend `main_test.go`/
  `environments_test.go` with the same assertions the other cases get (look at
  how existing vector files are covered by tests and match it).
- Run `make regenerate` then `make validate` and `make regenerate-check`. Commit
  the regenerated `manifest.json`, `vectors/`, and `expected/` bytes together with
  the generator change. If regeneration touches `release/1.0.0-rc.9.json`, stop
  and record it in the report instead of committing a release-metadata change
  (that file is byte-frozen); find out why and describe it.
- `conformance/README.md`: one bullet naming the new surface next to the
  agent-environments bullets. environments.md §13: add the byte-exactness vector
  to the conformance-surface list.

## Constraints

- Verify before asserting; label anything not reproduced on this machine as
  docs-confidence in the report, never in the normative text.
- No pushes, no tags, no PR; one signed commit; report attached; then
  `task-board handoff TASK-260905-2qojpx --role developer`. Never write LOGBOOK.md
  or anything into the control root.
