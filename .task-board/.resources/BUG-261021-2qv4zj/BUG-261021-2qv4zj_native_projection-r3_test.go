package envprofile
import("encoding/json";"os";"os/exec";"path/filepath";"strings";"testing")
// Compares the native skill publication with the core section 3 projection.
// Repository-only inputs may install, but must not become installed context.
func TestReviewerNativeSkillProjection(t *testing.T){
 binary:=filepath.Join(t.TempDir(),"curator")
 build:=exec.Command("go","build","-o",binary,"./cmd/curator");build.Dir="../.."
 if out,err:=build.CombinedOutput();err!=nil{t.Fatalf("build: %v %s",err,out)}
 for _,tc:=range []struct{name,path string;visible bool}{
 {"research",".research/notes.md",false},{"excluded-name","references/README.md",false},{"excluded-directory","assets/tests/a.md",false},{"visible-declared","references/guide.md",true},
 }{t.Run(tc.name,func(t *testing.T){
 home:=t.TempDir();pinOperatorHome(t,home);homes:=pinHomes(t);ids:=newGitIdentities(t)
 caps:="{}";if tc.visible{caps=`{"network":["api.example.net"]}`}
 files:=map[string]string{"SKILL.md":"---\nname: scoped-guide\ndescription: Context projection fixture\n---\n# Guide\n","agent-skill.json":`{"schema_version":3,"capabilities":`+caps+`}`,tc.path:"Read https://api.example.net/guide\n"}
 skill:=gitRepo(t,files,"v1.0.0");operand:=ids.serve(skill,"https://example.com/skills/scoped-guide")
 profile:=t.TempDir();writeManifestPackage(t,profile,`{"schema_version":1,"name":"scope-profile","version":"1.0.0","requires":{"skills":{"scoped-guide":{"git":"`+operand+`","range":"*"}}}}`,nil)
 cfg,_:=json.Marshal(map[string]any{"schema_version":1,"skills_root":filepath.Join(home,"skills"),"projects":map[string]any{},"audit":map[string]any{"enabled":true,"mode":"strict","fail_on":"high","backend":"null"}})
 config:=filepath.Join(home,"config.json");system:=filepath.Join(home,"system.json")
 if err:=os.WriteFile(config,cfg,0600);err!=nil{t.Fatal(err)};if err:=os.WriteFile(system,[]byte(`{"schema_version":1}`),0600);err!=nil{t.Fatal(err)}
 t.Setenv("CURATOR_CONFIG",config);t.Setenv("CURATOR_SYSTEM_CONFIG",system)
 out,err:=exec.Command(binary,"profile","install",profile).CombinedOutput()
 if err!=nil{t.Fatalf("expected admission: %v %s",err,out)}
 if _,err:=os.Stat(lockPath(home,"scope-profile"));err!=nil{t.Fatal(err)}
 for _,adapter:=range []string{"claude_code","codex_cli"}{
  path:=filepath.Join(homes[adapter],"skills","scoped-guide",filepath.FromSlash(tc.path))
  body,err:=os.ReadFile(path)
  if tc.visible {if err!=nil||!strings.Contains(string(body),"https://api.example.net/guide"){t.Fatalf("visible control absent: %v",err)}} else {
   if err==nil&&strings.Contains(string(body),"https://api.example.net/guide"){t.Errorf("publication mismatch: CLI exit 0, lock exists, %s native skills/scoped-guide/%s contains excluded contact bytes",adapter,tc.path)} else if !os.IsNotExist(err){t.Fatalf("unexpected publication read: %v",err)}
  }
 }
 })}
}
