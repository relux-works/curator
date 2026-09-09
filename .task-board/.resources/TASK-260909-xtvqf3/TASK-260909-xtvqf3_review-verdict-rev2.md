# Independent review of TASK-260909-xtvqf3, CR revision 2

Verdict: **changes_requested**. Route: **to-dev**. No external blocker.
repeat-of: TASK-260909-xtvqf3_review-verdict-rev1.md / R2 (claimed positive wrap coverage absent); R1 (manual adoption example remains non-executable).

CR-TASK-260909-xtvqf3-2 has base 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 and candidate tree ff61be4a8bd43fa4ffb179d31aa38e41891d4313. Independent exact diff exits 0 with no paths. Empty repository delta is the correct scope for this artifact-only gate task: tests, vectors, runner and adoption evidence belong in outcome resources, and production adoption belongs in a subsequent diagnostics revision. It is not acceptance evidence. This verdict does not accept diagnostics CR2 or waive future adoption/review/main integration, defaults/plan/Pi obligations.

## R2 — medium, repeat: fixed-code owners still lack claimed positive forms

The mutable owner matrix is now complete for this exact source: all 10 owned codes have direct/wrapped/joined positives (30/30). The exact prior joined resolve_invocation_failed survivor is killed. However, vectors.coverage.fixed_code_note explicitly claims UsageError and axconfig.Error are pinned as direct/wrapped/joined positives. The gate only supplies direct positives for those two owners (2/6 owner/form combinations). TestGateOwnFamilyAndNilPreserved's broader comment is likewise inaccurate. Nil coverage does not prove positive classification through those wrappers.

Independently inserted a single-family, single-form narrowing at CodeOf entry: if the outer error is a joined chain containing a direct nonnil *cli.UsageError, return no code. Direct/wrapped UsageError and all other owner behavior are retained. Every published diagnostics TestGate test still PASSES, Go exit 0. Independent TestReviewFixedOwnerForms passes baseline (exit 0) and fails the same mutant (exit 1), specifically `usage joined lost: "" false`. The mutation diff and executable probe are in the evidence. This is a demonstrated missing gate, not an inferred production bug: the untouched candidate classifies these forms correctly.

Required: cover both fixed-code owners in all three forms, keep existing nil and mutable-owner coverage, and add a named positive-form narrowing proof to the runner. Derive/assert the full owner-by-form coverage count so claimed combinations cannot silently be omitted. Correct the vectors/comments to match measured coverage. Do not limit rework to the newly remembered UsageError example.

## R1 — low, repeat: manual adoption commands remove the extraction destination

The first adoption block starts `mkdir -p /tmp/cr2-gate && rm -rf /tmp/cr2-gate`, then immediately runs `git archive ... | tar -x -C /tmp/cr2-gate`. It therefore deletes the destination it needs. Independent reproduction with only the path relocated under task .temp fails: tar exit 1, `could not chdir`. The manual block also needs a clearly specified starting directory for the attached files and Git archive source root. The automatic runner path now works; the manual example remains broken.

Required: use a fresh task-local destination that exists for extraction, resolve the Git root and attachment paths explicitly, and execute the published manual block end to end before attaching it. Do not use unconditional deletion of a shared /tmp path as freshness setup.

## Confirmed repairs and actual independent outcomes

The published rev2 filenames were preserved; no renaming or silent repair was needed for the successful runner replay. Invoked with bash from the assigned Git root, with a fresh task-local destination. The runner was byte-identical to the attached resource. Its setup checks now fail closed in the requested five scenarios. The published self-check also exits 0; independent end-to-end negatives additionally validate the real runner route rather than relying on helper-only self-checks.

| Independent check | Actual exit | Outcome |
| --- | ---: | --- |
| Published full runner | 0 | Baseline and all five intended mutants verified |
| Baseline diagnostics / main framing / full two-package suites | 0 / 0 / 0 | 6 named diagnostics + 2 named framing tests pass |
| Resolve admits mcp_layer_missing | 1 | TestGateResolveRejectsForeignCodes fails |
| Layer admits resolve_invocation_failed | 1 | TestGateLayerRejectsForeignCodes fails; prior R3 survivor killed |
| Refusal admits resolve_invocation_failed | 1 | TestGateRefusalRejectsForeignCodes fails; prior R3 survivor killed |
| Joined resolve_invocation_failed rejected | 1 | TestGateJoinedOwnedAccepted fails; prior gate R2 survivor killed |
| One exact complete hostile detail exempted | 1 | TestGateFramingSingleDetailAtRealResolver fails |
| Independent framing companions under that mutant | 0 | Separate named test, all 3 companions byte-exact |
| Published self-check | 0 | Five negative guards trip |
| Independent missing overlays / zero selection / failing baseline / surviving mutant / stale destination | 1 each | Real runner refuses each with its matching GATE-FAIL marker |
| Additional joined UsageError mutant, published TestGate | 0 | SURVIVES: R2 |
| Independent fixed-owner probe, baseline / same mutant | 0 / 1 | Actual missing acceptance coverage confirmed |
| Manual adoption extraction | 1 | Missing destination: R1 |

Expected negative Go exits above are associated with actual named assertion failures, not compilation/setup errors. The independent runner negatives use copied packages under review scratch: missing file, renamed test entry functions, an injected t.Fatal baseline, a no-op first mutation, and a nonempty destination. These are deliberate negative probes, not repairs to the successful published run.

## Source derivation, framing entry and preservation

Independently parsed SPEC section 6 and compared it with all JSON normative/owned/foreign sets. Confirmed 18/18 codes, owned families 6/2/2, normative foreign counts 12/16/16, 44/44 rejection pairs, and 168/168 cases including four extras across three forms. Owner constant declarations in fragment/resolve.go, composition/probe.go and systemprompt/systemprompt.go agree. The Go tests use explicit owner constants independent of diagnostics' allow maps; the normative list and foreign complement are derived at runtime. These are complete for the pinned source, not a claim that future owner declarations automatically enumerate themselves.

Framing mutation is equality on exactly one full detail, preserving other Line logic. Test drives actual `run -> fragment.Resolver.Resolve -> ExecRunner -> diagnostics.Emit -> Line` (main.go resolver line 91, Emit line 99), using a fixed missing executable and hostile profile. No real Curator/ax/model executable is invoked. Three independent companion subtests execute under the exception, fixing prior R3. CodeOf remains API-only in this source, with no non-test production caller; this matrix does not prove full-main launch integration.

Diagnostics baseline tree fbe90d5e60593a3a069721b2ad9e53cd071d8c02 was extracted with git archive. All eight pinned extracted blob OIDs and SPEC SHA-256 5a7ccf0ba95708cb573a586977eb0ba4bad1e46233a540dc99784c1d4d922d48 independently match. Successful runner restores diagnostics.go byte-equal after every intended mutant; the reviewer mutant was restored from saved bytes and verified. Deliberately failing negative runs remain disposable scratch; their abandoned mutations do not touch the baseline or original candidate. Final assigned worktree git status is empty; diff --check exits 0. No original diagnostics workspace writes, ordinary source changes, commits, branch changes, installs, tags, hosted CI or private history mutations.

Read producer briefs, both gate verdict/evidence generations relevant to the rework, current results, tests, vectors, runner and adoption guidance, plus diagnostics repeated-R3 verdict. Independently reran the focused two-package suites through the runner; did not rerun make check, unrelated packages, or claim historical producer CI as new review evidence. Darwin arm64, Go 1.25.5. Evidence was gathered at the pinned diagnostics tree, not inferred from current main.

Evidence resource: TASK-260909-xtvqf3_review-evidence-rev2.txt includes exact commands/exits, all current substantive resource SHA-256 hashes, reproducible reviewer probe source, baseline/mutant and negative-run logs, and preservation checks. Live reviewer checklist leaves acceptance/AC/green-gate completion unchecked. Attach evidence and this verdict before routing to-dev; no accept_cr, commit_ack, done transition, or unsupported reviewer handoff.
