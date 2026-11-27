package harnessregistry

import (
	upstream "github.com/relux-works/curator-runtimes/pkg/harnessregistry"
	"github.com/relux-works/curator/internal/harnessapi"
)

// ManagedHomeSpecs returns all registered managed-home harness specifications.
func ManagedHomeSpecs() []harnessapi.ManagedHomeSpec { return upstream.ManagedHomeSpecs() }
