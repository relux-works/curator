// Package contextaudit implements audit detectors for context and MCP
// snapshots: the scoped secret-material detector, the system-module warning,
// and the interim blocking rule for regular files containing NUL bytes.
package contextaudit

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/stateread"
)

// Classes and diagnostics.
const (
	ClassSecretMaterial      = "context-secret-material"
	ClassSystemModulePresent = "context-system-module-present"
	ClassOpaqueFile          = "opaque-file"
	PatternNULByte           = "nul-byte"
	SeverityBlocking         = "blocking"
	DiagWaiverApplied        = "context_secret_waiver_applied"
	DiagWaiverUnmatched      = "context_secret_waiver_unmatched"
)

// PatternClass is one closed detector pattern (vectors/context-detectors.json
// pattern_classes). Group names the capture whose span is reported.
type PatternClass struct {
	Pattern           string
	Regexp            *regexp.Regexp
	Group             int
	PlaceholderPrefix string
}

// Classes is the closed pattern set. Every class is vectored; the set is
// not extensible by configuration.
var Classes = []PatternClass{
	{Pattern: "aws-access-key-id", Regexp: regexp.MustCompile(`(?:^|[^A-Z0-9])(AKIA[0-9A-Z]{16})(?:[^A-Z0-9]|$)`), Group: 1, PlaceholderPrefix: "AKIA"},
	{Pattern: "private-key-block", Regexp: regexp.MustCompile(`(-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY(?: BLOCK)?-----)`), Group: 1},
	{Pattern: "bearer-token", Regexp: regexp.MustCompile(`(?:^|[^A-Za-z0-9])Bearer[ \t]+([A-Za-z0-9._~+/-]{20,}=*)`), Group: 1},
}

// ScopeFiles are the root-level files inside the detector scope; every file
// below context/ is inside it too.
var ScopeFiles = []string{contextpkg.ManifestName, contextpkg.MCPManifestName, contextpkg.InformativeDoc}

// InScope reports whether a snapshot-relative portable path is inside the
// detector scope.
func InScope(path string) bool {
	if strings.HasPrefix(path, contextpkg.ContextDir+"/") {
		return true
	}
	for _, name := range ScopeFiles {
		if path == name {
			return true
		}
	}
	return false
}

// Waiver is one scoped waiver of machine configuration
// (secret_material_waivers). Pin is the member's pin as the lock spells it —
// bare hex — or as the header spells it ("commit <hex>", "state sha256:<hex>").
type Waiver struct {
	Pin         string
	HashVersion hashing.Version
	File        string
	Span        [2]int
	Reason      string
}

// Finding is one detector finding.
type Finding struct {
	Class        string
	File         string
	Pattern      string
	Severity     string
	Span         [2]int
	Waived       bool
	WaiverReason string
}

// WaiverWarning reports an applied or unmatched waiver.
type WaiverWarning struct {
	Diagnostic string
	File       string
	Span       [2]int
	Reason     string // applied
	Pin        string // unmatched, as the waiver spelled it
}

// SystemModule is one context-system-module-present warning.
type SystemModule struct {
	Package  string
	Path     string
	Selector []string // nil when the module has no selector
}

// Report is the outcome for one snapshot.
type Report struct {
	Findings      []Finding
	Waivers       []WaiverWarning
	SystemModules []SystemModule
}

// Blocking reports whether an unwaived blocking finding remains. Scoped
// secret waivers may clear only their matching secret finding; opaque-file
// findings are never waivable.
func (r Report) Blocking() bool {
	for _, finding := range r.Findings {
		if finding.Severity == SeverityBlocking && !finding.Waived {
			return true
		}
	}
	return false
}

// FirstBlocking returns the first unwaived blocking finding in report order.
func (r Report) FirstBlocking() (Finding, bool) {
	for _, finding := range r.Findings {
		if finding.Severity == SeverityBlocking && !finding.Waived {
			return finding, true
		}
	}
	return Finding{}, false
}

// PinKey normalizes a pin spelling onto the bare lowercase key.
func PinKey(pin string) string {
	pin = strings.TrimSpace(pin)
	pin = strings.TrimPrefix(pin, "commit ")
	pin = strings.TrimPrefix(pin, "state ")
	pin = strings.TrimPrefix(pin, "sha256:")
	return strings.ToLower(pin)
}

// Detect walks the snapshot at root and runs the detector over the files in
// scope. It also checks every regular file for opaque NUL bytes. pin is the
// member's pin; only waivers at that pin apply, and every waiver handed in
// that clears nothing is reported unmatched.
func Detect(root, pin string, waivers []Waiver) (Report, error) {
	files := map[string][]byte{}
	for _, name := range ScopeFiles {
		payload, err := os.ReadFile(filepath.Join(root, name)) // #nosec G304 -- snapshot root chosen by the caller
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Report{}, fmt.Errorf("read %s: %w", name, err)
		}
		files[name] = payload
	}
	contextRoot := filepath.Join(root, contextpkg.ContextDir)
	if info, err := os.Lstat(contextRoot); err == nil && info.IsDir() {
		err := filepath.WalkDir(contextRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			payload, err := os.ReadFile(path) // #nosec G304 -- walked below the snapshot root
			if err != nil {
				return err
			}
			files[filepath.ToSlash(rel)] = payload
			return nil
		})
		if err != nil {
			return Report{}, err
		}
	}
	paths, err := opaquescan.NULPaths(root)
	if err != nil {
		return Report{}, err
	}
	var opaqueFindings []Finding
	for _, path := range paths {
		opaqueFindings = append(opaqueFindings, Finding{
			Class: ClassOpaqueFile, File: path, Pattern: PatternNULByte,
			Severity: SeverityBlocking,
		})
	}
	return detectFiles(files, pin, hashing.VersionV1, waivers, opaqueFindings), nil
}

// DetectAtVersion binds state-hash waivers to the content-hash framing
// version carried by the resolved context lock. Commit pins have no framing
// version and may match only unversioned commit waivers.
func DetectAtVersion(root, pin string, hashVersion hashing.Version, waivers []Waiver) (Report, error) {
	if hashVersion == hashing.VersionV1 {
		return Detect(root, pin, waivers)
	}
	if hashVersion != hashing.VersionV2 {
		return Report{}, fmt.Errorf("unsupported context content hash version %d", hashVersion)
	}
	files := map[string][]byte{}
	for _, name := range ScopeFiles {
		read, err := stateread.ReadFile(filepath.Join(root, name))
		if err != nil {
			return Report{}, fmt.Errorf("read %s: %w", name, err)
		}
		if read.Kind == stateread.KindPresent {
			files[name] = read.Bytes
		}
	}
	contextRoot := filepath.Join(root, contextpkg.ContextDir)
	info, err := stateread.Lstat(contextRoot)
	if err != nil {
		return Report{}, err
	}
	if info.Kind == stateread.KindPresent && info.Info.IsDir() {
		err := filepath.WalkDir(contextRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			read, err := stateread.ReadFile(path)
			if err != nil {
				return err
			}
			if read.Kind == stateread.KindPresent {
				files[filepath.ToSlash(rel)] = read.Bytes
			}
			return nil
		})
		if err != nil {
			return Report{}, err
		}
	}
	return detectFiles(files, pin, hashVersion, waivers, nil), nil
}

// DetectFiles runs the secret detector over its existing scope and applies the
// opaque NUL rule to every in-memory file keyed by snapshot-relative path.
func DetectFiles(files map[string][]byte, pin string, waivers []Waiver) Report {
	return DetectFilesAtVersion(files, pin, hashing.VersionV1, waivers)
}

// DetectFilesAtVersion is the in-memory counterpart of DetectAtVersion.
func DetectFilesAtVersion(files map[string][]byte, pin string, hashVersion hashing.Version, waivers []Waiver) Report {
	var opaqueFindings []Finding
	if hashVersion == hashing.VersionV1 {
		for path, content := range files {
			if offset := bytes.IndexByte(content, 0); offset >= 0 {
				opaqueFindings = append(opaqueFindings, Finding{
					Class: ClassOpaqueFile, File: path, Pattern: PatternNULByte,
					Severity: SeverityBlocking, Span: [2]int{offset, offset + 1},
				})
			}
		}
	} else if hashVersion != hashing.VersionV2 {
		return Report{Findings: []Finding{{
			Class: ClassOpaqueFile, Pattern: PatternNULByte, Severity: SeverityBlocking,
		}}}
	}
	return detectFiles(files, pin, hashVersion, waivers, opaqueFindings)
}

func detectFiles(files map[string][]byte, pin string, hashVersion hashing.Version, waivers []Waiver, opaqueFindings []Finding) Report {
	var report Report
	paths := make([]string, 0, len(files))
	for path := range files {
		if InScope(path) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		content := files[path]
		for _, class := range Classes {
			for _, match := range class.Regexp.FindAllSubmatchIndex(content, -1) {
				start, end := match[2*class.Group], match[2*class.Group+1]
				if start < 0 {
					continue
				}
				body := string(content[start:end])
				if isPlaceholder(class, body) {
					continue
				}
				report.Findings = append(report.Findings, Finding{
					Class: ClassSecretMaterial, File: path, Pattern: class.Pattern,
					Severity: SeverityBlocking, Span: [2]int{start, end},
				})
			}
		}
	}
	report.Findings = append(report.Findings, opaqueFindings...)
	sort.SliceStable(report.Findings, func(i, j int) bool {
		a, b := report.Findings[i], report.Findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Span[0] < b.Span[0]
	})
	memberPin := PinKey(pin)
	for _, waiver := range waivers {
		applied := false
		if PinKey(waiver.Pin) == memberPin && waiverMatchesHashVersion(pin, hashVersion, waiver.HashVersion) {
			for index := range report.Findings {
				finding := &report.Findings[index]
				if finding.Class != ClassOpaqueFile && finding.File == waiver.File && finding.Span == waiver.Span && !finding.Waived {
					finding.Waived = true
					finding.WaiverReason = waiver.Reason
					applied = true
					report.Waivers = append(report.Waivers, WaiverWarning{
						Diagnostic: DiagWaiverApplied, File: waiver.File, Span: waiver.Span, Reason: waiver.Reason,
					})
					break
				}
			}
		}
		if !applied {
			report.Waivers = append(report.Waivers, WaiverWarning{
				Diagnostic: DiagWaiverUnmatched, File: waiver.File, Span: waiver.Span, Pin: waiver.Pin,
			})
		}
	}
	return report
}

func waiverMatchesHashVersion(pin string, memberVersion, waiverVersion hashing.Version) bool {
	if strings.HasPrefix(strings.TrimSpace(pin), "commit ") {
		return waiverVersion == 0
	}
	if memberVersion == 0 {
		memberVersion = hashing.VersionV1
	}
	if waiverVersion == 0 {
		waiverVersion = hashing.VersionV1
	}
	return memberVersion == waiverVersion
}

// isPlaceholder applies the closed placeholder rule: a body whose remainder
// after the class prefix is one repeated character or ends in EXAMPLE.
func isPlaceholder(class PatternClass, body string) bool {
	rest := strings.TrimPrefix(body, class.PlaceholderPrefix)
	if rest == "" {
		return false
	}
	if strings.HasSuffix(rest, "EXAMPLE") {
		return true
	}
	first := rest[0]
	for index := 1; index < len(rest); index++ {
		if rest[index] != first {
			return false
		}
	}
	return true
}

// SystemModules lists every class: system module of a context manifest as
// the always-warn surfacing class reports it.
func SystemModules(packageName string, manifest *contextpkg.Manifest) []SystemModule {
	var out []SystemModule
	if manifest == nil {
		return nil
	}
	for _, module := range manifest.Modules {
		if module.Class != "system" {
			continue
		}
		out = append(out, SystemModule{Package: packageName, Path: module.Path, Selector: module.Environments})
	}
	return out
}
