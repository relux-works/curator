package install

import ("encoding/json"; "os"; "strings"; "testing"; "github.com/relux-works/curator/internal/audit")
func TestReviewMalformedRecordRev5(t *testing.T) {
 for _, mode := range []string{"trailing-object", "null-pinned", "null-revoked"} {
  t.Run(mode, func(t *testing.T) {
   project, home, _ := strictLocalProject(t, "")
   cfg := draftAuditConfig(home)
   assertGatePassed(t, draftProjectResult(cfg, project, home, false))
   objectPath, reportPath := auditBindingPaths(t, home)
   if mode == "trailing-object" {
    b, _ := os.ReadFile(objectPath)
    if err := os.WriteFile(objectPath, append(b, []byte("{}")...), 0600); err != nil { t.Fatal(err) }
   } else {
    b, _ := os.ReadFile(reportPath)
    var v map[string]any
    if err := json.Unmarshal(b, &v); err != nil { t.Fatal(err) }
    if mode == "null-pinned" { v["pinned"] = nil } else { v["revoked"] = nil }
    b, _ = json.Marshal(v)
    if err := os.WriteFile(reportPath, b, 0600); err != nil { t.Fatal(err) }
    rewriteAuditObject(t, objectPath, map[string]any{"evidence_sha256": audit.EvidenceDigest(b)})
   }
   result := draftProjectResult(cfg, project, home, false)
   if !strings.Contains(strings.Join(result.Errors, ";"), "source_audit") { t.Fatalf("malformed record admitted: %+v", result) }
  })
 }
}
