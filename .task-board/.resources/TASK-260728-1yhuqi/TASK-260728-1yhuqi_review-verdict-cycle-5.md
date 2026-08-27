# TASK-260728-1yhuqi — review verdict cycle 5

## Verdict

**CHANGES REQUESTED**

Route to `analysis`. This is ordinary contract/probe rework, not a stop-the-line
boundary.

Cycle 4's two stated findings are closed in the normative documents:

- `manager-worker-v2` is again one `swiftc -###` graph command followed by one
  `swiftc` compile command; the manager does not execute plan jobs.
- `curator-swift-source-admission-v1` rejects malformed UTF-8, NUL, and every
  raw `@`/`#` source byte before the graph phase. Independent replay rejected
  `@Observable` and `#Predicate` with zero manager-started commands, while
  admitted source compiled with the plugin paths present and zero macro-load
  remarks.

Acceptance is withheld because the probe's actual session omits a mandatory
security step that the contract states as part of the compile permit.

## Blocking finding — permit-time re-binding is not in the exercised session

Decision 0011 section 4 and reference section 4.1.4 require the manager,
immediately before starting `compile_argv`, to re-resolve and re-identify every
verified executable, plugin, search, source, and output binding, and to confirm
that every plugin path verified absent is still absent.

The submitted probe implements `Reverify`, and isolated checks S20/S21 call it,
but the cycle-4 `session` does not:

1. `AdmitSources`
2. run `swiftc -###`
3. `VerifyPlan`
4. run `swiftc compile_argv`

There is no `Reverify(b.Verify.Bindings)` call between steps 3 and 4.
Repository-wide search of the extracted probe finds `Reverify` only in its
unit test and the isolated S20/S21 checks; none of the S60/S62/S64 happy-path
sessions invokes it.

Consequences:

- an absent plugin path may appear after graph verification and the exercised
  session still starts the compile command;
- an executable, search path, SDK presentation, source binding, or output
  parent may change after graph verification and the exercised session still
  starts the compile command;
- the green S62 result proves command cardinality, but not the contract's
  stronger statement that command 2 received a permit after binding
  re-verification;
- the source-admission verdict is not exercised together with the source
  binding immediately before compile. The immutable-snapshot rule is normative,
  but the probe does not demonstrate the stated defense if a source is changed
  or re-pointed after Stage B/graph.

This is not only a missing call. The current output binding cannot be
re-verified on the normal absent-output path: `checkBucket` resolves and stores
the output parent's file identity but retains the not-yet-existing output path
as `Raw`; `Reverify` later calls `EvalSymlinks(Raw)`, which fails while the
output is correctly still absent. Integrating the current function unchanged
would therefore reject a valid happy path. The binding must preserve and
re-check the actual path whose identity was recorded.

The absent-plugin branch should also fail closed on `Lstat` errors other than
`ENOENT`; such an error is not proof of continued absence.

## Required rework

1. Put permit-time re-binding into the one authoritative session immediately
   before `compile_argv`. A finding must return
   `build_execution_control_unavailable`, a stable Swift detail, exactly one
   manager-started command, and no artifact.
2. Make output binding explicit: preserve the operation-private parent path
   whose identity was checked when the output does not yet exist, and re-check
   that same path. Do not make a valid absent output fail merely because the
   final file has not been created.
3. Treat every re-verification filesystem error fail-closed, including
   non-`ENOENT` errors while checking an absent plugin path.
4. Add integrated structural cases, not only unit checks:
   - happy path reaches two commands only after all bindings reverify;
   - an absent plugin appears after graph;
   - a source is replaced or changed after Stage B/graph, including a change
     that introduces `@` or `#`;
   - an executable/search binding changes;
   - the output parent is replaced or re-pointed.
   Each adversarial case must prove that the compile command did not start.
5. Add an expected-red control for a session that verifies the graph but skips
   permit-time re-binding, then replay the native/degraded fixtures and every
   control. Update decision/reference text only if the repaired binding shape or
   stable diagnostic differs from the current normative contract.

## Independent evidence

- Submitted archive SHA-256:
  `3446bfa42c24bad7102982b0b5e117f9a7f31cf9b3956c45eed37bb1510ba16b`,
  matching the producer gate log.
- Standalone module:
  - `gofmt -l .`: exit 0, no paths
  - `go vet ./...`: exit 0
  - `go test ./... -count=1`: exit 0
  - `go build`: exit 0
- Native replay on Apple Swift 6.3.2 / macOS 26.5 arm64:
  - 23/23 cases matched
  - 32 closure checks, 0 yielded a verdict
  - 15/15 controls failing as required
  - 62/62 structural checks matched
  - executed P2 admission held
  - aggregate green
- Degraded replay: 23 cases `not_run`, exit 0, nothing installed.
- C1-C15 were replayed individually; every control exited 1.
- Submitted decision/reference/archive hashes match the producer's recorded
  gate hashes.

The green results support the design decisions and close cycle 4, but they do
not override the missing compile-permit step in the session they exercise.
