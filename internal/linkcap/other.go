//go:build !unix && !windows

package linkcap

func Unavailable(error) bool { return false }
