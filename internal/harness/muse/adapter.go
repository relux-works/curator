// Package muse delegates the Muse harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/muse.
package muse

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/muse"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Muse.
const ID harnessapi.ID = "muse"

// Adapter delegates the Muse harness adapter type to curator-runtimes.
type Adapter = muse.HarnessAdapter

// New constructs the Muse harness adapter.
func New() harnessapi.Adapter { return muse.NewHarnessAdapter() }
