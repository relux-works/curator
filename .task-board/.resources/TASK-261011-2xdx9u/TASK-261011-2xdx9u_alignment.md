# TASK-261011-2xdx9u — R-CS2-contract-fix alignment

## Sources and scope

Canonical source: [platform contract table at 177cb554](https://github.com/relux-works/curator/blob/177cb554/.research/261010_platform-contract-table.md). Audit source: [platform docs audit at 48789217, D1 R-CS2 and C4](https://github.com/relux-works/curator/blob/48789217/.research/261010_platform-docs-audit.md). Later authority: the supplied task precondition `owner-decisions-20261010-audit.md`, D-GOALS/D-IDENTITY/D-EXECUTOR/D-PATH/D-DONOR-PROVISIONING/D-WIRE-NAMES/D-RESUME.

The shared disposition in `cips/platform-contract-disposition.md` records decided values, open owner questions/options, and one cross-CIP schema/owner/version table. No schema, runtime, provider qualification or previously measured test is presented as adopted runtime support. All changes are documentation; the worktree remains uncommitted for role handoff.

## R-CS2 audit alignment

D1 names seven A-rows plus C4. C4 covers eight CIPs in this task and separately discusses C08/C09, owned by R-CS1; those two are not claimed repaired here. Coverage below is 7/7 named A-rows and 8/8 in-scope C4 CIP rows, manually aligned to the cited authority. That denominator measures documentary disposition, not runtime behavior.

| Audit row | Section changed | Canonical value or explicit open owner question |
|---|---|---|
| A09.2 | CIP-0002 Operative disposition; shared Schema owner and version table | Draft operator model only: registration/project/composite commands and closed grammar remain proposed. Accepted direction does not make CLI syntax supported. Contract V1. |
| A09.3 | CIP-0002 Operative disposition and Open questions (existing admission/precedence retained); shared Managed and interactive PATH | Operator conditions retain unique names, alternatives/stacks with jointly satisfying versions, protected per-item approval and granted ceilings, drift checks and adapter refusal. Home/delivery/grammar questions retain their alternatives; no adapter coverage invented. Contract V1/X1. |
| A09.4 | CIP-0007 Operative disposition, Recommendation defaults, Signing keys / Built-in registry distribution questions; shared Open owner questions | Several signed registries; any verified revocation vetoes; dedicated rotating key sets; system/fleet config and separate release-shipped trust bundle, no compiled-in keys. Default entry is configuration. Concrete publisher/service and bundle/snapshot delivery remain open: reviewed files versus service, pinned shipped snapshot versus signed first-use fetch. Contract X2. |
| A09.5 | CIP-0002 Fragment v4 and dispatcher / PATH vector / question 6; CIP-0004 disposition; CIP-0007 launch example; shared Managed and interactive PATH | D-PATH: protected launch tool directory → admitted base PATH in order → one protected manager dispatcher, with checks. Never execute repository .agents/bin in managed launches; resolved PATH covered by plan digest. Interactive append is a separate scope. Exact coordinated transform remains unpublished. Overrides contract X1; retains X2. |
| A10.3 | CIP-0003 disposition/summary/historical design; CIP-0010 C1/C6; shared Credential source/channel | First node/broker backend: 0600 files in owner-only 0700 directories on macOS/Linux; Keychain/Secret Service later. No source/channel fallback. Native per-home evidence remains historical. Contract C8. |
| A10.5 | CIP-0003 disposition/summary and superseded default/Q7/shared-file questions; CIP-0010 C6; shared source/channel table | Source-policy v1 missing entry → per-home; v2 missing entry refuses, migration writes explicit per-home. CIP-0003 adopted narrowed; token becomes node, helper API-key, shared-file replaced, no-copy/migration/no-fallback retained. Provider phases separate; C4 Codex adoption and C7 Muse subscription explicitly open. Contract C1–C8/V1. |
| A10.9 | CIP-0010 Affects / Implementation plan; CIP-0011 Affects / Summary / §4 / Implementation plan / question 2; shared Executor phases (reference to trunk Decision 0021 amendment v1) | D-EXECUTOR: curator-run under agent account first, verifying/executing caller-composed plan only, no profile resolution/compose and no network except broker. Shared library with session runner. Board/session host call dispatcher; no embedded first client. Intake release/grammar pending R-LR1. Overrides contract N6/P1/C9. |
| C4 / C02 | CIP-0002 Operative disposition and shared precedence/PATH/version sections | Existing per-surface precedence retained; accepted operator inputs versus Draft grammar explicit. Protected dispatcher reconciled with CIP7. A new launch diagram is deferred to R-CS3. |
| C4 / C03 | CIP-0003 Operative disposition, Summary, Historical design/specification/implementation, Superseded questions; CIP-0010 C6 | No-acceptance claim dated as history; default and Keychain-first leaf superseded by C10 C1/C6/Revision 2.2. Research preserved, not treated as current authentication evidence. |
| C4 / C04 | CIP-0004 Operative disposition, Summary, accepted specification disposition, questions 1–4; shared PATH section | Accepted no-source hook/interactive order; approval/default questions superseded. Entry/switch/leave restoration and cache refresh/reload sequence explicit; release/cache grammar and compatibility timing still pending. |
| C4 / C05 | CIP-0005 Operative disposition, Summary, dated current state, question 2 | CLI secret transport first accepted; token precedence and explicit-file-failure/no-fallback retained. Remaining backend policy/version/qualification questions remain open. Existing backend/channel/coverage tables remain planned contracts; new figure deferred R-CS3. |
| C4 / C06 | CIP-0006 Operative disposition, Summary, dated current state, question 1 | Narrow machine-owned profile-settings authority accepted; general Decision 0014/packages deferred. Presence/scope/key ownership preserved; detailed rule/suppression/deny/release questions and coordinated versions remain pending. New transaction figure deferred R-CS3. |
| C4 / C07 | CIP-0007 Operative disposition, launch example, Recommendation/defaults, questions 2/3/8, publication leaf; shared registry/version sections | Accepted B conditions govern revocation-before-selection and locked alias exposure. Registry entry and separate bundle distinguished; no invented URL/keys. D-PATH replaces repository bin. New selection/publication figure deferred R-CS3. |
| C4 / C10 | CIP-0010 Operative disposition, C6, consumer plan and open Codex/Muse questions; shared source/channel and phases tables | File-first/default/narrowing decisions operative, first consumer aligned, source/channel matrix retains phases and reauthentication/qualification limits. New lifecycle figure deferred R-CS3. |
| C4 / C11 | CIP-0011 Operative disposition, Summary, §§1/4, consumer plan/questions; shared Executor phases / Wire naming / owner-version table | Separate dispatcher module; Curator commands are clients. D-EXECUTOR first consumer; D-WIRE-NAMES repository-prefixed platform identifiers. Exact module domains, transport versus signed-envelope roles and pending adapters distinguished. |

## Additional owner propagation and unpublished boundaries

| Contract / decision | Section changed | Disposition |
|---|---|---|
| W2 / D-WIRE-NAMES | Shared Wire naming and owner/version table; CIP-0011 §1 pointer | Broker domains now swarma-credential-broker grant/1 and revocation/1; newline + CCJ-1 and grant-ID preimage retained. No migration because nothing shipped. Historical curator-broker domains labelled replaced. |
| W3 / I4 | Shared schema table and open questions | Repository naming also applies to helper/launcher request kinds and eventual launch ticket. No helper v2 signature object/version/canonical bytes or ticket signature/host-role schema invented. Explicit component format versus shared adapter; joint activation versus disabled SH2 backend. |
| D-GOALS / G1/G2 | Shared related boundaries | Board/goals optional; caller/board authors now, no agent host-socket goal capability; interim rationale and deferred separate self-goal namespace decision recorded. |
| D-IDENTITY / R1/R2 | Shared related boundaries | One-shot flag at registration; dispatcher revokes only those identities after account retirement in a separate operation; standing identity only operator-signed revocation. |
| D-DONOR-PROVISIONING / D5 | Shared related boundaries | Separate donor privileged binary/spec/rules/removal; B disabled until specified. Donor drawings remain R-CS1 scope. |
| D-RESUME / P8 | Shared related boundaries | Same major, or same 0.x minor, native resume allowed on same machine without extra qualification row; breaking version refuses; harness failure → resume-failed, caller decides, no automatic fallback/portable seed. |

## Existing diagram disposition

R-CS2 D1 names no existing curator-spec diagram. The five sources under cips/diagrams (donor-side.dsl, donor-first-deployment, join-handshake, remote-worker-turn, remote-worker-retire) were inspected for scope; they describe donor flows owned by R-CS1. They contain neither the combined managed PATH nor the first local broker-consumer path repaired here. No existing R-CS2 figure was found or changed; no new figures were authored. New C4 illustrations are R-CS3/P1 work, as required by the brief. This does not claim donor figures are current or rendered.

## Verification and limitations

The prior content-recovery run manually re-read the seven named A-rows, the eight in-scope C4 entries, and the corresponding canonical contract rows and owner decisions. The candidate remains documentation only and uncommitted. No runtime qualification, provider experiment, diagram rendering, tests or builds were invoked directly.

The alignment table is a board outcome only; its predecessor worktree copy was removed before handoff. The shared owner/version table remains repository documentation. Decision 0021 belongs to TASK-261011-kojgyg and is unchanged in revision 3; its amendment is supplied by trunk ddafa1ff. The shared CIP disposition links to that existing amendment.

## Historical validation and resolved handoff constraint

The predecessor Change Request validation returned exit 1 because the validator treated two owner-precondition `board-resource://` citations as missing local files. Those links were replaced with plain references to the supplied owner-decision precondition. The successor's direct `python3 tools/validate.py` returned exit 0 (73 schemas, 1294 vectors), and `git diff --check` returned exit 0. These are prior results, not freshly rerun evidence. The predecessor packaging checker was also recorded as exit 0; that historical count is not adopted here.

The earlier recovery stopped before handoff because the then-current brief prohibited the wrapper's test/build steps. The current developer brief explicitly sanctions the configured automatic Change Request validation suite at handoff, with its Go step compile-only under R223. That resolves the recorded policy boundary. The suite's actual status and per-command exit codes are recorded by handoff in its Change Request validation outcome; no green result is presumed in this alignment artifact.

## Checklist applicability

Generic code, runtime-test, lint-suite and build checklist items are inapplicable to the explicit docs-only brief and were removed through board mutations. The conditional reviewer-verdict item belongs to later review routing. The retained implementation/architecture items mean documentary acceptance-criteria alignment and consistency with the recorded architecture. No test/build item is checked by inference. Documentation packaging and tracked whitespace are checked only after fresh exit-0 processes.

## Prior standalone documentation checks (not rerun counts)

| Command | Real exit code | Scope and result |
|---|---:|---|
| Initial temporary packaging checker | 1 | Checker defect: an unanchored personal-path pattern matched the conceptual integrity/home/profile notation, and a URI pattern rejected a historical citation explanation in board evidence. Patterns were narrowed to actual personal paths and Markdown board links. No gate pass claimed. |
| `python3 TASK-261011-2xdx9u_doc_packaging.py` (corrected temporary checker) | 0 | 12/12 changed/new Markdown documents including this board outcome, eight operative blocks, 27 local links/anchors and seven new tables; zero findings. No product tests or build. |
| `git diff --check` | 0 | Tracked documentation whitespace. |

## Revision 3 recovery

The current r4 brief supersedes the earlier recovery instructions. Standard `refresh-candidate` returned `refresh_already_current` (exit 0), with trunk, reviewed trunk and branch all ddafa1ff3ff82c2869cce8bc224654f37a4446bb. Decision 0021 has no candidate diff; its existing amendment is referenced, not edited. All eight CIP edits and the shared disposition are retained. The shared disposition link uses trunk's amendment v1 anchor. The untracked repository LOGBOOK is removed; historical notes are preserved in TASK-261011-2xdx9u_rev3-handoff.md on the board. The audit-alignment rows above are carried from the prior outcome; this recovery does not claim a fresh substantive audit or new runtime evidence. Fresh documentation check results are recorded below. No tests or builds are invoked directly. The configured automatic handoff suite remains sanctioned by the current brief.

| Recovery command | Real exit code | Result |
|---|---:|---|
| `python3 /tmp/TASK-261011-2xdx9u_doc_packaging.py` | 0 | 10/10 Markdown documents, 22 local links/anchors, eight new tables; zero findings. |
| `git diff --check` | 0 | Tracked whitespace clean. |
| `git diff --exit-code -- decisions/0021-sessions-enter-through-curator-run.md` | 0 | Decision 0021 excluded from candidate changes. |

The docs-only recovery retires re-injected generic code/test/build/lint obligations and the later reviewer-routing obligation through the board CLI, retaining the six applicable documentary/evidence checklist items. Automatic handoff validation reports its own exit codes; no result is presumed here.
