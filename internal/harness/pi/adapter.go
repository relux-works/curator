// Package pi delegates the Pi harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/pi.
package pi

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/pi"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Pi.
const ID harnessapi.ID = "pi"

// Adapter delegates the Pi harness adapter type to curator-runtimes.
type Adapter = pi.HarnessAdapter

// New constructs the Pi harness adapter.
func New() harnessapi.Adapter { return pi.NewHarnessAdapter() }
