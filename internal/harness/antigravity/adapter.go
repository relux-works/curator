// Package antigravity delegates the Antigravity managed-home adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/agy.
package antigravity

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/agy"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Antigravity.
const ID harnessapi.ID = "antigravity"

// Adapter delegates the Antigravity harness adapter type to curator-runtimes.
type Adapter = agy.HarnessAdapter

// New constructs the Antigravity harness adapter.
func New() harnessapi.Adapter { return agy.NewHarnessAdapter() }

// ManagedHomeSpec returns the Antigravity managed-home specification.
func ManagedHomeSpec() harnessapi.ManagedHomeSpec { return agy.ManagedHomeSpec() }
