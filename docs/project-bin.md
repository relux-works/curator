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
additional executable text is foreign. Legacy Unix symlinks must point to the
recorded managed runtime. Recognized legacy launchers update and remove through
the same journal as current launchers. A filename, shebang, or arbitrary link to
another file does not qualify. Legacy enforced binary launchers without a
fingerprint record are conservatively preserved rather than adopted.

Use `curator install` to reconcile the project after resolving a collision.
Move a foreign entry yourself if you intend Curator to own that command name.
No install or removal silently moves or deletes foreign entries.
