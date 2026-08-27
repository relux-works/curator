//go:build windows

package transaction

func durableRenameNoReplace(from, to string) error {
	return durableRename(from, to)
}
