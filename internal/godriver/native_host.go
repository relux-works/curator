//go:build !curator_test_platform

package godriver

import "runtime"

func nativeControlHost() string { return runtime.GOOS }
