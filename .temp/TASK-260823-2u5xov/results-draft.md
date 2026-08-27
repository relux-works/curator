# TASK-260823-2u5xov — Schema-8 consumption coverage and curator qualification

Branch `task/TASK-260823-2u5xov-schema8-consumption`, PR
[#37](https://github.com/relux-works/curator/pull/37). Open, not merged.

## 1. The defect this closes

`TASK-260823-omp8zt` ran the exact Go and Python commands of
`curator-spec/.github/workflows/implementations.yml` against a staged schema-8
suite and got exit 0 from both. That green is false: the Go pin's conformance
case switched only on the two schema-6 family names, and the Python pin never
reads `schema-cases/index.json`. A suite can gain a whole normative bump while
every pinned job stays green consuming none of it.

curator's own gate had the same shape of hole in two places.

**Presence was incompletely declared.** `.github/ci/root-artifacts.tsv` named
`schema-cases/agent-skill-v8` but not `csk-skill-v8`, so the legacy-alias half
of the bump could disappear from a candidate root without moving the suite
plan. `vectors/script-host-execution-policy.json` was named by nobody and read
by no Go test at all.

**Consumption was not declared anywhere.** `.github/ci/platform-cases.tsv`
carried no row for any schema-8 consumer. `TestReleasedSchemaCases`,
`TestReadAuthoritativeMarkerV4SchemaCases`, `TestModuleRootVectors` and
`TestModuleRootVectorsDriveTheWholeBuild` could each have been renamed, deleted
or filtered out and every lane would have stayed green.

## 2. What changed

| File | Change |
| --- | --- |
| `.github/ci/root-artifacts.tsv` | `schema-cases/csk-skill-v8` added to `internal/skillspec`; new `internal/scriptpolicy` row for `vectors/script-host-execution-policy.json` |
| `.github/ci/platform-cases.tsv` | eight rows naming the schema-8 consumption cases, required on all three runners |
| `internal/scriptpolicy/conformance_test.go` | new; first consumer of the script-worker behavioural family |
| `ci.yml`, `candidate-suite.sh`, `suite-plan.sh`, `root-artifacts.tsv`, `README.md` | candidate-lane prose generalised from "schema v6"; no check changed |
| `README.md` | new "Suite consumption, not suite presence" subsection |

Each ledger row tolerates exactly the skip class its case prints when it ran
without a root: `root-unset` for the four packages the artefact table defers,
`root-content` for the `internal/godriver` case that guards its own read.
`root-unset` is `deferred-only` in `skip-classes.tsv`, so in the candidate lane
no tolerated skip survives.

### What `internal/scriptpolicy/conformance_test.go` asserts

curator parses `script-worker-v1` and refuses it (manager profile §3.6), so what
it can assert is bounded. The file states the bound instead of decoding the
sections it likes and ignoring the rest:

- the closed policy identity and interpreter set this build hard-codes are
  checked against the suite's own bytes;
- every published `opt_in_case` decides both manifest acceptance and the
  enforced/declared-only classification, and an accepted enforced command must
  be refused with `script_execution_policy_unsupported`;
- every enforced shape the suite declares — the policy against each published
  interpreter — is refused before any worker surface, which is what makes the
  worker-side sections genuinely unreachable rather than merely unread;
- the file's top-level section set is asserted in **both** directions against a
  classification table, so a section the protocol adds fails on its first run.

## 3. Two things deliberately not covered, both declared

**`audit_label_cases` — a real gap, named with its owner.** The manager profile
§7 and core §4.1.1 require two audit warning classes,
`script-command-declared-only` and `script-command-unfiltered-declared-network`.
curator emits neither. `script-command-declared-only` is **not**
worker-dependent: it applies to every declared-only script command curator
already installs today. It is classified by name in
`scriptHostExecutionPolicySections` with owner `STORY-260822-2h0v9j`, so it
cannot read as covered. It was not implemented here because it changes audit
decision semantics — always `warn` in every mode, never subject to `fail_on`,
never reportable as an applied control — which is a surface change with its own
conformance needs, not something to fold into a coverage commit.

**`conformance-claim-v5` — not curator's surface.** It is the fourth
schema-8-era family. curator publishes no conformance claim and has never
consumed `conformance-claim-v1..v4`. Declaring a consumer for it would be a
fake gate, so it is deliberately absent from `root-artifacts.tsv`. The claim
schema is the specification's release-gate surface.

## 4. Consumption proof

### 4.1 Removing a family fails the candidate lane

`suite-plan.sh` under `CI_REQUIRE_FULL_ROOT=1` against the `6001dc3` root with
one artefact deleted, five separate runs:

| Removed artefact | Package deferred | `suite-plan.sh` exit |
| --- | --- | ---: |
| `schema-cases/agent-skill-v8` | `internal/skillspec` | 1 |
| `schema-cases/csk-skill-v8` | `internal/skillspec` | 1 |
| `schema-cases/install-marker-v4` | `internal/marker` | 1 |
| `vectors/module-roots.json` | `internal/moduleroots` | 1 |
| `vectors/script-host-execution-policy.json` | `internal/scriptpolicy` | 1 |

Full output: `TASK-260823-2u5xov_family-removal-suite-plan.txt`.
The intact root plans `served=43 deferred=0 excluded=0`, exit 0.

### 4.2 A vanished consumption case fails the ledger

Running `internal/scriptpolicy` with a `-run` filter that no longer matches
`TestScriptExecutionOptInCases` — the shape a rename or a deletion produces —
makes the platform-case gate exit 1 by name:

```
FAIL  required case never ran on darwin: internal/scriptpolicy :: TestScriptExecutionOptInCases
      a rename, a deleted test or a -run filter matching nothing all look like this.
```

Full output: `TASK-260823-2u5xov_vanished-case-platform-cases.txt`.

### 4.3 Both lanes stay correct

- Against the materialised committed `SPEC_PIN` `00b1688a` (which publishes
  none of the five artefacts), the four packages defer and every new row is
  tolerated by its declared class: gate ok.
  `TASK-260823-2u5xov_default-lane-platform-cases.txt`.
- Against the `6001dc3` root, all eight rows are observed passing on darwin:
  `TASK-260823-2u5xov_candidate-lane-platform-cases.txt`.

## 5. Local gate results

| Command | Exit |
| --- | ---: |
| `go build ./...` | 0 |
| `golangci-lint run ./...` — 0 issues | 0 |
| `bash .github/ci/no-broad-suppression.sh` | 0 |
| `bash .github/ci/gate-selftest.sh` — 81 passed, 0 failed | 0 |
| `bash .github/ci/ledger-consistency.sh` — 80 rows across linux/darwin/windows | 0 |
| `go test ./internal/{skillspec,marker,moduleroots,scriptpolicy,godriver}` against the `6001dc3` root | 0 |
| the same packages with the root unset (default-lane shape) | 0 |

