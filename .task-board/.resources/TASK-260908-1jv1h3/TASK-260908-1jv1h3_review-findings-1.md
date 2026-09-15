# TASK-260908-1jv1h3 — review findings, cycle 1

**Verdict: ACCEPT.** No blocking or major finding. Five minor/boundary items below,
none of which requires a change to the shipped config.

Subject: `chore/rc-stays-out-of-install-channels` @ `d1bb0a4e`, base `04550e28`,
PR https://github.com/relux-works/curator/pull/65. One signed commit
(`git log --show-signature`: `G`, `oparin@me.com`, ECDSA
`SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`); 1 ahead of `origin/main`,
0 behind.

---

## 0. Change Request delta

No `Change Request Under Review` section was handed to this run and this
`task-board` build exposes no `cr` subcommand, so acceptance is recorded through
the role handoff, not `accept_cr`. The Story workspace is a curator-spec
checkout while the leaf delivers into curator, so a CR delta would read empty;
that is structural in this epic (TASK-260906-284db9 cycle-1 verdict §0) and is
not a finding.

---

## 1. What `auto` actually means — attacked, not read

The brief asks for semantics "from the pinned toolchain". **There is no pinned
version.** `.github/workflows/release.yml:37-39` uses
`goreleaser/goreleaser-action@v6` with `version: "~> v2"` — a range that
resolves to the newest stable v2.x at run time. This is stated as a boundary,
not a defect; see finding F3.

Verified against **two** points of that range:

- **GoReleaser 2.17.0**, source commit `770a4fc7a8fb2dca874b6c98cb739dd64fc931c0`
  — byte-identical to the `GitCommit` the local `goreleaser` binary reports, so
  the source read and the binary are the same artifact.
- **GoReleaser 2.18.1**, the newest stable v2 today (`gh api
  repos/goreleaser/goreleaser/releases`), fetched at
  `internal/pipe/{cask,scoop,release,semver}` — the four decision lines are
  unchanged.

Decision chain, all three fields:

| Field | Code | Behaviour |
| --- | --- | --- |
| `homebrew_casks[0].skip_upload` | `cask.go:147` (2.17.0) / `:158` (2.18.1) | `TrimSpace(SkipUpload)=="auto" && ctx.Semver.Prerelease!=""` → `pipe.Skip`, no tap commit |
| `scoops[0].skip_upload` | `scoop.go:216` / `:218` | same predicate → `pipe.Skip("release is prerelease")` |
| `release.prerelease` | `release.go:74-83` | `case "auto": if ctx.Semver.Prerelease != "" { ctx.PreRelease = true }` |

`ctx.Semver.Prerelease` is `Masterminds/semver.NewVersion(ctx.Git.CurrentTag).Prerelease()`
(`internal/pipe/semver/semver.go:24-38`) — the SemVer prerelease suffix, exactly
as the comments claim.

**Bonus, stronger than the comment claims:** `ctx.PreRelease` does not only set
the prerelease flag. `internal/client/github.go:582-593` forces
`make_latest="false"` whenever it is true, so the comment's "instead of
appearing as the latest stable download" is literally enforced, and
`install.sh:29` (`api.github.com/.../releases/latest`, which excludes
prereleases) is a **third** install channel this leaf protects that the comments
do not name.

**On a stable tag** every gate falls through and publishes exactly as today —
proven in both directions below, not asserted.

### Executable evidence (`probe-run.log`)

Upstream tests, run green at commit `770a4fc7`:

- `release/TestDefaultPreRelease` — 5 subtests, including `auto-release` (stable
  tag → `PreRelease` false) and `auto-rc` (rc tag → true). Both directions.
- `cask/TestRunPipeNoUpload/skip_upload_auto`, `scoop/Test_doRun` (18 subtests,
  incl. `is_prerelease_and_skip_upload_set_to_auto`).

Upstream covers the skip direction for cask/scoop but **not** the publish
direction, so I wrote three review probes (`review-probes.go.txt`) that drive
the real `runAll`→`publishAll` entry points with this project's actual config
values. All green:

| Probe case | `skip_upload` | tag | expected | result |
| --- | --- | --- | --- | --- |
| cask, under review | `auto` | `v0.15.0-rc.1` | skip, no tap commit | PASS |
| cask, under review | `auto` | `v0.15.0` | **publishes** | PASS |
| cask, pre-change | unset | `v0.15.0-rc.1` | **publishes — the incident** | PASS |
| cask, pre-change | unset | `v0.15.0` | publishes | PASS |
| scoop, under review | `auto` | `v0.15.0-rc.1` | skip, no bucket commit | PASS |
| scoop, under review | `auto` | `v0.15.0` | **publishes** | PASS |
| scoop, pre-change | unset | `v0.15.0-rc.1` | **publishes — the incident** | PASS |
| semver | — | `v0.15.0-rc.1` → `rc.1`, `v0.15.0` → `""`, +4 | tag drives the gate | PASS 6/6 |

The gate flips on **both** axes — the config value and the tag shape — and the
pre-change row reproduces the v0.14.0-rc.1 incident mechanism executably. This
is not a delete-only mutant: the `auto`+stable rows assert the gate *admits*
what it must admit.

---

## 2. Comment truth — every clause checked against the record

> "v0.14.0-rc.1 reached the tap because this was unset."

- Tap commit **exists**: `relux-works/homebrew-tap` `3bbc3ada`,
  `2026-08-06T00:22:42Z`, "Brew cask update for curator version v0.14.0-rc.1",
  touching `Casks/curator.rb` (**modified**, i.e. it replaced the stable
  definition in place). It stood for a month, until `951cde1e` (v0.14.0,
  2026-09-06).
- Bucket commit exists too: `relux-works/scoop-bucket` `973742e6`,
  `2026-08-06T00:22:44Z`. The "same reason as the cask above" comment is
  measured, not assumed.
- The field **was** unset at that tag: `git show v0.14.0-rc.1:.goreleaser.yml`
  has no `skip_upload` under `homebrew_casks` or `scoops` and no
  `release.prerelease`.
- The **mechanism** is the unset field and not a neighbouring default:
  `cask.Default()` (`cask.go:53-92`) never assigns `SkipUpload`, so it stays
  `""`, which matches neither `"true"` nor `"auto"` and therefore publishes on
  every tag. The tap commit message is also GoReleaser's *default*
  `CommitMessageTemplate` (`"Brew cask update for {{ .ProjectName }} version
  {{ .Tag }}"`), matching the observed commit character for character — the
  causal loop closes.

---

## 3. Nothing else changes — proven structurally

`git diff 04550e28..HEAD --stat`: `.goreleaser.yml`, 1 file, +10/−0. (The brief
says +8/−0; the actual count is 10 — 3 field lines plus 7 comment lines. The
brief's figure is off, the diff is not.)

Eyeballing a diff is not proof, so both revisions were parsed and flattened
(`yaml.safe_load` → leaf-path dict) and the dicts compared:

```
added:   ['.homebrew_casks[0].skip_upload', '.release.prerelease', '.scoops[0].skip_upload']
removed: []
changed: []
values:  all three == 'auto'
```

Exactly three leaf keys added, nothing removed, nothing altered — no YAML
indentation or anchoring accident. No workflow, tap/bucket repo, footer,
changelog, signing, notarization or nfpm behaviour shifts. `goreleaser check`
on the head config: 1 file validated, exit 0.

The cask/scoop **generation** still runs on an rc; only publication is skipped
(`publishAll`'s own comment: "this is needed so we actually create the
`dist/foo.rb` file"), so nothing downstream that expects those files breaks.

No other install-channel publisher exists in the config — no `aur`, `nix`,
`winget`, `chocolatey`, `krew` or `dockers` stanza — so there is no unguarded
bypass channel. `go install …@latest` is excluded by the module proxy's own
prerelease rule.

---

## 4. Hosted lanes on `d1bb0a4e`

`gh pr checks 65`, run `34167479072`, `headSha d1bb0a4e…`, conclusion
`success`:

| Lane | Result |
| --- | --- |
| Gate self-test (macos / ubuntu / windows) | pass |
| Interop conformance gate | pass |
| Lint, Naming gate | pass |
| Race (macos / ubuntu) | pass |
| Test (macos / ubuntu / windows) | pass |
| Candidate suite (`${{ matrix.os }}`) | **skipping** |

11 green, 0 red. The skipped lane is correct by design, not a casualty:
`ci.yml:322-326` gates `candidate-conformance` on
`github.event_name == 'workflow_dispatch'`, and this is a `pull_request` event.
PR state `MERGEABLE`, base `main`, head `d1bb0a4e` — the reviewed head is the
PR head.

---

## 5. The docs-confidence boundary

**Named explicitly, as the AC requires.** The publish phase cannot be exercised
without a real remote write, so nothing here proves the *hosted* run behaves as
the code says. What this review upgraded is the *semantics* half: it is no
longer docs-confidence but source-and-test confidence at two points of the
`~> v2` range, driven through the production entry points.

What remains unproven and can only be proven by the tag:

- that the hosted runner resolves `~> v2` to a version whose semantics match
  (2.18.1 verified today; the tag may run something newer);
- that `TAP_GITHUB_TOKEN`, artifact upload, signing and notarization interact
  with the skip as expected in a real run.

**First real proof is the rc1 tag itself**: after `v0.15.0-rc.1`, the GitHub
release must read `prerelease: true` and must not be marked latest, and
`relux-works/homebrew-tap` and `relux-works/scoop-bucket` must gain **no** new
commit. That is a three-command check and it is the acceptance test for this
leaf.

---

## Findings

### F1 — minor, wording: "formula" should be "cask"

`.goreleaser.yml:72` — "A release candidate must not replace the stable
**formula**". The tap ships a **cask** (`homebrew_casks`, `directory: Casks`,
and the incident commit touched `Casks/curator.rb`). Homebrew treats formula and
cask as distinct; the comment names the wrong one. Cosmetic, does not affect
behaviour.

### F2 — minor, evidence shape: `goreleaser check` cannot fail on this class

The commit message lists under "Measured here": "`goreleaser check` accepting
the new values". I attacked that claim: a copy of the head config with
`skip_upload: "Auto"` and `prerelease: "ato"` **also passes `goreleaser check`**
(1 file validated, exit 0). The gate is a bare, case-sensitive `==` on an
unvalidated string, so `check` passing is positive-path evidence for a class it
cannot discriminate. The values in the shipped file are correct (verified by
literal grep), so this changes nothing about the change — but the citation
carries less weight than its placement implies, and the discriminating evidence
is §1 above, not `check`.

### F3 — boundary, not a defect: the version is ranged, not pinned

The commit message says "the 2.17.0 the workflow **pins** with `~> v2`". `~> v2`
pins nothing; 2.17.0 is what the producer's local binary happens to be. A tag cut
today runs **2.18.1**. I verified the semantics are identical there, so there is
no live risk — but a future v2.x could in principle change them, and no lane in
this repo would notice. Stated as the residual bound; a v3 would be excluded by
the constraint.

### F4 — coverage gap, follow-up (out of this leaf's scope)

`grep -rn goreleaser` over `*.sh`/`*.go`/`*.yml`/`Makefile` shows **no in-repo
check reads `.goreleaser.yml` at all**. `release-workflow-gate.sh` asserts only
workflow step ordering. Measured AC-row coverage for the three fields:
**0 of 3** guarded by any automated in-repo check; 3 of 3 covered by the
external executable evidence produced at review time, which does not run in CI.
Combined with F2, a future typo — `"Auto"`, `"ato"`, a dropped quote — silently
restores the exact v0.14.0-rc.1 incident with every lane green.

Recommend a follow-up leaf adding three literal assertions to
`release-workflow-gate.sh` (plus the matching negative mutants in
`gate-selftest.sh`, since a source-text checker is defeated by a mutant that
preserves the searched-for token). **Not requested as rework here**: this leaf's
AC is explicitly "nothing else in the release pipeline changes", and adding a
gate would violate it.

### F5 — record clarification, so nobody re-opens this

`gh api …/releases/tags/v0.14.0-rc.1` currently reads `prerelease: true`, which
appears to contradict the commit message's "the GitHub release for an rc is
published as a normal one". It does not. `internal/client/github.go:517-541`
sets `Prerelease: &ctx.PreRelease` **explicitly**, and with the config key unset
`ctx.PreRelease` is `false` — so GoReleaser published it as a normal release.
Of all seven releases, that one is also the only one whose title deviates from
GoReleaser's default `NameTemplate` `{{.Tag}}` ("Curator v0.14.0-rc.1" vs
"v0.14.0-rc.1" everywhere else), and `LOGBOOK.md:8` records the installer
following GitHub latest to v0.13.0.

**Deduction, with its bound stated:** two correlated deviations from an
otherwise uniform six-release pattern indicate the flag and the title were
edited by hand after publication. GitHub exposes no release edit history, so
this is inference from the deviation pattern, not a read — reported as such
rather than as fact. The material point either way: with `prerelease: "auto"`
the marking becomes deterministic from config instead of depending on a human
remembering.

---

## Method and scratch

Read-only against the producer worktree; nothing written there or to the control
root. All scratch under this Story worktree's
`.temp/review-1jv1h3/` (shallow goreleaser clone, config copies, probe sources,
logs). The three probe files live in the throwaway goreleaser clone, not in
curator; no curator file was modified. No publish dry-run was attempted and no
remote write of any kind was made — the only network calls were GitHub reads and
source fetches. No LOGBOOK.md entry, per the brief.
