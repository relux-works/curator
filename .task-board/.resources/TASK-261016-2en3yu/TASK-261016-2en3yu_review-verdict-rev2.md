# Revision 2 review: accepted

Scope: L15-1 slice 1, machine selector grammar, CLI, typed resolve transport and fail-closed unavailable selections. Exact candidate tree: `ddde39b56122ce8b78ecb49ab410914ab8693171`; base and unchanged Story HEAD: `dccac3199aa5c543c589c129f5ef2bc437523a2a`.

Coverage inspected before implementation review: **6 of 6 slice AC groups driven**, **3 of 3 selector CLI forms driven**. Named tests and production call sites are retained in `TASK-261016-2en3yu_rework-rev2.md`. This is bounded slice coverage: only **2 of 8 original whole-leaf paired rows fully driven**. Effective backend application, successor fragments, login/readiness and vendor qualification remain deferred to L15-1b. No claim of complete original-leaf implementation is accepted.

The previous `dotted-profile-selector-path` mechanism is fixed. Both original reviewer reproductions now pass, as do canonical/alias routing, repeated internal dots and invalid identifier tests. The rework keeps the complete profile operand, normalizes only the final environment operand and refuses invalid profiles before writes. A dotted-only narrowing mutant fails the named regression while plain-profile controls pass. The additional reserved-name mutant fails the public show assertion. No repeated blocking mechanism reproduced.

Fresh normative read: curator-spec main remains `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`. Reviewed login-readiness §§2/2.1/5, environments §§7.4/8.2/10.1–10.3/12.1–12.2, profile-codex-settings §§1–6 and common/subtree schemas against the explicit slice decision. Manager-config v5 allocation, profile identifier grammar, machine ownership and unavailable-delivery refusal agree within this slice. Deferred Muse/backend behavior is not a completed conformance claim.

## Surface sweep

# Revision 2 sweep
Candidate: ddde39b56122ce8b78ecb49ab410914ab8693171.
Coverage inspected before implementation: 6/6 slice AC groups with named driving tests; full original leaf remains explicitly 2/8.
Rework-first: retained dotted-profile reproductions and new invalid-profile/routing controls pass (rework.log, exit 0).

| Row | Result | Evidence |
|---|---|---|
| text grammar | held | config.Load malformed/null/unknown/alias/old-envelope attacks plus valid controls; grammar.log exit 0. |
| selector paths | held | Full public CLI set/show/unset and resolve with/without repair, invalid identifiers, unsupported stores across all adapters, state preservation and native fragment controls; selectors.log exit 0. All nine narrowing controls independently killed by named assertions, including dotted-only and reserved-profile-show (mutant-results.json). |
| error surface | held | Resolve unknown/malformed/permission native-reader attacks plus absent-key/keyring/auto controls; errors.log exit 0. Permission injected through existing I/O seam; no physical permission qualification. |

## Independent validation

- Rework first: `go test ./cmd/curator -run '^(TestReviewDottedCredentialProfile|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting)' -count=1`: exit 0.
- Grammar: `go test ./internal/config -run '^TestCredentialStore' -count=1 -v`: exit 0.
- Selectors: `go test ./cmd/curator ./internal/envprofile -run '^(TestCredentialStoreCLI|TestCredentialStorePolicyActivation)' -count=1 -v`: exit 0.
- Error surface: `go test ./internal/envprofile -run '^TestCredentialStorePolicyNativeInspection$' -count=1 -v`: exit 0.
- `go test ./internal/config ./internal/envregistry -count=1`: exit 0.
- Nine controls independently regenerated from candidate blobs and executed with Go overlays and `-count=1`: 9/9 exit 1 in intended named behavioral assertions. No setup/build failure counted as a kill. Definitions retained in the producer control resource; this review attaches fresh output.
- Free hunt: Go overlay extends the public CLI routing tests with valid 128-character profiles, repeated dots, nonreserved COM0, and invalid 129-character, reserved device, colon, newline and non-ASCII identifiers. `go test -overlay .temp/review-2/freehunt-overlay.json ./cmd/curator -run '^TestCredentialStoreCLI(DottedProfileRouting|InvalidProfileRouting)$' -count=1 -v`: exit 0; no additional finding.
- `git diff --check` and `git diff --cached --exit-code`: exit 0. All 16 changed paths match candidate blobs. Fresh origin/main equals the base; `git rev-list --count HEAD..main` is 0. README is the documented tool-command addition outside the original likely-path list.

Independent full hosted gate [38882423088](https://github.com/relux-works/curator/actions/runs/38882423088) completed/success: 20 successful jobs and two conditional skips. Head `f86c636bf7720ea0d5f3ad85e68df33e66c5e1a5` resolves to the exact candidate tree. Separate terminal API readback verifies the binding. Gate process exit is recorded in the evidence artifact. Producer build/vet/lint results are supporting evidence from the exact unchanged candidate, not represented as local reviewer reruns. Reviewer hosted workflow separately reruns those gates.

## Limits and logbook

No review/free-hunt wall-clock budget was supplied. The complete supplemental three-row surface table was swept. `held` means only that the named attacks did not reproduce. The free hunt is empty of findings, not a proof of absence.

Native permission failure uses the existing I/O seam; no physical permission or vendor runtime qualification is claimed. No source-text checker is introduced. The original hosted-override-drop plant is inapplicable to this narrowed slice; the independently killed CLI transport-drop control exercises its actual producer composition boundary.

Repository code is unchanged by review; test mutations exist only as scratch overlays. No Story commit, rebase, merge, or landing was performed. The full-gate scratch script changes only detached snapshot signing and uses configured identity plus signature verification, retaining the repository workflow and commands. It avoids the unsigned invented-author snapshot in the original script. No landing is claimed.

## Structured verdict

```json
{
  "verdict": "accepted",
  "revision": 2,
  "candidate_tree": "ddde39b56122ce8b78ecb49ab410914ab8693171",
  "findings": [],
  "notes": [
    "Only the explicitly narrowed slice is accepted; original whole-leaf completion and runtime qualification are not claimed.",
    "Previous dotted-profile-selector-path finding fixed with retained regression witnesses and killed narrowing control.",
    "All three surfaces held; free hunt produced no finding. This is not proof of absence.",
    "Native permission error uses the existing I/O seam.",
    "Full hosted reviewer rerun completed/success; no landing is claimed."
  ]
}
```

Evidence resource: `TASK-261016-2en3yu_review-evidence-rev2.md`. Acceptance routes the task to integrating; this review performs no landing.
