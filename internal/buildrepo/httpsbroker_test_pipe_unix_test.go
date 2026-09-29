//go:build !windows

package buildrepo

import (
	"io"
	"os"
	"os/exec"
	"strconv"
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
	handle, ok := parseHTTPSBrokerHandle(os.Getenv(EnvHTTPSBrokerTransport))
	if !ok {
		return 1
	}
	reader := os.NewFile(handle, "https-askpass-test-reader")
	if reader == nil {
		return 1
	}
	command := exec.Command(os.Getenv("GIT_ASKPASS"), "Password for 'https://oauth2@fixture.test': ")
	childHandle := strconv.Itoa(3 + len(command.ExtraFiles))
	command.ExtraFiles = append(command.ExtraFiles, reader)
	childEnv := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, testBrokerFetchChild+"=") {
			continue
		}
		childEnv = append(childEnv, entry)
	}
	command.Env = setEnvironmentValue(childEnv, EnvHTTPSBrokerTransport, childHandle)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	err := command.Run()
	_ = reader.Close()
	if err != nil {
		return 1
	}
	return 0
}

func testHTTPSBrokerTransportValue(transport HTTPSBrokerSecretTransport) string {
	return strconv.FormatUint(uint64(transport.(*unixHTTPSBrokerSecretTransport).reader.Fd()), 10)
}
