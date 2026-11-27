// Package toolprov implements manager-provisioned CLI tools (CIP-0007 v1 /
// protocol/tool-provisioning.md) and automatic Go toolchain provisioning
// (CIP-0039 / protocol/go-toolchain-provisioning.md) into the sealed
// <manager-home>/tools/store/ tree.
package toolprov

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// ReceiptFilename is the sealed store metadata file written beside each
	// provisioned tool or Go toolchain.
	ReceiptFilename = ".curator-tool-receipt.json"

	// ReceiptSchema is the closed schema identifier for tool receipts.
	ReceiptSchema = "curator-tool-receipt/1"

	maxDownloadBytes   = 512 * 1024 * 1024 // 512 MiB cap for official toolchain/tool archives
	maxArchiveFileSize = 256 * 1024 * 1024 // 256 MiB cap per tar member
	maxArchiveEntries  = 100_000           // generous cap for Go stdlib + tool trees
)

// ExecutableRecord records one verified executable inside a sealed tool store entry.
type ExecutableRecord struct {
	RelativePath string `json:"relative_path"`
	SHA256       string `json:"sha256"`
}

// Receipt is the curator-tool-receipt/1 record persisted in each sealed store entry.
type Receipt struct {
	Schema        string                      `json:"schema"`
	Backend       string                      `json:"backend"` // "go-official", "aqua", "pypi-uv"
	ID            string                      `json:"id"`
	Version       string                      `json:"version"`
	Platform      string                      `json:"platform"`
	ArchiveURL    string                      `json:"archive_url,omitempty"`
	ArchiveSHA256 string                      `json:"archive_sha256,omitempty"`
	Executables   map[string]ExecutableRecord `json:"executables"`
	ProvisionedAt string                      `json:"provisioned_at"`
}

// CurrentPlatform returns the canonical <os>-<arch> platform string.
func CurrentPlatform() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

// StoreRoot returns the sealed tool store root below managerHome.
func StoreRoot(managerHome string) string {
	return filepath.Join(managerHome, "tools", "store")
}

// FileSHA256 computes the lowercase 64-hex SHA-256 digest of path.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 -- caller passes validated store/staging path
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WriteReceipt atomically persists a Receipt inside dir.
func WriteReceipt(dir string, r Receipt) error {
	if r.Schema == "" {
		r.Schema = ReceiptSchema
	}
	if r.ProvisionedAt == "" {
		r.ProvisionedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tool receipt: %w", err)
	}
	raw = append(raw, '\n')
	tmp, err := os.CreateTemp(dir, ".receipt-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp receipt: %w", err)
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
		return fmt.Errorf("write temp receipt: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp receipt: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp receipt: %w", err)
	}
	if err := os.Rename(tmpPath, filepath.Join(dir, ReceiptFilename)); err != nil {
		return fmt.Errorf("publish receipt: %w", err)
	}
	cleanup = false
	return nil
}

// VerifySealedDir reads and verifies .curator-tool-receipt.json and every
// recorded executable inside dir.
func VerifySealedDir(dir string) (*Receipt, error) {
	receiptPath := filepath.Join(dir, ReceiptFilename)
	raw, err := os.ReadFile(receiptPath) // #nosec G304 -- store-contained receipt path
	if err != nil {
		return nil, err
	}
	var r Receipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("tool_receipt_invalid: %w", err)
	}
	if r.Schema != ReceiptSchema || r.Backend == "" || r.ID == "" || r.Version == "" || len(r.Executables) == 0 {
		return nil, fmt.Errorf("tool_receipt_invalid: missing required receipt fields in %s", receiptPath)
	}
	cleanDir := filepath.Clean(dir)
	for name, rec := range r.Executables {
		if rec.RelativePath == "" || strings.HasPrefix(rec.RelativePath, "/") || strings.Contains(rec.RelativePath, "..") {
			return nil, fmt.Errorf("tool_receipt_invalid: unsafe relative path for %q", name)
		}
		target := filepath.Join(cleanDir, filepath.FromSlash(rec.RelativePath))
		if !strings.HasPrefix(target, cleanDir+string(filepath.Separator)) {
			return nil, fmt.Errorf("tool_receipt_invalid: executable %q escapes store root", name)
		}
		info, err := os.Lstat(target)
		if err != nil {
			return nil, fmt.Errorf("tool_receipt_stale: executable %q missing: %w", name, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("tool_receipt_invalid: executable %q is not a regular file", name)
		}
		got, err := FileSHA256(target)
		if err != nil {
			return nil, fmt.Errorf("tool_receipt_stale: hash %q: %w", name, err)
		}
		if got != rec.SHA256 {
			return nil, fmt.Errorf("tool_receipt_drift: executable %q sha256 mismatch", name)
		}
	}
	return &r, nil
}

// DownloadAndVerify streams url to destPath while verifying expectedSHA256
// (lowercase 64-hex). On mismatch it removes destPath and returns errCode.
func DownloadAndVerify(ctx context.Context, client *http.Client, url, expectedSHA256, destPath, errCode string) error {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%s: build request: %w", errCode, err)
	}
	req.Header.Set("User-Agent", "curator-toolprov/1")
	resp, err := client.Do(req) // #nosec G704 -- URL originates exclusively from pinned official/Aqua catalog
	if err != nil {
		return fmt.Errorf("%s: download %s: %w", errCode, url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: download %s returned HTTP %d", errCode, url, resp.StatusCode)
	}
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- private staging file
	if err != nil {
		return fmt.Errorf("%s: create staging file: %w", errCode, err)
	}
	cleanup := true
	defer func() {
		_ = f.Close()
		if cleanup {
			_ = os.Remove(destPath)
		}
	}()

	h := sha256.New()
	limited := io.LimitReader(resp.Body, maxDownloadBytes+1)
	n, err := io.Copy(io.MultiWriter(f, h), limited)
	if err != nil {
		return fmt.Errorf("%s: stream %s: %w", errCode, url, err)
	}
	if n > maxDownloadBytes {
		return fmt.Errorf("%s: archive exceeds %d byte limit", errCode, maxDownloadBytes)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("%s: close staging file: %w", errCode, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, strings.TrimSpace(expectedSHA256)) {
		return fmt.Errorf("%s: sha256 mismatch for %s (got %s, want %s)", errCode, url, got, expectedSHA256)
	}
	cleanup = false
	return nil
}

// SafeExtractTarGz extracts a .tar.gz archive into destDir, rejecting path
// traversal, absolute paths, escaping symlinks, and special file types.
func SafeExtractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath) // #nosec G304 -- verified archive in private staging dir
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer func() { _ = gz.Close() }()

	cleanDest := filepath.Clean(destDir)
	if err := os.MkdirAll(cleanDest, 0o700); err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	count := 0
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}
		count++
		if count > maxArchiveEntries {
			return fmt.Errorf("archive exceeds maximum entry count (%d)", maxArchiveEntries)
		}
		cleanName, err := sanitizeArchiveMember(hdr.Name)
		if err != nil {
			return fmt.Errorf("unsafe archive member %q: %w", hdr.Name, err)
		}
		if cleanName == "." {
			continue
		}
		target := filepath.Join(cleanDest, filepath.FromSlash(cleanName))
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(filepath.Separator)) {
			return fmt.Errorf("archive member %q escapes destination", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
		case tar.TypeReg:
			if hdr.Size < 0 || hdr.Size > maxArchiveFileSize {
				return fmt.Errorf("archive member %q has invalid size %d", hdr.Name, hdr.Size)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			mode := os.FileMode(0o600)
			if hdr.Mode&0o111 != 0 {
				mode = 0o700
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode) // #nosec G304 -- contained target
			if err != nil {
				return err
			}
			written, copyErr := io.CopyN(out, tr, hdr.Size)
			closeErr := out.Close()
			if copyErr != nil {
				return fmt.Errorf("extract %q: %w", hdr.Name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close %q: %w", hdr.Name, closeErr)
			}
			if written != hdr.Size {
				return fmt.Errorf("short write extracting %q", hdr.Name)
			}
		case tar.TypeSymlink:
			linkTarget := hdr.Linkname
			if filepath.IsAbs(linkTarget) {
				return fmt.Errorf("archive symlink %q uses absolute target %q", hdr.Name, linkTarget)
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(target), filepath.FromSlash(linkTarget)))
			if resolved != cleanDest && !strings.HasPrefix(resolved, cleanDest+string(filepath.Separator)) {
				return fmt.Errorf("archive symlink %q escapes destination (%q)", hdr.Name, linkTarget)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			_ = os.Remove(target)
			if err := os.Symlink(linkTarget, target); err != nil {
				return err
			}
		case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
			// Handled transparently by archive/tar; skip if surfaced.
			continue
		default:
			return fmt.Errorf("archive member %q has unsupported entry type %d", hdr.Name, hdr.Typeflag)
		}
	}
	return nil
}

func sanitizeArchiveMember(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty path")
	}
	if strings.ContainsRune(name, '\x00') || strings.ContainsRune(name, '\\') {
		return "", errors.New("contains NUL or backslash")
	}
	if strings.HasPrefix(name, "/") {
		return "", errors.New("absolute path")
	}
	trimmed := strings.TrimPrefix(name, "./")
	if trimmed == "" || trimmed == "." {
		return ".", nil
	}
	for _, part := range strings.Split(trimmed, "/") {
		if part == ".." {
			return "", errors.New("contains parent directory segment")
		}
	}
	clean := filepath.ToSlash(filepath.Clean(trimmed))
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", errors.New("escapes root")
	}
	return clean, nil
}
