# TASK-261021-i2f2uq independent review, revision 1

Scope: proposal and its two attached evidence/results resources. Curator head: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`. Launcher head: `dd70beb42239fc80f81bcfd512358f8d3844b615`. No implementation branch or introduced commits exist; both snapshot commits verify `G`. Every precondition resource was read, including the named reviewer assignment. No repository content was edited.

## Surface sweep

| Surface | Held or broken | Independent attack / inspection | Evidence |
| --- | --- | --- | --- |
| F1 documentation maturity | Held | Compared root help with nested profile help and launcher FRONT-DOOR opening with README. | Root omits import; nested help includes it. FRONT-DOOR says nothing implemented; README says partially implemented. |
| F2 provider/nested help | Held | Executed umbrella register help, direct register help, complete station status help, and root launcher help. | Exits 0/2/2/0; umbrella prints generic run help. |
| F3 policy-dependent help | Held | Real launcher entry point with absent then malformed fixture launcher.json. | Root help changes from 0 to 1/defaults_config_invalid. docs/cli.md lines 9-11 and 632 contradict each other. |
| F4 spelling | Held at supplied head | Registered a second fixture manager without selection. Compared upstream feature branch separately. | Notice uses curator run; upstream feat/q1-signoff-activation changes it to curator-run. Upstream fix does not alter this snapshot verdict. |
| F9 registration/ownership | Held | Registered first and second managers; re-registered first; removed selected first; registered with no-select. | First selected, peer preserved selection, same-ID timestamp changed, removal selected nobody, no-select preserved absence. service.go and werkstatt/service_unix.go confirm the binding includes RegisteredAt. |
| F10 no install hook | Held | Re-read README security promise and FRONT-DOOR registration sections; inspected actual task-board wrapper help. | Zero code execution during install; explicit wrapper syntax exists. This is a proposal constraint, not a claim that lazy registration was executed. |
| F11 parser guidance | Held | Invalid station JSON flag and incompatible recovery flags through real binary. | Exit 2 with generic expected-grammar text. |
| F17 trust | Held | Put a harmless provider outside trust roots on PATH and invoked it through the actual Curator binary. | Warning subcommand_provider_outside_trust_roots, provider ran, exit 23 propagated. Registration resolves an absolute executable. |
| F18 help admission | Partly broken | Ran compose add help and inspected commandRegistry; independently inspected global add, install flag reuse and update dispatch. | Compose already advertises tag and revision. Other cited global discrepancies hold. R1 below. |
| Inventory | Held | Executed all 72 registry topics with help and root help; checked launcher launch/admin parsers and searched Curator flag declarations against proposal. | All 72 exits 0; no missing command topic or flag found in this sweep. Additional administrative rejection probes matched inventory. |
| Grammar, every row | Partly broken (P2 compatibility clarification) | Compared profile/env/project/run/station/manager/global rows with handlers and alias policies A/B. Exercised out-of-range queue depth and mutually exclusive recovery actions. | Existing specialist semantics and stable wire IDs are retained; no additional broken rename found. Implicit registration scope migration needs a concrete compatibility rule: R2. |
| Hook-free/lazy registration | Held | Traced proposed trigger, trust, authority, transaction and ownership constraints against registry implementation and independently exercised timestamp/selection transitions. | No package hook or permissive PATH authority proposed. Lazy discovery remains explicitly unimplemented and conditional. Explicit wrapper failure exits remain unverified, as producer states. |
| Ownership/top five | Held | Checked each slice against Curator dispatch, launcher SPEC/FRONT-DOOR sections 4/6/11/12 and wrapper help. | Owners are consistent. Upstream spec/project-management exact edit locations are explicitly deferred, not asserted as reviewed trees. |
| Public hygiene | Held | Read all delivered resources and scanned proposal for private home paths and personal/host identities; checked source links. | No personal path/name found in proposal. Generic example paths and task identifiers are not personal disclosures. |
| Free hunt | Held | Checked native/hosted conflict, forbidden down JSON, queue limit, recovery exclusivity, provider exit propagation and no-select with no current selection. | Intended refusals and preserved selection observed; no extra blocking proposal issue found. |

## Findings

R1 (P2): proposal line 109 says tag/revision are omitted from the registered synopsis; F8/F18 repeat the claim and the top-five list includes a compose-help fix. At this exact head command_help.go already advertises both. `curator profile compose example add --help` exits 0 and prints both flags. Correct those claims and narrow the recommended fix; the rest of F18 remains supported.

R2 (P2): proposal line 249 changes implicit registration destination to user scope while invoking alias policy B. That policy covers old spellings, but the old and new commands have the same spelling. The production fixture with writable system directory writes session-managers.json there and leaves user config absent. Define how omitted --scope preserves legacy targeting during the compatibility period, or explicitly label this as a separately versioned behavior change. This is a future proposal clarification, not an implemented privilege regression.

## Independent execution evidence

Private scratch evidence: `.temp/TASK-261021-i2f2uq/` under the review workspace. `probes.json` contains 88 production invocations (72 help topics included); `extra.json` contains seven further attacks; `scope.json` contains the writable-system destination probe. Assertion scripts exited 0. Launcher fixture changes only the two configuration-root link variables to isolated directories; parsing, registry writes and dispatch use production code. Registration used an inert executable and launched no manager. The writable fixture ancestor means registry writes landed in the relocated system directory; no real system registry was changed.

Both binaries built successfully with Go 1.25.5 on darwin/arm64. Normal-.git validation clones were checked out at the exact two heads; signature verification uses the same configured allowed-signers trust as the source clones. No introduced commit range exists for this resource-only task.

Exploratory failures: inherited GIT_DIR initially routed the launcher fetch to the wrong repository; explicit sibling git-dir with GIT_DIR/GIT_WORK_TREE unset resolved it. Two early builds failed because the short temporary directory did not yet exist; it was created and builds then passed. Standalone validation clones do not inherit repository-local signature configuration; explicit use of the existing allowed-signers setting verified G. None of these failures is counted as passing evidence.

Reviewed resource SHA-256 values (materialized board bytes):

- cli-review: `7fd9c81bd82a18fea897895f4c4fcd99e147f921c358ac8f82e3bb2a499363ce`
- help-evidence: `c31fdf8a2abef9939c3d2f4ec985b6adef265b3d5897101176cbf7a8e1354568`
- results-r1: `ad4afed3c4e44b7c951f071d2d43d12580475a514f31278b95a9c13909b3161f`

### Suite results and limits

Every test command used `env -u GIT_DIR -u GIT_WORK_TREE` and a short workspace-local `TMPDIR` (`.t`). Both full suites used normal `.git` clones at the exact reviewed heads, not linked worktree metadata. Commands below run from the respective clone.

| Command | Exit | Evidence / limit |
| --- | ---: | --- |
| Curator `go test ./...` | 1 | `curator-suite.log`: cmd/curator, artifactpolicy, crossconformance, envprofile and install reached the default 10-minute package deadline. Two rustsource tests failed because rustc was absent from PATH. Full-suite validation is not green. |
| Launcher `go test ./...` | 1 | `launcher-suite.log`: cmd/curator-run reached the default 10-minute deadline during TestWerkstattAlignmentBinary. All other launcher packages passed, including conformance and sessionmanager. This is not a passing launcher suite. |
| Launcher `go -C schemas/session-manager test ./...` | 0 | `schema-suite.log`: nested schema module passed. |
| Curator targeted help/passthrough/compose refusal tests, first attempt | 1 | `targeted.log`: TestMain could not acquire the package host GOROOT lock within five minutes while the full suite was running. No target-test pass is claimed for this attempt. |

The full-run failures are recorded as validation limitations of the supplied, unchanged snapshots. They do not establish a proposal defect, and this document verdict is not implementation/release qualification. No toolchain installation, test suppression, source patch or lock bypass was used to turn them green.

Targeted retry after the full run released its lock: `go test ./cmd/curator -run 'TestRegisteredSubcommandHelpNoWrites|TestCommandHelpRunPassthrough|TestProfileComposeRefusals' -count=1` exited **0**, package duration 62.985s (`targeted-retry.log`). This covers all registered help positions and persistent no-write checks, provider native-tail passthrough, and invalid compose combinations. Both normal validation clone working trees remained clean.

## Verdict

Accepted as a proposal, with two nonblocking P2 corrections. No P0/P1 proposal finding. Full-suite failures remain disclosed and are not waived as implementation evidence.

```json
{
  "verdict": "accepted",
  "head": "cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7",
  "findings": [
    {
      "id": "R1",
      "severity": "P2",
      "file_line": "TASK-261021-i2f2uq_cli-review.md:109",
      "mechanism": "The claim that registered compose help omits tag/revision is false at the supplied snapshot; repeated in F8/F18 and cleanup recommendations.",
      "reproduction": "Build cdcba3bd and run curator profile compose example add --help; exit 0 and both --tag and --revision are present in command_help.go and runtime output.",
      "required_fix": "Remove or correct the omission claims and narrow the suggested compose-help cleanup; preserve the independently supported global-help findings."
    },
    {
      "id": "R2",
      "severity": "P2",
      "file_line": "TASK-261021-i2f2uq_cli-review.md:249",
      "mechanism": "Changing omitted --scope to user scope changes the behavior of an unchanged command spelling; alias policy B alone does not preserve current implicit system targeting.",
      "reproduction": "Production launcher with isolated writable system config root: session-manager register --manifest fixture.json exits 0 and writes the system registry, leaving user configuration absent. Proposed default instead targets user scope.",
      "required_fix": "Specify a compatibility rule for omitted --scope during the alias window, or explicitly identify the default change as a separately versioned behavior migration."
    }
  ]
}
```
