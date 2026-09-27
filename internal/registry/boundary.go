package registry

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/stateread"
)

// Page-boundary diagnostics are closed by registry protocol §9.3.
const (
	PageBoundaryMissing  = "registry_page_boundary_missing"
	PageBoundaryMismatch = "registry_page_boundary_mismatch"
	PageBoundaryStale    = "registry_page_boundary_stale"
)

// PageBoundaryError reports a rejected records page and names its registry.
// The FetchFn returns no records for a rejected chain and does not use or
// replace a cache entry for that operation.
type PageBoundaryError struct {
	Diagnostic string
	URL        string
	Detail     string
}

func (e *PageBoundaryError) Error() string {
	return fmt.Sprintf("%s: registry %s %s", e.Diagnostic, e.URL, e.Detail)
}

// pageStateError reports a failure to read or persist protected rollback
// state. It fails closed like a boundary rejection, without inventing a
// page-boundary diagnostic.
type pageStateError struct{ msg string }

func (e *pageStateError) Error() string { return e.msg }

// pageChain applies §9.3 to one cursor chain. The first page's boundary is
// the chain boundary; later pages verify their own signature, then compare
// their exact JSON bytes before anything else. Only the first page checks or
// advances the persisted high-water.
type pageChain struct {
	name      string
	url       string
	keys      []string
	stateDir  string
	stateName string
	stateFile string
	persist   bool
	known     map[string]bool
	state     snapshotState
	chain     []byte
	pages     int
}

// openPageChain reads the existing high-water without creating or repairing
// state. A missing state file listed in the catalog, an unreadable state, or
// an invalid catalog is unavailable, never first use.
func openPageChain(name, registryURL string, keys []string, stateDir string, persist bool) (*pageChain, error) {
	known, err := loadSnapshotStateCatalogReadOnly(stateDir)
	if err != nil {
		return nil, &pageStateError{fmt.Sprintf("registry %s rollback state catalog is unavailable: %v", name, err)}
	}
	stateName := "snapshot-" + urlDigest(registryURL) + ".json"
	stateFile := filepath.Join(stateDir, stateName)
	state, exists, err := readSnapshotState(stateFile)
	if err != nil {
		return nil, &pageStateError{fmt.Sprintf("registry %s rollback state is unreadable: %v", name, err)}
	}
	if !exists && known[stateName] {
		return nil, &pageStateError{fmt.Sprintf("registry %s rollback state is missing after prior use", name)}
	}
	return &pageChain{
		name: name, url: registryURL, keys: keys,
		stateDir: stateDir, stateName: stateName, stateFile: stateFile,
		persist: persist, known: known, state: state,
	}, nil
}

func (c *pageChain) verifiedBoundary(boundary map[string]any) (parsedSnapshot, error) {
	parsed, err := parseSnapshot(boundary)
	if err != nil || !VerifySigned(boundary, c.keys) {
		return parsedSnapshot{}, &PageBoundaryError{
			Diagnostic: PageBoundaryMissing,
			URL:        c.url,
			Detail:     "served a records page whose boundary failed verification",
		}
	}
	return parsed, nil
}

// acceptPage processes one page boundary in §9.3 order. Call it before the
// page's records are added to the chain's result. A later page never checks or
// advances high-water state.
func (c *pageChain) acceptPage(boundary map[string]any, raw []byte) error {
	parsed, err := c.verifiedBoundary(boundary)
	if err != nil {
		return err
	}
	if c.pages > 0 {
		if !bytes.Equal(raw, c.chain) {
			return &PageBoundaryError{
				Diagnostic: PageBoundaryMismatch,
				URL:        c.url,
				Detail:     "served a records page whose boundary differs from the chain boundary",
			}
		}
		c.pages++
		return nil
	}
	if parsed.Version < c.state.HighestVersion {
		return &PageBoundaryError{
			Diagnostic: PageBoundaryStale,
			URL:        c.url,
			Detail:     "served a records page boundary below the persisted high-water",
		}
	}
	if parsed.Version == c.state.HighestVersion && c.state.Head != "" {
		if c.state.Head != parsed.Head || c.state.MerkleRoot != parsed.MerkleRoot || c.state.LogSize != parsed.LogSize {
			return &PageBoundaryError{
				Diagnostic: PageBoundaryStale,
				URL:        c.url,
				Detail:     "served a records page boundary that conflicts with the persisted high-water",
			}
		}
		c.chain = append([]byte(nil), raw...)
		c.pages = 1
		return nil
	}
	if c.persist {
		if err := os.MkdirAll(c.stateDir, 0o700); err != nil {
			return &pageStateError{fmt.Sprintf("registry %s rollback state directory is unavailable: %v", c.name, err)}
		}
		_ = os.Chmod(c.stateDir, 0o700) // #nosec G302 -- protected state directory needs traversal bits
		// The shared snapshot-state writer atomically publishes the higher
		// high-water and boundary posture before this page contributes.
		next := snapshotState{
			HighestVersion:   parsed.Version,
			Head:             parsed.Head,
			MerkleRoot:       parsed.MerkleRoot,
			LogSize:          parsed.LogSize,
			BoundaryVerified: true,
		}
		if err := writeSnapshotState(c.stateFile, next); err != nil {
			return &pageStateError{fmt.Sprintf("registry %s rollback state could not be persisted: %v", c.name, err)}
		}
		c.state = next
		if !c.known[c.stateName] {
			c.known[c.stateName] = true
			if err := writeSnapshotStateCatalog(c.stateDir, c.known); err != nil {
				return &pageStateError{fmt.Sprintf("registry %s rollback state catalog could not be persisted: %v", c.name, err)}
			}
		}
	}
	c.chain = append([]byte(nil), raw...)
	c.pages = 1
	return nil
}

// cacheBoundaryCurrent accepts a cached response only when its signed
// boundary matches the already-persisted high-water. A cache entry can never
// establish first-use state or advance a high-water.
func (c *pageChain) cacheBoundaryCurrent(boundary map[string]any, raw []byte) bool {
	parsed, err := c.verifiedBoundary(boundary)
	if err != nil || c.state.Head == "" || parsed.Version != c.state.HighestVersion {
		return false
	}
	return c.state.Head == parsed.Head && c.state.MerkleRoot == parsed.MerkleRoot && c.state.LogSize == parsed.LogSize && len(raw) > 0
}

// BoundaryPosture is one trusted registry's read-only §9.3 status row.
type BoundaryPosture struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	HighWaterVersion     *int   `json:"high_water_version"`
	HighWaterLogSize     *int   `json:"high_water_log_size"`
	LastBoundaryVerified bool   `json:"last_boundary_verified"`
	Diagnostic           string `json:"diagnostic,omitempty"`
}

// ReadBoundaryPosture reports persisted state in trusted registry order and
// never creates, repairs, or mutates protected state. An unreadable or
// missing-after-use state is reported as a diagnostic rather than absence.
func ReadBoundaryPosture(stateDir string, registries []Registry) []BoundaryPosture {
	known, catalogErr := loadSnapshotStateCatalogReadOnly(stateDir)
	rows := make([]BoundaryPosture, 0, len(registries))
	for _, reg := range registries {
		if len(reg.PublicKeys) == 0 {
			continue
		}
		row := BoundaryPosture{Name: reg.Name, URL: reg.URL}
		if catalogErr != nil {
			row.Diagnostic = fmt.Sprintf("rollback state catalog is unavailable: %v", catalogErr)
			rows = append(rows, row)
			continue
		}
		stateName := "snapshot-" + urlDigest(reg.URL) + ".json"
		state, exists, err := readSnapshotState(filepath.Join(stateDir, stateName))
		switch {
		case err != nil:
			row.Diagnostic = fmt.Sprintf("rollback state is unreadable: %v", err)
		case !exists && known[stateName]:
			row.Diagnostic = "rollback state is missing after prior use"
		case !exists:
			// First use has no persisted high-water yet.
		default:
			version, logSize := state.HighestVersion, state.LogSize
			row.HighWaterVersion = &version
			row.HighWaterLogSize = &logSize
			row.LastBoundaryVerified = state.BoundaryVerified
		}
		rows = append(rows, row)
	}
	return rows
}

// ReadOnlyStateDir preserves access to pre-migration rollback state. Only a
// proven absence selects the legacy directory; unreadable state keeps the
// protected path selected so its later read fails closed instead of falling
// back to legacy state.
func ReadOnlyStateDir(stateDir, legacyDir string) string {
	metadata, err := stateread.Stat(stateDir)
	if err == nil && metadata.Kind == stateread.KindAbsent {
		return legacyDir
	}
	return stateDir
}
