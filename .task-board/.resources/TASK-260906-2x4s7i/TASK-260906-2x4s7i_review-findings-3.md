# Review findings — TASK-260906-2x4s7i (cycle 3, PR #47 at `2f2dfa4`)

**Verdict: CHANGES REQUESTED.** `repeat-of: cycle-1 stray-file class` (F10), `repeat-of: F9` (F12, F14).
One blocking, one major, four minor.

**PR #47 is not safe to land as it stands** — but the reason is mechanics, not the discriminator.
Every behavioural finding cycle 2 raised (F5, F6, F7a, F7b, F8) is genuinely fixed, and I could not
construct a §1-legal or §6.1-legal spelling that the committed schema misclassifies or refuses. The
blocker is that the rework commit also committed a **5.4 MB compiled Mach-O arm64 executable** at the
repository root, which no CI lane catches and which becomes permanent in the history of a
specification repository the moment this lands.

Subject: `feat/path-overlay-declarable` at `2f2dfa4`, two signed commits past `origin/main` = `550579d`.
Reviewed delta `git diff origin/main..2f2dfa4` (65 files, +4150/−139); cycle-2 delta `bd39adb..2f2dfa4`
(15 files). No Change Request captured for this element; the artefact under review is the PR, per the
cycle-2 and cycle-3 briefs.

Everything below was measured by me against the **committed** schema through `Draft202012Validator`
using `tools/validate.py`'s own `schema_registry()` (`jsonschema==4.25.1`, per `requirements-dev.txt`),
on an `rsync` copy of the clean `2f2dfa4` worktree. `git archive` is **not** usable for this repo — it
applies `export-subst`/eol filters and breaks `conformance/v1/fixtures/byte-exact/subst.txt`; my first
control run failed for that reason and the copy was redone. Every mutant runs the repo's own gate
(`python3 tools/validate.py`) as a standalone process; the schema is restored byte-exactly afterwards
and the SHA-256 asserted (`4c52a22b…19a`).

---

## 0. What the fix got right — verified by me, not read

### The five cycle-2 findings, each re-driven

| cycle-2 finding | claim | my measurement |
|---|---|---|
| **F5** `C://…` refused in all 8 columns | scheme is now ≥2 chars | `C://Users/operator/context` and `C:///Users/operator/context` → **path**, bare VALID, every form INVALID. `M-schemelen` (`+`→`*`) dies on `valid-overlay-path-windows-double-slash.json`. **Fixed and pinned.** |
| **F6** `file:` admitted as a form-free path | `file:` dropped from the allowlist | `file:///x`, `FILE:///x`, `file://server/share/x` → **REFUSED in all 7 columns**. `M-file` (re-admit `file`) dies on `invalid-overlay-file-url-with-form.json`. **Fixed and pinned.** |
| **F7a** host grammar unpinned in both directions | two cases added | `M-host` (`[^/:\s]+`, the permissive class) dies on `invalid-overlay-scp-host-grammar-with-form.json`; `M-host2` (host ≥2 chars) dies on `valid-overlay-git-single-letter-host.json`. **Both directions now pinned.** |
| **F7b** invalid SCP forms silently local | arm 2 refuses them | `git@my_host:x`, `my_host:x` → **REFUSED**, no longer form-free paths. `M-arm2-del` dies on `invalid-overlay-scp-host-grammar.json`. **Fixed and pinned.** |
| **F8** `directory` on a path source unpinned | one negative case | `M-dir` dies on `invalid-overlay-path-directory.json`. All **five** §1-forbidden members (`range`, `tag`, `revision`, `directory`, `branch`) now have a killing mutant. **Fixed.** |

### The two judgement calls the brief asked to be sourced

**`c:example/team-context` pinned as a valid `git` overlay — right, and §1 and §6.1 do not conflict.**
The sentence that decides it is §1's own definition of the `git` kind: "a network git source under the
**core §6.1 canonical identity**". §1 does not describe git spellings itself; it incorporates §6.1 by
reference. core §6.1 gives the SCP form as `[user@]host:path` with "an ASCII host matching
`[A-Za-z0-9][A-Za-z0-9.-]*`" — a grammar that admits a single character. `c:example/team-context` is
therefore a well-formed §6.1 SCP form with canonical identity `c/example/team-context`. §1's `path`
kind has no competing claim: it is "named by an **absolute path**, or by a **project-relative path**",
and a Windows drive-relative spelling (`c:example` resolves against the current directory *of drive C*)
is neither. Nothing in §1 forces the other reading, so this is not a spec gap that needs an amendment —
it is §1 routing the spelling to §6.1, which accepts it. It is also the safe direction under the
cycle-1/F9 rule: misclassified toward `git` fails loudly at identity validation and can never be
silently accepted as a local directory.

**The drive carve-out provably cannot swallow a single-character-host SCP form.** The carve-out is
`^[A-Za-z]:[\\/]` — the character after the colon must be `/` or `\`. The SCP pattern's first path
character is `[^/\s\\]` — which excludes exactly `/` and `\`. The intersection of the two languages is
**empty**, so no string can be both. Confirmed empirically: `M-drive-del3` (delete the carve-out from
arm 3) is a semantic no-op — identical classification on all 55 probed sources. What distinguishes them
is one character: a separator immediately after the colon means drive path, anything else means SCP host.

### Structure, ordering, and the absence of a bypass

* **There is no arm ordering to get wrong.** JSON Schema `allOf` is order-independent, and each arm's
  preconditions are expressed as explicit `not` clauses rather than as fall-through. `M-reorder`
  (arms permuted to 3-2-1) is a semantic no-op across all 55 probes and leaves the corpus green. The
  brief's dimension-1 ordering concern is structurally impossible here, not merely untested.
* **No bypass surface.** `#/$defs/overlay` is referenced exactly once (line 374, the `overlays` map).
  `system-config-v2` only re-refs `overlays_allowed`; `context-lock-v1` and
  `agent-environment-marker-v1` carry a boolean `overlay` flag, not a declaration. There is no second
  schema through which a form-carrying `path` overlay could enter.
* **Declarability composes to the lock — driven, not read.** I built a `context-lock-v1` member for a
  `path` overlay (`kind: context`, `overlay: true`, `state_sha256`, no `source`/`commit`) and validated
  it against the committed lock schema: **VALID**, exactly §1's "its store key and pin are the core §8
  content hash of the snapshot — a state hash, exactly the `local` pin shape". The same member plus
  `directory` is **refused**. So the path overlay this change makes declarable is also lockable, and the
  `directory` prohibition holds on both surfaces.

### The residual edges, decided

| spelling | committed | decided by |
|---|---|---|
| `C:foo`, `c:example/x` | git, form required | §1 `git` → core §6.1 host grammar admits one character; §1 `path` claims neither absolute nor project-relative. Safe direction. |
| `git@host:/x`, `host:/x`, `git@github.com:/example/x` | REFUSED | core §2: a portable relative path "is not absolute"; §6.1 requires "a non-empty **portable repository path**"; the repo's own `validate_repository_path` (`tools/validate.py:420`) rejects `path.startswith("/")`. An invalid network form, and §6.1: "Invalid network forms MUST be rejected, not treated as local." |
| `git@my_host:x`, `my_host:x`, `proj_ect:v2/ctx` | REFUSED | host outside `[A-Za-z0-9][A-Za-z0-9.-]*` ⇒ invalid network form ⇒ same §6.1 sentence. |
| `git@host: x` | REFUSED | §6.1 path "with no whitespace". |
| `\\server\share\x` (UNC), `//host/x` | path, form-free | no colon before a slash ⇒ arm 2 never fires; §1 absolute path. Correct. |
| `a/b:c`, `./a:b`, `../a:b/c` | path, form-free | colon in a later segment ⇒ arm 2's `^[^/\s]*:` cannot cross `/`. Correct — but unpinned, see F12. |
| `.`, `..`, `team`, `packages/team-context` | path, form-free | §1 project-relative. |
| `https://host:8080/x`, `…/x?y`, `…/x#y`, `…/x%20y`, `…/x\y`, `https://user:pw@host/x` | git, form required | §6.1-invalid network forms, but the schema classifies rather than refuses them; they die at identity validation. Consistent with the layer's job — refusal is only needed where the alternative would be *silently local*. Not a defect. |

**F6's consequence, answered.** §1.1 maps "source kind not `git`, `local`, or `path`" to
`profile_source_kind_unsupported`, which is the right diagnostic for `file:` and `svn:`. But the schema
layer **cannot** express which §1.1 diagnostic applies: `svn://x` (`profile_source_kind_unsupported`)
and `git@my_host:x` (`profile_source_invalid` — invalid canonical identity) both produce the same
generic `then: false` failure. That is inherent to JSON Schema and belongs in the stated bounds, not in
the schema.

### Gates, mechanics, pin — re-run and re-read by me

* `make validate` → **exit 0**: `validated 60 schemas and 1043 vector files`; `Ran 227 tests in 44.321s / OK`;
  `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
* `make regenerate-check` → **exit 0**, empty diff. The generator is the only writer of the regenerated
  files: `go run ./tools/generate-vectors -root .` on the committed tree reproduces `conformance/v1` and
  `release/1.0.0-rc.9.json` byte for byte.
* `gh pr checks 47` → 9 lanes, **all pass** (Formatting, Links, Specification and Implementations on
  `ubuntu-latest`, `macos-latest`, `windows-latest`), "Release target provenance" skipping. PR head
  OID = `2f2dfa4`, matching the reviewed commit.
* **Signatures.** Both commits author *and* committer `Ivan Oparin <oparin@me.com>`; both verify **`G`**
  against the repository's own `maintainers.allowed_signers`.
* **The rename is clean.** 907 index entries, **0** missing files; 67 manager-config-v2 entries and 67
  files in that directory, **0** orphans. The old stem `overlay-path-file-url` appears nowhere in the
  tree. (The 25 orphan case files under `agent-environment-marker-v1` and `launch-env-fragment-v1` are
  pre-existing on `origin/main` and untouched by this diff.)
* **Pin consumption clean — read from `relux-works/curator@a3abcf34`'s source, not from a report.**
  `internal/interop/golden_test.go:290` reads `vectors/manager-config.json`; this diff touches only
  `vectors/manager-config-v2.json`. No consumer enumerates `schema-cases/manager-config-v2` — the
  by-name suites are `agent-skill-v*`, `csk-skill-v*`, `install-marker-v*`, `skill-build-v1`,
  `skillfile-dev-v2`, `build-receipt-v1`. `internal/buildrepo/buildrepo.go:32
  ConformanceManifestSHA256 = b6f56aac…` pins the **rc.5** manifest and rc.5 is untouched; only the
  rc.9 candidate pin advanced, generator-written. The three green Implementations lanes are that check
  actually executed.
* **AC row 4 holds.** Corpus scan of every published overlay declaration: 111 form-carrying entries,
  and every one inside an index `valid=True` case has a **git** source. No positive case gives a path
  source a form. Distinct published `source` spellings: 4 (cycle 1) → 16 (cycle 2) → **20**.

---

## Findings

### F10 — BLOCKING. `repeat-of: cycle-1 stray-file class`. The rework commit added a 5.4 MB compiled binary at the repository root

```
$ git log --oneline --diff-filter=A -- generate-vectors
2f2dfa4 Refuse the network spellings the discriminator was treating as local

$ git cat-file -s $(git rev-parse HEAD:generate-vectors)
5462002
$ file generate-vectors
generate-vectors: Mach-O 64-bit executable arm64
$ git cat-file -p HEAD:generate-vectors | strings | grep -m1 curator-spec/tools
github.com/relux-works/curator-spec/tools/generate-vectors
$ git check-ignore -v generate-vectors ; echo $?
1
```

`generate-vectors` is the default output of `go build ./tools/generate-vectors` — the binary's name is
the package directory's name and the build path is embedded in it. **No Makefile target produces it:**
`make regenerate` and `make regenerate-check` both use `go run ./tools/generate-vectors -root .`. It was
swept in by a whole-tree `git add`.

* It is **tracked at HEAD**, not ignored. `.gitignore` covers `.temp/`, `.DS_Store`, `__pycache__/` and
  `*.py[cod]` — the last two added by `550579d` for exactly this class of accident. Nothing covers a
  Go build output.
* It is **platform-specific** (arm64 macOS) in a repository whose CI runs three platforms, and whose
  entire premise is portable, byte-reproducible artefacts.
* **No lane catches it.** All nine checks on PR #47 pass with the binary present; I confirmed the diff
  and `git ls-tree -r HEAD` myself rather than relying on CI.
* Once landed it is **permanent**: 5.4 MB in every future clone of a specification repository, removable
  only by rewriting history.

This is the same shape as the cycle-1 `tools/__pycache__` trap, which the cycle-1 reviewer confirmed
was live and which `550579d` closed for Python only. The brief's dimension 5 names "no stray file"
explicitly, and this is the largest stray file the repository has ever been offered.

**Cost of the fix:** `git rm --cached generate-vectors`, add a `/generate-vectors` line to `.gitignore`
(the root-anchored form, so `tools/generate-vectors/` stays tracked), amend, force-with-lease. No
schema, case, vector or prose change; `make validate` and `make regenerate-check` are unaffected because
neither reads the file.

### F11 — MAJOR. The PR body describes `bd39adb` and contradicts its own head

The review artefact is the PR, and PR #47's description was not updated for the cycle-2 rework. Its
Evidence section quotes gate output that **does not reproduce** at the head:

| PR body claims | measured at `2f2dfa4` |
|---|---|
| "`make validate` exit 0 — 60 schemas, **1037** vector files, 227 tests OK" | `validated 60 schemas and **1043** vector files` |
| "**eighteen** cases, each the positive or negative its classification demands" | 27 overlay cases added across the two commits; 37 overlay schema cases at head |
| "Classification driven directly against the committed pattern, **16 spellings**" | 20 distinct `source` spellings published |
| the pinned spellings include "a **`file:` URL**", listed alongside the positives | `file:` URLs are **refused outright**; the two `file:` cases are both `valid=false` |
| "A scheme outside that set is now refused outright … `svn://host/x` is neither a legal git source nor a directory" | true, but the body never says `file:` joined that set — the reason cycle 2 raised F6 |

The commit message of `2f2dfa4` is accurate and complete; only the PR description is stale. But a
merge commit carries the PR body, so landing this records a description asserting that a `file:` URL is
among the change's positive classification cases when the change refuses it. This is the
"a claim that does not reproduce" shape applied to the artefact's own evidence section.

### F12 — MINOR. `repeat-of: F9`. Four anchors and one character class are unpinned; each changes real classification

24 of my 30 mutants die on a named case. Of the 6 survivors, 2 are provable semantic no-ops
(`M-drive-del3`, `M-reorder`, above). The other four each change how a **named source** classifies, so
each is a real unpinned decision:

| survivor | mutation | classification it silently changes |
|---|---|---|
| **M-unanchored2** | arm 2's `^[^/\s]*:` loses `^` | `a/b:c`, `./a:b`, `../a:b/c` → **path → REFUSED** |
| **M-unanchored1** | arm 1's `^[A-Za-z][A-Za-z0-9+.-]+://` loses `^` | `/x/svn://y`, `ctx/http://y` → path → REFUSED; `x:svn://y` → git → REFUSED |
| **M-unanchored-allow** | arm 1's scheme allowlist loses `^` | `ctx/http://y` → **path → git** |
| **M-scp-bslash** | SCP trailing class `[^/\s\\]` → `[^/\s]` | `host:\x`, `git@host:\x`, `ab:\share\y` → **REFUSED → git** |
| **M-arm2-colonplus** | arm 2's `^[^/\s]*:` → `^[^/\s]+:` | `:`, `:x` → **REFUSED → path** |

`M-unanchored1` is cycle 2's own `M-unanchored` survivor, unaddressed. The most substantive is
**M-unanchored2**: a project-relative path carrying a colon in a *later* segment (`docs/v1:draft/ctx`)
is admitted as a form-free path by exactly one anchor character, and nothing in the corpus notices if
that anchor goes away. `M-scp-bslash` is benign in outcome — both the committed and mutated readings
still reject, one at the schema and one at identity validation — but it means the backslash exclusion
inside the SCP class, which cycle 1 identified as the F1 repair, is now carried entirely by the drive
carve-out and has no evidence of its own.

Two cases close four of the five: a positive `path` case for `a/b:c` kills M-unanchored2, and a case
for a source with an embedded `://` in a later segment (e.g. `context/http-mirror://x`) kills
M-unanchored1, M-unanchored-allow and M-unanchored2 together. Each derivation follows from the
classification flip above — the mutant changes that source's classification, so a case pinning it fails.

### F13 — MINOR. The Windows-drive carve-out in arm 3 is dead code

`allOf[2].if.anyOf[1].allOf[1]` — the `not` on `^[A-Za-z]:[\\/]` inside arm 3's SCP branch — can never
change a result, because that pattern and the SCP pattern have empty intersection (§0). Arm 2's copy of
the same clause *is* load-bearing: `M-drive-del2` dies on `valid-overlay-path-windows-backslash.json`.
Either delete the arm-3 copy or say in the schema that it is deliberate belt-and-braces; as it stands a
future reader will assume it does something.

### F14 — MINOR. `repeat-of: F9`. `profiles/manager.md:2197` is still 107 characters

Cycle 2 asked for the reflow. `profiles/manager.md` is byte-identical between `bd39adb` and `2f2dfa4` —
the file was not touched in the rework. Line 2197 remains 107 characters in a paragraph whose other
lines run 53–75; the file's >90-character line count goes 92 (main) → 93.

### F15 — MINOR. The bounds cycle 2 asked to be stated are stated nowhere, and this cycle adds more

Cycle 2's item 5 asked that `C:foo` and the unanchored-scheme survivor be recorded as stated bounds.
Neither appears in the commit message, the PR body, or the repository. The current bound set is larger:

* `C:foo`, `c:example/x`, `a:b/c` — a colon in the first segment classifies **git**, so a POSIX
  directory literally named `a:b` is not declarable as a path.
* `g://host/x`, `x://y`, `C:/`, `C:\` — a one-letter prefix before `://` is a **drive path**, the
  unavoidable consequence of the F5 repair. RFC 3986 permits one-letter schemes; none of
  `{ssh,git,http,https}` is one, so nothing legal is lost, but the rule should be written down.
* `1://y`, `1:/y`, `:`, `:x`, `x:`, `C:` — **refused**; degenerate spellings that are neither §1 kind.
* `proj_ect:v2/ctx` — refused, because an underscore is outside §6.1's host grammar.
* `M-unanchored*` (F12) and the §1.1-diagnostic limitation (§0).

---

## Mutant table — 30 mutants, 24 killed, 6 survivors (2 of them provable no-ops)

Each runs `python3 tools/validate.py` as a standalone process against the committed corpus; the schema
is restored byte-exactly and its SHA-256 asserted after every run. Every mutant either **narrows** the
gate to admit exactly one member of the class it must reject, or preserves the searched-for tokens while
changing behaviour. There are three deletion mutants (`M-arm1-del`, `M-arm2-del`, `M-drive-del2/3`) and
they are present as controls, not as the evidence.

| mutant | exit | named case that fails | outcome |
|---|---|---|---|
| M0-control | 0 | `validated 60 schemas and 1043 vector files` | green (expected) |
| M-host → `[^/:\s]+` (cycle-1 permissive class) | 1 | `invalid-overlay-scp-host-grammar-with-form.json` | killed |
| M-host2 → host must be ≥2 chars | 1 | `valid-overlay-git-single-letter-host.json` | killed |
| M-dir drop `directory` from the path refusal | 1 | `invalid-overlay-path-directory.json` | killed |
| M-file re-admit `file` to the allowlist | 1 | `invalid-overlay-file-url-with-form.json` | killed |
| M-schemelen scheme `+` → `*` | 1 | `valid-overlay-path-windows-double-slash.json` | killed |
| M-drive-del2 drop the carve-out from arm 2 | 1 | `valid-overlay-path-windows-backslash.json` | killed |
| M-drive-noslash carve-out `[\\/]` → `[/]` | 1 | `valid-overlay-path-windows-backslash.json` | killed |
| M-drive-nobslash carve-out `[\\/]` → `[\\]` | 1 | `valid-overlay-path-windows-slash.json` | killed |
| M-drive-single carve-out `[\\/]` → `[\\/][^/]` | 1 | `valid-overlay-path-windows-double-slash.json` | killed |
| M-drive-wide carve-out → `^[A-Za-z]:` | 1 | `valid-overlay-git-single-letter-host.json` | killed |
| M-arm1-del drop the unknown-scheme arm | 1 | `invalid-overlay-unknown-scheme.json` | killed |
| M-arm2-del drop the invalid-SCP arm | 1 | `invalid-overlay-scp-host-grammar.json` | killed |
| M-arm2-noscp drop "not a valid SCP form" from arm 2 | 1 | `valid-overlay-git-single-letter-host.json` | killed |
| M-arm2-noscheme drop "not scheme-shaped" from arm 2 | 1 | `valid.json` | killed |
| M-oneof git `then` `oneOf` → `anyOf` | 1 | `invalid-overlay-two-requirement-forms.json` | killed |
| M-notallof path `not/anyOf` → `not/allOf` | 1 | `invalid-overlay-path-requirement-form.json` | killed |
| M-range drop `range` from the path refusal | 1 | `invalid-overlay-path-windows-slash-requirement-form.json` | killed |
| M-tag drop `tag` from the path refusal | 1 | `invalid-overlay-path-relative-requirement-form.json` | killed |
| M-revision drop `revision` from the path refusal | 1 | `invalid-overlay-path-requirement-form.json` | killed |
| M-lower allowlist lowercase only | 1 | `valid-overlay-git-ssh-uppercase.json` | killed |
| M-hostcls host class → `[A-Za-z0-9]+` (no dot/hyphen) | 1 | `valid-overlay-git-scp.json` | killed |
| M-unanchored-scp SCP pattern loses `^` | 1 | `invalid-overlay-scp-host-grammar-with-form.json` | killed |
| **M-drive-del3** drop the carve-out from arm 3 | 0 | — | **survivor — provable no-op (F13)** |
| **M-reorder** arms permuted 3-2-1 | 0 | — | **survivor — provable no-op, `allOf` is order-free** |
| **M-unanchored2** arm 2's colon test loses `^` | 0 | — | **SURVIVOR (F12)** |
| **M-unanchored1** arm 1's scheme test loses `^` | 0 | — | **SURVIVOR (F12)** |
| **M-unanchored-allow** allowlist loses `^` | 0 | — | **SURVIVOR (F12)** |
| **M-scp-bslash** SCP class re-admits backslash | 0 | — | **SURVIVOR (F12)** |
| **M-arm2-colonplus** colon test `*` → `+` | 0 | — | **SURVIVOR (F12)** |

---

## AC coverage — 6 of 7 rows pass, 1 fails on mechanics

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | §12.1 knob row: form git-only, form-free path admitted | PASS | `protocol/environments.md:2222`; `profiles/manager.md:2189-2197` agrees member for member, in the manager's voice, citing environments §12.1 and section 1 |
| 2 | schema requires a form only for a `git` source | **PASS** | committed schema via `Draft202012Validator`, 55 spellings; no §1- or §6.1-legal spelling misclassified or refused |
| 3 | path + `range`/`tag`/`branch`/`revision`/`directory` stays invalid | **PASS, 5 of 5** | M-range, M-tag, M-revision, M-dir each die on a named case; `branch` by `additionalProperties: false` + `invalid-unknown-overlay-field.json` |
| 4 | published cases stop giving a path source a form | PASS | corpus scan: 111 form-carrying overlays, every one in a `valid=True` case has a git source |
| 5 | positive + negative path-overlay cases published | PASS | 37 overlay schema cases, 29 vector twins, 20 distinct source spellings |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me, exit 0/0; pin consumption read from `curator@a3abcf34` source; 3 green Implementations lanes incl. `windows-latest` |
| 7 | signed commits, human identity, **no stray file** | **FAIL** | signatures `G`, identity correct — but `generate-vectors`, 5,462,002 bytes of Mach-O arm64, is tracked at HEAD (**F10**) |

---

## What unblocks acceptance

1. **F10** — untrack `generate-vectors`, ignore it (`/generate-vectors`, root-anchored), amend and
   force-with-lease. Nothing else changes.
2. **F11** — rewrite the PR body for the head: correct the vector-file count, the case count, the
   spelling count, and move `file:` out of the positive list.
3. **F12** — publish a positive `path` case for `a/b:c` and one for a source with an embedded `://` in
   a later segment; re-run M-unanchored1/2/-allow and report them dead. `M-scp-bslash` and
   `M-arm2-colonplus` may be recorded as bounds instead.
4. **F13** — delete the redundant arm-3 carve-out, or annotate it.
5. **F14** — reflow `profiles/manager.md:2197`.
6. **F15** — write the bound set down where an implementer will read it.
7. Re-run both gates, requote the exit codes, and re-read `gh pr checks 47` after the push.

**None of items 3–6 is a reason to withhold the schema itself.** If the orchestrator wants the smallest
path to a landable PR, F10 and F11 are the whole blocking set, and both are edits to the commit's
metadata and file list rather than to its content.

Reviewer artifacts (nothing written outside `.temp/`):
`.temp/review-2x4s7i-c3/{matrix.py,mutants.py,survivors.py,sources.txt,edges.txt,probe-sources.txt,
matrix-committed.md,edges-committed.md,mutants.md,survivors.md,survivors-probe.md,validate.log,regen.log}`
in the `curator-spec` main checkout.
