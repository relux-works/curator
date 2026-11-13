// Package storecompatibility classifies the frozen revision-1 store replay.
// It proves an exact migration partition, never full revision-1 conformance.
package storecompatibility

type expectation struct{ ID, Check, Witness string }

// This roster is pinned to the released store archive SHA-256
// 7efa5b09e450162cfb2937f25b3e29fab4699602e0990cfd295a4f695cb60413.
// The six old expectations intentionally fail after structural refusal tightens.
var roster = []expectation{
	{"TestStoreBoundaryResolveVectorsDriveResolve/intact-resolve-emits-fragment", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/swapped-system-prompt-bytes-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/swapped-root-context-bytes-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/symlinked-entry-root-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/wrong-ownership-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/wrong-permissions-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/containment-escape-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/non-regular-component-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/lock-file-wrong-owner-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/marker-symlink-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/environments-root-wrong-owner-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/store-root-symlinked-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/intact-updated-store-old-marker-stale", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/swapped-updated-store-old-marker-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/unprovisioned-intact-stale", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/unprovisioned-swapped-untrusted", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/unreadable-marker-non-current", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/swapped-bytes-emits-fragment", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/untrusted-reported-current", "", ""},
	{"TestStoreBoundaryResolveVectorsDriveResolve/untrusted-without-diagnostic", "", ""},
	{"TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan/dry-run-untrusted-reports-would-rebuild", "ownership", "want would-rebuild-untrusted-store"},
	{"TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan/dry-run-intact-plans-nothing", "", ""},
	{"TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan/dry-run-enclosing-no-rebuild", "", ""},
	{"TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan/dry-run-mutates", "ownership", "want would-rebuild-untrusted-store"},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-rebuilds-git-entry-from-snapshot", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-path-entry-cannot-rebuild", "ownership", "want no fragment and environment_repair_failed"},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-local-entry-cannot-rebuild", "permissions", "want no fragment and environment_repair_failed"},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-rebuilds-entry-boundary-failure", "ownership", "want a trusted fragment"},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-enclosing-refuses-no-rebuild", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-stale-old-marker-succeeds", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-swapped-old-marker-never-adopted", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-unprovisioned-intact-provisions", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-unprovisioned-swapped-rebuilds", "", ""},
	{"TestStoreBoundaryRepairVectorsDriveResolveRepair/repair-reapplies-untrusted", "ownership", "want a trusted fragment"},
	{"TestStoreBoundaryStatusVectorsDriveStatus/status-intact-current", "", ""},
	{"TestStoreBoundaryStatusVectorsDriveStatus/status-names-failing-check", "", ""},
	{"TestStoreBoundaryStatusVectorsDriveStatus/status-enclosing-names-boundary", "", ""},
	{"TestStoreBoundaryStatusVectorsDriveStatus/status-hides-failing-check", "", ""},
}
