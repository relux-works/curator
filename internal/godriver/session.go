// Package godriver implements the closed go-v1 build driver under the portable
// manager-worker-v1 execution policy of Protocol Core section 4.2.1. The
// package-independent toolchain probes run directly from the manager parent;
// every source-aware command runs inside one identity-verified, hidden-mode
// re-execution of the installed manager. The built artifact is verified and
// never started.
package godriver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildmeta"
)

const (
	defaultProbeTimeout       = 15 * time.Second
	defaultFingerprintTimeout = 2 * time.Minute
	defaultOutputLimit        = int64(64 * 1024)
	maxVersionOutput          = 4096
)

var (
	goVersionPattern = regexp.MustCompile(`^go version go(1\.([0-9]+)(?:\.[0-9]+)?(?:rc[0-9]+|beta[0-9]+)?) ([a-z0-9]+)/([a-z0-9]+)$`)
	allowedFamilies  = map[string]struct{}{"1.25": {}}
	probeEnvNames    = []string{
		"GOROOT", "GOHOSTOS", "GOHOSTARCH", "GOOS", "GOARCH", "GO386", "GOAMD64", "GOARM", "GOARM64",
		"GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM", "GOTELEMETRY", "GOTELEMETRYDIR",
	}
)

// Selection names the operator mechanisms that choose the trusted Go
// installation, in the order the driver consults them. They are the complete
// accepted set: the driver never searches PATH, never reads a package or
// manifest value, and never downloads a toolchain.
const (
	// SelectionCuratorGo is the explicit launcher variable. It must name an
	// absolute <GOROOT>/bin/go (bin/go.exe on Windows).
	SelectionCuratorGo = "CURATOR_GO"
	// SelectionGOROOT selects the trusted root; the launcher is derived from it.
	SelectionGOROOT = "GOROOT"
)

// toolchainSelectionRemedy is the operator remedy every
// toolchain_executable_mismatch carries. It answers exactly one host
// condition, which is the only one that produces the code in practice: a
// version-manager launcher — a goenv, asdf, or mise shim, or any wrapper that
// resolves outside a real GOROOT/bin — was selected, either because it is what
// the selection variables name or because it answers `go env` with a root
// other than the one selected. The verdict alone does not tell an operator
// that, and the fix is one host fact: the real GOROOT/bin ahead of the wrapper,
// which is where the trusted selection variables are then read from.
//
// It is advice for the operator's own shell. The driver itself still never
// searches PATH and never downloads a toolchain, and the guidance the CLI adds
// from the code says so.
//
// The text is short on purpose: an install boundary bounds a rendered
// diagnostic, and a remedy that pushed the line past that bound would be
// truncated away exactly where an operator reads it.
const toolchainSelectionRemedy = `put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`

// TestedFamilies lists the Go release families this manager has tested against
// the go-v1 conformance vectors, in ascending order. A release outside the list
// is rejected rather than downloaded or approximated, so operator-facing
// diagnostics name it instead of guessing at a remedy.
func TestedFamilies() []string {
	families := make([]string, 0, len(allowedFamilies))
	for family := range allowedFamilies {
		families = append(families, family)
	}
	sort.Strings(families)
	return families
}

// Config contains only manager/operator inputs. CuratorGo and GOROOT are
// explicit so the driver never searches PATH or consumes package environment.
type Config struct {
	PrivateBase        string
	CuratorGo          string
	GOROOT             string
	ForbiddenRoots     []string
	Executor           Executor
	ProbeTimeout       time.Duration
	OutputLimit        int64
	FingerprintTimeout time.Duration
}

// ConfigFromEnvironment reads only the two trusted-selection variables. No
// user PATH, Go flags, workspace, proxy, compiler, authentication, or locale
// state is admitted into Config.
func ConfigFromEnvironment(privateBase string, forbiddenRoots ...string) Config {
	return Config{
		PrivateBase:    privateBase,
		CuratorGo:      os.Getenv(SelectionCuratorGo),
		GOROOT:         os.Getenv(SelectionGOROOT),
		ForbiddenRoots: append([]string(nil), forbiddenRoots...),
	}
}

// Session owns one operation-private telemetry/config/cache root and the
// frozen native target plus portable toolchain identity derived from it.
type Session struct {
	executable         string
	goroot             string
	operation          string
	privateBase        string
	environment        []string
	target             buildmeta.Target
	toolchain          buildmeta.Toolchain
	rootInfo           fs.FileInfo
	state              []toolchainState
	fingerprintTimeout time.Duration

	closeOnce sync.Once
	closeErr  error
}

// Snapshot is the package-independent value needed by dry-run planning. Its
// private paths are descriptive only and have already been removed.
type Snapshot struct {
	Executable  string
	GOROOT      string
	Target      buildmeta.Target
	Toolchain   buildmeta.Toolchain
	Environment []string
}

type hostFacts struct {
	runtimeGOROOT string
	goos          string
	goarch        string
}

// Establish selects, probes, and fingerprints one trusted Go installation.
// The caller must Close the returned session after the last child exits.
func Establish(ctx context.Context, config Config) (*Session, error) {
	return establish(ctx, config, hostFacts{runtimeGOROOT: build.Default.GOROOT, goos: runtime.GOOS, goarch: runtime.GOARCH})
}

// Probe performs the same three commands and fingerprinting as Establish but
// removes all private state before returning. It does not memoize results.
func Probe(ctx context.Context, config Config) (Snapshot, error) {
	return probe(ctx, config, hostFacts{runtimeGOROOT: build.Default.GOROOT, goos: runtime.GOOS, goarch: runtime.GOARCH})
}

func probe(ctx context.Context, config Config, host hostFacts) (Snapshot, error) {
	session, err := establish(ctx, config, host)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := session.Snapshot()
	if err := session.Close(); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func establish(ctx context.Context, config Config, host hostFacts) (_ *Session, resultErr error) {
	if config.Executor == nil {
		config.Executor = OSExecutor{}
	}
	if config.ProbeTimeout <= 0 || config.ProbeTimeout > defaultProbeTimeout {
		config.ProbeTimeout = defaultProbeTimeout
	}
	if config.OutputLimit <= 0 || config.OutputLimit > defaultOutputLimit {
		config.OutputLimit = defaultOutputLimit
	}
	if config.FingerprintTimeout <= 0 || config.FingerprintTimeout > defaultFingerprintTimeout {
		config.FingerprintTimeout = defaultFingerprintTimeout
	}

	executable, goroot, rootInfo, err := selectToolchain(config, host.runtimeGOROOT)
	if err != nil {
		return nil, err
	}
	base, err := validatePrivateBase(config.PrivateBase, config.ForbiddenRoots)
	if err != nil {
		return nil, err
	}
	operation, err := os.MkdirTemp(base, ".curator-go-probe-")
	if err != nil {
		return nil, diagnosticErr("private_probe_failed", err, "cannot create operation-private probe")
	}
	defer func() {
		if resultErr != nil {
			if cleanupErr := os.RemoveAll(operation); cleanupErr != nil {
				resultErr = errors.Join(resultErr, diagnosticErr("private_probe_cleanup_failed", cleanupErr, "cannot remove failed probe"))
			}
		}
	}()

	layout, err := createProbeLayout(operation, host.goos)
	if err != nil {
		return nil, err
	}
	bootstrap := bootstrapEnvironment(layout, host.goos)
	run := func(arguments ...string) (Output, error) {
		if err := verifySelectedRoot(goroot, rootInfo, executable); err != nil {
			return Output{}, err
		}
		request := Process{
			Executable:  executable,
			Arguments:   append([]string(nil), arguments...),
			Directory:   layout.empty,
			Environment: append([]string(nil), bootstrap...),
			Timeout:     config.ProbeTimeout,
			OutputLimit: config.OutputLimit,
		}
		output, runErr := config.Executor.Run(ctx, request)
		if int64(len(output.Stdout)) > config.OutputLimit || int64(len(output.Stderr)) > config.OutputLimit {
			return output, diagnostic("process_output_limit", "Go probe exceeded its output bound")
		}
		if errors.Is(runErr, errProcessTimeout) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return output, diagnosticErr("process_timeout", runErr, "Go probe exceeded its deadline")
		}
		if errors.Is(runErr, errOutputLimit) {
			return output, diagnosticErr("process_output_limit", runErr, "Go probe exceeded its output bound")
		}
		return output, runErr
	}

	if _, err := run("telemetry", "off"); err != nil {
		if DiagnosticCode(err) != "" {
			return nil, err
		}
		return nil, diagnosticErr("telemetry_initialization_failed", err, "go telemetry off failed")
	}
	versionOutput, err := run("version")
	if err != nil {
		if DiagnosticCode(err) != "" {
			return nil, err
		}
		return nil, diagnosticErr("go_version_failed", err, "go version failed")
	}
	version, family, versionOS, versionArch, err := parseGoVersion(versionOutput.Stdout)
	if err != nil {
		return nil, err
	}
	if _, allowed := allowedFamilies[family]; !allowed {
		return nil, diagnostic("unsupported_go_family", "Go release family %s is not allowlisted", family)
	}

	envArguments := append([]string{"env", "-json"}, probeEnvNames...)
	envOutput, err := run(envArguments...)
	if err != nil {
		if DiagnosticCode(err) != "" {
			return nil, err
		}
		return nil, diagnosticErr("go_env_failed", err, "fixed go env probe failed")
	}
	probed, err := decodeProbeEnvironment(envOutput.Stdout)
	if err != nil {
		return nil, err
	}
	if err := validateProbe(probed, goroot, versionOS, versionArch, host, layout.config); err != nil {
		return nil, err
	}
	if entries, readErr := os.ReadDir(layout.empty); readErr != nil || len(entries) != 0 {
		return nil, diagnosticErr("process_environment_poisoned", readErr, "Go probe modified its empty working directory")
	}

	target, err := targetFromProbe(probed)
	if err != nil {
		return nil, err
	}
	fingerprintCtx, cancelFingerprint := context.WithTimeout(ctx, config.FingerprintTimeout)
	digest, state, err := fingerprintToolchain(fingerprintCtx, goroot, version)
	cancelFingerprint()
	if err != nil {
		return nil, err
	}
	if err := verifySelectedRoot(goroot, rootInfo, executable); err != nil {
		return nil, err
	}
	environment := buildEnvironment(bootstrap, goroot, target)
	return &Session{
		executable: executable, goroot: goroot, operation: operation, privateBase: base,
		environment: environment, target: target,
		toolchain: buildmeta.Toolchain{
			Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
			GoVersion: version, ContentSHA256: digest,
		},
		rootInfo: rootInfo, state: state, fingerprintTimeout: config.FingerprintTimeout,
	}, nil
}

// Snapshot returns defensive copies of all frozen session values.
func (session *Session) Snapshot() Snapshot {
	return Snapshot{
		Executable: session.executable, GOROOT: session.goroot,
		Target: copyTarget(session.target), Toolchain: session.toolchain,
		Environment: append([]string(nil), session.environment...),
	}
}

// Executable is the fingerprinted Go launcher this session selected.
func (session *Session) Executable() string { return session.executable }

// GOROOT is the canonical root of the fingerprinted toolchain.
func (session *Session) GOROOT() string { return session.goroot }

// Target is the frozen native target and its single tuning input.
func (session *Session) Target() buildmeta.Target { return copyTarget(session.target) }

// Toolchain is the portable curator-go-toolchain-v1 identity.
func (session *Session) Toolchain() buildmeta.Toolchain { return session.toolchain }

// Environment is the closed offline compiler environment.
func (session *Session) Environment() []string { return append([]string(nil), session.environment...) }

// VerifyToolchain revalidates the launcher and exact tree identity. Call it
// after the last process that may consume this session.
func (session *Session) VerifyToolchain(ctx context.Context) error {
	if err := verifySelectedRoot(session.goroot, session.rootInfo, session.executable); err != nil {
		return err
	}
	digest, state, err := fingerprintToolchain(ctx, session.goroot, session.toolchain.GoVersion)
	if err != nil {
		return diagnosticErr("toolchain_mutated", err, "cannot revalidate toolchain")
	}
	if digest != session.toolchain.ContentSHA256 || !reflect.DeepEqual(state, session.state) {
		return diagnostic("toolchain_mutated", "toolchain tree changed during operation")
	}
	return nil
}

// Close revalidates the toolchain through the last child exit, then removes
// telemetry mode and every other operation-private probe path even on drift.
//
// A caller that takes its own trust verdict through VerifyToolchain, before it
// acts on anything this session produced, uses Release instead: teardown that
// runs after that decision must not be able to revise it.
func (session *Session) Close() error {
	if session == nil {
		return nil
	}
	session.closeOnce.Do(func() {
		verifyCtx, cancel := context.WithTimeout(context.Background(), session.fingerprintTimeout)
		verifyErr := session.VerifyToolchain(verifyCtx)
		cancel()
		if cleanupErr := session.removePrivateState(); cleanupErr != nil {
			session.closeErr = errors.Join(verifyErr, cleanupErr)
			return
		}
		session.closeErr = verifyErr
	})
	return session.closeErr
}

// Release removes telemetry mode and every other operation-private probe path
// without revalidating the toolchain. It reports cleanup failures only, so it
// never reports drift for a session whose outputs were already trusted.
func (session *Session) Release() error {
	if session == nil {
		return nil
	}
	session.closeOnce.Do(func() { session.closeErr = session.removePrivateState() })
	return session.closeErr
}

// removePrivateState deletes the operation-private probe root and refuses any
// path that is not strictly below the private base the session was given.
func (session *Session) removePrivateState() error {
	if !strictlyBelow(session.operation, session.privateBase) {
		return diagnostic("private_probe_cleanup_failed", "refusing to remove a probe outside its private base")
	}
	if err := os.RemoveAll(session.operation); err != nil {
		return diagnosticErr("private_probe_cleanup_failed", err, "cannot remove operation-private probe")
	}
	return nil
}

type probeLayout struct {
	root, empty, emptyPath, gopath, gomodcache, gocache, gotmp, home, config, tmp string
	appdata, localappdata, userprofile                                            string
}

func createProbeLayout(root, goos string) (probeLayout, error) {
	layout := probeLayout{
		root: root, empty: filepath.Join(root, "empty"), emptyPath: filepath.Join(root, "empty-path"),
		gopath: filepath.Join(root, "gopath"), gomodcache: filepath.Join(root, "gomodcache"),
		gocache: filepath.Join(root, "gocache"), gotmp: filepath.Join(root, "gotmp"),
		home: filepath.Join(root, "home"), config: filepath.Join(root, "config"), tmp: filepath.Join(root, "tmp"),
		appdata: filepath.Join(root, "appdata"), localappdata: filepath.Join(root, "localappdata"),
		userprofile: filepath.Join(root, "userprofile"),
	}
	switch goos {
	case "darwin":
		layout.config = filepath.Join(layout.home, "Library", "Application Support")
	case "windows":
		layout.config = layout.appdata
	}
	directories := []string{layout.empty, layout.emptyPath, layout.gopath, layout.gomodcache, layout.gocache, layout.gotmp, layout.home, layout.config, layout.tmp}
	if goos == "windows" {
		directories = append(directories, layout.appdata, layout.localappdata, layout.userprofile)
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return probeLayout{}, diagnosticErr("private_probe_failed", err, "cannot create private probe layout")
		}
	}
	return layout, nil
}

func bootstrapEnvironment(layout probeLayout, goos string) []string {
	values := indispensableEnvironment()
	if values == nil {
		values = make(map[string]string)
	}
	for key, value := range map[string]string{
		"GOENV": "off", "GOTOOLCHAIN": "local", "LC_ALL": "C", "LANG": "C",
		"GOPATH": layout.gopath, "GOMODCACHE": layout.gomodcache, "GOCACHE": layout.gocache,
		"GOTMPDIR": layout.gotmp, "HOME": layout.home, "XDG_CONFIG_HOME": layout.config,
		"PATH": layout.emptyPath, "TMPDIR": layout.tmp,
	} {
		values[key] = value
	}
	if goos == "windows" {
		values["APPDATA"] = layout.appdata
		values["LOCALAPPDATA"] = layout.localappdata
		values["USERPROFILE"] = layout.userprofile
		values["TEMP"] = layout.tmp
		values["TMP"] = layout.tmp
	}
	return environmentSlice(values)
}

func buildEnvironment(bootstrap []string, goroot string, target buildmeta.Target) []string {
	values := environmentMap(bootstrap)
	for key, value := range map[string]string{
		"GOROOT": goroot, "GOOS": target.GOOS, "GOARCH": target.GOARCH,
		"GO111MODULE": "on", "GOFLAGS": "", "GOPROXY": "off", "GOSUMDB": "off",
		"GOPRIVATE": "", "GONOPROXY": "none", "GONOSUMDB": "none", "GOVCS": "*:off",
		"GOWORK": "off", "CGO_ENABLED": "0", "GO_EXTLINK_ENABLED": "0", "GOEXPERIMENT": "",
	} {
		values[key] = value
	}
	for key, value := range target.Tuning {
		values[key] = value
	}
	return environmentSlice(values)
}

func environmentMap(environment []string) map[string]string {
	values := make(map[string]string, len(environment))
	for _, item := range environment {
		key, value, _ := strings.Cut(item, "=")
		values[key] = value
	}
	return values
}

func environmentSlice(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	environment := make([]string, 0, len(keys))
	for _, key := range keys {
		environment = append(environment, key+"="+values[key])
	}
	return environment
}

func selectToolchain(config Config, runtimeRoot string) (string, string, fs.FileInfo, error) {
	var candidate, root string
	switch {
	case config.CuratorGo != "":
		candidate = config.CuratorGo
		if !filepath.IsAbs(candidate) || filepath.Base(candidate) != platformGoName || filepath.Base(filepath.Dir(candidate)) != "bin" {
			return "", "", nil, diagnostic("untrusted_go_executable", "CURATOR_GO must name an absolute GOROOT/bin/%s", platformGoName)
		}
		root = filepath.Dir(filepath.Dir(candidate))
	case config.GOROOT != "":
		root = config.GOROOT
		candidate = filepath.Join(root, "bin", platformGoName)
	default:
		root = runtimeRoot
		candidate = filepath.Join(root, "bin", platformGoName)
	}
	if !filepath.IsAbs(root) || !filepath.IsAbs(candidate) {
		return "", "", nil, diagnostic("untrusted_go_executable", "trusted Go root and executable must be absolute")
	}
	// physicalPath, not filepath.EvalSymlinks: the selected root and the
	// launcher under it must be resolved to the objects they really name before
	// either is typed or fingerprinted, and on Windows EvalSymlinks leaves a
	// directory junction unresolved. Both are resolved the same way, so the
	// derived-launcher identity below compares like with like.
	resolvedRoot, err := physicalPath(root)
	if err != nil {
		return "", "", nil, diagnosticErr("go_toolchain_missing", err, "trusted GOROOT is unusable")
	}
	rootInfo, err := os.Lstat(resolvedRoot)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		return "", "", nil, diagnosticErr("go_toolchain_missing", err, "trusted GOROOT is not a real directory")
	}
	resolvedCandidate, err := physicalPath(candidate)
	if err != nil {
		return "", "", nil, diagnosticErr("go_toolchain_missing", err, "trusted Go executable is unusable")
	}
	expected := filepath.Join(resolvedRoot, "bin", platformGoName)
	if resolvedCandidate != expected {
		return "", "", nil, diagnosticRemedy("toolchain_executable_mismatch", toolchainSelectionRemedy,
			"selected Go executable is not the regular executable under the derived GOROOT")
	}
	for _, forbidden := range config.ForbiddenRoots {
		forbiddenAbs, resolveErr := absolutePhysical(forbidden)
		if resolveErr == nil && (resolvedCandidate == forbiddenAbs || isWithin(resolvedCandidate, forbiddenAbs)) {
			return "", "", nil, diagnostic("untrusted_go_executable", "selected Go executable is under a repository or runtime root")
		}
	}
	if err := validateLauncher(expected); err != nil {
		return "", "", nil, err
	}
	return expected, resolvedRoot, rootInfo, nil
}

func validateLauncher(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || !executableMode(info.Mode()) {
		return diagnosticErr("untrusted_go_executable", err, "selected Go launcher is not a regular executable")
	}
	file, err := os.Open(path) // #nosec G304 -- path is independently derived from trusted GOROOT
	if err != nil {
		return diagnosticErr("untrusted_go_executable", err, "cannot open selected Go launcher")
	}
	defer func() { _ = file.Close() }()
	openedInfo, err := file.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return diagnosticErr("untrusted_go_executable", err, "selected Go launcher changed while opening")
	}
	var header [8]byte
	read, readErr := io.ReadFull(file, header[:])
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		return diagnosticErr("untrusted_go_executable", readErr, "cannot inspect selected Go launcher")
	}
	if !nativeExecutableHeader(header[:read]) {
		return diagnostic("untrusted_go_executable", "selected Go launcher is a wrapper rather than a native executable")
	}
	return nil
}

func verifySelectedRoot(goroot string, rootInfo fs.FileInfo, executable string) error {
	currentRoot, err := os.Lstat(goroot)
	if err != nil || !currentRoot.IsDir() || !os.SameFile(rootInfo, currentRoot) {
		return diagnosticErr("toolchain_mutated", err, "fingerprinted GOROOT was replaced")
	}
	if err := validateLauncher(executable); err != nil {
		return diagnosticErr("toolchain_mutated", err, "selected Go executable changed")
	}
	return nil
}

func validatePrivateBase(path string, forbidden []string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", diagnostic("private_probe_failed", "private probe base must be absolute")
	}
	physical, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", diagnosticErr("private_probe_failed", err, "private probe base is unavailable")
	}
	info, err := os.Lstat(physical)
	if err != nil || !info.IsDir() {
		return "", diagnosticErr("private_probe_failed", err, "private probe base is not a directory")
	}
	for _, root := range forbidden {
		forbiddenAbs, resolveErr := absolutePhysical(root)
		if resolveErr == nil && (physical == forbiddenAbs || isWithin(physical, forbiddenAbs)) {
			return "", diagnostic("private_probe_failed", "private probe base is under a repository or runtime root")
		}
	}
	return physical, nil
}

func parseGoVersion(stdout []byte) (normalized, family, goos, goarch string, err error) {
	if len(stdout) == 0 || len(stdout) > maxVersionOutput || stdout[len(stdout)-1] != '\n' || strings.Count(string(stdout), "\n") != 1 || !utf8.Valid(stdout) || strings.ContainsRune(string(stdout), 0) {
		return "", "", "", "", diagnostic("malformed_go_version", "go version output must be one bounded UTF-8 line with a terminal LF")
	}
	payload := stdout[:len(stdout)-1]
	if len(payload) > 0 && payload[len(payload)-1] == '\r' {
		payload = payload[:len(payload)-1]
	}
	if len(payload) == 0 || strings.ContainsAny(string(payload), "\r\n\x00") {
		return "", "", "", "", diagnostic("malformed_go_version", "go version output has invalid line structure")
	}
	matches := goVersionPattern.FindStringSubmatch(string(payload))
	if matches == nil {
		return "", "", "", "", diagnostic("malformed_go_version", "go version output does not identify a release and target")
	}
	if len(matches[2]) > 1 && matches[2][0] == '0' {
		return "", "", "", "", diagnostic("malformed_go_version", "go version release family has a leading zero")
	}
	minor, conversionErr := strconv.Atoi(matches[2])
	if conversionErr != nil {
		return "", "", "", "", diagnosticErr("malformed_go_version", conversionErr, "go version release family is invalid")
	}
	if minor < 23 {
		return "", "", "", "", diagnostic("unsupported_go_family", "Go release is older than 1.23")
	}
	return string(payload), fmt.Sprintf("1.%d", minor), matches[3], matches[4], nil
}

func decodeProbeEnvironment(payload []byte) (map[string]string, error) {
	if len(payload) == 0 || int64(len(payload)) > defaultOutputLimit || !utf8.Valid(payload) {
		return nil, diagnostic("invalid_go_env", "go env output is empty, oversized, or invalid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return nil, diagnosticErr("invalid_go_env", err, "go env output is not a JSON object")
	}
	values := make(map[string]string, len(probeEnvNames))
	for decoder.More() {
		token, tokenErr := decoder.Token()
		if tokenErr != nil {
			return nil, diagnosticErr("invalid_go_env", tokenErr, "go env output has an invalid key")
		}
		key, ok := token.(string)
		if !ok {
			return nil, diagnostic("invalid_go_env", "go env output has a non-string key")
		}
		if _, duplicate := values[key]; duplicate {
			return nil, diagnostic("invalid_go_env", "go env output repeats %s", key)
		}
		var value string
		if err := decoder.Decode(&value); err != nil {
			return nil, diagnosticErr("invalid_go_env", err, "go env value %s is not a string", key)
		}
		values[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, diagnosticErr("invalid_go_env", err, "go env object is incomplete")
	}
	if token, err := decoder.Token(); err != io.EOF || token != nil {
		return nil, diagnostic("invalid_go_env", "go env output has trailing data")
	}
	if len(values) != len(probeEnvNames) {
		return nil, diagnostic("invalid_go_env", "go env output does not contain the exact fixed field set")
	}
	for _, key := range probeEnvNames {
		if _, present := values[key]; !present {
			return nil, diagnostic("invalid_go_env", "go env output is missing %s", key)
		}
	}
	return values, nil
}

func validateProbe(values map[string]string, goroot, versionOS, versionArch string, host hostFacts, configRoot string) error {
	probedRoot, err := absolutePhysical(values["GOROOT"])
	if err != nil || probedRoot != goroot {
		return diagnosticErrRemedy("toolchain_executable_mismatch", toolchainSelectionRemedy, err,
			"go env GOROOT does not match the selected toolchain")
	}
	if values["GOHOSTOS"] != host.goos || values["GOOS"] != host.goos || versionOS != host.goos ||
		values["GOHOSTARCH"] != host.goarch || values["GOARCH"] != host.goarch || versionArch != host.goarch {
		return diagnostic("target_mismatch", "Go host, target, version, and manager platform must be identical")
	}
	if values["GOTELEMETRY"] != "off" {
		return diagnostic("telemetry_initialization_failed", "go env reports telemetry mode %q", values["GOTELEMETRY"])
	}
	telemetry, telemetryErr := absolutePhysical(values["GOTELEMETRYDIR"])
	physicalConfig, configErr := absolutePhysical(configRoot)
	if telemetryErr != nil || configErr != nil || !strictlyBelow(telemetry, physicalConfig) {
		return diagnostic("telemetry_directory_untrusted", "Go telemetry directory is outside the private platform configuration root")
	}
	return nil
}

func targetFromProbe(values map[string]string) (buildmeta.Target, error) {
	target := buildmeta.Target{GOOS: values["GOOS"], GOARCH: values["GOARCH"], Tuning: map[string]string{}}
	if key := tuningVariable(target.GOARCH); key != "" {
		if values[key] == "" || utf8.RuneCountInString(values[key]) > 8192 || !utf8.ValidString(values[key]) || strings.ContainsAny(values[key], "\r\n\x00") {
			return buildmeta.Target{}, diagnostic("target_mismatch", "native tuning variable %s is empty or malformed", key)
		}
		target.Tuning[key] = values[key]
	}
	return target, nil
}

func tuningVariable(goarch string) string {
	switch goarch {
	case "386":
		return "GO386"
	case "amd64":
		return "GOAMD64"
	case "arm":
		return "GOARM"
	case "arm64":
		return "GOARM64"
	case "mips", "mipsle":
		return "GOMIPS"
	case "mips64", "mips64le":
		return "GOMIPS64"
	case "ppc64", "ppc64le":
		return "GOPPC64"
	case "riscv64":
		return "GORISCV64"
	case "wasm":
		return "GOWASM"
	default:
		return ""
	}
}

func copyTarget(target buildmeta.Target) buildmeta.Target {
	copyValue := target
	copyValue.Tuning = make(map[string]string, len(target.Tuning))
	for key, value := range target.Tuning {
		copyValue.Tuning[key] = value
	}
	return copyValue
}

func absolutePhysical(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("path is not absolute")
	}
	return physicalPath(path)
}

func strictlyBelow(path, root string) bool { return path != root && isWithin(path, root) }

func isWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && relative != "." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

var chmod = os.Chmod
