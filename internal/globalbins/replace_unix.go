//go:build !windows

package globalbins

import "os"

func replaceLedgerFile(from, to string) error {
	return os.Rename(from, to)
}
