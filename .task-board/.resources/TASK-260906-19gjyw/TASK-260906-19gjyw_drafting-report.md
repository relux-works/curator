# TASK-260906-19gjyw — consume the landed git-source-only overlay rule

Drafting report. Repository `relux-works/curator`, branch `feat/consume-overlay-rule`,
worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume`, base
`7c74a49216f1da820ca562eb2643353ccea520b1` (stage (c) landed).
Authority: curator-spec main `87a0d0060bad64ab883d007dcdf35df7485368bf`.

Stage (c) was written against authority `550579d` and enforced form-on-every-overlay.
The reconciliation (curator-spec PR #47) landed after that authority was frozen and made
the requirement form required **only for a `git` source**. This is the consumption of that
rule, not a defect fix in stage (c).

---

## 1. The discriminator

One exported helper, `identity.ClassifySource(spelling) identity.SourceKind`, in
`internal/identity/sourcekind.go`. It returns `SourceGit`, `SourcePath`, or
`SourceInvalid`, decides from the spelling alone, and **never touches the filesystem**.

It is the Go transcription of the three `allOf` members of the committed
`schemas/v1/manager-config-v2.schema.json` `$defs/overlay`, evaluated in the order the
schema states them:

| Arm | Schema condition | Verdict |
|---|---|---|
| `allOf[0]` | `^[A-Za-z][A-Za-z0-9+.-]+://` and not `^([Ss][Ss][Hh]\|[Gg][Ii][Tt]\|[Hh][Tt][Tt][Pp][Ss]?)://` | `SourceInvalid` |
| `allOf[1]` | `^[^/\s]*:` and not URL and not `^[A-Za-z]:[\\/]` and not SCP | `SourceInvalid` |
| `allOf[2]` | git scheme or SCP `^(?:[^/:\s@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\s\\]` | `SourceGit`, else `SourcePath` |

Two transcription decisions worth naming:

* **ECMA-262 `\s`, not Go `\s`.** The committed patterns are evaluated by an ECMA-262
  engine, whose `\s` covers the vertical tab, the Unicode space separators, U+2028/U+2029
  and U+FEFF; Go's is only `[\t\n\f\r ]`. The class is spelled out (`ecmaSpace`) so an
  exotic space classifies here exactly as the schema classifies it, rather than leaving
  the difference as a bound. Pinned by `identity.TestClassifySourceEcmaWhitespace`;
  mutant **M9** replaces the class with Go's `\s` and kills that test.
* **No trimming.** The schema does not trim, and the corpus pins the untrimmed reading, so
  the helper does not trim. `profile install` trims before calling it, for the reason in §3.

### Classification matrix, against the committed corpus

The candidate root publishes **41 indexed `manager-config-v2` overlay schema cases** (14
valid, 27 invalid) and **33 overlay vector cases** (14 valid, 19 invalid). All of them are
driven through the production `config.Load` / `config.Parse` entry points by
`internal/config` `TestManagerConfigV2SchemaCases` and `TestManagerConfigV2Vectors`. The
matrix below is that corpus plus the decided edges the corpus and the brief state in prose;
`identity.TestClassifySourceMatrix` pins the helper directly so a misclassification names
the arm.

| Spelling | Verdict | Why | Corpus case |
|---|---|---|---|
| `https://github.com/example/x` | git | admitted scheme | `valid.json`, many |
| `HTTPS://…`, `HTTP://…`, `SSH://…`, `GIT://…` | git | scheme is case-insensitive | `valid-overlay-git-{https,http,ssh,git}-uppercase` |
| `git@github.com:example/team-context` | git | SCP with user | `valid-overlay-git-scp` |
| `github.com:example/team-context` | git | SCP without user | `valid-overlay-git-scp-no-user` |
| `c:example/team-context` | git | §6.1 host grammar admits one character | `valid-overlay-git-single-letter-host` |
| `/Users/operator/context` | path | absolute | `valid-overlay-path-source` |
| `packages/team-context` | path | project-relative, no colon | `valid-overlay-path-relative` |
| `packages/team:context` | path | the colon is in a later segment | `valid-overlay-path-colon-later-segment` |
| `C:\Users\operator\context` | path | drive letter, backslash | `valid-overlay-path-windows-backslash` |
| `C:/Users/operator/context` | path | drive letter, slash | `valid-overlay-path-windows-slash` |
| `C://Users/operator/context` | path | a **one-character** scheme is not a URL, so the drive rule takes it | `valid-overlay-path-windows-double-slash` |
| `c:\users\operator\context` | path | lowercase drive | `valid-overlay-path-windows-lowercase-drive` |
| `C:` | refused | a bare drive letter names nothing | `invalid-overlay-bare-drive-letter` |
| `file:///Users/operator/context` | refused | `file` is not an admitted transport | `invalid-overlay-file-url` |
| `svn://github.com/example/x` | refused | unsupported scheme | `invalid-overlay-unknown-scheme` |
| `my_host:example/x` | refused | underscore is outside the §6.1 host grammar | `invalid-overlay-scp-host-grammar-no-user` |
| `git@my_host:example/x` | refused | same, with a user | `invalid-overlay-scp-host-grammar` |
| `github.com:\example\x` | refused | a backslash first path character | `invalid-overlay-scp-backslash-path` |
| `file:` | refused | neither a URL, a drive, nor an SCP remote | matrix (prose edge) |
| `github.com:/example/x` | refused | a slash first path character is not an SCP remote | matrix (prose edge) |
| `-host:example/x`, `git@:example/x`, `github.com:` | refused | host/path grammar | matrix (prose edge) |
| `github.com/example/x` | path | a bare canonical identity carries no colon — see §3 | matrix (prose edge) |
| `""` | refused | an empty source is no source | matrix |

`identity.Parse` also changed: its `len(host) == 1` carve-out ("a single-letter host is a
Windows drive") is retired, because the landed corpus decides `c:example/x` is git on §6.1's
one-character host. The Windows drive spellings never reach that rule — `driveRE`
(`^[A-Za-z]:[\\/]`) returns them local first, in both the `C:/x` and `C:\x` forms — so
retiring it does not turn a drive path into a remote. Pinned by
`identity.TestParseOneCharacterHostIsANetworkIdentity`; mutant **M16** restores the
carve-out and kills it.

---

## 2. The three surfaces, and every addressing mode

### (a) The config reader — `internal/config/environments.go` `parseOverlay`

The blanket `if forms != 1` is replaced by a kind switch. Field grammar
(`range`/`tag`/`revision`/`directory`/`weight`) is validated first, exactly as the schema
validates it in `properties`, then the kind rule applies, exactly as the schema applies it
in `allOf`. `branch` is not a schema property at all and is refused as an unknown field.

`Environments.render()` also changed: a path overlay no longer renders an empty
`"revision": ""` key. That was the `default:` arm of the form switch; it is now
`case decl.Revision != "":`. The §12.1 row is `{source, range|tag|revision, directory?,
weight?}` for a git source and `{source, weight?}` for a path source, and the vector family
asserts the rendered effective configuration.

| Addressing mode | Driven by | Result |
|---|---|---|
| path source, no form | `TestPathOverlayDeclarationParses` (8 spellings) + 7 corpus valid cases | accepted |
| path source + range / tag / revision | `TestSchema2KnobRejections` + `invalid-overlay-path-*-requirement-form` | refused |
| path source + directory | `TestSchema2KnobRejections` + `invalid-overlay-path-directory` | refused |
| path source + branch | `TestSchema2KnobRejections/overlay path branch` | refused (unknown field) |
| git source, exactly one form | `TestGitOverlayDeclarationParses` (8 spellings) + corpus | accepted |
| git source, no form | `TestGitOverlayDeclarationParses` + `invalid-overlay-{no,git-scp-no,git-uppercase-no}-requirement-form` | refused |
| git source, two forms | `invalid-overlay-two-requirement-forms` | refused |
| neither kind | `TestSchema2KnobRejections` + `invalid-overlay-{unknown-scheme,file-url,bare-drive-letter,scp-host-grammar*,scp-backslash-path}` | refused |

Driven through `run()` by `cmd/curator` `TestOverlayFromMachineConfigIsRefusedByKind`: the
config source hands `run()` exactly the reader's error, `profile list` stops there, and the
stderr names the `environments.overlays` row.

### (b) `resolveOverlay` — `internal/envprofile/overlays.go`

Now classifies with the same helper instead of the old `isPathOperand`.

| Addressing mode | Driven by (production entry point) | Result |
|---|---|---|
| path overlay, no form | `TestPathOverlayJoinsClosure` (`Install`) | joins the closure, flagged `overlay`, state-hash pin, default weight |
| path overlay, declared weight | `TestOverlayExplicitWeightOverridesDefault` (`Install`) | rule-4 weight wins over manifest and default |
| path overlay + range / tag / revision / directory | `TestPathOverlayFormIsSourceInvalid` (`Install`, 4 subtests) | `profile_source_invalid` |
| git overlay, https | `TestGitOverlayJoinsClosure` (`Install`) | resolves jointly, commit pin, canonical identity |
| git overlay, SCP | `TestSCPOverlayResolvesAsGit` (`Install`, 2 spellings) | commit pin, canonical identity `example.com/personal` |
| overlay source of neither kind | `TestOverlayInvalidSourceKindIsSourceInvalid` (`Install`, 7 spellings) | `profile_source_invalid`, no clone |
| overlays forbidden | `TestForbiddenOverlaysResolveAlone` (`Install`) | every list emptied |

The refusal assertions require the error to carry the **kind refusal reason** ("neither a
git source nor a path", or the F12 "no network identity" wording), not merely
`profile_source_invalid`. Without that, a mutant that let the spelling fall through to the
git arm would still produce `profile_source_invalid` from a downstream clone failure and
prove nothing.

### (c) `profile compose add` — `cmd/curator/compose.go`

Same kind switch. The previously dead `default: form = "path"` branch of `compose list` is
now live.

| Addressing mode | Driven by (`run()`) | Exit |
|---|---|---|
| bare path source | `TestProfileComposeAddPathRow`, `…RefusesAFormOnAPathSource/relative path bare` | 0, parses back as a path declaration, lists as `path` |
| path source + `--range` / `--tag` / `--revision` | `…RefusesAFormOnAPathSource` (4 rows) | 2 |
| path source + `--directory` | `…RefusesAFormOnAPathSource` | 2 |
| git source, one form | `…RefusesAFormOnAPathSource` (range, tag) | 0 |
| git source, bare | `…RefusesAFormOnAPathSource` (URL and SCP) | 2 |
| git source, two forms | `…RefusesAFormOnAPathSource` | 2 |
| neither kind (`svn://`, `D:`, `git@my_host:…`) | `…RefusesAFormOnAPathSource` | 2 |

Every accepted row is re-read with `config.Load`, so the kind cannot change between writing
the declaration and reading it back.

### End to end: the §6 promise

`cmd/curator` `TestPathOverlayFromMachineConfigJoinsTheClosure` declares a path overlay in
machine configuration with `weight: 250`, runs `profile install <root> --use` through
`run()`, and asserts the environment marker records `personal` as an overlay-flagged member
with weight 250 and a state-hash pin (no commit), and that the emitted root context carries
its chapter. Before the landed rule no operator could declare this row at all.

---

## 3. What changed for `profile install <git-url|path>`

`isPathOperand` is gone. `installOperandKind` calls the same discriminator on the trimmed
operand, then applies **one stated widening**:

> a spelling the discriminator calls `path` that is also a valid core §6.1 canonical
> network identity (`identity.ValidCanonical`) stays `git`.

**Why the widening exists.** `canonicalGit` accepts an already-canonical `host/path`
identity and `ensureRepo` clones it as `https://` + identity, so
`curator profile install github.com/example/x` is a working **network** install today. The
landed discriminator classifies that same bare spelling as `path` — it has to, because
`packages/team-context` is a declarable path overlay and the two are syntactically
identical. Following the classification verbatim on this row would silently turn a network
install into a local one and hand a planted `./github.com/evil-org/pkg` directory exactly
the machine-allowlist bypass that finding F14 exists to prevent. Core §6.1 is explicit that
an invalid network form must not be treated as local; a *valid* one certainly must not be.

The producer brief predicted "`packages/team` and `C:\…` become path operands where they
were not". That prediction does not survive contact with the code: `C:\…` was already a
path operand, and `packages/team` **is** `ValidCanonical`, so treating it as a path would
violate the brief's own governing clause — "make sure nothing that was a `git` install
silently becomes a `path` install". The invariant wins over the prediction, and the
invariant is now a test.

The widening is install-only. It must not be applied to overlay sources: the corpus case
`valid-overlay-path-relative` requires `packages/team-context` (which is `ValidCanonical`)
to be a **path** overlay, so applying it there would fail conformance.

### Complete list of install kind changes — nothing here is silent

> **Corrected in rework 1 (review finding F2).** The table below originally omitted the
> colon-in-a-later-segment class, the one **refuse → accept** direction in the change. The
> row is added, and `packages/team:context` is now pinned in `installOperandCases`. The
> table was re-derived by measurement for the correction: 44 operands driven through the
> stage (c) classification (`isPathOperand` + `canonicalGit`, restored in a throwaway
> checkout of `7c74a492`) and through `installOperandKind` at this head — **31 unchanged,
> 13 changed**, and the 13 are exactly the five change classes listed here.

| Operand class | Before | After | Direction |
|---|---|---|---|
| `github.com/example/x`, `packages/team`, any `ValidCanonical` `host/path` | git | **git** | unchanged |
| `https://…`, `ssh://…`, `git://…`, `http://…`, `git@host:path` | git | **git** | unchanged |
| `/abs`, `./rel`, `../rel`, `.`, `..`, `.\rel`, `C:/x`, `C:\x`, `\\host\share\x` | path | **path** | unchanged |
| `c:example/x` (one-character §6.1 host) | path | **git** | path → git: *adds* the network allowlist gate |
| `D:` (bare drive letter) | path | **refused** | fail-closed |
| `""` (empty operand) | git (then failed at clone) | **refused** | fail-closed |
| `pkg`, `MyOrg/pkg`, `~/pkg` — no colon, not a canonical identity | git (a **local** `git clone <relative path>`, never a network fetch) | **path** | git → path, but no network identity is lost: these never reached the network |
| `github.com:/org/pkg` (slash first path char) | git (canonicalized to `github.com/org/pkg`) | **refused** | fail-closed, and it is what the schema decides |
| `packages/team:context`, `a/b:c` — a colon in a **later** segment | `profile_source_invalid` at `canonicalGit` ("malformed or unsupported network source") | **path** | **refused → path**, the one refuse→accept direction. The spelling carries no network identity (`identity.Parse` errors, `ValidCanonical` is false), so nothing reaches the network either way, and the landed `$defs/overlay` decides it is a path. Pinned by `installOperandCases` and by mutant **M19**. |
| `my_host:x`, `github.com:\x`, `-h:x`, `svn://…`, `file://…` | `profile_source_invalid` at `canonicalGit` | `profile_source_invalid` at the kind gate | same outcome, earlier and named |

The security-meaningful invariant — **no operand carrying a core §6.1 canonical network
identity is ever classified `path`** — is stated as a property test over the whole install
matrix (`TestInstallNeverDemotesANetworkIdentityToAPath`) and end to end by the unchanged F14
test `TestOperandShadowedByDirectoryResolvesAsGit`. Mutant **M10** drops the widening and
kills all three.

> **Corrected in rework 1 (review finding F1).** As first written, the property test
> **survived M10**: its loop filtered on `identity.Parse` being non-empty, and `Parse`
> returns `("", nil)` for a bare canonical identity, so `github.com/evil-org/pkg`,
> `github.com/relux-works/pkg` and `example.com/org/pkg` — the class the widening exists
> for — were skipped. It reported `checked 5` for a class of size 0. The loop now admits
> both spellings of the class (`Parse` non-empty **or** `ValidCanonical`), reports
> `checked 8 network-identity operands, 3 of them bare canonical`, and fails if the matrix
> ever loses the bare spelling. M10 was re-applied after the repair and kills all three
> named tests; the claim in this paragraph is now a measurement.

---

## 4. Mutant table

Harness: `.temp/mutants/harness.py`, results `.temp/mutants/results.json`, log
`.temp/mutants/run-24.log`. Each mutant keeps the gate present and weakens it to admit
exactly one member of the class it must reject (or flips exactly one classification arm);
the named test is then executed against the candidate root and its real exit code recorded.
The harness counts `=== RUN` lines and reports **NOT-RUN** rather than "killed" when the
`-run` filter matched nothing — an earlier revision of the harness anchored the whole
`Parent/child` filter as one regexp, which Go splits on `/`, producing an unbalanced group
that silently ran zero tests and reported eight false survivors. No mutant below is a
delete-only mutant.

| # | Gate narrowed to admit | Named failing test | Exit | Verdict |
|---|---|---|---|---|
| M1 | reader: a **revision** on a path source | `config TestManagerConfigV2SchemaCases/invalid-overlay-path-requirement-form.json` | 1 | killed |
| M2 | reader: a **form-free git** source | `config …/invalid-overlay-no-requirement-form.json` | 1 | killed |
| M3 | reader: exactly the **`svn://`** scheme | `config …/invalid-overlay-unknown-scheme.json` | 1 | killed |
| M4 | reader: exactly `directory: packages/team` on a path source | `config …/invalid-overlay-path-directory.json` | 1 | killed |
| M5 | discriminator: an **underscore** SCP host | `config …/invalid-overlay-scp-host-grammar-with-form.json` | 1 | killed |
| M6 | discriminator: a **one-character** URL scheme | `config …/valid-overlay-path-windows-double-slash.json` | 1 | killed |
| M7 | discriminator: drive rule covers only the **forward-slash** spelling | `config …/valid-overlay-path-windows-lowercase-drive.json` | 1 | killed |
| M8 | discriminator: drops plain **`http`** | `config …/valid-overlay-git-http-uppercase.json` | 1 | killed |
| M9 | discriminator: whitespace class becomes **Go `\s`** | `identity TestClassifySourceEcmaWhitespace` | 1 | killed |
| M10 | install: **drops the network-identity widening** | `envprofile TestOperandShadowedByDirectoryResolvesAsGit`, `TestInstallOperandKindIsSyntactic`, `TestInstallNeverDemotesANetworkIdentityToAPath` | 3 | killed — **3 of 3 after the F1 repair**; as first written the third *survived*, see rework 1 |
| M11 | install: neither-kind refusal admits exactly **`D:`** as a path | `envprofile TestInstallRefusesANonSourceOperand` | 1 | killed |
| M12 | `resolveOverlay`: neither-kind refusal admits exactly **`C:`** into the local arm | `envprofile TestOverlayInvalidSourceKindIsSourceInvalid` | 1 | killed |
| M13 | `resolveOverlay`: path-form gate admits exactly a **`directory`** | `envprofile TestPathOverlayFormIsSourceInvalid` | 1 | killed |
| M14 | `compose add`: path-form gate admits exactly **`--revision`** | `curator TestProfileComposeAddRefusesAFormOnAPathSource` | 1 | killed |
| M15 | `compose add`: git-form gate admits a **form-free git** source | `curator TestProfileComposeAddRefusesAFormOnAPathSource` | 1 | killed |
| M16 | canonicalization: restores the **single-letter-host** carve-out | `identity TestParseOneCharacterHostIsANetworkIdentity` | 1 | killed |
| M17 | rendering: restores the unconditional **`revision`** key on a path overlay | `config TestManagerConfigV2Vectors/schema2-overlay-path-source` | 1 | killed |

**17 of 17 killed, 0 survivors** *as first reported* — but the M10 row overstated its
kill count, and review finding F1 measured `TestInstallNeverDemotesANetworkIdentityToAPath`
as a **survivor** of M10. After the rework-1 repair M10 kills 3 of 3, and two further
narrowing mutants are added: **M18** (the property loop's filter narrowed back to
`Parse`-only, killed by the new bare-class guard) and **M19** (install refuses a colon in a
later segment, which survived the reviewed matrix and is killed by the added
`packages/team:context` operand). Current standing: **19 of 19 killed, 0 survivors.**

### Token-preserving mutants against the text-inspecting gate

No Go gate added here inspects source text, but this change edits
`.github/ci/platform-cases.tsv`, which `ledger-consistency.sh` reads as text and matches
against the test names `go list` reports per target GOOS. That is the source-text-inspecting
gate this change touches, and it is attacked by two mutants that **preserve the searched-for
token** and change behaviour. The harness (`.temp/mutants/text-gate-harness.py`, log
`.temp/mutants/text-gate-36.log`) runs **both** the static checker and the behavioural suite
for each, because the behavioural suite is green under both.

| # | Preserves | Changes | `ledger-consistency.sh` | behavioural suite | Verdict |
|---|---|---|---|---|---|
| T1 | the ledger row's `TestClassifySourceMatrix` token, verbatim | the Go source renames the case, so the row requires a case no build compiles | **1** — "required on linux but not compiled into that build" | **0 (green)** | killed |
| T2 | the row's token and the case itself | the row narrows `must_run_on` to `darwin`, leaving linux and windows undeclared | **1** — "compiled into the linux build but undeclared there" | **0 (green)** | killed |

Both mutants are invisible to `go test`. A harness that executed only the behavioural suite
would report them as survivors; that is the whole point of the clause.

The other static checkers (`gate-selftest.sh`, `no-broad-suppression.sh`) are unmodified by
this change and were run green.

---

## 5. Retired bound and ledger rows

**The bound is retired.** Stage (c) carried an explicit bound: *"no operator can declare a
path overlay from config or CLI until the spec fix lands, so it is not claimed as driven;
`cmdComposeList`'s `default: form = "path"` branch is dead for the same reason."* Both
halves are now false: `TestPathOverlayFromMachineConfigJoinsTheClosure` declares one from
machine configuration through `run()`, `TestProfileComposeAddPathRow` declares one from the
CLI row and observes the `path` form column, and the branch is live. The bound is not
carried forward in any repository file (it lived in the stage (c) board reports only).

**Ledger rows.** Rows 303–305 were rewritten to describe what their tests now assert; 12
new rows were registered.

| Row | Before | After |
|---|---|---|
| 303 `config TestPathOverlayDeclarationParses` | "a path overlay carrying a revision parses per the 550579d family; the section 1 refusal fires at resolution (… consumption gap: TASK-260906-19gjyw)" | "a path overlay parses with no requirement form at all, in every path spelling the landed discriminator admits (absolute, relative, colon-in-a-later-segment, and all four Windows drive spellings)" |
| 304 `envprofile TestPathOverlayJoinsClosure` | "… from a manufactured Policy (resolution machinery only; no operator surface can declare it until the TASK-260906-19gjyw consumption lands)" | "a form-free path overlay joins the closure flagged overlay at the machine default weight and survives update" |
| 305 `envprofile TestPathOverlayFormIsSourceInvalid` | "a form on a path overlay is reader grammar but resolution-invalid under section 1" | "a range, tag, revision, or directory on a path overlay is profile_source_invalid at resolution" |
| 376 `curator TestProfileComposeAddPathRow` | "… requires exactly one form and lists it" | "… takes no requirement form, parses back as a path declaration, and lists in the path form column" |

**Corrected in rework 1 (review finding F1).** Two of the newly registered rows were
reworded, because as first written one of them asserted a property no test asserted:

| Row | As registered | Corrected |
|---|---|---|
| `envprofile TestInstallNeverDemotesANetworkIdentityToAPath` | "no install operand carrying a core 6.1 canonical network identity is classified path, which is what would hand it the allowlist bypass" — **untruthful**: the named test skipped every bare canonical identity | "… over both spellings of that class — the canonicalizable URL or SCP remote and the bare already-canonical host/path the install widening exists for — …". The test now checks all 8 members, 3 of them bare. |
| `envprofile TestInstallOperandKindIsSyntactic` | "… widened only so a canonical network identity stays git" | "… , and a colon in a later segment is a path operand" appended, for the F2 operand added to the matrix |

Both rows were driven through `platform-case-gate.sh` after the repair (against a ledger
filtered to those two rows and a real `go test -json` stream): both observed **PASSING by
name**, gate exit 0. `ledger-consistency.sh` is exit 0 at 225 rows.

New rows, all `linux,darwin,windows`, no tolerated skip: `config
TestGitOverlayDeclarationParses`; `identity TestClassifySourceMatrix`,
`TestClassifySourceNeverProbesTheFilesystem`, `TestClassifySourceEcmaWhitespace`,
`TestParseOneCharacterHostIsANetworkIdentity`; `envprofile
TestInstallOperandKindIsSyntactic`, `TestInstallNeverDemotesANetworkIdentityToAPath`,
`TestInstallRefusesANonSourceOperand`, `TestOverlayInvalidSourceKindIsSourceInvalid`,
`TestSCPOverlayResolvesAsGit`; `curator TestProfileComposeAddRefusesAFormOnAPathSource`,
`TestPathOverlayFromMachineConfigJoinsTheClosure`,
`TestOverlayFromMachineConfigIsRefusedByKind`.

`ledger-consistency.sh` proves, from `go list` per target GOOS, that all **225** rows name
cases actually compiled into the linux, darwin and windows builds.

---

## 6. Two-root gate table

Both conformance roots were materialized as plain checkouts (`cp -R` of a real worktree, and
`git worktree add --detach` for the pin) and **verified file-by-file against their published
`manifest.json`**: candidate 1047/1047 files match, pin 691/691 match, 0 mismatches.

> An earlier attempt materialized the roots with `git archive`. That silently corrupted
> `conformance/v1/fixtures/byte-exact/subst.txt` (65 bytes on disk against the vector's 40)
> through the repository's `.gitattributes` filters, and `internal/interop
> TestConformanceSnapshotAcquisition` caught it. That was a materialization defect in this
> report's tooling, not a product failure; the manifest verification above now precedes
> every lane.

Every gate below ran as its own process, unpiped, with its real exit code recorded. The two
lanes were run **sequentially**, never concurrently and never alongside a `-race` suite.

| Gate | Root | Command | Exit | Log |
|---|---|---|---|---|
| build | — | `go build ./...` | **0** | `.temp/static-25.log` |
| vet | — | `go vet ./...` | **0** | `.temp/static-25.log` |
| gofmt | — | `gofmt -l cmd internal` (no paths listed) | **0** | `.temp/static-25.log` |
| lint | — | `golangci-lint run ./...` — 0 issues | **0** | `.temp/lint-26.log` |
| gate selftest | — | `gate-selftest.sh` — 94 passed, 0 failed | **0** | `.temp/gate-selftest-27.log` |
| suppression | — | `no-broad-suppression.sh` | **0** | `.temp/nbs-28.log` |
| ledger | candidate | `ledger-consistency.sh` — 225 rows across linux darwin windows | **0** | `.temp/ledger-29.log` |
| **test gate** | **pin `0ed5c691`** | `test-gate.sh` — served=69, deferred=3, excluded=0 | **0** | `.temp/test-gate-pin-31.log` |
| **test gate** | **candidate `87a0d006`, `CI_REQUIRE_FULL_ROOT=1`** | `test-gate.sh` — served=72, **deferred=0**, excluded=0 | **0** | `.temp/test-gate-candidate-30.log` |
| cmd suite | candidate | `go test -count=1 -timeout 30m ./cmd/curator` | **0** | `.temp/cmd-curator-22.log` |
| mutants | candidate | `.temp/mutants/harness.py` — 17 killed, 0 survivors | **0** | `.temp/mutants/run-24.log` |

The `SPEC_PIN` root still defers exactly the three registered packages as designed —
`internal/config`, `internal/envfragment`, `internal/envmarker` — each because the pin
publishes none of their registered artefacts; they run with `CURATOR_CONFORMANCE_ROOT`
unset and the platform-case gate records the deferral by name.

### The 30 subcases

Reproduced before the change against the candidate root: `go test ./internal/config` failed
with exactly **30** failing subcases (15 in `TestManagerConfigV2SchemaCases`, 15 in
`TestManagerConfigV2Vectors`) — 7 invalid cases accepted and 8 valid cases rejected in the
schema-case family, and the matching split in the vector family. All 30 now pass; the
package is green against the candidate root and the candidate lane is `deferred=0`.

---

## 7. Acceptance-criteria coverage

**7 of 7 AC rows driven**, with the production call site named for each.

| # | AC row | Production call site | Driving test | Status |
|---|---|---|---|---|
| 1 | A path overlay is reachable from all three surfaces and joins the closure with its weight, driven through `run()` | `config.Load` → `envprofile.PolicyFromConfig` → `Install` → `resolveOverlay`; `cli.cmdComposeAdd` | `curator TestPathOverlayFromMachineConfigJoinsTheClosure`, `TestProfileComposeAddPathRow`; `config TestPathOverlayDeclarationParses`; `envprofile TestPathOverlayJoinsClosure` | driven |
| 2 | One exported discriminator matches the landed schema, never probes the filesystem | `identity.ClassifySource`, called from `parseOverlay`, `resolveOverlay`, `installOperandKind`, `cmdComposeAdd` | `config TestManagerConfigV2SchemaCases` (41 cases) + `TestManagerConfigV2Vectors` (33); `identity TestClassifySourceMatrix`, `TestClassifySourceNeverProbesTheFilesystem` | driven |
| 3 | A path overlay carrying range/tag/branch/revision/directory stays `profile_source_invalid` | `parseOverlay` (reader) and `resolveOverlay` (resolution) | `config TestSchema2KnobRejections` + 5 corpus cases; `envprofile TestPathOverlayFormIsSourceInvalid`; `curator TestOverlayFromMachineConfigIsRefusedByKind` | driven |
| 4 | `profile install` uses the same helper with no source silently changing kind | `envprofile.installOperandKind` → `installLocked` | `envprofile TestInstallOperandKindIsSyntactic`, `TestInstallNeverDemotesANetworkIdentityToAPath`, `TestInstallRefusesANonSourceOperand`, `TestOperandShadowedByDirectoryResolvesAsGit` | driven; every kind change enumerated in §3, re-derived by measurement in rework 1 (44 operands, 31 unchanged, 13 changed) after F2 found the table incomplete |
| 5 | The candidate lane with `CI_REQUIRE_FULL_ROOT=1` is green, closing the 30 subcases | `.github/ci/test-gate.sh` | candidate lane exit 0, served=72 deferred=0 | driven on **darwin only** — see bounds |
| 6 | The stale bound and ledger rows 303–305 are retired and truthful | `.github/ci/platform-cases.tsv` | `ledger-consistency.sh` exit 0, 225 rows; two rows re-driven through `platform-case-gate.sh` | driven — after the rework-1 correction of the one untruthful newly registered row (F1) |
| 7 | Every new refusal is driven through `run()` and proven by a narrowing mutant | as above | 19 mutants, 19 killed (see rework 1: M10's third named test survived as first written and is repaired; M18 and M19 added) | driven |

---

## 8. Stated bounds

1. **Two of the three runners are unverified here — this is the one open blocker.** Every
   lane above ran on **darwin** only. What *is* proved portably: every ledger row's case is
   compiled into all three platforms' builds (`ledger-consistency.sh` from `go list` per
   target GOOS, 225 rows, exit 0), and `GOOS=linux go vet ./...` and `GOOS=windows go
   vet ./...` are both exit 0. The claim "green on all three runners" is nevertheless
   **unknown** for linux and windows and is not inferred from the darwin result. See §9.
2. **The one-character-host SCP spelling is not driven through a live clone.** The network
   fixture reaches git through a `url.<base>.insteadOf` rewrite, and a `c:`-prefixed remote
   is exactly the spelling git itself may read as a drive-relative local path on Windows,
   which would make that row a property of the fixture rather than of curator. It is pinned
   where it is portable instead: `identity.TestParseOneCharacterHostIsANetworkIdentity` for
   canonicalization, `identity.TestClassifySourceMatrix` and
   `config.TestGitOverlayDeclarationParses` for classification.
3. **`TestClassifySourceNeverProbesTheFilesystem` plants nothing on Windows.** Every git
   spelling carries a colon, which Windows forbids in a path component, so no directory of
   that name can exist there at all. The test logs how many spellings it planted and
   asserts the verdicts either way. The bound is in the safe direction: on Windows the
   shadowing attack the test defends against cannot be constructed.
4. **The transcription is pinned by the corpus, not by the schema text.** The conformance
   root publishes `conformance/v1` only — it does not publish `schemas/v1/`, so no test can
   compare the Go patterns against the committed pattern strings. The pin is the 41 overlay
   schema cases and 33 vector cases driven through `config.Load` / `config.Parse`. A future
   schema change that the corpus does not exercise would not be caught here.
5. **`branch` on an overlay is refused as an unknown field, not as `profile_source_invalid`.**
   The schema has no `branch` property and `additionalProperties: false`, and
   `OverlayDeclaration` has no field to carry one, so the declaration cannot reach
   resolution. AC row 3 names `branch`; the refusal is at the reader with the
   unknown-field diagnostic, which is the only place it can exist.
6. **`profile install` trims its operand before classifying; the overlay reader does not.**
   The schema does not trim and the corpus pins the untrimmed reading, while every
   downstream git entry point (`canonicalGit`, `identity.Parse`) already trims. A leading
   space therefore changes an overlay `source`'s classification and does not change an
   install operand's. This is stated, not tested per spelling.

---

## 9. The one open item: AC row 5 / checklist item 6

**Checklist item 6 — "the candidate lane … is green on all three runners" — is left
unchecked, and the handoff guard is fail-closed on it.** It is not checkable from a
developer machine, for a structural reason:

* `.github/workflows/ci.yml` `candidate-conformance` is guarded by
  `github.event_name == 'workflow_dispatch' && (inputs.candidate_ref != '' || inputs.candidate_root != '')`,
  over the matrix `[ubuntu-latest, macos-latest, windows-latest]`. It is **not** part of the
  push or pull-request CI, so it runs only when an operator dispatches the workflow with
  `candidate_ref=87a0d0060bad64ab883d007dcdf35df7485368bf`.
* The producer brief instructs: *"Do not push and do not open a PR."* Without a push there is
  no ref for that dispatch to run against.
* Locally there is no linux or windows execution environment on this machine (no running
  container runtime; `lima` is installed with no instance). Every git spelling also carries a
  colon, which Windows forbids in a path component, so some rows could not be constructed
  there in any case.

Everything reachable from here is done and green: the candidate lane on **darwin** is exit 0
with `served=72, deferred=0` under `CI_REQUIRE_FULL_ROOT=1`, all 30 previously failing
`internal/config` subcases pass, and the cross-platform *content* of the lane is proved by
`ledger-consistency.sh`. The residual risk this leaves unmeasured is a platform-conditional
execution difference on linux or windows; the change adds no platform-conditional code — it
is pure spelling classification — but that is an argument, not a measurement, and it is
reported as unknown rather than as green.

**Exact input needed:** a `workflow_dispatch` run of `candidate-conformance` with
`candidate_ref=87a0d0060bad64ab883d007dcdf35df7485368bf` against the integrated head, on
ubuntu-latest, macos-latest and windows-latest. That requires a push, which is the
orchestrator's step and outside this task's authorization. Alternatively, if item 6 was
intended to mean "the developer reproduced the candidate lane locally against a materialized
`87a0d006` root" — which is what the producer brief actually asked of this role — the item
should be reworded to say so. Rewording an acceptance row is not this role's call, which is
why the item is left unchecked rather than reinterpreted.

### Resolved — the measurement now exists (added 2026-09-07, publish-only run)

**The block is closed, and it was closed by measurement rather than by argument.** The
orchestrator pushed the branch, opened PR #62 and dispatched `candidate-conformance` against
head `38702164` with `candidate_ref=87a0d0060bad64ab883d007dcdf35df7485368bf`: run
[34071813375](https://github.com/relux-works/curator/actions/runs/34071813375). All three
`Candidate suite` jobs concluded **success** — ubuntu-latest (job 101590439643, 3m38s),
macos-latest (job 101590439723, 8m7s), windows-latest (job 101590439718, 32m17s) — and each
job's own log shows `CI_REQUIRE_FULL_ROOT: 1`, the resolved `candidate_revision
87a0d0060bad64ab883d007dcdf35df7485368bf`, `deferred=0`, and `test-gate: stage served exit=0`,
with `internal/config` served and `ok` on all three (`TestManagerConfigV2SchemaCases`,
`TestManagerConfigV2Vectors`, `TestPathOverlayDeclarationParses`,
`TestGitOverlayDeclarationParses` among them). Two facts are recorded rather than smoothed
over: ubuntu reports `served=71 deferred=0 excluded=1`, the one exclusion being
`internal/godriver`, excluded on linux by the root's own
`vectors/conformance-claim-v3-qualification.json` — a designed platform exclusion, not a gap
in this change's coverage; and at the time of writing the enclosing dispatch run is still
`in_progress` on an unrelated `Test (windows-latest)` job, which does not bear on the
candidate lane. Separately, PR #62 at the same head reports 11 SUCCESS checks with `Candidate
suite` SKIPPED, as expected for a `pull_request` event. **This paragraph supersedes §7 row 5
("driven on darwin only") and §8 bound 1 ("two of the three runners are unverified"): AC row 5
is driven on all three runners, and bound 1 is retired.** No other bound in §8 changes, and no
repository change was made in this run.
