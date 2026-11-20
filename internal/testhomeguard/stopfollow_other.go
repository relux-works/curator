//go:build !linux && !darwin

package testhomeguard

func childStopped(int) bool { return false }
