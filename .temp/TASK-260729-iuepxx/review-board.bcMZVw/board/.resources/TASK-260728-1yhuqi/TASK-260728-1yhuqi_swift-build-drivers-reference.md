# Swift build drivers: `swift-v1` and `swift-repository-v1`

Implementation-ready reference for the closed local `swift-v1` and external
`swift-repository-v1` drivers decided in
[decision 0011](../decisions/0011-swift-driver-pair.md), under the boundary of
[decision 0008](../decisions/0008-additional-language-driver-boundary.md), the
toolchain contract of
[decision 0007](../decisions/0007-compiled-build-toolchain-preflight.md) and the
execution policy of
[decision 0006](../decisions/0006-portable-manager-worker-execution.md).

Both identifiers are **reserved**, not admitted. Every schema, including manifest
schema 8 as first minted, MUST reject them until `TASK-260728-251p01` moves them
in the same change that mints the schema admitting them. Nothing in this document
is a platform claim.

Measured values below come from one host — macOS 26.5 arm64, Apple Swift 6.3.2
(`swiftlang-6.3.2.1.108`), `XcodeDefault.xctoolchain` and `MacOSX.sdk` from Xcode
26.5 — and from the `swift-boundary-fixture-v1` probe run recorded with
`TASK-260728-1yhuqi`.

---

## 1. `toolchain-registry-v1`: the `swift` entry

| Field | Value |
|---|---|
| `toolchain_id` | `swift` |
| `primary_relpath` | POSIX `usr/bin/swiftc`; Windows `usr\bin\swiftc.exe` |
| `fingerprint_algorithm` | `curator-swift-toolchain-v1` (section 2) |
| `baseline` | `{"kind":"at_least","min":"6.3.2"}` |
| `compatibility` | families `{(6, 3)}`, granularity `(major, minor)` |
| `platforms` | `[(macos, arm64)]` |
| `companions` | none |
| `link_support_roles` | `macos`: `[platform-sdk]`, data-only |
| `metadata_sources` | `Package.swift` first-line `swift-tools-version`; `.swift-version` |

`compatibility` and `baseline` are gates, never build inputs. Lowering the
baseline requires measuring the older release. Adding a family requires running
the driver's conformance vectors against it. Neither may be derived from version
ordering, and no package or descriptor byte reaches either.

### 1.1 Probe vectors

Run once per operation from the manager parent during Stage A, from a
manager-owned empty working directory, under the section 5 environment.

| # | argv | Reads |
|---|---|---|
| P1 | `swiftc -print-target-info` | stdout JSON: `compilerVersion`, `swiftCompilerTag`, `target.triple`, `target.unversionedTriple`, `paths.runtimeLibraryPaths` |
| P2 | `swiftc -print-target-info -target <native-triple>` | stdout JSON: `paths.runtimeLibraryPaths` for the exact triple the compile will use |
| P3 | `clang -print-prog-name=ld -target <native-triple>` | stdout: the absolute linker path the C driver will resolve |

`swiftc` and `clang` are the absolute paths `<root>/usr/bin/swiftc` and
`<root>/usr/bin/clang`. A probe MUST NOT be spelled as a bare name.

**Forbidden as version probes.** `swift --version` and `swiftc --version` each
split one banner across two streams: `swift-driver version: 1.148.6 …` on
stderr without a trailing newline, the Apple version line on stdout. A consumer
that merges the streams sees them concatenated into one line, and an anchored
rule stops matching. Measured on this host and independently in
`TASK-260729-rhjxtx`.

**Forbidden entirely.** No `swift build`, `swift package`, `swift run`,
`swift test`, `dump-package`, `describe`, `resolve` or `show-dependencies`
invocation may appear in any stage, for any purpose, including diagnostics and
dry runs. Reading what a SwiftPM package declares executes the package
(decision 0011 section 2).

### 1.2 Normalization — `swift.printTargetInfo.compilerVersion`

Input: the `compilerVersion` member of P1's **stdout** JSON.

Accepted form, anchored, whole value:

```
^Apple Swift version (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*) \(.+\)$
```

The three capture groups are the normalized version. `swiftCompilerTag` is
recorded in the identity and is **not** the version.

- No prerelease marker is admitted by this rule. A banner that does not match
  leaves the version **undetermined**, which fails Stage A with
  `build_toolchain_version_undetermined`. It is never guessed, never truncated,
  and never taken from a second surface.
- The open-source banner form (`Swift version 6.1 (swift-6.1-RELEASE)`) is
  **not** admitted. No host in this task carried one; admitting it is a
  qualification obligation with its own measurement, not a regex change.

### 1.3 Native target admission

| Step | Rule |
|---|---|
| identity | `target.unversionedTriple` from P1. Measured: `arm64-apple-macosx`. |
| compiler argument | `target.triple` from P1. Measured: `arm64-apple-macosx26.0`. |
| admission | every entry of P2's `paths.runtimeLibraryPaths` that lies inside the toolchain root MUST exist and MUST be a directory |

The identity is the **unversioned** triple, because the versioned one carries a
deployment-version component supplied by the SDK; using it would move cache
identity on an SDK update that changed nothing else in the closure.

The unversioned triple is **not** a valid `-target` argument. Measured:
`swiftc -print-target-info -target arm64-apple-macosx` exits 1 with
`error: Swift requires a minimum deployment target of macOS 10.9.0`.

`-print-target-info` is a **representability** surface, not an admission test.
Measured: `-print-target-info -target x86_64-unknown-linux-gnu` exits **0** and
names `<root>/usr/lib/swift/linux`, which does not exist in the tree. An unknown
triple (`not-a-real-triple`) exits 1.

Failure: `build_toolchain_platform_unsupported`, Stage A, before source
acquisition and before any compiler child.

---

## 2. `curator-swift-toolchain-v1`

### 2.1 Resolution

Two roots are resolved, in this order, through exactly the two declaration
channels decision 0007 section 3 fixes — a root bundled with the manager
distribution, or trusted operator configuration in manager-owned owner-protected
state:

| Role | Contents | Contributes a process |
|---|---|---|
| `swift-toolchain` | the toolchain root carrying `usr/bin/{swiftc,swift,swift-frontend,clang,ld}` | yes |
| `platform-sdk` | the SDK tree the compiler and linker read | no |

Forbidden origins for **both**, with the same force and the same diagnostics:
`PATH`, the inherited environment, `xcrun`, `xcode-select`, `DEVELOPER_DIR`,
`TOOLCHAINS`, a package byte, a descriptor byte, a version-manager shim
(`swiftly`, `swiftenv`), a network fetch, an installer.

A missing or unusable declaration fails Stage A with
`build_toolchain_root_undeclared` or `build_toolchain_root_unusable`, before any
source is acquired.

**Closure members.** The required set is defined structurally, not as a fixed
list, because it is platform-determined:

> Every executable the verified job plan names (section 4.1), plus the linker
> `clang -print-prog-name=ld` resolves, MUST resolve — following symlinks — to a
> regular executable file **inside** the `swift-toolchain` root.

On macOS that instantiates to exactly four: `usr/bin/swiftc`,
`usr/bin/swift-frontend`, `usr/bin/clang`, `usr/bin/ld`. Measured on this host:
`swiftc` is a symlink to the single `swift-frontend` binary, which dispatches on
`argv[0]`. A member resolving outside the root fails
`build_toolchain_root_unusable`.

**`usr/bin/swift` is NOT a member.** It is the SwiftPM launcher, section 1.1
forbids it from every stage, and the driver never invokes it, so requiring it to
resolve would add a portability constraint with no property behind it. Its bytes
are inside the fingerprinted root and are covered by `tree_sha256` regardless.
This document uses it in exactly one place — as the upstream oracle of section
7.4 — which is a conformance-probe role outside any manager pipeline. It is
therefore **probe-only: absent from the runtime closure, absent from the
registry's required member set, and forbidden from the pipeline.**

Linux was measured to add a fifth member, `swift-autolink-extract` (section
12.3). Windows is unmeasured; section 12.2 states the obligation without naming
a count it cannot support.

### 2.2 SDK presentation (mandatory)

The manager MUST NOT pass the declared SDK path to `-sdk`. It creates an
operation-private directory it owns entirely and presents the SDK through it, at
a fixed nesting depth:

```
<staging>/sdk/                                created by the manager; contains exactly `present/`
<staging>/sdk/present/                        created by the manager; contains exactly `SDKs/`
<staging>/sdk/present/SDKs/<name>        ->   <declared platform-sdk root>
```

and passes `-sdk <staging>/sdk/present/SDKs/<name>`, where `<name>` is the base
name of the declared root.

Why: the compiler derives external macro plugin search paths from the `-sdk`
argument, **three ancestor levels up and then into `Developer/usr`**. Measured
with the declared Xcode SDK path passed directly, **two distinct** derived paths
exist outside every fingerprinted root —
`…/MacOSX.platform/Developer/usr/lib/swift/host/plugins` and
`…/MacOSX.platform/Developer/usr/bin/swift-plugin-server` — and `#Predicate`
loads `FoundationMacros` through a `swift-plugin-server` process in that tree.

Why *this depth*: three levels up from `<staging>/sdk/present/SDKs/<name>` is
`<staging>/sdk/present`, so every derived tree lands inside `<staging>/sdk`,
which the manager creates and keeps empty apart from the presentation chain.

Measured with the presentation above, over two frontend jobs: 14 plugin
components in the plan; **6 distinct** paths, of which 2 exist and are both
inside the toolchain root, 4 do not exist, and **0 exist outside a fingerprinted
root**. Every SDK-derived component lands inside `<staging>/sdk` and none of
them exists. `#Predicate` fails to compile; `@Observable` compiles with exit 0
and loads `<root>/usr/lib/swift/host/plugins/libObservationMacros.dylib` **in
process**, with no server.

The manager MUST guarantee, at creation and before the graph phase, that:

1. `<staging>/sdk` is freshly created and contains exactly `present/`;
2. `<staging>/sdk/present` contains exactly `SDKs/`;
3. `<staging>/sdk/present/SDKs` contains exactly the one symlink;
4. no entry named `Developer`, `usr`, `SDKs` or `Toolchains` exists directly
   under `<staging>` or `<staging>/sdk` other than the presentation chain, so a
   derivation walking up to four levels still finds nothing; and
5. the whole `<staging>` tree is operation-private manager-owned state with no
   other admitted writer.

Failure: `build_execution_control_unavailable`.

The presentation is a defence, not the proof. Section 4.1's plan verification is
the proof, and it holds whatever the derivation rule becomes — a derived tree
that *did* exist outside a fingerprinted root would be rejected there regardless
of where it came from.

### 2.3 Identity

```json
{
  "algorithm": "curator-swift-toolchain-v1",
  "swift_version": "<compilerVersion, verbatim>",
  "swift_compiler_tag": "<swiftCompilerTag, verbatim>",
  "launcher_relpath": "usr/bin/swiftc",
  "frontend_relpath": "usr/bin/swift-frontend",
  "c_driver_relpath": "usr/bin/clang",
  "linker_relpath": "usr/bin/ld",
  "roots": [
    {"role": "swift-toolchain", "tree_sha256": "sha256:<hex>"},
    {"role": "platform-sdk",    "tree_sha256": "sha256:<hex>"}
  ],
  "closure_sha256": "sha256:<hex>"
}
```

- `roles` is a closed manager-owned token set: exactly `swift-toolchain` and
  `platform-sdk`, in that order, on macOS.
- **No member names a filesystem path.** Toolchain location is not portable
  identity (decision 0007 section 3.2).
- Each `tree_sha256` uses the same walk, ordering, record framing and link rules
  as `curator-go-toolchain-v1`, with the domain prefix
  `curator-swift-toolchain-v1/root`.
- `closure_sha256` is domain-separated over the ordered `(role, tree_sha256)`
  pairs with the prefix `curator-swift-toolchain-v1/closure`, so a two-root
  closure can never collide with a one-root closure over the same bytes.
- `curator-go-toolchain-v1` and `curator-rust-toolchain-v1` are untouched. This
  algorithm does not reuse, extend or alias either.

**Measured cost, per operation, per root** (walk plus content hash):

| Role | Regular files | Symlinks | Bytes | Wall clock |
|---|---|---|---|---|
| `swift-toolchain` | 5,109 | 91 | 2.57 GiB | 5.89 s |
| `platform-sdk` | 32,345 | 7,448 | 0.71 GiB | 5.60 s |

The cost is stated rather than optimised away. Memoising across operations is
forbidden: it would defeat the property the fingerprint exists to prove.

---

## 3. Source layout

Identical for both drivers; `swift-repository-v1` applies it to the descriptor's
selected build root.

| Requirement | Rule |
|---|---|
| metadata file | `<build_root>/Package.swift` MUST exist as a regular file and MUST be the nearest ancestor `Package.swift` of `source_dir` |
| metadata read | **exactly** the bytes up to the first `LF`, with one trailing `CR` removed. No byte after that `LF` is read, scanned, or able to change any verdict. The body is never parsed, compiled, executed, or passed to the compiler |
| `source_dir` | MUST equal `build_root` |
| sources directory | `<build_root>/Sources` MUST exist as a directory |
| compiled source set | every regular file under `<build_root>/Sources` whose name ends in `.swift`, recursively, ordered by relative path in Unicode-scalar order; MUST be non-empty |
| other entries under `Sources` | every non-`.swift` regular file, every symlink, device, socket and fifo anywhere in the subtree is a **rejection** |
| source path bytes | every compiled source relative path MUST be free of ASCII control bytes (`0x00`–`0x1F`, `0x7F`). See 3.1 |
| module name | `curator-swift-module-v1` over the consuming command key (section 3.2). No package string reaches an argument vector |

`Package.swift` is excluded from the compiled source set by name.

The mapping is **total** over `Sources`: it selects nothing, so there is no
candidate set, no heuristic and no ordering a package can exploit. That totality
is what satisfies decision 0008 section 4's non-discovering requirement, and it
is what makes the non-`.swift` rejection load-bearing — the compiled byte set is
exactly the audited byte set.

Two programs require two build roots.

### 3.1 Source path bytes

Rejected: any ASCII control byte in a compiled source relative path. Diagnostic
`build_source_layout_invalid`, Stage B, before the graph phase.

The rule is narrow on purpose and is **measured** rather than precautionary. A
source named `new\nline.swift` splits its job across physical lines of the
`swiftc -###` plan: measured, the graph command still exits 0 while the plan
carries 7 physical lines for 3 jobs, and the section 4.1 verifier rejects all
seven. Refusing the name at Stage A is the only place the rejection is
attributable to the snapshot rather than to a parse failure.

Admitted, and each **measured** to round-trip through the plan grammar
unambiguously:

| Name | Rendered in the plan as |
|---|---|
| `has space.swift` | single-quoted |
| `has'quote.swift` | single-quoted, the inner quote as `'\''` |
| `has#hash.swift` | single-quoted |
| `back\slash.swift` | single-quoted, the backslash literal |
| `@resp.swift` | bare, as an absolute path — never expanded as a response file |

### 3.2 `curator-swift-module-v1`

The command key grammar and the Swift module grammar do not overlap:

```
command key    ^[A-Za-z0-9][A-Za-z0-9._-]*$      unbounded; may start with a digit
Swift module   ^[A-Za-z_][A-Za-z0-9_]{0,63}$     bounded; may not start with a digit
```

A replacement rule is unsound: `my-tool`, `my.tool` and `my_tool` would all
become `my_tool`, merging three distinct protocol identities. The derivation is:

**Escape.** Map the key into `[A-Za-z0-9]` with a prefix-free code:

| Input byte | Output |
|---|---|
| `z` | `zz` |
| `.` | `zd` |
| `-` | `zh` |
| `_` | `zu` |
| any other `[A-Za-z0-9]` | itself |
| anything else | reject: the key is outside the protocol grammar |

**Branch.** Let `esc` be the escaped key.

| Condition | Result | Length |
|---|---|---|
| `len(esc) ≤ 61` | `Sk_` + `esc` | ≤ 64 |
| otherwise | `Tk_` + `hex40(SHA-256("curator-swift-module-v1\0" + key))` + `_` + `esc[:20]` | exactly 64 |

Properties, each executable rather than argued:

- **Total.** Every protocol-valid key has a result, including punctuation,
  leading digits, and keys longer than the module bound.
- **Deterministic.** No host, clock or filesystem input.
- **Injective on the short branch, by construction.** The escape is decodable;
  the decoder recovers the exact key.
- **Collision-resistant on the long branch.** A 160-bit digest of the **whole**
  key, not of a truncation.
- **Branch-separated.** `Sk_` and `Tk_` are disjoint prefixes, so a short result
  can never equal a long one.
- **Prefix-reserved.** The result always starts with an uppercase letter, so it
  can never be a Swift keyword and can never be `Swift` — which a bare escape
  would produce for the key `wift`. Measured: of 341 module and framework names
  inventoried from the toolchain root and the platform SDK, **0** carry either
  prefix.

Declared mapping, which the conformance vectors carry verbatim:

| Command key | Module name |
|---|---|
| `tool` | `Sk_tool` |
| `my-tool` | `Sk_myzhtool` |
| `my.tool` | `Sk_myzdtool` |
| `my_tool` | `Sk_myzutool` |
| `9.tool` | `Sk_9zdtool` |
| `0` | `Sk_0` |
| `z` | `Sk_zz` |
| `z-z` | `Sk_zzzhzz` |
| `zdz` | `Sk_zzdzz` |
| `a.b-c_d.e` | `Sk_azdbzhczudzde` |
| 41 bytes escaping to exactly 61 | `Sk_…`, exactly 64 long |
| one byte more | `Tk_…`, exactly 64 long |

A key outside the protocol grammar is rejected with
`build_source_layout_invalid`; the manager never coerces one.

**Identity does not depend on this being injective.** Section 8 binds the
command key itself into the canonical build input alongside the module name.

**External mode only.** A `swift-repository-v1` command requires
`skill-build.json` **schema 2**. Against schema 1 it fails
`build_descriptor_driver_unsupported`; against an unsupported version,
`build_descriptor_schema_unsupported`. Neither falls back to another target,
another driver, a script, a system command or a generic build facility.
`build_root` MAY be `.`. The whole external snapshot is the validation, identity
and audit subject; only the selected build root is compiler-visible.

---

## 4. The two argument vectors

Working directory: the canonical `source_dir`. The manager MUST use exactly
these vectors and MUST NOT alter, extend, reorder or repeat them.

```text
# graph phase
swiftc -### -swift-version 6 -O -target <native-triple> -sdk <presented-sdk> \
       -module-name <module> -no-color-diagnostics <sources…> -o <staged-artifact>

# compile phase
swiftc      -swift-version 6 -O -target <native-triple> -sdk <presented-sdk> \
       -module-name <module> -no-color-diagnostics <sources…> -o <staged-artifact>
```

They differ in exactly one token, at index 0. That is the property the graph
phase rests on: the plan verified is the plan executed. The manager MUST produce
both vectors from **one** builder, so they cannot drift, and the equality is
asserted token-for-token rather than assumed.

| Placeholder | Source |
|---|---|
| `<native-triple>` | Stage A `target.triple` (section 1.3) |
| `<presented-sdk>` | section 2.2 |
| `<module>` | `curator-swift-module-v1` over the consuming command key (section 3.2) |
| `<sources…>` | the ordered compiled source set (section 3) |
| `<staged-artifact>` | operation-private manager staging path |

`<staged-artifact>` MUST be stable for a given operation, because the output
path reaches the Mach-O `LC_UUID` (section 10).

The planned **job** argv is a different thing and is **not** reproducible:
measured, it carries a per-run `TemporaryDirectory.XXXXXX` the driver creates
under the operation-private `TMPDIR`. Section 4.1 is therefore a
bucket-and-boundary check over the plan, never a comparison against a fixed
expected plan.

### 4.1 Graph-phase plan verification

`swiftc -###` prints a job plan and executes nothing. Measured: exit **0** for a
source file containing `this is not swift @@@ (((`, exit **0** for a source path
that does not exist, and the source directory unchanged in both cases. The plan
is written to **stdout**; stderr is empty.

This section is a rejection engine. There is no skip, no ignore and no
best-effort recovery: anything the grammar does not account for fails
`build_execution_control_unavailable` before the compile phase.

#### 4.1.1 The output grammar

Measured on Apple Swift 6.3.2, macOS 26.5 arm64. The plan is POSIX
single-quote-quoted argv, one job per line.

```
plan   := line ( LF line )* LF?
line   := token ( SP+ token )*                  MUST be non-empty
token  := ( bare | quoted | escaped )+
bare   := any byte except SP, ', \, and every ASCII control byte
quoted := "'" ( any byte except ' )* "'"
escaped:= "\" any byte except an ASCII control byte
```

Measured token renderings: a value containing a space or a `#` is wrapped in
single quotes; an embedded single quote is rendered `'\''` — close the quoted
run, backslash-escape the quote outside it, reopen; a backslash inside a quoted
run is literal.

The tokenizer MUST fail — not return a shorter token list — when a line ends
inside a quoted run, ends with a dangling backslash, or carries any ASCII
control byte. A shorter token list is exactly how a malformed plan reads as a
clean one.

Line-level rules, all mandatory:

1. the graph command MUST have exited 0 and the plan MUST be non-empty;
2. splitting on `LF` yields the lines, with **exactly one** trailing empty
   element dropped for the final newline; any remaining empty or whitespace-only
   line is a rejection;
3. every line MUST tokenize, and its first token MUST be an absolute path.

#### 4.1.2 The five buckets, and totality

Every **path-shaped** token — one beginning with `/`, or with a drive-absolute
or UNC form on Windows — MUST be claimed by exactly one bucket and satisfy its
rule. A path-shaped token no bucket claims is a rejection.

| Bucket | Claimed by | Rule |
|---|---|---|
| executable | a line's first token; the value of `-new-driver-path` | MUST exist and resolve to a regular executable **inside the `swift-toolchain` root** |
| plugin | the value of `-plugin-path`, `-external-plugin-path`, `-in-process-plugin-server-path`, `-load-plugin-library`, `-load-plugin-executable`, `-load-resolved-plugin` | MUST resolve **inside a fingerprinted root**, **or** MUST NOT exist |
| search | the value of `-sdk`, `-isysroot`, `--sysroot`, `-resource-dir`, `-I`, `-F`, `-L`, and the joined spellings `-I<path>`, `-L<path>`, `-F<path>` | MUST exist and resolve **inside a fingerprinted root** |
| source | the value of `-primary-file`; a positional token equal to a source-set member | MUST equal a member of the manager's own ordered compiled source set, **byte for byte** |
| output | the value of `-o`; any other positional path | MUST resolve, or have a parent that resolves, **inside operation-private manager state** (staging or the operation `TMPDIR`) |

Flag/value grammar:

- a flag admits both `-flag value` and `-flag=value`; a flag with **no**
  following token, or with an empty value, is a rejection;
- `-external-plugin-path` carries exactly `<dir>#<server>`: exactly one `#`,
  both components non-empty and absolute. Any other shape is a rejection;
- every **other** plugin flag carries exactly one path; a `#` in its value is a
  rejection, because that flag defines no separator;
- a non-path-shaped token that is not a flag and names an existing entry
  relative to the child working directory is a rejection — that is the residual
  relative-path channel, closed explicitly.

**Stated limit.** Totality is over tokens the grammar recognises as paths. A
future toolchain that introduced a path channel in a shape this grammar does not
recognise would be caught only if the resulting token were path-shaped. Because
every unclaimed path-shaped token rejects, the failure direction is closed, and
extending the flag tables is a measured contract change rather than a silent
admission.

#### 4.1.3 Containment

Containment is computed on the **symlink-resolved** path of both the candidate
and the root, and compared byte-exactly on path components:

```
contained(p, r)  :=  resolve(p) == resolve(r)  OR  resolve(p) starts with resolve(r) + separator
```

- `/root-evil/bin/ld` is not inside `/root` — a shared prefix is not containment;
- a path lexically below a root but symlinked out of it is **rejected**;
- the root itself is inside itself, because `--sysroot` names it directly;
- on a case-insensitive volume a case variant does not match and therefore fails
  closed;
- a **dangling symlink** is an entry that exists and resolves nowhere. It is
  neither inside a root nor absent, so it is a rejection.

#### 4.1.4 Binding and the compile permit

The graph phase and the compile phase are two moments. The manager MUST record,
for every verified path, its resolved path and its file identity, and for every
plugin path it verified as **absent**, that absence. Immediately before the
compile child starts it MUST re-check every binding and reject unless:

- every bound path still resolves to the identical resolved path;
- every bound path still has the identical file identity; and
- every absent plugin path is still absent.

This **narrows** the window; it does not remove it. What closes it is the
ownership requirement section 2.1 already imposes: both fingerprinted roots are
manager-distribution or owner-protected manager-owned state, and `<staging>` is
operation-private, so no other principal is an admitted writer. The re-check is
the defence that does not depend on that assumption holding.

#### 4.1.5 Measured results

With the section 2.2 presentation, over the default two-source vector:

| Quantity | Measured |
|---|---|
| planned jobs | 3 — two `swift-frontend`, one `clang` |
| executable bucket | 5 bindings (3 job executables, 2 `-new-driver-path`), 0 outside the root |
| plugin bucket | 14 components; 6 distinct; 2 existing, both inside the toolchain root; 4 absent; **0 existing outside a fingerprinted root** |
| search bucket | 5 bindings, all inside a fingerprinted root |
| source bucket | 4 bindings, each byte-equal to a manager source-set member |
| output bucket | 5 bindings, all inside operation-private state |
| rejections | **0** |

Negative coverage, one vector per failure family, all **rejected**: relative
executable; unknown wrapper line; executable outside the root; executable that
does not exist; unmatched quote; dangling backslash; plugin flag with no value;
plugin flag with an empty value; `-external-plugin-path` with one component;
with three components; `#` in a single-path plugin flag; plugin path existing
outside every root; search path outside every root; joined search path outside
every root; search path that does not exist; source not in the manager's set;
output outside operation-private state; unclaimed absolute positional; blank
line; empty plan; ASCII control byte; `-new-driver-path` outside the root.
Measured: **20 of 20 rejected**, while the plan the toolchain actually emits
verifies with 0 rejections. The retired lenient scan admits **16 of the 20**.

The linker never appears in this plan: `swiftc` starts `clang`, and `clang`
starts the linker. Its resolution is checked by probe vector P3, which measured
`<root>/usr/bin/ld` and confirmed resolved containment.

Failure: `build_execution_control_unavailable`, before the compile phase.

---

## 5. Operation-private environment

The environment starts empty except for indispensable operating-system process
variables, and carries exactly:

| Variable | Value |
|---|---|
| `PATH` | a manager-owned **empty** directory |
| `HOME` | operation-private |
| `TMPDIR` | operation-private, MUST exist before the first child starts |
| `LC_ALL`, `LANG` | `C` |
| Windows: `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `TEMP`, `TMP` | operation-private |

Nothing else. In particular the following MUST be absent, not merely overridden:
`DEVELOPER_DIR`, `TOOLCHAINS`, `SDKROOT`, `SWIFT_EXEC`, `SWIFT_DRIVER_*`,
`SWIFTPM_*`, `SWIFT_BACKTRACE`, `MACOSX_DEPLOYMENT_TARGET`, `CPATH`,
`C_INCLUDE_PATH`, `LIBRARY_PATH`, `LD_LIBRARY_PATH`, `DYLD_*`, `CC`, `CXX`,
`LD`, `LDFLAGS`, `CFLAGS`, `NSUnbufferedIO`, every proxy variable, and every
version-manager variable.

There is no manager-written tool configuration file for Swift. There is no
SwiftPM in the pipeline, so there is no `$SWIFTPM_HOME`, no registry
configuration, no mirror file and no netrc to write or neutralise.

A `TMPDIR` that does not exist is a hard failure, not a warning: measured, the
driver aborts with `error: couldNotFindTmpDir(…)` and produces nothing.

---

## 6. Pre-compile rejection matrix

Total. Every surface has exactly one verdict, decided before the compile phase.
The shared semantic class is `build_package_code_execution_forbidden` unless a
row names another.

### 6.1 Properties

- **No package code runs.** Every row is decided from snapshot bytes the manager
  reads itself, or from the `-###` plan, which was measured to execute nothing.
- **Host-independent.** The only host-derived inputs are the resolved triple and
  the presented SDK path, both manager-selected.
- **Three verdicts, kept apart.** `reject` fails the operation with a named
  diagnostic. `bound` means the manager reads the bytes deliberately and
  completely. `inert` means the bytes are in the audit subject and the source
  identity, are never opened by the manager, never reach the compiler, and are
  never executed, because no channel in the pipeline names them.

The build-root subtree is partitioned totally:

| Region | Rule |
|---|---|
| inside `Sources` | `.swift` regular files and directories only; everything else rejected (6.2) |
| `<build_root>/Package.swift` | **bound**: first line only (section 3, section 7) |
| inside the build root, matching a rejected name or extension | rejected (6.2) |
| everything else inside the build root | **inert** (6.6) |

### 6.2 Rows decided by snapshot bytes

| Surface | Verdict | Diagnostic |
|---|---|---|
| `Package.resolved` anywhere in the build-root subtree | reject | `build_package_dependency_declaration_forbidden` |
| `Package@swift-*.swift` anywhere in the subtree | reject | `build_package_alternate_manifest_forbidden` |
| `.swiftpm`, `Plugins`, `Snippets` directory anywhere in the subtree | reject | `build_package_plugin_forbidden` |
| native or foreign input anywhere in the build-root subtree, by closed extension list: `.o .a .dylib .so .bundle .framework .c .cc .cpp .cxx .m .mm .h .hpp .modulemap .swiftinterface .swiftmodule .tbd .obj .lib .dll` | reject | `build_package_native_input_forbidden` |
| any non-`.swift` regular file under `Sources` — including a script, a `Makefile` or an executable-bit file | reject | `build_package_native_input_forbidden` |
| any symlink, device, socket or fifo in the build-root subtree | reject | `build_package_unsupported_entry_kind` |
| a compiled source relative path carrying an ASCII control byte | reject | `build_source_layout_invalid` |
| a command key outside `^[A-Za-z0-9][A-Za-z0-9._-]*$` | reject | `build_source_layout_invalid` |
| empty compiled source set | reject | `build_source_layout_invalid` |
| missing `Package.swift`, missing `Sources`, `source_dir` ≠ `build_root` | reject | `build_source_layout_invalid` |

Outside `Sources` the native-input list would be inert bytes. It is rejected
anyway, for the same reason as `Package.resolved`: its presence declares an
intent the driver cannot honour, and naming the mismatch at the boundary is
better than building something the author did not describe.

### 6.3 Rows decided by the graph-phase plan

| Surface | Verdict | Diagnostic |
|---|---|---|
| a plan line the section 4.1.1 grammar does not cover — unmatched quote, dangling backslash, control byte, blank line, non-absolute first token | reject | `build_execution_control_unavailable` |
| a flag with a missing or empty value; an `-external-plugin-path` value that is not `<dir>#<server>`; a `#` in a single-path plugin flag | reject | `build_execution_control_unavailable` |
| a job executable or `-new-driver-path` value outside the `swift-toolchain` root | reject | `build_execution_control_unavailable` |
| a plugin path that exists outside every fingerprinted root, or a dangling symlink | reject | `build_execution_control_unavailable` |
| a search path outside every fingerprinted root, or absent | reject | `build_execution_control_unavailable` |
| an output path outside operation-private manager state | reject | `build_execution_control_unavailable` |
| a source token that is not byte-equal to a manager source-set member | reject | `build_execution_control_unavailable` |
| any other path-shaped token no bucket claims | reject | `build_execution_control_unavailable` |
| a binding whose resolution or identity changed between graph and permit; an absent plugin path that appeared | reject | `build_execution_control_unavailable` |
| an empty plan, or a non-zero graph command | reject | `build_execution_control_unavailable` |

### 6.4 Rows decided by the fixed vectors and environment

| Surface | Verdict | Why it cannot occur |
|---|---|---|
| any SwiftPM invocation | reject | the two vectors of section 4 are the whole command set |
| `Package.swift` body — targets, products, dependencies, `unsafeFlags`, `swiftSettings`, `linkerSettings`, `cSettings`, plugins, macro targets, binary targets, system-library targets, prebuild and postbuild commands, manifest `#if` | reject as an input | the manager's read stops at the first `LF` (section 3); the file is excluded from the source set; and neither command shape has a flag member |
| a package-supplied compiler or linker flag, by any channel | reject | neither command shape has a flag member |
| a response file — an `@`-leading argument | reject as unreachable | measured that `swiftc` honours `@file`, and measured that a source named `@resp.swift` reaches the compiler as an absolute path and is compiled. No vector member begins with `@` |
| a build configuration selector — debug/release, `-Onone`, `.xcconfig`, `.xcodeproj`, a scheme | reject as unreachable | the compile vector fixes `-O`; no configuration member exists; none of those files is compiler-visible |
| a script outside `Sources` — `.sh`, `Makefile`, a hook, an executable-bit file | inert | no channel names it: the process graph is fixed, `PATH` is an empty directory, and section 4.1 rejects any executable it does not already account for |
| `-import-objc-header`, `-Xcc`, `-Xfrontend`, `-Xllvm`, `-Xlinker`, `-I`, `-L`, `-l` | reject | absent from both vectors |
| network access, dependency resolution, registry access, git fetch | reject | no network-capable command exists in the pipeline; `PATH` is an empty directory |
| package-selected toolchain path, root, channel, mirror, installer, version manager | reject | decision 0007 resolution; `.swift-version` is classified, not honoured |
| cross-compilation, non-native `-target` | reject | the compile vector fixes `-target` |

### 6.5 Admitted surfaces

| Surface | Bound |
|---|---|
| macros whose implementation is inside a fingerprinted root (`@Observable`, and the toolchain's own) | section 4.1 verifies the confinement from the executed plan |
| `#if` conditional compilation | selects which admitted source compiles; equivalent to a Go build constraint |
| `import` of a module the presented SDK exposes | the SDK is a fingerprinted data root |
| `@_cdecl`, `@_silgen_name` | section 11 |

A macro whose implementation lives outside every fingerprinted root is not
"rejected" by a byte scan — it simply cannot load. Measured: `#Predicate` under
the section 2.2 presentation fails with
`error: external macro implementation type 'FoundationMacros.PredicateMacro' could not be found`.
That is a compile error attributable to package source, not a manager
diagnostic.

### 6.6 Inert bytes

Files inside the build root, outside `Sources`, that no row above rejects, and
that are not `Package.swift`: `README.md`, `LICENSE`, `.gitignore`, a resources
directory, `Tests`, a `.editorconfig`, a CI configuration.

They are **inert**, which is a precise statement rather than a shrug:

- they are inside the audit subject and inside `curator-build-source-v1`, so
  they are identified and reviewable;
- the manager never opens them — the only files it reads are the compiled source
  set and the first line of `Package.swift`;
- they never reach the compiler — the compiler-visible set is exactly the
  ordered source set, and section 4.1's source bucket rejects any token that is
  not byte-equal to one of its members;
- they are never executed — the process graph is fixed (section 2.1), `PATH` is
  an empty directory (section 5), and section 4.1 rejects any executable the plan
  names that is not inside the toolchain root.

---

## 7. Stage B — metadata dispositions

| Source | Field | Disposition |
|---|---|---|
| `Package.swift` | first-line `swift-tools-version` | `classified` (7.2) |
| `Package.swift` | every byte after the first `LF` | not a metadata source and not read at all; rejected as an input by section 6 |
| `.swift-version` | the bare version string | `compared`, decision 0007 channel table |

Evaluation order is Unicode-scalar lexical order of relative source path, so
`.swift-version` precedes `Package.swift`. A snapshot carrying a section 6.2
surface is deterministically a package-influence rejection before any comparison
runs.

`.swift-version` is `compared` rather than `forbidden` because it names a
version, not an origin, and because it is inert against the direct resolution
decision 0007 mandates. Measured: a `.swift-version` of `5.9.9-nonexistent`
beside the sources changed nothing; the compile exited 0.

### 7.1 Curator reads the header itself, and reads only line 1

The manager reads `Package.swift`'s first line as bytes. It MUST NOT invoke
`swift package tools-version` or any other SwiftPM subcommand to do so. That
command appears in this document only as the upstream oracle the classifier is
measured against.

**First line = bytes up to the first `LF`, with one trailing `CR` removed. That
is the classifier's entire input.** No byte after that `LF` is read, scanned, or
able to change the verdict.

Upstream is different, and the difference is measured, not assumed:
`swift package tools-version` reports `9.9.0` with exit 0 for a manifest whose
only specification sits inside a **multi-line string literal** on line 3. A
whole-file scan would therefore let arbitrary manifest body bytes — including
bytes inside a string constant — set the version the manager compares, which is
the exact input this driver exists to exclude. The narrowing is declared as
section 7.3's security partition rather than hidden.

Two nearby cases are **not** narrowings, and both are measured. A canonical line
1 followed by a second specification on line 2 yields `6.0.0` from upstream —
line 1 decides for both. A specification inside a single-line string literal
(`let s = "// swift-tools-version:9.9"`) is found by neither, because the line
does not begin with the comment marker.

### 7.2 Classifier — `swift-tools-version`

Ordered, exhaustive, with a mandatory catch-all. Every rule reads line 1 and
nothing else.

| # | Class | Rule |
|---|---|---|
| 1 | `rejected-absent-header` | **line 1** carries no tools-version specification in any case or spacing form. Whether a later line does is not consulted |
| 2 | `rejected-non-canonical-header` | line 1 carries a specification, but not in the canonical form (see below) |
| 3 | `rejected-missing-specifier` | canonical prefix present, version text empty after trimming |
| 4 | `rejected-grammar` | version text present, fails the grammar, and upstream does not represent it either |
| 5 | `rejected-unsupported-floor` | version < `4.0.0` |
| 6 | `host-gate` | version > the resolved normalized compiler version |
| 7 | `accepted` | otherwise |

Class 1 is the honest name for what happened: *line 1 carried none*. It does not
claim the file carries none, and the manager never learns whether it does.

**Canonical form.** Line 1 matches, whole line:

```
^// swift-tools-version: ?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(\.(0|[1-9][0-9]*))?$
```

Exactly two slashes, one ASCII space, the lowercase keyword, a colon, at most
one space, a two- or three-component decimal version with no leading zeros, no
prerelease, no build metadata, nothing after it.

**Grammar** for the version text: two or three decimal components, no component
with a leading zero unless it is `0`, no prerelease, no build metadata.

**Floor**: `4.0.0`. Measured, not assumed: `3.1` is refused by the corroborating
command with `… is using Swift tools version 3.1.0 which is no longer supported…`
and `4.0` is accepted.

### 7.3 The security partition `F`

`F` is the set of forms upstream represents and Curator refuses. It has two
shapes.

**Shape A — line 1 carries the specification, and upstream reinterprets it.**
The compared version differs from the bytes the author wrote.

| Header | upstream normalizes to | Curator |
|---|---|---|
| `//swift-tools-version:6.0` | `6.0.0` | class 2 |
| `// SWIFT-TOOLS-VERSION:6.0` | `6.0.0` | class 2 |
| `//   swift-tools-version:  6.0` | `6.0.0` | class 2 |
| `// swift-tools-version:06.0` | `6.0.0` | class 2 |
| `// swift-tools-version:6.0-beta` | `6.0.0` | class 2 |
| `// swift-tools-version:6.0+build` | `6.0.0` | class 2 |

**Shape B — the specification is below line 1, so Curator never reads it.**

| Manifest | upstream normalizes to | Curator |
|---|---|---|
| `import Foundation` then the header on line 2 | `6.0.0` | class 1 |
| the header only inside a multi-line string literal | `9.9.0` | class 1 |

These eight are `F`. **No member of `F` may classify as `rejected-grammar`**:
calling it a grammar error would assert that upstream refuses it too. Shape A
members classify as `rejected-non-canonical-header`; shape B members classify as
`rejected-absent-header`, which is what Curator actually determined.

**Alignment properties.** P1 (no widening) is asserted over **all** cases:
Curator accepts nothing upstream refuses. P2 (no narrowing) is asserted over
cases **outside** `F`. Measured: both hold, `F` non-empty and enumerated.

A byte-order mark before the header is **not** in `F`: measured, upstream's own
recognition fails on it too and falls back to `3.1.0`, so both refuse it. A
second specification on line 2 below a canonical line 1 is **not** in `F`:
measured, upstream also takes line 1. A specification inside a single-line string
literal is **not** in `F`: measured, neither finds it.

### 7.4 Recognised command outcomes are a closed set

Used only by the conformance probe, which measures Curator's classifier against
upstream. Every line is predicted **before** the command runs, from the value
under test plus constants the probe fixes from the resolved toolchain:
`<full>` = normalized compiler version (`6.3.2`), `<mm>` = its major-minor form
(`6.3`), `<pkg>` = the package directory name, `<raw>` = the version text as
written, `<norm>` = its upstream normalization.

Isolated command — `swift package tools-version`:

| Whole line / condition | Stream | Class |
|---|---|---|
| `<norm>` with exit 0 | stdout | `accepted` |
| `error: the Swift tools version '<raw>' is misspelt or otherwise invalid; consider replacing it with '// swift-tools-version: <full>' to specify the current Swift toolchain version as the lowest Swift version supported by the project` | stderr | `rejected-grammar` |
| `error: the Swift tools version specification is possibly missing a version specifier; consider using '// swift-tools-version: <full>' to specify the current Swift toolchain version as the lowest Swift version supported by the project` | stderr | `rejected-missing-specifier` |
| `error: package 'package.swift' is using Swift tools version 3.1.0 which is no longer supported; consider using '// swift-tools-version: <mm>' to specify the current tools version` | stderr | `rejected-absent-header` |

Corroborating command — `swift build`. Every diagnostic carries a `'<pkg>': `
infix the isolated forms do not, which is why the two sets cannot share one
predictor:

| Whole line / condition | Class |
|---|---|
| exit 0 | `accepted` |
| `error: '<pkg>': package '<pkg>' is using Swift tools version <norm> but the installed version is <full>` | `host-gate` |
| `error: '<pkg>': package '<pkg>' is using Swift tools version <norm> which is no longer supported; consider using '// swift-tools-version: <mm>' to specify the current tools version` | `rejected-unsupported-floor` |
| `error: '<pkg>': the Swift tools version '<raw>' is misspelt or otherwise invalid; …` (as above, with the infix) | `rejected-grammar` |
| `error: '<pkg>': the Swift tools version specification is possibly missing a version specifier; …` (as above, with the infix) | `rejected-missing-specifier` |
| `error: '<pkg>': package 'package.swift' is using Swift tools version 3.1.0 which is no longer supported; …` | `rejected-absent-header` |

Rules:

- recognition is **whole trimmed line equality**; a lead with an unconstrained
  tail and a substring found anywhere are families, not outcomes, and MUST NOT
  be recognised;
- two expected lines of **different** classes matching inside one output is
  `unknown`, not first-wins;
- anything outside the set is `unknown`, yields no verdict, and fails the probe.

**Isolated vs corroborating.** `swift package tools-version` cannot be applying
the host gate: measured, it reports `99.0.0` with exit **0** on a 6.3.2 host
while `swift build` on the same package exits 1 with the host-gate line. The
corroborating outcome is required to be *reachable* from the isolated one — an
`accepted` isolated outcome may become `accepted`, `rejected-unsupported-floor`
or `host-gate` — never equal to it.

### 7.5 Closure is measured, not asserted

Three kinds, both laundering directions reported (A = a fabrication agreeing
with an isolated-accepted value, B = with an isolated-rejected one):

| Kind | Count | What it is |
|---|---|---|
| measured, unrelated | 4 | a real command, a real non-zero exit, nothing about the value |
| measured | 20 outcomes over 338 pairs | every value-bearing outcome classified under every other case's value |
| measured, extended (**constructed**) | 27 | a measured diagnostic with a tail appended, a wrapper in front, or embedded in a longer line |

The third kind is constructed and is labelled so in the fixture: a fail-closed
property is a claim about outcomes no host has emitted yet.

**Exclusions are printed with their reason.** 20 outcomes are excluded from the
cross-feed because they name no value under test: an exit-0 acceptance, a
missing-specifier diagnostic and an absent-header diagnostic. Feeding those
under another value would test the classifier against text that was never about
a value. The exclusion is by *measured property* — the recognised line does not
contain the value — not by an allowlist of case names.

Measured result: 32 emitted rows, **0** yielding a verdict.

### 7.6 Controls required to fail

Each restores a retired defect from the same binary and MUST exit non-zero.

| Control | Guards | Measured findings |
|---|---|---|
| `C1` lead-only recognition | an outcome is a whole line, not a lead | 3 |
| `C2` substring recognition | a diagnostic embedded in a longer line is not that diagnostic | 1 |
| `C3` exit status as semantics | the floor, the host gate and a grammar rejection are not one exit status | 7 |
| `C4` `swift build` as the isolated command | representability is measured by a command that cannot apply the host gate | 3 |
| `C5` unearned `PATH`-closure zero | a zero without a firing control is not evidence | 1 |
| `C6` plugin closure unchecked | the macro plugin surface is inside the fingerprinted closure | 2 |
| `C7` lenient plan parsing | a plan shape the grammar does not cover is a rejection, not a skip | 16 |
| `C8` lexical containment | containment is resolved, not a string prefix | 1 |
| `C9` collapsing module derivation | two distinct command keys never share one module name | 3 |

`C5` through `C9` are the Swift-specific ones. `C5` reports a zero obtained with
the shim directory off `PATH`. `C6` uses the declared SDK path and reports the 2
distinct live plugin paths it derives outside every fingerprinted root. `C7`
restores the lenient scan and reports that it admits 16 of the 20 malformed
plans of section 4.1.5. `C8` restores lexical prefix containment and reports the
symlink escape it admits. `C9` restores the replacement module rule and reports
that `my-tool`/`my.tool`/`my_tool`, `a.b-c`/`a-b.c` and `9.tool`/`9-tool` each
collapse to one name.

---

## 8. Identity, cache, receipt, marker, claim

The canonical build input binds, in addition to the members decision 0008
section 8 requires of every new driver:

| Member | Value |
|---|---|
| `toolchain_identity` | the whole `curator-swift-toolchain-v1` object of section 2.3, including both root digests and the closure digest |
| `native_target` | `target.unversionedTriple` |
| `source_set` | the ordered relative paths of the compiled source set |
| `command_key` | the consuming manifest command key, verbatim |
| `module_name` | the `curator-swift-module-v1` derivation of section 3.2 |
| `policy` | the closed object below |

`command_key` and `module_name` are both bound, and that is not redundancy. The
module name is what the compiler was given, so it belongs in the input that
identifies the build; the command key is what the protocol keeps distinct.
Binding both means two commands cannot share a cache identity even under a
hypothetical module-name collision, so identity does not depend on the section
3.2 derivation being injective.

```json
{
  "package_manager": "none",
  "dependency_mode": "source-inlined",
  "network": "none",
  "manifest_execution": false,
  "plugins": false,
  "macros": "toolchain-only",
  "binary_targets": false,
  "system_library_targets": false,
  "target_mode": "native",
  "optimization": "release",
  "linker": "toolchain-ld",
  "link_mode": "internal",
  "native_inputs": false,
  "sdk_presentation": "manager-owned-symlink-root",
  "plugin_closure_check": "job-plan-verified-v1",
  "compiler_directives": "reject-nonstandard-native-inputs-v1",
  "incremental": false,
  "execution_policy": "manager-worker-v2"
}
```

`execution_policy` is the `const` `manager-worker-v2`. `network: "none"` denotes
the absence of any network-capable command together with the empty `PATH`; it is
**not** a claim of kernel-enforced network denial, which remains the deferred
`total-network-denial` guarantee.

| Artifact | Rule |
|---|---|
| receipt schema 3 | local mode; strict `oneOf` on the `driver` `const`; carries the policy object, the toolchain identity and the native target |
| receipt schema 4 | external mode; same shape, plus the external source identity |
| marker schema 4 | records `driver`, `receipt_schema_version` and `execution_policy` per build entry; a reader rejects a `swift-v1` entry claiming `manager-worker-v1` rather than inferring the policy from the driver name |
| claim schema 4 | asserts `swift-v1` / `swift-repository-v1` with `execution_policy` bound by the assertion's own `driver` `const` |

The effective toolchain requirement and the `compatibility` set stay gates, never
build inputs.

---

## 9. Diagnostics

| Code | Stage | Trigger |
|---|---|---|
| `build_toolchain_root_undeclared` | A | no declaration for `swift-toolchain` or `platform-sdk` |
| `build_toolchain_root_unusable` | A | a required closure member (section 2.1) missing, non-executable, or resolving outside the root |
| `build_toolchain_version_undetermined` | A | `compilerVersion` does not match the section 1.2 rule |
| `build_toolchain_platform_unsupported` | A | host pair not in `platforms`, or a `runtimeLibraryPaths` entry inside the root absent |
| `build_toolchain_metadata_mismatch` | B | `swift-tools-version` classifier classes 2–6, or a `.swift-version` comparison mismatch |
| `build_package_code_execution_forbidden` | B | the shared class for section 6 rejections |
| `build_package_dependency_declaration_forbidden` | B | `Package.resolved` present |
| `build_package_alternate_manifest_forbidden` | B | `Package@swift-*.swift` present |
| `build_package_plugin_forbidden` | B | `.swiftpm`, `Plugins` or `Snippets` present |
| `build_package_native_input_forbidden` | B | a native/foreign input, or a non-`.swift` regular file under `Sources` |
| `build_package_unsupported_entry_kind` | B | symlink, device, socket or fifo in the subtree |
| `build_source_layout_invalid` | B | layout requirement of section 3 unmet, a control byte in a source path (3.1), or a command key outside the protocol grammar (3.2) |
| `build_execution_control_unavailable` | B | plan verification failed (4.1), a binding changed between graph and permit (4.1.4), or the SDK presentation could not be established (2.2) |
| `build_descriptor_driver_unsupported` | B | schema-1 descriptor for an external Swift command |
| `build_descriptor_schema_unsupported` | B | unknown descriptor schema version |
| `build_artifact_class_unsupported` | B | the platform cannot produce a single self-contained executable |

Each MUST remain distinguishable from the others, from a cache hit, from an
audit success, from source unavailability, and from a generic fallback.

---

## 10. Artifact

| Property | Measured |
|---|---|
| shape | `Mach-O 64-bit executable arm64` |
| dynamic dependencies | `/usr/lib/libSystem.B.dylib`, `/usr/lib/libobjc.A.dylib`, `/usr/lib/swift/*`, `/System/Library/Frameworks/Foundation.framework/…` — all base-installation |
| signature | `adhoc, linker-signed`, applied by `ld` during linking |
| runs | yes |

The signature is compiler output, not a manager signing step: it is produced by
the fixed vector, selects no identity, credential or notarization, and reaches
no network. The manager performs **no** post-build signing. A platform policy
requiring a locally signed binary waits for the separately reviewed signer
profile.

Published as `bin/<command>` or `bin/<command>.exe`, derived solely from the
consuming manifest command key.

**Reproducibility, stated precisely.** Measured: two compiles of the same sources
to the **same** output path are byte-identical; changing only the output path
changes the bytes, because the path reaches the Mach-O `LC_UUID`. Identity is
input-keyed, as decision 0008 section 3 requires. This is **not** a
reproducible-build claim, and a manager whose staging path varied per operation
would produce different bytes for the same inputs.

Every other compiler product — object files, `.swiftmodule`, `.swiftdoc`,
`.swiftsourceinfo`, `.dSYM`, dependency files — stays in operation-private
staging and is discarded with it. Measured: the compile phase writes nothing
into the source directory.

The manager MUST NOT execute the artifact, for validation, version discovery,
smoke testing, post-processing, receipt generation, rollback, or any other
reason.

---

## 11. Residual exposures

- **Toolchain macros run inside the compiler.** Source syntax selects which
  implementations the frontend loads; section 4.1 confines those to fingerprinted
  roots, verified from the executed plan. That bounds *where the code comes
  from*, not what macro expansion can do.
- **Compile-time filesystem reads are bounded, not proven absent.** The admitted
  language has no `include_str!` analogue. The surfaces checked and excluded are
  `-import-objc-header`, `-Xcc -include`, module maps, bridging headers and
  `.swiftinterface` inputs. A macro inside the fingerprinted toolchain can still
  read files during expansion; `STORY-260728-327soo` receives that as an input,
  since none of the six deferred hardened guarantees covers compile-time reads.
- **Foreign symbol declarations.** `@_silgen_name` and `@_cdecl` can only name
  symbols the pinned link environment already resolves, which are
  base-installation libraries the artifact class already admits. Not a claim
  that the produced program is safe.
- **Ordinary compiler-input exposure.** Resource-consumption denial of service
  and compiler vulnerabilities reached by adversarial source, bounded by the
  parent-enforced deadline, output and artifact limits, and whichever
  native-control inventory entries the host provides. The six deferred hardened
  guarantees are not claimed, named as controls, or implied.

---

## 12. Platform matrix and qualification

| Platform | Status |
|---|---|
| macOS arm64 | measured on one host; enters a claim only via `TASK-260728-2bu2q6` |
| macOS amd64 | qualification obligation |
| Windows | implementation contract only, **no** claim |
| Linux | excluded until `TASK-260728-1y8u4m` |

### 12.1 The acceptance test

Identical on every candidate platform. All six MUST hold:

1. **Poisoned `PATH`, pinned run**: the compile vector exits 0 with **zero**
   resolutions against a shim directory covering every plausible tool name.
2. **Firing control**: the same harness, with a linker named rather than
   defaulted, produces at least one resolution. *Without this the zero is
   unearned and MUST NOT be accepted.*
3. **Closure members**: the structural rule of section 2.1 holds — every
   executable the verified plan names, plus the linker `clang -print-prog-name`
   resolves, resolves inside the `swift-toolchain` root, following symlinks. The
   member set is read off the plan on that host and recorded; it is not asserted
   in advance. `swift` is not a member.
4. **Plan verification**: section 4.1 passes on that host — the grammar accepts
   the emitted plan with 0 rejections, every path-shaped token is claimed by a
   bucket, and the negative vectors of 4.1.5 all reject.
5. **Artifact**: the produced executable's dynamic dependencies are all
   base-installation libraries of the declared platform baseline, and it runs.
6. **Version rule**: `compilerVersion` matches the section 1.2 normalization
   rule, or the rule is extended by measurement first.

### 12.2 Windows implementation contract

No claim is made. An implementation MUST satisfy 12.1 with a Windows shim set
covering at least `link.exe`, `lld-link.exe`, `cl.exe`, `clang.exe`, `ld.exe`,
`swift-plugin-server.exe`, `where.exe` and `vswhere.exe`; MUST bind whatever the
Windows toolchain needs to link against the base installation as one or more
data-only `platform-sdk` roots through the two declaration channels, presented
per section 2.2; and MUST NOT resolve `link.exe`, `cl.exe`, `vswhere.exe` or a
Visual Studio activation script from `PATH`, the registry or an environment
variable, answer the gap with a host-resolved tool or a downgraded control, or
record a platform claim from a cross-compiled or emulated run.

**The Windows closure member count is not asserted here.** The expected shape is
`usr\bin\swiftc.exe`, `swift-frontend.exe`, `clang.exe` and the resolved linker,
matching the macOS four — but the plan on that host is unmeasured, and Linux was
measured to add a fifth member, so the count is evidently platform-determined.
An implementation MUST take the member set from the plan it verifies under 12.1
step 3, and MUST record it, rather than reading a count out of this paragraph.
`swift.exe` is not a member on any platform, for the section 2.1 reason: the
driver never invokes it.

Reason the claim is withheld: `TASK-260729-rhjxtx` measured **no Swift toolchain
on the reachable Windows host** (19 cases `not_run`), so the `PATH` property,
the linker resolution and the plugin closure are all unmeasured there.

### 12.3 Linux qualification rules

12.1 plus two Linux-specific questions this host cannot answer:

1. the linking vector was measured to require `swift-autolink-extract`, a
   **fifth** executable beyond the macOS four, which must be shown to resolve
   inside the fingerprinted root and to appear in the verified plan;
2. the open-source toolchain banner is not admitted by the section 1.2 rule, so
   a Linux qualification must measure that form and extend the rule, or be
   rejected. It must not be extended speculatively.

---

## 13. Conformance vector inventory

Owned by `TASK-260728-251p01` at admission time; enumerated here so the shape is
fixed before a vector exists.

| Group | Positive | Negative |
|---|---|---|
| manifest schema 8, local command | 2 | 8 — `swift-v1` while reserved; unknown driver; `source_dir` ≠ `build_root`; extra command member; `build_root` of `.`; missing `build_roots` entry; duplicate command key; `manager-worker-v1` claimed |
| descriptor schema 2, external target | 2 | 6 — `swift-repository-v1` against schema 1; unknown schema version; extra target member; `source_dir` ≠ `build_root`; unknown driver; missing target |
| receipt schema 3 / 4 | 4 | 12 — wrong `execution_policy`; missing policy member; extra policy member; missing root in the identity; a path-bearing identity member; wrong role token; role order swapped; missing `closure_sha256`; versioned triple as `native_target`; source set out of order; `module_name` not the 3.2 derivation of `command_key`; `command_key` absent |
| marker schema 4 | 2 | 4 — `swift-v1` with `manager-worker-v1`; missing `receipt_schema_version`; unknown driver; policy inferred from the driver name |
| claim schema 4 | 2 | 4 — claim for a platform outside `platforms`; `execution_policy` not bound by the `driver` `const`; reserved identifier asserted; unmeasured banner form asserted |
| `swift-tools-version` classifier | 10 | 14 — one per class 1–6, plus each of the eight `F` forms. The positives add the three agreement cases of 7.1: later specification ignored, single-line string literal found by neither, one canonical form per class |
| module name (3.2) | 12 | 8 — a key outside the protocol grammar; a leading `.`, `-` or `_`; an empty key; a non-ASCII key; a `Sk_` result for an overlength key; a `Tk_` result for a short key; a collapsed `my-tool`/`my.tool` pair; a name outside `^[A-Za-z_][A-Za-z0-9_]{0,63}$` |
| plan grammar (4.1) | 1 | 22 — one per negative family enumerated in 4.1.5 |
| layout | 2 | 11 — one per section 6.2 row |

Every negative MUST fail with the exact section 9 code, never with a generic
schema error.
