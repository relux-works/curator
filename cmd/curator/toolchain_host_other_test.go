//go:build !windows

package main

// platformPathFact adds nothing outside Windows: os.Lstat already classifies a
// directory as a directory on every other supported host.
func platformPathFact(string) string { return "" }
