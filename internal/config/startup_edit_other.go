//go:build !unix

package config

import "os"

func startupEditInfo(os.FileInfo) bool { return false }
