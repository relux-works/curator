// Package windsurf delegates the Windsurf harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/windsurf.
package windsurf

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/windsurf"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Windsurf.
const ID harnessapi.ID = "windsurf"

// Adapter delegates the Windsurf harness adapter type to curator-runtimes.
type Adapter = windsurf.HarnessAdapter

// New constructs the Windsurf harness adapter.
func New() harnessapi.Adapter { return windsurf.NewHarnessAdapter() }
