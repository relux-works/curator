# CIP-0006 evidence: legacy provider settings and MCP opt-outs

Research outcome for TASK-261004-2iewnz — research-105-design-and-implementation-plan, under STORY-261004-3v2zm2 — legacy-provider-settings-and-mcp-optouts-105. This accompanies the [CIP draft](261004_CIP-0006-legacy-provider-settings-and-mcp-optouts.md). It records observations separately from proposed behavior. No product code, tests, or LOGBOOK.md were changed by this research.

## E1. Source identity and bounds

Observed 2026-10-04:

| Read | Result | Exit |
|---|---|---:|
| `git rev-parse HEAD` | `ca1b776fb580ec0cee0173bf150daf063023aeaa` | 0 |
| `git rev-parse main` | Same commit | 0 |
| `git ls-remote https://github.com/relux-works/curator.git refs/heads/main` | Same advertised main | 0 |
| `git ls-remote --tags https://github.com/relux-works/curator-spec.git 'refs/tags/*rc.14*'` | Tag object `661bead088186db70c9f57ad0b115305ec2bef35`; peeled commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7` | 0 |
| Clone the public spec into scratch, `--depth 1 --branch v1.0.0-rc.14`, then `git rev-parse HEAD` | Clone exited 0 with an annotated-tag warning; HEAD equals the peeled commit above | 0 / 0 |
| `gh issue view 105 -R relux-works/curator --json number,title,body,url` | Issue title and acceptance read directly | 0 |

The issue's earlier reproduction used `2cb29dac…`; this research reran the refusal on the newer main above. The local worktree's initial status was clean. Curator's CI itself pins the inspected rc.14 commit: [`.github/workflows/ci.yml:35–43`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/.github/workflows/ci.yml#L35-L43). The source comparison is a research citation baseline, not integration authorization or release qualification.

All `C` references below mean the pinned curator commit; all `S` references mean the peeled rc.14 spec commit. No raw operator configuration, credential file, token, Keychain secret, other operator's product material, private host, or personal path was used as evidence. Scratch paths in commands/output are normalized to `$CIP_SCRATCH` or `<scratch>`; only synthetic fixture paths were normalized. No provider process or authenticated model turn was run.

## E2. Issue and specification facts

| Ref | Source | Verified fact / design implication |
|---|---|---|
| I1 | [Issue #105](https://github.com/relux-works/curator/issues/105) | Four Claude permission paths, plugin booleans, four Codex paths and explicit per-server MCP negatives are the concrete inventory. Models/engines and other migration work are excluded. Unsupported input must not be silently dropped. |
| S1 | [Decision 0014:3–10,40–89](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0014-tool-configuration-surfaces.md#L3-L10) | Status is proposed, not adopted. Its options distinguish machine settings, package declarations and provisioning-only state. It explicitly leaves ownership/merge/permission policy decisions open. |
| S2 | [Decision 0018:175–240](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0018-curator-run-permission-interface.md#L175-L240), [adoption choices:324–330](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0018-curator-run-permission-interface.md#L324-L330) | Permission mode and lock transport are adopted; provider flag grammar belongs to agents-management. Stored settings and launcher permission mode cannot be treated as the same field. |
| S3 | [Environments §1:35–46](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L35-L46), [§2:379–420](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L379-L420) | Current packages exclude settings and schema 1 rejects unknown members. Option A deliberately preserves that boundary. |
| S4 | [§6:1143–1237](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1143-L1237), [§9.3:2523–2538](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2523-L2538) | Overlays are per-profile machine declarations. Current scopes are adapters/targets, not cwd projects. Weights order chapters; MCP/skills resolve jointly. |
| S5 | [§7.4:1610–1708](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1610-L1708) | Claude has a written minimal `.claude.json`, no settings seed. Codex seeds once; A and B differ on inherited MCP. Old homes keep their historical bytes. |
| S6 | [§7.8:1826–1857](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1826-L1857), [§5.8:1110–1141](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1110-L1141) | Claude strict MCP suppresses other sources; Codex uses one reserved layer and retains A-seeded base servers. Empty positive sets currently mean no channel. |
| S7 | [§8.2:1979–2071](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1979-L2071), [§8.3–8.4.1:2073–2239](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2073-L2239) | Marker is a closed ledger, not authority proof; ownership, no-follow writes and unreadable-versus-absent rules constrain a settings extension. |
| S8 | [§9.1–9.2:2258–2521](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2258-L2521), [§9.6:2771–2867](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2771-L2867) | Install/update/switch/retention and takeover are distinct operations. Native import's detected list is context and skills, not arbitrary provider files. |
| S9 | [§10.1:2901–3134](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2901-L3134) | Read-only resolve verifies first; stale refuses without repair; same-user pre-exec race is a residual. Existing `curator run` owns launch composition; native mode is not a strict permission posture. |
| S10 | [§10.2–10.3:3155–3351](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3155-L3351), [§12.1–12.2:3672–3799](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3672-L3799) | Fragment revisions are closed; no private ad-hoc machine knobs; system lock semantics and profile path boundary already exist. |

Environments §§7–10 were read, including the credential, secondary-target, onboarding, import and race boundaries. The CIP cites only relevant constraints and does not reproduce unrelated historical operator narratives in those sections.

## E3. Current implementation map

| Ref | File:line on curator main | Observation |
|---|---|---|
| C1 | [`internal/contextpkg/contextpkg.go:104–117,168–255`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/contextpkg/contextpkg.go#L168-L255) | Manifest has modules and positive contexts/skills/MCP requirements; strict keys exclude settings/policy. |
| C2 | [`internal/config/environments.go:270–275,395–403,617–627`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/environments.go#L270-L275) | `permissions` accepts only native/yolo; knob and overlay member lists are closed. |
| C3 | [`cmd/curator/envconfig.go:58–102`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/envconfig.go#L58-L102), [`internal/config/environments.go:1351–1363`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/environments.go#L1351-L1363) | Existing public config mutation validates before atomic write. CLI value parsing uses ordinary `json.Unmarshal` and string fallback: a strict new record needs an explicit duplicate-key/error path. This is a code observation, not an adversarial test result. |
| C4 | [`internal/envprofile/overlays.go:15–78`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/overlays.go#L15-L78), [`cmd/curator/compose.go:95–130`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/compose.go#L95-L130) | Declaration edit and closure update are separate; path overlays also pass protected-source checks. |
| C5 | [`cmd/curator/env.go:123–181`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/env.go#L123-L181), [`internal/envprofile/managed.go:2389–2437`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L2389-L2437) | Production entry supplies launch directory; profile selection is named → env current → machine current. No project overlay selector is hidden in this call chain. |
| C6 | [`internal/envprofile/managed.go:288–357`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L288-L357), [`:2258–2317`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L2258-L2317) | `assembleHome` and `buildFragment` are the production integration points. Permission mode and MCP union are built independently. |
| C7 | [`internal/contextmaterialize/mcp.go:95–127,214–238,300–319`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/contextmaterialize/mcp.go#L214-L238), [`internal/envprofile/managed.go:204–223`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L204-L223) | MCP is filtered only by environment; names derive from lock members. Codex renderer emits only command/args or URL, no explicit enabled=false. |
| C8 | [`internal/envregistry/envregistry.go:28–37,242–288`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envregistry/envregistry.go#L28-L37), [`internal/envprofile/managed.go:823–895`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L823-L895), [`internal/envprofile/status.go:774–795`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/status.go#L774-L795) | Shipped seed selector is A. Code implements B but does not select it. Repair does not refresh a copied seed; status distinguishes historical seed records. |
| C9 | [`internal/envmarker/envmarker.go:27–51,307–315`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envmarker/envmarker.go#L307-L315), [`internal/envfragment/envfragment.go:23–70,107–136`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envfragment/envfragment.go#L23-L70) | Current marker supports v1–v3 and four surface keys. Ordinary adapters emit fragment v2; Muse emits v3. Proposed v4 tokens are therefore new, not current capability. |
| C10 | [`internal/envprofile/profiledelta.go:34–105`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/profiledelta.go#L34-L105), [`internal/envprofile/import.go:101–180`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/import.go#L101-L180) | Existing update confirmation inspects system modules/MCP declarations, not settings. Native onboarding import reassembles context/skills. Option B would need additional package-delta work; A avoids claiming that existing gate covers settings. |
| C11 | [`internal/config/config.go:25–35,358–361`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/config.go#L25-L35) | Machine schemas 1–3 are readable; system schemas 1–2. New knob must not be inserted into a frozen legacy shape. |

## E4. External adapter corroboration and limits

These are current official documentation observations, **not pinned native execution evidence**:

- [OpenAI configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference), fetched 2026-10-04 via the official Codex documentation URL: cap is numeric, status-line is ordered string items, tier is a native string, and MCP has an enabled boolean. Searching the fetched page found no `status_line_use_colors` occurrence. That bounds this documentation search; it does not prove the installed/pinned tool lacks the setting. L2 must verify it.
- [Claude settings precedence](https://code.claude.com/docs/en/settings), fetched 2026-10-04: project/local/user are different scopes, permission lists combine, and managed settings retain their precedence. This motivates explicit scope records and rejecting unproved flattening.
- [Claude permission rules](https://code.claude.com/docs/en/permissions), fetched 2026-10-04: slash-prefixed patterns depend on source scope, and some negation behavior depends on source-local order. The proposed bounded compiler/refusal rule is a design response, not a claim that Curator currently translates these patterns.

The first broad OpenAI search returned API service-tier material; it was not used as evidence of Codex configuration. Opening the direct official configuration reference supplied the relevant source. The Claude all-settings link returned an internal retrieval error and was not treated as verified evidence. No release-native behavior, provider permission enforcement or actual plugin activation was measured here.

## E5. Commands actually run and results

All gates/probes below ran as standalone processes, without `tee` or a pipeline. The small source-reading commands are navigation, not validation gates. `$CIP_SCRATCH` is a disposable directory outside the worktree, created with `mktemp`; its random suffix is intentionally not retained in public evidence.

### Build and selected baseline tests

```sh
go build -o "$CIP_SCRATCH/curator-proof" ./cmd/curator
```

Exit **0**. The build used the current worktree source at the main commit recorded in E1; no product source was edited.

```sh
go test ./internal/contextpkg ./internal/contextmaterialize -run 'TestUnknownFieldIsRejected|TestMCPGate|TestMCPFileEmptySetWritesNothing|TestCodexTOMLExactness|TestClaudeMCPJSONShape' -count=1 -v
```

Exit **0**. Observed output:

```text
=== RUN   TestUnknownFieldIsRejected
--- PASS: TestUnknownFieldIsRejected (0.00s)
=== RUN   TestMCPGate
--- PASS: TestMCPGate (0.00s)
PASS
ok  github.com/relux-works/curator/internal/contextpkg  0.879s
=== RUN   TestMCPFileEmptySetWritesNothing
--- PASS: TestMCPFileEmptySetWritesNothing (0.00s)
=== RUN   TestCodexTOMLExactness
--- PASS: TestCodexTOMLExactness (0.00s)
=== RUN   TestClaudeMCPJSONShape
--- PASS: TestClaudeMCPJSONShape (0.00s)
PASS
ok  github.com/relux-works/curator/internal/contextmaterialize  1.643s
```

Coverage is **5/5 named tests matched and green**, not a full suite or feature gate. No result was reused from a previous task outcome. The issue's historical test report was read as context only. No mutants were applied to product source during read-only research; the CIP names the implementation mutants to run later.

### Synthetic production-entry probes

Three isolated fixtures had private directories and regular manifest files. Each used a schema-2 user config with an absolute scratch `skills_root` and `projects: {}`, plus a separately written system config `{"schema_version":2}`. No context modules, MCP commands, credentials, or auth/session state were supplied.

Base manifest:

```json
{"schema_version":1,"name":"legacy-example","version":"1.0.0"}
```

The `settings` fixture added:

```json
{"settings":{"claude_code":{"permissions":{"allow":["Read"],"deny":["Bash"],"ask":[],"defaultMode":"acceptEdits"}}}}
```

The `mcp-policy` fixture added the hypothetical member below. It is a negative schema probe, not an existing or recommended package API:

```json
{"mcp_policy":{"claude_code":{"example-server":"disabled"}}}
```

For each case, run from that fixture's `work` directory, with `<case>` equal to `settings`, `mcp-policy`, or `control`:

```sh
env -i HOME="$CIP_SCRATCH/<case>/home" PATH=/usr/bin:/bin \
  CURATOR_CONFIG="$CIP_SCRATCH/<case>/config.json" \
  CURATOR_SYSTEM_CONFIG="$CIP_SCRATCH/<case>/no-system.json" \
  "$CIP_SCRATCH/curator-proof" profile install "$CIP_SCRATCH/<case>/package"
```

| Probe | Real exit | Observed diagnostic / result | Interpretation |
|---|---:|---|---|
| Settings member | **1** | `profile_source_invalid: context_manifest_invalid: unknown field settings` | Expected failing command; rc.14 has no package settings member. |
| MCP policy member | **1** | `profile_source_invalid: context_manifest_invalid: unknown field mcp_policy` | Expected failing command; positive MCP dependencies do not admit this policy object. |
| Plain package control | **0** | Installed and activated `legacy-example`; four scratch in-place adapter entries switched | Positive control confirms the package/profile pipeline is reachable. |

All cases printed the existing permissive-security-posture warning. The successful control also printed the empty MCP allowlist warning. These were warnings, not suppressed failures.

A separate Python assertion process exited **0**: for both failing cases, no `profiles` directory was published, no environment marker existed in the synthetic home, and the user config still equaled its original synthetic input. Mutation-lock bookkeeping files existed; therefore the claim is **no profile/config/surface publication**, not zero filesystem writes. For the control, `profiles/legacy-example/lock.json`, `source.json` and `profiles/current` existed with the expected selection. An initial exploratory inventory searched for the wrong lock basename; the assertions used the actual `profiles/<name>/lock.json` location and supersede that inventory.

### Real empty-MCP fragment

```sh
env -i HOME="$CIP_SCRATCH/control/home" PATH=/usr/bin:/bin \
  CURATOR_CONFIG="$CIP_SCRATCH/control/config.json" \
  CURATOR_SYSTEM_CONFIG="$CIP_SCRATCH/control/no-system.json" \
  "$CIP_SCRATCH/curator-proof" env resolve claude_code \
  --profile legacy-example --repair --format json
```

Exit **0**. Stdout, with only the scratch path normalized:

```json
{"env":{"CLAUDE_CONFIG_DIR":"<scratch>/control/environments/legacy-example/claude_code"},"environment":"claude_code","fragment":"launch-env-fragment-v2","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"ce647fbf565431a7200c3ec0d9b1b2ea15f08bd22f76547a973d0cab67d6ffa7","name":"legacy-example"}}
```

Stderr reported detected Claude release **unknown**, recorded release 2.1.261, and the ordinary first-resolve notice. The controlled PATH did not expose a Claude binary. This proves the real Curator entry emitted no MCP section for the empty set; it does not prove what Claude would load afterward.

### Existing config surface refusals

With the same hermetic control environment, separate invocations:

```sh
"$CIP_SCRATCH/curator-proof" env config set profile_settings.legacy-example '{}'
```

Exit **2**: `knob: unknown environments knob "profile_settings.legacy-example"`. Expected failing command: the proposed knob is not implemented.

```sh
"$CIP_SCRATCH/curator-proof" env config set permissions.legacy-example '"acceptEdits"'
```

Exit **1**: `environments.permissions.legacy-example: must be one of native, yolo`. Expected failing command: `acceptEdits` is not that launch-mode enum. A separate Python assertion exited **0**, confirming the synthetic config still equaled its original input after both refusals.

### Fixture error and correction

The first attempt at each of the three install probes used a `CURATOR_SYSTEM_CONFIG` override naming a nonexistent scratch file and an uncorrected `projects: []` draft fixture. All three exited **1** at the missing explicit system config, before reaching the package reader. Those attempts do **not** count as evidence for settings/MCP refusal. The fixture was corrected to a present empty schema-2 system config and `projects: {}`; all three commands were rerun with the results above. No failure was relabeled as a pass.

## E6. Review boundaries and findings record

- **Verified now:** current main/rc.14 identities; closed input surfaces; profile-versus-project distinction; shipped A seed selector; current MCP positive-only rendering and empty omission; adopted permission-mode separation; five baseline tests and the bounded scratch probes above.
- **Proposed, not verified:** the new machine knob, per-key marker ownership, project scope guard, all new schemas/fragments, adapter equivalence and negative enforcement, rollback semantics and implementation effort estimates.
- **Unresolved evidence for implementation:** exact pinned support for the legacy color field; native rule-mode grammar and source anchoring; Codex negative-only layer behavior; higher-precedence provider configuration/argv conflicts; PM's actual project-profile selection path. None was inferred from a passing helper test.
- **Evidence intentionally not collected:** real installer content, operator credentials, native account state, authenticated runtime turns, or a full Curator/launcher test suite. The brief forbids reading those private inputs; their absence is not evidence that the deployed configuration contains none.
- **LOGBOOK.md exception:** the binding brief explicitly prohibits editing it. Important findings are recorded in the CIP, this companion, and task notes instead. The generic logbook checklist item must not be checked as if an edit occurred.

Both files must be attached as new task-scoped outcome resources before the researcher handoff. The draft recommends option A and gives decision-ready alternatives and L1–L8 implementation leaves; review/adoption and implementation remain subsequent work.

## E7. Artifact verification and command-channel recovery

The first artifact-checker invocation and `git diff --check` returned no output while basic shell probes were also unresponsive. Both verification attempts were interrupted and exited **130**; neither is counted as passing. Scratch-checker creation and other interrupted non-TTY diagnostics also exited **130**; the interrupted TTY `pwd` probe exited **1**. These provided no successful validation or host-health evidence. The command channel subsequently recovered; no cause is asserted.

The standalone `git diff --check` retry exited **0**. This checks tracked changes only, so it is supplemented by a standalone `python3 -` inline artifact checker, exit **0**, which read only the two authored artifacts, the cited repository/spec sources, and Git metadata. It verified **10/10** required CIP sections, **38/38** pinned source-link paths and line ranges, **5/5** parseable JSON examples, balanced fences, final newlines and no trailing whitespace. Link checks establish that the cited locations exist at the pinned revisions; they do not substitute for the source reading recorded above.

The checker also verified that the only worktree changes are the **2/2** expected untracked research files, with no tracked product-code or `LOGBOOK.md` changes, and that their combined size is below the 80 KiB research bound. Its privacy check rejected personal-path prefixes, the internal-hostname sentinel and common credential-token/private-key patterns in the authored artifacts only. That is a bounded publication check, not a scan of native configuration, accounts or credentials. The earlier build, five selected tests and synthetic-probe exit codes remain as recorded in E5.

## E8. R1 reproducibility: replacement scripts and reruns (2026-10-05)

Review verdict R1 (2026-10-05, changes requested on TASK-261004-2iewnz — research-105-design-and-implementation-plan): the original inline Python bodies behind the E5 line-142 assertion, the E5 line-176 assertion, and the E7 lines-196-198 artifact checker were not recoverable. They ran as heredoc/stdin input in the prior session and only their exit codes were recorded. The three scripts below are replacement reproductions written for this rework. They are NOT the originals and are not claimed to be; E5/E7 historical results stay labeled historical and are not rewritten. Each replacement ran once, standalone, on 2026-10-05; the real exit code and date are recorded per script.

Pinned inputs shared by all three scripts: curator `ca1b776fb580ec0cee0173bf150daf063023aeaa` (the E1 baseline; verified present with `git rev-parse <sha>^{commit}` before use) and spec tag `v1.0.0-rc.14` (peeled `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`). Scripts A and B export the pinned curator commit with `git archive` into disposable scratch and build there, so the worktree source is never built in place and no repository file is modified. All fixture paths below are sanitized to `<scratch>`; the random `tempfile` suffix is intentionally not retained, matching E1. No real credential, token, Keychain secret, login, or personal path was used or printed.

### E8.1 Install probes and publication assertions (replaces the line-142 check)

Expected counts: probe exits 1/1/0 (settings, mcp-policy, control); 18/18 script checks.

Fixture construction is inside the script body: per-case `home`, `package`, and `work` directories; a schema-2 user config; a present schema-2 system config; one `agent-context.json` manifest per case. Exact file snapshots written by the script (only the real scratch root is replaced by `<scratch>`; each file also ends with one newline):

`config.json` per case (`<case>` is `settings`, `mcp-policy`, or `control`):

```json
{
  "projects": {},
  "schema_version": 2,
  "skills_root": "<scratch>/<case>/home/skills"
}
```

`sys.json`, identical in all cases:

```json
{
  "schema_version": 2
}
```

`package/agent-context.json` for the `settings` case:

```json
{
  "name": "legacy-example",
  "schema_version": 1,
  "settings": {
    "claude_code": {
      "permissions": {
        "allow": [
          "Read"
        ],
        "ask": [],
        "defaultMode": "acceptEdits",
        "deny": [
          "Bash"
        ]
      }
    }
  },
  "version": "1.0.0"
}
```

`package/agent-context.json` for the `mcp-policy` case:

```json
{
  "mcp_policy": {
    "claude_code": {
      "example-server": "disabled"
    }
  },
  "name": "legacy-example",
  "schema_version": 1,
  "version": "1.0.0"
}
```

`package/agent-context.json` for the `control` case:

```json
{
  "name": "legacy-example",
  "schema_version": 1,
  "version": "1.0.0"
}
```

Script body (`r1_install_assertions.py`):

```python
#!/usr/bin/env python3
"""R1 replacement reproduction for E5 install probes + line-142 assertions.

NOT the original inline script (unavailable); see evidence E8. Self-contained:
exports the pinned curator commit, builds it, constructs three synthetic
fixtures, reruns the three `profile install` probes, and asserts the
postconditions: no profile/config/surface publication on the two refusals,
and the expected selection on the control.

Usage: python3 r1_install_assertions.py <curator-repo-path>
Exit 0 iff every probe exit and every postcondition matches.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile

CURATOR_COMMIT = "ca1b776fb580ec0cee0173bf150daf063023aeaa"
BASE_MANIFEST = {"schema_version": 1, "name": "legacy-example", "version": "1.0.0"}
SETTINGS_EXTRA = {"settings": {"claude_code": {"permissions": {
    "allow": ["Read"], "deny": ["Bash"], "ask": [],
    "defaultMode": "acceptEdits"}}}}
MCP_EXTRA = {"mcp_policy": {"claude_code": {"example-server": "disabled"}}}
SYS_CONFIG = {"schema_version": 2}

fails = []


def check(name, cond, detail=""):
    print(("PASS " if cond else "FAIL ") + name
          + ("" if cond or not detail else " :: " + detail))
    if not cond:
        fails.append(name)


def run(argv, env, cwd):
    proc = subprocess.run(argv, env=env, cwd=cwd,
                          stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                          text=True)
    return proc.returncode, proc.stdout, proc.stderr


def write_json(path, obj):
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(json.dumps(obj, indent=2, sort_keys=True) + "\n")


def main():
    repo = sys.argv[1]
    scratch = tempfile.mkdtemp(prefix="r105-r1-install-")
    rel = lambda path: os.path.relpath(path, scratch)
    # Pinned input: the exact curator baseline the E5 probes used.
    subprocess.run(["git", "-C", repo, "rev-parse",
                    CURATOR_COMMIT + "^{commit}"],
                   stdout=subprocess.DEVNULL, check=True)
    export = os.path.join(scratch, "src")
    os.mkdir(export)
    archive = subprocess.run(["git", "-C", repo, "archive", CURATOR_COMMIT],
                             stdout=subprocess.PIPE, check=True)
    subprocess.run(["tar", "-x", "-C", export], input=archive.stdout,
                   check=True)
    binary = os.path.join(scratch, "curator-r1")
    build = subprocess.run(["go", "build", "-o", binary, "./cmd/curator"],
                           cwd=export, stdout=subprocess.PIPE,
                           stderr=subprocess.PIPE, text=True)
    check("build pinned curator exits 0", build.returncode == 0,
          build.stderr.strip().splitlines()[:1])
    if build.returncode != 0:
        print("RESULT: 0/1 checks passed (build failed)")
        return 1

    snapshots = {}
    for case, extra in (("settings", SETTINGS_EXTRA),
                        ("mcp-policy", MCP_EXTRA),
                        ("control", {})):
        casedir = os.path.join(scratch, case)
        home = os.path.join(casedir, "home")
        os.makedirs(os.path.join(home, "skills"))
        os.makedirs(os.path.join(casedir, "package"))
        os.makedirs(os.path.join(casedir, "work"))
        config_path = os.path.join(casedir, "config.json")
        write_json(config_path, {
            "projects": {},
            "schema_version": 2,
            "skills_root": os.path.join(home, "skills"),
        })
        sys_path = os.path.join(casedir, "sys.json")
        write_json(sys_path, SYS_CONFIG)
        manifest = dict(BASE_MANIFEST)
        manifest.update(extra)
        write_json(os.path.join(casedir, "package", "agent-context.json"),
                   manifest)
        with open(config_path, "rb") as handle:
            snapshots[case] = handle.read()

    results = {}
    for case in ("settings", "mcp-policy", "control"):
        casedir = os.path.join(scratch, case)
        env = {"HOME": os.path.join(casedir, "home"),
               "PATH": "/usr/bin:/bin",
               "CURATOR_CONFIG": os.path.join(casedir, "config.json"),
               "CURATOR_SYSTEM_CONFIG": os.path.join(casedir, "sys.json")}
        code, out, err = run(
            [binary, "profile", "install",
             os.path.join(casedir, "package")],
            env, os.path.join(casedir, "work"))
        results[case] = (code, out, err)

    check("settings probe exits 1", results["settings"][0] == 1,
          repr(results["settings"][0]))
    check("settings diagnostic names the field",
          "unknown field settings" in results["settings"][2])
    check("mcp-policy probe exits 1", results["mcp-policy"][0] == 1,
          repr(results["mcp-policy"][0]))
    check("mcp-policy diagnostic names the field",
          "unknown field mcp_policy" in results["mcp-policy"][2])
    check("control probe exits 0", results["control"][0] == 0,
          repr(results["control"][0]))
    check("control activates the profile",
          "installed and activated profile legacy-example"
          in results["control"][1])

    for case in ("settings", "mcp-policy"):
        casedir = os.path.join(scratch, case)
        check(case + ": no profiles dir published",
              not os.path.exists(os.path.join(casedir, "profiles")))
        markers = []
        for root, _, files in os.walk(os.path.join(casedir, "home")):
            if ".agent-environment.json" in files:
                markers.append(rel(root))
        check(case + ": no marker in synthetic home", markers == [],
              repr(markers))
        with open(os.path.join(casedir, "config.json"), "rb") as handle:
            check(case + ": user config bytes unchanged",
                  handle.read() == snapshots[case])
        lockdir = os.path.join(casedir, "state", "locks", "v1")
        locks = (os.listdir(lockdir) if os.path.isdir(lockdir) else [])
        check(case + ": lock bookkeeping exists (writes were nonzero)",
              len(locks) >= 1, rel(lockdir))

    casedir = os.path.join(scratch, "control")
    profdir = os.path.join(casedir, "profiles", "legacy-example")
    check("control: lock.json published",
          os.path.isfile(os.path.join(profdir, "lock.json")))
    check("control: source.json published",
          os.path.isfile(os.path.join(profdir, "source.json")))
    current_path = os.path.join(casedir, "profiles", "current")
    current = (open(current_path, encoding="utf-8").read()
               if os.path.isfile(current_path) else None)
    check("control: current selects legacy-example",
          current == "legacy-example\n", repr(current))

    total = 18
    passed = total - len(fails)
    print("RESULT: %d/%d checks passed" % (passed, total))
    shutil.rmtree(scratch, ignore_errors=True)
    return 0 if not fails else 1


if __name__ == "__main__":
    sys.exit(main())
```

Recorded run, standalone process without pipes (`<worktree>` was the story worktree; the script reads only the pinned commit object from its git store):

```text
$ python3 r1_install_assertions.py <worktree>
PASS build pinned curator exits 0
PASS settings probe exits 1
PASS settings diagnostic names the field
PASS mcp-policy probe exits 1
PASS mcp-policy diagnostic names the field
PASS control probe exits 0
PASS control activates the profile
PASS settings: no profiles dir published
PASS settings: no marker in synthetic home
PASS settings: user config bytes unchanged
PASS settings: lock bookkeeping exists (writes were nonzero)
PASS mcp-policy: no profiles dir published
PASS mcp-policy: no marker in synthetic home
PASS mcp-policy: user config bytes unchanged
PASS mcp-policy: lock bookkeeping exists (writes were nonzero)
PASS control: lock.json published
PASS control: source.json published
PASS control: current selects legacy-example
RESULT: 18/18 checks passed
```

Exit **0** on 2026-10-05. The manager home in these fixtures is the case directory itself (manager home is the config file's directory), so `profiles`, `state/locks/v1`, and the synthetic `home` all sit under `<scratch>/<case>`; that layout is asserted, not assumed.

### E8.2 Config refusals and unchanged-config assertion (replaces the line-176 check)

Expected counts: refusal exits 2/1 (unknown knob, wrong enum); 7/7 script checks. Fixture construction has the same shape as E8.1 (single `control` case; the plain package is installed first with expected exit 0; the config is snapshotted after that install and compared after both refusals), so no new snapshot files are introduced beyond E8.1.

Script body (`r1_config_refusal_assertions.py`):

```python
#!/usr/bin/env python3
"""R1 replacement reproduction for E5 config refusals + line-176 assertion.

NOT the original inline script (unavailable); see evidence E8. Self-contained:
exports the pinned curator commit, builds it, installs a plain control
package in a synthetic fixture, reruns the two refusing `env config set`
invocations, and asserts both refusal exits/diagnostics plus byte-identical
user config afterwards.

Usage: python3 r1_config_refusal_assertions.py <curator-repo-path>
Exit 0 iff every probe exit and every postcondition matches.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile

CURATOR_COMMIT = "ca1b776fb580ec0cee0173bf150daf063023aeaa"

fails = []


def check(name, cond, detail=""):
    print(("PASS " if cond else "FAIL ") + name
          + ("" if cond or not detail else " :: " + detail))
    if not cond:
        fails.append(name)


def run(argv, env, cwd):
    proc = subprocess.run(argv, env=env, cwd=cwd,
                          stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                          text=True)
    return proc.returncode, proc.stdout, proc.stderr


def write_json(path, obj):
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(json.dumps(obj, indent=2, sort_keys=True) + "\n")


def main():
    repo = sys.argv[1]
    scratch = tempfile.mkdtemp(prefix="r105-r1-config-")
    subprocess.run(["git", "-C", repo, "rev-parse",
                    CURATOR_COMMIT + "^{commit}"],
                   stdout=subprocess.DEVNULL, check=True)
    export = os.path.join(scratch, "src")
    os.mkdir(export)
    archive = subprocess.run(["git", "-C", repo, "archive", CURATOR_COMMIT],
                             stdout=subprocess.PIPE, check=True)
    subprocess.run(["tar", "-x", "-C", export], input=archive.stdout,
                   check=True)
    binary = os.path.join(scratch, "curator-r1")
    build = subprocess.run(["go", "build", "-o", binary, "./cmd/curator"],
                           cwd=export, stdout=subprocess.PIPE,
                           stderr=subprocess.PIPE, text=True)
    check("build pinned curator exits 0", build.returncode == 0,
          build.stderr.strip().splitlines()[:1])
    if build.returncode != 0:
        print("RESULT: 0/1 checks passed (build failed)")
        return 1

    casedir = os.path.join(scratch, "control")
    home = os.path.join(casedir, "home")
    os.makedirs(os.path.join(home, "skills"))
    os.makedirs(os.path.join(casedir, "package"))
    os.makedirs(os.path.join(casedir, "work"))
    config_path = os.path.join(casedir, "config.json")
    write_json(config_path, {
        "projects": {},
        "schema_version": 2,
        "skills_root": os.path.join(home, "skills"),
    })
    write_json(os.path.join(casedir, "sys.json"), {"schema_version": 2})
    write_json(os.path.join(casedir, "package", "agent-context.json"), {
        "name": "legacy-example",
        "schema_version": 1,
        "version": "1.0.0",
    })
    env = {"HOME": home,
           "PATH": "/usr/bin:/bin",
           "CURATOR_CONFIG": config_path,
           "CURATOR_SYSTEM_CONFIG": os.path.join(casedir, "sys.json")}
    work = os.path.join(casedir, "work")
    code, out, _ = run([binary, "profile", "install",
                        os.path.join(casedir, "package")], env, work)
    check("control install exits 0", code == 0, repr(code))
    with open(config_path, "rb") as handle:
        before = handle.read()

    code1, _, err1 = run([binary, "env", "config", "set",
                          "profile_settings.legacy-example", "{}"],
                         env, work)
    check("unknown-knob set exits 2", code1 == 2, repr(code1))
    check("unknown-knob diagnostic names the knob",
          'unknown environments knob "profile_settings.legacy-example"'
          in err1)
    code2, _, err2 = run([binary, "env", "config", "set",
                          "permissions.legacy-example", '"acceptEdits"'],
                         env, work)
    check("wrong-enum set exits 1", code2 == 1, repr(code2))
    check("wrong-enum diagnostic names the enum",
          "must be one of native, yolo" in err2)
    with open(config_path, "rb") as handle:
        check("user config bytes unchanged after both refusals",
              handle.read() == before)

    total = 7
    passed = total - len(fails)
    print("RESULT: %d/%d checks passed" % (passed, total))
    shutil.rmtree(scratch, ignore_errors=True)
    return 0 if not fails else 1


if __name__ == "__main__":
    sys.exit(main())
```

Recorded run, standalone process without pipes:

```text
$ python3 r1_config_refusal_assertions.py <worktree>
PASS build pinned curator exits 0
PASS control install exits 0
PASS unknown-knob set exits 2
PASS unknown-knob diagnostic names the knob
PASS wrong-enum set exits 1
PASS wrong-enum diagnostic names the enum
PASS user config bytes unchanged after both refusals
RESULT: 7/7 checks passed
```

Exit **0** on 2026-10-05.

### E8.3 Artifact checker (replaces the lines-196-198 check)

Enumerated families and expected counts: 10/10 required CIP sections (exact ordered list from the template); every enumerated pinned source link resolving in range (38/38 on the pre-E8 files; E8 adds no new source links); every JSON example parsing (5 pre-E8 plus the 5 E8.1 snapshots, 10/10 expected); balanced fences, final newline, and no trailing whitespace in both files; worktree changes exactly the 2 research files; `git diff --check` exit 0; LOGBOOK.md untouched; 0 publication-pattern matches. Combined size is reported against the historical 81920-byte budget, not gated (finding F-R1-size in E8.5). The link check enumerates only URL anchors of the pinned-blob form; prose range mentions and non-blob links (issue, documentation) are out of scope by design, as is link text that names a wider range than its own anchor.

Script body (`r1_artifact_checker.py`):

```python
#!/usr/bin/env python3
"""R1 replacement reproduction for the E7 artifact checker (lines 196-198).

NOT the original inline script (unavailable); see evidence E8. Verifies the
two research artifacts against pinned inputs: the 10 required CIP sections,
every pinned source-link path/line-range at its pinned revision, every JSON
example, fence/newline/whitespace hygiene, worktree change scope, and
bounded publication patterns. Combined size is REPORTED against the
historical 80 KiB budget, not gated: the R1-mandated script bodies necessarily
change the total (see E8 finding F-R1-size).

Usage: python3 r1_artifact_checker.py <curator-repo-path>
Exit 0 iff every gated check passes. Prints X/Y counts per family.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile

CURATOR_COMMIT = "ca1b776fb580ec0cee0173bf150daf063023aeaa"
SPEC_TAG = "v1.0.0-rc.14"
SPEC_COMMIT = "43bf0a2506d5c354a73bbc3ea4623d4653db10c7"
CIP = "261004_CIP-0006-legacy-provider-settings-and-mcp-optouts.md"
EVIDENCE = ("261004_CIP-0006-legacy-provider-settings-and-mcp-optouts"
            "_evidence.md")
REQUIRED_SECTIONS = [
    "Summary", "Motivation and user stories", "Current state", "Design",
    "Security considerations", "Compatibility and migration",
    "Specification changes", "Implementation plan", "Test plan",
    "Open questions for the operator",
]
SIZE_BUDGET = 81920  # Historical research budget; reported, not gated.
LINK_RE = re.compile(
    r"https://github\.com/relux-works/(curator|curator-spec)/blob/"
    r"([0-9a-f]{40})/([^#\s)]+)#L(\d+)(?:-L?(\d+))?")

fails = []


def check(name, cond, detail=""):
    print(("PASS " if cond else "FAIL ") + name
          + ("" if cond or not detail else " :: " + detail))
    if not cond:
        fails.append(name)


def main():
    repo = sys.argv[1]
    research = os.path.join(repo, ".research")
    cip_path = os.path.join(research, CIP)
    ev_path = os.path.join(research, EVIDENCE)
    check("both artifacts exist",
          os.path.isfile(cip_path) and os.path.isfile(ev_path))
    if fails:
        print("RESULT: 0/1 checks passed")
        return 1
    with open(cip_path, encoding="utf-8") as handle:
        cip_text = handle.read()
    with open(ev_path, encoding="utf-8") as handle:
        ev_text = handle.read()

    sections = [line[3:] for line in cip_text.splitlines()
                if line.startswith("## ") and not line.startswith("###")]
    check("CIP sections 10/10 required",
          sections == REQUIRED_SECTIONS,
          "found %d: %r" % (len(sections), sections))

    scratch = tempfile.mkdtemp(prefix="r105-r1-artifact-")
    clone = os.path.join(scratch, "spec")
    subprocess.run(["git", "clone", "--depth", "1", "--branch", SPEC_TAG,
                    "https://github.com/relux-works/curator-spec.git",
                    clone],
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                   check=True)
    head = subprocess.run(["git", "-C", clone, "rev-parse", "HEAD"],
                          stdout=subprocess.PIPE, text=True,
                          check=True).stdout.strip()
    check("spec clone HEAD equals pinned rc.14 commit", head == SPEC_COMMIT,
          head)
    subprocess.run(["git", "-C", repo, "rev-parse",
                    CURATOR_COMMIT + "^{commit}"],
                   stdout=subprocess.DEVNULL, check=True)

    links = LINK_RE.findall(cip_text) + LINK_RE.findall(ev_text)
    resolved = 0
    bad = []
    for owner, sha, path, start, end in links:
        want = CURATOR_COMMIT if owner == "curator" else SPEC_COMMIT
        rev = sha if owner == "curator" else "HEAD"
        home = repo if owner == "curator" else clone
        ok = sha == want
        if ok:
            proc = subprocess.run(["git", "-C", home, "show",
                                   rev + ":" + path],
                                  stdout=subprocess.PIPE,
                                  stderr=subprocess.DEVNULL)
            ok = proc.returncode == 0
            if ok:
                total = proc.stdout.count(b"\n")
                last = int(end) if end else int(start)
                ok = 1 <= int(start) <= last <= total
        resolved += ok
        if not ok:
            bad.append("%s/%s#L%s" % (owner, path, start))
    check("pinned source links %d/%d resolve in range"
          % (resolved, len(links)), resolved == len(links) and links,
          repr(bad[:3]))

    blocks = []
    for text in (cip_text, ev_text):
        in_json = False
        current = []
        for line in text.splitlines():
            if line.startswith("```"):
                if in_json:
                    blocks.append("\n".join(current))
                in_json = line.strip() == "```json" and not in_json
                current = []
            elif in_json:
                current.append(line)
    parsed = 0
    for block in blocks:
        try:
            json.loads(block)
            parsed += 1
        except ValueError:
            pass
    check("JSON examples %d/%d parse" % (parsed, len(blocks)),
          parsed == len(blocks) and blocks)

    hygiene_ok = True
    for label, text in (("CIP", cip_text), ("evidence", ev_text)):
        fence_lines = [line for line in text.splitlines()
                       if line.startswith("```")]
        balanced = len(fence_lines) % 2 == 0
        newline = text.endswith("\n")
        trailing = [line for line in text.splitlines()
                    if line != line.rstrip(" \t")]
        if not (balanced and newline and not trailing):
            hygiene_ok = False
        check("%s fences balanced, final newline, no trailing space"
              % label, balanced and newline and not trailing,
              "fences=%d trailing=%d" % (len(fence_lines), len(trailing)))

    status = subprocess.run(["git", "-C", repo, "status", "--porcelain"],
                            stdout=subprocess.PIPE, text=True,
                            check=True).stdout.splitlines()
    expected = {"?? .research/" + CIP, "?? .research/" + EVIDENCE}
    check("worktree changes are exactly the 2 research files",
          set(status) == expected, repr(status))
    diffcheck = subprocess.run(["git", "-C", repo, "diff", "--check"],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE, text=True)
    check("git diff --check exits 0", diffcheck.returncode == 0,
          diffcheck.stderr.strip()[:120])
    changed = subprocess.run(["git", "-C", repo, "diff", "--name-only"],
                             stdout=subprocess.PIPE, text=True,
                             check=True).stdout
    check("LOGBOOK.md untouched",
          "LOGBOOK.md" not in status and "LOGBOOK.md" not in changed)

    size = (os.path.getsize(cip_path) + os.path.getsize(ev_path))
    print("SIZE %d bytes combined vs historical %d budget "
          "(reported, not gated)" % (size, SIZE_BUDGET))

    # Fragmented literals: this body is pasted into the evidence file, so
    # contiguous secret/path spellings must not appear in this source.
    # Home prefixes require a username segment plus a subpath slash, so the
    # sanitized <scratch> fixture paths (which end in a bare "home" segment
    # or a quoted leaf) cannot match; a bare home dir without subpath is an
    # accepted bound (see E8).
    user = "[A-Za-z0-9_][A-Za-z0-9_.~-]*/"
    patterns = ["/User" + "s/" + user, "/hom" + "e/" + user,
                "/privat" + "e/" + user, "C:" + "\\\\Users\\\\" + user,
                "-----BEGIN " + "(?:RSA |EC |OPENSSH )?"
                + "PRIVATE KEY-----",
                "gh" + "p_[A-Za-z0-9]{36}",
                "github" + "_pat_[A-Za-z0-9_]{20,}",
                "s" + "k-[A-Za-z0-9]{20,}",
                "x" + "ox[bpas]-[A-Za-z0-9-]{10,}",
                "AK" + "IA[0-9A-Z]{16}"]
    hits = []
    for pattern in patterns:
        for label, text in (("CIP", cip_text), ("evidence", ev_text)):
            if re.search(pattern, text):
                hits.append(label + ":" + pattern[:12])
    check("publication patterns: 0 matches in artifacts", hits == [],
          repr(hits))

    print("RESULT: %d/%d gated checks passed"
          % (11 - len(fails), 11))
    shutil.rmtree(scratch, ignore_errors=True)
    return 0 if not fails else 1


if __name__ == "__main__":
    sys.exit(main())
```

Invocation is `python3 r1_artifact_checker.py <worktree>`, a standalone process without pipes; it clones the spec tag into disposable scratch and verifies the peeled commit before checking. Recorded run and observed counts: see E8.4, appended after the run so this section states only the pre-run expectations.

### E8.4 Recorded artifact-checker run

Standalone process without pipes (`<worktree>` sanitized as in E8.1):

```text
$ python3 r1_artifact_checker.py <worktree>
PASS both artifacts exist
PASS CIP sections 10/10 required
PASS spec clone HEAD equals pinned rc.14 commit
PASS pinned source links 38/38 resolve in range
PASS JSON examples 10/10 parse
PASS CIP fences balanced, final newline, no trailing space
PASS evidence fences balanced, final newline, no trailing space
PASS worktree changes are exactly the 2 research files
PASS git diff --check exits 0
PASS LOGBOOK.md untouched
SIZE 93338 bytes combined vs historical 81920 budget (reported, not gated)
PASS publication patterns: 0 matches in artifacts
RESULT: 11/11 gated checks passed
```

Exit **0** on 2026-10-05. Observed counts: 10/10 sections; spec clone HEAD equal to the peeled rc.14 commit; 38/38 pinned source links resolving in range; 10/10 JSON examples parsing; both files balanced with final newlines and no trailing whitespace; worktree changes exactly the 2 research files; `git diff --check` exit 0; LOGBOOK.md untouched; 0 publication-pattern matches.

Sequencing note: this subsection and E8.5 were appended after the recorded run above. They add no source links, no JSON examples, and no fenced code blocks, so the enumerated counts the checker verified are unaffected. That neutrality was confirmed by comparing the enumerated inputs before and after the append: identical blob-link multiset and identical JSON-fence count. The SIZE line above reflects the pre-E8.4/E8.5 total; the final total is in finding F-R1-size.

### E8.5 Bounds, findings, and open question

- **F-R1-size (finding for the operator).** The final combined artifact size is 96520 bytes (CIP draft unchanged at 42704; evidence 53816), over the historical 81920-byte research budget from the original brief. The overrun is forced by this rework brief itself, which mandates the full script bodies; no content can be cut without violating R1. Recommended: accept the overrun for this evidence file (a one-time reproducibility correction, not recurring research bloat), or direct a follow-up to move the bodies to a board-attached file. The checker reports size but does not gate on it for exactly this reason.
- **Sentinel bound.** The original internal-hostname sentinel value is unavailable (it was never published, by design), so the replacement publication check cannot reproduce it. It instead covers documented generic families: user-home path prefixes with a username plus subpath segment in four platform spellings, PEM private-key headers, and five credential-token shapes. A bare home directory without a subpath segment is an accepted gap, stated in the script.
- **Link-enumeration bound.** Only URL anchors of the pinned-blob form are enumerated and range-checked; prose range mentions, link text wider than its anchor, and non-blob links are out of scope by design.
- **Historical results preserved.** The E5/E7 numbers (5/5 tests, probe exits, 38/38 links, 5/5 JSON, 2/2 files) remain labeled historical; the reruns above neither confirm nor alter them. They add independently reproducible evidence for the same claims.
