package godriver

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Zombies cannot act; orphan reaping timing is outside the teardown contract.
func processAlive(pid int) bool {
	raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err == nil && strings.Contains(string(raw), ") Z ") {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
