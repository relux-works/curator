package managerlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const lockRetryInterval = 10 * time.Millisecond

type fileLock struct {
	mu        sync.Mutex
	file      *os.File
	held      bool
	onRelease func()
}

func acquireFileLock(ctx context.Context, path string, onRelease func()) (*fileLock, error) {
	if ctx == nil {
		return nil, fmt.Errorf("lock context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create manager lock state: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) // #nosec G304 -- path is derived below the canonical manager home
	if err != nil {
		return nil, fmt.Errorf("open manager lock: %w", err)
	}

	for {
		locked, lockErr := tryFileLock(file)
		if lockErr != nil {
			_ = file.Close()
			return nil, fmt.Errorf("acquire manager lock: %w", lockErr)
		}
		if locked {
			if err := ctx.Err(); err != nil {
				_ = unlockFile(file)
				_ = file.Close()
				return nil, err
			}
			return &fileLock{file: file, held: true, onRelease: onRelease}, nil
		}

		timer := time.NewTimer(lockRetryInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			_ = file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (lock *fileLock) AssertHeld() error {
	if lock == nil {
		return ErrNotHeld
	}
	lock.mu.Lock()
	defer lock.mu.Unlock()
	if !lock.held || lock.file == nil {
		return ErrNotHeld
	}
	return nil
}

func (lock *fileLock) Close() error {
	if lock == nil {
		return nil
	}
	lock.mu.Lock()
	if !lock.held || lock.file == nil {
		lock.mu.Unlock()
		return nil
	}
	file := lock.file
	lock.file = nil
	lock.held = false
	onRelease := lock.onRelease
	lock.onRelease = nil
	err := errors.Join(unlockFile(file), file.Close())
	lock.mu.Unlock()
	if onRelease != nil {
		onRelease()
	}
	return err
}
