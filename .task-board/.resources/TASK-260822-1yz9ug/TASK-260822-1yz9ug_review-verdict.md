# TASK-260822-1yz9ug — review verdict: changes requested

Reviewer run `RUN-260822-c8550c`, read-only. Verdict branch: **changes requested → `to-dev`**.

The decision landed and its normative core is sound. Two findings against the
document as merged require a follow-up decision-only PR amending
`decisions/0009-first-party-module-roots.md`. **Nothing here asks for a revert
of `b92b105`.**

## What passes

| DoD item | Result | Evidence |
| --- | --- | --- |
| Decision under `decisions/`, next free number vs `origin/main` | pass | `0009-first-party-module-roots.md` at `b92b105`; `origin/main` carried `0001`-`0008` (pre-existing duplicate at `0005`) before the merge |
| Covers `modules`, bijection, scan surface, cache identity per 8.1, rejected alternatives | pass | points 1, 2, 5, 7; both required rejections plus three more |
| Squash-merged to `main`, required checks green | pass | PR 24 `MERGED`, `mergeCommit b92b105`, single parent `a2d44eb`; Formatting, Links, Specification ×3, Implementations ×3 all `pass`; `Release target provenance` correctly skipped — `.github/workflows/ci.yml:87` gates it on `github.event_name != 'pull_request'`; both post-merge `main` workflows green |
| Branch and worktree removed | pass | no local or remote `decision/TASK-260822-1yz9ug-module-roots`; no `.temp/STORY-260822-1pm1c9` worktree |
| Outcome resources | pass | `_results.md`, `_decision-0009.md` |
| Logbook | pass | `LOGBOOK.md:37` and `:47`; entry numbering (`### 2024`) follows the file's HHMM convention, not a sequence |
| Governance format | pass | `GOVERNANCE.md:61-66` wants context, decision, alternatives, compatibility impact, security impact — all present |

Spot-checks of the technical claims all hold:

- `core.md` 4.2 (`origin/main:protocol/core.md:205`) matches the Context verbatim on the
  build-root/`go.mod`/intervening-module rules, vendor-only resolution, and the
  `GOPROXY=off`/`GOWORK=off`/`GOFLAGS=`/`CGO_ENABLED=0` environment.
- `profiles/manager.md` 2.3 (`:294`) matches the containment and scan-surface descriptions,
  including the `SFiles` and `golang.org/x/sys` `cgo_import_dynamic` exceptions the decision
  scopes in point 5.
- The "current total rejection" claim is exact against the reference implementation:
  `cocoaskills/src/csk/builds/go_v1.py:980` raises `vendor_metadata_inconsistent` when a
  non-standard result has a non-nil `Module.Replace`.
- `core.md` 8.1 (`:901`) does bind every regular file of the snapshot under
  `curator-build-source-v1`, and does make links invalid — so point 7's "identity unchanged"
  and point 4's "link-freeness already follows from 8.1" are both correct.
- No `decisions/README.md` index exists, so nothing needed indexing.

## F1 — the first consumer does not install through a `type: "system"` manifest

The claim appears three times and is false in all three:

- Context: "Because `go-v1` cannot package it, the skill installs through a `type: "system"`
  manifest instead — a strictly smaller audit surface than the compiled build it should be
  getting."
- Rejected alternative "Require repository consolidation": "It is exactly why the first
  consumer ships as a `type: "system"` manifest today, trading a closed, audited, vendor-only
  compiled build for an unmanaged installation path."
- Consequences: "First consumer: `skill-project-management` switches from `type: "system"` to
  a `go-v1` build".

Measured:

- `relux-works/skill-project-management` `origin/main:agent-skill.json` is `schema_version 6`,
  `build_roots ["tools/board-cli","tools/board-tui"]`, three commands, **all**
  `"type": "build", "driver": "go-v1"` (`task-board`, `tb-sessiond`, `task-board-tui`).
- Only two commits ever touched that file (`3dfb2d2`, `cc251cf`, both 2026-08-07); both are
  build/`go-v1`. There is no `"type": "system"` anywhere in the repo. `Skillfile.json` is a
  consumer lockfile, not a command manifest.
- The snapshot curator actually installs — `~/.curator/cache/project-management/<rev>/snapshot/agent-skill.json`
  at the revision pinned by `origin/main:Skillfile.json` (`ca5c4fd3`) and at the older
  `3dfb2d2` — is the same build/`go-v1` manifest.

The real situation is better motivation than the false one, and it should replace it: the
packaged snapshot ships **with the `replace` directives stripped and the first-party modules
pre-vendored as ordinary `v0.0.0` requirements**. At `ca5c4fd3`:

- `tools/board-cli/go.mod` requires `pkg/board v0.0.0` and `pkg/remoteconfig v0.0.0` and
  contains **zero** `replace` lines;
- `tools/board-cli/vendor/modules.txt` lists both as ordinary explicit modules with **zero**
  `=>` lines;
- `vendor/github.com/relux-works/skill-project-management/pkg/...` holds the copies.

Meanwhile the git tree diverged from that shape: `2ed3acd` added the directory-form replaces,
and `.gitignore:65-67` now excludes `tools/*/vendor/` with the rationale recorded under
`TASK-260819-3vr8j3` (a stale vendor tree auto-enables `-mod=vendor` and fails every build).
So `main` as it stands is unbuildable under `go-v1` for two independent reasons — no vendor
tree in the packaged snapshot and replaced modules rejected — which is the accurate version of
"`go-v1` cannot package it".

Minor, same area: the Context says `tools/board-cli` **and** `tools/board-tui` "each require
`pkg/board`, `pkg/remoteconfig`, and `pkg/providerlimits`". `tools/board-tui/go.mod` requires
and replaces only `pkg/remoteconfig`; only `tools/board-cli/go.mod` carries all three.

The epic note on `EPIC-260822-18ylpq` is the origin of both errors ("Until this lands the skill
installs via type=system manifest"; "Its vendor trees ... are already prepared on that repo and
stay valid for the switch" — they are gitignored on `main`). It needs the same correction, or
the next producer re-inherits it.

## F2 — point 5's laundering path is not actually closed, and the first consumer is in it today

Point 5 scopes decision 0005's vendored exceptions "to results whose module carries no
replacement, so that `go mod vendor` cannot launder package-controlled assembly or
dynamic-import directives into the build under a third-party allowance", and the Security
impact concludes this "narrows the current trust boundary rather than widening it".

The replacement predicate does not carry that weight. A package that strips its `replace`
directives from the packaged snapshot and vendors its own first-party modules as `v0.0.0`
requirements presents `Module.Replace == nil` to the manager. The 0005 exceptions therefore
apply to its own code, and the new bijection never fires — there is no directive to biject
against, so the package declares no `modules` and is validated exactly as today.

This is not hypothetical. It is the shipping shape of the named first consumer, verified above
at `ca5c4fd3`: first-party `pkg/board` and `pkg/remoteconfig` sitting in `vendor/` with no
replacement recorded anywhere, indistinguishable from third-party vendored code to a conforming
manager. `cc251cf` ("Reduce Go builds to CLI/TUI only, avoid assembly") suggests that boundary
has already been felt in practice.

Consequences for the document:

- The Compatibility impact line "replaced modules are rejected outright today, so nothing
  shipped can depend on the wider reading" is true as literally written but misleading: a
  shipped, accepted package already routes first-party code through the third-party allowance
  by not having a replacement at all, and decision 0009 leaves that route open.
- Either the security claim needs narrowing to what it actually delivers (it closes the
  laundering path only for packages that *keep* their replace directives, i.e. the newly
  admitted shape), or the decision needs an additional mechanism — and if it is the latter,
  that is a design question, not an editorial one.

## Requested changes

Follow-up decision-only PR to `curator-spec` `main`, same landing recipe:

1. Replace the three `type: "system"` statements with the verified current shape of the first
   consumer (stripped replaces + pre-vendored first-party modules, snapshot diverging from the
   git tree, `main` currently unbuildable under `go-v1`).
2. Fix the per-module attribution: `tools/board-cli` carries all three; `tools/board-tui`
   carries `pkg/remoteconfig` only.
3. Add "strip the `replace` directives from the packaged snapshot and pre-vendor the
   first-party modules as ordinary `v0.0.0` requirements" as a rejected alternative, with its
   real failure modes: the packaged tree diverges from the source repo with no drift gate, and
   first-party code enters under an allowance written for audited third-party dependencies.
4. Correct F2 — either narrow point 5 and the Security impact to the guarantee actually
   delivered, or record the residual as an explicit open question alongside the existing six.
5. Correct the `EPIC-260822-18ylpq` note so the normative task
   (`TASK-260822-3nvx91`) does not inherit the false premise.

Numbering warning still live, from the implementer's own findings: `draft/TASK-260728-12pnm1-rust-driver`,
`draft/TASK-260728-168smo-kotlin-native-driver`, and `draft/TASK-260728-1yhuqi-swift-driver`
hold `0009`, `0010`, `0011` locally and now collide with `main`.
