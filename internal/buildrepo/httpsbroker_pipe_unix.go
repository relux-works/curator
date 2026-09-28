//go:build !windows

package buildrepo

import (
	"os"
	"os/exec"
	"strconv"
)

func inheritHTTPSBrokerReader(command *exec.Cmd, reader *os.File) (string, error) {
	fd := 3 + len(command.ExtraFiles)
	command.ExtraFiles = append(command.ExtraFiles, reader)
	return strconv.Itoa(fd), nil
}
