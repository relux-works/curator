package toolprov

import (
	"context"
	"fmt"
	"go/build"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
)

const (
	// PinnedGoVersion is the official Go release provisioned automatically by
	// CIP-0040 when no operator-selected or compatible host Go root is present.
	// It belongs to godriver.TestedFamilies().
	PinnedGoVersion = "1.25.5"

	// EnvDisableAutoToolchain disables automatic Go toolchain downloads when set to "1".
	EnvDisableAutoToolchain = "CURATOR_NO_AUTO_TOOLCHAIN"
)

// goReleasePin holds verified official go.dev/dl/ archive metadata.
type goReleasePin struct {
	Version  string
	Filename string
	SHA256   string
}

// Official SHA-256 digests published by https://go.dev/dl/?mode=json for go1.25.5.
var pinnedGoReleases = map[string]goReleasePin{
	"darwin-arm64": {
		Version:  PinnedGoVersion,
		Filename: "go1.25.5.darwin-arm64.tar.gz",
		SHA256:   "bed8ebe824e3d3b27e8471d1307f803fc6ab8e1d0eb7a4ae196979bd9b801dd3",
	},
	"darwin-amd64": {
		Version:  PinnedGoVersion,
		Filename: "go1.25.5.darwin-amd64.tar.gz",
		SHA256:   "b69d51bce599e5381a94ce15263ae644ec84667a5ce23d58dc2e63e2c12a9f56",
	},
	"linux-amd64": {
		Version:  PinnedGoVersion,
		Filename: "go1.25.5.linux-amd64.tar.gz",
		SHA256:   "9e9b755d63b36acf30c12a9a3fc379243714c1c6d3dd72861da637f336ebb35b",
	},
	"linux-arm64": {
		Version:  PinnedGoVersion,
		Filename: "go1.25.5.linux-arm64.tar.gz",
		SHA256:   "b00b694903d126c588c378e72d3545549935d3982635ba3f7a964c9fa23fe3b9",
	},
}

// GoOptions customizes Go toolchain provisioning (primarily for tests).
type GoOptions struct {
	Platform            string
	BaseURL             string // defaults to "https://go.dev/dl"
	HTTPClient          *http.Client
	SkipSystemSearch    bool
	DisableAutoDownload bool
	Pins                map[string]goReleasePin
}

var goVersionFamilyRe = regexp.MustCompile(`go version go(1\.[0-9]+)(?:\.[0-9]+)?`)

// EnsureGoToolchain implements CIP-0040 toolchain resolution:
//  1. If CURATOR_GO or GOROOT is explicitly set in the environment, returns ""
//     so godriver honors the operator selection strictly without fallback.
//  2. Reuses an already-verified sealed Go root under <managerHome>/tools/store/go/.
//  3. Reuses a compatible Go root in fixed trusted host locations (without
//     searching caller PATH).
//  4. Unless disabled, downloads, SHA-256-verifies, and atomically publishes
//     the pinned official Go archive into <managerHome>/tools/store/go/.
func EnsureGoToolchain(ctx context.Context, managerHome string, opts GoOptions) (string, error) {
	if os.Getenv(godriver.SelectionCuratorGo) != "" || os.Getenv(godriver.SelectionGOROOT) != "" {
		return "", nil
	}
	platform := opts.Platform
	if platform == "" {
		platform = CurrentPlatform()
	}
	pins := opts.Pins
	if pins == nil {
		pins = pinnedGoReleases
	}
	pin, ok := pins[platform]
	if !ok {
		return "", fmt.Errorf("toolchain_provisioning_unsupported_platform: no pinned Go toolchain for %s", platform)
	}

	storeDir := filepath.Join(StoreRoot(managerHome), "go", pin.Version, platform)
	if r, err := VerifySealedDir(storeDir); err == nil {
		if rec, ok := r.Executables["go"]; ok {
			return filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)), nil
		}
	}

	if !opts.SkipSystemSearch {
		if hostGo := findTrustedHostGo(ctx); hostGo != "" {
			return hostGo, nil
		}
	}

	if opts.DisableAutoDownload || os.Getenv(EnvDisableAutoToolchain) == "1" {
		return "", fmt.Errorf("toolchain_provisioning_disabled: no trusted Go toolchain found and auto-provisioning is disabled")
	}

	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://go.dev/dl"
	}
	archiveURL := baseURL + "/" + pin.Filename

	parentDir := filepath.Dir(storeDir)
	if err := os.MkdirAll(parentDir, 0o700); err != nil {
		return "", fmt.Errorf("create Go store parent: %w", err)
	}
	stageDir, err := os.MkdirTemp(parentDir, ".go-stage-*")
	if err != nil {
		return "", fmt.Errorf("create Go staging dir: %w", err)
	}
	cleanupStage := true
	defer func() {
		if cleanupStage {
			_ = os.RemoveAll(stageDir)
		}
	}()

	archivePath := filepath.Join(stageDir, pin.Filename)
	if err := DownloadAndVerify(ctx, opts.HTTPClient, archiveURL, pin.SHA256, archivePath, "toolchain_provisioning_digest_mismatch"); err != nil {
		return "", err
	}
	extractDir := filepath.Join(stageDir, "root")
	if err := SafeExtractTarGz(archivePath, extractDir); err != nil {
		return "", fmt.Errorf("toolchain_provisioning_extract_failed: %w", err)
	}
	_ = os.Remove(archivePath)

	goBinRel := "go/bin/go"
	if runtime.GOOS == "windows" {
		goBinRel = "go/bin/go.exe"
	}
	goBinPath := filepath.Join(extractDir, filepath.FromSlash(goBinRel))
	info, err := os.Lstat(goBinPath)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("toolchain_provisioning_invalid: extracted archive missing %s", goBinRel)
	}
	binSHA, err := FileSHA256(goBinPath)
	if err != nil {
		return "", fmt.Errorf("hash extracted go binary: %w", err)
	}
	if err := WriteReceipt(extractDir, Receipt{
		Schema:        ReceiptSchema,
		Backend:       "go-official",
		ID:            "go",
		Version:       pin.Version,
		Platform:      platform,
		ArchiveURL:    archiveURL,
		ArchiveSHA256: pin.SHA256,
		Executables: map[string]ExecutableRecord{
			"go": {RelativePath: goBinRel, SHA256: binSHA},
		},
	}); err != nil {
		return "", err
	}

	// Atomic publish into storeDir (remove any partial/stale unverified target first).
	_ = os.RemoveAll(storeDir)
	if err := os.Rename(extractDir, storeDir); err != nil {
		return "", fmt.Errorf("publish Go toolchain store: %w", err)
	}
	cleanupStage = false
	_ = os.RemoveAll(stageDir)
	return filepath.Join(storeDir, filepath.FromSlash(goBinRel)), nil
}

// ResolveGodriverConfig builds a godriver.Config using explicit CURATOR_GO /
// GOROOT if set, otherwise resolving or auto-provisioning a trusted Go root
// under managerHome (CIP-0040).
func ResolveGodriverConfig(ctx context.Context, managerHome, privateBase string, forbiddenRoots ...string) (godriver.Config, error) {
	cfg := godriver.ConfigFromEnvironment(privateBase, forbiddenRoots...)
	if cfg.CuratorGo != "" || cfg.GOROOT != "" {
		return cfg, nil
	}
	if managerHome == "" {
		return cfg, nil
	}
	goPath, err := EnsureGoToolchain(ctx, managerHome, GoOptions{})
	if err != nil {
		return cfg, err
	}
	if goPath != "" {
		cfg.CuratorGo = goPath
	}
	return cfg, nil
}

func findTrustedHostGo(ctx context.Context) string {
	tested := make(map[string]bool)
	for _, fam := range godriver.TestedFamilies() {
		tested[fam] = true
	}
	candidates := []string{
		build.Default.GOROOT,
		"/opt/homebrew/opt/go/libexec",
		"/usr/local/go",
		"/usr/lib/go",
		"/usr/local/opt/go/libexec",
	}
	seen := map[string]bool{}
	binName := "go"
	if runtime.GOOS == "windows" {
		binName = "go.exe"
	}
	for _, root := range candidates {
		if root == "" {
			continue
		}
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		if seen[resolvedRoot] {
			continue
		}
		seen[resolvedRoot] = true
		goBin := filepath.Join(resolvedRoot, "bin", binName)
		info, err := os.Lstat(goBin)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		out, err := exec.CommandContext(ctx, goBin, "version").Output() // #nosec G204 -- fixed trusted GOROOT candidate
		if err != nil {
			continue
		}
		m := goVersionFamilyRe.FindStringSubmatch(strings.TrimSpace(string(out)))
		if len(m) >= 2 && tested[m[1]] {
			return goBin
		}
	}
	return ""
}
