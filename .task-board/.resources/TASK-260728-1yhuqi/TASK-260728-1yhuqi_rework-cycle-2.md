# TASK-260728-1yhuqi — rework cycle 2

Closes the five blockers of `TASK-260728-1yhuqi_review-verdict-cycle-2.md`. The
independently supported direct-`swiftc` architecture, the SwiftPM rejection and
every cycle-1 closure are preserved unchanged; nothing in this cycle reopens
them.

Reviewer's own passing evidence is unchanged and was re-established here: 23/23
cases, 32 closure checks with no verdict, all prior structural checks matching.

---

## R1 — one exact compiler-version normalization rule

**The finding.** Three descriptions were in circulation. Decision 0011 section
11 admitted `\(.*\)` (empty parentheses allowed); reference section 1.2 required
`\(.+\)`; and the implementation matched the prefix and the numeric token before
the first space and never looked at the rest, so `Apple Swift version 6.3.2 x`
passed the implementation and failed both written rules.

**What closed it.** One byte-exact whole-value grammar, stated once in reference
section 1.2 as ABNF over bytes, pointed at rather than restated by decision 0011
section 11 and the registry `normalization` field, and implemented by one total
parser:

```abnf
banner  = prefix version SP "(" suffix ")"
prefix  = %s"Apple Swift version "        ; exactly 20 bytes
version = num "." num "." num
num     = "0" / ( %x31-39 *8%x30-39 )     ; no leading zero, 1..9 digits
suffix  = 1*200sbyte
sbyte   = %x20-27 / %x2A-7E               ; printable ASCII, excluding ( and )
```

plus three whole-value rules the productions do not carry: every byte in
`%x20-7E` (one rule that rejects CR, LF, NUL, every other C0 control, DEL and
every non-ASCII byte, so the value is ASCII-only and its own canonical form); a
1..255-byte whole-value bound checked **before** any scan; and hard anchoring
with no trimming at either end.

Excluding `(` and `)` from `sbyte` is what makes the grammar unambiguous without
balanced-paren parsing: an admitted value carries exactly one `(` and one `)`,
and the `)` is the last byte.

The parser is **total** — every input yields a normalization or exactly one of
eleven typed codes (`banner_empty`, `banner_too_long`, `banner_byte`,
`banner_prefix`, `banner_parens`, `banner_trailing`, `banner_separator`,
`banner_version_shape`, `banner_version_component`, `banner_suffix_empty`,
`banner_suffix_too_long`), and the conformance vectors assert on the **code**,
not on a boolean.

Both forms the two documents disagreed about are now explicitly rejected:
`Apple Swift version 6.3.2 ()` is `banner_suffix_empty` and
`Apple Swift version 6.3.2 x` is `banner_parens`.

**Suffix policy, stated rather than implied.** The suffix is required, non-empty
and bounded, and it is *consumed*: fixture `S29` reconstructs the whole value
byte-for-byte from the parsed components and requires equality. A parser that
stops at the first space cannot satisfy that, which is exactly the point.

**Vectors.** 32 whole-value vectors, `BV01`–`BV32`: 4 positives (the measured
banner, the minimal `0.0.0 (x)` form, a nine-digit component, and a suffix
exercising every printable non-paren byte) and 28 negatives covering the prefix,
the separator, trailing bytes, an empty suffix, the suffix bound, nested
parentheses, two/four components, a leading zero, a prerelease marker, an empty
component, a ten-digit component, and the whole byte-class family — LF, CR, NUL,
TAB, DEL and non-ASCII — plus the empty value and the length bound.

**New expected-red control.** `C10` restores the prefix-only parser from the
same binary and reports that it admits **17 of the 28 negatives**, so the
narrowing is evidenced rather than asserted. A unit test additionally requires
the retired parser to be strictly *looser* — never tighter — so no host the
grammar accepts is a regression.

---

## R2 — close and actually execute native runtime-library admission

**The finding.** Reference 1.3 and decision 0011 section 5 required only that
each `runtimeLibraryPaths` entry *already inside* the toolchain root exist. They
said nothing about an empty list, a dangling symlink, or an existing entry
outside every root. And probe P2 was declared but **never run**: `readTargetInfo`
ran P1 only, set `StdlibPresent` incidentally, and the green predicate never
gated on it.

**P2 is now executed and bound.** `swiftc -print-target-info -target
<target.triple>`, the exact triple the compile passes, is run from the
manager-owned empty working directory under the operation-private environment;
its whole result is recorded in the fixture (`native_target_admission.probe`),
and `green` gates on its verdict. Fixture `S31` asserts the argv carries
`-print-target-info -target arm64-apple-macosx26.0` and exited 0. The command
evidence carries the invocation.

**A new measurement changed the rule the reviewer proposed.** The reviewer's
suggested repair — every entry must resolve inside a fingerprinted root — was
tested against the host and **would reject it**. Measured, P2 for
`arm64-apple-macosx26.0` returns:

```json
["<toolchain-root>/usr/lib/swift/macosx", "/usr/lib/swift"]
```

`/usr/lib/swift` exists, is a directory, and is outside every fingerprinted root:
it is the Swift runtime macOS ships in the OS. The reviewer explicitly allowed
"or explicitly define and justify a different closed set", and that is what this
is.

**The closed set is three classes** — **A** in-closure, **B** base-installation
inside a closed registry-declared per-platform prefix set, **C** reject — with
seven mandatory rules (reference 1.3, R2.1–R2.7): non-empty list; every entry
absolute; every entry resolves (a dangling symlink is a rejection *distinct*
from an absence) and resolves to a directory; every entry classifies A or B with
fingerprinted roots matched first; **at least one** entry class A; every entry
serialized into identity as a `(role, relpath)` pair, class B under the reserved
role `platform-base-installation` so a runtime moving between the OS and the
closure moves cache identity; and a declared base prefix must be absolute,
exist, be a directory and **not** lie inside any fingerprinted root, which would
launder a class-A obligation.

Class B is the same trust boundary the acceptance test already accepts when it
requires the produced executable's dynamic dependencies to be base-installation
libraries of the declared baseline. The prefix set is **empty for every platform
but macOS**, so an unmeasured platform fails closed.

**Why the hatch is narrow, measured rather than argued.** Fixture `S34`:
the class-B prefix appears as **0 tokens** in the verified job plan. The linker
job's `-L` paths are `<root>/usr/lib/swift/macosx` — the class-A entry — and the
presented SDK's `usr/lib/swift`. Section 4.1's plan verification stays total and
admits no base-installation path in any bucket. Class B is an admission fact
about where the runtime could live, never an input handed to a child.

**Identity representation.** `runtime_library_members` joins
`curator-swift-toolchain-v1` as an ordered, deduplicated `(role, relpath)` list.
Measured value on this host:
`[{platform-base-installation, "."}, {swift-toolchain, "usr/lib/swift/macosx"}]`.

**Negatives.** `RV01`–`RV10`, all exercised against real filesystem state:
empty list; relative entry; absent in-root entry (the unserved-target case);
dangling symlink; regular file; existing directory outside every root; a symlink
that is lexically nowhere but resolves outside every root; base-installation
only; a base prefix declared inside a fingerprinted root; an empty entry. All
ten reject (`S33`).

**New expected-red control.** `C11` restores the cycle-1 rule and reports it
admits **6 of 6** shapes the closed rule rejects.

**Two honest notes.** P1 and P2 return byte-identical JSON on this host
(`p1_equals_p2: true`) — P2 is run and bound anyway, because that equality is a
property of a host, not of the contract. And measured, P2 for
`x86_64-unknown-linux-gnu` returns one class-A-shaped entry that does not exist,
so R2.3 rejects it: that is the admission test the representability surface
cannot perform.

---

## R3 — Windows toolchain identity is serializable before implementation

**The finding.** `curator-swift-toolchain-v1` serialized four macOS constants
(`usr/bin/swiftc`, `usr/bin/swift-frontend`, `usr/bin/clang`, `usr/bin/ld`)
while the runtime closure was defined structurally from the job plan and P3.
Windows had no rule for a linker that is not `usr/bin/ld`, no rule for
plan-derived members entering identity, no rule for more than one SDK root, no
`link_support_roles` value, and no closed schema for the admission task to mint.

**What closed it, in four parts.**

**(a) One serialization, `curator-swift-relpath-v1` (reference 2.4).** Every
member enters identity as a `(role, relpath)` pair. Both root and member are
**fully resolved first** — POSIX `realpath(3)`, Windows final path with reparse
points/junctions/symlinks resolved. Containment is component-wise and
byte-exact, so a case variant on a case-insensitive volume fails closed. The
separator is **always U+002F** on every platform; no volume prefix, drive
letter, UNC prefix, leading or trailing separator; the Windows `C:\`,
`\\server\share`, `\\?\C:\` and `\\?\UNC\` forms are stripped before splitting;
no empty, `.` or `..` component; a Windows final component keeps its extension
verbatim; a member equal to its root is exactly `"."`; case comes from the
resolved path as the filesystem reports it, never from an argument spelling.
File identity for permit-time re-verification is POSIX `(st_dev, st_ino)` and
Windows `(dwVolumeSerialNumber, nFileIndexHigh:nFileIndexLow)`, and is **not**
part of the portable identity.

`usr/bin/ld` and an unmeasured `usr/bin/link.exe` go through the same function.

**(b) The linker member is P3's answer, not a constant.** `linker` in the
identity is whatever `clang -print-prog-name=<linker>` resolves, serialized
relative to its containing root. `manager_invoked` carries only the executables
the manager itself starts — the registry `primary_relpath` and the C driver —
which are known in Stage A before any plan exists.

**(c) Plan-derived members enter a second object, and the reason is stated.**
The structural closure rule is only decidable after the graph phase, while the
toolchain identity is computed in Stage A. So the verified plan's executables go
into `curator-swift-process-closure-v1`, minted at permit time and carried in
the **receipt**, in two projections: `invoked` (relpaths as spelled — measured
**4**) and `resolved` (what they resolve to — measured **3**, because `swiftc`
is a symlink to `swift-frontend`). Fixture `S35` asserts the 4/3 inequality
rather than restating "two of them are one file". Binding only `resolved` loses
that the driver invokes `usr/bin/swiftc`; binding only `invoked` lets a
re-pointed symlink keep one identity while executing other bytes.

Keeping it out of the **cache key** is a stated completeness claim, not an
omission: the plan-derived set cannot vary independently of inputs the key
already binds — both root digests, the compiler version and tag, the native
target, the fixed argument vectors and the ordered source set. Putting it in
would force a graph phase before every cache lookup and buy no distinction.

**(d) Multiple SDK roots have a closed rule.** Root roles carry a per-platform
**cardinality**: `exactly-one` roles serialize bare (`platform-sdk`);
`one-or-more` roles serialize with an ordinal (`platform-sdk[0]`) assigned by
**declaration order** — and the bracket is present **even for a single root**,
so declaring a second root later cannot silently change the identity of the
first. Windows declares `platform-sdk` as `one-or-more`. Each ordinal gets its
own presentation chain under `<staging>/sdk/<ordinal>/` per section 2.2 and its
own `roots` entry hashed in ordinal order; `-sdk` takes ordinal 0; two
declarations resolving to the same real root are a Stage A failure, not a silent
collapse. Members are ordered by `(role_token, relpath)` byte order over the
**serialized** bytes and deduplicated on that key. `closure_sha256` is
domain-separated over ordered role-token/digest pairs, so an *n*-root closure
can never collide with an *n+1*-root one.

**The Windows registry values are now concrete**: `link_support_roles` is
`platform-sdk`, cardinality `one-or-more`, `data_only: true`, `qualified: false`;
`base_installation_prefixes` is **empty**.

**Windows remains unqualified**, and two things remain genuinely unmeasured and
are named as obligations rather than filled in: the plan-derived member
**count**, and the per-platform **argument template** for any SDK root beyond
ordinal 0. Every other serialization question the count touches is answered
today by rules that do not depend on it. No Windows claim is recorded.

**Vectors.** `SV01`–`SV06` negatives (ancestor; sibling sharing a lexical
prefix; case variant; unresolved `..`; empty root; empty member) all reject, and
the root-relative-to-itself case yields `"."` (`S36`). `S37` asserts the ordinal
tokens, the `(role, relpath)` ordering and the duplicate collapse.

---

## R4 — the remaining exact-vector and line-grammar ambiguity

**The finding.** Both normative documents displayed `swiftc` in two command
lines and then said the vectors "differ in exactly one token, at index 0". Read
as complete argv that is wrong: index 0 is the program and `-###` is an
insertion after it. Separately, the reference grammar rejected every control
byte while `VerifyPlan` silently removed one trailing CR from every line, and
the terminal newline was `LF?` — optional.

**The vector relation is now stated as a construction**, verbatim in both
documents, in the conformance vectors and in the implementation:

```text
program      := the resolved absolute swiftc inside the swift-toolchain root
compile_args := [ "-swift-version","6","-O","-target",T,"-sdk",S,
                  "-module-name",M,"-no-color-diagnostics", <sources…>, "-o", A ]
graph_args   := [ "-###" ] ++ compile_args
compile_argv := [ program ] ++ compile_args
graph_argv   := [ program ] ++ graph_args
```

with seven mechanically asserted properties (V1–V7): `len(graph_args) ==
len(compile_args)+1`; `graph_args[0] == "-###"`; `graph_args[i+1] ==
compile_args[i]` byte-exact for every `i`; `graph_argv[1] == "-###"` — the
insertion point in complete argv; `-###` occurs **exactly once** in `graph_argv`;
**never** in `compile_argv`; and both vectors share one `program`.

Uniqueness is asserted even though `-###` cannot collide with another token,
because "cannot collide" is an argument and an assertion is a check. Five
negatives (`VV01`–`VV05`): no insertion; doubled; appended instead of inserted;
present in the compile vector; a second token co-mutated under cover of the
insertion. All five reject (`S39`); `S38` asserts V1–V7 on the live vectors.

**The physical-line grammar is now total and stated separately** (reference
4.1.1 layer 1), because it is the layer a Windows implementation would otherwise
have to guess at:

```abnf
plan        = 1*line-record
line-record = line LF                  ; the TERMINAL LF IS MANDATORY
line        = 1*lbyte
lbyte       = %x20-7E / %x80-FF        ; no C0 control, no CR, no DEL
```

Every previously open case is decided: LF is the only terminator on every
platform; the **terminal LF is mandatory** — measured, the plan's final byte is
0x0A, so a plan not ending in LF is a truncated read and rejects, which is what
removes the `LF?` ambiguity and makes "exactly one trailing empty element" a
consequence rather than a convention; a **bare CR rejects anywhere**, never
stripped, so CRLF is a rejection; every other control byte and DEL reject —
the *same* class the token grammar rejects, so the two layers cannot disagree;
bytes `0x80–0xFF` are **admitted** inside a line and compared byte-exactly,
because a POSIX path is a byte string and the plan is not required to be valid
UTF-8; blank lines reject; and the plan (8 MiB), each line (64 KiB) and the line
count (4096) are bounded, so an adversarial plan is never an unbounded read.

A Windows implementation therefore reads the child's stdout as a **binary
stream with no newline translation**. A Windows toolchain measured to emit CRLF
extends the grammar by measurement; it is never tolerated at runtime.

`VerifyPlan` no longer strips anything: a line reaching the token layer is
exactly the bytes between two LFs.

**Measured** for the driver's own vector: 4808 bytes, final byte 0x0A, 3 LF,
**0 CR**, 0 other control bytes, 0 non-ASCII bytes, longest line 1990 bytes, 3
job lines, stderr empty.

**Negatives.** `LV01`–`LV12`: empty plan; no terminal LF; CRLF; bare CR before
the terminator; bare CR mid-line; trailing blank line; leading blank line; blank
line between jobs; NUL; TAB; DEL; a line past the bound. All twelve reject; the
retired splitter admits **11** of them (`S41`).

**New expected-red control.** `C12` restores the CR-normalizing splitter and
reports the 4 shapes it admits.

---

## R5 — traceability for the measured prerequisite

`TASK-260729-rhjxtx` is now linked `blocked_by` on the board — verified in the
board projection, which reads
`["TASK-260728-2spy93","TASK-260728-1g0z69","TASK-260729-rhjxtx"]` — and both
normative documents carry an **inherited-measurements table** that traces each
consumed fact individually rather than crediting the task in general:

| # | Inherited | Consumed in | Status |
|---|---|---|---|
| M1 | the split-stream `--version` banner | reference 1.1, decision 8 and 11 | reproduced here; the two agree |
| M2 | finding 6, frontend spawn for an unserved target | reference 1.3, decision 5 | **refined**: reproduces under a compile-only vector, not under this driver's linking vector |
| M3 | Linux needs `swift-autolink-extract`, a fifth executable | reference 2.1 and 12.3, decision 3 and 13 | consumed unchanged |
| M4 | no Swift toolchain on the reachable Windows host, 19 `not_run` | reference 12.2, decision 13 | consumed unchanged |

The metadata-readability, manifest-execution and selector-inertness observations
were re-measured here and are cited from this task's own evidence rather than
inherited, and both documents say so.

---

## Probe results, before and after

| | cycle 1 | cycle 2 |
|---|---|---|
| cases | 23 matched / 0 divergences | **23 matched / 0 divergences** |
| closure checks | 32, 0 verdicts | **32, 0 verdicts** |
| expected-red controls | 9 of 9 red | **12 of 12 red** |
| structural checks | 30, 0 divergences | **44, 0 divergences** |
| executed native admission | not run | **run and gating**: in-closure 1, base-installation 1, 0 rejections |
| green | true, exit 0 | **true, exit 0** |
| degraded run | 23 `not_run`, exit 0 | **23 `not_run`, exit 0** |

Probe module: 22 sources, 6 test files, 44 test functions, 6387 lines. The
tarball extracts clean and `gofmt -l`, `go vet`, `go build` and `go test` all
exit 0 from the extracted copy.

Every control was also replayed **individually**; all twelve exit 1.

## Gates

- curator `gofmt -l .` and `make check` are **expected-red on this repo state
  and were red before this task started**: 1141 listed paths, **0** under
  `.temp/TASK-260728-1yhuqi`, **0** outside `.temp/`, **0** modified tracked
  files. Every listed path is a scratch tree belonging to another task.
- spec `tools/validate.py` fails **only** its link check, on 3 links in 2
  documents this task did not author and only copied into the worktree so the
  tree validates as a whole. The clean baseline at `57c1f56` exits 0 (30
  schemas, 93 vector files). Scoped over the two documents this task **did**
  author: **6 links, 0 broken**.
- Nothing staged, committed, pinned or published. Nothing installed on any host.
  No platform widening: macOS arm64 remains the only measured pair and Windows
  and Linux remain unqualified.
