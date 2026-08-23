package buildrepo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleasedSkillBuildSchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	directory := filepath.Join(root, "schema-cases", "skill-build-v1")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		t.Skipf("supplied conformance root publishes no schema-cases/skill-build-v1: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseDescriptor(payload)
			wantValid := strings.HasPrefix(entry.Name(), "valid")
			if (err == nil) != wantValid {
				t.Fatalf("valid=%v, error=%v", wantValid, err)
			}
		})
	}
}

func TestCanonicalRepositorySourceVectors(t *testing.T) {
	cases := []struct {
		input, identity, transport string
	}{
		{"git@git.example.com:skills/a.git", "git.example.com/skills/a", "ssh"},
		{"https://GIT.example.com/Skills/A.git", "git.example.com/Skills/A", "https"},
		{"ssh://git@git.example.com/skills/a", "git.example.com/skills/a", "ssh"},
		{"https://git.example.com/Skills/A.GIT", "git.example.com/Skills/A.GIT", "https"},
		{"https://git.example.com/文書/工具.git", "git.example.com/文書/工具", "https"},
	}
	for _, testCase := range cases {
		source, err := ParseSource(testCase.input)
		if err != nil {
			t.Errorf("ParseSource(%q): %v", testCase.input, err)
			continue
		}
		if source.Identity != testCase.identity || source.Transport != testCase.transport {
			t.Errorf("ParseSource(%q) = %+v", testCase.input, source)
		}
	}
}

func TestReleasedSourceIdentityVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "source-identities.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Input    string  `json:"input"`
		Identity *string `json:"identity"`
		Error    string  `json:"error"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, vector := range vectors {
		t.Run(vector.Input, func(t *testing.T) {
			source, err := ParseSource(vector.Input)
			if vector.Identity == nil {
				if err == nil {
					t.Fatalf("expected rejection (%s), got %+v", vector.Error, source)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if source.Identity != *vector.Identity {
				t.Fatalf("identity = %q, want %q", source.Identity, *vector.Identity)
			}
		})
	}
}

func TestRepositorySourceRejectsNonRC5Forms(t *testing.T) {
	for _, input := range []string{
		"file:///tmp/a", "http://git.example.com/a", "git://git.example.com/a",
		"https://user@git.example.com/a", "https://git.example.com:8443/a",
		"https://git.example.com/a%2Fb", "https://git.example.com/a?q=1",
		"https://git.example.com/a#x", "https://git.example.com/a/../b",
		"git@git.example.com:a b", "ssh://git@git.example.com/文書/a",
		"https://git.example.com/a\u00a0b", "https://git.example.com/a\u0085b",
		"https://git.example.com/a\u2003b", "https://git.example.com/a\u009fb",
	} {
		if _, err := ParseSource(input); err == nil {
			t.Errorf("ParseSource(%q) unexpectedly succeeded", input)
		}
	}
}

func TestRepositorySourceLimitCountsUnicodeScalars(t *testing.T) {
	const prefix = "https://git.example.com/"
	prefixScalars := len([]rune(prefix))
	valid := prefix + strings.Repeat("界", 4096-prefixScalars)
	if len(valid) <= 4096 {
		t.Fatalf("test source must exceed 4096 UTF-8 bytes, got %d", len(valid))
	}
	if _, err := ParseSource(valid); err != nil {
		t.Fatalf("rc.5-valid 4096-scalar source rejected: %v", err)
	}

	tooManyScalars := valid + "界"
	if _, err := ParseSource(tooManyScalars); err == nil {
		t.Fatal("source above 4096 Unicode scalars unexpectedly succeeded")
	}
	if _, err := ParseSource(prefix + string([]byte{0xff})); err == nil {
		t.Fatal("source containing invalid UTF-8 unexpectedly succeeded")
	}
}

func TestLockedCommitRequiresMatchingImmutableWidth(t *testing.T) {
	valid := []map[string]any{
		{"object_format": "sha1", "hex": strings.Repeat("a", 40)},
		{"object_format": "sha256", "hex": strings.Repeat("b", 64)},
	}
	for _, object := range valid {
		if _, err := ParseLockedCommit(object, "lock"); err != nil {
			t.Fatal(err)
		}
	}
	for _, object := range []map[string]any{
		{"object_format": "sha1", "hex": strings.Repeat("a", 64)},
		{"object_format": "sha256", "hex": strings.Repeat("b", 40)},
		{"object_format": "sha1", "hex": strings.Repeat("A", 40)},
		{"object_format": "sha1", "hex": strings.Repeat("a", 40), "branch": "main"},
	} {
		if _, err := ParseLockedCommit(object, "lock"); err == nil {
			t.Errorf("accepted mutable or mismatched lock: %#v", object)
		}
	}
}

func TestSkillBuildDescriptorClosedTargetsAndContainment(t *testing.T) {
	valid := `{"schema_version":1,"targets":{"root":{"driver":"go-repository-v1","build_root":".","source_dir":"cmd/tool"},"nested":{"driver":"go-repository-v1","build_root":"tools/admin","source_dir":"tools/admin/cmd/admin"}}}`
	descriptor, err := ParseDescriptor([]byte(valid))
	if err != nil || len(descriptor.Targets) != 2 {
		t.Fatalf("valid descriptor = %+v, %v", descriptor, err)
	}
	invalid := []string{
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"other/tool"}}}`,
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"..","source_dir":"."}}}`,
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":".","source_dir":".","output":"bin/tool"}}}`,
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":".","source_dir":".","argv":[]}}}`,
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":".","source_dir":".","credentials":{}}}}`,
		`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":".","source_dir":".","signing":{}}}}`,
	}
	for _, payload := range invalid {
		if _, err := ParseDescriptor([]byte(payload)); err == nil {
			t.Errorf("accepted invalid descriptor: %s", payload)
		}
	}
}

func TestLocalSelectorCanonicalization(t *testing.T) {
	cases := map[string]string{
		"tools/./golden":      "tools/golden",
		"tools/tmp/../golden": "tools/golden",
		"../../tools":         "../../tools",
		"a/../../b":           "../b",
	}
	for input, want := range cases {
		got, err := NormalizeLocalSelector(input)
		if err != nil || got != want {
			t.Errorf("NormalizeLocalSelector(%q) = %q, %v; want %q", input, got, err, want)
		}
		again, err := NormalizeLocalSelector(got)
		if err != nil || again != got {
			t.Errorf("normalization is not idempotent: %q -> %q", got, again)
		}
	}
	for _, input := range []string{"", "/absolute", "a//b", "a/"} {
		if _, err := NormalizeLocalSelector(input); err == nil {
			t.Errorf("accepted invalid selector %q", input)
		}
	}
}

func FuzzNormalizeLocalSelectorIdempotent(f *testing.F) {
	for _, selector := range []string{".", "tools/./golden", "tools/tmp/../golden", "../../tools", "a/../../b", "文書/工具"} {
		f.Add(selector)
	}
	f.Fuzz(func(t *testing.T, selector string) {
		normalized, err := NormalizeLocalSelector(selector)
		if err != nil {
			return
		}
		again, err := NormalizeLocalSelector(normalized)
		if err != nil {
			t.Fatalf("normalized selector %q was rejected: %v", normalized, err)
		}
		if again != normalized {
			t.Fatalf("normalization is not idempotent: %q -> %q", normalized, again)
		}
		if strings.Contains(normalized, "//") {
			t.Fatalf("normalized selector contains an empty component: %q", normalized)
		}
	})
}

func TestLocalIdentityDoesNotExposeHostPath(t *testing.T) {
	identity, err := LocalIdentity("project-123", "tools/tmp/../golden")
	if err != nil {
		t.Fatal(err)
	}
	if identity != "sha256:4c006e6f2d8c9ede6e6d5bc3bce3edea9780e2dfd3b442ef358f51a77b921969" {
		t.Fatalf("local identity = %q", identity)
	}
	if strings.Contains(identity, "tools") || strings.Contains(identity, "/") {
		t.Fatalf("local identity exposes selector: %s", identity)
	}
}
