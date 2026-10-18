# Rework round 3 — second recovery: CR validator signing blocker

Task: BUG-261018-uoisfe — in-place-profile-skills-miss-native-surface.
This run made no repository source changes. The row-12 fix remains at
cmd/curator/main.go:283 and cmd/curator/command_help.go:128; behavioral sweep,
regression witness and list-only narrowing mutant remain in
cmd/curator/native_skills_flags_test.go:17, :54 and :102.
Rows 1–11 are unchanged; coverage remains the previously documented
12 of 12 attack-surface rows and 5 of 5 AC rows, with the production call
sites, positive-control bounds and retained reviewer names in the existing
coverage map. Muse, opencode and managed homes remain out of scope.

Candidate: 3e00fcc8abe75e411a88773c65c4042ae2448eb0, signature G.
Observed worktree tree: 2fe068da665b203ed48e52cfcb7fba74de96e37e,
identical to the signed candidate. Story HEAD stays at checkpoint
5d8e6b07f7f485557d9bbb065c9692348ca065c2; all changes remain uncommitted.
Fresh origin/main is 9c1605a74436d2448f759bd4b66ae6685237d83b;
the changed-path intersection is empty and root README.md is unchanged.

| Check | Real exit | Evidence / bound |
| --- | ---: | --- |
| `go test -count=1 -timeout=4m -v ./cmd/curator -run '^TestNativeSkillsTakeover(EveryRegisteredCommand|ArgumentDiscardRegression|ActivationBoundary|RegistryNarrowingMutant)$'` | 0 | Four tests; all 67 unrelated command cases; run directly in this run |
| `go build -o .temp/BUG-261018-uoisfe/resume2-curator ./cmd/curator` | 0 | Direct build in this run |
| `git diff --check` | 0 | Empty diagnostic log |
| Temporary-index identity / signature / fresh-trunk overlap verification | 0 | No real index or branch changes |
| Host-prescribed environment + task-local `ssh-keygen -Y sign -n git` probe using the configured public signing key | 0 | Agent signing works in the correctly bootstrapped shell; signature output stayed private and ignored |
| Fresh provider read of full hosted gate | 0 | Run 39117060646, attempt 2, completed/success on the exact signed candidate |
| CR revision 2 runtime `sh scripts/remote-gate.sh` | 1 | Previously recorded; signing fails before hosted execution; not green |
| CR revision 3 runtime `sh scripts/remote-gate.sh` | 1 | Previously recorded; same signing failure before hosted execution; not green |

Full local suite, lint, vet and platform-case evidence are reused from the
unchanged candidate and environment, not claimed as freshly rerun: 85 packages,
3494 top-level names, 28 CLI/envprofile groups, with exact prior/current identity
bounds already attached in validation-r3 and its 121-payload evidence bundle.
Fresh provider URL: https://github.com/relux-works/curator/actions/runs/39117060646
No new full hosted run was dispatched; the exact candidate already has a green
full hosted gate. That result does not overwrite the two failing CR validations.

| Mutant | Gate narrowed to | Named tests that fail | Survivor bound |
| --- | --- | --- | --- |
| list-only-consent-bypass | Only `list` may silently accept takeover; all other consent rejections stay present | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Killed, child exit 1; no survivor |

## Constraint, recovery attempts and required owner action

The successor prompt described a missing handoff, but persisted notes and CR
validation logs show both preceding runs called handoff and the post-turn CR
construction subsequently failed. The failure is not a missing product change.
Both runtime logs stop at `git commit-tree -S` in scripts/remote-gate.sh:47:
SSH signing tries to load the private signing-key file and requests a passphrase.
There is no test, build or hosted result from those failed commands.

Sourcing the host-prescribed `agent-git-env` works in this run's shell, and a
fresh task-local signing probe exits 0. The preceding recovery already sourced
that environment before handoff; CR revision 3 still failed afterward. Therefore
shell-local bootstrap is not evidence that the post-turn validator has the same
environment. The runtime's actual environment was not exported or inspected;
its precise missing/overriding variable remains unverified.

The supported handoff, worktree and spawn CLI help exposes no producer command
to synchronously republish this candidate with an explicit validator environment.
Changing the gate, repository config, shared git config or runtime installation
to hide the failure would exceed the row-12 rework scope. Restarting another
session's shared runtime is prohibited by the host state-preservation rules.
No private key, alternate socket, unsigned commit, forged hosted status or
validation bypass was used. No control-root file was directly edited.

Recommended recovery: the runtime owner must bootstrap the process performing
post-turn CR validation with the prescribed agent transport/signing environment
(configured public signing key and swarm socket), then retry the producer's
normal CR publication. If that process needs a restart, its owner must arrange
it safely between affected sessions. Unlocking keys is needed only if that
correctly bootstrapped process actually reports unavailable identities; the
current shell signing probe shows available signing authority.

Alternatives: an owner-supported per-validator environment configuration would
avoid a shared-runtime restart if available. A product-code or gate-script
workaround would widen this leaf and is not recommended. More no-change
handoffs from this shell would repeat the already disproven recovery.

Status: blocked on runtime-owner signing-environment repair. No handoff is
claimed for this run, and no further unchanged CR attempt is knowingly queued.
The existing gaps remain: Muse multi-root ownership, adapter-ledger interplay,
and opencode's machine-global native surface.
