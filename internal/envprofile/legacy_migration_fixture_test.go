package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Legacy records are fixture inputs only. Production never writes this format:
// unsealed records lack origin/parent preimages and cannot authorize recovery.
// writeMigrationJournal writes an explicit historical fixture. This format
// is never emitted by production and cannot authorize forward recovery.
func writeMigrationJournal(home string, journal *migrationJournal) error {
	payload, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	dir := filepath.Dir(migrationJournalPath(home))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".env-migration-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, migrationJournalPath(home)); err != nil {
		return err
	}
	return nil
}
