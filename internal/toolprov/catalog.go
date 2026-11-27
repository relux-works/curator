package toolprov

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// PinnedUVVersion is the official astral-sh/uv version in the built-in Aqua
// catalog, used both directly and to bootstrap curated PyPI tools.
const PinnedUVVersion = "0.13.10"

// ProvisionedTool describes a verified executable materialized inside
// <manager-home>/tools/store/.
type ProvisionedTool struct {
	ID             string
	Version        string
	Backend        string // "aqua" or "pypi-uv"
	ExecutablePath string // absolute path inside sealed store
	ExecutableSHA  string // 64-hex SHA-256 of ExecutablePath
	NeedsForwarder bool   // true for venv scripts (pypi-uv) that must execute from their venv path
}

// CatalogOptions customizes catalog network endpoints and HTTP client (for tests).
type CatalogOptions struct {
	Platform            string
	HTTPClient          *http.Client
	AquaRegistryBaseURL string // defaults to "https://raw.githubusercontent.com/aquaproj/aqua-registry/main"
	GitHubAPIBaseURL    string // defaults to "https://api.github.com"
	UVBaseURL           string // optional override for built-in uv archive URL prefix
	UVPins              map[string]aquaPinnedAsset
	PyPIAllowlist       map[string]pypiCatalogEntry
}

type aquaPinnedAsset struct {
	Version  string
	Filename string
	URL      string
	SHA256   string
	Binaries map[string]string // binary name -> path inside extracted dir (or "" to search)
}

// Official SHA-256 digests from astral-sh/uv v0.13.10 GitHub Release .sha256 assets.
var defaultUVPins = map[string]aquaPinnedAsset{
	"darwin-arm64": {
		Version:  PinnedUVVersion,
		Filename: "uv-aarch64-apple-darwin.tar.gz",
		URL:      "https://github.com/astral-sh/uv/releases/download/0.13.10/uv-aarch64-apple-darwin.tar.gz",
		SHA256:   "e419bf07533bdbf48795fc793601c834dd79db57a8c49ce8e25cfe3a8639958d",
		Binaries: map[string]string{"uv": "uv-aarch64-apple-darwin/uv", "uvx": "uv-aarch64-apple-darwin/uvx"},
	},
	"darwin-amd64": {
		Version:  PinnedUVVersion,
		Filename: "uv-x86_64-apple-darwin.tar.gz",
		URL:      "https://github.com/astral-sh/uv/releases/download/0.13.10/uv-x86_64-apple-darwin.tar.gz",
		SHA256:   "153e03197588204c04fc4c81cc4bafbfa4280222a385e4ee36a87aeec0ae63b4",
		Binaries: map[string]string{"uv": "uv-x86_64-apple-darwin/uv", "uvx": "uv-x86_64-apple-darwin/uvx"},
	},
	"linux-amd64": {
		Version:  PinnedUVVersion,
		Filename: "uv-x86_64-unknown-linux-musl.tar.gz",
		URL:      "https://github.com/astral-sh/uv/releases/download/0.13.10/uv-x86_64-unknown-linux-musl.tar.gz",
		SHA256:   "5597970380083a340b7983af05a84824b4db1c62b7059c796935b7fd6dc95fd2",
		Binaries: map[string]string{"uv": "uv-x86_64-unknown-linux-musl/uv", "uvx": "uv-x86_64-unknown-linux-musl/uvx"},
	},
	"linux-arm64": {
		Version:  PinnedUVVersion,
		Filename: "uv-aarch64-unknown-linux-musl.tar.gz",
		URL:      "https://github.com/astral-sh/uv/releases/download/0.13.10/uv-aarch64-unknown-linux-musl.tar.gz",
		SHA256:   "68174084fe40b6300361c5b2038066120dc62c92920c22a0f31f6fc74b066b57",
		Binaries: map[string]string{"uv": "uv-aarch64-unknown-linux-musl/uv", "uvx": "uv-aarch64-unknown-linux-musl/uvx"},
	},
}

type pypiCatalogEntry struct {
	Package        string
	DefaultVersion string
	PythonVersion  string
	Executables    []string
}

// Built-in curated PyPI allowlist for CIP-0007 v1.
var defaultPyPIAllowlist = map[string]pypiCatalogEntry{
	"google-colab-cli": {
		Package:        "google-colab-cli",
		DefaultVersion: "0.8.0",
		PythonVersion:  "3.12",
		Executables:    []string{"colab"},
	},
}

// ProvisionTool resolves and provisions one declared tool dependency (CIP-0007 v1)
// into <managerHome>/tools/store/ and returns the verified executable for alias.
func ProvisionTool(ctx context.Context, managerHome, alias, id, versionConstraint string, opts CatalogOptions) (ProvisionedTool, error) {
	if opts.Platform == "" {
		opts.Platform = CurrentPlatform()
	}
	switch {
	case strings.HasPrefix(id, "aqua/"):
		return provisionAquaTool(ctx, managerHome, alias, strings.TrimPrefix(id, "aqua/"), versionConstraint, opts)
	case strings.HasPrefix(id, "pypi/"):
		return provisionPyPITool(ctx, managerHome, alias, strings.TrimPrefix(id, "pypi/"), versionConstraint, opts)
	default:
		return ProvisionedTool{}, fmt.Errorf("tool_catalog_unknown: unsupported tool catalog namespace %q", id)
	}
}

func provisionAquaTool(ctx context.Context, managerHome, alias, slug, versionConstraint string, opts CatalogOptions) (ProvisionedTool, error) {
	parts := strings.Split(slug, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ProvisionedTool{}, fmt.Errorf("tool_catalog_unknown: invalid Aqua tool id %q", "aqua/"+slug)
	}
	owner, repo := parts[0], parts[1]
	fullID := "aqua/" + slug

	// Fast path: built-in pinned uv catalog.
	if owner == "astral-sh" && repo == "uv" {
		pins := opts.UVPins
		if pins == nil {
			pins = defaultUVPins
		}
		pin, ok := pins[opts.Platform]
		if !ok {
			return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unsupported_platform: no built-in uv asset for %s", opts.Platform)
		}
		url := pin.URL
		if opts.UVBaseURL != "" {
			url = strings.TrimRight(opts.UVBaseURL, "/") + "/" + pin.Filename
		}
		return installAquaTarGz(ctx, managerHome, alias, fullID, owner, repo, pin.Version, opts.Platform, url, pin.SHA256, pin.Binaries, opts.HTTPClient)
	}

	// Dynamic Aqua Registry resolution for any aqua/<owner>/<repo> with SHA-256 checksums.
	return provisionDynamicAquaTool(ctx, managerHome, alias, owner, repo, versionConstraint, opts)
}

func installAquaTarGz(ctx context.Context, managerHome, alias, fullID, owner, repo, version, platform, url, expectedSHA string, expectedBins map[string]string, client *http.Client) (ProvisionedTool, error) {
	storeDir := filepath.Join(StoreRoot(managerHome), "aqua", owner, repo, version, platform)
	if r, err := VerifySealedDir(storeDir); err == nil {
		if rec, ok := r.Executables[alias]; ok {
			return ProvisionedTool{
				ID:             fullID,
				Version:        r.Version,
				Backend:        "aqua",
				ExecutablePath: filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)),
				ExecutableSHA:  rec.SHA256,
			}, nil
		}
		// Fallback to single executable if alias differs from canonical binary name.
		if len(r.Executables) == 1 {
			for _, rec := range r.Executables {
				return ProvisionedTool{
					ID:             fullID,
					Version:        r.Version,
					Backend:        "aqua",
					ExecutablePath: filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)),
					ExecutableSHA:  rec.SHA256,
				}, nil
			}
		}
	}

	parentDir := filepath.Dir(storeDir)
	if err := os.MkdirAll(parentDir, 0o700); err != nil {
		return ProvisionedTool{}, fmt.Errorf("create aqua store parent: %w", err)
	}
	stageDir, err := os.MkdirTemp(parentDir, ".aqua-stage-*")
	if err != nil {
		return ProvisionedTool{}, fmt.Errorf("create aqua stage: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(stageDir)
		}
	}()

	archivePath := filepath.Join(stageDir, "archive.tar.gz")
	if err := DownloadAndVerify(ctx, client, url, expectedSHA, archivePath, "tool_provisioning_digest_mismatch"); err != nil {
		return ProvisionedTool{}, err
	}
	extractDir := filepath.Join(stageDir, "root")
	if err := SafeExtractTarGz(archivePath, extractDir); err != nil {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_extract_failed: %w", err)
	}
	_ = os.Remove(archivePath)

	execMap := map[string]ExecutableRecord{}
	for binName, relHint := range expectedBins {
		rel, err := locateExtractedBinary(extractDir, binName, relHint)
		if err != nil {
			return ProvisionedTool{}, fmt.Errorf("tool_provisioning_missing_binary: %s: %w", binName, err)
		}
		binPath := filepath.Join(extractDir, filepath.FromSlash(rel))
		// #nosec G302 -- owner-only executable in sealed tool store
		if err := os.Chmod(binPath, 0o700); err != nil {
			return ProvisionedTool{}, err
		}
		sha, err := FileSHA256(binPath)
		if err != nil {
			return ProvisionedTool{}, err
		}
		execMap[binName] = ExecutableRecord{RelativePath: rel, SHA256: sha}
	}
	if len(execMap) == 0 {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_missing_binary: no executables recorded for %s", fullID)
	}

	if err := WriteReceipt(extractDir, Receipt{
		Schema:        ReceiptSchema,
		Backend:       "aqua",
		ID:            fullID,
		Version:       version,
		Platform:      platform,
		ArchiveURL:    url,
		ArchiveSHA256: expectedSHA,
		Executables:   execMap,
	}); err != nil {
		return ProvisionedTool{}, err
	}

	_ = os.RemoveAll(storeDir)
	if err := os.Rename(extractDir, storeDir); err != nil {
		return ProvisionedTool{}, fmt.Errorf("publish aqua tool store: %w", err)
	}
	cleanup = false
	_ = os.RemoveAll(stageDir)

	rec, ok := execMap[alias]
	if !ok {
		for _, first := range execMap {
			rec = first
			break
		}
	}
	return ProvisionedTool{
		ID:             fullID,
		Version:        version,
		Backend:        "aqua",
		ExecutablePath: filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)),
		ExecutableSHA:  rec.SHA256,
	}, nil
}

func locateExtractedBinary(root, name, hint string) (string, error) {
	if hint != "" {
		p := filepath.Join(root, filepath.FromSlash(hint))
		if info, err := os.Lstat(p); err == nil && info.Mode().IsRegular() {
			return hint, nil
		}
	}
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if d.Name() == name {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			found = filepath.ToSlash(rel)
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("binary %q not found in extracted archive", name)
	}
	return found, nil
}

type aquaRegistryDoc struct {
	Packages []aquaPackageSpec `yaml:"packages"`
}

type aquaPackageSpec struct {
	Type         string            `yaml:"type"`
	RepoOwner    string            `yaml:"repo_owner"`
	RepoName     string            `yaml:"repo_name"`
	Asset        string            `yaml:"asset"`
	Format       string            `yaml:"format"`
	Replacements map[string]string `yaml:"replacements"`
	Files        []struct {
		Name string `yaml:"name"`
		Src  string `yaml:"src"`
	} `yaml:"files"`
	Checksum *struct {
		Type      string `yaml:"type"`
		Asset     string `yaml:"asset"`
		Algorithm string `yaml:"algorithm"`
	} `yaml:"checksum"`
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func provisionDynamicAquaTool(ctx context.Context, managerHome, alias, owner, repo, versionConstraint string, opts CatalogOptions) (ProvisionedTool, error) {
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	regBase := strings.TrimRight(opts.AquaRegistryBaseURL, "/")
	if regBase == "" {
		regBase = "https://raw.githubusercontent.com/aquaproj/aqua-registry/main"
	}
	regURL := fmt.Sprintf("%s/pkgs/%s/%s/registry.yaml", regBase, owner, repo)
	regBody, status, err := fetchSmall(ctx, client, regURL, 2*1024*1024)
	if err != nil {
		return ProvisionedTool{}, fmt.Errorf("tool_registry_unreachable: fetch Aqua registry for %s/%s: %w", owner, repo, err)
	}
	if status == http.StatusNotFound {
		return ProvisionedTool{}, fmt.Errorf("tool_catalog_unknown: %s/%s not found in Aqua registry", owner, repo)
	}
	if status != http.StatusOK {
		return ProvisionedTool{}, fmt.Errorf("tool_registry_unreachable: Aqua registry HTTP %d for %s/%s", status, owner, repo)
	}

	var doc aquaRegistryDoc
	if err := yaml.Unmarshal(regBody, &doc); err != nil || len(doc.Packages) == 0 {
		return ProvisionedTool{}, fmt.Errorf("tool_registry_invalid: malformed Aqua registry YAML for %s/%s", owner, repo)
	}
	pkg := doc.Packages[0]
	if pkg.Type != "github_release" {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unsupported: Aqua package %s/%s has unsupported type %q", owner, repo, pkg.Type)
	}
	if pkg.RepoOwner == "" {
		pkg.RepoOwner = owner
	}
	if pkg.RepoName == "" {
		pkg.RepoName = repo
	}
	if pkg.Checksum == nil || (pkg.Checksum.Algorithm != "" && !strings.EqualFold(pkg.Checksum.Algorithm, "sha256")) {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unverified: Aqua package %s/%s lacks a SHA-256 checksum declaration", owner, repo)
	}

	ghBase := strings.TrimRight(opts.GitHubAPIBaseURL, "/")
	if ghBase == "" {
		ghBase = "https://api.github.com"
	}
	relURL := fmt.Sprintf("%s/repos/%s/%s/releases/latest", ghBase, pkg.RepoOwner, pkg.RepoName)
	if exact := exactPinnedVersion(versionConstraint); exact != "" {
		tag := exact
		if !strings.HasPrefix(tag, "v") {
			// Try exact tag first; fall back to v-prefixed tag if 404.
			if body, code, err := fetchSmall(ctx, client, fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", ghBase, pkg.RepoOwner, pkg.RepoName, tag), 2*1024*1024); err == nil && code == http.StatusOK {
				return finishDynamicAquaRelease(ctx, managerHome, alias, owner, repo, pkg, body, opts, client)
			}
			tag = "v" + tag
		}
		relURL = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", ghBase, pkg.RepoOwner, pkg.RepoName, tag)
	}
	relBody, status, err := fetchSmall(ctx, client, relURL, 4*1024*1024)
	if err != nil || status != http.StatusOK {
		return ProvisionedTool{}, fmt.Errorf("tool_registry_unreachable: fetch GitHub release %s (HTTP %d): %v", relURL, status, err)
	}
	return finishDynamicAquaRelease(ctx, managerHome, alias, owner, repo, pkg, relBody, opts, client)
}

func finishDynamicAquaRelease(ctx context.Context, managerHome, alias, owner, repo string, pkg aquaPackageSpec, relBody []byte, opts CatalogOptions, client *http.Client) (ProvisionedTool, error) {
	var rel ghRelease
	if err := json.Unmarshal(relBody, &rel); err != nil || rel.TagName == "" {
		return ProvisionedTool{}, fmt.Errorf("tool_registry_invalid: invalid GitHub release payload for %s/%s", owner, repo)
	}
	goos, goarch, _ := strings.Cut(opts.Platform, "-")
	osVal := goos
	if rep, ok := pkg.Replacements[goos]; ok {
		osVal = rep
	}
	archVal := goarch
	if rep, ok := pkg.Replacements[goarch]; ok {
		archVal = rep
	}
	format := pkg.Format
	if format == "" {
		format = "tar.gz"
	}
	render := func(tmpl string) string {
		r := strings.NewReplacer(
			"{{.Version}}", rel.TagName,
			"{{trimV .Version}}", strings.TrimPrefix(rel.TagName, "v"),
			"{{.SemVer}}", strings.TrimPrefix(rel.TagName, "v"),
			"{{.OS}}", osVal,
			"{{.Arch}}", archVal,
			"{{.Format}}", format,
		)
		return r.Replace(tmpl)
	}
	assetName := render(pkg.Asset)
	checksumAssetName := render(strings.ReplaceAll(pkg.Checksum.Asset, "{{.Asset}}", assetName))

	var assetURL, checksumURL string
	for _, a := range rel.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
		}
		if a.Name == checksumAssetName {
			checksumURL = a.BrowserDownloadURL
		}
	}
	if assetURL == "" {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_missing_asset: release %s missing asset %q", rel.TagName, assetName)
	}
	if checksumURL == "" {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unverified: release %s missing checksum asset %q", rel.TagName, checksumAssetName)
	}
	sumBody, status, err := fetchSmall(ctx, client, checksumURL, 1024*1024)
	if err != nil || status != http.StatusOK {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unverified: fetch checksum %s (HTTP %d): %v", checksumURL, status, err)
	}
	expectedSHA := extractSHA256FromChecksumFile(string(sumBody), assetName)
	if expectedSHA == "" {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_unverified: checksum asset %q has no SHA-256 entry for %q", checksumAssetName, assetName)
	}

	bins := map[string]string{}
	for _, f := range pkg.Files {
		if f.Name != "" {
			bins[f.Name] = render(f.Src)
		}
	}
	if len(bins) == 0 {
		bins[repo] = ""
	}
	version := strings.TrimPrefix(rel.TagName, "v")
	return installAquaTarGz(ctx, managerHome, alias, "aqua/"+owner+"/"+repo, owner, repo, version, opts.Platform, assetURL, expectedSHA, bins, client)
}

func provisionPyPITool(ctx context.Context, managerHome, alias, pkgName, versionConstraint string, opts CatalogOptions) (ProvisionedTool, error) {
	allowlist := opts.PyPIAllowlist
	if allowlist == nil {
		allowlist = defaultPyPIAllowlist
	}
	entry, ok := allowlist[pkgName]
	if !ok {
		return ProvisionedTool{}, fmt.Errorf("tool_catalog_unknown: PyPI package %q is not in Curator's curated allowlist", pkgName)
	}
	version := entry.DefaultVersion
	if exact := exactPinnedVersion(versionConstraint); exact != "" {
		version = exact
	}
	fullID := "pypi/" + pkgName
	storeDir := filepath.Join(StoreRoot(managerHome), "pypi", pkgName, version, opts.Platform)

	if r, err := VerifySealedDir(storeDir); err == nil {
		if rec, ok := r.Executables[alias]; ok {
			return ProvisionedTool{
				ID:             fullID,
				Version:        r.Version,
				Backend:        "pypi-uv",
				ExecutablePath: filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)),
				ExecutableSHA:  rec.SHA256,
				NeedsForwarder: true,
			}, nil
		}
	}

	// Provision uv through the sealed Aqua backend first.
	uvTool, err := ProvisionTool(ctx, managerHome, "uv", "aqua/astral-sh/uv", ">="+PinnedUVVersion, opts)
	if err != nil {
		return ProvisionedTool{}, fmt.Errorf("bootstrap uv for %s: %w", fullID, err)
	}

	// Create venv in final location so entry-point shebangs point directly to
	// <storeDir>/venv/bin/python. Receipt is written atomically last as the
	// commit marker; any failure removes storeDir cleanly.
	_ = os.RemoveAll(storeDir)
	if err := os.MkdirAll(storeDir, 0o700); err != nil {
		return ProvisionedTool{}, fmt.Errorf("create pypi store dir: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(storeDir)
		}
	}()

	uvPythonDir := filepath.Join(StoreRoot(managerHome), "uv-python")
	uvCacheDir := filepath.Join(managerHome, "cache", "uv")
	if err := os.MkdirAll(uvPythonDir, 0o700); err != nil {
		return ProvisionedTool{}, err
	}
	if err := os.MkdirAll(uvCacheDir, 0o700); err != nil {
		return ProvisionedTool{}, err
	}
	uvEnv := append(cleanBaseEnv(),
		"UV_PYTHON_INSTALL_DIR="+uvPythonDir,
		"UV_CACHE_DIR="+uvCacheDir,
		"UV_PYTHON_DOWNLOADS=automatic",
		"UV_MANAGED_PYTHON=1",
		"UV_NO_CONFIG=1",
	)

	venvDir := filepath.Join(storeDir, "venv")
	venvCmd := exec.CommandContext(ctx, uvTool.ExecutablePath, "venv", "--python", entry.PythonVersion, venvDir) // #nosec G204 -- sealed manager-provisioned uv binary
	venvCmd.Env = uvEnv
	if out, err := venvCmd.CombinedOutput(); err != nil {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_pypi_venv_failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	venvPython := filepath.Join(venvDir, "bin", "python")
	if runtime.GOOS == "windows" {
		venvPython = filepath.Join(venvDir, "Scripts", "python.exe")
	}
	specArg := fmt.Sprintf("%s==%s", entry.Package, version)
	pipCmd := exec.CommandContext(ctx, uvTool.ExecutablePath, "pip", "install", "--python", venvPython, specArg) // #nosec G204 -- sealed manager-provisioned uv binary and curated package pin
	pipCmd.Env = uvEnv
	if out, err := pipCmd.CombinedOutput(); err != nil {
		return ProvisionedTool{}, fmt.Errorf("tool_provisioning_pypi_install_failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	execMap := map[string]ExecutableRecord{}
	for _, binName := range entry.Executables {
		rel := "venv/bin/" + binName
		if runtime.GOOS == "windows" {
			rel = "venv/Scripts/" + binName + ".exe"
		}
		binPath := filepath.Join(storeDir, filepath.FromSlash(rel))
		info, err := os.Lstat(binPath)
		if err != nil || !info.Mode().IsRegular() {
			return ProvisionedTool{}, fmt.Errorf("tool_provisioning_missing_binary: expected %s in %s", rel, storeDir)
		}
		sha, err := FileSHA256(binPath)
		if err != nil {
			return ProvisionedTool{}, err
		}
		execMap[binName] = ExecutableRecord{RelativePath: rel, SHA256: sha}
	}

	if err := WriteReceipt(storeDir, Receipt{
		Schema:      ReceiptSchema,
		Backend:     "pypi-uv",
		ID:          fullID,
		Version:     version,
		Platform:    opts.Platform,
		Executables: execMap,
	}); err != nil {
		return ProvisionedTool{}, err
	}
	cleanup = false

	rec, ok := execMap[alias]
	if !ok {
		rec = execMap[entry.Executables[0]]
	}
	return ProvisionedTool{
		ID:             fullID,
		Version:        version,
		Backend:        "pypi-uv",
		ExecutablePath: filepath.Join(storeDir, filepath.FromSlash(rec.RelativePath)),
		ExecutableSHA:  rec.SHA256,
		NeedsForwarder: true,
	}, nil
}

func exactPinnedVersion(constraint string) string {
	c := strings.TrimSpace(constraint)
	c = strings.TrimPrefix(c, "==")
	c = strings.TrimPrefix(c, "=")
	if c == "" || c == "latest" || strings.ContainsAny(c, "><^~*, ") {
		return ""
	}
	return c
}

func fetchSmall(ctx context.Context, client *http.Client, url string, maxBytes int64) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "curator-toolprov/1")
	resp, err := client.Do(req) // #nosec G704 -- URL originates exclusively from Aqua/GitHub registry
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func extractSHA256FromChecksumFile(content, assetName string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		candidate := strings.ToLower(fields[0])
		if !isHex64(candidate) {
			continue
		}
		if len(fields) == 1 && len(lines) == 1 {
			return candidate
		}
		if len(fields) >= 2 {
			target := strings.TrimPrefix(fields[1], "*")
			if target == assetName || filepath.Base(target) == assetName {
				return candidate
			}
		}
	}
	return ""
}

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, ch := range s {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func cleanBaseEnv() []string {
	keep := map[string]bool{
		"HOME": true, "USER": true, "TMPDIR": true, "PATH": true,
		"SSL_CERT_FILE": true, "SSL_CERT_DIR": true,
		"HTTP_PROXY": true, "HTTPS_PROXY": true, "NO_PROXY": true,
		"http_proxy": true, "https_proxy": true, "no_proxy": true,
	}
	var out []string
	for _, kv := range os.Environ() {
		k, _, ok := strings.Cut(kv, "=")
		if ok && keep[k] {
			out = append(out, kv)
		}
	}
	return out
}
