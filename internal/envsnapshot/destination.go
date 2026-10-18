package envsnapshot

// DestinationInvalid is the normative destination refusal, distinct from a
// malformed or changed prior snapshot (Mismatch).
const DestinationInvalid = "environment_materialize_destination_invalid"

func invalidDestination(reason string) error { return &Error{DestinationInvalid, reason} }

// DestinationPolicy supplies the complete existing protected-root inventory.
// The materialize coordinator must supply manager roots, stores, managed/native
// homes, credential roots, repository administration and the active workspace.
// Unreadable or missing supplied roots refuse; they are never silently omitted.
// Inventory discovery and capture authorization are separate from admission.
type DestinationPolicy struct {
	ProtectedRoots []string
}

// DestinationKind describes an observation, never permission to rename a path.
type DestinationKind string

// Admitted destination states; none authorizes publication by pathname.
const (
	DestinationAbsent   DestinationKind = "absent"
	DestinationSnapshot DestinationKind = "snapshot"
)
