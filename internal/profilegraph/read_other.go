//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package profilegraph

import "os"

// Refuse on hosts without the descriptor-relative no-follow implementation.
// No weaker filesystem admission is substituted.
func (r *Reader) open(root, rel string, directory bool, links ...bool) (*os.File, error) {
	return nil, refusal("graph_invalid_input")
}
