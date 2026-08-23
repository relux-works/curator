# Reviewer verdict for TASK-260811-1u42b9

Verdict: **changes requested -> to-dev**

## Scope and goal evidence

- Reviewer run: `RUN-260822-af01e0`.
- `task-board spawn goal RUN-260822-af01e0`: no active goal; the run is not goal-bound.
- No run directives were recorded at the review checkpoints.
- Reviewed rework outcome: `TASK-260811-1u42b9_rework-evidence_RUN-260822-a5f43f.md`.
- Review was read-only; no product or test code was modified.

## Accepted rework

1. npm cache, `npm ci`, and Node calls now resolve concrete argv, cwd,
   environment, input mounts, work copies, and typed write roots from closed C5
   templates before committing a canonical `closureexec.DerivationPermit`.
2. Portable execution uses `closureexec.ManagerProcessRunner` and directly
   launches the C0-selected relative executable; portable receipts retain
   `network=not-observed` without synthesizing lossless process/read/write
   absence.
3. Materialized external packages are recursively re-admitted and compared
   file-for-file, including executable mode and embedded metadata, with exact
   extraction evidence from the admitted tarballs.
4. Verified negotiation retains the missing/incomplete/incompatible/cross-mode/
   drift zero-start matrix, and canonical executor comparison rejects an extra
   reported process.
5. Closed lock/schema parsing, workspace and target selection, SRI plus Curator
   digest binding, lifecycle/bundle/node-gyp/native/opaque denial, private-cache
   receipts, real offline npm materialization, and exact installed-package
   reconciliation remain green.

## Required changes

### 1. The real npm launcher still crosses an undeclared ambient executable

The integration path stages the resolved npm entry point and binds that file as
the package-manager executable (`internal/npmsource/conformance_test.go:842-860`,
`905-925`). On the reviewed host the selected file is
`/opt/homebrew/lib/node_modules/npm/bin/npm-cli.js` and its first line is
`#!/usr/bin/env node`. Executing that script necessarily crosses
`/usr/bin/env` before the staged Node runtime.

The permit allows only the npm script and staged Node
(`internal/npmsource/materialize.go:498-509`); `/usr/bin/env` is neither a C0
tool node nor an allowed process. The committed environment also appends
`/usr/bin:/bin` to `PATH` (`materialize.go:921-922`), preserving ambient
executable lookup even though the accepted no-hidden-edge contract explicitly
says host `PATH` cannot repair a missing process/tool edge. A genuine provider
claiming `exact-executable-allowlisting-v1` therefore cannot execute and report
this real command while matching the permit.

The new positive verified fixture does not expose the mismatch. It runs the
adapter fake runner and then constructs `Audit.Processes`, reads, writes, argv,
cwd, and environment by copying the permit (`conformance_test.go:577-599`,
`1295-1319`). It never observes the real npm shebang chain, so its successful
receipt is circular evidence rather than proof that the real compatible
lossless provider can enforce this command.

Required rework: remove the hidden interpreter edge. Prefer executing the
exact C0-bound Node binary directly with an admitted/fingerprinted npm CLI
script as an explicit read/tool input. Alternatively, model and fingerprint
the interpreter launcher as a separate exact tool/process and remove ambient
`PATH` fallbacks. Add a positive verified execution vector whose process
evidence comes from the actual real launch boundary, plus a negative proving an
unbound shebang/interpreter or PATH fallback fails before the npm operation is
accepted.

### 2. The claimed shared S03/S04/S08 vectors are names, not executable proofs

`TestS02S03S04S07S08N12N13OfflineMaterializationGates` contains only three
subtests (`internal/npmsource/conformance_test.go:338-386`): one discards a
pre-seeded project `node_modules`, one rejects an extra installed package, and
one removes a private-cache member. It does not:

- seed an ambient npm cache with the same package identity and different bytes
  (`S03`);
- make a declared build/download action attempt network and prove the stable
  `closure_network_attempted` boundary with no publication (`S04`); or
- replay once with empty caches and once with a poisoned inaccessible ambient
  cache and compare graph/output identities and audited network state (`S08`).

The real npm test performs one replay only. The verified extra-process test
copies the permit's `Network: none` rather than exercising a network-attempt
observation. Test names cannot satisfy the task's explicit S01-S08 acceptance
gate.

Required rework: add the actual npm-wrapped S03, S04, and two-run S08 fixtures,
assert their stable diagnostics/checkpoint boundaries/start and publication
counters, and preserve honest portable evidence where a lossless observation
is unavailable.

## Fresh validation

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/closureexec ./internal/nodesource ./internal/npmsource` | 0 | `focused-01.log`; artifactpolicy `36.986s`, closureexec `4.750s`, nodesource `1.480s`, npmsource `9.580s` |
| Focused materialized-byte, C5, real npm/Node, portable evidence, and verified positive/negative vectors | 0 | `reviewfocus-01.log`; `5.280s` |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-01.log`; `31.733s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-01.log`; `80.4%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `vet-01.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `lint-01.log`; zero issues |
| `go build ./...` | 0 | `build-01.log` |
| `git diff --check` | 0 | `diffcheck-01.log` |
| `task-board validate` | 0 | `board-01.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-01.log`; cmd/curator `392.929s`, artifactpolicy `136.153s`, npmsource `48.611s`, rustsource `145.123s` |

Green gates validate the implementation's current expectations. They do not
close the findings because the verified fixture manufactures the authority it
is meant to observe, the real launcher has an undeclared interpreter/process
edge, and three mandatory semantic vectors are absent.
