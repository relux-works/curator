package buildsource

import "testing"

// frozenRootCase is one validated-root fixture together with the replacement
// that takes its name away from the directory instance Validate hashed. The
// replacement is platform specific; the property it feeds is not.
type frozenRootCase struct {
	// root is the path handed to Validate.
	root string
	// replace leaves a different directory instance, carrying byte-identical
	// content, reachable at root.
	replace func(t *testing.T)
}
