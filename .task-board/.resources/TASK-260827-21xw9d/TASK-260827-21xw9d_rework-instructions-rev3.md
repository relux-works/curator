# TASK-260827-21xw9d rev3 rework: one blocking defect (D6)

Reviewer run against `CR-TASK-260827-21xw9d-3` revision 3 (candidate
`f88a627d`), 2026-08-28.

D5 is fixed and was independently reverified: `docs/authoring-cli-commands.md`
lines 68 and 176 now name
`$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact`, which
matches `internal/install/external.go:105` plus
`internal/install/targets.go:185-186`, and `cache/build_repositories` is gone
from the repository. `build_repositories` survives in the document only as the
manifest key at lines 17 and 139.

The whole document was reverified this round, not just the delta. One bullet
that both earlier rounds only grepped, and that I drove, is wrong.

## D6 (blocking): the closed-build-parameters bullet names an error that cannot fire

`docs/authoring-cli-commands.md:40` currently reads:

> **Closed build parameters**: Argument vectors are hardcoded in
> `internal/godriver/build.go`. Custom build flags, linker flags (`ldflags`),
> compiler flags (`gcflags`), environment overrides, or hook scripts in
> `agent-skill.json` are rejected with `build_execution_package_influence_forbidden`.

The first sentence is exact. The second names the wrong check. Taking the
document's own Example 1 manifest and adding one influence key:

    "mytool": { "type": "build", "driver": "go-v1",
                "source_dir": "src/cmd/mytool", "ldflags": "-s -w" }

    curator skill check ./influence
    error: skill.manifest_invalid (agent-skill.json): commands.mytool.ldflags: field is not supported for build commands

The parser refuses it, not the driver. `rejectUnknownBuildFields`
(`internal/skillspec/parse.go:839-855`) holds a build command to
`{type, driver, source_dir}` plus `modules` at schema 8.

`build_execution_package_influence_forbidden` is unreachable from a manifest.
`validatePackageCommandSurface` (`internal/godriver/build.go:299-345`) validates
`request.CommandObject`, which is not the raw manifest object: it is rebuilt
from the already-parsed command at `internal/install/plan.go:561-571`, whose own
comment says so.

    // commandObject reproduces the exact package-declared build-command surface.
    // The parser admits only these fields for a build command, so anything else in
    // the manifest has already been rejected.

So the `extra` slice at `build.go:305-318` is always empty for a parsed
manifest. The diagnostic is defense in depth at the execution boundary, not the
author-facing error. This is the same class as rev1 D1 (wrong error identifier)
and rev2 D5 (invented path), and the acceptance criterion is that every
constraint be grep-verified against `internal/` with literal evidence.

### Fix

Replace the bullet at line 40 with something like:

    - **Closed build parameters**: Argument vectors are hardcoded in
      `internal/godriver/build.go` (`listArguments`, `buildArgumentPrefix`). A
      build command in `agent-skill.json` admits only `type`, `driver`,
      `source_dir`, and, from schema 8, `modules`. Any other key, including
      `ldflags`, `gcflags`, environment overrides, or hook fields, is rejected
      during manifest parsing with `skill.manifest_invalid`:
      `commands.<name>.<field>: field is not supported for build commands`
      (`internal/skillspec/parse.go`, `rejectUnknownBuildFields`). The driver
      re-validates the presented command surface at the execution boundary and
      rejects any influence attempt with
      `build_execution_package_influence_forbidden`
      (`internal/godriver/build.go`, `validatePackageCommandSurface`); that
      check is defense in depth, because the parser has already refused the
      field.

Verify by adding `"ldflags": "-s -w"` to the Example 1 manifest and running
`curator skill check` against a tree binary. Paste the literal output.

## Optional, non-blocking

Section 3 may add one clause: because `interpreter` requires `execution_policy`
(`internal/skillspec/parse.go:876-880`), `execution_policy` must be exactly
`script-worker-v1` (`parse.go:883-884`), and `scriptpolicy.Admit()` refuses that
policy, no schema-8 command declaring an interpreter is installable by Curator
today. Both existing statements at lines 65 and 69 are true; only the
composition is unstated. Driven output for all four cases is in
`TASK-260827-21xw9d_review-verdict-rev3.md`.

## Do not change anything else

The rest of the document was reverified in full this round and holds: all
fourteen diagnostic identifiers at non-test call sites, the build vectors, the
network-off environment (`GOPROXY`, `GOSUMDB`, `GOVCS`, `GOFLAGS` at
`session.go:449-450`), the native-input and directive lists, the assembly
first-party gate, the `go.work` and `toolchain` rejections, the Linux refusal
before worker launch, the interpreter set and its rationale, the mutual
`execution_policy`/`interpreter` errors, the `script-worker-v1` refusal, both
`WriteBinShim` branches and the Windows `.cmd` note, the runtime store target,
the local build cache path, the external artifact store path, the `.git`
restatement, all three worked examples (`curator skill check` exit 0 each), all
six reserved driver identities and their verbatim error, the README link at
`README.md:60`, and prose style (no em-dashes, en-dashes, or guillemets).
