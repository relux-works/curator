// Package gemini delegates the Gemini harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/gemini.
package gemini

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/gemini"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Gemini.
const ID harnessapi.ID = "gemini"

// Adapter delegates the Gemini harness adapter type to curator-runtimes.
type Adapter = gemini.HarnessAdapter

// Capabilities returns the declared Gemini harness capabilities.
func Capabilities() harnessapi.CapabilitySet { return gemini.HarnessCapabilities() }

// New constructs the Gemini harness adapter.
func New() harnessapi.Adapter { return gemini.NewHarnessAdapter() }
