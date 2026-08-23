# TASK-260822-3fkfmf — manager profile and SECURITY.md prose for `script-worker-v1`

**Branch:** `spec/sw-manager-security` (curator-spec), commits `b9ca2ad` then
`c2371d3`, both signed and pushed to `origin`, no PR — TASK-260822-c0rxj7 merges
the sibling branches.
**Worktree:** `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/manager-security-worktree`
**Base:** `origin/main` = `b92b105` ("Add first-party module roots for local go-v1 builds (decision 0009) (#24)").
**Input consumed:** `TASK-260822-1l4r4f_analysis.md` revision 2 (all five decision-0008
open questions) and its reviewer verdict `TASK-260822-1l4r4f_review-verdict.md`.

Diff against `origin/main`: 2 files changed, 445 insertions, 12 deletions. No
other file touched.

**Sequencing note.** The core.md subsection did not exist when `b9ca2ad` landed —
`TASK-260822-1f533i` had not created its worktree yet — so that commit referenced
the policy by name only and used provisional spellings for the identifiers core.md
had not yet fixed. `TASK-260822-1f533i` then committed `41cf556` on
`spec/script-worker-v1-normative`, defining **`protocol/core.md` section 4.1.1
"Portable `script-worker-v1` execution policy"**. Commit `c2371d3` reconciles both
files against it. Core.md is normative for every identifier it defines and won on
all of them; the identifier sets were re-checked mechanically afterwards and the
only names in these two files that core.md does not contain are the three
worker-session diagnostics discussed under "Findings handed forward".

## What landed

### `profiles/manager.md`

1. **Section 3 carve-out.** A new paragraph immediately after the launcher-behavior
   paragraph scopes the existing launcher rules to declared-only commands and states
   that an enforced command goes through section 3.1 instead: inherited `PATH`
   discarded rather than preserved, and no shell, `.cmd`, or symlink shim between the
   manager and the interpreter. This resolves analysis finding F3 (the shipped
   launcher rules at `manager.md:503-509` structurally contradict the policy).
2. **New section 3.1 "Enforced `script-worker-v1` command launch"**, mirroring the
   structure of 2.2.1:
   - enforced vs declared-only definition; build commands explicitly out of scope;
   - fixed three-node process graph, worker = hidden-mode re-execution of the
     installed manager; shebang and Windows file association declared inert (Q2);
   - eight ordered per-invocation manager obligations;
   - F6 note: one capability set per skill, so every enforced command in a skill
     shares one containment profile;
   - the mandatory portable control set, and the single rejection rule
     (`script_execution_control_unavailable`) plus the additive install/update
     preflight (Q4);
   - the three build controls deliberately **not** carried across — closed standard
     input, bounded combined output, wall-clock deadline — with the replacement rule
     (explicit manager-controlled standard-stream binding);
   - the `script-worker-v1-native-control-inventory-v1` table, three platform columns
     (macOS / Linux / Windows), eight controls, the `host-conditional` availability
     state and its consistency rules, and the closed five-entry unavailable-reason
     vocabulary (Q5);
   - the closed `script-capability-evidence-v1` record: contents, per-invocation
     cadence, never re-probed mid-session, never on the command's stdout/stderr,
     never in cache key / receipt / marker / claim, bounded machine-local retention
     (Q4);
   - the eight deferred script guarantees and the rule that applying a
     `host-conditional` control never upgrades a policy-level guarantee;
   - the explicit per-platform honesty statement (no kernel confinement primitive on
     macOS or Windows in this release);
   - the single failure boundary;
   - the package-influence boundary, including why a one-value package field is not a
     package-visible policy choice (Q1);
   - the migration note (F7);
   - six stable `phase: execution` diagnostics.
3. **Section 7 audit policy.** Two REQUIRED warning classes, always `warn`, never
   blocking, never subject to `fail_on`: `script_command_declared_only` and
   `script_command_network_unfiltered` (decision 0008 §5 and Q3).

### `SECURITY.md`

1. **Security model qualifier.** The blanket sentence "Capability declarations and
   source auditing are review and policy surfaces; they are not runtime sandboxes"
   was true for every command before this change and is not true for an enforced
   script command. It is now split: source auditing stays a review surface;
   capability declarations stay a review surface except under
   `execution_policy: "script-worker-v1"`, where the same declaration is
   additionally the input to a **manager-mechanism** containment profile. The
   sentence "Neither form is a kernel sandbox" carries the honesty.
2. **New peer section "Enforced script execution boundary"**, placed after the
   external build repository boundary and before registry security state, with three
   subsections:
   - **Script trusted computing base** — including the explicit statement that the
     interpreter's standard library, `site-packages`, `node_modules`, and every other
     installed package tree are inside the trusted base and are *not* verified, and
     that this applies to a manager distributed as an interpreted package (Q2; the
     cocoaskills Python manager is exactly that case).
   - **Enforced script mechanisms and deferred guarantees** — the eight-row
     mechanism-versus-deferred-guarantee table (the deliverable's core), the
     reporting-only treatment of declared network host globs with the rationale that
     no supported host offers host-granular filtering as a kernel control (Q3), and
     the platform-reach paragraph.
   - **Script evidence and failure boundary** — separate record and inventory
     versions from the build side, probe cadence, result-only exposure and the
     stdout/stderr prohibition, bounded retention, the one failure rule, and the
     migration statement.

## Identifier reconciliation with core.md 4.1.1

Adopted from section 4.1.1 in `c2371d3`, replacing the provisional spellings of
`b9ca2ad`:

| Provisional in `b9ca2ad` | Normative in core.md 4.1.1 |
|---|---|
| `script-unconditional-network-denial` | `script-total-network-denial` |
| `script-unconditional-write-confinement` | `script-private-runtime-area-only-writes` |
| `script-verified-interpreter-tree` | `script-read-only-runtime-tree` |
| `script-unconditional-exec-denial` (separate row) | folded into `script-exact-executable-allowlisting` — seven guarantees, not eight |
| `script_execution_deferred_claim_forbidden` | `script_execution_hardened_claim_forbidden` |
| `script_command_declared_only` | `script-command-declared-only` |
| `script_command_network_unfiltered` | `script-command-unfiltered-declared-network` |
| `invocation-private` runtime area | `operation-private` runtime area |
| three-node process graph | four-node graph, three under `exec: "none"` |
| policy referenced by name only | `protocol/core.md` section 4.1.1, cited by number in both files |

Agreed on first contact, no change needed: `script-worker-v1-native-control-inventory-v1`;
`script-capability-evidence-v1`; the `available` / `host-conditional` / `unavailable`
states; all five unavailable reasons; all eight inventory controls and every cell of
the three-platform table; `script_execution_control_unavailable`;
`script_execution_capability_evidence_invalid`; `script-network-host-allowlisting`;
`script-exact-executable-allowlisting`;
`script-hard-aggregate-descendant-resource-bounds`;
`script-fail-closed-capability-preflight`.

Added in `c2371d3` beyond `b9ca2ad`:

- `script_execution_policy_unsupported`, plus the rule that a manager without this
  policy rejects the command rather than installing it declared-only.
- **The reserved environment-name set.** Core.md 4.1.1's `env_read` bullet ends
  "The exact reserved set per platform and per interpreter identifier is defined by
  the manager profile" — an explicit delegation that would otherwise have been an
  unsatisfied forward reference. Section 3.1 now defines it: a common set (`PATH`,
  `HOME`, the temp and XDG roots, the proxy and resolver names, `LD_PRELOAD`,
  `LD_LIBRARY_PATH`, `LD_AUDIT`, `IFS`, `CSK_PROJECT_ROOT`), a macOS set (every
  `DYLD_`-prefixed name), a Windows set (`USERPROFILE`, `APPDATA`, `LOCALAPPDATA`,
  `PATHEXT`, `COMSPEC`, `WINDIR`), a `python3-v1` set (every `PYTHON`-prefixed
  name), and a `node-v1` set (every `NODE_`- and `NPM_CONFIG_`-prefixed name), with
  the rule that a manager rejects an interpreter identifier it has no reserved set
  for rather than launch with an unfiltered environment, and reports any `env_read`
  entry it withholds.

Core.md's other delegation to this profile — "the script command launcher rules of
the manager profile govern declared-only commands" (`core.md:288`) — is satisfied by
the section 3 carve-out.

## The no-kernel-guarantee acceptance criterion

Every occurrence of "kernel" and "guarantee" in the added prose was audited. All of
them are negations, deferrals, or prohibitions:

- `SECURITY.md`: "Neither form is a kernel sandbox"; "each one stops short of a
  kernel-enforced guarantee that this release does not provide and MUST NOT claim";
  "no supported host offers host-granular filtering as a kernel control"; "Kernel
  confinement is not part of it"; "this release applies no kernel confinement
  primitive to an enforced invocation at all"; "it never upgrades the policy-level
  guarantee, which stays deferred"; "A reader MUST NOT infer from `exec: \"none\"` or
  `network: \"none\"` a denial that the host does not provide".
- `profiles/manager.md`: "Every rule this policy enforces is a manager mechanism";
  "the first release applies no kernel confinement primitive at all"; "a descendant
  that resolves an absolute path or opens a path outside the private roots is not
  stopped by this policy"; "MUST NOT describe any of them as a guarantee of the
  policy"; `script_execution_deferred_claim_forbidden`.

No sentence in the added prose asserts that the portable policy denies, prevents, or
confines anything at the kernel level on any platform.

## Gates (each run standalone, real exit codes)

Base `origin/main` `b92b105` was measured green first; every gate below was
re-run after the reconciliation and reflects the final tree at `c2371d3`.

| Gate | Command | Exit |
|---|---|---|
| Schemas, cases, manifest, local links | `python tools/validate.py` | 0 (`validated 49 schemas and 471 vector files`) |
| Stable release gate tests | `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 (`Ran 91 tests ... OK`) |
| Generator and cryptographic vectors | `go test ./tools/...` | 0 |
| Vector regeneration | `go run ./tools/generate-vectors -root .` | 0 |
| Deterministic regeneration | `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | 0 |
| Go formatting | `gofmt -l tools` | 0 (no output) |
| Whitespace | `git diff --check` | 0 |

`python3` in this environment has no `jsonschema`; a throwaway venv was created at
`<worktree>/.temp/venv` from `requirements-dev.txt` and used for both Python gates.
`<worktree>/.temp/` is untracked and was not committed; the repo has no `.gitignore`,
so the landing task should not `git add -A` from this worktree.

The `links` CI lane was also run locally over the two changed files:
`lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429
SECURITY.md profiles/manager.md` exited 0, 1 unique link, 1 OK, 0 errors.

## Findings handed forward (not fixed here)

1. **`active-process-count-limit` on Linux is marked `available: RLIMIT_NPROC`, and
   that verdict looks wrong.** `RLIMIT_NPROC` caps processes per real UID across the
   whole session; it is not a private aggregate domain for the worker. The existing
   rc5 ledger marks macOS `active-process-count-limit` **unavailable:
   `no-private-aggregate-domain`** for that very primitive. Marking the same
   primitive `available` on Linux is internally inconsistent with the ledger it is
   copied from, and an `available` entry MUST be reported `applied` — so this cell
   would produce exactly the false applied-claim that decision 0006 forbids. It was
   mirrored verbatim from the accepted analysis rather than silently diverged from,
   because the same table also lands in core.md (TASK-260822-1f533i) and the vectors
   (TASK-260822-f4qv7w) and a unilateral change in one of three mirrored copies reads
   as a copy error. **Recommendation: change the Linux cell to `host-conditional:
   delegated cgroup v2 `pids.max`` and drop `RLIMIT_NPROC`,** applied consistently in
   all three places. Reviewer's call.
2. **Three diagnostics in `profiles/manager.md` are not in core.md's set, and
   core.md calls its set complete.** Section 4.1.1 says "The complete diagnostic set
   of this policy is:" and lists four. Section 3.1 carries those four plus
   `script_execution_worker_identity_invalid`,
   `script_execution_worker_protocol_invalid`, and
   `script_execution_package_influence_forbidden`. That extension is the established
   shape — core.md 4.2.1 names three `build_execution_*` codes and manager.md 2.2.1
   carries six, including exactly these three analogues — and dropping them would
   leave the identity-verified worker's own failure modes undiagnosable, which is
   the point of verifying identity. But core.md 4.2.1 never says "complete", and
   4.1.1 does, so the two documents are in tension as written. Section 3.1 labels
   the split explicitly rather than papering over it. **Recommendation: soften
   4.1.1's "complete diagnostic set" to "policy-level diagnostic set", or promote
   the three.** Reviewer's call; either resolution leaves this profile unchanged.
3. **Reviewer finding 3 on TASK-260822-1l4r4f is still open.** Q5 routes
   host-conditional controls to platform-case class `host-capability`, but
   `platform-cases.tsv` defines that class as "the runner filesystem cannot create
   the artefact". A missing Landlock/cgroup/netns kernel feature is not that. This
   prose does not touch the ledger, so nothing here is wrong, but
   TASK-260822-f4qv7w needs a widened class or a new one, decided in prose rather
   than at gate time.
4. **Windows Job Object nesting is still unconfirmed** (analysis §8 item 7): whether
   `active-process-count-limit` can be applied inside an existing job the manager did
   not create. The inventory says `available: Job Object active-process limit` on
   Windows, carried over verbatim from rc5 where the same claim already ships, so
   this prose introduces no new exposure. It stays an open vendor-documentation item.
5. **Two subjective calls worth a reviewer's eye.** The `phase` for the seven
   diagnostics is `execution`, reusing the build execution phase. And the audit
   warning classes were placed in `profiles/manager.md` section 7 rather than in the
   audit-record schema, because section 7 is where audit decisions are normative
   today and because core.md 4.1.1 says the classes are "emitted on the surfaces the
   manager profile defines"; the schema-side representation belongs to
   TASK-260822-1mwy10 if it wants one.

6. **The reserved environment-name set is new normative content invented here.**
   Core.md delegated it without constraining it, so section 3.1's five lists are this
   profile's judgement, not a transcription. They are deliberately prefix-based where
   a vendor keeps adding names (`DYLD_`, `PYTHON`, `NODE_`, `NPM_CONFIG_`) so the set
   stays closed as interpreters evolve. TASK-260822-f4qv7w will need vectors for at
   least the withheld-`env_read` case and the unknown-interpreter rejection.
