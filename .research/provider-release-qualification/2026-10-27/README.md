# Exact-release native-project observations, 2026-10-27

TASK-261027-pu2vvu — provider-release-qualification-and-canary. **No provider
has a complete pass for the mandatory native-project contract. Keep refusal.**
This report observes installed release bytes; it changes no certification row,
adapter, product refusal, installation, login or launcher policy.

The baseline is research data, not a machine admission record. `pass` below
means the bounded surface/recipe named in its evidence, never whole-provider
qualification. `fail` means a witnessed violation of that recipe. `unknown`
means an incomplete discriminating positive/negative pair or lifecycle proof;
it requires refusal just as a failure does. Flag recognition is not evidence
of effective behavior. These results do not prove every possible native
configuration impossible.

## Frozen identities and provenance

Measured on macOS arm64 with Go 1.25.5. Binaries were copied read-only before
execution; updating launchers were not invoked. Versions were read inside the
offline fence. Claude 2.1.311 was installed, so 2.1.310 is **not measured**.

| Provider | Exact version | Binary SHA-256 | Size |
| --- | --- | --- | ---: |
| Codex CLI | 0.163.0 | `8813367285524ac798dd96d1b071aea30d58708c7226fd236cd275cf1cfdee31` | 249952688 bytes |
| Claude Code | 2.1.311 | `85d8c08707fcfd1e78ffa0204eefbc6ec9e4fe22f375fb2767ef85d957ed3161` | 276185504 bytes |
| Muse Code | 1.4.13-R6996.1 | `9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0` | 408650416 bytes |

Source witnesses came from main `f52b8bdd05d0705632b0142d31580e015aba2369`.
The research branch subsequently advanced to current main
`05102ddb4f4911f25c49fee5ba55221a5dd16a3e`; every source witness hash still
matches. The two intervening commits changed documentation outside this
research scope, so completed runtime measurements were preserved.

The [10f1wx research landed in PR174](https://github.com/relux-works/curator/pull/174)
at `362c2aad88677aa21eb2deaae0812d4ae00b2a99`. Its
[Codex/Muse witnesses](../../permissions-memory/README.md) and
[Claude witnesses](../../claude-permissions-memory/README.md) are replayed,
not silently promoted from their historical release pins. The
[Muse 1.4.13 resume/trust witness](../../muse1413/resume-trust/README.md)
is also replayed on the same exact digest.

## Per-form mandatory-surface matrix

`exec` means Codex/Muse exec or Claude print. `resume` means an actual
retained-session operation where explicitly identified below. A fresh
`--last`/`--continue` smoke probe without a retained session cannot establish
resume behavior. Interactive means a private PTY, never an exec-to-TUI
inference. Cell-level reasons are also recorded in [baseline.json](baseline.json).

| Provider / mandatory surface | Interactive | Exec / print | Resume | Evidence |
| --- | --- | --- | --- | --- |
| Codex / manager permissions | unknown | unknown | unknown | C1, S |
| Codex / memory off | unknown | unknown | **fail** | C2, S |
| Codex / project MCP restrictions | unknown | unknown | unknown | C3, S |
| Codex / project trust | unknown | unknown | unknown | C4, S |
| Codex / hooks | unknown | unknown | unknown | C5, S |
| Claude / manager permissions | unknown | **fail** | unknown | A1, S |
| Claude / memory off | unknown | unknown | **fail** | A2, S |
| Claude / project MCP restrictions | unknown | **pass** | unknown | A3, S |
| Claude / project trust | unknown | **fail** | unknown | A4, S |
| Claude / hooks | unknown | **fail** | unknown | A5, S |
| Muse / manager permissions | unknown | unknown | unknown | M1, S |
| Muse / memory off | unknown | **fail** | unknown | M2, S |
| Muse / project MCP restrictions | unknown | unknown | unknown | M3, S |
| Muse / project trust | unknown | unknown | unknown | M4, S |
| Muse / hooks | unknown | unknown | unknown | M5, S |

The permission obligation includes manager-only authority: repository rules
must not change the manager's selected grants, including by narrowing them.
The trust/hook obligation requires an explicit checkout/hook grant. Observing
a prompt or a trusted catalog alone does not establish the complete grant
binding. This report does not invent the unresolved product authority model.

### Codex evidence

**C1.** [permissions replay](evidence/permissions-replay.log),
`TestCodexShellDenial` and `TestNestedSandboxControl`: read-only and
workspace-write shell attempts fail nested `sandbox_apply`. An all-allow
nested control fails for the same environmental reason; a direct outer-fence
write succeeds. This is not proof of native permission enforcement. There is
no universal tool/default-deny matrix for any requested form.

**C2.** [bounded observations](evidence/permissions-memory.json),
`TestCodexMemoryRead` and `TestCodexDedicatedMemory`: enabled seeded summary
and real manual read/write positives are suppressed by the five explicit
manager off settings despite trusted-project memory enables; profile/project
instruction controls remain. Automatic memory generation/writes remain
unqualified. `TestCodexMemoryResume` performs real `exec resume --last` and
observes earlier memory-derived assistant text in the next model request
after all five settings are off. This is retained history, not new discovery.

**C3.** `TestCodexMCPDenial` and `TestCodexMCPDefaultPrompt` prove actual
annotated stdio tool success, specific denial, and no denied call marker.
These explicit tool gates do not establish a complete project MCP ceiling.
[Loader replay](evidence/codex-loader-replay.log), `TestMCPRuntime`, pairs
trusted/untrusted project startup with an approved MCP process/tool positive.
It uses app-server, so it does not certify interactive, exec or resume parity.
The historical repository `disabled_tools`/enumeration read-gap counterexample
in [the older Codex record](../../project-exclusion-codex/README.md) remains a
separate record: that adapter/read-gap test was **not rerun here**, and is not
claimed as a fresh 0.163.0 witness.

**C4/C5.** `TestConfigLayers` observes effective approval/sandbox settings,
MCP and `hooks/list` under trusted/untrusted app-server config.
`TestResumeApprovedRoot` does real thread resume, but both source-list
baselines omit the project; it cannot prove project exclusion.
[Loader projections](evidence/loader-observations.json) bind those observations.
They do not prove native hook execution, per-start operator binding or TUI
parity. The PTY displays the folder-trust prompt; no trust was granted.

### Claude evidence

**A1.** [Claude replay](evidence/claude-replay.log) and
[63 projected turns](evidence/claude.json), `TestPermissions`, pair real
Bash/Edit/Write/MCP effects with manager denies challenged by project/local
allows at fresh print, safe-origin resume and native child entries. Denied
files remain absent/original and denied MCP has no tool-call effect. WebFetch
has no successful offline positive and remains unknown. `TestBuiltinCeiling`
pairs empty/Read-only builtins with actual granted Read. However,
`TestProjectPermissions/none,project,local` proves both repository deny sources
defeat a manager Read grant while both managed and project instructions load.
That source-authority failure was measured at fresh print only. Resume's
individual passing atoms do not close its complete permission contract.

**A2.** `TestMemory` proves enabled/disabled seeded automatic-memory read at
fresh print, safe-origin resume and actual native child input.
`TestMemoryResumeHistory` performs real `--resume`: enabled-origin memory text
survives after `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1`. `TestMemoryDirectedWrite`
still writes into the memory directory with the flag off; this fixed tool
call is not a spontaneous automatic writer. Automatic write-off stays unknown.

**A3.** [Claude loader replay](evidence/claude-loader-replay.log),
`TestP5MCP/project-empty-{false,true}/exclude-false/strict-{false,true}/approved-{false,true}`:
real `.mcp.json` process witnesses appear without strict mode and disappear
under `--strict-mcp-config`, retaining the admitted process when selected.
An explicit empty set also works. `exclude-false` preserves native discovery;
this pass does not rely on `--setting-sources user`. It covers fresh print
source/process restriction, not permission to call every MCP tool or resume.

**A4/A5.** `TestP4Hooks/exclude-false` executes approved-home and project
SessionStart marker hooks without a persisted checkout trust grant.
`exclude-true` suppresses only the project hook, but also excludes native
project settings/context and is not a native-project solution. Startup smoke
independently records `claude-exec.hook_effect=true` in an untrusted checkout.
2.1.311 reports that its three project `permissions.allow` entries are ignored
until trust, while still executing that project hook. The PTY instead stops
at a trust prompt before any hook effect. The fresh `--continue` smoke also
executes the hook, but has no retained origin and is not a resume certificate.

### Muse evidence

**M1/M2.** `TestMuseToolDenials` in the
[permissions observations](evidence/permissions-memory.json) proves real
Write/Edit/WebFetch/add_memory and legacy `shell` success/denial pairs.
Managed `bash` lacks a successful positive because its proxy cannot bind
inside the privacy fence. These flags do not establish a manager-wide ceiling.
`read_memory` returns the seeded marker even with `--disable-write`.
`TestMuseMemoryConfigCandidate` rejects the attempted dedicated memory config;
it is not an exhaustive search of every configuration. `TestMuseMemoryReentry`
retains the previous memory result after deleting its source and disabling
writes. That is same-ID exec re-entry, **not canonical interactive resume**.

**M3/M4/M5.** [Muse resume replay](evidence/muse-resume-replay.log) and
[effect observations](evidence/muse-resume-observations.json) repeat the exact
1.4.13 P7 witness. Fresh untrusted exec has project rule/skill/MCP counts zero,
approved hook execution and no project-hook effect. Actual retained-session
`resume --last` offers trust/quit, without continue-untrusted. Granting trust
only to the synthetic checkout revives project skills/MCP and actually runs
the project hook, while saved root rules still report `project: none`.
That is a failure of the older *complete exclusion* recipe; native-project
intentionally allows project inputs after trust, so this is **not mislabeled
as a native-project trust failure**. Exact operator/hook-byte binding and
full manager MCP/hook enforcement remain unknown. `/usr/bin/true` MCP entries
measure configuration admission, not a working MCP tool server.

**S — bounded entry smoke.** [Startup observations](evidence/startup-observations.json)
cover all three forms on each exact binary with scratch HOME/config/XDG,
synthetic project allows, project MCP and hook inputs, memory-off candidates,
and no available external provider. All three PTYs display trust prompts;
they receive no trust/login answer and are stopped after six seconds.
Codex's missing-origin exec-resume starts fresh; Muse reports no retained
sessions; Claude's fresh continue is likewise not evidence of real resume.
Network-bound incomplete turns, absent markers before a usable turn, help
text and startup exits yield unknown, never permission/memory success.
`bounded_stop` records whether the process was still active at the last poll;
it can race a normal exit, so completed native terminal records/exit codes
are the evidence, not that boolean alone.

## Replay boundary, tools and validation

Every vendor child runs under macOS `sandbox-exec` and `env -i`, with scratch
HOME, native home, four XDG directories and TMPDIR under `.tmp`. The fence
denies actual native/config/credential homes, Keychains-shaped paths, security
lookups and managed/system policy locations. Network is denied except the
single local synthetic SSE provider port where required; there are no live
credentials or model inference. The synthetic request fixtures are transport
stimuli, not substitute provider implementations: real native tools produce
the effects. Incomplete native operations remain unknown.

`replay.py` copies existing Go witnesses into fresh `.tmp` stages. Staged changes
are exact binary/version pins, the required `.temp` → `.tmp` root guard and
optional Codex-only introspection selection; assertions and counterexamples
are preserved. Source SHA manifests are
checked before vendor execution. Every group has a PID, zero-exit receipt,
source hashes and sanitized public log. Native children are fixed-response
behavioral probes, not delegated development agents.

| Group | Native replay result | Recorded execution time |
| --- | --- | ---: |
| Codex + Muse permission/memory | pass as observation suite | 112.830s |
| Claude permission/memory | pass as observation suite | 102.506s |
| Codex config/MCP loader | pass as observation suite | 39.521s |
| Claude hooks/MCP loader | pass as observation suite | 28.072s |
| Muse retained resume/trust | pass as observation suite | 13.294s |
| 17 narrowing controls | all caught their intended failures | 56.347s |

Suite success includes successful detection of refusal/unknown witnesses. It
does not make the mandatory matrix pass. Narrowing controls include enabled
memory/tool plants, effect attribution, unknown-as-qualified, prompt export,
foreign digest, unrelated network and synthetic Keychains. Privacy-fence
mutants run only synthetic curl/cat/sh controls, never a vendor under a weaker
fence. Logs are linked above and hashed in [the manifest](evidence/manifest.json).

Two startup-harness attempts are retained privately. The first stopped at a
process-group signaling race before completing the matrix. The second uses
recorded foreground-PID cleanup and completed all entries; only its projected
observations are published. No surviving task-local process was found on
resume. Native replay receipts were already complete, so those six groups
were not repeated after the interrupted session.

| Tool | Purpose / command | Output |
| --- | --- | --- |
| Python 3 | `python3 .research/provider-release-qualification/2026-10-27/replay.py --help` | runner help; no vendor execution |
| Go 1.25.5 | Runner invokes selected copied Go research packages; no full product suite | scratch `replay.log`, `receipt.json`, PID and source hashes |
| `sandbox-exec`, `env`, Git, curl/cat/sh | Native privacy fence, empty environment, synthetic checkout and boundary controls | scratch per-case evidence |
| tmux | Existing Muse private socket/background terminal fixture, `/dev/null` config; recorded PID cleanup | scratch resume screen and marker effects |
| Frozen vendor binaries | `--version`, help, bounded entry and deterministic local protocol probes | projected version/argv/effects; raw inputs remain private |
| `startup.py` | Bounded private-PTY and headless entry smoke | scratch `observations.json`, raw output and PIDs |
| `publish.py` | `python3 .research/provider-release-qualification/2026-10-27/publish.py --scratch .tmp/TASK-261027-pu2vvu` | sanitized `evidence/` only |
| `canary.py` | Commands below | JSON disposition; no policy/board writes |
| `check_canary.py` | Paired CLI evidence controls; command below | fresh `.tmp` bundles and `results.json`; no vendor execution |
| `candidate.py`, `check_candidate.py` | Explicit all-provider research pins and paired native intake controls | candidate receipts and scratch control results; never admission |

Reproduce from the repository root with privately prepared read-only binary
copies named `codex`, `claude`, `muse` in `FROZEN_BINARIES`. Use a fresh scratch
directory for every invocation. The driver knows the actual HOME locator only
to construct deny rules; vendor processes receive synthetic homes.

```sh
R=.research/provider-release-qualification/2026-10-27
python3 "$R/replay.py" permissions --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/permissions-01
python3 "$R/replay.py" claude --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/claude-01
python3 "$R/replay.py" codex-loader --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/codex-loader-01
python3 "$R/replay.py" claude-loader --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/claude-loader-01
python3 "$R/replay.py" muse-resume --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/muse-resume-01
python3 "$R/replay.py" permissions --controls --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/controls-01
python3 "$R/startup.py" --binaries "$FROZEN_BINARIES" --scratch .tmp/qualification/startup-02
```

## Canary design and controls

`canary.py --binaries DIR` hashes frozen copies without executing foreign
bytes. A digest change, including a same-version replacement, emits exit 2 and
requires a new qualification task/review. A version change or any of the 45
verdict changes in `--candidate MATRIX` also emits exit 2 with before/after
values and evidence references. Missing forms, unknown schema/platform,
malformed verdicts or evidence and admission injection refuse with exit 3.
The comparison matrix is a separately reviewed observation input, not an
assertion that version/help output can infer behavior.

Each cell's `evidence` is now an exact relative artifact path;
the earlier explanatory text is preserved in `reason`. Blank-after-trimming
and dangling references are invalid. `evidence_manifest_sha256` pins the
reviewed bytes of `evidence/manifest.json`, which in turn pins every required
artifact. The canary verifies all manifest entries, not only the sole pass
cell's proof. Missing directories, manifests or artifacts, changed hashes,
and paths escaping the bundle yield exit 3, `qualification_required: true`
and `admitted: false` before any no-drift result, including in binary-only
mode. Baseline and candidate resolve proofs relative to their own JSON
file's directory; a copied candidate must bring its complete evidence bundle.
There is no fallback to the script's evidence when a candidate's proof is
absent. Manifest pins must be updated only with the reviewed evidence bytes.

Exit 0 means only no observed drift (or successful self-controls). Every mode
emits `admitted: false`, including an all-pass candidate. No mode edits product
tables, native homes or baseline, installs anything, creates a scheduler or
automatically approves a release. A proposed scheduled integration can open a
task after exit 2; that external automation is not installed by this research.
Future candidate bytes need an explicit research pin/review before replay.

The [0.164.0 follow-up](../2026-10-27-codex-0.164.0/README.md) adds that
explicit input without changing this record's baseline pins or evidence.
`replay.py --candidate FILE` and `startup.py --candidate FILE` accept a
`provider-release-candidate-v1` manifest with platform `darwin/arm64`,
`admitted: false` and exact `version`/`sha256` for all three providers.
Missing or mismatched pins refuse before a fixture or native execution.
Receipts retain the canonical candidate. `replay.py codex` and
`startup.py --provider codex` select only Codex observations. Without a
candidate, historical baseline pins remain the execution gate.

Canary `--candidate FILE --binaries DIR` accepts that pin format, checks all
three binary digests, and compares version/bytes to the unchanged baseline.
A pin manifest contains no verdict evidence, so this mode explicitly reports
verdict drift as unmeasured. The existing complete observation-matrix mode
still verifies its proof bundle and compares all 45 cells. Neither input form
can admit a release. Candidate pins do not establish effective behavior.

```sh
python3 "$R/canary.py" --self-test
python3 "$R/canary.py" --binaries "$FROZEN_BINARIES"
python3 "$R/canary.py" --candidate "$R/baseline.json"
python3 "$R/check_canary.py" --scratch .tmp/qualification/canary-controls-01
```

[Canary fixtures](canary-fixture.json) cover unchanged input, changed bytes,
new version/same verdicts, unknown/fail-to-pass, pass-to-fail/unknown, missing
form/evidence, invalid verdict and admission injection. The self-control also
checks that an all-pass matrix never admits. [Control receipts](evidence/canary-controls.json)
and [same-byte observations](evidence/canary-same-bytes.json) are persisted.
The [paired CLI receipts](evidence/canary-proof-controls.json) address PR187
review P2: blank, dangling, missing-directory, missing-artifact, missing-manifest,
hash corruption and escaping-path cases each produce the exact expected
refusal. Every case has intact and repaired exit-0 controls, with admission
false throughout. The dangling-reference repair supplies the previously
absent artifact plus its manifest entry and pin; it does not merely suppress
the failing check. Both baseline and candidate absence are exercised. These
controls use isolated copies and do not rerun or alter native probe evidence.
Remaining decisions and limits are tracked in
[UNRESOLVED_QUESTIONS.md](UNRESOLVED_QUESTIONS.md).
