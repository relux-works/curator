# TASK-260924-291k0q review verdict — CR rev1 (candidate tree 00bb23f5) — STOP-THE-LINE (blocked)

## Reproduced
- `python3 tools/verify_skillfile_sources_independence.py` on candidate: rc=0 (89/89 refs identical to rc.10; 28/28 clauses; 1 informative marker).
- `python3 -B -m unittest tools.test_skillfile_sources_independence`: 5 tests OK (includes $ref and prose mutants).
- Gate against the UNCHANGED released protocol text (base 2343512 protocol/skillfile-sources.md): rc=1,
  `protocol/skillfile-sources.md:19: environments section 9.4 is not present at v1.0.0-rc.10`. The gate works, and it proves the
  released rc.13 suite itself is NOT independent of post-rc.10 core.
- CI wiring: .github/workflows/ci.yml:46 in Specification job. CHANGELOG under Unreleased: OK. README baseline statement: OK.

## Findings
1. BLOCKING release/1.0.0-rc.13.json:44 — manifest_sha256 rewritten; `git diff v1.0.0-rc.13 -- release/1.0.0-rc.13.json` non-empty. Released record must stay byte-identical (review note item 1).
2. BLOCKING protocol/skillfile-sources.md:18-23 — not a pure marker: normative "Machine-global Skillfiles follow environments §9.4 profile locks when the manager implements that capability" becomes "may consult ... outside this extension". That drops a conditional normative requirement (review note item 2).
3. conformance/skillfile-sources-v1/manifest.json:513,517 — changes the released rc.13 suite manifest (differs from tag), following from 1–2.

## Why blocked, not to-dev
The only citation the gate rejects is genuinely normative in the released rc.13 suite. The gate cannot pass on main without either (a) a normative change to the released suite text plus altered released manifest/record, or (b) an exception. No producer-side fix satisfies both "passes on main" and "released records byte-identical / no normative change".

## Options (human decision needed)
A. (Recommended) Accept a normative relaxation of §1 as an Unreleased change for the next release (rc.14): new suite manifest pinned in a new release record; release/1.0.0-rc.13.json stays byte-identical to tag (validator must stop cross-checking rc.13 record against the live manifest, or the suite gets a new manifest revision).
B. Keep §1 normative; gate carries an explicit, tested allowlist entry for environments §9.4 as a known conditional capability (partial clients without profile locks are unaffected since the clause is conditional). Gate passes on main with no release/protocol changes.
C. Declare cocoaskills baseline requires environments §9.4 (drop rc.10-only claim) — contradicts operator decision.
Decision needed: A or B.
