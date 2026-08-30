# TASK-260827-21xw9d Results Summary

## Work Completed

Created `docs/authoring-cli-commands.md` (English, clean prose style) and linked it from `README.md`.

### Key Content & Technical Verification

1. **Placements**: Detailed embedded build roots (`build_roots`) vs external build repositories (`build_repositories`).
2. **`go-v1` / `go-repository-v1` Admission Matrix**:
   - Vendored dependencies (`-mod=vendor`, `vendor/modules.txt`, network proxy/checksum off: `session.go:449-450`).
   - Closed build parameters (`listArguments`, `buildArgumentPrefix` at `build.go:32-36`). Manifest keys beyond `{type, driver, source_dir, modules}` rejected during parsing with `skill.manifest_invalid` (`rejectUnknownBuildFields`, `parse.go:839-855`). Execution boundary defense-in-depth check `build_execution_package_influence_forbidden` (`validatePackageCommandSurface`, `build.go:299-345`).
   - Diagnostic identifiers verified at `internal/` non-test call sites:
     - `cgo_required` (`godriver/graph.go:235`, `build.go:594`)
     - `go_native_input_forbidden` (`godriver/graph.go:237`)
     - `go_pgo_forbidden` (`godriver/graph.go:302`)
     - `go_generator_forbidden` (`godriver/graph.go:286`)
     - `go_forbidden_compiler_directive` (`godriver/graph.go:278`)
     - `go_test_input_forbidden` (`godriver/graph.go:180`)
     - `go_assembly_forbidden` (`godriver/graph.go:243`)
     - `external_link_forbidden` (`godriver/build.go:607`)
     - `libgcc_fallback_forbidden` (`godriver/build.go:609`)
     - `toolchain_switch_forbidden` (`godriver/build.go:466`)
     - `workspace_dependency_forbidden` (`godriver/build.go:454`)
     - `build_execution_control_unavailable` (`godriver/controls.go:15`)
     - `build_execution_package_influence_forbidden` (`godriver/controls.go:18`)
     - `script_execution_policy_unsupported` (`scriptpolicy/scriptpolicy.go:37`)
   - Target execution & store paths verified against source:
     - Local build cache: `$HOME/.curator/cache/build/<driver>/<hash>` (`buildcache/cache.go:262`)
     - External build artifact store: `$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact` (`install/external.go:105`, `install/targets.go:185-186`)
     - Script runtime store: `$HOME/.curator/runtime/<skill>/<commit>/...` (`runtimestore/runtimestore.go:24-26`)
3. **Script Commands**:
   - Admitted interpreters `node-v1`, `python3-v1` (`internal/skillspec/types.go:29`).
   - Mutual requirement for `execution_policy` and `interpreter` (`parse.go:876-880`).
   - Refusal of `script-worker-v1` with `script_execution_policy_unsupported` (`scriptpolicy.go:37`).
4. **Worked Examples**:
   - Example 1: Embedded `go-v1` skill (validated exit 0 with tree binary `go run ./cmd/curator skill check`).
   - Example 2: External `go-repository-v1` skill (validated exit 0 with tree binary `go run ./cmd/curator skill check`).
   - Example 3: Script-command skill (validated exit 0 with tree binary `go run ./cmd/curator skill check`).
5. **Planned Language Drivers**:
   - `kotlin-native-v1`, `kotlin-native-repository-v1`, `swift-v1`, `swift-repository-v1`, `rust-v1`, `rust-repository-v1` (LOGBOOK.md:2849).
   - Validated error: `error: skill.manifest_invalid (agent-skill.json): commands.<name>.driver: must be go-v1 or go-repository-v1` (`parse.go`).

## Rev3 Rework Verification

- **D6 Fix**: Line 40 updated to explain parser rejection `skill.manifest_invalid: field is not supported for build commands` (`rejectUnknownBuildFields`) and execution boundary defense-in-depth `build_execution_package_influence_forbidden` (`validatePackageCommandSurface`). Driven test with `ldflags` manifest confirmed verbatim output match.
