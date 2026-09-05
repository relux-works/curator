# Review findings M3 cycle 1: snapshot byte-exactness (TASK-260905-2qojpx)

Subject: worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact`,
branch `draft/snapshot-byte-exactness`, head `d85c719`, base `b4f29cd` (= main). Exactly
one commit ahead of base, 18 files. Change Request `CR-TASK-260905-2qojpx-1` rev 1 on the
story branch carries `repository_delta=empty` by design: the producer brief located the
work in the draft worktree above and told the producer to leave the story worktree
untouched. Reviewed read-only; scratch only under the worktree's
`.temp/TASK-260905-2qojpx/review-m3-1/`.

## Verdict: ACCEPT

No blocking or major finding. Two minor findings and one nit below; they do not need
another producer cycle before integration, but the CHANGELOG line should be fixed
when the orchestrator integrates (or in the next cycle if one happens anyway).

## Findings

### F1 (minor) CHANGELOG.md, `## Unreleased`, second bullet
Quote: "a repository `.gitattributes` rule that keeps those fixture bytes unconverted on
every checkout". Wrong: the commit adds no such rule (drafting report deviation 2, which
I confirmed: a `conformance/v1/fixtures/byte-exact/** -text` rule at the root is dead
because the fixture's nested `* text=auto` outranks it; `git check-attr` still reports
`text: auto`). What keeps the bytes is the plumbing commit plus git leaving in-index CRLF
alone, guarded by `make validate`. Fix: replace the clause with "fixture blobs committed
through plumbing so they survive every checkout unconverted, a root `.gitattributes`
note explaining why no attribute rule can protect them".

### F2 (minor) Story checklist item 2 / producer brief vs. delivered `.gitattributes`
The brief and the story DoD ask for a `-text` rule. The delivered comment-only approach
is the correct one; the rule would have been a placebo. Reviewer decision: deviation 2
accepted, DoD item satisfied in substance. Evidence in "Reproduction" below.

### F3 (nit) tools/test_validate.py
`SnapshotAcquisitionVectorTests` is defined after the `if __name__ == "__main__":
unittest.main()` block, so running the file directly as a script never sees the class.
`make validate` uses `unittest discover`, which imports the module and does see it (5/5
ran). Move the class above the main guard when the file is next touched.

## Deviation 1 (rc.9 pin in the commit): accepted
`release/1.0.0-rc.9.json` changes only its two `manifest_sha256` pins. Both precedent
commits that added a conformance surface after rc.9 (`cef93fb`, `f8d7e7a`) moved the
same pin, and `make regenerate-check` diffs that file, so it cannot be left behind.
`release/1.0.0-rc.5..8.json` untouched (`git diff --stat`).

## Dimension 1: the rule (environments.md §1.2, §13)
- Normative text says the snapshot "is a function of the commit object graph alone,
  never of the acquiring machine's git configuration or of the repository's
  attributes", enumerates `core.autocrlf`, `text`/`eol`, clean/smudge, `ident`,
  `export-subst`, `export-ignore`, and adds "and no other entry". The "git archive
  under autocrlf=false with no attributes is fine in practice" reading is closed: the
  rule binds the output to the object graph, not to any config state.
- `path` non-normalization stated; consequence for content hash, `path`/`local` state
  hash, effective pin and hash-bound identities stated with the §5.6 premise; no new
  diagnostic and "no conforming acquisition path" stated; the informative note names
  object-database extraction as satisfying and `git archive` as not.
- Cites core §6.2 line 1200 and §6.5 line 1375 verbatim; `protocol/core.md` untouched.
  §1.1 numbering unchanged. §13 lists the vector.

## Dimension 2: the vector is real (reproduction)
Independent 6-line Python hash of core §8 (`hash.py`), scratch repo committed via
`git hash-object -w --no-filters` + `git update-index --cacheinfo`, git 2.50.1.

| acquisition | autocrlf=true | autocrlf=false |
|---|---|---|
| expected file | `sha256:500ea934…2bced0` | same |
| `ls-tree -r` + `cat-file blob` | `sha256:500ea934…2bced0` ✔, literal `$Format:%H` present | ✔ same |
| `git archive` | `sha256:ca1584d8…` ✘ (subst expanded to commit id; CRLF kept) | `sha256:9773dcbc…` ✘ (subst expanded) |

Blob digests in the scratch commit equal the per-file `sha256` records in
`snapshot-acquisition.json` (`ced60fac…`, `c8dba689…`, `4fdbc441…`, `c76a5bc3…`,
`ec9a6c8c…`). The vector discriminates under both settings.

## Dimension 3: fixture survives checkout
`git ls-files --eol` in the worktree: `i/crlf w/crlf` crlf.txt, `i/mixed w/mixed`
mixed.txt, `i/lf` for the other three. Fresh `git clone` of the branch under
`core.autocrlf=true`, `false`, and `input`: identical eol output, fixture hash
`sha256:500ea934…2bced0` in all three, `git add -A` leaves 0 dirty paths. Attack:
appending `conformance/v1/fixtures/byte-exact/** -text` to the root file in a clone
still yields `text: auto` from `check-attr` (rule dead, confirming deviation 2).

## Dimension 4: generator and gates
- `writeSnapshotAcquisitionVectors` is called from `main.go` (production path), reuses
  `regularFiles`/`contentHash`; Go `contentHash` and Python `environment_content_hash`
  both implement core §8 and agree with my independent hash.
- `validate_snapshot_acquisition_vectors` is in `main()`'s check list.
- Reran myself: `make validate` exit 0 (`Ran 152 tests … OK`, `go test ./tools/... ok`),
  `go test -count=1 ./tools/...` ok, targeted `-k SnapshotAcquisition` 5/5 OK,
  `make regenerate-check` exit 0 with a clean tree afterwards (manifest regenerated,
  not hand-edited). Logs: `make-validate.log`, `make-regenerate-check.log` in the
  scratch dir. jsonschema came from the producer's `.temp/TASK-260905-2qojpx/venv`.
- Live mutants on a clone through `python3 tools/validate.py`: mixed.txt normalized →
  `validation failed: vector digest mismatch`; `$Format:%h$` expanded → digest
  mismatch; extra fixture file → manifest inventory mismatch. Unit negatives narrow
  the new check itself (CRLF-normalized, expanded subst, hash omitting
  `.gitattributes`, stale expected file) and pass.

## Dimension 5: docs
CHANGELOG `Unreleased` present in rc.9 style, no version bump; one README bullet;
§13 updated. `git diff --stat` shows nothing outside the brief's scope except the
test/validator additions the brief implied. F1 is the one inaccuracy.

## Dimension 6: signed commit
`git log --show-signature -1`: `Good "git" signature with ECDSA key
SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, author `Ivan Oparin <oparin@me.com>`.
`git -c gpg.ssh.allowedSignersFile=maintainers.allowed_signers verify-commit HEAD` exit 0
("Good signature for oparin@me.com"). The global allowedSignersFile on this machine
points at a stale `/private/tmp` path; local config leftover, not a commit property.

## Empty repository delta on the Change Request
Accepting rev 1 with `repository_delta=empty` is correct because the brief placed the
deliverable on `draft/snapshot-byte-exactness` (`d85c719`) outside the story worktree.
Integration must take that commit, not the story branch tree. The orchestrator owns
that step; nothing on the story branch needs to change for this leaf.

## Docs-confidence (not reproduced here)
Git-for-Windows behaviour on the in-index-CRLF path; same code base, probed only with
git 2.50.1 (Apple).
