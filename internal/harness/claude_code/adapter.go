// Package claude_code delegates the Claude Code harness adapter to
// github.com/relux-works/curator-runtimes/pkg/agentic/systems/claude.
package claude_code

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/claude"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ID is the canonical Curator harness identifier for Claude Code.
const ID harnessapi.ID = "claude_code"

// Adapter delegates the Claude Code harness adapter type to curator-runtimes.
type Adapter = claude.HarnessAdapter

// New constructs the Claude Code harness adapter.
func New() harnessapi.Adapter { return claude.NewHarnessAdapter() }
