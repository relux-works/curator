//go:build !windows

package snapshot

func isDestinationSharingViolation(error) bool {
	return false
}
