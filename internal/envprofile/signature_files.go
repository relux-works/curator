package envprofile

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/privatedir"
)

// Windows verifiers accept regular pathname inputs. Public verification bytes
// live only in an atomically private directory/files and are removed on every
// return, including failed process execution. Managed state and keyrings remain
// read-only; creating/removing this staging directory changes its temp parent.
func runSignatureFiles(command *exec.Cmd, payload []byte, inputs ...[]byte) (output []byte, result error) {
	if len(command.Args) == 0 {
		return nil, fmt.Errorf("%w: empty verifier command", errSignerVerificationUnavailable)
	}
	args, err := signaturePipeArgs(command.Args[1:], len(inputs))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errSignerVerificationUnavailable, err)
	}
	command.Env = gitops.ReadOnlyEnv()
	command.Stdin = bytes.NewReader(payload)
	if len(inputs) == 0 {
		return signatureCommandOutput(command)
	}
	root, err := signaturePrivateInputDirectory()
	if err != nil {
		return nil, fmt.Errorf("%w: private signature staging: %v", errSignerVerificationUnavailable, err)
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil {
			result = fmt.Errorf("%w: remove signature staging: %v (verifier: %v)", errSignerVerificationUnavailable, err, result)
		}
	}()
	if err := privatedir.Validate(root); err != nil {
		return nil, fmt.Errorf("%w: private signature directory: %v", errSignerVerificationUnavailable, err)
	}
	flavour := "native"
	if runtime.GOOS == "windows" {
		flavour, err = signatureExecutableFlavour(command.Path)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errSignerVerificationUnavailable, err)
		}
	}
	for index, input := range inputs {
		path := filepath.Join(root, fmt.Sprintf("input-%d", index))
		file, err := privatedir.Create(path, 0600)
		if err != nil {
			return nil, fmt.Errorf("%w: private signature file: %v", errSignerVerificationUnavailable, err)
		}
		_, writeErr := file.Write(input)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			return nil, fmt.Errorf("%w: write signature input: %v", errSignerVerificationUnavailable, errors.Join(writeErr, closeErr))
		}
		if err := privatedir.ValidateFile(path); err != nil {
			return nil, fmt.Errorf("%w: private signature file boundary: %v", errSignerVerificationUnavailable, err)
		}
		converted := path
		if runtime.GOOS == "windows" {
			converted, err = signatureWindowsPath(flavour, path)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", errSignerVerificationUnavailable, err)
			}
		}
		for i, arg := range args {
			if arg == fmt.Sprintf("/dev/fd/%d", index+3) {
				args[i] = converted
			}
		}
	}
	command.Args = append(command.Args[:1], args...)
	return signatureCommandOutput(command)
}

func signaturePrivateInputDirectory() (string, error) {
	for attempt := 0; attempt < 10; attempt++ {
		var random [32]byte
		if _, err := rand.Read(random[:]); err != nil {
			return "", err
		}
		path := filepath.Join(os.TempDir(), ".curator-signature-"+hex.EncodeToString(random[:]))
		err := privatedir.Make(path)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		return path, nil
	}
	return "", errors.New("signature staging name collision")
}

func signaturePipeArgs(args []string, inputCount int) ([]string, error) {
	result := append([]string(nil), args...)
	for inputIndex := 0; inputIndex < inputCount; inputIndex++ {
		placeholder := fmt.Sprintf("/dev/fd/%d", inputIndex+3)
		found := false
		for _, arg := range result {
			if arg == placeholder {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("signature input %d has no verifier argument", inputIndex)
		}
	}
	for _, arg := range result {
		if strings.HasPrefix(arg, "/dev/fd/") {
			index, err := strconv.Atoi(strings.TrimPrefix(arg, "/dev/fd/"))
			if err != nil || fmt.Sprintf("/dev/fd/%d", index) != arg || index < 3 || index-3 >= inputCount {
				return nil, errors.New("verifier requested an unavailable signature input")
			}
		}
	}
	return result, nil
}
