# Compiled-command status, diagnostics, and repair

`curator status` reports read-only status diagnostics for project closures. The command reports one status code per declared skill and, when the closure activates compiled (`go-v1`) commands, one diagnostic line per active build command. Running `curator status --json` formats diagnostics as a JSON document containing a `builds` array; a closure without compiled commands produces the historical document unchanged, with no `builds` key.

## Status codes and checks

Status codes are machine-readable. Only `up-to-date` for skills and `current` for compiled commands indicate an exactly synchronized state. Running `curator status --check` exits non-zero for any other status value, including unrecognized state codes.

Reporting and verdict are separate. Plain `curator status` exits zero when it produces the requested report, including cases where the plan itself refused as long as every active compiled command received a row (raw details write as warnings to standard error). A refusal that leaves commands undescribed exits non-zero, as does any failure in a scope without compiled commands. The `--check` flag turns a non-current verdict into a non-zero exit code.

The command reports the following status codes for skills and compiled commands:

| Code | Meaning |
|---|---|
| `up-to-date` / `current` | every check passed |
| `not-installed` | the declaration has no installed skill |
| `invalid-marker` | the install marker is absent, unreadable, or invalid |
| `unsupported-marker` | the marker schema cannot be read by this manager |
| `needs-install` | the installation is behind its declaration, or its marker schema cannot describe a compiled command |
| `content-drift` | installed content no longer hashes to the recorded value |
| `unresolvable` | the declared ref cannot be resolved in the source repository |
| `build-context-exposed` | a build root reached agent-facing context |
| `build-command-drift` | the recorded and activated compiled command sets differ |
| `build-source-drift` | the recorded build-source identity no longer matches the raw snapshot |
| `build-input-drift` | the recorded logical key was derived from a different build input |
| `unsupported-build-driver` | a recorded or planned driver outside `go-v1` |
| `unusable-build-toolchain` | the trusted Go toolchain could not be resolved or verified, so nothing could be planned |
| `missing-build-artifact` | the protected cache holds no entry for the recorded key |
| `corrupt-build-receipt` | the entry's canonical receipt differs from the recorded one |
| `build-artifact-drift` | the entry's artifact path or hash differs from the recorded one |
| `corrupt-build-cache` | the protected entry cannot be interpreted |
| `untrusted-build-cache` | candidate bytes are outside a provable protected boundary |
| `unsupported-build-platform` | this host cannot prove protected build cache state |
| `build-state-changed` | the install marker or the protected cache evidence moved while status was classifying it |
| `unknown-build-state` | a planner outcome this manager does not know; it fails closed |

A row may carry a `cause`, a stable subcode that refines status details without widening the primary status vocabulary. A `build-input-drift` row carries one of three cause subcodes (`build-root`, `target`, `unattributed`), while `unusable-build-toolchain` carries the `go-v1` boundary code that refused the operation. Every other state leaves the `cause` field empty.

The cause subcodes for `build-input-drift` are:

| Cause | Meaning |
|---|---|
| `build-root` | the marker does not record the build root the closure now activates |
| `target` | the marker's recorded artifact path is not the one this target derives |
| `unattributed` | the key differs, and the marker records no prior input to attribute it |

## Logical cache keys and diagnostics

Curator derives logical cache keys by hashing the complete build input: schema version, driver, build source, build root, command, source directory, native target and tuning, trusted toolchain identity, and fixed manager build policy. An install marker records no prior input, so a key mismatch reports as input drift and attributes only as far as recorded build roots and artifact paths can prove. Curator does not guess which input moved.

Compiled verdicts are bracketed on both sides. `curator status` fingerprints install markers before planning and re-reads them afterwards; it also re-takes protected cache lookups once classification finishes. If markers or cache entries move during execution, Curator reports `build-state-changed` instead of publishing a stale verdict.

Diagnostics report driver, build root, source directory, build-source identity, native target and tuning, logical cache key, manager-derived artifact path, and read-only cache outcome. All published paths are protocol-relative (manager home, cache, staging, and probe locations are never published). Untrusted details collapse onto a single line, stripped of non-printable characters, path-redacted, and length-bounded. Toolchain failures prior to logical identity derivation report driver, build root, source directory, and validated build-source digest, omitting target, key, and cache outcome.

Dry runs with `curator install --dry-run` and `curator upgrade --dry-run` run no compiler. Per active build command they report plan statuses: `cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`, `corrupt`, `unsupported`, or `toolchain-unavailable`. These report a plan, never a completed compiler check.

## Reconciliation and repair

Curator provides no separate repair command: `curator install` and `curator upgrade` act as the reconciliation path. Commands rebuild missing, corrupt, drifted, or untrusted entries into new protected state after passing manifest, closure, collision, requirement, audit, registry, and moved-tag gates. Unusable entries are quarantined and replaced under the manager-home lock, never adopted by changing permissions or rewriting a marker. A failed gate, preflight, build, or commit leaves previous installations, consumers, and live caches unchanged, and the run says so.

The build cache is not a transaction target: launchers point only at live entries, so replacements are selected before dependent installations become durable. A run that fails subsequently restores the cache before releasing the manager-home lock. The replacement is withdrawn and the displaced entry is restored by renaming within the protected cache root. Nothing is deleted, so ordinary garbage collection sweeps any remaining artifacts.

Selecting a replacement involves quarantining the predecessor, renaming the entry into the freed slot, validating the entry, and syncing the cache root. A publication that fails part-way puts back what it moved before reporting. Quarantining an entry is a rename plus the sync that makes it durable; a sync failure returns the entry to the live slot rather than reporting failure with an empty slot. That return is synced as well; only if its own sync fails does the run report the cache as changed with the entry live and readable. A failed publication leaves the cache exactly as it found it, and reversal is only needed for a publication that fully succeeded.

Restoration is refused when an incomplete transaction still references the published entry, or when a journaled target moved from its initial state. Restoring an unusable predecessor over an entry a recovered commit will reference would break the installation, so the rebuilt entry is retained.

Whenever a run leaves the live cache changed (a kept entry, a publication that could not undo moves, or a partial reversal), Curator says so per command instead of repeating the ordinary claim that the live build cache is unchanged. The warning names which of the three conditions occurred. The installation and its consumers remain unchanged either way; nothing on these paths is ever deleted, so the state left behind is always one a later `install` or `upgrade` repairs.

Two states fail closed without repair attempts: a host that cannot prove protected cache state (`unsupported`) and a trusted Go toolchain that cannot be resolved or verified (`toolchain-unavailable`). Both conditions refuse execution before disk mutation.

## Global status

`curator global status` reports machine-wide scope status using the same stable vocabulary: one code per declared skill and one diagnostic line per active compiled command. The command accepts `--check` and `--json` flags.

`curator global status` runs a read-only plan equivalent to `curator global install --dry-run`. The command resolves the machine-wide closure and passes read-only audit and registry gates without running compilers or mutating target paths, cache entries, or trust state.

Global status differs from project status in two ways:

- The machine-readable document carries `alias`, `skills`, and (when compiled commands are active) `builds`. It carries no `path` attribute.
- Plain `curator global status` always exits zero. Declared skill reports read directly from install markers; scopes without compiled commands print standard lines even if planning fails (refusals write as warnings to standard error). The `--check` flag is the only surface that turns a verdict into a non-zero exit code. It fails closed twice over: once for every non-current code, and once when the plan refused before describing every compiled command, because such a run cannot prove the scope is current.

A machine-wide scope without a `Skillfile.json` declares and activates nothing; it prints nothing and passes `--check`.

Curator selects trusted Go installations exclusively via `CURATOR_GO` (an absolute `<GOROOT>/bin/go`, `bin/go.exe` on Windows) or `GOROOT`. Curator never searches `PATH` environment variables and never downloads toolchain binaries. The manager accepts only release families tested against the `go-v1` vectors. A missing, untrusted, or untested toolchain reports the failing boundary together with those selection mechanisms and the tested families.

## Maintenance and garbage collection

`curator gc` executes a single serialized maintenance pass. The command acquires the exclusive manager-home mutation lock, recovers any incomplete install transaction, and only then marks and sweeps. This ordering guarantees `gc` cannot race an install, a rollback, or a recovery, and cannot lose a consumer registry update. The same pass runs automatically at the end of every installation under the manager-home lock.

Marking reads live project, global, and hybrid scopes once. Runtime store entries are marked from all supported install marker schemas; compiled build cache entries are marked from marker v2 build state and transaction journals.

Anything the pass cannot prove keeps its artifacts across passes. A consumer registry that exists but does not match the exact shape Curator writes is reported and left untouched rather than rewritten; a registered checkout is unregistered only once its scope is proven absent or proven valid and empty. Ambiguous registries (documents stating `schema_version` or `consumers` more than once) are refused by every reader and writer. Skill directories or installed skills that are symbolic links or reparse points are refused instead of followed. Unreadable or invalid markers block the build sweep.

The sweep removes a protected build cache entry only when all of the following conditions hold:

- No marker and no journal references the entry.
- The cache root and entry remain verifiable manager-protected state.
- The entry is structurally exact and self-consistent with the logical key encoded in its directory name.
- The entry was published more than 24 hours ago (Curator's documented grace period).

Everything else is retained and reported as a maintenance warning (corrupt receipts, untrusted roots, symlink or reparse escapes, ownership or DACL failures). Entry content is never executed, adopted, or permission-repaired, and a receipt alone is never treated as proof of provenance or of a live consumer. Retaining an unreferenced entry remains safe: the only cost of removal is a subsequent rebuild.

Every decision and removal binds to the directory object the pass proved, not to the pathname. Candidate classification (its exact members, receipt, artifact bytes, and size) reads through the descriptor of the resolved entry, and renames or deletions resolve through the proven cache root. An entry whose parent is no longer that object is retained and reported. Exchanging the cache-root path after validation can neither redirect a removal outside the Curator cache root nor let a planted replacement supply the verdict for an entry being removed.
