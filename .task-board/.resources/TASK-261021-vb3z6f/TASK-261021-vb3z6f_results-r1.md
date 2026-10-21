# TASK-261021-vb3z6f results, revision 1

## Delivery

- Repository: `relux-works/curator-spec`.
- Branch: `exact-plan-final-form`, created from fetched `origin/main`.
- Base: `e41f96387287f5e76604627f9e63662f7d6d86d4`.
- Signed head: `0c1f830fc4006299f463f731af2310ec5d9f8eb1`; signature verdict `G`.
- Pushed branch; remote-head readback equals the signed head.
- No PR opened, default branch pushed or board status changed.

## Summary and proposal choice

Filed **Draft CIP-0023**, a new successor proposal rather than changing the
Accepted status of CIP-0016 or its historical release/qualification contract.
CIP-0001 uses new numbers for successors; CIP-0022 is already allocated to the
separately filed MCP proposal, so this proposal does not reuse it.

The Draft defines the proposed closed `launch-exact/3` payload and its full
CCJ-1 digest boundary, snapshot manifest/descriptor and receipt bindings,
prepared policy/endpoint/doorbell resources, durable session result,
attached/detached ingress, qualification and retirement lifecycle. It separates
caller terminal ingress from the provider's host-owned PTY or Codex app-server.
`werkstatt up` requests detached exact/3 with a prepared authorized doorbell and
receives a durable session record; no host component recomposes the process.

The initial joint Claude/Codex/Muse public exact-plan gate is retained, followed
by per-manager/runtime/platform/build retirement. Durable retirement blocks
fresh delegated launches even after an advertisement is withdrawn; existing
legacy sessions can continue through their recorded legacy route with visible
limits. Proposed refusal codes and no-downgrade selection are explicit.

CIP-0002 admitted project composition is a Curator capture input, with exact
profile/admission/composition/lease bindings and durable transfer of snapshot
references to the session. Resume, session versus manager recycle, indeterminate
starts, notification outbox and GC roots are defined without live-home fallback.
The test plan lists positive controls, negative vectors and narrowing mutants
for the future implementation gates; none is claimed as runtime evidence now.

## Source evidence and limits

- Read the task brief and its only precondition resource.
- Read CIP process/template, governance, CIP-0016's relevant design/test/owner
  sections, local-hosted and managed-home-snapshot normative contracts, and the
  existing CIP-0002 admission/composition/lease design.
- Read both supplied FRONT-DOOR resources from TASK-261021-yqt1qr.
- Examined `origin/cip16-delegated-launches` at
  `3a3e041b3fbf76292e69b6abb4a0eca855524007`. Its accepted owner direction and
  still-proposed lease/exclusive-profile conditions remain distinguished.
- A successful `git ls-remote --heads origin cip02-ready-for-acceptance` returned
  no matching branch. The Draft explicitly states the assumed project interface
  instead and requires reconciliation when the candidate is published.
- No native provider, credential store, private home or runtime was inspected.
  Supplied exact/2 limitations are identified as input facts, not new probes.

## Files changed

- `cips/CIP-0023-exact-plan-hosted-launches.md`: new complete Draft proposal.
- `cips/README.md`: unique CIP-0023 Draft index entry.
- `README.md`: proposal discovery pointer.
- `CHANGELOG.md`: Unreleased documentation entry.
- `UNRESOLVED_QUESTIONS.md`: linked owner-question discovery entry.

No wire schemas, normative protocol/profile files, release artifacts, validation
tools or conformance corpus changed. The proposal process does not require
vector stubs for a filed Draft; future vectors are listed in the CIP.

## Validation, commands and claims

Git commands used the specification checkout's explicit separate git-directory/
work-tree binding. Machine-local paths are intentionally omitted from this
public report. Python validation reused the existing workspace-local pinned
validation environment under `.temp/TASK-261021-yiyrus/venv`; no install was
needed for this task. All temporary checks/logs are under the current task's
`.temp/TASK-261021-vb3z6f/` directory.

| Command | Exit | Claim and limits |
|---|---:|---|
| `git fetch origin` | 0 | Updated remote references before creating the task branch. |
| `git switch -c exact-plan-final-form origin/main` | 0 | Started at the brief's named base, without absorbing the prior task branch. |
| `git ls-remote --heads origin cip02-ready-for-acceptance` | 0 | No matching candidate branch existed at this successful lookup; this is not a failed read treated as absence. |
| `python -B -c 'import jsonschema; print("validator dependency ready")'` | 0 | The existing local validation environment can load its dependency. |
| `PYTHONDONTWRITEBYTECODE=1 python -B tools/validate.py` | 0 | Final documentation tree passes existing schema/vector/release/link validation: 74 schemas and 1294 vector files. Specification consistency only; native/provider qualification remains unrun. |
| `PYTHONDONTWRITEBYTECODE=1 python -B .temp/TASK-261021-vb3z6f/check-draft-01.py` | 0 | Required template metadata/sections retained, Draft matches its unique index entry, new proposal contains no private path/seat names, and actual repository local-link checker passes. Remote URLs/anchors are outside that check. |
| Same check: valid-link fixture plus planted `missing.md` link | 0 | The actual checker accepts the valid target, then specifically raises its broken-local-link diagnostic for the planted violation. Disposable fixture removed. |
| `git diff --cached --check` | 0 | The staged reviewed scope has no whitespace errors. |
| `git diff HEAD^ HEAD --exit-code -- protocol profiles schemas conformance release tools` | 0 | Delivered commit changes no normative/schema/corpus/release/tool paths. |
| `git commit -S -m 'docs: draft exact-plan hosted successor with detached sessions'` | 0 | Created the signed task commit. |
| `git log -1 --format='%G? %H'` | 0 | Verified `G` at the delivered head. |
| `git push -u origin exact-plan-final-form` | 0 | Published only the named feature branch. |
| `git ls-remote origin refs/heads/exact-plan-final-form` plus head equality check | 0 | Remote head equals `0c1f830fc4006299f463f731af2310ec5d9f8eb1`. |
| `git status --short` | 0 | Clean working scope after commit; task scratch remains ignored. |

Final logs: `validate-03.log`, `draft-check-03.log`, `python-readiness-01.log`,
`cip02-branch-01.log` and `remote-head-01.log` in the task scratch directory.
Validation was repeated after final prose edits. Self-review checked the closed
payload/result inventories, separate caller/provider stdin, digest LF rules,
retirement versus grandfathered continuation, joint qualification versus
per-runtime retirement, source-capture lease transfer and all GC roots.

Full Go/Python unit suites and runtime mutants were not run: the change files a
Draft and index entries without modifying runtime behavior, validation tools or
schemas. Existing specification validation and the targeted documentation and
checker-negative controls are the relevant checks for this delta. No validation
clone, daemon, background service, provider process or destructive test was used.

## Open owner decisions

1. Retain the initial joint exact-plan cohort, then durable per-runtime
   retirement; explicitly decide if independent initial public rollout is wanted.
2. Joint launcher/manager containing control-version allocation and wire
   publication, preserving all older closed payloads.
3. Prepared protected doorbell routes, at-least-once delivery with dedup,
   retry/cancellation horizon and outbox retention policy.
4. Reconcile the assumed CIP-0002 interface with its candidate; publish Curator
   composition-capture/verification and explicit lease-transfer carriers.
5. Settle CIP-0014 runtime projection/shared-login gates and prove Codex native
   cold-resume policy; unsupported mechanisms refuse rather than rewriting config.
6. Grandfathered legacy continuation, explicit fresh exact migration and
   operator-governed retention/forget of resumable snapshots.

Recommendations and tradeoffs remain in the Draft. Filing it activates no
implementation, runtime capability, native adapter or published wire schema.
