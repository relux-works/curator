# TASK-260827-21xw9d review verdict: changes requested

Reviewer run `RUN-260828-c4e4c4`, 2026-08-28. Change Request
`CR-TASK-260827-21xw9d-1` revision 1, base `41ab53cd`, candidate `91e2abf5`.
Scope reviewed: `docs/authoring-cli-commands.md` (new, 258 lines) and the one
README link. The delta touches no `.go` file (`git diff --name-only ... | grep
-c '\.go$'` = 0), so no suite was rerun; the sibling docs in the same CR belong
to other story tasks and were not re-reviewed here.

Verification binary: `go build -o .temp/review-21xw9d/curator ./cmd/curator`
from the candidate tree.

## What holds

Most of the document is accurate and was reverified independently.

Section 2, the `go-v1` admission matrix: every diagnostic identifier except one
(see D1) resolves in `internal/` with a non-test call site. `cgo_required`
(`internal/godriver/graph.go:235`), `go_native_input_forbidden` (`graph.go:237`),
`go_pgo_forbidden` (`graph.go:302`), `go_generator_forbidden` (`graph.go:286`),
`go_forbidden_compiler_directive`, `go_test_input_forbidden` (`graph.go:180`),
`go_assembly_forbidden`, `external_link_forbidden` (`build.go:603`),
`libgcc_fallback_forbidden` (`build.go:609`), `toolchain_switch_forbidden`
(`build.go:466`), `workspace_dependency_forbidden` (`build.go:457`),
`build_execution_control_unavailable` (`controls.go:15`, raised at
`controls_other.go:26,30` with "the portable execution policy is specified for
macOS and Windows only"). The build vectors quoted in the "closed build
parameters" bullet match `internal/godriver/build.go:32-36` literally,
including `-mod=vendor` and `-ldflags=-linkmode=internal -libgcc=none`.

Section 3, script commands: `ScriptInterpreters` is
`{"node-v1": true, "python3-v1": true}` at `internal/skillspec/types.go:29`, and
the "every shell identifier is deliberately absent" claim is the code comment
above it. The `script-worker-v1` refusal was reproduced end to end, not just
grepped:

    curator skill check .temp/review-21xw9d/policy_skill
    error: script_execution_policy_unsupported (commands.run-script.execution_policy): this manager does not implement script-worker-v1, and the policy forbids installing the command declared-only, downgrading it, or ignoring the field
    exit=1

Section 4, worked examples: all three manifests were rebuilt verbatim from the
document and validate.

    curator skill check .temp/review-21xw9d/embedded_go        exit=0
    curator skill check .temp/review-21xw9d/external_go_repo   exit=0
    curator skill check .temp/review-21xw9d/script_skill       exit=0

(The one warning each is `skill.command_resolution_contract_missing`, caused by
the placeholder `SKILL.md` I wrote, not by the documented manifests.)

Section 5, planned drivers: the quoted error is exact. Substituting each of
`kotlin-v1`, `swift-v1`, `rust-v1` into a valid manifest yields, verbatim:

    error: skill.manifest_invalid (agent-skill.json): commands.mytool.driver: must be 'go-v1' or 'go-repository-v1'
    exit=1

README links the document at line 60. No em-dashes, en-dashes, or guillemets.

## Defects requiring rework

### D1 (blocking): `package_build_command_influence_forbidden` does not exist

`docs/authoring-cli-commands.md:40` names the diagnostic
`package_build_command_influence_forbidden`. That string appears nowhere in the
repository except in this document:

    grep -rn "command_influence" --include='*.go' --include='*.md' .
    docs/authoring-cli-commands.md:40:...

The real identifier is `build_execution_package_influence_forbidden`
(`internal/godriver/controls.go:18`), raised at `internal/godriver/build.go:302,
318, 322, 325, 328, 337, 341`. The substance of the bullet is right: extra keys
on the build-command object are rejected, and `ldflags`, `gcflags`, `env`,
`hooks`, `pre_build`, `post_build` are all literal keys in
`packageInfluenceSurfaces` (`build.go:106-129`). Only the identifier is
invented. The task's own results resource
(`TASK-260827-21xw9d_results.md`) cites the correct constant
`CodePackageInfluenceForbidden` at `build.go:300-320`, so the document
contradicts the evidence filed to support it. The acceptance criterion is
"every constraint grep-verified against internal/ with literal evidence"; this
one is not.

Fix: replace with `build_execution_package_influence_forbidden`.

### D2 (blocking): the three "resulting shim" blocks are invented output

The document presents a concrete shim body for each worked example. None of the
three matches what `curator install` writes. I installed the document's own
Example 3 skill into a scratch project and read the file:

    curator add script_skill -git <path> -tag v1.0.0 && curator install
    cat <project>/.agents/bin/run-script

    #!/bin/sh
    if [ -n "${PATH:-}" ]; then
      PATH='<project>/.agents/bin':"$PATH"
    else
      PATH='<project>/.agents/bin'
    fi
    export PATH
    exec '<home>/.curator/runtime/script_skill/04ee1429.../scripts/run.sh' "$@"

Against `docs/authoring-cli-commands.md:231-236`, three things are wrong.

The PATH preamble is missing entirely. `UnixShimContent`
(`internal/runtimestore/runtimestore.go:161-172`) emits it whenever
`pathEntries` is non-empty, which is the normal project-install case.

The target directory is wrong. The document points the shim at
`.agents/skills/script_skill/scripts/run.sh`, inside the project skill tree.
The real target is the manager runtime store,
`$HOME/.curator/runtime/<skill>/<commit>/<declared path>`
(`runtimestore.Dir`, `runtimestore.go:24-26`, used by `RuntimeCommandPath` at
`runtimestore.go:97-104` and reached through `PrepareScriptRuntime` at
`internal/install/targets.go:159-161`). An author who follows the document will
look in the wrong place when a command misbehaves.

The quoting is wrong. `shellQuote` (`runtimestore.go:208-213`) emits single
quotes; the document shows double quotes.

The same applies to the two compiled examples at lines 115-122 and 177-184.
`.curator/cache/builds/mytool` is not a path Curator ever creates. The local
build cache is `<home>/cache/build/<driver>/<sha256-hex>` (`buildcache.paths`,
`internal/buildcache/cache.go:262-263`): the directory is `build`, singular,
carries a driver segment, and is content-addressed by cache key, not named
after the command. The external artifact lands at
`<externalRoot>/artifacts/<sha256-hex>/artifact`
(`internal/install/targets.go:185-186`).

Fix: either reproduce a real shim (install one skill and paste the output, as
above) or drop the literal bodies and describe the contract in prose, naming
`.agents/bin` for the shim and the runtime store / build cache for the target.
Note also that the unconditional claim at line 66, "on POSIX systems, shims use
`#!/bin/sh` with `exec`", has a second branch: with no `pathEntries`,
`WriteBinShim` writes a relative symlink instead (`runtimestore.go:147-155`).

### D3 (blocking): the `.git` suffix is not a validation rule

`docs/authoring-cli-commands.md:162` states "The `git` URL must end with the
`.git` suffix." It does not. I removed the suffix from the document's own
Example 2 manifest and revalidated:

    "git": "https://github.com/example/remote-tool"
    curator skill check .temp/review-21xw9d/nogit    exit=0

`buildrepo.ParseSource` trims the suffix for identity derivation
(`internal/buildrepo/buildrepo.go:127`, `strings.TrimSuffix(repoPath, ".git")`)
and never requires it. The real rule already exists in this repository, at
`docs/build-https.md:140`: include `.git` for an HTTPS address because the
service may answer with a `301` and Curator's protected fetch rejects
redirects. That is transport-specific operational advice, not a manifest
constraint, and stating it as "must" sends an author looking for a parse error
that will never fire.

Fix: restate as the HTTPS redirect guidance and cross-reference
`docs/build-https.md`.

### D4 (blocking): the reserved Kotlin driver identity is `kotlin-native-v1`

`docs/authoring-cli-commands.md:250` names the planned Kotlin driver
`kotlin-v1`. The project's own record reserves a different identity.
`LOGBOOK.md:2849`: "Six closed identities reserved: `rust-v1`,
`rust-repository-v1`, `swift-v1`, `swift-repository-v1`, `kotlin-native-v1`,
`kotlin-native-repository-v1`." `swift-v1` and `rust-v1` in the document are
correct; only Kotlin is wrong. The same entry records that reservation is
explicitly not admission, which is worth a clause since the section's whole
point is the gap between reserved and shipped.

Fix: use `kotlin-native-v1`, and consider naming the `-repository-v1` halves
since the document already distinguishes `go-v1` from `go-repository-v1`.

## Non-blocking notes

`docs/authoring-cli-commands.md:65` attributes the absence of shell
interpreters to "cross-platform behavior". `internal/skillspec/types.go:26-29`
gives a different reason: admitting one is a specification revision rather than
a manager configuration option. Prefer the code's rationale.

Section 3 never mentions that `execution_policy` and `interpreter` are mutually
required (`internal/skillspec/parse.go:877-880`). An author copying the schema-8
snippet with only one of the two gets a parse error the document does not
predict.

`docs/authoring-cli-commands.md:27` calls `go-repository-v1` "its external
variant". The two are distinct closed driver identities validated separately
(`parse.go:347` and `parse.go:365`), and the document's own Example 2 treats
them as such.

## Routing

Status `to-dev`. D1 through D4 are localized text corrections against evidence
already cited above; no structural rewrite is needed. Re-review after the fix.
