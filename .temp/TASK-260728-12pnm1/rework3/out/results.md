# TASK-260728-12pnm1 — rework cycle 3 handoff

Both blocking findings of `TASK-260728-12pnm1_review-verdict-cycle-2.md` are
closed, each by a host measurement, a contract change and an executable check.
Every accepted cycle-1 and cycle-2 finding is retained. Decisions 0007 and 0008
and every frozen artifact are untouched. macOS arm64 remains the only measured
platform. Nothing was staged, committed, published or pinned.

---

## Blocker 1 — CRLF normalization was not total at the 4096-byte boundary

**The finding was correct and reproduced.** Replaying the superseded order on
the reviewer's exact construction: an LF stream of 4093 `x`, LF, `A`, LF is 4096
raw bytes and was **admitted**; its CRLF form is 4098 raw bytes and was
**rejected**, although folding it yields the first byte for byte. A 410-line
shape is worse — the same 4096-byte folded stream expands to 4506 raw bytes.

**The fix.** Section 2.3 now has two bounds with the fold between them:

1. a **raw capture bound** of 8192 bytes — a resource limit on how much stdout
   is read, never an admission decision;
2. UTF-8 and NUL rejection;
3. the CRLF fold;
4. the **semantic bound** of 4096 bytes on the *folded* stream — the only length
   rule that decides admission;
5. bare CR, terminal-LF, payload rules, unchanged.

**Why 8192.** Folding replaces two bytes with one and rewrites nothing else, so
`len(folded) >= ceil(len(raw)/2)` and any raw stream over 8192 bytes folds to
over 4096. The raw bound is therefore non-lossy by construction, and 8192 is the
smallest value with that property: 4096 CRLF pairs is exactly 8192 raw bytes and
folds to exactly 4096. **Measured** — that maximal expansion passes the capture
limit and is decided by the terminal-LF rule, not by the limit; one pair beyond
it is rejected by the limit and would have folded to 4097 anyway.

**Evidence.** Six new vectors, all host-independent:

- N14/N16 — LF and CRLF pairs whose folded stream is exactly **4096** bytes, in
  a two-line and a 410-line shape: both admitted, with identical payload **and**
  identical framed digest;
- N15/N17 — the same pairs at exactly **4097** bytes: both rejected;
- N18 — the maximal CRLF expansion reaches a semantic rule rather than the
  capture limit;
- N19 — no raw length from 0 to 32768 above the capture bound can fold within
  the semantic bound; 0 such lengths.

Control **C11** keeps the superseded order runnable and is required to keep
failing; it additionally re-checks the replacement on the same pairs and reports
itself as *not* failing if the replacement regresses.

**The published hashes are unchanged.** The measured 192-byte stream still
yields payload `7d8e0833…` / record `7fc35c11…`, and `cargo --version` still
yields `8d712854…` / `d677e668…`.

---

## Blocker 2 — Stage P required canonicalizing a path it permits not to exist

### 2a. The impossibility, replaced by a total algorithm

**The finding was correct.** Section 4.3 permitted a non-dependency target leaf
to be absent and, in the next sentence, required the canonical form of the joined
path. **Measured**: `realpath(3)` on an absent leaf returns `NULL` with `ENOENT`
and `filepath.EvalSymlinks` fails identically, so the rule had no
implementation. A second measurement makes it worse than merely impossible:
canonicalizing an absent leaf below a symbolic link **reports a path outside the
package directory** before failing, so canonicalization was never a containment
check to begin with.

**The fix.** Containment now rests on two total mechanisms, and canonicalization
is demoted to a cross-check where it is defined:

1. the lexical grammar, unchanged, which already proves the joined path is a
   descendant of the manifest directory and so of `R`;
2. a **non-following component walk** from `R` down, which rejects any symbolic
   link or reparse point at any depth *including the leaf itself*, and stops at
   the first component that does not exist or that exists and is not a
   directory — nothing below such a component is opened;
3. canonicalization of the **deepest existing ancestor** only, which is always
   defined — in the worst case `R` itself — required to keep `R` as a
   path-component prefix, so case folding and Windows short names are still
   caught;
4. the leaf policy, which is the only place the two key kinds differ: a
   dependency `path` must exist, be a directory and hold a `Cargo.toml` already
   in `M`; any other path-bearing key is admitted when absent, and nothing is
   canonicalized for it.

**Evidence.** Seven new vectors P31 through P37: the absent leaf that has no
canonical form (the premise, checked rather than asserted), the absent leaf with
safe ancestors reporting the deepest existing directory as its anchor, the
absent leaf below a symbolic-link ancestor rejected *at the link component*, the
same shape as a dependency, the path below an existing regular file where the
two leaf policies diverge as specified, the symbolic-link leaf pointing inside
`R` still rejected, and the build-script file gate. A totality unit test drives
nine shapes through both leaf policies and asserts none escapes without a
verdict.

### 2b. The G2-through-G8 counterpart claim, and the rule that makes it true

**The finding was correct and the gap is exploitable.** **Measured**: with
`package.build` absent and a `build.rs` file present, `cargo metadata` reports a
`custom-build` target and `cargo build` compiles and **executes** the script —
the marker file was written. The forbidden-key gate has nothing to reject, so
before this revision only G2 saw it, after Cargo had read and resolved the tree.

**The fix.** Section 4.2 gains a file rule: any filesystem entry named
`build.rs` directly in a manifest's own directory is
`build_rust_build_script_forbidden`. One `lstat` per manifest directory, on the
name alone.

**The discovery scope was measured, not assumed.** A differently named file, a
nested `build.rs`, and a **directory** named `build.rs` are all ignored by cargo;
`build = false` suppresses discovery; edition 2024 behaves as edition 2021. The
rule is nevertheless stated on the name alone so that it needs no reasoning
about Cargo's discovery semantics and stays correct if they change. It
over-rejects in exactly two measured-inert directions — `build = false` beside a
`build.rs`, and a directory named `build.rs` — neither of which `cargo vendor`
produces, since those manifests carry `build = false` precisely when no build
script file was packaged. Both over-rejections are stated in the contract and
predicted in the structural check rather than discovered by it.

**Section 7.2 is corrected.** The collective assertion is replaced by a per-row
table naming each row's fail-before-Cargo counterpart, with G2 citing both the
`package.build` key row and the new file rule, and an explicit note that the old
claim was false for exactly the auto-discovered shape.

**Evidence.** A structural check measures all seven shapes against both cargo and
the file rule. Positive vector 18 covers the default absent-build-script case;
negative vectors 51 and 52 cover the auto-discovered case and the stated
over-rejection. Control **C12** keeps the key-only gate runnable and is required
to keep failing; like C11 it also re-checks the replacement.

---

## Contract deltas

Both changes are narrowings, both reuse existing diagnostics
(`build_rust_build_script_forbidden`, `build_rust_input_outside_build_root`),
and neither mints a schema, a receipt or a pin.

| | before | after |
|---|---|---|
| conformance vectors | 134 | **150** |
| controls required to fail | 10 | **12** |
| probe host-independent vectors | 68 | **81** |
| probe structural checks | 13 | **14** |
| new diagnostics | — | **none** |

The version-record change **admits** exactly the CRLF encodings of streams whose
LF form was already admissible — the streams the fold exists to make equivalent
— and rejects nothing previously accepted. The Stage P changes reject strictly
more, or decide cases the superseded text could not decide at all.

---

## Verification

Probe, from the extracted tarball:

```text
gofmt -l .      0 files
go vet ./...    exit 0
go build ./...  exit 0
go test ./...   exit 0
```

Probe runs:

| run | result |
|---|---|
| full, macOS arm64 | **exit 0, green** — 19/19 cases, P1 and P2 hold, 24 closure checks with 0 verdicts, **12 of 12** controls failing as required, **81** host-independent vectors with 0 divergences, **14** structural checks with 0 divergences |
| no resolvable toolchain (`env -i`, `HOME=/nonexistent`) | exit 0, 19 cases `not_run` with the reason recorded, the 81 vectors still run with 0 divergences, nothing installed |
| controls-only replay | exit 0, 12 of 12 failing as required |

Repository: `go build ./...`, `go vet ./...`, `go test ./...` all exit 0;
`gofmt -l` reports 0 files outside `.temp/` and `.task-board/`.

Two expected-red gates, both fully attributed to **other tasks'** scratch trees:

- `make check` exit 2 — the `gofmt` half lists files under
  `.temp/TASK-260728-2jaw7h` (a vendored go1.25.1 source tree),
  `.temp/TASK-260720-1zntv0`, `.temp/TASK-260729-1t1z2l` and
  `.temp/TASK-260729-3jmqgl`. Zero occurrences of `TASK-260728-12pnm1`. Its
  `go vet` and `go test` halves are green.
- `tools/validate.py` exit 1 — a broken local link `../release/1.0.0-rc.5.json`
  in `docs/external-build-repositories.md`, an untracked scratch file belonging
  to another task. Moving the six untracked scratch paths aside gives a
  clean-tree baseline of **exit 0**, "validated 30 schemas and 93 vector files";
  they were restored afterwards.
- Scoped link check over the two authored documents: **0 broken of 4** local
  links.

`golangci-lint` was **NOT RUN** — it is not installed on this host.

---

## Board impact

**No new board element is created.** Both fixes add implementation obligations
to the already-linked `TASK-260728-q283m8`, `TASK-260728-13ioo0`,
`TASK-260728-2yxdo7` and `TASK-260728-gjxj1v`, and no schema obligation beyond
what `TASK-260728-251p01` already carries — the two changes mint no diagnostic,
no schema member and no pin. Creating elements for them would be decomposition
for symmetry rather than for a requirement.

## Standing limits, restated rather than quietly dropped

- `platforms` holds only `(macos, arm64)`. macOS amd64, Windows and Linux remain
  qualification obligations with a stated acceptance test, never claims. The
  Windows TOML serialization obligations of section 6.1 remain unmeasured.
- The two residual exposures of section 12 are unchanged and still open:
  compile-time file inclusion through `include_str!`, which needs a seventh
  deferred guarantee in `STORY-260728-327soo`, and FFI against base-installation
  libraries.
- Decision number `0009` stays reserved; it renumbers if Swift
  (`TASK-260728-1yhuqi`) or Kotlin (`TASK-260728-168smo`) lands it first.
