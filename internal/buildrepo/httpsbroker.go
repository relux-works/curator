package buildrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

const (
	// HTTPSBrokerName is the basename of the manager-owned Git askpass copy.
	// The manager binary dispatches this basename before its public CLI.
	HTTPSBrokerName = "curator-build-https-askpass"
	// EnvHTTPSBrokerState names the manager-owned, secret-free state file.
	EnvHTTPSBrokerState = "CURATOR_BUILD_HTTPS_ASKPASS_STATE"
	// EnvHTTPSBrokerSecret carries the resolved secret only to the fetch
	// process tree. It is never written to broker state.
	EnvHTTPSBrokerSecret = "CURATOR_BUILD_HTTPS_ASKPASS_SECRET" // #nosec G101 -- environment variable name, not a credential.
)

type httpsCredentialState struct {
	Host     string `json:"host"`
	Username string `json:"username"`
}

// HTTPSCredentials is one resolved credential bound to one HTTPS host. The
// secret is deliberately private and all diagnostic formatting redacts it.
type HTTPSCredentials struct {
	Host     string
	Username string
	secret   string
}

// NewHTTPSCredentials constructs the material passed to exactly one fetch.
func NewHTTPSCredentials(host, username, secret string) HTTPSCredentials {
	return HTTPSCredentials{Host: host, Username: username, secret: secret}
}

// Selected reports whether this fetch is authenticated.
func (c HTTPSCredentials) Selected() bool {
	return c.Host != "" && c.Username != "" && c.secret != ""
}

func (c HTTPSCredentials) String() string {
	return fmt.Sprintf("HTTPSCredentials{Host:%q, Username:%q, Secret:<redacted>}", c.Host, c.Username)
}

// GoString returns the same redacted diagnostic representation as String.
func (c HTTPSCredentials) GoString() string { return c.String() }

// IsHTTPSBrokerInvocation reports whether argv0 names the private broker copy.
func IsHTTPSBrokerInvocation(argv0 string) bool {
	name := filepath.Base(argv0)
	if runtime.GOOS == "windows" {
		name = strings.TrimSuffix(strings.ToLower(name), ".exe")
	}
	return name == HTTPSBrokerName
}

// RunHTTPSCredentialBroker answers only Git's two exact HTTPS credential
// prompts. Every malformed invocation or unavailable input fails silently.
func RunHTTPSCredentialBroker(args []string, getenv func(string) string, out io.Writer) int {
	if len(args) != 1 || getenv == nil || out == nil {
		return 1
	}
	statePath, secret := getenv(EnvHTTPSBrokerState), getenv(EnvHTTPSBrokerSecret)
	if statePath == "" || secret == "" {
		return 1
	}
	state, ok := readHTTPSCredentialState(statePath)
	if !ok {
		return 1
	}
	var answer string
	switch args[0] {
	case "Username for 'https://" + state.Host + "': ":
		answer = state.Username
	case "Password for 'https://" + state.Username + "@" + state.Host + "': ":
		answer = secret
	default:
		return 1
	}
	if _, err := io.WriteString(out, answer+"\n"); err != nil {
		return 1
	}
	return 0
}

func readHTTPSCredentialState(path string) (httpsCredentialState, bool) {
	if !filepath.IsAbs(path) {
		return httpsCredentialState{}, false
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 4096 {
		return httpsCredentialState{}, false
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- absolute manager-owned path supplied only to the fetch child.
	if err != nil {
		return httpsCredentialState{}, false
	}
	var state httpsCredentialState
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil || !validHTTPSBrokerField(state.Host) || !validHTTPSBrokerField(state.Username) {
		return httpsCredentialState{}, false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return httpsCredentialState{}, false
	}
	return state, true
}

func validHTTPSBrokerField(value string) bool {
	if value == "" || len(value) > 1024 {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.IsSpace(r) || r == '\'' {
			return false
		}
	}
	return true
}

func materializeHTTPSCredentialBroker(root, executable string, credentials HTTPSCredentials) (string, string, error) {
	if !credentials.Selected() {
		return "", "", nil
	}
	if !filepath.IsAbs(executable) {
		return "", "", admissionError(CodeIdentityInvalid, "credential broker source is not absolute")
	}
	managerRoot := filepath.Join(root, "manager-wrappers")
	if err := os.Mkdir(managerRoot, 0o700); err != nil {
		return "", "", err
	}
	name := HTTPSBrokerName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	wrapper := filepath.Join(managerRoot, name)
	if err := copyBrokerExecutable(executable, wrapper); err != nil {
		return "", "", err
	}
	statePath := filepath.Join(managerRoot, "https-askpass-state.json")
	payload, err := json.Marshal(httpsCredentialState{Host: credentials.Host, Username: credentials.Username})
	if err != nil {
		return "", "", err
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(statePath, payload, 0o600); err != nil {
		return "", "", err
	}
	return wrapper, statePath, nil
}

func copyBrokerExecutable(source, destination string) (err error) {
	if err := os.Link(source, destination); err == nil {
		return nil
	}
	input, err := os.Open(source) // #nosec G304 -- admitted absolute manager executable.
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, input.Close()) }()
	info, err := input.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("credential broker source is not a regular file")
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700) // #nosec G302,G304 -- manager-derived private path; executable wrapper requires its execute bit.
	if err != nil {
		return err
	}
	if _, copyErr := io.Copy(output, input); copyErr != nil {
		_ = output.Close()
		return copyErr
	}
	return output.Close()
}
