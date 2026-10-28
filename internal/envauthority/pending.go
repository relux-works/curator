package envauthority

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

type pendingEffect struct {
	Path          string
	DirectoryOnly bool
}

type Pending struct {
	LeaseJournals []*envjournal.Journal
	Effects       map[string][]pendingEffect
	Units         map[string]*envunit.Unit
	Homes         map[string]envunit.Home
	Artifacts     map[string]bool
}

// PendingUnits reads only the fixed private journal namespace. A pending
// exception requires the complete real journal and reachable preimage state;
// neither an arbitrary journal filename nor a selected caller home is authority.
func PendingUnits(manager string) (*Pending, error) {
	result := &Pending{Effects: map[string][]pendingEffect{}, Units: map[string]*envunit.Unit{}, Homes: map[string]envunit.Home{}, Artifacts: map[string]bool{}}
	root := filepath.Join(manager, "state", "transactions", "v1")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	if len(entries) > 4096 {
		return nil, fmt.Errorf("%s: journal inventory exceeds bound", envorigin.DiagInvalid)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") && !strings.HasSuffix(entry.Name(), ".json.delete") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		raw, present, err := envorigin.ReadPrivateAuthority(manager, path, 16<<20)
		if err != nil {
			return nil, err
		}
		if !present {
			return nil, fmt.Errorf("%s: journal disappeared", envorigin.DiagInvalid)
		}
		var journal envjournal.Journal
		if err := json.Unmarshal(raw, &journal); err != nil {
			return nil, err
		}
		canonical, err := json.Marshal(&journal)
		if err != nil || !bytes.Equal(raw, append(canonical, '\n')) {
			return nil, fmt.Errorf("%s: noncanonical journal", envorigin.DiagInvalid)
		}
		// Legacy journals are still handled by the engine's existing validation.
		// They cannot exempt a home or create a provisioning-origin exception.
		if journal.Unit == nil {
			if journal.ProjectIdentity == "project-leases" {
				if entry.Name() != journal.TransactionID+".json" {
					return nil, fmt.Errorf("%s: lease journal filename differs", envorigin.DiagInvalid)
				}
				copy := journal
				result.LeaseJournals = append(result.LeaseJournals, &copy)
			}
			continue
		}
		if err := envjournal.VerifyUnitState(&journal, manager); err != nil {
			return nil, err
		}
		name := journal.TransactionID + ".json"
		if strings.HasSuffix(entry.Name(), ".delete") {
			name += ".delete"
		}
		if name != entry.Name() {
			return nil, fmt.Errorf("%s: journal filename binding differs", envorigin.DiagInvalid)
		}
		if _, duplicate := result.Units[journal.TransactionID]; duplicate {
			return nil, fmt.Errorf("%s: duplicate pending lifecycle identity", envorigin.DiagInvalid)
		}
		result.Units[journal.TransactionID] = journal.Unit
		for _, target := range journal.Unit.Targets {
			result.Effects[journal.TransactionID] = append(result.Effects[journal.TransactionID], pendingEffect{Path: target.Path, DirectoryOnly: target.Class == "parent"})
		}
		for _, home := range journal.Unit.Homes {
			result.Effects[journal.TransactionID] = append(result.Effects[journal.TransactionID], pendingEffect{Path: home.Root})
		}
		for id, prior := range result.Effects {
			if id == journal.TransactionID {
				continue
			}
			for _, a := range prior {
				for _, b := range result.Effects[journal.TransactionID] {
					if pendingEffectsIntersect(a, b) {
						return nil, fmt.Errorf("%s: overlapping pending unit effects", envorigin.DiagInvalid)
					}
				}
			}
		}
		for _, home := range journal.Unit.Homes {
			if _, duplicate := result.Homes[home.Target]; duplicate {
				return nil, fmt.Errorf("%s: overlapping pending home units", envorigin.DiagInvalid)
			}
			result.Homes[home.Target] = home
		}
		for _, target := range journal.Targets {
			for _, artifact := range []string{target.StagedPath, target.BackupPath, target.RollbackPath} {
				if artifact != "" {
					result.Artifacts[artifact] = true
				}
			}
		}
		for _, directory := range journal.Directories {
			if directory.StagedPath != "" {
				result.Artifacts[directory.StagedPath] = true
			}
		}
		for _, action := range journal.Opaque {
			result.Artifacts[action.QuarantinePath] = true
		}
	}
	return result, nil
}

func pendingEffectsIntersect(a, b pendingEffect) bool {
	if a.DirectoryOnly && b.DirectoryOnly {
		return a.Path == b.Path
	}
	if a.DirectoryOnly {
		return a.Path == b.Path || strings.HasPrefix(a.Path, b.Path+string(filepath.Separator))
	}
	if b.DirectoryOnly {
		return b.Path == a.Path || strings.HasPrefix(b.Path, a.Path+string(filepath.Separator))
	}
	return overlaps(a.Path, b.Path)
}
