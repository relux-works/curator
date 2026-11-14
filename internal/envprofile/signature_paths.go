package envprofile

import (
	"debug/pe"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Windows packages may select a native, MSYS, or Cygwin GnuPG executable.
// Inspect its import table rather than treating executable presence as proof
// of a compatible path convention. No verifier/configuration is run here.
func signatureVerifierPath(executable, path string) (string, error) {
	if runtime.GOOS != "windows" {
		return path, nil
	}
	flavour, err := signatureExecutableFlavour(executable)
	if err != nil {
		return "", err
	}
	return signatureWindowsPath(flavour, path)
}

func signatureExecutableFlavour(executable string) (string, error) {
	resolved, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}
	file, err := pe.Open(resolved)
	if err != nil {
		return "", fmt.Errorf("gpg-path-flavour-unknown: %w", err)
	}
	// ImportedLibraries is a stub in the supported Go toolchain. ImportedSymbols
	// reads the PE import table and returns symbol:DLL pairs.
	symbols, readErr := signaturePEImports(file)
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return "", fmt.Errorf("gpg-path-flavour-unknown: import table unreadable")
	}
	var libraries []string
	for _, symbol := range symbols {
		if index := strings.LastIndexByte(symbol, ':'); index >= 0 {
			libraries = append(libraries, symbol[index+1:])
		}
	}
	return signatureRuntimeFlavour(libraries), nil
}

// The supported PE parser uses unchecked import RVAs internally. A corrupt
// verifier image must remain a named unavailable result, never panic a read.
func signaturePEImports(file *pe.File) (symbols []string, err error) {
	defer func() {
		if recover() != nil {
			symbols = nil
			err = errors.New("malformed PE import table")
		}
	}()
	return file.ImportedSymbols()
}

func signatureRuntimeFlavour(libraries []string) string {
	native := false
	for _, library := range libraries {
		switch strings.ToLower(library) {
		case "msys-2.0.dll":
			return "msys"
		case "cygwin1.dll":
			return "cygwin"
		case "kernel32.dll", "msvcrt.dll", "ucrtbase.dll":
			native = true
		}
	}
	if native {
		return "native"
	}
	return "unknown"
}

func signatureWindowsPath(flavour, path string) (string, error) {
	if flavour == "native" {
		return path, nil
	}
	if flavour != "msys" && flavour != "cygwin" {
		return "", fmt.Errorf("gpg-path-flavour-unknown: executable runtime is not recognized")
	}
	path = strings.ReplaceAll(path, `\`, "/")
	if len(path) < 3 || path[1] != ':' || path[2] != '/' || (path[0] < 'A' || path[0] > 'Z') && (path[0] < 'a' || path[0] > 'z') {
		return "", fmt.Errorf("gpg-path-unrepresentable: POSIX runtime requires an absolute drive path")
	}
	prefix := "/"
	if flavour == "cygwin" {
		prefix = "/cygdrive/"
	}
	return prefix + strings.ToLower(path[:1]) + path[2:], nil
}
