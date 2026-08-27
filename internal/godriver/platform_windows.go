//go:build windows

package godriver

import (
	"io/fs"
	"os"
)

const platformGoName = "go.exe"

func executableMode(mode fs.FileMode) bool { return mode.IsRegular() }

func nativeExecutableHeader(header []byte) bool {
	return len(header) >= 2 && header[0] == 'M' && header[1] == 'Z'
}

func indispensableEnvironment() map[string]string {
	result := make(map[string]string, 2)
	for _, key := range []string{"SYSTEMROOT", "WINDIR"} {
		if value, present := os.LookupEnv(key); present && value != "" {
			result[key] = value
		}
	}
	return result
}
