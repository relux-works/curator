# TASK-261016-2rpqii — revision 2 review

Candidate tree: bf3995a03bbdc346f1e16625d699f2650d7bd5f7.
Base: 205272e75ea7d40fd10e073c20ce49bbc0787cd2.
Fresh origin/main: 0d00a109ad1da77a5506747c89f24af9442eb341; only board paths differ from base.
Scope: documentation and diagrams, no runtime gate changed.

AC driving checks identified before review: 5/5 rows have named checks:
1. CLIAvailability + FactAndPolicyComparison (source pack, actual dispatch, normative spec).
2. RenderAndCompare (four SVGs).
3. PointerAndAnchorCheck (README and environment-config).
4. VersionAndHygieneCheck (candidate-only scan and source comparison).
5. LycheeLocalAndExternal, with MissingTargetControl.
These are documentation checks, not a claim of product test coverage.

## Surface sweep

1. Truth against code — held. CLIAvailability: env/profile help and dispatch confirm resolve/config exist and login/prepare do not. Native selector code and named admission tests exercise current isolation. Fresh main differs only in board records. Candidate marks future commands rc.7. No newly available command was mislabelled.
2. Truth against facts — held. FactAndPolicyComparison: compared every vendor table against the supplied versioned facts, looking specifically for invented Linux qualification, desktop/ssh equivalence, forced Claude storage, and keyring-isolation claims. None reproduced. Muse Linux and refresh remain explicitly unknown; Claude 2.1.299, Codex 0.153.2 and Muse 1.4.8 scope the observations. No live credentials inspected.
3. Credential store — held. Policy comparison against current login-readiness and CIP-0017: exact machine knob, native default, source/store/isolation/link order, Claude force-store refusals and no automatic migration preserved. Tried the apparent Muse native contradiction: CIP-0017 explicitly supersedes the generic mapping for managed Muse; effective file in the candidate is correct.
4. Visibility — held. PointerAndAnchorCheck: README pointer sits outside the collapsed details section, contains exact machine command and native default, and labels rc.7. environment-config links the reference. Local Lychee resolved the links and fragments (exit 0).
5. Diagrams — held. RenderAndCompare rerendered four sources in scratch with PlantUML; all SVGs match the exact candidate byte-for-byte and parse as XML. All four are embedded. Source/text comparison checked unknown versus absent preparation, shared versus isolated links, refusal branches, and measured store bounds. The prepare diagram is a summary over participants, not a runtime implementation.
6. Public hygiene — held. VersionAndHygieneCheck: naming gate on byte-exact copies of all 11 changed files passed. Additional literal path/identity/incident scan on new prose and diagrams found no matches. Reviewed added README/environment prose as well. Generic environment-variable storage contracts are intentional, not private local paths. This is bounded to the candidate delta, not the existing public board.
7. Checks — held. Lychee local with fragments: exit 0, 31 successful targets, 17 external exclusions. External reference check: exit 0, 12/12 targets, one redirect. Pinned full delta git diff --check: exit 0 (includes new files). MissingTargetControl retained a valid documentation link while retargeting one diagram; valid fixture exit 0, mutant exit 2 with exactly the missing SVG and one surviving valid target. This verifies missing-target detection only. Admission and config test commands below both passed.

## Free hunt

No blocking findings. Tested a nonexistent fragment alongside a valid fragment
through Lychee with --include-fragments: expected exit 2, one missing fragment,
one successful target. Rechecked the shell recipe's failed-resolve hazard:
the document explicitly requires successful resolve before login; no claim that
eval propagates the nested command exit is made. Rechecked future Muse native
mapping against the explicit CIP-0017 exception, rather than assuming the older
generic table governs it. No previous reviewer verdict is attached; this review
does not invent a previous accepted revision or count republishing as rework.
The brief supplied no numeric review/free-hunt budget; the sweep is complete.

## Commands and results

All commands below were rerun in this reviewer session. Logs are attached in
TASK-261016-2rpqii_review-checks-rev2.txt with private absolute roots redacted.

| Check | Command | Result |
|---|---|---|
| Local links | `lychee --offline --include-fragments --no-progress README.md docs/environment-config.md docs/agent-logins.md` | exit 0 |
| External links | `lychee --no-progress --timeout 30 docs/agent-logins.md` | exit 0 |
| Render | `plantuml -tsvg -nometadata .temp/review-rev2/render/agent-login-*.puml` | exit 0; four output blobs identical |
| Pinned whitespace | `git diff --check 205272e75ea7d40fd10e073c20ce49bbc0787cd2 bf3995a03bbdc346f1e16625d699f2650d7bd5f7` | exit 0 |
| Working whitespace | `git diff --check` | exit 0 |
| Hygiene | `bash .github/ci/naming-gate.sh .temp/review-rev2/public-scope` | exit 0; exact candidate copies |
| Admission | `go test -count=1 -timeout=3m ./internal/envprofile -run '^Test(CodexIsolatedAdmission|CodexUnknownStoreSharedRefuses|CodexCredentialStoreTOMLSpellings|CodexMalformedStoreSharedRefuses|MuseDirectoriesAndIsolation)$'` | exit 0 |
| Config CLI | `go test -count=1 -timeout=3m ./cmd/curator -run '^TestEnvConfig(ShowDefaults|SetUnsetRoundTrip|Refusals)$'` | exit 0 |
| Missing target control | `lychee --offline --no-progress .temp/review-rev2/control.md` | valid exit 0; retargeted SVG exit 2 |
| Missing fragment control | `lychee --offline --include-fragments --no-progress .temp/review-rev2/anchor-control.md` | exit 2, intended missing fragment |

Admission tests drive the exported Resolve entry used by the CLI. In particular,
unknown-store refusal checks no managed-home creation, then successful retry
with file storage. This review does not claim implementation of future rc.7 gates.

Existing hosted evidence reused after a fresh provider read:
[CI run 38823186484](https://github.com/relux-works/curator/actions/runs/38823186484),
attempt 2, completed success on cd4487462837154c94ae958aca4bea557836d0e1.
20 jobs succeeded; 2 conditional jobs skipped. All 11 candidate blobs match this
snapshot. The producer documents the original failed attempt separately; this
review does not call that attempt successful or infer its timeout cause.
No full hosted suite or real vendor login was rerun by this reviewer.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 2,
  "findings": [],
  "notes": [
    "Vendor measurements are supplied evidence, not fresh credential probes.",
    "Muse Linux login and refresh behavior remain explicitly unqualified.",
    "Checks establish the named bounds only; no proof of defect absence is claimed."
  ]
}
```

## Task-scoped logbook

Reviewed all seven brief surfaces; 7/7 held for the attacks described above.
Independent link controls rejected missing target and missing fragment while
retaining positive controls. Fresh main differs from the baseline only in board
records. No repository source or candidate file was changed. Acceptance is for
the exact revision; landing remains pending and belongs to the delivery owner.
