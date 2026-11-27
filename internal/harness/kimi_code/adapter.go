// Package kimi_code delegates the Kimi Code managed-home adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/kimi.
package kimi_code

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/kimi"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Kimi Code.
const ID harnessapi.ID = "kimi_code"

// Adapter delegates the Kimi Code harness adapter type to curator-runtimes.
type Adapter = kimi.HarnessAdapter

// New constructs the Kimi Code harness adapter.
func New() harnessapi.Adapter { return kimi.NewHarnessAdapter() }

// ManagedHomeSpec returns the Kimi Code managed-home specification.
func ManagedHomeSpec() harnessapi.ManagedHomeSpec { return kimi.ManagedHomeSpec() }
