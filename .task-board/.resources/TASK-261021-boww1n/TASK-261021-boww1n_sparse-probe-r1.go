package codexexclusion

import (
 "fmt"
 "os"
 "path/filepath"
 "testing"
)

// Measures only explicit startup/reload discovery in an external sparse tree.
// Source skill bytes remain committed and present in the original checkout.
// Materialization and late creation are adjacent positive controls, not claims
// of a protected overlay or a complete strict-context adapter.
func TestReviewerExternalSparse(t *testing.T) {
 f:=setup(t)
 original:=f.work
 skill(t,filepath.Join(f.codex,"skills"),"approved","APPROVED_SPARSE_SKILL")
 skill(t,filepath.Join(original,".agents","skills"),"source","SOURCE_ONLY_SKILL")
 skill(t,filepath.Join(original,".codex","skills"),"source-codex","SOURCE_CODEX_SKILL")
 put(t,filepath.Join(original,"AGENTS.md"),"SOURCE_ROOT_INSTRUCTIONS\n")
 put(t,filepath.Join(original,"src","main.txt"),"ordinary source data\n")
 f.config(fmt.Sprintf("\n[projects.%q]\ntrust_level=\"untrusted\"\n", original))
 out:=f.prompt("original-positive","-c","project_doc_max_bytes=0")
 require(t,out,0,"APPROVED_SPARSE_SKILL","SOURCE_ONLY_SKILL","SOURCE_CODEX_SKILL")
 out,code:=f.run("stage-fixture","","/usr/bin/git","add",".");require(t,out,code)
 out,code=f.run("commit-fixture","","/usr/bin/git","-c","user.name=Fixture","-c","user.email=fixture@example.invalid","-c","commit.gpgsign=false","commit","-qm","Synthetic sparse discovery fixture");require(t,out,code)
 external:=filepath.Join(f.root,"werkstatt","view")
 out,code=f.run("external-worktree","","/usr/bin/git","worktree","add","--detach","--no-checkout",external,"HEAD");require(t,out,code)
 f.work=external
 out,code=f.run("sparse-init","","/usr/bin/git","sparse-checkout","init","--no-cone");require(t,out,code)
 out,code=f.run("sparse-set","/src/\n","/usr/bin/git","sparse-checkout","set","--no-cone","--stdin");require(t,out,code)
 out,code=f.run("populate-sparse-index","","/usr/bin/git","read-tree","-mu","HEAD");require(t,out,code)
 if _,err:=os.Stat(filepath.Join(external,"src","main.txt"));err!=nil{t.Fatal("missing included source",err)}
 out,code=f.run("sparse-index-proof","","/usr/bin/git","ls-files","-t");require(t,out,code,"S .agents/skills/source/SKILL.md","H src/main.txt")
 for _,name:=range []string{".agents",".codex","AGENTS.md"} {
  if _,err:=os.Stat(filepath.Join(external,name));!os.IsNotExist(err){t.Fatalf("sparse source unexpectedly materialized: %s",name)}
 }
 if _,err:=os.Stat(filepath.Join(original,".agents","skills","source","SKILL.md"));err!=nil{t.Fatal(err)}
 f.config(fmt.Sprintf("\nproject_doc_max_bytes=0\n[projects.%q]\ntrust_level=\"untrusted\"\n",external))
 out=f.prompt("sparse-startup")
 require(t,out,0,"APPROVED_SPARSE_SKILL")
 absent(t,out,"SOURCE_ONLY_SKILL","SOURCE_CODEX_SKILL","SOURCE_ROOT_INSTRUCTIONS")
 a:=f.api("sparse-reload")
 skills:=encoded(a.call("skills/list",map[string]any{"cwds":[]string{external},"forceReload":true}))
 require(t,skills,0,"APPROVED_SPARSE_SKILL");absent(t,skills,"SOURCE_ONLY_SKILL","SOURCE_CODEX_SKILL","LATE_SPARSE_SKILL")
 skill(t,filepath.Join(external,".agents","skills"),"late","LATE_SPARSE_SKILL")
 skills=encoded(a.call("skills/list",map[string]any{"cwds":[]string{external},"forceReload":true}))
 require(t,skills,0,"APPROVED_SPARSE_SKILL","LATE_SPARSE_SKILL");absent(t,skills,"SOURCE_ONLY_SKILL","SOURCE_CODEX_SKILL")
 out,code=f.run("materialize-source-control","/.agents/\n/.codex/\n","/usr/bin/git","sparse-checkout","add","--stdin");require(t,out,code)
 if _,err:=os.Stat(filepath.Join(external,".agents","skills","source","SKILL.md"));err!=nil{t.Fatal("materialization control failed",err)}
 out=f.prompt("materialized-positive")
 require(t,out,0,"APPROVED_SPARSE_SKILL","SOURCE_ONLY_SKILL","SOURCE_CODEX_SKILL","LATE_SPARSE_SKILL")
 f.save("reviewer-result.txt","External sparse startup and explicit reload exclude unmaterialized source skill; approved home skill retained. Materializing source or creating a late skill makes it load. No lifecycle or write-protection certification.\n")
}
