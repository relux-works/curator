# CHANGES_REQUESTED — CR-TASK-260910-hwxr26-2 revision 2

Candidate tree: 249b0d72167f13b98676f1b0634378b4e9cf8909; base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90. All 12 changed worktree files independently compared byte-for-byte to candidate blobs. No repository code modified.

## F1 — P2: policy renewal bypasses evidence integrity

internal/audit/sourceaudit.go:376 returns the renewable policy mismatch before evidence validation at :379 onward. CheckSourceAudit at :749–752 then establishes and overwrites the existing report. This violates skillfile-sources §4: a missing, unreadable or mismatching report fails, and leaves rework F1/F2 incomplete on the renewal path.

Concrete production reproduction: establish a clean local package through install.Project with audit enabled, strict mode and FailOn=high; replace its persisted *.report.json with {}; change trusted FailOn to critical; call install.Project with DraftSourcesV1=true and DryRun=false. Source audit admits the corrupt report and overwrites it. The fixture subsequently fails only at the known sibling marker-schema boundary, proving passage through this gate rather than successful whole installation. Independent TestReviewPolicyRenewalBrokenEvidence fails both refusal and non-overwrite assertions (exit 1). Attached Go overlay contains the exact runnable probe; it replaces internal/install/draftaudit_test.go using go test -overlay without changing candidate code.

Required rework: validate existing evidence integrity/completeness and nonrenewable bindings before permitting policy renewal; do not let the first renewable error hide another refusal. Preserve valid policy renewal and first issuance. Add production regression rows combining policy drift with corrupt/recomputed wrong evidence and invalid time/decision, plus genuine narrowing mutants. Current claimed narrowing variants are input variations, not demonstrated implementation mutants.

## Independent verification

Shell zsh, set -o pipefail; commands completed and exit codes observed:
- go test -p 1 ./internal/audit ./internal/install ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(ParseSourceAudit|ValidateSourceAudit|SourceAudit|CheckSourceAudit|PolicyDigest|DraftAudit|Local|EffectiveLabels)' -count=1: exit 0. Registry mask selected no tests; no registry coverage claimed from this command.
- go test -p 1 ./internal/registry ./internal/install ./internal/interop -run 'Test(CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1 -timeout=60s: exit 0.
- go test -p 1 -overlay /tmp/hwx-overlay.json ./internal/install -run '^TestReviewPolicyRenewalBrokenEvidence$' -count=1: exit 1, expected refusal absent and corrupt report overwritten (0/1 adversarial scenario refused).

Attached hosted validation log reports run 35262500804 success, exit 0. Independently resolved gate commit 857575d5523b057275752c1d8a54880dd553fc71 tree to the exact candidate tree above. Hosted platform results accepted from attached evidence, not rerun locally; rose-air and candidate-suite lanes skipped. Full local suite not run. Broad acceptance/currentness and mutation coverage are not established by these narrow checks; no acceptance claimed.

Lifecycle: spawn goal queried immediately before verdict: run is not goal-bound. Route to to-dev for ordinary implementation rework; no human decision needed. This verdict is also the task-scoped finding record; LOGBOOK.md remains untouched per campaign rules.
