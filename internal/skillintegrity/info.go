//go:build !windows

package skillintegrity

import (
	"io/fs"
	"os"
)

func captureEntryInfo(_ string, entry fs.DirEntry) (os.FileInfo, error) {
	return entry.Info()
}
