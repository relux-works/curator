# TASK-261021-vb3z6f — independent review r1

Verdict: **changes_requested**. One P1 contract-integration finding; the executable checks pass.

Reviewed head: `0c1f830fc4006299f463f731af2310ec5d9f8eb1`.
Base: `e41f96387287f5e76604627f9e63662f7d6d86d4`.
CIP-0002 comparison: `752307c6a825c07256521e818fef25ba5ae78fc7`.
CIP-0016 amendment comparison: `3a3e041b3fbf76292e69b6abb4a0eca855524007`.

The authorized request and board reviewer assignment agree. Read every task precondition, the author's results-r1, and the two referenced FRONT-DOOR resources through task-board. Fetched using the specification repository's explicit git directory/worktree binding. Detached review head matched the request; the sole introduced commit verified `G`. The worktree had zero commits behind local main. The complete test run used a separate clean normal-`.git` clone at this head, with a short workspace-local TMPDIR and GIT_DIR/GIT_WORK_TREE unset. Final clone status remained clean and HEAD unchanged.

## Surface sweep

Held describes the proposed documentation contract unless executable evidence is named. This change publishes no exact/3 runtime or schema. Walkthroughs are design counterexamples, not invented production execution.

| Surface | Held or broken | Attack run | Evidence |
|---|---|---|---|
| Closure of detached/doorbell gap | Held | Walked a detached `werkstatt up`: qualification and route selection, authorized doorbell, preparation, Curator capture, caller composition/sealing, receiver admission, ready record, later attach and notification retry. Challenged lost response and missing TTY. | Sections 1–4 provide the necessary inputs/result and separate caller ingress from provider PTY/app-server. Only ready is success; notification retry cannot launch again. Runtime activation is explicitly pending. |
| Snapshot binding and ownership | Held | Compared the complete digest boundary and no-host-composition rule with CIP-0014/0016. Ran the repository snapshot checker with a valid bundle, then changed content bytes, added/removed entries, injected descriptor env, changed descriptor digest and removed canonical LF. Compared portable identity against exact manifest identity. | Valid control passed; each mutation failed with its intended `file-identity`, `entry-set`, `launch-schema`, `launch-digest` or `metadata-canonical` reason. Portable equality did not imply equal exact manifest hashes. This checks specification data, not materializer filesystem safety. |
| Runtime qualification and retirement | Held | Walked first-release partial qualification, complete exact/3 activation, stale client, withdrawn capability after retirement, another runtime and grandfathered legacy continuation. Compared the actual pinned CIP-0016 amendment and supplied front-door resources. | Initial joint Claude/Codex/Muse gate retained; subsequent retirement is per tuple, manager-persisted and enforced at receiver and launcher. No fresh delegated/native downgrade; legacy continuation remains explicitly legacy. New refusal spellings have no conflicting definition in inspected protocol/profile/CIP text. |
| Project-context interface | **Broken: F1** | Compared the newly published CIP-0002 candidate's closed lease state machine and exact-route retention clause with CIP-0023's capture-to-session transfer. Used an active detached session plus a concurrent source-home repair/sync as the witness. | CIP-0023 lines 332–338 introduce a capture-only early-release path. Candidate lines 365–375 and 478–486 require the original session-owned home lease through authoritative termination/reconciliation. The assumed interface needs reconciliation now. |
| Resume, recycle and GC | Held except F1 | Walked known-stopped cold resume with original content/composition, refreshed permitted secrets and newly sealed generation. Challenged missing/unreadable receipts, unknown start/liveness, daemon recycle, stopped-but-resumable roots and pending outbox events. | Section 7 forbids live-profile/fresh-session fallback, retains uncertain references and separates session recycle from manager recycle. The specific original-home lease inconsistency is F1, not masked by otherwise adequate snapshot GC roots. |
| Security and free hunt | Held except F1 | Checked unauthorized doorbell destinations, secret-bearing events, post-seal env changes, unapproved project inventory, native discovery, policy drift and admission retry. Read all five changed files and searched added lines for personal paths/host/person identifiers. | Prepared bindings, metadata-only outbox, full process digest, protected receipt/admission checks and adapter qualification remain gates. No introduced hygiene finding or code/schema widening found. |
| Scope, process and completeness | Held | Checked CIP template, Draft/index agreement, separate successor rationale, all owner questions/recommendations, vector families, local links and actual heading anchors. Ran the real link checker with a valid target then an independently planted missing target. | Required sections present; link control specifically rejected `missing.md`. Normative/schema/corpus/tool/release paths unchanged. Number 0023 follows the separately allocated MCP proposal. No implementation qualification is asserted. |

## F1 — retain the original project-home session lease

Location: `cips/CIP-0023-exact-plan-hosted-launches.md:335` (context lines 332–338).

CIP-0023 calls its only project lease input `capture_lease_ref` and permits releasing a **capture-only stable-home lease** after durable transfer when the projection has no later live-home dependency. It does not identify a separate mandatory original-home session lease that survives that release.

The requested CIP-0002 candidate defines a different interface:

- Lines 365–375: preparation creates `prepared`; transfer creates `active` before provider start; `active` becomes `released` only after authoritative confirmed termination. Unknown state retains the home and artifacts.
- Lines 478–486: specifically on the exact-plan route, the original Werkstatt remains leased until authority/state ownership is reconciled on termination.
- Line 245: repair, switching and sync targeting a live/unknown leased home wait or refuse.

Reproduction as a contract trace: start an approved project composition in a detached exact session; finish capture and durable reference transfer; prove a self-contained projection; leave the provider running; release the original home as capture-only and request repair/sync. CIP-0023's early-release branch can classify the source home as idle, while CIP-0002 requires its session lease to remain active and the update to wait/refuse. Retaining snapshot/artifact roots does not itself preserve the home mutation exclusion. The general instruction to retain independently required leases does not define how the sole capture-lease input maps to the candidate's mandatory session lease.

Required fix: reconcile against the pinned published candidate. Preserve the original home lease as a session-owned binding through termination/reconciliation, separately from snapshot lifetime roots. If early release is intended, make it an explicit coordinated amendment/owner decision with a defined capture-only versus session-lease interface; do not present it as the current CIP-0002 contract. Update the candidate reference/assumption and add a negative vector for repair/sync while the exact session is alive or unknown. A terminal/reconciled session is the positive control for release.

This is a specification integration defect, not a claim that the current runtime already releases such a lease.

## Independent validation

Platform: macOS arm64; Python 3.14.7; Go 1.25.5; repository-pinned jsonschema 4.25.1. Local evidence does not claim the hosted multi-platform CI matrix or native-provider qualification.

| Command/check | Exit | Result |
|---|---:|---|
| `make validate` | 0 | Complete Makefile validation chain. |
| `python3 tools/validate.py` | 0 | 74 schemas and 1294 vector files. |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 0 | 879 tests in 851.385s; OK, no skip summary. |
| `go test ./tools/...` | 0 | generate-vectors passed in 16.176s. |
| Targeted `test_managed_home_snapshots.py` | 0 | 16 tests passed. |
| Targeted `test_local_hosted_launches.py` | 0 | 10 tests passed. |
| Independent snapshot/document probe | 0 | Positive/negative snapshot checks, exact versus portable identity, template/links/anchors, missing-link control, hygiene and unchanged normative paths. Also confirmed the conflicting lease clauses. |
| `git diff HEAD^ HEAD --check` | 0 | No whitespace errors. |
| Final status/head check | 0 | Clean normal-`.git` clone at the requested SHA. |

Private local evidence under review-workspace `.temp/vb3z6f-r1/`: `suite-01.log`, `suite-01.exit`, `snapshot-tests-01.log`, `hosted-tests-01.log`, `probe.py`, `probe-01.log`, `walkthrough.md`, `cip02-candidate.md` and `clone-01.log`. The actual repository validator/snapshot/link entry points were exercised. No exact/3 provider, control transport, runtime doorbell or retirement implementation exists in this diff, so native runtime attacks on those proposed surfaces were not executed. No mocked route is presented as runtime proof. The author's historical branch-absence statement was not treated as current: the named candidate was fetched and compared independently.

```json
{"verdict":"changes_requested","head":"0c1f830fc4006299f463f731af2310ec5d9f8eb1","findings":[{"id":"F1","severity":"P1","file_line":"cips/CIP-0023-exact-plan-hosted-launches.md:335","mechanism":"The proposed capture-only stable-home lease can be released after admission/resource transfer, but CIP-0002 candidate 752307c6 requires the original session-owned Werkstatt lease to remain active until termination and authority/state reconciliation. The interface supplies only a capture lease and does not define a separate retained original-home session lease.","reproduction":"Keep an admitted detached exact project session running after successful snapshot capture and resource transfer. Take CIP-0023's capture-only release branch and request repair/sync of the original home. The proposal can make that home idle; CIP-0002 lines 245, 365-375 and 478-486 require it to remain leased and the update to wait/refuse. This is a static contract witness, not an executed runtime defect.","required_fix":"Reconcile with the pinned CIP-0002 candidate and retain the original home session lease through termination/reconciliation, separately from snapshot roots. Alternatively obtain and state a coordinated amendment defining distinct capture/session lease semantics. Update the candidate reference and add active/unknown exact-session versus repair/sync negatives with terminal-release positive control."}]}
```
