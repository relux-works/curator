# TASK-260827-21xw9d review verdict rev4: accepted

Reviewer run `RUN-260828-a70495` lineage continued. Change Request
`CR-TASK-260827-21xw9d-4` revision 4, base `41ab53cd`, candidate `7e43aa4e`.
Scope reviewed: `docs/authoring-cli-commands.md` (245 lines) and the one README
link. The delta touches no `.go` file
(`git diff --name-only 41ab53cd 7e43aa4e | grep -c '\.go$'` = 0), so no suite
was rerun; the sibling docs in the same CR belong to other story tasks and were
not re-reviewed here.

The worktree was byte-compared against the candidate tree before review
(`git show 7e43aa4e:<path>` vs working file: MATCH for `docs/authoring-cli-commands.md`
and `README.md`).

Verification binary: `go build -o .temp/review-21xw9d-r4/curator ./cmd/curator`
from the candidate tree (exit 0), after
`git submodule update --init --recursive`.

## rev3 -> rev4 delta

Exactly one file, one line: the prescribed D6 fix.

    git diff --name-only f88a627d 7e43aa4e
    docs/authoring-cli-commands.md

Line 40 replaced. Nothing else moved.

## D6 fixed

The bullet now reads:

> **Closed build parameters**: Argument vectors are hardcoded in
> `internal/godriver/build.go` (`listArguments`, `buildArgumentPrefix`). A build
> command in `agent-skill.json` admits only `type`, `driver`, `source_dir`, and,
> from schema 8, `modules`. Any other key, including `ldflags`, `gcflags`,
> environment overrides, or hook fields, is rejected during manifest parsing
> with `skill.manifest_invalid`: `commands.<name>.<field>: field is not
> supported for build commands` (`internal/skillspec/parse.go`,
> `rejectUnknownBuildFields`). The driver re-validates the presented command
> surface at the execution boundary and rejects any influence attempt with
> `build_execution_package_influence_forbidden` (`internal/godriver/build.go`,
> `validatePackageCommandSurface`); that check is defense in depth, because the
> parser has already refused the field.

Every clause was driven against the tree binary, not grepped.

The reproduction the rev3 verdict prescribed. The document's own Example 1
manifest plus `"ldflags": "-s -w"`:

    curator skill check ./influence
    error: skill.manifest_invalid (agent-skill.json): commands.mytool.ldflags: field is not supported for build commands
    exit=1

Matches the documented string exactly, including the `skill.manifest_invalid`
class and the `commands.<name>.<field>` shape. The unmodified Example 1 still
validates (`curator skill check ./embedded_go` exit=0), so the manifest itself
is not what changed the outcome.

The bullet's list of influence keys was driven key by key, each added to the
otherwise-valid Example 1 manifest:

    gcflags      error: skill.manifest_invalid (agent-skill.json): commands.mytool.gcflags: field is not supported for build commands
    env          error: skill.manifest_invalid (agent-skill.json): commands.mytool.env: field is not supported for build commands
    hooks        error: skill.manifest_invalid (agent-skill.json): commands.mytool.hooks: field is not supported for build commands
    pre_build    error: skill.manifest_invalid (agent-skill.json): commands.mytool.pre_build: field is not supported for build commands
    post_build   error: skill.manifest_invalid (agent-skill.json): commands.mytool.post_build: field is not supported for build commands

The "only `type`, `driver`, `source_dir`, and, from schema 8, `modules`" claim
holds at the source and was driven at the schema boundary:

    internal/skillspec/parse.go:839-855
        allowed := map[string]bool{"type": true, "driver": true, "source_dir": true}
        if schema >= 8 { allowed["modules"] = true }
        ...
        return verr.New(label+"."+unknown[0], "field is not supported for build commands")

    schema 6 + "modules"  error: ... commands.mytool.modules: field is not supported for build commands
    schema 8 + "modules"  error: ... commands.mytool.modules: build_module_root_declaration_invalid: ...

At schema 6 `modules` is an unknown field; at schema 8 it passes the allow-list
and reaches its own value validation. That is precisely the boundary the bullet
states.

`rejectUnknownBuildFields` is called from the `go-v1` branch at
`internal/skillspec/parse.go:368`, which is where the bullet lives (the `go-v1`
admission matrix). Correct placement.

The defense-in-depth clause is backed by the code's own comment.
`validatePackageCommandSurface` (`internal/godriver/build.go:141,299`) validates
`request.CommandObject`, which `commandObject`
(`internal/install/plan.go:561-571`) rebuilds from the already-parsed command
with exactly the four admitted keys:

    // ... The parser admits only these fields for a build command, so anything else in
    // the manifest has already been rejected. `modules` appears only when the
    // schema-8 command declared a non-empty list ...

So the document now names the error an author can actually trigger and
correctly demotes `build_execution_package_influence_forbidden` to the
execution-boundary check it is. The three earlier defects of this class (rev1
D1 wrong identifier, rev2 D5 invented path, rev3 D6 wrong cause) are all closed.

Build vectors named in the first sentence are literal:

    internal/godriver/build.go:32  listArguments = []string{"list", "-mod=vendor", "-deps", "-json", "-buildvcs=false", "-compiler=gc", "-pgo=off", "."}
    internal/godriver/build.go:34  buildArgumentPrefix = []string{
      "build", "-mod=vendor", "-trimpath", "-buildvcs=false", "-buildmode=exe", "-compiler=gc", "-pgo=off",
      "-ldflags=-linkmode=internal -libgcc=none", "-o", }

## Full reverification this round

Not accepted from the prior verdicts. Re-derived and re-driven.

### Section 2, admission matrix

All fourteen diagnostic identifiers resolve at non-test call sites in
`internal/`:

    cgo_required                                 godriver/build.go:594
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

Network-off during compilation (`internal/godriver/session.go:449-451`):
`GOFLAGS: ""`, `GOPROXY: "off"`, `GOSUMDB: "off"`, `GOVCS: "*:off"`,
`GOWORK: "off"`, `CGO_ENABLED: "0"`, `GO_EXTLINK_ENABLED: "0"`.

Linux refusal before worker launch: `internal/godriver/controls_other.go` is
`//go:build !darwin && !windows`, and `prepareControlDomain` returns
`diagnostic(CodeControlUnavailable, "the portable execution policy is specified
for macOS and Windows only")`. The document's phrasing matches the code's own
reason.

### Section 3, script commands

`ScriptInterpreters = {"node-v1": true, "python3-v1": true}`
(`internal/skillspec/types.go:29`), and line 65's rationale is the code comment
verbatim (`types.go:26-28`: "admitting one is a specification revision, not a
manager configuration option").

Mutual requirement and policy refusal driven at schema 8 with `capabilities`
present:

    interpreter + execution_policy
      error: script_execution_policy_unsupported (commands.run-script.execution_policy): this manager does not implement script-worker-v1, and the policy forbids installing the command declared-only, downgrading it, or ignoring the field
    execution_policy only
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.execution_policy: requires 'interpreter'
    interpreter only
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.interpreter: requires 'execution_policy'

matching `internal/skillspec/parse.go:876-881`.

Shim contract (line 67) matches both `WriteBinShim` branches
(`internal/runtimestore/runtimestore.go:137-155`): `UnixShimContent` with the
`#!/bin/sh` PATH preamble and `exec` when path entries exist, relative symlink
when they do not; `WindowsShimContent` emits the `.cmd` launcher.

Target resolution contract (line 68) re-derived from code:

    runtime store      runtimestore.Dir  = filepath.Join(home, "runtime", skillName, commit)      runtimestore.go:24-26
    local build cache  base = filepath.Join(store.home, "cache", "build", buildmeta.DriverGoV1)   buildcache/cache.go:262
                       DriverGoV1 = "go-v1"                                                        buildmeta/models.go:21
    external store     deps.StoreRoot = filepath.Join(home, "external-build-cache")                install/external.go:105
                       finalArtifact = filepath.Join(externalRoot, "artifacts", keyName, "artifact")  install/targets.go:185-186
    manager home       Config.Home() = filepath.Dir(c.Path); path = <home>/.curator/config.json    config/config.go:153,197

All three documented paths are correct. `cache/build_repositories` remains gone
from the repository.

### Section 4, worked examples

All three manifests rebuilt verbatim from the document and validated against
the tree binary:

    curator skill check ./embedded_go        exit=0
    curator skill check ./external_go_repo   exit=0
    curator skill check ./script_skill       exit=0

The single warning on each is `skill.command_resolution_contract_missing`,
caused by the placeholder `SKILL.md` written for the harness, not by the
documented manifests.

### Section 5, planned drivers

Each of the six reserved identities substituted into a valid manifest produces
the documented error verbatim:

    kotlin-native-v1              error: skill.manifest_invalid (agent-skill.json): commands.mytool.driver: must be 'go-v1' or 'go-repository-v1'
    kotlin-native-repository-v1   (same)
    swift-v1                      (same)
    swift-repository-v1           (same)
    rust-v1                       (same)
    rust-repository-v1            (same)

### Cross-cutting

README links the document at `README.md:60`. `docs/build-ssh.md` and
`docs/build-https.md` both exist and are linked from Example 2. No em-dashes,
en-dashes, or guillemets in the document.

## Acceptance criteria

- Admission matrix and every constraint grep-verified against `internal/` with
  literal evidence: yes, fourteen identifiers plus the build vectors, the
  network-off environment, the native-input and directive lists, the assembly
  first-party gate, the `go.work` and `toolchain` rejections, the Linux refusal,
  the interpreter set, the shim branches, and all three filesystem paths.
- Three worked examples validate against the tree binary: yes,
  `curator skill check` exit 0 on each, binary built from the candidate tree.
- Planned-language paragraph names the actual validation error: yes, driven for
  all six reserved identities.
- Prose style clean: yes.
- README links it: yes, `README.md:60`.

## Routing

Accepted. `accept_cr(TASK-260827-21xw9d, revision=4)`. No `commit_ack` supplied;
the orchestrator makes the `done` transition after committing the scope.
