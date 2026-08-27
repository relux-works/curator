//go:build unix

package transaction

func durableReplaceFile(from, to string) error {
	return durableRename(from, to)
}
