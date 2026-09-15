package main
import (
 "fmt"
 "os"
 "path/filepath"
 "github.com/relux-works/curator/internal/contextpkg"
)
func main() {
 roots, err := filepath.Glob(filepath.Join(os.Args[1], "packages", "*")); if err != nil { panic(err) }
 for _, root := range roots {
  m, err := contextpkg.LoadManifest(root); if err != nil { panic(err) }
  warnings, err := contextpkg.ValidateModules(root, m, map[string]bool{"claude_code":true,"codex_cli":true,"opencode":true,"pi":true}); if err != nil { panic(err) }; if len(warnings)>0 { panic(warnings) }
  fmt.Println(filepath.Base(root), "PASS")
 }
 if len(roots)!=6 { panic("expected six packages") }
}
