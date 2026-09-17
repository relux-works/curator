# TASK-260917-e6nwdj — landing review verdict

Verdict: **ACCEPTED** for exact curator-spec commit `e8b53a003256433761cebce6080d6a955d777f25`, parent `9912db7c1c5082fe167d07f1b26573fd1bf17eaa`. The S5/E5/R3-P2 union may land. No implementation changes requested. Delivery worktree remained clean and read-only throughout. No PR merge or producer checkpoint was performed.

## Per-hunk verdict

Full S5-side, E5-side, and numbered delivery quotations are attached as `TASK-260917-e6nwdj_union-evidence.md`.

| Hunk | Delivery quotation and location | Assessment |
|---|---|---|
| §10.1 link currency | protocol/environments.md:2346–2350: “link-target identity is necessary but / no longer sufficient currency: the store entry’s integrity is verified, not / assumed (section 4); the link target is read with `lstat`-class semantics / (section 8.3.1).” | PASS. Preserves S5 integrity verification and E5 nofollow inspection. The obsolete sufficient-currency/no-rehash claim is replaced as S5 requires; it is not left alongside its negation. |
| §10.1 repair | protocol/environments.md:2373–2404: “Repair writes are section 8.3.1 writes: repair / replaces directory entries and refuses with / `environment_write_would_follow_link`”; then “A store entry is re-applied only after that entry passed / the section 4 contract and its pin hash”. | PASS. E5 write discipline precedes the complete S5 failure-class paragraph. Enclosing refusal, entry staging/publication, git snapshot rebuild, path/local repair failure, and distinct dry-run outcomes are all retained verbatim. Trust verification precedes reapplication; safe entry replacement does not authorize following a foreign link. |
| §12 non-current list | protocol/environments.md:2814–2817: “store-untrusted (`environment_store_untrusted`), / refused-provider (section 11), link-blocked (section 8.3.1), or / unreadable state is non-current”. | PASS. Both S5 store-untrusted and E5 link-blocked remain, alongside the pre-existing conditions. |
| §13 conformance surfaces | protocol/environments.md:3038–3067: “the section 8.3.1 / write-discipline vectors / (`vectors/environments-write-nofollow.json`)”; “former target is byte-identical afterwards. ; / and the section 4 / protected-boundary cases / (`vectors/environments-store-boundary.json`)”. | PASS, editorial nit. Both complete case descriptions survive; only conjunction/punctuation/whitespace are composed. All E5 foreign-target assertions and all S5 boundary/currency/repair/status cases remain. |

## File and architecture fidelity

- Both supplied patch digests match the task metadata. The attached union patch is byte-identical to `git diff 9912db7 e8b53a0`.
- All 8/8 changed paths are exactly the accepted candidate's path set; no unrelated changes. The new store-boundary vector is byte-identical to accepted S5.
- `profiles/manager.md`, `tools/validate.py`, and `tools/test_validate.py` equal independent automatic three-way merges (base 684c9f1, reconstructed accepted S5, landed 9912db7), byte for byte. No validator family or intervening R3/P2 change is lost.
- `protocol/environments.md` produces exactly four conflicts. Reconstructing those unions gives the complete delivery file byte for byte, accounting for the explicit punctuation/whitespace joins in §13. All other protocol hunks merge automatically unchanged.
- `CHANGELOG.md` is the literal additive concatenation of E5 and S5 within its one conflict. Non-blocking editorial nit: E5 ends in “checked” at line 37, while “semantically by `tools/validate.py`.” follows S5 at line 90. No normative behavior depends on that prose.
- The regenerated manifest retains E5 and adds S5; the two rc.9 candidate manifest pins advance together to `sha256:7342f14dc74ccb56c4f1fac6ffdaa10694d791bd39f6627786c603d304ba3ad5`. Fresh regeneration is exact.
- Architecture remains the accepted separation: S5 authenticates store boundaries/pins and distinguishes currency; E5 constrains writes to directory-entry replacement without following foreign links. Existing manager/core references remain intact. The specification validators check modeled conformance data; they do not implement manager filesystem security.

## Validation method and limits

I reran both required make gates myself on a disposable raw-byte copy of e8b53a0 with a temporary Git baseline, using `PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`. No earlier producer green was used as a substitute. The temporary baseline was committed only in `/tmp`, with disposable reviewer identity; no project branch was committed or modified.

The first attempt used `git archive`, which expanded the export-subst fixture. Both initial make commands exited 2 for that copy-only defect. Those logs are retained as setup diagnostics and excluded from the candidate verdict. Raw `git cat-file blob` extraction corrected the harness; post-regeneration comparison verifies 1382/1382 raw tracked blobs equal e8b53a0, with zero mismatches.

Production call sites: `tools/validate.py:7469` main constructs and invokes the checks, including S5 at 7485, E5 at 7490, and registry checkpoint validation at 7492. The suite includes S5 negative tests for forged output, missing diagnostics, incorrect currency, absent/unreadable distinctions, inventory removal, and a five-to-one boundary narrowing. E5 includes symlink-parent/target refusals and unchanged foreign-target assertions.

Independent production-entry attack in a separate disposable copy narrowed 5/5 boundary shapes to ownership-only (1/5), updating scenario observations consistently. After successful manifest regeneration, `python3 tools/validate.py` exited 1 specifically on the missing link-safety check. Thus the semantic gate, not a stale manifest hash, rejects the narrowing. This is one targeted narrowing experiment, not an exhaustive mutation audit. S5 corpus inventory is 20 resolve + 4 dry-run + 10 repair + 4 status cases (38 total); runtime manager behavior is outside this landing review and is not claimed proven.

The run is not goal-bound (`task-board spawn goal` reported none). No Change Request Under Review was assigned for this landing-review leaf; the already-accepted S5 change request is untouched.

## Final gate results

- `make regenerate-check`: exit 0.
- `make validate`: exit 0; 62 schemas, 1095 vector files, 407 Python tests in 548.894 seconds, Go generate-vectors tests passed (0.996 seconds).
- Post-suite disposable `git diff --exit-code`: exit 0; delivery worktree status clean and HEAD unchanged.
- Environment: Python 3.14.6; go1.26.0 darwin/amd64.

### exact-regenerate-check.log

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json

EXIT_CODE=0

```

### exact-validate.log

```text
python3 tools/validate.py
validated 62 schemas and 1095 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.......................................................................................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 407 tests in 548.894s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.996s

EXIT_CODE=0

```

### negative-entrypoint.log

```text
Attack: collapse five boundary-check branches to ownership-only, updating failing_check consistently (5/5 -> 1/5). Regenerate manifest before invoking production CLI so digest mismatch cannot mask semantic refusal.
go run ./tools/generate-vectors -root .
REGENERATE_EXIT=0
validation failed: store-boundary case symlinked-entry-root-untrusted: this branch must fail link_safety
VALIDATOR_EXIT=1

```

### byte-identity.txt

```text
Post-regeneration raw Git blob identity: 1382/1382 tracked paths equal e8b53a0; mismatches=[]

```

