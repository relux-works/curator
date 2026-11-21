package ciartifact

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingCloseEvidenceInput struct {
	*bytes.Reader
	info     os.FileInfo
	closeErr error
}

func (f *failingCloseEvidenceInput) Stat() (os.FileInfo, error) { return f.info, nil }

func (f *failingCloseEvidenceInput) Close() error { return f.closeErr }

func TestSanitizeJSONPreservesGoEventContract(t *testing.T) {
	hostPath := strings.Join([]string{"/", "home", "runner", "checkout", "internal", "fixture.go"}, "")
	input := `{"Time":"2026-01-02T03:04:05Z","Action":"output","Package":"github.com/relux-works/curator/internal/ciartifact","Test":"TestSanitizedEvidence","Elapsed":0.25,"Output":"at ` + hostPath + ` TOKEN=secret-value GOAMD64=v3 runner.internal.invalid 192.0.2.44 environment_store_untrusted: store entry failed ownership check: fixture"}`
	clean, found, err := sanitizeGoTestStream([]byte(input))
	if err != nil {
		t.Fatal("sanitize JSON:", err)
	}
	if !json.Valid(clean) {
		t.Fatal("sanitized event is invalid JSON")
	}
	rawPrefix, _, _ := strings.Cut(input, `,"Output":`)
	cleanPrefix, _, _ := strings.Cut(string(clean), `,"Output":`)
	if rawPrefix == "" || rawPrefix != cleanPrefix {
		t.Fatal("non-Output event bytes changed")
	}
	var event map[string]any
	if err := json.Unmarshal(clean, &event); err != nil {
		t.Fatal("decode sanitized event:", err)
	}
	for key, want := range map[string]any{
		"Time": "2026-01-02T03:04:05Z", "Action": "output", "Package": curatorModulePath + "/internal/ciartifact", "Test": "TestSanitizedEvidence", "Elapsed": 0.25,
	} {
		if event[key] != want {
			t.Errorf("event field %q changed: got %v, want %v", key, event[key], want)
		}
	}
	output, ok := event["Output"].(string)
	if !ok {
		t.Fatal("Output was not preserved as a string")
	}
	for _, forbidden := range []string{hostPath, "secret-value", "GOAMD64=v3", "runner.internal.invalid", "192.0.2.44"} {
		if strings.Contains(output, forbidden) {
			t.Errorf("sensitive fixture survived: %q", forbidden)
		}
	}
	for _, witness := range []string{"environment_store_untrusted:", "failed ownership check:"} {
		if !strings.Contains(output, witness) {
			t.Errorf("compatibility witness %q was removed", witness)
		}
	}
	if found.paths == 0 || found.assignments == 0 || found.hostnames == 0 || found.ipAddresses == 0 {
		t.Errorf("negative controls were not all exercised: %+v", found)
	}
}

func TestSanitizeGoTestStreamNormalizesUnsafeSubtestMetadata(t *testing.T) {
	input := []byte(
		`{"Action":"run","Package":"` + curatorModulePath + `/internal/ciartifact","Test":"TestSanitizedEvidence/worker.internal.invalid","Output":"safe witness"}` + "\n" +
			`{"Action":"run","Package":"` + curatorModulePath + `/internal/ciartifact","Test":"TestSanitizedEvidence/child-/home/fixture/private","Output":"safe witness"}` + "\n" +
			`{"Action":"run","Package":"` + curatorModulePath + `/internal/ciartifact","Test":"TestSanitizedEvidence/safe-child","Output":"safe witness"}` + "\n",
	)
	clean, found, err := sanitizeGoTestStream(input)
	if err != nil {
		t.Fatal("sanitize subtest metadata:", err)
	}
	var got []string
	for _, line := range bytes.Split(bytes.TrimSpace(clean), []byte("\n")) {
		if !json.Valid(line) {
			t.Fatalf("normalized event is invalid JSON: %s", line)
		}
		var event struct {
			Package string
			Test    string
		}
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatal("decode normalized event:", err)
		}
		if event.Package != curatorModulePath+"/internal/ciartifact" {
			t.Fatalf("package metadata changed: %q", event.Package)
		}
		got = append(got, event.Test)
	}
	want := []string{"TestSanitizedEvidence/" + redactedSubtest, "TestSanitizedEvidence/" + redactedSubtest, "TestSanitizedEvidence/safe-child"}
	if len(got) != len(want) {
		t.Fatalf("normalized event count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("normalized test metadata %d = %q, want %q", i, got[i], want[i])
		}
	}
	for _, forbidden := range []string{"worker.internal.invalid", "/home/fixture/private"} {
		if bytes.Contains(clean, []byte(forbidden)) {
			t.Errorf("unsafe subtest metadata survived: %q", forbidden)
		}
	}
	if found.hostnames == 0 || found.paths == 0 {
		t.Fatalf("host and path controls were not both exercised: %+v", found)
	}
}

func TestSanitizeTextIsIdempotent(t *testing.T) {
	input := "https://worker.internal.invalid/private/data TOKEN=fixture-value 192.0.2.44"
	first, _ := sanitizeText(input)
	second, found := sanitizeText(first)
	if first != second {
		t.Fatal("sanitized output changed on a second pass")
	}
	if found != (counts{}) {
		t.Fatalf("second pass found residual sensitive values: %+v", found)
	}
}

func TestSanitizeJSONRedactsSensitiveObjectValues(t *testing.T) {
	fixturePath := "/" + strings.Join([]string{"private", "runner", "item"}, "/")
	mixedValue := "fixture-key-value " + fixturePath + " worker.internal.invalid"
	input := []byte(`{"evidence":{"api_key":"` + mixedValue + `","password":{"nested":"fixture-password-value"}},"safe":"keep this witness"}`)
	clean, _, err := sanitizeTextOrJSON(input)
	if err != nil {
		t.Fatal("sanitize JSON:", err)
	}
	for _, forbidden := range []string{"fixture-key-value", fixturePath, "worker.internal.invalid", "fixture-password-value"} {
		if strings.Contains(string(clean), forbidden) {
			t.Errorf("sensitive object value retained %q", forbidden)
		}
	}
	if !strings.Contains(string(clean), `"api_key":"[redacted credential]"`) || !strings.Contains(string(clean), `"password":"[redacted credential]"`) {
		t.Fatal("sensitive object value survived")
	}
	if !strings.Contains(string(clean), "keep this witness") {
		t.Fatal("safe evidence value was removed")
	}
}

func TestSanitizeGoTestStreamRedactsEscapedPath(t *testing.T) {
	hostPath := "/" + strings.Join([]string{"home", "runner", "checkout", "file.go"}, "/")
	escapedPath := strings.ReplaceAll(hostPath, "/", `\u002f`)
	input := []byte(`{"Action":"output","Package":"example/pkg","Output":"at ` + escapedPath + `:12"}`)
	clean, _, err := sanitizeGoTestStream(input)
	if err != nil {
		t.Fatal("sanitize escaped JSON:", err)
	}
	var event map[string]string
	if err := json.Unmarshal(clean, &event); err != nil {
		t.Fatal("decode sanitized event:", err)
	}
	if strings.Contains(event["Output"], hostPath) {
		t.Fatal("escaped path survived JSON decoding and redaction")
	}
}

func TestSanitizePlainTextAndUnparsedLines(t *testing.T) {
	hostPath := strings.Join([]string{"D:", `\`, "a", `\`, "checkout", `\`, "src", `\`, "item.go"}, "")
	input := []byte("not-json " + hostPath + " host=worker.internal.invalid API_TOKEN=fixture-secret PID=123\n")
	clean, found, err := sanitizeTextOrJSON(input)
	if err != nil {
		t.Fatal("sanitize text:", err)
	}
	for _, forbidden := range []string{hostPath, "worker.internal.invalid", "fixture-secret", "PID=123"} {
		if strings.Contains(string(clean), forbidden) {
			t.Errorf("sensitive fixture survived: %q", forbidden)
		}
	}
	if found.paths == 0 || found.hostnames == 0 || found.assignments == 0 {
		t.Errorf("text negative controls were not all exercised: %+v", found)
	}
}

func TestSanitizeTreeKeepsSourceAndRejectsExistingDestination(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "sanitized")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	fixturePath := "/" + strings.Join([]string{"home", "runner", "private.txt"}, "/")
	original := []byte(`{"Output":"` + fixturePath + ` token=fixture-secret","Action":"output"}` + "\n")
	if err := os.WriteFile(filepath.Join(source, "go-test.json"), original, 0o600); err != nil {
		t.Fatal("write source:", err)
	}
	counts, err := SanitizeTree(source, destination)
	if err != nil {
		t.Fatal("sanitize tree:", err)
	}
	if counts.Files != 1 || counts.Paths == 0 || counts.Credentials == 0 {
		t.Fatalf("unexpected sanitizer totals: %+v", counts)
	}
	gotSource, err := os.ReadFile(filepath.Join(source, "go-test.json"))
	if err != nil {
		t.Fatal("read source:", err)
	}
	if string(gotSource) != string(original) {
		t.Fatal("source evidence was modified")
	}
	gotClean, err := os.ReadFile(filepath.Join(destination, "go-test.json"))
	if err != nil {
		t.Fatal("read sanitized copy:", err)
	}
	if strings.Contains(string(gotClean), fixturePath) || strings.Contains(string(gotClean), "fixture-secret") {
		t.Fatal("sanitized destination retained a forbidden value")
	}
	if _, err := SanitizeTree(source, destination); err == nil {
		t.Fatal("existing destination was accepted")
	}
}

func TestSanitizeTreeRejectsOverlappingAndSymlinkInputs(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	if _, err := SanitizeTree(source, filepath.Join(source, "nested")); err == nil {
		t.Fatal("overlapping destination was accepted")
	}
	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(source, "link")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	if _, err := SanitizeTree(source, filepath.Join(root, "sanitized")); err == nil {
		t.Fatal("symlink entry was accepted")
	}
}

func TestSanitizeTreeRejectsBinaryContent(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "sanitized")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	if err := os.WriteFile(filepath.Join(source, "evidence.bin"), []byte{0, 1, 2}, 0o600); err != nil {
		t.Fatal("write binary fixture:", err)
	}
	if _, err := SanitizeTree(source, destination); err == nil {
		t.Fatal("binary evidence was copied without inspection")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatal("failed sanitization left a destination behind")
	}
}

func TestSanitizeTreeDoesNotPublishWhenEvidenceCloseFails(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "sanitized")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	original := []byte(`{"Output":"token=fixture-secret","Action":"output"}` + "\n")
	if err := os.WriteFile(filepath.Join(source, "go-test.json"), original, 0o600); err != nil {
		t.Fatal("write source:", err)
	}
	closeErr := errors.New("injected close failure")
	openInput := func(_ *os.Root, relative string) (evidenceInput, error) {
		info, err := os.Stat(filepath.Join(source, relative))
		if err != nil {
			return nil, err
		}
		return &failingCloseEvidenceInput{
			Reader:   bytes.NewReader(original),
			info:     info,
			closeErr: closeErr,
		}, nil
	}
	_, err := sanitizeTree(source, destination, os.RemoveAll, openInput)
	if !errors.Is(err, errEvidenceFileClose) {
		t.Fatalf("close failure was not reported: %v", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("close failure published a destination: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal("read staging parent:", err)
	}
	if len(entries) != 1 || entries[0].Name() != "source" {
		t.Fatalf("close failure left staging output behind: %v", entries)
	}
}

func TestSanitizeTreeDoesNotPublishWhenStageRemovalFails(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "sanitized")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	if err := os.WriteFile(filepath.Join(source, "evidence.bin"), []byte{0, 1, 2}, 0o600); err != nil {
		t.Fatal("write binary fixture:", err)
	}
	var stagedPath string
	t.Cleanup(func() {
		if stagedPath != "" {
			if err := os.RemoveAll(stagedPath); err != nil {
				t.Errorf("remove test staging directory: %v", err)
			}
		}
	})
	removeErr := errors.New("injected staging removal failure")
	removeStage := func(path string) error {
		stagedPath = path
		return removeErr
	}
	_, err := sanitizeTree(source, destination, removeStage, openEvidenceInput)
	if !errors.Is(err, errEvidenceStageRemove) || !strings.Contains(err.Error(), "binary content") {
		t.Fatalf("sanitization and cleanup failures were not reported: %v", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stage removal failure published a destination: %v", err)
	}
}

func TestSanitizeTreeSuccessfulPublishSkipsStageCleanup(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "sanitized")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal("create source:", err)
	}
	original := []byte(`{"Action":"output","Output":"safe evidence"}` + "\n")
	if err := os.WriteFile(filepath.Join(source, "go-test.json"), original, 0o600); err != nil {
		t.Fatal("write source evidence:", err)
	}
	removeCalls := 0
	removeStage := func(string) error {
		removeCalls++
		return errors.New("unexpected cleanup after successful publish")
	}
	if _, err := sanitizeTree(source, destination, removeStage, openEvidenceInput); err != nil {
		t.Fatalf("successful publish reported a stage cleanup error: %v", err)
	}
	if removeCalls != 0 {
		t.Fatalf("published stage was cleaned up %d times", removeCalls)
	}
	clean, err := os.ReadFile(filepath.Join(destination, "go-test.json"))
	if err != nil {
		t.Fatal("read published evidence:", err)
	}
	if !strings.Contains(string(clean), "safe evidence") {
		t.Fatalf("published evidence is missing the safe witness: %s", clean)
	}
}

func TestSanitizeTreeRefusesForbiddenNonOutputEventIdentifiers(t *testing.T) {
	tests := []struct {
		name        string
		packagePath string
		testName    string
	}{
		{
			name:        "account path in test name",
			packagePath: curatorModulePath + "/internal/ciartifact",
			testName:    "/home/fixture/account/private",
		},
		{
			name:        "host name in package",
			packagePath: curatorModulePath + "/runner.internal.invalid",
			testName:    "TestSafeMetadata",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			destination := filepath.Join(root, "sanitized")
			if err := os.Mkdir(source, 0o755); err != nil {
				t.Fatal("create source:", err)
			}
			event, err := json.Marshal(map[string]string{
				"Action":  "run",
				"Output":  "safe output",
				"Package": test.packagePath,
				"Test":    test.testName,
				"Time":    "2026-01-02T03:04:05Z",
			})
			if err != nil {
				t.Fatal("encode event fixture:", err)
			}
			if err := os.WriteFile(filepath.Join(source, "go-test.json"), append(event, '\n'), 0o600); err != nil {
				t.Fatal("write event fixture:", err)
			}
			if _, err := SanitizeTree(source, destination); err == nil {
				t.Fatal("forbidden non-Output event identifier was accepted")
			}
			if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("refused event published a destination: %v", err)
			}
		})
	}
}
