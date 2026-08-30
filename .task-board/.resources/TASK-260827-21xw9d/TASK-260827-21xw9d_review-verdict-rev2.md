# TASK-260827-21xw9d review verdict rev2: changes requested

Reviewer run `RUN-260828-a70495`, 2026-08-28. Change Request
`CR-TASK-260827-21xw9d-2` revision 2, base `41ab53cd`, candidate `2509c7d3`.
Scope reviewed: `docs/authoring-cli-commands.md` (new, 245 lines) and the one
README link. The delta touches no `.go` file
(`git diff --name-only 41ab53cd 2509c7d3 | grep -c '\.go$'` = 0), so no suite
was rerun; the sibling docs in the same CR belong to other story tasks and were
not re-reviewed here. The worktree was byte-compared against the candidate tree
before review (`git show 2509c7d3:<path>` vs working file: MATCH for both scope
files).

Verification binary: `go build -o .temp/review-21xw9d-r2/curator ./cmd/curator`
from the candidate tree (exit 0).

## Rev1 defects: three fixed, one incompletely fixed

### D1 fixed

`docs/authoring-cli-commands.md:40` now names
`build_execution_package_influence_forbidden`.

    grep -rn "build_execution_package_influence_forbidden" --include='*.go' internal/
    internal/godriver/controls.go:18:  CodePackageInfluenceForbidden = "build_execution_package_influence_forbidden"

    grep -rn "package_build_command_influence_forbidden" .
    (no matches)

The invented identifier is gone from the repository.

### D3 fixed

Line 159 no longer states the `.git` suffix as a validation rule. It is now
operational advice about a `301` redirect against Curator's protected fetch,
which is what `docs/build-https.md:140` says verbatim, and it notes that the
parser trims the suffix for identity derivation
(`internal/buildrepo/buildrepo.go:127`). Reconfirmed the suffix is not required:

    "git": "https://github.com/example/remote-tool"
    curator skill check ./nogit    exit=0

### D4 fixed

Line 237 uses `kotlin-native-v1` and `kotlin-native-repository-v1`, matching
`LOGBOOK.md:2849`, and now also names the `-repository-v1` halves for Swift and
Rust and states that reservation is explicitly not admission. Each of the six
reserved identities was substituted into a valid manifest; all six produce the
documented error verbatim:

    kotlin-native-v1              error: skill.manifest_invalid (agent-skill.json): commands.mytool.driver: must be 'go-v1' or 'go-repository-v1'
    kotlin-native-repository-v1   (same)
    swift-v1                      (same)
    swift-repository-v1           (same)
    rust-v1                       (same)
    rust-repository-v1            (same)

### D2 fixed for the shim and the two local paths, NOT fixed for the external one

The invented shim bodies are gone, replaced by "Execution target contract"
prose. Everything in that prose was reverified, and one path is still invented.

Fixed and confirmed:

- Line 67 now carries both `WriteBinShim` branches: `#!/bin/sh` + `exec` with a
  PATH preamble when path entries are present, relative symlink when they are
  not (`internal/runtimestore/runtimestore.go:137-155`), plus the Windows `.cmd`
  double-expansion note (`WindowsShimContent`, `runtimestore.go:177-205`).
- Line 225 names `$HOME/.curator/runtime/<skill>/<commit>/scripts/run.sh`. I
  installed the document's own Example 3 skill into a scratch project under a
  scratch HOME and read the real shim:

      curator bootstrap --skills-root <scratch>/skills --non-interactive
      curator add script_skill -git ../gitrepo -tag v1.0.0 && curator install
      cat proj/.agents/bin/run-script

      #!/bin/sh
      if [ -n "${PATH:-}" ]; then
        PATH='<proj>/.agents/bin':"$PATH"
      else
        PATH='<proj>/.agents/bin'
      fi
      export PATH
      exec '<scratch>/.curator/runtime/script_skill/3887f117a72397410b50c01046856e42aa081bb1/scripts/run.sh' "$@"

  Shim location, both PATH branches, single-quote quoting, and the runtime store
  target all match the document.
- Lines 68 and 119 name `$HOME/.curator/cache/build/<driver>/<hash>` and
  `$HOME/.curator/cache/build/go-v1/<hash>`. Confirmed:
  `base = filepath.Join(store.home, "cache", "build", buildmeta.DriverGoV1)`
  (`internal/buildcache/cache.go:262`), `DriverGoV1 = "go-v1"`
  (`internal/buildmeta/models.go:21`), and the store home is the manager home,
  `filepath.Dir(config.Path)` = `$HOME/.curator`
  (`internal/config/config.go:153,197`). The scratch install produced
  `<scratch>/.curator/cache/` under exactly that root.

## Defect requiring rework

### D5 (blocking): the external artifact store path is invented

`docs/authoring-cli-commands.md:176` states the external compiled binary lands
in `$HOME/.curator/cache/build_repositories/.../artifacts/<hash>/artifact`.

`cache/build_repositories` is not a directory Curator ever creates. That string
occurs nowhere in the repository except in this document:

    grep -rn "cache/build_repositories" . --exclude-dir=.git
    docs/authoring-cli-commands.md:176:...

`build_repositories` exists only as an `agent-skill.json` manifest key
(`internal/skillspec/parse.go:135,172,417,421,429,491`); it is never a path
segment. The real store root is a sibling of `cache`, not a child of it:

    internal/install/external.go:105:
        deps.StoreRoot = filepath.Join(home, "external-build-cache")

reached with `home = cfg.Home()` (`internal/install/install.go:541`), which is
`$HOME/.curator`. The artifact path is then built at
`internal/install/targets.go:185-186`:

    keyName := strings.TrimPrefix(entry.result.CacheKey, "sha256:")
    finalArtifact := filepath.Join(externalRoot, "artifacts", keyName, "artifact")

and the same root is what garbage collection sweeps
(`internal/scopes/gc.go:125`, `filepath.Join(request.Home, "external-build-cache")`).

So the correct path is:

    $HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact

This is the same class rev1 raised as D2: a filesystem path presented as fact
without being read out of the code. Two of the three paths were re-derived
correctly in this revision and the third was guessed. An author debugging an
external build will look under `~/.curator/cache/` and find nothing there.

Fix: replace the path at line 176 with
`$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact`, and while
there, make the generic mention at line 68 name the same store instead of the
unnamed "build artifact store".

## Everything else reverified and holding

Section 2, `go-v1` admission matrix. Every diagnostic identifier resolves in
`internal/` at a non-test site:

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

Substance checks beyond the identifiers:

- Build vectors are literal at `internal/godriver/build.go:32-36`, including
  `-mod=vendor` and `-ldflags=-linkmode=internal -libgcc=none`.
- "Network access is disabled during compilation" is backed by
  `internal/godriver/session.go:449-450`: `GOPROXY: "off"`, `GOSUMDB: "off"`,
  `GOVCS: "*:off"`, `GOFLAGS: ""`.
- The native-input list (CgoFiles, CFiles, CXXFiles, MFiles, HFiles, FFiles,
  SwigFiles, SwigCXXFiles) matches `graph.go:229-230` exactly, and `CgoFiles`
  is the one that diverts to `cgo_required` (`graph.go:234-236`).
- The assembly clause "outside third-party vendor directories or inside
  replaced modules" is exact: `firstParty := item.Module != nil &&
  item.Module.Replace != nil` (`graph.go:201`) gates it at `graph.go:242`.
- The `//go:cgo_import_dynamic` and `//go:generate` rules match
  `graph.go:276-286`, including the vendor-tree exemption for generators.
- Linux refusal before worker launch: `controls_other.go` is `//go:build
  !darwin && !windows` and `prepareControlDomain` returns
  `diagnostic(CodeControlUnavailable, "the portable execution policy is
  specified for macOS and Windows only")` at `controls_other.go:26`.

Section 3, script commands. `ScriptInterpreters = {"node-v1", "python3-v1"}`
(`internal/skillspec/types.go:29`) and the rev1 note is addressed: line 65 now
gives the code's own rationale (a specification revision, not a manager
configuration option), matching the comment at `types.go:26-28`. The
mutual-requirement note added at line 66 was driven, not assumed:

    halfa (execution_policy, no interpreter)
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.execution_policy: requires 'interpreter'
    halfb (interpreter, no execution_policy)
      error: skill.manifest_invalid (agent-skill.json): commands.run-script.interpreter: requires 'execution_policy'

matching `internal/skillspec/parse.go:876-880`. The `script-worker-v1` refusal
reproduced verbatim:

    curator skill check ./policy_skill
    error: script_execution_policy_unsupported (commands.run-script.execution_policy): this manager does not implement script-worker-v1, and the policy forbids installing the command declared-only, downgrading it, or ignoring the field

Section 4, worked examples. All three manifests were rebuilt verbatim from the
document and validated against the tree binary:

    curator skill check ./embedded_go        exit=0
    curator skill check ./external_go_repo   exit=0
    curator skill check ./script_skill       exit=0

(The one warning each is `skill.command_resolution_contract_missing`, caused by
the placeholder `SKILL.md` I wrote, not by the documented manifests.)

Line 27's claim that `go-v1` and `go-repository-v1` are the only compiled build
drivers holds from an author's position: the `swiftpm*` packages in `internal/`
are not referenced from `internal/skillspec/`, `internal/install/`, or `cmd/`,
so no manifest can select them.

README links the document at `README.md:60`; `docs/build-ssh.md` and
`docs/build-https.md` both exist. No em-dashes, en-dashes, or guillemets in the
document.

## Routing

Status `to-dev`. D5 is a one-line path correction with the exact replacement
given above (plus the optional companion touch at line 68). Everything else in
the document is verified and needs no change. Re-review after the fix.
