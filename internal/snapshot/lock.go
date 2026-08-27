package snapshot

import (
	"errors"
	"fmt"
	"os"
)

type snapshotLock struct {
	file *os.File
}

func acquireSnapshotLock(path string) (*snapshotLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) // #nosec G304 -- path is manager-derived
	if err != nil {
		return nil, fmt.Errorf("open snapshot lock: %w", err)
	}
	if err := lockFile(file); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("lock snapshot: %w", err)
	}
	return &snapshotLock{file: file}, nil
}

func (lock *snapshotLock) Close() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	return errors.Join(unlockFile(lock.file), lock.file.Close())
}
