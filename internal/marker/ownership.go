package marker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
)

// DiagOwnershipUnproven identifies canonical entries whose preceding ownership
// cannot be established. It does not imply authorization to adopt their bytes.
const DiagOwnershipUnproven = "source_ownership_unproven"

// CheckOwnership permits proven absence or a preceding owned canonical node.
// Source markers prove ownership through their object/schema/name subset, not
// build validity. Legacy markers retain their full preceding validation rule.
func CheckOwnership(installedDir string) error {
	if err := checkOwnership(installedDir); err != nil {
		return fmt.Errorf("%s: %s: %w", DiagOwnershipUnproven, installedDir, err)
	}
	return nil
}

func checkOwnership(installedDir string) error {
	// The project/manager home is the caller's trusted root. Below that root,
	// neither the store parent, store, node nor marker may redirect through links.
	for _, path := range []string{filepath.Dir(filepath.Dir(installedDir)), filepath.Dir(installedDir), installedDir} {
		entry, err := stateread.Lstat(path)
		if err != nil {
			return err
		}
		if entry.Kind == stateread.KindAbsent {
			continue
		}
		if !entry.Info.IsDir() {
			return fmt.Errorf("%s: canonical component %s must be a directory without links", stateread.DiagUnreadable, path)
		}
	}
	entry, err := stateread.Lstat(installedDir)
	if err != nil {
		return err
	}
	if entry.Kind == stateread.KindAbsent {
		return nil
	}
	path := filepath.Join(installedDir, Name)
	evidence, err := stateread.Lstat(path)
	if err != nil {
		return err
	}
	if evidence.Kind == stateread.KindAbsent || !evidence.Info.Mode().IsRegular() {
		return fmt.Errorf("ownership marker must be a regular readable file")
	}
	file, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || !os.SameFile(info, evidence.Info) {
		return fmt.Errorf("ownership marker changed while opening")
	}
	payload, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if err := protocoljson.Validate(payload); err != nil {
		return err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return fmt.Errorf("ownership marker must be a JSON object")
	}
	var version int
	var name string
	if err := json.Unmarshal(object["schema_version"], &version); err != nil {
		return fmt.Errorf("ownership schema identity must be an integer")
	}
	if err := json.Unmarshal(object["name"], &name); err != nil || name != filepath.Base(installedDir) {
		return fmt.Errorf("ownership marker must name this node")
	}
	if version == SchemaV5 || version == SchemaV6 {
		return nil
	}
	// Existing legacy/global installations predate source markers. This rule
	// preserves their full-schema ownership contract without accepting unknown versions.
	if version < LegacySchemaVersion || version > PolicySchemaVersion {
		return fmt.Errorf("unrecognized ownership schema %d", version)
	}
	recorded, kind, err := ReadState(installedDir)
	if err != nil {
		return err
	}
	if kind != stateread.KindPresent || recorded == nil {
		return fmt.Errorf("no valid preceding legacy marker")
	}
	return nil
}
