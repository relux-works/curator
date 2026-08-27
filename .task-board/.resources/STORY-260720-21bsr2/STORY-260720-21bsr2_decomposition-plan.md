# STORY-260720-21bsr2 development decomposition

## Outcome

The story is decomposed into twelve atomic tasks over eight internal phases.
The protocol suite remains the only interoperability oracle, Curator and csk
receive separate native consumers, and release pin movement is separated from
candidate testing by two immutable-evidence gates.

## External prerequisites

- STORY-260720-35dck7, Protocol schema v6, owns the normative schema, base Go
  fixture, build-driver and lifecycle vectors, protocol authoring/CLI text, and
  candidate verification.
- STORY-260720-3plyvy, Curator Go build driver, owns Curator product behavior.
- STORY-260720-1uv5gi, csk Go build driver, owns csk product behavior.

Implementation prerequisites are enforced at task level. Links to
TASK-260720-cw39jh, TASK-260720-3lo9jc, and TASK-260720-3ag6pi serialize edits
to shared curator-spec artifacts and prevent duplicate ownership. The Curator
consumer extends TASK-260720-jrrgw9, the csk consumer extends
TASK-260720-12r55p, and the two post-release audits consume the CI handoffs from
TASK-260720-1pvfj5 and TASK-260720-3s27te. Redundant story-to-story blockers are
not retained because task-board treats prerequisites at `to-review` as
unfinished implementation blockers and would prevent this planning-only story
from reaching its required solution-architecture review handoff.

## Work breakdown

| Task | Repository | Atomic deliverable |
|---|---|---|
| TASK-260720-2g7avf, Define shared executable compiled-build interoperability cases | curator-spec | Generated case contract and deterministic executable fixture behavior |
| TASK-260720-1673lr, Consume the shared compiled-build suite in Curator | Curator | Independent Go suite consumer |
| TASK-260720-31zeo2, Consume the shared compiled-build suite in csk | csk | Independent Python suite consumer |
| TASK-260720-3nj1r6, Run both managers through one black-box interoperability gate | curator-spec | Isolated two-manager parity runner |
| TASK-260720-14jjgt, Publish the portable compiled-skill authoring guide | curator-spec | Complete practical schema v6 author guide |
| TASK-260720-p7sdhg, Define the language-driver matrix and admission gate | curator-spec | Ecosystem classification, unsafe non-goals, and future-driver checklist |
| TASK-260720-3pvihp, Qualify manager releases for the specification gate | Evidence only | Immutable manager-release and candidate-suite qualification |
| TASK-260720-vs6den, Promote curator-spec implementation pins after manager releases | curator-spec | Released manager pins and cross-platform implementation gate |
| TASK-260720-25d05o, Qualify the released schema v6 protocol suite | Evidence only | Signed protocol-release and suite qualification |
| TASK-260720-38l1sy, Audit the Curator released-suite pin and CI gate | Curator | Release-safe pin provenance and required Go gate audit |
| TASK-260720-1utsx8, Audit the csk released-suite pin and CI gate | csk | Release-safe pin provenance and required Python gate audit |
| TASK-260720-22ynoi, Verify compiled-build interoperability across released implementations | All three, read-only | Acceptance, coverage, parity, and provenance report |

## Internal phases

1. Shared executable cases and language-driver matrix.
2. Curator consumer, csk consumer, and author guide in parallel.
3. Black-box parity runner.
4. Immutable manager-release qualification.
5. curator-spec implementation-pin promotion and cross-platform gate.
6. Immutable protocol-release qualification.
7. Curator and csk released-suite pin and CI audits in parallel.
8. Independent interoperability verification.

The critical path is shared cases to one manager consumer to the parity runner,
then both evidence gates and pin promotions, then final verification.

## Acceptance coverage

| Story requirement | Owning tasks |
|---|---|
| Same fixture and expected outcomes | TASK-260720-2g7avf, TASK-260720-1673lr, TASK-260720-31zeo2 |
| Same negative-case rejection and built-command launch behavior | TASK-260720-3nj1r6, TASK-260720-22ynoi |
| Complete schema v6 example and Go prerequisites | TASK-260720-14jjgt |
| Cache behavior, dry-run, security limits, and portability | TASK-260720-14jjgt, TASK-260720-22ynoi |
| Language matrix and unsafe build-system non-goals | TASK-260720-p7sdhg |
| Standard process for future compile-only drivers | TASK-260720-p7sdhg |
| No release pin before real evidence | TASK-260720-3pvihp, TASK-260720-25d05o, TASK-260720-38l1sy, TASK-260720-1utsx8, TASK-260720-vs6den |
| Cross-platform independent-consumer proof | TASK-260720-vs6den, TASK-260720-22ynoi |

## Architectural decisions

- curator-spec is the only source of fixture bytes, stable case IDs, expected
  outcomes, and suite identity.
- Consumers are native adapters in separate repositories and may not import
  one another or commit local expected-value copies.
- The runner compares normalized decisions and launch behavior, not
  platform-dependent compiled binary bytes or physical cache paths.
- Candidate testing may precede protocol publication, but release pins may move
  only after the corresponding immutable evidence task succeeds.
- Cross-story notes explicitly require candidate conformance to use a
  caller-supplied CURATOR_CONFORMANCE_ROOT while committed manager pins remain
  on the previous released suite.
- The release sequence deliberately breaks the circular dependency: released
  managers qualify first against one committed candidate suite, curator-spec
  pins those releases and publishes the protocol, then manager CI pins the
  released protocol.

## Gaps and stop conditions

No unresolved product or architecture decision remains because the accepted
TASK-260720-poa3ze research fixes the go-v1 security and portability boundary.
External releases are intentionally not assumed: TASK-260720-3pvihp and
TASK-260720-25d05o must transition to blocked with the exact missing evidence
if the required releases, artifacts, signatures, digests, or CI records do not
exist when execution reaches them.

## Non-goals

- Generic shell hooks, package-selected executable or argv, compiler plugins,
  package build recipes, automatic compiler installation, remote build
  services, cross-compilation, and shared physical caches.
- Treating implementation behavior as normative, copying expected values, or
  comparing toolchain-dependent binary bytes.
- Advancing current rc.3 implementation or suite pins during decomposition.
