# TASK-260728-1g0z69 — developer handoff evidence

**Task:** define-toolchain-requirement-and-guidance-contract
**Parent:** STORY-260728-2fsqtv (compiled-build-toolchain-preflight)
**Board status at handoff:** `to-review`
**Revision:** rework cycle 1 — this document reflects the contract *after* the
five blocking reviewer findings were addressed. The finding-by-finding delta is
`TASK-260728-1g0z69_rework-01.md`.
**Task class:** research — the deliverable is a canonical, implementation-ready
decision plus its normative-ready reference, not a protocol change. The
normative landing is `TASK-260728-2jaw7h`.

## Working base

- Task worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1g0z69/curator-spec-worktree`
  (`git worktree add --detach` at `57c1f56` in `curator-spec`, then the accepted
  `TASK-260728-2kp3tv` candidate rsynced in, excluding `.git`).
- Predecessor `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`
  preserved read-only: still 127 dirty entries at `57c1f56` after this task.
- Nothing staged, committed, published, pinned, or released. No platform
  validation claimed.

## Delta versus the accepted predecessor

Exactly three changes (`diff -rq --exclude=.git` against the predecessor):

| Path | Change |
|---|---|
| `decisions/0007-compiled-build-toolchain-preflight.md` | new — the canonical decision |
| `docs/compiled-build-toolchain-requirements.md` | new — implementation-ready reference |
| `CHANGELOG.md` | one bullet under `1.0.0-rc.5` / Added, explicitly stating that the decision changes no rc.5 wire surface |

No schema, vector, generated case, receipt, marker, claim, cache key, or
release-metadata byte changed. `conformance/v1` and `release/1.0.0-rc.5.json`
are byte-identical to the accepted predecessor after regeneration.

## The contract, in short

1. **Closed identifiers.** `toolchain-registry-v1` maps each driver to exactly
   one primary toolchain plus an ordered companion list. Identifiers are the
   closed set `go`, `rust`, `swift`, `kotlin`, and companion-only `jdk`. No
   generic or inferred mapping; an absent entry is unsupported. The `go` entry
   is complete; `rust`, `swift`, `kotlin`, `jdk` are reserved with a fixed
   obligation list their own driver decisions must fill on a qualified host.
1b. **Manager-owned tested-release families.** Every entry declares a REQUIRED
   closed `compatibility` set of release families the manager has tested against
   that driver's conformance vectors. Membership is exact set membership, never
   an ordering test, so it carries the released `profiles/manager.md` §2.2 rule
   forward unchanged: `go` = `{(1, 23)}`, and a Go 1.99.0 that satisfies
   `at_least 1.23.0` is still refused as `build_toolchain_untested_release`. It
   is an independent gate — a requirement can neither widen nor narrow it, and
   nothing on the wire surface can reach it.
2. **One version domain.** Each entry normalizes bounded, locale-independent
   probe output to the canonical triple `major.minor.patch`. Requirements are
   always written canonically: `toolchain-requirement-v1` is exactly `id` and
   `version`, with `exact`, `at_least`, or `range` (inclusive lower, exclusive
   upper). Combination is interval intersection — associative, commutative, so
   source order cannot change the result. An empty intersection is rejected
   before the host is probed, so it fails identically on every machine.
3. **Prerelease policy.** A prerelease host satisfies nothing and a requirement
   can never name one. Rationale: `1.24rc1`, `1.79.0-nightly`, `2.1.0-Beta1`
   and dated Swift snapshots have no shared order, and an unstable compiler
   identity does not belong in a cache key.
4. **Trusted resolution only.** Two admissible origins: manager-bundled, or
   trusted operator configuration in owner-protected state. The forbidden list
   is exhaustive and explicit (ambient/user `PATH`, any package or repository
   byte, runtime roots, `.agents/bin`, shims, manifest/descriptor values,
   inherited environment, every version-manager shim). The exclusion rule is
   stated per surface: on the **manager-defined wire surface** (manifest command,
   descriptor target) the field set is closed and no path, root, URL, mirror,
   channel, track, install command, environment override, credential, keyring,
   checksum, or trust root field exists or may be added; on the
   **source-ecosystem metadata surface** each entry declares a closed
   disposition table assigning every field `forbidden`, `compared`, or
   `ignored`, with forbidden evaluated before compared in a fixed lexical order.
   A version constraint is admitted because it can only **filter** the
   manager-trusted set — it can never add a candidate, and it cannot reach the
   `compatibility` set either.
5. **Two-stage preflight.** Stage A (availability, version, platform) runs right
   after manifest parsing and build-command validation, before external
   acquisition, before cache lookup, before any persistent mutation. Stage B
   (source-metadata cross-check) runs after local snapshot validation or after
   exact external acquisition and audit, and before any cache candidate read or
   compiler child. No cache hit, dry run, or offline mode skips either stage,
   and no gate is reordered ahead of source validation or audit.
6. **Metadata is an assertion, never a selector.** `go.mod`, `Cargo.toml`,
   `rust-toolchain.toml`, `swift-tools-version`, `.swift-version` and the
   selected Kotlin field are compared against the already-resolved toolchain and
   discarded. `rust-toolchain.toml` `path` is `forbidden` and, being forbidden,
   is evaluated first, so a file carrying both `path` and `nightly` is
   deterministically a package-influence rejection. Channel values are
   classified before comparison: a version literal is an `at_least` assertion,
   `stable` is permitted and never honored, `beta`/`nightly`/dated are a
   mismatch, anything else is a mismatch — never a default and never a selector.
7. **Twelve stable `build_toolchain_*` codes**, each with its stage and exact
   trigger, including `build_toolchain_untested_release`. Where two gates would
   both reject a host the Stage A step order decides the reported code, so it is
   deterministic. Diagnostics carry a `guidance_id` and never prose or a URL.
8. **`toolchain-guidance-catalog-v1`** is manager-owned, versioned, and total
   over toolchain × **all twelve** reasons × supported platform, release-gate
   enforced, so there is no runtime guidance-missing case for any code. `reason`
   is exactly the code suffix, so the code-to-reason mapping is the identity and
   stays total as codes are added. Each reason declares a `guidance_class`
   (`host`, `configuration`, `authoring`) that fixes the admissible
   `primary_source` origin — the language's official origin, the manager's
   operator documentation, or this specification — all manager-trusted, text
   plus URL, never a command. Identifiers carry an immutable revision
   (`toolchain.<id>.<reason>.<platform>.r<N>`), exactly one entry per tuple is
   active, and `superseded_by` names a strictly greater revision of the same
   tuple.
9. **No auto-install in v1** — named explicitly for `rustup`, `swiftly`,
   `sdkman`, `asdf`, `mise`, Homebrew, `winget`, the Gradle wrapper, and
   `GOTOOLCHAIN`.
10. **Identity effects.** The resolved toolchain identity stays a build input;
    new drivers bind an ordered `toolchain_identities` array while `go-v1` and
    `go-repository-v1` keep their existing single-field shape byte-for-byte. The
    effective requirement and the `compatibility` set are **gates, not build
    inputs** — same reasoning that kept host capability evidence out of the key
    in decision 0006 — and there is no bypass because Stage A gates before cache
    lookup. Two cases are stated separately so neither is misread: an
    incompatible current toolchain fails at Stage A with **no** cache lookup and
    **no** rebuild, while a cache candidate carrying a *different* toolchain
    identity under a compatible current toolchain is an ordinary key miss and
    rebuild. The guidance catalog is likewise not a cache, receipt, marker, or
    claim input.

## AC mapping

| AC clause | Where |
|---|---|
| canonical and implementation-ready | decision 0007 plus the reference: registry field list including `compatibility`, complete `go` entry, grammar with regex bounds, intersection algorithm, seven ordered Stage A steps, 12-code diagnostic table, catalog schema with classes and revisioned lifecycle, 58-case vector inventory |
| deterministic cross-language version semantics | reference §2: canonical triple, one lexicographic order, anchored per-entry normalization, order-independent intersection, prerelease rejection instead of cross-language prerelease ranking; §1.1.1 exact-membership compatibility as a separate non-ordering gate |
| prevents package-selected paths, URLs, channels, commands, trust roots | reference §3 exhaustive origin lists, §3.1 wire-surface versus metadata-surface exclusion with the closed disposition table and fixed forbidden-before-compared precedence; `build_toolchain_package_influence_forbidden`; metadata-as-assertion rule in §4.2 |
| cache and receipt identity effects | decision §6 and reference §4.3: identity stays an input, requirement, compatibility set, and catalog are not; the incompatible-at-Stage-A case and the different-cached-identity case stated separately; currentness unchanged; status/audit versus install/upgrade/repair split |
| both fail-fast stages without bypassing source audit | reference §4.1 and §4.2: Stage A consumes only manager/operator configuration plus one validated manifest field; Stage B is placed after validation or acquisition-and-audit, so no gate loses a predecessor |
| official platform guidance without stale package-controlled instructions | reference §6: manager-owned versioned catalog, identity code-to-reason mapping total over all twelve reasons, per-class primary-source origins, revisioned IDs with `superseded_by`, catalog excluded from every hashed identity so a URL fix or new revision invalidates nothing |

## Known asymmetry, stated deliberately

A descriptor-declared requirement is not readable at Stage A for an external
command, because the descriptor arrives with the repository. Stage A therefore
gates on baseline ∩ manifest, and the descriptor requirement joins the
intersection at Stage B. This is why the consuming manifest requirement is
REQUIRED in the next schema: the cheap gate always has something to evaluate.
It is documented in reference §4.1 rather than hidden.

## Deliberately not done

- No `protocol/core.md`, `profiles/manager.md`, `SECURITY.md`, schema, or vector
  change. Those are `TASK-260728-2jaw7h`, and the schema version numbers are
  `TASK-260728-2spy93`.
- No `README.md` change. The README table is the released `1.0.0-rc.5`
  specification set; adding a not-yet-normative design reference to it would be
  a false release claim. The new doc is reachable from decision 0007.
- No probe argv, normalization rule, or root layout invented for `rust`,
  `swift`, `kotlin`, or `jdk`. Those need a qualified host; fabricating them
  here would be an unverified platform claim. They are reserved with an exact
  obligation list instead.

## Verification

Every command run standalone; real exit codes below. These are the **rework
cycle 1** runs, against the final content.

Task worktree `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1g0z69/curator-spec-worktree`:

| Command | Exit |
|---|---|
| `python tools/validate.py` (after rework) | 0 — `validated 42 schemas and 422 vector files` |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 — 29 tests |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `gofmt -l tools` | 0, no output |
| `git diff --check` | 0 |
| `go run ./tools/generate-vectors -root .` | 0 |
| `diff -rq conformance/v1 <predecessor>/conformance/v1` | 0 — byte-identical |
| `diff -q release/1.0.0-rc.5.json <predecessor>/release/1.0.0-rc.5.json` | 0 |

Clean-checkout probe `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1g0z69/clean-probe`
(final content copied, `git init`, one unsigned probe commit `9a3fc65`, clean
tree):

| Command | Exit |
|---|---|
| `make regenerate-check` (run 1) | 0 |
| `make regenerate-check` (run 2) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 — `release gate passed for 1.0.0-rc.5 at 9a3fc65370c7…` |

Negative probes against the reworked files, each expected red and reported as
red:

| Probe | Exit | Result |
|---|---|---|
| retired descriptor stem appended to the reworked reference doc | 1 (expected failure) | `docs/compiled-build-toolchain-requirements.md:626: retired repository descriptor name is not an alias and must be absent` — the guard covers new files |
| broken local link appended to the reworked decision 0007 | 1 (expected failure) | `broken local link: ../docs/no-such-file-0007.md` — the link guard covers new files |

Both probes were reverted and the final `python tools/validate.py` run exits 0.

Python gates were run with the existing `validation-venv`, because ambient
`python3` lacks `jsonschema`.

Note on `make regenerate-check` inside the task worktree: it uses
`git diff --exit-code` against `57c1f56`, which cannot pass in a worktree
carrying the uncommitted rc.5 candidate. Determinism was therefore proven two
ways instead — byte-identical regeneration output versus the accepted
predecessor, and two clean-probe `regenerate-check` runs at exit 0.

Logs: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1g0z69/*.log`.

## Reviewer focus

Standing questions from cycle 0:

- Is the "constraints filter, never select" claim actually airtight across the
  manifest, the descriptor, and every stage-B metadata source?
- Is Stage A's earlier position genuinely gate-neutral — does any phase it now
  precedes lose a predecessor it depended on?
- Is excluding the effective requirement from the cache key correct, given that
  Stage A always precedes cache lookup?
- Are the reserved `rust`/`swift`/`kotlin`/`jdk` obligation lists complete
  enough that those driver decisions cannot quietly reopen selection?
- Does the prerelease rejection break a legitimate case worth admitting?

New for cycle 1 — see `TASK-260728-1g0z69_rework-01.md` for the full list:

- Does `compatibility` fully restore the `profiles/manager.md` §2.2 gate, and is
  requirement-before-compatibility the right Stage A precedence?
- Is the §3.1 surface split a real distinction rather than a relabeling?
- Is `channel = "stable"` permitted rather than mismatched the correct call? It
  is the one place this cycle relaxes the previous text.
- Do the identity code-to-reason mapping and `guidance_class` keep
  `primary_source` honest for authoring and configuration reasons?
- Does splitting the old cache vector into positive 14 and negative 45 leave any
  cache-versus-fail-fast case unstated?
