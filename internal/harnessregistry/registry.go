// Package harnessregistry delegates the closed compile-time harness catalog to
// github.com/relux-works/curator-runtimes/pkg/harnessregistry.
package harnessregistry

import (
	upstream "github.com/relux-works/curator-runtimes/pkg/harnessregistry"
	"github.com/relux-works/curator/internal/harness/codex_cli"
	"github.com/relux-works/curator/internal/harness/opencode"
	"github.com/relux-works/curator/internal/harnessapi"
)

// DefaultID is the default Curator harness identifier.
const DefaultID harnessapi.ID = upstream.DefaultID

// CanonicalID normalizes known harness aliases to their canonical identifier.
func CanonicalID(id harnessapi.ID) harnessapi.ID { return upstream.CanonicalID(id) }

// IDs returns all registered harness identifiers in canonical order.
func IDs() []harnessapi.ID { return upstream.IDs() }

// IDsWithCapability returns registered harness identifiers declaring capability.
func IDsWithCapability(capability harnessapi.Capability) []harnessapi.ID {
	var matched []harnessapi.ID
	for _, id := range IDs() {
		adapter, err := Lookup(id)
		if err != nil {
			continue
		}
		if adapter.Capabilities().Has(capability) {
			matched = append(matched, id)
		}
	}
	return matched
}

// Lookup returns the registered adapter for id.
func Lookup(id harnessapi.ID) (harnessapi.Adapter, error) {
	switch id {
	case codex_cli.ID:
		return codex_cli.New(), nil
	case opencode.ID:
		return opencode.New(), nil
	default:
		return upstream.Lookup(id)
	}
}

// Verify checks adapter API version, identifier, and declared capabilities.
func Verify(version harnessapi.APIVersion, id harnessapi.ID, capabilities harnessapi.CapabilitySet, constructor func() harnessapi.Adapter) (harnessapi.Adapter, error) {
	return upstream.Verify(version, id, capabilities, constructor)
}

// ProbePolicy returns the release probe policy for id.
func ProbePolicy(id harnessapi.ID) (harnessapi.ProbeSpec, error) { return upstream.ProbePolicy(id) }
