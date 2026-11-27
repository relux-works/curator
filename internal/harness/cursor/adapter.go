// Package cursor delegates the Cursor harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/cursor.
package cursor

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/cursor"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Cursor.
const ID harnessapi.ID = "cursor"

// Adapter delegates the Cursor harness adapter type to curator-runtimes.
type Adapter = cursor.HarnessAdapter

// New constructs the Cursor harness adapter.
func New() harnessapi.Adapter { return cursor.NewHarnessAdapter() }
