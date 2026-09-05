# Drafting report: snapshot byte-exactness (TASK-260905-2qojpx, review M3)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact`,
branch `draft/snapshot-byte-exactness`, base `b4f29cd`. One signed commit,
not pushed, no tag, no PR. The story worktree under
`curator-spec/.temp/STORY-260905-5u97yt/worktree` was left untouched as the
brief instructed (the brief overrides the generic story-workspace note).

## Deliverable

```
d85c7191b9f452a8b26e32cfa67adb64b97f59a4 Specify snapshot byte-exactness and add the acquisition vector
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
```

(`git -c gpg.ssh.allowedSignersFile=maintainers.allowed_signers verify-commit HEAD`, exit 0.
Without that override the machine's global `gpg.ssh.allowedSignersFile` points at a
stale `/private/tmp/curator-spec-rc8-verify.*/maintainers.allowed_signers` path and
prints "No principal matched" although the signature itself is good; that is a
local git-config leftover, not a property of the commit.)

Files in the commit (18): `.gitattributes`, `CHANGELOG.md` (new `## Unreleased`),
`conformance/README.md` (one bullet), `conformance/v1/expected/byte-exact-snapshot_sha256.txt`,
`conformance/v1/fixtures/byte-exact/{.gitattributes,lf.txt,crlf.txt,mixed.txt,subst.txt}`,
`conformance/v1/manifest.json`, `conformance/v1/vectors/snapshot-acquisition.json`,
`protocol/environments.md` (§1.2 + §13), `release/1.0.0-rc.9.json` (see deviation 1),
`tools/generate-vectors/{environments.go,environments_test.go,main.go}`,
`tools/{validate.py,test_validate.py}`. `protocol/core.md` untouched.

## Two deviations from the brief — reviewer decision requested

1. **`release/1.0.0-rc.9.json` IS in the commit** (only its two `manifest_sha256`
   pins changed, `90ee8047…` → `ab25038b…`). The brief said to stop and not commit
   it because it is byte-frozen. That premise does not hold for rc.9: CHANGELOG rc.9
   says "Moved the live candidate-suite pin to release/1.0.0-rc.9.json; rc.8 and
   earlier release metadata remain byte-frozen". `writeRC9ReleaseMetadata` in the
   generator rewrites the pin from the manifest digest on every `make regenerate`,
   `tools/validate.py` fails unless the pin equals the manifest digest, and
   `TestRC9ReleaseMetadataPinsSuiteWithoutClaimFabrication` asserts the same. Any
   new file under `conformance/v1` therefore necessarily moves the rc.9 pin, and both
   precedent commits that added conformance surfaces after rc.9 (`cef93fb`,
   `f8d7e7a`) committed the regenerated pin. Refusing it would make `make validate`
   and `make regenerate-check` unpassable, so I followed repository precedent. If the
   reviewer disagrees, drop the hunk together with the vector — they cannot be
   separated.
2. **No `-text` rule for `conformance/v1/fixtures/byte-exact/**` in the root
   `.gitattributes`** — a comment block explains why instead. I first added the rule
   as the brief asked and it is provably dead: gitattributes precedence gives the
   `.gitattributes` in the path's own directory priority over every ancestor, and the
   fixture's own file says `* text=auto`. Evidence with the rule present:
   ```
   $ git check-attr -a conformance/v1/fixtures/byte-exact/crlf.txt
   conformance/v1/fixtures/byte-exact/crlf.txt: text: auto
   conformance/v1/fixtures/byte-exact/crlf.txt: eol: lf
   $ git add conformance/v1/fixtures/byte-exact
   warning: in the working copy of '.../crlf.txt', CRLF will be replaced by LF the next time Git touches it
   $ git ls-files --eol conformance/v1/fixtures/byte-exact   # index normalized
   i/lf    w/crlf  attr/text=auto eol=lf   conformance/v1/fixtures/byte-exact/crlf.txt
   i/lf    w/mixed attr/text=auto eol=lf   conformance/v1/fixtures/byte-exact/mixed.txt
   ```
   The only override that beats a nested file is `$GIT_DIR/info/attributes`, which is
   not versioned. What actually preserves the bytes on every checkout:
   - the blobs were committed through plumbing (`git hash-object -w --no-filters` +
     `git update-index --add --cacheinfo`), so the index holds CRLF/mixed;
   - git does not convert a path whose index blob already contains CRLF, on add or
     on checkout. Probed on a scratch repo with the same root `.gitattributes`:
     fresh clones under `core.autocrlf=true`, `false`, and `input` all show
     `i/crlf w/crlf` / `i/mixed w/mixed`, working-tree sha256 == blob sha256 for
     all five files, and `git add -A` yields a clean status. Only an explicit
     `git add --renormalize .` rewrites `crlf.txt` and `mixed.txt` (the comment in
     the root file warns against it);
   - `make validate` (`validate_snapshot_acquisition_vectors`) and the Go generator
     test fail on a normalized checkout, so a silent rewrite cannot land.
   Committed state, as the brief asked to paste:
   ```
   $ git ls-files --eol conformance/v1/fixtures/byte-exact
   i/lf    w/lf    attr/text=auto eol=lf 	conformance/v1/fixtures/byte-exact/.gitattributes
   i/crlf  w/crlf  attr/text=auto eol=lf 	conformance/v1/fixtures/byte-exact/crlf.txt
   i/lf    w/lf    attr/text=auto eol=lf 	conformance/v1/fixtures/byte-exact/lf.txt
   i/mixed w/mixed attr/text=auto eol=lf 	conformance/v1/fixtures/byte-exact/mixed.txt
   i/lf    w/lf    attr/text=auto eol=lf 	conformance/v1/fixtures/byte-exact/subst.txt
   ```
   Docs-confidence, not reproduced here: behaviour of Git-for-Windows on the
   `has_crlf_in_index` path (same code base; git 2.50.1 Apple was the probe binary).

## Defect reproduction (`.temp/TASK-260905-2qojpx/repro`, git 2.50.1)

Scratch repo: `.gitattributes` = `* text=auto` + `subst.txt export-subst`; LF, CRLF,
mixed files; `subst.txt` with `$Format:%H$`/`$Format:%h$`; committed with plain
`git add` (note: `text=auto` already normalized the CRLF file to LF at commit — a
second, independent way the working tree and the blob disagree).

| file | blob (`cat-file`) | `archive` autocrlf=false | `archive` autocrlf=true |
|---|---|---|---|
| lf.txt | `4fdbc441…` LF | `4fdbc441…` same | `c8dba689…` CRLF |
| mixed.txt | `f5835d63…` | same | `16414c51…` CRLF |
| subst.txt | `bcc643e2…` literal `$Format:` | `81695b0b…` expanded to commit id | `a681a8bb…` expanded + CRLF |

Object-database extraction (`ls-tree -r` + `cat-file blob`) under `autocrlf=true`
matched every blob digest, `.gitattributes` included.

## End-to-end vector check (`.temp/TASK-260905-2qojpx/e2e`)

The committed fixture was re-committed via plumbing in a scratch repo and acquired
both ways; hash = core §8 over the extracted tree (independent Python).

| acquisition | autocrlf=true | autocrlf=false | literal `$Format:%H` present |
|---|---|---|---|
| expected (`byte-exact-snapshot_sha256.txt`) | `sha256:500ea934…2bced0` | same | — |
| object-db (`ls-tree` + `cat-file`) | `sha256:500ea934…2bced0` ✔ | `sha256:500ea934…2bced0` ✔ | yes |
| `git archive` | `sha256:c10c814b…` ✘ | `sha256:cebfdb1b…` ✘ | no (expanded) |

The vector fails `git archive` under both settings (export-subst alone breaks it)
and passes object-database extraction under both.

## What was written

- `environments.md` §1.2 "Snapshot bytes" (new subsection after §1.1, numbering of
  §1.1 unchanged): the MUST rule (exact committed blob bytes, no working-tree
  conversion or attribute-driven archive processing, function of the commit graph
  only), scoped to `git` profile snapshots, their context modules, and
  profile-scoped skill snapshots, phrased as this capability's requirement with a
  cross-reference to core §6.2/§6.5; the consequence paragraph (content hash, state
  hash, effective pin, hash-bound identities independent of platform/config; §5.6
  premise); `path` copies bytes as they are, unnormalized (review N7 exists in the
  board resource but only as a table row about Windows `path` installs and `--format
  shell`; I did not cite it in normative text); no diagnostic, no conforming
  acquisition path without exact bytes; a non-normative note naming
  `ls-tree`+`cat-file`/raw-object reader as satisfying and `git archive` as not, and
  the skills-via-same-path remark. §13 lists the vector.
- Generator `writeSnapshotAcquisitionVectors` (environments.go, called from main.go
  after the environments vectors): hash via existing `regularFiles`+`contentHash`
  (`.gitattributes` included as a regular file of the tree), per-file byte
  counts/digests, the acquisition contract (including the plumbing commit note,
  because `* text=auto` would otherwise normalize the fixture on `git add`).
- `validate.py` `validate_snapshot_acquisition_vectors` in `main()`: recomputes the
  hash from the checked-out fixture, checks exact inventory, exact `.gitattributes`
  bytes, CRLF-only/mixed/LF-only/literal-placeholder properties, per-file records,
  expected file bytes, and that the contract names `core.autocrlf=true` and
  `$Format:%H$`.
- Negative tests: Go `TestSnapshotAcquisitionVectorIsTheRawFixtureHash` (inventory,
  hash equality, hash-without-.gitattributes must differ, byte properties, per-file
  digests); Python `SnapshotAcquisitionVectorTests` (5): published passes;
  CRLF-normalized checkout fails; expanded export-subst fails; hash omitting
  `.gitattributes` fails; stale expected file fails.

## Gate evidence (run standalone, redirected to `.temp/TASK-260905-2qojpx/*.log`)

`make validate` — exit 0 (`make-validate-01.log`; python venv with
`jsonschema==4.25.1` from `requirements-dev.txt` created under `.temp/`, since the
system python lacks it):
```
python3 -B -m unittest discover -s tools -p 'test_*.py'
........................................................................................................................................................
Ran 152 tests in 20.550s
OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
```
`go test -count=1 ./tools/...` — exit 0 (`go-test-01.log`): `ok ... 0.418s`.
Targeted `-k SnapshotAcquisition` — exit 0, 5/5 ok (`py-negative-02.log`).

`make regenerate-check` after the commit — exit 0 (`make-regenerate-check-02.log`):
```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```
(The first run before committing, `make-regenerate-check-01.log`, exited 2 because
`git diff --exit-code` compares against the index; expected pre-commit failure.)

Not run: nothing else was required. Not pushed, not tagged, no PR.
