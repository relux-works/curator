# TASK-260822-1l4r4f — decision 0008 open questions: resolutions for the normative prose

**Task:** TASK-260822-1l4r4f (parent STORY-260822-3k3hbs, epic EPIC-260822-3ar0tv)
**Date:** 2026-08-22
**Subject:** the five questions deferred by `decisions/0008-enforced-script-capabilities.md` §"Open questions deferred to the normative change" (lines 127–139)
**Consumers:** TASK-260822-1f533i (core.md prose), TASK-260822-1mwy10 (manifest schema 8), TASK-260822-3fkfmf (manager profile + SECURITY.md prose), TASK-260822-f4qv7w (conformance vectors)

## 0. Sources read

All citations are `path:line` in `/Users/iv/Developer/ReluxWorks/curator-spec` at `main` = `a2d44eb`
("Add enforced script capabilities (decision 0008) (#23)"), release `1.0.0-rc.8`.

| Source | What it settles |
|---|---|
| `decisions/0008-enforced-script-capabilities.md` | the decision under interpretation; the five questions |
| `decisions/0006-portable-manager-worker-execution.md` | the doctrine this must extend: mandatory portable / native inventory / deferred hardened |
| `decisions/0007-portable-and-verified-assurance.md` | the `portable` vs `verified` assurance overlay that landed after 0006 |
| `protocol/core.md:320-473` (§4.2.1) | the portable execution policy in normative form |
| `protocol/core.md:193-204` (§4.1) | script command shape, shim ownership |
| `protocol/core.md:551-563` (§4.3) | the capability declaration that becomes the policy input |
| `protocol/core.md:930-968` (§8.2) | Go toolchain identity — the fingerprint model Q2 asks about |
| `protocol/core.md:1288-1315` (§12.3) | closed-identifier admission rule for drivers and execution policies |
| `profiles/manager.md:202-293` (§2.2.1) | the manager-side worker session obligations |
| `profiles/manager.md:495-542` (§3) | how script command launchers work today |
| `profiles/manager.md:614-635` (§7) | source audit decisions and the pre-capability-schema pin precedent |
| `SECURITY.md` | the enforcement/guarantee split prose to mirror |
| `schemas/v1/common.schema.json` (`$defs.scriptCommand`, `$defs.capabilities`, `$defs.commandV7`) | the exact shapes schema 8 has to extend |
| `conformance/v1/vectors/go-host-execution-policy.json` | the machine-readable authority for the inventory, evidence record, failure boundary |
| `schemas/v1/provider-capability-receipt-v1.schema.json` | the six verified provider capabilities |

---

## 1. Cross-cutting findings the prose must absorb first

These are not answers to the five questions; they are constraints that change what a correct
answer looks like. Each is a verified defect or gap in the current text, not an opinion.

### F1 (blocking for all five). Capability *values* have no normative meaning today

`protocol/core.md:553-563` defines capabilities syntactically and nothing more. `"repo"` and
`"home-config"` appear exactly twice in the whole repository — once in that syntax list and once
in `schemas/v1/common.schema.json:535` — and are never defined. `secrets` identifiers, `env_read`
names, and `network` host globs have no defined effect either. That was acceptable while
§4.3 said "audit surface, not a runtime sandbox" (`core.md:553-554`).

Decision 0008 promotes the same object to a containment profile. A containment profile derived
from undefined terms is not implementable and not testable. **The core.md subsection must define,
for `script-worker-v1` only, the exact meaning of every capability value it consumes** — what
`repo` and `home-config` resolve to, what a `secrets` identifier does to the environment, what an
`env_read` name passes through, what a host glob matches. Leaving §4.3 semantics open and writing
enforcement prose on top of it produces vectors that cannot be written.

Concrete instance of the damage: 0008 §3 last bullet says absent fields mean "no writes outside
the private runtime area", but the schema default for a missing `filesystem` is `"repo"`
(`common.schema.json`, `$defs.capabilities.filesystem.default`), which is not the private runtime
area under any reading. **The decision text and the schema default already disagree.**

### F2 (blocking for Q4 and Q5). `capability-evidence-v1` cannot carry script evidence

0008 §4 says inventory gaps are "recorded as unavailable in one closed `capability-evidence-v1`
record". That record is closed against exactly this: `core.md:419-425` requires exactly one entry
per **build** inventory control, and an `execution_policy` other than `manager-worker-v1` is
`build_execution_hardened_claim_forbidden` (`core.md:437`, mirrored as vector case
`hardened-execution-policy-in-evidence-record` in `go-host-execution-policy.json`). A script
record naming `script-worker-v1` is, by the current normative text, a forbidden hardened claim.

**Resolution: a new closed record version `script-capability-evidence-v1` over a new inventory
`script-worker-v1-native-control-inventory-v1`, with its own diagnostics prefix
(`script_execution_*`).** Reuse 0006's *shape and rules*, never its version constants. This is
exactly what 0006 already requires of any new inventory: "Adding, removing, or re-scoping an entry
requires a new inventory version" (`core.md:414-417`), and non-aliasing identities are the
doctrine's core move (0007's identity table).

### F3 (blocking for Q2 and Q5). The existing script launcher contradicts the policy

`profiles/manager.md:503-509`: a launcher "prepends the project or global command directory …
**preserves the inherited `PATH`**, forwards arguments without reinterpretation, and returns the
child command's exit status", and on Unix it "is either a relative symlink … or an executable
POSIX-shell wrapper". Verified against the Go implementation: `internal/runtimestore/runtimestore.go:119-145`
writes a bare symlink on Unix and `@echo off / "<path>" %*` on Windows.

Under `script-worker-v1` none of that survives:

- inherited `PATH` is the opposite of the controlled `PATH` 0008 §3 requires for `exec`;
- a symlink shim has no process in which to install controls;
- a POSIX-shell wrapper puts a shell in the graph, which the build boundary forbids outright
  (`SECURITY.md`: "MUST NOT use `sh -c`, `cmd.exe /c`, PowerShell, or any other shell");
- the Windows `.cmd` shim routes through `cmd.exe` for the same reason.

**The enforced shim must be a manager-owned native launcher that re-executes the installed manager
in the hidden script-worker mode.** §3 needs an explicit carve-out: the launcher rules of
`manager.md:503-509` govern declared-only commands; enforced commands use the worker launch.

### F4 (naming, affects Q3 and Q5). "Hardened" is stale; 0007 renamed the far end

0008 defers to "a hardened profile … named separately, exactly as decision 0006 does for builds".
Since rc.7, the far end of that axis is decision 0007's **`verified`** mode:
`verified-provider-policy-v1` / `verified-provider-execution-v1` over the
`host-execution-provider-v1` contract, with six `established` capability ids
(`provider-capability-receipt-v1.schema.json`). The six *deferred* build guarantees of 0006 and
the six *provider* capabilities of 0007 are the same six concepts under two naming schemes.

**The script prose should defer to a future verified/provider script execution policy, name its own
deferred-guarantee set with script-specific names, and keep those names disjoint from the frozen
build six** so a script evidence record can never trip `build_execution_hardened_claim_forbidden`
or be mistaken for build evidence.

### F5 (scope, affects Q3 and Q5). There is no Linux cell in any execution-policy ledger

`protocol/core.md` and `profiles/manager.md` never mention Linux. `rc5-native-control-inventory-v1`
has exactly two platforms, `macos` and `windows` (`go-host-execution-policy.json`,
`native_control_inventory.platforms`), and capability-evidence `platform` is `macos` or `windows`
(`core.md:421-422`). Decision 0008's Context offers "unshare/netns on Linux" as an example of a
network control. **That example implies a platform cell that does not exist.** The prose must not
inherit it; the first script release ships macOS and Windows cells only, matching 0006. (Claim
schemas 3 and 4 do admit `linux` in `operating_systems`, which is a different axis — where the
implementation was tested — not an execution-policy inventory cell.)

### F6 (consequence, affects Q1). Capabilities are skill-wide; enforcement is per command

`capabilities` is a top-level property of the manifest (`agent-skill-v7.schema.json`, top-level
properties are `schema_version, runtime_roots, build_roots, build_repositories, commands,
capabilities, dependencies`), while commands are individually typed objects. So one skill has one
capability set, and every enforced command in it gets the same containment profile. An `exec` name
that only one command needs widens the profile of every enforced command in that skill. This is
coarse but it is the existing shape; **state it explicitly in the prose and do not invent
per-command capabilities in this change** — that would touch §4.3 for every schema version and is
its own decision.

### F7 (migration, affects Q4 and Q5). Enforcement makes `env_read` and `exec` breaking

Today a script inherits the caller's whole environment and `PATH`. Under enforcement, an undeclared
variable is absent and an undeclared executable is not on `PATH`. Skills whose declarations were
written as documentation will break the first time they opt in. That is the intended effect — it is
the entire point of 0008 — but the prose owes an explicit migration note, and the audit layer
should surface the diff at opt-in time rather than at first invocation.

---

## 2. Q1 — `execution_policy` per command vs manifest-level default with per-command override

### Recommendation

**Per command, on the script command object only. One OPTIONAL field, one closed value, no
manifest-level default, no inheritance, no override resolution.**

Schema 8 adds to `$defs.scriptCommand` (or a `scriptCommandV8` sibling, so schema 7 bytes stay
frozen):

```jsonc
"execution_policy": { "const": "script-worker-v1" }   // OPTIONAL; absent = declared-only
```

Absent means exactly what it means in schema 7: declared-only, launcher exec, no enforcement claim,
plus the new declared-only audit warning class of 0008 §5.

### Rationale

1. **It preserves the "policy identity is not a package option" doctrine, correctly read.**
   `core.md:325-327` says the policy identity "is never a package-visible option, a host label, or
   an operator preference". A one-value enum does not violate that rule, because the package is not
   *choosing* a policy — it is choosing whether the command is enforced at all. The manager still
   owns every control, limit, path, and environment value. The influence is monotonically
   restrictive: opting in can only narrow what the command may do, never widen it. That asymmetry is
   what makes it safe, and the prose should say so in one sentence, because a reviewer comparing
   §4.2.1 with the script subsection will otherwise read a contradiction.
2. **The audit statement is per command, so the declaration should be too.** 0008 §5 requires a
   reviewer to distinguish enforced from declared-only commands, and `core.md:203-204` establishes
   the command as the ownership unit: "The command name is the shim name and one active name has
   exactly one owner." A per-command field makes the audit claim local and directly readable rather
   than derived from a default plus an override.
3. **A default plus override has a worse cost/benefit than it looks.** The value space is a single
   constant. A manifest-level default therefore saves at most one line per command, while adding a
   precedence rule, four (default × override) combinations, and a set of conformance vectors for
   each — for a field that can only ever say one thing.
4. **Per-command opt-in fails safe over the manifest's lifetime.** A manifest-level default silently
   enforces every command added later by a contributor who never considered enforcement. Absent-means-
   declared-only fails toward the pre-existing, honestly labelled behavior. This is the same instinct
   as 0008's rejected "enforce on legacy schemas implicitly".
5. **Mixed mode is a required migration path, not an edge case.** A skill must be able to enforce a
   new command while an old one stays declared-only under its warning class, in one release. Per-command
   is the only shape that expresses that without a negative override value.

### Rejected options

- **Manifest-level default with per-command override.** Rejected: non-local read (you cannot tell
  what a command does from the command object), a precedence rule and its vectors, and it needs a
  negative override token (`"none"`? `null`? `false`?) to express "this one command stays declared-only",
  which adds a second way to spell the schema-7 meaning.
- **Manifest-level only, no per-command field.** Rejected: all-or-nothing per skill blocks incremental
  migration, and a skill with one network-dependent legacy command could never enforce any of its
  other commands.
- **Implicit enforcement for every schema-8 script command, no field.** Rejected: it makes the schema
  version the enforcement switch. Schema 8 will carry other features; an author bumping the version
  for one of them would silently acquire enforcement, and the audit record would have nothing local
  to point at.
- **`execution_policy` on build commands too, for symmetry.** Rejected: for builds the policy is
  implied by the closed driver and is deliberately not package-visible (`core.md:325-327`). Adding
  the field there would convert a manager-owned identity into a package-visible one and break the
  cache/receipt/marker/claim identity argument of 0006.

### Prose hooks

- new `$defs` in `common.schema.json` (do not mutate `scriptCommand`; schema 7 bytes are frozen per
  `COMPATIBILITY.md`), reachable only from `commandV8`;
- one sentence in the script subsection on why a one-value package field is not a package-visible
  policy choice;
- one sentence recording F6: one capability set, N enforced commands, same profile.

---

## 3. Q2 — Interpreter identity: §8.2-style fingerprint vs declared `exec` entry

### Recommendation

**Neither of the two options as posed. Use the `driver` precedent: the package names a closed
interpreter *identifier*; the manager resolves it to an operator-trusted executable under §8.2's
*resolution and trust* rules; the manager verifies that executable *file's* identity per invocation;
the interpreter's library and installed-package trees are declared trusted computing base, not
verified identity. The shebang and Windows file association become inert.**

Concretely, on an enforced script command schema 8 requires a closed identifier, e.g.

```jsonc
"interpreter": { "enum": ["python3-v1", "node-v1", "bash-v1", "powershell-v1"] }   // exact set = prose task's call
```

and the manager, per invocation:

- resolves the interpreter **independently of the package and before entering any package-controlled
  directory**, and rejects resolution from the repository, a runtime root, project `.agents/bin`, the
  user `PATH`, or a manifest — the exact rule `core.md:932-937` already states for `bin/go`;
- requires a canonical regular executable file, rejects symlink / reparse-point / hard-link
  substitution, records strong file identity, and hashes the executable's bytes — the rule
  `manager.md:214-216` already states for the worker executable;
- re-checks that identity at the launch boundary so a replacement race cannot widen the graph
  (`manager.md:217-219`);
- treats the interpreter's stdlib, site-packages, and any installed package tree as trusted computing
  base and **says so**, per the sentence already in `core.md:343-347`: an implementation whose graph
  contains "an interpreter and an installed package tree" MUST "treat every mutable component of that
  graph … as trusted computing base".

### Rationale

1. **The §8.2 whole-tree fingerprint does not transfer.** `core.md:942-968` hashes every file, link,
   and directory under `GOROOT` and requires the tree to "remain unchanged through the last child
   exit". For an interpreter, the equivalent tree is the whole installation including `site-packages`
   / `node_modules` — user-writable, mutated by any unrelated `pip install`, and legitimately different
   between two invocations of the same skill. Hashing it per invocation is both prohibitively expensive
   for a command expected to start promptly and semantically wrong: the identity would change for
   reasons that have nothing to do with the skill. A build is a once-per-install operation with a
   cache key to protect; a script invocation is neither.
2. **A declared `exec` entry is a direct package-influence violation.** `exec` is "unique bare
   executable names" (`core.md:558`) supplied by the package. Letting it name the program that
   *launches* the command is precisely what `core.md:461-468` forbids: package bytes "MUST NOT select
   or modify … the Go or tool executable paths". `exec` also has the wrong job — it bounds what the
   script may *spawn*, and 0008 §3 already says the interpreter is *added* to that bound by the
   manager ("the interpreter plus the declared executable names, resolved by the manager to fixed
   paths"), which presupposes the manager already knows the interpreter.
3. **The shebang is package-controlled content, so it cannot be the selector.** Under enforcement,
   letting `#!/usr/bin/env python3` pick the program hands program selection to the package and, via
   `env`, to `PATH`. It also does not exist on Windows, so any shebang- or extension-derived scheme is
   non-portable by construction — decisive, given that 0006 requires parity on macOS and Windows.
4. **The closed identifier is a precedent this project already reviewed and shipped.** `driver:
   "go-v1"` is exactly this shape: the package names a closed identifier, the manager owns the
   executable. `core.md:1290-1292` and `core.md:1306-1311` makes the admission rule explicit — each new closed identifier is a
   specification revision with its own review, not a manager configuration option. Interpreters inherit
   that rule for free, which is what keeps "add Ruby" from being a config change.
5. **Per-invocation file identity is affordable and honest.** Digesting a 5–30 MB interpreter binary
   costs single-digit milliseconds; walking a Python installation does not. Verifying the launched file
   and openly declaring the library tree as TCB says exactly what is true.
6. **Do not run the interpreter to fetch a version line.** §8.2 appends normalized `go version` output
   (`core.md:955-961`). For scripts that would add a process to the graph on every invocation, before
   controls are meaningful, for a string. Reject it.

### Rejected options

- **§8.2-style whole-tree fingerprint of the interpreter installation** — cost and false-positive
  identity churn, as above; it would make `pip install` in an unrelated venv invalidate a skill.
- **Interpreter as a declared `exec` capability entry** — package selects the launching program;
  violates `core.md:461-468`; conflates "may spawn" with "is launched by".
- **Shebang / file-extension derived resolution** — package-controlled selector, `env`/`PATH`
  indirection, and no Windows equivalent.
- **Interpreter path or digest as a package-declared field** — worse than the `exec` variant: it puts
  a host path in the manifest, is not portable across machines, and hands the package the exact value
  the manager must own.
- **Interpreter digest as part of a cache or identity input** — script commands have no build cache;
  and by the same argument 0006 used to exclude capability evidence from the cache key
  (`core.md:439-442`), a per-host digest is reporting state. The *closed identifier* may join the audit
  record; the host path and digest may not.

### Prose hooks

- a `script-worker-v1` interpreter-resolution paragraph in core.md that cites §8.2 for resolution and
  trust and explicitly declines its tree hash;
- an explicit "the shebang and file association are inert; they MUST NOT select the executed program"
  sentence;
- a SECURITY.md TCB sentence naming the interpreter library tree and installed package tree, mirroring
  `SECURITY.md` ("An implementation that launches the worker through a mutable interpreter or installed
  package tree adds those to the same trusted base and MUST say so") — note this already applies to the
  cocoaskills manager itself, which is a Python package (`cocoaskills-production/pyproject.toml`, `src/csk`).

---

## 4. Q3 — Do `network` host globs configure portable filtering, or stay reporting-only?

### Recommendation

**Reporting-only in the first release. A non-empty host-glob list is an audit and reporting
declaration; the manager MUST NOT represent it as an applied control, MUST NOT add a host-filtering
entry to the script inventory, and MUST NOT claim filtering in any record. An enforced command whose
`network` is a host-glob list is admitted, not rejected, and is labelled with a distinct audit
warning class ("enforced command with unfiltered declared network").**

`network: "none"` keeps the mandatory portable treatment 0008 §3 describes — offline environment
configuration plus proxy and resolver scrubbing — stated as a manager mechanism, never as denial.
Per-host filtering is deferred and gets its own named guarantee (e.g. `script-network-host-allowlisting`)
that MUST NOT appear in the mandatory set, the inventory, or an evidence record.

### Rationale

1. **The spec already rules on this, in the vectors.** `go-host-execution-policy.json`,
   `capability_evidence_cases`, contains the case `unknown-native-control-is-rejected` whose control
   name is literally `host-firewall-profile`, `in_inventory: false`, expected error
   `build_execution_capability_evidence_invalid`. A host-filtering control is already the worked
   example of something a conforming implementation must not report. Adding one to the script inventory
   would contradict the suite that ships alongside.
2. **No portable mechanism exists at the required strength.** Everything available unprivileged on both
   macOS and Windows — `HTTP(S)_PROXY`/`NO_PROXY`, resolver configuration, interpreter-specific proxy
   settings — is honored only by cooperating libraries. A script using a raw socket or an IP literal
   ignores all of it. 0006's central rule is that a mechanism which does not hold is not claimed:
   "Claim the hardened guarantees on macOS and Windows anyway. Rejected: the claim would be false."
3. **The alternatives that would actually filter are out of scope and out of doctrine.** A manager-owned
   local filtering proxy adds a network-listening process to a graph the policy defines as fixed, and
   either terminates TLS (unacceptable) or is CONNECT-only (bypassable). Kernel/LSM filtering is not
   available unprivileged on either supported platform — the same finding 0006 recorded.
4. **Admit rather than reject, because 0006 already made that call.** Rejecting host-glob commands from
   enforcement would deny them the `exec`, `filesystem`, and environment controls they *can* have, and
   push them back to declared-only. That is a smaller version of the "hardened Linux or nothing" option
   0006 explicitly rejected as "worse than the problem". Honest partial enforcement wins; a distinct
   warning class carries the honesty.
5. **The glob syntax is too weak to filter on anyway.** `common.schema.json` constrains each entry to
   `^[^\s/\\]+$` — a bare host pattern with no scheme, port, or path. Even a working filter could only
   act on hostnames, which an IP literal sidesteps. Promising filtering on this vocabulary would
   over-promise twice.
6. **The deferral target exists but does not cover this yet.** 0007's provider contract has exactly six
   `established` capabilities, and `total-network-denial-v1` is all-or-nothing
   (`provider-capability-receipt-v1.schema.json`). Per-host allowlisting is not among them. The prose
   should say the deferral is to a *future* verified capability id, not to an existing one.

### Rejected options

- **Configure a portable filter from the globs (proxy/resolver/`NO_PROXY`).** Rejected: cooperating-library
  only, trivially bypassed, and reporting it as an applied control is exactly the false claim 0006 forbids.
- **Reject enforced commands that declare host globs.** Rejected: withholds real controls to avoid an
  unclaimed one; recreates the shape 0006 rejected.
- **Add a `host-allowlist` entry to the script inventory marked `unavailable` on both platforms.** Rejected:
  an inventory entry that is unavailable everywhere at birth is a promise in a machine-readable file. 0006
  keeps unmet guarantees in the *deferred* list, not in the inventory, and forbids their appearance in
  evidence.
- **Silently treat host globs as `network: "none"` under enforcement.** Rejected: it changes declared
  behavior without the author's consent and breaks working skills, which is the "enforce implicitly"
  failure mode 0008 already rejected.

### Prose hooks

- one row in the script mechanism/deferred-guarantee table: mechanism = "declared host globs are recorded
  and reported; no portable filtering is applied", deferred = `script-network-host-allowlisting`;
- an audit warning class distinct from the declared-only one;
- a negative vector: host-filtering control name in a script evidence record ⇒
  `script_execution_capability_evidence_invalid`.

---

## 5. Q4 — Evidence cadence: per invocation vs per install generation

### Recommendation

**Per invocation, for the probe and the record. Exactly one closed `script-capability-evidence-v1`
record per enforced-command invocation, probed once before worker launch, never re-probed mid-session
however long the command runs. Additionally preflight the mandatory portable controls at install/update
time, so a shim that can never run is never published — same diagnostic, different moment.
The record is result-only, exposed on manager result surfaces, and MUST NOT be written to the command's
stdout or stderr, nor into a marker, cache key, receipt, or claim.**

Replace 0008 §4's phrase "per invocation-policy identity" — it is ambiguous and was probably meant to
read "per invocation" — with "exactly one record per enforced-command invocation".

### Rationale

1. **Per install generation is not a probe.** `core.md:412-414`: "Availability MUST be probed once per
   operation before worker launch; **a cached, inherited, or configured result is not a probe**", repeated
   in `manager.md:211-213`. An install-generation record is by construction a cached result replayed at
   invocation time. Choosing it would require weakening the one rule that makes the whole evidence
   mechanism trustworthy.
2. **The host genuinely changes between install and invocation.** A script command runs arbitrarily long
   after install, possibly thousands of times: OS updates land, policy changes, the user session differs,
   Windows Job Object availability can differ under a nested job, resource limits differ per session. A
   build is one operation bounded by its own install; a script command is not. Install-generation evidence
   would be a claim about a host that no longer exists.
3. **Once per invocation, not periodically.** For a long-running command, a mid-run re-probe cannot
   retroactively change controls that were installed before `exec`, and two records for one session would
   be a contradiction the closed-record rules exist to forbid. The record describes the controls installed
   at launch; that is a complete and honest statement for the whole session.
4. **Cost is real but bounded and already accepted.** The probe is local feature detection, not I/O against
   the network or a package tree, and the build side already mandates it per operation. Latency is bounded
   by the same work the worker launch does anyway (identity check, environment construction).
5. **Install-time preflight is additive, not an alternative.** Publishing a shim for an enforced command on
   a host that cannot apply the mandatory set produces a command that fails on every invocation. Rejecting
   at install with `script_execution_control_unavailable` is strictly better UX and matches
   `manager.md:294-337`'s preflight posture. It does **not** license skipping the invocation-time check.
6. **The output channel is the part that differs from builds, and it must be stated.** Build evidence is
   reported "in install, dry-run plan, and status results" (`core.md:439-442`) because the manager owns the
   result. A script invocation's stdout/stderr belong to the user's pipeline —`manager.md:508-509` requires
   the launcher to forward arguments unreinterpreted and return the child's exit status, i.e. behave
   transparently. Writing an evidence record there would corrupt the command's output. Expose it through a
   dry-run/plan projection, a status-style read, and an explicitly operator-selected diagnostic destination
   that package data can never choose.
7. **Retention must be bounded by default.** One record per invocation, persisted forever, is an unbounded
   log with a privacy surface. Default to retaining at most the most recent record per command, machine-local,
   operator-configurable — and keep it out of every hashed identity, exactly as `core.md:439-442` does for
   builds — and for the same reason decision 0006 gives for excluding build evidence from the cache key
   ("letting it into the key would fragment portable cache identity while telling readers nothing they may
   rely on").

### Rejected options

- **Per install generation.** Rejected: it is definitionally a cached result (`core.md:412-414`), and it
  makes a stale claim about a host that may have changed since install.
- **Per invocation *and* periodic re-probe during long sessions.** Rejected: controls are installed before
  `exec` and cannot change afterwards; a second record for one session is a contradiction, not information.
- **Hybrid — install-generation record reused when a cheap "host unchanged" token still matches.** Rejected:
  the token is a cache validity heuristic, and there is no portable, trustworthy definition of "unchanged
  host". It is the cached-result rule with an extra step.
- **Emit the record on the command's stderr.** Rejected: breaks output transparency and would make evidence
  a package-visible side channel.
- **Put the record in the install marker.** Rejected: markers are identity/state, and `core.md:439-442` and
  `manager.md:271-273` already forbid evidence in marker, cache, receipt, and claim. Reuse the rule, do not
  carve an exception.

### Prose hooks

- the cadence sentence, plus the explicit "not re-probed mid-session";
- two preflight moments, one diagnostic (`script_execution_control_unavailable`), stated as
  install/update **and** invocation;
- an explicit exposure/retention paragraph, because the build text's "install, dry-run plan, and status
  results" does not cover the invocation case.

---

## 6. Q5 — Windows scope: mandatory portable set vs native inventory

### Recommendation

**Windows ships at full parity in the mandatory portable set, exactly as 0006 requires for builds. Every
platform asymmetry lives in the script native-control inventory, with each (control, platform) cell
explicitly available-with-named-mechanism or unavailable-with-named-reason from a closed reason
vocabulary. Two platform columns only — `macos` and `windows`.**

#### Mandatory portable set (both platforms; a host that cannot apply all of it rejects with `script_execution_control_unavailable` before the worker starts)

| # | Control | Note |
|---|---|---|
| 1 | identity-verified manager-owned script worker, launched as a hidden-mode re-execution of the installed manager | replaces the symlink/`.cmd`/shell shim for enforced commands (F3) |
| 2 | fixed manager-owned environment: empty bootstrap plus manager-set variables, plus exactly the host variables named by `env_read` | makes `env_read` enforced (F7) |
| 3 | manager-built `PATH` containing exactly the resolved interpreter and the resolved declared `exec` names, inherited `PATH` discarded | the `exec` mechanism of 0008 §3 |
| 4 | interpreter executable resolution and per-invocation identity verification | Q2 |
| 5 | offline network configuration plus proxy and resolver scrubbing when `network: "none"` | Q3; mechanism, never denial |
| 6 | operation-private temporary root and per-command private configuration/cache roots; manager-selected working directory | the `filesystem` mechanism |
| 7 | release of unrelated descriptors and handles before `exec` | see divergence note below |
| 8 | per-invocation availability probe plus exactly one closed `script-capability-evidence-v1` record | Q4, F2 |
| 9 | termination and joining of the complete worker domain before the invocation returns | mirrors `manager.md:237-239` |

**Deliberate divergence from the build policy, and it must be written down: standard input stays open.**
`core.md:369-370` makes "closed standard input" mandatory for builds. A script command legitimately reads
stdin — that is the point of a CLI shim, and `manager.md:508-509` requires transparent forwarding.
Carrying the build rule across would break interactive and piped commands. Keep descriptor/handle hygiene
(item 7); drop stdin closure.

#### Script native-control inventory, first release (`script-worker-v1-native-control-inventory-v1`)

| Control | macOS | Windows |
|---|---|---|
| `descendant-domain-termination` | available: process group and session teardown | available: Job Object kill-on-close |
| `active-process-count-limit` | unavailable: `no-private-aggregate-domain` | available: Job Object active-process limit |
| `aggregate-memory-limit` | unavailable: `no-private-aggregate-domain` | available: Job Object process and job memory limit |
| `per-file-size-limit` | available: `RLIMIT_FSIZE` | unavailable: `no-private-aggregate-domain` |
| `inherited-handle-restriction` | available: close-on-exec plus explicit descriptor release | available: explicit handle inheritance list |
| `descendant-exec-denial` **(new)** | unavailable: `no-unprivileged-per-process-exec-policy` | available **only when the effective `exec` set is empty**: all-or-none child-process creation policy |

The first five cells are copied verbatim from `rc5-native-control-inventory-v1`
(`core.md:402-408`, `manager.md:246-257`, `go-host-execution-policy.json`) because the underlying host
facts do not depend on what the child is. Copying rather than referencing keeps each inventory
independently versionable, which `core.md:414-417` requires.

`descendant-exec-denial` is the one genuinely new cell, and it is where Windows is *stronger* than macOS.
0008 §3 anticipates it ("Where the platform inventory provides descendant-exec denial, it is applied and
recorded"). Windows offers an all-or-none child-process creation restriction; macOS has no unprivileged
per-process equivalent (its dynamic sandbox interface is deprecated, and App Sandbox behavior is
entitlement- and packaging-dependent — both recorded as findings in decision 0006's rejected alternatives).

**All-or-none forces one schema consequence:** the control is applicable only when the effective `exec`
set is empty, which is the deny-by-default case and therefore the common one. Express it as an
`applicable_when` predicate on the inventory entry plus a third `status` value `not-applicable` in the
script evidence record. That is legal precisely because the script record is a new closed version (F2) —
no existing bytes change.

### Rationale

1. **Withholding Windows would repeat the mistake 0006 was written to correct.** 0006 rejected
   "ship real builds only on a hardened Linux profile and reject `go-v1` on macOS and Windows" because it
   "withholds the entire compiled-skill feature from the platforms this release actually targets". Scripts
   are the *more* portable command type; a macOS-first enforcement release would be strictly worse.
2. **Every item in the mandatory set is environment-, process-, and path-level, which is where Windows is
   at parity.** Nothing in items 1–9 needs a kernel confinement primitive. Manager-built environment,
   controlled `PATH`, private temp, handle hygiene, identity verification, and domain teardown all have
   direct Windows implementations — teardown is in fact *stronger* there (Job Object kill-on-close vs
   process-group teardown).
3. **The asymmetries are exactly the ones already balanced in the build ledger.** macOS gives
   `RLIMIT_FSIZE`; Windows gives Job Object process-count and memory limits. Neither platform dominates,
   which is why the inventory rather than the mandatory set is the right home for both.
4. **The platform-case ledger discipline is what makes this safe.** Note for the prose task: "platform-case
   ledger" is not a defined term in the specification — it is descriptive of the discipline 0006 established
   (verified: the phrase appears nowhere in the repo outside 0008 line 138). That discipline is: an exhaustive,
   versioned, per-platform table; every cell explicitly available-with-mechanism or unavailable-with-reason;
   a closed reason vocabulary; probed per operation, never inferred from a host label
   (`core.md:410-417`). Either adopt the phrase and define it, or keep using "native-control inventory".
   Do not use it undefined in normative text.
5. **A new unavailable-reason code is required and must be added deliberately.** The current vocabulary has
   exactly one entry, `no-private-aggregate-domain` (`go-host-execution-policy.json`,
   `native_control_inventory.unavailable_reasons`). `descendant-exec-denial` on macOS is unavailable for a
   different reason, so the script inventory's reason vocabulary gains
   `no-unprivileged-per-process-exec-policy`. Reusing the aggregate-domain reason would be false.
6. **Fact-check caveat, stated plainly.** The Windows child-process restriction and the macOS
   deprecated-sandbox finding are corroborated *inside this repository* by decision 0006's rejected
   alternatives ("Deprecated dynamic sandbox interfaces, entitlement- and packaging-dependent App Sandbox
   behavior, all-or-none Windows child-process policy…"). The exact Windows attribute name and its behavior
   under nested jobs were **not** verified against vendor documentation in this analysis. The prose task
   must confirm both before either becomes a normative mechanism string, and the manager-profile task must
   confirm Job Object nesting behavior before relying on `active-process-count-limit` inside an existing job.

### Rejected options

- **macOS-only first release; Windows declared-only.** Rejected: repeats 0006's rejected shape and leaves
  the majority-platform story unspecified and unversioned when it lands.
- **Put `descendant-exec-denial` in the mandatory portable set.** Rejected: macOS cannot provide it
  unprivileged, so every macOS host would reject every enforced command — the fail-closed outcome 0006
  replaced.
- **Promote `descendant-domain-termination` and `inherited-handle-restriction` to mandatory because both
  cells are available.** Rejected: 0006 keeps them in the inventory even though both cells are available, so
  the ledger stays comparable when a third platform column is added. Mirror it; do not optimize.
- **Reference `rc5-native-control-inventory-v1` from the script policy instead of copying it.** Rejected:
  a shared inventory means a build-motivated revision silently re-scopes script conformance, which
  `core.md:414-417` treats as a specification revision requiring its own version.
- **Drop `descendant-exec-denial` from the first release entirely.** Rejected: it is available on Windows
  for the deny-by-default case that most enforced commands will be in, and 0008 §3 already promises it where
  the inventory provides it. The cost is one predicate and one status value inside a brand-new record version.
- **Add a Linux column now.** Rejected: F5 — no execution-policy ledger in this specification has one, and
  adding it here would fork the platform set between the build and script policies.

---

## 7. Fact-check ledger

| # | Claim | Source | Verdict |
|---|---|---|---|
| 1 | `capabilities` is manifest-level, not per command | `schemas/v1/agent-skill-v7.schema.json` top-level properties; `common.schema.json` `$defs.commandV7` has no capability field | verified |
| 2 | `"repo"` / `"home-config"` have no normative semantics | 2 hits repo-wide: `core.md:557`, `common.schema.json:535` | verified |
| 3 | `filesystem` defaults to `"repo"`, contradicting 0008's "private runtime area" default | `common.schema.json` `$defs.capabilities.filesystem.default` vs `0008` §3 last bullet | verified conflict |
| 4 | `capability-evidence-v1` rejects a non-`manager-worker-v1` policy | `core.md:437`; vector case `hardened-execution-policy-in-evidence-record` | verified |
| 5 | A host-filtering control name is already a rejected out-of-inventory example | vector case `unknown-native-control-is-rejected`, control `host-firewall-profile` | verified |
| 6 | Network host globs are bare hosts — no scheme, port, or path | `common.schema.json` pattern `^[^\s/\\]+$` | verified |
| 7 | A cached availability result is not a probe | `core.md:412-414`; `manager.md:211-213` | verified |
| 8 | Evidence is result-only and excluded from cache/receipt/marker/claim | `core.md:439-442`; `manager.md:271-273` | verified |
| 9 | Today's script launcher preserves inherited `PATH` and may be a symlink or shell wrapper | `manager.md:503-509`; `curator/internal/runtimestore/runtimestore.go:119-145` | verified in spec and implementation |
| 10 | Build policy mandates closed stdin | `core.md:369-370` | verified (and deliberately not carried over) |
| 11 | No Linux platform cell exists in any execution-policy inventory | zero `linux` hits in `core.md` / `manager.md`; `native_control_inventory.platforms = [macos, windows]` | verified |
| 12 | The far end of the axis is now 0007's `verified` mode, not an unnamed "hardened profile" | `decisions/0007`; `provider-capability-receipt-v1.schema.json` six capability ids | verified |
| 13 | "platform-case ledger" is not a defined term | only occurrence outside 0008 line 138 is in two `.temp/` copies of the same file | verified |
| 14 | §8.2 hashes the entire `GOROOT` tree and requires it unchanged through last child exit | `core.md:942-968` | verified |
| 15 | Closed identifiers are admitted only by specification revision | `core.md:1290-1292` and `core.md:1306-1311` | verified |
| 16 | Windows all-or-none child-process policy; macOS deprecated dynamic sandbox | `decisions/0006` rejected alternatives | corroborated in-repo only — **vendor confirmation still required** |
| 17 | cocoaskills manager is itself a Python package (so its worker graph includes a mutable interpreter) | `cocoaskills-production/pyproject.toml`, `src/csk` | verified |

## 8. Items for the maintainer (non-blocking; recorded, not asked)

Under the 2026-08-22 pre-authorization these are decided as recommended above and need no approval.
They are listed because they are the places where a later reviewer is most likely to want a different call:

1. F1 — defining capability-value semantics inside the `script-worker-v1` subsection rather than fixing
   §4.3 globally. Scoped narrowly on purpose; a global fix would touch every schema version.
2. F2 — a new `script-capability-evidence-v1` record version instead of reusing `capability-evidence-v1`
   as 0008's text literally says. The decision text is not implementable as written; this is the minimal
   correction.
3. Q2 — introducing a closed `interpreter` identifier is a schema surface 0008 did not anticipate. It is
   the smallest construct that keeps program selection out of package control.
4. Q5 — `descendant-exec-denial` brings an `applicable_when` predicate and a `not-applicable` status into
   the evidence record shape. Both are confined to the new record version.
5. Claim 16 needs vendor-documentation confirmation before the mechanism strings become normative.
