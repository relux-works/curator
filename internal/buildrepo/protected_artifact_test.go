package buildrepo

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Drive the production cache writer and reader for each target OS, including
// Windows on non-Windows hosts, and for both receipt namespaces.
func TestProtectedArtifactNameMatchesTargetPlatform(t *testing.T) {
	for _, target := range []struct{ goos, filename, logical string }{
		{"darwin", "artifact", "bin/tool"},
		{"linux", "artifact", "bin/tool"},
		{"windows", "artifact.exe", "bin/tool.exe"},
	} {
		for _, version := range []int{LegacyReceiptSchemaVersion, SourceAwareReceiptSchemaVersion} {
			t.Run(target.goos+"/"+ArtifactsDir(version), func(t *testing.T) {
				input := adoptionArtifactInput(version)
				driverInputOf(input)["target"].(map[string]any)["goos"] = target.goos
				key, err := cacheKey(input)
				if err != nil {
					t.Fatal(err)
				}
				store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
				artifact := []byte("cached executable")
				execution := testExecutionReceiptBytes(t, driverInputOf(input), artifact)
				receipt, err := store.StoreArtifact(key, input, "tool", artifact, execution)
				if err != nil {
					t.Fatal(err)
				}
				entry := filepath.Join(store.Root, ArtifactsDir(version), strings.TrimPrefix(key, "sha256:"))
				physical := filepath.Join(entry, target.filename)
				if got, err := os.ReadFile(physical); err != nil || !bytes.Equal(got, artifact) {
					t.Fatalf("physical artifact = %q, err=%v", got, err)
				}
				var record struct {
					CacheKey string                `json:"cache_key"`
					Artifact struct{ Path string } `json:"artifact"`
				}
				if err := json.Unmarshal(receipt, &record); err != nil {
					t.Fatal(err)
				}
				if record.CacheKey != key || record.Artifact.Path != target.logical {
					t.Fatalf("receipt key=%q path=%q; want %q, %q", record.CacheKey, record.Artifact.Path, key, target.logical)
				}
				hit, err := store.LookupArtifact(key, input, false)
				if err != nil || hit == nil || !bytes.Equal(hit.Bytes, artifact) || !bytes.Equal(hit.Receipt, receipt) {
					t.Fatalf("cache hit=%+v err=%v", hit, err)
				}
				// An extensionless legacy Windows file cannot satisfy the new
				// executable path, even when its bytes and receipt are intact.
				if target.goos == "windows" {
					if err := os.Rename(physical, filepath.Join(entry, "artifact")); err != nil {
						t.Fatal(err)
					}
					if hit, err := store.LookupArtifact(key, input, false); hit != nil || ErrorCode(err) != CodeArtifactInvalid {
						t.Fatalf("extensionless Windows artifact accepted: hit=%+v err=%v", hit, err)
					}
				}
			})
		}
	}
}
