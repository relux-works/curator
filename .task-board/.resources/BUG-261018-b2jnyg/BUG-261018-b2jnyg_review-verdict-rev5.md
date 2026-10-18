# Review verdict: accepted

Element: BUG-261018-b2jnyg — repair refuses floats in managed Claude state.
Revision: 5. Candidate commit: `6cc8b035bb1709dd8d116de28d3649867501f935` (signature G).
Tree: `d1d16a8b497530c1c5ddca460fab693daf3bd069`.
Base: `2e5ab6dab61fc78a525cbcb0de6c5fcc9f5ae882`.
Repository delta: present, four task paths. No product files modified by reviewer.

Verdict schema:
```json
{"verdict":"accepted","revision":5,"findings":[],"notes":["Coverage measures 13 of 13 behavioral rows through Resolve; direct CLI parsing is explicitly outside that ratio.","Necessary merges may normalize whitespace, key order and string escapes; decoded values and number text are preserved. Satisfied files remain byte-identical.","Special-file attacks cover directories; concurrent filesystem mutation and every OS special-file class are not claimed covered.","Original full local suite was reused after all 57 raw-log size/hash checks, not rerun. Targeted tests and six narrowing controls were rerun by this reviewer.","No review/free-hunt wall-clock budget was supplied. All ten requested rows were swept."]}
```

# Review sweep revision 5

Coverage: 13 of 13 behavioral AC rows named; CLI parsing is a stated bound. Exact four candidate blobs verified. Fresh trunk 892dab098aef1497df9ccacdb173faaf4f8b8418 has zero changed-path overlap.

1. held — specification ownership: environments §7.4 provisioning seeds permits only the launch-entry merge on repair; no post-provisioning CCJ-1 requirement. Checked supplied spec source, docs claim and Resolve call sites.
2. held — TestClaudeRepairPreservesNumberText, all 14 variants via Resolve; numbers.log; go test ./internal/envprofile -count=1 -timeout=3m -v -run '^TestClaudeRepairPreservesNumberText$'; exit 0. Includes overflow 1e999, signed zero, huge integer, exponents and nested arrays. managed.go:868,884.
3. held — TestClaudeRepairSatisfiedIsNoOp via Resolve; noop.log; go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestClaudeRepairSatisfiedIsNoOp$; exit 0. Both forms preserve bytes, inode and mtime during other-surface repair; managed.go:862.
4. held — TestClaudeRepairRefusesUnsafeState through Resolve (all seven cases, no skips), refusals.log, go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestClaudeRepairRefusesUnsafeState$, exit 0. Numeric neighboring variants already exercised in row 2. managed.go:821,833 and transaction backup/nofollow guards. Refusals preserve leaf identity, metadata, bytes and symlink target.
5. held — TestClaudeAbsentSeedGolden both forms through Resolve; golden.log; go test ./internal/envprofile -count=1 -timeout=3m -v -run ^TestClaudeAbsentSeedGolden$, exit 0. Explicit expected pre-fix bytes, managed.go:871.
6. held — TestClaudeRepairPreservesNumberText second Resolve (Repair=false), including 1e999; same numbers.log and exit 0 as row 2. claudeProjects:910,925 called at managed.go:2204; UseNumber at :884.
7. held — independent rg inventories all-encodes.log and harness-files.log (exit 0), inspected production callers and compared gaps sweep. Codex TOML at managed.go:689 is provisioning-only; Claude is the only existing tool-state JSON merge in envprofile. MCP/referenced CCJ-1 surfaces are manager-authored. No additional same-class tool-owned re-encode found within searched production internal/ tree. This is bounded source/ownership inspection; no executable sweep test exists or is claimed.
8. held — TestClaudeSeedNarrowingMutants, mutants.log; go test ./internal/envprofile -count=1 -timeout=4m -v -run ^TestClaudeSeedNarrowingMutants$, exit 0. Six actual behavioral child runs exit 1 for named failures, including whole-object CCJ-1, retained-token number normalization, both flag weakenings, narrowed no-op and admitting null. No setup failure counted.
9. held — candidate README diff is empty (git diff base tree -- README.md, exit 0); source inspection confirms claudeRepairFixture -> writeManagedFixture assigns temporary home/native/xdg/launch at managed_test.go:35-42; regression and refusal execution above uses those public Resolve fixtures. New tests never resolve the real HOME as the repair target. Tool processes may read ordinary ambient tool configuration; no stronger sandbox claim.
10. held — gh run view 39166396311 --json headSha,conclusion,jobs,url, exit 0; success, 20 successful jobs / 2 conditional skips on signed commit 6cc8b035bb1709dd8d116de28d3649867501f935, exact candidate tree. Full local evidence reused: all 57 raw log hashes/sizes verified; 82 whole packages plus completed inventories 374/374 envprofile, 363/363 install, 425/425 CLI. Original failures/skips retained, not promoted to passes; successful completed reruns supply coverage. validation-audit.json records evidence. Configured CR5 gate log terminates with exit 0. git diff --check exit 0. No local mirror used.


## Free hunt

No blocking finding. Additional valid JSON values were injected into the existing regression fixture through a Go overlay, leaving candidate code untouched: `-0.0000`, `0e-0`, `1.23000e+0004`, a large negative integer, `1E-999`, and a string containing HTML-sensitive characters, escaped NUL and a Unicode surrogate pair. Both forms passed through Resolve repair and verification. Command: `go test -overlay .temp/review-b2jnyg/freehunt-overlay.json ./internal/envprofile -count=1 -timeout=3m -v -run '^TestClaudeRepairPreservesNumberText$/^(referenced|monolithic)$/^new-launch$'`, exit 0. Reproduction recipe is attached separately.

## Evidence and limits

- `BUG-261018-b2jnyg_review-evidence-rev5.log`: complete targeted output, local paths redacted; includes intended nested exit-1 mutant failures and successful parent harness.
- `BUG-261018-b2jnyg_review-validation-rev5.json`: independent hosted API response, signature, local raw-log integrity checks and coverage inventories.
- Producer gaps resource supplies the AC map, sweep list, original failed attempts and declared bounds. The specification was independently read in the exact local spec fixture (`43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, environments §7.4 provisioning-seeds/project-entry clauses); its ownership rule agrees with the producer's earlier cited revision.
- All ten row results are held for the named checks only. Non-executable spec/sweep/scope rows were checked against source and actual production callers rather than misrepresented as standalone behavioral tests.
- Full hosted run: https://github.com/relux-works/curator/actions/runs/39166396311 . Current success was read independently, not inferred from a previous attempt or publication note.
- Acceptance authorizes the named revision; no landing, default-branch update or post-landing result is claimed. Primary/run goal read reports this run is not goal-bound.

## Task-scoped logbook

Publication-only retries retained the same candidate tree. Current revision finally has an actual green hosted gate; historical infrastructure failures remain historical failures. Fresh protected trunk has no overlap with the four candidate paths. Review found no additional same-class re-encode inside envprofile. Acceptance is recorded for integration, with no code edits and no commit by the reviewer.
