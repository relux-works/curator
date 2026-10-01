//go:build windows

package buildrepo

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

func runHTTPSBrokerFetchChild() int {
	reportPath := os.Getenv(testBrokerReportPath)
	presence := "absent"
	if _, ok := os.LookupEnv(legacyHTTPSBrokerSecretEnv); ok {
		presence = "present"
	}
	if reportPath == "" || os.WriteFile(reportPath+".fetch", []byte(presence), 0o600) != nil {
		return 1
	}
	childEnv := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, testBrokerFetchChild+"=") {
			continue
		}
		childEnv = append(childEnv, entry)
	}
	command := exec.Command(os.Getenv("GIT_ASKPASS"), "Password for 'https://oauth2@fixture.test': ")
	command.Env = childEnv
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return 1
	}
	return 0
}
