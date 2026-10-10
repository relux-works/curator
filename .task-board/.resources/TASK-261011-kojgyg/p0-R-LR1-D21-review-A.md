# TASK-261011-kojgyg — R-LR1-D21-contract-fix — reviewer A

Record-only review of CR-TASK-261011-kojgyg-1 revision 1. Recommendation: accept the bounded Decision 0021 amendment. No accept_cr, reject_cr or terminal status mutation performed, as required by p0-R-LR1-D21-review-A-brief.md. The deciding reviewer retains the verdict transaction.

Reviewed exact delta: base a60fd34173abc82401b5aee00c6204dce2456997 to candidate tree 1486fc81f4b3225b564e57f7912916890c3ddb27. Only decisions/0021-sessions-enter-through-curator-run.md changes. No repository edits made by this reviewer. The run goal query reports no active goal.

## Findings

```yaml
findings:
  - id: A-N1
    severity: note
    severity_reason: "P3 informational scope boundary; no blocking defect found."
    location: "decisions/0021-sessions-enter-through-curator-run.md:214-245"
    finding: "The amendment authorizes a mode but does not publish or qualify the launcher intake. It explicitly keeps activation disabled until publication and qualification."
    action: "Retain this distinction in consuming launcher work; no revision requested here."
```

## Sources and coverage

Read the contract table from curator commit 177cb554 and audit from curator commit 48789217 using local Git objects, plus the supplied owner-decisions-20261010-audit.md. Read the full candidate document and board outcome TASK-261011-kojgyg_alignment.md. No remote freshness or runtime capability is inferred.

R-LR1 in audit D1 assigns exactly four A-rows, not ten. Checked all 4/4, plus 13/13 supporting contract rows below to exceed the requested ten-row substantive comparison without inventing audit IDs.

| Audit row | Candidate location | Result |
|---|---|---|
| A09.1 | 195-210, 226-231 | Caller composes; agent executor verifies/executes; root launcher distinct; shared launch library and broker-only executor network match D-EXECUTOR. |
| A09.6 | 214-245 | Intake, transport, binding, signals, exits, receipts and refusals assigned to launcher publication; activation gate explicit. No invented schema or runtime claim. |
| A09.10 | 46-48, 135-141, 170-174, 176-210, 343-345 | Versioned agent-account exception consistently replaces blanket headless prohibition; human interactive contract preserved. |
| A10.9 | 230-231, 237-245, 257-268 | D-EXECUTOR chooses curator-run first; board/host call dispatcher with no embedded broker client; slice 0 remains Claude with peer/binding qualification. |

| Contract row | Candidate lines | Checked canonical value or open question |
|---|---|---|
| N6 | 195-203 | curator-run under worker-<label>, distinct from root swarma-user-launcher; human mode separate. |
| P1 | 195-210, 226-230 | Local headless v0, immutable caller-composed plan, dispatcher admission and agent execution; owner amendment supplied by D-EXECUTOR. |
| I2 | 226, 230 | Execution, board run and correlation launch_id remain distinct; correlation is not authorization. |
| I3 | 227, 230 | Active owned generation and broker binding; UID/label alone not authority. No new binding identifier format invented. |
| L1 | 229 | Privileged launcher request 80 KiB within 5 s. Helper's separate 16 KiB bound is not incorrectly applied to launch input. |
| L2 | 228-229 | FD3 plan; 48 KiB decoded, 64 KiB base64, v0 96 KiB/5 s, v1 128 KiB, broker 64 KiB; smallest bound before provisioning. |
| P2 | 227-228, 232 | Atomic root receipt before group/GID/UID drop, existing receipt refuses replay, no inherited broker connection. |
| P6 | 232 | Owned execution/generation and receipt/process checks, TERM/KILL, confirmed absence; no sudo relay or PID-only proof. |
| C3 | 231 | Enrolled setup-token through CLAUDE_CODE_OAUTH_TOKEN at exec; no credential file, native fallback or refresh. |
| C9 | 230-245 | D-EXECUTOR resolves first consumer; cross-user proof and protected Claude launch precede later Codex/network slices. |
| W3 | 249-251, 263-268 | Component-specific signed grammar versus shared adapter stays open; naming does not invent canonical signed bytes. |
| I4 | 252-256 | Joint ticket/host-role versioning versus disabled SH2 stays open; issuer/redeemer and epochs remain distinct. |
| C4 | 257-261 | Single owner with supported-release evidence versus defer remains open; app-server external token target, no auth.json or broker-mode codex exec. |

D-PATH is reproduced in its exact order at line 226, excludes repository .agents/bin and binds resolved PATH to the plan digest. D-WIRE-NAMES is reproduced at 263-268 without inventing signed formats. D-GOALS, D-IDENTITY, D-DONOR-PROVISIONING and D-RESUME add no new behavior to this narrow one-shot decision; their separate target contracts are not silently rewritten. Existing interactive profile-pin resume language is not a claim about harness-version admission.

## Swept surfaces and alignment

| Surface | Result and bound |
|---|---|
| Authority and human-mode preservation | Reviewed full original decision and delta; amendment is dated/versioned, original human composition/ax/untracked semantics preserved. |
| Privilege, identity, credential and transport boundaries | 13 supporting rows checked above; no contradictory value found in text within scope. |
| Activation and owner questions | Qualification required, not claimed; W3/I4/C4 retain their options; D-EXECUTOR is a supplied decision, not an editorial recommendation. |
| Diagrams | Candidate Decision 0021 has zero diagram blocks or image references. Audit R-LR1 D1 names no existing Decision 0021 figure. Related F-L figures have other repository owners; M-LR1/M-LR2 are future P1 work. No diagram rendering or external diagram acceptance claimed. |
| Alignment outcome | Complete and true for 4/4 assigned A-rows; accurately limits changes to Decision 0021 and distinguishes launcher follow-on work. |
| Scope and hygiene | Exactly one Markdown path changed; full-text review found no personal first names, local personal paths or session links. Product paths/identifiers are legitimate contract text. |

## Verification and limits

Directly reran against the candidate tree: document lint (exit 0), followed by exact-base/candidate git diff --check (exit 0). Lint checked 1/1 allowed changed path, 10/10 Markdown table lines (8 data rows), 4/4 relative file links against candidate Git objects, 4/4 audit references, whitespace/NUL and selected personal-path/session-link patterns. These checks establish syntax/packaging only; semantic review is the manual comparison above. Hygiene scanning is bounded, not an exhaustive detector.

The producer outcome reports post-edit document lint and diff check with exit 0; accepted as historical reporting only, not substituted for the direct reviewer checks. No product tests, builds, renderer, provider experiment, commit, deployment or source change ran. Generic tests-green is inapplicable under the explicit docs-only/no-tests brief. No blocking findings or rework requests for this revision.
