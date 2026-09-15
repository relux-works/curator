package main
import("fmt";"os";"path/filepath";"github.com/relux-works/curator/internal/contextpkg")
func main(){ roots,err:=filepath.Glob(filepath.Join(os.Args[1],"packages","*"));if err!=nil {panic(err)};for _,root:=range roots {m,e:=contextpkg.LoadManifest(root);if e!=nil {panic(e)};_,e=contextpkg.ValidateModules(root,m,map[string]bool{"claude_code":true,"codex_cli":true,"opencode":true,"pi":true});if e!=nil {panic(e)};fmt.Printf("%s: LoadManifest + ValidateModules OK (%d modules)\n",m.Name,len(m.Modules))}}
