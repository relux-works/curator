# Project registration and item admission

CIP-0002 admission is an operator-controlled, local manager operation. It
creates no launch fragment, composition or lease, and advertises no native
source-exclusion qualification. Existing `project add`, `resolve` and `refresh`
retain their existing Skillfile meanings.

## Explicit instruction packaging

Start with an installed profile and a committed Git checkout. The manager home
must be outside the checkout. Use the checkout ID and digests printed by these
commands:

```sh
curator project register ./demo --name demo --profile dev --environment codex_cli
curator project inspect CHECKOUT_ID
curator project package CHECKOUT_ID instructions --source AGENTS.md
curator project preview CHECKOUT_ID
curator project approve CHECKOUT_ID instructions --digest ITEM_SHA256 --package-digest REVIEW_SHA256
curator project revoke CHECKOUT_ID instructions
```

`register` defaults the name to the directory basename. `package` is an explicit
selection of one inventoried `AGENTS.md` or `CLAUDE.md`; discovery never selects
one automatically. It preserves the selected LF Markdown bytes, writes a
schema-2 `agent-context.json` with exactly that module, and retains the complete
snapshot and schema-4 context lock outside the repository. `--optional` records
explicit optionality. Preview shows the generated manifest, selected module
bytes, complete inventory, pinned closure, file changes and package digest.
The commands return JSON; approval and revocation affect protected state only.

A changed source file requires packaging and review again. It never rewrites
previously approved bytes. Approval of an older preview refuses if its live
selected source changed. A later candidate digest cannot reuse an old item
approval. A digest identifies content; it does not grant authority.

The admitted context's rendering intent is `append-after-profile`. The
composition consumer must keep the selected profile's root/system context and
append admitted project context in its recorded order. This package does not
implement composition or native delivery.

## Registration identity

Checkout IDs are random manager allocations. Registration records bind a
canonical Git root, its canonical network origin identity (or the defined
null-origin case), its single root commit, filesystem device/inode, selected
profile lock and environment. Profile bytes, names, `.git` and repository files
cannot select an allocation. `LocateFor` takes explicit manager-selected
profile/environment parameters; the deepest registered root shadows a parent,
including when its requested environment slot is absent. Unqualified ambiguous
selection refuses.

Moved, replaced or unreadable checkouts require an explicit repair:

```sh
curator project rebind CHECKOUT_ID ./demo
```

Rebind retires the old record and allocates a fresh ID. It carries no approval.
Revocation remains possible when a source checkout is absent. Filesystem
binding is supported on macOS and Linux; other platforms refuse registration.
Home/worktree references reserve protected allocations for the later composition
and lease lanes; they do not claim a populated sparse worktree or usable home.

## Explicit package requests

`curator project preview CHECKOUT_ID --declaration request.json` accepts a
closed manager-operation request. Relative paths are under that registration:

```json
{
  "version": 1,
  "items": [
    {
      "id": "guidance",
      "type": "context",
      "environment": "codex_cli",
      "required": true,
      "source": "project-context",
      "dependencies": {}
    }
  ]
}
```

All six item members are required; nulls, duplicate keys, unknown members and
unknown versions refuse. `source` names a package directory. Context packages
must declare explicit modules in `agent-context.json`; bare-file selection uses
`project package` instead. Context requirements resolve offline with the existing
context resolver, against local Git object sources explicitly mapped by package
name in `dependencies`. Each source is a relative checkout-root path whose
canonical network origin matches its declared source. Exact revisions, tags and
semantic ranges use only existing local objects; missing pins never fetch.
Every selected closure member is copied to protected immutable storage before
approval, with inherited context parsing, pins, weights, ordering and module
byte rules. Context approval cannot import skill/MCP/native-settings effects.
Native import directives are unsupported even when their targets are inventoried.
The conservative first-slice import check rejects whitespace-prefixed `@path`
tokens in selected modules; ordinary prose links remain inert data.

Whole skill requests use `type: "skill"`. They require an explicit portable
skill manifest, complete bounded snapshot, content pin and the existing strict
read-only audit/canary. This entry point refuses skill requirements, MCP
requirements and external build repositories until their owning complete pinned
activation closure is available. It does not replace command activation, build,
capability or permission checks in a later consumer. An arbitrary directory is
never sufficient executable admission.

Inline MCP is inventoried and displayed by identity only. Stage-A admission
refuses required MCP items as unsupported. Explicitly optional requests are
returned as named unsupported omissions with their source digest; they must
still be bounded, readable JSON. There is no server approval or connection until
CIP-0022 stage B. Rules, knowledge and native configuration are inert discovery
candidates; they have no extra item arms in leaf 1.

## Authority and bounds

State lives in the private `project-context/` namespace below the configured
manager home. Its ledger, snapshots and locks use no-follow reads, bounded
inventories, owner/write-boundary checks, complete byte/entry/mode verification,
canonical records, the manager-home lock and atomic publication. Missing or
malformed authority never becomes an empty successful state. The ledger binds
its manager allocation; copied records or identical hashes from another
checkout confer no approval. Snapshot references encode the complete digest.

Approvals record the manager's protected operator principal, checkout, item
digest, environment and ceiling digest. This operator entry point accepts no
caller-selected approver, permission payload or agent impersonation flag. Agent
approval through a manager-issued grant is unsupported here and requires its
own grant-aware entry point. Permission files never apply, and native memory
remains off. Optional omission is explicit and named; a required item cannot be
silently disabled or removed while its approval remains active.

Limits include 1 MiB/depth 16 per JSON record, 256 items, snapshot depth 32,
240-byte portable paths, at most 10,000 inventory nodes/files, 16 MiB per file
and 64 MiB per snapshot. Protected ledger and Git-output limits can reject a
large request before those outer caps. Paths, links, case/normalization aliases,
cycles, unsupported imports and unbounded patterns refuse. Discovery only walks
known control roots and never sources hooks, installs, fetches or connects.

The first slice assumes trusted code. An unrestricted process with the same
UID can modify operator-owned manager state; filesystem permissions and hashes
do not establish hostile-code containment. Native discovery suppression, grant
enforcement, source-exclusion qualification and launches belong to later lanes.

## Verification

Use Go and Git with disposable fixture repositories:

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/projectcontext -count=1
env -u GIT_DIR -u GIT_WORK_TREE go test ./cmd/curator -run 'TestProjectAdmissionCLIFlow|TestRegisteredSubcommandHelpNoWrites' -count=1
CURATOR_PROJECT_ADMISSION_MUTANTS=1 env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/projectcontext -run '^TestAdmissionNarrowingMutants$' -count=1 -v
```

The mutation lane uses Go overlays in disposable test directories. It requires
intended regression assertion failures for root-only closure verification and
cross-checkout hash approval; compilation/setup failures are not successful
controls. Store task logs and temporary validation state under `.temp/`.
