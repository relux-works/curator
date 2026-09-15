# Review verdict: changes requested

Task: TASK-260910-1xph2y
CR: CR-TASK-260910-1xph2y-1, revision 1
Base/HEAD: d019f0e7179520b5c8dcde321c4fe51e04552f58
Candidate tree: 96b33ebf97e4dfe01a705124219fbec1adf2f5c1
repeat-of: none

## F1 — Preserve the marker attestation and substitution contract (P1)

Location: schemas/draft-sources-v1/install-marker-v5.schema.json:5-128;
protocol/skillfile-sources.md:190-198.

The new prose replaces only source/git/ref_kind/ref/commit with package and
says other marker fields retain their existing meanings. Yet the closed v5
schema omits both attestation and substituted and rejects either field.
The unchanged v4 schema supports both. Core section 10 requires attestation
and substitution to match the effective plan for currentness, and the new
extension preserves registry/audit requirements and legacy entries. Neither
package nor lock_sha256 records the selected registry/status/key or the fact
and identity of a legacy skill development substitution. The new source-audit
object is a separate machine report binding, not a specified replacement for
these marker/currentness fields. No mapping explains how a schema-2 legacy
substitution or attested network package retains those state comparisons.

Independent reproduction at the schema validation entrypoint:
- Load the shipped valid v4 and v5 marker cases.
- Add attestation={registry:"trusted",status:"audited"}: v4 accepts, v5 rejects.
- Add substituted="operator-development": v4 accepts, v5 rejects.
- v5 rejects on additionalProperties, not on bad field content.

Impact: a writer cannot both emit the retained marker evidence and satisfy the
new schema. Dropping it silently leaves status/repair and substitution/audit
handling without the specified persisted comparison state. This is a normative
contract inconsistency, not a claim of a running manager exploit.

Requested rework: preserve the relevant existing fields in marker v5 (with
appropriate local/network applicability), or specify an explicit equivalent
versioned replacement and its exact currentness/refresh/repair rules. Keep
registry evidence non-authorizing and strict substitution refusal intact.
Add positive v5 attested-network and legacy-substitution cases plus negative
malformed/stale/mismatching evidence requirements. Check the complete v4-to-v5
field migration, including external build records, against retained semantics
rather than only testing the current empty-build marker fixture. A new CR
revision and independent review are required.

## Validation rerun by this reviewer

All 91 changed candidate paths match the CR tree byte-for-byte. HEAD remains
the base; HEAD..main is 0. Repository is specification/tooling, so validation
used its platform-neutral Python/Go lane, not an unrelated app build.

| Check | Exit | Result |
| --- | ---: | --- |
| System Python schema import | 1 | jsonschema missing; prerequisite failure recorded |
| Initial system-Python make validate | 2 | stopped at missing jsonschema; not a pass |
| Existing task venv import readiness | 0 | jsonschema and referencing available |
| PATH="$PWD/.temp/source-contract/venv/bin:$PATH" make validate | 0 | 60 schemas, 1047 vectors, 227 unit tests, Go tooling tests pass |
| venv Python extracted documented draft validation command | 0 | 67/67 schema cases, 53/53 negatives, 7/7 wire schemas, 3/3 snapshot vectors |
| venv Python reviewer adversarial.py | 0 | F1 reproduced; 11 independent selector attacks rejected; root selector admitted; mutant killed; 91/91 CR paths matched |
| git diff --check BASE TREE | 0 | no whitespace errors |
| git diff --exit-code BASE TREE -- schemas/v1/*.json conformance/v1 release tools | 0 | frozen schema/vector/release/tooling files unchanged |
| git diff --cached --exit-code | 0 | real index unchanged |

No check result above is accepted solely from producer evidence. The producer
outcome was consulted for its existing task-local venv; no installation was
needed. The original failed environment check remains documented separately.

## Negative-evidence bounds

The actual specification check call site is the documented loop in
conformance/draft-sources-v1/README.md invoking
Draft202012Validator(..., registry=registry).is_valid(instance). Reviewer
attacks use that same validator and registered schemas. The endpoint-refusal
mutant narrows refusal from more than 2 endpoints to more than 3; the shipped
invalid-too-many case rejects with the original schema and is admitted by the
mutant, so the expected-negative assertion kills it. This is not a delete-only
mutant. Independently tried selector acquisition/ref mixes and containment
path forms against the v2 schema.

Manager semantic execution remains 0/33 and unverified: no production manager
call site was changed or tested, and no local v2 installation, physical race,
credential fallback, audit authorization or native containment success is
claimed. The downstream semantic vectors are requirements, not executed tests.
Their explicit limitation is appropriate to this specification-only scope.

## Logbook and handoff

The marker-field loss is the material finding. Other reviewed contracts cover
source/output separation, frozen full-input identity, runtime/build handling,
transport-neutral identity and bounded fail-closed transport policy. Draft
namespace isolation and unsupported-implementation disclaimers are preserved.

Only scratch review scripts/logs under .temp/review-1xph2y and board outcome
resources were written. No normative/schema/example files were repaired during
review. No commit, tag, push, PR, checkpoint, integration, switch, rebase or merge
was performed. The candidate stays uncommitted. Route to to-dev; do not accept
revision 1 and do not request a human commit acknowledgement.

Reproduction script and combined raw logs are attached as task-scoped outcomes.
