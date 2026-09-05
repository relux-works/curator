# Review findings 0013-1: Decision 0013 at 71ac9d1

Subject: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
branch `draft/decision-0013-execution-ownership`, head `71ac9d1`, base `b4f29cd`.
Change Request `CR-TASK-260905-2ft7ts-1` rev 1 on the story branch has
`repository_delta=empty`: by the producer brief's design the document lives only
in the external worktree above, so the empty story delta is expected and not a
defect. The deliverable was judged in that worktree.

Verdict: **CHANGES REQUESTED** (two major findings). Route: `to-dev`.

## What was verified (my own checks, not the report's)

- One signed commit `71ac9d1` on the named branch, base `b4f29cd`, delta exactly
  one new file (704 lines). `git log --show-signature`: `Good "git" signature with
  ECDSA key SHA256:V6JiKG…`; the "No principal matched" line is the same local
  allowed-signers artifact the base commit `b4f29cd` shows, not a signing defect.
- Numbering: `origin/main` = `b4f29cd`, `decisions/` on main has 0012 and no 0011;
  `draft/TASK-260728-1yhuqi-swift-driver` head `604d525` carries
  `0011-swift-driver-pair.md`. 0012 Context lines 60–63 name Option A and "next
  free number after reconciliation with the swift-driver draft's 0011".
- ax `28bf96d` = tag v0.5.0: §1.6 (`urn:ax:schema:*`, reverse-DNS keys ≤64,
  extensions object ≤65,536 bytes, depth 4, `sha256:` digest form, bytes as
  unpadded base64url), §2.1 name grammar, §2.4, §5.1 limits (argv 1..128 × 1–4,096
  B, 65,536 total; env_names 0..64; env_literals 0..64 × 4,096, disjoint;
  contains_secrets false), §7.3 seven-name registry, §7.5 `SpawnPlan` row and
  "trusted terminal backend performs process creation", §7.7 yolo mappings,
  §13.1 steps 2/4, §13.2, §13.10, §14.1, §15.1 `details`, §15.3 codes
  (`invalid_arguments` 2, `capability_unavailable` 6, `provider_protocol_error` 13,
  `policy_refused` and `secret_policy_violation` 16), §16.2.
- PR #1 head `d7075e1`: three keys, "launcher merges … env_literals", "adds no
  member to SpawnPlan", "commit of a git profile or the state hash", warn-and-
  continue drift default, `curator session` note — all present in the diff.
- agents-management `944c7b4` (decision's pinned commit; main has since advanced
  one commit to `91bf945`, alias resolution, no LaunchMode change): `LaunchMode`
  three values + `Valid`/`String`; `Home` "load-bearing" comment; `EffortTransport`
  None/Argv/Stdin; `StdinPayload{Attached, Bytes}`; `Composition.Prefix`;
  `Capabilities.LaunchModes`; `Plan{Binary,Argv,Env,Stdin,WorkDir}`; `BuildPlan`;
  `ErrUnsupportedLaunchMode`; `ErrEffortMissing`; claude argv `-p --output-format
  json --model … --effort … --max-budget-usd … --dangerously-skip-permissions` and
  `compositionArgvPrefix`; codex `exec -m`, `--dangerously-bypass-approvals-and-sandbox`,
  `-c model_reasoning_effort=`; qwen is the only `EffortTransportStdin` system,
  pi is `EffortTransportNone` (the report's brief-divergence is correct);
  `vendorplugin.Lineup(models []Model) []RankedModel` in `pkg/vendorplugin/lineup.go`;
  `Effort.Recommended`; architecture invariants "Observable launch surface is the
  parity bar", "Single source per fact", "no defaults are injected anywhere".
- curator-spec `b4f29cd`: 0010 D6/D10; 0012 D3 (lock hash is the effective pin),
  D6 (single composer, `env_names` allowlist, `--mcp-config`/`--strict-mcp-config`,
  `-p curator-mcp`, `OPENCODE_CONFIG`), D8; environments §7.3, §10.1–10.4, §11,
  and §10.2 "`system_prompt` is present exactly when the resolved chain carries
  at least one applicable system module"; registry §1 CCJ-1; core §2 grammar
  `^[A-Za-z0-9][A-Za-z0-9._-]*$` (the report's docs-confidence item — confirmed,
  it falls inside ax §2.1 except for length, which the decision handles).
- launcher `6de42d8` `0.1.2-draft`: §1 "No plan rebuilding", §3, §4.1–4.5 (§4.5
  "ALWAYS routes the composed launch through ax's instrumentation", no untracked
  fallback), §4.2 mapping, §5, §6 `ax_handoff_failed`, §8, §9.
- Review v3 M1 recommendation text quoted in Rejected alternatives matches the
  source verbatim; M2 "Reject launcher-owned interactive argv", M15, M16 present.

All eight contract items of the producer brief are specified; the report's
item→section table is accurate. Dimensions 1, 3, 4, 6, 7 of the review brief pass.
Dimensions 2 and 5 produced the findings below.

## Findings

### F1 — major — Decision 3.4 / 3.5: `ax.launch-plan-request` cannot fit a valid plan

Quote: "The document itself is not stored: the ax §1.6 extension bound is 65,536
canonical bytes, and a document may carry 65,536 bytes of argv … `argv_suffix:
[...] | null`" (stored under the extension key).

What is wrong: the decision rejects storing the whole document because of the
65,536-byte extensions bound, then stores the full `argv_suffix` inside that same
bounded extensions object — alongside the four Curator keys and any caller
`extensions`. A plan whose `argv_suffix` approaches the §5.1 argv total (which
§3.3 admits) yields a Session Record whose canonical `extensions` object exceeds
ax §1.6. Attack shape: a document that passes every §3.3 check produces a record
ax cannot persist; the decision names no refusal for it, so it is either an
unnamed late failure after §3.3 said "valid" or a silent bound violation.
Resume replay (§3.5) is also built on this stored copy.

Fix: drop `argv_suffix` from `ax.launch-plan-request`. `base_argv_length` plus
the record's own `launch_plan.argv` already determine the suffix
(`argv[base_argv_length:]`); make §3.5 replay it from there. Keep `form`,
`base_argv_length`, `request_digest`. Then state the residual bound explicitly:
caller `extensions` plus the ax key must fit §1.6, refused as
`launch_plan_invalid`, `field: "extensions"` at §3.3 time.

### F2 — major — Decision 3.6: permission-bypass gate is `MAY`, leaving a record-misrepresenting bypass path

Quote: "`ax` MAY refuse a known bypass spelling in a caller-supplied element …
this decision lists no spelling normatively — the list is the plugin's".

What is wrong: an operator typing `-- --dangerously-bypass-approvals-and-sandbox`
(or the composer, if buggy) puts the unrestricted flag into `argv_suffix`; under
`MAY`, ax records `execution_profile: standard`, skips the ax §2.4 yolo
confirmation and the per-launch-event repetition, and launches a bypassed process.
The immutable record then misstates the profile — the exact property Option A
rests on. The decision's stated reason for not requiring the refusal ("the
vocabulary is the vendor's") does not hold: ax §7.7 already carries the
normative yolo spelling per built-in provider (`--dangerously-bypass-approvals-and-sandbox`
for Codex, `--yolo` for Muse, `--approval-mode=yolo` for Gemini), and a
`caller_launch_plan` plugin emits exactly those flags under its own profile
mapping. The information to refuse is already inside ax and the plugin; no
spelling needs freezing in 0013.

Fix: make it `MUST`: a `caller_launch_plan` plugin MUST refuse any caller element
equal to a flag of its own §7.7 `yolo` profile mapping (long or documented alias)
with `launch_plan_invalid`, `reason: "profile_flag"`, `details.argv_index`,
before process creation; the same rule for the `argv` form. 0013 still lists no
spellings; the rule keys on §7.7. Add the matching negative case to the PR's
required conformance (a suffix carrying the provider's yolo flag under
`--profile standard` is refused).

### F3 — minor — Decision 3.3: secret violation code conflicts with ax §15.3

Quote: "and, when the secret rule fired, `reason: "secret"`" under
`launch_plan_invalid`, exit class 2.

What is wrong: ax §15.3 already classifies secret findings as
`secret_policy_violation`, exit 16. Two codes for one condition, in different exit
classes, is the kind of split the PR reviewer will refuse.

Fix: secret-rule violations in a caller document refuse with the existing
`secret_policy_violation` (16), with `details.field`; `launch_plan_invalid` covers
shape, limit, exclusivity, and extension-collision only. Or justify the new code
against §15.3 explicitly.

### F4 — minor — Decision 4: `base64` contradicts ax §1.6 byte encoding

Quote: "`encoding` is exactly `utf-8` or `base64`".

What is wrong: ax §1.6 "bytes MUST be represented as a content-addressed blob or
unpadded base64url". A `base64` (standard alphabet, padded) member inside an ax
schema object breaks the common data rule the same section imposes on every ax
schema — and the decision cites §1.6 for the schema id in the paragraph above.

Fix: `encoding` is `utf-8` or `base64url` (unpadded, per ax §1.6).

### F5 — minor — Decision 6.4 / 3.2: composer collision between `env_literals` and `env_names` unspecified

What is wrong: `env_literals` = plan `Env` ⊕ fragment `env` ⊕ variable-kind
channel; `env_names` = fragment `mcp.env_names`. The reserved-name exclusion
(0012 D6) keeps registry adapter names out of `env_names`, but a system plugin's
`ChildEnv` (plan `Env`) is not bounded by that exclusion, so a plan `Env` name
equal to an allowed `mcp.env_names` entry produces a document that ax refuses
under §5.1 disjointness, surfacing as `ax_handoff_failed` with no composer rule.

Fix: one sentence in 6.3/6.4: on a name present in both, the composer drops it
from `env_names` (a literal the composer chose wins over a destination-local
lookup) and warns — or the reverse; pick one and state it, so the document is
disjoint by construction.

### F6 — nit — Decision 6.4 session name default will often exceed 64

`<env-id>-<profile-name>-<utc-stamp>` with the decision's own example profile
name (44 chars) + `claude_code` (11) + stamp (16) + 2 = 73 > 64, refused as a
`usage` error. The default is thus unusable for the example the document gives.
Consider `<env-id>-<stamp>` as the default with the profile name in the
extension key it already occupies, or document the consequence in Open
question 1.

### F7 — nit — Status/report: agents-management main moved

Decision pins `944c7b4` explicitly (correct at drafting time); main is now
`91bf945` (model-alias resolution, no LaunchMode/Lineup change). No text change
required; noted so the next cycle re-verifies against the pin, not `main`.

## Not verified by me (unchanged docs-confidence from the report)

- Idempotence of the current ax plugin `launch` operation (the decision requires
  it rather than asserting it — acceptable).
- Per-tool "everything after the last flag is the user turn" behavior — left to
  launcher SPEC 0.2 as instructed.
