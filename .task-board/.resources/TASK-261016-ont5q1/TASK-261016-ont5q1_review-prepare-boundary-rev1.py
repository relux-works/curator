from pathlib import Path
import json
root=Path.cwd()
test=root/'cmd/curator/muse_root_context_test.go'
extra=r'''
// Corrupt only the proposed fragment before the real boundary call; the public
// resolve must refuse each shape with no output or marker write.
func TestM1ReviewBoundary(t *testing.T) {
 source := museRootProfile(t, nil)
 if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK { t.Fatal(stderr) }
 marker := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.Muse), envmarker.Name)
 before, err := os.ReadFile(marker); if err != nil { t.Fatal(err) }
 for _, tc := range []struct{shape, diagnostic string}{
 {"extra", "not registry-declared"}, {"literal", envregistry.DiagCredentialStoreUnsupported},
 {"suffix", "does not name its declared parent"}, {"parent", "outside the environments root"},
 {"home", "not registry-declared"}, {"adapter", "fragment environment disagrees with adapter"},
 } {
  t.Run(tc.shape, func(t *testing.T) {
   t.Setenv("CURATOR_M1_REVIEW_POISON", tc.shape)
   code, out, stderr := runProfile(t, source, "env", "resolve", "muse")
   if code != exitFail || out != "" || !strings.Contains(stderr, tc.diagnostic) { t.Fatalf("boundary poison %s: %d %q %s", tc.shape, code, out, stderr) }
   after, err := os.ReadFile(marker); if err != nil || !bytes.Equal(before, after) { t.Fatal("boundary refusal wrote marker") }
  })
 }
}
'''
out=root/'.temp/m1-review/boundary_test.go'; out.write_text(test.read_text()+extra)
p=root/'internal/envprofile/managed.go'
source=p.read_text()
anchor='if err := envfragment.CheckBoundary(adapter, EnvRoot(req.Home), fragment); err != nil {'
assert source.count(anchor)==1
inject=r'''
 if adapter.ID == envregistry.Muse {
 switch os.Getenv("CURATOR_M1_REVIEW_POISON") {
 case "extra": fragment.Env["M1_FOREIGN"] = "file"
 case "literal": fragment.Env["TBH_CREDENTIAL_BACKEND"] = "keychain"
 case "suffix": fragment.Env["XDG_DATA_HOME"] = filepath.Join(plan.parent, "config")
 case "parent": for _, v := range adapter.HomeVariables { fragment.Env[v[0]] = filepath.Join(os.TempDir(), "outside", v[1]) }
 case "home": fragment.Env["HOME"] = plan.parent
 case "adapter": fragment.Environment = "codex_cli"
 }
 }
'''
managed=root/'.temp/m1-review/poison-managed.go'; managed.write_text(source.replace(anchor,inject+anchor))
(root/'.temp/m1-review/boundary-overlay.json').write_text(json.dumps({'Replace':{str(test):str(out),str(p):str(managed)}}))
(root/'.temp/m1-review/boundary-test.txt').write_text(extra+inject)
