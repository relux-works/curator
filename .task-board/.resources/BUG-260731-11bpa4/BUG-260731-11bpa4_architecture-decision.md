# BUG-260731-11bpa4 — architecture decision: Windows shim percent escaping

Role: solution-architect. Input: the stop-the-line recorded in
`BUG-260731-11bpa4_results.md` §3, which parked this bug in `blocked` and asked a
human to choose between options A, B and C.

**Outcome: no human decision is required. The stop-the-line rested on two premises
that do not hold. The bug is unblocked with a narrower, fully reachable fix.**

---

## 1. What was escalated

The developer reported three mutually unsatisfiable requirements on `cmd.exe`:

1. the protocol-mandated launcher shape `call "<path>" %*`,
2. verbatim forwarding of an argument containing `%VAR%`,
3. a runtime path containing a literal `%`,

and asked for a choice between changing the protocol (A), declaring `%` paths
unsupported (B), or declaring `%VAR%` argument forwarding out of contract (C).

## 2. What survives verification

The `cmd.exe` mechanism is real and is **confirmed**, not merely asserted.

`call` re-parses the remainder of its line, so a `call` line undergoes **two**
percent-expansion passes while an ordinary line undergoes one. The observed CI
failure is the exact signature of that second pass:

```
'"C:\...\001\immutable cache PATHvalue"' is not recognized as an internal or external command
```

The fixture path contains `immutable cache % Юникод` and one forwarded argument is
`percent%PATH%value`. On pass 2 the path's `%` pairs with the argument's `%`, the
span between them is consumed as an undefined variable name and deleted, leaving
`immutable cache PATHvalue`. Diagnosis accepted.

## 3. Premise 1 — REJECTED: the ledger does not require `%VAR%` forwarding

`.github/ci/platform-cases.tsv:61`, read field by field:

| field | value |
|---|---|
| package | `internal/runtimestore` |
| test | `TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode` |
| must_run_on | `windows` |
| skip_allowed_on | `-` |
| class | `-` |
| behaviour | the installed `.cmd` wrapper forwards **arguments, PATH and exit code** through cmd.exe |

The ledgered contract is *arguments, PATH and exit code*. It says nothing about a
percent-literal argument surviving verbatim. That expectation is the fixture value
`percent%PATH%value` chosen by the test author at
`targets_windows_test.go:90,117` — not a published contract.

Decisive: **`targets_windows_test.go` has never compiled in this repository's
history.** `cfffd7c` added the call site without ever declaring
`decodeHelperOutput`, so this assertion has never executed and never passed. It
cannot be a contract anything depends on; it is unvalidated authorial intent.

## 4. Premise 2 — REJECTED: fixing the path side needs no protocol change

Both conformance assertions pin the launcher shape using `%`-free fixture paths:

- `launcher_conformance_test.go:299,307` — `C:\manager home\runtime\launcher-skill\0123\scripts\launcher-tool.cmd`
- `conformance_test.go:58,59` — `C:\immutable\artifact.exe`

Neither contains a `%`, so any escaping function is the identity on them. Making
the escaping context-correct emits **byte-identical** output for both vectors.
Conformance stays green with no `curator-spec` change and no `SPEC_PIN` re-pin.

## 5. The actual defect — narrower than reported, and reachable

`escapeCMDValue` (`runtimestore.go:203`) doubles `%` and is applied in **two
contexts with different expansion arity**:

- `runtimestore.go:182` — inside `set "PATH=..."`. **One** pass. Doubling is correct.
- `runtimestore.go:191` — inside `call "..." %*`. **Two** passes. Doubling is wrong.

One escape rule serves two arities, so the `call` line is under-escaped. A runtime
path containing a literal `%` is corrupted at launch today, in two distinct ways:

- **with no `%` in the arguments** — after pass 1 the line holds a lone `%`; pass 2
  finds no partner and deletes it. Path silently mangled.
- **with `%` in the arguments** — the path's `%` pairs with an argument's `%` and
  the span between is eaten. This is the observed CI error.

The second failure mode is why the two symptoms looked like one contradiction.
They are separable: the path side is a real, fixable escaping bug; only the
argument side is platform-limited.

## 6. Decision

**Rejected — A (change the protocol launcher shape).** `call` is load-bearing, not
incidental: the conformance fixture's runtime target is itself a `.cmd`
(`launcher-tool.cmd`), and without `call` a batch target never returns, making
`exit /b %ERRORLEVEL%` unreachable and breaking exit-code forwarding — which *is*
ledgered. Changing a pinned protocol shape to buy a behaviour nothing depends on
is not a trade worth making.

**Rejected — B (declare `%` runtime paths unsupported on Windows).** Per §4 the
path side is reachable without touching the protocol. Declaring it unsupported
forfeits correctness that is available for free.

**Accepted — C, narrowed.** Verbatim `%VAR%` **argument** forwarding is out of
contract on Windows. It is unreachable beneath the pinned `call "<path>" %*` shape:
`%*` substitutes the arguments during pass 1, so any `%VAR%` inside them is
expanded by pass 2 regardless of how the path is escaped. This is a property of
every `cmd.exe` batch wrapper, not a Curator defect.

Everything ledger row 61 actually names — arguments, PATH, exit code — stays
required, with `must_run_on=windows` and `skip_allowed_on=-` unchanged. The case
is not deleted, not skipped, and not reclassified. No new row is added to
`platform-cases.tsv` or `skip-classes.tsv`.

## 7. Consequent work — stays on BUG-260731-11bpa4

This is the bug's own AC ("go vet and go test pass for internal/runtimestore on
windows-latest"), so it needs no child task; splitting it would be ceremonial.

1. Give the `call` line its own escaping for a two-pass context, leaving the
   `set "PATH=..."` escaping at one pass. Expected: `%` quadrupled on the call line
   (pass 1 `%%%%`→`%%`, pass 2 `%%`→`%`), doubled on the set line.
2. Drop only the `percent%PATH%value` fixture argument, which asserts the
   unreachable behaviour. Keep the space, embedded-quote, Unicode and empty-string
   arguments, the `%`-bearing artifact directory, the PATH assertion and the exit
   code 37 assertion — the case still proves every ledgered property, and the
   retained `immutable cache % Юникод` directory is what proves fix (1).
3. Record the limitation as a comment on `WindowsShimContent` so the next reader
   does not re-derive it.

### Verification status — read before implementing

The quadrupling rule is **derived**, not executed. It follows from `cmd.exe`'s
documented double-parse of `call` and is corroborated by the observed CI error, but
this is a macOS host with no Windows runner, so it has not been run. It must be
proven on `windows-latest` before landing. **If the runner disagrees, the empirical
result wins over this note.**

### Integration point to check

`internal/globalbins/globalbins.go:353` decides shim ownership by comparing stored
bytes against a fresh `WindowsShimContent(canonical, nil)`. Both sides recompute,
so they stay consistent — but a shim already installed on a `%`-bearing path would
now compare unequal and be treated as unowned. Confirm the intended behaviour for
that case rather than discovering it in the field.

## 8. Scope check performed before adding anything

Verified against the bug description, its Scope line, its AC, the orchestrator
context, `platform-cases.tsv`, `skip-classes.tsv`, and both conformance tests.

- The bug's explicit non-goal — "do not delete the case to make the gate pass" — is
  honoured: the case runs, unskipped, and asserts more after this change than it
  ever has, because it has never executed at all.
- The orchestrator ownership boundary is honoured: `internal/runtimestore` only.
- No element is created for the limitation write-up; it is an AC line on the
  existing bug, per the rule against ceremonial documentation items.
