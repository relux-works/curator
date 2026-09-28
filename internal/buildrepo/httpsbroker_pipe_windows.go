//go:build windows

package buildrepo

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func inheritHTTPSBrokerReader(command *exec.Cmd, reader *os.File) (string, error) {
	handle := syscall.Handle(reader.Fd())
	if handle == 0 {
		return "", fmt.Errorf("HTTPS broker pipe has an invalid handle")
	}
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.AdditionalInheritedHandles = append(command.SysProcAttr.AdditionalInheritedHandles, handle)
	return strconv.FormatUint(uint64(handle), 10), nil
}
