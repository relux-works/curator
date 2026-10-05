# Review verdict: accepted

TASK-261004-34brhn — research-claude-login-transfer-modes
Parent: STORY-261004-1pwxri — claude-macos-managed-home-login-modes
Change request: CR-TASK-261004-34brhn-2, revision 2.
Reviewed 2026-10-05. Base ca1b776fb580ec0cee0173bf150daf063023aeaa; candidate tree eb0495a995fda34ad318f1be532f623144a04f40.

Accept the research draft for producer integration. This is acceptance of a decision proposal and bounded evidence, not authentication qualification or approval to enact its open operator decisions. No blocking findings remain.

## Previous rejection and rework

Read TASK-261004-34brhn_review-verdict-rev1.md. Its sole blocking mechanism R1 was unavailable probe/fixture/sandbox/verifier definitions. Revision 2 fixes it: the evidence companion lines 151–915 supplies the reconstructed wrapper, complete synthetic fixtures and configuration snapshots, eleven matrix assertions, security shim, fake helper, sandbox template, setup/run sequence, pinned bundle identity and repository inputs, expected counts, real dated exit ledger and artifact checker. Historical observations are labelled historical and the unavailable original harness is explicitly distinguished from the reconstructed one. The draft itself is unchanged since revision 1; rework touches only the companion. R1 is resolved, so there is no repeated blocking mechanism and no repeat-of entry to assign.

## Swept surfaces

| Surface | Result | Evidence |
| --- | --- | --- |
| Scope and operator decisions | held | Draft lines 45–114 offers all three requested selectors and capability-gated shared-file, four options with concrete tradeoffs, profile/environment control, defaults and whole-selector precedence. |
| Reproducibility | held | Six exact script/template bodies, eleven cases, fifteen-row historical and reconstructed ledgers, pinned inputs and dated exits. Independent extraction compiled all three Python bodies; artifact checker reran with exit 0. |
| Evidence accuracy and confidence | held | Sixteen citation groups spot-checked below. Draft lines 29–41 and evidence gate ledger distinguish synthetic status from real authentication, inspection from measurement, and repeat reads from refresh. Q1 remains 0/3 disposable-account facts. |
| Safety and public publication | held | Candidate contains synthetic DUMMY values, neutral references and no concrete personal paths, internal hostnames, client/employer names or credential material observed on inspection. Harness whitelists child env, pins executable, redirects HOME/config, uses a non-forwarding shim and denies credential-location access/network. Historical full-home/IPC denial is not attributed to the reconstructed sandbox: its narrower bounds are explicit at evidence lines 623–631. This review read no real credential stores and ran no Claude login/logout/status command. |
| Decision readiness | held | Concrete source registry and selector union; launch/tracked boundary, rotation, expiry, migration, specification amendments, ordered implementation leaves, planned negative tests and six decision-ready questions with recommendations. |
| Q1/Q7 and architecture | held | Native credential copying remains refused; setup-token enrollment requires a narrow normative revision; file sharing remains gated; final-executor lookup occurs after tracked serialization. Backend metadata is not treated as authentication. |
| Diff hygiene | held | Exact candidate adds only the two named .research files; both working files match candidate blobs byte for byte. LOGBOOK.md and product code untouched. |
| Validation | held | Supplied artifact verifier exit 0: 12/12 headings, 3/3 JSON blocks, 15/15 historical and 15/15 rerun rows, two matching research paths, clean whitespace, links and public-pattern checks, 84,970 bytes under 98,304. No Go builds/tests performed. |

## Citation spot checks

Read curator Git objects at ca1b776fb580ec0cee0173bf150daf063023aeaa:

1. internal/config/environments.go:20 — isolation map and no credential-source modes.
2. internal/config/environments.go:799 — shared/isolated parser.
3. internal/envregistry/envregistry.go:240 — Darwin passthrough empty, 2.1.261 pin.
4. internal/envregistry/envregistry.go:450 — Darwin default/refusal policy.
5. internal/envprofile/managed.go:319 — production isolation resolution.
6. internal/envprofile/managed.go:563 and :598 — passthrough selection and unconditional keychain marker.
7. internal/envprofile/managed.go:1504 — credential-link preservation/refusal.
8. internal/envfragment/envfragment.go:54 — fragment contract, ordinary v2/Muse v3.

Read immutable public source at curator-spec rc.14 commit 43bf0a2506d5c354a73bbc3ea4623d4653db10c7:

9. protocol/environments.md:1426 — native-store passthrough and unverified Linux refresh.
10. decisions/0017-environment-credential-modes.md:70 — historical managed file observation.
11. decisions/0017-environment-credential-modes.md:168 — Q1 gate, bounded Q2 isolation, Q7 refusal.
12. profiles/manager.md:2847 — credential mutation boundary.
13. protocol/environments.md:3758 — both lock directions, no credential-selection lock.

Read immutable public launcher source at 2517d2753945d0a8b0c40a291e4aef883a7c16ee:

14. internal/composition/composition.go:35 — full direct env excluded from JSON; env_literals serialized.
15. internal/composition/composition.go:75 — composition constructs serialized literals.
16. internal/execution/execution.go:117 and :142 — production direct/tracked execution boundary.

All sixteen checked groups support the bounded claims. Sources: https://github.com/relux-works/curator/tree/ca1b776fb580ec0cee0173bf150daf063023aeaa ; https://github.com/relux-works/curator-spec/tree/43bf0a2506d5c354a73bbc3ea4623d4653db10c7 ; https://github.com/relux-works/curator-agent-launcher/tree/2517d2753945d0a8b0c40a291e4aef883a7c16ee .

Official authentication documentation was independently opened on the review date and supports separate config-directory logins and the Console-profile exception: https://code.claude.com/docs/en/authentication . Settings-reference rendering failed; no independently verified helper-documentation result is claimed from that failed read.

## Verification bounds and disposition

Independently reran document verification and Python syntax compilation only; inspected supplied fixture/selection assertions. Accepted the attached producer-reported reconstructed R1–R15 ledger as bounded synthetic-selection evidence, without independently executing the CLI or attesting historical safety. Did not rerun shipped-binary offsets, live login, refresh or any product suites; no prior product suites are promoted to current-release passes. These limitations do not qualify shared-file and do not discharge Q1.

Fresh public main advertisement: 9bc8e1a1eace93377e41c36465b26a58ee5ce05a, minimal-config ls-remote exit 0. Read-only changed-path comparison against the CR base returned exit 0: 124 upstream paths, zero overlap with the two candidate research paths. This is a review freshness check, not landing authorization.

Board outcome listing confirms task-scoped draft and updated evidence attachments. Run goal query reported none (not goal-bound). LOGBOOK.md prohibition overrides the generic logbook checklist; this verdict persists review findings on the board. Leave integration to the bound researcher/analyst producer; no commit_ack, commit, checkpoint or done transition is supplied by this reviewer.

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row": "Scope and operator decisions", "result": "held"},
    {"row": "Reproducibility", "result": "held"},
    {"row": "Evidence accuracy and confidence", "result": "held"},
    {"row": "Safety and public publication", "result": "held"},
    {"row": "Decision readiness", "result": "held"},
    {"row": "Q1/Q7 and architecture", "result": "held"},
    {"row": "Diff hygiene", "result": "held"},
    {"row": "Validation", "result": "held"}
  ],
  "free_hunt": []
}
```
