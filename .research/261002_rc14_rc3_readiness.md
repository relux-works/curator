# TASK-261002-1pif8m — rc14-rc3-readiness-checklist

Ready for review. Research snapshot: 2026-10-02, 01:05–01:06 UTC. Intended order: **spec v1.0.0-rc.14 → curator v0.15.0-rc.3 → next launcher tag**. No publication, repository commit, PR comment, or product/configuration edit was performed.

## Decision and principal findings

1. Prepare rc.14 from the seven merged spec changes below. Its version preparation, new release record, generated digest, and downstream qualification are still open. Current green spec CI is candidate evidence, not rc.14 release evidence.
2. Curator rc.3 must pin the **full peeled commit of the published rc.14 tag**, then enable `hashing.EnableV2Writers`. The resulting rc.14 manifest digest is unknown until generation; do not copy the present Muse candidate digest into a release plan as if it were immutable rc.14 evidence.
3. Keep **Codex seed A and security-posture A** in rc.3. Both landed after rc.2 and require an intervening published warning release before B. A spec pin advancement is not that release. Neither B leaf is a blocker for a truthful revision-A rc.3.
4. Do not include B3 cache pruning until macmini-infra repairs and requalifies the paired PRs. They remain open at the exact heads reviewed as CHANGES. If B3 is omitted, explicitly defer it; it is not presently part of main or inherently necessary for the hash/Muse release.
5. **Launcher v0.1.1 already exists**, is a verified annotated tag, and points to current public main, which includes Muse. The local launcher checkout is stale. Of the proposed versions, reserve **v0.2.0** for the next feature-labelled release after metadata/compatibility preparation; never reuse v0.1.1. There is currently zero code delta between public main and v0.1.1. A future patch-only correction could instead use v0.1.2; SemVer does not require a minor bump merely to correct documentation under major zero.
6. Readiness is conditional: 97 declared gap rows accompany the current Muse candidate, the Windows HTTPS incident has no established root cause, and curator main's required test/race jobs were still running at observation. A green scoped implementation job must not become a claim of complete rc.14 conformance.

## Evidence identities and limits

| Repository / object | Verified identity | Observation |
|---|---|---|
| [Spec main](https://github.com/relux-works/curator-spec/commit/e41c561b3300a0a4d3425437fddc5a7048d7b11e) | `e41c561b3300a0a4d3425437fddc5a7048d7b11e` | Local main/HEAD and authenticated GitHub main agree; 9/9 observed check runs successful. |
| [Curator main](https://github.com/relux-works/curator/commit/2247509a476423bfdd9dd57a5098056fdd2be8cb) | `2247509a476423bfdd9dd57a5098056fdd2be8cb` | Assigned worktree HEAD and authenticated main agree; 6 successful, 5 running, 2 optional/candidate skipped check runs. |
| [Launcher public main](https://github.com/relux-works/curator-agent-launcher/commit/1ac7eafba38d2a62bb679602401fb0db2f2e8e56) | `1ac7eafba38d2a62bb679602401fb0db2f2e8e56` | 8 successful and 2 optional skipped check runs across observed runs. Local main is only `d0920353556dd0a3cea2864c616c230915977bc7`; remote files were read at the full public SHA. |
| Launcher [v0.1.1](https://github.com/relux-works/curator-agent-launcher/tree/v0.1.1) | tag object `7db30dd6b1e11e36826895ab75d7f32ef559a857`; target `1ac7eaf…` | GitHub signature verified; compare v0.1.1...main is identical, 0 commits ahead/behind. |
| Released [spec rc.13](https://github.com/relux-works/curator-spec/tree/v1.0.0-rc.13) | peeled commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` | Published suite digest `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`. |
| Current spec candidate | manifest SHA-256 `bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783` | Still labels protocol rc.13; contains hash-v2 plus Muse. Not the published rc.13 corpus. |

[Machine-readable evidence](TASK-261002-1pif8m_evidence.json) records timestamp, API command argv, real retrieval exit codes, check URLs, board scope totals, and declared gap counts. Board reads were performed via `task-board q`, not by reading board files. No release tests were rerun in this read-only research task; prior validation is explicitly attributed below. Running CI is unknown, not success. Snapshot freshness must be checked again before merging or tagging.

No `docs/release*` file exists in the inspected curator main tree. The operational references are [docs/ci-gates.md](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/docs/ci-gates.md), the release workflow, source/workflow gates, GoReleaser config, and release assets. This is a successful tree inventory, not an inference from a failed read.

## Spec rc.14: merged payload since rc.13

`git log v1.0.0-rc.13..main` establishes seven merged commits; [Unreleased](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/CHANGELOG.md) describes them.

| Merged change | Commit / PR | Release content and bound |
|---|---|---|
| Source-suite independence | `17d8879`, [#99](https://github.com/relux-works/curator-spec/pull/99) | Gate skillfile-sources schema references/citations against rc.10; only the precise conditional environments §9.4 citation is exempt. |
| Global lock publication | `4ad8042`, [#113](https://github.com/relux-works/curator-spec/pull/113) | Publish extended profile lock before materialization; retain lock on unmanaged conflict; transactional rollback and `profile sync --takeover` recovery vectors. |
| Repair persistence and MCP asymmetry | `2b649c4`, [#115](https://github.com/relux-works/curator-spec/pull/115) | Clarifies replay of passing store entries on stale-home repair and Claude strict-MCP versus Codex merged-profile behavior; no requirement/vector change claimed. |
| Length-framed content hashes | `b1a2efb`, [#116](https://github.com/relux-works/curator-spec/pull/116) | `curator-content-v2`; hash-version identities, equal-version registry matching, blocking NUL guard for v1; eight new carrier schemas listed below. Source-suite carriers deliberately deferred. |
| New-driver threat review | `add5023`, [#118](https://github.com/relux-works/curator-spec/pull/118) | Informative checklist linked to existing normative requirements. |
| Adopt Decisions 0019/0021 | `0400fea`, [#120](https://github.com/relux-works/curator-spec/pull/120) | Records operator adoption and backlinks. Normative/launcher amendments remain follow-ups; does not itself deliver an ax bypass or session-entry redesign. |
| Muse environment | `e41c561`, [#121](https://github.com/relux-works/curator-spec/pull/121) | Muse adapter, four XDG parents without HOME replacement, settings/trust seeds, shared file-link credential liveness, refusal boundaries and v3 fragments. Refresh/context surfaces and foreign personal-context isolation remain unverified. |

New hash carriers: `install-marker-v5`, `context-lock-v2`, `audit-record-v2`, `agent-environment-marker-v3`, `registry-log-entry-v2`, `registry-bundle-v2`, `manager-config-v3`, `log-response-v3`. Muse adds `launch-env-fragment-v3`. Comparing tag to main found **9 added `.schema.json` files and zero modified pre-existing `.schema.json` files**; `schemas/v1/README.md` changed. That is a bounded Git comparison, not a fresh cross-platform schema test.

### Required and optional work before rc.14

- [ ] **Required:** release-preparation leaf updates active version declarations, new rc.14 record, generation/gates/tests, compatibility/security notes, and changelog heading. There is no rc.14 release record/tag in the observed inputs.
- [ ] **Required:** retain published schema/claim/metadata history. Main currently regenerates `release/1.0.0-rc.13.json` with candidate digest `bd03456…`; the tagged record pins `be11bb1e…`. For rc.14, restore/preserve the tagged rc.13 bytes, emit a separate rc.14 record, and prevent the generator from rewriting historical records. This is release-preparation work, not permission to move rc.13's tag or assets.
- [ ] **Required:** qualify actual generated rc.14 bytes against full-SHA implementation pins and each implementation's declared coverage; exact case-count tables and digest dispatch must recognize the newly versioned manifest.
- [ ] **Required for a full-conformance claim:** close measured implementation gaps. The normative release checklist says full suite, no skips, three OSes. Current [COMPATIBILITY.md partial-client boundary](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/COMPATIBILITY.md) and implementations workflow exercise a narrower declared scope. Document this explicitly in rc.14, keep unsupported claim sets empty, and do not interpret that scoped CI as proof of the broader checklist. If the release owner requires the literal full-client gate, it remains open; do not silently waive it.
- [ ] **Conditional B3 scope:** macmini-infra repairs curator #101's remaining link-free reference-discovery failures; spec #119 advances its consumer pin to the repaired head, rebases/refreshes onto current Muse main, then repeats paired qualification and review. Spec #119 is currently merge-conflicted (`mergeable=false`, `dirty`). Include only after this work; otherwise defer both PRs together.
- [ ] **Should defer:** TASK-260930-3ny11n — spec-skillfile-sources-content-hash-v2 (backlog). It requires lockstep with the csk owner. Current source suite stays at digest `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`, compatible rc.10 core. No need to bundle this separate schema/version transition into rc.14; no `skillfile-lock-v2`, `install-marker-v6`, or `source-audit-v2` claim until delivered.
- [ ] **Should track separately:** normative follow-ups to Decisions 0019/0021. Shipping adoption records must not imply those follow-ups are implemented.

### Exact release procedure and version files

[RELEASE.md](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/RELEASE.md) is normative; [release workflow](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/.github/workflows/release.yml), [generator](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/tools/generate-vectors/main.go), and [release gate](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/tools/release_gate.py) supply the concrete mechanics. RELEASE.md itself does not enumerate every bump file; this inventory is derived from those files.

1. Prepare `README.md` active Version/release links, `CHANGELOG.md` rc.14 dated heading, `COMPATIBILITY.md`, and `SECURITY.md` compatibility/security impact notes. Advance active `PROTOCOL_VERSION` in `tools/validate.py` and `tools/release_gate.py`, and `protocolVersion`/release writer/date in `tools/generate-vectors/main.go`. Adapt current-record assertions in `tools/test_validate.py`, `tools/test_release_gate.py`, and generator tests while retaining historical assertions. Do not globally replace historical rc.13 or rc.9 identifiers.
2. Generate `conformance/v1/manifest.json` and vectors as rc.14. Add `release/1.0.0-rc.14.json` containing exact generated manifest SHA-256 in `candidate_protocol_pin` and `downstream_consumption`, source-suite digest/baseline, assurance identities, and truthful empty unsupported implementation/platform/claim sets. Preserve portable default, explicit fail-closed verified mode and deferred hardened guarantees; `security_posture=hardened` is not an attestation of the reserved hardened execution policy. Preserve historical claim protocol/schema identities instead of automatically minting rc.14 claims.
3. Update generated-file coverage in `Makefile` and `.github/workflows/release.yml` to include the new record. Current explicit diff lists end at rc.13; leaving them unchanged would fail to check rc.14 drift. Preserve every released schema and published historical record byte-for-byte, including rc.13, against tagged provenance.
4. Run `make validate` on Linux/macOS/Windows; `make regenerate-check`; independently digest the suite and check downstream digest verification before consumption. Qualify Go curator, csk's declared rc.10 core plus source suite, and registry at full commit pins; examine no-skip, measured coverage and immutable native evidence for every actual claim. Candidate acceptance alone is not a release claim.
5. Before integrating the reviewed release PR, run `python tools/verify_release_merge_policy.py` with its documented credentials/options. Repository settings must enable squash and disable rebase/merge commits. Required review/checks precede GitHub squash merge. Verify the resulting signed squash commit and its main Release target provenance job (`tools/verify_release_commit.py`). Check active version with `python tools/release_gate.py --version 1.0.0-rc.14 --commit <release-commit>` (or `make release-check VERSION=1.0.0-rc.14` on that exact target). These commands are proposed, not executed here.
6. **Operator “да”:** create a signed annotated `v1.0.0-rc.14` tag on that accepted main squash commit, verify against `maintainers.allowed_signers`, then push the tag. Never tag a PR/unmerged branch or rewrite an existing tag. The tag-push workflow verifies tag, fresh main provenance, validators/tests/regeneration, and the version release gate before packaging.
7. Inspect the published prerelease: `curator-protocol-1.0.0-rc.14.tar.gz`, `.zip`, `checksums.txt`, and build-provenance attestations. Archives include normative docs, schemas, corpus, decisions, release and review records. Verify immutable assets/checksums and record tag-object, peeled commit and manifest digest for curator promotion.

Stable 1.0.0's independent security/interoperability review reports, different reviewer identities, no open critical/high findings, and reviews-only post-freeze rule are **stable-release requirements**, not a reason to fabricate reports for rc.14.

### Implementation pins

Current [implementations.yml](https://github.com/relux-works/curator-spec/blob/e41c561b3300a0a4d3425437fddc5a7048d7b11e/.github/workflows/implementations.yml) pins:

| Consumer | Full pin | Scope / readiness |
|---|---|---|
| Go curator | `e4a6a8d5df879183c4d8b5c648aa440bd50ff7dc` | Its required hosted jobs observed green; accepts rc.13/hash-v2/Muse candidate identities. A new rc.14 digest still requires count/digest updates and requalification. |
| csk | `4a88aa0e47ab2b34ded16162cec7f97d0c7cf6f1` | Own observed merge jobs green; declares rc.10 core plus separately versioned source corpus, not full rc.14. |
| Registry | `d690bea6fab1c8e6392e05d3a3cdfcf1168bc914` | 9 observed jobs green; a pin advance is needed if its actual candidate acceptance changes. |

There are no published rc.14 identities yet. Prefer retaining known-green pins only when their declared consumers accept the exact rc.14 suite. If curator needs a new qualification commit before spec publication, qualify it as a candidate **with the v2 writer still OFF**, advance the spec consumer pin to it, then publish spec; the later curator pin/write cut-over follows. This avoids a publication dependency cycle.

## Curator rc.3: payload, pin and release gates

### What should ship

Use the merged tree after rc.2, not every line currently under Unreleased: that section includes older carried entries and omits some newly landed work. Source history from `v0.15.0-rc.2..2247509a` includes:

- Security remediation: protected profile/store boundaries; nofollow managed writes/path kind and disappearing-path handling; absence-versus-read-failure fixes; direct-only system modules; bounded MCP declarations; trust-root provider warning A; scoped source signer/update checks; registry TOFU/equivocation protections and record boundaries; Codex seed A and machine security-posture A; HTTPS secrets via broker pipe and Unix refusal-path EPIPE correction.
- Profile/environment work: global adopt and publication ordering, backup/unmanage restore, 0018 permission fragments, first-run/operator guidance and reinstall fixes.
- Build work: external build repository delivery and acquisition consumers; native/scriptworker identity/toolchain follow-ups and hosted-lane corrections.
- Hash-v2 reader/carrier implementation with writer disabled; Muse environment adapter and v3 fragments; non-git product-root install with a hygiene notice and refusal of unexpected Git errors.

Evidence: [curator history comparison](https://github.com/relux-works/curator/compare/v0.15.0-rc.2...2247509a476423bfdd9dd57a5098056fdd2be8cb), [current changelog](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/CHANGELOG.md). Release-prep must reconcile this history into an accurate rc.3 entry; R1–R5 prose alone does not prove all those features are newly introduced after rc.2.

### Pin and switch checklist

- [ ] Promote `.github/workflows/ci.yml` `SPEC_PIN` from rc.13's `23435129…` to the **full rc.14 tag target**, with the exact released manifest digest verified. Update `docs/ci-gates.md`, exact-count/gap rows for the new digest, artifact/family ledgers and digest-dispatch code as required. Record tag object and SHA in evidence.
- [ ] Flip the one write cut-over, [`internal/hashing/hashing.go:47`](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/internal/hashing/hashing.go#L47), from false to true **only with the published rc.14 pin**. `WriteVersion()` is used by install targets/publication, markers, context resolution/materialization/store/locks and environment marker/migration publication. Exercise real entry points and legacy-reader/NUL/mismatched-version negatives; test-seam success is not evidence of the shipped default. New identities must not silently alias old ones.
- [ ] Reconcile candidate ledgers: current Muse digest has **97** gap rows: **81** owned by TASK-260917-2tx81l — manager-content-hash-v2 (board done, but inherited gaps remain), **9** TASK-260927-1e5qqm — flip-codex-seed-to-revision-b, **5** BUG-260923-2afgyq — marker-reader-cross-field-validation, **2** STORY-260925-1v7pvn — windows-exec-hardlink-origin-validation. This measures rows, not a pass ratio or 97 distinct absent features. Remove only rows whose exact production case is now green; pin promotion cannot simply drop the ledger.
- [ ] Keep `CodexSeedRevision=A` and `SecurityPostureRevision="A"`. TASK-260927-31gaka — ship-codex-seed-revision-a and TASK-260927-4pv4au — manager-security-posture-revision-a are board done and landed at `7444178d` and `3f60f7f0`, both after rc.2. Thus rc.3 is the first currently planned tagged warning release containing these implementations. Their B leaves require that shipped release, not merely done prerequisites.
- [ ] Keep warning-stage hook/passthrough/provider gates as documented; no evidence authorizes a blanket A→B flip at rc.14. Update confirmation is already reported `B-flip` by status code, not a new rc.14-triggered change. Portable assurance remains default and no reserved hardened execution claim is implied.
- [ ] Preserve separate immutable module release pin: [`internal/buildrepo/release_pin.go`](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/internal/buildrepo/release_pin.go) remains rc.8 (`f8c405aa3ad0a39d260c2ed93684e55c5a346359`, tag object `ad247840292487d5d88ac44331798b6b4182a79f`). This is independent of CI `SPEC_PIN`; advancing it requires separate release-pin qualification and must not be an incidental search/replace.

### Open leaves and blockers

Board traversal included tasks and bugs recursively through both requested epics, including closed history. EPIC-260910-2hw1xb — security-remediation has **68 leaves**: 62 done, 1 closed, 3 backlog, 1 blocked, this 1 analysis research leaf. EPIC-260910-16qce1 — registry-security-remediation has **12/12 leaves done** and no active leaf. Counts describe board state, not fresh test coverage.

| Leaf / evidence | Observed status | Release classification |
|---|---|---|
| TASK-260930-fp8vx7 — diagnose-and-fix-windows-broker-realgit-flake | blocked | Open release risk: two historical real-Git failures; diagnostics snapshot had 21,000 hosted passes, no reproduced failure, no established root cause/fix. Full Windows CI on the actual candidate is required. Root-cause closure needs the external failing instrumented run or reproducible runner conditions; shipping with the incident open requires an explicit operator risk decision, not a false “fixed” statement. |
| TASK-260930-3ny11n — spec-skillfile-sources-content-hash-v2 | backlog | Separate lockstep source-suite transition; defer if existing source suite remains unchanged. Blocks claims/new source carriers, not core rc.14 publication by itself. |
| TASK-260927-1e5qqm — flip-codex-seed-to-revision-b | backlog; depends on seed-A leaf | Post-rc.3 release work. Do not land before the intervening warning release. |
| TASK-260927-25hk87 — flip-security-posture-default-hardened | backlog; depends on posture-A leaf | Post-rc.3 release work; warning-release prerequisite still matters although implementation parent is done. |
| TASK-261001-2d1obt — review-b3-cache-prune-prs | to-review | Outcome says CHANGES; board handoff is not PR acceptance. Conditional blocker if B3 is in scope; otherwise defer both PRs. |
| BUG-260923-2afgyq — marker-reader-cross-field-validation | backlog | Five invalid v3/v4 external-repository cases admitted; security/conformance debt. Recommended pre-rc.3 fix, mandatory before claiming those cases/full conformance. |
| STORY-260925-1v7pvn — windows-exec-hardlink-origin-validation | backlog; no children returned | Two named executable-origin gaps; split into two bounded implementation leaves. Recommended pre-release fix, mandatory before claiming those security cases/full conformance. |

Actual unconditional publication blockers are release-prep/pin/write-cutover qualification and green required CI; conditional blockers are B3 inclusion and full-conformance claims. The Windows unresolved incident needs a recorded risk disposition before publication. This report does not manufacture hard board dependency edges that are absent.

B3's exact heads are spec `51eb12008aae017e59a785d3d9965dadcbc86aef` and curator `d0e486f8f900995b19aa6c3192d4e0309e5047d8`. The latest [round-2 outcome](board-resource://TASK-261001-2d1obt/outcome/TASK-261001-2d1obt_verdict-r2.md) independently reports two actual CLI failures: linked `.agents` marker ancestor and linked `consumers.json` allow deletion under an uncertain reference boundary. Its 17/17 decision cases and 7/7 killed mutants do not cover these remaining paths. The spec pin must follow the repaired curator commit; merge **spec first, curator second**, preserving candidate qualification in lockstep. No GitHub review rows were returned; CHANGES comes from board evidence, not an invented GitHub review state.

### Curator publication steps and assets

1. Prepare rc.3 changelog and docs, integrate reviewed pin/write-switch fixes, qualify the exact combined candidate on Linux/macOS/Windows plus required race/lint/coverage/gate jobs. Current main CI is still running in the evidence snapshot. Optional rose-air/candidate skips are not mandatory-platform passes.
2. Check [release source gate](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/.github/ci/release-source-gate.sh): no `replace`/`exclude` directives, and tag target is contained in freshly fetched public main. Keep negative module/ancestry and workflow-order checks. Set version through GoReleaser ldflags (`internal/version.value=v{{ .Version }}`); curator has no fixed rc.2 source constant to bump.
3. **Operator “да” publication boundary:** rc.3 release intent is already approved in the assignment. Record that standing authorization against the concrete accepted SHA/assets; do not ask for the same permission again. Signed tag/push/release execution belongs to the release operator/coordinator, outside this research task. New scope/risk acceptance is separate if required.
4. Push `v0.15.0-rc.3` on accepted main. The [Release workflow](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/.github/workflows/release.yml) triggers on `v*` tags, checks out full history/submodules on macOS, installs Go, fetches public main, runs source/ancestry gate, installs cosign/Syft, executes `goreleaser release --clean`, then attests build provenance. It does not itself rerun the full test suite: pre-tag exact-SHA CI is necessary.
5. Inspect [GoReleaser](https://github.com/relux-works/curator/blob/2247509a476423bfdd9dd57a5098056fdd2be8cb/.goreleaser.yml) output: six archives, darwin/linux amd64+arm64 `.tar.gz`, windows amd64+arm64 `.zip`; LICENSE/NOTICE/README inside; six archive SBOM JSON files; four Linux deb/rpm packages; `checksums.txt`, `checksums.txt.sig`, `checksums.txt.pem`; provenance for archives/packages/checksums. macOS signing/notarization is conditional on configured credentials, so cannot be attested from source alone.
6. Verify prerelease=true, draft=false; Homebrew/Scoop `skip_upload:auto` must keep rc.3 from replacing stable channel metadata. Validate provenance using `gh attestation verify <artifact> --owner relux-works`, and checksums signature with cosign using the workflow identity/OIDC issuer shown in release notes. Verify the exact published downloads, not an old local dist directory. The [published rc.2 assets](https://github.com/relux-works/curator/releases/tag/v0.15.0-rc.2) confirm this 19-asset packaging pattern; rc.3 assets do not yet exist.

## Launcher tag, install line, and compatibility

### Version recommendation

**Next requested feature-labelled tag: v0.2.0.** `v0.1.1` is already occupied by signed tag `7db30dd6…` targeting public main `1ac7eaf…`. That tree already carries Muse interactive plans, v3 fragments, agents-management v0.5.37 and shared typed context construction. It also still has `buildVersion="0.1.0"`, an Unreleased feature section, and a README install line at v0.1.0. A new reviewed version/documentation preparation commit is necessary; simply tagging the same bytes as v0.2.0 would preserve the wrong reported version.

This minor-version recommendation communicates the new feature set relative to 0.1.0. It is a release-policy recommendation, not a requirement to pretend those features are untagged: v0.1.1 already contains them. [SemVer](https://semver.org/) treats 0.y.z as initial development and suggests minor increments for successive initial releases; v0.1.2 remains possible for patch-only corrections if the operator chooses that narrower scope. Do not move/delete v0.1.1.

Required README line for the proposed next tag:

```sh
go install github.com/relux-works/curator-agent-launcher/cmd/curator-run@v0.2.0
curator-run --version
```

Update `cmd/curator-run/main.go` buildVersion to `0.2.0`, retain the separately owned `specVersion="0.5.0-draft"` unless the launcher contract itself changes, add a dated 0.2.0 changelog entry, update README version text and version goldens. Fix README's later supported-environments sentence, which still lists only Claude/Codex/Pi despite its new Muse paragraph. Go toolchain comes from go.mod (currently 1.25.5); no replace/workspace override. Install in an operator-owned trusted provider directory, not curator's user-bin shims, managed skill bins, or environments root.

### Compatibility with rc.3

| Interface | Observed matching contract | Required evidence before claiming compatibility |
|---|---|---|
| Fragment resolver | Launcher calls `curator env resolve --repair --format json`; reads v1/v2/v3; Muse manager emits v3 and four XDG parents, preserving HOME. | Installed rc.3 → proposed launcher smoke at actual entry points, with digest/read-error/wrong-version refusals. |
| Existing providers/context | Canonical Claude/Codex/Pi remain; Claude/Codex channels use shared agents-management construction, reserved `path_prepend` remains non-operative and digest-covered. | Direct/tracked entry-point goldens, noncolliding pass-through and colliding prompt/MCP/profile negative cases. |
| Permission transport | v2/v3 `{mode,locked,source}`; explicit yolo, force-native and tracked-yolo fail-closed rules preserved. | Exercise real curator fragment emission and launcher resolution, including headless and locked rejection. |
| Muse | Public README: plugin probes/admit verified releases 1.4.1/1.4.2; native no posture flag; yolo exactly one `--yolo`; no MCP/prompt carrier. Curator registry specifically verifies `1.4.1-R4503.1`, with unverified-release warning boundaries. | Use common verified release 1.4.1 first; independently test 1.4.2 before extending manager verification. No fresh-login isolation, refresh, personal-context isolation, or unlisted release claim. |
| Umbrella discovery | Curator resolves `curator-run`; trust-root warning A and refusal of manager-published/managed directories remain. | Installed `curator run` discovery in trusted provider directory; refuse managed shim locations. |
| ax tracking | Optional `ax start … --launch-plan` remains. Decision 0021's desired per-launch untracked policy is not delivered by adoption records. | Verify installed ax bridge; fake-ax tests do not attest real daemon behavior. Keep current no-fallback/no-bypass statement. |

Source compatibility is well supported; actual installed rc.3/new-tag compatibility remains **unverified** because neither proposed release pair exists. Public launcher CI covers Linux/macOS; it does not establish native Windows launcher support. No prebuilt launcher release workflow exists in the inspected tree: present delivery is the versioned Go module. [Public README](https://github.com/relux-works/curator-agent-launcher/blob/1ac7eafba38d2a62bb679602401fb0db2f2e8e56/README.md), [CHANGELOG](https://github.com/relux-works/curator-agent-launcher/blob/1ac7eafba38d2a62bb679602401fb0db2f2e8e56/CHANGELOG.md), [go.mod](https://github.com/relux-works/curator-agent-launcher/blob/1ac7eafba38d2a62bb679602401fb0db2f2e8e56/go.mod), and [CI](https://github.com/relux-works/curator-agent-launcher/blob/1ac7eafba38d2a62bb679602401fb0db2f2e8e56/.github/workflows/ci.yml) are the primary evidence.

## Ordered leaf plan

Effort is a rough engineering estimate, excludes queue/review latency, and is not a booked run. P1–P19 are proposed leaves, not newly created board items. Release preparation must use managed review/integration; this researcher does not execute it.

| Order / leaf-sized output | Depends on | Rough effort | Parallelism / approval |
|---|---|---|---|
| P1 — freeze release scope and measured claim boundary; decide B3 in/out and Windows incident disposition | This report | 30–60 min | Coordinator/operator; any new incident-risk acceptance needs “да”. Existing rc.3 release authorization persists. |
| P2 — repair B3 linked marker ancestor; actual CLI preservation/refusal regression and narrowing mutant | P1, only if B3 included | 1–3 h | macmini-infra; may run with P3–P6 on isolated leaves. |
| P3 — repair B3 linked consumer registry boundary with same evidence shape | P1, only if B3 included | 1–3 h | macmini-infra; coordinate shared files with P2. |
| P4 — implement five marker cross-field negatives at production reader | P1's claim scope | 2–4 h | BUG-260923-2afgyq; parallel security leaf. Mandatory for its claims; recommended before rc.3. |
| P5 — validate Windows executable platform ownership | P1's claim scope | 2–4 h | First leaf under STORY-260925-1v7pvn; native Windows evidence. |
| P6 — validate Windows extra hard-link component-store origin | P5 boundary design | 2–4 h | Second leaf; serialize shared resolver edits, qualification may overlap other packages. |
| P7 — B3 paired pin refresh/current-Muse corpus qualification and independent re-review | P2, P3 | 1–3 h + hosted CI | Update spec consumer pin to repaired green curator head; spec merge first, curator merge second. No release tags here. |
| P8 — preserve tagged rc.13 metadata/schema history and add rc.14 generator/validator/gate support | P1; P7 only if included | 2–4 h | Spec release owner; can prepare alongside independent security fixes. |
| P9 — generate rc.14 record/corpus/digest; update exact downstream candidate count dispatch and pins, writer OFF | P8; any selected P4–P7 payload | 1–3 h | Candidate manager qualification commit can precede spec tag; csk source extension remains deferred unless jointly selected. |
| P10 — qualify rc.14 spec, regeneration, declared implementation coverage and native claims on three OSes | P9 | 30–90 min worker + CI | Independent OS lanes parallel; split local long gates into bounded subsets, preserve each exit code. No skips/gaps relabelled green. |
| P11 — review release PR, verify merge policy, squash merge, signed target/provenance and release gate | P10 | 30–60 min + review | Maintainer/coordinator. Fresh authority check before integration. |
| P12 — signed spec rc.14 tag/publish and verify archives/digest/provenance | P11 | 15–30 min + CI | **Operator “да” needed** for spec publication; subsequent asset verification read-only. |
| P13 — curator committed rc.14 SPEC_PIN + v2 writer cut-over, updated ledgers/docs, real writer/migration regressions | P12; selected fixes | 2–4 h | One coupled leaf; retain A seed/posture, retain separate rc.8 immutable module pin. |
| P14 — curator rc.3 changelog accuracy and release-channel/module packaging checks | P13 | 30–90 min | Can prepare draft notes earlier; validate against actual merged tree. |
| P15 — exact combined curator CI and installed candidate smoke; inspect Windows incident evidence/disposition | P13, P14, P1 disposition | 30–90 min worker + CI | Three OS lanes and race lanes parallel. Any red result returns to its owning leaf. |
| P16 — signed curator rc.3 tag, publish 19-asset pattern and verify prerelease/stable-channel isolation | P15 | 15–30 min + CI | **Operator “да” already supplied for rc.3 release intent**; apply to concrete reviewed target, no duplicate request. |
| P17 — launcher v0.2.0 version/changelog/README consistency and tagged-module checks | P1 version scope | 30–90 min | Can run while spec/curator gates run; starts from public main, not stale local checkout. |
| P18 — launcher required Linux/macOS gates plus installed rc.3 compatibility matrix | P16, P17 | 1–2 h + CI | Existing provider and Muse cases can run in parallel; real ax verification separately when available. |
| P19 — signed new launcher tag and verify versioned Go install/version/umbrella smoke | P18 | 15–30 min | **Operator “да” needed** for launcher publication. Current workflow publishes no binaries automatically. |

Critical publication path: P8→P9→P10→P11→P12→P13→P14/P15→P16→P18→P19. P17 overlaps it. If B3 is selected, P2/P3→P7 precede P8's scope freeze. Security fixes P4–P6 precede whichever release/claim includes them. External Windows root-cause work cannot be assigned a credible finish time without a failing diagnostic run; no amount of repeated green execution proves causality.

**After rc.3 is demonstrably published**, separately schedule TASK-260927-1e5qqm — flip-codex-seed-to-revision-b and TASK-260927-25hk87 — flip-security-posture-default-hardened, each with real production-default negatives and updated gap rows, for a subsequent curator release. TASK-260930-3ny11n — spec-skillfile-sources-content-hash-v2 remains paired with its independent csk owner and may proceed in parallel only on a separately versioned source-suite release path.

## Verification ledger and task logbook

Fresh research checks: authenticated source/commit/tag/PR/check-run reads, recursive board projections, source history/schema diff, record/digest inspection and declared-gap counting. Evidence collector exited **0**; its 17 API subprocesses record individual exits: **15 exited 0**, and the rc.14/rc.3 tag-ref reads each exited **1 / HTTP 404**, expected because those exact tags were absent in accessible repositories at observation. These are retrieval failures with stated interpretation, not passing gates. Preliminary invalid board `resources`/`search` projections exited 1 and were replaced with supported projections. Initial missing local paths/unsupported source directories produced exit 2 searches and were corrected through tree inventory. An initial csk request used the wrong repository and returned 404; the actual `ivanopcode/cocoaskills` pin was then read successfully. None counts as release validation.

Document verification: standalone Python consistency check exited **0**, independently asserting 7/7 merged spec commits, 9/9 added schema files with no pre-existing schema modifications, 17/17 recorded API exits, both board scope totals, the 97-row owner distribution, all P1–P19 steps and current write/seed/posture switches. `git diff --check` exited **0**. The report and JSON are new untracked files; the Python check read them explicitly, while Git's diff check covers tracked changes only. These checks establish report consistency, not release qualification.

Accepted prior evidence, not rerun here:

- [B3 round-2 verdict](board-resource://TASK-261001-2d1obt/outcome/TASK-261001-2d1obt_verdict-r2.md): local validation/generation targeted passes exit 0; 17/17 decisions; 7/7 mutants killed with real exit 1 (expected-red); two current-head production CLI probes exit 1 (actual defects); broad transaction package command exit 1 on timeout. No green full-suite replay inferred. PR heads freshly match the reviewed heads; upstream spec main has since advanced to Muse and is now conflicted.
- [Windows HTTPS results](board-resource://TASK-260930-fp8vx7/outcome/TASK-260930-fp8vx7_results.md): local selected tests/vet/Windows cross-compilation/lint/diff checks exit 0; hosted 21,000 stress passes are a stability observation, not reproduced incident/root-cause evidence. Native Windows mutants were not rerun in that task.
- [Hash-v2 revision-4 results](board-resource://TASK-260917-2tx81l/outcome/TASK-260917-2tx81l_refresh4-results.md): exact counts and real command evidence for the accepted implementation, with shipped writer still OFF. That acceptance does not qualify a future rc.14 pin/cut-over by itself.
- Current hosted check conclusions are preserved with URLs in the evidence JSON; successful retrieval exit 0 is not the remotely executed test's locally measured exit code.

Task logbook, 2026-10-02:

- **Finding:** publication identities lag candidate changes. Preserve published rc.13 history and use a new rc.14 record/digest; do not confuse main's `bd03456…` with tagged `be11bb1e…`.
- **Finding:** 97 declared Muse-candidate gaps remain despite the content-hash implementation owner's board-done status. Promote writer/count ledgers through actual production checks; do not infer absence of debt from board closure.
- **Finding:** seed/posture A landed after rc.2, so the planned rc.3 warning release is a prerequisite to B. The rc.14 pin only authorizes hash-v2 carrier publication.
- **Anomaly:** local launcher main is stale; public v0.1.1 already equals public main and includes the requested features, while source/README still identify v0.1.0. Correct version metadata on a new reviewed commit before a new tag.
- **Decision:** defer B3 unless owner repairs the two reproduced boundary defects, advances the paired consumer pin, resolves current main drift and obtains fresh review. Prior pure decision/mutant coverage is not proof of reference-discovery safety.

This embedded logbook is the task-scoped record; read-only scope leaves the repository-wide LOGBOOK.md untouched.

## Research deliverable checklist

- [x] Evidence-backed required/open checklists for spec rc.14, curator rc.3 and launcher tag.
- [x] All four release-brief questions answered, with ordered dependencies, rough effort, parallel paths and operator publication points.
- [x] Findings written to this file; primary sources and attributed board outcomes cited.
- [x] Important findings/decisions/anomalies recorded in the embedded task logbook.
- [x] Task-scoped findings and machine-readable evidence prepared for board attachment; board activity records the attachment and researcher handoff separately. No release-readiness checkbox above is checked merely because research exists.
