//go:build !windows

package gitops

func checkCloneDestination(string) error { return nil }
