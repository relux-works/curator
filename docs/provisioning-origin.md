# Provisioning origin and retained homes (unreleased)

This describes the source implementation in this branch. It is not a release,
installation, provider qualification, or launcher activation announcement. The
normative contract is [provisioning origin](https://github.com/relux-works/curator-spec/blob/a9c5e03b2857a10a66433e1a60bb255a89b9d741/protocol/provisioning-origin.md),
together with the existing environment provisioning and lifecycle rules.

## Capture and maintenance

Before first managed-home publication, Curator resolves the native account
anchor independently of `HOME`. It captures the original selector's presence
and value: absent, explicitly empty, and nonempty are distinct. A selector
inside a managed namespace is excluded using lexical and physical containment,
including existing case aliases and aliases whose final tail is absent. Errors
while inspecting an ancestor are not absence. Pi's selector is the agent root
itself: `/fixture/pi/agent` targets `/fixture/pi/agent/auth.json`; Curator does
not append another `agent`.

The owner-private, regular, no-follow `.agent-provisioning-origin.json` record
uses `provisioning-origin/1`. Its binding names the exact profile, lock and
managed target. It contains provenance and locators, never credential contents.
The account lookup transport is implemented for Darwin/Linux with cgo. Other
target/toolchain combinations refuse initial capture; changing `HOME` does not
provide an account-lookup substitute. Fixture account lookup tests are not real
OS account or native provider qualification.

Existing homes use only their independently verified recorded native root A,
even when today's caller selects B. Maintenance does not infer authority from
`HOME`, a default native directory, a credential link, marker provenance labels,
or the existence of a journal. A substituted root or unsafe parent refuses;
ordinary repair does not bless drift, recapture history, change backend policy,
or relink a wrong-target credential. Lock transitions preserve capture facts
and rebind every materialized home, including inactive adapters.

Credential custody is separate from native skills, context, seeds, settings,
hooks and sessions. OpenCode reconciles its recorded XDG config seed links
against the current operator XDG parent as required by its seed contract; an
unrecorded shadow is preserved with a warning. Muse does not seed foreign XDG
config entries. Neither operation retargets ambient XDG data/state directories.
Copy-seed contents remain tool-owned after provisioning.

A missing credential leaf differs from a missing or invalid origin root.
Codex and Linux Claude retain their supported detached-pending behavior; Pi
shared resolution reports `environment_credential_conflict reason=dangling`
without a fragment, and Muse shared file login retains its strict liveness
requirement. Ambient backends and isolated login-pending homes retain their
own policy. Regular credential forks, wrong links, unreadable metadata and
broken ancestors do not become empty/login-pending states. Curator never uses
credential copying as a link-capability fallback.

## Explicit legacy attestation

An existing home without a verified origin returns
`environment_provisioning_origin_required`. An unsafe or inconsistent record
returns `environment_provisioning_origin_invalid`. Preserve the home and
establish its history independently. Ordinary resolve, repair, reinstall and
selector changes are not adoption routes.

This branch adds the following explicit attestation grammar:

```sh
curator env adopt codex_cli --profile demo \
  --managed-target /fixture/manager/environments/demo/codex_cli \
  --lock-sha256 sha256:REPLACE_WITH_EXACT_INSTALLED_64_HEX_DIGEST \
  --operator-home /fixture/operator \
  --selector-state present-value --selector-value /fixture/native-a \
  --native-root /fixture/native-a
```

These are fixture placeholders, not a command to run against an existing home.
Every flag shown is required. For `--selector-state absent`, omit
`--selector-value`; for `present-empty`, supply `--selector-value ''` explicitly.
The operand accepts the existing `codex`/`claude` aliases. The operator attests
original facts; Curator validates the exact installed binding, physical root,
backend and existing credential entries before publishing only the missing
origin record. Matching repeated attestation is idempotent; a different existing
capture refuses. Adoption neither repairs links nor takes over settings or
credential files. An unsealed legacy journal cannot supply the missing history.
The trust boundary is the account owner's private state; a hostile same-UID
writer is outside this schema contract.

## Publication and recovery

Managed and native publication uses complete, phase-bound transaction evidence:
profile identities, every affected home/marker/origin, exact target/preimages,
parent existence and protection, durable boundary proofs and subsidiary
publications. Credential entries are recorded as no-follow entry metadata/link
text. Arbitrary home/session trees and native credential payloads are never
snapshotted. Directory metadata actions and private-file/tree protection survive
restart; a process-local undo callback is insufficient.

Returned ordinary failures restore the admitted mutable preimages. Genuine
interrupted units recover according to their saved preparing, forward, rollback
or cleanup phase. Cleanup failure after commit remains pending cleanup, not a
claim of complete rollback. Unrelated or substituted journals do not authorize
continuation. Old unsealed link-only migration journals are preserved and refuse
all continuation; only a completed legacy receipt with independently verified
recorded postimages can be cleaned without changing those postimages.

Internal `environment-lifecycle/1` state records retained outcomes after journal
cleanup. It does not add an origin carrier, public marker/lock schema, native
selection rule or launch authority.

## Global retained conflict

The existing handled `environment_surface_unmanaged_conflict` outcome preserves
the newly published global lock/store L1, plus narrowly required first-use source
and private receipt evidence. Every prior mutable home, origin, marker, native
surface and activation remains at its invocation preimage. This state is
retained and nonlaunchable, not ordinary stable lock drift.

Plain restart only compensates or completes receipt cleanup according to its
saved phase; it does not take over a surface. Explicit `curator profile sync
--takeover` or `curator profile use NAME --takeover` may finish the recorded whole
group, preserving original capture. Retry the global operation separately to
produce L2. A failed takeover restores the retained L1 baseline. Other failures
do not inherit this exception: they restore L0. Manually changing a lock without
bound durable intent never authorizes rebind or repair.

## Remove, later purge, and name reuse

Successful `curator profile remove NAME` removes installed lock/config and
revokes live references while retaining owned homes, origins and session bytes
by default. A permanent removal receipt survives final journal cleanup. The
retained profile is intentionally nonlaunchable; unrelated valid profiles,
leases, install and GC can continue. This is not a blanket orphan exemption,
and removal history does not pin the old store closure as live.

This branch extends the existing `curator profile remove NAME --purge` operand
to an absent installed profile only when its unique, valid removal receipt
proves the exact retained inventory and namespace. Partial, unreadable or
substituted installed state is not absence. Purge quarantines whole owned homes
with no-follow metadata identity and boundary checks; it does not read, hash,
copy or follow credential/session contents. Precommit failure restores the
original directory identity; postcommit cleanup remains pending until complete.
Terminal absence is verified before an idempotent retry succeeds. Fresh same-name
installation requires that terminal receipt and consumes it atomically before
new origin capture; a retained same-name home or different-lock substitute
refuses. `env unmanage` remains a native-in-place operation, not managed purge.

## Contribution and evidence limits

Physical containment, no-create/error handling and nearby fixture/assertion
kernels adapt material from [PR #179](https://github.com/relux-works/curator/pull/179),
originally authored by Alexis, at signed commits
`50f488dbfe09d22d4c8fd6df7d1c57aa51e2c020` and
`2166919225995cbfdb702c67ffad03c664ba64d5`. This branch preserves attribution;
it does not imply original-author agreement, transfer signed authorship, or
claim that the bounded contribution implemented recorded origin or lifecycle.
Unpublished successor material has not been reconciled here.

Named synthetic tests exercise raw account-input selection, A/B maintenance,
backend refusals, transaction/recovery owners, all-home rebind, retained global
conflict, remove/purge and detector plants. Their comments state specific claims
and limits. They do not establish real isolated-account/container/native-provider
behavior, external launcher no-start behavior, Windows compilation/ACL execution,
or full-suite/platform qualification. Before activation, older maintenance
consumers must be excluded from affected homes: compatible marker syntax does
not make an older binary enforce this sidecar contract. No release or host
activation follows from those tests.
