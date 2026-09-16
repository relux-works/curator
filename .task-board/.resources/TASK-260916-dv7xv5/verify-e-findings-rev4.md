# Verification of E1–E7 against the shipped implementations — rev4 (2026-09-16)

## Rev4 answers verdict rev3

Rev3 review (`TASK-260916-dv7xv5_review-verdict-rev3.md`, verdict **rework**) found
two transcript defects and nothing substantive: the E3 literal output line for
`internal/skillspec/parse.go:697` carried two leading tabs instead of the three
in the source, and the rev3 introduction called `profile.go:260` the usage-error
return (the `return exitUsage` is `:259`; `:260` is the closing brace). Both are
corrected below; every verdict, citation, sibling consequence and bound of rev3
is otherwise unchanged (the E3 block now matches the reviewer's independent
capture byte for byte).

## Rev3 answers verdict rev2

Rev2 review (`TASK-260916-dv7xv5_review-verdict-rev2.md`, verdict **rework**)
required four producer corrections. All four are applied here; verdicts are
unchanged (E1–E4 confirmed, E5 mitigated-in-code with stated limits, E6 split,
E7 confirmed):

1. **E1 citations corrected.** `cmdProfileUpdate` starts at
   `cmd/curator/profile.go:252` (not `:260`; the usage-error `return exitUsage`
   inside the function is `:259` and `:260` is its closing brace). Range selection is `internal/contextresolve/
   contextresolve.go:554-575` (comment "Ranges only: the highest satisfying
   candidate" at `:554-555`), with candidate ordering at `:489`. `:511-515`
   is the exact-commit branch (`if exactCommit != ""` at `:509`) and is
   dropped from the range claim.
2. **E2 literal output + caller classification.** The
   `context_system_module_transitive|SystemModules` search returns exactly
   four lines (not "declarations only"); `internal/envprofile/
   envprofile.go:1138-1139` is the production caller that surfaces
   `SystemModules` as warnings — warning surfacing, not admission. Verdict
   stays confirmed.
3. **E3 seed-write citation replaced.** The actual seed writes are
   `internal/envprofile/managed.go:887-893` (loop over `seeds.files`,
   `os.Remove` then `os.WriteFile` with the payload). `:378-388` renders
   root-context documents via `contextmaterialize.Referenced` and proves
   nothing about seeds — replaced. The 15-line `mcp_servers` output is shown
   literally, including `internal/skillspec/types.go:98` (McpServer type
   comment) with its disposition.
4. **Literal output blocks.** Every quoted command now shows the command, its
   literal output and `exit=` — no paraphrases. E5/E6 limits and the E7
   strict-MCP asymmetry are kept exactly as rev2 states them; E4–E7 verdicts
   unchanged.

Scope: read-only review of `relux-works/curator` `main` @ `80483355` and
`relux-works/curator-agent-launcher` `main` @ `b34e1e27`. Method: `git show
<pin>:<path>` (`| nl -ba` for line numbers) plus greps run from the root of a
temporary `git archive <pin>` extraction (no checkout, no tree modification).
No tests executed, no dynamic reproduction: **static** verdicts throughout —
"confirmed" means the inspected code path contains no mitigation, "mitigated"
means it closes the hole the specification leaves open, and neither is a
runtime observation.

## Per-finding verdicts

| Finding | Implementation site | Verdict | Evidence |
|---|---|---|---|
| E1 signer rule / update delta | `cmd/curator/profile.go`, `internal/envprofile/envprofile.go`, `internal/contextresolve/contextresolve.go`, `internal/pkgversion/pkgversion.go` | **confirmed** (static) | Entry point: `cmdProfileUpdate` (`cmd/curator/profile.go:252`) → `envprofile.UpdateWithPolicy` (`:292`; `internal/envprofile/envprofile.go:887`) → `updateLocked` (`:897`); the command reports only `updated` / `unchanged` (`profile.go:301-305`) — no resolved-version delta, no confirmation gate. Resolution builds candidates from the source's version tags peeled to commits (`contextresolve.go:79-94` `Candidate`/`Candidates`/`ResolveTag`), sorts them (`:489`) and, for ranges only, picks the highest satisfying candidate (`:554-575`); `Resolve` has no signature step. `latest` is the spelling of `*` (`pkgversion.go:286-289`). Signature-related code that does exist is outside this path — see literal transcript E1 below: 13 locations, none reached from `contextresolve` or `envprofile`. |
| E2 transitive `class: system` | `internal/contextmaterialize/contextmaterialize.go`, `internal/envprofile/managed.go`, `internal/envprofile/envprofile.go`, `internal/contextaudit/contextaudit.go` | **confirmed** as specified (static) | `SystemPrompt` (`contextmaterialize.go:248-265`) walks `EmittedOrder(lock, precedence)` — every member of the closure — and appends every `Applicable(pkg, "system", environment)` module with no admission check between the two; the production call is `homePlan.systemPrompt` (`managed.go:1785`). The audit layer classifies system modules as `context-system-module-present` and `Report.Blocking` (`contextaudit.go:109-115`) blocks on `Findings` only, never on `SystemModules`. There is no direct-vs-transitive distinction anywhere on this path; the search returns exactly four lines (transcript E2 below) — three declarations plus the production warning caller `envprofile.go:1138-1139`, which appends each `SystemModules` entry to `warnings` (warning surfacing, not admission). The broad `transitiv\|directly` grep of rev1 matched unrelated prose in `internal/closure/closure.go:1` and `internal/envprofile/gitsource.go:32` and stays withdrawn. |
| E3 codex seed imports `mcp_servers` | `internal/envregistry/envregistry.go`, `internal/envprofile/managed.go`, `internal/mcp/mcp.go`, `internal/contextmaterialize/mcp.go`, `internal/skillspec/types.go` | **confirmed** (static) | The codex adapter declares `Seeds: []string{"config.toml"}` (`envregistry.go:219`); `gatherSeeds` reads the native file whole into `bundle.files` (`managed.go:570-584`); at provisioning the plan writes those bytes unchanged — `if provisioned { for seed, payload := range seeds.files {` with `os.Remove` then `os.WriteFile(..., payload, 0o644)` (`managed.go:887-893`). Nothing strips a table. (Correction: rev2 cited `:378-388` here; that block iterates `contextmaterialize.Referenced` root-context documents into `p.fileHashes`/`p.copies` and proves nothing about seeds.) The other `mcp_servers` code paths are not the seed — see literal transcript E3 and dispositions below: 15 lines; none writes to or filters the seed. |
| E4 umbrella discovery on ambient `PATH` | `cmd/curator/umbrella.go` | **confirmed** (static) | `findProvider` resolves `exec.LookPath("curator-"+name)` on the process `PATH` (`umbrella.go:30-38`); `providerUntrustedDir` (`:43-63`) refuses only the directories the manager itself publishes (`globalbins.Select` + `underDir`); `cmdUmbrella` executes the result with operator argv (`:81-91`). No ownership or writability test of the resolved directory exists on this path, so a project hook's `PATH` entry is accepted. |
| E5 write-through-symlink on takeover | `internal/envprofile/switch.go` | **mitigated in code** for the inspected case; spec rule absent; limits below | The `claude_code` copied root-context write removes the target before `os.WriteFile` (`switch.go:523-530`), the copy fallback removes again (`:539-545`), and `replaceLink` removes before `os.Symlink` (`:694-696`). Limits: this covers the pre-existing foreign-manager link at the root-context target only; it is remove-then-create, not `O_NOFOLLOW`, so a link recreated between the two calls is followed; `writeStoreDocument` still writes with plain `os.WriteFile` (`:686`) and other managed/store writes were not inspected. The rule itself is absent from environments §8.3/§9.5. |
| E6 `path`-kind sources | `internal/contextpkg/contextpkg.go`, `internal/envprofile/managed.go`, `internal/envprofile/switch.go`, `internal/envprofile/envprofile.go`, `internal/contextresolve/contextresolve.go` | **partially confirmed** (static) | MCP half — **not applicable**: dependency declarations accept only `git` with `range\|tag\|revision` (+`directory`, `weight`) (`contextpkg.go:289-305`), MCP members are loaded only from lock members of kind MCP (`managed.go:163-170`), and the MCP source-identity allowlist is applied to every MCP dependency at resolution regardless of the root's kind (`contextresolve.go:477-483`), so a `path` root's git MCP dependencies do not bypass admission and a `path`-kind MCP package cannot exist. System-module half — **confirmed**: module class parsing is kind-agnostic (`contextpkg.go:263`), so a `path` root or overlay carries `class: system` modules. Directory boundary — **confirmed within the inspected flow**: the `path` source is ingested at `envprofile.go:609` (`Source{Kind: KindPath, Path: options.Operand}`) and pinned by state hash (`:928-952`, `:1157`, `:1252`); the marker records no source identity for non-git kinds (`switch.go:760-777`) and the path as provenance (`:793-797`); no ownership/permission/containment validation of the directory was found on that ingestion path. Not inferred from marker fields alone. |
| E7 launcher configuration family and strict-MCP asymmetry | launcher `internal/axconfig/config.go`, `internal/defaults/defaults.go`, `internal/fragment/resolve.go`, `internal/defaults/lineup.go`, `internal/fragment/fragment.go`; curator `internal/envregistry/envregistry.go`, `internal/envprofile/managed.go` | **confirmed** (static), all three minors | (a) Ownership: both loaders `Lstat` the file and on `ModeSymlink` `os.Stat` the target and accept a regular file — links are followed, not refused (`axconfig/config.go:58-74`, `defaults/defaults.go:108-118`; bytes read at `defaults.go:90`); no uid/gid/permission check anywhere (transcript E7 below: permission grep returns `internal/execution/execution.go:197` only, the executable-bit check). (b) Repair-as-persistence: `--repair` is unconditional (`fragment/resolve.go:75-80`) as the SPEC requires; the residual is in S5, not the launcher. The stderr line-group covers model/effort origin only (`lineup.go:271-290`); the resolved provider path is not printed. (c) strict-MCP asymmetry — **confirmed**: the launcher's channel table carries `--mcp-config <path>` with `--strict-mcp-config` for `claude_code` and `-p curator-mcp` for `codex_cli` (`fragment.go:172-173`), matching curator's registry declarations (`envregistry.go:192`, `:215`); combined with the seed half (E3: `envregistry.go:219`, `managed.go:576-583` read + `:887-893` write), a managed claude home runs only the profile's MCP set while a managed codex home runs the seeded native `mcp_servers` plus the profile layer. |

## Literal command transcripts (re-run at the pins by the producer)

Method for each: `git archive <pin> | tar -x -C $TMPD`, command run from
`$TMPD` (repository root equivalent), output paths shown relative to the
root. `git show` line numbers via `git show <pin>:<path> | nl -ba`.

### E1 — signer-related code search (curator @ 80483355)

```sh
grep -rnE 'verify-tag|verify-commit|allowed_signers|gpg|ssh-keygen|Signature|signer' internal cmd --include='*.go' | grep -v _test.go | cut -d: -f1,2 | sort -u
```

```text
internal/buildrepo/pipeline.go:177
internal/buildrepo/pipeline.go:33
internal/buildrepo/pipeline.go:34
internal/install/external.go:24
internal/install/install.go:57
internal/registry/registry.go:181
internal/registry/registry.go:198
internal/registry/registry.go:212
internal/registry/registry.go:223
internal/registry/registry.go:245
internal/registry/registry.go:260
internal/registry/snapshot.go:478
internal/swiftpmsource/git.go:514
exit=0; stderr=''
```

Disposition (unchanged from rev2): `registry.go:181,198,212,223,245,260`
and `snapshot.go:478` validate the **registry record and snapshot signature
envelopes** (registry protocol); `buildrepo/pipeline.go:33-34,177` is the
**build-repository signer policy** (refuses any policy but `none`);
`install/external.go:24` and `install.go:57` are comments on the
protected-store/signer policy of external builds; `swiftpmsource/git.go:514`
labels a `cat-file -e <rev>^{commit}` existence probe "verify-commit" — an
existence check, not a signature check. None is reached from `contextresolve`
or `envprofile`: the update path (`profile.go:252` → `:292` →
`envprofile.go:887` → `:897`) and the range-selection path
(`contextresolve.go:489`, `:554-575`) contain no signature step.

### E2 — system-module search (curator @ 80483355)

```sh
grep -rn 'context_system_module_transitive\|SystemModules' internal --include='*.go' | grep -v _test.go
```

```text
internal/contextaudit/contextaudit.go:104:	SystemModules []SystemModule
internal/contextaudit/contextaudit.go:251:// SystemModules lists every class: system module of a context manifest as
internal/contextaudit/contextaudit.go:253:func SystemModules(packageName string, manifest *contextpkg.Manifest) []SystemModule {
internal/envprofile/envprofile.go:1138:				for _, system := range contextaudit.SystemModules(resolved.Name, manifest) {
exit=0; stderr=''
```

Disposition: `:104` is the `Report.SystemModules` field; `:251-253` is the
constructor that lists every `class: system` module of a manifest (always-warn
surfacing class); `envprofile.go:1138-1139` is the production caller that
appends each entry to `warnings`
(`warnings = append(warnings, contextaudit.ClassSystemModulePresent+": "+...)`)
— warning surfacing, not admission. No `context_system_module_transitive`
diagnostic exists anywhere: transitive system modules are not refused.

### E3 — mcp_servers search (curator @ 80483355)

```sh
grep -rn 'mcp_servers' internal cmd --include='*.go' | grep -v _test.go
```

```text
internal/contextmaterialize/mcp.go:231:// codexMCP renders the TOML layer document whose only table is mcp_servers,
internal/contextmaterialize/mcp.go:232:// one [mcp_servers.<name>] table per server in sorted name order, keys in
internal/contextmaterialize/mcp.go:240:		fmt.Fprintf(&out, "[mcp_servers.%s]\n", server.Name)
internal/skillspec/types.go:98:// McpServer is a dependencies.mcp_servers entry (Spec §5.8).
internal/skillspec/parse.go:514:	if _, present := obj["mcp_servers"]; present && schema < 5 {
internal/skillspec/parse.go:515:		return nil, nil, nil, verr.New("dependencies.mcp_servers", "requires schema_version 5")
internal/skillspec/parse.go:522:		allowed["mcp_servers"] = true
internal/skillspec/parse.go:532:	mcpServers, err := parseMcpServers(obj["mcp_servers"])
internal/skillspec/parse.go:692:		return nil, verr.New("dependencies.mcp_servers", "must be an object")
internal/skillspec/parse.go:695:		label := "dependencies.mcp_servers." + name
internal/skillspec/parse.go:697:			return nil, verr.New("dependencies.mcp_servers", "MCP server names must be non-empty strings")
internal/mcp/mcp.go:253:		servers = data["mcp_servers"]
internal/marker/marker.go:173:	McpServers         map[string][]string   `json:"mcp_servers,omitempty"`
internal/marker/marker.go:253:		"git", "requirements", "mcp_servers", "attestation", "activation", "requirers", "substituted",
internal/marker/marker.go:300:	if _, present := raw["mcp_servers"]; present {
exit=0; stderr=''
```

Disposition: `contextmaterialize/mcp.go:231-240` (`codexMCP`) **renders the
profile's own layer file** `curator-mcp.config.toml` — not the seed;
`skillspec/types.go:98` is the `McpServer` type comment (skill-manifest
`dependencies.mcp_servers` entry, Spec §5.8); `skillspec/parse.go:514-697`
parses skill-manifest `dependencies.mcp_servers` — not the seed;
`mcp/mcp.go:253` (`entriesInFile`, called from `homeEntries`) **reads** a
home's `mcp_servers` for the read-only MCP verification report — not the
seed; `marker/marker.go:173,253,300` is the marker field — not the seed.
None writes to or filters the seed. The seed path is `envregistry.go:219`
(declaration) → `managed.go:576-583` (whole-file read into `bundle.files`) →
`managed.go:887-893` (provisioning-time write of those bytes unchanged).

### E7 — permission-check search (launcher @ b34e1e27)

```sh
grep -rnE 'Sys\(\)|Uid|Gid|Perm\(\)' --include='*.go' | grep -v _test.go
```

```text
./internal/execution/execution.go:197:		if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
exit=0; stderr=''
```

Disposition (unchanged from rev2): the only ownership/permission-adjacent
check in the launcher is the executable-bit check in `execution.go:197` —
neither config loader (`axconfig/config.go:58-74`,
`defaults/defaults.go:108-118`) performs any uid/gid/permission validation.

## Consequences for the sibling stories (recorded on each story)

- `STORY-260916-ioemse` (E1): scope stands; the manager task starts from zero
  on the context/profile path — the registry, build-repository and swiftpm
  signature code is unrelated and must not be mistaken for a base.
- `STORY-260916-2d9coh` (E2): the warning class exists
  (`contextaudit.go:22`); the admission rule slots in between `EmittedOrder`
  and `Applicable` in `SystemPrompt`, plus `Report.Blocking`.
- `STORY-260916-1i1gfo` (E3): the strip-or-report choice lives in
  `gatherSeeds` / the registry seed row; the actual seed writes are
  `managed.go:887-893`; the layer renderer and the read-only verifier are
  unaffected.
- `STORY-260916-2otjbn` (E4): the change is local to `findProvider` /
  `providerUntrustedDir`.
- `STORY-260916-73a5zg` (E5): the spec task writes down what
  `switch.go:523-545` does; the manager task adds the vector plus an
  `O_NOFOLLOW`/atomicity review that also covers `writeStoreDocument`
  (`:686`) and the other managed writes.
- `STORY-260916-wgt8vz` (E6): drop the MCP half (not applicable); keep the
  system-module admission and the directory boundary of the `path` ingestion
  flow.
- `STORY-260916-33vuzm` (E7): ownership and symlink refusal are new code in
  both loaders; the strict-MCP asymmetry is confirmed and belongs to the
  §7.8 table row shared with E3.
