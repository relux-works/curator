package install

import (
	"fmt"
	"os"
)

// operationPrivatePrefix names every operation-private root the manager
// creates. One prefix keeps the complete ephemeral footprint of a run
// greppable, both for an operator and for a regression that asserts a scope
// created nothing else.
const operationPrivatePrefix = "curator-install-private-"

// privateRoot is the single operation-private ephemeral root of one
// installation. Every directory a run needs outside persistent state lives
// inside it: the read-only closure workspace of a dry run and the private base
// the trusted toolchain probes or builds in.
//
// One root per operation is the whole point. A dry run must be auditable as
// "created exactly this, removed exactly this", and a real run must be able to
// drop everything it staged with one removal. The root is allocated on first
// use, so an operation that needs nothing ephemeral — a closure with no build
// command, a manifest that never clones — creates nothing at all.
type privateRoot struct {
	prefix  string
	path    string
	created bool
}

// dir allocates a uniquely named directory inside the operation-private root,
// creating the root itself on first use.
func (root *privateRoot) dir(prefix string) (string, error) {
	if !root.created {
		path, err := os.MkdirTemp("", root.prefix)
		if err != nil {
			return "", fmt.Errorf("create the operation-private root: %w", err)
		}
		root.path, root.created = path, true
	}
	directory, err := os.MkdirTemp(root.path, prefix)
	if err != nil {
		return "", fmt.Errorf("create the operation-private %s directory: %w", prefix, err)
	}
	return directory, nil
}

// remove drops the whole operation-private root. It is idempotent and reports
// only a path it could not remove; it takes no verdict on anything else.
func (root *privateRoot) remove() error {
	if !root.created {
		return nil
	}
	path := root.path
	root.path, root.created = "", false
	return os.RemoveAll(path)
}

// releasePrivateRoot removes the operation-private root of one scope after
// every phase below it has run. A leftover private path is an operator
// warning, never a verdict on live state that has already changed.
func releasePrivateRoot(result *Result, scope string, root *privateRoot) {
	if err := root.remove(); err != nil {
		result.Messages = append(result.Messages,
			fmt.Sprintf("%s: could not remove the operation-private root: %v", scope, err))
	}
}
