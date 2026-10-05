# Review verdict: changes requested

Task: TASK-261004-34brhn \u2014 research-claude-login-transfer-modes
Parent: STORY-261004-1pwxri \u2014 claude-macos-managed-home-login-modes
Change request: CR-TASK-261004-34brhn-1, revision 1.
Reviewed base ca1b776fb580ec0cee0173bf150daf063023aeaa and candidate tree 486e1bddf306b2ac6ebe2b7457e528a1c48b2897. Verdict: changes_requested; route to analysis for research-evidence rework, then another reviewer cycle. No human decision or external blocker is needed for this correction.

## Finding R1 \u2014 reproducible measured evidence is missing (acceptance-blocking)

In .research/261004_CIP-0003-claude-managed-home-credential-modes_evidence.md:79, the only reproduction commands are `python3 $SCRATCH/probe.py --version`, `python3 $SCRATCH/probe.py setup-token --help`, `python3 $SCRATCH/probe.py auth status --json`, and `python3 $SCRATCH/matrix.py <case>`, invoking `$SCRATCH/probe.sb`. None of those files is supplied or defined in the companion. Lines 73\u201377 describe a whitelist, sandbox, shim and synthetic objects without their exact definitions. Lines 89\u2013105 nevertheless report the 15 measured outcomes and coverage. Line 144 similarly reports verification through an unavailable `verify_artifacts.py`.

The binding review instruction requires every measured/verified claim to have a reproducible probe in the evidence file. A future reviewer cannot reproduce the exact fixture, environment, helper behavior, sandbox denials or Keychain shim from these commands. In particular, the safety claims at lines 74\u201375 and 124 cannot be independently checked against the harness. This is missing evidence, not an allegation that real credentials were accessed.

Required correction: include complete sanitized, self-contained harness and sandbox/shim/fixture definitions as fenced text in the evidence companion (or an exact equivalent runnable recipe), including the environment whitelist, per-case setup/reset, timeout/exit capture, P14 persistence sequence, and output normalization. Use runtime-generated scratch paths and synthetic values only. Include the artifact verification recipe or downgrade that assertion to producer-reported evidence. Preserve the two-document research-only scope. Do not recreate the probes against the operator account, run login/logout, or access real credential stores. If the original harness is unavailable, distinguish historical reported observations from a newly reconstructed harness and its results; do not imply identity between them.

## Swept review surfaces

| Surface | Result |
| --- | --- |
| Operator scope and mechanisms | Pass: token, helper, per-home and conditional shared-file; four genuine options with tradeoffs. |
| Decision readiness | Pass: concrete source registry and selector union, defaults, whole-selector precedence, expiry/rotation, launch boundary, migration and versioned schema proposal. |
| Q1/Q7 and architecture | Pass: no native credential copying; sharing remains gated; synthetic selection is not successful authentication; launcher secret handling stays after tracked serialization. |
| Citation accuracy | Pass for the spot checks listed below. |
| Measured evidence reproducibility | Fail R1. Shipped-code offsets are supplied; scratch harness definitions are not. |
| Safety/publication | Artifact inspection found no concrete personal paths, internal hostnames, employer/client names or real credential values. Producer asserts no real-store access; historical execution safety cannot be independently established without R1. This review read no credential stores and ran no Claude authentication command. |
| Diff hygiene | Pass: exact candidate adds only the named CIP and evidence under .research/; LOGBOOK.md and product code unchanged. Both working files are byte-identical to candidate blobs. |
| Tests/checks | No go build/test, live login or account probes run, per hosted-evidence instruction. Candidate git diff --check passed; both JSON examples parsed; combined artifact size 51,905 bytes. No attached product suites accepted as current passes. |

## Citation spot checks (12/12 checked groups agree)

Checked from pinned Git objects, not inferred from prose:

1. Curator internal/config/environments.go:20 \u2014 Isolation map, no credential-mode fields.
2. Curator internal/config/environments.go:799 \u2014 shared/isolated closed parser.
3. Curator internal/envregistry/envregistry.go:240 \u2014 empty Darwin passthrough and 2.1.261 pin.
4. Curator internal/envregistry/envregistry.go:450 \u2014 macOS isolated default/shared refusal.
5. Curator internal/envprofile/managed.go:319 \u2014 production isolation resolution.
6. Curator internal/envprofile/managed.go:563 and :598 \u2014 link selection and unconditional Keychain marker.
7. Curator internal/envprofile/managed.go:1504 \u2014 credential-link preservation/refusal.
8. Curator internal/envfragment/envfragment.go:1 and :54 \u2014 closed fragment and v2/v3 behavior.
9. Spec rc.14 protocol/environments.md:1426 \u2014 native-store passthrough and unverified Linux refresh.
10. Spec rc.14 decisions/0017-environment-credential-modes.md:70 and :168 \u2014 historical file observation, Q1 gate, Q7 refusal.
11. Spec rc.14 profiles/manager.md:2847 and protocol/environments.md:3758 \u2014 credential mutation boundary, both lock directions, no credential-source lock.
12. Launcher internal/composition/composition.go:35 and :75, internal/execution/execution.go:117 and :142 \u2014 excluded direct environment, serialized env_literals and direct/tracked split.

Spec commit: 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. Launcher commit: 2517d2753945d0a8b0c40a291e4aef883a7c16ee.

Also checked the current official authentication documentation: config-directory scoping, macOS fallback and credential precedence support the bounded statements in the draft. Source: https://code.claude.com/docs/en/authentication (reviewed 2026-10-05). The settings-reference web rendering failed; no independent helper-documentation verification is claimed from that failed read. Binary offsets and producer scratch results were not independently rerun.

Fresh public main advertisement was 54bed271b7609bf206a04369202473c430d0d96a. Its changed-path set since the CR base does not overlap either candidate research path. Citation checks above remain explicitly pinned to the submitted source versions; this is not an integration authorization.

Run goal query returned none (not goal-bound). One focused rework finding after the complete surface sweep; no code or LOGBOOK changes made by the reviewer.
