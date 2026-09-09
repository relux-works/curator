# TASK-260909-xtvqf3 CR4 independent review

Verdict: ACCEPT revision 4 only. No blocking findings. Reviewer RUN-260909-697e8c. Scope is the recovered publication described by gate-cr4-review.md, superseding historical development/release/review briefs. Acceptance is not completion or acceptance transfer.

## Actual candidate and remote provenance

Public worktree status independently reports CR-TASK-260909-xtvqf3-4 ready/story_final, producer RUN-260909-ee8fbe (developer/implementer), repository_delta=present, base 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3, candidate tree 489e695df7ada7598347233c6553b791dacebb60, exactly nine diagnostics paths. The patch is 97077 bytes, SHA256 e300d649343c5ae8771f9c841f1bf805b6414ba4d6d0a92cbb2c7035052ffe1d.

From the runtime repository binding, freshly queried ssh://git@github.com/relux-works/curator-agent-launcher with git ls-remote --symref HEAD refs/heads/main, then fetched that exact ref. Advertised HEAD/main, fetched OID and clean Story HEAD all equal 289ff42f037b9f86411fe7852000c466b3fe970d; its tree equals the candidate tree. git verify-commit succeeds with the configured human SSH signature. PR12 is MERGED into main at 2026-09-09T12:23:33Z; its head and merge commit equal that same OID. The actual GitHub COMMENTED review records ACCEPT for this exact head and independently accepted diagnostics CR3 (RUN-260909-88927a); it is not misrepresented as an APPROVED review.

Independently downloaded gh pr diff 12: its complete bytes equal both the attached CR4 patch and git diff --binary from the recorded base to candidate. Thus every source byte and all nine paths have the existing PR12 delivery provenance. No new source change was authored by this recovery. The recorded comparison base remains old, so calling CR4 empty would be false. Parent must use the real landed commit if bound Complete requires it; do not rewrite old base/kind/history or fabricate an empty delta.

## Public recovery and immutable history

Read release-outcome and republish-provenance first. Public release receipt returned exit 0, accepted revision 3 retained and task moved to-dev, after bound sibling checkpoint readiness; a new producer run subsequently published story_final. Independently rehashed frozen historical rev-000003.json: 693e8797d0383a4b50e42c6b34c8783bdd27e08fee79abf3345761786bc8ddcc, exactly unchanged. Read-only record still states the historical accepted task_delta. Current public status independently retains sibling TASK-260909-3d1589 CR3 checkpointed/task_delta/empty, its original reviewer and producer bindings and ff61be4a8bd43fa4ffb179d31aa38e41891d4313 tree. This agrees with its public checkpoint receipt. No private records were written.

## Artifact identity, supersession and adoption

Downloaded each of the five exact resources from both tasks through public resource get. All five pairs are byte-identical, filenames preserved, and match every manifest hash/size:

| Resource basename | SHA256 |
| --- | --- |
| TASK-260909-3d1589_rev2_gate-conformance_test.go | 8495946dce980b9e6a69a17f20fd40703917da51880f4aaa50d7b43f0da754e9 |
| TASK-260909-3d1589_rev2_gate-framing_test.go | 5031d27fdcdb65c3a69ee723baa202d5f928335a26ac225a935b3763810b5f9e |
| TASK-260909-3d1589_rev2_vectors.json | a404a9b703ea8f3fa585b4de45d04f5d8e9079b19129c44fc28164e286ddb76d |
| TASK-260909-3d1589_rev2_run-gate.sh | 6b278b35604507f741ffd03f1f1e3b8ccf57d8c841f69573ab93e479686b224d |
| TASK-260909-3d1589_rev3_adoption.md | 861431c62047aa5968813689a88afada9867bd7089e299260b724786fffb3845 |

Final manifest SHA256 ba6a3fd47ac36c499e64905c179bde5897cb0cb514a9dfa4191a47b8334ff02f also matches accepted original CR3. Inspected registry, vectors, framing entry, runner and adoption composition. Original fixed-owner positive-form gap is covered by five owners, 36 positives and 23 nil cases; mutable-family rejection derives 44 normative foreign pairs / 168 cases including extra codes/forms. Fixed owners have no mutable code field; their wrapped/joined narrowings supplement the three single-foreign-code admissions. Count formulas are arithmetic assertions, not independent execution counters.

Exact adoption staging is <git-root>/.temp/TASK-260909-3d1589/gate/, even when resources are downloaded from xtvqf3. Preserve basenames and chmod +x the runner. Copy conformance overlay to internal/diagnostics/gate_conformance_test.go and framing overlay to cmd/curator-run/gate_framing_test.go. Vectors remain reference. Manual and automatic blocks use the same runner; automatic primary invocation retains || exit $? before self-check. Fresh destinations are required; populated data is preserved. Manifest's <task> placeholder means 3d1589 for verbatim commands; its historical eight-owner wording is a typo: the registry has FIVE owners. These previously accepted nonblocking clarifications remain applicable.

Historical runner pin fbe90d5e60593a3a069721b2ad9e53cd071d8c02 is diagnostics CR2, not current delivered 489e695 tree. Historical manifest empty-delta statements describe CR3, not current CR4. The pinned artifact gate does not automatically test a changed future candidate; future source adoption requires its own exact-source validation/review. Existing PR12 delivery is separately established above.

## Verification actually rerun versus inherited

Reran public downloads, full hashing and byte comparisons, JSON parsing, bash -n, clean-tree/exact-tree checks, git diff --check, fresh remote authority/fetch, commit signature and live PR review/diff reads: all exit 0. Attached evidence includes scripts, commands, outputs and exits. Two exploratory unsupported CLI calls (resources query and task-board cr) returned exit 1; recovered with documented get outcomeResources and public worktree status/resource get. They were not counted as passing validation or absence evidence.

Did NOT rerun make check, Go suites or the full gate: recovery changed no source or gate bytes. Read actual CR4 validation resource (SHA256 8e0b402db9fa84a54a51590eab00301469d914902adf3819871e67a085618ae2): runtime make check, build/vet/tests/race, [exit 0]. This is runtime evidence, not this review's execution.

Explicitly reuse sibling independent review-verdict-rev3.md (SHA256 9fc81165544de5438f8d5ad94baf4f311f1e3843553d6a83d1f27401c0989f14), review-evidence-rev3.txt (e65eddb0cf8b8dfdad5c3db43c3823ade834be063c17de29efbbc8a2a0995216), and preceding rev1/rev2 accepted behavioral evidence for unchanged package. Inspected recorded replay: baseline exit 0; 8 diagnostics + 2 framing named tests; nine named narrowing mutants each Go exit 1, not compiler/setup failures; five runner negatives; restoration; exact automatic stale refusal exit 1 and short-circuit-removal counterexample exit 0. Framing exception exempts exactly one complete hostile detail through run -> Resolver.Resolve (ExecRunner) -> diagnostics.Emit -> Line; three independent companions remain framed. These are inherited attacks, not newly executed attacks. This is not positive-path-only acceptance.

CodeOf remains API-only; PR12 records 15/18 rows at stage APIs and 8/18 through main. Full main/defaults/plan/Pi obligations are not waived. No real ax/model execution or hosted CI green claim.

## Lifecycle and preservation

Live merged checklist inspected: all 12 items checked, supported by current identity checks plus explicitly inherited independent behavior evidence. No LOGBOOK writes per recovery brief. No source/managed branch edits, commits, installs, restarts, tags, CI operations or runtime-home changes. Original resources and candidate preserved. Only task-local scratch and public outcome attachments written.

Accept only CR4 through public accept_cr. Task routes to integrating; reviewer does not set done, handoff, Complete, or commit_ack. Parent routes the bound developer/implementer for Complete and signed Curator board-state publication.

Evidence resource: TASK-260909-xtvqf3_review-evidence-rev4.txt.
