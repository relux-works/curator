# Context installation and skill CLI distribution: visual proposal

Date: 2026-09-10. Task: TASK-260910-16vtxi.

Status: discussion material, not an accepted schema or implemented feature.
This follows TASK-260910-nys3f2 and resolves its binary-distribution ambiguity:
the requested binaries are CLI packages needed by skills, not the Curator
manager executable. No normative specification files were edited.

## Evidence boundary

- Specification checkout: `d019f0e7179520b5c8dcde321c4fe51e04552f58`.
- Adjacent implementation checkout: `683364ce233df872d6cbb194e0e6b205127f5bca`.
- Source inspection, official documentation and interactive explanatory models
  were used. No skill package was installed or compiled for this research.
- No release binary was downloaded, signed, notarized or executed to validate
  native distribution. macOS offline delivery remains an empirical acceptance
  question, not a demonstrated capability.
- The adjacent implementation contains unrelated working-tree changes. They
  were preserved. Relevant implementation files were inspected read-only.
- Visual fragments are response content. Standalone wrappers and headless
  screenshots are temporary QA inputs, not a deployed site or product UI.

## Visuals

`context-installation.html` explores relative path, absolute path, Git and MCP
metadata sources, with Claude Code and Codex output examples. It separates
proposed project declarations from existing managed-home MCP output channels.

`cli-acquisition.html` models current source/cache behavior and proposed
prebuilt admission. Change toolchain presence, artifact validity, fallback
permission and native OS policy to see the resulting route or failure.

The visuals use illustrative package names and URLs. All proposed field names
are provisional and intentionally omit a schema version. They are not valid
examples for the current closed Skillfile schema.

## Changes ready for specification work

| Slice | Concrete contract | Specification locations |
| --- | --- | --- |
| P1: sources and collections | Root-declared relative/absolute paths; shared Git/path source aliases; immediate-child skill collections; immutable snapshots and frozen membership. | New Skillfile schema version; core source identity and markers; manager install plan; project lock decision in ADR 0012; CLI and conformance vectors. |
| P2: typed context | Separate semantic kind, activation and existing root/system class; rules and addressable knowledge; common source selectors; dependency declarations in versioned agent-skill/agent-context metadata. | Core and environments protocols; new context/skill/project schemas; lock identity; adapter capabilities and activation vectors; ADR 0012 integration. |
| P2b: instruction outputs | Authored public/private inputs, explicit composition order, private effective state and output ownership. | Outline an ADR first; project-scope adapter and ownership contracts must precede an automatic writer. |
| P3: MCP import | Typed metadata acquisition, exact descriptor lock, normalization, credential-name bindings and existing managed-home wiring. | Environments protocol and manager profile; project manifest/lock; adapters. Native project config editing is a separate new ownership surface. |
| Skill CLI prebuilt | A distinct artifact acquisition and receipt contract beside source-build drivers. | Core command/provider identity; artifact descriptor, registry subject and receipt versions; manager planning/cache/transaction; native admission profiles. |

### P1 example

```text
project/
├── Skillfile.json
└── agents/
    └── skills/
        ├── review/SKILL.md
        └── docs/SKILL.md
```

```json
{
  "sources": {
    "project": { "path": "./agents" },
    "shared": {
      "git": "https://example.org/toolkit.git",
      "tag": "v1.2.0"
    }
  },
  "skills": [
    { "from": "project", "directory": "skills", "include": ["*"] },
    { "from": "shared", "directory": "skills", "include": ["release"] }
  ]
}
```

An absolute `path` is also valid in the proposed contract. Relative paths are
resolved against the declaring root Skillfile directory, never the current
shell directory. Local bytes include uncommitted/new files; the presence of
`.git` must not switch the meaning to HEAD. The selected contents become an
immutable snapshot, not an implicit live link to mutable authored files.

The project lock records selected member directories, identities and hashes;
Git additionally records the exact commit. A matching hash establishes byte
identity, not availability of another machine's absolute path. Portable source
aliases can have local, uncommitted overrides. The public-lock filename and
private effective-state layout remain decisions; no name is reserved here.

Start with immediate-child collections, literal member names and `*`, with
explicit exclusions. Specify missing selections, malformed members, duplicate
names, symlink escapes, input/output overlap and snapshot limits. Expansion is
atomic for the required set. Launch/status never discovers a new member and
silently changes the installation; install/update/explicit refresh does that.

Root/operator declarations can bind host paths. Downloaded skill metadata may
address contained sibling packages or declared sources; it cannot grant itself
permission to read arbitrary absolute paths on the user's machine.

### P2 example and context-management integration

```json
{
  "rules": [
    { "from": "project", "files": ["rules/*.md"],
      "activation": { "mode": "always" } }
  ],
  "knowledge": [
    { "from": "project", "files": ["knowledge/*.md"],
      "activation": { "mode": "on-demand" } }
  ]
}
```

`kind` answers what a unit means; `activation` answers when it should load;
the existing `class: root | system` answers which instruction channel carries
it. Add versioned fields rather than overloading class or retroactively
reclassifying all old mixed-prose modules as rules. Retain ordering/weight and
lock machinery, while defining source identity for the newly admitted units.

For reusable named exports use versioned `agent-context.json`; for a skill's
machine-readable context dependencies extend `agent-skill.json`. `SKILL.md`
remains the agent-facing document. The same selectors should support project
composition and contained local/Git dependencies without one repo per unit.

Knowledge materializes as documents plus a small title/description/path index.
The agent reads documents as needed; this is not a guarantee of automatic
retrieval. Skill-local reference links must remain resolvable after installation.
Do not concatenate an entire knowledge directory into permanent instructions.
Preserve module bodies as opaque content and specify any generated wrapper.

Native rules are common but not uniform. Claude supports `.claude/rules/`;
Cursor has native rules and separate native skills. There is no universal
`KNOWLEDGE.md` filename established by the environments reviewed. See the prior
task's cited environment survey for broader comparisons.
[Claude memory/rules](https://code.claude.com/docs/en/memory),
[Cursor rules](https://cursor.com/docs/context/rules),
[Cursor skills](https://cursor.com/docs/context/skills).

Current Curator maps Cursor skills to `.cursor/rules` in manager section 10 and
`internal/adapters/adapters.go`; reconcile that mapping with current native
skills when updating the adapter revision. Do not silently flatten a scoped
rule into always-on instructions if an adapter cannot express its activation.

The visual's Codex `AGENTS.md` output is a proposed owned project composition,
not an existing project writer. Claude native generated rule files also need
collision/drift ownership. Typed semantics can be specified now; writing
unmanaged project instruction files must wait for that ownership contract.

### P2b: public/private project instructions

Public source: `agents/context/project.md` or a shared context package.
Private source: explicitly declared `agents.private.md`, or a local override
outside the checkout. Compose public inputs plus private refinements once.
Global context retains its global scope and must not be loaded a second time
through the project output.

Generated `AGENTS.md`/`CLAUDE.md` can be ignored, together with private source
and local effective state. Public locks must omit private contents, private
paths and private hashes. Existing tracked/unmanaged files need an explicit
ownership transition and drift protection; installation must not silently
untrack, overwrite or append into them. Uncommitted content is still visible
to the selected agent/model and is not a credential vault.

Recommend explicit sync as the first lifecycle. Per-launch generation requires
a project-scoped delivery channel or isolation that prevents concurrent
profiles from rewriting the same file. Codex's per-directory override behavior
and Claude's additive local file are not interchangeable implementations of
public + private concatenation. This part remains an outline.

### P3: metadata import is not an MCP connection

An MCP runtime URL is not necessarily a configuration download endpoint.
Accept an explicitly declared metadata format from local/Git sources or a
catalogue URL, normalize inert metadata into `agent-mcp.json`, lock the exact
descriptor bytes and bind credential names locally. The service behind a
remote runtime URL can still change behavior independently of that hash.

The existing protocol already emits Claude
`<home>/.agent-context/mcp/claude_code.json` and Codex
`<home>/curator-mcp.config.toml`; environment launch channels consume them.
See `protocol/environments.md` sections 5.8 and 7.8 and
`internal/contextmaterialize/mcp.go`. The displayed launch flags are facts
about that admitted Curator adapter revision, not a fresh compatibility test
against every newly installed client version.

Project-native writes must merge only owned entries, detect drift and preserve
unrelated configuration. Imports must not execute hooks, `npx`, installers or
servers. The agent host owns connection and OAuth. MCP discovery metadata,
Curator audit policy and CLI/server binary admission remain separate roles.
[MCP Registry](https://modelcontextprotocol.io/registry/about).

Agent Plugins 1.0 provides a portable skills/MCP packaging input, not a universal
rules/knowledge/runtime-install standard. It can become an explicit importer
for representable components without replacing Curator locks or auditing.
[Agent Plugins specification](https://agent-plugins.org/specification).

## How compilation currently works

The current admitted compiled drivers are `go-v1` and `go-repository-v1`.
Other language-driver work exists in the repository, but its authoring guide
still lists those drivers as planned/rejected by manifest parsing. Do not
describe the admitted installer as a general Swift/Rust/Kotlin compiler today.
Source: `../curator/docs/authoring-cli-commands.md`, sections 1 and Planned
language drivers.

A context-only skill with no active compiled command does not enter the Go
toolchain gate: both build planners return early for an empty command set.

| Stage | Current behavior | Missing-environment consequence |
| --- | --- | --- |
| Resolve | Validate source snapshots, closure, commands and audit/registry policy before package-aware build commands. | Required providers or rejected sources block the plan before compilation. |
| Select Go | `CURATOR_GO`, then `GOROOT`, then the implementation's build.Default.GOROOT fallback. The real GOROOT/bin/go must be a regular native executable outside forbidden roots. No PATH search or wrapper acceptance. | Missing/unusable root or executable gives `go_toolchain_missing`; trust/identity failures have separate diagnostics. |
| Probe and fingerprint | Manager-owned empty cwd; telemetry-off, version and selected `go env` probes; native target and tested-family allowlist; toolchain content fingerprint. | Merely having a `go` command or an arbitrarily newer compiler is insufficient. |
| Cache decision | Toolchain identity enters the cache key. Both embedded and external build plans probe before reusable cache lookup. | Missing compiler blocks this install/revalidation route even if build output was cached earlier. |
| Source preflight/build | Closed arguments, vendor-only dependencies, CGO off, GOTOOLCHAIN local, proxy/network/toolchain switching disallowed. No package hooks or code generation. | Unvendored deps, a requested newer toolchain or unsupported native build needs remain blockers; Curator does not automatically provision them. |
| Installed command | Existing Unix shims directly exec the recorded artifact path. | Running an already-installed self-contained CLI does not require Go merely because Go produced it. Install/repair/currentness is a different operation. |

Evidence: `profiles/manager.md` sections 2.1-2.4;
`../curator/internal/godriver/session.go` lines 40-160 and selectToolchain;
`../curator/internal/install/plan.go` lines 365-430;
`../curator/internal/install/external.go` lines 113-134;
`../curator/internal/runtimestore/runtimestore.go` lines 155-172.

A concrete spec/implementation discrepancy was observed: the manager profile
requires support for the tested Go 1.23 family, while the checked implementation
allowlist contains only 1.25. Record and reconcile that contract separately;
neither version string should be presented as a universal installed-host fact.

`go-repository-v1` revision 1 also explicitly excludes manager post-signing,
timestamping and notarization (`protocol/core.md` section 12.2). OS-required
local signing cannot be added as an undocumented source-build fix. Missing
toolchain, source rejection and native execution policy are distinct blockers.

## Proposed prebuilt acquisition

A CLI package declares candidates and exported commands. The operator/project
policy selects `source-only`, `prefer-prebuilt` or `prebuilt-only` and whether a
declared source recipe may be used as fallback. Package-provided metadata must
not install its own trust root or relax verification requirements.

The package declaration should stay compact; platform artifacts belong in a
versioned distribution descriptor, while exact selected identities belong in
the lock and receipt. Illustrative declaration fragment, not current syntax:

```json
{
  "cli_packages": {
    "repo-tools": {
      "version": "2.4.1",
      "commands": ["repo-check"],
      "prebuilt": {
        "metadata": "https://packages.example.org/repo-tools/2.4.1.json"
      },
      "source": {
        "driver": "go-repository-v1",
        "repository": "repo-tools",
        "target": "repo-check"
      }
    }
  }
}
```

The repository alias above would resolve through the existing exact source
declaration/lock; the excerpt does not define a floating source. The final
location of these fields and whether shared CLI providers get their own
package manifest are open design decisions.

### Route and failure contract

1. Resolve required package/command identity and locked candidates; apply
   source/package policy before considering a download.
2. Select an artifact by OS, architecture, CPU baseline, minimum OS and ABI.
   Linux glibc/musl and shared libraries, macOS deployment target and linked
   dylibs, and Windows runtime DLL requirements cannot be reduced to OS/arch.
3. For a prebuilt candidate verify authenticated descriptor, expected producer,
   source/build provenance, freshness/revocation and exact archive/payload
   digests. No local compiler probe is needed in this branch.
4. Check native signing/admission where the OS provides it and validate declared
   runtime requirements. These checks cannot guarantee all future executions.
5. Publish into protected immutable storage atomically with a distinct prebuilt
   receipt and managed shim; preserve the old working installation on failure.
   Lock acquisition mode, final artifact hash and policy/attestation identity.
6. If no compatible candidate is declared in authenticated metadata, a declared
   source recipe may run only when local policy permits it. Toolchain checks
   start here. Trust failure, revocation, expired strict metadata, OS denial or
   a corrupt locked download do not silently trigger a source build.

Transport outages may use an already verified cached artifact with acceptable
freshness or another authorized mirror serving identical bytes. Treat a missing
locked artifact differently from authenticated absence of a platform build.
Any broader network-failure fallback needs a separate explicit policy; the
default should not allow a malicious mirror to trigger compilation.

Start with declared self-contained CLI payloads and bounded archives, no
install scripts, package hooks, privileged installers or automatic dependency
package managers. Validate archive paths, symlinks, size/expansion limits,
executable format and payload digests. Do not execute a downloaded `--version`
as an identity check. A prebuilt receipt cannot impersonate a locally verified
source-build session or inherit its execution-assurance level.

### Benefits, costs and threats

| Change | Benefit | Remaining cost or threat |
| --- | --- | --- |
| Build once in a controlled release pipeline | No compiler/SDK installation on consumer hosts; less setup work and no compilation delay on that route. | Publisher must maintain target matrix, baseline ABI and dependency updates. No timing claims were measured here. |
| Fixed final artifact bytes | Repeatable download/cache/install; shared content cache; simple exact rollback to an admitted version. | A bad or compromised build is distributed consistently to everyone too. |
| Producer signatures and provenance | Bind an artifact to a trusted release identity and declared build inputs. | Compromised trusted CI/keys can produce authentic malicious artifacts; signature is not a source audit or sandbox. |
| Central deny/revocation policy | A newly known bad release can be excluded without editing every Skillfile. | Offline clients cannot learn a new revocation immediately; hard freshness improves control at an availability cost. |
| Per-project pins and shims | Multiple CLI versions can coexist without changing a global package. | Curator now owns its artifacts' security updates, disk lifecycle, rollback and compatibility policy. |

### Cross-platform trust: reuse primitives, define a new subject

Existing `protocol/registry.md` already has Ed25519/CCJ-1 envelopes, pinned
keys, signed snapshots/logs, rollback high-water state, deny-wins revocations,
bounded cached/offline data and out-of-band trust-anchor rotation. Reuse those
primitives; a new crypto algorithm or a new mandatory hosting service is not
necessary for the first binary contract.

However, current audit-record matching accepts either a content hash OR a
matching source identity and commit. That is insufficient to admit a binary:
many different binaries, platforms and build pipelines share one commit.
Define a new versioned binary subject requiring the exact artifact digest and
package/target binding; never let a source-only match satisfy binary admission.

Keep three roles explicit even if one deployment serves all of them:

- Catalogue: which versions/targets and candidate URLs exist.
- Artifact host/mirror: serves bytes; compromise cannot change an admitted digest.
- Trust policy/registry: which producer/build/source/artifact is allowed now.

Suggested signed descriptor facts: package/version; source identity and exact
source/tree hash; target/ABI/minimum runtime; artifact format/size/archive
SHA-256 and payload digests; expected publisher and builder identity; provenance
reference; public native verification identifiers (such as an expected Team ID,
never private signing material or notarization credentials); schema version. Separately signed policy
records provide admission/revocation and freshness. A key included in a
downloaded descriptor is not a new trust anchor.

Release sequence: frozen source -> controlled build -> platform signing and
notarization where applicable -> final packaging/stapling -> final hashes ->
provenance/registry publication -> mirrors. Track unsigned build output and
signed/distributed output as different subjects when checking reproducibility;
native signing/timestamps can change bytes. Never modify a signed artifact
after recording the final digest, or silently re-sign it on the consumer host.

TUF is an established basis for delegated/threshold key rotation, freshness and
rollback/freeze protection. Curator already implements part of this problem;
extending those contracts is not automatically TUF conformance. Use a vetted
implementation if choosing full TUF, rather than inventing a near-compatible
replacement. [TUF specification](https://theupdateframework.github.io/specification/latest/).

Sigstore bundles can be an additional interoperable producer attestation
format. Verify the expected certificate identity and issuer as well as the
blob's signature/digest; a mathematically valid certificate from an arbitrary
workflow is insufficient. Transparency records are not malware verdicts.
[Sigstore verification](https://docs.sigstore.dev/cosign/verifying/verify/).

Curator's admission is a manager policy, not a universal OS Gatekeeper. Current
shims directly exec artifacts and do not perform a registry call on every
invocation. Enforcing later revocations at each managed execution requires a
new launch gate and freshness contract. Direct execution outside Curator can
bypass that gate, and already-running processes/offline clients have explicit
limits. Installation-time validation must not be marketed as runtime isolation.

## Native OS policy and distribution UX

### macOS

Yes, macOS can reject a downloaded CLI even after a Curator signature check.
For a supported public release path, the producer should use Developer ID
Application signing, the required hardened-runtime/signing settings, and
Apple notarization. Curator's Ed25519 signature is an independent layer and
does not replace Apple's trust chain. Successful notarization is an automated
check, not a guarantee of harmless behavior or App Review.
[Apple notarization](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution).

Apple documents that standalone binaries receive notarization tickets, but a
ticket cannot currently be stapled directly to a standalone binary. A ZIP can
be submitted for notarization but cannot itself be stapled. Appropriate app
bundles, disk images and flat installer packages can carry tickets for offline
assessment. This does not establish that extracting a CLI from a DMG into an
arbitrary private store always preserves offline execution admission.
[Apple custom workflow](https://developer.apple.com/documentation/security/customizing-the-notarization-workflow).

Therefore do not promise a universally frictionless single-file/offline path.
Test the actual packaging/download/extraction/launch chain on clean supported
macOS versions and architectures, with normal quarantine, first use online and
offline, and after moving into the protected store. Include codesign/notary
failure, native certificate/ticket revocation where testable, and executable
launch. Verify native diagnostics plus actual behavior; a checksum or one
`spctl` result alone is insufficient evidence. Preserve xattrs/native signatures.

Candidate MVP: a notarized CLI archive with an explicitly supported first-use
online path; offline first-use remains unsupported until validated. A stapled
DMG/bundle/package path is a separate packaging spike. Executing a `.pkg` or
adding system installation hooks would exceed the initial inert-payload
contract. Do not resolve this by clearing quarantine, disabling Gatekeeper,
silently ad-hoc signing, or treating curl's metadata behavior as a security UX.

### Windows and Linux

Windows code signing and SmartScreen/Smart App Control are independent native
checks. A correctly signed new executable can still warn because reputation
is insufficient; EV certificates no longer provide the historical automatic
SmartScreen reputation bypass. Enterprise policy can be stricter. Thus no
cross-platform registry can guarantee prompt-free Windows execution either.
[Microsoft SmartScreen guidance](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation).

Linux has no single universal Gatekeeper-style admission service to substitute
for. Target ABI/runtime compatibility, executable permissions, mount options
and distribution/enterprise execution policy still apply. The binary target
matrix needs more precision than a generic `linux-amd64` label.
For example, RHEL's fapolicyd can enforce separate allow/deny execution rules
and a trusted-file database; an authentic download does not override that policy.
[Red Hat application control](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/security_hardening/blocking-and-allowing-applications-by-using-fapolicyd).

## Existing distribution models

| Model | Mechanism and trust | What Curator should reuse / keep distinct |
| --- | --- | --- |
| Homebrew formula/bottle | Uses a matching prebuilt bottle by default; can build from source under formula constraints. Reviewed formula metadata pins bottle hashes. | Exact target hashes, explicit source fallback, CPU/relocation/runtime baselines. Curator should keep project pins and a closed install contract. |
| Homebrew cask | Distributes upstream application/installer artifacts; a different producer trust and native-policy path from formula bottles. | A useful reminder that a package-manager checksum and upstream native signing establish different properties. Do not equate brew installation with a generic safety guarantee. |
| WinGet | Manifest selects an installer by architecture/type/version, URL and InstallerSha256; installation behavior depends on installer format. | Typed candidates and exact integrity. Native installer side effects and Windows reputation remain separate. |
| APT/Debian | Signed archive Release metadata authenticates Packages indexes and their package hashes. End-user trust is the repository chain, not necessarily a signature from each individual binary author. | Signed indexes, curated releases and key lifecycle. System dependency resolution and global installation ownership are a different scope from project CLI storage. |
| RPM ecosystem | RPM supports checking package signatures against trusted keys with rpmkeys; repository policy and native execution policy are additional layers. | Distinguish package authenticity from permission to execute and from the broader package manager's dependency/installation contract. |

[Homebrew bottles](https://docs.brew.sh/Bottles),
[Homebrew security](https://docs.brew.sh/Homebrew-Security-and-Supply-Chain),
[WinGet manifests](https://learn.microsoft.com/en-us/windows/package-manager/package/manifest),
[APT trust](https://manpages.debian.org/trixie/apt/apt-secure.8.en.html),
[RPM signature verification](https://rpm.org/docs/6.1.x/man/rpmkeys.8).

As of the reviewed Homebrew documentation, bottle provenance verification is
opt-in through `HOMEBREW_VERIFY_ATTESTATIONS`, not universally enabled by
default. Formula post-install steps can run software even for bottles; those
steps and source builds use Homebrew's documented sandbox. Curator's proposed
inert payload installation is a narrower contract.

Recommend two explicit provider ownership modes rather than making Curator a
replacement for every system package manager:

1. Curator-managed CLI: immutable prebuilt/source artifacts, exact project pins,
   receipts and shims, multiple versions in parallel.
2. System-managed dependency: operator-installed brew/WinGet/APT package or
   system command, verified as an external provider under its own contract.
   Curator must not claim a downloaded-artifact receipt or reproducible bytes
   for a mutable PATH command. Automatic package-manager installation would
   require explicit ownership/side-effect policy and is outside this proposal.

## Recommended acceptance sequence

1. Specify P1 source/collection semantics and a minimal exact project lock.
2. Extend typed context and dependency schemas with the same source model;
   preserve root/system and define adapter activation capabilities.
3. Outline instruction ownership/public-private lifecycle, then choose the
   project delivery channel before implementing generated files.
4. Extend MCP acquisition into the existing managed environment machinery;
   separately define project-native write ownership and authentication bindings.
5. Specify CLI prebuilt subjects/receipts and route policy. Run a narrowly scoped
   real macOS/Windows/Linux packaging/admission spike before making native
   installation guarantees or choosing a universal archive format.

The binary work should not block local skills or typed context work. Its
cross-platform primitives can be designed now, while native packaging evidence
and the runtime revocation boundary remain explicit implementation gates.

## Verification and operational notes

Tool readiness is recorded under this task's scratch directory. Chrome was
launched headlessly with a separate temporary profile through an existing
Playwright installation; no interactive browser was focused and no new tool
was installed. See `visual-qa-01.json`, passing `visual-qa-02.log`, screenshots and
`verify-visuals.cjs` for layout/interaction verification.

The first layout pass found inline-code overflow at 320px; the code block
layout was corrected. The final pass verified both explainers at 736, 360 and
320px in light and dark themes, source/agent/tab changes, missing toolchain,
permitted and forbidden source fallback, integrity failure and native OS denial.
No script errors or horizontal overflow remained in those checks.

Some initial read-only searches used nonexistent candidate paths/globs
(`adapters/*.md`, `profiles/agent-environments.md`, `specification.md`,
`internal/shim`, `runtimestore/targets_unix.go`). They produced no evidence;
subsequent reads used discovered files. Apple documentation's Markdown route
did not return usable content, so the official DocC JSON endpoint was read;
the extracted text is retained in `apple-*.txt`. None of these failures blocked
or weakened an implementation validation claim.
