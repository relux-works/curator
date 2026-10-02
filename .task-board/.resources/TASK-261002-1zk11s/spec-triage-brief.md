# THE ONLY CURRENT INSTRUCTION — TASK-261002-1zk11s: curator-spec owner-review triage (researcher, read-only)

Ivan has DELEGATED the triage of these curator-spec issues to curator:
- #101: accept Go families newer than the newest tested one, with a warning, under the go-v1 lockdown;
- #106: a manager-provisioned Go toolchain;
- #107: resolve version-manager Go shims to the real GOROOT;
- #108: manager-provisioned CLI tools with layered signed registries;
- #109: detect build-repository pins made unreachable by a history rewrite;
- #110: clarify skillfile-sources §2, managed output scoped by physical identity;
- #111: editorial cross-reference of the manager-profile exceptions;
- #112: tag resolution records the terminal commit of the tag chain.

Bring to Ivan only the issues that truly need HIM: product direction, scope, security posture, or anything irreversible. **#106 and #108 stay unimplemented until Ivan decides**; summarise them as decision cards.

For each issue, read the full issue and its comments (`gh issue view N --repo relux-works/curator-spec --comments`). Check the current spec text (curator-spec main e41c561b) and the curator implementation status (curator main). Then give:
- **Kind:** editorial, clarification, normative change, or new feature.
- **Current state:** already covered, partly covered, or absent, with file:line evidence.
- **Recommendation:**
  - ACCEPT and implement: name the leaf scope and the repos touched (spec / curator / csk lockstep?), and whether it goes into rc.14 or later;
  - ACCEPT and defer;
  - REJECT, with the reason;
  - NEEDS-IVAN: one paragraph with the options and your recommendation.
- **Rough effort,** plus conformance and lockstep impact. Who else consumes it? Note ivan-cocoa (csk) for go-v1 and toolchain items: the toolchain-preflight question is open with them.

Output:
- the triage table;
- a short "for Ivan" list, ideally at most three cards;
- the leaf list for the ACCEPT items, in a sensible order.

Read-only: no issue comments, labels or posts. Never spell any employer name.

Write the results. Then run `task-board handoff TASK-261002-1zk11s --role researcher` and END TURN.
