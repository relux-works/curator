# TASK-260906-1xitqi release candidate evidence

## Candidate and source authority

- Audited candidate source: completed `main` commit `7320bc2adbd15aa5ade4e78ef7ef9008274e7478`.
- A fresh `git ls-remote --symref origin HEAD refs/heads/main` and exact-ref fetch both resolved `refs/heads/main` to that same OID. The isolated Story worktree HEAD also equals it before this task's uncommitted changes.
- Successful baseline main CI: https://github.com/relux-works/curator/actions/runs/34015435043. The run is for exact SHA `7320bc2`; Test passed on macOS/Linux/Windows, Race passed on macOS/Linux, and Lint, Naming, Interop, and all three Gate self-test jobs passed. Candidate conformance was correctly skipped because this was a normal push run.
- No commit or tag was created by this task. The parent orchestrator retains publication and signed-commit ownership.

## Exact public-channel versus completed-main inventory

| Surface | Public source/version on 2026-09-06 | Gap to completed main |
| --- | --- | --- |
| GitHub latest stable release | `v0.13.0`, commit `cfffd7cd7be33aba9ca44d26993ce5ab19b5fa4d`, 19 assets | 180 commits behind main; lacks the later schema-7/8, source-closure, credential, and environment-profile deliveries. |
| GitHub prerelease | `v0.14.0-rc.1`, commit `af471504aafd0ea5ea16f712469883ddde8b827e`, 19 assets | 152 commits behind main; supports skill schema 7/marker schema 3, but not main's skill schema 8/marker schema 4 or later capabilities. |
| `v0.14.0-rc.2` tag | Commit `74a24fdbded64381de512dc916022ca9bae2edb3`, 151 commits behind | Tag exists, but no GitHub Release and no Release workflow run exists. |
| `v0.14.0-rc.3` tag | Commit `124654a051f60995e826b75ab43a3c4f3359f5ff`, 150 commits behind | Tag exists, but no GitHub Release and no Release workflow run exists. |
| Homebrew cask | `0.14.0-rc.1` in `relux-works/homebrew-tap` | Installs the schema-7 prerelease, not completed main. |
| Scoop manifest | `0.14.0-rc.1` in `relux-works/scoop-bucket` | Installs the schema-7 prerelease, not completed main. |
| Installer script | Script source is current `main`, but its default uses GitHub's `/releases/latest` | Downloads stable `v0.13.0`, so delivered binary is 180 commits behind main. |
| README Go install | `go install github.com/relux-works/curator/cmd/curator@latest` resolves `v0.13.0` | Reproduced failure, exit 1: released `go.mod` contains a local `replace`; Go refuses versioned install before compilation. |
| Debian/RPM/direct archives | Stable `v0.13.0`; prerelease assets at rc.1 | No rc.2, rc.3, or completed-main binaries/packages. Each published release has six OS/arch archives, four deb/rpm packages, six archive SBOMs, checksums, cosign signature, and certificate (19 assets total). |

The rc.2/rc.3 investigation found **no failed workflow to repair**: Release workflow history has no execution for either tag. They are unpublished tag events, not failed GoReleaser runs. The last release execution was rc.1 and succeeded: https://github.com/relux-works/curator/actions/runs/31059325487. Existing rc.2/rc.3 tags must not be reused or moved.

Completed main adds, beyond rc.1: skill schema 8 with marker schema 4 and policy/script/module-root consumption; portable source-closure adapters for Rust, npm, pnpm, Yarn Classic/Modern, and SwiftPM plus cross-adapter conformance; byte-exact Git acquisition; build-repository SSH/HTTPS credential selection; environment-profile stage (a) install/list/use/update/remove/sync flows; and the accompanying platform, audit, artifact-policy, cache, and CI hardening. Main is therefore the only source with the completed capability set.

## Release-blocker changes

- Replaced the zero pseudo-version plus local submodule `replace` for `tuitestkit` with published `v0.1.1`; refreshed `go.sum`.
- Added `.github/ci/release-source-gate.sh`, which refuses release modules containing `replace` or `exclude` directives and treats a missing/unreadable module as failure.
- The same gate optionally peels the release candidate ref and proves it is an ancestor of a resolved main ref. The Release workflow freshly force-fetches public `origin/main` and passes the pushed tag SHA plus that remote-tracking ref, preventing an unmerged branch tag from publishing.
- Wired the module check into normal CI and both checks into the tag Release workflow before GoReleaser.
- Added negative gate self-tests for `replace`, `exclude`, missing module input, an unmerged candidate, missing main-ref evidence, and half-specified ancestry evidence, plus positive tests for clean module and contained-main cases.

No product behavior was changed.

## Validation and actual exit codes

- Public `go install ...@latest` before the fix: **exit 1**, expected reproduction; Go resolved `v0.13.0` and refused its `replace` directive.
- `bash .github/ci/release-source-gate.sh`: **exit 0**.
- `bash .github/ci/gate-selftest.sh` after the complete gate change: **exit 0**, 91 passed / 0 failed, including all new negative cases and a production-wiring assertion that fresh-main fetch precedes the tag ancestry gate before GoReleaser.
- `bash .github/ci/release-source-gate.sh go.mod HEAD refs/remotes/origin/main`: **exit 0** at the production entry; candidate and freshly resolved main both peel to `7320bc2`.
- `go test -count=1 ./internal/ui`: **exit 0**; proves the real tests compile and execute against published `tuitestkit v0.1.1`.
- `go test -count=1 ./cmd/...`: **exit 0** (`cmd/curator` 394.250s; spec-pin has no tests).
- Broader internal package groups: **exit 1**. All observed failures share a local macOS capture-store error: rename/cleanup beneath deliberately read-only temporary trees returns `permission denied`. Unaffected packages passed, including slow `envprofile`, `godriver`, `install`, and `install/atomicity`. A targeted archive of pristine main `7320bc2` reproduced `TestImmutableAdmittedTreeReplayAndTimeOfUseRechecks` on this Intel Mac with installed Go 1.26.0 (**exit 1**) and the hosted toolchain Go 1.25.5 darwin/amd64 (**exit 1**) with the same rename and cleanup errors. This proves the failure is present without this task's delta and is not Go 1.26 drift; the exact main source remains green in the hosted macOS run linked above. This task does not claim those local groups passed.
- `go build ./...`: **exit 0**.
- `go vet ./...`: **exit 0**.
- bare `golangci-lint run`: **exit 127** because the binary is absent locally; rerun as `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run`: **exit 0**, 0 issues.
- `go mod verify`: **exit 0**.
- `bash -n .github/ci/release-source-gate.sh .github/ci/gate-selftest.sh`: **exit 0**.
- Ruby YAML parse initially used an unsupported Ruby 2.6 keyword and exited 1; rerun without that keyword parsed both workflows with **exit 0**.
- `git diff --check`: **exit 0**.
- Current GoReleaser `v2.18.1` config check: **exit 0**, one configuration validated.
- Current GoReleaser `v2.18.1 release --snapshot --clean --skip=publish,sign,sbom`: **exit 0**. It built darwin/linux/windows amd64+arm64 binaries, six archives, two debs, two rpms, checksums, Homebrew cask, and Scoop manifest. Signing, notarization, SBOM, and publication were deliberately skipped locally and remain workflow-owned.

Two synthetic local-module-proxy attempts were not counted as release validation: one correctly failed checksum-database lookup for the intentionally unpublished tag; the second used a hand-built zip that retained nested fixture `go.mod` files and was rejected by Go's module-zip validator. Real Go proxy/VCS packaging excludes nested modules. The release-source gate plus post-publication `go install` check below are the valid evidence path.

## Version recommendation and parent-owned publication plan

Recommend **stable `v0.14.0`**. `v0.14.0` already has a public rc lineage, no stable `0.14` exists, and semver does not promise compatibility between prereleases. The large completed-main delta is appropriate to stabilize under the already-announced `0.14.0` line; do not mint rc.4 merely to compensate for rc.2/rc.3's missing automation history once reviewer and exact-SHA CI are green.

After review and integration, the parent orchestrator should:

1. Land only the reviewed candidate using the repository's signed-commit path; confirm the resulting main SHA and wait for the full normal CI matrix on that exact SHA.
2. Create a new signed annotated `v0.14.0` tag on that exact SHA and push it. Do not move/reuse rc.2 or rc.3. No tag or release should be created before exact-SHA CI is green.
3. Observe the Release workflow through terminal status. Require success before declaring availability; a missing run is a publication failure, as demonstrated by rc.2/rc.3.
4. Verify GitHub marks `v0.14.0` as latest non-prerelease and publishes exactly the configured set: six archives, six archive SBOMs, four deb/rpm packages, `checksums.txt`, `.sig`, and `.pem` (19/19).
5. Download artifacts; verify GitHub provenance attestations and the cosign checksum signature; verify every checksum and run `curator --version` from representative macOS/Linux/Windows archives expecting `v0.14.0`.
6. Verify GoReleaser updated Homebrew cask and Scoop manifest to `0.14.0`; install/upgrade through both channels and check `curator --version`.
7. Run the default installer in an isolated writable bin directory; verify it resolves GitHub latest to `v0.14.0`, validates the checksum, and reports `v0.14.0`.
8. After Git proxy/checksum propagation, run `go install github.com/relux-works/curator/cmd/curator@latest` in a clean module/cache context; require exit 0 and `go version -m`/`curator --version` to identify `v0.14.0`.
9. Install/query both amd64 and arm64 deb/rpm metadata where runners are available and confirm package version `0.14.0` and `/usr/bin/curator`.

Publication remains explicitly deferred to the parent orchestrator.
