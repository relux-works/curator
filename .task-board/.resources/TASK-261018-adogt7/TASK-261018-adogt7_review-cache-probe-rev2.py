import pathlib,json,subprocess
root=pathlib.Path.cwd(); out=root/'.temp/review-adogt7'
test=r'''
// A pre-upgrade public Gate allowance must not suppress new Markdown detection.
func TestReviewPriorMarkdownAllowance(t *testing.T) {
 root := os.Getenv("REVIEW_CACHE_ROOT")
 snapshot := filepath.Join(root, "snapshot")
 if err := os.MkdirAll(snapshot,0700); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(snapshot,"notes.md"), []byte("curl https://undeclared.example.net/x\n"),0600); err != nil { t.Fatal(err) }
 cfg := newCfg(t,"strict","high")
 cfg.Path = filepath.Join(root,"home","config.json")
 subject := Subject{Name:"formatter",Source:"example.com/skills",Git:"https://example.com/skills.git",Commit:strings.Repeat("1",40),Snapshot:snapshot,SchemaVersion:3,Capabilities:capabilities.ImplicitNone(),HashVersion:hashing.VersionV2,Directory:"skills/formatter"}
 subject.Package,_ = SkillPackage(subject.Source,"",subject.Name,subject.Commit,subject.Directory)
 if os.Getenv("REVIEW_PHASE") == "seed" {
   _,errs:=Gate(cfg,[]Subject{subject}); if len(errs)>0 { t.Fatalf("old Gate seed failed: %v",errs) }; return
 }
 fresh:=newCfg(t,"strict","high")
 _,freshErrors:=Gate(fresh,[]Subject{subject})
 if len(freshErrors)==0 || !strings.Contains(strings.Join(freshErrors," "),"audit.capability.network-undeclared") {t.Fatalf("fresh control did not block: %v",freshErrors)}
 t.Logf("fresh Gate blocks: %v",freshErrors)
 _,errs:=Gate(cfg,[]Subject{subject})
 if len(errs)==0 {t.Fatal("prior ruleset-4 allowance bypassed current Markdown network gate")}
}
'''
p=out/'audit_test.go'; p.write_text((root/'internal/audit/audit_test.go').read_text()+test)
old=out/'old-audit.go'; old.write_bytes(subprocess.check_output(['git','show','5ecb68da:internal/audit/audit.go']))
base={str(root/'internal/audit/audit_test.go'):str(p)}
(out/'cache-current.json').write_text(json.dumps({'Replace':base}))
(out/'cache-old.json').write_text(json.dumps({'Replace':dict(base,**{str(root/'internal/audit/audit.go'):str(old)})}))
