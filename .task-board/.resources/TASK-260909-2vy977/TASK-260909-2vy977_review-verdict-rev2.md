# Review verdict: changes requested; route to analysis

Task TASK-260909-2vy977; CR-TASK-260909-2vy977-2 revision 2 is NOT accepted.
Base: 289ff42f037b9f86411fe7852000c466b3fe970d.
Exact candidate tree: 0d4c177f22f2fa412ddbbb4ccbf4645ceacf4b6f.
Assigned patch SHA256: ea3f343e9ac81262530b89d750258e3651e6035547edb0e402b84d7ed93ed3d9.
Preserved HEAD/checkpoint: ae8676cf03b61ea52ad6af507b04fcca9f62d3f1.
All 14 changed candidate paths independently compared byte-for-byte with the worktree (Python assertion exit 0). No source/index/branch/private-record changes.

## F4 — Unsupported cross-vendor default selection (blocking, new)

Production cmd/curator-run.run calls Files.Complete, whose internal/defaults/lineup.go systemCandidates gathers all pi-native vendors; Complete calls vendorplugin.Lineup(models)[0] on their union and binds its unique contributor. This chooses gpt-5.6-sol/max via pi-openai because 120 exceeds Anthropic 80 and Google 50. Contributor binding preserves the selected row but cannot justify selecting that vendor.

The exact pinned v0.5.11 normative source says:
- pkg/vendorplugin/lineup.go: the derived total order is over a vendor's models; position is presentation, not evidence. LineupOf takes a vendor's own list.
- pkg/vendorplugin/vendor.go CapabilityRank.Score: the scale's own units are the vendor's business.
- pkg/vendorplugin/vendors/local-models/models.go: a capability score is comparable only within one vendor's own lineup.
Thus accepting an arbitrary []Model is mechanical capability, not authority to compare vendors. The candidate itself acknowledges the comparison is not meaningful. Determinism, origins and operator overrides do not establish the missing policy.

Independent scratch probe using real NewRegistry and tagged Lineup yielded:
pi-anthropic top=claude-fable-5 score=80
pi-google top=gemini-3.1-pro-preview score=50
pi-openai top=gpt-5.6-sol score=120
union original=gpt-5.6-sol; google units x10=gemini-3.1-pro-preview
Each vendor's internal ordering is unchanged. This is an explanatory probe on copied model values, not an upstream edit or a claimed gate mutant. Source/log are in .temp/review-rev2/score-probe.go and score-probe.log.

SPEC 4.3 already asks for the highest-ranked admitted model for a system, but does not resolve the now-visible multiple-vendor scale conflict. Rev2 only updates the 4.2 release/implementation paragraph; it supplies no evidence-backed resolution. README adds union policy; that is not an accepted design decision. Do not silently replace it with first-runtime ordering, refusal-only Pi, score normalization, a hard-coded vendor or an invented model.

Required next step: scoped design analysis against the accepted design and upstream contract. If no existing authority selects a Pi runtime, explicitly resolve the product choice (e.g. a reviewed runtime preference followed by that vendor's Lineup, or an explicitly approved cross-vendor policy with its own basis). Then implement and test the chosen policy at run. Retain working native-Pi registration and per-member resolution. This is analysis/rework, not the former missing-tag external blocker.

## F3 — Validation and coverage accounting remain inaccurate (repeat)

repeat-of: review-verdict-rev1.md F3, validation/coverage attestation class.
Producer log RUN-260909-bac28b contains one manual full suite:
make check 2>&1 | tail -25; echo "MAKE_EXIT:${PIPESTATUS[0]}"
Its result reports outer exit 0 and MAKE_EXIT:0, with passing build/vet/normal/race output. Unlike rev1's ambiguous pipeline, rev2 has a printed pipeline status; however the outer exit remains echo's status, and the log excerpt does not independently establish the executing shell. Do not describe it as fail-closed shell exit propagation. This does not suggest the tests failed.
Separately, TASK-260909-2vy977_change-request_rev2-validation.log begins '$ make check' and ends '[exit 0]', with different test timings. This is the credible runtime validation result, reused here. There were TWO full-suite executions in rev2: manual plus runtime, contrary to the explicit runtime-only/no-duplication brief. The outcome's 'make check exactly once' and 'F3 corrected' need an append-only correction; preserve originals. A second same-class finding warrants the orchestrator's focused attestation/accounting gate route under the CR lifecycle skill, not another memory-only instruction to the producer. This reviewer does not spawn that separate task.

The claimed '12 of 12 implementation rows driven' also mixes direct helper tests, production-entry tests and dependency inspection. Own-row recommendation and no-effort tests named in that table directly call Complete; pin verification is not a run test. Supply an honest production-entry denominator, add missing run cases or state precise bounds without declaring owned requirements fulfilled.

## Verified progress and evidence boundaries

Prior F1 release blocker is closed: independent git ls-remote returned exit 0 and exact v0.5.11 tag object 0ea486e46765ecf12fff7c2ac526e12da02e95ed and peeled commit a2a6e9f377f62a5872d99ecdfff0d1690e385f2a. Good configured-human ECDSA signature is accepted from supplied primary/producer evidence, not reverified by this reviewer. go.mod pins v0.5.11 without replace; GOWORK is empty, and the independent narrow test used GOWORK=off. Module API source was inspected directly in the tagged module cache.

Prior F2 registration finding is closed: NewRegistry registers pinative, required gemini/agy systems and Google vendor, and seeds the three real native-Pi runtime declarations. No legacy Pi wrapper substitution. This is real supported module usage, even though the default vendor policy is not justified.

Reviewer independently reran only:
GOWORK=off go test ./cmd/curator-run -run '^(TestRunPiResolvesNativeLineup|TestRunLineupEnvsPrintGroupBeforeRefusal)$' -count=1
Real command exit 0, package 0.415s. No full suite or mutant harness rerun.

**3 of 3 environment origin/completion AC rows driven** in that narrow production-entry run:
| Row | Production call site | Named candidate test |
|---|---|---|
| claude_code defaults/origins | cmd/curator-run.run -> Files.Complete -> EmitGroup | TestRunLineupEnvsPrintGroupBeforeRefusal, claude_code table row |
| codex_cli defaults/origins | same | TestRunLineupEnvsPrintGroupBeforeRefusal, codex_cli table row |
| pi defaults/origins | same | TestRunLineupEnvsPrintGroupBeforeRefusal, pi table row; TestRunPiResolvesNativeLineup |

**2 of 3 environment default-selection rows are normatively supported**: Pi's observed successful completion protects the unsupported union policy, not its correctness. The tests reach the current pending-plan refusal after the group; they do not prove BuildLaunch, real Pi execution, ax or main integration. Those remain stated downstream bounds, not new defects in this leaf.

Per-member flag output is exercised by TestRunPiResolvesNativeLineup. Locks/failures and other negative cases retain named candidate TestRunDiagnosticsContract coverage and attached runtime validation, not fresh reviewer execution. The producer's 34/34 mutant result is retained as reported evidence; no independent mutant certification is claimed in this decisive rejection. Source-text gate applicability is not inferred from prose.

Prior file-resolution defaults.go/defaults_test.go bytes match accepted checkpoint ae8676c (git diff --quiet exit 0), preserving that leaf's evidence. No tag absence is claimed. No failed gate is recast as passing, and green runtime validation is not recast as failing.

## Checklist and lifecycle

Unfulfilled checklist rows (1-based): 1 (authoritative default semantics across all environments), 5 (complete AC implementation), 8 (honest complete production-entry coverage), 16 (implementation matches AC), 17 (architecture fit of cross-vendor selection). Rows 3/4 and existing green build/test evidence remain valid within the stated scope. Row 20 is fulfilled by this attached rejection and public analysis routing.

Preserve the candidate and previous accepted checkpoint; do not integrate rev2. Record one changes-requested verdict and route analysis through set_status. No source edits, hosted CI, real ax, installs/restarts, runtime-home writes, tags/releases, commits, checkpoint or integration performed. Scratch readiness/logs reside under .temp/review-rev2. One log parser initially failed on a non-JSON header (exit 1), then was corrected to skip non-JSON records; a dry-run uncheck_item with index was refused (exit 1), repaired using scoped schema and item. Neither failed read was treated as absence evidence.
