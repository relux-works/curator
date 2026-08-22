package buildrepo

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1" // #nosec G505 -- Git SHA-1 object identity is a protocol input, not a security digest.
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"golang.org/x/text/unicode/norm"
)

// Stable admission error codes from the rc.5 manager profile.
const (
	CodeIdentityInvalid              = "build_repository_identity_invalid"
	CodeSourceUnavailable            = "build_repository_source_unavailable"
	CodeRefMoved                     = "build_repository_ref_moved"
	CodeIncompleteSource             = "build_repository_incomplete_source"
	CodeObjectSemanticsInvalid       = "build_repository_git_object_semantics_invalid"
	CodeLFSUnsupported               = "build_repository_git_lfs_unsupported"
	CodeLocalGitfileUnsupported      = "build_repository_local_gitfile_unsupported"
	CodeLocalBareUnsupported         = "build_repository_local_bare_unsupported"
	CodeLocalLinkedUnsupported       = "build_repository_local_linked_worktree_unsupported"
	CodeLocalLayoutUnsafe            = "build_repository_local_layout_unsafe"
	CodeLocalFormatUnsupported       = "build_repository_local_format_unsupported"
	CodeLocalObjectFormatUnsupported = "build_repository_local_object_format_unsupported"
)

// AdmissionError preserves the stable protocol code without exposing remote
// diagnostics or object contents.
type AdmissionError struct {
	Code string
	err  error
}

func (e *AdmissionError) Error() string { return e.Code + ": " + e.err.Error() }
func (e *AdmissionError) Unwrap() error { return e.err }

func admissionError(code, format string, args ...any) error {
	return &AdmissionError{Code: code, err: fmt.Errorf(format, args...)}
}

// ErrorCode returns the stable admission code carried by err, if any.
func ErrorCode(err error) string {
	var target *AdmissionError
	if errors.As(err, &target) {
		return target.Code
	}
	return ""
}

// Limits bounds raw-object expansion and repository processing.
type Limits struct {
	Timeout          time.Duration
	MaxObjects       int
	MaxObjectBytes   int64
	MaxExpandedBytes int64
	MaxFiles         int
	MaxPathBytes     int
	MaxTreeDepth     int
	MaxTagDepth      int
}

// DefaultLimits returns the fail-closed repository processing limits.
func DefaultLimits() Limits {
	return Limits{
		Timeout: 2 * time.Minute, MaxObjects: 200_000,
		MaxObjectBytes: 512 << 20, MaxExpandedBytes: 2 << 30,
		MaxFiles: 200_000, MaxPathBytes: 4096, MaxTreeDepth: 128, MaxTagDepth: 16,
	}
}

// GitTool is an operator-selected Git distribution. Executable and ExecPath
// must be absolute, ordinary paths; AllowedVersions pins tested release
// prefixes such as "git version 2.50.".
type GitTool struct {
	Executable      string
	ExecPath        string
	AllowedVersions []string
	AskPass         string
	SSHWrapper      string
}

// SSHPolicy contains the manager-owned inputs for the fixed SSH wrapper.
type SSHPolicy struct {
	Wrapper         string
	SSH             string
	ExpectedHost    string
	RepositoryPath  string
	EmptyConfig     string
	KnownHosts      string
	EmptyKnownHosts string
	Identity        string
	AgentSocket     string
	ConnectTimeout  int
}

// ExactSSHCommand is the data-free core of the manager binary SSH wrapper.
// It accepts only Git's exact wrapper tuple and returns the fixed OpenSSH argv.
func ExactSSHCommand(policy SSHPolicy, argv []string) ([]string, error) {
	expectedCommand := "git-upload-pack '" + policy.RepositoryPath + "'"
	if len(argv) != 3 || argv[0] != policy.Wrapper || argv[1] != policy.ExpectedHost || argv[2] != expectedCommand {
		return nil, admissionError(CodeIdentityInvalid, "SSH wrapper invocation does not match protected policy")
	}
	for _, value := range []string{policy.Wrapper, policy.SSH, policy.EmptyConfig, policy.KnownHosts, policy.EmptyKnownHosts} {
		if !filepath.IsAbs(value) {
			return nil, admissionError(CodeIdentityInvalid, "SSH policy path is not absolute")
		}
	}
	if policy.ConnectTimeout <= 0 || !sshPathRE.MatchString(policy.RepositoryPath) {
		return nil, admissionError(CodeIdentityInvalid, "SSH policy is invalid")
	}
	result := []string{policy.SSH, "-F", policy.EmptyConfig, "-T",
		"-o", "BatchMode=yes", "-o", "NumberOfPasswordPrompts=0", "-o", "PasswordAuthentication=no",
		"-o", "KbdInteractiveAuthentication=no", "-o", "PreferredAuthentications=publickey",
		"-o", "HostbasedAuthentication=no", "-o", "GSSAPIAuthentication=no", "-o", "StrictHostKeyChecking=yes",
		"-o", "UserKnownHostsFile=" + policy.KnownHosts, "-o", "GlobalKnownHostsFile=" + policy.EmptyKnownHosts,
		"-o", "CheckHostIP=no", "-o", "VerifyHostKeyDNS=no", "-o", "UpdateHostKeys=no",
		"-o", "ForwardAgent=no", "-o", "ForwardX11=no", "-o", "ClearAllForwardings=yes",
		"-o", "PermitLocalCommand=no", "-o", "ProxyCommand=none", "-o", "ProxyJump=none",
		"-o", "ControlMaster=no", "-o", "ControlPath=none", "-o", "ControlPersist=no",
		"-o", "RequestTTY=no", "-o", "EscapeChar=none", "-o", "EnableEscapeCommandline=no",
		"-o", "CanonicalizeHostname=no", "-o", "ConnectionAttempts=1", "-o", "ConnectTimeout=" + strconv.Itoa(policy.ConnectTimeout)}
	switch {
	case policy.Identity != "" && policy.AgentSocket == "":
		if !filepath.IsAbs(policy.Identity) {
			return nil, admissionError(CodeIdentityInvalid, "SSH identity path is not absolute")
		}
		result = append(result, "-o", "IdentitiesOnly=yes", "-o", "IdentityAgent=none", "-i", policy.Identity)
	case policy.AgentSocket != "" && policy.Identity == "":
		if !filepath.IsAbs(policy.AgentSocket) {
			return nil, admissionError(CodeIdentityInvalid, "SSH agent path is not absolute")
		}
		result = append(result, "-o", "IdentitiesOnly=no", "-o", "IdentityFile=none", "-o", "IdentityAgent="+policy.AgentSocket)
	case policy.Identity != "" && policy.AgentSocket != "":
		// The pinned-agent form: the operator agent holds the private key and
		// the named identity (conventionally the public half) pins which
		// single key the agent offers. One authentication attempt, one
		// disclosed public key, and passphrase-protected keys authenticate
		// without a prompt.
		if !filepath.IsAbs(policy.Identity) {
			return nil, admissionError(CodeIdentityInvalid, "SSH identity path is not absolute")
		}
		if !filepath.IsAbs(policy.AgentSocket) {
			return nil, admissionError(CodeIdentityInvalid, "SSH agent path is not absolute")
		}
		result = append(result, "-o", "IdentitiesOnly=yes", "-o", "IdentityAgent="+policy.AgentSocket, "-i", policy.Identity)
	default:
		return nil, admissionError(CodeIdentityInvalid, "SSH authentication policy must select an identity, an agent, or both")
	}
	return append(result, policy.ExpectedHost, expectedCommand), nil
}

// ValidateGitTool performs trusted discovery/version pinning before any
// package or substitution path is read.
func ValidateGitTool(ctx context.Context, tool GitTool) error {
	for name, value := range map[string]string{"git": tool.Executable, "git exec path": tool.ExecPath} {
		if !filepath.IsAbs(value) {
			return admissionError(CodeIdentityInvalid, "%s path must be absolute", name)
		}
		info, err := os.Lstat(value)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || (name == "git" && !info.Mode().IsRegular()) || (name != "git" && !info.IsDir()) {
			return admissionError(CodeIdentityInvalid, "%s path is not an admitted ordinary object", name)
		}
	}
	for name, value := range map[string]string{"credential broker": tool.AskPass, "SSH wrapper": tool.SSHWrapper} {
		if value == "" {
			continue
		}
		info, err := os.Lstat(value)
		if !filepath.IsAbs(value) || err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return admissionError(CodeIdentityInvalid, "%s is not an admitted absolute regular file", name)
		}
	}
	cmd := exec.CommandContext(ctx, tool.Executable, "--version") // #nosec G204 -- operator-selected absolute executable.
	cmd.Env = cleanDiscoveryEnvironment()
	out, err := cmd.Output()
	if err != nil || len(out) > 256 {
		return admissionError(CodeIdentityInvalid, "trusted Git version probe failed")
	}
	version := strings.TrimSpace(string(out))
	for _, allowed := range tool.AllowedVersions {
		if allowed != "" && strings.HasPrefix(version, allowed) {
			return nil
		}
	}
	return admissionError(CodeIdentityInvalid, "Git release family is not operator-pinned")
}

// NetworkRequest binds exact source, Git tool, and processing limits.
type NetworkRequest struct {
	Source Source
	Lock   LockedCommit
	Tag    string
	// RefKind/RefValue select an operator substitution. Revision retains the
	// immutable lock path; branch and tag are resolved to their actual terminal
	// commit and object format in the private repository.
	RefKind  string
	RefValue string
	Tool     GitTool
	Limits   Limits
}

// File is one normalized regular file in an admitted snapshot.
type File struct {
	Path       string
	Content    []byte
	Executable bool
}

// Snapshot is a proved immutable repository tree and its canonical identity.
type Snapshot struct {
	ObjectFormat   string
	Commit         string
	Files          []File
	CanonicalBytes []byte
	Digest         string
	TagVerified    bool
}

// AcquireNetwork fetches exactly one locked object or exact tag into a fresh
// private bare repository and proves the snapshot from raw objects.
func AcquireNetwork(ctx context.Context, request NetworkRequest) (*Snapshot, error) {
	if err := ValidateGitTool(ctx, request.Tool); err != nil {
		return nil, err
	}
	parsedSource, err := ParseSource(request.Source.Git)
	if err != nil || parsedSource != request.Source {
		return nil, admissionError(CodeIdentityInvalid, "network source is not canonical parsed input")
	}
	if request.Source.Transport != "https" && request.Source.Transport != "ssh" {
		return nil, admissionError(CodeIdentityInvalid, "unsupported transport")
	}
	if request.Source.Transport == "https" && request.Tool.AskPass == "" {
		return nil, admissionError(CodeIdentityInvalid, "HTTPS requires a manager credential broker")
	}
	if request.Source.Transport == "ssh" && request.Tool.SSHWrapper == "" {
		return nil, admissionError(CodeIdentityInvalid, "SSH requires the exact manager wrapper")
	}
	if _, err := ParseLockedCommit(map[string]any{"object_format": request.Lock.ObjectFormat, "hex": request.Lock.Hex}, "lock"); err != nil {
		return nil, admissionError(CodeIdentityInvalid, "invalid immutable lock")
	}
	if request.Tag != "" && !ValidRefName(request.Tag) {
		return nil, admissionError(CodeIdentityInvalid, "invalid exact tag")
	}
	if request.RefKind != "" && request.RefKind != "revision" && request.RefKind != "branch" && request.RefKind != "tag" {
		return nil, admissionError(CodeIdentityInvalid, "invalid substitution ref kind")
	}
	if request.RefKind != "" && !ValidRefName(request.RefValue) {
		return nil, admissionError(CodeIdentityInvalid, "invalid substitution ref")
	}
	limits := normalizedLimits(request.Limits)
	ctx, cancel := context.WithTimeout(ctx, limits.Timeout)
	defer cancel()
	return acquireNetworkFormat(ctx, request, limits)
}

func acquireNetworkFormat(ctx context.Context, request NetworkRequest, limits Limits) (*Snapshot, error) {
	root, err := os.MkdirTemp("", "curator-buildrepo-")
	if err != nil {
		return nil, admissionError(CodeSourceUnavailable, "cannot create private state")
	}
	defer func() { _ = os.RemoveAll(root) }()
	paths, err := makePrivatePaths(root)
	if err != nil {
		return nil, admissionError(CodeSourceUnavailable, "cannot initialize private state")
	}
	env := cleanGitEnvironment(paths, request.Tool, request.Source.Transport)
	initArgs := []string{"--git-dir=" + paths.repo, "-c", "init.defaultBranch=curator-invalid", "init", "--bare", "--quiet", "--template=" + paths.template, "--object-format=" + request.Lock.ObjectFormat, "--ref-format=files"}
	if err := runGit(ctx, request.Tool.Executable, paths.work, env, initArgs...); err != nil {
		return nil, admissionError(CodeSourceUnavailable, "private Git initialization failed")
	}
	destination := "refs/curator/locked"
	sourceRef := request.Lock.Hex
	expectedCommit := request.Lock.Hex
	exactTag := request.Tag
	if request.Tag != "" {
		destination = "refs/curator/tag"
		sourceRef = "refs/tags/" + request.Tag
	}
	if request.RefKind != "" {
		destination = "refs/curator/effective"
		sourceRef = request.RefValue
		exactTag = ""
		if request.RefKind == "branch" {
			sourceRef = "refs/heads/" + request.RefValue
			expectedCommit = ""
		}
		if request.RefKind == "tag" {
			sourceRef = "refs/tags/" + request.RefValue
			expectedCommit = ""
			exactTag = request.RefValue
		}
	}
	refspec := sourceRef + ":" + destination
	fetchArgs := strictFetchArgs(paths.repo, paths.hooks, request.Tool.AskPass, request.Source.Transport, request.Source.Git, refspec)
	if err := runGit(ctx, request.Tool.Executable, paths.work, env, fetchArgs...); err != nil {
		return nil, admissionError(CodeSourceUnavailable, "exact source fetch failed")
	}
	if err := validatePrivateRepository(paths.repo, request.Lock.ObjectFormat); err != nil {
		return nil, err
	}
	selected, err := readSingleOID(filepath.Join(paths.repo, filepath.FromSlash(destination)), request.Lock.ObjectFormat)
	if err != nil {
		return nil, admissionError(CodeIncompleteSource, "manager destination ref is invalid")
	}
	snapshot, err := proveRepository(ctx, request.Tool, env, paths, request.Lock.ObjectFormat, selected, exactTag, limits)
	if err != nil {
		return nil, err
	}
	if request.Tag != "" && snapshot.Commit != expectedCommit {
		return nil, admissionError(CodeRefMoved, "exact tag terminal commit differs from lock")
	}
	if expectedCommit != "" && request.Tag == "" && snapshot.Commit != expectedCommit {
		return nil, admissionError(CodeIncompleteSource, "locked object is not the selected commit")
	}
	snapshot.TagVerified = exactTag != ""
	return snapshot, nil
}

func validTreeComponent(name string) bool {
	return identifiers.PortableComponent(name) && !strings.Contains(name, "/")
}

type privatePaths struct{ root, repo, work, home, config, template, hooks, path string }

func makePrivatePaths(root string) (privatePaths, error) {
	p := privatePaths{root: root, repo: filepath.Join(root, "repo.git"), work: filepath.Join(root, "work"), home: filepath.Join(root, "home"), config: filepath.Join(root, "config"), template: filepath.Join(root, "empty-template"), hooks: filepath.Join(root, "empty-hooks"), path: filepath.Join(root, "empty-path")}
	for _, dir := range []string{p.work, p.home, p.config, p.template, p.hooks, p.path} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return privatePaths{}, err
		}
	}
	for _, name := range []string{"global.gitconfig", "system.gitconfig"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			return privatePaths{}, err
		}
	}
	return p, nil
}

func strictFetchArgs(repo, hooks, askPass, transport, source, refspec string) []string {
	return []string{"--git-dir=" + repo, "--no-replace-objects", "--no-lazy-fetch", "--no-optional-locks",
		"-c", "protocol.allow=never", "-c", "protocol." + transport + ".allow=always", "-c", "protocol.version=0",
		"-c", "credential.helper=", "-c", "core.askPass=" + askPass, "-c", "core.hooksPath=" + hooks,
		"-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "submodule.recurse=false",
		"-c", "fetch.recurseSubmodules=false", "-c", "maintenance.auto=false", "-c", "fetch.writeCommitGraph=false",
		"-c", "fetch.fsckObjects=true", "-c", "transfer.fsckObjects=true", "-c", "http.followRedirects=false",
		"-c", "http.sslVerify=true", "-c", "http.proxy=", "-c", "https.proxy=", "fetch", "--quiet", "--atomic",
		"--no-tags", "--no-recurse-submodules", "--no-auto-maintenance", "--no-write-fetch-head", "--no-write-commit-graph",
		"--refmap=", "--jobs=1", "--upload-pack=git-upload-pack", "--", source, refspec}
}

func cleanDiscoveryEnvironment() []string {
	env := []string{"LANG=C", "LC_ALL=C"}
	for _, key := range []string{"SYSTEMROOT", "WINDIR", "COMSPEC", "PATHEXT"} {
		if value := os.Getenv(key); value != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func cleanGitEnvironment(p privatePaths, tool GitTool, transport string) []string {
	env := cleanDiscoveryEnvironment()
	env = append(env,
		"GIT_CONFIG_GLOBAL="+filepath.Join(p.root, "global.gitconfig"), "GIT_CONFIG_SYSTEM="+filepath.Join(p.root, "system.gitconfig"),
		"GIT_CONFIG_NOSYSTEM=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_NO_LAZY_FETCH=1", "GIT_OPTIONAL_LOCKS=0",
		"GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat", "GIT_PROTOCOL_FROM_USER=0", "GIT_LITERAL_PATHSPECS=1", "GIT_ATTR_NOSYSTEM=1",
		"GIT_EXEC_PATH="+tool.ExecPath, "HOME="+p.home, "XDG_CONFIG_HOME="+p.config, "PATH="+p.path,
	)
	if transport == "https" && tool.AskPass != "" {
		env = append(env, "GIT_ASKPASS="+tool.AskPass)
	}
	if transport == "ssh" && tool.SSHWrapper != "" {
		env = append(env, "GIT_SSH="+tool.SSHWrapper, "GIT_SSH_VARIANT=ssh")
	}
	return env
}

func runGit(ctx context.Context, executable, dir string, env []string, args ...string) error {
	cmd := exec.CommandContext(ctx, executable, args...) // #nosec G204 -- absolute trusted executable; closed argument construction.
	cmd.Dir, cmd.Env, cmd.Stdin = dir, env, bytes.NewReader(nil)
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.Discard, &boundedWriter{writer: &stderr, remaining: 64 << 10}
	return cmd.Run()
}

type boundedWriter struct {
	writer    io.Writer
	remaining int
}

func (w *boundedWriter) Write(payload []byte) (int, error) {
	original := len(payload)
	if len(payload) > w.remaining {
		payload = payload[:w.remaining]
	}
	if len(payload) > 0 {
		if _, err := w.writer.Write(payload); err != nil {
			return 0, err
		}
		w.remaining -= len(payload)
	}
	return original, nil
}

func normalizedLimits(l Limits) Limits {
	d := DefaultLimits()
	if l.Timeout > 0 {
		d.Timeout = l.Timeout
	}
	if l.MaxObjects > 0 {
		d.MaxObjects = l.MaxObjects
	}
	if l.MaxObjectBytes > 0 {
		d.MaxObjectBytes = l.MaxObjectBytes
	}
	if l.MaxExpandedBytes > 0 {
		d.MaxExpandedBytes = l.MaxExpandedBytes
	}
	if l.MaxFiles > 0 {
		d.MaxFiles = l.MaxFiles
	}
	if l.MaxPathBytes > 0 {
		d.MaxPathBytes = l.MaxPathBytes
	}
	if l.MaxTreeDepth > 0 {
		d.MaxTreeDepth = l.MaxTreeDepth
	}
	if l.MaxTagDepth > 0 {
		d.MaxTagDepth = l.MaxTagDepth
	}
	return d
}

func validatePrivateRepository(repo, objectFormat string) error {
	entries, err := os.ReadDir(repo)
	if err != nil {
		return admissionError(CodeIncompleteSource, "private repository is unreadable")
	}
	allowed := map[string]bool{"HEAD": true, "config": true, "objects": true, "refs": true}
	for _, entry := range entries {
		if !allowed[entry.Name()] || entry.Type()&os.ModeSymlink != 0 {
			return admissionError(CodeIncompleteSource, "private repository contains unexpected state")
		}
	}
	for _, forbidden := range []string{"FETCH_HEAD", "shallow", "objects/info/alternates", "objects/info/http-alternates", "info/grafts", "objects/info/commit-graph", "objects/pack/multi-pack-index", "refs/replace"} {
		if _, err := os.Lstat(filepath.Join(repo, filepath.FromSlash(forbidden))); err == nil {
			return admissionError(CodeIncompleteSource, "private repository contains forbidden state")
		}
	}
	config, err := os.ReadFile(filepath.Join(repo, "config")) // #nosec G304 -- fixed Git admin filename below a private manager-created repository.
	if err != nil || (objectFormat == "sha256" && !bytes.Contains(bytes.ToLower(config), []byte("objectformat = sha256"))) {
		return admissionError(CodeIncompleteSource, "private repository object format is invalid")
	}
	return nil
}

func readSingleOID(path, format string) (string, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- path is a manager-derived exact ref below the private repository.
	if err != nil || len(payload) > 66 {
		return "", errors.New("invalid ref")
	}
	value := strings.TrimSuffix(strings.TrimSuffix(string(payload), "\n"), "\r")
	if !validOID(value, format) {
		return "", errors.New("invalid ref")
	}
	return value, nil
}

func validOID(value, format string) bool {
	if format == "sha1" {
		return sha1RE.MatchString(value)
	}
	return format == "sha256" && sha256RE.MatchString(value)
}

type rawObject struct {
	oid, kind string
	data      []byte
}

type objectReader struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	format   string
	limits   Limits
	count    int
	expanded int64
	cache    map[string]rawObject
}

func newObjectReader(ctx context.Context, tool GitTool, env []string, paths privatePaths, format string, limits Limits) (*objectReader, error) {
	args := []string{"--git-dir=" + paths.repo, "--no-replace-objects", "--no-lazy-fetch", "--no-optional-locks", "-c", "core.hooksPath=" + paths.hooks, "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "maintenance.auto=false", "cat-file", "--batch=%(objectname) %(objecttype) %(objectsize)"}
	cmd := exec.CommandContext(ctx, tool.Executable, args...) // #nosec G204 -- trusted absolute executable and fixed argv.
	cmd.Dir, cmd.Env, cmd.Stderr = paths.work, env, io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &objectReader{cmd: cmd, stdin: stdin, stdout: bufio.NewReaderSize(stdout, 64<<10), format: format, limits: limits, cache: map[string]rawObject{}}, nil
}

func (r *objectReader) read(oid string) (rawObject, error) {
	if object, ok := r.cache[oid]; ok {
		return object, nil
	}
	if !validOID(oid, r.format) || r.count >= r.limits.MaxObjects {
		return rawObject{}, admissionError(CodeIncompleteSource, "invalid or excessive object request")
	}
	if _, err := io.WriteString(r.stdin, oid+"\n"); err != nil {
		return rawObject{}, admissionError(CodeIncompleteSource, "object request failed")
	}
	header, err := r.stdout.ReadString('\n')
	if err != nil || len(header) > 256 {
		return rawObject{}, admissionError(CodeIncompleteSource, "malformed object header")
	}
	parts := strings.Fields(strings.TrimSuffix(header, "\n"))
	if len(parts) != 3 || parts[0] != oid || (parts[1] != "commit" && parts[1] != "tag" && parts[1] != "tree" && parts[1] != "blob") {
		return rawObject{}, admissionError(CodeIncompleteSource, "malformed object response")
	}
	if parts[2] == "" || (len(parts[2]) > 1 && parts[2][0] == '0') {
		return rawObject{}, admissionError(CodeIncompleteSource, "non-canonical object size")
	}
	size, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || size < 0 || size > r.limits.MaxObjectBytes || r.expanded+size > r.limits.MaxExpandedBytes {
		return rawObject{}, admissionError(CodeIncompleteSource, "object size limit exceeded")
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r.stdout, data); err != nil {
		return rawObject{}, admissionError(CodeIncompleteSource, "truncated object")
	}
	terminator, err := r.stdout.ReadByte()
	if err != nil || terminator != '\n' {
		return rawObject{}, admissionError(CodeIncompleteSource, "malformed object terminator")
	}
	if computeOID(r.format, parts[1], data) != oid {
		return rawObject{}, admissionError(CodeIncompleteSource, "object identity mismatch")
	}
	object := rawObject{oid: oid, kind: parts[1], data: data}
	r.count++
	r.expanded += size
	r.cache[oid] = object
	return object, nil
}

func (r *objectReader) close() error {
	_ = r.stdin.Close()
	trailing, _ := io.ReadAll(io.LimitReader(r.stdout, 2))
	err := r.cmd.Wait()
	if err != nil || len(trailing) != 0 {
		return admissionError(CodeIncompleteSource, "object reader did not terminate cleanly")
	}
	return nil
}

func computeOID(format, kind string, data []byte) string {
	var h hash.Hash
	if format == "sha1" {
		h = sha1.New() // #nosec G401 -- Git SHA-1 object identity is protocol compatibility, not a security digest.
	} else {
		h = sha256.New()
	}
	_, _ = fmt.Fprintf(h, "%s %d%c", kind, len(data), byte(0))
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func proveRepository(ctx context.Context, tool GitTool, env []string, paths privatePaths, format, selected, exactTag string, limits Limits) (_ *Snapshot, resultErr error) {
	reader, err := newObjectReader(ctx, tool, env, paths, format, limits)
	if err != nil {
		return nil, admissionError(CodeIncompleteSource, "object reader could not start")
	}
	defer func() {
		if err := reader.close(); resultErr == nil && err != nil {
			resultErr = err
		}
	}()
	commit, err := peelCommit(reader, selected, exactTag, limits.MaxTagDepth)
	if err != nil {
		return nil, err
	}
	commitObject, err := reader.read(commit)
	if err != nil {
		return nil, err
	}
	tree, err := parseCommit(commitObject, format)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0)
	seenPaths := map[string]string{}
	if err := walkTree(reader, tree, "", 0, limits, seenPaths, &files); err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	canonical := frameSnapshot(files)
	digest := sha256.Sum256(canonical)
	return &Snapshot{ObjectFormat: format, Commit: commit, Files: files, CanonicalBytes: canonical, Digest: "sha256:" + hex.EncodeToString(digest[:])}, nil
}

func peelCommit(r *objectReader, oid, exactTag string, maxDepth int) (string, error) {
	seen := map[string]bool{}
	for depth := 0; ; depth++ {
		if seen[oid] || depth > maxDepth {
			return "", admissionError(CodeObjectSemanticsInvalid, "tag chain is cyclic or too deep")
		}
		seen[oid] = true
		object, err := r.read(oid)
		if err != nil {
			return "", err
		}
		if object.kind == "commit" {
			return oid, nil
		}
		if object.kind != "tag" {
			return "", admissionError(CodeObjectSemanticsInvalid, "selected object is not a commit or tag")
		}
		target, targetType, tag, err := parseTag(object, r.format)
		if err != nil {
			return "", err
		}
		if depth == 0 && exactTag != "" && tag != exactTag {
			return "", admissionError(CodeObjectSemanticsInvalid, "outer annotated tag name mismatch")
		}
		targetObject, err := r.read(target)
		if err != nil {
			return "", err
		}
		if targetObject.kind != targetType {
			return "", admissionError(CodeObjectSemanticsInvalid, "annotated tag target type mismatch")
		}
		oid = target
	}
}

type objectHeader struct {
	key, value   string
	continuation bool
}

func parseHeaders(data []byte) ([]objectHeader, error) {
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, errors.New("invalid header bytes")
	}
	separator := bytes.Index(data, []byte("\n\n"))
	if separator < 0 {
		return nil, errors.New("missing separator")
	}
	if bytes.IndexByte(data[:separator], '\r') >= 0 {
		return nil, errors.New("invalid header bytes")
	}
	lines := bytes.Split(data[:separator], []byte{'\n'})
	headers := make([]objectHeader, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			return nil, errors.New("empty header")
		}
		if line[0] == ' ' {
			if len(headers) == 0 {
				return nil, errors.New("orphan continuation")
			}
			headers = append(headers, objectHeader{value: string(line[1:]), continuation: true})
			continue
		}
		space := bytes.IndexByte(line, ' ')
		if space <= 0 {
			return nil, errors.New("malformed header")
		}
		key := string(line[:space])
		if !headerKey(key) {
			return nil, errors.New("invalid header key")
		}
		headers = append(headers, objectHeader{key: key, value: string(line[space+1:])})
	}
	return headers, nil
}

func headerKey(value string) bool {
	for i, b := range []byte(value) {
		//nolint:staticcheck // The positive grammar mirrors Git's header production directly.
		if !(b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || i > 0 && (b >= '0' && b <= '9' || b == '-')) {
			return false
		}
	}
	return value != ""
}

func parseCommit(object rawObject, format string) (string, error) {
	if object.kind != "commit" {
		return "", admissionError(CodeObjectSemanticsInvalid, "terminal object is not a commit")
	}
	headers, err := parseHeaders(object.data)
	if err != nil {
		return "", admissionError(CodeObjectSemanticsInvalid, "invalid commit headers")
	}
	if len(headers) < 3 || headers[0].continuation || headers[0].key != "tree" || !validOID(headers[0].value, format) {
		return "", admissionError(CodeObjectSemanticsInvalid, "commit tree header is invalid")
	}
	i := 1
	for i < len(headers) && headers[i].key == "parent" && !headers[i].continuation {
		if !validOID(headers[i].value, format) {
			return "", admissionError(CodeObjectSemanticsInvalid, "commit parent is invalid")
		}
		i++
	}
	for _, required := range []string{"author", "committer"} {
		if i >= len(headers) || headers[i].continuation || headers[i].key != required || headers[i].value == "" {
			return "", admissionError(CodeObjectSemanticsInvalid, "commit required header is invalid")
		}
		i++
	}
	structural := map[string]bool{"tree": true, "parent": true, "author": true, "committer": true}
	lastExtra := false
	for ; i < len(headers); i++ {
		h := headers[i]
		if h.continuation {
			if !lastExtra {
				return "", admissionError(CodeObjectSemanticsInvalid, "invalid commit continuation")
			}
			continue
		}
		if structural[h.key] {
			return "", admissionError(CodeObjectSemanticsInvalid, "duplicate or misplaced commit header")
		}
		lastExtra = true
	}
	return headers[0].value, nil
}

func parseTag(object rawObject, format string) (target, targetType, tag string, err error) {
	headers, parseErr := parseHeaders(object.data)
	if parseErr != nil {
		err = admissionError(CodeObjectSemanticsInvalid, "invalid tag headers")
		return
	}
	required := []string{"object", "type", "tag"}
	if len(headers) < 3 {
		err = admissionError(CodeObjectSemanticsInvalid, "missing tag headers")
		return
	}
	for i, key := range required {
		if headers[i].continuation || headers[i].key != key || headers[i].value == "" {
			err = admissionError(CodeObjectSemanticsInvalid, "invalid tag header order")
			return
		}
	}
	target, targetType, tag = headers[0].value, headers[1].value, headers[2].value
	if !validOID(target, format) || (targetType != "commit" && targetType != "tag") || !ValidRefName(tag) {
		err = admissionError(CodeObjectSemanticsInvalid, "invalid annotated tag semantics")
		return
	}
	i := 3
	if i < len(headers) && headers[i].key == "tagger" && !headers[i].continuation {
		if headers[i].value == "" {
			err = admissionError(CodeObjectSemanticsInvalid, "empty tagger")
			return
		}
		i++
	}
	structural := map[string]bool{"object": true, "type": true, "tag": true, "tagger": true}
	lastExtra := false
	for ; i < len(headers); i++ {
		h := headers[i]
		if h.continuation {
			if !lastExtra {
				err = admissionError(CodeObjectSemanticsInvalid, "invalid tag continuation")
				return
			}
			continue
		}
		if structural[h.key] {
			err = admissionError(CodeObjectSemanticsInvalid, "duplicate tag header")
			return
		}
		lastExtra = true
	}
	return
}

func walkTree(r *objectReader, oid, prefix string, depth int, limits Limits, seen map[string]string, files *[]File) error {
	if depth > limits.MaxTreeDepth {
		return admissionError(CodeIncompleteSource, "tree depth limit exceeded")
	}
	object, err := r.read(oid)
	if err != nil {
		return err
	}
	if object.kind != "tree" {
		return admissionError(CodeObjectSemanticsInvalid, "tree entry target type mismatch")
	}
	idBytes := 20
	if r.format == "sha256" {
		idBytes = 32
	}
	data := object.data
	local := map[string]bool{}
	for len(data) > 0 {
		space := bytes.IndexByte(data, ' ')
		nul := bytes.IndexByte(data, 0)
		if space <= 0 || nul <= space+1 || len(data) < nul+1+idBytes {
			return admissionError(CodeObjectSemanticsInvalid, "malformed tree")
		}
		mode, nameBytes := string(data[:space]), data[space+1:nul]
		if !utf8.Valid(nameBytes) {
			return admissionError(CodeObjectSemanticsInvalid, "tree path is not valid UTF-8")
		}
		name := string(nameBytes)
		if !validTreeComponent(name) {
			return admissionError(CodeObjectSemanticsInvalid, "invalid tree component")
		}
		collision := strings.ToLower(norm.NFC.String(name))
		if local[collision] {
			return admissionError(CodeObjectSemanticsInvalid, "platform-colliding tree names")
		}
		local[collision] = true
		child := hex.EncodeToString(data[nul+1 : nul+1+idBytes])
		data = data[nul+1+idBytes:]
		path := name
		if prefix != "" {
			path = prefix + "/" + name
		}
		if len([]byte(path)) > limits.MaxPathBytes {
			return admissionError(CodeIncompleteSource, "path length limit exceeded")
		}
		if old, exists := seen[collisionKey(path)]; exists && old != path {
			return admissionError(CodeObjectSemanticsInvalid, "platform path collision")
		}
		seen[collisionKey(path)] = path
		switch mode {
		case "40000":
			if err := walkTree(r, child, path, depth+1, limits, seen, files); err != nil {
				return err
			}
		case "100644", "100755":
			blob, err := r.read(child)
			if err != nil {
				return err
			}
			if blob.kind != "blob" {
				return admissionError(CodeObjectSemanticsInvalid, "blob entry target type mismatch")
			}
			if isLFSPointer(blob.data) {
				return admissionError(CodeLFSUnsupported, "reachable Git LFS pointer is unsupported")
			}
			if len(*files) >= limits.MaxFiles {
				return admissionError(CodeIncompleteSource, "file count limit exceeded")
			}
			content := append([]byte(nil), blob.data...)
			*files = append(*files, File{Path: path, Content: content, Executable: mode == "100755"})
		default:
			return admissionError(CodeObjectSemanticsInvalid, "unsupported tree mode")
		}
	}
	return nil
}

func collisionKey(path string) string { return strings.ToLower(norm.NFC.String(path)) }

func frameSnapshot(files []File) []byte {
	result := append([]byte(nil), []byte("curator-build-source-v1\x00")...)
	for _, file := range files {
		result = append(result, 'F')
		result = binary.BigEndian.AppendUint64(result, uint64(len([]byte(file.Path))))
		result = append(result, file.Path...)
		result = binary.BigEndian.AppendUint64(result, uint64(len(file.Content)))
		result = append(result, file.Content...)
	}
	return result
}

// Materialize writes only proved blob bytes as regular files. Existing or
// non-empty destinations are rejected.
func (s *Snapshot) Materialize(destination string) error {
	if entries, err := os.ReadDir(destination); err == nil && len(entries) != 0 {
		return admissionError(CodeLocalLayoutUnsafe, "snapshot destination is not empty")
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	for _, file := range s.Files {
		target := filepath.Join(destination, filepath.FromSlash(file.Path))
		rel, err := filepath.Rel(destination, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return admissionError(CodeObjectSemanticsInvalid, "snapshot path escapes destination")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if file.Executable {
			mode = 0o700
		}
		if err := os.WriteFile(target, file.Content, mode); err != nil {
			return err
		}
	}
	return nil
}

func isLFSPointer(data []byte) bool {
	if len(data) == 0 || len(data) >= 1024 {
		return false
	}
	trimmed := strings.TrimFunc(string(data), unicode.IsSpace)
	lines := strings.Split(trimmed, "\n")
	state := 0
	extensions := map[string]string{}
	priorities := map[byte]string{}
	versions := map[string]bool{"https://git-lfs.github.com/spec/v1": true, "https://hawser.github.com/spec/v1": true, "http://git-media.io/v/2": true}
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		space := strings.IndexByte(line, ' ')
		if space <= 0 {
			return false
		}
		key, value := line[:space], line[space+1:]
		if state == 3 {
			return false
		}
		switch key {
		case "version":
			if state != 0 || !versions[value] {
				return false
			}
			state = 1
		case "oid":
			if state != 1 || !strings.HasPrefix(value, "sha256:") || !sha256RE.MatchString(strings.TrimPrefix(value, "sha256:")) {
				return false
			}
			state = 2
		case "size":
			if state != 2 {
				return false
			}
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil || n < 0 {
				return false
			}
			state = 3
		default:
			if !strings.HasPrefix(key, "ext-") || len(key) < 7 || key[4] < '0' || key[4] > '9' || key[5] != '-' {
				return false
			}
			word := key[6]
			//nolint:staticcheck // The positive grammar mirrors the extension-key production directly.
			if !(word >= 'A' && word <= 'Z' || word >= 'a' && word <= 'z' || word >= '0' && word <= '9' || word == '_') {
				return false
			}
			if !strings.HasPrefix(value, "sha256:") || !sha256RE.MatchString(strings.TrimPrefix(value, "sha256:")) {
				return false
			}
			if previous, ok := priorities[key[4]]; ok && previous != key {
				return false
			}
			priorities[key[4]] = key
			extensions[key] = value
		}
	}
	return state == 3
}
