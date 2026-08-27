# Integrate atomic global builds

## Description
Give global installs the same schema 6 planning, private build, protected publication, marker v2, shim activation, transaction, rollback, and dry-run semantics as project installs.

## Scope
Own src/csk/global_install.py and the global adapters, env files, global bins, and CLI wiring required for transaction targets. Reuse shared planner, driver, cache, marker, shims, and transaction infrastructure instead of implementing a second build policy. Preserve global manifest and existing partial-source diagnostics, but replace partial materialization with the normative all-or-rollback boundary.

## Acceptance Criteria
Global install and global upgrade accept valid active build commands, build misses outside the home lock, revalidate and publish under the home lock, write marker v2, refresh global and user-bin shims, environments, and adapters transactionally, and preserve argument and exit behavior on Unix and Windows. Any build, publication, shim, user-bin, environment, adapter, marker, or stale-removal failure preserves the prior working global install through reverse rollback. Global dry-run acquires no mutation lock and leaves manager-home bytes unchanged. Script and system-only global installs remain compatible. Focused global, bins, adapter, rollback, dry-run, cross-platform, and strict-mypy gates pass.
