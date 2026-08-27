# TASK-260822-1yz9ug — first-party module roots decision: outcome

## Result

Decision **0009** landed on `curator-spec` `main`.

- File: `decisions/0009-first-party-module-roots.md` (247 lines)
- PR: https://github.com/relux-works/curator-spec/pull/24 (squash-merged)
- Merge commit on `main`: `b92b105`
- Branch commit: `2c58449`, signed, signature verified `G` (`oparin@me.com`)
- Branch `decision/TASK-260822-1yz9ug-module-roots` and its worktree removed
  after merge; remote branch deleted by `--delete-branch`.

## Numbering

`0009` was verified free against `origin/main` twice: at branch creation and
again immediately before the squash merge. `origin/main` carried `0001`-`0008`
with a pre-existing duplicate at `0005`
(`0005-external-build-repositories.md` and
`0005-vendored-go-boundary-relaxation.md`), which is the collision history the
task warned about. Unmerged `draft/*` branches hold `0009`, `0010`, and `0011`
locally; those are not on `main` and were correctly ignored per the
"next free against `origin/main`" rule. They are a live collision risk for
whoever lands those drafts.

## Decision content

Covers every element the task required:

- `modules` declaration on local build commands (absent/empty keeps current
  single-module meaning); schema bump coordinated with decision 0008.
- Bijection between declared module directories and directory-form `replace`
  directives; undeclared directive rejects, unused declaration rejects.
- Admitted directive form: directory form only, no version on either side, no
  module-to-module redirects.
- Containment: portable relative path other than `.`, real link-free directory
  strictly inside the snapshot, own `go.mod`, unique and pairwise disjoint,
  disjoint from every build root and runtime root.
- Declared directories join the directive, cgo, and assembly scan surface.
- External dependencies stay vendor-only and versioned.
- Snapshot-wide `curator-build-source-v1` identity per core.md 8.1 keeps cache
  keys sound with no algorithm, framing, or receipt change.
- Rejected alternatives: reading `replace` directives as implicit manager input
  (package-controlled steering of the manager) and requiring repository
  consolidation (a packaging shape requirement that costs third-party
  adoption). Also rejected: Go workspaces, extending decision 0005's vendored
  exceptions to replaced modules, and admitting module-to-module redirects.
- Explicit `Compatibility impact` and `Security impact` sections, per
  GOVERNANCE.md's decision-record requirement.

## Probe evidence that shaped the decision

Behaviour was measured, not assumed, with a minimal two-module repro
(`tools/cli` requiring `pkg/lib` at `v0.0.0` via `replace ... => ../../pkg/lib`)
and the spec's exact fixed argument vectors. Three findings changed the shape
of the document:

1. **The build never reads the replacement directory.** `go mod vendor` copies
   a locally replaced module into `vendor/`, and under `-mod=vendor` the
   compiled bytes come from that copy. `go list` reports the package `Dir`
   below `<build root>/vendor`, non-empty `Module.Version` (`v0.0.0`), and a
   populated `Module.Replace`:

   ```json
   {"Path":"../../pkg/lib","Dir":"<snapshot>/pkg/lib",
    "GoMod":"<snapshot>/pkg/lib/go.mod","GoVersion":"1.23"}
   ```

   Those paths are lexical. With the replacement directory renamed away, both
   `go build` and `go list` still succeeded (rc=0) and still reported the same
   `Dir`. A manager therefore cannot treat `Module.Replace` as evidence about
   the tree, which is why the decision validates the *declared* directory
   against the snapshot and explicitly forbids trusting `Replace.Dir`/`GoMod`.

2. **Directory form and module-to-module redirect are exactly separable.** A
   redirect (`replace github.com/spf13/pflag => github.com/spf13/pflag v1.0.6`)
   reports `{"Path":"github.com/spf13/pflag","Version":"v1.0.6",...}` — module
   path plus `Version`, no `Dir`/`GoMod`. Directory form reports `Dir`/`GoMod`
   and no `Version`. The discriminator is precise and implementable identically
   in Go and Python from the same `go list` stream.

3. **No replace directive can hide by being unused.** Adding a `replace` to
   `go.mod` that `vendor/modules.txt` does not record fails the fixed `go list`
   before `go build`:

   ```text
   go: inconsistent vendoring in <build root>:
       example.com/repo/pkg/unused: is replaced in go.mod, but not marked as
       replaced in vendor/modules.txt
   ```

   The complete effective replace set is therefore materialised in
   `vendor/modules.txt`, a regular file below the build root already hashed by
   `curator-build-source-v1`. The bijection is fully observable within the
   existing fixed command set — no new argument vector is needed.

## Security finding recorded in the decision

Because `go mod vendor` copies a replaced first-party module into `vendor/`,
that module matches decision 0005's "vendored non-standard package" predicate.
Left unscoped, enabling module roots would let package-controlled assembly
(`SFiles`) and `//go:cgo_import_dynamic` reach the build under an allowance
written for widely audited third-party dependencies — `go mod vendor` as a
laundering path. Decision 0009 scopes those exceptions to results whose module
carries **no** replacement, which narrows the current trust boundary rather
than widening it, and is not a regression for any accepted package because
replaced modules are rejected outright today.

## Verification

Full CI parity was run locally before push, each command as a standalone
process with its real exit code:

| Command | rc |
| --- | --- |
| `python tools/validate.py` (49 schemas, 471 vector files) | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` (91 tests) | 0 |
| `go test ./tools/...` | 0 |
| `go run ./tools/generate-vectors -root .` | 0 |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | 0 |
| `gofmt -l tools` (empty) | 0 |
| `git diff --check` | 0 |

PR #24 checks: Formatting, Links, Specification (ubuntu/macos/windows), and
Implementations (ubuntu/macos/windows) all **pass**. `Release target
provenance` reported `skipping` on the PR by design — `ci.yml` gates it on
`github.event_name != 'pull_request'`.

Post-merge on `main`: `Specification CI` completed/success and
`Implementation conformance` completed/success, so the release-provenance job
that only runs off-PR also passed against the signed squash commit.

## Deviation from the recipe in task notes

The task notes suggested a worktree under
`curator-spec/.temp/STORY-260822-1pm1c9/` and branch `spec/module-roots-decision`.
The worktree was instead placed under
`curator/.temp/TASK-260822-1yz9ug/curator-spec-worktree` with branch
`decision/TASK-260822-1yz9ug-module-roots`. Reason: `curator-spec` has no
`.gitignore`, so an in-repo `.temp/` shows up as untracked noise in every
`git status` on that checkout. Functionally identical; both are removed.

## Follow-on

Unblocks TASK-260822-3nvx91 (core.md prose and schema). The decision's six open
questions are written to be settled by that task, notably (1) whether the
manager reconciles the vendor copy against the declared directory, and (2)
which surface is authoritative for the effective replace set.
