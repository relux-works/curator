package audit

import (
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// Identical effective capabilities hit; changing any capability input misses without reusing the allowance.
func TestVerdictCacheBindsEffectiveCapabilities(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	caps := capabilities.Manifest{
		Network: []string{"github.com"}, Exec: []string{"git"},
		Filesystem: capabilities.Filesystem{Keyword: "repo"},
		Secrets:    []string{"credential-name"}, EnvRead: []string{"HOME"}, PromptScope: "repo",
	}
	subject := subjectWith(t, "curl https://github.com/data\n", caps, 3)
	digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	storeCachedFindings(cfg, digest, subject, nil, nil)
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, subject); !hit {
		t.Fatal("same effective capabilities missed")
	}
	for _, change := range []struct {
		name  string
		apply func(*capabilities.Manifest)
	}{
		{"network", func(c *capabilities.Manifest) { c.Network = nil }},
		{"exec", func(c *capabilities.Manifest) { c.Exec = nil }},
		{"filesystem-keyword", func(c *capabilities.Manifest) { c.Filesystem.Keyword = "home-config" }},
		{"filesystem-paths", func(c *capabilities.Manifest) { c.Filesystem.Paths = []string{"src"} }},
		{"secrets", func(c *capabilities.Manifest) { c.Secrets = nil }},
		{"env-read", func(c *capabilities.Manifest) { c.EnvRead = nil }},
		{"prompt-scope", func(c *capabilities.Manifest) { c.PromptScope = "other" }},
	} {
		t.Run(change.name, func(t *testing.T) {
			probe := subject
			change.apply(&probe.Capabilities)
			if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, probe); hit {
				t.Fatal("changed capabilities reused the allowance")
			}
		})
	}
}
