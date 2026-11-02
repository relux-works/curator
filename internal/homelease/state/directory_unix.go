//go:build !windows

package state

import (
	"github.com/relux-works/curator/internal/pathboundary"
	"os"
)

func openDirectory(path string) (*os.File, error) { return pathboundary.OpenReadNoFollow(path) }
