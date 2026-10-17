// Package nativeseed reads operator-owned seed inputs without writing them.
// It is separate from the no-symlink reader for manager-owned state.
package nativeseed

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/stateread"
)

// ReadFile resolves the complete symlink chain once, including relative links.
// Dangling links and cycles are unreadable, while an absent input stays absent.
// The resolved leaf is opened without following a replacement final symlink.
// Type, ownership and mutation permissions are checked on that open file before
// any bytes are read from the same descriptor. No native path is ever written.
func ReadFile(path string) (stateread.File, error) {
	return readWithOpen(path, openSeed)
}

func readWithOpen(path string, open func(string) (*os.File, error)) (stateread.File, error) {
	unreadable := func(err error) (stateread.File, error) {
		return stateread.File{Kind: stateread.KindUnreadable}, stateread.UnusableError(path, err)
	}
	state, err := stateread.Lstat(path)
	if err != nil {
		return stateread.File{Kind: stateread.KindUnreadable}, err
	}
	if state.Kind == stateread.KindAbsent {
		return stateread.File{Kind: stateread.KindAbsent}, nil
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return unreadable(fmt.Errorf("resolve native seed: %w", err))
	}
	// Avoid opening known special files; the descriptor check below also covers
	// replacement after this inspection. Unix open is nonblocking for raced FIFOs.
	info, err := os.Lstat(resolved)
	if err != nil {
		return unreadable(err)
	}
	if !info.Mode().IsRegular() {
		return unreadable(fmt.Errorf("native seed target must be a regular file"))
	}
	file, err := open(resolved)
	if err != nil {
		return unreadable(err)
	}
	defer func() { _ = file.Close() }()
	info, err = file.Stat()
	if err != nil {
		return unreadable(err)
	}
	if !info.Mode().IsRegular() {
		return unreadable(fmt.Errorf("native seed target must be a regular file"))
	}
	if err := checkSeedFile(file, info); err != nil {
		return unreadable(err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return unreadable(err)
	}
	return stateread.File{Kind: stateread.KindPresent, Bytes: data}, nil
}
