# Revision 1 review: changes requested

Scope: L15-1 slice 1, machine selector grammar, CLI, typed resolve transport and unsupported-selection refusals.

P2: a valid dotted profile cannot be addressed by the new selector CLI. The production JSON parser accepts `team.prod`, as does the normative identifier schema, but the CLI splits that profile into two keys. `show` silently returns `null` for an existing `file` selection; `set` and `unset` refuse. Preserve the whole profile operand, add named regression cases for set/show/unset (canonical and alias environment forms), and a narrowing mutant that breaks only dotted-profile routing. Keep plain-profile controls and failed-write state assertions.

Normative source: curator-spec main `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`, login-readiness section 2 and its credential-store-config-v1/common identifier schemas. Fresh API read agrees with producer pin; protocol sources match the producer copies. The identifier grammar explicitly admits internal dots.

Coverage checked before source inspection: 6 of 6 slice AC groups have named driving tests and production call sites in the handoff. This measures presence, not correctness: the CLI identity group misses the valid dotted-profile class. All 3 supplemental surface rows were attacked. The source remains unmodified and all 16 changed paths match the exact candidate. README is the justified extra path. No prior review finding exists for this revision.

# Review sweep, revision 1
Candidate tree: 7efa9fa6153d25c5cf051aaa77661ddf9a1803e7

| Row | Result | Evidence |
| --- | --- | --- |
| text grammar | held | Re-ran TestCredentialStoreConfigGrammar and VersionBoundary through config.Load; malformed/null/alias/unknown values refused beside valid controls. focused.log exit 0. |
| selector paths | broken | F1 dotted-profile-cli: TestReviewDottedCredentialProfile/team.prod refuses a valid profile while /team passes; dotted-profile.log exit 1. |
| error surface | held | TestCredentialStorePolicyNativeInspection exercised Resolve: permission, malformed and unknown native storage refused; absent-key/keyring/auto controls reached the intended unsupported-delivery refusal. error-surface.log exit 0; permission injection uses existing I/O seam. |

## Validation

- `go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run '^TestCredentialStore' -count=1`: exit 0, independently rerun.
- `go test ./internal/config ./internal/envregistry -count=1`: exit 0, independently rerun.
- `go test ./internal/envprofile -run '^TestCredentialStorePolicyNativeInspection$' -count=1 -v`: exit 0, independently rerun.
- Seven producer narrowing controls independently rebuilt from exact candidate bytes and rerun with Go overlays and `-count=1`: all exit 1 in their intended named behavioral assertions, no build/setup failures. See mutant-results.json and sanitized logs.
- Reviewer overlay tests: exit 1, three reproduced symptoms; plain-profile positive control passes. Replay script and test source included in the task-scoped evidence bundle.
- `git diff --check`: exit 0.
- Original full hosted gate independently read back completed/success, snapshot tree equals the exact candidate. Reviewer full-gate rerun https://github.com/relux-works/curator/actions/runs/38864359459 was cancelled after the completed surface sweep reproduced the blocking defect. The owned local polling process was stopped (exit 137) and its throwaway branch deleted. It is NOT a passing or completed full-gate result; earlier producer green is reused only as exact-tree supporting evidence. This is an explicit limitation against the brief's requested completed independent full-gate rerun. The scratch copy changes only snapshot signing/identity and verifies the signature before publication, preserving the repository script's jobs and commands. No Story branch commit/index mutation was made.

## Logbook

The failure is a selector-path identity defect, not a platform qualification issue. Route the complete finding once to implementation rework; do not create another research leaf. No independent additional finding reproduced in free hunt of schema promotion and selector/alias call sites.

## Structured verdict

```json
{
  "verdict": "changes_requested",
  "revision": 1,
  "candidate_tree": "7efa9fa6153d25c5cf051aaa77661ddf9a1803e7",
  "findings": [
    {
      "id": "dotted-profile-selector-path",
      "row": "selector paths",
      "invariant": "Every valid profile identifier, including team.prod, remains one profile operand in credential_store.<profile>.<env-id> set/show/unset; show reports the selected value.",
      "mechanism": "internal/config/environments.go:1360 splits every dot into a path segment; cmd/curator/envalias.go:35 normalizes segment 2 and cmd/curator/envconfig.go:70 handles effective selection only for exactly three segments. The new selector CLI therefore cannot address valid dotted profile names.",
      "reproductions": [
        {
          "test_file": "attacks.go.txt (overlaid on cmd/curator/env_credential_store_test.go)",
          "tests": [
            "TestReviewDottedCredentialProfile",
            "TestReviewDottedCredentialProfileExisting"
          ],
          "command": "python3 replay.py (from candidate repository; replay.py and attacks.go.txt from the evidence bundle)",
          "expected_failure": "Exit 1: team positive control passes; team.prod set refuses unknown canonical environment prod; existing team.prod show returns null instead of file; unset refuses not set.",
          "log": "dotted-profile.log",
          "candidate_tree": "7efa9fa6153d25c5cf051aaa77661ddf9a1803e7"
        }
      ],
      "severity": "bypass",
      "priority": "P2",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "The same path mechanism also causes valid set/unset regressions. One finding covers all three symptoms.",
    "Slice 1 only: effective backend application, runtime qualification, login/readiness and hosted override transport remain explicitly deferred.",
    "No review/free-hunt time budget was supplied. Supplemental producer surface table was used in its given order.",
    "Native permission failure attack uses the existing read seam; this is not a physical permission-mode qualification."
  ]
}
```
