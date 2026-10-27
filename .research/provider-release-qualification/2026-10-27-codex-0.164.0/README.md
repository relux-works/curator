# Codex 0.164.0 pinned canary replay, 2026-10-27

**No mandatory native-project verdict changed.** Codex 0.164.0 still has one
witnessed **fail** (memory-off on actual resume) and fourteen **unknown** cells.
Version and binary SHA-256 changed; behavior in the bounded mandatory matrix
is unchanged. **Admission remains false.** Claude and Muse are byte-identical
to the baseline and were not behaviorally replayed.

TASK-261027-pu2vvu — provider-release-qualification-and-canary follow-up r2.
This draft is stacked on [PR187](https://github.com/relux-works/curator/pull/187),
base head `01d86d8ab2e5851cb96c9c5bfc5e8784678e164e`.

## Release identity drift

| Provider | Baseline → candidate | SHA-256 drift | New qualification required |
| --- | --- | --- | --- |
| Codex | 0.163.0 → 0.164.0 | yes | yes; observations still refuse admission |
| Claude | 2.1.311 → 2.1.311 | no | no new byte-driven replay |
| Muse | 1.4.13-R6996.1 → 1.4.13-R6996.1 | no | no new byte-driven replay |

| Provider | Candidate binary SHA-256 | Size |
| --- | --- | ---: |
| Codex | `d8990e3ee7acfbc26f91265b12a1a83f0fee40d0e0f8f1f40a37d4ec21f499cf` | 260293200 bytes |
| Claude | `85d8c08707fcfd1e78ffa0204eefbc6ec9e4fe22f375fb2767ef85d957ed3161` | 276185504 bytes |
| Muse | `9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0` | 408650416 bytes |

[Inventory](proofs/inventory.json) records the read-only freeze and bounded
version queries. The same frozen copies were used in this replay. Updating
launchers were never invoked. Native processes received scratch HOME/native,
XDG and TMP paths under `.tmp`, an empty environment, an offline macOS fence,
and synthetic local protocol responses. No real native config/credential
stores, live provider credentials, login or trust answers were supplied.
Configured trusted/untrusted synthetic project pairs reproduce the existing
loader witness; they do not answer a TUI trust prompt or trust a real checkout.

## All fifteen Codex cells: before → after

| Mandatory surface | Interactive | Exec | Actual resume | Proof |
| --- | --- | --- | --- | --- |
| Manager permissions | unknown → unknown | unknown → unknown | unknown → unknown | C1 |
| Memory off | unknown → unknown | unknown → unknown | **fail → fail** | C2 |
| Project MCP restrictions | unknown → unknown | unknown → unknown | unknown → unknown | C3 |
| Project trust | unknown → unknown | unknown → unknown | unknown → unknown | C4 |
| Hooks | unknown → unknown | unknown → unknown | unknown → unknown | C5 |

[Baseline delta](baseline-delta.json) records all 45 cells. The 30 non-Codex
cells explicitly carry prior evidence on identical bytes. Unknowns remain
unknown; success of an observation suite does not qualify a provider.

**C1.** [Permission/memory suite](proofs/codex-replay.log) and
[41 bounded records](proofs/codex-observations.json) repeat `TestCodexShellDenial`
and `TestNestedSandboxControl`: sandboxed shell has no successful filesystem
positive, while an all-allow nested sandbox also fails and the direct outer
write succeeds. This distinguishes environment nesting from native permission
enforcement, and leaves the complete manager ceiling unknown.

**C2.** Enabled seeded summary and manual memory read/write positives are
suppressed by the five memory-off settings. Automatic writes remain unproved.
`TestCodexMemoryResume` completes an enabled-origin turn and real
`exec resume --last`; the resumed main request still contains the earlier
memory-derived assistant marker after all five settings are off:
`positive_effect=true`, `forbidden_effect=true`, `specific_denial=false`.
This is retained history, not new memory discovery; the mandatory resume cell
remains fail. Fresh read/manual-tool atoms do not close complete memory-off.

**C3.** Explicit MCP deny and default-prompt pairs retain actual permitted tool
positives and remove denied effects. The [loader replay](proofs/codex-loader-replay.log)
repeats trusted/untrusted config and MCP process/tool pairs on the candidate.
It uses app-server, so its passing atoms do not establish the full project
MCP ceiling across interactive, exec and resume.

**C4/C5.** [Loader projections](proofs/codex-loader-observations.json) retain
approved/project settings and hooks/list catalogs under the synthetic trust
pairs. `TestResumeApprovedRoot` performs actual thread resume, but its source
baselines omit project input; it cannot prove exclusion. Hook execution/denial
and exact operator/hook-byte binding across forms remain unproved.
[Entry smoke](proofs/startup-observations.json) reaches the private PTY folder
trust prompt without an answer. All three entry probes stop at six seconds;
headless incomplete turns do not establish policy enforcement. Missing-origin
startup resume is not the retained-origin resume used in C2.

## Explicit candidate-pin contract and controls

[Candidate manifest](candidate.json) uses `provider-release-candidate-v1`,
`darwin/arm64`, explicit `admitted: false`, and exact version/SHA-256 pins for
**all three** providers. [candidate.py](../2026-10-27/candidate.py) validates
that contract. Replay/startup hash the selected binaries before creating any
fixture. Wrong/missing pins refuse before execution; the unchanged baseline
pins remain the default when no candidate is supplied. Candidate version
strings are also substituted into existing native identity assertions.

The input authorizes bounded observation only. It cannot admit, install,
change baseline evidence or change a product certification/refusal table.
The canonical candidate and all release pins are recorded in
[replay](proofs/codex-receipt.json), [loader](proofs/codex-loader-receipt.json)
and [startup](proofs/startup-receipt.json) receipts. The selected `codex` group
reuses the permission/memory witnesses and only Codex introspection; it does
not execute unchanged Muse probes. Staged edits are release pins/version
expectations, scratch-root guards and that provider selection, with original
source hashes checked before vendor execution. Assertions and fences stay.

[Nineteen candidate controls](proofs/candidate-controls.json) pass: mismatched
bytes, missing provider, missing digest, missing version and admission injection
are each checked in replay, startup and canary. Two further controls omit
`--candidate` and verify that replay/startup retain the baseline byte gate.
Refusals must name the intended
reason, create no native fixture and never execute the planted foreign file.
A matching pin runs real loader probes to completion and records the candidate;
the matching canary still demands qualification (exit 2), with admission false.
The existing fourteen canary self-controls and nine paired proof-bundle controls
also pass, including missing evidence and artifact hash corruption.

[Byte canary](proofs/canary-bytes.json), [candidate identity canary](proofs/canary-identities.json)
and [complete observation comparison](proofs/canary-verdicts.json) all require
qualification for Codex version/byte drift. The complete comparison finds
**zero verdict changes**. Its private matrix combines the fifteen new Codex
proofs with unchanged baseline Claude/Muse proofs; both bundles' evidence
hashes are checked. [Input receipt](proofs/canary-input-receipt.json) binds that
matrix and manifest; baseline delta and source pins retain its provenance.
No version/flag text is treated as proof of behavior.

## Reproduction and validation

Tests, probes and replays use:
`env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS`.
Git signing/publication uses the normal environment after `agent-git-env`,
per the corrected owner instruction. The earlier fixed-pin and signing
blockers are resolved by this authorized candidate input and signing scope.

```sh
R=.research/provider-release-qualification/2026-10-27
C=.research/provider-release-qualification/2026-10-27-codex-0.164.0/candidate.json
# FROZEN_BINARIES contains privately prepared read-only codex/claude/muse copies.
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS python3 -B "$R/check_candidate.py" --candidate "$C" --binaries "$FROZEN_BINARIES" --scratch .tmp/canary-controls-01
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS python3 -B "$R/replay.py" codex --candidate "$C" --binaries "$FROZEN_BINARIES" --scratch .tmp/codex-observations-01
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS python3 -B "$R/startup.py" --provider codex --candidate "$C" --binaries "$FROZEN_BINARIES" --scratch .tmp/codex-entry-01
```

| Tool | Purpose / output |
| --- | --- |
| Python 3 | Candidate parsing, SHA checks, projections and controls; public `proofs/`, private `.tmp` |
| Go 1.25.5, macOS arm64 | Existing native Go witnesses selected by replay; zero-exit receipts, source manifests and logs |
| `sandbox-exec`, `env`, Git, synthetic stdio MCP/SSE | Existing offline boundary and real native tool effects |
| Frozen Codex | Exact candidate native execution; version/help, paired probes and private PTY |
| Canary | Pin/byte/version/verdict drift and evidence integrity; every outcome refuses admission |
| Git / GitHub CLI | Signed stacked draft and exact-head publication checks |

Both native suites completed with exit 0 and no skipped tests. Candidate
controls include one further matching loader replay after strengthening the
control assertions; it also passed. Native observation success includes the
resume failure witness. All public proofs are projected/hashes only; no raw
prompts, native policies, executable locations or credentials are published.
The [proof manifest](proofs/manifest.json) and [driver provenance](proofs/driver-provenance.json)
bind this run. Independent review and hosted CI remain required; no product
certification, adapter or refusal changed, and no release is admitted.
