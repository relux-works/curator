# CIP-0005 evidence: audit backends and CLI secret transport

- **Task:** TASK-261004-hy8zmn — audit-token-argv-and-backend-env-allowlist-design
- **Parent:** STORY-261004-2b8pnx — design-audit-backends-and-secret-transport
- **Date:** 2026-10-04
- **Proposal:** [CIP-0005 draft](261004_CIP-0005-audit-backends-and-cli-secret-transport.md)
- **Scope:** read-only source research and synthetic scratch-HOME probes; no implementation or provider inference.

## Research boundary and exit criterion

Decision to unblock: safe publication credential transport and a supported audit-backend execution contract under manager profile §7. The design also has to preserve the adopted single owner of native harness flags. Operator acceptance and priority remain pending.

Bound: 60 worker minutes, one CIP plus this evidence file, at most 96 KiB combined, no archive, one serial research prerequisite. Exit criterion: options, recommendation, proposed spec sentences, consuming implementation leaves and decision-ready questions supported by pinned source or measured behavior. No grammar is accepted for implementation here; the draft identifies the smallest protocol/descriptor revision that a later spec leaf must freeze. First consuming production slice proposed: token-option refusal and secure token input; first backend slice: shared runner plus command protocol.

No real credential, token file, provider auth/config directory or Keychain secret was inspected. No login/logout or model invocation occurred. LOGBOOK.md is expressly out of scope despite the generic role checklist: findings are recorded in these research outcomes and task notes. Research documents are the only intended repository delta, uncommitted.

## Pinned sources

| Source | Observed identity | Verification |
| --- | --- | --- |
| curator | ca1b776fb580ec0cee0173bf150daf063023aeaa | Local HEAD; public refs/heads/main from a direct git ls-remote matched in the initial attempt. That attempt began with a clean working tree; recovery inherited only these two untracked research files. |
| curator-spec | 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 | Local HEAD and peeled v1.0.0-rc.14 tag matched; curator CI pins it at .github/workflows/ci.yml:37–43. |
| cocoaskills comparison | 0da153a54e7f8b446e63a6a177a9b5392e30da11 | Unauthenticated GitHub API resolved short commit 0da153a; PR #142 reports this head and merged=true. |
| Toolchain for measured Curator binary | go1.26.0 darwin/amd64 | GOENV=off GOTOOLCHAIN=local go version, exit 0. This is a research build, not the project's release toolchain qualification. |

[Public reference PR #142](https://github.com/ivanopcode/cocoaskills/pull/142) is titled “fix(security): secrets out of argv and audit backend env (K24)”. The browser could not render that PR page, so the research used its unauthenticated [API metadata](https://api.github.com/repos/ivanopcode/cocoaskills/pulls/142), [commit metadata](https://api.github.com/repos/ivanopcode/cocoaskills/commits/0da153a54e7f8b446e63a6a177a9b5392e30da11), and pinned raw public source files. Only relevant source/tests were fetched into temporary storage; no reference repository corpus or another operator's material is attached.

## Curator source observations

All links below point to curator at the inspected main revision. These are source observations; source inspection is not a claim that future controls ran.

| Claim | Pinned file and line |
| --- | --- |
| cmdAudit accepts a literal token and environment fallback | [cmd/curator/main.go:2111–2173](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L2111); token registration :2120; selection :2150–2157 |
| Standard flag errors write to stderr; interspersed parsing handles equals/separate forms | [cmd/curator/main.go:233–236](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L233); [parseInterspersed :417–454](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L417) |
| Exit 2 is usage; config load failures use exit 1 | [cmd/curator/main.go:54–58](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L54); [loadConfig :352–361](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L352) |
| Backend-specific validation is delegated by comment; raw object is retained | [internal/config/config.go:129–150](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/config.go#L129); [parseAudit :910–1011](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/config.go#L910) |
| Current size defaults; class settings | [internal/config/config.go:42–43](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/config.go#L42); [source_policy :1042–1078](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/config/config.go#L1042) |
| Gate and read-only variant share pipeline; static canary is enforced | [internal/audit/audit.go:155–205](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L155) |
| Static-only findings currently receive configured backend cache label | [internal/audit/audit.go:280–336](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L280); [:413–429](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L413); [:465–485](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L465) |
| Current verifiable-finding threshold behavior | [internal/audit/audit.go:131–148](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L131) |
| Canary checks both deterministic finding IDs | [internal/audit/audit.go:612–637](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L612) |
| Static detector scope is manifest/scripts and some read/walk errors are ignored | [internal/audit/audit.go:521–561](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L521). This is an additional reason to require independently enumerated backend request coverage; no claim that all existing detectors inspect all files. |
| Legacy hash helper versus explicit versioned helper | [internal/hashing/hashing.go:75–131](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/hashing/hashing.go#L75); local pin write at [audit/audit.go:378–395](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/audit/audit.go#L378) |
| Profile path hardcodes null | [internal/envprofile/envprofile.go:1684–1735](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/envprofile.go#L1684) |
| Registry publication validates record before request construction and sets Bearer header | [internal/registry/http.go:438–461](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/registry/http.go#L438). Explains why invalid synthetic records below stop before HTTP. |
| Git environment must remain a separate policy | [internal/gitops/gitops.go:36–39](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/gitops/gitops.go#L36) inherits environment on the ordinary lane; [:198–230](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/gitops/gitops.go#L198) shows the separate isolated lane and explicit exec Env assignment. |
| Existing platform-specific no-follow building blocks | [internal/pathboundary/open_read_nofollow_unix.go:11–16](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/pathboundary/open_read_nofollow_unix.go#L11); [Windows counterpart :12–36](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/pathboundary/open_read_nofollow_windows.go#L12). Neither alone supplies the proposed bounded private token-reader contract. |

Production integration inventory from scoped source search:

| Surface | Current entry into audit |
| --- | --- |
| Explicit CLI audit | cmd/curator/main.go:2265 → audit.Gate |
| External repository audit callback | cmd/curator/main.go:1963–1965 → GateReadOnly/Gate |
| Skill installation including dry-run | internal/install/install.go:583–585 → GateReadOnly/Gate |
| Global installation including dry-run | internal/install/global.go:198–200 → GateReadOnly/Gate |
| Profile closure member | internal/envprofile/envprofile.go:1732 → Gate, with separate null config |

This is an inspected call-site inventory, not an executed coverage ratio. Repair/upgrade callers must be traced into these owners in implementation tests; names in a proposal do not establish their coverage.

Scoped rg across production Go files found MaxRequestBytes references only in configuration definitions/parsing and found no audit child exec path. The stronger basis for the unused-backend claim is the inspected audit pipeline plus P1, not a zero-hit text search alone.

## rc.14 specification observations

| Contract | Pinned reference |
| --- | --- |
| Pipeline, canary, egress and mode semantics | [profiles/manager.md §7, lines 1101–1120](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L1101) |
| Versioned local audit identities/pins | [profiles/manager.md:1122–1129](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L1122) |
| Profile strictness and secret findings | [profiles/manager.md §12.6, lines 2977–3004](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L2977) |
| Existing open backend schema/defaults | [schemas/v1/manager-config-v1.schema.json:71–112](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/schemas/v1/manager-config-v1.schema.json#L71); [v3:37–38](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/schemas/v1/manager-config-v3.schema.json#L37) |
| Publication uses bearer, excludes tokens from diagnostics/cache | [protocol/registry.md:442–486](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/registry.md#L442) |
| CLI currently documents publication but no credential source | [cli/curator.md:44–46](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/cli/curator.md#L44) |
| Canonical source identities; allowlist prefix semantics | [protocol/core.md §6.1, lines 1192–1217](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/core.md#L1192). It does not define an audit source-policy glob grammar; the CIP proposes one explicitly. |
| Native harness flag owner, including exec mode | [Decision 0019:64–106](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0019-fragment-consumers-and-one-construction-site.md#L64); adopted status :5 |
| Interactive entry ownership; no headless curator-run workaround | [Decision 0021:64–68](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0021-sessions-enter-through-curator-run.md#L64); [Decision 0019:117–124](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0019-fragment-consumers-and-one-construction-site.md#L117) |

## Public reference comparison

Inspected at the exact PR head; these tests and mutants were **read, not rerun**.

| Reference source | Observation | Proposed use / limitation |
| --- | --- | --- |
| [environment.py:9–35](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/audit/backends/environment.py#L9) | Common allowlist; Windows canonicalization; deny after overrides; CODEX_HOME only per-Codex addition. | Adopt the invariant for every adapter and launch phase, adding a narrowly documented Claude directory variable. |
| [registry_token.py:13–67](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/registry_token.py#L13) | Handle identity/type checks, 64 KiB stat and read limits, newline rule; Unix permission check is conditional on non-Windows. O_NOFOLLOW may be absent. | Useful boundary tests, not Windows ACL/no-follow attestation. Proposed RFC padding grammar is narrower than its character-class-only regular expression. |
| [codex_backend.py:31–121](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/audit/backends/codex_backend.py#L31) | Canary and extraction share the environment constructor; source arrives on stdin; argv requests temporary structured output. | Do not copy native flag spelling into Curator: its adopted ownership differs. No Claude compatibility conclusion follows. |
| [command_backend.py:30–85](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/audit/backends/command_backend.py#L30) | Explicit env, stdin JSON, timeout and response parsing; configured cwd/env are available. | Retain direct process observation; Curator's proposed v1 removes arbitrary cwd and environment expansion. |
| [pipeline.py:92–118](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/audit/pipeline.py#L92), [:173–186](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/src/csk/audit/pipeline.py#L173) | Static/backend canaries and cloud checks; backend failures use the mode distinction. Oversize inserts a finding instead of extracting. | Compare error classes; Curator should explicitly record incomplete coverage and bound actual wrapped stdin as well. Do not infer enforcement from shared terminology. |
| [test_audit_secrets.py:410–543](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/tests/test_audit_secrets.py#L410) | Actual child-environment observations and a Git-auth/audit-child boundary fixture. | Model real Curator entry-point tests after this shape; helper-only predicates are insufficient. |
| [audit_secrets_mutants.py:23–122](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/tests/audit_secrets_mutants.py#L23) | Narrowing shapes include one-extra-byte, one permission class, only one backend, parent-only env filtering and locale exemptions. | Future Curator gates must kill their own analogous mutants. No imported killed/attempted ratio is claimed. |
| [test_audit_secrets_review.py:70–104](https://github.com/ivanopcode/cocoaskills/blob/0da153a54e7f8b446e63a6a177a9b5392e30da11/tests/test_audit_secrets_review.py#L70) | Tests mixed-case Windows-style filtering through launches. | Simulated platform behavior is not a substitute for native Windows ACL/reparse/process evidence. |

## External technical fact checks

Retrieved 2026-10-04; these are documentation observations, not live provider validation.

- [OpenAI noninteractive documentation](https://learn.chatgpt.com/docs/non-interactive-mode) documents stdin prompting, ephemeral exec, schema-constrained output and saved CLI authentication. [CLI reference](https://learn.chatgpt.com/docs/developer-commands?surface=cli) distinguishes ignoring user configuration from auth lookup and lists native exec controls. [Configuration reference](https://learn.chatgpt.com/docs/config-file/config-basic) exposes separate feature controls. These support the proposed typed launch requirements, but do not prove complete tool isolation or installed-version compatibility.
- [Claude CLI reference](https://code.claude.com/docs/en/cli-reference) documents print/structured output, persistence controls and setting-source selection. Empty built-in tools alone does not disable MCP. [Environment reference](https://code.claude.com/docs/en/env-vars) documents CLAUDE_CONFIG_DIR and directs credential-storage questions separately. No claim is made that this variable alone isolates Keychain credentials.
- [RFC 6750 §2.1](https://www.rfc-editor.org/rfc/rfc6750#section-2.1) defines the bearer alphabet with padding only at the end. This supports the proposed grammar; the registry spec independently governs actual token entropy.
- [Microsoft CreateFileW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew) documents OPEN_EXISTING, handle inheritance and opening reparse points without normal processing. [GetSecurityInfo](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-getsecurityinfo) retrieves security information by handle. Using those primitives for a conservative DACL reader is a proposal, not native verification.
- Public POSIX open documentation searches identified O_NOFOLLOW/O_NONBLOCK semantics, but full page retrieval failed. The proposal's local feasibility evidence instead points to Curator's existing Unix no-follow code and [internal/buildcache/protection_unix.go:226](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/buildcache/protection_unix.go#L226), which combines no-follow with nonblocking open. No filesystem-race or ACL probe was run in this research.

## Measured probes in a scratch HOME

### Construction and safety

Built the production cmd/curator binary from unchanged source. The build used a fresh scratch HOME, TMPDIR, GOCACHE, GOPATH and GOMODCACHE; GOENV=off and GOTOOLCHAIN=local; an explicitly constructed PATH; and no inherited credential variables. Dependencies were downloaded from public module sources. The standalone go build exited **0**.

Fixture preparation created a synthetic local skill Git repository, tagged v1, and a registered fixture project. The fixture Git init/add/commit/tag processes each exited **0**, with synthetic author/committer identity, disabled commit signing and empty global/system Git config. This committed only a disposable fixture repository, not the Story worktree.

The skill has two regular files: SKILL.md and an agent-skill.json schema-3 manifest with no environment, executable, network or secret capabilities. Its project manifest is schema 1 with one declaration of probe-skill at tag v1. There are no remote sources. Built-in registries are disabled and audit_registries is empty.

Effective configuration for P1 (all path labels below stand for scratch paths):

    {
      "schema_version": 1,
      "skills_root": "<scratch>/skills",
      "projects": {
        "fixture": {"path": "<scratch>/project", "agents": ["codex_cli"]}
      },
      "disable_builtin_registries": true,
      "audit_registries": [],
      "audit": {
        "enabled": true,
        "mode": "strict",
        "backend": "unsupported-probe",
        "backends": {
          "unsupported-probe": {
            "kind": "command",
            "command": ["/usr/bin/touch", "<scratch>/backend-launched"]
          }
        },
        "max_request_bytes": 1
      }
    }

The descriptor intentionally uses the currently opaque object shape; it is **not** an accepted CIP descriptor. The production test asks whether current code invokes anything at all.

For each CLI probe, the entire environment was constructed explicitly: PATH, scratch HOME, scratch TMPDIR, scratch CURATOR_CONFIG; P1 additionally set GIT_CONFIG_NOSYSTEM=1 and GIT_CONFIG_GLOBAL=/dev/null. P3 alone added the synthetic CURATOR_REGISTRY_TOKEN value. No real env inventory was dumped.

Commands below use <curator>, <record> and <config> as portable aliases for the built binary and scratch files. Each CLI ran directly as a standalone process without tee or a gate pipeline. The invalid publication record contains exactly {}. The loopback destination was a guard against accidental remote contact; record parsing prevents reaching HTTP.

### Results and real exit codes

| Probe | Direct CLI operation | Actual exit | Observation / interpretation |
| --- | --- | --- | --- |
| P1 | <curator> audit fixture --json | **0** | Strict audit returned warnings=null, errors=null. No backend marker. A verdict file named verdict-unsupported-probe--1-2.json held backend=unsupported-probe, zero findings and no hash_version. Confirms misleading successful static-only coverage for this selected backend, not successful analyzer execution. |
| P2 | <curator> audit --publish <record> --registry http://127.0.0.1:1 --token synthetic-fixture | **1** | Expected failure: invalid record lacks required name. Reaching record validation proves the literal option was accepted. This is a failing publication probe, not a passing publication gate. |
| P3 | CURATOR_REGISTRY_TOKEN=synthetic-fixture <curator> audit --publish <record> --registry http://127.0.0.1:1 | **1** | Expected failure: same invalid record. Confirms the existing environment source reaches record validation. No HTTP success/authentication claim. |
| P4 | <curator> audit --publish <record> --registry http://127.0.0.1:1 --token-file token.txt | **2** | Expected failure on current code: unknown -token-file option; usage still lists -token. No file read claimed. |
| P5 | Scratch-only marker/cache observation | **0** | Printed backend_child_observed=false; the cache record observation is shown below. This was a read of self-created synthetic state only. |

P1 stdout:

    [
      {
        "scope": "fixture",
        "warnings": null,
        "errors": null,
        "script_policies": null
      }
    ]

P5 output:

    backend_child_observed=false
    {"cache_file":"verdict-unsupported-probe--1-2.json","backend":"unsupported-probe","findings_count":0,"hash_version":"absent"}

P2/P3 terminal diagnostic:

    curator: record does not conform to audit-record-v1: audit record requires a non-empty string "name"

P4's first diagnostic:

    flag provided but not defined: -token-file

P1–P3 also emitted the existing security_posture_permissive warning. It is unrelated to the backend result; P1's audit mode was explicitly strict. None of these probes observes ps or shell history, tests a real provider, tests token-file enforcement, or claims future environment filtering is implemented.

Measured bound: **4/4 planned current-behavior CLI probes executed**, plus the scratch-state observation. **0/3 proposed backend implementations** were exercised, because the feature does not exist in this pipeline. No Linux/Windows native run, real credential flow, future security mutant, full Go test suite or reference-project test suite was executed. No passing test outcome from an earlier board artifact was accepted as new evidence.

### Reproduction outline

1. Create a fresh scratch directory and empty HOME; use the explicit environment above rather than the operator's home or Git configuration.
2. Build ./cmd/curator into that directory with GOENV=off GOTOOLCHAIN=local and fresh Go caches. Preserve the build exit status.
3. Create the two-file schema-3 skill, initialize/commit/tag the disposable local repository, and write the project manifest/configuration described above.
4. Run P1 once with a fresh manager state directory. Inspect only the synthetic child marker and local verdict JSON.
5. Write {} as the record and run P2–P4 separately. Preserve exits 1, 1 and 2; never report them as green publication gates.

## Artifact verification

The first artifact-verification attempt encountered an execution-environment stall: all four commands below returned no output or natural exit, as did independent date/pwd/shell-builtin probes outside the repository. The four validation processes were interrupted; each reported **exit 130**, not success. The temporary verifier-creation process was also interrupted with exit 130. A PTY-only shell diagnostic was interrupted with exit 1; non-PTY date/native-executable diagnostics reported 130. These are infrastructure observations, not expected-red security gates.

| Interrupted standalone validation command | Exit | Status |
| --- | --- | --- |
| git diff --no-index --check /dev/null .research/261004_CIP-0005-audit-backends-and-cli-secret-transport.md | 130 | Unverified; retry required |
| git diff --no-index --check /dev/null .research/261004_CIP-0005-audit-backends-and-cli-secret-transport_evidence.md | 130 | Unverified; retry required |
| git diff --exit-code HEAD -- LOGBOOK.md cmd internal go.mod go.sum | 130 | Unverified by this attempt |
| git -C <spec-checkout> diff --exit-code v1.0.0-rc.14 -- profiles/manager.md schemas/v1/manager-config-v1.schema.json schemas/v1/manager-config-v2.schema.json schemas/v1/manager-config-v3.schema.json decisions/0019-fragment-consumers-and-one-construction-site.md decisions/0021-sessions-enter-through-curator-run.md protocol/core.md protocol/registry.md | 130 | Unverified by this attempt |

Earlier successful source/status reads and P1–P5 remain the research evidence; the interrupted commands do not add verification. Any successful retries are recorded below separately. No interrupted or unrun validation is counted as green.

## Handoff interpretation

Ready for review means the research packet is attached to its task and handed off under the researcher role. It does not mean design acceptance, implementation, spec publication or production security qualification. The orchestrator can publish the reviewed CIP to curator-spec/cips and seek the operator's design/priority decision.

### Earlier execution-service blocker

After the research artifacts were written, command execution stopped returning output or natural exits, including a fresh non-login shell outside the worktree and a native no-op executable. Interrupting and recreating sessions did not restore execution. No provider/account or product-state change was attempted as a workaround.

Both task-scoped resource-add attempts also stalled and were interrupted with **exit 130**. Their persistence is **unconfirmed**, not assumed absent; query task-scoped outcomes before retrying a create and update an existing resource if one was stored. The subsequent CLI mutation to record the blocker and change status to blocked likewise stalled and was interrupted with **exit 130**. Its persistence is also unconfirmed. The last successfully observed board status was analysis; the researcher handoff was **not run**. No passing checklist or to-review status is claimed.

At that point the required external input was restoration of the local command runner. The recovery procedure was to validate the two research files and tracked-delta scope, inspect the authoritative board through task-board, attach/update both task-scoped outcomes, resolve the conditional LOGBOOK checklist item as not applicable under the explicit no-edit instruction, and run the researcher handoff. No implementation producer was authorized: design acceptance remains pending independently of that execution outage.

### Recovery verification, 2026-10-04

The command runner responded normally in the resumed researcher run. The authoritative board query exited **0** and showed analysis, eight unchecked checklist items, and only system spawn logs as outcomes: neither attempted research attachment had persisted. This establishes the earlier attachment state without inferring absence from the interrupted writes. Recovery preserves the CIP recommendation and the prior probe record; it adds verification and packaging only, not another research prerequisite or implementation work.

**Evidence provenance:** P1–P5, the scratch-HOME build, fixture setup, and their exit codes above are retained measurements from the initial attempt. They were **not rerun** in recovery and are not presented as new executions. Recovery directly re-read the production token parser, backend configuration/cache path, profile null selection, the rc.14 audit policy, and adopted launch ownership. HEAD and the rc.14 peeled tag still have the recorded identities. No full Go suite, provider invocation, credential read, Windows/Linux runtime probe or future narrowing mutant was added.

| Recovery command / observation | Actual exit | Interpretation |
| --- | --- | --- |
| git diff --exit-code HEAD -- LOGBOOK.md cmd internal go.mod go.sum | **0** | Tracked product sources, dependencies and LOGBOOK.md unchanged. |
| git diff --no-index --check /dev/null .research/261004_CIP-0005-audit-backends-and-cli-secret-transport.md | **1** | Non-zero comparison against an empty file, with no whitespace diagnostic. Not counted as a passing validation. |
| git diff --no-index --check /dev/null .research/261004_CIP-0005-audit-backends-and-cli-secret-transport_evidence.md | **1** | Same non-zero comparison result. Replaced as a whitespace gate by the explicit document validator below. |
| git -C <spec-checkout> diff --exit-code v1.0.0-rc.14 -- profiles/manager.md schemas/v1/manager-config-v1.schema.json schemas/v1/manager-config-v2.schema.json schemas/v1/manager-config-v3.schema.json decisions/0019-fragment-consumers-and-one-construction-site.md decisions/0021-sessions-enter-through-curator-run.md protocol/core.md protocol/registry.md | **0** | Cited local specification files match the released baseline. |
| python3 <scratch>/verify_cip.py | **0** | 24/24 document checks; 2/2 required companion artifacts. The same standalone check is rerun after this recovery annotation before attachment. |
| Unauthenticated reference-source/PR retrieval with Python urllib | **0** | PR #142 is merged and its head is 0da153a54e7f8b446e63a6a177a9b5392e30da11; four source files below were re-read. No reference tests executed. |

The temporary document validator checks twelve required template headings; each artifact's terminal newline, trailing whitespace and explicit private-path/host markers; relative companion links; board IDs accompanied by their titles; the combined 96 KiB artifact ceiling; and an exact Git delta containing only the two requested research files. Its denominator is these **24 explicit document assertions**, not runtime security coverage. Manual review also inspected the packet for private material. Neither review establishes absence of every possible confidential string; no real account state was used as input. The validator is a scratch-only packaging aid, not product code or a proposed production gate.

Reference-source digests measured in recovery (all at the pinned public PR head):

| Relative reference path | Bytes | SHA-256 |
| --- | ---: | --- |
| src/csk/audit/backends/environment.py | 1411 | d18575709f5a0f1931aaff8469b1c67d780f00327b03f710d83f852f8410f549 |
| src/csk/registry_token.py | 3329 | 40a2e0279df4fe2e8b0e7a67d47de4eff82041fd8d798ddc37e9ee04ce953957 |
| src/csk/audit/backends/codex_backend.py | 6231 | 0ab8867135b3fa57906b16936233a5d0d99258bc77e42dd11a942bdcbb7417de |
| src/csk/audit/backends/command_backend.py | 3769 | 8f5fca8dd58ff01f29a704eee9cecd507dfe1d74d77734a330fec2f7da597214 |

Browser reads of the two raw GitHub URLs failed again; the successful unauthenticated retrieval above supplies the evidence instead. Official documentation was independently reopened: [Codex exec reference](https://learn.chatgpt.com/docs/developer-commands?surface=cli) confirms that ignoring user configuration retains CODEX_HOME authentication lookup; [Claude CLI reference](https://code.claude.com/docs/en/cli-reference) distinguishes disabling built-in tools from MCP and documents setting-source controls; [Claude environment reference](https://code.claude.com/docs/en/env-vars) documents CLAUDE_CONFIG_DIR without promising that it relocates every credential store. These reconfirm the proposal's qualification boundaries, not a working native adapter.

Checklist interpretation: the conditional generic logbook item is **not applicable** because the task explicitly requires LOGBOOK.md untouched. Findings, anomaly history and verification belong to this evidence file, the CIP and task notes instead. Closing that conditional checklist item records this exception; it does not assert a logbook edit. Both research artifacts must be attached before the researcher handoff. The board handoff receipt, not this document's wording, determines whether the task reached to-review.
