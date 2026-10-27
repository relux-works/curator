# Bundled release

The Curator release maintainers own the existing `relux-works/homebrew-tap`
`curator` cask, `relux-works/scoop-bucket` manifest, release deb/rpm files and
the AUR `curator-bin` package. GoReleaser **open source v2.17.0** builds them.

| Platform | Payload | Channel |
| --- | --- | --- |
| macOS amd64, arm64 | `curator`, `curator-run` | Existing Homebrew cask; tar.gz |
| Linux amd64, arm64 | `curator`, `curator-run` | deb, rpm, AUR `curator-bin`; tar.gz |
| Windows amd64, arm64 | `curator.exe` | Existing Scoop bucket; zip |

There is no `ax` executable, separate launcher tap/bucket or winget publication.
Native packages install both Unix commands in the same bin directory. AUR
downloads the Linux release archives; it does not rebuild or re-sign the pair.
The AUR package name/account authorization and `AUR_SSH_PRIVATE_KEY` secret must
be provisioned by the release maintainers before a real stable release. Existing
`TAP_GITHUB_TOKEN` must permit both existing channel repositories.

## Source and artifact gates

`curator-run.json` pins the public launcher repository, annotated tag, commit,
version, Go toolchain and release-key fingerprint. `bundle.py prepare` fetches
only that tag into an isolated temporary repository, verifies its signature
against `RELEASE_SIGNERS`, checks its signed tag name, peeled commit and exact
release key, and only then exports it to `.tmp/curator-run`. Inherited Git
configuration cannot redirect this verification. No fetched code runs before
verification. Neither a lightweight/unsigned tag nor another allowed historical
commit key qualifies as the pinned release signer.

GoReleaser uses the exported module with the launcher's release build flags,
read-only module resolution, pinned Go toolchain and pinned launcher version.
Updating the pin also requires updating those two build settings and rerunning
the snapshot. No launcher repository changes are made here.

Both Darwin binaries use the existing Developer ID certificate and App Store
Connect inputs. Real release CI refuses missing credentials before GoReleaser;
snapshots deliberately skip signing and notarization. The checksum-signing
wrapper inspects all six archives and all four deb/rpm packages before invoking
cosign, refusing missing/empty/nonexecutable Unix companions or an excluded
`ax`. This gate runs before GoReleaser publication. Release checksums and GitHub
provenance cover the resulting paired archives and native packages. The pin and
upstream license/NOTICE accompany the payloads. Package-manager checksums do
not replace source-signer or runtime-provider verification.

Homebrew/AUR/Scoop keep stable channels unchanged for prereleases (`skip_upload:
auto`). Snapshot CI has read-only permissions, no signing secrets and no publish
step. Actual credential availability, Apple signing/notarization, AUR ownership
and published installation remain release-time checks, not snapshot evidence.

## Local checks

Run from a normal Git checkout, with Go 1.25.5, Python 3.9+, SSH tools,
GoReleaser OSS 2.17.0 and Syft on PATH:

```sh
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK python3 -m unittest discover -s release -p 'test_*.py' -v
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK go test -count=1 ./tools/goreleaserconfig/
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK goreleaser check
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK goreleaser release --snapshot --clean --skip=sign,notarize
python3 release/bundle.py artifacts .tmp/dist
python3 release/bundle.py channels .tmp/dist
```

Output stays in ignored `.tmp/`; the source and missing-payload controls use
temporary fixture repositories and disposable SSH keys, never release keys.
The controls exercise wrong/missing tags and commits, unsigned/lightweight tags,
unknown signers, a historical allowed key, renamed signed tags and a removed
launcher before the checksum signer. The snapshot inspects actual package
contents rather than substituting fixture packages for build evidence.

## Follow-ups

This slice changes release payloads and native package channels. `install.sh`
still installs Curator alone, as does `go install`. Installing the pair from the
script and expected-versus-observed `providers.curator-run` version JSON remain
separate work. Bundling does not qualify, register, trust or activate a runtime
provider, and a package-manager upgrade does not grant provider trust.
