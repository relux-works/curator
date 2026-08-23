# TASK-260729-1vpytz — review verdict cycle 1

**Verdict: ACCEPTED**  
**Reviewed:** 2026-07-29T00:57:29Z  
**Reviewer run:** RUN-260729-694233

## Independent evidence

- Re-fetched the official [downloads JSON](https://go.dev/dl/?mode=json), [installation guidance](https://go.dev/doc/install), [installation-management and removal guidance](https://go.dev/doc/manage-install), [toolchain selection guidance](https://go.dev/doc/toolchain), and [release policy/history](https://go.dev/doc/devel/release). The live catalog exposes stable Go 1.26.5 and Go 1.25.12; the release page records Go 1.25.12 as released 2026-07-07 and the policy supports a major family until two newer major releases exist.
- Revalidated all ten selected Go 1.25.12 URLs by HTTP and matched every artifact name, shape, architecture, kind, and SHA-256 byte-for-byte against official JSON: darwin amd64/arm64 pkg and tar.gz; Windows amd64/arm64 MSI and ZIP; Linux amd64/arm64 tar.gz. Official JSON exposes SHA-256 fields and no separate signature field.
- Confirmed official platform guidance: macOS pkg installs under `/usr/local/go` and creates `/etc/paths.d/go`; Windows MSI defaults to Program Files or Program Files (x86); Linux requires a fresh `/usr/local/go` tree and warns not to untar over an existing tree. Removal guidance matches the packet, including separate Go user config/cache/module/bin state and MSI environment cleanup.
- Confirmed from accepted `go-v1` protocol evidence that Go must be 1.23 or newer **and** from a manager-tested release family; launcher selection is independent of package/PATH, uses the regular `<GOROOT>/bin/go[.exe]`, requires clean GOROOT equality, `GOENV=off`, `GOTOOLCHAIN=local`, and `curator-go-toolchain-v1` full-tree identity. Confirmed implementation evidence has `allowedFamilies = {1.25}` and absolute-root/native-launcher/fingerprint enforcement. Therefore 1.25.12 is compatible; 1.26 is not admitted without conformance and allowlist expansion.
- Fresh read-only macOS inventory reproduced macOS 26.5 arm64; PATH first resolves the 411-byte goenv shim; physical candidates remain goenv 1.25.5, Homebrew 1.25.5, and `/usr/local/go` 1.25.1 with the documented launcher hashes and clean GOROOT results. None is current 1.25.12 or admitted by fingerprint.
- Fresh corrected encoded PowerShell inventory on `ssh win` reproduced Windows 10 10.0.19045 AMD64 / PowerShell 5.1, no `go.exe` from Get-Command or where.exe, no Go segment in process/user/machine PATH, no conventional Go root, no GoProgrammingLanguage registry key, and no Go Programming Language uninstall entry. SSH again emitted the non-PQ key-exchange warning; it is not toolchain evidence.
- Fresh read-only `ssh lev` inventory reproduced Ubuntu 26.04 LTS x86_64, no Go on PATH, no conventional/shallow Go root, no installed `golang*` package, root-owned 0755 system parents, and `apt-cache policy golang-go` candidate `2:1.26~1`. The packet correctly makes no Linux native-qualification claim and rejects that package family for the accepted contract.
- `go test ./...` completed with exit 0 across the repository on the review host. The board outcome and `.research/260729_official-go-bootstrap-guidance.md` are byte-identical (SHA-256 `c28eec5a35e794b3e264d707afcc095a42e8f8179dec45ebba75983b3939ae90`).

## Probe handling and review observations

- The reviewers first Windows attempt stopped after expected `where.exe` not-found stderr was promoted to a PowerShell error; its partial output was discarded. The corrected UTF-8 encoded, non-terminating negative-discovery probe exited 0 and is the accepted Windows evidence. No remote state was changed.
- One first jq projection had an expression-precedence error after all ten URL HEAD checks had already returned 200; that selector output was discarded and a corrected independent JSON projection exited 0.
- Non-blocking evidence caveat: the producer-recorded `git diff --check -- LOGBOOK.md .research/...` exit 0 did not inspect those untracked files. A no-index check reports intentional Markdown hard-break spaces. This is not a content, source, architecture, or consumption defect and is superseded by the independent checks above.

## Acceptance mapping

The packet answers all requested platform, package-shape, root, verification, upgrade/uninstall, security-boundary, inventory, and compatibility questions. It clearly distinguishes discovery from admission; supplies display-only URLs, hashes, exact platform verification commands and cautions; records actual macOS/Windows/Linux observations; maps Go 1.25.12 to the accepted family; and provides direct sections for `TASK-260728-ypbuav` and CocoaSkills Windows qualification. No product, catalog, PATH/profile/registry, package, release, stage, commit, or publication mutation was performed.

## Verdict branch

**ACCEPTED → done.** No changes requested and no stop-the-line boundary. Reviewer supplies no `commit_ack`.