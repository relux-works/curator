# TASK-260827-21xw9d review verdict rev3: changes requested

Reviewer run `RUN-260828-a70495` lineage continued; this run reviewed Change
Request `CR-TASK-260827-21xw9d-3` revision 3, base `41ab53cd`, candidate
`f88a627d`. Scope: `docs/authoring-cli-commands.md` (245 lines) and the one
README link. The delta touches no `.go` file, so no suite was rerun; the sibling
docs in the same CR belong to other story tasks and were not re-reviewed here.

The worktree was byte-compared against the candidate tree before review
(`git show f88a627d:<path>` vs working file: MATCH for both scope files).

Verification binary: `go build -o .temp/review-21xw9d-r3/curator ./cmd/curator`
from the candidate tree (exit 0).

## rev2 -> rev3 delta

Exactly two lines, both the prescribed D5 fix:

    docs/authoring-cli-commands.md:68   "/ build artifact store" -> named external store
    docs/authoring-cli-commands.md:176  cache/build_repositories -> external-build-cache

## D5 fixed

The path now reads `$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact`
at line 176, and line 68 names the same store. Re-derived from the code, not
accepted from the prior verdict:

    internal/install/external.go:105   deps.StoreRoot = filepath.Join(home, "external-build-cache")
    internal/install/external.go:117   deps.resolved(home)
    internal/install/install.go:541    planExternalBuilds(..., cfg.Home(), ...)
    internal/config/config.go:153      func (c *Config) Home() string { return filepath.Dir(c.Path) }
    internal/config/config.go:197      return filepath.Join(home, ".curator", "config.json")
    internal/install/targets.go:185-186
        keyName := strings.TrimPrefix(entry.result.CacheKey, "sha256:")
        finalArtifact := filepath.Join(externalRoot, "artifacts", keyName, "artifact")
    internal/scopes/gc.go:125          same root swept by GC

    grep -rn "cache/build_repositories" . --exclude-dir=.git   ->  (no matches)
    grep -n "build_repositories" docs/authoring-cli-commands.md
        17:  ... the `build_repositories` map in `agent-skill.json`.
        139: "build_repositories": {

Both remaining mentions are the manifest key, not a path segment. D5 is closed.

## D6 (blocking, new): the closed-build-parameters bullet names an error that cannot fire for the stated cause

`docs/authoring-cli-commands.md:40` reads:

> **Closed build parameters**: Argument vectors are hardcoded in
> `internal/godriver/build.go`. Custom build flags, linker flags (`ldflags`),
> compiler flags (`gcflags`), environment overrides, or hook scripts in
> `agent-skill.json` are rejected with `build_execution_package_influence_forbidden`.

The first sentence is exact (`listArguments` and `buildArgumentPrefix`,
`internal/godriver/build.go:32-36`). The second is wrong about which check
fires. I drove it rather than grepping it. Taking the document's own Example 1
manifest and adding one influence key:

    "commands": { "mytool": { "type": "build", "driver": "go-v1",
                              "source_dir": "src/cmd/mytool", "ldflags": "-s -w" } }

    curator skill check ./influence
    error: skill.manifest_invalid (agent-skill.json): commands.mytool.ldflags: field is not supported for build commands

The rejection comes from the parser, not the driver, and carries a different
identifier. `rejectUnknownBuildFields` (`internal/skillspec/parse.go:839-855`)
holds a build command to `{type, driver, source_dir}` plus `modules` at schema
8, and emits `field is not supported for build commands` under
`skill.manifest_invalid`.

`build_execution_package_influence_forbidden` is unreachable from a manifest.
`validatePackageCommandSurface` (`internal/godriver/build.go:299-345`) validates
`request.CommandObject`, which is not the raw manifest object: it is
reconstructed at `internal/install/plan.go:561-571` from the already-parsed
command, with a comment that states the point directly.

    // commandObject reproduces the exact package-declared build-command surface.
    // The parser admits only these fields for a build command, so anything else in
    // the manifest has already been rejected.

So the `extra` slice at `build.go:305-318` is always empty for a parsed
manifest, and the `type` / `driver` / `source_dir` mismatch branches cannot fire
either, because the same parsed command feeds both sides. The diagnostic is
defense in depth at the execution boundary, not the author-facing error.

Why this blocks. It is the third instance of the class rev1 blocked as D1 (an
error identifier attributed to a cause it does not have) and rev2 blocked as D5
(a filesystem path stated without being read out of the code). The acceptance
criterion is "every constraint grep-verified against internal/ with literal
evidence"; this is the one bullet in section 2 whose stated error does not
reproduce when driven. Every other bullet in that section names an error an
author can actually trigger, and all fourteen identifiers were reconfirmed at
non-test call sites this round. An author who adds `ldflags` and searches the
tree for `build_execution_package_influence_forbidden` to understand the refusal
will not find their error.

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

Verify with the `ldflags` manifest above and confirm the printed error matches
what the document claims.

## Non-blocking note: schema-8 interpreters are spec-admitted but Curator-unreachable

Line 65 says schema 8 admits `node-v1` and `python3-v1`; line 69 says Curator
refuses `script-worker-v1`. Both are true, and the composition is not stated:
because `interpreter` requires `execution_policy`
(`internal/skillspec/parse.go:876-880`) and `execution_policy` must be exactly
`script-worker-v1` (`parse.go:883-884`, `ScriptExecutionPolicy` at
`internal/skillspec/types.go:24`), and `scriptpolicy.Admit()` refuses that
policy, no schema-8 command declaring an interpreter can be installed by Curator
today. Driven:

    interpreter node-v1 + execution_policy script-worker-v1
      error: script_execution_policy_unsupported (commands.run-script.execution_policy): this manager does not implement script-worker-v1, and the policy forbids installing the command declared-only, downgrading it, or ignoring the field
    execution_policy only
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.execution_policy: requires 'interpreter'
    interpreter only
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.interpreter: requires 'execution_policy'
    interpreter bash + execution_policy manager-v1
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.execution_policy: must be "script-worker-v1"; omit the field for declared-only execution

Worth one clause in section 3 for an authoring guide, since an author reading
"admitted interpreters" may plan to ship one. Not blocking: no statement in the
document is false.

## Everything else reverified this round

Not accepted from the prior verdicts; re-derived.

Section 2, all fourteen diagnostic identifiers resolve at non-test call sites:

    cgo_required                                 godriver/graph.go:235, build.go:594
    go_native_input_forbidden                    godriver/graph.go:237
    go_pgo_forbidden                             godriver/graph.go:302
    go_generator_forbidden                       godriver/graph.go:286
    go_forbidden_compiler_directive              godriver/graph.go:278
    go_test_input_forbidden                      godriver/graph.go:180
    go_assembly_forbidden                        godriver/graph.go:243
    external_link_forbidden                      godriver/build.go:607
    libgcc_fallback_forbidden                    godriver/build.go:609
    toolchain_switch_forbidden                   godriver/build.go:466
    workspace_dependency_forbidden               godriver/build.go:454
    build_execution_control_unavailable          godriver/controls.go:15
    build_execution_package_influence_forbidden  godriver/controls.go:18
    script_execution_policy_unsupported          scriptpolicy/scriptpolicy.go:37

Substance beyond the identifiers:

- Build vectors literal at `build.go:32-36`, including `-mod=vendor` and
  `-ldflags=-linkmode=internal -libgcc=none`.
- Network off during compilation: `session.go:449-450` sets `GOPROXY: "off"`,
  `GOSUMDB: "off"`, `GOVCS: "*:off"`, `GOFLAGS: ""`.
- Native-input list matches `graph.go:229-230` exactly, and `CgoFiles` is the
  one field that diverts to `cgo_required` (`graph.go:234-236`).
- Assembly clause exact: `firstParty := item.Module != nil &&
  item.Module.Replace != nil` (`graph.go:201`) gates `graph.go:242`.
- `go.work` rejection: `rejectWorkspaceAndToolchainDirectives` walks the build
  root for `go.work` (`build.go:441-458`); the `toolchain` directive scan
  follows at `build.go:463-468`.
- Linux refusal before worker launch: `controls_other.go` is
  `//go:build !darwin && !windows`, `prepareControlDomain` returns
  `diagnostic(CodeControlUnavailable, "the portable execution policy is
  specified for macOS and Windows only")`.

Section 3. `ScriptInterpreters = {"node-v1", "python3-v1"}`
(`internal/skillspec/types.go:29`), and line 65's rationale is the code's own
comment at `types.go:26-28`. Shim contract at line 67 matches `WriteBinShim`
(`runtimestore.go:137-155`): `UnixShimContent` with the `#!/bin/sh` PATH
preamble and `exec` when path entries exist, relative symlink when they do not;
`WindowsShimContent` (`runtimestore.go:177-205`) emits `.cmd` with the
documented double-expansion escaping. Local build cache at lines 68 and 119
matches `base = filepath.Join(store.home, "cache", "build",
buildmeta.DriverGoV1)` (`buildcache/cache.go:262`).

Section 4. All three manifests rebuilt verbatim from the document and validated
against the tree binary:

    curator skill check ./embedded_go        exit=0
    curator skill check ./external_go_repo   exit=0
    curator skill check ./script_skill       exit=0

The single warning on each is `skill.command_resolution_contract_missing`,
caused by the placeholder `SKILL.md` written for the harness, not by the
documented manifests.

Section 5. Each of the six reserved identities substituted into a valid
manifest produces the documented error verbatim:

    kotlin-native-v1, kotlin-native-repository-v1, swift-v1,
    swift-repository-v1, rust-v1, rust-repository-v1
      error: skill.manifest_invalid (agent-skill.json): commands.mytool.driver: must be 'go-v1' or 'go-repository-v1'

README links the document at `README.md:60`. No em-dashes, en-dashes, or
guillemets in the document.

## Routing

Status `to-dev`. D6 is a single-bullet correction with the replacement text and
the reproduction command given above. Nothing else in the document needs to
change; the non-blocking interpreter note is optional. Re-review after the fix.
