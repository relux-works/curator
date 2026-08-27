# TASK-260822-1so0ym review verdict — ACCEPTED

Reviewer run: RUN-260823-82df18 (not goal-bound). Read-only review; no code modified.
All evidence below was reproduced independently by the reviewer, not taken from the
implementer's report.

## Delivery identity (verified)

| Item | Value | How verified |
| --- | --- | --- |
| Branch | `spec/module-roots-prose` | `git ls-remote origin` → `bac193cadb7d26aabf006c92924b4a05f6574e31` |
| Commit | `bac193c` "Add module root conformance vectors" | `git show --stat`; author Ivan Oparin, no AI attribution |
| Diff shape | 5 files, +588/-0 | vectors, manifest entry, generator, regression test, README |
| Ancestry in candidate | `git merge-base --is-ancestor bac193c 6001dc3` exit 0 | confirmed ancestor |
| Vector digest | `b13dcd6177bef7fcb4c96b4dc29a8435883d22624855ddfbfd2de6e760596dac` | `shasum -a 256` on committed bytes equals the `conformance/v1/manifest.json` entry |
| Delta into candidate | only `protocol_version` `1.0.0-rc.8` → `1.0.0-rc.9` | `git diff bac193c 6001dc3 -- conformance/v1/vectors/module-roots.json` = 1 line |

## AC 1 — vector coverage

`conformance/v1/vectors/module-roots.json` carries ten cases. Each was checked line
by line against the normative prose in `protocol/core.md` §4.2.3 and the diagnostic
table in `profiles/manager.md` at the same commit.

| Required category | Case | Diagnostic | `fails_before` | Prose basis |
| --- | --- | --- | --- | --- |
| acceptance | `valid-declared-module-roots` | — | — | bijection satisfied; 4 annotations = 2 directive + 2 selection pairs |
| escape path | `replacement-target-escapes-snapshot` | `..._directive_undeclared` | `go-build` | bijection: replacement naming no declared directory |
| module-to-module redirect | `module-to-module-redirect` | `..._directive_form_unsupported` | `go-build` | admitted directive form: two-token right side |
| undeclared replace | `undeclared-directory-replacement` | `..._directive_undeclared` | `go-build` | bijection |
| unused declaration | `declared-module-without-replacement` | `..._declaration_unused` | `go-build` | bijection |
| nested modules | `nested-declared-module-roots` | `..._containment_invalid` | `go-list` | containment: pairwise disjoint |
| build-root overlap | `module-root-contained-by-build-root` | `..._containment_invalid` | `go-list` | containment |
| runtime-root overlap | `module-root-contained-by-runtime-root` | `..._containment_invalid` | `go-list` | containment |
| versioned-left directive | `versioned-left-directory-replacement` | `..._directive_form_unsupported` | `go-build` | reconciliation: two-token-left with no matching one-token-left |
| Windows path collision | `windows-case-colliding-declared-module-roots` | `..._containment_invalid` | `go-list` | containment under §2 platform path mapping |

Failure-boundary assignment matches the prose exactly: declaration and containment
before `go list`, form and bijection after `go list` and before `go build`. Every
rejection pins `build_permitted=false`, `go_build_started=false`, and
`persistent_state_changed=false`, so no case lets a rejected build touch installation
state. `TestGeneratedModuleRootConformanceVectors` locks the full name→(diagnostic,
boundary) map, the case count, the evaluation order, and the positive declaration.

## AC 2 — double regeneration

Reproduced in a fresh detached worktree at `bac193c`
(`curator-spec/.temp/TASK-260822-1so0ym/review-wt`):

| Command | Exit | Result |
| --- | ---: | --- |
| `go run ./tools/generate-vectors -root .` pass 1 | 0 | `git status` shows only `release/1.0.0-rc.8.json` modified — the committed `conformance/v1` tree already equals generator output |
| `go run ./tools/generate-vectors -root .` pass 2 | 0 | — |
| `cmp` of 688-file `conformance/v1` SHA-256 inventories | 0 | byte-identical |
| `go test ./tools/generate-vectors -run '^TestGeneratedModuleRootConformanceVectors$'` on committed bytes | ok | targeted regression green |
| `gofmt -l tools` | 0 | empty |

Stronger than the checklist asked: not only are passes 1 and 2 identical, pass 1 is
already identical to the committed bytes, so the vectors are reproducible from a
clean checkout.

## AC 3 — spec CI green

The branch's own CI (run 32632733803 on `bac193c`) was red on all three OSes with
`rc.8 downstream candidate pin does not match the suite manifest`. That failure is
structural, not a defect in this delivery: any authenticated new vector changes
`conformance/v1/manifest.json`, and the pre-existing prose commit `61ab801` had already
moved the rc.8 candidate pin. `bac193c` correctly left `release/1.0.0-rc.8.json`
untouched per the task instruction rather than mutating tagged history. The stop-the-line
packet named the exact external input; TASK-260822-c0rxj7 delivered it.

Green evidence on the qualified candidate, verified by the reviewer against GitHub, not
reported second-hand:

- curator-spec **Specification CI 32659168954**, head SHA exactly
  `6001dc33281b94a4ec7442ab15278550dd0f51d9`, conclusion `success`, six jobs green:
  Formatting, Links, Release target provenance, Specification on ubuntu/macos/windows.
- curator **candidate-conformance 32659157687**, conclusion `success`, fourteen jobs green.
  Job logs show `CANDIDATE_REF: 6001dc33...`, `revision accepted (immutable, full 40-hex)`,
  `manifest digest matches the supplied expectation`, `SPEC_PIN 00b1688a...` unchanged, and
  Candidate suite green on ubuntu/macos/windows.

Local reproduction on a detached worktree at `6001dc3`
(`curator-spec/.temp/TASK-260822-1so0ym/cand-wt`), jsonschema 4.25.1 from the pinned venv:

| Command | Exit | Result |
| --- | ---: | --- |
| `python3 tools/validate.py` | 0 | validated 53 schemas and 691 vector files |
| `python3 -m unittest discover -s tools` | 0 | 98 tests OK |
| `go test ./tools/...` | 0 | ok |
| `make validate` (all three above) | 0 | green |

## rc.8 immutability (independently checked)

`release/1.0.0-rc.8.json` blob IDs: `b92b105` `e05e4e92`, `61ab801` `c4bc6aae`,
`bac193c` `c4bc6aae`, `6001dc3` `e05e4e92`, `origin/main` `e05e4e92`. The candidate
restores rc.8 byte-for-byte to the `origin/main` blob. This task's commit introduced no
rc.8 change of its own.

## Non-blocking finding — carry forward, do not reopen this task

`build_module_root_declaration_invalid` is the fifth normative diagnostic in
`profiles/manager.md` and has **no** conformance vector anywhere in the suite (grep at
`6001dc3` hits only the prose). The JSON-Schema `schema-cases/*-skill-v8/invalid-module-roots-*`
files cover its syntactic subset — `.`, absolute, backslash, duplicate, parent, Windows
device — but not the two filesystem clauses that only a vector can express: a declared
directory that is not a real link-free directory in the snapshot, and one with no `go.mod`
directly inside it. The file's own doc comment says it "emits the filesystem and build-graph
cases that JSON Schema cannot express", so this is the gap in its own terms.

Two smaller bijection branches are likewise unvectored: two directives resolving to the
same declaration, and an unreadable annotation shape such as a three-token side.

None of these are named by this task's AC or checklist, and reopening for them would be
scope creep against an explicit AC. Recommended owner: fold into TASK-260822-10udu1's
landing scope or a follow-up under STORY-260822-1pm1c9 before the consumer implementation
stories build against these vectors.

## Verdict

Accepted. Implementation matches AC, fits the spec-repo generator/manifest/regression
architecture, and all gates are green on the qualified immutable candidate. Work is
already committed and pushed at `bac193c`, so no commit hand-off is outstanding.
