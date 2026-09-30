# TASK-260930-2mtgv7 review verdict — CR rev1 — ACCEPTED

Delta: only `.research/260930_compiled-build-leaves-reconciliation.md` (140 lines), no code.

## 1. OBSOLETE verdicts (3eqseq, 2sxx7k, 3pvihp, vs6den, 25d05o, 38l1sy, 22ynoi, d8ktna, 14jjgt) — sound
- `gh run list --commit 0e3169bb…`: CI run 36641526700 conclusion `success` (confirmed). 3eqseq obsolete.
- Disposable spec clone at 23435129 (rc.13): `release/1.0.0-rc.13.json:18 "claim_v5"`, `:20 "claims_emitted": []` (confirmed) → 3pvihp/25d05o/d8ktna.
- `.github/workflows/implementations.yml:38` ref 80fd617f…, `:56` ref 4a88aa0e… — full-commit pins (confirmed) → vs6den.
- `.temp/TASK-260720-1ljev5/worktree` absent (confirmed) → 2sxx7k; residual native-Linux intent is preserved in 1skseh.
- Curator ci.yml:43 SPEC_PIN = rc.13 tag commit (confirmed) → 38l1sy; v6 consumer residual carried by 1673lr.
- Nothing still-wanted is lost: each residual is named with its carrier leaf (1skseh, 1673lr, re-scoped doc/interop leaves). 1utsx8 is csk-owned and flagged as such.

## 2. "No consumer" claims (1uepyd, rjxrgs) — confirmed
`grep -rln` over internal/ cmd/:
- external-repository-acquisition: 0 files.
- external-repository-lifecycle: only `internal/scopes/gc_conformance_test.go` (reads `status_repair_gc_cases` only, :44/:60) — matches "all but status_repair_gc".
- install-marker-v3: 0 files. `grep -c install-marker-v3 .github/ci/conformance-case-counts.tsv` = 0. rc.13 has 27 case files.

## 3. Re-runs (CURATOR_CONFORMANCE_ROOT = rc.13 clone, -count=1 -v)
- T2: exit 0, 27 top-level PASS, `ok github.com/relux-works/curator/internal/buildrepo 4.274s`
- T6: exit 0, 3 PASS, `ok github.com/relux-works/curator/internal/marker 0.388s`

## 4. Plan
Dependency order is right: board closes → C vector consumers (2,3,4; 4 depends on marker-v4 BUG-260923-2afgyq coordination) ∥ 20ao7p → S contract before 2gbtb9/ypbuav/3vtl57 → S+K interop → H Linux last. Both human decisions are correctly framed as product/scope questions (still want manifest-declared toolchain preflight? still want cross-manager black-box parity?) and correctly gate only their own branches. csk (K) and lev (H) needs are flagged per leaf.

Minor (non-blocking): 2u5u14 PARTIAL is honestly reported as "unknown" against its AC; the next S leaf should settle it.
