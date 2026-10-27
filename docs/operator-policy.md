# Operator trust and hook consent

CIP-0020 section 2.1 and Decision 0031 define operator policy-derived consent.
This implementation records the separate private
`workspace-trust-policy-grant-v1` event. It never creates an A1
`explicit-grant` on somebody else's behalf.

## Commands

```sh
curator env trust key --principal principal:operator --public-key BASE64_ED25519_PUBLIC_KEY
curator env trust grant --command signed-command.json
curator env trust revoke --command signed-revocation.json
curator env trust list
curator env trust read codex_cli /work/project --checkout CHECKOUT_ID --lane lane:build --profile base --lock-sha256 LOCK_DIGEST
```

The names `key`, `grant`, `revoke`, and `read` are implementation proposals:
the policy-grant amendment does not publish command names. They leave the A1
`trust add/remove/list <env> <path>` grammar available for separate adoption.
`list` here inspects policy entries and derived event history across environments.

`key` enrolls one operator public key in private manager state. An identical
repeat is idempotent; a different principal/key refuses. This is a local
operator bootstrap command, outside package/profile ingestion. There is no
key replacement or implicit enrollment from a signed request. The approving
operator's signing key remains outside Curator. Keeper automation supplies
signed command files; no native TUI response or manual configuration edit is
needed. The precise approval transport and CLI names require maintainer review.

`grant` admits exactly the trust boolean and hook ID/digest list in the
operator-approved policy. A new signed approval with a narrowed hook list
revokes every event derived from the old entry before recording the replacement.
`revoke` revokes the whole entry, including its hook consent. Replaying a
revoked approval cannot recreate consent. To revoke only hooks while retaining
project trust, submit a fresh signed `grant` command for the same entry with
an empty or narrower hook list. Every effective hook definition digest binds
its executable, arguments, trigger and execution-affecting settings.

Only one active entry may match a complete checkout/path/environment/lane/
profile/lock selection. A conflicting grant refuses without changing state.
Reads also refuse ambiguous legacy state, before consulting an adapter.

## Signed command input

The closed `operator-policy-command/1` body contains:

- `operation`: `grant` or `revoke`; `entry_id`, UTC `time`, nonempty `reason`;
- `issuer`: role `operator`, approving `principal_ref`, unique immutable
  `approval_ref` (64 lowercase hex), and `actor` with kind `operator` or `keeper`
  and its own principal reference;
- `actor_public_key`: base64 Ed25519 key authorized for this exact command;
- `entry_revision`: the entry's current monotonically increasing revision
  from `list.entry_revisions[entry_id]`; omit it for revision zero;
- for `grant`, `policy` and `selection`; for `revoke`, both are absent.

The envelope contains `command`, `approval_signature`, and `actor_signature`.
Both signatures cover the exact CCJ-1 bytes of the complete command body.
The first verifies against the independently enrolled operator key. The second
verifies possession of the actor key bound by that approval. An operator actor
must match the enrolled operator principal/key. Reusing `approval_ref` for
other bytes refuses. Caller role strings, package signatures and unsigned
requests confer no authority. Duplicate/unknown/misspelled members refuse.

The revision counts retained successful grant/revoke commands and terminal
lapses for that entry, including commands for removed policies and revokes
with no event. A lapse advances it in the same atomic write as the terminal
event state. Moving the lapse into history on a later grant or revoke retains
that increment; listing, repeated denied reads and serialization never count
it twice.
An absent entry in `entry_revisions` has revision zero. Every new command must
bind the current revision in both signatures. Accepted commands increase it;
rejected commands and idempotent retries do not. Revoking or lapsing an entry
therefore invalidates pending unused approvals for that entry, even if their
signed timestamps equal or follow the transition's timestamp. A keeper must obtain fresh
operator approval after reading the new revision; changing the revision alone
invalidates the signatures. This boundary also serializes competing updates.
Existing audit commands with no revision remain verifiable at their original
bytes, but cannot be newly applied to an entry with a nonzero revision.

Stale revisions and ambiguous entries use the existing refusal code
`workspace_trust_policy_authority` with distinct sanitized reasons. CIP-0020
and Decision 0031 provide no dedicated code for either case; maintainer
confirmation of that diagnostic mapping remains an open question.

`policy` contains the complete CIP-0020 entry: `entry_id`, `root_commit`,
`canonical_destination`, `remote_spellings`, `environments`, `lanes`,
`project_trust`, `hooks`, `issuer`, `time`, `reason`, `review_by`. Its issuer,
time and reason equal the command's. This first implementation admits one
direct canonical spelling, one selected checkout/environment/lane per command,
and an existing protected project registration. Origins must use the exact
canonical URL; SCP shorthand and aliases are refused. Hook lists are explicit,
including an empty list. No hook payloads or credentials belong in these inputs.

`selection` contains `checkout_id`, exact absolute `path`, `environment`,
`lane_id`, and `profile` with `name` and `lock_sha256`. The registration,
retained filesystem identity, repository tuple and installed lock are checked
under the manager mutation lock. The policy digest is SHA-256 of the complete
entry's CCJ-1 bytes. Approval, event, selection, policy replacement and revocation
history are committed in one private atomic state write.

Private state lives at `<manager-home>/operator-policy/state.json`; it is
excluded from portable profiles, locks, snapshots and packages. Unknown event
kinds, malformed/missing allocated state, changed policy digests, unreadable
state and foreign ownership journals fail closed. This release projects no
native keys, so its ownership journal and touched-home lists are empty. It
refuses a foreign/nonempty journal rather than guessing cleanup ownership.
Same-UID tampering remains the explicit pre-CIP-0024 integrity limitation.

## Per-start delivery and limits

`read` returns effective consent only after all required admission checks.
Unknown/unmatched projects refuse with `workspace_untrusted`. Revoked events
refuse with `workspace_trust_grant_revoked`; profile/lock drift refuses with
`project_context_drift`. Changed repository/filesystem identity and expiry
retain a terminal lapse, requiring fresh consent. Refusals emit no stdout.

A missing, renamed, non-directory or linked registered checkout retains that
lapse before the initial `workspace_trust_symlink` refusal. Restoring the old
directory still refuses `workspace_untrusted`; only a new signed grant at
the current entry revision can create new consent. Prior lapse history remains.
`list` is read-only and retains prior policy digests, issuers and state history.

**No real provider release is qualified for policy-grant destination evaluation
and hook discovery/native approval/reload/revocation by this implementation.**
Consequently production `read` refuses `workspace_trust_adapter_unqualified`
even for a valid grant with no hooks. Tests of successful reads use an internal
observer and prove core policy checks only. There are no CLI qualification
flags, seed fallbacks or hook bypasses. This is a command/storage and gated
read implementation, not an activated unattended launch path.

The read API is independent of proposed effective-posture/1 (spec PR #197).
It emits no launch fragment, prepares no lease, starts no provider, and writes
no managed/native home or profile lock. The launcher must consume a fresh
successful read at every start/resume/restart once adapters are qualified;
a cached delivery never extends consent. Permissions, sandbox selection,
credentials, memory-off, native-project/strict mode and project-item admission
remain independent operator/runtime policies. A trust grant changes none of them.

## Validation

Run inside an already scrubbed test child: remove every inherited `GIT_*` and
`XDG_*` variable, SSH_AUTH_SOCK/SSH_ASKPASS, GOFLAGS and vendor-home overrides
before launching Go. Use an absolute temporary validation checkout and cache
paths. The flags below repeat the minimum cleanup; they alone do not scrub
all Git overrides. HOME uses an absolute mktemp template:

```sh
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS -u GIT_CONFIG_COUNT -u GIT_CONFIG_PARAMETERS HOME=$(mktemp -d /tmp/curator-policy-home.XXXXXX) go test ./internal/operatorpolicy -count=1
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS -u GIT_CONFIG_COUNT -u GIT_CONFIG_PARAMETERS HOME=$(mktemp -d /tmp/curator-policy-home.XXXXXX) go test ./cmd/curator -run '^TestOperatorPolicy' -count=1
```

Both test packages call the neutral `testgitenv.Isolate` in `TestMain` and pin
`USERPROFILE` to the resulting fixture HOME on every OS. The helper owns HOME
and XDG config directories and scrubs inherited Git/SSH/native-root overrides.
Explicit fixture homes are transported to CLI helpers and PATH/native children
after isolation. An absent neutral HOME boundary fails before tests run.

Until PR #196 lands, validation requires its exact neutral helper as a source
overlay: head `613e730d847476f0f422e1474f001f5a94e99f66`, file
`internal/testgitenv/testgitenv.go`, blob
`5e84fdd4342f937e27ef65c19d6bc0c9fe5eadf0`. That dependency is not copied into
this PR. Only directly relevant CLI fixture-home transport is reused here.
No separate Git-prefix scrub implementation is introduced. Test launchers must
clear every inherited `GIT_*` and `XDG_*` variable, `SSH_AUTH_SOCK`, `GOFLAGS`,
`CODEX_HOME`, and `CLAUDE_CONFIG_DIR`, with only explicit fixture overrides
allowed. All HOME, working-directory and cache paths must be absolute and
temporary. Logs belong under `.temp/`.

| Contract/vector | Paired production-path test |
| --- | --- |
| Package trust/hooks, forged or absent approval/actor, changed signed scope | `TestPolicyAuthorityPairs`, `TestPolicyPackageDeclarationPairs`, `TestOperatorPolicyCLI` |
| Revoked consent on next start, replay/idempotent revoke | `TestPolicyRevocationNextStart` |
| Unused pre-revoke approval, equal/newer signed clocks, fresh post-revoke control | `TestReviewerDelayedApprovalAfterRevoke` |
| Revocation without an event, retained revision, forged revision, fresh signature | `TestPolicyRevisionRetainedWithoutEvent` |
| Missing checkout/restore terminal lapse and new approval control | `TestReviewerMissingCheckoutLapse` |
| Pre-lapse future/equal-second approval, post-lapse signature positive | `TestR2PendingApprovalAcrossLapse` |
| Lapse counted once, retained across read/list/history and later revoke | `TestPolicyLapseRevisionCountsOnce`, `TestR2LapseSurvivesRewrite` |
| Duplicate active write and distinct lane control | `TestReviewerAmbiguousEntries` |
| Authenticated legacy ambiguity in both event orders, public/core reads | `TestReviewerLegacyAmbiguousRead` |
| Changed effective hook, unlisted native approval, empty list, unreadable hook state | `TestPolicyHookPairs` (internal observer only) |
| Exact project/environment/lane, selected or installed profile/lock drift | `TestPolicyScopePairs` |
| Policy digest, missing/malformed private state, unknown event, audit/journal | `TestPolicyPrivateStatePairs` |
| Entry edit revokes former events and retains provenance | `TestPolicyEditRevokesOldEvent` |
| Deadline boundary and durable lapse | `TestPolicyExpiryLapse` |
| Lost/restored inode, unchanged root after ordinary commit, incomplete history | `TestPolicyFilesystemLapse`, `TestPolicyRepositoryHistoryPair` |
| Pinned-key replacement is refused | `TestPolicyOperatorPin` |
| Published event schema/semantic cases and valid control | `TestPolicyPublishedEventSchema` |
| Closed request parser; duplicate/missing/unknown/case-alias controls | `TestSignedCommandClosedParser` |
| Production refuses unqualified adapters, including empty hooks | `TestProductionReadQualification`, `TestOperatorPolicyCLI` |
| CLI package startup home isolation | `TestOperatorPolicyHomeIsolation` |
| Explicit child fixture HOME, malformed carrier and PATH/native-owner controls | `TestCLIUnitFixtureHomesPreserveExplicitSelection`, `TestCLIUnitFixtureHomesRejectInvalidPayloads`, `TestCLIPathIsolation`, native-owner tests |
| Inherited Git overrides removed before fixture signing | `TestSignerRowsSurviveHostileGlobalGitConfig`, package startup assertions |

The embedded event schema and event schema corpus are pinned verbatim from
curator-spec `e1051a7ad8e7e8e91637a22d6347755754ea0953`. They establish
specification compatibility. They do not certify runtime vendor behavior,
SSH routing evaluation, transport host pins or active sandbox policy.
