# Review verdict: CIP-0004 shell-hook design research

- Task: TASK-261004-1z2pgb — shell-hook-no-source-and-path-append-design
- Parent: STORY-261004-7fglii — shell-hook-path-and-backend-env-hardening
- Change Request: CR-TASK-261004-1z2pgb-1 — shell-hook research revision 1
- Date: 2026-10-05
- Role: reviewer
- Verdict: accepted (research deliverable); route through accept_cr to integrating.
- Blocking findings: none. Free hunt: empty.
- Goal query: `task-board spawn goal` returned no active goal; this run is not goal-bound.

Acceptance approves the quality and scope of the research revision. It does not accept the proposed product design on the operator's behalf, authorize implementation or scheduling, or claim that a candidate hook is qualified. The operator decision remains pending.

## Exact revision and hygiene

Reviewed base `54bed271b7609bf206a04369202473c430d0d96a` against candidate tree `685f82e26cc7dc73e34027422e0582c93bc41c4a`. `git diff --name-status` reports exactly two added research files, and `git diff --check` exits 0. The exact `git diff` bytes hash to `5794745fcc2ad7cb9bb3b2b2f538fe33679408e60a8a166c6760c54b08c29993`, matching the assigned patch binding. Both worktree artifacts equal the corresponding candidate blobs byte-for-byte.

| Artifact | Bytes | SHA-256 |
|---|---:|---|
| `.research/261004_CIP-0004-shell-hook-no-source-path-append.md` | 35873 | `d7ac7832726d05a4b747ff38f70eafbecd1c9d79f3fcc8dce8169b4f37d74bc3` |
| `.research/261004_CIP-0004-shell-hook-no-source-path-append_evidence.md` | 41223 | `30bcdfd9d4a5a3b021d1ab7406f87a6f07304945901c8590c232be24087731b3` |

No product code or LOGBOOK.md changed. The reviewer made no repository edits, account changes, credential reads, commits, branch operations or integration actions. Safety inspection found no personal paths, internal hostnames, client/employer names or secret values in either artifact. The reproductions use synthetic payloads and explicitly isolated HOME/config environments. Historical researcher operations cannot be independently reconstructed from prose; the supplied scripts and recorded operations show no credential/account operation.

Fresh public `refs/heads/main` observation is `9bc8e1a1eace93377e41c36465b26a58ee5ce05a`, equal to local main. The diff from the research baseline to this tip is empty on the cited shell/envfiles/approval/install/CLI surfaces; the two candidate research paths have no upstream overlap since the CR base. This is a review freshness observation, not a landing-authority attestation.

## Swept review surfaces

| Surface | Result | Evidence in the submitted revision |
|---|---|---|
| Scope and operator authority | Holds | CIP lines 3, 13, 109–119, 150–158 retain draft status and prohibit implementation/scheduling before the operator decision. |
| Required option comparison | Holds | CIP lines 40–53 compare B-enforcing arbitrary digest approval, explicit canonical verify-and-trust, and direct no-source activation; include global bypass, TOCTOU and source-location compatibility. |
| Concrete recommendation and controls | Holds | CIP lines 55–70 specify no-source-v1, physical nearest-marker discovery, inherited/global/project order, snapshot restoration, manager-home binding, opt-outs, delimiter refusal and shell integration. |
| Hostile repository and PATH boundaries | Holds | CIP lines 59–76 distinguish data discovery from shell execution and append precedence from a command sandbox; inherited entries and provider trust roots remain explicit boundaries. |
| Terminal and copy-paste safety | Holds | CIP lines 78–80 require one-line ASCII-safe rendering, quoted printable operands, and omission of executable path hints for escaped operands. Invalid encodings, bidi, smart quotes and untrusted error strings are covered. |
| Compatibility and cached hooks | Holds | CIP lines 84–91 cover legacy helpers/approvals, exact managed-cache recognition, closed metadata shape, flag precedence, retained opt-outs, cache byte currentness, reload and unknown live-shell state. |
| Spec and architecture fit | Holds | CIP lines 93–105 change Manager §§3/8/10 and Environments §§11/12, replace only K3's staged plan, and explicitly preserve launcher and provider/update boundaries. |
| Decision-ready leaves and open questions | Holds | CIP lines 107–119 and 150–158 propose bounded L1–L5 leaves and five decisions with recommended answers; they create no board implementation tasks. |
| Production-entry test plan | Holds | CIP lines 121–146 cover all 18 supplied A1–D5 rows, real CLI/cache entry points, native Windows, negative and narrowing mutants, ratios and latency limits. Alternative A/B add all section E obligations. |
| Measurement provenance and bounds | Holds | Evidence lines 45–112, 116–559, 572–591 distinguish original measurements, recovery tests and future qualification. Baseline 9/18 touched rows is bounded; candidate coverage is explicitly 0/18. |
| Public artifact and diff hygiene | Holds | Exact-tree comparison, independent character/path checks and the embedded audit passed. Total 77096 bytes is below the 100 KB bound; companion links/template sections are present. |

## Citation spot checks

Independently inspected all 17 indexed C1–C10/S1–S7 reference groups, exceeding the required eight. Curator source is pinned to `ca1b776fb580ec0cee0173bf150daf063023aeaa`; curator-spec rc.14 peeled commit is `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. Reads used `git show <OID>:<file>` with numbered lines. The local rc.14 tag peels to the stated spec commit.

| Reference | Checked file and lines | Confirmed claim |
|---|---|---|
| C1 | `internal/envfiles/envfiles.go:24–63,72–123` | Both helper families prepend; POSIX helpers derive roots from source location; writers record digests. |
| C2 | `cmd/curator/main.go:2331–2372` | Real shell-init calls Hook/InstallHook with the global flag and prints cache/source instructions. |
| C3 | `internal/shell/shell.go:48–70,90–152,165–197,452–568,588–610,734–821` | A-warning default, global bypass, raw warnings, project check then source, restoration without root clearing and cached hook boundary. |
| C4 | `cmd/curator/hook.go:44–113` | Approval hashes selected current bytes without canonical-template equality. |
| C5 | `internal/install/commit.go:695–718` | Post-commit approval recording failures warn; this is not cache refresh. |
| C6 | `internal/shell/shell_hook_trust_test.go:521–594` | Test explicitly rejects hostile execution under B and requires it under A; unavailable pwsh skips. |
| C7 | `internal/hookapproval/hookapproval.go:602–631` | 0600 staging, sync and rename do not bind later shell execution. |
| C8 | `internal/shell/shell_test.go:40–122,141–178,302–335,413–440` | Existing activation/invalid-PWD/registration/cache/native prompt tests; leaving does not assert root clearing. |
| C9 | `internal/envfiles/envfiles_test.go:34–115` | Line 79 explicitly asserts PATH prefix. |
| C10 | `cmd/curator/security_posture.go:10–20` | HookTrust derives from the shipped binary profile, not live loaded code. |
| S1 | `profiles/manager.md:665–701` | Interactive helpers currently prepend; launcher prepend is a separate contract. |
| S2 | `profiles/manager.md:1242–1427` | Sourcing/discovery/global precedence, approvals, A/B staging and downstream execution binding match the draft. |
| S3 | `conformance/v1/vectors/shell-hook-trust.json:43–108` plus JSON parse | A/B profile expectations; 14 cases; structural spec validation does not execute hooks. |
| S4 | `protocol/environments.md:3440–3498` | Provider trust roots address the historical PATH attack with a separate rollout. |
| S5 | `protocol/environments.md:3575–3592` | Shared posture rows are read-only; non-current status makes --check non-zero. |
| S6 | `profiles/manager.md:2237–2244` | Compiled-repository environment exposure has the additional prepend sentence. |
| S7 | `profiles/manager.md:1502–1514` | Closed hook-trust values/provenance and other independent gate revisions. |

Additional checks: `.github/workflows/ci.yml:35–43` confirms the corpus pin; `internal/envfiles/stage.go:13–21` shares helper generators; `cmd/curator/main.go:416–459` accepts boolean flag=value syntax. Scoped rg found no InstallHook call in the reviewed install package and only the stated CLI install call in main.go; that absence claim remains scoped.

Reopened the four cited primary platform documents and confirmed alias parsing, command precedence, quoting and version-specific environment behavior: [zsh Functions](https://zsh.sourceforge.io/Doc/Release/Functions.html), [PowerShell command precedence](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_command_precedence?view=powershell-7.5), [PowerShell quoting rules](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_quoting_rules?view=powershell-7.5), [PowerShell environment variables](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-7.5). These support platform semantics only.

## Verification performed versus inherited evidence

Reviewer-executed checks, all exit 0: exact-tree/name/whitespace inspection; byte-for-byte candidate/worktree comparison; artifact audit recipe extracted from evidence lines 603–627; separate disclosure/control-character audit; pinned spec-vector JSON parse (14 cases); raw-observation JSON parsing; measurement consistency assertions; canonical POSIX fixture reconstruction and digest comparison; fresh public main-ref read and source-path comparison.

The raw observation corpus has 52 rows: 48 exit-0 child rows, two intentional exit-127 pasted-injection demonstrations and two non-process summaries. All ten project and two global marker observations agree with their counters. The separate control-byte corpus has four exit-0 rows. Recomputed fixture digest matches `ebe6c8ce75ef68bf4ff8eff7c549cccebf1c5c72b0671a3c82afb2a9b395b88c`.

Accepted as attached historical evidence, not rerun by this reviewer: the CLI build, P1–P4 real-shell probes, focused existing hostile-checkout/envfiles tests and recovery reruns. The scripts are reproducible, bounded, use real shell entry points and accurately limit the claims to the baseline. The passing hostile-checkout test is an old-contract result, including unsafe A execution; it is not proof of K3 safety. No local go build/go test ran, in accordance with hosted-evidence mode. No full suite, native Windows execution, TOCTOU exploit, candidate mutation result or implemented no-source behavior is attested.

The generic checklist is satisfied for this research scope: matches AC means the CIP/evidence deliverable; architecture fit means the proposed contract fits the stated boundaries; tests green means reviewer document/consistency checks plus the accurately bounded attached baseline tests. Product conformance remains unqualified. The conditional non-acceptance routing and logbook items are inapplicable: this verdict accepts the research, and the binding task explicitly forbids LOGBOOK.md edits.

## Recommendation for the next owner

Accept this revision and route it to the researcher/analyst integration lifecycle. After research publication, present the five operator questions in CIP lines 152–156 before prioritizing or scheduling L1–L5. The recommended package is C, literal-tail project append, no global sourcing, safe output and explicit cache migration. A retains approved arbitrary code and execution-race compatibility costs; B restricts bytes but retains store/snapshot complexity. The implementation must produce the planned real-shell evidence before any qualification claim.
