//go:build linux

package testhomeguard

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func childStopped(pid int) bool {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	closeParen := bytes.LastIndexByte(raw, ')')
	if closeParen < 0 {
		return false
	}
	fields := strings.Fields(string(raw[closeParen+1:]))
	return len(fields) > 0 && (fields[0] == "T" || fields[0] == "t")
}
