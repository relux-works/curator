# TASK-260930-12i5zr review verdict — rev1 ACCEPTED
CR-TASK-260930-12i5zr-1 rev1, base cbe52078, candidate tree 17bfe7dd (11 paths). Suites: rc.13 manifest sha256 be11bb1e… (v1.0.0-rc.13 worktree), candidate curator-spec 526a9aa067a0 manifest sha256 950ee74a… (both digests recomputed, match the TSV pins).

## 1. Counts keyed by suite identity
- Header suite_manifest_sha256/family/expected_cases; Load selects the table by sha256(manifest.json) of CURATOR_CONFORMANCE_ROOT; no root → rc.13 identity. No lower bound anywhere (RequirePublishedCount and coverage harness are exact `!=`).
- rc.13 table vs main: every one of main's 87 rows is byte-identical (comm diff empty); 4 rows ADDED (audit-record-v1 2, context-lock-v1 31, install-marker-v3 27, registry-log-entry-v1 2) for the new frozen-case consumers — a tightening, not a change.
- Unknown digest fails closed: manifest copy + 1 byte → EXIT=1 "no published-case count pins for conformance manifest sha256:f7899951…".
- Mutant M1: candidate install-marker-v4 28→29 → candidate EXIT=1 ("publishes 28 cases, want pinned count 29"), rc.13 EXIT=0. Mutant M2: rc.13 context-lock-v1 31→30 → rc.13 EXIT=1, candidate EXIT=0. TSV restored (sha256 verified).

## 2. Frozen-shape negatives DRIVEN (candidate, -v PASS)
install-marker-v4 + v3 (marker.Read, with hash_version-stripped positive control), context-lock-v1 (contextlock.Parse + control), agent-environment-marker-v2 (envmarker.Parse + control), audit-record-v1 (registry.ParseRecord + control), registry-log-entry-v1 v2-record-in-frozen-entry. No frozen case appears in conformance-gaps.tsv (grep count 0). Under rc.13 the cases are absent and the tests assert absence on disk and that the digest is not the candidate's.

## 3. Gap rows
103 candidate-digest rows = 87 owned by TASK-260917-2tx81l + the 16 pre-existing rows re-keyed; rc.13 digest carries exactly main's 16 rows. The 2tx81l rows load only under 950ee74a; guard tests check owner, disk/index equality and the 5 content-hashes-v2 vectors.

## 4. Package set, real exit codes
marker contextlock envmarker registry manifest skillspec conformancecoverage: rc.13 EXIT=0, candidate EXIT=0.
crossconformance (full, -timeout 9m): candidate EXIT=0 (393.7 s), rc.13 EXIT=0 (386.1 s). No interruption reproduced.

## 5. Scope
Only conformancecoverage (test-support), tests and the two CI TSVs changed; no CHANGELOG/LOGBOOK changes. docs/ci-gates.md prose still accurate ("at its selected corpus revision").

Residual (non-blocking): the candidate digest is pinned to 526a9aa0; any later PR #116 push changes the manifest and needs a new table (fails closed by design).
