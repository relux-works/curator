// Package buildrepo owns the schema-7 external repository wire models. It is
// deliberately read-only: acquisition, audit, compilation, and installation
// live outside this package.
package buildrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/verr"
)

const (
	// DescriptorName is the closed schema-7 build descriptor filename.
	DescriptorName = "skill-build.json"
	// ProtocolVersion is the authoritative external-repository protocol release.
	ProtocolVersion = "1.0.0-rc.8"
	// ConformanceManifestSHA256 pins the released conformance manifest.
	ConformanceManifestSHA256 = "d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1"
)

var (
	hostRE    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*$`)
	sshUserRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	sshPathRE = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
	scpRE     = regexp.MustCompile(`^(?:([A-Za-z0-9][A-Za-z0-9._-]{0,63})@)?([A-Za-z0-9][A-Za-z0-9.-]*):([A-Za-z0-9._/-]+)$`)
	sha1RE    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256RE  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// LockedCommit is a full immutable Git object identifier and its algorithm.
type LockedCommit struct {
	ObjectFormat string
	Hex          string
}

// Source is a declared Git transport and its canonical network identity.
type Source struct {
	Git       string
	Identity  string
	Transport string
}

// Target is one manager-owned command target from skill-build.json.
type Target struct {
	Name      string
	Driver    string
	BuildRoot string
	SourceDir string
}

// Descriptor is the parsed closed external-build descriptor.
type Descriptor struct {
	SchemaVersion int
	Targets       map[string]Target
}

// ParseLockedCommit validates a full lowercase SHA-1 or SHA-256 commit object.
func ParseLockedCommit(raw any, field string) (LockedCommit, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return LockedCommit{}, verr.New(field, "must be an object")
	}
	if err := rejectUnknown(obj, field, "object_format", "hex"); err != nil {
		return LockedCommit{}, err
	}
	format, _ := obj["object_format"].(string)
	hexValue, _ := obj["hex"].(string)
	valid := format == "sha1" && sha1RE.MatchString(hexValue) || format == "sha256" && sha256RE.MatchString(hexValue)
	if !valid {
		return LockedCommit{}, verr.New(field, "requires a full lowercase sha1 or sha256 object id matching object_format")
	}
	return LockedCommit{ObjectFormat: format, Hex: hexValue}, nil
}

// ParseSource accepts exactly the rc.5 HTTPS, SSH URI, and SCP spellings and
// returns their shared canonical network-git identity.
func ParseSource(raw string) (Source, error) {
	if raw == "" || !utf8.ValidString(raw) || utf8.RuneCountInString(raw) > 4096 || strings.ContainsAny(raw, "%?#\\") || containsRepositoryWhitespaceOrControl(raw) {
		return Source{}, fmt.Errorf("repository git source is not in the released rc.5 grammar")
	}
	var host, repoPath, transport string
	switch {
	case strings.HasPrefix(raw, "https://"):
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Host == "" || parsed.Port() != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return Source{}, fmt.Errorf("repository HTTPS source is invalid")
		}
		host, repoPath, transport = parsed.Hostname(), strings.TrimPrefix(parsed.Path, "/"), "https"
	case strings.HasPrefix(raw, "ssh://"):
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "ssh" || parsed.Host == "" || parsed.Port() != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return Source{}, fmt.Errorf("repository SSH source is invalid")
		}
		if parsed.User != nil {
			if _, password := parsed.User.Password(); password || !sshUserRE.MatchString(parsed.User.Username()) {
				return Source{}, fmt.Errorf("repository SSH username is invalid")
			}
		}
		host, repoPath, transport = parsed.Hostname(), strings.TrimPrefix(parsed.Path, "/"), "ssh"
		if !sshPathRE.MatchString(repoPath) {
			return Source{}, fmt.Errorf("repository SSH path must be portable ASCII")
		}
	default:
		match := scpRE.FindStringSubmatch(raw)
		if match == nil {
			return Source{}, fmt.Errorf("repository source must use HTTPS or SSH")
		}
		host, repoPath, transport = match[2], match[3], "ssh"
	}
	if !hostRE.MatchString(host) || !validRepositoryPath(repoPath, transport == "ssh") {
		return Source{}, fmt.Errorf("repository host or path is invalid")
	}
	repoPath = strings.TrimSuffix(repoPath, ".git")
	if repoPath == "" {
		return Source{}, fmt.Errorf("repository path is empty after canonicalization")
	}
	identity := strings.ToLower(host) + "/" + repoPath
	if utf8.RuneCountInString(identity) > 4096 {
		return Source{}, fmt.Errorf("canonical repository identity exceeds 4096 Unicode scalars")
	}
	return Source{Git: raw, Identity: identity, Transport: transport}, nil
}

func containsRepositoryWhitespaceOrControl(value string) bool {
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func validRepositoryPath(value string, asciiOnly bool) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.Contains(value, "//") {
		return false
	}
	if asciiOnly && !sshPathRE.MatchString(value) {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || component == "." || component == ".." || strings.ContainsAny(component, ":\x00\r\n\t ") {
			return false
		}
	}
	return true
}

// ValidRefName reports whether value is in the closed portable Git ref grammar.
func ValidRefName(value string) bool {
	if value == "" || len([]byte(value)) > 255 || !utf8.ValidString(value) || value == "@" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.HasSuffix(value, ".") || strings.Contains(value, "//") || strings.Contains(value, "..") || strings.Contains(value, "@{") {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r == 0x7f || strings.ContainsRune("~^:?*[\\", r) {
			return false
		}
	}
	for _, component := range strings.Split(value, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}

// LoadDescriptor reads and validates skill-build.json below repositoryRoot.
func LoadDescriptor(repositoryRoot string) (*Descriptor, error) {
	payload, err := os.ReadFile(filepath.Join(repositoryRoot, DescriptorName)) // #nosec G304 -- fixed name below caller-owned root
	if err != nil {
		return nil, err
	}
	return ParseDescriptor(payload)
}

// ParseDescriptor parses the closed canonical skill-build.json schema.
func ParseDescriptor(payload []byte) (*Descriptor, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("malformed JSON in %s: %w", DescriptorName, err)
	}
	var raw any
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New(DescriptorName, "must contain an object")
	}
	if err := rejectUnknown(obj, DescriptorName, "schema_version", "targets"); err != nil {
		return nil, err
	}
	if number, ok := obj["schema_version"].(json.Number); !ok || string(number) != "1" {
		return nil, verr.New("schema_version", "%s requires integer schema 1", DescriptorName)
	}
	rawTargets, ok := obj["targets"].(map[string]any)
	if !ok || len(rawTargets) == 0 {
		return nil, verr.New("targets", "must be a non-empty object")
	}
	targets := make(map[string]Target, len(rawTargets))
	names := sortedKeys(rawTargets)
	for _, name := range names {
		label := "targets." + name
		if !identifiers.Valid(name) {
			return nil, verr.New(label, "target name %s", identifiers.Rule)
		}
		entry, ok := rawTargets[name].(map[string]any)
		if !ok {
			return nil, verr.New(label, "must be an object")
		}
		if err := rejectUnknown(entry, label, "driver", "build_root", "source_dir"); err != nil {
			return nil, err
		}
		driver, _ := entry["driver"].(string)
		if driver != "go-repository-v1" {
			return nil, verr.New(label+".driver", "must be 'go-repository-v1'")
		}
		buildRoot, buildOK := rootOrPath(entry["build_root"])
		sourceDir, sourceOK := rootOrPath(entry["source_dir"])
		if !buildOK {
			return nil, verr.New(label+".build_root", "must be '.' or a portable relative path")
		}
		if !sourceOK {
			return nil, verr.New(label+".source_dir", "must be '.' or a portable relative path")
		}
		if buildRoot != "." && sourceDir != buildRoot && !strings.HasPrefix(sourceDir, buildRoot+"/") {
			return nil, verr.New(label+".source_dir", "must be contained by build_root")
		}
		targets[name] = Target{Name: name, Driver: driver, BuildRoot: buildRoot, SourceDir: sourceDir}
	}
	return &Descriptor{SchemaVersion: 1, Targets: targets}, nil
}

func rootOrPath(raw any) (string, bool) {
	value, ok := raw.(string)
	return value, ok && (value == "." || identifiers.PortablePath(value))
}

// NormalizeLocalSelector applies the project-relative selector algorithm
// without exposing or resolving an absolute host path.
func NormalizeLocalSelector(selector string) (string, error) {
	selector = filepath.ToSlash(selector)
	if selector == "" || strings.HasPrefix(selector, "/") || strings.HasSuffix(selector, "/") || strings.Contains(selector, "//") {
		return "", fmt.Errorf("local selector must be a non-empty project-relative selector")
	}
	parts := make([]string, 0)
	for _, component := range strings.Split(selector, "/") {
		switch component {
		case ".":
			continue
		case "..":
			if len(parts) > 0 && parts[len(parts)-1] != ".." {
				parts = parts[:len(parts)-1]
			} else {
				parts = append(parts, component)
			}
		default:
			if component == "" || strings.ContainsRune(component, '\x00') {
				return "", fmt.Errorf("local selector contains an empty or invalid component")
			}
			parts = append(parts, component)
		}
	}
	if len(parts) == 0 {
		return ".", nil
	}
	return strings.Join(parts, "/"), nil
}

// LocalIdentity derives a host-path-free identity for a local substitution.
func LocalIdentity(projectIdentity, selector string) (string, error) {
	normalized, err := NormalizeLocalSelector(selector)
	if err != nil {
		return "", err
	}
	payload, err := registry.CanonicalBytesChecked(map[string]any{
		"algorithm": "curator-operator-local-git-v1",
		"project":   projectIdentity,
		"selector":  normalized,
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func rejectUnknown(obj map[string]any, field string, allowed ...string) error {
	set := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		set[key] = true
	}
	var unknown []string
	for key := range obj {
		if !set[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return verr.New(field, "unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	return nil
}

func sortedKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
