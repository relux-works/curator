# Project command ownership

Project installs publish command shims under `.agents/bin`. A valid filename
alone never proves Curator owns an entry.

Before replacing a declared command, Curator checks the previous ownership
record and the live entry. A foreign collision fails the entire install with
`project_bin_unmanaged_conflict: <path>`. An unrelated foreign entry remains
unchanged and is reported with `project_bin_unmanaged_preserved: <path>`.
The same rule preserves a foreign file placed over a formerly managed shim,
including when the skill is removed.

The machine-local `.agents/.curator-bin-ownership.json` ledger records the SHA-256
fingerprint of each published shim and enforced-launcher sidecar. It is committed
in the installation transaction, so rollback restores the prior ledger and
launchers together. A fingerprint includes the entry kind; a symlink is checked
as a link, without reading its destination. A malformed ledger refuses with
`project_bin_ownership_invalid`; it never authorizes an overwrite. Unreadable
ownership inventories or legacy markers refuse with `manager_state_unreadable`,
so a failed read cannot authorize legacy adoption as though state were absent.

For installations predating the ledger, ownership requires a previous installed
skill marker naming the command and a recognized launcher to that marker's
runtime or exact artifact cache entry. Unix script and Windows `.cmd` wrappers
must match the complete generated template, including their PATH prefix;
additional executable text is foreign. Windows recognition also includes the
original direct wrapper, the lifecycle PATH-prefix wrapper, and the earlier
single-pass `call` escaping template. Each must match its complete historical
bytes and the previously recorded target. Legacy Unix symlinks must point to the
recorded managed runtime. Recognized legacy launchers update and remove through
the same journal as current launchers. A filename, shebang, or arbitrary link to
another file does not qualify. Legacy enforced binary launchers use the explicit marker-and-sidecar
adoption evidence described below.

Use `curator install` to reconcile the project after resolving a collision.
Move a foreign entry yourself if you intend Curator to own that command name.
No install or removal silently moves or deletes foreign entries.

Ownership inspection is read-only and precedes builds and the dry-run return,
after planning has already fetched the declared sources.
Only desired paths, ledger entries and legacy candidates are fingerprinted;
unrelated foreign files are reported without reading their contents.
Unlike a user-bin collision (which warns and skips one forwarding command),
a project-bin collision refuses the entire installation.

Legacy enforced commands are adopted only when a previous installed marker
names the command, the marker-selected frozen manifest declares it enforced,
and the sidecar bytes exactly match the contract derived from that manifest
and marker. The old manager executable's bytes cannot be proved; this is the
explicit limit of this legacy adoption evidence. Removing an adopted command
removes both launcher and sidecar. Missing or mismatched evidence refuses with
`project_bin_legacy_enforced`; move the launcher and sidecar aside manually
before reinstalling. Deleting a current ledger uses the same legacy checks.

Recovery: preserve your entries and ledger before making repairs. For a foreign
collision, move the named foreign entry before reinstalling. For malformed
ownership state (`project_bin_ownership_invalid`), restore a trusted ledger
backup. If no backup exists, move the malformed ledger aside manually; the next
install attempts legacy adoption and refuses any unprovable enforced pair.
Unreadable state (`manager_state_unreadable`) needs its permissions or parent
path repaired before retrying. A refusal never moves a file automatically.

A stale marker with no live launcher or sidecar does not require loading its
frozen manifest. If a live candidate needs that manifest and it cannot be
loaded, ownership remains unknown: `project_bin_legacy_enforced` asks you to
move the launcher and sidecar aside before reinstalling. Desired, ledger-named
and legacy file read failures retain `manager_state_unreadable` and its remedy.
The dry-run report includes preserved unrelated entries as well as refusals.
