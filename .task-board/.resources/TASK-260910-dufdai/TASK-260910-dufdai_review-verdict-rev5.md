# TASK-260910-dufdai revision 5 independent review

Verdict: CHANGES_REQUESTED. Route to to-dev. One actionable finding within the receipt-3 build/assurance integration scope.

## F1 — external execution receipts bind a different input from the published receipt-3 input (high)

Locations: internal/buildrepo/pipeline.go:311-322,335,366; internal/install/external.go:467-476.

The accepted skillfile-sources.md section 4, lines 289-295, requires assurance permits, execution receipts and checkpoints to bind the exact receipt-3 build input digest in build_input_sha256, and requires refusal when the implementation cannot bind it. The receipt input defined at lines 250-254 is the complete {schema_version:3,package,build} wrapper, with the existing external schema-2 driver input inside build.

RunPipeline derives the external wrapper and cache key, but derives a separate compiler buildInput through Go.BuildInput. The production externalBuildInput constructs a go-v1 schema-1 compiler view with build root "build", wrapped with the package. It omits the external driver's declared/effective identities, locked commit, descriptor target, substitution and other receipt-2 fields. The new equality guard compares only Package. Both cache-hit validation and fresh compilation ValidateFor then validate execution evidence against this different compiler input. Adding the package to the compiler view does not make its digest equal to the exact external receipt-3 input digest.

Concrete reproduction on the unchanged candidate:

    go test -overlay .temp/review-dufdai-r5/install-probe.json -p 1 ./internal/install -run '^TestReviewExternalExecutionBindsExactReceipt3Input$' -count=1 -timeout=120s

Exit 1. The probe drives install.Project using the existing draft build fixture with both build arms, requires successful installation, reads the published external receipt and execution receipt, independently hashes canonical receipt.input, and first verifies that this hash equals receipt.cache_key. It then requires execution.build_input_sha256 to equal the same digest. Actual result:

    successful install accepted wrong execution binding:
    got sha256:1c2be0235173eff8444b80c3bb356e3cd3111c020fa37c11042ce3a89902b3e4
    exact receipt-3 input digest sha256:1d8047f4a67dd5d171afba37708dea80f58a3839905c8be591d564795f7649db

The host fixture's package digest may vary between executions; equality is the invariant. This is a production-entry observation, not a helper-only assertion. No product files were modified: the probe is an overlaid copy of the existing test file.

Required rework: thread the exact external receipt-3 logical input/digest through the assurance/compiler session and validation path, preserving existing evidence shapes and all provider/nonce/toolchain/artifact/checkpoint checks. Refuse a session binding only the compiler go-v1 view even when its package matches. Cover fresh publication and cache reuse with production-entry tests, and retain legacy schema-2 behavior. Do not patch the digest after execution merely to make the record agree. Add a negative row for matching package but wrong full external input; the existing package-blind negative misses this class.

## Candidate and scope

CR-TASK-260910-dufdai-5; base b56089e22075b0f209fbbe75df3178f99864233a; candidate tree 6d631372cc9beff4cc5bf19ba6bd39ee7fea11ec. Reviewed the leaf's 32-path delta against checkpoint bfc0b33, not the accepted sibling changes. All 53 CR working files compared byte-for-byte with the candidate. Workspace remains 32 candidate paths; only ignored review artifacts added. No production/test edits, commits or commit_ack.

Read producer results revisions 1-5, sibling hwxr26 results/verdict rev10 and 17ps6u results/verdict rev4. Their accepted decisions remain outside this finding. Git marker/runtime migration remains an inherited bound, not claimed completed by this review.

## Independent verification

Shell zsh; every test uses -p 1 -count=1. Long processes retained and polled to terminal exit.

- go test ./internal/buildmeta ./internal/buildcache ./internal/buildrepo -run 'Receipt|SourceAware|Adopt|Protected|Prepare' -timeout=180s: exit 0 (0.465s / 5.329s / 22.452s).
- go test ./internal/install -run 'TestDraftBuild' -timeout=400s: exit 0 (94.810s). Covers both arms, cache reuse, package refresh, package-blind refusal, toolchain failure and protected final parents.
- go test -overlay .temp/review-dufdai-r5/probe.json ./internal/buildmeta -run TestReviewReceiptRawShapes -timeout=90s: exit 0. Three valid package controls; 16/16 null/empty foreign-field cases and 9/9 duplicate/trailing-document cases refused across local/network/configured arms.
- Narrowing mutant: LookupArtifact admits receipt version 2 or 3 irrespective of the requested version, retaining input/key/artifact checks. go test -overlay .temp/review-dufdai-r5/mutant.json ./internal/buildrepo -run '^TestExternalReceipt3RefusesEveryEvidenceFieldMismatch$/^receipt_schema_version$' -timeout=90s: exit 1, killed, re-labelled receipt adopted. Independent mutant coverage 1/1, zero survivors; other producer mutants inspected but not independently rerun.
- Production exact-execution-binding probe above: exit 1, finding F1, 22.924s.
- git diff --check: exit 0.

Full repository tests/lint and platform evidence reused from the exact hosted candidate, not rerun locally. Independently queried GitHub run 35344794309 and downloaded complete 924822-byte log. Conclusion success. Head be4a0ed4016d376689263480d9c14c9765585060 resolves to candidate tree 6d631372cc9beff4cc5bf19ba6bd39ee7fea11ec exactly. Ubuntu/macOS/Windows tests, Ubuntu/macOS race, lint, interop, naming and self-tests green. Candidate-suite and rose-air skipped; no passing claim for these lanes. Windows validation is hosted only.

Hosted URL: https://github.com/relux-works/curator/actions/runs/35344794309

Evidence archive contains overlay JSON and overlaid sources, all independent logs, hosted metadata/log. Overlay paths refer to this assigned Story workspace. No full-suite replay or new hosted gate triggered.

Run goal queried: not goal-bound. No directives. Campaign prohibits LOGBOOK.md edits; finding persisted in this outcome and board notes instead. No external blocker or human decision is required; ordinary implementation rework followed by another review.

Additional independent legacy control: `go test -p 1 ./internal/buildmeta ./internal/marker -run 'LegacyReceipt|Golden|MarkerV5Build' -count=1 -timeout=90s` exited 0 (0.473s / 0.427s); pinned legacy receipt/golden and marker-5 build-version checks pass. This does not resolve F1.
