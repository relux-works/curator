# Launcher documentation review gaps — 2026-10-13

Task: TASK-261013-2borgq — wave6-launcher-review-findings
Scope: retained documentation delta plus the explicitly authorised release-pin test edit; nine changed repository files, uncommitted. Revision 2.

## Finding coverage

| Accepted finding | Applied launcher documentation | Still open / ownership bound |
|---|---|---|
| Wire register and duplicated status drift | SPEC §10.2 identifies the dispatcher's published draft result fields separately from the full transfer transport/lifecycle. §10.5 replaces consumer-derived spelling/status assertions with an index of existing format/rule IDs and exact owner revision links. It retains owner short labels, makes removed launch tickets a negative reference, and uses Slice 0 in the assigned rollout passages. README and all four figures use the same draft/transport distinction. | Shared naming/glossary/contract registers are outside this module. The shared naming-scope clarification, shared owner-linked index and mechanical cross-repository reference/status gate remain open; §10.5 records them as publication work. No checker or new naming decision is claimed. Pinned disposition text still needs its owning-lane reconciliation; the launcher explicitly avoids deriving its conflicting spellings. |
| Muse subscription decision state | SPEC §10.3 lane table and open-question text, README, and applicable figure notes separate priority decided 2026-10-13, proposed broker channel (b) with fallback (c), and qualification open. The broker §§10.2/18.5 is pinned in the §10.5 index. | Refresh, rotation, lifetime, billing and the fallback protocol remain unknown/unpublished. API-key support is separate. No supported subscription channel or qualified tuple is claimed; Codex's independent open question remains. |
| Complete exec invocation and interpreted inputs | SPEC §10.2 requires executable/entrypoint/arguments/environment and all transitive code/configuration inputs to be qualified; inert inspection remains local, worker-produced-code execution remains R-VERIFY. Protected entrypoints/configuration and pinned/disabled code loading are required. Negative cases cover readable scripts, inline evaluation, imports, project plugins/executable config, startup/environment injection and indirect shared-tool bypass; inert positive controls remain. README and every target figure carry the obligation. | Tool/release-specific profiles, qualification evidence and the verification contract remain unpublished. A tool that cannot establish the boundary refuses locally and goes off-host. No wrapper service, implementation or qualification is introduced. |
| Protected exec-profile authority and immutable body pin | SPEC §10.1 maps admitted Curator profile identity/lock digest, placement and class minimums through the existing protected machine policy beside R-NETBIND. Dispatcher prepare pins identity/body digest in input/1; start/retry/recovery retain the pin and refuse drift/out-of-bound caller references. §10.2 requires the launcher to apply/verify it before child code, preserving existing class/host/qualification checks and adding refusal qualification. README and all target figures reflect the binding. | Policy/profile encoding and identity-to-body digest relation remain publication work. Workplaces default to the narrowest entry; an interpreter-bearing placement needs an explicit entry. No separate service, register, selectable-set policy or grant dimension is added. |
| Exact base archive/manifest/repository binding | SPEC §10.1 derives one mode-aware manifest from immutable repository B. §10.2 verifies every archive entry, including missing/extra entries, before work; each rework round repeats it. Board import independently recomputes the same canonical manifest from repository B, never the composed plan, and checks result base_commit/manifest_digest while keeping admission checks. §10.5 requires content/mode mismatch, missing/extra path, dirty-checkout and stale-rework qualification. README and all target figures agree. | Canonical manifest/archive/intake encoding is unpublished. Dispatcher remains repository-free; import belongs to the board. Consistent base representations do not prove who edited a path or replace exact-C review. |
| Workspace modes and kinds | SPEC §§10.1–10.2 compare path, supported Git kind/mode and content identity, without following links. Mode-only and equal-blob regular↔symlink changes export as modify; deletion carries prior mode. §10.5 requires corresponding vectors and immutable manifest access. README and every target figure reflect these semantics. | Canonical encoding/exporter immutable access and off-host qualification remain open. Existing workspace-result/1 fields are unchanged; no alternate symlink manifest representation, wire field, refusal code, cap or numeric value was invented. |

## Release and figure corrections

- CHANGELOG opens an empty Unreleased and dates 0.3.0 as 2026-10-13.
- README link and release/install references, SPEC §10 preamble and current agents-management dependency references align with 0.5.1-draft / v0.5.62 / v0.3.0. The existing entrypoint reference is described as consumed from the pin, without claiming fresh runtime verification.
- Revision notes use “accepted architecture-review findings (2026-10-13)”.
- components.puml, launch-paths.puml, launch-states.puml, artifact-layout.puml and their linked index show hosted launch unavailable until the swarma dispatcher path ships. session_host_missing (exit 1) occurs before handoff even with task-board on PATH; earlier validation and native/ax routes remain.
- Superseded active hosted receiver/descriptor paths were removed from current panels. Target panels remain expressly unimplemented/unqualified.
- Code follow-up: the session_host_missing detail still says “task-board receiver is not available on PATH”. Correct that diagnostic in a separate code task; this pass did not change it.

## Revision-2 repair and validation

The revision-2 scope decision authorises exactly the release-document pins in
cmd/curator-run/main_test.go, TestReleaseVersionPinned:

- Line 415: install command pinned to @v0.3.0.
- Line 416: released v0.3.0 README wording replaces the candidate/pending-tag sentence.
- Line 418: CHANGELOG heading pinned to the dated released heading.

wantBuild = "0.3.0" (line 400), wantSpec = "0.5.1-draft" (line 401),
both spec-version documentation pins (lines 413–414), and every other byte
of the test file remain unchanged. This is the only code edit. The eight-file
revision-1 documentation delta is retained byte-for-byte. No schema,
implementation, validation configuration or LOGBOOK.md changes were made.

Commands run directly in revision 2, without pipes or tee:

| Command | Real exit | Evidence / limitation |
|---|---|---|
| git diff --check | 0 | No whitespace errors after the test edit. |
| python3 .temp/resources/TASK-261013-2borgq/TASK-261013-2borgq_r2_audit.py | 0 | 9/9 scoped files; 3/3 exact test-pin changes, all other test bytes unchanged; 7/7 added-text hygiene patterns; 20/20 documentation consistency clauses; 101/101 local file targets; 4/4 figure-source structure checks. |

| sh scripts/remote-gate.sh | 0 | Hosted CI: Test and Race succeeded on ubuntu-latest and macos-latest (4/4 required lanes); hosted Test includes gofmt, build, vet and tests. Optional self-hosted lane skipped, not counted as passing. |

Hosted result: https://github.com/relux-works/curator-agent-launcher/actions/runs/38372728944
Snapshot commit: 9fa7acd22cd75e6066295c4a42f7eed94b274579.
The gate created a detached snapshot and cleaned up its temporary branch;
the Story worktree and index remain uncommitted. This is a fresh hosted run
executed by this revision-2 producer, not reused predecessor evidence.
The CR handoff also applies its configured validation to the frozen candidate;
that separate log is board-generated. No CR-validation result is claimed
before that command returns.

The source audit is not implementation qualification or a PlantUML parser.
External HTTP availability, link anchors and rendered layout remain unverified.
Rendering was not rerun; the predecessor recorded java -version exit 1 because
no runtime was available. That failing probe is prior evidence, not a fresh
probe or a passing rendering gate. No local Go tests or builds were run, as
required by the revision-2 instruction; Go validation uses hosted CI.

Public hygiene checks cover additions and the seven stated patterns, not a
universal name/hostname classifier. Existing out-of-scope historical text is
untouched. Optional standing export discard, local-reviewer/requester-exclusion
and metadata-confidentiality policies remain unchosen/off. No implementation,
qualification or activation is claimed for the six documentation findings.

## Prior failure and resolved scope constraint

Revision 1's configured remote gate exited 1 on the obsolete release-document
assertions in TestReleaseVersionPinned. Its successor retained the docs delta
and reported the conflict between the required released heading and the then
strict docs-only scope. The current scope decision resolves that conflict by
allowing the three assertion edits above and requiring hosted validation.
The earlier blocked recommendation is superseded; the failing revision-1 gate
remains recorded as failing and is not reused as green evidence.

The stale session_host_missing detail is still a separate code follow-up;
this revision deliberately leaves that production string unchanged. Every
publication and qualification gap in the finding table remains open.

Findings, decisions and limitations persist in task outcomes and notes.
LOGBOOK.md is expressly excluded by the brief. The duplicate generic logbook
checklist requirement is inapplicable. Reviewer verdict routing is a later
review-role action, not evidence produced by this developer handoff.
