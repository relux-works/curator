package projectadapter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/privatedir"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const MuseVersion = "1.4.11-R6667.1"
const ModeStrict = "strict"
const ModeNativeProject = "native-project"

// MuseCertification reports full strict surfaces, not individual loader
// observations. The measured exec foreign-context flag does not qualify strict
// exclusion, permissions or memory-off. No version/platform is launchable.
func MuseCertification(version, platform, mode, form string) Certification {
	report := Certification{Environment: envregistry.Muse, ProviderVersion: version, Platform: platform, Mode: mode, Form: form, Certified: []SurfaceName{}, Refused: []Refusal{}}
	if mode == ModeNativeProject {
		report.Notice = "project sources not isolated"
	}
	for _, surface := range []SurfaceName{Commands, Context, MCP, MemoryOff, NativeSourceExclusion, Permissions, Skills} {
		reason := "exact_native_tuple_unqualified"
		if version == MuseVersion && platform == "macos" {
			reason = "strict_surface_unqualified"
			if surface == NativeSourceExclusion {
				reason = "complete_project_source_exclusion_unqualified"
			}
			if surface == MCP {
				reason = "native_writer_unqualified"
			}
			if surface == Permissions {
				reason = "manager_permissions_unqualified"
			}
			if surface == MemoryOff {
				reason = "memory_off_unqualified"
			}
		}
		report.Refused = append(report.Refused, Refusal{surface, reason})
	}
	if form != "exec" {
		report.Refused = append(report.Refused, Refusal{LaunchForm, "interactive_and_resume_unqualified"})
	}
	if mode != ModeStrict && mode != ModeNativeProject {
		report.Refused = append(report.Refused, Refusal{LaunchForm, "unknown_composition_mode"})
	}
	return report
}

// MuseLaunch contributes nothing on refusal, including no XDG variables,
// trust/exclusion flags, legacy fragment or fallback route.
func MuseLaunch(version, platform, mode, form string) (LaunchContribution, error) {
	report := MuseCertification(version, platform, mode, form)
	if mode != ModeStrict && mode != ModeNativeProject {
		return LaunchContribution{}, &Unsupported{LaunchForm, "unknown_composition_mode"}
	}
	if form != "exec" {
		return LaunchContribution{}, &Unsupported{LaunchForm, "interactive_and_resume_unqualified"}
	}
	if mode == ModeNativeProject {
		return LaunchContribution{}, &Unsupported{Permissions, "manager_permissions_unqualified; project sources not isolated"}
	}
	for _, required := range []SurfaceName{NativeSourceExclusion, Permissions, MemoryOff} {
		for _, refusal := range report.Refused {
			if refusal.Surface == required {
				return LaunchContribution{}, &Unsupported{required, refusal.Reason}
			}
		}
	}
	// No successful path exists until exact native qualification is added.
	return LaunchContribution{}, &Unsupported{NativeSourceExclusion, "strict_surface_unqualified"}
}

type MuseMCPItem struct {
	ItemID   string
	Required bool
}
type Omission struct{ ItemID, Reason string }

// MuseOfflineInput must come from retained I-COMP data after protected
// authority resolution. This API is a byte copier, not an approval interface.
// PolicyRecord is retained verbatim as inert manager data; it is never encoded
// as invented native settings. Memory describes that manager plan only.
type MuseOfflineInput struct {
	Files        map[string][]byte
	Generated    map[string]Surface
	PolicyRecord []byte
	PolicySHA256 string
	Memory       string
	MCPItems     []MuseMCPItem
}

// MuseOfflineHome is deliberately separate from envfragment.Fragment and
// LaunchContribution. Parents describes layout, never an admitted child env.
type MuseOfflineHome struct {
	Home         string
	Parents      map[string]string
	Generated    map[string]Surface
	PolicyRecord []byte
	Memory       string
	Omitted      []Omission
	Mode         string
}

func sum(b []byte) string              { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func offlineError(reason string) error { return fmt.Errorf("project_context_drift: %s", reason) }

func musePath(p string) bool {
	if !utf8.ValidString(p) || !fs.ValidPath(p) || len(p) > 240 || strings.ContainsAny(p, "\\:\x00") || strings.Count(p, "/") >= 32 {
		return false
	}
	if p == "config/muse/AGENTS.md" {
		return true
	}
	parts := strings.Split(p, "/")
	if len(parts) == 3 && parts[0] == ".agent-context" && parts[1] == "commands" && identifiers.Valid(parts[2]) {
		return true
	}
	return len(parts) >= 5 && parts[0] == "data" && parts[1] == "muse" && parts[2] == "skills" && identifiers.Valid(parts[3])
}

func museInput(input MuseOfflineInput) (MuseOfflineInput, []Omission, error) {
	if input.Memory != "off" || !json.Valid(input.PolicyRecord) || sum(input.PolicyRecord) != input.PolicySHA256 {
		return input, nil, offlineError("manager policy or memory differs")
	}
	if len(input.Files) == 0 || len(input.Files) > 10000 || len(input.Files) != len(input.Generated) {
		return input, nil, offlineError("complete receipt required")
	}
	out := input
	out.Files = map[string][]byte{}
	out.Generated = map[string]Surface{}
	out.PolicyRecord = bytes.Clone(input.PolicyRecord)
	var total int64
	aliases := map[string]bool{}
	for p, b := range input.Files {
		if !musePath(p) {
			return out, nil, &Unsupported{Context, "unapproved_control_path"}
		}
		s, ok := input.Generated[p]
		if !ok || s.SHA256 != sum(b) || s.Size != int64(len(b)) || (s.Mode != 0600 && s.Mode != 0700) || len(b) > 16<<20 {
			return out, nil, offlineError("generated bytes or metadata differ")
		}
		total += s.Size
		if total > 64<<20 {
			return out, nil, offlineError("generated byte bound exceeded")
		}
		alias := cases.Fold().String(norm.NFC.String(p))
		if aliases[alias] {
			return out, nil, offlineError("generated path alias")
		}
		aliases[alias] = true
		out.Files[p] = bytes.Clone(b)
		out.Generated[p] = s
	}
	for p := range out.Files {
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if aliases[cases.Fold().String(norm.NFC.String(parent))] {
				return out, nil, offlineError("generated file ancestor collision")
			}
		}
	}
	omitted := []Omission{}
	seen := map[string]bool{}
	for _, item := range input.MCPItems {
		if !identifiers.Valid(item.ItemID) || seen[item.ItemID] {
			return out, nil, offlineError("invalid MCP item identity")
		}
		seen[item.ItemID] = true
		if item.Required {
			return out, nil, &Unsupported{MCP, "native_writer_unqualified"}
		}
		omitted = append(omitted, Omission{item.ItemID, "unsupported"})
	}
	return out, omitted, nil
}

// MaterializeMuseOffline copies strict additive output into a NEW private
// manager allocation only. Native-project copying requires the future mode-aware
// I-COMP receipt; this function never strips or rebaselines additive output. It never
// updates an existing/live home, acquires a lease, emits a fragment, reads the
// repository/native home, starts a provider or applies native policy. I-LEASE
// remains the sole future owner of live publication and session admission.
// Interrupted offline allocations are not live homes and confer no authority.
func MaterializeMuseOffline(ctx context.Context, home string, input MuseOfflineInput) (MuseOfflineHome, error) {
	var result MuseOfflineHome
	data, omitted, err := museInput(input)
	if err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if !filepath.IsAbs(home) || filepath.Clean(home) != home {
		return result, offlineError("invalid offline allocation")
	}
	parentPath, name := filepath.Dir(home), filepath.Base(home)
	if err = privatedir.Validate(parentPath); err != nil {
		return result, offlineError("offline parent unavailable")
	}
	parent, err := os.OpenRoot(parentPath)
	if err != nil {
		return result, offlineError("offline parent unreadable")
	}
	defer parent.Close()
	if err = parent.Mkdir(name, 0700); err != nil {
		return result, offlineError("offline allocation already exists or is unavailable")
	}
	complete := false
	defer func() {
		if !complete {
			_ = parent.RemoveAll(name)
		}
	}()
	root, err := parent.OpenRoot(name)
	if err != nil {
		return result, offlineError("offline allocation unreadable")
	}
	defer root.Close()
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		return result, err
	}
	parents := map[string]string{}
	for _, pair := range adapter.HomeVariables {
		if err = root.MkdirAll(pair[1], 0700); err != nil {
			return result, offlineError("offline parent creation failed")
		}
		parents[pair[0]] = filepath.Join(home, pair[1])
	}
	paths := make([]string, 0, len(data.Files))
	for p := range data.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		if err = root.MkdirAll(filepath.FromSlash(path.Dir(p)), 0700); err != nil {
			return result, offlineError("offline directory creation failed")
		}
		f, e := root.OpenFile(filepath.FromSlash(p), os.O_CREATE|os.O_EXCL|os.O_WRONLY, fs.FileMode(data.Generated[p].Mode))
		if e != nil {
			return result, offlineError("offline file creation failed")
		}
		_, e = f.Write(data.Files[p])
		closeErr := f.Close()
		if e != nil || closeErr != nil {
			return result, offlineError("offline file write failed")
		}
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = VerifyMuseOffline(home, data.Generated); err != nil {
		return result, err
	}
	complete = true
	return MuseOfflineHome{home, parents, data.Generated, data.PolicyRecord, "off", omitted, ModeStrict}, nil
}

// VerifyMuseOffline verifies the exact offline entry set against the retained
// receipt, including sizes/modes and links. It never rebaselines changed bytes.
// This is not a live native-state verifier: mutable native state is absent.
func VerifyMuseOffline(home string, generated map[string]Surface) error {
	if len(generated) == 0 || len(generated) > 10000 {
		return offlineError("complete receipt required")
	}
	if err := privatedir.Validate(home); err != nil {
		return offlineError("offline allocation unavailable")
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return offlineError("offline allocation unreadable")
	}
	defer root.Close()
	seen := 0
	err = fs.WalkDir(root.FS(), ".", func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return offlineError("offline entry unreadable")
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return offlineError("offline link detected")
		}
		if entry.IsDir() {
			return nil
		}
		s, ok := generated[p]
		if !ok || !musePath(p) || s.Size < 0 || s.Size > 16<<20 || (s.Mode != 0600 && s.Mode != 0700) {
			return offlineError("offline entry set differs")
		}
		info, e := entry.Info()
		if e != nil || !info.Mode().IsRegular() || info.Size() != s.Size || uint32(info.Mode().Perm()) != s.Mode {
			return offlineError("offline metadata differs")
		}
		f, e := root.Open(filepath.FromSlash(p))
		if e != nil {
			return offlineError("offline entry unreadable")
		}
		b, e := io.ReadAll(io.LimitReader(f, (16<<20)+1))
		closeErr := f.Close()
		if e != nil || closeErr != nil || sum(b) != s.SHA256 {
			return offlineError("offline bytes differ")
		}
		seen++
		return nil
	})
	if err != nil {
		return err
	}
	if seen != len(generated) {
		return offlineError("offline entry set differs")
	}
	return nil
}
