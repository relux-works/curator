// Package nofollow reads and mutates manager paths through pinned directory
// handles. Each parent is inspected without following links, then checked
// against the opened directory identity before it is used.
package nofollow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/stateread"
)

// OpenParent opens the parent of rel below anchor. The anchor's own parent
// route may contain platform aliases (such as /var on macOS); the anchor and
// every component below it must be plain directories. The caller closes it.
func OpenParent(anchor, rel string) (*os.Root, string, error) {
	if !filepath.IsLocal(rel) || rel == "." {
		return nil, "", fmt.Errorf("invalid relative manager path %q", rel)
	}
	state, err := stateread.Lstat(anchor)
	if err != nil {
		return nil, "", err
	}
	if state.Kind == stateread.KindAbsent {
		return nil, "", stateread.AbsentError(anchor)
	}
	info := state.Info
	if !info.IsDir() || redirect(info) {
		return nil, "", fmt.Errorf("manager anchor %s is not a plain directory", anchor)
	}
	root, err := os.OpenRoot(anchor)
	if err != nil {
		return nil, "", err
	}
	opened, err := root.Stat(".")
	if err != nil || !os.SameFile(info, opened) {
		_ = root.Close()
		return nil, "", fmt.Errorf("manager anchor %s changed while opening: %v", anchor, err)
	}
	parts := strings.Split(filepath.Clean(rel), string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		info, err := root.Lstat(part)
		if err != nil {
			_ = root.Close()
			return nil, "", err
		}
		if !info.IsDir() || redirect(info) {
			_ = root.Close()
			return nil, "", fmt.Errorf("manager parent %s is not a plain directory", part)
		}
		child, err := root.OpenRoot(part)
		if err != nil {
			_ = root.Close()
			return nil, "", err
		}
		opened, err := child.Stat(".")
		_ = root.Close()
		if err != nil || !os.SameFile(info, opened) {
			_ = child.Close()
			return nil, "", fmt.Errorf("manager parent %s changed while opening: %v", part, err)
		}
		root = child
	}
	return root, parts[len(parts)-1], nil
}

// ReadRegularFile proves the parent route and the leaf identity. Missing
// directories or leaves are absence; a linked or blocked route is unreadable.
func ReadRegularFile(anchor, rel string) (stateread.File, error) {
	root, leaf, parentErr := OpenParent(anchor, rel)
	if parentErr != nil {
		if os.IsNotExist(parentErr) {
			return stateread.File{Kind: stateread.KindAbsent}, nil
		}
		return stateread.File{Kind: stateread.KindUnreadable}, parentErr
	}
	defer func() { _ = root.Close() }()
	state, err := stateread.LstatWith(filepath.Join(anchor, rel), func(string) (os.FileInfo, error) { return root.Lstat(leaf) })
	if err != nil {
		return stateread.File{Kind: stateread.KindUnreadable}, err
	}
	if state.Kind == stateread.KindAbsent {
		return stateread.File{Kind: stateread.KindAbsent}, nil
	}
	info := state.Info
	if !info.Mode().IsRegular() || redirect(info) {
		return stateread.File{Kind: stateread.KindUnreadable}, fmt.Errorf("manager record %s must be a regular file", rel)
	}
	file, err := root.Open(leaf)
	if err != nil {
		return stateread.File{Kind: stateread.KindUnreadable}, err
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return stateread.File{Kind: stateread.KindUnreadable}, fmt.Errorf("manager record %s changed while opening: %v", rel, err)
	}
	payload, err := io.ReadAll(file)
	if err != nil {
		return stateread.File{Kind: stateread.KindUnreadable}, err
	}
	return stateread.File{Kind: stateread.KindPresent, Bytes: payload}, nil
}
