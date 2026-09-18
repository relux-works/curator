# CHANGES_REQUESTED — revision 4

Candidate: 7c39d3bb3315f4e8b0008a44bcd7daf7a1ab4c5b; base 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90. All 12 changed filesystem paths independently compared byte-for-byte with candidate blobs: equal. No repository code modified; adversarial test uses a Go overlay in /tmp.

## F1 — P2: report text can turn a non-renewable failure into renewal

internal/audit/sourceaudit.go:550–560 classifies errors using strings.Contains. The wrong-skill error at line 451 interpolates the untrusted report.Skill. A report with Skill = "is stale; re-run the gates", with its evidence digest recomputed in the binding, produces an evidence rejection that isRenewableAuditError misclassifies as renewable. CheckSourceAudit then calls establishSourceAudit (line 770), overwriting both existing records. The phase ordering does not prevent this classification bypass.

Production reproduction: attached overlay replaces draftaudit_test.go with the candidate test plus this adversarial wrong-skill value and a fresh-time cause. Run:
`go test -p 1 -overlay /tmp/TASK-260910-hwxr26-review-overlay-rev4.json ./internal/install -run '^TestDraftAuditRenewalNeverHidesRefusal/.*wrong-skill$' -count=1 -timeout=90s`
Overlay JSON maps the absolute workspace internal/install/draftaudit_test.go to the attached review-probe-rev4.go (adjust paths when replaying).

Observed exit 1: 3/3 attacks (fresh, stale, stale+policy drift) bypass source-audit refusal and overwrite BOTH binding and evidence. Install stops only at the sibling marker-schema boundary; a subsequent read-only Project returns ok on the replaced evidence. This is not a claim that full installation completed. Expected: wrong evidence fails without overwrite under every renewal condition, as skillfile-sources section 4 requires.

Required rework: use typed/sentinel validation outcomes or an explicit renewal result whose type cannot be derived from diagnostic text. Do not let interpolated caller data select policy or stale renewal. Add production refusal rows for both renewal marker strings in untrusted report fields, with and without real renewal conditions, asserting unchanged records. Preserve valid policy/stale renewal controls. Kill an implementation mutant restoring text-based classification.

## Verification and bounds

Independent zsh narrow test command: `go test -p 1 ./internal/audit ./internal/install ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(SourceAudit|DraftAudit|LocalPackage|EffectiveLabels)' -count=1 -timeout=90s` exited 0. Registry reported no tests matched; no registry rerun coverage claimed. Attached log records results. Adversarial overlay command above exited 1 and is attached separately. Full local suite was not run.

Accepted as attached evidence, not independently rerun: rev4 hosted validation log reports success for run 35273810967, lint, ubuntu/macOS/Windows tests and listed race/conformance jobs. Independently resolved gate commit fcda5e840f5699eed7ebae84b543c29a525fc0f1^{tree} to the exact candidate 7c39d3bb3315f4e8b0008a44bcd7daf7a1ab4c5b. Rose-air and candidate-suite jobs were skipped, not passing. Producer ordering-mutant evidence was inspected, not replayed; the new adversarial inputs demonstrate the remaining gap regardless of the green gate.

Reviewer goal queried: run not goal-bound. Findings persisted in board notes; repository LOGBOOK.md edits prohibited by campaign. Route to to-dev for ordinary implementation rework. No human-only blocker and no acceptance.
