# Schema 6 vendored Go skill rollout guide

Task: `TASK-260805-nkfu88` — publish-go-skill-rollout-guide
Audience: agentic-infra skill, CLI, and CocoaSkills maintainers
Status: maintainer guide based on locally accepted artifacts; remote publication and landing are not asserted

## Accepted reference pilot

The reusable reference is the independently accepted `skill-sentry` fallback
pilot, not the earlier `skill-bi` candidate. `skill-bi` reached a real RC3
assembly-policy conflict through its selected `golang.org/x/sys/unix` graph and
must not be presented as a green schema 6 pilot.

| Input or evidence | Immutable identity | Acceptance evidence |
| --- | --- | --- |
| CocoaSkills RC3 | commit `17547365d0a525b1d7b0c2d29f9f7ec792a87c0a`, tree `394a77f11bb4051108910b23bad529cbf77f1311`, wheel SHA-256 `8da896a44000094f0f966a0d90d70bebc456ef59d518a836d99e0e4c9c7d2bfd` | Signed upstream RC3; exact parser, build, audit, toolchain, cache, launcher, and lifecycle contract audited. |
| `sentry-cli` source | accepted environment-token candidate commit `1f39eca3f1edce71dc66bf568e2a1f78187a3a7c`, tree `cec7ce0c17a11df5ce203377ffcf2f725c9d8d67`; derived from authenticated `master@109af705e0025cc9e34f2a483e9a41ab722cdc64` | Go tests, vet, format, offline build, RC3 graph, real help, and missing-token behavior independently accepted. |
| `skill-sentry` candidate | commit `2d19f9117b99dd4b2479347b94940a79b353997a`, tree `c10b69b6e3f061bebd10027feb3d03bf6988dad8` | Independent implementation review accepted. |
| Portable candidate | Git bundle SHA-256 `ef0feaf430a6cc2cc7a61c4fc07241e837367ce5a5136e6f295a17ac92594863` | `git bundle verify` exited 0, records complete history, and exposes `refs/heads/task/TASK-260805-3sgr2e-schema6` at the accepted commit. |
| Vendor closure | `build/vendor/modules.txt` SHA-256 `aa292df0837bb1a53a7f6407e116c75b791fc5046402ce7f71839d411826f933`; four modules | Package tests and RC3 package-graph validation accepted. |
| Black-box E2E | exact RC3 plus exact skill candidate on physical Go `1.25.5 darwin/arm64` | Producer harness passed 4/4 with exit 0; independent reviewer rerun passed 4/4 with exit 0. |

The accepted macOS E2E covers clean install and real `sentry --help`, repeat
cache/currentness, source-changing update, corrupt-cache repair, status, remove,
rollback atomicity, advisory and strict audit behavior, missing/wrong Go,
corrupt vendored source, and Linux fail-closed inventory. Native Windows
execution is still owned by `TASK-260805-1up812` and is not claimed here.

## Exact schema 6 package

### Pilot manifest

This is the accepted `skill-sentry` `agent-skill.json`, byte-for-byte in field
content and values:

```json
{
  "schema_version": 6,
  "build_roots": ["build"],
  "capabilities": {
    "network": ["sentry.io"],
    "filesystem": "none",
    "exec": "none",
    "secrets": ["SENTRY_TOKEN"],
    "env_read": ["SENTRY_TOKEN", "SENTRY_BASE_URL", "SENTRY_ORG", "SENTRY_PROJECT", "SENTRY_TIMEOUT"],
    "prompt_scope": "Read-only Sentry issue, event, release, session, and project analysis using environment-only authentication."
  },
  "commands": {
    "sentry": {
      "type": "build",
      "driver": "go-v1",
      "source_dir": "build/cmd/sentry"
    }
  }
}
```

Do not add package-controlled build programs, arguments, environment, hooks,
tags, linker flags, output paths, or another driver. A build command is legal
only in schema 6 and its driver is exactly `go-v1`. Prefer only canonical
`agent-skill.json`; if legacy `csk-skill.json` is temporarily retained, its
parsed JSON must equal the canonical manifest or RC3 rejects the package.

### Pilot source and runtime layout

```text
skill-sentry/
├── agent-skill.json
├── SKILL.md
├── agents/
│   └── runtime.json              # logical command: {"sentry": "sentry"}
├── build/                        # compile-only; not agent context or runtime copy
│   ├── SOURCE.md                 # immutable upstream/source/vendor identities
│   ├── go.mod                    # module root must be the build root itself
│   ├── go.sum
│   ├── cmd/
│   │   └── sentry/
│   │       └── main.go           # exactly one non-test root package, package main
│   ├── internal/
│   │   ├── cli/
│   │   ├── client/
│   │   ├── config/
│   │   └── output/
│   └── vendor/
│       ├── modules.txt
│       └── <all non-standard dependency source>
├── scripts/
│   └── refresh-sentry-source.sh
└── tests/
    ├── test_package.py
    └── test_rc3_manifest.py
```

Every `source_dir` must be below exactly one declared, link-free build root.
Build roots must be nonempty, unique, disjoint from each other and from runtime
roots, and each must be used. The nearest enclosing `go.mod` must be the direct
`go.mod` of that build root; nested modules, workspaces, root `.` build sources,
and ambiguous roots reject. Every non-standard package must resolve inside the
committed `vendor/` closure.

The skill contains no platform executable. CocoaSkills owns compilation,
immutable cache publication, and the generated `sentry`/`sentry.cmd` launcher.
Skill documentation and runtime metadata must resolve that manager launcher and
must never compile or execute `build/` directly.

## Immutable source pin and vendor refresh

Each migrated skill owns its copied source. Record at least the original skill
base, authenticated CLI base, accepted CLI source commit and tree, relevant
source paths, Go version, and post-refresh `vendor/modules.txt` digest in
`build/SOURCE.md`. A branch name or tag alone is not a source pin.

Use a task-specific refresh script based on this template. Fill in constants
during review and commit the instantiated script; do not accept environment or
caller overrides for the immutable commit.

```bash
#!/usr/bin/env bash
set -euo pipefail

readonly PINNED_COMMIT="<40-hex accepted CLI commit>"
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
readonly SKILL_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd -P)"
readonly BUILD_ROOT="${SKILL_ROOT}/build"
readonly SOURCE_ROOT="${1:-}"

if [[ -z "${SOURCE_ROOT}" ]]; then
  echo "usage: $0 /absolute/path/to/clean/cli-checkout" >&2
  exit 2
fi

if [[ "$(git -C "${SOURCE_ROOT}" rev-parse HEAD)" != "${PINNED_COMMIT}" ]]; then
  echo "CLI checkout must be at ${PINNED_COMMIT}" >&2
  exit 1
fi

if [[ -n "$(git -C "${SOURCE_ROOT}" status --porcelain --untracked-files=all)" ]]; then
  echo "CLI checkout must be clean" >&2
  exit 1
fi

go_version="$(go version)"
if [[ ! "${go_version}" =~ ^go\ version\ go1\.25\.[0-9]+\  ]]; then
  echo "Go 1.25.x is required; found: ${go_version}" >&2
  exit 1
fi

rm -rf -- "${BUILD_ROOT}/cmd" "${BUILD_ROOT}/internal" "${BUILD_ROOT}/vendor"
mkdir -p -- "${BUILD_ROOT}"
cp -R -- "${SOURCE_ROOT}/cmd" "${BUILD_ROOT}/cmd"
cp -R -- "${SOURCE_ROOT}/internal" "${BUILD_ROOT}/internal"
cp -- "${SOURCE_ROOT}/go.mod" "${SOURCE_ROOT}/go.sum" "${BUILD_ROOT}/"
GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go -C "${BUILD_ROOT}" mod vendor
```

Adapt the copied path allowlist when a CLI has a different module layout, but
keep it explicit and minimal. After refresh, review the source diff, run the
full CLI tests, check the RC3 package graph, confirm the module count/digest,
and run integrated lifecycle E2E. Never refresh from a dirty checkout or a
moving ref. Never fetch dependencies during the manager build.

## Audit and security-policy disclosure

RC3 source auditing is opt-in. `audit.enabled` defaults to `false` and the
default mode is `advisory`; therefore a successful schema 6 install is not, by
itself, proof that an audit ran.

When enabled, the audit gate runs before toolchain selection, cache lookup, or
compiler work and covers the entire frozen skill snapshot, including `build/`,
`go.mod`, `vendor/modules.txt`, and all vendored source. Strict backend failure
blocks installation. Advisory backend failure may warn and proceed without an
audit verdict. Static canary and egress-policy failures block even in advisory
mode. Cloud audit requires explicit enablement and obeys source/redaction
policy. Evidence must always record whether auditing was disabled, advisory, or
strict.

Registry attestations authenticate source input. Native-header and artifact
hash validation protect the compiled output, but schema 6 does not audit that
output as a separate subject. Do not substitute artifact validation for source
audit and do not claim RC3 guarantees total network denial, read-only source or
toolchain roots, private-root-only writes, hard descendant resource bounds, or
exact executable allowlisting; those hardening guarantees are deferred.

## Required qualification gates

Run every gate as a standalone command and preserve its real exit code.
Expected-red checks remain failures with their nonzero exits; they prove
fail-closed behavior and must not be reported as passing commands.

| Boundary | Required gate and acceptance |
| --- | --- |
| Manifest/package | Exact RC3 `csk skill check`; package tests assert schema 6, one unambiguous root, canonical/legacy behavior, logical launcher, no committed Mach-O/ELF/PE, exact source pin, and vendor digest. |
| CLI quality | From `build/`: `go test -count=1 ./...`, `go vet ./...`, and `gofmt -l` over first-party Go paths. Any output from the format check rejects. Record coverage without inventing a universal threshold. |
| Closed build | Physical native Go `1.25.x`; `GOPROXY=off`, `GOSUMDB=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`; committed vendor closure; RC3-owned internal-link build. |
| Toolchain negatives | Missing Go, PATH/tool-manager shim, symlinked or repository-local executable, wrong version, and wrong GOOS/GOARCH fail before publication with no partial target. Package content cannot select Go. |
| Platform | Native success is supported only on macOS and Windows. Qualify both before broad rollout. Linux and other hosts must fail closed before a worker or Go compiler starts; never add an ad-hoc Linux bypass. |
| Clean install | Empty manager state and initialized disposable project; compile, publish immutable cache, materialize skill, generate launcher, run real safe `<command> --help`, and record manager, skill, source, toolchain, receipt, cache, marker, launcher, and artifact identities. |
| Cache/currentness | Repeat the identical install; require verified cache hit, stable key/digest, and current `status --check --json`. |
| Update | Change the pinned vendored input deterministically; require a new source identity/cache key and atomic target replacement. Rerun source, vendor, CLI, audit, and lifecycle gates. |
| Repair | Remove or corrupt launcher/cache material; status must report stale or untrusted state and reinstall must restore verified material. |
| Remove and GC | Remove target launcher/marker atomically. Record that unreferenced cache material follows RC3's 24-hour lifecycle grace; do not manually treat cache deletion as package rollback. |
| Atomic rollback | Broken Go update, compile failure, corrupt vendor, strict audit failure, and manager mismatch must leave the previously stable target unchanged and publish no partial launcher/cache/skill target. |
| Audit modes | Record disabled behavior; enabled advisory backend failure may proceed with warning; enabled strict backend failure blocks; canary/egress failures block in either mode. |
| Manager compatibility | Use a physical manager matching the accepted RC3 implementation. RC3 source paired with an older installed manager is an expected fail-closed mismatch, not a supported configuration. |

RC3 owns the closed semantic equivalents of `go list -mod=vendor` and `go
build -mod=vendor -trimpath -buildmode=exe` with internal linking. Skill CI may
run an offline smoke build, but package scripts must not replace or broaden the
manager invocation.

## Remaining-skill rollout

The pilot changes the sequence from the earlier planning draft: `skill-sentry`
is now the accepted direct-Go reference, while `skill-bi` is held at its current
runtime pending a clean dependency-graph decision. Each row has an explicit
risk owner; that owner must open and accept a separate task before mutation.

| Skill | Current classification | Proposed delta | Primary risk and owner | Disposition |
| --- | --- | --- | --- | --- |
| `skill-sentry` | Direct Go; accepted pilot | Publish the accepted env-token `sentry-cli` input and schema 6 skill candidate without behavior drift; preserve logical `sentry` and manager launcher. | Publication/landing identity and pipeline policy — `sentry-cli` and `skill-sentry` maintainers, coordinated by `TASK-260805-1l0u4n`. | Reference pilot; no further behavior delta in publication task. |
| `skill-tracker` | Direct Go with committed platform binary | Pin an accepted `tracker-cli` revision, vendor into `build/`, use schema 6 `tracker`, remove bundled binary, and normalize physical `tracker-cli` versus logical `tracker`. | Mutating/destructive command coverage plus naming compatibility — `tracker-cli` and `skill-tracker` maintainers. | Next direct-Go candidate only after its own source-graph audit and macOS/Windows E2E. |
| `skill-bi` | Direct Go with committed platform binary | Keep the current delivery path until the selected `golang.org/x/sys/unix` assembly dependency is removed without regression or a separate protocol decision is approved; then re-enter the standard template. | RC3 forbids selected assembly and must not be weakened — `bi-cli` and `skill-bi` maintainers; protocol changes belong to CocoaSkills/protocol owners. | Hold; do not call the existing candidate schema 6 compatible. |
| `skill-wiki` | Hybrid Python wrapper plus external Go CLI | Preserve the `wk` Python auth/context wrapper. Separately qualify a pinned `wiki-cli` manager-built `wiki` launcher, update wrapper resolution, and correct stale Grafana/`wb-wiki-cli make setup` docs. | Cross-runtime launcher and auth-context compatibility — `skill-wiki` and `wiki-cli` maintainers. | Separate hybrid pilot, not a mechanical direct-Go conversion. |
| `skill-gitlab` | Python plus system `glab` | Retain Python `gmr` and the system `glab` dependency. Canonical-manifest cleanup, if wanted, is a separate metadata requirement. | Wrapper/system-tool compatibility — `skill-gitlab` maintainer. | Retain current; no Go rewrite. |
| `skill-grafana` | Python/venv REST CLI | Retain the existing Python runtime and venv lifecycle. | Python dependency and credential behavior — `skill-grafana` maintainer. | Retain current; no Go rewrite. |
| `skill-youtrack` | Python wrappers plus pinned PyPI CLI | Retain Python wrappers and `youtrack-cli==0.22.2`. | Wrapper/upstream package compatibility — `skill-youtrack` maintainer. | Retain current; no Go rewrite. |

Non-Go skills remain non-Go unless a separate product and architecture rewrite
is explicitly approved. Repository language, perceived consistency, or the
existence of schema 6 is not such approval. Schema 7 external repositories are
also a separate manager/protocol initiative; never flatten
`build_repositories`, `skill-build.json`, or `go-repository-v1` into schema 6,
because doing so would discard locked-object proof, separate audit subjects,
receipt v2, and marker v3 identity.

## Publication and rollback runbook

Publication changes remote state and is outside this documentation task. The
following order is the handoff contract for the assigned publication owner.

1. Restore authenticated GitLab DNS/VPN, API, Git transport, and registered
   signing identity. Fetch fresh protected default branches.
2. Require the CLI base to equal authenticated
   `109af705e0025cc9e34f2a483e9a41ab722cdc64` and the skill base to equal
   `8ebe1515dc8d167f2f7adbb145e62a6671c8dfe6`. If either moved, recompute
   ancestry and diffs; tree-identical application may proceed, but any behavior
   change returns to implementation and independent review.
3. Apply the accepted env-token CLI patch so the resulting source commit/tree
   is identical to accepted `1f39eca3...`/`cec7ce0c...`. Rerun CLI tests, vet,
   format, closed offline build, and RC3 graph gates before push.
4. Import and verify the skill Git bundle, then apply its accepted tree
   `c10b69b6...` to the fresh skill base without behavior changes. Rerun package,
   vendor, RC3, Go, launcher, and negative gates.
5. Publish and review the CLI MR first, then the skill MR. Require GitLab
   security/signature and repository policy pipelines. Record exact MR commits,
   terminal pipeline IDs, and their relationship to the accepted local trees.
6. After both default-branch landings, repeat the black-box lifecycle run at
   the exact landed commits. Broader rollout remains gated on native Windows
   evidence; macOS acceptance must not be relabeled cross-platform acceptance.

### External landing placeholders

These fields must remain placeholders until `TASK-260805-1l0u4n` records the
authenticated remote facts. This guide deliberately does not guess URLs or
landing state.

| External fact | Value | Owner |
| --- | --- | --- |
| `sentry-cli` MR URL | `<PENDING — TASK-260805-1l0u4n>` | `TASK-260805-1l0u4n` |
| `skill-sentry` MR URL | `<PENDING — TASK-260805-1l0u4n>` | `TASK-260805-1l0u4n` |
| `sentry-cli` protected-default landing commit/status | `<NOT LANDED / NOT VERIFIED — TASK-260805-1l0u4n>` | `TASK-260805-1l0u4n` |
| `skill-sentry` protected-default landing commit/status | `<NOT LANDED / NOT VERIFIED — TASK-260805-1l0u4n>` | `TASK-260805-1l0u4n` |
| Required GitLab pipeline IDs/statuses | `<PENDING — TASK-260805-1l0u4n>` | `TASK-260805-1l0u4n` |

### Rollback

Rollback is a new protected-branch change, not history rewriting and not manual
cache surgery.

1. Stop rollout to additional skills and preserve the failing manager/skill,
   receipt, marker, cache key, artifact hash, platform, toolchain, and pipeline
   evidence.
2. Revert `skill-sentry` first to the last accepted pre-schema-6 protected
   revision. This removes the package's dependency on the new build contract.
   Reinstall/apply that revision and verify the previous logical command path.
3. The env-token `sentry-cli` commit may remain because the installed skill no
   longer consumes it. Revert it separately only for a demonstrated CLI
   regression, using its own tests and review.
4. CocoaSkills schema 6 is backward compatible with earlier skill schemas and
   need not be reverted for a skill-only failure. For a manager regression,
   first revert dependent schema 6 skills, then use the separately owned,
   signed manager rollback process; do not bypass internal signature policy.
5. Let manager lifecycle remove targets and age unreferenced cache entries
   through the documented grace period. Do not edit receipts, markers, cache
   directories, or launchers by hand to simulate rollback.
6. Re-open rollout only after the corrected exact revision passes independent
   review and the complete platform, audit, toolchain, cache, update, repair,
   remove, and atomicity matrix.

No merges, tags, releases, protected-branch writes, manager changes, skill
changes, or schema 7 work are authorized by this guide.

## Evidence map

- `TASK-260805-3sgr2e_validation-evidence.md` and accepted review run
  `RUN-260805-373678`: exact candidate, bundle, vendor, parser, Go, build, and
  launcher evidence.
- `TASK-260805-1d7gk3_results.md` and accepted review run
  `RUN-260805-08cc42`: independent revision-bound lifecycle and negative E2E.
- `TASK-260805-2ukk8d_cocoaskills-rc3-internal-fork-build-contract.md`: exact
  RC3 schema, audit, toolchain, platform, cache, lifecycle, and schema 7
  boundaries.
- `TASK-260805-dkhy7w_agentic-infra-go-skill-topology.md`: seven-skill runtime
  classification and matching CLI ownership.
- `TASK-260805-22kli8_assembly-free-fallback-go-pilot.md`: fallback selection
  and the admissibility distinction that led to the accepted Sentry pilot.
