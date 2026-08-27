# STORY-260720-21bsr2 decomposition audit

Date: 2026-07-20

## Audit verdict

Retain the existing twelve-task, eight-phase decomposition. No task was
duplicated or replaced. The ownership split is sound: curator-spec owns the
fixture, expected outcomes, result contract, runner, portable author guidance,
and driver-admission guidance; Curator and csk own separate native consumers;
evidence-only tasks guard release and pin transitions.

The audit corrected contract and dependency defects that would otherwise have
left developers to infer behavior or allowed release qualification to start too
early.

## Corrections made

1. TASK-260720-2g7avf, Define shared executable compiled-build interoperability
   cases, now owns a test-only result sink and deterministic JSON Lines record
   contract in addition to shared cases and expected outcomes. The contract
   fixes the case, suite-digest, decision-boundary, mutation, and launch fields
   required by both native consumers and the runner.
2. TASK-260720-1673lr, Consume the shared compiled-build suite in Curator, and
   TASK-260720-31zeo2, Consume the shared compiled-build suite in csk, now
   explicitly implement that suite-owned record contract and exercise shared
   positive lifecycle cases as well as rejection cases.
3. TASK-260720-31zeo2, Consume the shared compiled-build suite in csk, is now
   blocked by TASK-260720-3pemm6, Add cross-platform Go build end-to-end tests,
   matching its stated real-Go E2E prerequisite.
4. TASK-260720-3nj1r6, Run both managers through one black-box interoperability
   gate, now consumes and validates the suite-owned result contract rather than
   defining a runner-local envelope after the consumers already exist.
5. TASK-260720-3pvihp, Qualify manager releases for the specification gate, is
   now blocked by TASK-260720-1pvfj5, Enforce cross-platform compiled-build CI
   gates, and TASK-260720-3s27te, Verify integrated csk schema v6 implementation.
   Published manager releases cannot be qualified before both manager
   integration handoffs and their exact candidate-suite evidence exist.
6. TASK-260720-vs6den, Promote curator-spec implementation pins after manager
   releases, now explicitly precedes manager-side released-suite pin promotion.
   It pins only qualified manager releases in curator-spec and makes no protocol
   release claim.
7. The candidate-versus-release boundary was aligned in TASK-260720-12r55p,
   Consume shared schema v6 vectors; TASK-260720-3pemm6, Add cross-platform Go
   build end-to-end tests; and TASK-260720-1pvfj5, Enforce cross-platform
   compiled-build CI gates. Candidate evidence uses an explicit immutable
   non-default revision or CURATOR_CONFORMANCE_ROOT and records its digest;
   default committed manager suite pins remain on the previous released
   protocol. The stale first checklist line on TASK-260720-1pvfj5 is explicitly
   superseded in its logbook by its current description, scope, AC, and later
   checklist gates.

## Atomic ownership and coverage

| Deliverable | Owner |
|---|---|
| Shared fixture, stable cases, expected outcomes, and consumer result contract | TASK-260720-2g7avf, Define shared executable compiled-build interoperability cases |
| Independent Go consumer | TASK-260720-1673lr, Consume the shared compiled-build suite in Curator |
| Independent Python consumer | TASK-260720-31zeo2, Consume the shared compiled-build suite in csk |
| Strict black-box comparison | TASK-260720-3nj1r6, Run both managers through one black-box interoperability gate |
| Portable schema v6 author guide | TASK-260720-14jjgt, Publish the portable compiled-skill authoring guide |
| Language matrix, unsafe non-goals, and future-driver admission process | TASK-260720-p7sdhg, Define the language-driver matrix and admission gate |
| Immutable manager-release qualification | TASK-260720-3pvihp, Qualify manager releases for the specification gate |
| Qualified manager pins in curator-spec | TASK-260720-vs6den, Promote curator-spec implementation pins after manager releases |
| Immutable protocol-release qualification | TASK-260720-25d05o, Qualify the released schema v6 protocol suite |
| Released-suite pin and no-skip audit in Curator | TASK-260720-38l1sy, Audit the Curator released-suite pin and CI gate |
| Released-suite pin and no-skip audit in csk | TASK-260720-1utsx8, Audit the csk released-suite pin and CI gate |
| Acceptance, parity, documentation, and provenance evidence | TASK-260720-22ynoi, Verify compiled-build interoperability across released implementations |

Searches for each owned artifact path found one owning task. Intentional
implementation-file overlap is serialized: the Curator consumer follows
TASK-260720-jrrgw9, Verify rc.4 build-driver conformance end to end, and the csk
consumer follows TASK-260720-12r55p plus TASK-260720-3pemm6.

## Dependency and release sequence

1. Protocol build and lifecycle vectors unblock the shared executable case set.
2. The two native consumers implement the suite-owned result contract after
   their manager-specific conformance prerequisites.
3. The black-box runner compares disjoint manager states against one suite
   digest and fails closed on missing, extra, duplicate, skipped, malformed,
   crashed, timed-out, wrong-digest, or divergent records.
4. Both manager integration handoffs and the parity runner precede immutable
   manager-release qualification.
5. curator-spec pins only actual qualified manager releases and runs the in-tree
   candidate suite across Linux, macOS, and Windows.
6. The actual signed protocol release and conformance archive are qualified.
7. Only then may Curator and csk advance their committed suite pins to the
   qualified protocol release.
8. Independent verification audits parity, documentation, and pin provenance at
   the qualified refs.

If a required public release, signature, artifact, checksum, digest, or CI
record does not exist, the corresponding evidence task must record the exact
missing external item and enter blocked. A branch, mutable tag, guessed hash,
local-only pass, board status, or planned version is never substitute evidence.

## Validation evidence

- Canonical story plan: 12 tasks, 8 phases. Critical path remains
  TASK-260720-2g7avf -> TASK-260720-1673lr -> TASK-260720-3nj1r6 ->
  TASK-260720-3pvihp -> TASK-260720-vs6den -> TASK-260720-25d05o ->
  TASK-260720-1utsx8 -> TASK-260720-22ynoi.
- Related cross-story plan: 59 tasks over 26 dependency phases; planning
  succeeded with no cycle.
- Structured audit: 12 child tasks, zero empty descriptions/scopes/AC, zero
  task-name duplicates, and zero tasks with fewer than three checklist items.
- `task-board validate` reports only the pre-existing twelve EPIC-260712 prose
  dependency references and unrelated orphan
  `.resources/TASK-260713-7a9c1e/review.md`; no issue names this story, its tasks,
  or its resources.
- PlantUML 1.2026.6 check and SVG rendering passed with Smetana; both SVGs pass
  `xmllint --noout` and both PNGs were visually inspected.
- TASK-260720-2g7avf diagram hashes: source
  `4098f619fc1e6360c6cc7a2471f97031334ddcfed6ba4486805ed5b88b64f3af`, SVG
  `6bfe2e36d1063b66d9a4ac46c8336e8df26b015337ae86b10775447e892f0699`.
- TASK-260720-3pvihp diagram hashes: source
  `40f26a50defa95092b62cc1f54415558a6091c817962715bd437bf828ee0f9de`, SVG
  `de3804ff72a04db3f9f02bad14def20791dabfe5eaa5d32348742adc3f200434`.
- Materialized board resources are byte-identical to the canonical plan and both
  diagram sources/renders.
- Tracked diffs for Curator, curator-spec, and csk workflow pin files are empty;
  no implementation, test, documentation, release metadata, or pin changed in
  this planning run.

Task-board briefly auto-escalated the new cross-story task links into coarse
story blockers. Those redundant story links were removed after verifying the
precise child-task dependencies remained. This preserves implementation order
without preventing this planning-only story from entering solution-architecture
review.
