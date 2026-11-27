package envprofile

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/toolprov"
)

const (
	// CommandRootReceiptName is the closed Slice B-a receipt written inside a
	// profile command root bin directory.
	CommandRootReceiptName = ".curator-profile-command-root.json"
	// CommandRootReceiptSchema is the closed schema tag of CommandRootReceipt.
	CommandRootReceiptSchema = "profile-command-root/1"
	// CommandRootMarkerName is the sibling generation marker file.
	CommandRootMarkerName = ".curator-command-generation.json"
)

// CommandRootOptions customizes command root materialization (primarily for tests).
type CommandRootOptions struct {
	Catalog toolprov.CatalogOptions
	Force   bool // force new generation even when current root matches lock
}

// CommandRootEntry describes one materialized executable inside bin/.
type CommandRootEntry struct {
	Name          string `json:"name"`
	Kind          string `json:"kind"` // "build-go-v1", "script", "tool-aqua", "tool-pypi-uv"
	ProviderSkill string `json:"provider_skill"`
	SourceID      string `json:"source_id,omitempty"`
	SHA256        string `json:"sha256"`
}

// CommandRootReceipt is the closed .curator-profile-command-root.json record.
type CommandRootReceipt struct {
	Schema          string             `json:"schema"`
	Profile         string             `json:"profile"`
	LockSHA256      string             `json:"lock_sha256"`
	PlanSHA256      string             `json:"plan_sha256"`
	Generation      string             `json:"generation"`
	Platform        string             `json:"platform"`
	Root            string             `json:"root"`
	Commands        []CommandRootEntry `json:"commands"`
	InventorySHA256 string             `json:"inventory_sha256"`
}

// CommandRootMarker is the sibling .curator-command-generation.json record.
type CommandRootMarker struct {
	Schema          string `json:"schema"`
	Profile         string `json:"profile"`
	LockSHA256      string `json:"lock_sha256"`
	PlanSHA256      string `json:"plan_sha256"`
	Generation      string `json:"generation"`
	Platform        string `json:"platform"`
	Root            string `json:"root"`
	ReceiptSHA256   string `json:"receipt_sha256"`
	InventorySHA256 string `json:"inventory_sha256"`
}

// ActiveCommandRoot is persisted in current-<platform>.json for fast lookup.
type ActiveCommandRoot struct {
	Root        string                  `json:"root"`
	CommandRoot envfragment.CommandRoot `json:"command_root"`
}

type plannedExport struct {
	name          string
	kind          string // "build", "script", "tool"
	providerSkill string
	skillRoot     string
	spec          *skillspec.Spec
	cmd           skillspec.Command
	tool          skillspec.ToolDependency
}

func commandRootCurrentPointer(home, profile, platform string) string {
	return filepath.Join(home, "environments", profile, "commands", "current-"+platform+".json")
}

// ProfileHasExecutableExports reports whether any active skill in the profile
// lock declares exported commands or tool dependencies.
func ProfileHasExecutableExports(home, profile string, policy Policy) (bool, error) {
	lock, _, err := readLock(home, profile)
	if err != nil {
		return false, err
	}
	manager := newGitManager(home).withPolicy(policy)
	exports, err := collectProfileExports(home, manager, lock)
	if err != nil {
		return false, err
	}
	return len(exports) > 0, nil
}

// MaterializeCommandRoot builds and provisions all exported commands (go-v1,
// script) and tool dependencies (CIP-0007 v1) for the installed profile lock
// into a private Slice B-a generation command root under
// <home>/environments/<profile>/commands/<generation>/<platform>/bin.
//
// Profiles with no exported commands or tools return (nil, nil) and leave
// fragment emission on its existing v2/v4/v7 format.
func MaterializeCommandRoot(ctx context.Context, home, profile string, policy Policy, opts CommandRootOptions) (*ActiveCommandRoot, error) {
	lock, lockHash, err := readLock(home, profile)
	if err != nil {
		return nil, fmt.Errorf("read profile lock for command root: %w", err)
	}
	lockHex := strings.TrimPrefix(lockHash, "sha256:")
	platform := opts.Catalog.Platform
	if platform == "" {
		platform = toolprov.CurrentPlatform()
		opts.Catalog.Platform = platform
	}

	manager := newGitManager(home).withPolicy(policy)
	exports, err := collectProfileExports(home, manager, lock)
	if err != nil {
		return nil, err
	}
	pointerPath := commandRootCurrentPointer(home, profile, platform)
	if len(exports) == 0 {
		_ = os.Remove(pointerPath)
		return nil, nil
	}

	// Reuse current verified root if lock matches and not forced.
	if !opts.Force {
		if existing, rootPath, vErr := LoadVerifiedCommandRootForPlatform(home, profile, lockHex, platform); vErr == nil && existing != nil {
			return &ActiveCommandRoot{Root: rootPath, CommandRoot: *existing}, nil
		}
	}

	planHex, err := computeCommandPlanSHA256(profile, lockHex, platform, exports)
	if err != nil {
		return nil, err
	}
	generation, err := newGenerationID()
	if err != nil {
		return nil, err
	}

	platformDir := filepath.Join(home, "environments", profile, "commands", generation, platform)
	binDir := filepath.Join(platformDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		return nil, fmt.Errorf("create profile command root: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(filepath.Join(home, "environments", profile, "commands", generation))
		}
	}()

	// Establish a single go-v1 session lazily only if at least one build command exists.
	var goSession *godriver.Session
	var goPrivateBase string
	defer func() {
		if goSession != nil {
			_ = goSession.Release()
		}
		if goPrivateBase != "" {
			_ = os.RemoveAll(goPrivateBase)
		}
	}()
	ensureGoSession := func() (*godriver.Session, error) {
		if goSession != nil {
			return goSession, nil
		}
		base, err := os.MkdirTemp("", "curator-profile-go-*")
		if err != nil {
			return nil, fmt.Errorf("create private go build base: %w", err)
		}
		goPrivateBase = base
		cfg, err := toolprov.ResolveGodriverConfig(ctx, home, base, home)
		if err != nil {
			return nil, err
		}
		sess, err := godriver.Establish(ctx, cfg)
		if err != nil {
			return nil, err
		}
		goSession = sess
		return goSession, nil
	}

	entries := make([]CommandRootEntry, 0, len(exports))
	for _, exp := range exports {
		destPath := filepath.Join(binDir, exp.name)
		switch exp.kind {
		case "build":
			sess, err := ensureGoSession()
			if err != nil {
				return nil, fmt.Errorf("establish go-v1 toolchain for %s.%s: %w", exp.providerSkill, exp.name, err)
			}
			buildRoot, err := findContainingBuildRoot(exp.spec.BuildRoots, exp.cmd.SourceDir)
			if err != nil {
				return nil, fmt.Errorf("command %s.%s: %w", exp.providerSkill, exp.name, err)
			}
			cmdObj := godriver.BuildCommand{
				"type":       "build",
				"driver":     exp.cmd.Driver,
				"source_dir": exp.cmd.SourceDir,
			}
			if len(exp.cmd.Modules) > 0 {
				mods := make([]any, len(exp.cmd.Modules))
				for i, m := range exp.cmd.Modules {
					mods[i] = m
				}
				cmdObj["modules"] = mods
			}
			var buildArtifact string
			if err := buildsource.WithValidated(exp.skillRoot, func(tok *buildsource.Token) error {
				res, bErr := godriver.Build(ctx, godriver.BuildRequest{
					Session:       sess,
					Source:        tok,
					CommandObject: cmdObj,
					BuildRoot:     buildRoot,
					SourceDir:     exp.cmd.SourceDir,
					Command:       exp.name,
					Modules:       exp.cmd.Modules,
					BuildRoots:    exp.spec.BuildRoots,
					RuntimeRoots:  exp.spec.RuntimeRoots,
				})
				if bErr != nil {
					return bErr
				}
				buildArtifact = res.Artifact.StagedPath
				return nil
			}); err != nil {
				return nil, fmt.Errorf("build %s.%s: %w", exp.providerSkill, exp.name, err)
			}
			sha, err := copyPrivateExecutable(buildArtifact, destPath)
			if err != nil {
				return nil, fmt.Errorf("install built command %s: %w", exp.name, err)
			}
			entries = append(entries, CommandRootEntry{
				Name:          exp.name,
				Kind:          "build-go-v1",
				ProviderSkill: exp.providerSkill,
				SourceID:      exp.cmd.SourceDir,
				SHA256:        sha,
			})

		case "script":
			scriptRel := exp.cmd.UnixPath
			if runtime.GOOS == "windows" && exp.cmd.WinPath != "" {
				scriptRel = exp.cmd.WinPath
			}
			if scriptRel == "" {
				return nil, fmt.Errorf("command %s.%s has no script path for %s", exp.providerSkill, exp.name, runtime.GOOS)
			}
			runtimeDst := filepath.Join(platformDir, "runtime", exp.providerSkill)
			if err := copySkillRuntime(exp.skillRoot, runtimeDst, exp.spec.RuntimeRoots, scriptRel); err != nil {
				return nil, fmt.Errorf("stage script runtime for %s.%s: %w", exp.providerSkill, exp.name, err)
			}
			targetScript := filepath.Join(runtimeDst, filepath.FromSlash(scriptRel))
			// #nosec G302 -- owner-only executable script in private command root
			if err := os.Chmod(targetScript, 0o500); err != nil {
				return nil, err
			}
			sha, err := writePOSIXForwarder(destPath, targetScript)
			if err != nil {
				return nil, fmt.Errorf("write script forwarder %s: %w", exp.name, err)
			}
			entries = append(entries, CommandRootEntry{
				Name:          exp.name,
				Kind:          "script",
				ProviderSkill: exp.providerSkill,
				SourceID:      scriptRel,
				SHA256:        sha,
			})

		case "tool":
			prov, err := toolprov.ProvisionTool(ctx, home, exp.name, exp.tool.ID, exp.tool.Version, opts.Catalog)
			if err != nil {
				return nil, fmt.Errorf("provision tool %s (%s) for skill %s: %w", exp.name, exp.tool.ID, exp.providerSkill, err)
			}
			var sha string
			if prov.NeedsForwarder {
				sha, err = writePOSIXForwarder(destPath, prov.ExecutablePath)
			} else {
				sha, err = copyPrivateExecutable(prov.ExecutablePath, destPath)
			}
			if err != nil {
				return nil, fmt.Errorf("install provisioned tool %s: %w", exp.name, err)
			}
			entries = append(entries, CommandRootEntry{
				Name:          exp.name,
				Kind:          "tool-" + prov.Backend,
				ProviderSkill: exp.providerSkill,
				SourceID:      prov.ID + "@" + prov.Version,
				SHA256:        sha,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	invBytes, err := marshalCanonicalValue(entries)
	if err != nil {
		return nil, err
	}
	invSum := sha256.Sum256(invBytes)
	invHex := hex.EncodeToString(invSum[:])

	receipt := CommandRootReceipt{
		Schema:          CommandRootReceiptSchema,
		Profile:         profile,
		LockSHA256:      lockHex,
		PlanSHA256:      planHex,
		Generation:      generation,
		Platform:        platform,
		Root:            binDir,
		Commands:        entries,
		InventorySHA256: invHex,
	}
	receiptBytes, err := marshalCanonicalValue(receipt)
	if err != nil {
		return nil, err
	}
	receiptBytes = append(receiptBytes, '\n')
	receiptPath := filepath.Join(binDir, CommandRootReceiptName)
	if err := os.WriteFile(receiptPath, receiptBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write command root receipt: %w", err)
	}
	receiptSum := sha256.Sum256(receiptBytes)
	receiptHex := hex.EncodeToString(receiptSum[:])

	markerDoc := CommandRootMarker{
		Schema:          "profile-command-generation/1",
		Profile:         profile,
		LockSHA256:      lockHex,
		PlanSHA256:      planHex,
		Generation:      generation,
		Platform:        platform,
		Root:            binDir,
		ReceiptSHA256:   receiptHex,
		InventorySHA256: invHex,
	}
	markerBytes, err := marshalCanonicalValue(markerDoc)
	if err != nil {
		return nil, err
	}
	markerBytes = append(markerBytes, '\n')
	markerPath := filepath.Join(platformDir, CommandRootMarkerName)
	if err := os.WriteFile(markerPath, markerBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write command generation marker: %w", err)
	}
	markerSum := sha256.Sum256(markerBytes)
	markerHex := hex.EncodeToString(markerSum[:])

	active := &ActiveCommandRoot{
		Root: binDir,
		CommandRoot: envfragment.CommandRoot{
			LockSHA256:      lockHex,
			PlanSHA256:      planHex,
			Generation:      generation,
			Platform:        platform,
			ReceiptSHA256:   receiptHex,
			MarkerSHA256:    markerHex,
			InventorySHA256: invHex,
		},
	}
	if err := writeJSONAtomic(pointerPath, active, 0o600); err != nil {
		return nil, fmt.Errorf("write active command root pointer: %w", err)
	}
	cleanup = false
	return active, nil
}

// LoadVerifiedCommandRoot loads and verifies the current profile command root
// for the host platform if it matches expectedLockHex. Returns (nil, "", nil)
// when no active command root is registered or when the pointer is stale.
func LoadVerifiedCommandRoot(home, profile, expectedLockHex string) (*envfragment.CommandRoot, string, error) {
	return LoadVerifiedCommandRootForPlatform(home, profile, expectedLockHex, toolprov.CurrentPlatform())
}

// LoadVerifiedCommandRootForPlatform loads and verifies the current profile
// command root for the specified platform.
func LoadVerifiedCommandRootForPlatform(home, profile, expectedLockHex, platform string) (*envfragment.CommandRoot, string, error) {
	pointerPath := commandRootCurrentPointer(home, profile, platform)
	raw, err := os.ReadFile(pointerPath) // #nosec G304 -- manager-owned environments tree
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", nil
		}
		return nil, "", err
	}
	var active ActiveCommandRoot
	if err := json.Unmarshal(raw, &active); err != nil {
		return nil, "", fmt.Errorf("command_root_invalid: parse active pointer: %w", err)
	}
	if active.CommandRoot.LockSHA256 != expectedLockHex || active.CommandRoot.Platform != platform {
		return nil, "", nil
	}
	expectedBin := filepath.Join(filepath.Clean(home), "environments", profile, "commands", active.CommandRoot.Generation, platform, "bin")
	if active.Root != expectedBin {
		return nil, "", fmt.Errorf("command_root_invalid: unexpected root path %q", active.Root)
	}
	info, err := os.Lstat(active.Root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, "", fmt.Errorf("command_root_invalid: root %q is not a real directory", active.Root)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		return nil, "", fmt.Errorf("command_root_invalid: root %q has group/other write bits", active.Root)
	}

	receiptPath := filepath.Join(active.Root, CommandRootReceiptName)
	receiptBytes, err := os.ReadFile(receiptPath) // #nosec G304 -- verified command root directory
	if err != nil {
		return nil, "", fmt.Errorf("command_root_invalid: read receipt: %w", err)
	}
	rSum := sha256.Sum256(receiptBytes)
	if hex.EncodeToString(rSum[:]) != active.CommandRoot.ReceiptSHA256 {
		return nil, "", fmt.Errorf("command_root_invalid: receipt digest mismatch")
	}
	var receipt CommandRootReceipt
	if err := json.Unmarshal(receiptBytes, &receipt); err != nil {
		return nil, "", fmt.Errorf("command_root_invalid: decode receipt: %w", err)
	}
	if receipt.Schema != CommandRootReceiptSchema || receipt.Profile != profile || receipt.LockSHA256 != expectedLockHex || receipt.Generation != active.CommandRoot.Generation || receipt.Platform != platform {
		return nil, "", fmt.Errorf("command_root_invalid: receipt identity mismatch")
	}

	markerPath := filepath.Join(filepath.Dir(active.Root), CommandRootMarkerName)
	markerBytes, err := os.ReadFile(markerPath) // #nosec G304 -- verified generation directory
	if err != nil {
		return nil, "", fmt.Errorf("command_root_invalid: read marker: %w", err)
	}
	mSum := sha256.Sum256(markerBytes)
	if hex.EncodeToString(mSum[:]) != active.CommandRoot.MarkerSHA256 {
		return nil, "", fmt.Errorf("command_root_invalid: marker digest mismatch")
	}

	// Verify exact bin inventory: every executable matches its SHA-256 and no
	// undeclared entries exist in bin/.
	expectedCmds := make(map[string]string, len(receipt.Commands))
	for _, c := range receipt.Commands {
		expectedCmds[c.Name] = c.SHA256
	}
	dirEntries, err := os.ReadDir(active.Root)
	if err != nil {
		return nil, "", fmt.Errorf("command_root_invalid: list bin: %w", err)
	}
	seen := 0
	for _, de := range dirEntries {
		name := de.Name()
		if name == CommandRootReceiptName {
			continue
		}
		wantSHA, ok := expectedCmds[name]
		if !ok {
			return nil, "", fmt.Errorf("command_root_invalid: undeclared entry %q in command root", name)
		}
		fullPath := filepath.Join(active.Root, name)
		fi, err := os.Lstat(fullPath)
		if err != nil || !fi.Mode().IsRegular() || fi.Mode()&os.ModeSymlink != 0 {
			return nil, "", fmt.Errorf("command_root_invalid: entry %q is not a regular file", name)
		}
		gotSHA, err := toolprov.FileSHA256(fullPath)
		if err != nil || gotSHA != wantSHA {
			return nil, "", fmt.Errorf("command_root_invalid: entry %q digest mismatch", name)
		}
		seen++
	}
	if seen != len(expectedCmds) {
		return nil, "", fmt.Errorf("command_root_invalid: missing declared commands in root")
	}

	cr := active.CommandRoot
	return &cr, active.Root, nil
}

func collectProfileExports(home string, manager *gitManager, lock *contextlock.Lock) ([]plannedExport, error) {
	seen := map[string]string{}
	var exports []plannedExport
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		entry := manager.entryPath(home, resolvedOf(member))
		root := packageRoot(entry, member.Directory)
		spec, err := skillspec.Load(root)
		if err != nil {
			return nil, fmt.Errorf("load skill %q manifest: %w", member.Name, err)
		}
		cmdNames := make([]string, 0, len(spec.Commands))
		for name := range spec.Commands {
			cmdNames = append(cmdNames, name)
		}
		sort.Strings(cmdNames)
		for _, name := range cmdNames {
			cmd := spec.Commands[name]
			if cmd.Type == "system" {
				continue // system commands are host prerequisites, not exported binaries
			}
			if err := checkExportName(name); err != nil {
				return nil, fmt.Errorf("skill %q command %q: %w", member.Name, name, err)
			}
			if prev, dup := seen[name]; dup {
				return nil, fmt.Errorf("profile_command_collision: command %q claimed by both %q and %q", name, prev, member.Name)
			}
			switch cmd.Type {
			case "build":
				if cmd.Driver != "go-v1" {
					return nil, fmt.Errorf("profile_command_unsupported: skill %q command %q uses unsupported build driver %q", member.Name, name, cmd.Driver)
				}
				seen[name] = member.Name
				exports = append(exports, plannedExport{
					name: name, kind: "build", providerSkill: member.Name, skillRoot: root, spec: spec, cmd: cmd,
				})
			case "script":
				seen[name] = member.Name
				exports = append(exports, plannedExport{
					name: name, kind: "script", providerSkill: member.Name, skillRoot: root, spec: spec, cmd: cmd,
				})
			}
		}

		toolNames := make([]string, 0, len(spec.Tools))
		for alias := range spec.Tools {
			toolNames = append(toolNames, alias)
		}
		sort.Strings(toolNames)
		for _, alias := range toolNames {
			dep := spec.Tools[alias]
			if err := checkExportName(alias); err != nil {
				return nil, fmt.Errorf("skill %q tool %q: %w", member.Name, alias, err)
			}
			if prev, dup := seen[alias]; dup {
				return nil, fmt.Errorf("profile_command_collision: command %q claimed by both %q and %q", alias, prev, member.Name)
			}
			seen[alias] = member.Name
			exports = append(exports, plannedExport{
				name: alias, kind: "tool", providerSkill: member.Name, skillRoot: root, spec: spec, tool: dep,
			})
		}
	}
	sort.Slice(exports, func(i, j int) bool { return exports[i].name < exports[j].name })
	return exports, nil
}

func checkExportName(name string) error {
	if name == "" || len(name) > 63 {
		return fmt.Errorf("profile_command_invalid: name %q length out of bounds", name)
	}
	if strings.HasPrefix(strings.ToLower(name), "curator-") {
		return fmt.Errorf("profile_command_reserved: name %q uses reserved curator- prefix", name)
	}
	for i, ch := range name {
		if i == 0 && (ch < 'a' || ch > 'z') {
			return fmt.Errorf("profile_command_invalid: name %q must begin with lowercase ASCII letter", name)
		}
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') && ch != '-' && ch != '_' {
			return fmt.Errorf("profile_command_invalid: name %q contains disallowed character %q", name, ch)
		}
	}
	return nil
}

func computeCommandPlanSHA256(profile, lockHex, platform string, exports []plannedExport) (string, error) {
	items := make([]map[string]string, 0, len(exports))
	for _, e := range exports {
		item := map[string]string{
			"name":     e.name,
			"kind":     e.kind,
			"provider": e.providerSkill,
		}
		if e.kind == "tool" {
			item["id"] = e.tool.ID
			item["version"] = e.tool.Version
		}
		items = append(items, item)
	}
	raw, err := marshalCanonicalValue(map[string]any{
		"profile":     profile,
		"lock_sha256": lockHex,
		"platform":    platform,
		"exports":     items,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func newGenerationID() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate command root id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

func findContainingBuildRoot(buildRoots []string, sourceDir string) (string, error) {
	for _, br := range buildRoots {
		if sourceDir == br || strings.HasPrefix(sourceDir, br+"/") {
			return br, nil
		}
	}
	return "", fmt.Errorf("source_dir %q is not contained in any declared build_roots", sourceDir)
}

func copyPrivateExecutable(src, dst string) (string, error) {
	in, err := os.Open(src) // #nosec G304 -- validated build artifact or sealed store binary
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700) // #nosec G302,G304 -- owner-only private executable in command root bin
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, h), in)
	closeErr := out.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	// #nosec G302 -- owner-only executable binary in private command root
	if err := os.Chmod(dst, 0o500); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writePOSIXForwarder(dst, targetAbsPath string) (string, error) {
	escaped := strings.ReplaceAll(targetAbsPath, "'", `'\''`)
	body := fmt.Sprintf("#!/bin/sh\nexec '%s' \"$@\"\n", escaped)
	if err := os.WriteFile(dst, []byte(body), 0o700); err != nil {
		return "", err
	}
	// #nosec G302 -- owner-only executable forwarder in private command root
	if err := os.Chmod(dst, 0o500); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:]), nil
}

func copySkillRuntime(skillRoot, runtimeDst string, runtimeRoots []string, scriptRel string) error {
	roots := runtimeRoots
	if len(roots) == 0 {
		roots = []string{filepath.ToSlash(filepath.Dir(scriptRel))}
	}
	for _, r := range roots {
		srcRoot := filepath.Join(skillRoot, filepath.FromSlash(r))
		dstRoot := filepath.Join(runtimeDst, filepath.FromSlash(r))
		if err := copyDirTree(srcRoot, dstRoot); err != nil {
			return err
		}
	}
	return nil
}

func copyDirTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not allowed in script runtime: %s", rel)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		in, err := os.Open(path) // #nosec G304 -- validated snapshot subtree
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		mode := os.FileMode(0o600)
		if info.Mode()&0o111 != 0 {
			mode = 0o700
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode) // #nosec G304 -- private runtime dir
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func marshalCanonicalValue(v any) ([]byte, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))
	dec.UseNumber()
	var domain any
	if err := dec.Decode(&domain); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(domain)
}

func writeJSONAtomic(path string, v any, mode os.FileMode) error {
	raw, err := marshalCanonicalValue(v)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}
