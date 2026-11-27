package harnessapi

import upstream "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// OperationNotPorted is returned by staged harness adapters for operations not yet ported.
const OperationNotPorted = upstream.OperationNotPorted

// UnportedOperations provides default OperationNotPorted stubs for staged adapters.
type UnportedOperations = upstream.UnportedOperations
