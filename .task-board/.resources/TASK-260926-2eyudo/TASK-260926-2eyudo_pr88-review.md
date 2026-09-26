# TASK-260926-2eyudo — exact-head review of relux-works/curator-spec#88 @ 94e8c78

**Verdict: CHANGES-REQUESTED** — items 1 and 2 (content) pass; item 3 fails: the pin bump is not lockstep with the spec-side coverage declaration, and the new pin breaks the Windows checkout. Required checks on 94e8c78 are red.

Method: disposable clones in $TMPDIR (curator-spec incl. `pull/88/head`, relux-works/curator); nothing pushed, no repo files changed.

## 1. b202b5d vs dcc7f01 — PASS
- b202b5d parent = 5746367 (main); dcc7f01 parent = 3d2c611.
- Per-file `git diff <c>^ <c> -- f | git patch-id --stable`, identical for all 9 non-CHANGELOG paths:
  conformance/v1/manifest.json 4ebbfb533c83; conformance/v1/vectors/script-host-execution-policy.json d59e11807f06; profiles/manager.md 156ae32e4e4a; protocol/core.md ac6f05350490; release/1.0.0-rc.9.json 450b27cfb888; tools/generate-vectors/main.go 861e929fcbc1; tools/generate-vectors/main_test.go de77aa7d2809; tools/test_validate.py 43c4217ac417; tools/validate.py a13e88b06643.
- CHANGELOG.md: patch-ids differ (d4602c6b vs 2c821865) only by context: the same 10-line "Defines hard-link substitution…" entry is inserted under `### Changed` after main's Decision-0022 transport entry (CHANGELOG.md:607). `comm` of main 5746367 vs b202b5d sorted lines: no line lost; entry occurs exactly once. Clean union.
- TASK-260924-mcmova_results.md (dcc7f01 patch-id 9140f840) absent in b202b5d — the only other difference.

## 2. 94e8c78 — content PASS
- Parent b202b5d; diff touches only .github/workflows/implementations.yml (+6/-6): comment + `ref: a3abcf34… -> 0a62862130fd6cbd9db8b246da63fdb5aecdfaaf` (implementations.yml:34-39). cocoaskills ref 3ecca1db (:51) and registry ref d690bea6 (:58) untouched.
- curator: 0a628621 is ancestor of origin/main (ON_MAIN) and descends from a3abcf34. Range contains e337d979 (STORY-260922-2goxjs, carries TASK-260922-18ex37); 0a628621:internal/scriptpolicy/conformance_test.go:68-69,169,172 decode and classify `executable_identity_cases` and `hard_link_substitution_definition`. Comment is accurate.

## 3. Required checks on 94e8c78 — FAIL (`gh pr checks 88 --required`, exit 1)
pass: Formatting, Links, Specification (macos, ubuntu); pending: Specification (windows) at check time;
fail: Implementations (ubuntu, macos, windows), run 36202537423. main's last Implementation run (36182557840 @ 5746367) was green, so the failures are introduced by this PR.

### Findings
- **F1 (blocking) .github/ci/implementation-coverage.tsv:48** still declares `go internal/scriptpolicy.TestARefusalPrecedesEveryWorkerSurface`; that test was removed in curator 9a59d568 (STORY-260822-2h0v9j curator-go-script-worker), which lies in a3abcf34..0a628621. Ubuntu/macOS fail at "Gate the declared Go consumption cases": `implementation-coverage: FAILED … TestARefusalPrecedesEveryWorkerSurface was not observed in this run`. The lockstep bump must also update the coverage TSV to the replacement production test(s) (and add rows for the new sections if the gate expects them).
- **F2 (blocking) .github/workflows/implementations.yml:30-41** Windows "Checkout Go manager" fails: `cannot create directory at '.task-board/EPIC-260720-21aq1i_…/TASK-260728-3b8qym_…': Filename too long`. At 0a628621 tracked paths reach 230 chars (max 160 at a3abcf34). Needs `core.longpaths` on the runner (e.g. `git config --system core.longpaths true` step before checkout) or a sparse checkout excluding `.task-board`.
